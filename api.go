package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

const muxAPI = "https://api.mux.com/video/v1"

// doMuxRequest performs an HTTP request against the Mux Video API using Basic Auth
func doMuxRequest(method, path string, body any, token, secret string) ([]byte, int, error) {
	var rb io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		rb = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, muxAPI+path, rb)
	if err != nil {
		return nil, 0, err
	}
	req.SetBasicAuth(token, secret)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return b, resp.StatusCode, nil
}
