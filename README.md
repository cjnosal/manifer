# manifer
Tools for generating yaml documents from templates using `bosh interpolate`

To reduce the size and complexity of yaml documents, named sets of ops files
can be organized into a 'library'. The 'template' is now only 
responsible for simplified high level definitions. Running `manifer` will 
combine the library and template to compose the final document.

# build
`./scripts/build.sh [--setup]`
- `--setup` will `go get` required branches of forked dependencies

# test
`./scripts/test.sh [go test flags]`
- `-count=1` can be used to disable test caching of integration tests in `cmd/manifer`

# run
```
./manifer list [-a] (-l <library path>...):
  list scenarios in selected libraries.
  -a    Include all referenced libraries
  -l value
        Path to library file
```
```
./manifer compose -t <template path> (-l <library path>...) (-s <scenario>...) [-p] [-d] [-- passthrough flags ...]:
  compose a yml file from snippets.
  -d    Show diff after each snippet is applied
  -l value
        Path to library file
  -p    Show snippets and arguments being applied
  -s value
        Scenario name in library
  -t string
        Path to template file
```

# interpolation and arguments

There are four argument scopes:
- snippet args
- scenario args
- scenario global args
- CLI global args

Every snippet is used in two interpolations:
- first the snippet itself is interpolated with snippet args, scenario args, and global args
- then the template is interpolated with the interpolated snippet and global args

# multiple libraries
Libraries can include other libraries by specifying the path and an alias under 
`libraries:`. If a scenario needs to include a scenario from a referenced 
library the name should be prefixed with `<library alias>.`.

If multiple independant libraries are provided to the CLI all scenario names 
should be unambiguous.

# sample
```base_library.yml
type: opsfile

scenarios:
- name: "reusable"
  snippets:
  - path: ./common_opsfile.yml
```
```derived_library.yml
type: opsfile

libraries:
- alias: other
  path: "./base_library.yml"

scenarios:
- name: "base"
  args:
  - -v 
  - foo=bar
  snippets:
  - path: ./opsfile.yml
    args:
    - -v 
    - path=value

- name: "derived"
  scenarios:
  - "other.reusable"
  - "base"
  global_args:
  - -v
  - deployment=test
  args:
  - -v 
  - foo=overridden
  snippets:
  - path: ./another_opsfile.yml
```
```template.yml
arbitrary: yaml
```

Invoke with:
`./manifer compose -l derived_library.yml -t template.yml -s derived -- -v extra=arg`

Add `-p` (show plan) and `-d` (show diff) to see the input/output of each interpolation