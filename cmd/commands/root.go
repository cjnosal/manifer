package commands

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/cjnosal/manifer/lib"
)

var rootCmd = &cobra.Command{
	Use:              "manifer",
	Short:            "a yaml composer",
	TraverseChildren: true,
}
var libraryPaths []string

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

	return rootCmd
}
