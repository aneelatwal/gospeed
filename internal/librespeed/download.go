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

func BuildDownloadURL(server Server) string {
	baseURL := strings.TrimRight(server.ServerURL, "/")
	downloadPath := strings.TrimLeft(server.DlURL, "/")
	return baseURL + "/" + downloadPath
}

func RunDownloadTest(server ServerResult) (float64, []byte, error) {
	fullURL := BuildDownloadURL(server.Server)

	const numStreams = 8
	const testDuration = 15 * time.Second
	const maxBufferSize = 5 * 1024 * 1024 // 5 MB

	var totalBytes int64
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, numStreams)

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	var downloadedData bytes.Buffer

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

				limitedReader := io.LimitReader(resp.Body, maxBufferSize)
				buf := make([]byte, 32*1024)
				for {
					n, err := limitedReader.Read(buf)
					if n > 0 {
						mu.Lock()
						totalBytes += int64(n)
						if downloadedData.Len() < maxBufferSize {
							remaining := maxBufferSize - downloadedData.Len()
							if n > remaining {
								downloadedData.Write(buf[:remaining])
							} else {
								downloadedData.Write(buf[:n])
							}
						}
						mu.Unlock()
					}
					if err != nil {
						if err == io.EOF {
							break
						}
						resp.Body.Close()
						errChan <- err
						return
					}
				}
				resp.Body.Close()
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	if err, ok := <-errChan; ok {
		return 0, nil, err
	}

	duration := time.Since(startTime)
	speedMbps := float64(totalBytes*8) / (duration.Seconds() * 1_000_000)

	fmt.Printf("Downloaded %d bytes in %v, speed: %.2f Mbps\n", totalBytes, duration, speedMbps)

	return speedMbps, downloadedData.Bytes(), nil
}
