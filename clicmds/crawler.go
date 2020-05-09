package clicmds

import (
	"context"
	"os"
	"os/signal"
	"syscall"

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
	os.RemoveAll(ctx.String("datadir"))
	crawl := store.NewCrawlGraph(ctx.String("datadir") + "/crawl")
	attack := store.NewAttackGraph(ctx.String("datadir") + "/attack")
	browserk := scanner.New(cfg, crawl, attack)
	log.Logger.Info().Msg("Starting browserker")

	scanContext := context.Background()
	if err := browserk.Init(scanContext); err != nil {
		log.Logger.Error().Err(err).Msg("failed to init engine")
		return err
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info().Msg("Ctrl-C Pressed, shutting down")
		err := browserk.Stop()
		if err != nil {
			log.Error().Err(err).Msg("failed to stop browserk")
		}
		os.Exit(1)
	}()

	err := browserk.Start()
	if err != nil {
		log.Error().Err(err).Msg("browserk failure occurred")
	}

	return browserk.Stop()
}
