package librespeed

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

func BuildUploadURL(server Server) string {
	baseURL := strings.TrimRight(server.ServerURL, "/")
	uploadPath := strings.TrimLeft(server.UlURL, "/")
	return baseURL + "/" + uploadPath
}

func RunUploadTest(server ServerResult, data []byte) (float64, error) {
	fullURL := BuildUploadURL(server.Server)

	const numStreams = 16
	const testDuration = 15 * time.Second

	var totalBytes int64
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, numStreams)

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	startTime := time.Now()

	wg.Add(numStreams)
	for i := 0; i < numStreams; i++ {
		go func(streamID int) {
			defer wg.Done()
			for {
				elapsed := time.Since(startTime)
				if elapsed >= testDuration {
					return
				}
				// Append unique cache buster query param
				urlWithParam := fmt.Sprintf("%s?ck=%d", fullURL, time.Now().UnixNano())

				resp, err := client.Post(urlWithParam, "application/octet-stream", bytes.NewReader(data))
				if err != nil {
					errChan <- err
					return
				}

				_, err = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if err != nil {
					errChan <- err
					return
				}

				mu.Lock()
				totalBytes += int64(len(data))
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return 0, <-errChan
	}

	elapsed := time.Since(startTime)
	uploadSpeedMbps := (float64(totalBytes) * 8) / elapsed.Seconds() / 1_000_000

	fmt.Printf("Uploaded %d bytes in %v, speed: %.2f Mbps\n", totalBytes, elapsed, uploadSpeedMbps)

	return uploadSpeedMbps, nil
}
