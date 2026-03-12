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

## ✨ Features

- 🚀 **Pure Go & CGO-Free**: Cross-compile to anywhere (Windows, macOS, Linux, edge devices) without messy C/C++ dependencies.
- ⚡ **High Performance**: Optimized for single-node performance, supporting millions of vectors with sub-millisecond search latency.
- 🧠 **HNSW Indexing**: Industrial-grade graph-based index for $O(\log N)$ search complexity.
- 💾 **Protobuf & BoltDB**: Ultra-fast persistence using Protocol Buffers and bbolt. Data survives restarts with automatic collection discovery.
- 🔍 **Advanced Filtering**: Support for payload filtering (Exact, Range, Prefix, Regex, Contains) just like Qdrant.
- 📉 **SQ8 Quantization**: Built-in 8-bit scalar quantization to reduce disk footprint for large-scale data.
- 🔌 **Dual Modes**: 
  - **Embedded Library**: Import it into your Go backend/desktop app with zero network overhead.
  - **Standalone Server**: Run it as a lightweight microservice with a Qdrant-compatible REST API.

---

## 📦 Installation

### Option A: Use as a Go Library
```bash
go get github.com/DotNetAge/govector/core
```

### Option B: Install via Homebrew (Mac/Linux Daemon)
```bash
brew tap DotNetAge/govector
brew install govector
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
go run cmd/govector-server/main.go -port 18080 -db ./govector.db -hnsw=true

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
