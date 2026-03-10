package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"govector/api"
	"govector/core"
)

func main() {
	// Parse command-line flags for microservice usage
	port := flag.Int("port", 18080, "Port for the HTTP API server")
	dbPath := flag.String("db", "govector.db", "Path to the BoltDB storage file")
	useHNSW := flag.Bool("hnsw", true, "Use high-performance HNSW index (false = Flat Index)")
	flag.Parse()

	fmt.Printf("=== GoVector: Lightweight Microservice (Port: %d) ===\n", *port)

	// 1. Initialize Storage Engine (BoltDB)
	store, err := core.NewStorage(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer func() {
		log.Println("Closing local storage engine...")
		store.Close()
	}()
	fmt.Printf("Storage engine loaded: %s\n", *dbPath)

	// 2. Load or Create Default Collection
	dim := 3
	colName := "test_collection"
	col, err := core.NewCollection(colName, dim, core.Cosine, store, *useHNSW)
	if err != nil {
		log.Fatalf("Failed to create/load collection: %v", err)
	}
	
	indexType := "Flat"
	if *useHNSW {
		indexType = "HNSW"
	}
	fmt.Printf("Loaded collection: %s (Dim: %d, Metric: %s, Index: %s, Current Points: %d)\n", col.Name, dim, col.Metric, indexType, col.Count())

	// 3. Initialize API Server
	addr := fmt.Sprintf(":%d", *port)
	server := api.NewServer(addr, store)
	server.AddCollection(col) // Register the collection with the HTTP server

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		fmt.Printf("API Server ready on http://localhost%s\n", addr)
		serverErrors <- server.Start()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		// ErrServerClosed is normal during graceful shutdown
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	case sig := <-osSignals:
		fmt.Printf("\nStart shutdown... Signal: %v\n", sig)

		// Create context for shutdown call.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := server.Stop(ctx); err != nil {
			log.Printf("Graceful shutdown did not complete in time: %v", err)
		}
	}
	
	fmt.Println("Server gracefully stopped.")
}
