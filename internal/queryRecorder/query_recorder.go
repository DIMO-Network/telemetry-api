package queryRecorder

import (
	"container/heap"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/DIMO-Network/telemetry-api/internal/dtcmiddleware"
)

const (
	// MaxQueries is the maximum number of queries to keep
	MaxQueries = 1000
)

// QueryRecorder tracks unique raw queries that have been seen
type QueryRecorder struct {
	queries map[string]*QueryInfo
	heap    *queryHeap
	mu      sync.RWMutex
}

// QueryInfo contains information about a recorded query
type QueryInfo struct {
	Query string              `json:"query"`
	Count int                 `json:"count"`
	Devs  map[string]struct{} `json:"devs"`

	index int // heap index for efficient updates
}

// queryHeap implements heap.Interface for maintaining least frequent queries
type queryHeap []*QueryInfo

func (h queryHeap) Len() int { return len(h) }

func (h queryHeap) Less(i, j int) bool {
	// Min heap: least frequent queries first
	return h[i].Count < h[j].Count
}

func (h queryHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *queryHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*QueryInfo)
	item.index = n
	*h = append(*h, item)
}

func (h *queryHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*h = old[0 : n-1]
	return item
}

// New creates a new QueryRecorder instance
func New() *QueryRecorder {
	h := &queryHeap{}
	heap.Init(h)
	return &QueryRecorder{
		queries: make(map[string]*QueryInfo),
		heap:    h,
	}
}

// cleanup removes least frequent queries to stay under the limit
func (qr *QueryRecorder) cleanup() {
	for len(qr.queries) > MaxQueries {
		leastFrequent := heap.Pop(qr.heap).(*QueryInfo)
		delete(qr.queries, leastFrequent.Query)
	}
}

// Add records a new query or updates an existing one
func (qr *QueryRecorder) Add(query string, devID string) {
	go func() {
		qr.mu.Lock()
		defer qr.mu.Unlock()

		if info, exists := qr.queries[query]; exists {
			// Update existing query count
			info.Count++
			info.Devs[devID] = struct{}{}
			// Fix heap after count change
			heap.Fix(qr.heap, info.index)
		} else {
			// Add new query
			info := &QueryInfo{
				Query: query,
				Count: 1,
				Devs:  map[string]struct{}{devID: {}},
			}
			qr.queries[query] = info
			heap.Push(qr.heap, info)

			// Clean up if we exceed the limit
			if len(qr.queries) > MaxQueries {
				qr.cleanup()
			}
		}
	}()
}

// GetQueries returns a copy of all recorded queries
func (qr *QueryRecorder) GetQueries() []QueryInfo {
	qr.mu.RLock()
	defer qr.mu.RUnlock()

	queries := make([]QueryInfo, 0, len(qr.queries))
	for _, info := range qr.queries {
		queries = append(queries, *info)
	}

	return queries
}

// Handler returns an HTTP handler that serves the list of recorded queries
func (qr *QueryRecorder) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		queries := qr.GetQueries()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")

		response := map[string]interface{}{
			"queries":     queries,
			"total_count": len(queries),
			"max_queries": MaxQueries,
			"recorded_at": time.Now().UTC(),
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})
}

// QueryRecordingExtension creates a GraphQL extension that records queries
type QueryRecordingExtension struct {
	Recorder *QueryRecorder
}

func (QueryRecordingExtension) ExtensionName() string {
	return "QueryRecorder"
}

func (QueryRecordingExtension) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (q QueryRecordingExtension) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	op := graphql.GetOperationContext(ctx)
	if op != nil && op.RawQuery != "" {
		developerID, _, _ := dtcmiddleware.GetSubjectAndTokenID(ctx)
		q.Recorder.Add(op.RawQuery, developerID)
	}
	return next(ctx)
}
