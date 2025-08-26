package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

// GCSClient handles Google Cloud Storage operations
type GCSClient struct {
	client     *storage.Client
	bucketName string
}

// NewGCSClient creates a new GCS client
func NewGCSClient(ctx context.Context, bucketName string) (*GCSClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}
	
	return &GCSClient{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// Close closes the GCS client
func (g *GCSClient) Close() error {
	return g.client.Close()
}

// StoreReport stores an HTML report in GCS with the specified path structure
func (g *GCSClient) StoreReport(ctx context.Context, htmlContent string, timestamp time.Time) (string, error) {
	// Generate the object path: YYYY/MM/DD/PropagationReport-YYYY-MM-DD-HH-MM-SS.html
	objectPath := g.generateObjectPath(timestamp)
	
	log.Printf("Storing report to GCS: gs://%s/%s", g.bucketName, objectPath)
	
	// Get bucket handle
	bucket := g.client.Bucket(g.bucketName)
	
	// Create object handle
	obj := bucket.Object(objectPath)
	
	// Create writer
	writer := obj.NewWriter(ctx)
	writer.ContentType = "text/html"
	writer.CacheControl = "public, max-age=3600" // Cache for 1 hour
	
	// Set metadata
	writer.Metadata = map[string]string{
		"generated-at":    timestamp.Format(time.RFC3339),
		"content-type":    "radio-propagation-report",
		"report-version":  "1.0",
	}
	
	// Write content
	if _, err := io.WriteString(writer, htmlContent); err != nil {
		writer.Close()
		return "", fmt.Errorf("failed to write content to GCS: %w", err)
	}
	
	// Close writer to finalize upload
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to finalize GCS upload: %w", err)
	}
	
	// Generate public URL
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.bucketName, objectPath)
	
	log.Printf("Report successfully stored at: %s", publicURL)
	return publicURL, nil
}

// GetReport retrieves a specific report content from GCS
func (g *GCSClient) GetReport(ctx context.Context, objectPath string) (string, error) {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectPath)
	
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create reader for %s: %w", objectPath, err)
	}
	defer reader.Close()
	
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read content from %s: %w", objectPath, err)
	}
	
	return string(content), nil
}

// ListReports lists recent reports from GCS
func (g *GCSClient) ListReports(ctx context.Context, limit int) ([]string, error) {
	bucket := g.client.Bucket(g.bucketName)
	
	query := &storage.Query{
		Prefix: "",
		// Sort by name (which includes timestamp) in descending order
	}
	
	it := bucket.Objects(ctx, query)
	
	var reports []string
	count := 0
	
	for count < limit {
		attrs, err := it.Next()
		if err == storage.ErrObjectNotExist {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}
		
		// Only include HTML reports
		if strings.HasSuffix(attrs.Name, ".html") && strings.Contains(attrs.Name, "PropagationReport") {
			reports = append(reports, attrs.Name)
			count++
		}
	}
	
	return reports, nil
}

// DeleteOldReports deletes reports older than the specified duration
func (g *GCSClient) DeleteOldReports(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	bucket := g.client.Bucket(g.bucketName)
	
	query := &storage.Query{
		Prefix: "",
	}
	
	it := bucket.Objects(ctx, query)
	
	var toDelete []string
	
	for {
		attrs, err := it.Next()
		if err == storage.ErrObjectNotExist {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to list objects for cleanup: %w", err)
		}
		
		// Only consider HTML reports older than cutoff
		if strings.HasSuffix(attrs.Name, ".html") && 
		   strings.Contains(attrs.Name, "PropagationReport") && 
		   attrs.Created.Before(cutoff) {
			toDelete = append(toDelete, attrs.Name)
		}
	}
	
	// Delete old reports
	for _, objectName := range toDelete {
		obj := bucket.Object(objectName)
		if err := obj.Delete(ctx); err != nil {
			log.Printf("Warning: Failed to delete old report %s: %v", objectName, err)
		} else {
			log.Printf("Deleted old report: %s", objectName)
		}
	}
	
	log.Printf("Cleanup completed: %d old reports deleted", len(toDelete))
	return nil
}

// generateObjectPath creates the GCS object path for a report
func (g *GCSClient) generateObjectPath(timestamp time.Time) string {
	return fmt.Sprintf("%04d/%02d/%02d/PropagationReport-%s.html",
		timestamp.Year(),
		timestamp.Month(),
		timestamp.Day(),
		timestamp.Format("2006-01-02-15-04-05"),
	)
}

// ReportInfo contains information about a stored report
type ReportInfo struct {
	Name    string    `json:"name"`
	URL     string    `json:"url"`
	Size    int64     `json:"size"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}
