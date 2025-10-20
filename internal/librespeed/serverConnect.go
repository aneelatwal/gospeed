package librespeed

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func PingServers(servers []Server) Server {
	const attempts = 3
	const timeout = 3 * time.Second
	const workerCount = 5

	type result struct {
		server     Server
		avgLatency time.Duration
		ok         bool
	}

	serverCh := make(chan Server)
	resultCh := make(chan result)

	// Start worker goroutines
	for w := 0; w < workerCount; w++ {
		go func() {
			client := &http.Client{
				Timeout: timeout,
			}
			for server := range serverCh {
				pingPath := "/backend/empty.php"
				if server.PingURL != "" {
					pingPath = server.PingURL
				}
				baseURL := strings.TrimRight(server.Server, "/")
				path := strings.TrimLeft(pingPath, "/")
				url := baseURL + "/" + path

				var totalDuration time.Duration
				var successCount int
				for i := 0; i < attempts; i++ {
					start := time.Now()
					resp, err := client.Get(url)
					if err != nil {
						continue
					}
					resp.Body.Close()
					duration := time.Since(start)
					totalDuration += duration
					successCount++
				}

				if successCount == 0 {
					resultCh <- result{server: server, ok: false}
					continue
				}
				avgDuration := totalDuration / time.Duration(successCount)
				resultCh <- result{server: server, avgLatency: avgDuration, ok: true}
			}
		}()
	}

	// Send servers to workers
	go func() {
		for _, server := range servers {
			serverCh <- server
		}
		close(serverCh)
	}()

	lowestAvg := time.Duration(1<<63 - 1) // max duration
	var fastestServer Server

	for i := 0; i < len(servers); i++ {
		res := <-resultCh
		if !res.ok {
			fmt.Printf("Server %s: all ping attempts failed or timed out\n", res.server.Server)
			continue
		}
		fmt.Printf("Server %s average latency: %v\n", res.server.Server, res.avgLatency)
		if res.avgLatency < lowestAvg {
			lowestAvg = res.avgLatency
			fastestServer = res.server
		}
	}

	return fastestServer
}
