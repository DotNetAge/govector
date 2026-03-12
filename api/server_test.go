package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DotNetAge/govector/core"
)

func createTestServerAndMux(t *testing.T) (*Server, *http.ServeMux, func()) {
	tempFile, err := os.CreateTemp("", "api-test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()

	store, err := core.NewStorage(tempPath)
	if err != nil {
		os.Remove(tempPath)
		t.Fatalf("Failed to create storage: %v", err)
	}

	server := NewServer(":8080", store)

	col, err := core.NewCollection("test", 3, core.Cosine, nil, false)
	if err != nil {
		store.Close()
		os.Remove(tempPath)
		t.Fatalf("Failed to create collection: %v", err)
	}

	points := []core.PointStruct{
		{ID: "1", Vector: []float32{1.0, 0.0, 0.0}},
		{ID: "2", Vector: []float32{0.0, 1.0, 0.0}},
		{ID: "3", Vector: []float32{0.0, 0.0, 1.0}},
	}
	if err := col.Upsert(points); err != nil {
		store.Close()
		os.Remove(tempPath)
		t.Fatalf("Failed to upsert points: %v", err)
	}

	server.AddCollection(col)

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /collections/{name}/points", server.handleUpsert)
	mux.HandleFunc("POST /collections/{name}/points/search", server.handleSearch)
	mux.HandleFunc("POST /collections/{name}/points/delete", server.handleDelete)

	cleanup := func() {
		store.Close()
		os.Remove(tempPath)
	}

	return server, mux, cleanup
}

func TestNewServer(t *testing.T) {
	tempFile, err := os.CreateTemp("", "api-test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	store, err := core.NewStorage(tempPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	server := NewServer(":8080", store)

	if server == nil {
		t.Error("Expected server to be non-nil")
	}

	if server.addr != ":8080" {
		t.Errorf("Expected addr :8080, got %s", server.addr)
	}

	if server.collections == nil {
		t.Error("Expected collections map to be initialized")
	}
}

func TestAddCollection(t *testing.T) {
	tempFile, err := os.CreateTemp("", "api-test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	store, err := core.NewStorage(tempPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	server := NewServer(":8080", store)

	col, err := core.NewCollection("test", 3, core.Cosine, nil, false)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	server.AddCollection(col)

	if server.collections["test"] == nil {
		t.Error("Expected collection to be added")
	}
}

func TestHandleUpsert(t *testing.T) {
	_, mux, cleanup := createTestServerAndMux(t)
	defer cleanup()

	t.Run("Success", func(t *testing.T) {
		newPoints := []core.PointStruct{
			{ID: "4", Vector: []float32{0.5, 0.5, 0.0}},
		}

		body, err := json.Marshal(map[string]interface{}{
			"points": newPoints,
		})
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest(http.MethodPut, "/collections/test/points", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/collections/test/points", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("CollectionNotFound", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"points": []core.PointStruct{},
		})

		req := httptest.NewRequest(http.MethodPut, "/collections/non-existent/points", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

func TestHandleSearch(t *testing.T) {
	_, mux, cleanup := createTestServerAndMux(t)
	defer cleanup()

	t.Run("Success", func(t *testing.T) {
		body, err := json.Marshal(map[string]interface{}{
			"vector": []float32{1.0, 0.0, 0.0},
			"limit":  2,
		})
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/collections/test/points/search", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v, body: %s", err, w.Body.String())
		}

		result := response["result"].([]interface{})
		if len(result) != 2 {
			t.Errorf("Expected 2 results, got %d", len(result))
		}
	})

	t.Run("DefaultLimit", func(t *testing.T) {
		body, err := json.Marshal(map[string]interface{}{
			"vector": []float32{1.0, 0.0, 0.0},
		})
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/collections/test/points/search", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/collections/test/points/search", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("CollectionNotFound", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"vector": []float32{1.0, 0.0, 0.0},
			"limit":  10,
		})

		req := httptest.NewRequest(http.MethodPost, "/collections/non-existent/points/search", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("WithFilter", func(t *testing.T) {
		body, err := json.Marshal(map[string]interface{}{
			"vector": []float32{1.0, 0.0, 0.0},
			"limit":  10,
			"filter": map[string]interface{}{
				"must": []map[string]interface{}{
					{"key": "category", "match": map[string]interface{}{"value": "test"}},
				},
			},
		})
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/collections/test/points/search", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

func TestHandleDelete(t *testing.T) {
	server, mux, cleanup := createTestServerAndMux(t)
	defer cleanup()

	t.Run("Success", func(t *testing.T) {
		body, err := json.Marshal(map[string]interface{}{
			"points": []string{"1", "2"},
		})
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/collections/test/points/delete", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		col := server.collections["test"]
		if col.Count() != 1 {
			t.Errorf("Expected count 1 after deletion, got %d", col.Count())
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/collections/test/points/delete", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("CollectionNotFound", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"points": []string{"1"},
		})

		req := httptest.NewRequest(http.MethodPost, "/collections/non-existent/points/delete", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("WithFilter", func(t *testing.T) {
		body, err := json.Marshal(map[string]interface{}{
			"filter": map[string]interface{}{
				"must": []map[string]interface{}{
					{"key": "category", "match": map[string]interface{}{"value": "test"}},
				},
			},
		})
		if err != nil {
			t.Fatalf("Failed to marshal request: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/collections/test/points/delete", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

func TestServerStop(t *testing.T) {
	server, _, cleanup := createTestServerAndMux(t)
	defer cleanup()

	ctx := context.Background()
	err := server.Stop(ctx)
	if err != nil {
		t.Errorf("Expected no error from Stop, got %v", err)
	}
}

func TestServerStopWithoutStart(t *testing.T) {
	tempFile, err := os.CreateTemp("", "api-test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	store, err := core.NewStorage(tempPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	server := NewServer(":8080", store)

	ctx := context.Background()
	err = server.Stop(ctx)
	if err != nil {
		t.Errorf("Expected no error from Stop without start, got %v", err)
	}
}
