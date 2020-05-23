package clicmds

import (
	"context"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/scanner"
	"gitlab.com/browserker/store"
)

var defaultFormValues = browserk.FormData{
	UserName:          "testuser",
	Password:          "testP@assw0rd1",
	FirstName:         "Test",
	MiddleInitial:     "V",
	MiddleName:        "Vuln",
	LastName:          "User",
	FullName:          "Test User",
	Address:           "99 W. 3rd Street",
	AddressLine1:      "Apt B",
	AddressLine2:      "",
	AddressLineExtra:  "",
	StatePrefecture:   "CA",
	Country:           "USA",
	ZipCode:           "90210",
	City:              "Beverly Hills",
	NameOnCard:        "Test User",
	CardNumber:        "4242424242424242",
	CardCVC:           "434",
	ExpirationMonth:   "12",
	ExpirationYear:    "2022",
	Email:             "testuser@test.com",
	PhoneNumber:       "5055151",
	CountryCode:       "+1",
	AreaCode:          "555",
	Extension:         "9024",
	PassportNumber:    "20942422424",
	TravelOrigin:      "NRT",
	TravelDestination: "GCM",
	Default:           "browserker",
	SearchTerm:        "browserker",
	CommentTitle:      "browserker",
	CommentText:       "why yes indeed",
	DocumentName:      "file.txt",
	URL:               "https://example.com/browserker",
	Network:           "192.168.1.1",
	IPV4:              "192.168.1.20",
	IPV6:              "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
}

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
			Value: "",
		},
		&cli.StringFlag{
			Name:  "datadir",
			Usage: "data directory",
			Value: "browserktmp",
		},
		&cli.BoolFlag{
			Name:  "profile",
			Usage: "enable to profile cpu/mem",
			Value: false,
		},
		&cli.IntFlag{
			Name:  "numbrowsers",
			Usage: "max number of browsers to use in parallel",
			Value: 3,
		},
		&cli.IntFlag{
			Name:  "maxdepth",
			Usage: "max depth of nav paths to traverse",
			Value: 10,
		},
	}
}

// Crawler runs browserker crawler
func Crawler(ctx *cli.Context) error {
	if ctx.Bool("profile") {
		go func() {
			http.ListenAndServe(":6060", nil)
		}()
	}

	cfg := &browserk.Config{}
	cfg.FormData = &defaultFormValues

	if ctx.String("config") == "" {
		cfg = &browserk.Config{
			URL:         ctx.String("url"),
			NumBrowsers: ctx.Int("numbrowsers"),
			MaxDepth:    ctx.Int("maxdepth"),
		}
	} else {
		data, err := ioutil.ReadFile(ctx.String("config"))
		if err != nil {
			return err
		}

		if err := toml.NewDecoder(strings.NewReader(string(data))).Decode(cfg); err != nil {
			return err
		}

		if cfg.URL == "" && ctx.String("url") != "" {
			cfg.URL = ctx.String("url")
		}
		if cfg.DataPath == "" && ctx.String("datadir") != "" {
			cfg.DataPath = ctx.String("datadir")
		}
	}
	os.RemoveAll(cfg.DataPath)
	crawl := store.NewCrawlGraph(cfg.DataPath + "/crawl")
	attack := store.NewAttackGraph(cfg.DataPath + "/attack")
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
		log.Info().Msg("Giving a few seconds to sync db...")
		time.Sleep(time.Second * 10)
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

func profile() {

}
