package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bbsify-landed/clog"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Args); err != nil {
		clog.Error(ctx, "hwgen failed", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
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

	clog.Info(ctx, "generated endpoints", "count", len(defs))
	return nil
}
