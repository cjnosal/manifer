package commands

import (
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"

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

var imp importCmd

func NewImportCommand(l io.Writer, w io.Writer, m lib.Manifer) *cobra.Command {

	imp.logger = log.New(l, "", 0)
	imp.writer = w
	imp.manifer = m

	cobraImport := &cobra.Command{
		Use:   "import",
		Short: "create a library from a directory of snippets.",
		Long: `import [--recursive] --path <import path> --out <library path>:
  create a library from a directory of snippets.
`,
		Run:              imp.execute,
		TraverseChildren: true,
	}

	cobraImport.Flags().StringVarP(&imp.out, "out", "o", "", "Path to save generated library file")
	cobraImport.Flags().StringVarP(&imp.path, "path", "p", "", "Directory or opsfile to import")
	cobraImport.Flags().BoolVarP(&imp.recursive, "recursive", "r", false, "Import snippets from subdirectories")

	return cobraImport
}

func (p *importCmd) execute(cmd *cobra.Command, args []string) {

	if len(p.path) == 0 {
		p.logger.Printf("Import path not specified")
		p.logger.Printf(cmd.Long)
		os.Exit(1)
	}
	if len(p.out) == 0 {
		p.logger.Printf("Output path not specified")
		p.logger.Printf(cmd.Long)
		os.Exit(1)
	}

	lib := library.Library{}
	for _, t := range library.Types {
		tlib, err := p.manifer.Import(t, p.path, p.recursive, p.out)

		if err != nil {
			p.logger.Printf("%v\n  while importing %s snippets", err, t)
			os.Exit(1)
		}

		// imported scenarios should specify processor type at snippet level,
		// and import does not reference other libraries
		lib.Scenarios = append(lib.Scenarios, tlib.Scenarios...)
	}

	yaml := &yaml.Yaml{}
	outBytes, err := yaml.Marshal(lib)
	if err != nil {
		p.logger.Printf("%v\n  while marshaling generated library", err)
		os.Exit(1)
	}

	file := &file.FileIO{}
	err = file.Write(p.out, outBytes, 0644)
	if err != nil {
		p.logger.Printf("%v\n  while writing generated library", err)
		os.Exit(1)
	}
}
