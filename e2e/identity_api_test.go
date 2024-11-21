package e2e_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
)

type mockIdentityServer struct {
	server    *httptest.Server
	responses map[string]interface{} // request payload hash -> response
	mu        sync.RWMutex
}

func setupIdentityServer() *mockIdentityServer {
	m := &mockIdentityServer{
		responses: make(map[string]interface{}),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Read and hash the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Get response for this exact request payload
		m.mu.RLock()
		response, exists := m.responses[string(body)]
		m.mu.RUnlock()

		if !exists {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err = json.NewEncoder(w).Encode(response); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}))

	m.server = server
	return m
}

// SetRequestResponse sets a response for an exact request payload
func (m *mockIdentityServer) SetRequestResponse(request, response any) error {
	reqBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[string(reqBytes)] = response
	return nil
}

func (m *mockIdentityServer) URL() string {
	return m.server.URL
}

func (m *mockIdentityServer) Close() {
	m.server.Close()
}
