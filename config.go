package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli"
)

type Config struct {
	Profile string
}

var DefaultConfig = &Config{
	Profile: "Default",
}

func LoadConfig() (*Config, error) {
	// 設定ファイルパスを取得
	cfgPath, err := FindCredentialPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(cfgPath); err != nil {
		// 設定ファイルが存在しないならデフォルトの設定でファイルを作る
		if os.IsNotExist(err) {
			cfg, err := CreateNewConfig(cfgPath)
			if err != nil {
				return nil, err
			}
			return cfg, nil
		}
		return nil, err
	}

	// 設定ファイルを読み込み
	cfg := &Config{}
	if _, err := toml.DecodeFile(cfgPath, cfg); err != nil {
		return nil, fmt.Errorf("cannot decode config file, %v", err)
	}
	return cfg, nil
}

func CreateNewConfig(cfgPath string) (*Config, error) {
	f, err := os.Create(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("cannot create default config file, %v", err)
	}
	if err := toml.NewEncoder(f).Encode(DefaultConfig); err != nil {
		return nil, fmt.Errorf("cannot encode default config, %v", err)
	}
	return DefaultConfig, nil
}

func FindCredentialPath() (string, error) {
	appName := "crcd"
	var dir string
	if runtime.GOOS == "windows" {
		dir = os.Getenv("APPDATA")
		if dir == "" {
			dir = filepath.Join(os.Getenv("USERPROFILE"), "Application Data", appName)
		}
		dir = filepath.Join(dir, appName)
	} else {
		dir = filepath.Join(os.Getenv("HOME"), ".config", appName)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("cannot create config directory, %v", err)
	}
	cfgPath := filepath.Join(dir, "config.toml")
	return cfgPath, nil
}

func config(c *cli.Context) error {
	credPath, err := FindCredentialPath()
	if err != nil {
		return err
	}

	editorEnv := os.Getenv("EDITOR")
	if editorEnv == "" {
		editorEnv = "vim"
	}

	cmd := exec.Command(editorEnv, credPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
