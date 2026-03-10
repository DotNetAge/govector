package main

import (
	"fmt"
	"log"

	"govector/core" // This is how users will import your library
)

func main() {
	fmt.Println("=== GoVector: Embedded Library Usage Example ===")

	// 1. You don't need the API server. Just initialize the local storage engine!
	store, err := core.NewStorage("embedded_govector.db")
	if err != nil {
		log.Fatalf("Failed to init storage: %v", err)
	}
	defer store.Close() // Clean up safely

	// 2. Create your internal collection (Dimensionality 3, Cosine Distance, Use HNSW=true)
	col, err := core.NewCollection("my_app_vectors", 3, core.Cosine, store, true)
	if err != nil {
		log.Fatalf("Failed to init collection: %v", err)
	}

	// 3. Upsert data directly via Go structs (No HTTP overhead!)
	err = col.Upsert([]core.PointStruct{
		{
			ID:      "user_101",
			Vector:  []float32{0.5, 0.5, 0.5},
			Payload: core.Payload{"role": "admin", "age": 30},
		},
		{
			ID:      "user_202",
			Vector:  []float32{0.8, 0.2, 0.0},
			Payload: core.Payload{"role": "guest", "age": 22},
		},
	})
	if err != nil {
		log.Fatalf("Upsert error: %v", err)
	}

	// 4. Perform a rich query using Go structs directly
	fmt.Println("\nSearching for an admin closest to [1.0, 0.0, 0.0]...")
	query := []float32{1.0, 0.0, 0.0}
	
	filter := &core.Filter{
		Must: []core.Condition{
			{Key: "role", Match: core.MatchValue{Value: "admin"}},
		},
	}

	results, err := col.Search(query, filter, 1) // Get Top 1
	if err != nil {
		log.Fatalf("Search error: %v", err)
	}

	// 5. Native Go struct access
	for _, res := range results {
		fmt.Printf("Found Native Object! ID: %s | Score: %.4f | Payload: %+v\n", res.ID, res.Score, res.Payload)
	}
}
