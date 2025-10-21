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

	const numGoroutines = 4
	var totalBytes int64
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, numGoroutines)

	startTime := time.Now()

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			resp, err := http.Get(fullURL)
			if err != nil {
				errChan <- err
				return
			}
			defer resp.Body.Close()

			n, err := io.Copy(io.Discard, resp.Body)
			if err != nil {
				errChan <- err
				return
			}

			mu.Lock()
			totalBytes += n
			mu.Unlock()
		}()
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
