package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cjnosal/manifer/lib"
	"github.com/cjnosal/manifer/pkg/file"
	"github.com/cjnosal/manifer/pkg/library"
	"github.com/cjnosal/manifer/pkg/yaml"
)

var rootCmd = &cobra.Command{
	Use:              "manifer",
	Short:            "a yaml composer",
	TraverseChildren: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if len(libraryPaths) == 0 {
			libraryPaths = defaultLibPaths
		}
	},
}

var libraryPaths []string
var defaultLibPaths []string

func Init(logger io.Writer, writer io.Writer, maniferLib lib.Manifer) *cobra.Command {
	rootCmd.PersistentFlags().StringSliceVarP(&libraryPaths, "library", "l", []string{}, "Path to library file")

	// register subcommands
	rootCmd.AddCommand(NewComposeCommand(logger, writer, maniferLib))
	rootCmd.AddCommand(NewListCommand(logger, writer, maniferLib))
	rootCmd.AddCommand(NewSearchCommand(logger, writer, maniferLib))
	rootCmd.AddCommand(NewInspectCommand(logger, writer, maniferLib))
	rootCmd.AddCommand(NewImportCommand(logger, writer, maniferLib))
	rootCmd.AddCommand(NewGenerateCommand(logger, writer, maniferLib))
	rootCmd.AddCommand(NewAddCommand(logger, writer, maniferLib))

	// viper.SetEnvPrefix("manifer")
	viper.BindEnv("lib_path", "MANIFER_LIB_PATH")
	viper.BindEnv("libs", "MANIFER_LIBS")

	// specific libraries from env
	envLibsString := viper.GetString("libs")
	if len(envLibsString) > 0 {
		defaultLibPaths = strings.Split(envLibsString, string(os.PathListSeparator))
	}

	if len(defaultLibPaths) == 0 {
		// search paths from env
		envLibPathString := viper.GetString("lib_path")
		if len(envLibPathString) > 0 {
			envLibPath := strings.Split(envLibPathString, string(os.PathListSeparator))

			for _, dir := range envLibPath {
				matches, _ := filepath.Glob(dir + string(os.PathSeparator) + "*.yml")
				for _, path := range matches {
					if validateLib(path) {
						defaultLibPaths = append(defaultLibPaths, path)
					}
				}
			}
		}
	}

	return rootCmd
}

func validateLib(path string) bool {
	fileAccess := &file.FileIO{}
	yamlAccess := &yaml.Yaml{}
	content, err := fileAccess.Read(path)
	if err == nil {
		lib := library.Library{}
		err = yamlAccess.Unmarshal(content, &lib)
		if err == nil && lib.Type != "" && (len(lib.Scenarios) > 0 || len(lib.Libraries) > 0) {
			return true
		}
	}
	return false
}
