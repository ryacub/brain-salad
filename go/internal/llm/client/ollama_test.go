package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOllamaClient_NewClient(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		timeout     time.Duration
		wantBaseURL string
		wantTimeout time.Duration
	}{
		{
			name:        "default values",
			baseURL:     "",
			timeout:     0,
			wantBaseURL: "http://localhost:11434",
			wantTimeout: 30 * time.Second,
		},
		{
			name:        "custom values",
			baseURL:     "http://custom:11434",
			timeout:     60 * time.Second,
			wantBaseURL: "http://custom:11434",
			wantTimeout: 60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOllamaClient(tt.baseURL, tt.timeout)
			if client == nil {
				t.Fatal("expected client to be non-nil")
			}
			if client.baseURL != tt.wantBaseURL {
				t.Errorf("expected baseURL %s, got %s", tt.wantBaseURL, client.baseURL)
			}
			if client.timeout != tt.wantTimeout {
				t.Errorf("expected timeout %v, got %v", tt.wantTimeout, client.timeout)
			}
		})
	}
}

func TestOllamaClient_Generate_Success(t *testing.T) {
	// Create a mock server that returns a successful response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("expected path /api/generate, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		// Verify request body
		var req GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Model == "" {
			t.Error("expected model to be set")
		}
		if req.Prompt == "" {
			t.Error("expected prompt to be set")
		}
		if req.Stream {
			t.Error("expected stream to be false")
		}

		// Return mock response
		resp := GenerateResponse{
			Model:     req.Model,
			CreatedAt: time.Now(),
			Response:  "This is a test response",
			Done:      true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, 5*time.Second)
	ctx := context.Background()

	resp, err := client.Generate(ctx, GenerateRequest{
		Model:  "llama2",
		Prompt: "test prompt",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected response to be non-nil")
	}
	if resp.Response != "This is a test response" {
		t.Errorf("expected response text 'This is a test response', got %s", resp.Response)
	}
	if !resp.Done {
		t.Error("expected done to be true")
	}
}

func TestOllamaClient_Generate_Timeout(t *testing.T) {
	// Create a mock server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, 100*time.Millisecond)
	ctx := context.Background()

	_, err := client.Generate(ctx, GenerateRequest{
		Model:  "llama2",
		Prompt: "test prompt",
	})

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	// Check if error is timeout-related
	if !isTimeoutError(err) {
		t.Errorf("expected timeout error, got %v", err)
	}
}

func TestOllamaClient_Generate_ConnectionError(t *testing.T) {
	// Use an invalid URL to simulate connection error
	client := NewOllamaClient("http://localhost:99999", 5*time.Second)
	ctx := context.Background()

	_, err := client.Generate(ctx, GenerateRequest{
		Model:  "llama2",
		Prompt: "test prompt",
	})

	if err == nil {
		t.Fatal("expected connection error, got nil")
	}
}

func TestOllamaClient_Generate_ModelNotFound(t *testing.T) {
	// Create a mock server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "model not found"}`))
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, 5*time.Second)
	ctx := context.Background()

	_, err := client.Generate(ctx, GenerateRequest{
		Model:  "nonexistent-model",
		Prompt: "test prompt",
	})

	if err == nil {
		t.Fatal("expected model not found error, got nil")
	}
}

func TestOllamaClient_ListModels(t *testing.T) {
	// Create a mock server that returns a list of models
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Errorf("expected path /api/tags, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}

		resp := map[string]interface{}{
			"models": []map[string]string{
				{"name": "llama2"},
				{"name": "mistral"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, 5*time.Second)
	ctx := context.Background()

	models, err := client.ListModels(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(models) != 2 {
		t.Errorf("expected 2 models, got %d", len(models))
	}
	if models[0] != "llama2" {
		t.Errorf("expected first model to be 'llama2', got %s", models[0])
	}
}

func TestOllamaClient_HealthCheck(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantError  bool
	}{
		{
			name:       "healthy",
			statusCode: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "unhealthy",
			statusCode: http.StatusServiceUnavailable,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/tags" {
					t.Errorf("expected path /api/tags, got %s", r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := NewOllamaClient(server.URL, 5*time.Second)
			ctx := context.Background()

			err := client.HealthCheck(ctx)
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestOllamaClient_Generate_ContextCancellation(t *testing.T) {
	// Create a mock server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewOllamaClient(server.URL, 30*time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err := client.Generate(ctx, GenerateRequest{
		Model:  "llama2",
		Prompt: "test prompt",
	})

	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}

// Helper function to check if an error is a timeout error
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return errMsg == "context deadline exceeded" ||
		errMsg == "i/o timeout" ||
		errMsg == "Client.Timeout exceeded while awaiting headers" ||
		// Handle wrapped errors
		contains(errMsg, "context deadline exceeded") ||
		contains(errMsg, "Client.Timeout exceeded")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
