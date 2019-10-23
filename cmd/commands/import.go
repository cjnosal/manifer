package commands

import (
	"context"
	"flag"
	"io"
	"log"

	"github.com/google/subcommands"

	"github.com/cjnosal/manifer/lib"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/yaml"
)

type importCmd struct {
	out       string
	path      string
	recursive bool

	logger  *log.Logger
	writer  io.Writer
	manifer lib.Manifer
}

func NewImportCommand(l io.Writer, w io.Writer, m lib.Manifer) subcommands.Command {
	return &importCmd{
		logger:  log.New(l, "", 0),
		writer:  w,
		manifer: m,
	}
}

func (*importCmd) Name() string     { return "import" }
func (*importCmd) Synopsis() string { return "create a library from a directory of opsfiles." }
func (*importCmd) Usage() string {
	return `import [--recursive] --path <import path> --out <library path>:
  create a library from a directory of opsfiles.
`
}

func (p *importCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.out, "out", "", "Path to save generated library file")
	f.StringVar(&p.out, "o", "", "Path to save generated library file")
	f.StringVar(&p.path, "path", "", "Directory or opsfile to import")
	f.StringVar(&p.path, "p", "", "Directory or opsfile to import")
	f.BoolVar(&p.recursive, "recursive", false, "Import opsfiles from subdirectories")
	f.BoolVar(&p.recursive, "r", false, "Import opsfiles from subdirectories")
}

func (p *importCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	if len(p.path) == 0 {
		p.logger.Printf("Import path not specified")
		p.logger.Printf(p.Usage())
		return subcommands.ExitFailure
	}
	if len(p.out) == 0 {
		p.logger.Printf("Output path not specified")
		p.logger.Printf(p.Usage())
		return subcommands.ExitFailure
	}

	lib, err := p.manifer.Import(library.OpsFile, p.path, p.recursive, p.out)

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
	err = file.Write(p.out, outBytes, 0644)
	if err != nil {
		p.logger.Printf("%v\n  while writing generated library", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
