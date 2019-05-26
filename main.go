package main

import (
	"os"

	"github.com/gsblue/dynamite/cmd"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Usage = "tool to archive and restore dynamoDB tables"
	app.Version = "0.0.1"
	app.EnableBashCompletion = true
	app.Commands = []cli.Command{
		cmd.BuildArchive(),
		cmd.BuildRestore(),
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
