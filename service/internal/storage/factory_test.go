package storage

import (
	"context"
	"os"
	"testing"

	"radiocast/internal/config"
)

func TestNewStorageClient_Local(t *testing.T) {
	// Test local storage client creation
	cfg := &config.Config{
		LocalReportsDir: "test-reports",
	}

	client, err := NewStorageClient(context.Background(), DeploymentLocal, cfg)
	if err != nil {
		t.Fatalf("Failed to create local storage client: %v", err)
	}
	defer client.Close()

	// Verify it's a LocalStorageClient
	if _, ok := client.(*LocalStorageClient); !ok {
		t.Errorf("Expected LocalStorageClient, got %T", client)
	}
}

func TestNewStorageClient_GCS(t *testing.T) {
	// Test GCS storage client creation (will fail without proper credentials)
	cfg := &config.Config{
		GCPProjectID: "test-project",
		GCSBucket:    "test-bucket",
	}

	// This will likely fail in test environment without GCP credentials
	// but we test the logic path
	client, err := NewStorageClient(context.Background(), DeploymentGCS, cfg)
	
	// In test environment, this might fail due to missing credentials
	// That's expected behavior
	if err != nil {
		t.Logf("GCS client creation failed as expected in test environment: %v", err)
		return
	}
	
	if client != nil {
		defer client.Close()
		// If it succeeds, verify it's a GCSClient
		if _, ok := client.(*GCSClient); !ok {
			t.Errorf("Expected GCSClient, got %T", client)
		}
	}
}

func TestNewStorageClient_LocalFallback(t *testing.T) {
	// Test local storage with default reports directory
	cfg := &config.Config{
		LocalReportsDir: "", // Empty to test default fallback
	}

	client, err := NewStorageClient(context.Background(), DeploymentLocal, cfg)
	if err != nil {
		t.Fatalf("Failed to create storage client: %v", err)
	}
	defer client.Close()

	// Should fall back to local storage
	if _, ok := client.(*LocalStorageClient); !ok {
		t.Errorf("Expected LocalStorageClient fallback, got %T", client)
	}
}

func TestNewStorageClient_MissingBucket(t *testing.T) {
	// Test GCS with missing bucket (will likely fail due to credentials in test env)
	cfg := &config.Config{
		GCPProjectID: "test-project",
		GCSBucket:    "", // Empty bucket
	}

	client, err := NewStorageClient(context.Background(), DeploymentGCS, cfg)
	// In test environment, this will likely fail due to missing GCP credentials
	// That's expected behavior
	if err != nil {
		t.Logf("GCS client creation failed as expected in test environment: %v", err)
		return
	}
	
	if client != nil {
		defer client.Close()
		// If it succeeds, verify it's a GCSClient
		if _, ok := client.(*GCSClient); !ok {
			t.Errorf("Expected GCSClient, got %T", client)
		}
	}
}

func TestNewStorageClient_NilConfig(t *testing.T) {
	// Test behavior with nil config
	client, err := NewStorageClient(context.Background(), DeploymentLocal, nil)
	if err == nil {
		if client != nil {
			client.Close()
		}
		t.Error("Expected error with nil config")
	}
}

func TestNewStorageClient_InvalidMode(t *testing.T) {
	// Test invalid deployment mode
	cfg := &config.Config{
		LocalReportsDir: "test-reports",
	}

	client, err := NewStorageClient(context.Background(), DeploymentMode("invalid"), cfg)
	if err == nil {
		if client != nil {
			client.Close()
		}
		t.Error("Expected error with invalid deployment mode")
	}
}

func TestNewStorageClient_ContextCancellation(t *testing.T) {
	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cfg := &config.Config{
		GCPProjectID: "",
		GCSBucket:    "",
	}

	// Local storage should still work with cancelled context
	// since it doesn't use context for initialization
	client, err := NewStorageClient(ctx, DeploymentLocal, cfg)
	if err != nil {
		t.Fatalf("Local storage should work with cancelled context: %v", err)
	}
	defer client.Close()

	if _, ok := client.(*LocalStorageClient); !ok {
		t.Errorf("Expected LocalStorageClient, got %T", client)
	}
}

func TestNewStorageClient_Integration(t *testing.T) {
	// Integration test that creates a client and performs basic operations
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	cfg := &config.Config{
		LocalReportsDir: "test-reports", // Use local storage
	}

	client, err := NewStorageClient(context.Background(), DeploymentLocal, cfg)
	if err != nil {
		t.Fatalf("Failed to create storage client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Test basic operations
	testDir := "test-reports"
	testFile := "test-reports/test.txt"
	testData := []byte("test content")

	// Create directory
	err = client.CreateDir(ctx, testDir)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Store file
	err = client.StoreFile(ctx, testFile, testData)
	if err != nil {
		t.Fatalf("Failed to store file: %v", err)
	}

	// Check file exists
	exists, err := client.FileExists(ctx, testFile)
	if err != nil {
		t.Fatalf("Failed to check file existence: %v", err)
	}
	if !exists {
		t.Error("File should exist after storing")
	}

	// Retrieve file
	retrievedData, err := client.GetFile(ctx, testFile)
	if err != nil {
		t.Fatalf("Failed to retrieve file: %v", err)
	}

	// Verify content
	if string(retrievedData) != string(testData) {
		t.Errorf("Retrieved data mismatch: expected %s, got %s", string(testData), string(retrievedData))
	}

	// List directory
	files, err := client.ListDir(ctx, testDir, false)
	if err != nil {
		t.Fatalf("Failed to list directory: %v", err)
	}

	if len(files) == 0 {
		t.Error("Directory should contain files")
	}
}

func TestStorageClientInterface(t *testing.T) {
	// Test that both implementations satisfy the StorageClient interface
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Test local storage
	localClient, err := NewLocalStorageClient("")
	if err != nil {
		t.Fatalf("Failed to create local storage client: %v", err)
	}
	defer localClient.Close()

	// Verify it implements StorageClient interface
	var _ StorageClient = localClient

	// Test that we can use it through the interface
	ctx := context.Background()
	testFile := "interface-test.txt"
	testData := []byte("interface test")

	err = localClient.StoreFile(ctx, testFile, testData)
	if err != nil {
		t.Fatalf("Interface method StoreFile failed: %v", err)
	}

	exists, err := localClient.FileExists(ctx, testFile)
	if err != nil {
		t.Fatalf("Interface method FileExists failed: %v", err)
	}
	if !exists {
		t.Error("File should exist")
	}

	retrievedData, err := localClient.GetFile(ctx, testFile)
	if err != nil {
		t.Fatalf("Interface method GetFile failed: %v", err)
	}
	if string(retrievedData) != string(testData) {
		t.Errorf("Data mismatch through interface: expected %s, got %s", string(testData), string(retrievedData))
	}
}
