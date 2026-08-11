package core

import (
	"strings"
	"testing"
)

func TestHNSWIndex(t *testing.T) {
	// Test Cosine distance
	t.Run("CosineDistance", func(t *testing.T) {
		index := NewHNSWIndex(Cosine)

		// Test Upsert
		points := []PointStruct{
			{ID: "1", Vector: []float32{1.0, 0.0, 0.0}, Payload: map[string]interface{}{"category": "A"}},
			{ID: "2", Vector: []float32{0.0, 1.0, 0.0}, Payload: map[string]interface{}{"category": "B"}},
			{ID: "3", Vector: []float32{0.0, 0.0, 1.0}, Payload: map[string]interface{}{"category": "A"}},
			{ID: "4", Vector: []float32{0.5, 0.5, 0.0}, Payload: map[string]interface{}{"category": "A"}},
		}

		err := index.Upsert(points)
		if err != nil {
			t.Fatalf("Failed to upsert points: %v", err)
		}

		// Test Count
		if index.Count() != 4 {
			t.Errorf("Expected count 4, got %d", index.Count())
		}

		// Test Search
		query := []float32{1.0, 0.0, 0.0}
		results, err := index.Search(query, nil, 2)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}

		if results[0].ID != "1" {
			t.Errorf("Expected first result to be ID 1, got %s", results[0].ID)
		}

		// Test Search with filter
		filter := &Filter{
			Must: []Condition{
				{Key: "category", Match: MatchValue{Value: "A"}},
			},
		}

		filteredResults, err := index.Search(query, filter, 2)
		if err != nil {
			t.Fatalf("Failed to search with filter: %v", err)
		}

		if len(filteredResults) != 2 {
			t.Errorf("Expected 2 filtered results, got %d", len(filteredResults))
		}

		// Test Delete
		err = index.Delete("1")
		if err != nil {
			t.Fatalf("Failed to delete point: %v", err)
		}

		if index.Count() != 3 {
			t.Errorf("Expected count 3 after deletion, got %d", index.Count())
		}

		// Test Delete non-existent point
		err = index.Delete("non-existent")
		if err == nil {
			t.Error("Expected error when deleting non-existent point")
		}

		// Test DeleteByFilter
		deletedIDs, err := index.DeleteByFilter(filter)
		if err != nil {
			t.Fatalf("Failed to delete by filter: %v", err)
		}

		if len(deletedIDs) != 2 {
			t.Errorf("Expected to delete 2 points by filter, got %d", len(deletedIDs))
		}

		if index.Count() != 1 {
			t.Errorf("Expected count 1 after filter deletion, got %d", index.Count())
		}
	})

	// Test Euclidean distance
	t.Run("EuclideanDistance", func(t *testing.T) {
		index := NewHNSWIndex(Euclid)

		points := []PointStruct{
			{ID: "1", Vector: []float32{0.0, 0.0}},
			{ID: "2", Vector: []float32{1.0, 0.0}},
			{ID: "3", Vector: []float32{1.0, 1.0}},
		}

		err := index.Upsert(points)
		if err != nil {
			t.Fatalf("Failed to upsert points: %v", err)
		}

		query := []float32{0.0, 0.0}
		results, err := index.Search(query, nil, 3)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}

		if results[0].ID != "1" {
			t.Errorf("Expected first result to be ID 1, got %s", results[0].ID)
		}
	})

	// Test Dot product
	t.Run("DotProduct", func(t *testing.T) {
		index := NewHNSWIndex(Dot)

		points := []PointStruct{
			{ID: "1", Vector: []float32{1.0, 0.0}},
			{ID: "2", Vector: []float32{0.0, 1.0}},
			{ID: "3", Vector: []float32{1.0, 1.0}},
		}

		err := index.Upsert(points)
		if err != nil {
			t.Fatalf("Failed to upsert points: %v", err)
		}

		query := []float32{1.0, 0.0}
		results, err := index.Search(query, nil, 2)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}

		// Check that ID 1 is in results (order may vary due to HNSW)
		found := false
		for _, r := range results {
			if r.ID == "1" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected ID 1 to be in results, got %v", results)
		}
	})
}

