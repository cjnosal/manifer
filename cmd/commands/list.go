package commands

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/cjnosal/manifer/lib"
	"github.com/cjnosal/manifer/pkg/scenario"
	"github.com/cjnosal/manifer/pkg/yaml"
)

type listCmd struct {
	allScenarios bool
	printJson    bool

	logger  *log.Logger
	writer  io.Writer
	manifer lib.Manifer
}

var list listCmd

func NewListCommand(l io.Writer, w io.Writer, m lib.Manifer) *cobra.Command {

	list.logger = log.New(l, "", 0)
	list.writer = w
	list.manifer = m

	cobraList := &cobra.Command{
		Use:   "list",
		Short: "list scenarios in selected libraries.",
		Long: `list [--all] (--library <library path>...):
  list scenarios in selected libraries.
`,
		Run:              list.execute,
		TraverseChildren: true,
	}

	cobraList.Flags().StringSliceVarP(&libraryPaths, "library", "l", []string{}, "Path to library file")
	cobraList.Flags().BoolVarP(&list.printJson, "json", "j", false, "Print output in json format")
	cobraList.Flags().BoolVarP(&list.printJson, "allScenarios", "a", false, "Include all referenced libraries")

	return cobraList
}

func (p *listCmd) execute(cmd *cobra.Command, args []string) {

	if len(libraryPaths) == 0 {
		p.logger.Printf("Library not specified")
		p.logger.Printf(cmd.Long)
		os.Exit(1)
	}

	entries, err := p.manifer.ListScenarios(libraryPaths, p.allScenarios)

	if err != nil {
		p.logger.Printf("%v\n  while looking up scenarios", err)
		os.Exit(1)
	}

	var outBytes []byte
	if p.printJson {
		outBytes = p.formatJson(entries)
	} else {
		outBytes = p.formatYaml(entries)
	}

	_, err = p.writer.Write(outBytes)
	if err != nil {
		p.logger.Printf("%v\n  while writing list output", err)
		os.Exit(1)
	}
}

func (p *listCmd) formatJson(entries []scenario.ScenarioEntry) []byte {
	bytes, _ := json.Marshal(entries)
	return bytes
}

func (p *listCmd) formatYaml(entries []scenario.ScenarioEntry) []byte {
	yaml := &yaml.Yaml{}
	bytes, _ := yaml.Marshal(entries)
	return bytes
}
