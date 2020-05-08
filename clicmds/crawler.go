package clicmds

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner"
	"gitlab.com/browserker/store"
)

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
		&cli.StringFlag{
			Name:  "datadir",
			Usage: "data directory",
			Value: "browserktmp",
		},
	}
}

func Crawler(ctx *cli.Context) error {
	cfg := &browserk.Config{
		URL:          ctx.String("url"),
		AllowedURLs:  nil,
		ExcludedURLs: nil,
		DataPath:     "",
		AuthScript:   "",
		AuthType:     0,
		Credentials: &browserk.Credentials{
			Username: "",
			Password: "",
			Email:    "",
		},
		NumBrowsers: 3,
	}
	crawl := store.NewCrawlGraph(ctx.String("datadir") + "/crawl")
	attack := store.NewAttackGraph(ctx.String("datadir") + "/attack")
	browserk := scanner.New(cfg, crawl, attack)
	log.Logger.Info().Msg("Starting browserker")

	scanContext := context.Background()
	if err := browserk.Init(scanContext); err != nil {
		log.Logger.Error().Err(err).Msg("failed to init engine")
		return err
	}
	return browserk.Start()
}
