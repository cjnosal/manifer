package commands

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cjnosal/manifer/v2/lib"
	"github.com/cjnosal/manifer/v2/pkg/file"
	"github.com/cjnosal/manifer/v2/pkg/library"
	"github.com/cjnosal/manifer/v2/pkg/yaml"
)

type generateCmd struct {
	lib      string
	template string
	dir      string
	libType  string

	logger  *log.Logger
	writer  io.Writer
	manifer lib.Manifer
}

var generate generateCmd

func NewGenerateCommand(l io.Writer, w io.Writer, m lib.Manifer) *cobra.Command {

	generate.logger = log.New(l, "", 0)
	generate.writer = w
	generate.manifer = m

	cobraGenerate := &cobra.Command{
		Use:   "generate",
		Short: "create a library based on the structure of a yaml file.",
		Long: `generate --template <yaml path> --out <library path> [--directory <snippet path>]:
  create a library based on the structure of a yaml file.
`,
		Run:              generate.execute,
		TraverseChildren: true,
	}

	cobraGenerate.Flags().StringVarP(&generate.lib, "out", "o", "", "Path to save generated library file")
	cobraGenerate.Flags().StringVarP(&generate.template, "template", "t", "", "Template to generate from")
	cobraGenerate.Flags().StringVarP(&generate.dir, "directory", "d", "", "Directory to save generated snippets (default out/snippets)")
	cobraGenerate.Flags().StringVarP(&generate.libType, "processor", "y", "", "Yaml backend for this library (opsfile or yq)")

	return cobraGenerate
}

func (p *generateCmd) execute(cmd *cobra.Command, args []string) {

	if len(p.template) == 0 {
		p.logger.Printf("Template path not specified")
		p.logger.Printf(cmd.Long)
		os.Exit(1)
	}
	if len(p.lib) == 0 {
		p.logger.Printf("Output path not specified")
		p.logger.Printf(cmd.Long)
		os.Exit(1)
	}
	if len(p.libType) == 0 {
		p.logger.Printf("Yaml processor not specified")
		p.logger.Printf(cmd.Long)
		os.Exit(1)
	}
	if len(p.dir) == 0 {
		p.dir = filepath.Join(filepath.Dir(p.lib), "snippets")
	}

	lib, err := p.manifer.Generate(library.Type(p.libType), p.template, p.lib, p.dir)

	if err != nil {
		p.logger.Printf("%v\n  while generating library", err)
		os.Exit(1)
	}

	yaml := &yaml.Yaml{}
	outBytes, err := yaml.Marshal(lib)
	if err != nil {
		p.logger.Printf("%v\n  while marshaling generated library", err)
		os.Exit(1)
	}

	file := &file.FileIO{}
	err = file.Write(p.lib, outBytes, 0644)
	if err != nil {
		p.logger.Printf("%v\n  while writing generated library", err)
		os.Exit(1)
	}
}
