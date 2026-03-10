<div align="center">
  <h1>🎯 GoVector</h1>
  <p><b>纯 Go 编写的轻量级、可嵌入的向量数据库（可以把它当作向量界的 SQLite）</b></p>

  [![Go Reference](https://pkg.go.dev/badge/github.com/DotNetAge/govector.svg)](https://pkg.go.dev/github.com/DotNetAge/govector)
  [![Go Version](https://img.shields.io/github/go-mod/go-version/DotNetAge/govector)](https://golang.org/)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

  <p>
    <a href="README.md">English</a> | <a href="README_zh.md">简体中文</a>
  </p>
</div>

在本地 AI、桌面应用和边缘计算时代，你并不总是需要像 Milvus 或 Qdrant 这样沉重的分布式向量集群。

**GoVector** 是一个完全使用 Go 语言编写的高性能、嵌入式向量搜索引擎。它提供**与 Qdrant 兼容**的 API 接口，内置用于极速近似最近邻 (ANN) 搜索的 **HNSW** 索引，并通过 **BoltDB** 提供本地持久化存储。

---

## ✨ 核心特性

- 🚀 **纯 Go & 无 CGO**: 无需处理繁杂的 C/C++ 依赖，可轻松交叉编译至任何平台（Windows、macOS、Linux、边缘设备）。
- 🧠 **HNSW 索引**: 采用工业级图索引技术 (`github.com/coder/hnsw`)，搜索复杂度仅为 $O(\log N)$。同时也提供 Flat（暴力扫描）索引供选择。
- 💾 **本地持久化**: 数据在重启后依然存在。所有数据通过 `bbolt` 安全地存储在单个本地文件中。
- 🔍 **元数据过滤 (Payload)**: 支持使用 JSON Payload 对向量搜索结果进行过滤（精确匹配、must/must_not 条件），与 Qdrant 的体验完全一致。
- 🔌 **双模式运行**: 
  - **嵌入式库**: 零网络开销，直接导入到你的 Go 后端或桌面应用中。
  - **独立服务器**: 作为一个轻量级微服务运行，提供兼容 Qdrant 的 REST API。

---

## 📦 安装指南

### 选项 A: 作为 Go 依赖库使用
```bash
go get github.com/DotNetAge/govector/core
```

### 选项 B: 通过 Homebrew 安装 (Mac/Linux 守护进程)
如果你希望将 GoVector 作为后台微服务运行：
```bash
brew tap DotNetAge/govector
brew install govector

# 启动后台守护服务 (开机自启!)
brew services start govector
```
默认情况下，Brew 服务会在 `http://localhost:18080` 运行，并已启用 HNSW 索引。

---

## 🚀 性能基准测试

我们在纯内存环境下，对比了 `Flat`（暴力线性扫描）索引和经过优化的 `HNSW`（图）索引。

**测试参数:** 
- **数据集:** 10,000 条向量
- **维度:** 128
- **查询次数:** 1,000 次查询
- **硬件:** 标准本地开发机

| 索引策略               | 构建时间         | 搜索延迟           | 吞吐量         |
| :--------------------- | :--------------- | :----------------- | :------------- |
| **Flat 索引** (线性扫描) | **0 ms** (瞬间)  | 4.267 ms / 查询    | 234 QPS        |
| **HNSW 索引** (图 ANN)  | 1,688 ms         | **0.058 ms / 查询** | **17,150 QPS** |

> *测试结果: HNSW 在查询吞吐量上实现了惊人的 **~73 倍加速**，轻松达到每秒 17,000 次以上的查询，同时保持了极高的准确率！*

---

## 💻 使用模式 1: 嵌入式 Go 库 (零网络开销)

非常适合本地 RAG (检索增强生成) 应用、桌面客户端或本地化微服务。

```go
package main

import (
        "fmt"
        "github.com/DotNetAge/govector/core"
)

func main() {
        // 1. 初始化本地存储 (会创建一个 .db 文件)
        store, _ := core.NewStorage("my_local_vectors.db")
        defer store.Close()

        // 2. 创建集合 (名称, 维度, 距离度量方式, 存储引擎, 是否启用 HNSW)
        col, _ := core.NewCollection("documents", 3, core.Cosine, store, true)

        // 3. 插入包含元数据 (Payload) 的数据
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

        // 4. 携带元数据过滤条件进行搜索
        query := []float32{1.0, 0.0, 0.0}
        filter := &core.Filter{
                Must: []core.Condition{
                        {Key: "category", Match: core.MatchValue{Value: "tech"}},
                },
        }

        results, _ := col.Search(query, filter, 1) // 获取 Top 1

        fmt.Printf("最佳匹配: %s (得分: %f)\n", results[0].ID, results[0].Score)
}
```

---

## 🌐 使用模式 2: 独立微服务 (兼容 Qdrant)

想在 Python、Node.js 或 Rust 中使用？直接把 GoVector 当作独立服务器运行！

### 1. 启动服务器

```bash
# 克隆仓库并运行服务器
go run cmd/govector-server/main.go -port 18080 -db ./govector.db -hnsw=true

# 输出示例:
# === GoVector: Lightweight Microservice (Port: 18080) ===
# Storage engine loaded: ./govector.db
# API Server ready on http://localhost:18080
```

### 2. 通过 HTTP 插入向量

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

### 3. 带过滤条件的向量搜索

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

### 4. 删除向量

```bash
# 按 ID 列表删除
curl -X POST http://localhost:18080/collections/test_collection/points/delete \
-H "Content-Type: application/json" \
-d '{
  "points": ["1", "2"]
}'

# 按过滤条件删除
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

## 🏗️ 架构说明

- **存储引擎**: `go.etcd.io/bbolt` (极速、纯 Go 实现的键值存储)。
- **图索引**: `github.com/coder/hnsw` (多层可导航小世界图)。
- **距离度量**: 余弦相似度 (Cosine Similarity)、欧氏距离 (Euclidean Distance)、点积 (Dot Product)。

## 🤝 贡献指南

非常欢迎提交 PR！这是一个轻量级项目，旨在成为 Go 生态中最优秀的嵌入式向量数据库。

## 📄 许可证

MIT License. 详情请参阅 [LICENSE](LICENSE) 文件。
