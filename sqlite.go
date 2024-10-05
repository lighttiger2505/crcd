package main

import "database/sql"

func selectHistory(path string, lastdate string) ([]*History, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// lastdate = "2019-03-27 12:46:22"
	q := "select title, url, last_visit_time from urls order by last_visit_time desc"
	if lastdate != "" {
		q = "select title, url, datetime(last_visit_time / 1000000 + (strftime('%s', '1601-01-01')), 'unixepoch') from urls where datetime(last_visit_time / 1000000 + (strftime('%s', '1601-01-01')), 'unixepoch') >= ? order by last_visit_time desc"
	}
	rows, err := db.Query(q, lastdate)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var historys []*History
	for rows.Next() {
		var title, url, lastVisitDate string
		if err := rows.Scan(&title, &url, &lastVisitDate); err != nil {
			panic(err)
		}
		historys = append(historys, &History{
			Title:         title,
			URL:           url,
			LastVisitDate: lastVisitDate,
		})
	}
	return historys, nil
}
