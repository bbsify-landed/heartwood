package schema_test

import (
	"testing"

	"github.com/bbsify-landed/heartwood/pkg/schema"
	"github.com/stretchr/testify/assert"
)

func TestDefineBasicEndpoint(t *testing.T) {
	def := schema.Define(
		schema.POST("/health"),
		schema.Request(
			schema.String("are_you_healthy"),
		),
		schema.Response(
			schema.String("healthy"),
		),
	)

	assert.Equal(t, "POST", def.Method)
	assert.Equal(t, "/health", def.Path)
	assert.Len(t, def.ReqFields, 1)
	assert.Equal(t, "are_you_healthy", def.ReqFields[0].Name)
	assert.Equal(t, schema.StringType, def.ReqFields[0].Type)
	assert.Len(t, def.ResFields, 1)
	assert.Equal(t, "healthy", def.ResFields[0].Name)
}

func TestDefineAllHTTPMethods(t *testing.T) {
	tests := []struct {
		opt    schema.Option
		method string
	}{
		{schema.GET("/a"), "GET"},
		{schema.POST("/b"), "POST"},
		{schema.PUT("/c"), "PUT"},
		{schema.DELETE("/d"), "DELETE"},
		{schema.PATCH("/e"), "PATCH"},
	}

	for _, tt := range tests {
		def := schema.Define(tt.opt)
		assert.Equal(t, tt.method, def.Method)
	}
}

func TestFieldTypes(t *testing.T) {
	def := schema.Define(
		schema.POST("/test"),
		schema.Request(
			schema.String("s"),
			schema.Int("i"),
			schema.Int64("i64"),
			schema.Float64("f64"),
			schema.Bool("b"),
		),
	)

	assert.Len(t, def.ReqFields, 5)
	assert.Equal(t, schema.StringType, def.ReqFields[0].Type)
	assert.Equal(t, schema.IntType, def.ReqFields[1].Type)
	assert.Equal(t, schema.Int64Type, def.ReqFields[2].Type)
	assert.Equal(t, schema.Float64Type, def.ReqFields[3].Type)
	assert.Equal(t, schema.BoolType, def.ReqFields[4].Type)
}

func TestFieldValidationChaining(t *testing.T) {
	f := schema.String("name").Required().MinLength(1).MaxLength(100)

	assert.True(t, f.IsRequired)
	assert.NotNil(t, f.MinLen)
	assert.Equal(t, 1, *f.MinLen)
	assert.NotNil(t, f.MaxLen)
	assert.Equal(t, 100, *f.MaxLen)
}

func TestNumericValidation(t *testing.T) {
	f := schema.Int("age").MinValue(0).MaxValue(150)

	assert.NotNil(t, f.Min)
	assert.Equal(t, float64(0), *f.Min)
	assert.NotNil(t, f.Max)
	assert.Equal(t, float64(150), *f.Max)
}

func TestRefField(t *testing.T) {
	address := schema.Define(
		schema.POST("/address"),
		schema.Request(
			schema.String("street"),
			schema.String("city"),
		),
	)
	address.Name = "Address"

	def := schema.Define(
		schema.POST("/user"),
		schema.Request(
			schema.String("name"),
			schema.Ref("address", address),
		),
	)

	assert.Len(t, def.ReqFields, 2)
	assert.Equal(t, schema.RefType, def.ReqFields[1].Type)
	assert.Equal(t, "Address", def.ReqFields[1].RefName)
}

func TestSliceField(t *testing.T) {
	f := schema.Slice("tags", schema.StringType)
	assert.Equal(t, schema.SliceType, f.Type)
	assert.Equal(t, schema.StringType, f.ElemType)
}

func TestSliceOfField(t *testing.T) {
	item := schema.Define(schema.POST("/item"))
	item.Name = "Item"

	f := schema.SliceOf("items", item)
	assert.Equal(t, schema.SliceRefType, f.Type)
	assert.Equal(t, "Item", f.RefName)
}
