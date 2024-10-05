package main

import (
	"os/user"
	"path/filepath"
)

func getHistoryPath(goos string) (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
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

	return filepath.Join(u.HomeDir, browserHistoryPath), nil
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
