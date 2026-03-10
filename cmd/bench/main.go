package main

import (
	"fmt"
	"math/rand"
	"time"

	"govector/core"
)

// generateRandomVectors creates n vectors of given dimension
func generateRandomVectors(n, dim int) [][]float32 {
	vectors := make([][]float32, n)
	for i := 0; i < n; i++ {
		vec := make([]float32, dim)
		for j := 0; j < dim; j++ {
			vec[j] = rand.Float32()
		}
		vectors[i] = vec
	}
	return vectors
}

func runBenchmark(name string, dim int, useHNSW bool, numPoints, numQueries int) {
	fmt.Printf("--- Benchmark: %s ---\n", name)

	// Create a collection strictly in memory (store = nil) to test pure indexing performance
	col, err := core.NewCollection("bench_col", dim, core.Cosine, nil, useHNSW)
	if err != nil {
		fmt.Printf("Error creating collection: %v\n", err)
		return
	}

	fmt.Printf("Generating %d random vectors (Dim: %d)...\n", numPoints, dim)
	pointsData := generateRandomVectors(numPoints, dim)

	// Prepare struct points
	var points []core.PointStruct
	for i, vec := range pointsData {
		points = append(points, core.PointStruct{
			ID:     fmt.Sprintf("pt_%d", i),
			Vector: vec,
			// No payload to test pure distance calculation and indexing
		})
	}

	// 1. Benchmark Upsert (Indexing)
	startBuild := time.Now()
	err = col.Upsert(points)
	if err != nil {
		fmt.Printf("Error upserting: %v\n", err)
		return
	}
	buildDuration := time.Since(startBuild)
	buildMs := buildDuration.Milliseconds()

	// Generate Query Vectors
	fmt.Printf("Generating %d random queries...\n", numQueries)
	queries := generateRandomVectors(numQueries, dim)

	// 2. Benchmark Search
	startSearch := time.Now()
	for _, q := range queries {
		_, err := col.Search(q, nil, 10) // Top 10
		if err != nil {
			fmt.Printf("Error searching: %v\n", err)
			return
		}
	}
	searchDuration := time.Since(startSearch)

	// Metrics Calculation
	qps := float64(numQueries) / searchDuration.Seconds()
	avgLatencyMs := float64(searchDuration.Microseconds()) / float64(numQueries) / 1000.0

	fmt.Printf("✅ Index Build Time:   %d ms\n", buildMs)
	fmt.Printf("✅ Search Avg Latency: %.3f ms/query\n", avgLatencyMs)
	fmt.Printf("✅ Search QPS:         %.0f queries/sec\n\n", qps)
}

func main() {
	// Initialize random seed
	rand.Seed(42)

	fmt.Println("🚀 Starting GoVector Benchmark Suite...")
	fmt.Println("Dataset: 10,000 points | Dim: 128 | Queries: 1,000 | TopK: 10")
	fmt.Println("===============================================================")

	dim := 128
	numPoints := 10000
	numQueries := 1000

	// Run Flat (Brute-Force) Benchmark
	runBenchmark("Flat Index (Linear Scan)", dim, false, numPoints, numQueries)

	// Run HNSW Benchmark
	runBenchmark("HNSW Index (Graph ANN)", dim, true, numPoints, numQueries)
}
