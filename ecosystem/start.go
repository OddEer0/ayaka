package ecosystem

import (
	"fmt"
	ayaka "github.com/OddEer0/ayaka/core"
	"os"

	cli "github.com/jawher/mow.cli"
)

const CliHelpString = `
CLI usage commands:

bin [COMMAND]

bin - binary file

COMMAND:
1) version | getting app version
2) run     | starting application
`

type Printer interface {
	Printf(format string, args ...interface{})
}

type DefaultPrinter struct{}

func (p DefaultPrinter) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func StartWithCli[T any](app *ayaka.App[T], print Printer) error {
	var err error
	info := app.Info()

	if print == nil {
		print = DefaultPrinter{}
	}

	cliApp := cli.App(info.Name, info.Description)

	cliApp.Command("version", "get version", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			print.Printf("VERSION: %s\n", info.Version)
		}
	})

	cliApp.Command("help", "cli usage help", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			print.Printf("%s\n", CliHelpString)
		}
	})

	cliApp.Command("run", "run application", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			err = app.Start()
		}
	})

	if e := cliApp.Run(os.Args); e != nil {
		return e
	}

	return err
}
