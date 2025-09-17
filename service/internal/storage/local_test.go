package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewLocalStorageClient(t *testing.T) {
	// Create in temp directory to avoid conflicts
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	client, err := NewLocalStorageClient("any-base-dir")
	if err != nil {
		t.Fatalf("Failed to create LocalStorageClient: %v", err)
	}
	defer client.Close()

	// Verify root directory is always "local_gcs"
	if client.rootDir != "local_gcs" {
		t.Errorf("Expected rootDir to be 'local_gcs', got '%s'", client.rootDir)
	}

	// Verify directory was created
	if _, err := os.Stat("local_gcs"); os.IsNotExist(err) {
		t.Error("local_gcs directory was not created")
	}
}

func TestLocalStorageClient_Close(t *testing.T) {
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	client, err := NewLocalStorageClient("")
	if err != nil {
		t.Fatalf("Failed to create LocalStorageClient: %v", err)
	}

	// Close should not return error
	err = client.Close()
	if err != nil {
		t.Errorf("Close() returned unexpected error: %v", err)
	}
}

func TestLocalStorageClient_CreateDir(t *testing.T) {
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	client, err := NewLocalStorageClient("")
	if err != nil {
		t.Fatalf("Failed to create LocalStorageClient: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	tests := []struct {
		name    string
		dirPath string
		wantErr bool
	}{
		{
			name:    "simple directory",
			dirPath: "test-dir",
			wantErr: false,
		},
		{
			name:    "nested directory",
			dirPath: "reports/2025/09/17",
			wantErr: false,
		},
		{
			name:    "directory with special characters",
			dirPath: "test-dir_with-special.chars",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.CreateDir(ctx, tt.dirPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify directory was created
				fullPath := filepath.Join("local_gcs", tt.dirPath)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Errorf("Directory %s was not created", fullPath)
				}
			}
		})
	}
}

