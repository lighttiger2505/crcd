package main

import (
	"os/user"
	"path/filepath"
)

func getHistoryPath(goos string) (string, error) {
	profilePath, err := getChromeProfilePath(goos)
	if err != nil {
		return "", err
	}
	return filepath.Join(profilePath, "History"), nil
}

func getBookmarkPath(goos string) (string, error) {
	profilePath, err := getChromeProfilePath(goos)
	if err != nil {
		return "", err
	}
	return filepath.Join(profilePath, "Bookmarks"), nil
}

func getChromeProfilePath(goos string) (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	cfg, err := LoadConfig()
	if err != nil {
		return "", err
	}

	var chromeProfilePath string
	switch goos {
	case "darwin":
		chromeProfilePath = "Library/Application Support/Google/Chrome"
	case "windows":
		// chromeProfilePath = ".config/google-chrome/Default/History"
	default:
		chromeProfilePath = ".config/google-chrome"
	}
	return filepath.Join(u.HomeDir, chromeProfilePath, cfg.Profile), nil
}
