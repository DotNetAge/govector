# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased] - 2026-03-12

### Added
- **Protobuf Storage Engine**: Replaced JSON serialization with Protocol Buffers for high-performance vector and point persistence.
- **Collection Auto-Discovery**: The API server now automatically scans and registers all collections found in the BoltDB storage on startup.
- **Data Versioning**: Introduced 64-bit nanosecond-precision versioning for all points, enabling future consistency checks.
- **SQ8 Vector Quantization**: Integrated Scalar Quantization (8-bit) to significantly reduce disk footprint for large-scale datasets.
- **Dynamic Collection Metadata**: Support for persisting and loading collection-specific parameters (Dim, Metric, HNSW config).
- **Safe Concurrent Access**: Full mutex protection for the API server's collection registry, eliminating data races.
- **Advanced Payload Filtering**: Added support for `Range`, `Prefix`, `Contains`, and `Regex` filtering in both search and deletion.
- **Consistent Deletion**: Refactored deletion logic to ensure storage and memory index are always in sync (Storage-First-then-Memory).
- **Large-Scale Benchmark Suite**: New benchmark tool supporting 10K, 100K, 1M, and 10M scale testing with memory usage reporting.

### Optimized
- **HNSW Recall**: Improved heuristic for post-filtering recall by over-fetching candidates during filtered searches.
- **Test Coverage**: Increased core package coverage to **92.7%** and API package coverage to **92.8%**.
- **Memory Management**: Batch-processing in benchmarks and better GC triggering for large datasets.

### Fixed
- Fixed a data race in `api/server.go` when starting/stopping the HTTP server.
- Fixed a potential integer overflow in the SQ8 quantization algorithm.
- Fixed a bug where collections created via API were lost after server restart.
