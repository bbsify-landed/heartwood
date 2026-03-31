package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const goModTmpl = `module {{.Name}}

go 1.25.0

require github.com/bbsify-landed/heartwood v0.0.0
replace github.com/bbsify-landed/heartwood => {{.ModRoot}}
`

func writeTestGoMod(t *testing.T, dir string, name string) {
	t.Helper()
	absPath, err := filepath.Abs(".")
	require.NoError(t, err)
	modRoot, err := findModRoot(absPath)
	require.NoError(t, err)

	tmpl, err := template.New("gomod").Parse(goModTmpl)
	require.NoError(t, err)

	f, err := os.Create(filepath.Join(dir, "go.mod"))
	require.NoError(t, err)
	defer f.Close()

	err = tmpl.Execute(f, map[string]string{
		"Name":    name,
		"ModRoot": modRoot,
	})
	require.NoError(t, err)

	// Copy go.sum if it exists
	if goSum, err := os.ReadFile(filepath.Join(modRoot, "go.sum")); err == nil {
		err = os.WriteFile(filepath.Join(dir, "go.sum"), goSum, 0o644)
		require.NoError(t, err)
	}
}

func TestGenerateBasicSchema(t *testing.T) {
	testDir := filepath.Join("testdata", "basic")

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

	// Create temp dir inside testdata to inherit heartwood module context
	outDir := filepath.Join("testdata", "tmp-gen")
	_ = os.MkdirAll(outDir, 0o755)
	defer os.RemoveAll(outDir)

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

	writeTestGoMod(t, outDir, "test")

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = outDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, out)
	}

	cmd = exec.Command("go", "build", ".")
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
		tmpDir := t.TempDir()
		_, _, err := loadDefinitions(tmpDir)
		assert.Error(t, err)
	})

	t.Run("Compilation Error", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tmpDir, "schema.go"), []byte("package bad\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(undefined) }"), 0o644)
		assert.NoError(t, err)
		_, _, err = loadDefinitions(tmpDir)
		assert.Error(t, err)
	})

	t.Run("No Definitions", func(t *testing.T) {
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
	// Create temp dir inside testdata to inherit heartwood module context
	outDir := filepath.Join("testdata", "tmp-run")
	_ = os.MkdirAll(outDir, 0o755)
	defer os.RemoveAll(outDir)

	// Copy schema.go to outDir so it can be loaded
	schemaSrc, err := os.ReadFile(filepath.Join(testDir, "schema.go"))
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(outDir, "schema.go"), schemaSrc, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	writeTestGoMod(t, outDir, "testrun")

	var stdout, stderr bytes.Buffer
	ctx := t.Context()

	// Success
	err = run(ctx, []string{"hwgen", outDir}, &stdout, &stderr)
	assert.NoError(t, err)
	// clog outputs to stderr by default, but we might need to capture it differently
	// or just check that it didn't return an error since clog output is harder to capture
	// when it's global. But let's see if we can just assert success.

	// Error - no definitions
	subDir := filepath.Join("testdata", "empty_run")
	_ = os.MkdirAll(subDir, 0o755)
	defer os.RemoveAll(subDir)
	err = os.WriteFile(filepath.Join(subDir, "schema.go"), []byte("package empty\n\ntype Foo struct{}"), 0o644)
	assert.NoError(t, err)

	err = run(ctx, []string{"hwgen", subDir}, &stdout, &stderr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no schema.Definition variables found")
}
