package scenario

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/cjnosal/manifer/pkg/library"
)

func TestListScenarios(t *testing.T) {

	independantLibrary := &library.LoadedLibrary{
		Path: "./lib/library3.yml",
		Library: &library.Library{
			Type: library.OpsFile,
			Scenarios: []library.Scenario{
				{
					Name:        "extra",
					Description: "an additional scenario",
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
					Name:        "big_dependency",
					Description: "a bigger utility",
					GlobalArgs:  []string{},
					Args:        []string{},
					Snippets:    []library.Snippet{},
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
					Name:        "main",
					Description: "the default",
					GlobalArgs:  []string{"g1"},
					Args:        []string{"a1"},
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
					Name:        "big",
					Description: "include everything",
					GlobalArgs:  []string{},
					Args:        []string{},
					Snippets:    []library.Snippet{},
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

	t.Run("list provided libraries", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		loader := library.NewMockLibraryLoader(ctrl)
		subject := &Lister{
			Loader: loader,
		}

		loader.EXPECT().Load([]string{"lib1", "lib2"}).Times(1).Return(providedLibraries, nil)

		entries, err := subject.ListScenarios([]string{"lib1", "lib2"}, false)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := []ScenarioEntry{
			{
				Name:        "main",
				Description: "the default",
			},
			{
				Name:        "big",
				Description: "include everything",
			},
			{
				Name:        "extra",
				Description: "an additional scenario",
			},
		}

		if !reflect.DeepEqual(expected, entries) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, entries)
		}
	})

	t.Run("list all referenced libraries", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		loader := library.NewMockLibraryLoader(ctrl)
		subject := &Lister{
			Loader: loader,
		}

		loader.EXPECT().Load([]string{"lib1", "lib2"}).Times(1).Return(providedLibraries, nil)

		entries, err := subject.ListScenarios([]string{"lib1", "lib2"}, true)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := []ScenarioEntry{
			{
				Name:        "main",
				Description: "the default",
			},
			{
				Name:        "big",
				Description: "include everything",
			},
			{
				Name:        "ref.dependency",
				Description: "",
			},
			{
				Name:        "ref.big_dependency",
				Description: "a bigger utility",
			},
			{
				Name:        "extra",
				Description: "an additional scenario",
			},
		}

		if !reflect.DeepEqual(expected, entries) {
			t.Errorf("Expected:\n'''%v'''\nActual:\n'''%v'''\n", expected, entries)
		}
	})

	t.Run("load error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		loader := library.NewMockLibraryLoader(ctrl)
		subject := &Lister{
			Loader: loader,
		}

		loader.EXPECT().Load([]string{"lib1", "lib2"}).Times(1).Return(nil, errors.New("test"))

		_, err := subject.ListScenarios([]string{"lib1", "lib2"}, false)

		if err == nil || err.Error() != "test" {
			t.Errorf("Loader error not reported")
		}
	})

}
