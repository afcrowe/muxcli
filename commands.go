package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func cmdCreate(token, secret, inputFile string) error {
	body := map[string]any{"new_asset_settings": map[string]any{"playback_policies": []string{"public"}}}
	b, code, err := doMuxRequest("POST", "/uploads", body, token, secret)
	if err != nil {
		return err
	}
	if code < 200 || code >= 300 {
		return fmt.Errorf("create upload failed: %d %s", code, string(b))
	}
	var resp map[string]any
	if err := json.Unmarshal(b, &resp); err != nil {
		return err
	}
	data := resp["data"].(map[string]any)
	uploadURL := data["url"].(string)
	uploadID := data["id"].(string)

	fmt.Printf("Upload URL: %s\nUpload ID: %s\n", uploadURL, uploadID)

	if err := uploadFile(uploadURL, inputFile); err != nil {
		return fmt.Errorf("uploadFile: %w", err)
	}

	var assetID string
	for i := 0; i < 60; i++ {
		b, code, err := doMuxRequest("GET", "/uploads/"+uploadID, nil, token, secret)
		if err != nil {
			log.Printf("Get upload error: %v", err)
		} else if code >= 200 && code < 300 {
			var up map[string]any
			if err := json.Unmarshal(b, &up); err == nil {
				if d, ok := up["data"].(map[string]any); ok {
					if aid, ok := d["asset_id"].(string); ok && aid != "" {
						assetID = aid
						break
					}
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
	if assetID == "" {
		return fmt.Errorf("timed out waiting for asset to be created from upload %s", uploadID)
	}

	for i := 0; i < 120; i++ {
		b, code, err := doMuxRequest("GET", "/assets/"+assetID, nil, token, secret)
		if err != nil {
			log.Printf("Get asset error: %v", err)
		} else if code >= 200 && code < 300 {
			var a map[string]any
			if err := json.Unmarshal(b, &a); err == nil {
				if d, ok := a["data"].(map[string]any); ok {
					if status, _ := d["status"].(string); status == "ready" {
						fmt.Printf("Asset ready: %s\n", assetID)
						if pids, ok := d["playback_ids"].([]any); ok {
							for _, p := range pids {
								if pm, ok := p.(map[string]any); ok {
									fmt.Printf("Playback ID: %v (policy=%v)\n", pm["id"], pm["policy"])
								}
							}
						}
						return nil
					}
					log.Printf("asset %s status: %v", assetID, d["status"])
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timed out waiting for asset %s to become ready", assetID)
}

func cmdDelete(token, secret, assetID string) error {
	_, code, err := doMuxRequest("DELETE", "/assets/"+assetID, nil, token, secret)
	if err != nil {
		return err
	}
	if code != 204 {
		return fmt.Errorf("delete asset returned status %d", code)
	}
	fmt.Printf("Deleted asset %s\n", assetID)
	return nil
}

func cmdGet(token, secret, assetID string) error {
	b, code, err := doMuxRequest("GET", "/assets/"+assetID, nil, token, secret)
	if err != nil {
		return err
	}
	if code < 200 || code >= 300 {
		return fmt.Errorf("get asset failed: %d %s", code, string(b))
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, b, "", "  "); err != nil {
		fmt.Println(string(b))
		return nil
	}
	fmt.Println(pretty.String())
	return nil
}

func cmdCreateRendition(token, secret, assetID, resolution string) error {
	body := map[string]any{"resolution": resolution}
	b, code, err := doMuxRequest("POST", "/assets/"+assetID+"/static-renditions", body, token, secret)
	if err != nil {
		return err
	}
	if code < 200 || code >= 300 {
		return fmt.Errorf("create static rendition failed: %d %s", code, string(b))
	}
	fmt.Println(string(b))
	return nil
}

func cmdListRenditions(token, secret, assetID string) error {
	b, code, err := doMuxRequest("GET", "/assets/"+assetID, nil, token, secret)
	if err != nil {
		return err
	}
	if code < 200 || code >= 300 {
		return fmt.Errorf("get asset failed: %d %s", code, string(b))
	}
	var a map[string]any
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	if d, ok := a["data"].(map[string]any); ok {
		if sr, ok := d["static_renditions"].(map[string]any); ok {
			if files, ok := sr["files"].([]any); ok {
				for _, f := range files {
					if fm, ok := f.(map[string]any); ok {
						fmt.Printf("id=%v name=%v status=%v size=%v bitrate=%v\n", fm["id"], fm["name"], fm["status"], fm["filesize"], fm["bitrate"])
					}
				}
				return nil
			}
			fmt.Printf("static_renditions status=%v\n", sr["status"])
			return nil
		}
		fmt.Println("no static_renditions for this asset")
		return nil
	}
	return fmt.Errorf("unexpected response")
}

// cmdGetMasterDownload requests temporary master access, and optionally polls the asset
// until a master.url appears. If wait is true, it will poll the GET asset endpoint
// every 5 seconds until timeoutSec seconds elapse.
func cmdGetMasterDownload(token, secret, assetID string, wait bool, timeoutSec int) error {
	body := map[string]any{"master_access": "temporary"}
	b, code, err := doMuxRequest("PUT", "/assets/"+assetID+"/master-access", body, token, secret)
	if err != nil {
		return err
	}
	if code < 200 || code >= 300 {
		return fmt.Errorf("master-access request failed: %d %s", code, string(b))
	}
	var a map[string]any
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	if d, ok := a["data"].(map[string]any); ok {
		if master, ok := d["master"].(map[string]any); ok {
			fmt.Printf("master.status=%v\n", master["status"])
			if url, ok := master["url"].(string); ok && url != "" {
				fmt.Printf("master.url=%s\n", url)
				return nil
			}
		}
	}

	if !wait {
		// No polling requested; print whatever response we got and return
		fmt.Println(string(b))
		return nil
	}

	// Poll GET /assets/{assetID} every 5s until master.url appears or timeout
	deadline := time.Now().Add(time.Duration(timeoutSec) * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(5 * time.Second)
		b2, code2, err := doMuxRequest("GET", "/assets/"+assetID, nil, token, secret)
		if err != nil {
			log.Printf("Get asset error: %v", err)
			continue
		}
		if code2 < 200 || code2 >= 300 {
			log.Printf("unexpected status from GET asset: %d %s", code2, string(b2))
			continue
		}
		var a2 map[string]any
		if err := json.Unmarshal(b2, &a2); err != nil {
			log.Printf("failed to parse GET asset response: %v", err)
			continue
		}
		if d2, ok := a2["data"].(map[string]any); ok {
			if master2, ok := d2["master"].(map[string]any); ok {
				if url, ok := master2["url"].(string); ok && url != "" {
					fmt.Printf("master.url=%s\n", url)
					return nil
				}
			}
		}
	}

	return fmt.Errorf("timed out waiting for master.url for asset %s", assetID)
}

// DeleteAssetsFromFile deletes asset IDs listed (one per line) in fileName using HTTP DELETE to Mux API
func DeleteAssetsFromFile(fileName, tokenID, secretKey string) error {
	if fileName == "" || tokenID == "" || secretKey == "" {
		return fmt.Errorf("fileName, tokenID and secretKey are required")
	}

	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	client := &http.Client{}
	for scanner.Scan() {
		assetID := scanner.Text()
		if assetID == "" {
			continue
		}
		fmt.Printf("Deleting asset: %s... ", assetID)
		req, _ := http.NewRequest("DELETE", muxAPI+"/assets/"+assetID, nil)
		req.SetBasicAuth(tokenID, secretKey)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode != 204 {
			fmt.Printf("FAILED: status %d\n", resp.StatusCode)
			continue
		}
		fmt.Println("OK")
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	log.Println("Done deleting assets from file")
	return nil
}
