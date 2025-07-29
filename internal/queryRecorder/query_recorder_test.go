package queryRecorder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	qr := New()
	if qr == nil {
		t.Fatal("New() returned nil")
	}
	if qr.queries == nil {
		t.Fatal("queries map is nil")
	}
	if qr.heap == nil {
		t.Fatal("heap is nil")
	}
	if len(qr.queries) != 0 {
		t.Fatalf("expected 0 queries, got %d", len(qr.queries))
	}
}

func TestAddNewQuery(t *testing.T) {
	qr := New()
	query := "query { user { id name } }"

	qr.Add(query)

	// Wait for goroutine to complete
	time.Sleep(10 * time.Millisecond)

	if len(qr.queries) != 1 {
		t.Fatalf("expected 1 query, got %d", len(qr.queries))
	}

	info, exists := qr.queries[query]
	if !exists {
		t.Fatal("query not found in map")
	}
	if info.Query != query {
		t.Fatalf("expected query %s, got %s", query, info.Query)
	}
	if info.Count != 1 {
		t.Fatalf("expected count 1, got %d", info.Count)
	}
}

func TestAddExistingQuery(t *testing.T) {
	qr := New()
	query := "query { user { id } }"

	// Add query twice
	qr.Add(query)
	qr.Add(query)

	// Wait for goroutines to complete
	time.Sleep(20 * time.Millisecond)

	if len(qr.queries) != 1 {
		t.Fatalf("expected 1 query, got %d", len(qr.queries))
	}

	info, exists := qr.queries[query]
	if !exists {
		t.Fatal("query not found in map")
	}
	if info.Count != 2 {
		t.Fatalf("expected count 2, got %d", info.Count)
	}
}

func TestCleanupWithHeap(t *testing.T) {
	qr := New()

	// Add more queries than the limit
	for i := 0; i < MaxQueries+10; i++ {
		query := fmt.Sprintf("query { user%d { id } }", i)
		qr.Add(query)
	}

	// Wait for goroutines to complete
	time.Sleep(100 * time.Millisecond)

	if len(qr.queries) != MaxQueries {
		t.Fatalf("expected %d queries after cleanup, got %d", MaxQueries, len(qr.queries))
	}
}

func TestHeapMaintainsOrder(t *testing.T) {
	qr := New()

	// Add queries with different frequencies
	queries := []string{
		"query { user1 { id } }", // will be added once
		"query { user2 { id } }", // will be added twice
		"query { user3 { id } }", // will be added once
	}

	// Add queries
	qr.Add(queries[0]) // count = 1
	qr.Add(queries[1]) // count = 1
	qr.Add(queries[1]) // count = 2
	qr.Add(queries[2]) // count = 1

	// Wait for goroutines to complete
	time.Sleep(50 * time.Millisecond)

	// Verify all queries are present
	if len(qr.queries) != 3 {
		t.Fatalf("expected 3 queries, got %d", len(qr.queries))
	}

	// Check that the heap is properly ordered
	// The query with count=1 should be at the top of the heap
	if qr.heap.Len() != 3 {
		t.Fatalf("expected heap size 3, got %d", qr.heap.Len())
	}

	// Verify the least frequent query is at the top
	leastFrequent := (*qr.heap)[0]
	if leastFrequent.Count != 1 {
		t.Fatalf("expected least frequent query to have count 1, got %d", leastFrequent.Count)
	}
}

func TestGetQueries(t *testing.T) {
	qr := New()

	// Add some queries
	queries := []string{
		"query { user1 { id } }",
		"query { user2 { id } }",
		"query { user3 { id } }",
	}

	for _, query := range queries {
		qr.Add(query)
	}

	// Wait for goroutines to complete
	time.Sleep(30 * time.Millisecond)

	result := qr.GetQueries()
	if len(result) != 3 {
		t.Fatalf("expected 3 queries, got %d", len(result))
	}

	// Verify all queries are present
	querySet := make(map[string]bool)
	for _, info := range result {
		querySet[info.Query] = true
	}

	for _, query := range queries {
		if !querySet[query] {
			t.Fatalf("query %s not found in result", query)
		}
	}
}

func TestHandler(t *testing.T) {
	qr := New()

	// Add a query
	query := "query { user { id } }"
	qr.Add(query)

	// Wait for goroutine to complete
	time.Sleep(10 * time.Millisecond)

	// Create test request
	req := httptest.NewRequest("GET", "/queries", nil)
	w := httptest.NewRecorder()

	// Call handler
	qr.Handler().ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Check response structure
	if _, exists := response["queries"]; !exists {
		t.Fatal("response missing 'queries' field")
	}
	if _, exists := response["total_count"]; !exists {
		t.Fatal("response missing 'total_count' field")
	}
	if _, exists := response["max_queries"]; !exists {
		t.Fatal("response missing 'max_queries' field")
	}
	if _, exists := response["recorded_at"]; !exists {
		t.Fatal("response missing 'recorded_at' field")
	}

	// Check values
	totalCount := response["total_count"].(float64)
	if totalCount != 1 {
		t.Fatalf("expected total_count 1, got %v", totalCount)
	}

	maxQueries := response["max_queries"].(float64)
	if maxQueries != float64(MaxQueries) {
		t.Fatalf("expected max_queries %d, got %v", MaxQueries, maxQueries)
	}
}

func TestHandlerMethodNotAllowed(t *testing.T) {
	qr := New()

	// Create test request with POST method
	req := httptest.NewRequest("POST", "/queries", nil)
	w := httptest.NewRecorder()

	// Call handler
	qr.Handler().ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", w.Code)
	}
}

func TestConcurrentAccess(t *testing.T) {
	qr := New()
	var wg sync.WaitGroup

	// Add queries concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			query := fmt.Sprintf("query { user%d { id } }", id)
			qr.Add(query)
		}(i)
	}

	wg.Wait()

	// Wait a bit more for goroutines to complete
	time.Sleep(50 * time.Millisecond)

	// Verify we have queries (some may be cleaned up if over limit)
	if len(qr.queries) == 0 {
		t.Fatal("no queries recorded after concurrent access")
	}

	// Verify we don't exceed the limit
	if len(qr.queries) > MaxQueries {
		t.Fatalf("exceeded max queries limit: %d > %d", len(qr.queries), MaxQueries)
	}
}

func TestHeapIndexTracking(t *testing.T) {
	qr := New()

	// Add a query
	query := "query { user { id } }"
	qr.Add(query)

	// Wait for goroutine to complete
	time.Sleep(10 * time.Millisecond)

	info, exists := qr.queries[query]
	if !exists {
		t.Fatal("query not found")
	}

	// Check that index is set
	if info.index < 0 {
		t.Fatal("heap index not set")
	}

	// Update the query count
	qr.Add(query)

	// Wait for goroutine to complete
	time.Sleep(10 * time.Millisecond)

	// Verify the heap is still properly ordered
	if qr.heap.Len() != 1 {
		t.Fatalf("expected heap size 1, got %d", qr.heap.Len())
	}

	// The query should still be in the heap with updated count
	if (*qr.heap)[0].Count != 2 {
		t.Fatalf("expected count 2, got %d", (*qr.heap)[0].Count)
	}
}
