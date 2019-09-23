package file

import (
	"testing"
)

func TestResolveRelativeTo(t *testing.T) {
	t.Run("absolute path", func(t *testing.T) {
		subject := &FileIO{}
		expected := "/tmp/foo.yml"
		actual, err := subject.ResolveRelativeTo("/tmp/foo.yml", "./bar/other.yml")
		if err != nil {
			t.Errorf(err.Error())
		}
		if expected != actual {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, actual)
		}
	})
	t.Run("relative path", func(t *testing.T) {
		subject := &FileIO{}
		expected := "../../test/data/opsfile.yml"
		actual, err := subject.ResolveRelativeTo("./opsfile.yml", "../../test/data/library.yml")
		if err != nil {
			t.Errorf(err.Error())
		}
		if expected != actual {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, actual)
		}
	})
}
