package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "os/exec"

    "log/slog"

    "github.com/gin-gonic/gin"
)

const baseHost = "localtest.me:8080"

// Define custom error variables
var (
    ErrInvalidURL        = fmt.Errorf("invalid URL")
    ErrHTTPRequestFailed = fmt.Errorf("HTTP request failed")
    ErrReadBodyFailed    = fmt.Errorf("failed to read response body")
    ErrJSONUnmarshal     = fmt.Errorf("failed to unmarshal JSON")
    ErrFileDownload      = fmt.Errorf("failed to download file")
    ErrFileWrite         = fmt.Errorf("failed to write file")
    ErrImageResize       = fmt.Errorf("failed to resize image")
)

type FileInfo struct {
    Filename string `json:"filename"`
    Download string `json:"download"`
}

func downloadAndResize(tenantID, fileID, fileSize string) error {
    slog.Info("Processing request", "tenantID", tenantID, "fileID", fileID)

    urlStr := fmt.Sprintf("http://%s.%s/storage/%s.json", tenantID, baseHost, fileID)
    slog.Info("Resolved URL", "url", urlStr)

    // Parse the URL to extract the hostname
    parsedURL, err := url.Parse(urlStr)
    if err != nil {
        slog.Error("Invalid URL", "error", err)
        return fmt.Errorf("%w: %v", ErrInvalidURL, err)
    }
    slog.Info("Resolved Hostname", "hostname", parsedURL.Hostname())

    // Make HTTP request
    resp, err := http.Get(urlStr)
    if err != nil {
        slog.Error("HTTP request failed", "error", err)
        return fmt.Errorf("%w: %v", ErrHTTPRequestFailed, err)
    }
    defer resp.Body.Close()

    // Read response body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        slog.Error("Failed to read response body", "error", err)
        return fmt.Errorf("%w: %v", ErrReadBodyFailed, err)
    }

    // Decode JSON data
    var info FileInfo
    err = json.Unmarshal(body, &info)
    if err != nil {
        slog.Error("Failed to unmarshal JSON", "error", err)
        return fmt.Errorf("%w: %v", ErrJSONUnmarshal, err)
    }

    // Download file
    downloadResp, err := http.Get(info.Download)
    if err != nil {
        slog.Error("Failed to download file", "error", err)
        return fmt.Errorf("%w: %v", ErrFileDownload, err)
    }
    defer downloadResp.Body.Close()

    // Create target filename
    targetFilename := fmt.Sprintf("uploads/%s", info.Filename)

    // Read the downloaded file into memory
    fileBytes, err := io.ReadAll(downloadResp.Body)
    if err != nil {
        slog.Error("Failed to read downloaded file", "error", err)
        return fmt.Errorf("%w: %v", ErrReadBodyFailed, err)
    }

    // Save downloaded file
    err = os.WriteFile(targetFilename, fileBytes, 0600)
    if err != nil {
        slog.Error("Failed to write file", "error", err)
        return fmt.Errorf("%w: %v", ErrFileWrite, err)
    }

    convertCmd := fmt.Sprintf("convert %s -resize %sx%s %s", targetFilename, fileSize, fileSize, targetFilename)
    slog.Info("Running command", "command", convertCmd)
    _, err = exec.Command("sh", "-c", convertCmd).Output()
    if err != nil {
        slog.Error("Error resizing image", "error", err)
        return fmt.Errorf("%w: %v", ErrImageResize, err)
    }

    slog.Info("Downloaded and resized image", "filename", targetFilename)
    return nil
}

func main() {
    // Create a Gin router
    router := gin.Default()

    // Define a POST endpoint
    router.POST("/cloudpawnery/image", func(c *gin.Context) {
        tenantID := c.Query("tenantID")
        fileID := c.Query("fileID")
        fileSize := c.Query("fileSize")

        if fileSize == "" {
            fileSize = "200"
        }

        // Validate tenantID and fileID
        if tenantID == "" || fileID == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Missing tenantID or fileID"})
            return
        }

        // Call the download and resize function
        err := downloadAndResize(tenantID, fileID, fileSize)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // Return a success response
        c.JSON(http.StatusOK, gin.H{"message": "File downloaded and resized successfully"})
    })

    // Start the HTTP server
    router.Run(":7000")
}