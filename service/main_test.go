package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"radiocast/internal/config"
)

func TestHealthEndpoint(t *testing.T) {
	cfg := &config.Config{
		Port:         "8080",
		OpenAIAPIKey: "test-key",
		GCPProjectID: "test-project",
		GCSBucket:    "test-bucket",
		Environment:  "test",
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Skip("Skipping test - GCS client creation failed (expected in test environment)")
	}
	defer server.Close()

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.handleHealth(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "healthy") {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}

func TestRootEndpoint(t *testing.T) {
	cfg := &config.Config{
		Port:         "8080",
		OpenAIAPIKey: "test-key",
		GCPProjectID: "test-project",
		GCSBucket:    "test-bucket",
		Environment:  "test",
	}

	server, err := NewServer(cfg)
	if err != nil {
		t.Skip("Skipping test - GCS client creation failed (expected in test environment)")
	}
	defer server.Close()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.handleRoot(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "Radio Propagation Service") {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}

func TestConfigLoad(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This will fail without proper env vars, but tests the structure
	_, err := config.Load(ctx)
	if err == nil {
		t.Error("Expected error when loading config without required env vars")
	}
}
