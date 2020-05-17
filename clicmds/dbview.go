package clicmds

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"gitlab.com/browserker/browserk"
	"gitlab.com/browserker/store"
)

func DBViewFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "datadir",
			Usage: "data directory",
			Value: "browserktmp",
		},
		&cli.BoolFlag{
			Name:  "navs",
			Usage: "prints navs",
			Value: true,
		},
		&cli.BoolFlag{
			Name:  "urls",
			Usage: "prints urls",
			Value: false,
		},
	}
}

func DBView(ctx *cli.Context) error {
	crawl := store.NewCrawlGraph(ctx.String("datadir"))
	if err := crawl.Init(); err != nil {
		log.Error().Err(err).Msg("failed to init database for viewing")
		return err
	}

	if ctx.Bool("urls") {
		results, err := crawl.GetNavigationResults()
		if err != nil {
			return err
		}

		if results == nil {
			return fmt.Errorf("No result entries found")
		}
		fmt.Printf("Had %d results\n", len(results))
		for _, entry := range results {
			if entry.Messages != nil {
				for _, m := range entry.Messages {
					if m.Request == nil {
						continue
					}
					fmt.Printf("URL visited: %s\n", m.Request.DocumentURL)
				}
			}
		}
		entries := crawl.Find(nil, browserk.NavInProcess, browserk.NavInProcess, 999)
		fmt.Printf("Had %d entries\n", len(entries))

		for _, paths := range entries {
			fmt.Printf("Path: \n")
			for i, path := range paths {
				if len(paths)-1 == i {
					fmt.Printf("%s %s\n", browserk.ActionTypeMap[path.Action.Type], printActionDetails(path.Action))
					break
				}
				fmt.Printf("%s %s -> ", browserk.ActionTypeMap[path.Action.Type], printActionDetails(path.Action))

			}
		}
	}
	log.Info().Msg("Closing db & syncing, please wait")
	err := crawl.Close()
	time.Sleep(5 * time.Second)
	return err
}

func printActionDetails(act *browserk.Action) string {
	ret := ""
	switch act.Type {
	case browserk.ActLoadURL:
		ret += "[" + string(act.Input) + "]"
	case browserk.ActLeftClick:
		ret += "[" + browserk.HTMLTypeToStrMap[act.Element.Type] + " "
		for k, v := range act.Element.Attributes {
			ret += k + "=" + v
		}
		ret += "]"
	case browserk.ActFillForm:
		ret += "[ FORM "
		for k, v := range act.Form.Attributes {
			ret += k + "=" + v
		}
		ret += "]"
	}
	return ret
}
