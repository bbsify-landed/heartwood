package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGenerateBasicSchema(t *testing.T) {
	testDir := filepath.Join("testdata", "basic")

	// Clean old generated files
	for _, name := range []string{"hw_gen.go", "hw_client_gen.go"} {
		_ = os.Remove(filepath.Join(testDir, name))
	}

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

	if err := generate(testDir, pkgName, defs); err != nil {
		t.Fatalf("generate: %v", err)
	}

	// Verify the generated files exist
	for _, name := range []string{"hw_gen.go", "hw_client_gen.go"} {
		path := filepath.Join(testDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected generated file %s to exist", name)
		}
	}

	// Verify the generated package compiles
	absDir, _ := filepath.Abs(testDir)
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = absDir

	// Need to set GOPATH context by running from module root
	modRoot, err := findModRoot(absDir)
	if err != nil {
		t.Fatalf("findModRoot: %v", err)
	}
	cmd = exec.Command("go", "build", "./cmd/hwgen/testdata/basic/")
	cmd.Dir = modRoot
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
