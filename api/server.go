// Package api provides a Qdrant-compatible HTTP API server for GoVector.
// It exposes REST endpoints for managing vector collections and performing similarity search.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/DotNetAge/govector/core"
)

// Server handles Qdrant-compatible HTTP requests for vector operations.
// It manages collections and provides endpoints for upsert, search, and delete operations.
type Server struct {
	addr        string
	collections map[string]*core.Collection
	store       *core.Storage // Persistence Engine
	httpServer  *http.Server  // Needed for graceful shutdown
}

// NewServer initializes the API server with the given address and storage backend.
// The storage parameter can be nil for in-memory-only operation.
func NewServer(addr string, store *core.Storage) *Server {
	return &Server{
		addr:        addr,
		collections: make(map[string]*core.Collection),
		store:       store,
	}
}

// AddCollection manually mounts a collection to the server.
// This is typically used during server initialization before Start is called.
func (s *Server) AddCollection(col *core.Collection) {
	s.collections[col.Name] = col
}

// Start boots the HTTP server and begins listening for requests.
// This method blocks until the server is stopped or encounters an error.
// Endpoints:
//   - PUT /collections/{name}/points - Upsert points
//   - POST /collections/{name}/points/search - Search vectors
//   - POST /collections/{name}/points/delete - Delete points
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("PUT /collections/{name}/points", s.handleUpsert)
	mux.HandleFunc("POST /collections/{name}/points/search", s.handleSearch)
	mux.HandleFunc("POST /collections/{name}/points/delete", s.handleDelete)

	s.httpServer = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	log.Printf("Starting GoVector API Server on %s", s.addr)
	return s.httpServer.ListenAndServe()
}

// Stop gracefully shuts down the server without interrupting active connections.
// It uses the provided context to determine how long to wait for connections to close.
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		log.Println("Shutting down HTTP server...")
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// handleUpsert processes PUT /collections/{name}/points requests.
// It accepts a JSON body with a "points" array and upserts them into the specified collection.
// Returns 404 if the collection doesn't exist, 400 for invalid JSON, or 500 for internal errors.
func (s *Server) handleUpsert(w http.ResponseWriter, r *http.Request) {
	colName := r.PathValue("name")
	col, exists := s.collections[colName]
	if !exists {
		http.Error(w, fmt.Sprintf("Collection %s not found", colName), http.StatusNotFound)
		return
	}

	var req struct {
		Points []core.PointStruct `json:"points"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := col.Upsert(req.Points); err != nil {
		http.Error(w, fmt.Sprintf("Failed to upsert: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"result": map[string]string{"operation": "completed"},
	})
}

// handleSearch processes POST /collections/{name}/points/search requests.
// It accepts a JSON body with "vector", optional "filter", and "limit" fields.
// Returns the topK most similar vectors matching the query and filter criteria.
// Returns 404 if the collection doesn't exist, 400 for invalid JSON, or 500 for internal errors.
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	colName := r.PathValue("name")
	col, exists := s.collections[colName]
	if !exists {
		http.Error(w, fmt.Sprintf("Collection %s not found", colName), http.StatusNotFound)
		return
	}

	var req struct {
		Vector []float32    `json:"vector"`
		Filter *core.Filter `json:"filter,omitempty"` // Added Payload Filter
		Limit  int          `json:"limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10 // default
	}

	results, err := col.Search(req.Vector, req.Filter, req.Limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"result": results,
	})
}

// handleDelete processes POST /collections/{name}/points/delete requests.
// It accepts a JSON body with either "points" (array of IDs) or "filter" to delete by.
// Returns the count of deleted points in the response.
// Returns 404 if the collection doesn't exist, 400 for invalid JSON, or 500 for internal errors.
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	colName := r.PathValue("name")
	col, exists := s.collections[colName]
	if !exists {
		http.Error(w, fmt.Sprintf("Collection %s not found", colName), http.StatusNotFound)
		return
	}

	var req struct {
		Points []string     `json:"points,omitempty"`
		Filter *core.Filter `json:"filter,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	deletedCount, err := col.Delete(req.Points, req.Filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Delete failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"result": map[string]interface{}{
			"operation": "completed",
			"deleted":   deletedCount,
		},
	})
}
