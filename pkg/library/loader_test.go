package library

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/cjnosal/manifer/v2/pkg/file"
	"github.com/cjnosal/manifer/v2/pkg/yaml"
	"github.com/cjnosal/manifer/v2/test"
	"github.com/google/go-cmp/cmp"
)

func TestLoad(t *testing.T) {
	t.Run("single library", func(t *testing.T) {
		lib1 := Library{
			Type: OpsFile,
			Scenarios: []Scenario{
				{
					Name: "s",
					Snippets: []Snippet{
						{
							Path: "./snippet.yml",
						},
					},
				},
			},
		}
		loadedlib1 := &Library{
			Type: OpsFile,
			Scenarios: []Scenario{
				{
					Name: "s",
					Snippets: []Snippet{
						{
							Path: "/wd/lib/snippet.yml",
						},
					},
				},
			},
		}
		expectedLoadedLibs := LoadedLibrary{
			TopLibraries: []*Library{
				loadedlib1,
			},
			Libraries: map[string]*Library{
				"/wd/lib/library.yml": loadedlib1,
			},
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFile := file.NewMockFileAccess(ctrl)
		mockYaml := yaml.NewMockYamlAccess(ctrl)
		subject := &Loader{
			File: mockFile,
			Yaml: mockYaml,
		}

		mockFile.EXPECT().GetWorkingDirectory().Times(1).Return("/wd", nil)
		mockFile.EXPECT().ResolveRelativeTo("./lib/library.yml", "/wd").Times(1).Return("/wd/lib/library.yml", nil)
		mockFile.EXPECT().Read("/wd/lib/library.yml").Times(1).Return([]byte("bytes"), nil)
		mockYaml.EXPECT().Unmarshal([]byte("bytes"), &Library{}).Times(1).Return(nil).Do(func(bytes []byte, lib *Library) {
			*lib = lib1
		})
		mockFile.EXPECT().ResolveRelativeTo("./snippet.yml", "/wd/lib/library.yml").Times(1).Return("/wd/lib/snippet.yml", nil)

		loadedLibs, err := subject.Load([]string{"./lib/library.yml"})

		if err != nil {
			t.Errorf("Unexpected error: %v\n", err)
		}

		if !cmp.Equal(expectedLoadedLibs, *loadedLibs) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedLoadedLibs, *loadedLibs)
		}
	})

	t.Run("two library", func(t *testing.T) {
		lib1 := Library{
			Type: OpsFile,
			Scenarios: []Scenario{
				{
					Name: "s",
					Snippets: []Snippet{
						{
							Path: "./snippet.yml",
						},
					},
				},
			},
		}
		loadedlib1 := &Library{
			Type: OpsFile,
			Scenarios: []Scenario{
				{
					Name: "s",
					Snippets: []Snippet{
						{
							Path: "/wd/lib/snippet.yml",
						},
					},
				},
			},
		}
		lib2 := Library{
			Type: OpsFile,
			Scenarios: []Scenario{
				{
					Name: "t",
					Snippets: []Snippet{
						{
							Path: "./snippet2.yml",
						},
					},
				},
			},
		}
		loadedlib2 := &Library{
			Type: OpsFile,
			Scenarios: []Scenario{
				{
					Name: "t",
					Snippets: []Snippet{
						{
							Path: "/wd/lib2/snippet2.yml",
						},
					},
				},
			},
		}
		expectedLoadedLibs := LoadedLibrary{
			TopLibraries: []*Library{
				loadedlib1,
				loadedlib2,
			},
			Libraries: map[string]*Library{
				"/wd/lib/library.yml":   loadedlib1,
				"/wd/lib2/library2.yml": loadedlib2,
			},
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFile := file.NewMockFileAccess(ctrl)
		mockYaml := yaml.NewMockYamlAccess(ctrl)
		subject := &Loader{
			File: mockFile,
			Yaml: mockYaml,
		}

		mockFile.EXPECT().GetWorkingDirectory().Times(1).Return("/wd", nil)

		mockFile.EXPECT().ResolveRelativeTo("./lib/library.yml", "/wd").Times(1).Return("/wd/lib/library.yml", nil)
		mockFile.EXPECT().Read("/wd/lib/library.yml").Times(1).Return([]byte("bytes"), nil)
		mockYaml.EXPECT().Unmarshal([]byte("bytes"), &Library{}).Times(1).Return(nil).Do(func(bytes []byte, lib *Library) {
			*lib = lib1
		})
		mockFile.EXPECT().ResolveRelativeTo("./snippet.yml", "/wd/lib/library.yml").Times(1).Return("/wd/lib/snippet.yml", nil)

		mockFile.EXPECT().ResolveRelativeTo("./lib2/library2.yml", "/wd").Times(1).Return("/wd/lib2/library2.yml", nil)
		mockFile.EXPECT().Read("/wd/lib2/library2.yml").Times(1).Return([]byte("bytes2"), nil)
		mockYaml.EXPECT().Unmarshal([]byte("bytes2"), &Library{}).Times(1).Return(nil).Do(func(bytes []byte, lib *Library) {
			*lib = lib2
		})
		mockFile.EXPECT().ResolveRelativeTo("./snippet2.yml", "/wd/lib2/library2.yml").Times(1).Return("/wd/lib2/snippet2.yml", nil)

		loadedLibs, err := subject.Load([]string{"./lib/library.yml", "./lib2/library2.yml"})

		if err != nil {
			t.Errorf("Unexpected error: %v\n", err)
		}

		if !cmp.Equal(expectedLoadedLibs, *loadedLibs) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedLoadedLibs, *loadedLibs)
		}
	})

	t.Run("referenced library", func(t *testing.T) {
		lib1 := Library{
			Type: OpsFile,
			Libraries: []LibraryRef{
				{
					Alias: "foo",
					Path:  "../lib2/library2.yml",
				},
			},
			Scenarios: []Scenario{
				{
					Name: "s",
					Snippets: []Snippet{
						{
							Path: "./snippet.yml",
						},
					},
				},
			},
		}
		loadedlib1 := &Library{
			Type: OpsFile,
			Libraries: []LibraryRef{
				{
					Alias: "foo",
					Path:  "/wd/lib2/library2.yml",
				},
			},
			Scenarios: []Scenario{
				{
					Name: "s",
					Snippets: []Snippet{
						{
							Path: "/wd/lib/snippet.yml",
						},
					},
				},
			},
		}
		lib2 := Library{
			Type: OpsFile,
			Scenarios: []Scenario{
				{
					Name: "t",
					Snippets: []Snippet{
						{
							Path: "./snippet2.yml",
						},
					},
				},
			},
		}
		loadedlib2 := &Library{
			Type: OpsFile,
			Scenarios: []Scenario{
				{
					Name: "t",
					Snippets: []Snippet{
						{
							Path: "/wd/lib2/snippet2.yml",
						},
					},
				},
			},
		}
		expectedLoadedLibs := LoadedLibrary{
			TopLibraries: []*Library{
				loadedlib1,
			},
			Libraries: map[string]*Library{
				"/wd/lib/library.yml":   loadedlib1,
				"/wd/lib2/library2.yml": loadedlib2,
			},
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFile := file.NewMockFileAccess(ctrl)
		mockYaml := yaml.NewMockYamlAccess(ctrl)
		subject := &Loader{
			File: mockFile,
			Yaml: mockYaml,
		}

		mockFile.EXPECT().GetWorkingDirectory().Times(1).Return("/wd", nil)

		mockFile.EXPECT().ResolveRelativeTo("./lib/library.yml", "/wd").Times(1).Return("/wd/lib/library.yml", nil)
		mockFile.EXPECT().Read("/wd/lib/library.yml").Times(1).Return([]byte("bytes"), nil)
		mockYaml.EXPECT().Unmarshal([]byte("bytes"), &Library{}).Times(1).Return(nil).Do(func(bytes []byte, lib *Library) {
			*lib = lib1
		})
		mockFile.EXPECT().ResolveRelativeTo("./snippet.yml", "/wd/lib/library.yml").Times(1).Return("/wd/lib/snippet.yml", nil)

		mockFile.EXPECT().ResolveRelativeTo("../lib2/library2.yml", "/wd/lib/library.yml").Times(1).Return("/wd/lib2/library2.yml", nil)
		mockFile.EXPECT().Read("/wd/lib2/library2.yml").Times(1).Return([]byte("bytes2"), nil)
		mockYaml.EXPECT().Unmarshal([]byte("bytes2"), &Library{}).Times(1).Return(nil).Do(func(bytes []byte, lib *Library) {
			*lib = lib2
		})
		mockFile.EXPECT().ResolveRelativeTo("./snippet2.yml", "/wd/lib2/library2.yml").Times(1).Return("/wd/lib2/snippet2.yml", nil)

		loadedLibs, err := subject.Load([]string{"./lib/library.yml"})

		if err != nil {
			t.Errorf("Unexpected error: %v\n", err)
		}

		if !cmp.Equal(expectedLoadedLibs, *loadedLibs) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expectedLoadedLibs, *loadedLibs)
		}
	})

	t.Run("referenced library error", func(t *testing.T) {
		lib1 := Library{
			Type: OpsFile,
			Libraries: []LibraryRef{
				{
					Alias: "foo",
					Path:  "../lib2/library2.yml",
				},
			},
			Scenarios: []Scenario{
				{
					Name: "s",
					Snippets: []Snippet{
						{
							Path: "./snippet.yml",
						},
					},
				},
			},
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFile := file.NewMockFileAccess(ctrl)
		mockYaml := yaml.NewMockYamlAccess(ctrl)
		subject := &Loader{
			File: mockFile,
			Yaml: mockYaml,
		}

		mockFile.EXPECT().GetWorkingDirectory().Times(1).Return("/wd", nil)

		mockFile.EXPECT().ResolveRelativeTo("./lib/library.yml", "/wd").Times(1).Return("/wd/lib/library.yml", nil)
		mockFile.EXPECT().Read("/wd/lib/library.yml").Times(1).Return([]byte("bytes"), nil)
		mockYaml.EXPECT().Unmarshal([]byte("bytes"), &Library{}).Times(1).Return(nil).Do(func(bytes []byte, lib *Library) {
			*lib = lib1
		})
		mockFile.EXPECT().ResolveRelativeTo("./snippet.yml", "/wd/lib/library.yml").Times(1).Return("/wd/lib/snippet.yml", nil)

		mockFile.EXPECT().ResolveRelativeTo("../lib2/library2.yml", "/wd/lib/library.yml").Times(1).Return("", errors.New("test"))

		_, err := subject.Load([]string{"./lib/library.yml"})

		expectedError := errors.New(`test
  while resolving library path ../lib2/library2.yml from /wd/lib/library.yml
  while loading library from path ./lib/library.yml`)
		if !cmp.Equal(&err, &expectedError, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%v'''\n", expectedError, err)
		}
	})

	t.Run("resolve snippet error", func(t *testing.T) {
		lib1 := Library{
			Type: OpsFile,
			Scenarios: []Scenario{
				{
					Name: "s",
					Snippets: []Snippet{
						{
							Path: "./snippet.yml",
						},
					},
				},
			},
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFile := file.NewMockFileAccess(ctrl)
		mockYaml := yaml.NewMockYamlAccess(ctrl)
		subject := &Loader{
			File: mockFile,
			Yaml: mockYaml,
		}

		mockFile.EXPECT().GetWorkingDirectory().Times(1).Return("/wd", nil)

		mockFile.EXPECT().ResolveRelativeTo("./lib/library.yml", "/wd").Times(1).Return("/wd/lib/library.yml", nil)
		mockFile.EXPECT().Read("/wd/lib/library.yml").Times(1).Return([]byte("bytes"), nil)
		mockYaml.EXPECT().Unmarshal([]byte("bytes"), &Library{}).Times(1).Return(nil).Do(func(bytes []byte, lib *Library) {
			*lib = lib1
		})
		mockFile.EXPECT().ResolveRelativeTo("./snippet.yml", "/wd/lib/library.yml").Times(1).Return("/wd/lib/snippet.yml", errors.New("test"))

		_, err := subject.Load([]string{"./lib/library.yml"})

		expectedError := errors.New(`test
  while resolving snippet path ./snippet.yml from /wd/lib/library.yml
  while loading library from path ./lib/library.yml`)
		if !cmp.Equal(&err, &expectedError, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%v'''\n", expectedError, err)
		}
	})

	t.Run("unmarshal library error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFile := file.NewMockFileAccess(ctrl)
		mockYaml := yaml.NewMockYamlAccess(ctrl)
		subject := &Loader{
			File: mockFile,
			Yaml: mockYaml,
		}

		mockFile.EXPECT().GetWorkingDirectory().Times(1).Return("/wd", nil)

		mockFile.EXPECT().ResolveRelativeTo("./lib/library.yml", "/wd").Times(1).Return("/wd/lib/library.yml", nil)
		mockFile.EXPECT().Read("/wd/lib/library.yml").Times(1).Return([]byte("bytes"), nil)
		mockYaml.EXPECT().Unmarshal([]byte("bytes"), &Library{}).Times(1).Return(errors.New("test"))

		_, err := subject.Load([]string{"./lib/library.yml"})

		expectedError := errors.New(`test
  while parsing library at /wd/lib/library.yml
  while loading library from path ./lib/library.yml`)
		if !cmp.Equal(&err, &expectedError, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%v'''\n", expectedError, err)
		}
	})

	t.Run("read library error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFile := file.NewMockFileAccess(ctrl)
		mockYaml := yaml.NewMockYamlAccess(ctrl)
		subject := &Loader{
			File: mockFile,
			Yaml: mockYaml,
		}

		mockFile.EXPECT().GetWorkingDirectory().Times(1).Return("/wd", nil)

		mockFile.EXPECT().ResolveRelativeTo("./lib/library.yml", "/wd").Times(1).Return("/wd/lib/library.yml", nil)
		mockFile.EXPECT().Read("/wd/lib/library.yml").Times(1).Return(nil, errors.New("test"))

		_, err := subject.Load([]string{"./lib/library.yml"})

		expectedError := errors.New(`test
  while reading library at /wd/lib/library.yml
  while loading library from path ./lib/library.yml`)
		if !cmp.Equal(&err, &expectedError, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%v'''\n", expectedError, err)
		}
	})

	t.Run("resolve library error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFile := file.NewMockFileAccess(ctrl)
		mockYaml := yaml.NewMockYamlAccess(ctrl)
		subject := &Loader{
			File: mockFile,
			Yaml: mockYaml,
		}

		mockFile.EXPECT().GetWorkingDirectory().Times(1).Return("/wd", nil)

		mockFile.EXPECT().ResolveRelativeTo("./lib/library.yml", "/wd").Times(1).Return("/wd/lib/library.yml", errors.New("test"))

		_, err := subject.Load([]string{"./lib/library.yml"})

		expectedError := errors.New(`test
  while resolving library path ./lib/library.yml from /wd`)
		if !cmp.Equal(&err, &expectedError, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%v'''\n", expectedError, err)
		}
	})

	t.Run("working directory error", func(t *testing.T) {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFile := file.NewMockFileAccess(ctrl)
		mockYaml := yaml.NewMockYamlAccess(ctrl)
		subject := &Loader{
			File: mockFile,
			Yaml: mockYaml,
		}

		mockFile.EXPECT().GetWorkingDirectory().Times(1).Return("", errors.New("test"))

		_, err := subject.Load([]string{"./lib/library.yml"})
		expectedError := errors.New("test\n  while finding working directory")
		if !cmp.Equal(&err, &expectedError, cmp.Comparer(test.EqualMessage)) {
			t.Errorf("Expected:\n'''%s'''\nActual:\n'''%v'''\n", expectedError, err)
		}
	})
}
