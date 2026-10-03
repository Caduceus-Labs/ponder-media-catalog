// Package cli declares the ponder-catalog command surface.
//
// The command names are part of the system's contract with its external
// scheduler (see docs/ARCHITECTURE.md): "ingest <source>", "publish" and
// "run-all". The job bodies are implemented in later phases; until then each
// declared command fails fast with a non-zero exit so nothing pretends to
// succeed and no database or output file is touched.
package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

// ErrNotImplemented marks a command that is declared but not yet built.
var ErrNotImplemented = errors.New("not implemented")

// Execute builds the root command and runs it, returning any error.
func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "ponder-catalog",
		Short:         "Build and publish the public Ponder media catalog",
		Long: "ponder-catalog ingests media metadata from external sources into a\n" +
			"local catalog database and publishes the public JSON consumed by the\n" +
			"Ponder application. Each job is a one-shot command started by an\n" +
			"external scheduler.",
		// Keep errors terse: print the error, not a full usage dump.
		SilenceUsage: true,
	}
	root.AddCommand(newIngestCommand(), newPublishCommand(), newRunAllCommand())
	return root
}

func newIngestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ingest <source>",
		Short: "Sync one external source into the catalog database",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return errors.New("ingest " + args[0] + ": " + ErrNotImplemented.Error())
		},
	}
}

func newPublishCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "publish",
		Short: "Render the public JSON files from the catalog database",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("publish: " + ErrNotImplemented.Error())
		},
	}
}

func newRunAllCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run-all",
		Short: "Run every source job, then publish (local and recovery use)",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return errors.New("run-all: " + ErrNotImplemented.Error())
		},
	}
}
