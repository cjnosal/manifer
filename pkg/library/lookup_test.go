package library

import (
	"errors"
	"reflect"
	"testing"
)

func TestGetContainingLibrary(t *testing.T) {

	independantLibrary := &LoadedLibrary{
		Path: "./lib/library3.yml",
		Library: &Library{
			Type: OpsFile,
			Scenarios: []Scenario{
				{
					Name: "extra",
					Snippets: []Snippet{
						{
							Path: "./lib/snippet3.yml",
						},
					},
				},
			},
		},
		References: map[string]*LoadedLibrary{},
	}

	referencedLibrary := &LoadedLibrary{
		Path: "./lib/library2.yml",
		Library: &Library{
			Type: OpsFile,
			Scenarios: []Scenario{
				{
					Name: "dependency",
					Snippets: []Snippet{
						{
							Path: "./lib/snippet2.yml",
						},
					},
				},
			},
		},
		References: map[string]*LoadedLibrary{},
	}

	referencingLibrary := &LoadedLibrary{
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
					Name: "main",
					Snippets: []Snippet{
						{
							Path: "./lib/snippet.yml",
						},
					},
					Scenarios: []ScenarioRef{
						{
							Name: "ref.dependency",
						},
					},
				},
			},
		},
		References: map[string]*LoadedLibrary{
			"ref": referencedLibrary,
		},
	}

	loadedLibraries := []LoadedLibrary{
		*referencingLibrary,
		*independantLibrary,
	}

	cases := []struct {
		name            string
		scenarioName    string
		expectedLibrary *LoadedLibrary
		expectedError   error
	}{
		{
			name:            "scenario in first provided library",
			scenarioName:    "main",
			expectedLibrary: referencingLibrary,
		},
		{
			name:            "scenario in second provided library",
			scenarioName:    "extra",
			expectedLibrary: independantLibrary,
		},
		{
			name:            "scenario in referenced library",
			scenarioName:    "ref.dependency",
			expectedLibrary: referencedLibrary,
		},
		{
			name:          "unknown scenario",
			scenarioName:  "doesnotexist",
			expectedError: errors.New("Unable to find scenario 'doesnotexist'"),
		},
		{
			name:          "unknown library reference",
			scenarioName:  "invalidref.scenario",
			expectedError: errors.New("Unable to find scenario 'invalidref.scenario'"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			subject := &Lookup{}

			lib, err := subject.GetContainingLibrary(c.scenarioName, loadedLibraries)

			if !reflect.DeepEqual(c.expectedError, err) {
				t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedError, err)
			}

			if c.expectedError == nil {
				if !reflect.DeepEqual(c.expectedLibrary, lib) {
					t.Errorf("Expected:\n'''%s'''\nActual:\n'''%s'''\n", c.expectedLibrary, lib)
				}
			}
		})
	}
}
