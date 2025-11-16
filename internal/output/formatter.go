package output

import (
	"fmt"
	"io"
	"os"
)

// Output handles formatted output to configurable writers
type Output struct {
	Stdout io.Writer
	Stderr io.Writer
}

// NewOutput creates an Output with default writers
func NewOutput(stdout, stderr io.Writer) *Output {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	return &Output{
		Stdout: stdout,
		Stderr: stderr,
	}
}

// PrintInfo prints informational message to stdout
func (o *Output) PrintInfo(format string, args ...interface{}) {
	fmt.Fprintf(o.Stdout, format+"\n", args...)
}

// PrintSuccess prints success message to stdout
func (o *Output) PrintSuccess(format string, args ...interface{}) {
	fmt.Fprintf(o.Stdout, format+"\n", args...)
}

// PrintError prints error message to stderr
func (o *Output) PrintError(format string, args ...interface{}) {
	fmt.Fprintf(o.Stderr, format+"\n", args...)
}
