# Windows defaults to the embedded raylib DLL, which needs no C compiler.
# The race detector needs cgo. Override with: make test TAGS= RACE=-race
ifeq ($(OS),Windows_NT)
TAGS ?= -tags nocgo
else
RACE ?= -race
endif

build:
	go build $(TAGS) -o bin/ .

run:
	go run $(TAGS) .

test:
	go test $(TAGS) $(RACE) -v -vet=off ./...

LINT_VERSION := v2.15
LINT_PKG := github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(LINT_VERSION)
lint:
	@golangci-lint version >/dev/null 2>&1 || { echo "Installing golangci-lint..."; go install ${LINT_PKG}; }
	@echo "Found golangci-lint, running..."
	golangci-lint run
