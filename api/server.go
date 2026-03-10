package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/DotNetAge/govector/core"
)

// Server handles Qdrant-compatible HTTP requests
type Server struct {
	addr        string
	collections map[string]*core.Collection
	store       *core.Storage // Persistence Engine
	httpServer  *http.Server  // Needed for graceful shutdown
}

// NewServer initializes the API server
func NewServer(addr string, store *core.Storage) *Server {
	return &Server{
		addr:        addr,
		collections: make(map[string]*core.Collection),
		store:       store,
	}
}

// AddCollection manually mounts a collection (normally done via API, simplifying for now)
func (s *Server) AddCollection(col *core.Collection) {
	s.collections[col.Name] = col
}

// Start boots the HTTP server
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

// Stop gracefully shuts down the server without interrupting active connections
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		log.Println("Shutting down HTTP server...")
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// handleUpsert processes PUT /collections/{name}/points
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

// handleSearch processes POST /collections/{name}/points/search
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

// handleDelete processes POST /collections/{name}/points/delete
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
