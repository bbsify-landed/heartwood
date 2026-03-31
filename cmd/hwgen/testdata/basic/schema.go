package basic

import "github.com/bbsify-landed/heartwood/pkg/schema"

var HealthCheck = schema.Define(
	schema.POST("/health"),
	schema.Request(
		schema.String("are_you_healthy").Required(),
	),
	schema.Response(
		schema.String("healthy"),
	),
)

var CreateUser = schema.Define(
	schema.POST("/users"),
	schema.Request(
		schema.String("name").Required().MinLength(1).MaxLength(100),
		schema.String("email").Required(),
		schema.Int("age").MinValue(0).MaxValue(150),
	),
	schema.Response(
		schema.String("id"),
		schema.String("name"),
		schema.String("email"),
		schema.Int("age"),
	),
)

var ComplexValidation = schema.Define(
	schema.POST("/complex"),
	schema.Request(
		schema.String("s_req").Required().MinLength(5).MaxLength(10),
		schema.Int("i_val").MinValue(10).MaxValue(20),
		schema.Float64("f_val").MinValue(1.5).MaxValue(2.5),
		schema.Slice("sl_val", schema.StringType).Required().MinLength(1).MaxLength(3),
	),
	schema.Response(
		schema.Bool("success"),
	),
)
