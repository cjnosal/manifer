package commands

import (
	"context"
	"flag"
	"io"
	"log"
	"path/filepath"

	"github.com/google/subcommands"

	"github.com/cjnosal/manifer/lib"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/yaml"
)

type generateCmd struct {
	lib      string
	template string
	dir      string

	logger  *log.Logger
	writer  io.Writer
	manifer lib.Manifer
}

func NewGenerateCommand(l io.Writer, w io.Writer, m lib.Manifer) subcommands.Command {
	return &generateCmd{
		logger:  log.New(l, "", 0),
		writer:  w,
		manifer: m,
	}
}

func (*generateCmd) Name() string { return "generate" }
func (*generateCmd) Synopsis() string {
	return "create a library based on the structure of a yaml file."
}
func (*generateCmd) Usage() string {
	return `generate --template <yaml path> --out <library path> [--directory <snippet path>]:
  create a library based on the structure of a yaml file.
`
}

func (p *generateCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.lib, "out", "", "Path to save generated library file")
	f.StringVar(&p.lib, "o", "", "Path to save generated library file")
	f.StringVar(&p.template, "template", "", "Template to generate from")
	f.StringVar(&p.template, "t", "", "Template to generate from")
	f.StringVar(&p.dir, "directory", "", "Directory to save generated snippets (default out/ops)")
	f.StringVar(&p.dir, "d", "", "Directory to save generated snippets (default out/ops)")
}

func (p *generateCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	if len(p.template) == 0 {
		p.logger.Printf("Template path not specified")
		p.logger.Printf(p.Usage())
		return subcommands.ExitFailure
	}
	if len(p.lib) == 0 {
		p.logger.Printf("Output path not specified")
		p.logger.Printf(p.Usage())
		return subcommands.ExitFailure
	}
	if len(p.dir) == 0 {
		p.dir = filepath.Join(filepath.Dir(p.lib), "ops")
	}

	lib, err := p.manifer.Generate(library.OpsFile, p.template, p.lib, p.dir)

	if err != nil {
		p.logger.Printf("%v\n  while generating library", err)
		return subcommands.ExitFailure
	}

	yaml := &yaml.Yaml{}
	outBytes, err := yaml.Marshal(lib)
	if err != nil {
		p.logger.Printf("%v\n  while marshaling generated library", err)
		return subcommands.ExitFailure
	}

	file := &file.FileIO{}
	err = file.Write(p.lib, outBytes, 0644)
	if err != nil {
		p.logger.Printf("%v\n  while writing generated library", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
