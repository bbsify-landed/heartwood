PACKAGES := $(shell go list ./... | grep -v '/bin/')

.PHONY: generate
generate:
	go run ./cmd/hwgen/ ./cmd/hwgen/testdata/basic/
	go run ./cmd/hwgen/ ./bin/schema-example/schema/

.PHONY: test
test: generate
	go test $(PACKAGES) -timeout 60s

.PHONY: coverage
coverage: generate
	go test $(PACKAGES) -coverprofile=coverage.out -timeout 60s
	go-test-coverage --config .testcoverage.yml

.PHONY: vet
vet:
	go vet $(PACKAGES)

.PHONY: lint
lint: generate
	golangci-lint run ./cmd/... ./pkg/...
