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

// StoreReport stores an HTML report in GCS with the specified path structure
func (g *GCSClient) StoreReport(ctx context.Context, htmlContent string, timestamp time.Time) (string, error) {
	// Generate the object path: YYYY/MM/DD/PropagationReport-YYYY-MM-DD-HH-MM-SS.html
	objectPath := g.generateObjectPath(timestamp)
	
	log.Printf("Storing report to GCS: gs://%s/%s", g.bucket, objectPath)
	
	// Get bucket handle
	bucket := g.client.Bucket(g.bucket)
	
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
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.bucket, objectPath)
	
	log.Printf("Report successfully stored at: %s", publicURL)
	return objectPath, nil
}

// StoreFile stores any file (JSON, text, etc.) in GCS in the same folder as the report
func (g *GCSClient) StoreFile(ctx context.Context, fileData []byte, filename string, timestamp time.Time) error {
	// Generate the folder path for this report
	folderPath := g.generateFolderPath(timestamp)
	objectPath := folderPath + filename
	
	log.Printf("Storing file to GCS: gs://%s/%s", g.bucket, objectPath)
	
	// Get bucket handle
	bucket := g.client.Bucket(g.bucket)
	
	// Create object handle
	obj := bucket.Object(objectPath)
	
	// Create writer
	writer := obj.NewWriter(ctx)
	
	// Set content type based on file extension
	if strings.HasSuffix(filename, ".json") {
		writer.ContentType = "application/json"
	} else if strings.HasSuffix(filename, ".txt") {
		writer.ContentType = "text/plain"
	} else if strings.HasSuffix(filename, ".md") {
		writer.ContentType = "text/markdown"
	} else if strings.HasSuffix(filename, ".png") {
		writer.ContentType = "image/png"
	} else if strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".jpeg") {
		writer.ContentType = "image/jpeg"
	} else if strings.HasSuffix(filename, ".gif") {
		writer.ContentType = "image/gif"
	} else {
		writer.ContentType = "application/octet-stream"
	}
	
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

// StoreChartImage stores a chart image in GCS in the same folder as the report
func (g *GCSClient) StoreChartImage(ctx context.Context, imageData []byte, filename string, timestamp time.Time) (string, error) {
	// Generate the folder path for this report
	folderPath := g.generateFolderPath(timestamp)
	objectPath := folderPath + filename
	
	log.Printf("Storing chart image to GCS: gs://%s/%s", g.bucket, objectPath)
	
	// Get bucket handle
	bucket := g.client.Bucket(g.bucket)
	
	// Create object handle
	obj := bucket.Object(objectPath)
	
	// Create writer
	writer := obj.NewWriter(ctx)
	writer.ContentType = "image/png"
	writer.CacheControl = "public, max-age=86400" // Cache for 24 hours
	
	// Set metadata
	writer.Metadata = map[string]string{
		"generated-at":   timestamp.Format(time.RFC3339),
		"content-type":   "chart-image",
		"chart-filename": filename,
	}
	
	// Write image data
	if _, err := writer.Write(imageData); err != nil {
		writer.Close()
		return "", fmt.Errorf("failed to write image to GCS: %w", err)
	}
	
	// Close writer to finalize upload
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to finalize GCS image upload: %w", err)
	}
	
	// Return the public URL
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.bucket, objectPath)
	log.Printf("Chart image successfully stored at: %s", publicURL)
	return publicURL, nil
}

// GetChartImage retrieves a chart image from GCS
func (g *GCSClient) GetChartImage(ctx context.Context, imagePath string) ([]byte, error) {
	bucket := g.client.Bucket(g.bucket)
	obj := bucket.Object(imagePath)
	
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader for chart image %s: %w", imagePath, err)
	}
	defer reader.Close()
	
	imageData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read chart image %s: %w", imagePath, err)
	}
	
	return imageData, nil
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
func (c *GCSClient) GetLatestReport() (string, error) {
	// This is a simplified implementation - in practice you'd want to
	// list objects and find the most recent one
	return "", fmt.Errorf("GetLatestReport not implemented for GCS yet")
}

