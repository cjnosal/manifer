package commands

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cjnosal/manifer/v2/lib"
	"github.com/cjnosal/manifer/v2/pkg/scenario"
	"github.com/cjnosal/manifer/v2/pkg/yaml"
)

type searchCmd struct {
	printJson bool

	logger  *log.Logger
	writer  io.Writer
	manifer lib.Manifer
}

var search searchCmd

func NewSearchCommand(l io.Writer, w io.Writer, m lib.Manifer) *cobra.Command {

	search.logger = log.New(l, "", 0)
	search.writer = w
	search.manifer = m

	cobraSearch := &cobra.Command{
		Use:   "search",
		Short: "search scenarios in selected libraries by name and description.",
		Long: `search (--library <library path>...) (query...):
  search scenarios in selected libraries by name and description.
`,
		Args:             cobra.MinimumNArgs(1),
		Run:              search.execute,
		TraverseChildren: true,
	}

	cobraSearch.Flags().StringSliceVarP(&libraryPaths, "library", "l", []string{}, "Path to library file")
	cobraSearch.Flags().BoolVarP(&search.printJson, "json", "j", false, "Print output in json format")

	return cobraSearch
}

func (p *searchCmd) execute(cmd *cobra.Command, args []string) {

	if len(libraryPaths) == 0 {
		p.logger.Printf("Library not specified")
		p.logger.Printf(cmd.Long)
		os.Exit(1)
	}

	entries, err := p.manifer.ListScenarios(libraryPaths, true)

	if err != nil {
		p.logger.Printf("%v\n  while looking up scenarios", err)
		os.Exit(1)
	}

	matches := []scenario.ScenarioEntry{}
	for _, e := range entries {
		for _, query := range args {
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
		outBytes = p.formatYaml(matches)
	}

	_, err = p.writer.Write(outBytes)
	if err != nil {
		p.logger.Printf("%v\n  while writing search output", err)
		os.Exit(1)
	}
}

func (p *searchCmd) formatJson(entries []scenario.ScenarioEntry) []byte {
	bytes, _ := json.Marshal(entries)
	return bytes
}

func (p *searchCmd) formatYaml(entries []scenario.ScenarioEntry) []byte {
	yaml := yaml.Yaml{}
	bytes, _ := yaml.Marshal(entries)
	return bytes
}
