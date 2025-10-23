package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// uploadFile uploads a local file to the provided signed URL via PUT
func uploadFile(uploadURL, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", uploadURL, f)
	if err != nil {
		return err
	}
	// content-type inference
	ct := "application/octet-stream"
	switch filepath.Ext(filePath) {
	case ".mp4", ".mov", ".m4v":
		ct = "video/mp4"
	case ".m4a":
		ct = "audio/mp4"
	}
	req.Header.Set("Content-Type", ct)
	req.ContentLength = fi.Size()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %d %s", resp.StatusCode, string(b))
	}
	return nil
}
