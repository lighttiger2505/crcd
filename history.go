package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"

	_ "github.com/mattn/go-sqlite3"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
)

type History struct {
	Title         string
	URL           string
	LastVisitDate string
}

func history(c *cli.Context) error {
	dbPath := getHistoryPath(runtime.GOOS)

	readFilePath, err := copyHisotryDB(dbPath)
	if err != nil {
		return err
	}
	defer os.Remove(readFilePath)

	lastdate := ""
	if c.String("range") != "" {
		year, month, day, err := parseDate(c.String("range"))
		if err != nil {
			return err
		}
		d := time.Now().AddDate(-1*year, -1*month, -1*day)
		lastdate = d.Format("2006-01-02 15:04:05")
	}

	historys, err := selectHistory(readFilePath, lastdate)
	if err != nil {
		return err
	}

	table := [][]string{}
	for _, history := range historys {
		table = append(table, []string{history.Title, history.URL})
	}

	lines := Format(table, 2, []int{40}, "", []int{1})
	for _, line := range lines {
		fmt.Println(line)
	}

	return nil
}

func getHistoryPath(goos string) string {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	var browserHistoryPath string
	switch goos {
	case "darwin":
		browserHistoryPath = "Library/Application Support/Google/Chrome/Default/History"
	case "windows":
		// browserHistoryPath = ".config/google-chrome/Default/History"
	default:
		browserHistoryPath = ".config/google-chrome/Default/History"
	}

	return filepath.Join(home, browserHistoryPath)
}

func copyHisotryDB(dbPath string) (string, error) {
	dbFile, err := os.Open(dbPath)
	if err != nil {
		return "", err
	}
	tmpFile, err := ioutil.TempFile("/tmp", "bhb")
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(tmpFile, dbFile); err != nil {
		return "", err
	}
	readFilePath := tmpFile.Name()
	dbFile.Close()
	tmpFile.Close()

	return readFilePath, nil
}

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

var unitMap = map[string]string{
	"D": "Day",
	"M": "Month",
	"Y": "Year",
	"d": "Day",
	"m": "Month",
	"y": "Year",
}

func parseDate(s string) (int, int, int, error) {
	// ([0-9]+[a-z]+)+
	orig := s
	var timeDeltaMap = map[string]int{}

	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, 0, 0, nil
	}
	if s == "" {
		return 0, 0, 0, errors.New("invalid date " + orig)
	}
	for s != "" {
		var v int64
		var err error

		// The next character must be [0-9]
		if !('0' <= s[0] && s[0] <= '9') {
			return 0, 0, 0, errors.New("invalid date " + orig)
		}

		// Consume [0-9]*
		v, s, err = leadingInt(s)
		if err != nil {
			return 0, 0, 0, errors.New("invalid duration " + orig)
		}

		// Consume unit.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if '0' <= c && c <= '9' {
				break
			}
		}
		if i == 0 {
			return 0, 0, 0, errors.New("missing unit in date " + orig)
		}
		u := s[:i]
		s = s[i:]
		unit, ok := unitMap[u]
		if !ok {
			return 0, 0, 0, errors.New("unknown unit " + u + " in duration " + orig)
		}
		timeDeltaMap[unit] = int(v)
	}

	return timeDeltaMap["Year"], timeDeltaMap["Month"], timeDeltaMap["Day"], nil
}

var errLeadingInt = errors.New("time: bad [0-9]*") // never printed

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (x int64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > (1<<63-1)/10 {
			// overflow
			return 0, "", errLeadingInt
		}
		x = x*10 + int64(c) - '0'
		if x < 0 {
			// overflow
			return 0, "", errLeadingInt
		}
	}
	return x, s[i:], nil
}
