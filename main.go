package main

import (
	"os"

	"github.com/kapetan-io/claude-md.go/cmd"
)

func main() {
	args := os.Args[1:]
	exitCode := cmd.Run(args, cmd.RunOptions{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	os.Exit(exitCode)
}
