# manifer
Tools for generating yaml documents from templates using `bosh interpolate`

To reduce the size and complexity of yaml documents, named sets of ops files
can be organized into a 'library'. The 'template' is now only 
responsible for simplified high level definitions. Running `manifer` will 
combine the library and template to compose the final document.

# glossary
- template:    an arbitrary yaml file to be modified
- snippet:     a file used to modify the template (e.g. BOSH ops files)
- scenario:    a named collection of snippets that should be used together
- library:     manifer's yaml file format that defines a collection of scenarios
- interpolate: apply a snippet to the template
- compose:     interpolate for all snippets defined in a set of scenarios

# getting started
1a) generate a library from your collection of opsfiles  
  `manifer import -r -p ./ops-dir -o ./new-lib.yml`  
or  
1b) generate a library from your yaml file  
  `manifer generate -t ./template.yml -o ./new-lib.yml -d ./opsdir`  
2) view the generated scenarios  
  `manifer list -l new-lib.yml`  
3) add scenarios for your common use cases, which can define variables or invoke other scenarios  
  `manifer add -l new-lib.yml -n use_case -d "thing I need frequently" -s "dependency_name" -- -v foo=bar -o extra-op.yml`  
4) inspect the scenario you created  
  `manifer inspect -l new-lib.yml -s use_case`  
5) use your new scenario to modify a template  
  `manifer compose -l new-lib.yml -t base.yml -s use_case`

# subcommands
## import
```
./manifer import [--recursive] --path <import path> --out <library path>:
  create a library from a directory of opsfiles.
  -o string
        Path to save generated library file
  -out string
        Path to save generated library file
  -p string
        Directory or opsfile to import
  -path string
        Directory or opsfile to import
  -r    Import opsfiles from subdirectories
  -recursive
        Import opsfiles from subdirectories
```
## generate
```
./manifer generate --template <yaml path> --out <library path> [--directory <snippet path>]:
  create a library based on the structure of a yaml file.
  -d string
        Directory to save generated snippets (default out/ops)
  -directory string
        Directory to save generated snippets (default out/ops)
  -o string
        Path to save generated library file
  -out string
        Path to save generated library file
  -t string
        Template to generate from
  -template string
        Template to generate from
```
## add
```
./manifer add --library <library path> --name <scenario name> [--description <text>] [--scenario <dependency>...] [-- passthrough flags ...]:
  add a new scenario to a library.
  -d string
        Informative description of the new scenario
  -description string
        Informative description of the new scenario
  -l string
        Path to library file
  -library string
        Path to library file
  -n string
        Name to identify the new scenario
  -name string
        Name to identify the new scenario
  -s value
        Dependency of the new scenario
  -scenario value
        Dependency of the new scenario
```
## list
```
./manifer list [--all] (--library <library path>...):
  list scenarios in selected libraries.
  -a    Include all referenced libraries
  -all
        Include all referenced libraries
  -j    Print output in json format
  -json
        Print output in json format
  -l value
        Path to library file
  -library value
        Path to library file
```
## search
```
./manifer search (--library <library path>...) (query...):
  search scenarios in selected libraries by name and description.
  -j    Print output in json format
  -json
        Print output in json format
  -l value
        Path to library file
  -library value
        Path to library file
```
## inspect
```
./manifer inspect (--library <library path>...) [--tree|--plan] (-s <scenario name>...) [-- passthrough flags ...]:
  inspect scenarios as a dependency tree or execution plan.
  -j    Print output in json format
  -json
        Print output in json format
  -l value
        Path to library file
  -library value
        Path to library file
  -p    Print execution plan
  -plan
        Print execution plan
  -s value
        Scenario name in library
  -scenario value
        Scenario name in library
  -t    Print dependency tree (default)
  -tree
        Print dependency tree (default)
```
## compose
```
./manifer compose --template <template path> (--library <library path>...) (--scenario <scenario>...) [--print] [--diff] [-- passthrough flags ...] [\;] :
  compose a yml file from snippets. Use '\;' as a separator when reusing a scenario with different variables.
  -d    Show diff after each snippet is applied
  -diff
        Show diff after each snippet is applied
  -l value
        Path to library file
  -library value
        Path to library file
  -p    Show snippets and arguments being applied
  -print
        Show snippets and arguments being applied
  -s value
        Scenario name in library
  -scenario value
        Scenario name in library
  -t string
        Path to initial template file
  -template string
        Path to initial template file
```
### appending additional compositions
Additional compositions can be appended using `\;` as a separator. For each additional composition:
- the output of the last composition is used as the template
- the list of libraries will be preserved
- new libraries can be referenced
- new scenarios and passthrough arguments can be specified
- global variables cleared

