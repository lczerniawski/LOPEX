package main

import (
	"log"
	"os"

	"github.com/lczerniawski/LOPEX/git"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:   "lopex",
		Usage:  "LOPEX is a powerful command-line tool designed to exploit misconfigured web servers and extract leftover files from source control repositories.",
		Action: dumpGitRepo,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "url",
				Aliases:  []string{"u"},
				Value:    "",
				Usage:    "Url to look for the files",
				Required: true,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func dumpGitRepo(c *cli.Context) error {
	var urlFlag = c.String("url")

	println("Downloading .git files")
	err := git.TryDumpGitRepo(urlFlag)
	if err != nil {
		println(err.Error())
	}

	return nil
}
