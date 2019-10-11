package commands

import (
	"context"
	"flag"
	"io"
	"log"

	"github.com/google/subcommands"

	"github.com/cjnosal/manifer/lib"
	"github.com/cjnosal/manifer/pkg/file"
)

type composeCmd struct {
	templatePath string
	libraryPaths arrayFlags
	scenarios    arrayFlags
	showPlan     bool
	showDiff     bool

	manifer lib.Manifer

	logger *log.Logger
	writer io.Writer
}

func NewComposeCommand(l io.Writer, w io.Writer, m lib.Manifer) subcommands.Command {
	return &composeCmd{
		logger:  log.New(l, "", 0),
		writer:  w,
		manifer: m,
	}
}

func (*composeCmd) Name() string     { return "compose" }
func (*composeCmd) Synopsis() string { return "compose a yml file from snippets." }
func (*composeCmd) Usage() string {
	return `compose --template <template path> (--library <library path>...) (--scenario <scenario>...) [--print] [--diff] [-- passthrough flags ...] [\;] :
  compose a yml file from snippets. Use '\;' as a separator when reusing a scenario with different variables.
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

	initialArgs, additionalCompositions := p.split(f.Args())

	libraryPaths := p.libraryPaths
	outBytes, err := p.manifer.Compose(
		p.templatePath,
		libraryPaths,
		p.scenarios,
		initialArgs,
		p.showPlan,
		p.showDiff,
	)

	if err != nil {
		p.logger.Printf("%v\n  while composing initial output", err)
		return subcommands.ExitFailure
	}

	// additionalCompositions will:
	// - invoke Compose using outBytes as template,
	// - accumulate libraryPaths,
	// - preserve plan/diff,
	// - reset scenarios, passthrough args
	for i, comp := range additionalCompositions {
		template := &file.TaggedBytes{Tag: p.templatePath, Bytes: outBytes}
		set := flag.NewFlagSet("additional composition", flag.ContinueOnError)
		set.SetOutput(&nullWriter{}) // suppress default error output
		var newLibraryPaths arrayFlags
		var newScenarios arrayFlags
		set.Var(&newLibraryPaths, "library", "Path to library file")
		set.Var(&newLibraryPaths, "l", "Path to library file")
		set.Var(&newScenarios, "scenario", "Scenario name in library")
		set.Var(&newScenarios, "s", "Scenario name in library")
		err := set.Parse(comp)
		if err != nil {
			p.logger.Printf("%v\n  while parsing flags for composition %d\nUsage for additional compositions:\n", err, i+1)
			set.SetOutput(p.logger.Writer())
			set.PrintDefaults()
			return subcommands.ExitFailure
		}
		libraryPaths = append(libraryPaths, newLibraryPaths...)

		outBytes, err = p.manifer.ComposeFromBytes(
			template,
			libraryPaths,
			newScenarios,
			set.Args(),
			p.showPlan,
			p.showDiff,
		)

		if err != nil {
			p.logger.Printf("%v\n  during composition %d", err, i+1)
			return subcommands.ExitFailure
		}
	}

	_, err = p.writer.Write(outBytes)
	if err != nil {
		p.logger.Printf("%v\n  while writing composed output", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func (p *composeCmd) split(args []string) ([]string, [][]string) {
	comps := [][]string{}

	start := 0
	for i, a := range args {
		if a == ";" {
			if i > start { // omit empty sets
				comps = append(comps, args[start:i])
			}
			start = i + 1
		}
	}
	// trailing args after last ;
	if start < len(args) {
		comps = append(comps, args[start:])
	}

	// args before first ; are part of first composition
	if len(comps) > 0 && args[0] != ";" {
		return comps[0], comps[1:]
	}

	return []string{}, comps
}

type nullWriter struct {
}

func (n *nullWriter) Write(b []byte) (int, error) {
	return len(b), nil
}
