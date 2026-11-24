// Command generate writes the doc comments of the package in the working
// directory as Go source: run by `go generate` in pkg/config.
//
//	go run ../../internal/fielddescriptions/generate <import path> <output file>
package main

import (
	"fmt"
	"os"
	"path"

	"github.com/litebitus/go-publisher/internal/fielddescriptions"
)

const generator = "go run ./internal/fielddescriptions/generate"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(arguments []string) error {
	if len(arguments) != 2 {
		return fmt.Errorf("usage: generate <import path> <output file>")
	}
	importPath, output := arguments[0], arguments[1]
	descriptions, err := fielddescriptions.Extract(".", importPath)
	if err != nil {
		return err
	}
	source, err := fielddescriptions.Source(path.Base(importPath), generator, descriptions)
	if err != nil {
		return err
	}
	if err := os.WriteFile(output, source, 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", output, err)
	}
	return nil
}
