package main

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
)

type History struct {
	Title string
	URL   string
}

func history(c *cli.Context) error {
	dbPath := getHistoryPath(runtime.GOOS)

	readFilePath, err := copyHisotryDB(dbPath)
	if err != nil {
		return err
	}

	historys, err := selectHistory(readFilePath)
	if err != nil {
		return err
	}

	for _, history := range historys {
		fmt.Println(history.Title, history.URL)
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

func selectHistory(path string) ([]*History, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var q = ""
	q = "select title, url from urls order by last_visit_time desc limit 100"
	rows, err := db.Query(q)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var historys []*History
	for rows.Next() {
		var title, url string
		if err := rows.Scan(&title, &url); err != nil {
			panic(err)
		}
		historys = append(historys, &History{
			Title: title,
			URL:   url,
		})
	}
	return historys, nil
}
