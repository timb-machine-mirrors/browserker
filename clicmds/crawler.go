package clicmds

import "github.com/urfave/cli/v2"

func TestCrawlerFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "url",
			Usage: "url as a start point",
			Value: "http://localhost/",
		},
	}
}

func TestCrawler(ctx *cli.Context) error {
	return nil
}
