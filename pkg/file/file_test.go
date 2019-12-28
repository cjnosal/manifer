package file

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRelativeTo(t *testing.T) {
	t.Run("absolute path", func(t *testing.T) {
		subject := &FileIO{}
		expected := "/tmp/foo.yml"
		actual, err := subject.ResolveRelativeTo("/tmp/foo.yml", "../../test/data")
		if err != nil {
			t.Errorf(err.Error())
		}
		if expected != actual {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, actual)
		}
	})
	t.Run("relative to file", func(t *testing.T) {
		subject := &FileIO{}
		expected := "../../test/data/v2/opsfile.yml"
		actual, err := subject.ResolveRelativeTo("./opsfile.yml", "../../test/data/v2/library.yml")
		if err != nil {
			t.Errorf(err.Error())
		}
		if expected != actual {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, actual)
		}
	})
	t.Run("relative to directory", func(t *testing.T) {
		subject := &FileIO{}
		expected := "../../test/data/v2/opsfile.yml"
		actual, err := subject.ResolveRelativeTo("./opsfile.yml", "../../test/data/v2")
		if err != nil {
			t.Errorf(err.Error())
		}
		if expected != actual {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, actual)
		}
	})
}

func TestResolveRelativeFrom(t *testing.T) {
	t.Run("absolute path", func(t *testing.T) {
		subject := &FileIO{}
		wd, _ := os.Getwd()
		absWd, _ := filepath.Abs(wd)
		expected := "../../pkg/file/opsfile.yml"
		actual, err := subject.ResolveRelativeFrom(filepath.Join(absWd, "opsfile.yml"), "../../test/data")
		if err != nil {
			t.Errorf(err.Error())
		}
		if expected != filepath.Clean(actual) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, actual)
		}
	})
	t.Run("relative to file", func(t *testing.T) {
		subject := &FileIO{}
		expected := "../../../pkg/file/opsfile.yml"
		actual, err := subject.ResolveRelativeFrom("./opsfile.yml", "../../test/data/v2/library.yml")
		if err != nil {
			t.Errorf(err.Error())
		}
		if expected != actual {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, actual)
		}
	})
	t.Run("relative to directory", func(t *testing.T) {
		subject := &FileIO{}
		expected := "../../pkg/file/opsfile.yml"
		actual, err := subject.ResolveRelativeFrom("./opsfile.yml", "../../test/data")
		if err != nil {
			t.Errorf(err.Error())
		}
		if expected != actual {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, actual)
		}
	})
}
