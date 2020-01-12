# manifer
Manifer uses yaml processors and variable interpolation to generate yaml documents from a starting template and a set of reusable yaml snippets.

To reduce the size and complexity of yaml documents, named sets of ops files or yq scripts 
can be organized into a 'library'. The 'template' is now only 
responsible for simplified high level definitions. Running `manifer` will 
combine the library and template to compose the final document.

# glossary
- template:    an arbitrary yaml file to be modified
- snippet:     a file used to modify the template (e.g. BOSH ops files or yq scripts)
- scenario:    a named collection of snippets that should be used together
- library:     manifer's yaml file format that defines a collection of scenarios and dependencies between scenarios
- interpolate: replace variable placeholders with values from the CLI, files, or environment variables
- process:     apply structural changes to a template as defined by a snippet
- compose:     process and interpolate all snippets defined in a set of scenarios

# getting started
1a) generate a library from your collection of opsfiles and yq scripts  
  `manifer import -r -p ./ops-dir -o ./new-lib.yml`  
or  
1b) generate a library from your yaml file  
  `manifer generate -t ./template.yml -y opsfile -o ./new-lib.yml -d ./opsdir`  
2) view the generated scenarios  
  `manifer list -l new-lib.yml`  
3) add scenarios for your common use cases, which can define variables or invoke other scenarios  
  `manifer add -l new-lib.yml -n use_case -d "thing I need frequently" -s "dependency_name" -- -v foo=bar -o extra-op.yml`  
4) inspect the scenario you created  
  `manifer inspect -l new-lib.yml -s use_case`  
5) use your new scenario to modify a template  
  `manifer compose -l new-lib.yml -t base.yml -s use_case`

# global flags and environment variables
For convenience there are different ways to specify which libraries to read.  
In order of precedence:  
1) a local flag  
  `manifer list -l mylib.yml`  
2) a global flag  
  `manifer -l mylib.yml list`  
3) specific libraries via MANIFER_LIBS and the system path separator  
  `MANIFER_LIBS=mylib.yml:myotherlib.yml manifer list`  
4) directories to search via MANIFER_LIB_PATH and the system path separator  
  `MANIFER_LIB_PATH=./mylibs/:./sharedlibs manifer list`  

# subcommands
## import
```
./manifer import [--recursive] --path <import path> --out <library path>:
  create a library from a directory of snippets.

Usage:
  manifer import [flags]

Flags:
  -h, --help               help for import
  -o, --out string         Path to save generated library file
  -p, --path string        Directory or snippet to import
  -r, --recursive          Import snippets from subdirectories

Global Flags:
  -l, --library strings   Path to library file
```
## generate
```
./manifer generate --template <yaml path> --out <library path> [--directory <snippet path>]:
  create a library based on the structure of a yaml file.

Usage:
  manifer generate [flags]

Flags:
  -d, --directory string   Directory to save generated snippets (default out/snippets)
  -h, --help               help for generate
  -o, --out string         Path to save generated library file
  -y, --processor string   Yaml backend for this library (opsfile or yq)
  -t, --template string    Template to generate from

Global Flags:
  -l, --library strings   Path to library file
```
## add
```
./manifer add --library <library path> --name <scenario name> [--description <text>] [--scenario <dependency>...] [-- passthrough flags ...]:
  add a new scenario to a library.

Usage:
  manifer add [flags]

Flags:
  -d, --description string   Informative description of the new scenario
  -h, --help                 help for add
  -n, --name string          Name to identify the new scenario
  -s, --scenario strings     Dependency of the new scenario

Global Flags:
  -l, --library strings   Path to library file
```
## list
```
./manifer list [--all] (--library <library path>...):
  list scenarios in selected libraries.

Usage:
  manifer list [flags]

Flags:
  -a, --allScenarios   Include all referenced libraries
  -h, --help           help for list
  -j, --json           Print output in json format

Global Flags:
  -l, --library strings   Path to library file
```
## search
```
./manifer search (--library <library path>...) (query...):
  search scenarios in selected libraries by name and description.

Usage:
  manifer search [flags]

Flags:
  -h, --help   help for search
  -j, --json   Print output in json format

Global Flags:
  -l, --library strings   Path to library file
```
## inspect
```
./manifer inspect (--library <library path>...) [--tree|--plan] (-s <scenario name>...) [-- passthrough flags ...]:
  inspect scenarios as a dependency tree or execution plan.

Usage:
  manifer inspect [flags]

Flags:
  -h, --help               help for inspect
  -j, --json               Print output in json format
  -p, --plan               Print execution plan
  -s, --scenario strings   Scenario name in library
  -t, --tree               Print dependency tree (default)

Global Flags:
  -l, --library strings   Path to library file
```
## compose
```
./manifer compose --template <template path> (--library <library path>...) (--scenario <scenario>...) [--print] [--diff] [-- passthrough flags ...] [\;] :
  compose a yml file from snippets. Use '\;' as a separator when reusing a scenario with different variables.

Usage:
  manifer compose [flags]

Flags:
  -d, --diff               Show diff after each snippet is applied
  -h, --help               help for compose
  -p, --print              Show snippets and arguments being applied
  -s, --scenario strings   Scenario name in library
  -t, --template string    Path to initial template file

Global Flags:
  -l, --library strings   Path to library file
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
Any valid yaml document you would like to modify with snippets and [implicit bosh variables](https://bosh.io/docs/cli-int/#implicit)

e.g. foo-template.yml
```
foo:
  bar: bizz
  buzz: ((bazz))
  extra: redundant
