package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DotNetAge/govector/core"
)

func createTestServerAndMux(t *testing.T) (*Server, *http.ServeMux, func()) {
	tempDir, err := os.MkdirTemp("", "api-test-dir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	tempPath := filepath.Join(tempDir, "storage.db")

	store, err := core.NewStorage(tempPath)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create storage: %v", err)
	}

	server := NewServer(":0", store)

	col, err := core.NewCollection("test", 3, core.Cosine, store, false)
	if err != nil {
		store.Close()
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create collection: %v", err)
	}

	points := []core.PointStruct{
		{ID: "1", Vector: []float32{1.0, 0.0, 0.0}},
		{ID: "2", Vector: []float32{0.0, 1.0, 0.0}},
		{ID: "3", Vector: []float32{0.0, 0.0, 1.0}},
	}
	if err := col.Upsert(points); err != nil {
		store.Close()
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to upsert points: %v", err)
	}

	server.AddCollection(col)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /collections", server.handleCreateCollection)
	mux.HandleFunc("DELETE /collections/{name}", server.handleDeleteCollection)
	mux.HandleFunc("GET /collections", server.handleListCollections)
	mux.HandleFunc("GET /collections/{name}", server.handleGetCollection)
	mux.HandleFunc("PUT /collections/{name}/points", server.handleUpsert)
	mux.HandleFunc("POST /collections/{name}/points/search", server.handleSearch)
	mux.HandleFunc("POST /collections/{name}/points/delete", server.handleDelete)

	cleanup := func() {
		store.Close()
		os.RemoveAll(tempDir)
	}

	return server, mux, cleanup
}

func TestNewServer(t *testing.T) {
	server := NewServer(":8080", nil)
	if server == nil {
		t.Fatal("Expected server to be non-nil")
	}
	if server.addr != ":8080" {
		t.Errorf("Expected addr :8080, got %s", server.addr)
	}
}

func TestNewServerWithQuantization(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "quant-test")
	defer os.RemoveAll(tempDir)
	dbPath := filepath.Join(tempDir, "test.db")

	server, err := NewServerWithQuantization(":0", dbPath, true)
	if err != nil {
		t.Fatalf("Failed to create server with quantization: %v", err)
	}
	if server == nil {
		t.Fatal("Expected server to be non-nil")
	}

	// Test error case
	_, err = NewServerWithQuantization(":0", "/non/existent/path/db.db", true)
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestServerStartAndStop(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "start-test")
	defer os.RemoveAll(tempDir)
	dbPath := filepath.Join(tempDir, "test.db")

	server, _ := NewServerWithQuantization(":0", dbPath, false)
	
	// Pre-create a collection to test loadCollections
	col, _ := core.NewCollection("preload", 4, core.Euclid, server.store, false)
	server.AddCollection(col)

	// Start server in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start()
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test Stop
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := server.Stop(ctx); err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}

	err := <-errChan
	if err != nil && err != http.ErrServerClosed {
		t.Errorf("Server exited with error: %v", err)
	}
}

func TestLoadCollections(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "load-test")
	defer os.RemoveAll(tempDir)
	dbPath := filepath.Join(tempDir, "test.db")

	store, _ := core.NewStorage(dbPath)
	// Create collection directly in storage
	core.NewCollectionWithParams("col1", 3, core.Cosine, store, false, core.DefaultHNSWParams())
	store.Close()

	// Reopen with server
	server, _ := NewServerWithQuantization(":0", dbPath, false)
	err := server.loadCollections()
	if err != nil {
		t.Fatalf("loadCollections failed: %v", err)
	}

	if server.GetCollectionsMapSize() != 1 {
		t.Errorf("Expected 1 collection, got %d", server.GetCollectionsMapSize())
	}
	if server.GetCollection("col1") == nil {
		t.Error("Collection col1 not found")
	}
}

