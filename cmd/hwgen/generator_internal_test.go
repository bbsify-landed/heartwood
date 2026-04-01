package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bbsify-landed/heartwood/pkg/schema"
	"github.com/stretchr/testify/assert"
)

func TestGoType(t *testing.T) {
	tests := []struct {
		field schema.Field
		want  string
	}{
		{schema.Field{Type: schema.StringType}, "string"},
		{schema.Field{Type: schema.IntType}, "int"},
		{schema.Field{Type: schema.Int64Type}, "int64"},
		{schema.Field{Type: schema.Float64Type}, "float64"},
		{schema.Field{Type: schema.BoolType}, "bool"},
		{schema.Field{Type: schema.RefType, RefName: "User"}, "*UserMessage"},
		{schema.Field{Type: schema.SliceType, ElemType: schema.StringType}, "[]string"},
		{schema.Field{Type: schema.SliceRefType, RefName: "User"}, "[]*UserMessage"},
		{schema.Field{Type: schema.FieldType(99)}, "any"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, goType(tt.field))
	}
}

func TestScalarGoType(t *testing.T) {
	tests := []struct {
		ft   schema.FieldType
		want string
	}{
		{schema.StringType, "string"},
		{schema.IntType, "int"},
		{schema.Int64Type, "int64"},
		{schema.Float64Type, "float64"},
		{schema.BoolType, "bool"},
		{schema.FieldType(99), "any"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, scalarGoType(tt.ft))
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"foo", "Foo"},
		{"foo_bar", "FooBar"},
		{"foo_bar_baz", "FooBarBaz"},
		{"_foo", "Foo"},
		{"foo_", "Foo"},
		{"foo__bar", "FooBar"},
		{"", ""},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, toPascalCase(tt.in))
	}
}

func TestWriteTemplate_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_gen.go")

	data := templateData{Package: "test"}

	// Invalid template
	err := writeTemplate(tmpFile, "{{ invalid", data)
	assert.Error(t, err)

	// Formatting error
	err = writeTemplate(tmpFile, "package test \n invalid go code", data)
	assert.Error(t, err)
	_, statErr := os.Stat(tmpFile)
	assert.NoError(t, statErr, "should have written unformatted file")
}
