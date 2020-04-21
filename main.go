package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"gitlab.com/browserker/clicmds"
)

func main() {
	app := cli.NewApp()
	app.Name = "Simple Scanner"
	app.Version = "0.1"
	app.Authors = []*cli.Author{{Name: "isaac dawson", Email: "isaac.dawson@gmail.com"}}
	app.Usage = "Run some DAST goodness baby!"
	app.Commands = []*cli.Command{
		{
			Name:    "testauth",
			Aliases: []string{"ta"},
			Usage:   "test authentication",
			Action:  clicmds.TestAuth,
			Flags:   clicmds.TestAuthFlags(),
		},
		{
			Name:    "crawl",
			Aliases: []string{"c"},
			Usage:   "crawl only",
			Action:  clicmds.Crawler,
			Flags:   clicmds.CrawlerFlags(),
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
