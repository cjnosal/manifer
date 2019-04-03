package command

import (
	"context"
	"flag"
	"io"
	"log"

	"github.com/google/subcommands"

	"github.com/cjnosal/manifer/pkg/scenario"
)

type listCmd struct {
	libraryPaths arrayFlags

	logger *log.Logger
	writer io.Writer
	lister scenario.ScenarioLister
}

func NewListCommand(l io.Writer, w io.Writer, sl scenario.ScenarioLister) subcommands.Command {
	return &listCmd{
		logger: log.New(l, "ListCommand ", 0),
		writer: w,
		lister: sl,
	}
}

func (*listCmd) Name() string     { return "list" }
func (*listCmd) Synopsis() string { return "list scenarios in selected libraries." }
func (*listCmd) Usage() string {
	return `list (-l <library path>...):
  list scenarios in selected libraries.
`
}

func (p *listCmd) SetFlags(f *flag.FlagSet) {
	f.Var(&p.libraryPaths, "l", "Path to library file")
}

func (p *listCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	outBytes, err := p.lister.ListScenarios(p.libraryPaths)

	if err != nil {
		p.logger.Printf("Error looking up scenarios: %v", err)
		return subcommands.ExitFailure
	}

	_, err = p.writer.Write(outBytes)
	if err != nil {
		p.logger.Printf("Error writing list output: %v", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
