package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

const (
	ExitCodeOK    int = iota //0
	ExitCodeError int = iota //1
)

func main() {
	err := newApp().Run(os.Args)
	var exitCode = ExitCodeOK
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		exitCode = ExitCodeError
	}
	os.Exit(exitCode)
}

func newApp() *cli.App {
	app := cli.NewApp()
	app.Name = "browser-hb"
	app.HelpName = "bhb"
	app.Usage = "CLI tool to list browser history and bookmarks."
	app.UsageText = "bhb [options]"
	app.Version = "0.0.1"
	app.Author = "lighttiger2505"
	app.Email = "lighttiger2505@gmail.com"
	app.Flags = []cli.Flag{
		// 		cli.StringFlag{
		// 			Name:  "suffix, x",
		// 			Usage: "Diary file suffix",
		// 		},
	}
	app.Commands = cli.Commands{
		{
			Name:    "history",
			Aliases: []string{"s"},
			Usage:   "Show browser history",
			Action:  history,
		},
		{
			Name:    "bookmark",
			Aliases: []string{"b"},
			Usage:   "Show browser bookmark",
			Action:  bookmark,
		},
	}
	return app
}
