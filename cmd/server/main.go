package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aneelatwal/gospeed/internal/librespeed"
)

type SpeedResult struct {
	Server       string  `json:"server"`
	DownloadMbps float64 `json:"download_mbps"`
	UploadMbps   float64 `json:"upload_mbps"`
	Timestamp    string  `json:"timestamp"`
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("internal/web")))

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

		result := SpeedResult{
			Server:       best.Server.ServerURL,
			DownloadMbps: downloadMbps,
			UploadMbps:   uploadMbps,
			Timestamp:    time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	fmt.Println("Server running on http://0.0.0.0:9090")
	http.ListenAndServe(":9090", nil)
}
