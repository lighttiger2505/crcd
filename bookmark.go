package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
)

type BookmarkRoot struct {
	Checksum string `json:"checksum"`
	Version  int    `json:"version"`
	Roots    struct {
		SyncTransactionVersion string           `json:"sync_transaction_version"`
		BookmarkBar            BookmarkChildren `json:"bookmark_bar"`
		Other                  BookmarkChildren `json:"other"`
		Synced                 BookmarkChildren `json:"synced"`
	} `json:"roots"`
}

type BookmarkChildren struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	Type                   string `json:"type"`
	URL                    string `json:"url,omitempty"`
	DateAdded              string `json:"date_added"`
	DateModified           string `json:"date_modified,omitempty"`
	SyncTransactionVersion string `json:"sync_transaction_version"`
	MetaInfo               struct {
		LastVisitedDesktop string `json:"last_visited_desktop"`
	} `json:"meta_info,omitempty"`
	Children []BookmarkChildren `json:"children,omitempty"`
}

func bookmark(c *cli.Context) error {
	dbPath := getBookmarkPath(runtime.GOOS)

	bytes, err := ioutil.ReadFile(dbPath)
	if err != nil {
		return err
	}

	var root BookmarkRoot
	if err := json.Unmarshal(bytes, &root); err != nil {
		return err
	}

	bs := collectBookmark(root.Roots.BookmarkBar.Children)

	table := [][]string{}
	for _, b := range bs {
		table = append(table, []string{b.Name, b.URL})
	}

	lines := Format(table, 2, []int{40}, "", []int{1})
	for _, line := range lines {
		fmt.Println(line)
	}
	return nil
}

func getBookmarkPath(goos string) string {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	var browserBookmarkPath string
	switch goos {
	case "darwin":
		browserBookmarkPath = "Library/Application Support/Google/Chrome/Default/Bookmarks"
	case "windows":
		// browserBookmarkPath = ".config/google-chrome/Default/History"
	default:
		browserBookmarkPath = ".config/google-chrome/Default/Bookmarks"
	}

	return filepath.Join(home, browserBookmarkPath)
}

func collectBookmark(children []BookmarkChildren) []BookmarkChildren {
	var results []BookmarkChildren
	for _, child := range children {
		if child.Type == "url" {
			results = append(results, child)
		}
		results = append(results, collectBookmark(child.Children)...)
	}
	return results
}
