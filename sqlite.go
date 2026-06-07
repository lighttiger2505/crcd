package main

import (
	"database/sql"
	"net/url"
	"time"
)

// webkitEpochOffset is the number of seconds between the WebKit epoch
// (1601-01-01 00:00:00 UTC) and the Unix epoch (1970-01-01 00:00:00 UTC).
const webkitEpochOffset = int64(11644473600)

// maxVisitSamples is the maximum number of individual visit timestamps sampled per URL.
const maxVisitSamples = 10

type History struct {
	Title         string
	URL           string
	VisitCount    int
	LastVisitTime time.Time
	Visits        []time.Time // sample of recent visit times (most recent first)
}

func selectHistory(path string, minTime *time.Time) (histories []*History, err error) {
	dsn := (&url.URL{Scheme: "file", Path: path, RawQuery: "immutable=1"}).String()
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := db.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	q := `
SELECT u.title, u.url, u.visit_count, u.last_visit_time, v.visit_time
FROM urls u
JOIN visits v ON v.url = u.id
ORDER BY u.id, v.visit_time DESC
`
	var args []interface{}
	if minTime != nil {
		q = `
SELECT u.title, u.url, u.visit_count, u.last_visit_time, v.visit_time
FROM urls u
JOIN visits v ON v.url = u.id
WHERE u.last_visit_time >= ?
ORDER BY u.id, v.visit_time DESC
`
		// Convert time.Time to WebKit microsecond timestamp for the filter.
		webkitMin := (minTime.Unix() + webkitEpochOffset) * 1000000
		args = append(args, webkitMin)
	}

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	_, tz := time.Now().Zone()
	loc := time.FixedZone("Local", tz)

	urlIndex := map[string]int{} // url -> index in histories

	for rows.Next() {
		var (
			title       string
			urlStr      string
			visitCount  int
			lastVisitWK int64
			visitWK     int64
		)
		if err := rows.Scan(&title, &urlStr, &visitCount, &lastVisitWK, &visitWK); err != nil {
			return nil, err
		}

		visitTime := webkitToTime(visitWK).In(loc)

		idx, exists := urlIndex[urlStr]
		if !exists {
			h := &History{
				Title:         title,
				URL:           urlStr,
				VisitCount:    visitCount,
				LastVisitTime: webkitToTime(lastVisitWK).In(loc),
			}
			urlIndex[urlStr] = len(histories)
			histories = append(histories, h)
			idx = len(histories) - 1
		}

		if len(histories[idx].Visits) < maxVisitSamples {
			histories[idx].Visits = append(histories[idx].Visits, visitTime)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return histories, nil
}

func webkitToTime(webkitTimestamp int64) time.Time {
	// WebKit timestamps are microseconds since 1601-01-01 00:00:00 UTC.
	seconds := webkitTimestamp / 1000000
	return time.Unix(seconds-webkitEpochOffset, 0).UTC()
}
