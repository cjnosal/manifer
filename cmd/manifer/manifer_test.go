package main

import (
	"github.com/cjnosal/manifer/v2/test"
	"github.com/google/go-cmp/cmp"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

func TestBuild(t *testing.T) {

	cmd := exec.Command(
		"../../scripts/build.sh",
	)
	err := cmd.Run()
	if err != nil {
		t.Errorf("Failed to build: %v", err)
	}

	t.Run("TestCompose", func(t *testing.T) {
		cmd := exec.Command(
			"../../manifer",
			"compose",
			"-l",
			"../../test/data/v2/library.yml",
			"-t",
			"../../test/data/v2/template.yml",
			"-s",
			"placeholder",
			"--",
			"-v",
			"path3=/final?",
			"-v",
			"value3=touch",
			"-l",
			"../../test/data/v2/vars.yml",
		)
		outWriter := &test.StringWriter{}
		errWriter := &test.StringWriter{}
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
		}

		expectedOut := `base2: basic_from_placeholder
final: touch
fixed: from_scenario
foo: bar
reused: by_second
set: by_first
`

		if !cmp.Equal(outWriter.String(), expectedOut) {
			t.Errorf("Expected Stdout:\n'''%v'''\nActual:\n'''%v'''\nDiff:\n'''%v'''\n",
				expectedOut, outWriter.String(), cmp.Diff(expectedOut, outWriter.String()))
		}
	})

	t.Run("TestCompose show plan", func(t *testing.T) {
		cmd := exec.Command(
			"../../manifer",
			"compose",
			"-l",
			"../../test/data/v2/library.yml",
			"-t",
			"../../test/data/v2/template.yml",
			"-s",
			"placeholder",
			"-p",
			"--",
			"-v",
			"path3=/final?",
			"-v",
			"value3=touch",
			"-l",
			"../../test/data/v2/vars.yml",
		)
		outWriter := &test.StringWriter{}
		errWriter := &test.StringWriter{}
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
		}

		expectedOut := `base2: basic_from_placeholder
final: touch
fixed: from_scenario
foo: bar
reused: by_second
set: by_first
`

		if !cmp.Equal(outWriter.String(), expectedOut) {
			t.Errorf("Expected Stdout:\n'''%v'''\nActual:\n'''%v'''\nDiff:\n'''%v'''\n",
				expectedOut, outWriter.String(), cmp.Diff(expectedOut, outWriter.String()))
		}

		expectedErr := `
snippet: ../../test/data/v2/placeholder_opsfile.yml
processor:
    type: opsfile
interpolator:
    vars:
        path1: /fixed?
        path2: /base2?
        path3: /base3?
        value1: from_scenario
        value2: basic_from_placeholder
    raw_args:
      - -v
      - path3=/final?
      - -v
      - value3=touch
      - -l
      - ../../test/data/v2/vars.yml

snippet: ../../test/data/v2/placeholder_opsfile.yml
processor:
    type: opsfile
interpolator:
    vars:
        path1: /fixed?
        path2: /set?
        value1: from_scenario
        value2: by_first
    raw_args:
      - -v
      - path3=/final?
      - -v
      - value3=touch
      - -l
      - ../../test/data/v2/vars.yml

snippet: ../../test/data/v2/placeholder_opsfile.yml
processor:
    type: opsfile
interpolator:
    vars:
        path1: /fixed?
        path2: /reused?
        value1: from_scenario
        value2: by_second
    raw_args:
      - -v
      - path3=/final?
      - -v
      - value3=touch
      - -l
      - ../../test/data/v2/vars.yml

interpolator:
    raw_args:
      - -v
      - path3=/final?
      - -v
      - value3=touch
      - -l
      - ../../test/data/v2/vars.yml
`
		if !cmp.Equal(errWriter.String(), expectedErr) {
			t.Errorf("Expected Stderr:\n'''%v'''\nActual:\n'''%v'''\nDiff:\n'''%v'''\n",
				expectedErr, errWriter.String(), cmp.Diff(expectedErr, errWriter.String()))
		}
	})

	t.Run("TestCompose show diff", func(t *testing.T) {
		cmd := exec.Command(
			"../../manifer",
			"compose",
			"-l",
			"../../test/data/v2/library.yml",
			"-t",
			"../../test/data/v2/template.yml",
			"-s",
			"placeholder",
			"-d",
			"--",
			"-v",
			"path3=/final?",
			"-v",
			"value3=touch",
			"-l",
			"../../test/data/v2/vars.yml",
		)
		outWriter := &test.StringWriter{}
		errWriter := &test.StringWriter{}
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
		}

		expectedOut := `base2: basic_from_placeholder
final: touch
fixed: from_scenario
foo: bar
reused: by_second
set: by_first
`

		if !cmp.Equal(outWriter.String(), expectedOut) {
			t.Errorf("Expected Stdout:\n'''%v'''\nActual:\n'''%v'''\nDiff:\n'''%v'''\n",
				expectedOut, outWriter.String(), cmp.Diff(expectedOut, outWriter.String()))
		}

		ansiGreen := []byte("\x1b\x5b\x33\x32\x6d")
		ansiStop := []byte("\x1b\x5b\x30\x6d")

		expectedWriter := test.ByteWriter{}
		expectedWriter.Write([]byte(`
Diff:
`))
		expectedWriter.Write(ansiGreen)
		expectedWriter.Write([]byte(`base2: basic_from_placeholder
final: touch
fixed: from_scenario
`))
		expectedWriter.Write(ansiStop)
		expectedWriter.Write([]byte(`foo: bar

Diff:
base2: basic_from_placeholder
final: touch
fixed: from_scenario
foo: bar
`))
		expectedWriter.Write(ansiGreen)
		expectedWriter.Write([]byte(`set: by_first
`))
		expectedWriter.Write(ansiStop)
		expectedWriter.Write([]byte(`
Diff:
base2: basic_from_placeholder
final: touch
fixed: from_scenario
foo: bar
`))
		expectedWriter.Write(ansiGreen)
		expectedWriter.Write([]byte(`reused: by_second
`))
		expectedWriter.Write(ansiStop)
		expectedWriter.Write([]byte(`set: by_first

Diff:
`))

		expectedErr := string(expectedWriter.Bytes())

		if !cmp.Equal(errWriter.String(), expectedErr) {
			t.Errorf("Expected Stderr:\n'''%v'''\nActual:\n'''%v'''\nDiff:\n'''%v'''\n",
				expectedErr, errWriter.String(), cmp.Diff(expectedErr, errWriter.String()))
		}
	})

	t.Run("TestCompose Additional Compositions", func(t *testing.T) {
		cmd := exec.Command(
			"../../manifer",
			"compose",
			"-l",
			"../../test/data/v2/library.yml",
			"-t",
			"../../test/data/v2/template.yml",
			"-s",
			"placeholder",
			"--",
			"-v",
			"path3=/final?",
			"-v",
			"value3=touch",
			";",
			"-l",
			"../../test/data/v2/base_library.yml",
			"-s",
			"placeholder",
			"-s",
			"base",
			"--",
			"-v",
			"path3=/what?",
			"-v",
			"value3=now",
		)
		outWriter := &test.StringWriter{}
		errWriter := &test.StringWriter{}
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
		}

		expectedOut := `base1: a
base2: b
final: touch
fixed: from_scenario
foo: bar
reused: by_second
set: by_first
what: now
`

		if !cmp.Equal(outWriter.String(), expectedOut) {
			t.Errorf("Expected Stdout:\n'''%v'''\nActual:\n'''%v'''\nDiff:\n'''%v'''\n",
				expectedOut, outWriter.String(), cmp.Diff(expectedOut, outWriter.String()))
		}
	})

	t.Run("TestListYaml", func(t *testing.T) {
		cmd := exec.Command(
			"../../manifer",
			"list",
			"-l",
			"../../test/data/v2/library.yml",
		)
		outWriter := &test.StringWriter{}
		errWriter := &test.StringWriter{}
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
		}

		expected := `- name: bizz
  description: adds an op
- name: empty
  description: contributes nothing
- name: placeholder
  description: replaces placeholder values
- name: basic
  description: a starting point
`

		if !cmp.Equal(outWriter.String(), expected) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\nDiff:\n'''%s'''\n",
				expected, outWriter.String(), cmp.Diff(expected, outWriter.String()))
		}
	})

	t.Run("TestListJson", func(t *testing.T) {
		cmd := exec.Command(
			"../../manifer",
			"list",
			"-l",
			"../../test/data/v2/library.yml",
			"-j",
		)
		outWriter := &test.StringWriter{}
		errWriter := &test.StringWriter{}
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
		}

		expected := `[{"Name":"bizz","Description":"adds an op"},{"Name":"empty","Description":"contributes nothing"},{"Name":"placeholder","Description":"replaces placeholder values"},{"Name":"basic","Description":"a starting point"}]`

		if !cmp.Equal(outWriter.String(), expected) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\nDiff:\n'''%s'''\n",
				expected, outWriter.String(), cmp.Diff(expected, outWriter.String()))
		}
	})

	t.Run("TestSearchYaml", func(t *testing.T) {
		cmd := exec.Command(
			"../../manifer",
			"search",
			"-l",
			"../../test/data/v2/library.yml",
			"bizz",
			"contributes",
		)
		outWriter := &test.StringWriter{}
		errWriter := &test.StringWriter{}
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
		}

		expected := `- name: bizz
  description: adds an op
- name: empty
  description: contributes nothing
`

		if !cmp.Equal(outWriter.String(), expected) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\nDiff:\n'''%s'''\n",
				expected, outWriter.String(), cmp.Diff(expected, outWriter.String()))
		}
	})

	t.Run("TestSearchJson", func(t *testing.T) {
		cmd := exec.Command(
			"../../manifer",
			"search",
			"-l",
			"../../test/data/v2/library.yml",
			"-j",
			"bizz",
			"contributes",
		)
		outWriter := &test.StringWriter{}
		errWriter := &test.StringWriter{}
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
		}

		expected := `[{"Name":"bizz","Description":"adds an op"},{"Name":"empty","Description":"contributes nothing"}]`

		if !cmp.Equal(outWriter.String(), expected) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\nDiff:\n'''%s'''\n",
				expected, outWriter.String(), cmp.Diff(expected, outWriter.String()))
		}
	})

	t.Run("TestInspect", func(t *testing.T) {
		t.Run("Yaml Tree", func(t *testing.T) {
			cmd := exec.Command(
				"../../manifer",
				"inspect",
				"-l",
				"../../test/data/v2/ref_library.yml",
				"--tree",
				"-s",
				"meta",
				"--",
				"-o=../../test/data/v2/ops_file_with_vars.yml",
				"-v=value=lastbit",
			)
			outWriter := &test.StringWriter{}
			errWriter := &test.StringWriter{}
			cmd.Stdout = outWriter
			cmd.Stderr = errWriter

			err := cmd.Run()
			if err != nil {
				t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
			}

			expected := `- name: meta
  library_path: ../../test/data/v2/ref_library.yml
  dependencies:
    - name: base
      library_path: ../../test/data/v2/base_library.yml
      snippets:
        - path: ../../test/data/v2/placeholder_opsfile.yml
          interpolator:
              vars:
                  path1: /base1?
                  path2: /base2?
                  path3: /base3?
                  value1: a
                  value2: b
                  value3: c
          processor:
              type: opsfile
    - name: placeholder
      description: replaces placeholder values
      library_path: ../../test/data/v2/library.yml
      interpolator:
          vars:
              path1: /fixed?
              value1: from_scenario
      snippets:
        - path: ../../test/data/v2/placeholder_opsfile.yml
          interpolator:
              vars:
                  path2: /set?
                  value2: by_first
          processor:
              type: opsfile
        - path: ../../test/data/v2/placeholder_opsfile.yml
          interpolator:
              vars:
                  path2: /reused?
                  value2: by_second
          processor:
              type: opsfile
      dependencies:
        - name: basic
          description: a starting point
          library_path: ../../test/data/v2/library.yml
          ref_interpolator:
              vars:
                  value2: basic_from_placeholder
          interpolator:
              vars:
                  path1: /base1?
                  value1: from_basic
          snippets:
            - path: ../../test/data/v2/placeholder_opsfile.yml
              interpolator:
                  vars:
                      path2: /base2?
                      path3: /base3?
              processor:
                  type: opsfile
- name: passthrough opsfile
  description: args passed after --
  library_path: <cli>
  snippets:
    - path: ../../test/data/v2/ops_file_with_vars.yml
      processor:
          type: opsfile
- name: passthrough variables
  description: vars passed after --
  library_path: <cli>
  global_interpolator:
      raw_args:
        - -v=value=lastbit
`
			if !cmp.Equal(outWriter.String(), expected) {
				t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\nDiff:\n'''%s'''\n",
					expected, outWriter.String(), cmp.Diff(expected, outWriter.String()))
			}
		})

		t.Run("Yaml Plan", func(t *testing.T) {
			cmd := exec.Command(
				"../../manifer",
				"inspect",
				"-l",
				"../../test/data/v2/ref_library.yml",
				"--plan",
				"-s",
				"meta",
				"--",
				"-o=../../test/data/v2/ops_file_with_vars.yml",
				"-v=value=lastbit",
			)
			outWriter := &test.StringWriter{}
			errWriter := &test.StringWriter{}
			cmd.Stdout = outWriter
			cmd.Stderr = errWriter

			err := cmd.Run()
			if err != nil {
				t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
			}

			expected := `global:
    raw_args:
      - -v=value=lastbit
steps:
  - snippet: ../../test/data/v2/placeholder_opsfile.yml
    params:
      - tag: snippet
        interpolator:
            vars:
                path1: /base1?
                path2: /base2?
                path3: /base3?
                value1: a
                value2: b
                value3: c
      - tag: base
      - tag: meta
    processor:
        type: opsfile
  - snippet: ../../test/data/v2/placeholder_opsfile.yml
    params:
      - tag: snippet
        interpolator:
            vars:
                path2: /base2?
                path3: /base3?
      - tag: basic
        interpolator:
            vars:
                path1: /base1?
                value1: from_basic
                value2: basic_from_placeholder
      - tag: placeholder
        interpolator:
            vars:
                path1: /fixed?
                value1: from_scenario
      - tag: meta
    processor:
        type: opsfile
  - snippet: ../../test/data/v2/placeholder_opsfile.yml
    params:
      - tag: snippet
        interpolator:
            vars:
                path2: /set?
                value2: by_first
      - tag: placeholder
        interpolator:
            vars:
                path1: /fixed?
                value1: from_scenario
      - tag: meta
    processor:
        type: opsfile
  - snippet: ../../test/data/v2/placeholder_opsfile.yml
    params:
      - tag: snippet
        interpolator:
            vars:
                path2: /reused?
                value2: by_second
      - tag: placeholder
        interpolator:
            vars:
                path1: /fixed?
                value1: from_scenario
      - tag: meta
    processor:
        type: opsfile
  - snippet: ../../test/data/v2/ops_file_with_vars.yml
    params:
      - tag: snippet
      - tag: passthrough opsfile
    processor:
        type: opsfile
`

			if !cmp.Equal(outWriter.String(), expected) {
				t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\nDiff:\n'''%s'''\n",
					expected, outWriter.String(), cmp.Diff(expected, outWriter.String()))
			}
		})

	})

	t.Run("TestGenerate from template", func(t *testing.T) {

		exec.Command(
			"rm",
			"-rf",
			"../../test/data/v2/generated_ops",
			"../../test/data/v2/generated.yml",
		).Run()

		cmd := exec.Command(
			"../../manifer",
			"generate",
			"-y",
			"opsfile",
			"-t",
			"../../test/data/v2/base_library.yml",
			"-d",
			"../../test/data/v2/generated_ops",
			"-o",
			"../../test/data/v2/generated.yml",
		)

		outWriter := &test.StringWriter{}
		errWriter := &test.StringWriter{}
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
		}

		cat := exec.Command(
			"cat",
			"../../test/data/v2/generated.yml",
		)
		outWriter = &test.StringWriter{}
		cat.Stdout = outWriter

		err = cat.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expectedOut := `scenarios:
  - name: add_scenario
    description: replace /scenarios?/- (imported from generated_ops/add_scenario.yml)
    snippets:
      - path: generated_ops/add_scenario.yml
        processor:
            type: opsfile
  - name: add_snippet
    description: replace /scenarios/((scenario_index))/snippets?/- (imported from
        generated_ops/scenario/add_snippet.yml)
    snippets:
      - path: generated_ops/scenario/add_snippet.yml
        processor:
            type: opsfile
  - name: set_interpolator
    description: replace /scenarios/((scenario_index))/snippets/((snippet_index))/interpolator?
        (imported from generated_ops/scenario/snippet/set_interpolator.yml)
    snippets:
      - path: generated_ops/scenario/snippet/set_interpolator.yml
        processor:
            type: opsfile
  - name: set_scenario
    description: replace /scenarios/((scenario_index)) (imported from generated_ops/set_scenario.yml)
    snippets:
      - path: generated_ops/set_scenario.yml
        processor:
            type: opsfile
  - name: set_snippet
    description: replace /scenarios/((scenario_index))/snippets/((snippet_index))
        (imported from generated_ops/scenario/set_snippet.yml)
    snippets:
      - path: generated_ops/scenario/set_snippet.yml
        processor:
            type: opsfile
  - name: set_type
    description: replace /type? (imported from generated_ops/set_type.yml)
    snippets:
      - path: generated_ops/set_type.yml
        processor:
            type: opsfile
  - name: set_vars
    description: replace /scenarios/((scenario_index))/snippets/((snippet_index))/interpolator/vars?
        (imported from generated_ops/scenario/snippet/interpolator/set_vars.yml)
    snippets:
      - path: generated_ops/scenario/snippet/interpolator/set_vars.yml
        processor:
            type: opsfile
`

		if !cmp.Equal(outWriter.String(), expectedOut) {
			t.Errorf("Expected Stdout:\n'''%v'''\nActual:\n'''%v'''\nDiff:\n'''%v'''\n",
				expectedOut, outWriter.String(), cmp.Diff(expectedOut, outWriter.String()))
		}
	})

	t.Run("TestImport file", func(t *testing.T) {

		exec.Command(
			"rm",
			"-rf",
			"../../test/data/v2/generated_ops",
			"../../test/data/v2/generated.yml",
		).Run()

		cmd := exec.Command(
			"../../manifer",
			"import",
			"-p",
			"../../test/data/v2/opsfile.yml",
			"-o",
			"../../test/data/v2/generated.yml",
		)

		outWriter := &test.StringWriter{}
		errWriter := &test.StringWriter{}
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
		}

		cat := exec.Command(
			"cat",
			"../../test/data/v2/generated.yml",
		)
		outWriter = &test.StringWriter{}
		cat.Stdout = outWriter

		err = cat.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expectedOut := `scenarios:
  - name: opsfile
    description: replace /bizz? (imported from opsfile.yml)
    snippets:
      - path: opsfile.yml
        processor:
            type: opsfile
`

		if !cmp.Equal(outWriter.String(), expectedOut) {
			t.Errorf("Expected Stdout:\n'''%v'''\nActual:\n'''%v'''\nDiff:\n'''%v'''\n",
				expectedOut, outWriter.String(), cmp.Diff(expectedOut, outWriter.String()))
		}
	})

	t.Run("TestImport directory", func(t *testing.T) {

		exec.Command(
			"rm",
			"-rf",
			"../../test/data/v2/generated_ops",
			"../../test/data/v2/generated.yml",
		).Run()

		cmd := exec.Command(
			"../../manifer",
			"import",
			"-r",
			"-p",
			"../../test/data/v2",
			"-o",
			"../../test/data/v2/generated.yml",
		)

		outWriter := &test.StringWriter{}
		errWriter := &test.StringWriter{}
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
		}

		cat := exec.Command(
			"cat",
			"../../test/data/v2/generated.yml",
		)
		outWriter = &test.StringWriter{}
		cat.Stdout = outWriter

		err = cat.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expectedOut := `scenarios:
  - name: opsfile
    description: replace /bizz? (imported from opsfile.yml)
    snippets:
      - path: opsfile.yml
        processor:
            type: opsfile
  - name: opsfile_with_vars
    description: replace /elsewhere? (imported from opsfile_with_vars.yml)
    snippets:
      - path: opsfile_with_vars.yml
        processor:
            type: opsfile
  - name: placeholder_opsfile
    description: replace ((path1)) (imported from placeholder_opsfile.yml)
    snippets:
      - path: placeholder_opsfile.yml
        processor:
            type: opsfile
  - name: base_library
    description: write type (imported from base_library.yml)
    snippets:
      - path: base_library.yml
        processor:
            type: yq
  - name: library
    description: write type (imported from library.yml)
    snippets:
      - path: library.yml
        processor:
            type: yq
  - name: ref_library
    description: write type (imported from ref_library.yml)
    snippets:
      - path: ref_library.yml
        processor:
            type: yq
  - name: template
    description: write foo (imported from template.yml)
    snippets:
      - path: template.yml
        processor:
            type: yq
  - name: template_with_var
    description: write foo (imported from template_with_var.yml)
    snippets:
      - path: template_with_var.yml
        processor:
            type: yq
  - name: vars
    description: write foo (imported from vars.yml)
    snippets:
      - path: vars.yml
        processor:
            type: yq
  - name: yq_library
    description: write type (imported from yq_library.yml)
    snippets:
      - path: yq_library.yml
        processor:
            type: yq
  - name: yq_script
    description: write foo (imported from yq_script.yml)
    snippets:
      - path: yq_script.yml
        processor:
            type: yq
  - name: yq_template
    description: write bazz (imported from yq_template.yml)
    snippets:
      - path: yq_template.yml
        processor:
            type: yq
`

		if !cmp.Equal(outWriter.String(), expectedOut) {
			t.Errorf("Expected Stdout:\n'''%v'''\nActual:\n'''%v'''\nDiff:\n'''%v'''\n",
				expectedOut, outWriter.String(), cmp.Diff(expectedOut, outWriter.String()))
		}
	})

	t.Run("TestAddScenario", func(t *testing.T) {

		exec.Command(
			"rm",
			"-rf",
			"../../test/data/v2/generated_ops",
			"../../test/data/v2/generated.yml",
		).Run()

		emptyLib := []byte(`
type: opsfile
scenarios:
 - name: dep`)
		err := ioutil.WriteFile("../../test/data/v2/generated.yml", emptyLib, 0644)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		cmd := exec.Command(
			"../../manifer",
			"add",
			"-l",
			"../../test/data/v2/generated.yml",
			"-n",
			"new scenario",
			"-d",
			"scenario description",
			"-s",
			"dep",
			"--",
			"-o",
			"../../test/data/v2/opsfile_with_vars.yml",
			"-v",
			"value=foo",
		)

		outWriter := &test.StringWriter{}
		errWriter := &test.StringWriter{}
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter

		err = cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
		}

		cat := exec.Command(
			"cat",
			"../../test/data/v2/generated.yml",
		)
		outWriter = &test.StringWriter{}
		cat.Stdout = outWriter

		err = cat.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expectedOut := `type: opsfile
scenarios:
  - name: dep
  - name: new scenario
    description: scenario description
    interpolator:
        raw_args:
          - -v
          - value=foo
    snippets:
      - path: opsfile_with_vars.yml
        processor:
            type: opsfile
    scenarios:
      - name: dep
`

		if !cmp.Equal(outWriter.String(), expectedOut) {
			t.Errorf("Expected Stdout:\n'''%v'''\nActual:\n'''%v'''\nDiff:\n'''%v'''\n",
				expectedOut, outWriter.String(), cmp.Diff(expectedOut, outWriter.String()))
		}
	})

	t.Run("TestLocalLibFlag", func(t *testing.T) {
		cmd := exec.Command(
			"../../manifer",
			"list",
			"-l",
			"../../test/data/v2/library.yml",
		)
		outWriter := &test.StringWriter{}
		errWriter := &test.StringWriter{}
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
		}
	})

	t.Run("TestGlobalLibFlag", func(t *testing.T) {
		cmd := exec.Command(
			"../../manifer",
			"-l",
			"../../test/data/v2/library.yml",
			"list",
		)
		outWriter := &test.StringWriter{}
		errWriter := &test.StringWriter{}
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
		}
	})

	t.Run("TestEnvLibs", func(t *testing.T) {
		cmd := exec.Command(
			"../../manifer",
			"list",
		)
		cmd.Env = append(os.Environ(), "MANIFER_LIBS=../../test/data/v2/library.yml")
		outWriter := &test.StringWriter{}
		errWriter := &test.StringWriter{}
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
		}
	})

	t.Run("TestEnvLibPath", func(t *testing.T) {
		cmd := exec.Command(
			"../../manifer",
			"list",
		)
		cmd.Env = append(os.Environ(), "MANIFER_LIB_PATH=../../test/data/v2")
		outWriter := &test.StringWriter{}
		errWriter := &test.StringWriter{}
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v\n%s", err, errWriter.String())
		}
	})
}
