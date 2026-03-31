PACKAGES := $(shell go list ./... | grep -v '/bin/')

.PHONY: test
test:
	go test $(PACKAGES) -timeout 60s

.PHONY: coverage
coverage:
	go test $(PACKAGES) -coverprofile=coverage.out -timeout 60s
	go tool cover -func=coverage.out

.PHONY: vet
vet:
	go vet $(PACKAGES)

.PHONY: generate
generate:
	go run ./cmd/hwgen/ ./cmd/hwgen/testdata/basic/
	go run ./cmd/hwgen/ ./bin/schema-example/schema/
