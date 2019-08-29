package composer

import (
	"fmt"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/plan"
	"path/filepath"
)

type Composer interface {
	Compose(executor plan.Executor,
		templatePath string,
		libraryPaths []string,
		scenarioNames []string,
		passthrough []string,
		showPlan bool,
		showDiff bool) ([]byte, error)
}

type ComposerImpl struct {
	Resolver ScenarioResolver
	File     file.FileAccess
}

func (c *ComposerImpl) Compose(executor plan.Executor,
	templatePath string,
	libraryPaths []string,
	scenarioNames []string,
	passthrough []string,
	showPlan bool,
	showDiff bool) ([]byte, error) {

	plan, err := c.Resolver.Resolve(libraryPaths, scenarioNames, passthrough)
	if err != nil {
		return nil, fmt.Errorf("Unable to resolve scenarios: %s", err.Error())
	}

	temp, err := c.File.TempDir("", "manifer")
	if err != nil {
		return nil, fmt.Errorf("Unable to create temporary directory: %s", err.Error())
	}
	defer c.File.RemoveAll(temp)

	in := templatePath
	var out string

	if len(plan.Snippets) > 0 || len(plan.GlobalArgs) > 0 {

		for i, snippet := range plan.Snippets {
			out = fmt.Sprintf(filepath.Join(temp, "composed_%d.yml"), i)
			err = executor.Execute(showPlan, showDiff, in, out, snippet.Path, snippet.Args, plan.GlobalArgs)
			if err != nil {
				return nil, fmt.Errorf("Unable to apply snippet %s: %s", snippet.Path, err.Error())
			}

			in = out
		}

		if len(plan.GlobalArgs) > 0 {
			out = fmt.Sprintf(filepath.Join(temp, "composed_final.yml"))
			err = executor.Execute(showPlan, showDiff, in, out, "", nil, plan.GlobalArgs)
			if err != nil {
				return nil, fmt.Errorf("Unable to apply passthrough args %v: %s", plan.GlobalArgs, err.Error())
			}
		}
	} else {
		out = templatePath
	}

	outBytes, err := c.File.Read(out)
	if err != nil {
		return nil, fmt.Errorf("Unable to read composed output: %v", err)
	}

	return outBytes, nil
}
