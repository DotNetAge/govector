// Package api provides a Qdrant-compatible HTTP API server for GoVector.
// It exposes REST endpoints for managing vector collections and performing similarity search.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/DotNetAge/govector/core"
)

// Server handles Qdrant-compatible HTTP requests for vector operations.
// It manages collections and provides endpoints for upsert, search, and delete operations.
type Server struct {
	addr        string
	collections map[string]*core.Collection
	store       *core.Storage // Persistence Engine
	httpServer  *http.Server  // Needed for graceful shutdown
	serverMu    sync.Mutex    // Protects httpServer and collections
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

// NewServerWithQuantization initializes the API server with vector quantization enabled.
// It creates a storage engine with the specified quantization settings.
func NewServerWithQuantization(addr string, dbPath string, useQuant bool) (*Server, error) {
	store, err := core.NewStorageWithQuantization(dbPath, useQuant, nil, false)
	if err != nil {
		return nil, err
	}
	return NewServer(addr, store), nil
}

// AddCollection manually mounts a collection to the server.
// This is typically used during server initialization before Start is called.
func (s *Server) AddCollection(col *core.Collection) {
	s.serverMu.Lock()
	defer s.serverMu.Unlock()
	s.collections[col.Name] = col
}

// Start boots the HTTP server and begins listening for requests.
// This method blocks until the server is stopped or encounters an error.
// Endpoints:
//   - POST /collections - Create collection
//   - DELETE /collections/{name} - Delete collection
//   - GET /collections - List collections
//   - GET /collections/{name} - Get collection info
//   - PUT /collections/{name}/points - Upsert points
//   - POST /collections/{name}/points/search - Search vectors
//   - POST /collections/{name}/points/delete - Delete points
func (s *Server) Start() error {
	// Automatically load collections from storage
	if s.store != nil {
		if err := s.loadCollections(); err != nil {
			return fmt.Errorf("failed to load collections on start: %w", err)
		}
	}

	mux := http.NewServeMux()

	// Collection management endpoints
	mux.HandleFunc("POST /collections", s.handleCreateCollection)
	mux.HandleFunc("DELETE /collections/{name}", s.handleDeleteCollection)
	mux.HandleFunc("GET /collections", s.handleListCollections)
	mux.HandleFunc("GET /collections/{name}", s.handleGetCollection)

	// Points operations endpoints
	mux.HandleFunc("PUT /collections/{name}/points", s.handleUpsert)
	mux.HandleFunc("POST /collections/{name}/points/search", s.handleSearch)
	mux.HandleFunc("POST /collections/{name}/points/delete", s.handleDelete)

	s.serverMu.Lock()
	s.httpServer = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}
	server := s.httpServer
	s.serverMu.Unlock()

	log.Printf("Starting GoVector API Server on %s", s.addr)
	return server.ListenAndServe()
}

// loadCollections reads all collection metadata from storage and initializes them.
func (s *Server) loadCollections() error {
	if s.store == nil {
		return nil
	}

	metas, err := s.store.ListCollectionMetas()
	if err != nil {
		return err
	}

	for _, meta := range metas {
		log.Printf("Loading collection %s from storage (dim=%d, metric=%s, hnsw=%v)",
			meta.Name, meta.VectorLen, meta.Metric, meta.UseHNSW)

		col, err := core.NewCollectionWithParams(
			meta.Name,
			meta.VectorLen,
			meta.Metric,
			s.store,
			meta.UseHNSW,
			meta.HNSWParams,
		)
		if err != nil {
			return fmt.Errorf("failed to load collection %s: %w", meta.Name, err)
		}
		s.serverMu.Lock()
		s.collections[meta.Name] = col
		s.serverMu.Unlock()
	}

	return nil
}

// Stop gracefully shuts down the server without interrupting active connections.
// It uses the provided context to determine how long to wait for connections to close.
func (s *Server) Stop(ctx context.Context) error {
	s.serverMu.Lock()
	server := s.httpServer
	s.serverMu.Unlock()

	if server != nil {
		log.Println("Shutting down HTTP server...")
		return server.Shutdown(ctx)
	}
	return nil
}

