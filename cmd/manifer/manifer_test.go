package main

import (
	"github.com/cjnosal/manifer/test"
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
		cmd.Stdout = outWriter

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
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
			t.Errorf("Unexpected error: %v", err)
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
			t.Errorf("Unexpected error: %v", err)
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
			t.Errorf("Unexpected error: %v", err)
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
		writer := &test.StringWriter{}
		cmd.Stdout = writer

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
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

		if !cmp.Equal(writer.String(), expected) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\nDiff:\n'''%s'''\n",
				expected, writer.String(), cmp.Diff(expected, writer.String()))
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
		writer := &test.StringWriter{}
		cmd.Stdout = writer

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := `[{"Name":"bizz","Description":"adds an op"},{"Name":"empty","Description":"contributes nothing"},{"Name":"placeholder","Description":"replaces placeholder values"},{"Name":"basic","Description":"a starting point"}]`

		if !cmp.Equal(writer.String(), expected) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\nDiff:\n'''%s'''\n",
				expected, writer.String(), cmp.Diff(expected, writer.String()))
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
		writer := &test.StringWriter{}
		cmd.Stdout = writer

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := `- name: bizz
  description: adds an op
- name: empty
  description: contributes nothing
`

		if !cmp.Equal(writer.String(), expected) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\nDiff:\n'''%s'''\n",
				expected, writer.String(), cmp.Diff(expected, writer.String()))
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
		writer := &test.StringWriter{}
		cmd.Stdout = writer

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := `[{"Name":"bizz","Description":"adds an op"},{"Name":"empty","Description":"contributes nothing"}]`

		if !cmp.Equal(writer.String(), expected) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\nDiff:\n'''%s'''\n",
				expected, writer.String(), cmp.Diff(expected, writer.String()))
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
			writer := &test.StringWriter{}
			cmd.Stdout = writer

			err := cmd.Run()
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
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
        - path: ../../test/data/v2/placeholder_opsfile.yml
          interpolator:
              vars:
                  path2: /reused?
                  value2: by_second
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
- name: passthrough
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
			if !cmp.Equal(writer.String(), expected) {
				t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\nDiff:\n'''%s'''\n",
					expected, writer.String(), cmp.Diff(expected, writer.String()))
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
			writer := &test.StringWriter{}
			cmd.Stdout = writer

			err := cmd.Run()
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
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
  - snippet: ../../test/data/v2/ops_file_with_vars.yml
    params:
      - tag: snippet
      - tag: passthrough
    processor:
        type: opsfile
`

			if !cmp.Equal(writer.String(), expected) {
				t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\nDiff:\n'''%s'''\n",
					expected, writer.String(), cmp.Diff(expected, writer.String()))
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
			"-t",
			"../../test/data/v2/base_library.yml",
			"-d",
			"../../test/data/v2/generated_ops",
			"-o",
			"../../test/data/v2/generated.yml",
		)

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		cat := exec.Command(
			"cat",
			"../../test/data/v2/generated.yml",
		)
		outWriter := &test.StringWriter{}
		cat.Stdout = outWriter

		err = cat.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expectedOut := `type: opsfile
scenarios:
  - name: add_scenario
    description: imported from add_scenario.yml
    snippets:
      - path: generated_ops/add_scenario.yml
  - name: add_snippet
    description: imported from add_snippet.yml
    snippets:
      - path: generated_ops/scenario/add_snippet.yml
  - name: set_snippet
    description: imported from set_snippet.yml
    snippets:
      - path: generated_ops/scenario/set_snippet.yml
  - name: set_vars
    description: imported from set_vars.yml
    snippets:
      - path: generated_ops/scenario/snippet/interpolator/set_vars.yml
  - name: set_interpolator
    description: imported from set_interpolator.yml
    snippets:
      - path: generated_ops/scenario/snippet/set_interpolator.yml
  - name: set_scenario
    description: imported from set_scenario.yml
    snippets:
      - path: generated_ops/set_scenario.yml
  - name: set_type
    description: imported from set_type.yml
    snippets:
      - path: generated_ops/set_type.yml
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

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		cat := exec.Command(
			"cat",
			"../../test/data/v2/generated.yml",
		)
		outWriter := &test.StringWriter{}
		cat.Stdout = outWriter

		err = cat.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expectedOut := `type: opsfile
scenarios:
  - name: opsfile
    description: imported from opsfile.yml
    snippets:
      - path: opsfile.yml
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

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		cat := exec.Command(
			"cat",
			"../../test/data/v2/generated.yml",
		)
		outWriter := &test.StringWriter{}
		cat.Stdout = outWriter

		err = cat.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expectedOut := `type: opsfile
scenarios:
  - name: empty_opsfile
    description: imported from empty_opsfile.yml
    snippets:
      - path: empty_opsfile.yml
  - name: opsfile
    description: imported from opsfile.yml
    snippets:
      - path: opsfile.yml
  - name: opsfile_with_vars
    description: imported from opsfile_with_vars.yml
    snippets:
      - path: opsfile_with_vars.yml
  - name: placeholder_opsfile
    description: imported from placeholder_opsfile.yml
    snippets:
      - path: placeholder_opsfile.yml
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

		err = cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		cat := exec.Command(
			"cat",
			"../../test/data/v2/generated.yml",
		)
		outWriter := &test.StringWriter{}
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
		writer := &test.StringWriter{}
		cmd.Stdout = writer

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("TestGlobalLibFlag", func(t *testing.T) {
		cmd := exec.Command(
			"../../manifer",
			"-l",
			"../../test/data/v2/library.yml",
			"list",
		)
		writer := &test.StringWriter{}
		cmd.Stdout = writer

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("TestEnvLibs", func(t *testing.T) {
		cmd := exec.Command(
			"../../manifer",
			"list",
		)
		cmd.Env = append(os.Environ(), "MANIFER_LIBS=../../test/data/v2/library.yml")
		writer := &test.StringWriter{}
		cmd.Stdout = writer

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("TestEnvLibPath", func(t *testing.T) {
		cmd := exec.Command(
			"../../manifer",
			"list",
		)
		cmd.Env = append(os.Environ(), "MANIFER_LIB_PATH=../../test/data/v2")
		writer := &test.StringWriter{}
		cmd.Stdout = writer

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}
