package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cjnosal/manifer/lib"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/plan"
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

	nodes := []*library.ScenarioNode{}
	for _, name := range p.scenarios {
		node, err := p.manifer.GetScenarioTree(libraryPaths, name)
		if err != nil {
			p.logger.Printf("%v\n  while inspecting scenario %s", err, name)
			os.Exit(1)
		}
		nodes = append(nodes, node)
	}
	passthroughNode, err := p.manifer.GetSnippetScenarioNode(args)
	if err != nil {
		p.logger.Printf("%v\n  while trying to parse passthrough args", err)
		os.Exit(1)
	}
	if passthroughNode != nil {
		nodes = append(nodes, passthroughNode)
	}
	varNode, err := p.manifer.GetVarScenarioNode(args)
	if err != nil {
		p.logger.Printf("%v\n  while trying to parse variable args", err)
		os.Exit(1)
	}
	if varNode != nil {
		nodes = append(nodes, varNode)
	}

	var outBytes []byte
	if p.printPlan {
		executionPlan := &plan.Plan{}
		for _, node := range nodes {
			executionPlan = plan.Append(executionPlan, plan.FromScenarioTree(node))
		}
		if p.printJson {
			outBytes = p.formatJson(executionPlan)
		} else {
			outBytes = p.formatPlainPlan(executionPlan)
		}
	} else {
		if p.printJson {
			outBytes = p.formatJson(nodes)
		} else {
			builder := strings.Builder{}
			for _, node := range nodes {
				p.formatPlainTree(node, &builder, "")
			}
			outBytes = []byte(builder.String())
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

func (p *inspectCmd) formatPlainTree(node *library.ScenarioNode, builder *strings.Builder, indent string) {
	builder.WriteString(fmt.Sprintf("%sname:        %s (from %s)\n", indent, node.Name, node.LibraryPath))
	builder.WriteString(fmt.Sprintf("%sdescription: %s\n", indent, node.Description))
	builder.WriteString(fmt.Sprintf("%sglobal:  %+v (applied to all scenarios)\n", indent, node.GlobalInterpolator))
	builder.WriteString(fmt.Sprintf("%srefvars: %+v (applied to snippets and subscenarios)\n", indent, node.RefInterpolator))
	builder.WriteString(fmt.Sprintf("%svars:    %+v (applied to snippets and subscenarios)\n", indent, node.Interpolator))

	builder.WriteString(fmt.Sprintf("%ssnippets:\n", indent))
	for _, snippet := range node.Snippets {
		builder.WriteString(fmt.Sprintf("%s  %s\n", indent, snippet.Path))
		builder.WriteString(fmt.Sprintf("%s  vars: %+v\n", indent, snippet.Interpolator))
		builder.WriteString("\n")
	}
	builder.WriteString(fmt.Sprintf("%sdependencies:\n", indent))
	for _, dep := range node.Dependencies {
		p.formatPlainTree(dep, builder, indent+"  ")
		builder.WriteString("\n")
	}
}

func (p *inspectCmd) formatPlainPlan(executionPlan *plan.Plan) []byte {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("global: %+v\n", executionPlan.Global))
	for _, step := range executionPlan.Steps {
		builder.WriteString(fmt.Sprintf("- %s\n", step.Snippet))
		builder.WriteString(fmt.Sprintf("  vars:\n"))
		for _, argSet := range step.Params {
			builder.WriteString(fmt.Sprintf("    %s: %+v\n", argSet.Tag, argSet.Params))
		}
	}
	return []byte(builder.String())
}
