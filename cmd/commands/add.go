package commands

import (
	"context"
	"flag"
	"io"
	"log"

	"github.com/google/subcommands"

	"github.com/cjnosal/manifer/lib"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/yaml"
)

type addCmd struct {
	name        string
	description string
	libraryPath string
	scenarios   arrayFlags

	manifer lib.Manifer

	logger *log.Logger
	writer io.Writer
}

func NewAddCommand(l io.Writer, w io.Writer, m lib.Manifer) subcommands.Command {
	return &addCmd{
		logger:  log.New(l, "", 0),
		writer:  w,
		manifer: m,
	}
}

func (*addCmd) Name() string     { return "add" }
func (*addCmd) Synopsis() string { return "add a new scenario to a library." }
func (*addCmd) Usage() string {
	return `add --library <library path> --name <scenario name> [--description <text>] [--scenario <dependency>...] [-- passthrough flags ...]:
  add a new scenario to a library.
`
}

func (p *addCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.libraryPath, "library", "", "Path to library file")
	f.StringVar(&p.libraryPath, "l", "", "Path to library file")
	f.StringVar(&p.name, "name", "", "Name to identify the new scenario")
	f.StringVar(&p.name, "n", "", "Name to identify the new scenario")
	f.StringVar(&p.description, "description", "", "Informative description of the new scenario")
	f.StringVar(&p.description, "d", "", "Informative description of the new scenario")
	f.Var(&p.scenarios, "scenario", "Dependency of the new scenario")
	f.Var(&p.scenarios, "s", "Dependency of the new scenario")
}

func (p *addCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	if p.libraryPath == "" {
		p.logger.Printf("Library path not specified")
		p.logger.Printf(p.Usage())
		return subcommands.ExitFailure
	}

	if p.name == "" {
		p.logger.Printf("Name not specified")
		p.logger.Printf(p.Usage())
		return subcommands.ExitFailure
	}

	lib, err := p.manifer.AddScenario(p.libraryPath, p.name, p.description, p.scenarios, f.Args())

	if err != nil {
		p.logger.Printf("%v\n  while adding scenario to library", err)
		return subcommands.ExitFailure
	}

	yaml := &yaml.Yaml{}
	outBytes, err := yaml.Marshal(lib)
	if err != nil {
		p.logger.Printf("%v\n  while marshaling updated library", err)
		return subcommands.ExitFailure
	}

	file := &file.FileIO{}
	err = file.Write(p.libraryPath, outBytes, 0644)
	if err != nil {
		p.logger.Printf("%v\n  while overwriting updated library", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
