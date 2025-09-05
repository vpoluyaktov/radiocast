package imagery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GenerateSunGIF downloads last 24 hourly solar images from Helioviewer and assembles them into a GIF.
// The GIF duration is ~12s (2 fps across 24 frames) and loops 10 times. Each frame is annotated with the UTC hour label.
func GenerateSunGIF(ctx context.Context, reportDir string, ts time.Time, outputGIF string) error {
	// Create a single temporary directory for all processing
	tmpRootDir := filepath.Join(os.TempDir(), fmt.Sprintf("helio_gif_%d", time.Now().UnixNano()))
	framesDir := filepath.Join(tmpRootDir, "frames")
	annDir := filepath.Join(tmpRootDir, "annotated")
	finalDir := filepath.Join(tmpRootDir, "final")
	
	// Create all necessary directories
	for _, dir := range []string{framesDir, annDir, finalDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	// Ensure cleanup of all temporary files when done
	defer os.RemoveAll(tmpRootDir)

	client := &hvHTTP{timeout: 30 * time.Second}

	// Resolve a Helioviewer sourceId for SDO/AIA with preferred measurements
	sourceID, err := client.lookupSourceID(ctx, "SDO", "AIA", "AIA", "304")
	if err != nil {
		return fmt.Errorf("helio datasource lookup failed; %w", err)
	}
	log.Printf("SunGIF: Using SDO/AIA 304 with sourceID: %d", sourceID)

	// Generate time points for the last 24 hours
	base := ts.UTC().Truncate(time.Hour)
	var hours []time.Time
	for i := 23; i >= 0; i-- { 
		hours = append(hours, base.Add(-time.Duration(i)*time.Hour)) 
	}

	// Download and process images for each hour
	var frames []string
	for i, t := range hours {
		dateStr := t.Format("2006-01-02T15:04:05Z")
		id, err := client.getClosestImageIDBySourceID(ctx, dateStr, sourceID)
		if err != nil {
			log.Printf("SunGIF: getClosestImage failed for %s: %v", dateStr, err)
			continue
		}
		
		// Download the image
		data, ext, err := client.downloadThumbnail(ctx, id)
		if err != nil {
			log.Printf("SunGIF: downloadThumbnail failed for %s: %v", dateStr, err)
			continue
		}
		
		// Save the original image
		framePath := filepath.Join(framesDir, fmt.Sprintf("frame_%02d.%s", i, ext))
		if err := os.WriteFile(framePath, data, 0644); err != nil {
			log.Printf("SunGIF: failed to write frame: %v", err)
			continue
		}
		
		// Convert PNG to JPG if needed for consistency
		jpgFramePath := filepath.Join(framesDir, fmt.Sprintf("frame_%02d.jpg", i))
		if ext == "png" {
			convCmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", framePath, jpgFramePath)
			convCmd.Stdout = io.Discard
			convCmd.Stderr = io.Discard
			if err := convCmd.Run(); err != nil {
				log.Printf("SunGIF: PNG to JPG conversion failed: %v", err)
				continue
			}
			framePath = jpgFramePath
		}
		
		// Add timestamp annotation with the correct time for each frame
		timestamp := t.Format("Jan 02 15:04 UTC")
		annPath := filepath.Join(annDir, fmt.Sprintf("frame_%02d.jpg", i))
		
		// Use ffmpeg to add timestamp
		cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", framePath,
			"-vf", fmt.Sprintf("drawtext=text='%s':fontcolor=white:fontsize=24:box=1:boxcolor=black@0.5:boxborderw=5:x=(w-text_w)/2:y=h-th-10", timestamp),
			annPath)
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		if err := cmd.Run(); err != nil {
			log.Printf("SunGIF: annotation failed: %v", err)
			continue
		}
		
		frames = append(frames, annPath)
	}

	// Check if we have any frames
	if len(frames) == 0 {
		return fmt.Errorf("no frames were successfully processed")
	}
	log.Printf("SunGIF: Found %d frames for GIF assembly", len(frames))
	
	// Sort frames to ensure correct order
	sort.Strings(frames)
	
	// Copy frames to final directory with sequential numbering
	for i, frame := range frames {
		destPath := filepath.Join(finalDir, fmt.Sprintf("frame%02d.jpg", i))
		srcData, err := os.ReadFile(frame)
		if err != nil {
			return fmt.Errorf("failed to read frame %s: %w", frame, err)
		}
		if err := os.WriteFile(destPath, srcData, 0644); err != nil {
			return fmt.Errorf("failed to write final frame: %w", err)
		}
	}
	
	// Use the optimized ffmpeg command for GIF generation
	cmd := exec.CommandContext(ctx, "ffmpeg", "-y",
		"-framerate", "2", 
		"-i", filepath.Join(finalDir, "frame%02d.jpg"),
		"-vf", "scale=512:-1:flags=lanczos,split[s0][s1];[s0]palettegen=max_colors=64:stats_mode=diff[p];[s1][p]paletteuse=dither=bayer:bayer_scale=5:diff_mode=rectangle",
		"-loop", "10",
		outputGIF,
	)
	
	// Capture stderr for logging
	var stderr bytes.Buffer
	cmd.Stdout = io.Discard
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		log.Printf("SunGIF: ffmpeg stderr: %s", stderr.String())
		return fmt.Errorf("assemble gif: %w", err)
	}
	
	// GIF generation completed successfully
	
	// Verify the GIF was created
	if _, err := os.Stat(outputGIF); err != nil {
		return fmt.Errorf("GIF not created: %w", err)
	}
	
	log.Printf("SunGIF: Successfully created %s", outputGIF)
	return nil
}

