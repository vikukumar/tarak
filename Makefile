##############################################################################
# Tarak Makefile
#
# Targets:
#   make build          – Build Next.js UI and all 5 binaries
#   make build-ui       – Build Next.js dashboard UI into internal/ui/dist
#   make test           – Run all Go tests
#   make test-race      – Run tests with race detector
#   make lint           – Run golangci-lint
#   make clean          – Remove build artifacts
#   make fmt            – Format all Go code
#   make vet            – Run go vet
#   make dev            – Build and run the server in dev mode
##############################################################################

.PHONY: all build build-ui build-tarak build-server build-agent build-cli test test-race test-cover lint clean fmt vet dev help

# ─── Build settings ──────────────────────────────────────────────────────────

VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "1.0.6")
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
DATE        := $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
MODULE      := github.com/vikukumar/tarak
LDFLAGS     := -s -w \
               -X $(MODULE)/internal/version.Version=$(VERSION) \
               -X $(MODULE)/internal/version.Commit=$(COMMIT) \
               -X $(MODULE)/internal/version.BuildDate=$(DATE) \
               -X $(MODULE)/internal/version.Author=vikukumar

BIN_DIR     := bin

all: build

# ─── Build UI ─────────────────────────────────────────────────────────────────

build-ui:
	@echo "Building Next.js Dashboard UI..."
	cd dashboard && npm run build
	mkdir -p internal/ui/dist
	rm -rf internal/ui/dist/*
	cp -r dashboard/out/* internal/ui/dist/
	@echo "Embedded UI assets ready in internal/ui/dist/"

# ─── Build Binaries ───────────────────────────────────────────────────────────

build: build-ui build-tarak build-server build-agent build-cli

build-tarak:
	@echo "Building tarak (all-in-one)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -trimpath -o $(BIN_DIR)/tarak ./cmd/tarak

build-server:
	@echo "Building tarakd (server daemon)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -trimpath -o $(BIN_DIR)/tarakd ./cmd/tarakd

build-agent:
	@echo "Building taraks (node worker agent)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -trimpath -o $(BIN_DIR)/taraks ./cmd/taraks

build-cli:
	@echo "Building tarakctl & taraktl (CLI)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -trimpath -o $(BIN_DIR)/tarakctl ./cmd/tarakctl
	go build -ldflags "$(LDFLAGS)" -trimpath -o $(BIN_DIR)/taraktl ./cmd/taraktl

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
		--data-dir ./data \
		--bind-address 127.0.0.1:6443 \
		--insecure \
		--log-level debug

# ─── Cleanup ──────────────────────────────────────────────────────────────────

clean:
	rm -rf $(BIN_DIR) dist coverage.out coverage.html

# ─── Help ─────────────────────────────────────────────────────────────────────

help:
	@echo "Available targets:"
	@echo "  build          Build Next.js UI and all 5 binaries"
	@echo "  build-ui       Build Next.js dashboard UI into internal/ui/dist"
	@echo "  build-tarak    Build unified tarak binary"
	@echo "  build-server   Build tarakd server binary"
	@echo "  build-agent    Build taraks worker agent binary"
	@echo "  build-cli      Build tarakctl and taraktl CLI binaries"
	@echo "  test           Run all tests"
	@echo "  test-race      Run tests with race detector"
	@echo "  test-cover     Run tests and generate HTML coverage report"
	@echo "  lint           Run golangci-lint"
	@echo "  fmt            Format Go code"
	@echo "  vet            Run go vet"
	@echo "  dev            Run the server in dev mode"
	@echo "  clean          Remove build artifacts"
