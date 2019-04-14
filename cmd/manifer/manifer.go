package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/cjnosal/manifer/pkg/command"
	"github.com/cjnosal/manifer/pkg/composer"
	"github.com/cjnosal/manifer/pkg/diff"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/interpolator/opsfile"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/plan"
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
	patch := diffmatchpatch.New()
	diff := &diff.FileDiff{
		File:  file,
		Patch: patch,
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
	opsFileExecutor := &plan.InterpolationExecutor{
		Interpolator: opsFileInterpolator,
		Diff:         diff,
		Output:       logger,
	}
	composer := &composer.ComposerImpl{
		Resolver: resolver,
		File:     file,
	}
	lister := &scenario.Lister{
		Loader: loader,
	}

	executorMap := map[library.Type]plan.Executor{
		library.OpsFile: opsFileExecutor,
	}

	// register subcommands
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(command.NewComposeCommand(logger, writer, composer, executorMap), "")
	subcommands.Register(command.NewListCommand(logger, writer, lister), "")

	// run
	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
