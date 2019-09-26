package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/google/subcommands"

	"github.com/cjnosal/manifer/lib"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/plan"
)

type inspectCmd struct {
	libraryPaths arrayFlags
	scenarios    arrayFlags
	printJson    bool
	printPlan    bool
	printTree    bool

	logger  *log.Logger
	writer  io.Writer
	manifer lib.Manifer
}

func NewInspectCommand(l io.Writer, w io.Writer, m lib.Manifer) subcommands.Command {
	return &inspectCmd{
		logger:  log.New(l, "", 0),
		writer:  w,
		manifer: m,
	}
}

func (*inspectCmd) Name() string { return "inspect" }
func (*inspectCmd) Synopsis() string {
	return "inspect scenarios as a dependency tree or execution plan."
}
func (*inspectCmd) Usage() string {
	return `inspect (--library <library path>...) [--tree|--plan] (-s <scenario name>...):
  inspect scenarios as a dependency tree or execution plan.
`
}

func (p *inspectCmd) SetFlags(f *flag.FlagSet) {
	f.Var(&p.libraryPaths, "library", "Path to library file")
	f.Var(&p.libraryPaths, "l", "Path to library file")
	f.BoolVar(&p.printJson, "json", false, "Print output in json format")
	f.BoolVar(&p.printJson, "j", false, "Print output in json format")
	f.BoolVar(&p.printPlan, "plan", false, "Print execution plan")
	f.BoolVar(&p.printPlan, "p", false, "Print execution plan")
	f.BoolVar(&p.printTree, "tree", false, "Print dependency tree (default)")
	f.BoolVar(&p.printTree, "t", false, "Print dependency tree (default)")
	f.Var(&p.scenarios, "scenario", "Scenario name in library")
	f.Var(&p.scenarios, "s", "Scenario name in library")
}

func (p *inspectCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	if len(p.libraryPaths) == 0 {
		p.logger.Printf("Library not specified")
		p.logger.Printf(p.Usage())
		return subcommands.ExitFailure
	}

	if len(p.scenarios) == 0 {
		p.logger.Printf("A scenario must be specified")
		p.logger.Printf(p.Usage())
		return subcommands.ExitFailure
	}

	nodes := []*library.ScenarioNode{}
	for _, name := range p.scenarios {
		node, err := p.manifer.GetScenarioTree(p.libraryPaths, name)
		if err != nil {
			p.logger.Printf("%v\n  while inspecting scenario %s", err, name)
			return subcommands.ExitFailure
		}
		nodes = append(nodes, node)
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

	_, err := p.writer.Write(outBytes)
	if err != nil {
		p.logger.Printf("%v\n  while writing inspect output", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func (p *inspectCmd) formatJson(i interface{}) []byte {
	bytes, _ := json.Marshal(i)
	return bytes
}

func (p *inspectCmd) formatPlainTree(node *library.ScenarioNode, builder *strings.Builder, indent string) {
	builder.WriteString(fmt.Sprintf("%sname:        %s (from %s)\n", indent, node.Name, node.LibraryPath))
	builder.WriteString(fmt.Sprintf("%sdescription: %s\n", indent, node.Description))
	builder.WriteString(fmt.Sprintf("%sglobal:  %v (applied to all scenarios)\n", indent, node.GlobalArgs))
	builder.WriteString(fmt.Sprintf("%srefargs: %v (applied to snippets and subscenarios)\n", indent, node.RefArgs))
	builder.WriteString(fmt.Sprintf("%sargs:    %v (applied to snippets and subscenarios)\n", indent, node.Args))

	builder.WriteString(fmt.Sprintf("%ssnippets:\n", indent))
	for _, snippet := range node.Snippets {
		builder.WriteString(fmt.Sprintf("%s  %s\n", indent, snippet.Path))
		builder.WriteString(fmt.Sprintf("%s  args: %v\n", indent, snippet.Args))
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
	builder.WriteString(fmt.Sprintf("global: %v\n", executionPlan.Global.Args))
	for _, step := range executionPlan.Steps {
		builder.WriteString(fmt.Sprintf("- %s\n", step.Snippet))
		builder.WriteString(fmt.Sprintf("  args:\n"))
		for _, argSet := range step.Args {
			builder.WriteString(fmt.Sprintf("    %s: %v\n", argSet.Tag, argSet.Args))
		}
	}
	return []byte(builder.String())
}
