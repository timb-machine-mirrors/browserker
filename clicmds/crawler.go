package clicmds

import "github.com/urfave/cli/v2"

func CrawlerFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "url",
			Usage: "url as a start point",
			Value: "http://localhost/",
		},
		&cli.StringFlag{
			Name:  "config",
			Usage: "config to use",
			Value: "browserk.toml",
		},
	}
}

func Crawler(ctx *cli.Context) error {
	return nil
}
