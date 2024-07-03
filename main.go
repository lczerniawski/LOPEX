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
		Action: appMain,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "url",
				Aliases:  []string{"u"},
				Value:    "",
				Usage:    "Url to look for the files",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "outputFolder",
				Aliases:  []string{"o"},
				Value:    "git-dump",
				Usage:    "Output folder for the dumped .git files",
				Required: false,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func appMain(c *cli.Context) error {
	var urlFlag = c.String("url")
	var outputFolder = c.String("outputFolder")

	println("Downloading .git files")
	err := git.TryDumpGitRepo(urlFlag, outputFolder)
	if err != nil {
		println(err.Error())
	}

	return nil
}