// GetReport retrieves a specific report content from GCS
func (g *GCSClient) GetReport(ctx context.Context, folderPath string) (string, error) {
	bucket := g.client.Bucket(g.bucket)
	
	// Ensure folderPath ends with / and append index.html
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}
	indexPath := folderPath + "index.html"
	
	obj := bucket.Object(indexPath)
	
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create reader for %s: %w", indexPath, err)
	}
	defer reader.Close()
	
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read content from %s: %w", indexPath, err)
	}
	
	return string(content), nil
}

// ListReports lists recent reports from GCS, sorted by creation time (newest first)
func (g *GCSClient) ListReports(ctx context.Context, limit int) ([]string, error) {
	bucket := g.client.Bucket(g.bucket)
	
	query := &storage.Query{
		Prefix: "",
	}
	
	it := bucket.Objects(ctx, query)
	
	// Collect all reports with their creation times
	type reportInfo struct {
		name    string
		created time.Time
	}
	
	var allReports []reportInfo
	
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}
		
		// Only include report folders (look for index.html files in PropagationReport folders)
		if strings.HasSuffix(attrs.Name, "/index.html") && strings.Contains(attrs.Name, "PropagationReport") {
			// Extract folder path by removing /index.html
			folderPath := strings.TrimSuffix(attrs.Name, "/index.html") + "/"
			allReports = append(allReports, reportInfo{
				name:    folderPath,
				created: attrs.Created,
			})
		}
	}
	
	// Sort by creation time (newest first)
	sort.Slice(allReports, func(i, j int) bool {
		return allReports[i].created.After(allReports[j].created)
	})
	
	// Return only the requested number of reports
	var reports []string
	for i, report := range allReports {
		if i >= limit {
			break
		}
		reports = append(reports, report.name)
	}
	
	return reports, nil
}

// DeleteOldReports deletes reports older than the specified duration
func (g *GCSClient) DeleteOldReports(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	bucket := g.client.Bucket(g.bucket)
	
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
		
		// Only consider report folders older than cutoff (look for index.html files)
		if strings.HasSuffix(attrs.Name, "/index.html") && 
		   strings.Contains(attrs.Name, "PropagationReport") && 
		   attrs.Created.Before(cutoff) {
			// Delete the entire folder by getting the folder path
			folderPath := strings.TrimSuffix(attrs.Name, "/index.html") + "/"
			toDelete = append(toDelete, folderPath)
		}
	}
	
	// Delete old report folders
	for _, folderPath := range toDelete {
		// List all objects in the folder
		folderQuery := &storage.Query{Prefix: folderPath}
		folderIt := bucket.Objects(ctx, folderQuery)
		
		var folderObjects []string
		for {
			attrs, err := folderIt.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Printf("Warning: Failed to list objects in folder %s: %v", folderPath, err)
				break
			}
			folderObjects = append(folderObjects, attrs.Name)
		}
		
		// Delete all objects in the folder
		for _, objName := range folderObjects {
			obj := bucket.Object(objName)
			if err := obj.Delete(ctx); err != nil {
				log.Printf("Warning: Failed to delete object %s: %v", objName, err)
			}
		}
		
		if len(folderObjects) > 0 {
			log.Printf("Deleted old report folder: %s (%d objects)", folderPath, len(folderObjects))
		}
	}
	
	log.Printf("Cleanup completed: %d old reports deleted", len(toDelete))
	return nil
}

// generateObjectPath creates the GCS object path for a report folder
func (g *GCSClient) generateObjectPath(timestamp time.Time) string {
	return fmt.Sprintf("%04d/%02d/%02d/PropagationReport-%s/index.html",
		timestamp.Year(),
		timestamp.Month(),
		timestamp.Day(),
		timestamp.Format("2006-01-02-15-04-05"),
	)
}

// generateFolderPath creates the GCS folder path for a report
func (g *GCSClient) generateFolderPath(timestamp time.Time) string {
	return fmt.Sprintf("%04d/%02d/%02d/PropagationReport-%s/",
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
