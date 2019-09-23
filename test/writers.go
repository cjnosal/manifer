package test

import (
	"bytes"
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

type ByteWriter struct {
	buffer bytes.Buffer
}

func (s *ByteWriter) Write(b []byte) (int, error) {
	s.buffer.Write(b)
	return len(b), nil
}

func (s *ByteWriter) Bytes() []byte {
	return s.buffer.Bytes()
}
