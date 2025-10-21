package librespeed

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

func BuildDownloadURL(server Server) string {
	baseURL := strings.TrimRight(server.ServerURL, "/")
	downloadPath := strings.TrimLeft(server.DlURL, "/")
	return baseURL + "/" + downloadPath
}

func RunDownloadTest(server ServerResult) (float64, []byte, error) {
	fullURL := BuildDownloadURL(server.Server)

	const maxBufferBytes = 5 * 1024 * 1024
	var downloadedData bytes.Buffer
	var mu sync.Mutex

	tester := SpeedTester{
		Client:     &http.Client{Timeout: 20 * time.Second},
		NumStreams: 8,
		Duration:   15 * time.Second,
		TransferFunc: func(client *http.Client, url string) (int64, []byte, error) {
			resp, err := client.Get(url)
			if err != nil {
				return 0, nil, err
			}
			defer resp.Body.Close()
			data, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxBufferBytes)))
			return int64(len(data)), data, err
		},
		DataCaptureFunc: func(data []byte) {
			mu.Lock()
			defer mu.Unlock()
			if downloadedData.Len() < maxBufferBytes {
				remaining := maxBufferBytes - downloadedData.Len()
				if len(data) > remaining {
					downloadedData.Write(data[:remaining])
				} else {
					downloadedData.Write(data)
				}
			}
		},
	}

	speed, _, err := tester.Run(fullURL)
	return speed, downloadedData.Bytes(), err
}
