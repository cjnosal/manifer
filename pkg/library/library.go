package library

type Type string

const (
	OpsFile Type = "opsfile"
)

type Library struct {
	Libraries []LibraryRef
	Type      Type
	Scenarios []Scenario
}

type LibraryRef struct {
	Alias string
	Path  string
}

type Scenario struct {
	Name         string
	Description  string
	GlobalArgs   []string `yaml:"global_args"`
	TemplateArgs []string `yaml:"template_args"`
	Args         []string
	Snippets     []Snippet
	Scenarios    []ScenarioRef
}

type ScenarioRef struct {
	Name string
	Args []string
}

type Snippet struct {
	Path string
	Args []string
}
