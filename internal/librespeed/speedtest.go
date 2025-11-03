package librespeed

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/aneelatwal/gospeed/internal/storage"
)

type SpeedTester struct {
	Client          *http.Client
	NumStreams      int
	Duration        time.Duration
	TransferFunc    func(client *http.Client, url string) (int64, []byte, error)
	DataCaptureFunc func([]byte) // optional hook to store or handle data
}

type ResultJson struct {
	Server       string  `json:"server"`
	PingMs       int64   `json:"ping_ms"`
	DownloadMbps float64 `json:"download_mbps"`
	UploadMbps   float64 `json:"upload_mbps"`
	Timestamp    string  `json:"timestamp"`
}

func (s *SpeedTester) Run(baseURL string) (float64, []byte, error) {
	var totalBytes int64
	var wg sync.WaitGroup
	var mu sync.Mutex
	startTime := time.Now()
	errChan := make(chan error, s.NumStreams)

	wg.Add(s.NumStreams)
	for i := 0; i < s.NumStreams; i++ {
		go func() {
			defer wg.Done()
			for time.Since(startTime) < s.Duration {
				url := fmt.Sprintf("%s?ck=%d", baseURL, time.Now().UnixNano())
				n, data, err := s.TransferFunc(s.Client, url)
				if err != nil {
					errChan <- err
					return
				}
				if data != nil && s.DataCaptureFunc != nil {
					s.DataCaptureFunc(data)
				}
				mu.Lock()
				totalBytes += n
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	close(errChan)
	if err, ok := <-errChan; ok {
		return 0, nil, err
	}

	elapsed := time.Since(startTime)
	speedMbps := float64(totalBytes*8) / (elapsed.Seconds() * 1_000_000)
	return speedMbps, nil, nil
}

// runSpeedtest performs a speedtest and saves the result
func RunSpeedtest(showResults bool) (ResultJson, error) {
	result := ResultJson{}
	// Select the best server
	servers, err := FetchServerList()
	if err != nil {
		return result, fmt.Errorf("failed to fetch servers: %w", err)
	}
	best := PingServers(servers)

	// Run download and upload tests
	downloadMbps, sampleData, err := RunDownloadTest(best)
	if err != nil {
		return result, fmt.Errorf("download test failed: %w", err)
	}
	uploadMbps, err := RunUploadTest(best, sampleData)
	if err != nil {
		return result, fmt.Errorf("upload test failed: %w", err)
	}

	timestamp := time.Now()

	if showResults {
		result = ResultJson{
			Server:       best.Server.ServerURL,
			PingMs:       best.Latency.Milliseconds(),
			DownloadMbps: downloadMbps,
			UploadMbps:   uploadMbps,
			Timestamp:    timestamp.Format(time.RFC3339),
		}
	}

	// Save to history CSV
	err = storage.SaveResult(storage.Result{
		Timestamp:    timestamp,
		PingMs:       float64(best.Latency.Milliseconds()),
		DownloadMbps: downloadMbps,
		UploadMbps:   uploadMbps,
	})
	if err != nil {
		return result, fmt.Errorf("failed to save result: %w", err)
	}

	fmt.Printf("Automatic speedtest completed: Download=%.2f Mbps, Upload=%.2f Mbps, Ping=%d ms\n",
		downloadMbps, uploadMbps, best.Latency.Milliseconds())
	return result, nil
}
