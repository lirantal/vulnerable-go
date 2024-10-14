package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
)

const baseHost = "localtest.me:8080"

type FileInfo struct {
	Filename string `json:"filename"`
	Download string `json:"download"`
}

func downloadAndResize(tenantID, fileID string) error {
	// Example input based on the mocked storage server in fixtures/http directory of the project:
	// tenantID := "3971533981712"
	// fileID := "fid-1f8b6b1e-1f8b-4b1e-8b6b-1e4b1e8b6b1e"

	url := fmt.Sprintf("http://%s.%s/storage/%s.json", tenantID, baseHost, fileID)
	fmt.Println("Resolved URL: ", url)

	// Make HTTP request
	resp, err := http.Get(url)
	if (err != nil) {
		panic(err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if (err != nil) {
		panic(err)
	}

	// Decode JSON data
	var info FileInfo
	err = json.Unmarshal(body, &info)
	if (err != nil) {
		panic(err)
	}

	// Download file
	downloadResp, err := http.Get(info.Download)
	if (err != nil) {
		panic(err)
	}
	defer downloadResp.Body.Close()

	// Create target filename
	targetFilename := fmt.Sprintf("uploads/%s", info.Filename)

	// read the downloaded file into memory
	fileBytes, err := ioutil.ReadAll(downloadResp.Body)
	if (err != nil) {
		panic(err)
	}

	// Save downloaded file
	err = ioutil.WriteFile(targetFilename, fileBytes, 0644)
	if (err != nil) {
		panic(err)
	}

	// Resize image using system command (replace 'convert' if needed)
	_, err = exec.Command("convert", targetFilename, "-resize", "180x180", targetFilename).Output()
	if (err != nil) {
		fmt.Println("Error resizing image:", err)
	} else {
		fmt.Println("Downloaded and resized image:", targetFilename)
	}

	return nil
}

func main() {
	// Create a Gin router
	router := gin.Default()

	// Define a POST endpoint
	router.POST("/cloudpawnery/image", func(c *gin.Context) {
		// If data lives on the query string we can use this:
		tenantID := c.Query("tenantID")
		fileID := c.Query("fileID")

		// Validate tenantID and fileID
		if (tenantID == "" || fileID == "") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing tenantID or fileID"})
			return
		}

		// Call the download and resize function
		err := downloadAndResize(tenantID, fileID)
		if (err != nil) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return a success response
		c.JSON(http.StatusOK, gin.H{"message": "File downloaded and resized successfully"})
	})

	// Start the HTTP server
	router.Run(":7000")
}