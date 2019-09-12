package command

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"strings"

	"github.com/google/subcommands"

	"github.com/cjnosal/manifer/pkg/scenario"
)

type listCmd struct {
	libraryPaths arrayFlags
	allScenarios bool
	printJson    bool

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
	f.BoolVar(&p.printJson, "json", false, "Print output in json format")
	f.BoolVar(&p.printJson, "j", false, "Print output in json format")
}

func (p *listCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	if len(p.libraryPaths) == 0 {
		p.logger.Printf("Library not specified")
		p.logger.Printf(p.Usage())
		return subcommands.ExitFailure
	}

	entries, err := p.lister.ListScenarios(p.libraryPaths, p.allScenarios)

	if err != nil {
		p.logger.Printf("%v\n  while looking up scenarios", err)
		return subcommands.ExitFailure
	}

	var outBytes []byte
	if p.printJson {
		outBytes = p.formatJson(entries)
	} else {
		outBytes = p.formatPlain(entries)
	}

	_, err = p.writer.Write(outBytes)
	if err != nil {
		p.logger.Printf("%v\n  while writing list output", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func (p *listCmd) formatJson(entries []scenario.ScenarioEntry) []byte {
	bytes, _ := json.Marshal(entries)
	return bytes
}

func (p *listCmd) formatPlain(entries []scenario.ScenarioEntry) []byte {
	builder := strings.Builder{}
	for _, entry := range entries {
		builder.WriteString(entry.Name)
		builder.WriteString("\n\t")
		description := entry.Description
		if len(description) == 0 {
			description = "no description"
		}
		builder.WriteString(description)
		builder.WriteString("\n\n")
	}
	return []byte(builder.String())
}
