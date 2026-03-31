package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateBasicSchema(t *testing.T) {
	testDir := filepath.Join("testdata", "basic")
	outDir := t.TempDir()

	// Run the generator
	defs, pkgName, err := loadDefinitions(testDir)
	if err != nil {
		t.Fatalf("loadDefinitions: %v", err)
	}
	if len(defs) == 0 {
		t.Fatal("expected definitions, got none")
	}
	if pkgName != "basic" {
		t.Fatalf("expected package name 'basic', got %q", pkgName)
	}

	if err := generate(outDir, pkgName, defs); err != nil {
		t.Fatalf("generate: %v", err)
	}

	// Verify the generated files exist
	for _, name := range []string{"hw_gen.go", "hw_client_gen.go"} {
		path := filepath.Join(outDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected generated file %s to exist", name)
		}
	}

	// To verify the generated package compiles, we need to copy the original
	// schema.go to the same directory as the generated files.
	schemaSrc, err := os.ReadFile(filepath.Join(testDir, "schema.go"))
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(outDir, "schema.go"), schemaSrc, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// We also need a go.mod file in the outDir so it's a valid module
	// that can find the heartwood dependency.
	absDir, _ := filepath.Abs(".")
	modRoot, err := findModRoot(absDir)
	if err != nil {
		t.Fatalf("findModRoot: %v", err)
	}

	goMod := fmt.Sprintf("module test\n\ngo 1.25.0\n\nrequire github.com/bbsify-landed/heartwood v0.0.0\n\nreplace github.com/bbsify-landed/heartwood => %s\n", modRoot)
	err = os.WriteFile(filepath.Join(outDir, "go.mod"), []byte(goMod), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Copy go.sum to avoid "updates to go.mod needed"
	goSum, err := os.ReadFile(filepath.Join(modRoot, "go.sum"))
	if err == nil {
		_ = os.WriteFile(filepath.Join(outDir, "go.sum"), goSum, 0o644)
	}

	cmd := exec.Command("go", "build", ".")
	cmd.Dir = outDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated code does not compile: %v\n%s", err, out)
	}

	// Verify expected definitions were found
	if _, ok := defs["HealthCheck"]; !ok {
		t.Error("expected HealthCheck definition")
	}
	if _, ok := defs["CreateUser"]; !ok {
		t.Error("expected CreateUser definition")
	}
	if _, ok := defs["ComplexValidation"]; !ok {
		t.Error("expected ComplexValidation definition")
	}

	// Verify HealthCheck definition
	hc := defs["HealthCheck"]
	if hc.Method != "POST" {
		t.Errorf("HealthCheck.Method = %q, want POST", hc.Method)
	}
	if hc.Path != "/health" {
		t.Errorf("HealthCheck.Path = %q, want /health", hc.Path)
	}
	if len(hc.ReqFields) != 1 {
		t.Errorf("HealthCheck.ReqFields has %d fields, want 1", len(hc.ReqFields))
	}

	// Verify CreateUser definition
	cu := defs["CreateUser"]
	if len(cu.ReqFields) != 3 {
		t.Errorf("CreateUser.ReqFields has %d fields, want 3", len(cu.ReqFields))
	}
	if len(cu.ResFields) != 4 {
		t.Errorf("CreateUser.ResFields has %d fields, want 4", len(cu.ResFields))
	}
}

func TestLoadDefinitions_Errors(t *testing.T) {
	t.Run("Invalid Directory", func(t *testing.T) {
		_, _, err := loadDefinitions("/non/existent/path")
		assert.Error(t, err)
	})

	t.Run("No Go Files", func(t *testing.T) {
		tmpDir, _ := os.MkdirTemp("", "hwgen-test-*")
		defer os.RemoveAll(tmpDir)
		_, _, err := loadDefinitions(tmpDir)
		assert.Error(t, err)
	})

	t.Run("Compilation Error", func(t *testing.T) {
		tmpDir, _ := os.MkdirTemp("", "hwgen-test-*")
		defer os.RemoveAll(tmpDir)
		err := os.WriteFile(filepath.Join(tmpDir, "schema.go"), []byte("package bad\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(undefined) }"), 0o644)
		assert.NoError(t, err)
		_, _, err = loadDefinitions(tmpDir)
		assert.Error(t, err)
	})

	t.Run("No Definitions", func(t *testing.T) {
		// Instead of making a full new module which is hard to point back to heartwood,
		// just make a subdirectory in the current module.
		subDir := filepath.Join("testdata", "empty")
		_ = os.MkdirAll(subDir, 0o755)
		defer os.RemoveAll(subDir)
		err := os.WriteFile(filepath.Join(subDir, "schema.go"), []byte("package empty\n\ntype Foo struct{}"), 0o644)
		assert.NoError(t, err)

		defs, _, err := loadDefinitions(subDir)
		assert.NoError(t, err)
		assert.Empty(t, defs)
	})
}

func TestRun(t *testing.T) {
	testDir := filepath.Join("testdata", "basic")
	outDir := t.TempDir()

	// Copy schema.go to outDir so it can be loaded
	schemaSrc, err := os.ReadFile(filepath.Join(testDir, "schema.go"))
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(outDir, "schema.go"), schemaSrc, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// We also need a go.mod file in the outDir
	absDir, _ := filepath.Abs(".")
	modRoot, err := findModRoot(absDir)
	if err != nil {
		t.Fatalf("findModRoot: %v", err)
	}

	goMod := fmt.Sprintf("module testrun\n\ngo 1.25.0\n\nrequire github.com/bbsify-landed/heartwood v0.0.0\n\nreplace github.com/bbsify-landed/heartwood => %s\n", modRoot)
	err = os.WriteFile(filepath.Join(outDir, "go.mod"), []byte(goMod), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// Copy go.sum
	goSum, err := os.ReadFile(filepath.Join(modRoot, "go.sum"))
	if err == nil {
		_ = os.WriteFile(filepath.Join(outDir, "go.sum"), goSum, 0o644)
	}

	var stdout, stderr bytes.Buffer

	// Success
	err = run([]string{"hwgen", outDir}, &stdout, &stderr)
	assert.NoError(t, err)
	assert.Contains(t, stderr.String(), "generated 3 endpoint(s)")

	// Error - no definitions
	subDir := filepath.Join("testdata", "empty_run")
	_ = os.MkdirAll(subDir, 0o755)
	defer os.RemoveAll(subDir)
	err = os.WriteFile(filepath.Join(subDir, "schema.go"), []byte("package empty\n\ntype Foo struct{}"), 0o644)
	assert.NoError(t, err)

	err = run([]string{"hwgen", subDir}, &stdout, &stderr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no schema.Definition variables found")
}
