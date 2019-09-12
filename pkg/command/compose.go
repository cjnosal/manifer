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
		logger:    log.New(l, "", 0),
		writer:    w,
		composer:  c,
		executors: em,
	}
}

func (*composeCmd) Name() string     { return "compose" }
func (*composeCmd) Synopsis() string { return "compose a yml file from snippets." }
func (*composeCmd) Usage() string {
	return `compose --template <template path> (--library <library path>...) (--scenario <scenario>...) [--print] [--diff] [-- passthrough flags ...]:
  compose a yml file from snippets.
`
}

func (p *composeCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.templatePath, "template", "", "Path to initial template file")
	f.StringVar(&p.templatePath, "t", "", "Path to initial template file")
	f.Var(&p.libraryPaths, "library", "Path to library file")
	f.Var(&p.libraryPaths, "l", "Path to library file")
	f.Var(&p.scenarios, "scenario", "Scenario name in library")
	f.Var(&p.scenarios, "s", "Scenario name in library")
	f.BoolVar(&p.showPlan, "print", false, "Show snippets and arguments being applied")
	f.BoolVar(&p.showPlan, "p", false, "Show snippets and arguments being applied")
	f.BoolVar(&p.showDiff, "diff", false, "Show diff after each snippet is applied")
	f.BoolVar(&p.showDiff, "d", false, "Show diff after each snippet is applied")
}

func (p *composeCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	if p.templatePath == "" {
		p.logger.Printf("Template path not specified")
		p.logger.Printf(p.Usage())
		return subcommands.ExitFailure
	}

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
		p.logger.Printf("%v\n  while composing output", err)
		return subcommands.ExitFailure
	}

	_, err = p.writer.Write(outBytes)
	if err != nil {
		p.logger.Printf("%v\n  while writing composed output", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
