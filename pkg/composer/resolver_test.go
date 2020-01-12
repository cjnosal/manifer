package composer

import (
	"errors"
	"testing"

	"github.com/cjnosal/manifer/test"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/cjnosal/manifer/pkg/interpolator"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/plan"
	"github.com/cjnosal/manifer/pkg/processor"
	"github.com/cjnosal/manifer/pkg/processor/factory"
)

func TestResolve(t *testing.T) {

	cases := []struct {
		name                           string
		libraryPaths                   []string
		scenarioNames                  []string
		passthrough                    []string
		yamlError                      error
		expectedLibraries              *library.LoadedLibrary
		expectedOpsFilePassthroughNode *library.ScenarioNode
		expectedVarNode                *library.ScenarioNode
		parseError                     error
		expectedPlan                   *plan.Plan
		expectedError                  error
		parseVarError                  error
	}{
		{
			name: "generate plan",
			libraryPaths: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			passthrough:                    []string{},
			expectedOpsFilePassthroughNode: nil,
			expectedLibraries: &library.LoadedLibrary{
				TopLibraries: []*library.Library{
					{
						Type: library.OpsFile,
						Scenarios: []library.Scenario{
							{
								Name: "a scenario",
								Snippets: []library.Snippet{
									{
										Path: "/foo.yml",
										Interpolator: library.InterpolatorParams{
											Vars: map[string]interface{}{"arg": "value"},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedPlan: &plan.Plan{
				Global: library.InterpolatorParams{
					Vars:      map[string]interface{}{},
					RawArgs:   []string{},
					VarFiles:  map[string]string{},
					VarsFiles: []string{},
					VarsEnv:   []string{},
				},
				Steps: []*plan.Step{
					{
						Snippet: "/foo.yml",
						Params: []plan.TaggedParams{
							{
								Tag: "snippet",
								Interpolator: library.InterpolatorParams{
									Vars: map[string]interface{}{"arg": "value"},
								},
							},
							{
								Tag:          "a scenario",
								Interpolator: library.InterpolatorParams{},
							},
						},
						Processor: library.Processor{
							Type: library.OpsFile,
						},
					},
				},
			},
		},
		{
			name: "append passthrough snippets",
			libraryPaths: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{},
			passthrough: []string{
				"-o=opsfile",
			},
			expectedOpsFilePassthroughNode: &library.ScenarioNode{
				Snippets: []library.Snippet{
					{
						Path:         "opsfile",
						Interpolator: library.InterpolatorParams{},
						Processor: library.Processor{
							Type: library.OpsFile,
						},
					},
				},
				Name:        "passthrough opsfile",
				Description: "args passed after --",
				LibraryPath: "<cli>",
			},
			expectedLibraries: &library.LoadedLibrary{
				TopLibraries: []*library.Library{
					{
						Type: library.OpsFile,
						Scenarios: []library.Scenario{
							{
								Name: "a scenario",
								Snippets: []library.Snippet{
									{
										Path: "/foo.yml",
										Interpolator: library.InterpolatorParams{
											Vars: map[string]interface{}{"arg": "value"},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedPlan: &plan.Plan{
				Global: library.InterpolatorParams{
					Vars:      map[string]interface{}{},
					RawArgs:   []string{},
					VarFiles:  map[string]string{},
					VarsFiles: []string{},
					VarsEnv:   []string{},
				},
				Steps: []*plan.Step{
					{
						Snippet: "opsfile",
						Params: []plan.TaggedParams{
							{
								Tag:          "snippet",
								Interpolator: library.InterpolatorParams{},
							},
							{
								Tag:          "passthrough opsfile",
								Interpolator: library.InterpolatorParams{},
							},
						},
						Processor: library.Processor{
							Type: library.OpsFile,
						},
					},
				},
			},
		},
		{
			name: "passthrough snippets error",
			libraryPaths: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			passthrough: []string{
				"-oextra=o",
			},
			parseError: errors.New("test"),
			expectedLibraries: &library.LoadedLibrary{
				TopLibraries: []*library.Library{
					{
						Type: library.OpsFile,
						Scenarios: []library.Scenario{
							{
								Name: "a scenario",
								GlobalInterpolator: library.InterpolatorParams{
									Vars: map[string]interface{}{"extra": "glob"},
								},
								Snippets: []library.Snippet{
									{
										Path: "/foo.yml",
										Interpolator: library.InterpolatorParams{
											Vars: map[string]interface{}{"arg": "value"},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedError: errors.New("test\n  while trying to parse opsfile passthrough args"),
		},
		{
			name: "yaml error",
			libraryPaths: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			yamlError:     errors.New("test"),
			expectedError: errors.New("test\n  while trying to load libraries"),
		},
		{
			name: "append passthrough variables",
			libraryPaths: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			passthrough: []string{
				"-vextra=e",
			},
			expectedVarNode: &library.ScenarioNode{
				GlobalInterpolator: library.InterpolatorParams{
					RawArgs: []string{"-vextra=e"},
				},
				Snippets:    []library.Snippet{},
				Name:        "passthrough variables",
				Description: "vars passed after --",
				LibraryPath: "<cli>",
			},
			expectedLibraries: &library.LoadedLibrary{
				TopLibraries: []*library.Library{
					{
						Type: library.OpsFile,
						Scenarios: []library.Scenario{
							{
								Name: "a scenario",
								GlobalInterpolator: library.InterpolatorParams{
									Vars: map[string]interface{}{"extra": "glob"},
								},
								Snippets: []library.Snippet{
									{
										Path: "/foo.yml",
										Interpolator: library.InterpolatorParams{
											Vars: map[string]interface{}{"arg": "value"},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedPlan: &plan.Plan{
				Global: library.InterpolatorParams{
					Vars:      map[string]interface{}{"extra": "glob"},
					RawArgs:   []string{"-vextra=e"},
					VarFiles:  map[string]string{},
					VarsFiles: []string{},
					VarsEnv:   []string{},
				},
				Steps: []*plan.Step{
					{
						Snippet: "/foo.yml",
						Params: []plan.TaggedParams{
							{
								Tag: "snippet",
								Interpolator: library.InterpolatorParams{
									Vars: map[string]interface{}{"arg": "value"},
								},
							},
							{
								Tag:          "a scenario",
								Interpolator: library.InterpolatorParams{},
							},
						},
						Processor: library.Processor{
							Type: library.OpsFile,
						},
					},
				},
			},
		},
		{
			name: "passthrough variables error",
			libraryPaths: []string{
				"/tmp/library/lib.yml",
			},
			scenarioNames: []string{
				"a scenario",
			},
			passthrough: []string{
				"-vextra=e",
			},
			parseVarError: errors.New("test"),
			expectedLibraries: &library.LoadedLibrary{
				TopLibraries: []*library.Library{
					{
						Type: library.OpsFile,
						Scenarios: []library.Scenario{
							{
								Name: "a scenario",
								GlobalInterpolator: library.InterpolatorParams{
									Vars: map[string]interface{}{"extra": "glob"},
								},
								Snippets: []library.Snippet{
									{
										Path: "/foo.yml",
										Interpolator: library.InterpolatorParams{
											Vars: map[string]interface{}{"arg": "value"},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedError: errors.New("test\n  while trying to parse passthrough vars"),
		},
	}

	for _, c := range cases {

		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLoader := library.NewMockLibraryLoader(ctrl)
			mockProcessorFactory := factory.NewMockProcessorFactory(ctrl)
			mockOpsProcessor := processor.NewMockProcessor(ctrl)
			mockInterpolator := interpolator.NewMockInterpolator(ctrl)

			mockLoader.EXPECT().Load(c.libraryPaths).Times(1).Return(c.expectedLibraries, c.yamlError)
			if c.yamlError == nil {
				mockProcessorFactory.EXPECT().Create(library.OpsFile).Times(1).Return(mockOpsProcessor, nil)
				mockOpsProcessor.EXPECT().ParsePassthroughFlags(c.passthrough).Times(1).Return(c.expectedOpsFilePassthroughNode, []string{"opsremainder"}, c.parseError)
			}
			if c.yamlError == nil && c.parseError == nil {
				mockInterpolator.EXPECT().ParsePassthroughVars([]string{"opsremainder"}).Times(1).Return(c.expectedVarNode, []string{}, c.parseVarError)
			}

			subject := Resolver{
				Loader:           mockLoader,
				ProcessorFactory: mockProcessorFactory,
				Interpolator:     mockInterpolator,
			}

			plan, err := subject.Resolve(c.libraryPaths, c.scenarioNames, c.passthrough)

			if !cmp.Equal(&c.expectedError, &err, cmp.Comparer(test.EqualMessage)) {
				t.Errorf("Expected error:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedError, err)
			}

			if err == nil {
				if !cmp.Equal(c.expectedPlan, plan) {
					t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\nDiff:\n'''%s'''\n",
						c.expectedPlan, plan, cmp.Diff(c.expectedPlan, plan))
				}
			}
		})
	}
}