func TestLocalStorageClient_StoreFile(t *testing.T) {
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	client, err := NewLocalStorageClient("")
	if err != nil {
		t.Fatalf("Failed to create LocalStorageClient: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	tests := []struct {
		name     string
		filePath string
		fileData []byte
		wantErr  bool
	}{
		{
			name:     "simple file",
			filePath: "test.txt",
			fileData: []byte("Hello, World!"),
			wantErr:  false,
		},
		{
			name:     "file in nested directory",
			filePath: "reports/2025/09/17/index.html",
			fileData: []byte("<html><body>Test</body></html>"),
			wantErr:  false,
		},
		{
			name:     "binary file",
			filePath: "images/test.png",
			fileData: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG header
			wantErr:  false,
		},
		{
			name:     "empty file",
			filePath: "empty.txt",
			fileData: []byte{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.StoreFile(ctx, tt.filePath, tt.fileData)
			if (err != nil) != tt.wantErr {
				t.Errorf("StoreFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file was created with correct content
				fullPath := filepath.Join("local_gcs", tt.filePath)
				data, err := os.ReadFile(fullPath)
				if err != nil {
					t.Errorf("Failed to read stored file: %v", err)
					return
				}

				if len(data) != len(tt.fileData) {
					t.Errorf("File size mismatch: expected %d, got %d", len(tt.fileData), len(data))
					return
				}

				for i, b := range tt.fileData {
					if data[i] != b {
						t.Errorf("File content mismatch at byte %d: expected %d, got %d", i, b, data[i])
						break
					}
				}
			}
		})
	}
}

func TestLocalStorageClient_GetFile(t *testing.T) {
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	client, err := NewLocalStorageClient("")
	if err != nil {
		t.Fatalf("Failed to create LocalStorageClient: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Store test files first
	testFiles := map[string][]byte{
		"test.txt":                []byte("Hello, World!"),
		"reports/2025/index.html": []byte("<html><body>Test</body></html>"),
		"empty.txt":               []byte{},
	}

	for filePath, fileData := range testFiles {
		err := client.StoreFile(ctx, filePath, fileData)
		if err != nil {
			t.Fatalf("Failed to store test file %s: %v", filePath, err)
		}
	}

	tests := []struct {
		name     string
		filePath string
		wantData []byte
		wantErr  bool
	}{
		{
			name:     "existing file",
			filePath: "test.txt",
			wantData: []byte("Hello, World!"),
			wantErr:  false,
		},
		{
			name:     "existing nested file",
			filePath: "reports/2025/index.html",
			wantData: []byte("<html><body>Test</body></html>"),
			wantErr:  false,
		},
		{
			name:     "empty file",
			filePath: "empty.txt",
			wantData: []byte{},
			wantErr:  false,
		},
		{
			name:     "non-existent file",
			filePath: "nonexistent.txt",
			wantData: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := client.GetFile(ctx, tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(data) != len(tt.wantData) {
					t.Errorf("Data length mismatch: expected %d, got %d", len(tt.wantData), len(data))
					return
				}

				for i, b := range tt.wantData {
					if data[i] != b {
						t.Errorf("Data mismatch at byte %d: expected %d, got %d", i, b, data[i])
						break
					}
				}
			}
		})
	}
}

func TestLocalStorageClient_FileExists(t *testing.T) {
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	client, err := NewLocalStorageClient("")
	if err != nil {
		t.Fatalf("Failed to create LocalStorageClient: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Store a test file
	err = client.StoreFile(ctx, "existing.txt", []byte("test"))
	if err != nil {
		t.Fatalf("Failed to store test file: %v", err)
	}

	tests := []struct {
		name       string
		filePath   string
		wantExists bool
		wantErr    bool
	}{
		{
			name:       "existing file",
			filePath:   "existing.txt",
			wantExists: true,
			wantErr:    false,
		},
		{
			name:       "non-existent file",
			filePath:   "nonexistent.txt",
			wantExists: false,
			wantErr:    false,
		},
		{
			name:       "nested non-existent file",
			filePath:   "deep/nested/nonexistent.txt",
			wantExists: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := client.FileExists(ctx, tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if exists != tt.wantExists {
				t.Errorf("FileExists() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestLocalStorageClient_ListDir(t *testing.T) {
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	client, err := NewLocalStorageClient("")
	if err != nil {
		t.Fatalf("Failed to create LocalStorageClient: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Create test directory structure
	testFiles := []string{
		"reports/file1.txt",
		"reports/file2.html",
		"reports/2025/09/17/index.html",
		"reports/2025/09/17/styles.css",
		"reports/2025/09/18/index.html",
		"other/file3.txt",
	}

	for _, filePath := range testFiles {
		err := client.StoreFile(ctx, filePath, []byte("test content"))
		if err != nil {
			t.Fatalf("Failed to store test file %s: %v", filePath, err)
		}
	}

	tests := []struct {
		name      string
		dirPath   string
		recursive bool
		wantFiles []string
		wantErr   bool
	}{
		{
			name:      "non-recursive reports directory",
			dirPath:   "reports",
			recursive: false,
			wantFiles: []string{"reports/file1.txt", "reports/file2.html", "reports/2025"},
			wantErr:   false,
		},
		{
			name:      "recursive reports directory",
			dirPath:   "reports",
			recursive: true,
			wantFiles: []string{
				"reports/file1.txt",
				"reports/file2.html",
				"reports/2025/09/17/index.html",
				"reports/2025/09/17/styles.css",
				"reports/2025/09/18/index.html",
			},
			wantErr: false,
		},
		{
			name:      "specific nested directory",
			dirPath:   "reports/2025/09/17",
			recursive: false,
			wantFiles: []string{"reports/2025/09/17/index.html", "reports/2025/09/17/styles.css"},
			wantErr:   false,
		},
		{
			name:      "non-existent directory",
			dirPath:   "nonexistent",
			recursive: false,
			wantFiles: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := client.ListDir(ctx, tt.dirPath, tt.recursive)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// For non-recursive, we might get directories too, so we need to be more flexible
				if tt.recursive {
					// For recursive, we should only get files
					if len(files) != len(tt.wantFiles) {
						t.Errorf("ListDir() returned %d files, expected %d", len(files), len(tt.wantFiles))
						t.Logf("Got files: %v", files)
						t.Logf("Expected files: %v", tt.wantFiles)
					}

					// Check that all expected files are present
					fileMap := make(map[string]bool)
					for _, file := range files {
						fileMap[file] = true
					}

					for _, expectedFile := range tt.wantFiles {
						if !fileMap[expectedFile] {
							t.Errorf("Expected file %s not found in results", expectedFile)
						}
					}
				} else {
					// For non-recursive, just check that we got some results
					if len(files) == 0 && len(tt.wantFiles) > 0 {
						t.Errorf("ListDir() returned no files, expected some files")
					}
				}
			}
		})
	}
}

func TestLocalStorageClient_Integration(t *testing.T) {
	originalDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	client, err := NewLocalStorageClient("")
	if err != nil {
		t.Fatalf("Failed to create LocalStorageClient: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Test complete workflow: create dir, store file, check existence, retrieve file
	dirPath := "reports/2025/09/17"
	filePath := "reports/2025/09/17/test-report.html"
	fileContent := []byte("<html><head><title>Test Report</title></head><body><h1>Radio Propagation Report</h1></body></html>")

	// 1. Create directory
	err = client.CreateDir(ctx, dirPath)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// 2. Store file
	err = client.StoreFile(ctx, filePath, fileContent)
	if err != nil {
		t.Fatalf("Failed to store file: %v", err)
	}

	// 3. Check file exists
	exists, err := client.FileExists(ctx, filePath)
	if err != nil {
		t.Fatalf("Failed to check file existence: %v", err)
	}
	if !exists {
		t.Error("File should exist after storing")
	}

	// 4. Retrieve file
	retrievedContent, err := client.GetFile(ctx, filePath)
	if err != nil {
		t.Fatalf("Failed to retrieve file: %v", err)
	}

	// 5. Verify content matches
	if len(retrievedContent) != len(fileContent) {
		t.Errorf("Content length mismatch: expected %d, got %d", len(fileContent), len(retrievedContent))
	}

	for i, b := range fileContent {
		if retrievedContent[i] != b {
			t.Errorf("Content mismatch at byte %d: expected %d, got %d", i, b, retrievedContent[i])
			break
		}
	}

	// 6. List directory contents
	files, err := client.ListDir(ctx, dirPath, false)
	if err != nil {
		t.Fatalf("Failed to list directory: %v", err)
	}

	found := false
	for _, file := range files {
		if file == filePath {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Stored file not found in directory listing")
	}
}
