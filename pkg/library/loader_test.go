package library

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/yaml"
)

type resolution struct {
	source string
	target string
	result string
}

func TestLoad(t *testing.T) {

	cases := []struct {
		name                string
		paths               []string
		expectedPaths       []string
		expectedResolutions []resolution
		testLibs            []Library
		yamlError           error
		expectedLoadedLibs  []LoadedLibrary
		expectedError       error
	}{
		{
			name: "single library",
			paths: []string{
				"./lib/library.yml",
			},
			expectedPaths: []string{
				"./lib/library.yml",
			},
			testLibs: []Library{
				{
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
				},
			},
			expectedResolutions: []resolution{
				{
					source: "./lib/library.yml",
					target: "./snippet.yml",
					result: "./lib/snippet.yml",
				},
			},
			expectedLoadedLibs: []LoadedLibrary{
				{
					Path: "./lib/library.yml",
					Library: &Library{
						Type: OpsFile,
						Scenarios: []Scenario{
							{
								Name: "s",
								Snippets: []Snippet{
									{
										Path: "./lib/snippet.yml",
									},
								},
							},
						},
					},
					References: map[string]*LoadedLibrary{},
				},
			},
		},
		{
			name: "two libraries",
			paths: []string{
				"./lib/library.yml",
				"./lib2/library2.yml",
			},
			expectedPaths: []string{
				"./lib/library.yml",
				"./lib2/library2.yml",
			},
			testLibs: []Library{
				{
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
				},
				{
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
				},
			},
			expectedResolutions: []resolution{
				{
					source: "./lib/library.yml",
					target: "./snippet.yml",
					result: "./lib/snippet.yml",
				},
				{
					source: "./lib2/library2.yml",
					target: "./snippet2.yml",
					result: "./lib2/snippet2.yml",
				},
			},
			expectedLoadedLibs: []LoadedLibrary{
				{
					Path: "./lib/library.yml",
					Library: &Library{
						Type: OpsFile,
						Scenarios: []Scenario{
							{
								Name: "s",
								Snippets: []Snippet{
									{
										Path: "./lib/snippet.yml",
									},
								},
							},
						},
					},
					References: map[string]*LoadedLibrary{},
				},
				{
					Path: "./lib2/library2.yml",
					Library: &Library{
						Type: OpsFile,
						Scenarios: []Scenario{
							{
								Name: "t",
								Snippets: []Snippet{
									{
										Path: "./lib2/snippet2.yml",
									},
								},
							},
						},
					},
					References: map[string]*LoadedLibrary{},
				},
			},
		},
		{
			name: "referenced libraries",
			paths: []string{
				"./lib/library.yml",
			},
			expectedPaths: []string{
				"./lib/library.yml",
				"./lib/library2.yml",
			},
			expectedResolutions: []resolution{

				{
					source: "./lib/library.yml",
					target: "./snippet.yml",
					result: "./lib/snippet.yml",
				},
				{
					source: "./lib/library.yml",
					target: "./library2.yml",
					result: "./lib/library2.yml",
				},
				{
					source: "./lib/library2.yml",
					target: "./snippet2.yml",
					result: "./lib/snippet2.yml",
				},
			},
			testLibs: []Library{
				{
					Type: OpsFile,
					Libraries: []LibraryRef{
						{
							Alias: "library2",
							Path:  "./library2.yml",
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
				},
				{
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
				},
			},
			expectedLoadedLibs: []LoadedLibrary{
				{
					Path: "./lib/library.yml",
					Library: &Library{
						Type: OpsFile,
						Libraries: []LibraryRef{
							{
								Alias: "library2",
								Path:  "./library2.yml",
							},
						},
						Scenarios: []Scenario{
							{
								Name: "s",
								Snippets: []Snippet{
									{
										Path: "./lib/snippet.yml",
									},
								},
							},
						},
					},
					References: map[string]*LoadedLibrary{
						"library2": &LoadedLibrary{
							Path: "./lib/library2.yml",
							Library: &Library{
								Type: OpsFile,
								Scenarios: []Scenario{
									{
										Name: "t",
										Snippets: []Snippet{
											{
												Path: "./lib/snippet2.yml",
											},
										},
									},
								},
							},
							References: map[string]*LoadedLibrary{},
						},
					},
				},
			},
		},
		{
			name: "yaml error",
			paths: []string{
				"./library.yml",
			},
			expectedPaths: []string{
				"./library.yml",
			},
			yamlError:     errors.New("test"),
			expectedError: errors.New("Unable to load library at ./library.yml: test"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFile := file.NewMockFileAccess(ctrl)
			mockYaml := yaml.NewMockYamlAccess(ctrl)
			subject := &Loader{
				File: mockFile,
				Yaml: mockYaml,
			}

			invocations := 0
			if c.yamlError == nil {
				for _, lp := range c.expectedPaths {
					mockYaml.EXPECT().Load(lp, &Library{}).Times(1).Return(nil).Do(func(path string, lib *Library) {
						if c.testLibs != nil && len(c.testLibs) >= invocations {
							*lib = c.testLibs[invocations]
							invocations++
						}
					})
				}
				for _, r := range c.expectedResolutions {
					mockFile.EXPECT().ResolveRelativeTo(r.target, r.source).Times(1).Return(r.result)
				}
			} else {
				mockYaml.EXPECT().Load(c.expectedPaths[0], &Library{}).Times(1).Return(c.yamlError)
			}

			loadedLibs, err := subject.Load(c.paths)

			if !reflect.DeepEqual(c.expectedError, err) {
				t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedError, err)
			}

			if c.expectedError == nil {
				if !reflect.DeepEqual(c.expectedLoadedLibs, loadedLibs) {
					t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", c.expectedLoadedLibs, loadedLibs)
				}
			}
		})
	}
}