func TestHandleCreateCollection(t *testing.T) {
	server, mux, cleanup := createTestServerAndMux(t)
	defer cleanup()

	t.Run("Success", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name":        "new_col",
			"vector_size": 128,
			"distance":    "euclidean",
			"hnsw":        true,
			"parameters": map[string]interface{}{
				"m":                16,
				"ef_construction": 128,
				"ef_search":       64,
				"k":               10,
			},
		})
		req := httptest.NewRequest(http.MethodPost, "/collections", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if server.GetCollection("new_col") == nil {
			t.Error("Collection was not created")
		}
	})

	t.Run("AlreadyExists", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name":        "test",
			"vector_size": 3,
			"distance":    "cosine",
		})
		req := httptest.NewRequest(http.MethodPost, "/collections", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Expected 409, got %d", w.Code)
		}
	})

	t.Run("InvalidMetric", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name":        "invalid_metric",
			"vector_size": 3,
			"distance":    "unknown",
		})
		req := httptest.NewRequest(http.MethodPost, "/collections", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}
	})

	t.Run("InvalidVectorSize", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name":        "invalid_size",
			"vector_size": 0,
			"distance":    "cosine",
		})
		req := httptest.NewRequest(http.MethodPost, "/collections", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}
	})

	t.Run("OtherMetrics", func(t *testing.T) {
		metrics := []string{"dot", "cosine", "Euclid", "Dot", "Cosine"}
		for _, m := range metrics {
			name := fmt.Sprintf("col_%s", m)
			body, _ := json.Marshal(map[string]interface{}{
				"name":        name,
				"vector_size": 3,
				"distance":    m,
			})
			req := httptest.NewRequest(http.MethodPost, "/collections", bytes.NewReader(body))
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("Expected 200 for metric %s, got %d", m, w.Code)
			}
		}
	})
	
	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/collections", bytes.NewReader([]byte("{invalid}")))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}
	})
}

func TestHandleDeleteCollection(t *testing.T) {
	server, mux, cleanup := createTestServerAndMux(t)
	defer cleanup()

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/collections/test", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		if server.GetCollection("test") != nil {
			t.Error("Collection still exists after deletion")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/collections/nonexistent", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", w.Code)
		}
	})
}

func TestHandleListCollections(t *testing.T) {
	_, mux, cleanup := createTestServerAndMux(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/collections", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	results := resp["result"].([]interface{})
	if len(results) != 1 {
		t.Errorf("Expected 1 collection in list, got %d", len(results))
	}
}

func TestHandleGetCollection(t *testing.T) {
	_, mux, cleanup := createTestServerAndMux(t)
	defer cleanup()

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/collections/test", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)
		result := resp["result"].(map[string]interface{})
		if result["name"] != "test" {
			t.Errorf("Expected name 'test', got %v", result["name"])
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/collections/nonexistent", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", w.Code)
		}
	})
}

func TestHandleUpsert(t *testing.T) {
	server, mux, cleanup := createTestServerAndMux(t)
	defer cleanup()

	t.Run("Success", func(t *testing.T) {
		newPoints := []core.PointStruct{
			{ID: "4", Vector: []float32{0.5, 0.5, 0.0}},
		}
		body, _ := json.Marshal(map[string]interface{}{
			"points": newPoints,
		})
		req := httptest.NewRequest(http.MethodPut, "/collections/test/points", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		if server.GetCollection("test").Count() != 4 {
			t.Errorf("Expected count 4, got %d", server.GetCollection("test").Count())
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/collections/test/points", bytes.NewReader([]byte("invalid")))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/collections/nonexistent/points", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", w.Code)
		}
	})
}

func TestHandleSearch(t *testing.T) {
	_, mux, cleanup := createTestServerAndMux(t)
	defer cleanup()

	t.Run("Success", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"vector": []float32{1.0, 0.0, 0.0},
			"limit":  2,
		})
		req := httptest.NewRequest(http.MethodPost, "/collections/test/points/search", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
	})

	t.Run("DefaultLimit", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"vector": []float32{1.0, 0.0, 0.0},
		})
		req := httptest.NewRequest(http.MethodPost, "/collections/test/points/search", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/collections/test/points/search", bytes.NewReader([]byte("invalid")))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/collections/nonexistent/points/search", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", w.Code)
		}
	})
}

func TestHandleDelete(t *testing.T) {
	server, mux, cleanup := createTestServerAndMux(t)
	defer cleanup()

	t.Run("SuccessByID", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"points": []string{"1"},
		})
		req := httptest.NewRequest(http.MethodPost, "/collections/test/points/delete", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
		if server.GetCollection("test").Count() != 2 {
			t.Errorf("Expected count 2, got %d", server.GetCollection("test").Count())
		}
	})

	t.Run("SuccessByFilter", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"filter": map[string]interface{}{
				"must": []map[string]interface{}{
					{"key": "id", "match": map[string]interface{}{"value": "2"}},
				},
			},
		})
		req := httptest.NewRequest(http.MethodPost, "/collections/test/points/delete", bytes.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/collections/test/points/delete", bytes.NewReader([]byte("invalid")))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/collections/nonexistent/points/delete", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", w.Code)
		}
	})
}

func TestServerStopWithoutStart(t *testing.T) {
	server := NewServer(":8080", nil)
	ctx := context.Background()
	if err := server.Stop(ctx); err != nil {
		t.Errorf("Stop failed: %v", err)
	}
}
