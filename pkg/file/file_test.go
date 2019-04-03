package file

import (
	"testing"
)

func TestResolveRelativeTo(t *testing.T) {
	t.Run("absolute path", func(t *testing.T) {
		subject := &FileIO{}
		expected := "/tmp/foo.yml"
		actual := subject.ResolveRelativeTo("/tmp/foo.yml", "./bar/other.yml")
		if expected != actual {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, actual)
		}
	})
	t.Run("relative path", func(t *testing.T) {
		subject := &FileIO{}
		expected := "bar/foo.yml"
		actual := subject.ResolveRelativeTo("./foo.yml", "./bar/other.yml")
		if expected != actual {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, actual)
		}
	})
}
