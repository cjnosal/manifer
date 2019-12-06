package commands

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/cjnosal/manifer/lib"
	"github.com/cjnosal/manifer/pkg/file"
)

type composeCmd struct {
	templatePath string
	scenarios    []string
	showPlan     bool
	showDiff     bool

	manifer lib.Manifer

	logger *log.Logger
	writer io.Writer
}

var compose composeCmd

func NewComposeCommand(l io.Writer, w io.Writer, m lib.Manifer) *cobra.Command {

	compose.logger = log.New(l, "", 0)
	compose.writer = w
	compose.manifer = m

	cobraCompose := &cobra.Command{
		Use:   "compose",
		Short: "compose a yml file from snippets.",
		Long: `compose --template <template path> (--library <library path>...) (--scenario <scenario>...) [--print] [--diff] [-- passthrough flags ...] [\;] :
  compose a yml file from snippets. Use '\;' as a separator when reusing a scenario with different variables.
`,
		Run:              compose.execute,
		TraverseChildren: true,
	}

	cobraCompose.Flags().StringVarP(&compose.templatePath, "template", "t", "", "Path to initial template file")
	cobraCompose.Flags().StringSliceVarP(&libraryPaths, "library", "l", []string{}, "Path to library file")
	cobraCompose.Flags().StringSliceVarP(&compose.scenarios, "scenario", "s", []string{}, "Scenario name in library")
	cobraCompose.Flags().BoolVarP(&compose.showPlan, "print", "p", false, "Show snippets and arguments being applied")
	cobraCompose.Flags().BoolVarP(&compose.showDiff, "diff", "d", false, "Show diff after each snippet is applied")

	return cobraCompose
}

func (p *composeCmd) execute(cmd *cobra.Command, args []string) {

	if p.templatePath == "" {
		p.logger.Printf("Template path not specified")
		p.logger.Printf(cmd.Long)
		os.Exit(1)
	}

	initialArgs, additionalCompositions := p.split(args)

	libraryPaths := libraryPaths
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
		os.Exit(1)
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
			os.Exit(1)
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
			os.Exit(1)
		}
	}

	_, err = p.writer.Write(outBytes)
	if err != nil {
		p.logger.Printf("%v\n  while writing composed output", err)
		os.Exit(1)
	}
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
