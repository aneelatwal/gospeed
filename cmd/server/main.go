package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aneelatwal/gospeed/internal/librespeed"
	"github.com/aneelatwal/gospeed/internal/storage"
	"github.com/aneelatwal/gospeed/internal/web"
)

type ResultJson struct {
	Server       string  `json:"server"`
	PingMs       int64   `json:"ping_ms"`
	DownloadMbps float64 `json:"download_mbps"`
	UploadMbps   float64 `json:"upload_mbps"`
	Timestamp    string  `json:"timestamp"`
}

func main() {
	http.Handle("/", http.FileServer(http.FS(web.Files)))

	http.HandleFunc("/api/speedtest", func(w http.ResponseWriter, r *http.Request) {
		// Select the best server
		servers, err := librespeed.FetchServerList()
		if err != nil {
			http.Error(w, "Failed to fetch servers", http.StatusInternalServerError)
			return
		}
		best := librespeed.PingServers(servers)

		// Run download and upload tests
		downloadMbps, sampleData, err := librespeed.RunDownloadTest(best)
		if err != nil {
			http.Error(w, "Download test failed", http.StatusInternalServerError)
			return
		}
		uploadMbps, err := librespeed.RunUploadTest(best, sampleData)
		if err != nil {
			http.Error(w, "Upload test failed", http.StatusInternalServerError)
			return
		}

		timestamp := time.Now()

		result := ResultJson{
			Server:       best.Server.ServerURL,
			PingMs:       best.Latency.Milliseconds(),
			DownloadMbps: downloadMbps,
			UploadMbps:   uploadMbps,
			Timestamp:    timestamp.Format(time.RFC3339),
		}

		// Save to history CSV
		err = storage.SaveResult(storage.Result{
			Timestamp:    timestamp,
			PingMs:       float64(result.PingMs),
			DownloadMbps: result.DownloadMbps,
			UploadMbps:   result.UploadMbps,
		})
		if err != nil {
			// Log error, but don't fail API
			fmt.Printf("Failed to save result: %v\n", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	http.HandleFunc("/api/history", func(w http.ResponseWriter, r *http.Request) {
		results, err := storage.LoadLastResults(5)
		if err != nil {
			http.Error(w, "Failed to load history", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	fmt.Println("Server running on http://0.0.0.0:9090")
	http.ListenAndServe(":9090", nil)
}
