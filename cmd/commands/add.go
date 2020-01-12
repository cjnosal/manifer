package commands

import (
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/cjnosal/manifer/v2/lib"
	"github.com/cjnosal/manifer/v2/pkg/file"
	"github.com/cjnosal/manifer/v2/pkg/yaml"
)

type addCmd struct {
	name        string
	description string
	scenarios   []string

	manifer lib.Manifer

	logger *log.Logger
	writer io.Writer
}

var add addCmd

func NewAddCommand(l io.Writer, w io.Writer, m lib.Manifer) *cobra.Command {

	add.logger = log.New(l, "", 0)
	add.writer = w
	add.manifer = m

	cobraAdd := &cobra.Command{
		Use:   "add",
		Short: "add a new scenario to a library.",
		Long: `add --library <library path> --name <scenario name> [--description <text>] [--scenario <dependency>...] [-- passthrough flags ...]:
  add a new scenario to a library.
`,
		Run:              add.execute,
		TraverseChildren: true,
	}

	cobraAdd.Flags().StringSliceVarP(&libraryPaths, "library", "l", []string{}, "Path to library file")
	cobraAdd.Flags().StringVarP(&add.name, "name", "n", "", "Name to identify the new scenario")
	cobraAdd.Flags().StringVarP(&add.description, "description", "d", "", "Informative description of the new scenario")
	cobraAdd.Flags().StringSliceVarP(&add.scenarios, "scenario", "s", []string{}, "Dependency of the new scenario")

	return cobraAdd
}

func (p *addCmd) execute(cmd *cobra.Command, args []string) {
	if len(libraryPaths) != 1 {
		p.logger.Printf("Library path not specified")
		p.logger.Printf(cmd.Long)
		os.Exit(1)
	}

	if p.name == "" {
		p.logger.Printf("Name not specified")
		p.logger.Printf(cmd.Long)
		os.Exit(1)
	}

	lib, err := p.manifer.AddScenario(libraryPaths[0], p.name, p.description, p.scenarios, args)

	if err != nil {
		p.logger.Printf("%v\n  while adding scenario to library", err)
		os.Exit(1)
	}

	yaml := &yaml.Yaml{}
	outBytes, err := yaml.Marshal(lib)
	if err != nil {
		p.logger.Printf("%v\n  while marshaling updated library", err)
		os.Exit(1)
	}

	file := &file.FileIO{}
	err = file.Write(libraryPaths[0], outBytes, 0644)
	if err != nil {
		p.logger.Printf("%v\n  while overwriting updated library", err)
		os.Exit(1)
	}
}
