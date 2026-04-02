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

var GetTime = schema.Define(
	schema.GET("/time"),
	schema.Response(
		schema.String("time"),
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