// TestHNSWIndexNilGraph verifies that all public methods return an error
// instead of panicking when the underlying HNSW graph is nil.
// This covers the P0 fix: defensive nil checks in Upsert/Search/Delete/DeleteByFilter.
func TestHNSWIndexNilGraph(t *testing.T) {
	t.Run("NilGraph_UpsertReturnsError", func(t *testing.T) {
		idx := &HNSWIndex{points: make(map[string]*PointStruct)}
		err := idx.Upsert([]PointStruct{{ID: "1", Vector: []float32{1.0}}})
		if err == nil {
			t.Fatal("expected error when upserting with nil graph")
		}
		if !strings.Contains(err.Error(), "nil") {
			t.Errorf("error should mention nil graph, got: %v", err)
		}
	})

	t.Run("NilGraph_SearchReturnsError", func(t *testing.T) {
		idx := &HNSWIndex{points: make(map[string]*PointStruct)}
		_, err := idx.Search([]float32{1.0}, nil, 1)
		if err == nil {
			t.Fatal("expected error when searching with nil graph")
		}
	})

	t.Run("NilGraph_DeleteReturnsError", func(t *testing.T) {
		idx := &HNSWIndex{points: make(map[string]*PointStruct)}
		err := idx.Delete("1")
		if err == nil {
			t.Fatal("expected error when deleting with nil graph")
		}
	})

	t.Run("NilGraph_DeleteByFilterReturnsError", func(t *testing.T) {
		idx := &HNSWIndex{points: make(map[string]*PointStruct)}
		filter := &Filter{
			Must: []Condition{{Key: "x", Match: MatchValue{Value: "y"}}},
		}
		_, err := idx.DeleteByFilter(filter)
		if err == nil {
			t.Fatal("expected error when DeleteByFilter with nil graph")
		}
	})

	t.Run("NilGraph_CountIsZero", func(t *testing.T) {
		idx := &HNSWIndex{points: make(map[string]*PointStruct)}
		if idx.Count() != 0 {
			t.Errorf("expected count 0 for nil graph, got %d", idx.Count())
		}
	})
}

// TestHNSWIndexOverwriteUpdate 验证对已存在 ID 做覆盖更新（单点）不会 panic。
// 这是内嵌补丁 fork 的回归用例：上游 coder/hnsw v0.6.1 的 Add 在替换场景会
// panic("node not added")，此前由 workspace 外部 replace 兜底修复。
func TestHNSWIndexOverwriteUpdate(t *testing.T) {
	idx := NewHNSWIndex(Cosine)

	if err := idx.Upsert([]PointStruct{
		{ID: "a", Vector: []float32{1.0, 0.0}, Payload: map[string]interface{}{"v": 1}},
	}); err != nil {
		t.Fatalf("failed to upsert initial point: %v", err)
	}

	// 覆盖更新同一个 ID：旧向量 (1,0) 替换为新向量 (0,1)
	if err := idx.Upsert([]PointStruct{
		{ID: "a", Vector: []float32{0.0, 1.0}, Payload: map[string]interface{}{"v": 2}},
	}); err != nil {
		t.Fatalf("failed to upsert overwrite: %v", err)
	}

	if idx.Count() != 1 {
		t.Fatalf("expected count 1 after overwrite, got %d", idx.Count())
	}

	// 覆盖后，与新向量一致的查询应命中该点，且新 payload 生效
	results, err := idx.Search([]float32{0.0, 1.0}, nil, 1)
	if err != nil {
		t.Fatalf("failed to search: %v", err)
	}
	if len(results) != 1 || results[0].ID != "a" {
		t.Fatalf("expected point a as nearest neighbor, got %v", results)
	}
	if results[0].Payload["v"] != 2 {
		t.Errorf("expected payload v=2 after overwrite, got %v", results[0].Payload)
	}
}

// TestHNSWIndexReinsertAfterEmpty 验证图被删空后重新 Upsert 不会 panic。
// 内嵌补丁 fork 修复了 assertDims 在「layers 残留但无节点」的空图上的维度检查，
// 上游 v0.6.1 在此场景会对 nil entry 解引用 panic。
func TestHNSWIndexReinsertAfterEmpty(t *testing.T) {
	idx := NewHNSWIndex(Cosine)

	if err := idx.Upsert([]PointStruct{
		{ID: "1", Vector: []float32{1.0, 0.0}, Payload: map[string]interface{}{"cat": "A"}},
		{ID: "2", Vector: []float32{0.0, 1.0}, Payload: map[string]interface{}{"cat": "B"}},
	}); err != nil {
		t.Fatalf("failed to upsert points: %v", err)
	}

	// 逐类删光所有点，使图进入空态
	for _, cat := range []string{"A", "B"} {
		if _, err := idx.DeleteByFilter(&Filter{
			Must: []Condition{{Key: "cat", Match: MatchValue{Value: cat}}},
		}); err != nil {
			t.Fatalf("failed to delete by filter %s: %v", cat, err)
		}
	}
	if idx.Count() != 0 {
		t.Fatalf("expected count 0 after deletion, got %d", idx.Count())
	}

	// 删空后重新插入新点，不应 panic
	if err := idx.Upsert([]PointStruct{
		{ID: "3", Vector: []float32{0.0, 0.0, 1.0}},
	}); err != nil {
		t.Fatalf("failed to re-upsert after empty: %v", err)
	}
	if idx.Count() != 1 {
		t.Fatalf("expected count 1 after re-upsert, got %d", idx.Count())
	}
}
