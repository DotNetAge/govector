<div align="center">
  <h1>🎯 GoVector</h1>
  <p><b>The Lightweight, Embeddable Vector Database in Pure Go. (Think SQLite for Vectors)</b></p>

  [![Go Reference](https://pkg.go.dev/badge/github.com/DotNetAge/govector.svg)](https://pkg.go.dev/github.com/DotNetAge/govector)
  [![Go Version](https://img.shields.io/github/go-mod/go-version/DotNetAge/govector)](https://golang.org/)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

  <p>
    <a href="README.md">English</a> | <a href="README_zh.md">简体中文</a>
  </p>
</div>

In the era of Local AI, desktop applications, and edge computing, you don't always need a heavy, distributed vector cluster like Milvus or Qdrant. 

**GoVector** is a high-performance, embedded vector search engine written entirely in Go. It offers **Qdrant-compatible** API endpoints, **HNSW** indexing for blazing-fast Approximate Nearest Neighbor (ANN) search, and persistent local storage via **BoltDB**.

---

## ✨ Features

- 🚀 **Pure Go & CGO-Free**: Cross-compile to anywhere (Windows, macOS, Linux, edge devices) without messy C/C++ dependencies.
- 🧠 **HNSW Indexing**: Industrial-grade graph-based index (`github.com/coder/hnsw`) for $O(\log N)$ search complexity. Flat (brute-force) index is also available.
- 💾 **Local Persistence**: Data survives restarts. Everything is stored securely in a single local file via `bbolt`.
- 🔍 **Metadata Filtering (Payload)**: Filter vector search results using JSON payloads (exact matches, must/must_not), just like Qdrant.
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
If you want to run GoVector as a background microservice:
```bash
brew tap DotNetAge/govector
brew install govector

# Start the background daemon service (starts on boot!)
brew services start govector
```
By default, the brew service runs on `http://localhost:18080` with HNSW enabled.

---

## 🚀 Benchmark Performance

We ran a pure-memory benchmark comparing the `Flat` (brute-force linear scan) index against our optimized `HNSW` (Graph) index.

**Parameters:** 
- **Dataset:** 10,000 vectors
- **Dimensions:** 128
- **Query count:** 1,000 queries
- **Hardware:** Standard local desktop

| Index Strategy               | Build Time         | Search Latency       | Throughput     |
| :--------------------------- | :----------------- | :------------------- | :------------- |
| **Flat Index** (Linear Scan) | **0 ms** (Instant) | 4.267 ms / query     | 234 QPS        |
| **HNSW Index** (Graph ANN)   | 1,688 ms           | **0.058 ms / query** | **17,150 QPS** |

> *Result: HNSW provides an astonishing **~73x speedup** in query throughput, effortlessly delivering over 17,000 queries per second while maintaining high accuracy!*

---

## 💻 Usage Mode 1: Embedded Go Library (Zero Network)

Perfect for local RAG (Retrieval-Augmented Generation) applications, desktop apps, or localized microservices.

```go
package main

import (
        "fmt"
        "govector/core"
)

func main() {
        // 1. Initialize local storage (creates a single .db file)
        store, _ := core.NewStorage("my_local_vectors.db")
        defer store.Close()

        // 2. Create a Collection (Name, Dimension, Distance Metric, Storage, UseHNSW)
        col, _ := core.NewCollection("documents", 3, core.Cosine, store, true)

        // 3. Upsert Data with Metadata (Payload)
        col.Upsert([]core.PointStruct{
                {
                        ID:      "doc_1",
                        Vector:  []float32{0.9, 0.1, 0.0},
                        Payload: core.Payload{"category": "tech", "author": "Alice"},
                },
                {
                        ID:      "doc_2",
                        Vector:  []float32{0.1, 0.9, 0.0},
                        Payload: core.Payload{"category": "art", "author": "Bob"},
                },
        })

        // 4. Search with Metadata Filtering
        query := []float32{1.0, 0.0, 0.0}
        filter := &core.Filter{
                Must: []core.Condition{
                        {Key: "category", Match: core.MatchValue{Value: "tech"}},
                },
        }

        results, _ := col.Search(query, filter, 1) // Top 1

        fmt.Printf("Best Match: %s (Score: %f)\n", results[0].ID, results[0].Score)
}
```

---

## 🌐 Usage Mode 2: Standalone Microservice (Qdrant Compatible)

Want to use it from Python, Node.js, or Rust? Run GoVector as a standalone server!

### 1. Start the Server

```bash
# Clone the repo and run the server
go run cmd/govector-server/main.go -port 18080 -db ./govector.db -hnsw=true

# Output:
# === GoVector: Lightweight Microservice (Port: 18080) ===
# Storage engine loaded: ./govector.db
# API Server ready on http://localhost:18080
```

### 2. Insert Vectors via HTTP

```bash
curl -X PUT http://localhost:18080/collections/test_collection/points \
-H "Content-Type: application/json" \
-d '{
  "points": [
    {"id": "1", "vector": [1.0, 0.0, 0.0], "payload": {"category": "fruit", "name": "Apple"}},
    {"id": "2", "vector": [0.0, 1.0, 0.0], "payload": {"category": "animal", "name": "Dog"}}
  ]
}'
```

### 3. Search Vectors with Filters

```bash
curl -X POST http://localhost:18080/collections/test_collection/points/search \
-H "Content-Type: application/json" \
-d '{
  "vector": [0.8, 0.2, 0.0],
  "limit": 1,
  "filter": {
    "must": [
      { "key": "category", "match": { "value": "fruit" } }
    ]
  }
}'
```

### 4. Delete Vectors

```bash
# Delete by ID
curl -X POST http://localhost:18080/collections/test_collection/points/delete \
-H "Content-Type: application/json" \
-d '{
  "points": ["1", "2"]
}'

# Delete by Filter
curl -X POST http://localhost:18080/collections/test_collection/points/delete \
-H "Content-Type: application/json" \
-d '{
  "filter": {
    "must": [
      { "key": "category", "match": { "value": "fruit" } }
    ]
  }
}'
```

---

## 🏗️ Architecture

- **Storage Engine**: `go.etcd.io/bbolt` (Fast, pure Go key-value store).
- **Graph Indexing**: `github.com/coder/hnsw` (Multi-layer Navigable Small World Graph).
- **Distance Metrics**: Cosine Similarity, Euclidean Distance, Dot Product.

## 🤝 Contributing

PRs are welcome! This is a lightweight project aiming to be the ultimate embedded vector database for the Go ecosystem.

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.
