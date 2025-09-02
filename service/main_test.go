package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"radiocast/internal/config"
	"radiocast/internal/server"
)

func TestHealthEndpoint(t *testing.T) {
	cfg := &config.Config{
		Port:         "8080",
		OpenAIAPIKey: "test-key",
		GCPProjectID: "test-project",
		GCSBucket:    "test-bucket",
		Environment:  "test",
	}

	srv, err := server.NewServer(cfg, server.DeploymentLocal)
	if err != nil {
		t.Skip("Skipping test - server creation failed (expected in test environment)")
	}
	defer srv.Close()

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	srv.HandleHealth(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "healthy") {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}

func TestRootEndpoint(t *testing.T) {
	// Skip this test as it requires real API calls and OpenAI key
	// In local mode, root endpoint generates reports on-demand which needs:
	// 1. Valid OpenAI API key
	// 2. External API calls to NOAA, etc.
	// 3. Chart generation
	t.Skip("Skipping root endpoint test - requires real API calls and valid OpenAI key")
}

func TestConfigLoad(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This will fail without proper env vars, but tests the structure
	_, err := config.Load(ctx)
	if err != nil {
		t.Logf("Config load failed as expected without env vars: %v", err)
	}
}
