package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/badAkne/order-service/cmd"
)

func main() {
	flag := &cli.BoolFlag{
		Name:    "no-json",
		Value:   true,
		Usage:   "Человеко-читаемый формат для логов вместо JSON",
		Aliases: []string{"nj"},
	}

	app := cli.App{
		Name:     "order-service",
		Version:  "1.0",
		Usage:    "order-service [global options] command [command options]",
		Commands: []*cli.Command{cmd.WebServer()},
		Flags:    []cli.Flag{flag},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
