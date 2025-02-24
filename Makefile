build:
	mkdir -p bin
	go build -o bin/boids main.go

run:
	go run .

LINT_PKG := github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64
LINT_BIN := $(shell go env GOPATH)/bin/golangci-lint
lint:
	@if \[ ! -f ${LINT_BIN} \]; then \
		echo "Installing golangci-lint..."; \
    go install ${LINT_PKG}; \
  fi
	@if \[ -f ${LINT_BIN} \]; then \
  	echo "Found golangci-lint at '$(LINT_BIN)', running..."; \
    ${LINT_BIN} run; \
	else \
    echo "golangci-lint not found or the file does not exist"; \
    exit 1; \
  fi
