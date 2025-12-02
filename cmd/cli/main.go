// Package main provides the CLI entry point for the Telos Idea Matrix application.
package main

import (
	"fmt"
	"os"

	"github.com/ryacub/telos-idea-matrix/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
