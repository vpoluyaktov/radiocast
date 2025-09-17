package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// GCSClient handles Google Cloud Storage operations
// Uses bucketName as the root for all operations
type GCSClient struct {
	client *storage.Client
	bucket string
}

// NewGCSClient creates a new GCS client
func NewGCSClient(ctx context.Context, bucketName string) (*GCSClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}
	
	return &GCSClient{
		client: client,
		bucket: bucketName,
	}, nil
}

// Close closes the GCS client
func (g *GCSClient) Close() error {
	return g.client.Close()
}

// CreateDir creates a directory (no-op for GCS as directories are implicit)
func (g *GCSClient) CreateDir(ctx context.Context, dirPath string) error {
	// GCS doesn't have explicit directories, they're created implicitly when files are stored
	return nil
}

// StoreFile stores a file at the specified path
func (g *GCSClient) StoreFile(ctx context.Context, filePath string, fileData []byte) error {
	log.Printf("Storing file to GCS: gs://%s/%s", g.bucket, filePath)
	
	// Get bucket handle
	bucket := g.client.Bucket(g.bucket)
	
	// Create object handle
	obj := bucket.Object(filePath)
	
	// Create writer
	writer := obj.NewWriter(ctx)
	
	// Set content type based on file extension
	writer.ContentType = GetContentType(filePath)
	
	writer.CacheControl = "public, max-age=3600" // Cache for 1 hour
	
	// Write file data
	if _, err := writer.Write(fileData); err != nil {
		writer.Close()
		return fmt.Errorf("failed to write file to GCS: %w", err)
	}
	
	// Close writer to finalize upload
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to finalize GCS file upload: %w", err)
	}
	
	log.Printf("File successfully stored: %s", filePath)
	return nil
}

// GetFile retrieves a file from the specified path
func (g *GCSClient) GetFile(ctx context.Context, filePath string) ([]byte, error) {
	bucket := g.client.Bucket(g.bucket)
	obj := bucket.Object(filePath)
	
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader for file %s: %w", filePath, err)
	}
	defer reader.Close()
	
	fileData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	
	return fileData, nil
}

// ListDir lists contents of a directory
func (g *GCSClient) ListDir(ctx context.Context, dirPath string, recursive bool) ([]string, error) {
	bucket := g.client.Bucket(g.bucket)
	
	// Ensure dirPath ends with / for prefix matching
	prefix := dirPath
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	
	query := &storage.Query{
		Prefix: prefix,
	}
	
	// If not recursive, set delimiter to get only immediate children
	if !recursive {
		query.Delimiter = "/"
	}
	
	it := bucket.Objects(ctx, query)
	
	var files []string
	
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}
		
		// Keep the full path to maintain consistency with local storage
		if attrs.Name != prefix { // Skip the directory itself
			files = append(files, attrs.Name)
		}
	}
	
	return files, nil
}

// FileExists checks if a file exists at the specified path
func (g *GCSClient) FileExists(ctx context.Context, filePath string) (bool, error) {
	bucket := g.client.Bucket(g.bucket)
	obj := bucket.Object(filePath)
	
	_, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence %s: %w", filePath, err)
	}
	return true, nil
}
