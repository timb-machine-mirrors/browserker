package clicmds

import "github.com/urfave/cli/v2"

func TestAuthFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "user",
			Usage: "username to auth with",
			Value: "test@test.com",
		},
		&cli.StringFlag{
			Name:  "pass",
			Usage: "password to auth with",
			Value: "testtest",
		},
		&cli.StringFlag{
			Name:  "url",
			Usage: "url to authenticate to",
			Value: "http://localhost/login",
		},
	}
}

func TestAuth(ctx *cli.Context) error {
	return nil
}
