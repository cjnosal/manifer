package main

import (
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
		"--",
		"-v",
		"path3=/final?",
		"-v",
		"value3=touch",
		"-l",
		"../../test/data/vars.yml",
	)
	writer := &test.StringWriter{}
	cmd.Stdout = writer

	err := cmd.Run()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := `base2: basic_from_placeholder
final: touch
fixed: from_scenario
foo: bar
reused: by_second
set: by_first
`

	if writer.String() != expected {
		t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, writer.String())
	}
}

func TestList(t *testing.T) {
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
	no description

empty
	no description

placeholder
	no description

basic
	no description

`

	if writer.String() != expected {
		t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, writer.String())
	}
}
