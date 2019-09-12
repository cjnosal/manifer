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
	allScenarios bool

	logger *log.Logger
	writer io.Writer
	lister scenario.ScenarioLister
}

func NewListCommand(l io.Writer, w io.Writer, sl scenario.ScenarioLister) subcommands.Command {
	return &listCmd{
		logger: log.New(l, "", 0),
		writer: w,
		lister: sl,
	}
}

func (*listCmd) Name() string     { return "list" }
func (*listCmd) Synopsis() string { return "list scenarios in selected libraries." }
func (*listCmd) Usage() string {
	return `list [--all] (--library <library path>...):
  list scenarios in selected libraries.
`
}

func (p *listCmd) SetFlags(f *flag.FlagSet) {
	f.Var(&p.libraryPaths, "library", "Path to library file")
	f.Var(&p.libraryPaths, "l", "Path to library file")
	f.BoolVar(&p.allScenarios, "all", false, "Include all referenced libraries")
	f.BoolVar(&p.allScenarios, "a", false, "Include all referenced libraries")
}

func (p *listCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	if len(p.libraryPaths) == 0 {
		p.logger.Printf("Library not specified")
		p.logger.Printf(p.Usage())
		return subcommands.ExitFailure
	}

	outBytes, err := p.lister.ListScenarios(p.libraryPaths, p.allScenarios)

	if err != nil {
		p.logger.Printf("%v\n  while looking up scenarios", err)
		return subcommands.ExitFailure
	}

	_, err = p.writer.Write(outBytes)
	if err != nil {
		p.logger.Printf("%v\n  while writing list output", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
