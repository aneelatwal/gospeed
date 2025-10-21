package librespeed

import (
	"fmt"
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

func RunDownloadTest(server ServerResult) (float64, error) {
	fullURL := BuildDownloadURL(server.Server)

	const numStreams = 8
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

				resp, err := client.Get(urlWithParam)
				if err != nil {
					errChan <- err
					return
				}

				n, err := io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
				if err != nil {
					errChan <- err
					return
				}

				mu.Lock()
				totalBytes += n
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	if err, ok := <-errChan; ok {
		return 0, err
	}

	duration := time.Since(startTime)
	speedMbps := float64(totalBytes*8) / (duration.Seconds() * 1_000_000)

	fmt.Printf("Downloaded %d bytes in %v, speed: %.2f Mbps\n", totalBytes, duration, speedMbps)

	return speedMbps, nil
}
