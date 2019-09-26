package main

import (
	"fmt"
	"github.com/cjnosal/manifer/test"
	"os/exec"
	"testing"
)

func TestCompose(t *testing.T) {
	cmd := exec.Command(
		"go",
		"run",
		"manifer.go",
		"compose",
		"-l",
		"../../test/data/library.yml",
		"-t",
		"../../test/data/template.yml",
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
		"../../test/data/vars.yml",
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

	if outWriter.String() != expectedOut {
		t.Errorf("Expected Out:\n'''%v'''\nActual:\n'''%v'''\n", expectedOut, outWriter.String())
	}
	// includes ascii color codes
	// use test.ByteWriter and strings.ReplaceAll(fmt.Sprintf("%v", errWriter.Bytes()), " ", ",") to print actual output to update
	expectedStdErr := []byte{10, 83, 110, 105, 112, 112, 101, 116, 32, 46, 46, 47, 46, 46, 47, 116, 101, 115, 116, 47, 100, 97, 116, 97, 47, 112, 108, 97, 99, 101, 104, 111, 108, 100, 101, 114, 95, 111, 112, 115, 102, 105, 108, 101, 46, 121, 109, 108, 59, 32, 65, 114, 103, 32, 91, 45, 118, 32, 112, 97, 116, 104, 50, 61, 47, 98, 97, 115, 101, 50, 63, 32, 45, 118, 32, 112, 97, 116, 104, 51, 61, 47, 98, 97, 115, 101, 51, 63, 32, 45, 118, 32, 112, 97, 116, 104, 49, 61, 47, 98, 97, 115, 101, 49, 63, 32, 45, 118, 32, 118, 97, 108, 117, 101, 49, 61, 102, 114, 111, 109, 95, 98, 97, 115, 105, 99, 32, 45, 118, 32, 118, 97, 108, 117, 101, 50, 61, 98, 97, 115, 105, 99, 95, 102, 114, 111, 109, 95, 112, 108, 97, 99, 101, 104, 111, 108, 100, 101, 114, 32, 45, 118, 32, 112, 97, 116, 104, 49, 61, 47, 102, 105, 120, 101, 100, 63, 32, 45, 118, 32, 118, 97, 108, 117, 101, 49, 61, 102, 114, 111, 109, 95, 115, 99, 101, 110, 97, 114, 105, 111, 93, 59, 32, 71, 108, 111, 98, 97, 108, 32, 91, 45, 118, 32, 112, 97, 116, 104, 51, 61, 47, 102, 105, 110, 97, 108, 63, 32, 45, 118, 32, 118, 97, 108, 117, 101, 51, 61, 116, 111, 117, 99, 104, 32, 45, 108, 32, 46, 46, 47, 46, 46, 47, 116, 101, 115, 116, 47, 100, 97, 116, 97, 47, 118, 97, 114, 115, 46, 121, 109, 108, 93, 10, 68, 105, 102, 102, 58, 10, 27, 91, 51, 50, 109, 98, 97, 115, 101, 50, 58, 32, 98, 97, 115, 105, 99, 95, 102, 114, 111, 109, 95, 112, 108, 97, 99, 101, 104, 111, 108, 100, 101, 114, 10, 102, 105, 110, 97, 108, 58, 32, 116, 111, 117, 99, 104, 10, 102, 105, 120, 101, 100, 58, 32, 102, 114, 111, 109, 95, 115, 99, 101, 110, 97, 114, 105, 111, 10, 27, 91, 48, 109, 102, 111, 111, 58, 32, 98, 97, 114, 10, 10, 83, 110, 105, 112, 112, 101, 116, 32, 46, 46, 47, 46, 46, 47, 116, 101, 115, 116, 47, 100, 97, 116, 97, 47, 112, 108, 97, 99, 101, 104, 111, 108, 100, 101, 114, 95, 111, 112, 115, 102, 105, 108, 101, 46, 121, 109, 108, 59, 32, 65, 114, 103, 32, 91, 45, 118, 32, 112, 97, 116, 104, 50, 61, 47, 115, 101, 116, 63, 32, 45, 118, 32, 118, 97, 108, 117, 101, 50, 61, 98, 121, 95, 102, 105, 114, 115, 116, 32, 45, 118, 32, 112, 97, 116, 104, 49, 61, 47, 102, 105, 120, 101, 100, 63, 32, 45, 118, 32, 118, 97, 108, 117, 101, 49, 61, 102, 114, 111, 109, 95, 115, 99, 101, 110, 97, 114, 105, 111, 93, 59, 32, 71, 108, 111, 98, 97, 108, 32, 91, 45, 118, 32, 112, 97, 116, 104, 51, 61, 47, 102, 105, 110, 97, 108, 63, 32, 45, 118, 32, 118, 97, 108, 117, 101, 51, 61, 116, 111, 117, 99, 104, 32, 45, 108, 32, 46, 46, 47, 46, 46, 47, 116, 101, 115, 116, 47, 100, 97, 116, 97, 47, 118, 97, 114, 115, 46, 121, 109, 108, 93, 10, 68, 105, 102, 102, 58, 10, 98, 97, 115, 101, 50, 58, 32, 98, 97, 115, 105, 99, 95, 102, 114, 111, 109, 95, 112, 108, 97, 99, 101, 104, 111, 108, 100, 101, 114, 10, 102, 105, 110, 97, 108, 58, 32, 116, 111, 117, 99, 104, 10, 102, 105, 120, 101, 100, 58, 32, 102, 114, 111, 109, 95, 115, 99, 101, 110, 97, 114, 105, 111, 10, 102, 111, 111, 58, 32, 98, 97, 114, 10, 27, 91, 51, 50, 109, 115, 101, 116, 58, 32, 98, 121, 95, 102, 105, 114, 115, 116, 10, 27, 91, 48, 109, 10, 83, 110, 105, 112, 112, 101, 116, 32, 46, 46, 47, 46, 46, 47, 116, 101, 115, 116, 47, 100, 97, 116, 97, 47, 112, 108, 97, 99, 101, 104, 111, 108, 100, 101, 114, 95, 111, 112, 115, 102, 105, 108, 101, 46, 121, 109, 108, 59, 32, 65, 114, 103, 32, 91, 45, 118, 32, 112, 97, 116, 104, 50, 61, 47, 114, 101, 117, 115, 101, 100, 63, 32, 45, 118, 32, 118, 97, 108, 117, 101, 50, 61, 98, 121, 95, 115, 101, 99, 111, 110, 100, 32, 45, 118, 32, 112, 97, 116, 104, 49, 61, 47, 102, 105, 120, 101, 100, 63, 32, 45, 118, 32, 118, 97, 108, 117, 101, 49, 61, 102, 114, 111, 109, 95, 115, 99, 101, 110, 97, 114, 105, 111, 93, 59, 32, 71, 108, 111, 98, 97, 108, 32, 91, 45, 118, 32, 112, 97, 116, 104, 51, 61, 47, 102, 105, 110, 97, 108, 63, 32, 45, 118, 32, 118, 97, 108, 117, 101, 51, 61, 116, 111, 117, 99, 104, 32, 45, 108, 32, 46, 46, 47, 46, 46, 47, 116, 101, 115, 116, 47, 100, 97, 116, 97, 47, 118, 97, 114, 115, 46, 121, 109, 108, 93, 10, 68, 105, 102, 102, 58, 10, 98, 97, 115, 101, 50, 58, 32, 98, 97, 115, 105, 99, 95, 102, 114, 111, 109, 95, 112, 108, 97, 99, 101, 104, 111, 108, 100, 101, 114, 10, 102, 105, 110, 97, 108, 58, 32, 116, 111, 117, 99, 104, 10, 102, 105, 120, 101, 100, 58, 32, 102, 114, 111, 109, 95, 115, 99, 101, 110, 97, 114, 105, 111, 10, 102, 111, 111, 58, 32, 98, 97, 114, 10, 27, 91, 51, 50, 109, 114, 101, 117, 115, 101, 100, 58, 32, 98, 121, 95, 115, 101, 99, 111, 110, 100, 10, 27, 91, 48, 109, 115, 101, 116, 58, 32, 98, 121, 95, 102, 105, 114, 115, 116, 10, 10, 83, 110, 105, 112, 112, 101, 116, 32, 59, 32, 65, 114, 103, 32, 91, 93, 59, 32, 71, 108, 111, 98, 97, 108, 32, 91, 45, 118, 32, 112, 97, 116, 104, 51, 61, 47, 102, 105, 110, 97, 108, 63, 32, 45, 118, 32, 118, 97, 108, 117, 101, 51, 61, 116, 111, 117, 99, 104, 32, 45, 108, 32, 46, 46, 47, 46, 46, 47, 116, 101, 115, 116, 47, 100, 97, 116, 97, 47, 118, 97, 114, 115, 46, 121, 109, 108, 93, 10, 68, 105, 102, 102, 58, 10}

	if errWriter.String() != string(expectedStdErr) {
		t.Errorf("Expected Stderr:\n'''%v'''\nActual:\n'''%v'''\n", string(expectedStdErr), errWriter.String())
	}
}

func TestListPlain(t *testing.T) {
	cmd := exec.Command(
		"go",
		"run",
		"manifer.go",
		"list",
		"-l",
		"../../test/data/library.yml",
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

	if writer.String() != expected {
		t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, writer.String())
	}
}

func TestListJson(t *testing.T) {
	cmd := exec.Command(
		"go",
		"run",
		"manifer.go",
		"list",
		"-l",
		"../../test/data/library.yml",
		"-j",
	)
	writer := &test.StringWriter{}
	cmd.Stdout = writer

	err := cmd.Run()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := `[{"Name":"bizz","Description":"adds an op"},{"Name":"empty","Description":"contributes nothing"},{"Name":"placeholder","Description":"replaces placeholder values"},{"Name":"basic","Description":"a starting point"}]`

	if writer.String() != expected {
		t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, writer.String())
	}
}

