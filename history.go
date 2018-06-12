package main

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/ryanuber/columnize"
	"github.com/urfave/cli"
)

func history(c *cli.Context) error {
	dbPath := getHistoryPath(runtime.GOOS)

	dbFile, err := os.Open(dbPath)
	if err != nil {
		return err
	}
	tmpFile, err := ioutil.TempFile("/tmp", "bhb")
	if err != nil {
		return err
	}
	if _, err := io.Copy(tmpFile, dbFile); err != nil {
		return err
	}
	readFilePath := tmpFile.Name()
	dbFile.Close()
	tmpFile.Close()

	db, err := sql.Open("sqlite3", readFilePath)
	if err != nil {
		return err
	}
	defer db.Close()

	var q = ""
	q = "select title, url from urls order by last_visit_time desc"
	rows, err := db.Query(q)
	if err != nil {
		panic(err)
	}

	var outputs []string
	for rows.Next() {
		var title string
		var url string
		if err := rows.Scan(&title, &url); err != nil {
			panic(err)
		}
		output := strings.Join([]string{title, url}, "|")
		outputs = append(outputs, output)
	}
	result := columnize.SimpleFormat(outputs)
	fmt.Println(result)
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
