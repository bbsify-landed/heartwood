package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	if err := run(os.Args, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "hwgen: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	dir := "."
	if len(args) > 1 {
		dir = args[1]
	}

	// Remove old generated files before loading the package,
	// since stale generated code can cause package load errors.
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	for _, name := range []string{"hw_gen.go", "hw_client_gen.go"} {
		_ = os.Remove(filepath.Join(absDir, name))
	}

	defs, pkgName, err := loadDefinitions(dir)
	if err != nil {
		return err
	}

	if len(defs) == 0 {
		return fmt.Errorf("no schema.Definition variables found in %s", dir)
	}

	if err := generate(dir, pkgName, defs); err != nil {
		return err
	}

	fmt.Fprintf(stderr, "hwgen: generated %d endpoint(s)\n", len(defs))
	return nil
}