This allows the value of a variable to be changed, without having to re-enter file paths for the libraries or template.
The following invocations are equivalent:
```
./manifer compose -t my-template -l my-library -s my-scenario -- -v arg=foo > temp
./manifer compose -t temp -l my-library -s my-scenario -- -v arg=bar > final
```
```
./manifer compose -t my-template -l my-library -s my-scenario -- -v arg=foo \; \
  -s my-scenario -- -v arg=bar > final
```

# schemas

## template
Any valid yaml document you would like to modify with opsfiles and [implicit bosh variables](https://bosh.io/docs/cli-int/#implicit)

e.g. foo-template.yml
```
foo:
  bar: bizz
  buzz: ((bazz))
  extra: redundant
```

## snippets
Yaml snippets to compose into the template use [go-patch](https://github.com/cppforlife/go-patch) format, 
Also known as [BOSH Ops Files](https://bosh.io/docs/cli-ops-files).

e.g. base-case.yml
```
- path: /foo/bar
  type: replace
  value: ((newbar))

- path: /foo/extra
  type: remove

- path: /foo/((sub))? # note: bosh/go-patch do not natively support variables in opsfile paths
  type: replace
  value:
    new: struct
```

## library
[library.go](https://github.com/cjnosal/manifer/blob/master/pkg/library/library.go)

Opsfiles can be grouped into scenarios, a named set that will be applied in order with associated variables to the template.
Scenarios are defined in a library file.

e.g. commonlib.yml
```
type: opsfile
scenarios:
- name: base-case
  description: helpful text displayed by `./manifer list`
  args: # applied to all snippets in this scenario
  - -v
  - sub=nested
  snippets: # opsfiles to apply, in order
  - path: ./base-case.yml
    args: # applied to single snippet
    - -v
    - newbar=trendy
```

e.g. mainlib.yml
```
type: opsfile # other yaml templating tools could be supported in the future
libraries:
- alias: common # reference to another library file
  path: ./commonlib.yml
scenarios:
- name: my-use-case
  scenarios: # scenarios can reference other scenarios. The referenced scenario's snippets are applied first.
  - name: common.base-case # prefix library alias if scenario name is in referenced library
    args: # arguments will be applied to all snippets in the referenced scenario
    - -v
    - e=f
  global_args: # applied all snippets in all scenarios as well as the template
  - -v
  - bazz=123
```

## Invocation
Running `manifer compose --library mainlib.yml --template foo-template.yml --scenario my-use-case` should produce:
```
foo:
  bar: trendy
  buzz: 123
  nested:
    new: struct
```

# interpolation and arguments
There are four argument scopes:
- snippet args
- scenario args
- scenario global args
- CLI global args

Every snippet is used in two interpolations:
- the snippet itself is interpolated with snippet args, scenario args, and global args
- then the interpolated snippet is applied to the template, interpolated with global args

# multiple libraries
Libraries can include other libraries by specifying the path and an alias under 
`libraries:`. If a scenario needs to include a scenario from a referenced 
library the name should be prefixed with `<library alias>.`.

If multiple independant libraries are provided to the CLI all scenario names 
should be unambiguous.

# build
`./scripts/build.sh [all]`
- `all` will build `manifer_darwin` and `manifer_linux`

# test
`./scripts/test.sh [unit|integration|go test flags]`
- `-count=1` can be used to disable test caching of integration tests in `cmd/manifer`

# import
`lib.Manifer` can be imported to list scenarios or compose yaml

```
package main

import (
  "os"
  "fmt"
  "github.com/cjnosal/manifer/lib"
)

func main() {
  logger := os.Stderr
  output := os.Stdout
  manifer := lib.NewManifer(logger)

  // starting yaml file
  template := "test/data/template.yml"

  // collection of scenarios
  libraries := []string{"test/data/library.yml"}

  // sets of ops files to apply
  scenarios := []string{"placeholder"}

  // arguments to pass through to `bosh interpolate`
  interpolationArgs := []string{"-vpath3=/foo", "-vvalue3=tweaks"}

  // list scenario names with descriptions
  scenarioSummary, err := manifer.ListScenarios(libraries, false)
  logger.Write([]byte(fmt.Sprintf("%v\n", scenarioSummary)))
  logger.Write([]byte(fmt.Sprintf("%v\n", err)))

  // apply ops files from the selected scenarios to the provided template
  composedYaml, err := manifer.Compose(template, libraries, scenarios, interpolationArgs, false, false)
  output.Write(composedYaml)
  logger.Write([]byte(fmt.Sprintf("%v\n", err)))
}
```
