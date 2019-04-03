package test

import (
	"errors"
	"strings"
)

type StringWriter struct {
	builder strings.Builder
}

func (s *StringWriter) Write(b []byte) (int, error) {
	s.builder.WriteString(string(b))
	return len(b), nil
}

func (s *StringWriter) String() string {
	return s.builder.String()
}

type BrokenWriter struct {
}

func (s *BrokenWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("broken writer failed")
}