// handleUpsert processes PUT /collections/{name}/points requests.
// It accepts a JSON body with a "points" array and upserts them into the specified collection.
// Returns 404 if the collection doesn't exist, 400 for invalid JSON, or 500 for internal errors.
func (s *Server) handleUpsert(w http.ResponseWriter, r *http.Request) {
	colName := r.PathValue("name")
	s.serverMu.Lock()
	col, exists := s.collections[colName]
	s.serverMu.Unlock()
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
	s.serverMu.Lock()
	col, exists := s.collections[colName]
	s.serverMu.Unlock()
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
	s.serverMu.Lock()
	col, exists := s.collections[colName]
	s.serverMu.Unlock()
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

// handleCreateCollection processes POST /collections requests.
// It accepts a JSON body with collection configuration and creates a new collection.
// Returns 400 for invalid JSON, 409 if collection already exists, or 500 for internal errors.
func (s *Server) handleCreateCollection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name       string                 `json:"name"`
		VectorLen  int                    `json:"vector_size"`
		Metric     string                 `json:"distance"`
		HNSW       bool                   `json:"hnsw,omitempty"`
		Parameters map[string]interface{} `json:"parameters,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Check if collection already exists
	s.serverMu.Lock()
	_, exists := s.collections[req.Name]
	s.serverMu.Unlock()
	if exists {
		http.Error(w, fmt.Sprintf("Collection %s already exists", req.Name), http.StatusConflict)
		return
	}

	// Validate vector length
	if req.VectorLen <= 0 {
		http.Error(w, "Vector size must be positive", http.StatusBadRequest)
		return
	}

	// Parse distance metric
	metric := core.Cosine
	switch req.Metric {
	case "euclidean", "Euclid":
		metric = core.Euclid
	case "dot", "Dot":
		metric = core.Dot
	case "cosine", "Cosine":
		metric = core.Cosine
	default:
		http.Error(w, fmt.Sprintf("Invalid distance metric: %s", req.Metric), http.StatusBadRequest)
		return
	}

	// Parse HNSW parameters if provided
	params := core.DefaultHNSWParams()
	if req.Parameters != nil {
		if m, ok := req.Parameters["m"].(float64); ok {
			params.M = int(m)
		}
		if efConstruction, ok := req.Parameters["ef_construction"].(float64); ok {
			params.EfConstruction = int(efConstruction)
		}
		if efSearch, ok := req.Parameters["ef_search"].(float64); ok {
			params.EfSearch = int(efSearch)
		}
		if k, ok := req.Parameters["k"].(float64); ok {
			params.K = int(k)
		}
	}

	// Create collection
	col, err := core.NewCollectionWithParams(req.Name, req.VectorLen, metric, s.store, req.HNSW, params)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create collection: %v", err), http.StatusInternalServerError)
		return
	}

	// Add to server
	s.serverMu.Lock()
	s.collections[req.Name] = col
	s.serverMu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"result": map[string]interface{}{
			"name":        req.Name,
			"vector_size": req.VectorLen,
			"distance":    string(metric),
			"hnsw":        req.HNSW,
			"parameters": map[string]interface{}{
				"m":               params.M,
				"ef_construction": params.EfConstruction,
				"ef_search":       params.EfSearch,
				"k":               params.K,
			},
			"operation": "completed",
		},
	})
}

// handleDeleteCollection processes DELETE /collections/{name} requests.
// It deletes the specified collection and all its data.
// Returns 404 if the collection doesn't exist, or 500 for internal errors.
func (s *Server) handleDeleteCollection(w http.ResponseWriter, r *http.Request) {
	colName := r.PathValue("name")
	s.serverMu.Lock()
	_, exists := s.collections[colName]
	s.serverMu.Unlock()
	if !exists {
		http.Error(w, fmt.Sprintf("Collection %s not found", colName), http.StatusNotFound)
		return
	}

	// Remove from server
	s.serverMu.Lock()
	delete(s.collections, colName)
	s.serverMu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"result": map[string]string{
			"operation":  "completed",
			"collection": colName,
		},
	})
}

// handleListCollections processes GET /collections requests.
// It returns a list of all collections and their basic information.
func (s *Server) handleListCollections(w http.ResponseWriter, r *http.Request) {
	var collections []map[string]interface{}
	for name, col := range s.collections {
		collections = append(collections, map[string]interface{}{
			"name":         name,
			"vector_size":  col.VectorLen,
			"distance":     string(col.Metric),
			"points_count": col.Count(),
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"result": collections,
	})
}

// handleGetCollection processes GET /collections/{name} requests.
// It returns detailed information about the specified collection.
// Returns 404 if the collection doesn't exist.
func (s *Server) handleGetCollection(w http.ResponseWriter, r *http.Request) {
	colName := r.PathValue("name")
	col, exists := s.collections[colName]
	if !exists {
		http.Error(w, fmt.Sprintf("Collection %s not found", colName), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"result": map[string]interface{}{
			"name":         colName,
			"vector_size":  col.VectorLen,
			"distance":     string(col.Metric),
			"points_count": col.Count(),
		},
	})
}
