package commands

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"strings"

	"github.com/google/subcommands"

	"github.com/cjnosal/manifer/lib"
	"github.com/cjnosal/manifer/pkg/scenario"
)

type searchCmd struct {
	libraryPaths arrayFlags
	printJson    bool

	logger  *log.Logger
	writer  io.Writer
	manifer lib.Manifer
}

func NewSearchCommand(l io.Writer, w io.Writer, m lib.Manifer) subcommands.Command {
	return &searchCmd{
		logger:  log.New(l, "", 0),
		writer:  w,
		manifer: m,
	}
}

func (*searchCmd) Name() string { return "search" }
func (*searchCmd) Synopsis() string {
	return "search scenarios in selected libraries by name and description."
}
func (*searchCmd) Usage() string {
	return `search (--library <library path>...) (query...):
  search scenarios in selected libraries by name and description.
`
}

func (p *searchCmd) SetFlags(f *flag.FlagSet) {
	f.Var(&p.libraryPaths, "library", "Path to library file")
	f.Var(&p.libraryPaths, "l", "Path to library file")
	f.BoolVar(&p.printJson, "json", false, "Print output in json format")
	f.BoolVar(&p.printJson, "j", false, "Print output in json format")
}

func (p *searchCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	if len(p.libraryPaths) == 0 {
		p.logger.Printf("Library not specified")
		p.logger.Printf(p.Usage())
		return subcommands.ExitFailure
	}

	entries, err := p.manifer.ListScenarios(p.libraryPaths, true)

	if err != nil {
		p.logger.Printf("%v\n  while looking up scenarios", err)
		return subcommands.ExitFailure
	}

	matches := []scenario.ScenarioEntry{}
	for _, e := range entries {
		for _, query := range f.Args() {
			if strings.Contains(e.Name, query) || strings.Contains(e.Description, query) {
				matches = append(matches, e)
				break
			}
		}
	}

	var outBytes []byte
	if p.printJson {
		outBytes = p.formatJson(matches)
	} else {
		outBytes = p.formatPlain(matches)
	}

	_, err = p.writer.Write(outBytes)
	if err != nil {
		p.logger.Printf("%v\n  while writing search output", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func (p *searchCmd) formatJson(entries []scenario.ScenarioEntry) []byte {
	bytes, _ := json.Marshal(entries)
	return bytes
}

func (p *searchCmd) formatPlain(entries []scenario.ScenarioEntry) []byte {
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
