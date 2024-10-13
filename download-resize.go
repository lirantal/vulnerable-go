package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
)

const baseHost = "localhost:8080"

type FileInfo struct {
	Filename string `json:"filename"`
	Download string `json:"download"`
}

func main() {
	// Replace with your actual tenant id and file id
	tenantID := "3971533981712"
	fileID := "fid-1f8b6b1e-1f8b-4b1e-8b6b-1e4b1e8b6b1e"

	url := fmt.Sprintf("http://%s.%s/storage/%s.json", tenantID, baseHost, fileID)
	fmt.Println("Resolved URL: ", url)

	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Decode JSON data
	var info FileInfo
	err = json.Unmarshal(body, &info)
	if err != nil {
		panic(err)
	}

	// Download file
	downloadResp, err := http.Get(info.Download)
	if err != nil {
		panic(err)
	}
	defer downloadResp.Body.Close()

	// Create target filename
	targetFilename := fmt.Sprintf("uploads/%s", info.Filename)

	// read the downloaded file into memory
	fileBytes, err := ioutil.ReadAll(downloadResp.Body)
	if err != nil {
		panic(err)
	}

	// Save downloaded file
	err = ioutil.WriteFile(targetFilename, fileBytes, 0644)
	if err != nil {
		panic(err)
	}

	// Resize image using system command (replace 'convert' if needed)
	_, err = exec.Command("convert", targetFilename, "-resize", "180x180", targetFilename).Output()
	if err != nil {
		fmt.Println("Error resizing image:", err)
	} else {
		fmt.Println("Downloaded and resized image:", targetFilename)
	}
}