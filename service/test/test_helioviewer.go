package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "time"
)

// Step 1: GET /?action=getClosestImage...
type ClosestImage struct {
    ID          string `json:"id"`
    Date        string `json:"date"`
    Observatory string `json:"observatory"`
    Instrument  string `json:"instrument"`
    Detector    string `json:"detector"`
    Measurement string `json:"measurement"`
}

func main() {
    // 1. Specify your desired observation time in UTC (or use now)
    t := time.Now().UTC()
    date := t.Format("2006-01-02T15:04:05Z")
    // Optionally, use a fixed date/time like: date := "2024-07-15T00:00:00Z"

    // 2. Get closest image ID for SDO AIA 304
    getIDurl := fmt.Sprintf(
        "https://api.helioviewer.org/v2/getClosestImage/?date=%s&sourceId=13",
        date,
    )
    fmt.Println("Requesting closest image:", getIDurl)

    resp, err := http.Get(getIDurl)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Response:", string(body))
    
    var img ClosestImage
    if err := json.Unmarshal(body, &img); err != nil {
        panic(fmt.Sprintf("Error parsing API response: %v", err))
    }

    fmt.Printf("Closest image ID: %s on %s\n", img.ID, img.Date)

    // 3. Download the full PNG image for this ID
    imgURL := fmt.Sprintf("https://api.helioviewer.org/v2/downloadImage/?id=%s&width=2048", img.ID)
    fmt.Println("Downloading full Sun PNG:", imgURL)

    out, err := os.Create("sun_sdo_aia304.png")
    if err != nil {
        panic(err)
    }
    defer out.Close()

    imgResp, err := http.Get(imgURL)
    if err != nil {
        panic(err)
    }
    defer imgResp.Body.Close()

    if imgResp.StatusCode != http.StatusOK {
        fmt.Printf("Failed to download image: %s\n", imgResp.Status)
        return
    }

    _, err = io.Copy(out, imgResp.Body)
    if err != nil {
        panic(err)
    }

    fmt.Println("Image saved as sun_sdo_aia304.png")
}
