package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"

	"github.com/cjnosal/manifer/cmd/commands"
	"github.com/cjnosal/manifer/lib"
)

func main() {
	// setup dependencies
	logger := os.Stderr
	writer := os.Stdout

	maniferLib := lib.NewManifer(logger)

	// register subcommands
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(commands.NewComposeCommand(logger, writer, maniferLib), "")
	subcommands.Register(commands.NewListCommand(logger, writer, maniferLib), "")
	subcommands.Register(commands.NewSearchCommand(logger, writer, maniferLib), "")
	subcommands.Register(commands.NewInspectCommand(logger, writer, maniferLib), "")

	// run
	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
