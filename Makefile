##############################################################################
# Tarak Makefile
#
# Targets:
#   make build          – Build tarakd and tarakctl
#   make test           – Run all tests
#   make test-race      – Run tests with race detector
#   make lint           – Run golangci-lint
#   make clean          – Remove build artifacts
#   make fmt            – Format all Go code
#   make vet            – Run go vet
#   make dev            – Build and run the server in insecure dev mode
##############################################################################

.PHONY: all build build-server build-cli test test-race test-cover lint clean fmt vet dev

# ─── Build settings ──────────────────────────────────────────────────────────

VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "1.0.0")
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
DATE        := $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
MODULE      := github.com/vikukumar/tarak
LDFLAGS     := -s -w \
               -X $(MODULE)/internal/version.Version=$(VERSION) \
               -X $(MODULE)/internal/version.Commit=$(COMMIT) \
               -X $(MODULE)/internal/version.BuildDate=$(DATE) \
               -X $(MODULE)/internal/version.Author=vikukumar

BIN_DIR     := bin

.PHONY: all build build-tarak build-server build-agent build-cli test test-race test-cover lint clean fmt vet dev

all: build

# ─── Build ────────────────────────────────────────────────────────────────────

build: build-tarak build-server build-agent build-cli

build-tarak:
	@echo "Building tarak (all-in-one)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/tarak ./cmd/tarak

build-server:
	@echo "Building tarakd (server daemon)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/tarakd ./cmd/tarakd

build-agent:
	@echo "Building taraks (node worker agent)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/taraks ./cmd/taraks

build-cli:
	@echo "Building tarakctl & taraktl (CLI)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/tarakctl ./cmd/tarakctl
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/taraktl ./cmd/taraktl

# ─── Test ─────────────────────────────────────────────────────────────────────

test:
	go test ./... -v -timeout 120s

test-race:
	go test ./... -race -v -timeout 120s

test-cover:
	go test ./... -coverprofile=coverage.out -covermode=atomic -timeout 120s
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ─── Code quality ─────────────────────────────────────────────────────────────

lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt:
	gofmt -w -s .
	@which goimports > /dev/null && goimports -w . || true

vet:
	go vet ./...

# ─── Dev mode ─────────────────────────────────────────────────────────────────

dev: build-server
	@echo "Starting Tarak API server in dev mode..."
	$(BIN_DIR)/tarakd server \
		--data-dir ./dev-data \
		--bind-address 0.0.0.0:6443 \
		--insecure \
		--log-level debug

# ─── Cleanup ──────────────────────────────────────────────────────────────────

clean:
	rm -rf $(BIN_DIR) coverage.out coverage.html

# ─── Help ─────────────────────────────────────────────────────────────────────

help:
	@echo "Available targets:"
	@echo "  build          Build tarakd and tarakctl"
	@echo "  build-server   Build only the tarakd server binary"
	@echo "  build-cli      Build only the tarakctl CLI binary"
	@echo "  test           Run all tests"
	@echo "  test-race      Run tests with race detector"
	@echo "  test-cover     Run tests and generate HTML coverage report"
	@echo "  lint           Run golangci-lint"
	@echo "  fmt            Format Go code"
	@echo "  vet            Run go vet"
	@echo "  dev            Run the server in dev mode"
	@echo "  clean          Remove build artifacts"