```

## snippets
Yaml snippets to compose into the template:  
- opsfile snippets use [go-patch](https://github.com/cppforlife/go-patch) format, 
Also known as [BOSH Ops Files](https://bosh.io/docs/cli-ops-files).  
- yq snippets use [yq](https://mikefarah.github.io/yq/) script format

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

Library files are used to organize sets of snippets, interpolation variables, and processor options
to allow easy dynamic composition of large yaml files.  

Libraries consist of:  
- a default processor type for all snippets  
  `type: opsfile`  
- aliases to other libraries  
  ```
  libraries:
  - alias: common
    path: ./commonlib.yml
  ```
- a list of scenarios, consisting of:  
  - a unique name  
  - a user-friendly description  
  - references to other scenarios this scenario depends on:  
    - by name, prefixed with `.` delimited library aliases  
    - interpolator variables to apply to the referenced scenario  
  - snippets to transform the template yaml:  
    - path to the snippet file  
    - interpolator variables for this snippet  
    - processor options for this snippet  
  - interpolator variables to use with all snippets in this scenario and referenced scenarios  
  - global interpolator variables to use with all snippets in this composition  
  ```
  scenarios:
  - name: first
    description: my first scenario
    scenarios:
    - name: second
      interpolator:
        vars:
          foo: bar
    - name: common.setup
    snippets:
    - path: ./opsfile.yml
      interpolator:
        vars:
          bizz: bazz
      processor:
        type: opsfile
        options:
          path: /buzz
  - name: second
    snippets:
    - path: ./secondop.yml
    - path: ./thirdop.yml
    global_interpolator:
      vars_store: ./generated.yml
  ```

### migrating from v1
In v1 libraries interpolator variables were specified as CLI `args`. In v2 `args` is replaced by the `interpolator` struct.  

When upgrading from manifer v1=>v2 you can either:  
- move each `args` element to `interpolator.raw_args` or  
- replace `args` with the appropriate `interpolator.var*` field
  
### interpolator variables
Variables can be defined by adding an `interpolator` block to a snippet, scenario reference, scenario, or via passthrough flags from the CLI
```
interpolator:
  vars: {} # map variable names to static values [--var=key=val (-v)]
  var_files: {} # map variable names to files that contain the value [--var-file=key=path]
  vars_files: [] # file paths that contain a map of variable names to static values [--vars-file=path (-l)]
  vars_env: [] # environment variables with the given prefixes [--vars-env=prefix]
  vars_store: "" # a vars-file that can lazily generate random passwords or certificates [--vars-store=path]
  raw_args: [] # insert CLI flags into the scenario definition (for internal use)
```

See [bosh interpolate](https://bosh.io/docs/cli-int/) and [variable types](https://bosh.io/docs/variable-types/) for more details

### processor options
A snippet can override the library's default processor type, or provide options

#### opsfile processor
```
type: opsfile
options:  
  path: "/foo" # return a section of the composed yaml instead of the full document
```
See the [go-patch](https://github.com/cppforlife/go-patch) and [ops-file](https://bosh.io/docs/cli-ops-files) docs for more details

Differentiating features: field matching and index selection

#### yq processor
```
type: yq
options:  
  command: # write, read, delete, prefix, merge (default write)
  path: # yaml element to read or delete
  prefix: # key to nest the current yaml structure under
  overwrite: # boolean for merges to replace existing elements
  append: # boolean for merges to append new array elements
```
See the [yq docs](https://mikefarah.github.io/yq/) for more details

Differentiating features: wildcards, prefix, and merge

## Invocation
Running `manifer compose --library mainlib.yml --template foo-template.yml --scenario my-use-case` should produce:
```
foo:
  bar: trendy
  buzz: 123
  nested:
    new: struct
```

# interpolation and variables
There are four variable scopes:
- snippet vars
- scenario vars
- scenario global vars
- CLI global vars

Every snippet is used in two interpolations:
- the snippet itself is interpolated with snippet vars, scenario vars, and global vars
- then the interpolated snippet is applied to the template, interpolated with global vars

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

# use manifer in your project
`lib.Manifer` can be imported to list scenarios or compose yaml

```
package main

import (
  "os"
  "fmt"
  "github.com/cjnosal/manifer/v2/lib"
)

func main() {
  logger := os.Stderr
  output := os.Stdout
  manifer := lib.NewManifer(logger)

  // starting yaml file
  template := "test/data/v2/template.yml"

  // collection of scenarios
  libraries := []string{"test/data/v2/library.yml"}

  // sets of snippets to apply
  scenarios := []string{"placeholder"}

  // arguments to pass through to `bosh interpolate`
  interpolationVars := []string{"-vpath3=/foo", "-vvalue3=tweaks"}

  // list scenario names with descriptions
  scenarioSummary, err := manifer.ListScenarios(libraries, false)
  logger.Write([]byte(fmt.Sprintf("%v\n", scenarioSummary)))
  logger.Write([]byte(fmt.Sprintf("%v\n", err)))

  // apply snippets from the selected scenarios to the provided template
  composedYaml, err := manifer.Compose(template, libraries, scenarios, interpolationVars, false, false)
  output.Write(composedYaml)
  logger.Write([]byte(fmt.Sprintf("%v\n", err)))
}
```
