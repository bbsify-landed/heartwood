package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/bbsify-landed/heartwood/pkg/schema"
)

// loadDefinitions loads a Go package from dir, finds all *schema.Definition
// variables, and returns their runtime values by compiling and executing a
// temporary program.
func loadDefinitions(dir string) (map[string]*schema.Definition, string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, "", fmt.Errorf("resolving path: %w", err)
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		Dir:  absDir,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, "", fmt.Errorf("loading package: %w", err)
	}
	if len(pkgs) == 0 {
		return nil, "", fmt.Errorf("no packages found in %s", dir)
	}

	pkg := pkgs[0]
	if len(pkg.Errors) > 0 {
		var msgs []string
		for _, e := range pkg.Errors {
			msgs = append(msgs, e.Msg)
		}
		return nil, "", fmt.Errorf("package errors: %s", strings.Join(msgs, "; "))
	}

	// Walk AST to find var declarations whose type is *schema.Definition
	var varNames []string
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range genDecl.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range vs.Names {
					obj := pkg.TypesInfo.ObjectOf(name)
					if obj == nil {
						continue
					}
					typStr := obj.Type().String()
					if typStr == "*github.com/bbsify-landed/heartwood/pkg/schema.Definition" {
						varNames = append(varNames, name.Name)
					}
				}
			}
		}
	}

	if len(varNames) == 0 {
		return nil, pkg.Name, nil
	}

	// Generate and run a temporary program to extract runtime values
	defs, err := execExtractor(pkg.PkgPath, varNames, absDir)
	if err != nil {
		return nil, "", err
	}

	return defs, pkg.Name, nil
}

// execExtractor generates a temporary Go program that imports the user's
// schema package, serializes the Definition variables to JSON, and runs it.
func execExtractor(pkgPath string, varNames []string, schemaDir string) (map[string]*schema.Definition, error) {
	tmpDir, err := os.MkdirTemp("", "hwgen-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Build the assignments: "HealthCheck": userschema.HealthCheck,
	var assignments []string
	for _, name := range varNames {
		assignments = append(assignments, fmt.Sprintf("\t\t%q: userschema.%s,", name, name))
	}

	src := fmt.Sprintf(`package main

import (
	"encoding/json"
	"os"

	"github.com/bbsify-landed/heartwood/pkg/schema"
	userschema %q
)

// Ensure schema import is used
var _ *schema.Definition

func main() {
	defs := map[string]*schema.Definition{
%s
	}

	for name, def := range defs {
		def.Name = name
	}

	if err := json.NewEncoder(os.Stdout).Encode(defs); err != nil {
		os.Stderr.WriteString("hwgen extractor: " + err.Error() + "\n")
		os.Exit(1)
	}
}
`, pkgPath, strings.Join(assignments, "\n"))

	mainFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(mainFile, []byte(src), 0o644); err != nil {
		return nil, fmt.Errorf("writing temp main.go: %w", err)
	}

	// Find the module root (directory containing go.mod) so we can set up
	// a go.work or use replace directives.
	modRoot, err := findModRoot(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("finding module root: %w", err)
	}

	// Create a go.mod in the temp dir that requires and replaces the user's module
	goMod := fmt.Sprintf(`module hwgen_extractor

go 1.25.0

require github.com/bbsify-landed/heartwood v0.0.0

replace github.com/bbsify-landed/heartwood => %s
`, modRoot)

	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		return nil, fmt.Errorf("writing temp go.mod: %w", err)
	}

	// Copy go.sum if it exists to satisfy dependencies like clog
	if goSum, err := os.ReadFile(filepath.Join(modRoot, "go.sum")); err == nil {
		_ = os.WriteFile(filepath.Join(tmpDir, "go.sum"), goSum, 0o644)
	}

	// Also check if the user's package is in a different module
	// For now, we assume the user's schema package is within the heartwood module

	cmd := exec.Command("go", "run", ".")
	cmd.Dir = tmpDir
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running extractor: %w", err)
	}

	var defs map[string]*schema.Definition
	if err := json.Unmarshal(out, &defs); err != nil {
		return nil, fmt.Errorf("parsing extractor output: %w", err)
	}

	return defs, nil
}

// findModRoot walks up from dir looking for go.mod.
func findModRoot(dir string) (string, error) {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}
