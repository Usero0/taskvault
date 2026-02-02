.PHONY: help build test clean fmt lint run-example init-demo

help:
	@echo "TaskVault Development Commands"
	@echo "=============================="
	@echo ""
	@echo "Setup & Build:"
	@echo "  make init-demo      Initialize demo environment with config"
	@echo "  make build          Build CLI binary (taskvault or taskvault.exe)"
	@echo "  make clean          Remove build artifacts and cache"
	@echo ""
	@echo "Testing & Quality:"
	@echo "  make test           Run all tests with coverage"
	@echo "  make test-race      Run tests with race detector"
	@echo "  make fmt            Format all Go code"
	@echo "  make lint           Run linter (requires golangci-lint)"
	@echo ""
	@echo "Examples & Demos:"
	@echo "  make run-example    Run example code with cached results"
	@echo "  make demo           Full demo: save → retrieve → stats"
	@echo ""
	@echo "Utilities:"
	@echo "  make deps           Download all dependencies"
	@echo "  make watch          Watch for changes and rebuild (requires entr)"
	@echo "  make stats          Show cache statistics"
	@echo "  make audit-log      Tail audit log (follow mode)"

build:
	@echo "🔨 Building TaskVault..."
	@go build -o taskvault ./cmd/taskvault
	@echo "✓ Built: ./taskvault"

test:
	@echo "🧪 Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -5
	@echo "✓ Coverage report: coverage.out"

test-race:
	@echo "🏃 Running race detector..."
	@go test -race ./...

fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✓ Formatted"

lint:
	@echo "🔍 Linting..."
	@golangci-lint run ./... --timeout=5m

deps:
	@echo "📦 Downloading dependencies..."
	@go mod download
	@echo "✓ Dependencies ready"

clean:
	@echo "🧹 Cleaning..."
	@rm -rf taskvault taskvault.exe
	@rm -rf .taskvault/cache .taskvault/*.log
	@rm -f coverage.out
	@rm -f example_input.txt example_output.txt
	@go clean -testcache
	@echo "✓ Cleaned"

init-demo:
	@echo "📋 Initializing demo..."
	@mkdir -p .taskvault
	@./taskvault init || go run ./cmd/taskvault init
	@echo "✓ Demo initialized at .taskvault/config.yaml"

run-example: build init-demo
	@echo "▶️  Running examples..."
	@go run ./cmd/examples/main.go

demo: build init-demo
	@echo "🎬 Full Demo"
	@echo "============"
	@echo ""
	@echo "Step 1: Creating sample input..."
	@echo "test input data v1" > demo_input.txt
	@echo ""
	@echo "Step 2: Saving to cache..."
	@./taskvault cache save demo_task demo_input.txt demo_output.txt
	@echo ""
	@echo "Step 3: Retrieving from cache (should be instant HIT)..."
	@echo "test input data v1" > demo_input2.txt
	@./taskvault cache get demo_task demo_input2.txt demo_output2.txt
	@echo ""
	@echo "Step 4: Cache statistics..."
	@./taskvault cache stats
	@echo ""
	@echo "✓ Demo complete!"
	@rm -f demo_input.txt demo_input2.txt demo_output.txt demo_output2.txt

stats: build
	@./taskvault cache stats

audit-log:
	@if [ -f .taskvault/cache/audit.log ]; then \
		tail -f .taskvault/cache/audit.log; \
	else \
		echo "Audit log not found. Run 'make demo' first."; \
	fi

watch:
	@command -v entr >/dev/null 2>&1 || { echo "entr not found. Install: brew install entr"; exit 1; }
	@find . -name "*.go" ! -path "./.git/*" | entr -r make build
