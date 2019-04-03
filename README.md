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
`./scripts/test.sh`

# run
```
./manifer compose -t <template path> (-l <library path>...) (-s <scenario>...) [-- passthrough flags ...]:
  compose a yml file from snippets.
  -l string
        Path to library file
  -s value
        Scenario name in library
  -t string
        Path to template file
```

# interpolation and arguments

There are four argument scopes:
- snippet args
- scenario args
- scenario template args
- scenario global args
- CLI global args

Every snippet is used in two interpolations:
- first the snippet itself is interpolated with snippet args, scenario args, and global args
- then the template is interpolated with the interpolated snippet, template args, and global args

# multiple libraries
Libraries can include other libraries by specifying the path and an alias under 
`libraries:`. If a scenario needs to include a scenario from a referenced 
library the name should be prefixed with `<library alias>.`.

If multiple independant libraries are provided to the CLI all scenario names 
should be unambiguous.

# sample
```
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
  - "base"
  args:
  - -v 
  - foo=overridden
  snippets:
  - path: ./another_opsfile.yml
```

Invoking:
`./manifer compose -l library.yml -t template.yml -s derived -- -v extra=arg`

Is equivalend to:
```
# apply base scenario
bosh interpolate -v path=value -v foo=bar -v foo=overridden -v extra=arg ./opsfile.yml > /tmp/op1
bosh interpolate -v foo=bar -v foo=overridden -v extra=arg -o /tmp/op1 template.yml > /tmp/t1
# apply derived scenario
bosh interpolate -v foo=overridden -v extra=arg ./another_opsfile.yml > /tmp/op2
bosh interpolate -v foo=overridden -v extra=arg -o /tmp/op2 /tmp/t1
```
