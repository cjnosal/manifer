package importer

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cjnosal/manifer/test"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/processor"
	"github.com/cjnosal/manifer/pkg/processor/factory"
)

type TestFileInfo struct {
	dir  bool
	name string
}

func (t *TestFileInfo) Name() string       { return t.name }
func (t *TestFileInfo) Size() int64        { return 0 }
func (t *TestFileInfo) Mode() os.FileMode  { return 0000 }
func (t *TestFileInfo) ModTime() time.Time { return time.Now() }
func (t *TestFileInfo) IsDir() bool        { return t.dir }
func (t *TestFileInfo) Sys() interface{}   { return nil }

func TestImport(t *testing.T) {
	t.Run("check dir error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProcessor := processor.NewMockProcessor(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		subject := NewImporter(mockFile, mockProcessorFactory)

		mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockProcessor, nil)
		mockFile.EXPECT().IsDir("/in").Times(1).Return(false, errors.New("oops"))

		expectedErr := errors.New("oops\n  checking import path /in")
		_, err := subject.Import(library.OpsFile, "/in", true, "/out")

		if !cmp.Equal(&expectedErr, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'%v'\nActual:\n'%v'\n", expectedErr, err)
		}
	})

	t.Run("validate single file error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProcessor := processor.NewMockProcessor(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		subject := NewImporter(mockFile, mockProcessorFactory)

		mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockProcessor, nil)
		mockFile.EXPECT().IsDir("/in").Times(1).Return(false, nil)
		mockProcessor.EXPECT().ValidateSnippet("/in").Times(1).Return(false, errors.New("oops"))

		expectedErr := errors.New("oops\n  validating file /in\n  importing file /in")
		_, err := subject.Import(library.OpsFile, "/in", true, "/dir/out")

		if !cmp.Equal(&expectedErr, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'%v'\nActual:\n'%v'\n", expectedErr, err)
		}
	})

	t.Run("invalid single file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProcessor := processor.NewMockProcessor(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		subject := NewImporter(mockFile, mockProcessorFactory)

		mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockProcessor, nil)
		mockFile.EXPECT().IsDir("/in").Times(1).Return(false, nil)
		mockProcessor.EXPECT().ValidateSnippet("/in").Times(1).Return(false, nil)

		expectedLib := &library.Library{
			Scenarios: []library.Scenario{},
		}
		lib, err := subject.Import(library.OpsFile, "/in", true, "/dir/out")

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		if !cmp.Equal(expectedLib, lib) {
			t.Errorf("Expected:\n'%v'\nActual:\n'%v'\n", expectedLib, lib)
		}
	})

	t.Run("resolve file path error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProcessor := processor.NewMockProcessor(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		subject := NewImporter(mockFile, mockProcessorFactory)

		mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockProcessor, nil)
		mockFile.EXPECT().IsDir("/in").Times(1).Return(false, nil)
		mockProcessor.EXPECT().ValidateSnippet("/in").Times(1).Return(true, nil)
		mockFile.EXPECT().ResolveRelativeFrom("/in", "/dir").Times(1).Return("", errors.New("oops"))

		expectedErr := errors.New("oops\n  resolving relative path from /dir\n  importing file /in")
		_, err := subject.Import(library.OpsFile, "/in", true, "/dir/out")

		if !cmp.Equal(&expectedErr, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'%v'\nActual:\n'%v'\n", expectedErr, err)
		}
	})

	t.Run("import file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProcessor := processor.NewMockProcessor(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		subject := NewImporter(mockFile, mockProcessorFactory)

		mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockProcessor, nil)
		mockFile.EXPECT().IsDir("/in").Times(1).Return(false, nil)
		mockProcessor.EXPECT().ValidateSnippet("/in").Times(1).Return(true, nil)
		mockFile.EXPECT().ResolveRelativeFrom("/in", "/dir").Times(1).Return("../in", nil)

		expectedLib := &library.Library{
			Scenarios: []library.Scenario{
				{
					Name:        "in",
					Description: "imported from in",
					Snippets: []library.Snippet{
						library.Snippet{
							Path: "../in",
							Processor: library.Processor{
								Type: library.OpsFile,
							},
						},
					},
				},
			},
		}
		lib, err := subject.Import(library.OpsFile, "/in", true, "/dir/out")

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		if !cmp.Equal(expectedLib, lib) {
			t.Errorf("Expected:\n'%v'\nActual:\n'%v'\n", expectedLib, lib)
		}
	})

	t.Run("walk dir error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProcessor := processor.NewMockProcessor(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		subject := NewImporter(mockFile, mockProcessorFactory)

		mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockProcessor, nil)
		mockFile.EXPECT().IsDir("/in").Times(1).Return(true, nil)
		mockFile.EXPECT().Walk("/in", gomock.Any()).Times(1).Return(errors.New("oops"))

		expectedErr := errors.New("oops\n  walking directory /in\n  importing directory /in")
		_, err := subject.Import(library.OpsFile, "/in", true, "/dir/out")

		if !cmp.Equal(&expectedErr, &err, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'%v'\nActual:\n'%v'\n", expectedErr, err)
		}
	})

	t.Run("walk file error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProcessor := processor.NewMockProcessor(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		subject := NewImporter(mockFile, mockProcessorFactory)

		mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockProcessor, nil)
		mockFile.EXPECT().IsDir("/in").Times(1).Return(true, nil)
		mockFile.EXPECT().Walk("/in", gomock.Any()).Times(1).Do(func(path string, callback func(path string, info os.FileInfo, err error) error) error {
			err := callback("f", nil, errors.New("oops"))
			expectedErr := errors.New("oops\n  walking to f")
			if !cmp.Equal(&expectedErr, &err, cmp.Comparer(test.EqualMessage)) {
				t.Errorf("Expected:\n'%v'\nActual:\n'%v'\n", expectedErr, err)
			}
			return err
		})

		subject.Import(library.OpsFile, "/in", true, "/dir/out")
	})

	t.Run("non-recursive skips dir", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProcessor := processor.NewMockProcessor(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		subject := NewImporter(mockFile, mockProcessorFactory)

		mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockProcessor, nil)
		mockFile.EXPECT().IsDir("/in").Times(1).Return(true, nil)
		mockFile.EXPECT().Walk("/in", gomock.Any()).Times(1).Do(func(path string, callback func(path string, info os.FileInfo, err error) error) error {
			err := callback("f", &TestFileInfo{dir: true}, nil)
			expectedErr := filepath.SkipDir
			if !cmp.Equal(&expectedErr, &err, cmp.Comparer(test.EqualMessage)) {
				t.Errorf("Expected:\n'%v'\nActual:\n'%v'\n", expectedErr, err)
			}
			return err
		})

		subject.Import(library.OpsFile, "/in", false, "/dir/out")
	})

	t.Run("non-recursive skips dir", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProcessor := processor.NewMockProcessor(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		subject := NewImporter(mockFile, mockProcessorFactory)

		mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockProcessor, nil)
		mockFile.EXPECT().IsDir("/in").Times(1).Return(true, nil)
		mockFile.EXPECT().Walk("/in", gomock.Any()).Times(1).Do(func(path string, callback func(path string, info os.FileInfo, err error) error) error {
			err := callback("f", &TestFileInfo{dir: true}, nil)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}
			return err
		})

		subject.Import(library.OpsFile, "/in", true, "/dir/out")
	})

	t.Run("validate file in dir error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProcessor := processor.NewMockProcessor(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		subject := NewImporter(mockFile, mockProcessorFactory)

		mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockProcessor, nil)
		mockFile.EXPECT().IsDir("/in").Times(1).Return(true, nil)
		mockProcessor.EXPECT().ValidateSnippet("f").Times(1).Return(false, errors.New("oops"))
		mockFile.EXPECT().Walk("/in", gomock.Any()).Times(1).Do(func(path string, callback func(path string, info os.FileInfo, err error) error) error {
			err := callback("f", &TestFileInfo{dir: false}, nil)
			expectedErr := errors.New("oops\n  validating file f\n  importing file f")
			if !cmp.Equal(&expectedErr, &err, cmp.Comparer(test.EqualMessage)) {
				t.Errorf("Expected:\n'%v'\nActual:\n'%v'\n", expectedErr, err)
			}
			return err
		})

		subject.Import(library.OpsFile, "/in", false, "/dir/out")
	})

	t.Run("resolve file path in dir error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProcessor := processor.NewMockProcessor(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		subject := NewImporter(mockFile, mockProcessorFactory)

		mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockProcessor, nil)
		mockFile.EXPECT().IsDir("/in").Times(1).Return(true, nil)
		mockProcessor.EXPECT().ValidateSnippet("f").Times(1).Return(true, nil)
		mockFile.EXPECT().ResolveRelativeFrom("f", "/dir").Times(1).Return("", errors.New("oops"))

		mockFile.EXPECT().Walk("/in", gomock.Any()).Times(1).Do(func(path string, callback func(path string, info os.FileInfo, err error) error) error {
			err := callback("f", &TestFileInfo{dir: false}, nil)
			expectedErr := errors.New("oops\n  resolving relative path from /dir\n  importing file f")
			if !cmp.Equal(&expectedErr, &err, cmp.Comparer(test.EqualMessage)) {
				t.Errorf("Expected:\n'%v'\nActual:\n'%v'\n", expectedErr, err)
			}
			return err
		})

		subject.Import(library.OpsFile, "/in", false, "/dir/out")
	})

	t.Run("import opsfiles from directory", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockProcessor := processor.NewMockProcessor(ctrl)
		mockFile := file.NewMockFileAccess(ctrl)
		mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
		subject := NewImporter(mockFile, mockProcessorFactory)

		mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockProcessor, nil)
		mockFile.EXPECT().IsDir("/in").Times(1).Return(true, nil)
		mockProcessor.EXPECT().ValidateSnippet("f").Times(1).Return(true, nil)
		mockProcessor.EXPECT().ValidateSnippet("g").Times(1).Return(false, nil)
		mockProcessor.EXPECT().ValidateSnippet("h").Times(1).Return(true, nil)
		mockFile.EXPECT().ResolveRelativeFrom("f", "/dir").Times(1).Return("../f", nil)
		mockFile.EXPECT().ResolveRelativeFrom("h", "/dir").Times(1).Return("../h", nil)

		mockFile.EXPECT().Walk("/in", gomock.Any()).Times(1).Do(func(path string, callback func(path string, info os.FileInfo, err error) error) error {
			err := callback("f", &TestFileInfo{dir: false}, nil)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}
			err = callback("g", &TestFileInfo{dir: false}, nil)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}
			err = callback("h", &TestFileInfo{dir: false}, nil)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}
			return nil
		})

		expectedLib := &library.Library{
			Scenarios: []library.Scenario{
				{
					Name:        "f",
					Description: "imported from f",
					Snippets: []library.Snippet{
						library.Snippet{
							Path: "../f",
							Processor: library.Processor{
								Type: library.OpsFile,
							},
						},
					},
				},
				{
					Name:        "h",
					Description: "imported from h",
					Snippets: []library.Snippet{
						library.Snippet{
							Path: "../h",
							Processor: library.Processor{
								Type: library.OpsFile,
							},
						},
					},
				},
			},
		}
		lib, err := subject.Import(library.OpsFile, "/in", false, "/dir/out")

		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		if !cmp.Equal(expectedLib, lib) {
			t.Errorf("Expected:\n'%v'\nActual:\n'%v'\n", expectedLib, lib)
		}
	})
}
