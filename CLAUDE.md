# Heartwood

## Build & Test

Use the Makefile for all build, test, and lint commands — don't run `go test` or `go vet` directly.

```
make test        # generate + run tests
make coverage    # generate + run tests with coverage thresholds
make vet         # go vet
make lint        # golangci-lint
make generate    # run hwgen code generation
```