// hvHTTP is a lightweight HTTP helper for Helioviewer calls
type hvHTTP struct { timeout time.Duration }

func (c *hvHTTP) get(ctx context.Context, url string) ([]byte, error) {
	log.Printf("Helioviewer API request: GET %s", url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	hc := &http.Client{Timeout: c.timeout}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != 200 {
		log.Printf("Helioviewer API Error: Status=%d, URL=%s, Body=%s", resp.StatusCode, url, string(body))
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	return body, nil
}

func (c *hvHTTP) getClosestImageIDBySourceID(ctx context.Context, date string, sourceID int64) (int64, error) {
	url := fmt.Sprintf("https://api.helioviewer.org/v2/getClosestImage/?date=%s&sourceId=%d", date, sourceID)
	b, err := c.get(ctx, url)
	if err != nil { return 0, err }
	
	// The API returns the ID as a string, so we need to parse it manually
	var resp map[string]interface{}
	if err := json.Unmarshal(b, &resp); err != nil { return 0, err }
	
	// Get the ID field as a string or number
	idValue, ok := resp["id"]
	if !ok { return 0, fmt.Errorf("no id field in response") }
	
	// Convert to int64 based on the type
	var imageID int64
	switch v := idValue.(type) {
	case string:
		// Parse string to int64
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil { return 0, fmt.Errorf("invalid id format: %s", v) }
		imageID = parsed
	case float64:
		// JSON numbers are parsed as float64
		imageID = int64(v)
	default:
		return 0, fmt.Errorf("unexpected id type: %T", idValue)
	}
	
	if imageID == 0 { return 0, fmt.Errorf("no image id") }
	return imageID, nil
}

func (c *hvHTTP) downloadThumbnail(ctx context.Context, id int64) ([]byte, string, error) {
	// Use the downloadImage endpoint to get a pre-colorized PNG image
	url := fmt.Sprintf("https://api.helioviewer.org/v2/downloadImage/?id=%d&width=1024", id)
	log.Printf("SunGIF: Downloading colorized image from %s", url)
	b, err := c.get(ctx, url)
	if err != nil {
		log.Printf("SunGIF: downloadImage failed: %v, falling back to JP2", err)
		// Fall back to JP2 if downloadImage endpoint fails
		url = fmt.Sprintf("https://api.helioviewer.org/v2/getJP2Image/?id=%d", id)
		b, err = c.get(ctx, url)
		if err != nil { return nil, "", err }
		
		// Convert JP2 to JPEG using ffmpeg with color mapping for SDO/AIA 304
		tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("helio_%d.jp2", id))
		tmpJpg := filepath.Join(os.TempDir(), fmt.Sprintf("helio_%d.jpg", id))
		if err := os.WriteFile(tmpFile, b, 0644); err != nil { return nil, "", err }
		
		// Apply a color palette to the image - SDO/AIA 304 is typically rendered in reddish-orange
		// Using a colormap filter to convert grayscale to color
		cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", tmpFile, 
			"-vf", "eq=gamma=1.5:saturation=2,colorchannelmixer=.5:.8:.1:0:.2:.5:.1:0:.1:.2:.5:0", 
			tmpJpg)
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		if err := cmd.Run(); err != nil { return nil, "", fmt.Errorf("failed to convert JP2 to JPG: %w", err) }
		
		// Read the converted JPG
		jpgData, err := os.ReadFile(tmpJpg)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read converted JPG: %w", err)
		}
		
		// Clean up temp files
		os.Remove(tmpFile)
		os.Remove(tmpJpg)
		
		return jpgData, "jpg", nil
	}
	
	// Check content type of downloaded image
	contentType := http.DetectContentType(b)
	log.Printf("SunGIF: Downloaded image content type: %s", contentType)
	
	if strings.Contains(contentType, "image/png") {
		return b, "png", nil
	} else if strings.Contains(contentType, "image/jpeg") {
		return b, "jpg", nil
	}
	
	// If we got here, we have an unknown content type
	return nil, "", fmt.Errorf("unknown content type: %s", contentType)
}

