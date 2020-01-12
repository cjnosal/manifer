package main

import (
	"fmt"
	"os"

	"github.com/cjnosal/manifer/v2/cmd/commands"
	"github.com/cjnosal/manifer/v2/lib"
)

func main() {
	// setup dependencies
	logger := os.Stderr
	writer := os.Stdout

	maniferLib := lib.NewManifer(logger)

	rootCmd := commands.Init(logger, writer, maniferLib)

	// run
	if err := rootCmd.Execute(); err != nil {
		logger.Write([]byte(fmt.Sprintf("%v\n", err)))
		os.Exit(1)
	}
}
