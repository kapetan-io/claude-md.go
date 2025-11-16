package main

import (
	"os"

	"github.com/kapetan-io/claude-md.go/internal/cli"
)

func main() {
	args := os.Args[1:]
	exitCode := cli.Run(args, cli.RunOptions{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	os.Exit(exitCode)
}
