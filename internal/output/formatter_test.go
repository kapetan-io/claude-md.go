package output_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/kapetan-io/claude-md.go/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintInfo(t *testing.T) {
	var stdout, stderr bytes.Buffer

	out := output.NewOutput(&stdout, &stderr)
	out.PrintInfo("test message: %s", "value")

	assert.Equal(t, "test message: value\n", stdout.String())
	assert.Empty(t, stderr.String())
}

func TestPrintSuccess(t *testing.T) {
	var stdout, stderr bytes.Buffer

	out := output.NewOutput(&stdout, &stderr)
	out.PrintSuccess("success: %d items", 5)

	assert.Equal(t, "success: 5 items\n", stdout.String())
	assert.Empty(t, stderr.String())
}

func TestPrintError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	out := output.NewOutput(&stdout, &stderr)
	out.PrintError("error: %v", "something went wrong")

	assert.Empty(t, stdout.String())
	assert.Equal(t, "error: something went wrong\n", stderr.String())
}

func TestNewOutputDefaults(t *testing.T) {
	out := output.NewOutput(nil, nil)

	require.NotNil(t, out)
	assert.Equal(t, os.Stdout, out.Stdout)
	assert.Equal(t, os.Stderr, out.Stderr)
}
