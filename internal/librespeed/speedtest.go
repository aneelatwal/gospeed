package librespeed

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type SpeedTester struct {
	Client          *http.Client
	NumStreams      int
	Duration        time.Duration
	TransferFunc    func(client *http.Client, url string) (int64, []byte, error)
	DataCaptureFunc func([]byte) // optional hook to store or handle data
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
