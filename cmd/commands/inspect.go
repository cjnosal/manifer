package commands

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/cjnosal/manifer/v2/lib"
	"github.com/cjnosal/manifer/v2/pkg/library"
	"github.com/cjnosal/manifer/v2/pkg/plan"
	"github.com/cjnosal/manifer/v2/pkg/yaml"
)

type inspectCmd struct {
	scenarios []string
	printJson bool
	printPlan bool
	printTree bool

	logger  *log.Logger
	writer  io.Writer
	manifer lib.Manifer
}

var inspect inspectCmd

func NewInspectCommand(l io.Writer, w io.Writer, m lib.Manifer) *cobra.Command {
	inspect.logger = log.New(l, "", 0)
	inspect.writer = w
	inspect.manifer = m

	cobraInspect := &cobra.Command{
		Use:   "inspect",
		Short: "inspect scenarios as a dependency tree or execution plan.",
		Long: `inspect (--library <library path>...) [--tree|--plan] (-s <scenario name>...) [-- passthrough flags ...]:
  inspect scenarios as a dependency tree or execution plan.
`,
		Run:              inspect.execute,
		TraverseChildren: true,
	}

	cobraInspect.Flags().StringSliceVarP(&libraryPaths, "library", "l", []string{}, "Path to library file")
	cobraInspect.Flags().BoolVarP(&inspect.printJson, "json", "j", false, "Print output in json format")
	cobraInspect.Flags().BoolVarP(&inspect.printPlan, "plan", "p", false, "Print execution plan")
	cobraInspect.Flags().BoolVarP(&inspect.printTree, "tree", "t", false, "Print dependency tree (default)")
	cobraInspect.Flags().StringSliceVarP(&inspect.scenarios, "scenario", "s", []string{}, "Scenario name in library")

	return cobraInspect
}

func (p *inspectCmd) execute(cmd *cobra.Command, args []string) {

	if len(libraryPaths) == 0 {
		p.logger.Printf("Library not specified")
		p.logger.Printf(cmd.Long)
		os.Exit(1)
	}

	if len(p.scenarios) == 0 {
		p.logger.Printf("A scenario must be specified")
		p.logger.Printf(cmd.Long)
		os.Exit(1)
	}

	nodes := library.ScenarioNodes{}
	for _, name := range p.scenarios {
		node, err := p.manifer.GetScenarioTree(libraryPaths, name)
		if err != nil {
			p.logger.Printf("%v\n  while inspecting scenario %s", err, name)
			os.Exit(1)
		}
		nodes = append(nodes, node)
	}
	for _, t := range library.Types {
		passthroughNode, remainder, err := p.manifer.GetSnippetScenarioNode(t, args)
		if err != nil {
			p.logger.Printf("%v\n  while trying to parse passthrough args", err)
			os.Exit(1)
		}
		args = remainder
		if passthroughNode != nil {
			nodes = append(nodes, passthroughNode)
		}
	}
	varNode, remainder, err := p.manifer.GetVarScenarioNode(args)
	if err != nil {
		p.logger.Printf("%v\n  while trying to parse variable args", err)
		os.Exit(1)
	}
	if varNode != nil {
		nodes = append(nodes, varNode)
	}
	if len(remainder) > 0 {
		p.logger.Printf("Invalid passthrough arguments %v", args)
		os.Exit(1)
	}

	var outBytes []byte
	if p.printPlan {
		executionPlan := &plan.Plan{
			Global: library.InterpolatorParams{
				Vars:      map[string]interface{}{},
				VarFiles:  map[string]string{},
				VarsFiles: []string{},
				VarsEnv:   []string{},
				VarsStore: "",
				RawArgs:   []string{},
			},
			Steps: []*plan.Step{},
		}
		for _, node := range nodes {
			executionPlan = plan.Append(executionPlan, plan.FromScenarioTree(node))
		}
		if p.printJson {
			outBytes = p.formatJson(executionPlan)
		} else {
			outBytes = p.formatYaml(executionPlan)
		}
	} else {
		if p.printJson {
			outBytes = p.formatJson(nodes)
		} else {
			outBytes = p.formatYaml(nodes)
		}
	}

	_, err = p.writer.Write(outBytes)
	if err != nil {
		p.logger.Printf("%v\n  while writing inspect output", err)
		os.Exit(1)
	}
}

func (p *inspectCmd) formatJson(i interface{}) []byte {
	bytes, _ := json.Marshal(i)
	return bytes
}

func (p *inspectCmd) formatYaml(i interface{}) []byte {
	yaml := yaml.Yaml{}
	bytes, _ := yaml.Marshal(i)
	return bytes
}
