.PHONY: build run release clean test

# Binary name
BINARY_NAME=govector

# Default target
all: build

build:
	@echo "🛠️  Building $(BINARY_NAME) to bin/ ..."
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/$(BINARY_NAME) ./cmd/govector
	@echo "✅ Build complete: bin/$(BINARY_NAME)"

run:
	@echo "🚀 Starting GoVector Server..."
	go run ./cmd/govector/main.go serve -port 18080 -db ./govector.db

release:
	@echo "📦 Building cross-platform release packages..."
	@chmod +x scripts/build_release.sh
	./scripts/build_release.sh $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.3")

test:
	@echo "🧪 Running benchmarks and tests..."
	go run cmd/bench/main.go

clean:
	@echo "🧹 Cleaning up..."
	@rm -rf bin/
	@rm -rf dist/
	@rm -f govector.db
	@rm -f server.log
	@echo "✅ Clean complete."
