build:
	mkdir -p bin
	go build -o bin/boids .

run:
	go run .

test:
	go test -v -race ./...

LINT_VERSION := v2.9
LINT_PKG := github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(LINT_VERSION)
lint:
	@golangci-lint version >/dev/null 2>&1 || { echo "Installing golangci-lint..."; go install ${LINT_PKG}; }
	@echo "Found golangci-lint, running..."
	golangci-lint run
