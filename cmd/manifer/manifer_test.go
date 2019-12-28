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
			"-p",
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

		// includes ascii color codes
		// use test.ByteWriter and strings.ReplaceAll(fmt.Sprintf("%v", errWriter.Bytes()), " ", ",") to print actual output to update
		expectedStdErr := string([]byte{10, 83, 110, 105, 112, 112, 101, 116, 32, 46, 46, 47, 46, 46, 47, 116, 101, 115, 116, 47, 100, 97, 116, 97, 47, 118, 50, 47, 112, 108, 97, 99, 101, 104, 111, 108, 100, 101, 114, 95, 111, 112, 115, 102, 105, 108, 101, 46, 121, 109, 108, 59, 32, 80, 97, 114, 97, 109, 115, 32, 123, 86, 97, 114, 115, 58, 109, 97, 112, 91, 112, 97, 116, 104, 49, 58, 47, 102, 105, 120, 101, 100, 63, 32, 112, 97, 116, 104, 50, 58, 47, 98, 97, 115, 101, 50, 63, 32, 112, 97, 116, 104, 51, 58, 47, 98, 97, 115, 101, 51, 63, 32, 118, 97, 108, 117, 101, 49, 58, 102, 114, 111, 109, 95, 115, 99, 101, 110, 97, 114, 105, 111, 32, 118, 97, 108, 117, 101, 50, 58, 98, 97, 115, 105, 99, 95, 102, 114, 111, 109, 95, 112, 108, 97, 99, 101, 104, 111, 108, 100, 101, 114, 93, 32, 86, 97, 114, 70, 105, 108, 101, 115, 58, 109, 97, 112, 91, 93, 32, 86, 97, 114, 115, 70, 105, 108, 101, 115, 58, 91, 93, 32, 86, 97, 114, 115, 69, 110, 118, 58, 91, 93, 32, 86, 97, 114, 115, 83, 116, 111, 114, 101, 58, 32, 82, 97, 119, 65, 114, 103, 115, 58, 91, 45, 118, 32, 112, 97, 116, 104, 51, 61, 47, 102, 105, 110, 97, 108, 63, 32, 45, 118, 32, 118, 97, 108, 117, 101, 51, 61, 116, 111, 117, 99, 104, 32, 45, 108, 32, 46, 46, 47, 46, 46, 47, 116, 101, 115, 116, 47, 100, 97, 116, 97, 47, 118, 50, 47, 118, 97, 114, 115, 46, 121, 109, 108, 93, 125, 10, 68, 105, 102, 102, 58, 10, 27, 91, 51, 50, 109, 98, 97, 115, 101, 50, 58, 32, 98, 97, 115, 105, 99, 95, 102, 114, 111, 109, 95, 112, 108, 97, 99, 101, 104, 111, 108, 100, 101, 114, 10, 102, 105, 110, 97, 108, 58, 32, 116, 111, 117, 99, 104, 10, 102, 105, 120, 101, 100, 58, 32, 102, 114, 111, 109, 95, 115, 99, 101, 110, 97, 114, 105, 111, 10, 27, 91, 48, 109, 102, 111, 111, 58, 32, 98, 97, 114, 10, 10, 83, 110, 105, 112, 112, 101, 116, 32, 46, 46, 47, 46, 46, 47, 116, 101, 115, 116, 47, 100, 97, 116, 97, 47, 118, 50, 47, 112, 108, 97, 99, 101, 104, 111, 108, 100, 101, 114, 95, 111, 112, 115, 102, 105, 108, 101, 46, 121, 109, 108, 59, 32, 80, 97, 114, 97, 109, 115, 32, 123, 86, 97, 114, 115, 58, 109, 97, 112, 91, 112, 97, 116, 104, 49, 58, 47, 102, 105, 120, 101, 100, 63, 32, 112, 97, 116, 104, 50, 58, 47, 115, 101, 116, 63, 32, 118, 97, 108, 117, 101, 49, 58, 102, 114, 111, 109, 95, 115, 99, 101, 110, 97, 114, 105, 111, 32, 118, 97, 108, 117, 101, 50, 58, 98, 121, 95, 102, 105, 114, 115, 116, 93, 32, 86, 97, 114, 70, 105, 108, 101, 115, 58, 109, 97, 112, 91, 93, 32, 86, 97, 114, 115, 70, 105, 108, 101, 115, 58, 91, 93, 32, 86, 97, 114, 115, 69, 110, 118, 58, 91, 93, 32, 86, 97, 114, 115, 83, 116, 111, 114, 101, 58, 32, 82, 97, 119, 65, 114, 103, 115, 58, 91, 45, 118, 32, 112, 97, 116, 104, 51, 61, 47, 102, 105, 110, 97, 108, 63, 32, 45, 118, 32, 118, 97, 108, 117, 101, 51, 61, 116, 111, 117, 99, 104, 32, 45, 108, 32, 46, 46, 47, 46, 46, 47, 116, 101, 115, 116, 47, 100, 97, 116, 97, 47, 118, 50, 47, 118, 97, 114, 115, 46, 121, 109, 108, 93, 125, 10, 68, 105, 102, 102, 58, 10, 98, 97, 115, 101, 50, 58, 32, 98, 97, 115, 105, 99, 95, 102, 114, 111, 109, 95, 112, 108, 97, 99, 101, 104, 111, 108, 100, 101, 114, 10, 102, 105, 110, 97, 108, 58, 32, 116, 111, 117, 99, 104, 10, 102, 105, 120, 101, 100, 58, 32, 102, 114, 111, 109, 95, 115, 99, 101, 110, 97, 114, 105, 111, 10, 102, 111, 111, 58, 32, 98, 97, 114, 10, 27, 91, 51, 50, 109, 115, 101, 116, 58, 32, 98, 121, 95, 102, 105, 114, 115, 116, 10, 27, 91, 48, 109, 10, 83, 110, 105, 112, 112, 101, 116, 32, 46, 46, 47, 46, 46, 47, 116, 101, 115, 116, 47, 100, 97, 116, 97, 47, 118, 50, 47, 112, 108, 97, 99, 101, 104, 111, 108, 100, 101, 114, 95, 111, 112, 115, 102, 105, 108, 101, 46, 121, 109, 108, 59, 32, 80, 97, 114, 97, 109, 115, 32, 123, 86, 97, 114, 115, 58, 109, 97, 112, 91, 112, 97, 116, 104, 49, 58, 47, 102, 105, 120, 101, 100, 63, 32, 112, 97, 116, 104, 50, 58, 47, 114, 101, 117, 115, 101, 100, 63, 32, 118, 97, 108, 117, 101, 49, 58, 102, 114, 111, 109, 95, 115, 99, 101, 110, 97, 114, 105, 111, 32, 118, 97, 108, 117, 101, 50, 58, 98, 121, 95, 115, 101, 99, 111, 110, 100, 93, 32, 86, 97, 114, 70, 105, 108, 101, 115, 58, 109, 97, 112, 91, 93, 32, 86, 97, 114, 115, 70, 105, 108, 101, 115, 58, 91, 93, 32, 86, 97, 114, 115, 69, 110, 118, 58, 91, 93, 32, 86, 97, 114, 115, 83, 116, 111, 114, 101, 58, 32, 82, 97, 119, 65, 114, 103, 115, 58, 91, 45, 118, 32, 112, 97, 116, 104, 51, 61, 47, 102, 105, 110, 97, 108, 63, 32, 45, 118, 32, 118, 97, 108, 117, 101, 51, 61, 116, 111, 117, 99, 104, 32, 45, 108, 32, 46, 46, 47, 46, 46, 47, 116, 101, 115, 116, 47, 100, 97, 116, 97, 47, 118, 50, 47, 118, 97, 114, 115, 46, 121, 109, 108, 93, 125, 10, 68, 105, 102, 102, 58, 10, 98, 97, 115, 101, 50, 58, 32, 98, 97, 115, 105, 99, 95, 102, 114, 111, 109, 95, 112, 108, 97, 99, 101, 104, 111, 108, 100, 101, 114, 10, 102, 105, 110, 97, 108, 58, 32, 116, 111, 117, 99, 104, 10, 102, 105, 120, 101, 100, 58, 32, 102, 114, 111, 109, 95, 115, 99, 101, 110, 97, 114, 105, 111, 10, 102, 111, 111, 58, 32, 98, 97, 114, 10, 27, 91, 51, 50, 109, 114, 101, 117, 115, 101, 100, 58, 32, 98, 121, 95, 115, 101, 99, 111, 110, 100, 10, 27, 91, 48, 109, 115, 101, 116, 58, 32, 98, 121, 95, 102, 105, 114, 115, 116, 10, 10, 83, 110, 105, 112, 112, 101, 116, 32, 59, 32, 80, 97, 114, 97, 109, 115, 32, 123, 86, 97, 114, 115, 58, 109, 97, 112, 91, 93, 32, 86, 97, 114, 70, 105, 108, 101, 115, 58, 109, 97, 112, 91, 93, 32, 86, 97, 114, 115, 70, 105, 108, 101, 115, 58, 91, 93, 32, 86, 97, 114, 115, 69, 110, 118, 58, 91, 93, 32, 86, 97, 114, 115, 83, 116, 111, 114, 101, 58, 32, 82, 97, 119, 65, 114, 103, 115, 58, 91, 45, 118, 32, 112, 97, 116, 104, 51, 61, 47, 102, 105, 110, 97, 108, 63, 32, 45, 118, 32, 118, 97, 108, 117, 101, 51, 61, 116, 111, 117, 99, 104, 32, 45, 108, 32, 46, 46, 47, 46, 46, 47, 116, 101, 115, 116, 47, 100, 97, 116, 97, 47, 118, 50, 47, 118, 97, 114, 115, 46, 121, 109, 108, 93, 125, 10, 68, 105, 102, 102, 58, 10})

		if !cmp.Equal(errWriter.String(), expectedStdErr) {
			t.Errorf("Expected Stderr:\n'''%v'''\nActual:\n'''%v'''\nDiff:\n'''%v'''\n",
				expectedStdErr, errWriter.String(), cmp.Diff(expectedStdErr, errWriter.String()))
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

	t.Run("TestListPlain", func(t *testing.T) {
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

		expected := `bizz
	adds an op

empty
	contributes nothing

placeholder
	replaces placeholder values

basic
	a starting point

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

	t.Run("TestSearchPlain", func(t *testing.T) {
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

		expected := `bizz
	adds an op

empty
	contributes nothing

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
		t.Run("Plain Tree", func(t *testing.T) {
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

			expected := `name:        meta (from ../../test/data/v2/ref_library.yml)
description: 
global:  {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to all scenarios)
refvars: {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to snippets and subscenarios)
vars:    {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to snippets and subscenarios)
snippets:
dependencies:
  name:        base (from ../../test/data/v2/base_library.yml)
  description: 
  global:  {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to all scenarios)
  refvars: {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to snippets and subscenarios)
  vars:    {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to snippets and subscenarios)
  snippets:
    ../../test/data/v2/placeholder_opsfile.yml
    vars: {Vars:map[path1:/base1? path2:/base2? path3:/base3? value1:a value2:b value3:c] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}

  dependencies:

  name:        placeholder (from ../../test/data/v2/library.yml)
  description: replaces placeholder values
  global:  {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to all scenarios)
  refvars: {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to snippets and subscenarios)
  vars:    {Vars:map[path1:/fixed? value1:from_scenario] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to snippets and subscenarios)
  snippets:
    ../../test/data/v2/placeholder_opsfile.yml
    vars: {Vars:map[path2:/set? value2:by_first] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}

    ../../test/data/v2/placeholder_opsfile.yml
    vars: {Vars:map[path2:/reused? value2:by_second] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}

  dependencies:
    name:        basic (from ../../test/data/v2/library.yml)
    description: a starting point
    global:  {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to all scenarios)
    refvars: {Vars:map[value2:basic_from_placeholder] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to snippets and subscenarios)
    vars:    {Vars:map[path1:/base1? value1:from_basic] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to snippets and subscenarios)
    snippets:
      ../../test/data/v2/placeholder_opsfile.yml
      vars: {Vars:map[path2:/base2? path3:/base3?] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}

    dependencies:


name:        passthrough (from <cli>)
description: args passed after --
global:  {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to all scenarios)
refvars: {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to snippets and subscenarios)
vars:    {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to snippets and subscenarios)
snippets:
  ../../test/data/v2/ops_file_with_vars.yml
  vars: {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}

dependencies:
name:        passthrough variables (from <cli>)
description: vars passed after --
global:  {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[-v=value=lastbit]} (applied to all scenarios)
refvars: {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to snippets and subscenarios)
vars:    {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]} (applied to snippets and subscenarios)
snippets:
dependencies:
`
			if !cmp.Equal(writer.String(), expected) {
				t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\nDiff:\n'''%s'''\n",
					expected, writer.String(), cmp.Diff(expected, writer.String()))
			}
		})

		t.Run("Plain Plan", func(t *testing.T) {
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

			expected := `global: {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[-v=value=lastbit]}
- ../../test/data/v2/placeholder_opsfile.yml
  vars:
    snippet: {Vars:map[path1:/base1? path2:/base2? path3:/base3? value1:a value2:b value3:c] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}
    base: {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}
    meta: {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}
- ../../test/data/v2/placeholder_opsfile.yml
  vars:
    snippet: {Vars:map[path2:/base2? path3:/base3?] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}
    basic: {Vars:map[path1:/base1? value1:from_basic value2:basic_from_placeholder] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}
    placeholder: {Vars:map[path1:/fixed? value1:from_scenario] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}
    meta: {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}
- ../../test/data/v2/placeholder_opsfile.yml
  vars:
    snippet: {Vars:map[path2:/set? value2:by_first] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}
    placeholder: {Vars:map[path1:/fixed? value1:from_scenario] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}
    meta: {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}
- ../../test/data/v2/placeholder_opsfile.yml
  vars:
    snippet: {Vars:map[path2:/reused? value2:by_second] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}
    placeholder: {Vars:map[path1:/fixed? value1:from_scenario] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}
    meta: {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}
- ../../test/data/v2/ops_file_with_vars.yml
  vars:
    snippet: {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}
    passthrough: {Vars:map[] VarFiles:map[] VarsFiles:[] VarsEnv:[] VarsStore: RawArgs:[]}
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

		expectedOut := `libraries: []
type: opsfile
scenarios:
  - name: add_scenario
    description: imported from add_scenario.yml
    snippets:
      - path: generated_ops/add_scenario.yml
    scenarios: []
  - name: add_snippet
    description: imported from add_snippet.yml
    snippets:
      - path: generated_ops/scenario/add_snippet.yml
    scenarios: []
  - name: set_snippet
    description: imported from set_snippet.yml
    snippets:
      - path: generated_ops/scenario/set_snippet.yml
    scenarios: []
  - name: set_vars
    description: imported from set_vars.yml
    snippets:
      - path: generated_ops/scenario/snippet/interpolator/set_vars.yml
    scenarios: []
  - name: set_interpolator
    description: imported from set_interpolator.yml
    snippets:
      - path: generated_ops/scenario/snippet/set_interpolator.yml
    scenarios: []
  - name: set_scenario
    description: imported from set_scenario.yml
    snippets:
      - path: generated_ops/set_scenario.yml
    scenarios: []
  - name: set_type
    description: imported from set_type.yml
    snippets:
      - path: generated_ops/set_type.yml
    scenarios: []
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

		expectedOut := `libraries: []
type: opsfile
scenarios:
  - name: opsfile
    description: imported from opsfile.yml
    snippets:
      - path: opsfile.yml
    scenarios: []
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

		expectedOut := `libraries: []
type: opsfile
scenarios:
  - name: empty_opsfile
    description: imported from empty_opsfile.yml
    snippets:
      - path: empty_opsfile.yml
    scenarios: []
  - name: opsfile
    description: imported from opsfile.yml
    snippets:
      - path: opsfile.yml
    scenarios: []
  - name: opsfile_with_vars
    description: imported from opsfile_with_vars.yml
    snippets:
      - path: opsfile_with_vars.yml
    scenarios: []
  - name: placeholder_opsfile
    description: imported from placeholder_opsfile.yml
    snippets:
      - path: placeholder_opsfile.yml
    scenarios: []
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

		expectedOut := `libraries: []
type: opsfile
scenarios:
  - name: dep
    description: ""
    snippets: []
    scenarios: []
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
