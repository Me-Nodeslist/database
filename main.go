package main

import (
	"fmt"
	"os"

	"github.com/Me-Nodeslist/database/cmd"
	"github.com/urfave/cli/v2"
)

// @title NodeList API
// @version 1.0
// @description This is a server API for NodeList program
// @host localhost:8088
// @BasePath /v1
func main() {
	local := make([]*cli.Command, 0, 2)
	local = append(local, cmd.ServerRunCmd, cmd.VersionCmd)
	app := cli.App{
		Commands: local,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Show application version",
			},
		},
		Action: func(ctx *cli.Context) error {
			if ctx.Bool("version") {
				fmt.Println(cmd.Version + "+" + cmd.BuildFlag)
			}
			return nil
		},
	}
	app.Setup()

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n\n", err) // nolint:errcheck
		os.Exit(1)
	}
}
