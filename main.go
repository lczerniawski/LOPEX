package main

import (
	"log"
	"os"

	"github.com/lczerniawski/LOPEX/git"
	"github.com/lczerniawski/LOPEX/svn"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "lopex",
		Usage: "LOPEX is a powerful command-line tool designed to exploit misconfigured web servers and extract leftover files from source control repositories.",
		Commands: []*cli.Command{
			{
				Name:  "all",
				Usage: "Run all available exploits.",
				Action: func(ctx *cli.Context) error {
					err := runGit(ctx)
					if err != nil {
						return err
					}

					err = runMercurial(ctx)
					if err != nil {
						return err
					}

					err = runSvn(ctx)
					if err != nil {
						return err
					}

					return nil
				},
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
						Value:    "repo-dump",
						Usage:    "Output folder for the dumped repo",
						Required: false,
					},
				},
			},
			{
				Name:   "git",
				Usage:  "Try to dump files from git repositories.",
				Action: runGit,
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
						Value:    "repo-dump",
						Usage:    "Output folder for the dumped repo",
						Required: false,
					},
				},
			},
			{
				Name:   "mercurial",
				Usage:  "Try to dump files from mercurial repositories.",
				Action: runMercurial,
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
						Value:    "repo-dump",
						Usage:    "Output folder for the dumped repo",
						Required: false,
					},
				},
			},
			{
				Name:   "svn",
				Usage:  "Try to dump files from subversion repositories.",
				Action: runSvn,
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
						Value:    "repo-dump",
						Usage:    "Output folder for the dumped repo",
						Required: false,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runGit(c *cli.Context) error {
	var urlFlag = c.String("url")
	var outputFolder = c.String("outputFolder")

	println("Try to download git repository files.")
	err := git.TryDownloadGitRepository(urlFlag, outputFolder)
	if err != nil {
		println(err.Error())
	}

	return nil
}

func runMercurial(c *cli.Context) error {
	panic("Unimplemented")
}

func runSvn(c *cli.Context) error {
	var urlFlag = c.String("url")
	var outputFolder = c.String("outputFolder")

	println("Try to download svn repository files.")
	err := svn.TryDownloadSvnRepository(urlFlag, outputFolder)
	if err != nil {
		println(err.Error())
	}

	return nil
}
