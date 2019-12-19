package library

type Type string

const (
	OpsFile Type = "opsfile"
)

type Library struct {
	Libraries []LibraryRef
	Type      Type `yaml:"type,omitempty"`
	Scenarios []Scenario
}

type LibraryRef struct {
	Alias string
	Path  string
}

type Scenario struct {
	Name        string
	Description string
	GlobalArgs  []string `yaml:"global_args,omitempty"`
	Args        []string `yaml:"args,omitempty"`
	Snippets    []Snippet
	Scenarios   []ScenarioRef
}

type ScenarioRef struct {
	Name string
	Args []string `yaml:"args,omitempty"`
}

type Snippet struct {
	Path string
	Args []string `yaml:"args,omitempty"`
}