// lookupSourceID queries getDataSources and finds a sourceId matching the given parameters.
func (c *hvHTTP) lookupSourceID(ctx context.Context, observatory, instrument, detector, measurement string) (int64, error) {
	// Endpoint returns a JSON of available sources
	url := "https://api.helioviewer.org/v2/getDataSources/"
	b, err := c.get(ctx, url)
	if err != nil { return 0, err }
	
	// The actual structure is a flat map: { "SDO": { "AIA": { "304": { "sourceId": N } } } }
	// Parse as generic map to handle the variable structure
	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil { 
		return 0, fmt.Errorf("failed to parse data sources: %w", err) 
	}
	
	// Navigate through the nested maps to find the sourceId
	obsData, ok := data[observatory]
	if !ok { return 0, fmt.Errorf("observatory not found: %s", observatory) }
	
	obsMap, ok := obsData.(map[string]interface{})
	if !ok { return 0, fmt.Errorf("invalid observatory data format") }
	
	instData, ok := obsMap[instrument]
	if !ok { return 0, fmt.Errorf("instrument not found: %s", instrument) }
	
	// For SDO/AIA, the structure is SDO->AIA->304->sourceId (no detector level)
	var sourceID int64
	if detector == "AIA" && instrument == "AIA" {
		// Special case for SDO/AIA where detector is skipped
		instMap, ok := instData.(map[string]interface{})
		if !ok { return 0, fmt.Errorf("invalid instrument data format") }
		
		measData, ok := instMap[measurement]
		if !ok { return 0, fmt.Errorf("measurement not found: %s", measurement) }
		
		measMap, ok := measData.(map[string]interface{})
		if !ok { return 0, fmt.Errorf("invalid measurement data format") }
		
		sourceIDFloat, ok := measMap["sourceId"].(float64)
		if !ok { return 0, fmt.Errorf("sourceId not found or invalid") }
		
		sourceID = int64(sourceIDFloat)
	} else {
		// General case with detector level
		instMap, ok := instData.(map[string]interface{})
		if !ok { return 0, fmt.Errorf("invalid instrument data format") }
		
		detData, ok := instMap[detector]
		if !ok { return 0, fmt.Errorf("detector not found: %s", detector) }
		
		detMap, ok := detData.(map[string]interface{})
		if !ok { return 0, fmt.Errorf("invalid detector data format") }
		
		measData, ok := detMap[measurement]
		if !ok { return 0, fmt.Errorf("measurement not found: %s", measurement) }
		
		measMap, ok := measData.(map[string]interface{})
		if !ok { return 0, fmt.Errorf("invalid measurement data format") }
		
		sourceIDFloat, ok := measMap["sourceId"].(float64)
		if !ok { return 0, fmt.Errorf("sourceId not found or invalid") }
		
		sourceID = int64(sourceIDFloat)
	}
	
	if sourceID == 0 { 
		return 0, fmt.Errorf("no sourceId for %s/%s/%s/%s", observatory, instrument, detector, measurement) 
	}
	
	return sourceID, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil { return err }
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil { return err }
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil { return err }
	return out.Sync()
}
