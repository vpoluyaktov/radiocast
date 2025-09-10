package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// GCSClient handles Google Cloud Storage operations
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


// StoreFile stores any file (JSON, text, etc.) in GCS in the same folder as the report
func (g *GCSClient) StoreFile(ctx context.Context, fileData []byte, filename string, timestamp time.Time) error {
	// Generate the object path for this file
	objectPath := GenerateReportFolderPath(timestamp) + "/" + filename
	
	log.Printf("Storing file to GCS: gs://%s/%s", g.bucket, objectPath)
	
	// Get bucket handle
	bucket := g.client.Bucket(g.bucket)
	
	// Create object handle
	obj := bucket.Object(objectPath)
	
	// Create writer
	writer := obj.NewWriter(ctx)
	
	// Set content type based on file extension
	writer.ContentType = GetContentType(filename)
	
	writer.CacheControl = "public, max-age=3600" // Cache for 1 hour
	
	// Set metadata
	writer.Metadata = map[string]string{
		"generated-at": timestamp.Format(time.RFC3339),
		"filename":     filename,
	}
	
	// Write file data
	if _, err := writer.Write(fileData); err != nil {
		writer.Close()
		return fmt.Errorf("failed to write file to GCS: %w", err)
	}
	
	// Close writer to finalize upload
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to finalize GCS file upload: %w", err)
	}
	
	log.Printf("File successfully stored: %s", filename)
	return nil
}



// GetFile retrieves any file from GCS
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

// GetLatestReport gets the most recent report from GCS
func (g *GCSClient) GetLatestReport() (string, error) {
	ctx := context.Background()
	reports, err := g.ListReports(ctx, 1)
	if err != nil {
		return "", err
	}
	
	if len(reports) == 0 {
		return "", fmt.Errorf("no reports found")
	}
	
	return reports[0], nil
}


// ListReports lists recent reports from GCS, sorted by creation time (newest first)
func (g *GCSClient) ListReports(ctx context.Context, limit int) ([]string, error) {
	bucket := g.client.Bucket(g.bucket)
	
	query := &storage.Query{
		Prefix: "reports/",
	}
	
	it := bucket.Objects(ctx, query)
	
	var reportPaths []string
	
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}
		
		// Look for index.html files
		if strings.HasSuffix(attrs.Name, "/index.html") {
			reportPaths = append(reportPaths, attrs.Name)
		}
	}
	
	// Sort alphabetically, then reverse for newest first
	sort.Strings(reportPaths)
	for i, j := 0, len(reportPaths)-1; i < j; i, j = i+1, j-1 {
		reportPaths[i], reportPaths[j] = reportPaths[j], reportPaths[i]
	}
	
	// Apply limit
	if limit > 0 && limit < len(reportPaths) {
		reportPaths = reportPaths[:limit]
	}
	
	return reportPaths, nil
}



// ReportInfo contains information about a stored report
type ReportInfo struct {
	Name    string    `json:"name"`
	URL     string    `json:"url"`
	Size    int64     `json:"size"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}
