package schema

import "github.com/bbsify-landed/heartwood/pkg/schema"

//go:generate go run github.com/bbsify-landed/heartwood/cmd/hwgen

var HealthCheck = schema.Define(
	schema.POST("/health"),
	schema.Request(
		schema.String("are_you_healthy").Required(),
	),
	schema.Response(
		schema.String("healthy"),
	),
)
