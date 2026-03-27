<div align="center">
  <h1>🎯 GoVector</h1>
  <p><b>The Lightweight, Embeddable Vector Database in Pure Go. (Think SQLite for Vectors)</b></p>

  [![Go Reference](https://pkg.go.dev/badge/github.com/DotNetAge/govector.svg)](https://pkg.go.dev/github.com/DotNetAge/govector)
  [![Go Version](https://img.shields.io/github/go-mod/go-version/DotNetAge/govector)](https://golang.org/)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
  [![Go Report Card](https://goreportcard.com/badge/github.com/DotNetAge/govector)](https://goreportcard.com/report/github.com/DotNetAge/govector)
  [![codecov](https://codecov.io/gh/DotNetAge/govector/graph/badge.svg?token=placeholder)](https://codecov.io/gh/DotNetAge/govector)

  <p>
    <a href="README.md">English</a> | <a href="README_zh.md">简体中文</a>
  </p>
</div>

In the era of Local AI, desktop applications, and edge computing, you don't always need a heavy, distributed vector cluster like Milvus or Qdrant. 

**GoVector** is a high-performance, embedded vector search engine written entirely in Go. It offers **Qdrant-compatible** API endpoints, **HNSW** indexing for blazing-fast Approximate Nearest Neighbor (ANN) search, and persistent local storage via **BoltDB**.

---

## 🛡️ Status & Performance Report

GoVector has been audited and recognized as a **"Best-in-Class"** embedded vector database. It achieves industrial-grade reliability with over 92% test coverage and sub-millisecond latency at scale.

![GoVector Status Report](assets/govector-report.png)

---

## ✨ Features

- 🚀 **Pure Go & CGO-Free**: Cross-compile to anywhere (Windows, macOS, Linux, edge devices) without messy C/C++ dependencies.
- 🛠️ **Unified CLI & Rich TUI**: A single `govector` binary for all operations (`serve`, `upsert`, `search`, `ls`, `rm`). Includes a beautiful, green-themed interactive Terminal User Interface (TUI) out of the box.
- ⚡ **High Performance**: Optimized for single-node performance, supporting millions of vectors with sub-millisecond search latency.
- 🧠 **HNSW Indexing**: Industrial-grade graph-based index for $O(\log N)$ search complexity. Enabled by default with auto-creation.
- 💾 **Protobuf & BoltDB**: Ultra-fast persistence using Protocol Buffers and bbolt. Data survives restarts with automatic collection discovery.
- 🔍 **Advanced Filtering**: Support for payload filtering (Exact, Range, Prefix, Regex, Contains) just like Qdrant.
- 📉 **SQ8 Quantization**: Built-in 8-bit scalar quantization to reduce disk footprint for large-scale data.
- 🛡️ **Reliability**: Over **92% test coverage** with nanosecond-precision versioning, storage-first consistency, and safe collection management (e.g., protecting the "default" collection).
- 🔌 **Dual Modes**: 
  - **Embedded Library**: Import it into your Go backend/desktop app with zero network overhead.
  - **Standalone Server / CLI**: Run it as a lightweight microservice with a Qdrant-compatible REST API, or manage data directly via the CLI.

---

## 📦 Installation & Quick Start

### Option A: Install via Homebrew (Mac/Linux)
The easiest way to install GoVector and run it as a service:
```bash
brew tap DotNetAge/govector
brew install govector

# Start the background service (starts the REST API server)
brew services start govector
```

### Option B: Build from Source
```bash
git clone https://github.com/DotNetAge/govector.git
cd govector
make build
# The binary will be available at ./bin/govector
```

### Option C: Use as a Go Library
```bash
go get github.com/DotNetAge/govector/core
```

---

## 🛠️ CLI & TUI Usage

GoVector comes with a powerful Command-Line Interface (CLI) and an interactive Terminal User Interface (TUI).

Run the binary without arguments to enter the **Interactive TUI**:
```bash
govector
```
*(Provides a beautiful, green-themed environment with a `/?` help menu, auto-creation of DBs/collections, and graceful `serve` interruption).*

Or use **Direct Commands**:
```bash
# General Syntax
govector <command> [dbfile] [options]

# Start the Qdrant-compatible REST API server
govector serve mydata.db -port 18080

# Insert data (auto-creates DB and HNSW-enabled collection if missing)
govector upsert mydata.db -c documents -j '{"id":"1", "vector":[0.1, 0.2], "payload":{"tag":"A"}}'

# Search data (with existence validation)
govector search mydata.db -c documents -v 0.1,0.2 -l 5

# List all collections
govector ls mydata.db

# Get point count in a collection
govector count mydata.db -c documents

# Delete a collection (safeguards protect the "default" collection)
govector rm mydata.db -c documents
```

---

## 🚀 Benchmark Performance (Large Scale)

Measured on a standard machine with 16GB RAM, 128-dimensional vectors.

| Index | Scale (N) | Build Time | Latency (Avg) | Throughput (QPS) | Memory (Alloc) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Flat** | 100K | 186 ms | 54.46 ms | 18 QPS | 59 MB |
| **HNSW** | 100K | 20.9 s | **0.08 ms** | **11,812 QPS** | 311 MB |
| **HNSW** | 1M | 4m 17s | **0.11 ms** | **8,709 QPS** | 3.32 GB |

> *Note: HNSW maintained sub-millisecond latency even at 1 million scale, providing 480x speedup over Flat index at 100K.*

---

## 💻 Usage Mode 1: Embedded Go Library (Zero Network)

```go
package main

import (
        "fmt"
        "github.com/DotNetAge/govector/core"
)

func main() {
        // 1. Initialize local storage (creates a single .db file)
        store, _ := core.NewStorage("govector.db")
        defer store.Close()

        // 2. Create a Collection (Automatic Persistence)
        col, _ := core.NewCollection("documents", 384, core.Cosine, store, true)

        // 3. Upsert Data with Versioning
        col.Upsert([]core.PointStruct{
                {
                        ID:      "doc_1",
                        Vector:  []float32{...},
                        Payload: core.Payload{"category": "tech"},
                },
        })

        // 4. Search with Metadata Filtering
        results, _ := col.Search(query, filter, 10)
        fmt.Printf("Best Match: %s (Score: %f)\n", results[0].ID, results[0].Score)
}
```

---

## 🌐 Usage Mode 2: Standalone Microservice (Qdrant Compatible)

```bash
# Start the server
govector serve ./govector.db -port 18080

# API Server ready on http://localhost:18080
```

GoVector supports the standard Qdrant-like REST API for `/collections`, `/points`, and `/search`.

---

## 🏗️ Architecture

- **Storage Engine**: `go.etcd.io/bbolt` with **Protocol Buffers**.
- **Graph Indexing**: `github.com/coder/hnsw`.
- **Distance Metrics**: Cosine, Euclidean, Dot Product.

## 🤝 Contributing

PRs are welcome! Help us make GoVector the ultimate embedded vector database for Go.

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.
