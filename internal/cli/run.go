package cli

import (
	"io"
	"os"

	"github.com/kapetan-io/claude-md.go/internal/output"
)

// Package-level variable for commands to access output
var currentOutput *output.Output

// RunOptions provides injectable dependencies for testing
type RunOptions struct {
	Stdout io.Writer
	Stderr io.Writer
}

// Run executes the CLI with given arguments and options
func Run(args []string, opts RunOptions) int {
	// Set defaults
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	// Make output available to commands
	currentOutput = output.NewOutput(opts.Stdout, opts.Stderr)

	// Configure Cobra's output streams (for help text, errors)
	rootCmd.SetOut(opts.Stdout)
	rootCmd.SetErr(opts.Stderr)

	// Let Cobra parse args and route to commands
	rootCmd.SetArgs(args)

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		return 1
	}
	return 0
}