func TestSearchPlain(t *testing.T) {
	cmd := exec.Command(
		"go",
		"run",
		"manifer.go",
		"search",
		"-l",
		"../../test/data/library.yml",
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

	if writer.String() != expected {
		t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, writer.String())
	}
}

func TestSearchJson(t *testing.T) {
	cmd := exec.Command(
		"go",
		"run",
		"manifer.go",
		"search",
		"-l",
		"../../test/data/library.yml",
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

	if writer.String() != expected {
		t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, writer.String())
	}
}

func TestInspect(t *testing.T) {
	t.Run("Plain Tree", func(t *testing.T) {
		cmd := exec.Command(
			"go",
			"run",
			"manifer.go",
			"inspect",
			"-l",
			"../../test/data/ref_library.yml",
			"--tree",
			"-s",
			"meta",
		)
		writer := &test.StringWriter{}
		cmd.Stdout = writer

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := `name:        meta (from ../../test/data/ref_library.yml)
description: 
global:  [] (applied to all scenarios)
refargs: [] (applied to snippets and subscenarios)
args:    [] (applied to snippets and subscenarios)
snippets:
dependencies:
  name:        base (from ../../test/data/base_library.yml)
  description: 
  global:  [] (applied to all scenarios)
  refargs: [] (applied to snippets and subscenarios)
  args:    [] (applied to snippets and subscenarios)
  snippets:
    ../../test/data/placeholder_opsfile.yml
    args: [-v path1=/base1? -v value1=a -v path2=/base2? -v value2=b -v path3=/base3? -v value3=c]

  dependencies:

  name:        placeholder (from ../../test/data/library.yml)
  description: replaces placeholder values
  global:  [] (applied to all scenarios)
  refargs: [] (applied to snippets and subscenarios)
  args:    [-v path1=/fixed? -v value1=from_scenario] (applied to snippets and subscenarios)
  snippets:
    ../../test/data/placeholder_opsfile.yml
    args: [-v path2=/set? -v value2=by_first]

    ../../test/data/placeholder_opsfile.yml
    args: [-v path2=/reused? -v value2=by_second]

  dependencies:
    name:        basic (from ../../test/data/library.yml)
    description: a starting point
    global:  [] (applied to all scenarios)
    refargs: [-v value2=basic_from_placeholder] (applied to snippets and subscenarios)
    args:    [-v path1=/base1? -v value1=from_basic] (applied to snippets and subscenarios)
    snippets:
      ../../test/data/placeholder_opsfile.yml
      args: [-v path2=/base2? -v path3=/base3?]

    dependencies:


`
		diffBytes(writer.String(), expected)
		if writer.String() != expected {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, writer.String())
		}
	})

	t.Run("Plain Plan", func(t *testing.T) {
		cmd := exec.Command(
			"go",
			"run",
			"manifer.go",
			"inspect",
			"-l",
			"../../test/data/ref_library.yml",
			"--plan",
			"-s",
			"meta",
		)
		writer := &test.StringWriter{}
		cmd.Stdout = writer

		err := cmd.Run()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := `global: []
- ../../test/data/placeholder_opsfile.yml
  args:
    snippet: [-v path1=/base1? -v value1=a -v path2=/base2? -v value2=b -v path3=/base3? -v value3=c]
    base: []
    meta: []
- ../../test/data/placeholder_opsfile.yml
  args:
    snippet: [-v path2=/base2? -v path3=/base3?]
    basic: [-v path1=/base1? -v value1=from_basic -v value2=basic_from_placeholder]
    placeholder: [-v path1=/fixed? -v value1=from_scenario]
    meta: []
- ../../test/data/placeholder_opsfile.yml
  args:
    snippet: [-v path2=/set? -v value2=by_first]
    placeholder: [-v path1=/fixed? -v value1=from_scenario]
    meta: []
- ../../test/data/placeholder_opsfile.yml
  args:
    snippet: [-v path2=/reused? -v value2=by_second]
    placeholder: [-v path1=/fixed? -v value1=from_scenario]
    meta: []
`

		if writer.String() != expected {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, writer.String())
		}
	})

}

func diffBytes(str1 string, str2 string) {
	b1 := []byte(str1)
	b2 := []byte(str2)
	if len(b1) != len(b2) {
		fmt.Printf("Different lengths: %d %d\n", len(b1), len(b2))
	}
	for i, b := range b1 {
		if i < len(b2) {
			if b != b2[i] {
				fmt.Printf("byte %d differs: %d, %d\n", i, b, b2[i])
			}
		}
	}
}
