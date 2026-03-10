.PHONY: build run release clean test

# Binary name
BINARY_NAME=govectord

# Default target
all: build

build:
	@echo "🛠️  Building $(BINARY_NAME) to bin/ ..."
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/$(BINARY_NAME) ./cmd/govector-server
	@echo "✅ Build complete: bin/$(BINARY_NAME)"

run:
	@echo "🚀 Starting GoVector Server..."
	go run ./cmd/govector-server/main.go -port 18080 -db ./govector.db -hnsw=true

release:
	@echo "📦 Building cross-platform release packages..."
	@chmod +x scripts/build_release.sh
	./scripts/build_release.sh v0.1.0

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
