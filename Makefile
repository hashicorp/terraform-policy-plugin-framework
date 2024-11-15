
.PHONY: test
test:
	@echo "Running tests..."
	@go test ./...

.PHONY: generate
generate:
	@echo "Generating code..."
	@go generate ./...

.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run

.PHONY: protobuf
protobuf:
	cd tools && go run ./protobuf-compile .

.PHONY: all
all: test generate lint protobuf
