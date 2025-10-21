package librespeed

import (
	"net/http"
	"strings"
	"time"
)

func BuildPingURL(server Server) string {
	baseURL := strings.TrimRight(server.ServerURL, "/")
	pingPath := "/backend/empty.php"
	if server.PingURL != "" {
		pingPath = server.PingURL
	}
	path := strings.TrimLeft(pingPath, "/")
	return baseURL + "/" + path
}

func PingServers(servers []Server) ServerResult {
	const attempts = 3
	const timeout = 3 * time.Second
	const workerCount = 10

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
				url := BuildPingURL(server)

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
	var fastestServerResult ServerResult

	for i := 0; i < len(servers); i++ {
		res := <-resultCh
		if !res.ok {
			// fmt.Printf("Server %s: all ping attempts failed or timed out\n", res.server.ServerURL)
			continue
		}
		// fmt.Printf("Server %s average latency: %v\n", res.server.ServerURL, res.avgLatency)
		if res.avgLatency < lowestAvg {
			lowestAvg = res.avgLatency
			fastestServerResult = ServerResult{
				Server:  res.server,
				Latency: res.avgLatency,
			}
		}
	}

	return fastestServerResult
}
