package command

import (
	"context"
	"flag"
	"io"
	"log"

	"github.com/google/subcommands"

	"github.com/cjnosal/manifer/pkg/composer"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/plan"
)

type composeCmd struct {
	templatePath string
	libraryPaths arrayFlags
	scenarios    arrayFlags
	passthrough  arrayFlags
	showPlan     bool
	showDiff     bool

	composer  composer.Composer
	executors map[library.Type]plan.Executor

	logger *log.Logger
	writer io.Writer
}

func NewComposeCommand(l io.Writer, w io.Writer, c composer.Composer, em map[library.Type]plan.Executor) subcommands.Command {
	return &composeCmd{
		logger:    log.New(l, "ComposeCommand ", 0),
		writer:    w,
		composer:  c,
		executors: em,
	}
}

func (*composeCmd) Name() string     { return "compose" }
func (*composeCmd) Synopsis() string { return "compose a yml file from snippets." }
func (*composeCmd) Usage() string {
	return `compose -t <template path> (-l <library path>...) (-s <scenario>...) [-p] [-d] [-- passthrough flags ...]:
  compose a yml file from snippets.
`
}

func (p *composeCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.templatePath, "t", "", "Path to template file")
	f.Var(&p.libraryPaths, "l", "Path to library file")
	f.Var(&p.scenarios, "s", "Scenario name in library")
	f.BoolVar(&p.showPlan, "p", false, "Show snippets and arguments being applied")
	f.BoolVar(&p.showDiff, "d", false, "Show diff after each snippet is applied")
}

func (p *composeCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	// TODO load libary, use type to select interpolator
	outBytes, err := p.composer.Compose(p.executors[library.OpsFile],
		p.templatePath,
		p.libraryPaths,
		p.scenarios,
		f.Args(),
		p.showPlan,
		p.showDiff,
	)

	if err != nil {
		p.logger.Printf("Error composing output: %v", err)
		return subcommands.ExitFailure
	}

	_, err = p.writer.Write(outBytes)
	if err != nil {
		p.logger.Printf("Error writing composed output: %v", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
