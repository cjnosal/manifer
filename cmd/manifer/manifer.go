package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"

	"github.com/cjnosal/manifer/pkg/command"
	"github.com/cjnosal/manifer/pkg/composer"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/interpolator"
	"github.com/cjnosal/manifer/pkg/interpolator/opsfile"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/scenario"
	"github.com/cjnosal/manifer/pkg/yaml"
)

func main() {
	// setup dependencies
	logger := os.Stderr
	writer := os.Stdout
	file := &file.FileIO{}
	yaml := &yaml.Yaml{
		File: file,
	}
	lookup := &library.Lookup{}
	selector := &scenario.Selector{
		Lookup: lookup,
	}
	loader := &library.Loader{
		File: file,
		Yaml: yaml,
	}
	resolver := &composer.Resolver{
		Loader:   loader,
		Selector: selector,
	}
	opsFileInterpolator := opsfile.NewOpsFileInterpolator(file, yaml)
	composer := &composer.ComposerImpl{
		Resolver: resolver,
		File:     file,
	}
	lister := &scenario.Lister{
		Loader: loader,
	}

	interpolatorMap := map[library.Type]interpolator.Interpolator{
		library.OpsFile: opsFileInterpolator,
	}

	// TODO subcommand/flag for showing plan steps
	// TODO subcommand/flag for showing diffs between steps

	// TODO integration tests

	// register subcommands
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(command.NewComposeCommand(logger, writer, composer, interpolatorMap), "")
	subcommands.Register(command.NewListCommand(logger, writer, lister), "")

	// run
	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
