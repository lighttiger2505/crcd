package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/urfave/cli"
)

const (
	TypeURL    = "url"
	TypeFolder = "folder"
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
	// Append parend folders
	Path string
}

func (c *BookmarkChildren) fullpath() string {
	return c.Path + "/" + c.Name
}

func bookmark(c *cli.Context) error {
	dbPath, err := getBookmarkPath(runtime.GOOS)
	if err != nil {
		return err
	}

	bytes, err := os.ReadFile(dbPath)
	if err != nil {
		return err
	}

	var root BookmarkRoot
	if err := json.Unmarshal(bytes, &root); err != nil {
		return err
	}

	bs := collectBookmarkWithPath(root.Roots.BookmarkBar.Children, "")

	table := [][]string{}
	for _, b := range bs {
		table = append(table, []string{b.fullpath(), b.URL})
	}

	lines := Format(table, 2, []int{96}, "", []int{1})
	for _, line := range lines {
		fmt.Println(line)
	}
	return nil
}

func getBookmarkPath(goos string) (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
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

	return filepath.Join(u.HomeDir, browserBookmarkPath), nil
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

func collectBookmarkWithPath(children []BookmarkChildren, p string) []BookmarkChildren {
	var results []BookmarkChildren
	for _, child := range children {
		if child.Type == TypeFolder {
			results = append(results, collectBookmarkWithPath(child.Children, p+"/"+child.Name)...)
		}
		if child.Type == TypeURL {
			child.Path = p
			results = append(results, child)
		}
	}
	return results
}
