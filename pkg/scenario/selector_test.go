package scenario

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/cjnosal/manifer/pkg/library"
)

func TestSelectScenarios(t *testing.T) {

	independantLibrary := &library.LoadedLibrary{
		Path: "./lib/library3.yml",
		Library: &library.Library{
			Type: library.OpsFile,
			Scenarios: []library.Scenario{
				{
					Name: "extra",
					Snippets: []library.Snippet{
						{
							Path: "./lib/snippet3.yml",
						},
					},
				},
			},
		},
		References: map[string]*library.LoadedLibrary{},
	}

	referencedLibrary := &library.LoadedLibrary{
		Path: "./lib/library2.yml",
		Library: &library.Library{
			Type: library.OpsFile,
			Scenarios: []library.Scenario{
				{
					Name:       "dependency",
					GlobalArgs: []string{"g2"},
					Args:       []string{"a2"},
					Snippets: []library.Snippet{
						{
							Path: "./lib/snippet2.yml",
							Args: []string{"s2"},
						},
					},
				},
				{
					Name:       "big_dependency",
					GlobalArgs: []string{},
					Args:       []string{},
					Snippets:   []library.Snippet{},
					Scenarios: []library.ScenarioRef{
						{
							Name: "dependency",
							Args: []string{},
						},
					},
				},
			},
		},
		References: map[string]*library.LoadedLibrary{},
	}

	referencingLibrary := &library.LoadedLibrary{
		Path: "./lib/library.yml",
		Library: &library.Library{
			Type: library.OpsFile,
			Libraries: []library.LibraryRef{
				{
					Alias: "library2",
					Path:  "./library2.yml",
				},
			},
			Scenarios: []library.Scenario{
				{
					Name:       "main",
					GlobalArgs: []string{"g1"},
					Args:       []string{"a1"},
					Snippets: []library.Snippet{
						{
							Path: "./lib/snippet1.yml",
							Args: []string{"s1"},
						},
					},
					Scenarios: []library.ScenarioRef{
						{
							Name: "ref.dependency",
							Args: []string{"r1"},
						},
					},
				},
				{
					Name:       "big",
					GlobalArgs: []string{},
					Args:       []string{},
					Snippets:   []library.Snippet{},
					Scenarios: []library.ScenarioRef{
						{
							Name: "ref.big_dependency",
							Args: []string{},
						},
					},
				},
			},
		},
		References: map[string]*library.LoadedLibrary{
			"ref": referencedLibrary,
		},
	}

	providedLibraries := []library.LoadedLibrary{
		*referencingLibrary,
		*independantLibrary,
	}

	type lookup struct {
		scenarioName string
		searchLibs   []library.LoadedLibrary
		loadLib      *library.LoadedLibrary
	}

	cases := []struct {
		name          string
		scenarioNames []string
		lookups       []lookup
		lookupError   error
		expectedPlan  *Plan
		expectedError error
	}{
		{
			name: "scenario with dependency",
			scenarioNames: []string{
				"main",
			},
			lookups: []lookup{
				{
					scenarioName: "main",
					searchLibs:   providedLibraries,
					loadLib:      referencingLibrary,
				},
				{
					scenarioName: "ref.dependency",
					searchLibs: []library.LoadedLibrary{
						*referencingLibrary,
					},
					loadLib: referencedLibrary,
				},
			},
			expectedPlan: &Plan{
				GlobalArgs: []string{
					"g2",
					"g1",
				},
				Snippets: []library.Snippet{
					{
						Path: "./lib/snippet2.yml",
						Args: []string{
							"s2",
							"a2",
							"r1",
							"a1",
						},
					},
					{
						Path: "./lib/snippet1.yml",
						Args: []string{
							"s1",
							"a1",
						},
					},
				},
			},
		},
		{
			name: "local reference inside referenced library",
			scenarioNames: []string{
				"ref.big_dependency",
			},
			lookups: []lookup{
				{
					scenarioName: "ref.big_dependency",
					searchLibs:   providedLibraries,
					loadLib:      referencedLibrary,
				},
				{
					scenarioName: "dependency",
					searchLibs: []library.LoadedLibrary{
						*referencedLibrary,
					},
					loadLib: referencedLibrary,
				},
			},
			expectedPlan: &Plan{
				GlobalArgs: []string{
					"g2",
				},
				Snippets: []library.Snippet{
					{
						Path: "./lib/snippet2.yml",
						Args: []string{
							"s2",
							"a2",
						},
					},
				},
			},
		},
		{
			name: "scenario in second provided library",
			scenarioNames: []string{
				"extra",
			},
			lookups: []lookup{
				{
					scenarioName: "extra",
					searchLibs:   providedLibraries,
					loadLib:      independantLibrary,
				},
			},

			expectedPlan: &Plan{
				GlobalArgs: []string{},
				Snippets: []library.Snippet{
					{
						Path: "./lib/snippet3.yml",
					},
				},
			},
		},
		{
			name: "lookup error",
			scenarioNames: []string{
				"doesnotexist",
			},
			lookups: []lookup{
				{
					scenarioName: "doesnotexist",
					searchLibs:   providedLibraries,
				},
			},
			lookupError:   errors.New("test"),
			expectedError: errors.New("test"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLookup := library.NewMockLibraryLookup(ctrl)

			for _, l := range c.lookups {
				mockLookup.EXPECT().GetContainingLibrary(l.scenarioName, l.searchLibs).Times(1).Return(l.loadLib, c.lookupError)
			}
			subject := &Selector{
				Lookup: mockLookup,
			}

			lib, err := subject.SelectScenarios(c.scenarioNames, providedLibraries)

			if !reflect.DeepEqual(c.expectedError, err) {
				t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedError, err)
			}

			if c.expectedError == nil {
				if !reflect.DeepEqual(c.expectedPlan, lib) {
					t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedPlan, lib)
				}
			}
		})
	}
}
