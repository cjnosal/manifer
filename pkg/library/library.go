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
	Name               string
	Description        string
	GlobalInterpolator InterpolatorParams `yaml:"global_interpolator,omitempty"`
	Interpolator       InterpolatorParams `yaml:"interpolator,omitempty"`
	Snippets           []Snippet
	Scenarios          []ScenarioRef
}

type ScenarioRef struct {
	Name         string
	Interpolator InterpolatorParams `yaml:"interpolator,omitempty"`
}

type Snippet struct {
	Path         string
	Interpolator InterpolatorParams `yaml:"interpolator,omitempty"`
	Processor    Processor          `yaml:"processor,omitempty"`
}

type Processor struct {
	Type    Type                   `yaml:"type,omitempty"`
	Options map[string]interface{} `yaml:"options,omitempty"`
}

func (p Processor) IsZero() bool {
	return p.Type == "" && len(p.Options) == 0
}

type InterpolatorParams struct {
	Vars      map[string]interface{} `yaml:"vars,omitempty"`
	VarFiles  map[string]string      `yaml:"var_files,omitempty"`
	VarsFiles []string               `yaml:"vars_files,omitempty"`
	VarsEnv   []string               `yaml:"vars_env,omitempty"`
	VarsStore string                 `yaml:"vars_store,omitempty"`
	RawArgs   []string               `yaml:"raw_args,omitempty"`
}

func (i InterpolatorParams) IsZero() bool {
	return len(i.Vars) == 0 && len(i.VarFiles) == 0 && len(i.VarsFiles) == 0 &&
		len(i.VarsEnv) == 0 && len(i.VarsStore) == 0 && len(i.RawArgs) == 0
}

func (ip InterpolatorParams) Merge(other InterpolatorParams) InterpolatorParams {
	for k, v := range other.Vars {
		ip.Vars[k] = v
	}
	for k, v := range other.VarFiles {
		ip.VarFiles[k] = v
	}
	ip.VarsFiles = append(ip.VarsFiles, other.VarsFiles...)
	ip.VarsEnv = append(ip.VarsEnv, other.VarsEnv...)
	if other.VarsStore != "" {
		ip.VarsStore = other.VarsStore
	}
	ip.RawArgs = append(ip.RawArgs, other.RawArgs...)
	return ip
}
