package ecosystem

import (
	"fmt"
	ayaka "github.com/illusory-server/ayaka/core"
	cli "github.com/jawher/mow.cli"
	"os"
)

const helpString = `
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

type DefaultPrint struct{}

func (p DefaultPrint) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func StartWithCli(app *ayaka.App, print Printer) error {
	var err error
	info := app.Info()

	if print == nil {
		print = DefaultPrint{}
	}

	cliApp := cli.App(info.Name, info.Description)

	cliApp.Command("version", "get version", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			print.Printf("VERSION: %s\n", info.Version)
		}
	})

	cliApp.Command("help", "cli usage help", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			print.Printf("%s\n", helpString)
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
