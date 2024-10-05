package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/fatih/color"
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

	lines := []string{}
	for _, b := range bs {
		name := color.YellowString(b.Name)
		url := color.HiBlackString(b.URL)
		line := fmt.Sprintf("%s\n%s", name, url)
		lines = append(lines, line)
	}
	selectedURL, err := fzfOpen(lines)
	if err != nil {
		return err
	}
	if selectedURL != "" {
		if err := openbrowser(selectedURL); err != nil {
			return err
		}
	}
	return nil
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
