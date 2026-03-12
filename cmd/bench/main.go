package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/DotNetAge/govector/core"
)

// generateRandomVector creates a single vector of given dimension
func generateRandomVector(dim int) []float32 {
	vec := make([]float32, dim)
	for j := 0; j < dim; j++ {
		vec[j] = rand.Float32()
	}
	return vec
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("   Alloc = %v | TotalAlloc = %v | Sys = %v | NumGC = %v\n",
		formatBytes(m.Alloc), formatBytes(m.TotalAlloc), formatBytes(m.Sys), m.NumGC)
}

func runBenchmark(name string, dim int, useHNSW bool, numPoints, numQueries int) {
	fmt.Printf("\n=== Benchmark: %s (N=%d, Dim=%d) ===\n", name, numPoints, dim)
	printMemUsage()

	// Create a collection strictly in memory (store = nil) to test pure indexing performance
	col, err := core.NewCollection("bench_col", dim, core.Cosine, nil, useHNSW)
	if err != nil {
		fmt.Printf("Error creating collection: %v\n", err)
		return
	}

	// 1. Benchmark Upsert (Indexing in batches to save memory during generation)
	batchSize := 10000
	if numPoints < batchSize {
		batchSize = numPoints
	}

	startBuild := time.Now()
	for i := 0; i < numPoints; i += batchSize {
		currentBatchSize := batchSize
		if i+batchSize > numPoints {
			currentBatchSize = numPoints - i
		}

		batch := make([]core.PointStruct, currentBatchSize)
		for j := 0; j < currentBatchSize; j++ {
			batch[j] = core.PointStruct{
				ID:     fmt.Sprintf("pt_%d", i+j),
				Vector: generateRandomVector(dim),
			}
		}

		err = col.Upsert(batch)
		if err != nil {
			fmt.Printf("Error upserting at %d: %v\n", i, err)
			return
		}

		if (i+currentBatchSize)%(batchSize*10) == 0 || i+currentBatchSize == numPoints {
			fmt.Printf("   Progress: %d/%d (%.1f%%)\n", i+currentBatchSize, numPoints, float64(i+currentBatchSize)/float64(numPoints)*100)
		}
	}
	buildDuration := time.Since(startBuild)

	printMemUsage()

	// 2. Benchmark Search
	fmt.Printf("   Running %d random queries (TopK=10)...\n", numQueries)
	startSearch := time.Now()
	for i := 0; i < numQueries; i++ {
		q := generateRandomVector(dim)
		_, err := col.Search(q, nil, 10)
		if err != nil {
			fmt.Printf("Error searching: %v\n", err)
			return
		}
	}
	searchDuration := time.Since(startSearch)

	// Metrics Calculation
	qps := float64(numQueries) / searchDuration.Seconds()
	avgLatencyMs := float64(searchDuration.Microseconds()) / float64(numQueries) / 1000.0

	fmt.Printf("✅ Index Build Time:   %v (Avg: %.3f ms/point)\n", buildDuration, float64(buildDuration.Milliseconds())/float64(numPoints))
	fmt.Printf("✅ Search Avg Latency: %.3f ms/query\n", avgLatencyMs)
	fmt.Printf("✅ Search QPS:         %.0f queries/sec\n", qps)
}

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	fmt.Println("🚀 Starting GoVector Large Scale Benchmark Suite...")
	fmt.Println("Metric: Cosine Similarity | TopK: 10")

	dim := 128
	numQueries := 100

	// Scale levels
	scales := []struct {
		Name  string
		Count int
	}{
		{"10K (Baseline)", 10000},
		{"100K (Small)", 100000},
		{"1M (Medium)", 1000000},
		// {"10M (Large)", 10000000},
	}

	for _, s := range scales {
		// For very large scales, we only run HNSW because Flat is O(N)
		if s.Count <= 100000 {
			runBenchmark("Flat Index - "+s.Name, dim, false, s.Count, numQueries)
		}

		runBenchmark("HNSW Index - "+s.Name, dim, true, s.Count, numQueries)

		// Force GC after each major scale
		runtime.GC()
	}
}
