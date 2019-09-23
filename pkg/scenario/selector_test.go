package scenario

import (
	"errors"
	"reflect"
	"testing"

	"github.com/cjnosal/manifer/pkg/library"
)

func TestSelectScenarios(t *testing.T) {

	independantLibrary := &library.Library{
		Type: library.OpsFile,
		Scenarios: []library.Scenario{
			{
				Name: "extra",
				Snippets: []library.Snippet{
					{
						Path: "/wd/snippet3.yml",
					},
				},
			},
		},
	}

	referencedLibrary := &library.Library{
		Type: library.OpsFile,
		Scenarios: []library.Scenario{
			{
				Name:       "dependency",
				GlobalArgs: []string{"g2"},
				Args:       []string{"a2"},
				Snippets: []library.Snippet{
					{
						Path: "/wd/snippet2.yml",
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
	}

	referencingLibrary := &library.Library{
		Type: library.OpsFile,
		Libraries: []library.LibraryRef{
			{
				Alias: "ref",
				Path:  "/wd/library2.yml",
			},
		},
		Scenarios: []library.Scenario{
			{
				Name:       "main",
				GlobalArgs: []string{"g1"},
				Args:       []string{"a1"},
				Snippets: []library.Snippet{
					{
						Path: "/wd/snippet1.yml",
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
	}

	providedLibraries := &library.LoadedLibrary{
		TopLibraries: []*library.Library{referencingLibrary, independantLibrary},
		Libraries: map[string]*library.Library{
			"/wd/library.yml":  referencingLibrary,
			"/wd/library2.yml": referencedLibrary,
			"/wd/library3.yml": independantLibrary,
		},
	}

	cases := []struct {
		name          string
		scenarioNames []string
		lookupError   error
		expectedPlan  *Plan
		expectedError error
	}{
		{
			name: "scenario with dependency",
			scenarioNames: []string{
				"main",
			},
			expectedPlan: &Plan{
				GlobalArgs: []string{
					"g2",
					"g1",
				},
				Snippets: []library.Snippet{
					{
						Path: "/wd/snippet2.yml",
						Args: []string{
							"s2",
							"a2",
							"r1",
							"a1",
						},
					},
					{
						Path: "/wd/snippet1.yml",
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
			expectedPlan: &Plan{
				GlobalArgs: []string{
					"g2",
				},
				Snippets: []library.Snippet{
					{
						Path: "/wd/snippet2.yml",
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

			expectedPlan: &Plan{
				GlobalArgs: []string{},
				Snippets: []library.Snippet{
					{
						Path: "/wd/snippet3.yml",
					},
				},
			},
		},
		{
			name: "lookup error",
			scenarioNames: []string{
				"doesnotexist",
			},
			expectedError: errors.New("Unable to resolve scenario doesnotexist"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			subject := &Selector{}

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
