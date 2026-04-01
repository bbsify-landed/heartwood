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

The `hwgen` tool generates typed handlers and HTTP clients from declarative schema definitions:

```go
var HealthCheck = &schema.Definition{
    Method: "POST",
    Path:   "/health",
    ReqFields: []schema.Field{schema.String("name").Required()},
    ResFields: []schema.Field{schema.String("status")},
}
```

```
go run github.com/bbsify-landed/heartwood/cmd/hwgen ./path/to/schema
```

## License

[MIT](LICENSE)
