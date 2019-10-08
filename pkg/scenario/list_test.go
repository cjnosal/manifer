package scenario

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/cjnosal/manifer/pkg/library"
	"github.com/google/go-cmp/cmp"
)

func TestListScenarios(t *testing.T) {

	independantLibrary := &library.Library{
		Type: library.OpsFile,
		Scenarios: []library.Scenario{
			{
				Name:        "extra",
				Description: "an additional scenario",
				Snippets: []library.Snippet{
					{
						Path: "/wd/lib/snippet3.yml",
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
						Path: "/wd/lib/snippet2.yml",
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
	}

	referencingLibrary := &library.Library{
		Type: library.OpsFile,
		Libraries: []library.LibraryRef{
			{
				Alias: "ref",
				Path:  "/wd/lib/library2.yml",
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
						Path: "/wd/lib/snippet1.yml",
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
	}

	loaded := &library.LoadedLibrary{
		TopLibraries: []*library.Library{
			referencingLibrary,
			independantLibrary,
		},
		Libraries: map[string]*library.Library{
			"/wd/lib/library3.yml": independantLibrary,
			"/wd/lib/library2.yml": referencedLibrary,
			"/wd/lib/library.yml":  referencingLibrary,
		},
	}

	t.Run("list provided libraries", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		loader := library.NewMockLibraryLoader(ctrl)
		subject := &Lister{
			Loader: loader,
		}

		loader.EXPECT().Load([]string{"./lib/library.yml", "./lib/library3.yml"}).Times(1).Return(loaded, nil)

		entries, err := subject.ListScenarios([]string{"./lib/library.yml", "./lib/library3.yml"}, false)

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

		if !cmp.Equal(expected, entries) {
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

		loader.EXPECT().Load([]string{"./lib/library.yml", "./lib/library3.yml"}).Times(1).Return(loaded, nil)

		entries, err := subject.ListScenarios([]string{"./lib/library.yml", "./lib/library3.yml"}, true)

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

		if !cmp.Equal(expected, entries) {
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
