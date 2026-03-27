# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- **Unified CLI (`govector`)**: Merged `govectord` and `govector-server` into a single, comprehensive CLI tool.
- **Interactive TUI**: Added a green-themed Terminal User Interface (TUI) as the default mode when running `govector` without arguments, featuring a custom ASCII banner and a `/?` help menu.
- **`rm` Command**: Introduced a command to safely delete collections from the database, with explicit safeguards to prevent the deletion of the "default" collection.
- **Auto-Creation**: Write operations (like `upsert` and TUI startup) now automatically create the specified database file and collections if they don't exist, defaulting to HNSW indexing.
- **Existence Validation**: Read operations (`search`, `count`, `delete`) now strictly validate the existence of databases and collections, gracefully exiting with an error if missing.
- **Dynamic Versioning**: Integrated `git describe --tags` into the build and release pipelines (`Makefile`, `build_release.sh`) for dynamic version resolution, replacing hardcoded version strings.

### Changed
- **CLI Syntax**: Updated the CLI command syntax to `govector <command> [dbfile] [options]`, making the database file optional (defaults to `govector.db`).
- **Server Shutdown**: The `serve` command running inside the TUI now handles `Ctrl+C` gracefully, stopping the HTTP server and returning to the TUI prompt instead of terminating the application.
- **Release Scripts**: Updated Homebrew `.rb` formula and `.service` templates to use the unified `govector serve` command and removed legacy flags (e.g., `-hnsw=true`).
- **Documentation Sync**: Synchronized all global documentation, including `README.md`, `README_zh.md`, `.docs/`, and `.qoder/repowiki/` content to match the new CLI architecture.

### Removed
- **`cmd/govector-server`**: Entirely deleted the standalone server entrypoint in favor of the new `govector serve` subcommand.
- **`govectord`**: Removed the `govectord` binary alias from the Homebrew build pipeline.

## [0.2.0] - 2026-03-12

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
