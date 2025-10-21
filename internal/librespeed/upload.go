package librespeed

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"
)

func BuildUploadURL(server Server) string {
	baseURL := strings.TrimRight(server.ServerURL, "/")
	uploadPath := strings.TrimLeft(server.UlURL, "/")
	return baseURL + "/" + uploadPath
}

func RunUploadTest(server ServerResult, data []byte) (float64, error) {
	fullURL := BuildUploadURL(server.Server)

	tester := SpeedTester{
		Client:     &http.Client{Timeout: 20 * time.Second},
		NumStreams: 16,
		Duration:   15 * time.Second,
		TransferFunc: func(client *http.Client, url string) (int64, []byte, error) {
			resp, err := client.Post(url, "application/octet-stream", bytes.NewReader(data))
			if err != nil {
				return 0, nil, err
			}
			defer resp.Body.Close()
			io.Copy(io.Discard, resp.Body)
			return int64(len(data)), nil, nil
		},
	}

	speed, _, err := tester.Run(fullURL)
	return speed, err
}
