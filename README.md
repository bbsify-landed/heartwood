# heartwood

A lightweight Go HTTP framework for type-safe request/response handlers using generics.

## Install

```
go get github.com/bbsify-landed/heartwood
```

## Usage

Define request/response types, then register handlers with concrete types — heartwood handles deserialization, validation, and serialization.

```go
app := hw.New()

hw.Use(app, "POST", "/greet", func(ctx context.Context, req *GreetRequest) (*GreetResponse, error) {
    return &GreetResponse{Message: "Hello, " + req.Name + "!"}, nil
})

hw.ListenAndServe(app, ctx, ":9000")
```

Request types implement `Deserializable` (decode + validate), response types implement `Serializable` (encode). See [examples/](examples/) for complete working code.

## Code Generation

The `hwgen` tool generates typed request/response structs, validation, handler registration, and HTTP clients from declarative schema definitions.

### Define a schema

```go
package myapi

import "github.com/bbsify-landed/heartwood/pkg/schema"

//go:generate go run github.com/bbsify-landed/heartwood/cmd/hwgen

var HealthCheck = schema.Define(
    schema.POST("/health"),
    schema.Request(
        schema.String("name").Required().MinLength(1),
        schema.Int("age").MinValue(0).MaxValue(150),
    ),
    schema.Response(
        schema.String("status"),
    ),
)
```

### Generate

```
go generate ./myapi
```

This produces two files:

- `hw_gen.go` — `HealthCheckRequest` / `HealthCheckResponse` structs with `Serialize`, `Deserialize`, `Validate` methods, plus a `RegisterHealthCheck` function
- `hw_client_gen.go` — a typed `Client` with a `HealthCheck(ctx, req)` method

### Use the generated code

```go
// Server
app := hw.New()
myapi.RegisterHealthCheck(app, func(ctx context.Context, req *myapi.HealthCheckRequest) (*myapi.HealthCheckResponse, error) {
    return &myapi.HealthCheckResponse{Status: "ok"}, nil
})

// Client
client := myapi.NewClient("http://localhost:9000")
res, err := client.HealthCheck(ctx, &myapi.HealthCheckRequest{Name: "alice", Age: 30})
```

### Available field types

| Builder | Go type | Constraints |
|---------|---------|-------------|
| `String(name)` | `string` | `Required`, `MinLength`, `MaxLength` |
| `Int(name)` | `int` | `Required`, `MinValue`, `MaxValue` |
| `Int64(name)` | `int64` | `Required`, `MinValue`, `MaxValue` |
| `Float64(name)` | `float64` | `Required`, `MinValue`, `MaxValue` |
| `Bool(name)` | `bool` | `Required` |
| `Ref(name, def)` | `*T` | — |
| `Slice(name, elemType)` | `[]T` | `MinLength`, `MaxLength` |
| `SliceOf(name, def)` | `[]*T` | `MinLength`, `MaxLength` |

## License

[MIT](LICENSE)
