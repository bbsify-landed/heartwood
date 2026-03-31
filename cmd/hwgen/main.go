package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	// Remove old generated files before loading the package,
	// since stale generated code can cause package load errors.
	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hwgen: %v\n", err)
		os.Exit(1)
	}
	for _, name := range []string{"hw_gen.go", "hw_client_gen.go"} {
		_ = os.Remove(filepath.Join(absDir, name))
	}

	defs, pkgName, err := loadDefinitions(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "hwgen: %v\n", err)
		os.Exit(1)
	}

	if len(defs) == 0 {
		fmt.Fprintf(os.Stderr, "hwgen: no schema.Definition variables found in %s\n", dir)
		os.Exit(1)
	}

	if err := generate(dir, pkgName, defs); err != nil {
		fmt.Fprintf(os.Stderr, "hwgen: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "hwgen: generated %d endpoint(s)\n", len(defs))
}
