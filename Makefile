build:
	mkdir -p bin
	go build -o bin/boids main.go

run:
	go run .

test:
	go test -race -vet=off ./...

LINT_PKG := github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4
lint:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install ${LINT_PKG}; \
	fi
	@echo "Running golangci-lint..."
	@golangci-lint run
