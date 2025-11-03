package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aneelatwal/gospeed/internal/librespeed"
	"github.com/aneelatwal/gospeed/internal/scheduler"
	"github.com/aneelatwal/gospeed/internal/storage"
	"github.com/aneelatwal/gospeed/internal/web"
)

func main() {
	// Load config and start scheduler
	config, err := storage.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: failed to load config: %v\n", err)
		config = storage.Config{FrequencyHours: 0}
	}

	scheduler.Start(config.FrequencyHours)
	http.Handle("/", http.FileServer(http.FS(web.Files)))

	http.HandleFunc("/api/speedtest", func(w http.ResponseWriter, r *http.Request) {
		result, err := librespeed.RunSpeedtest(true)
		if err != nil {
			http.Error(w, "Failed to run speedtest", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	http.HandleFunc("/api/frequency", scheduler.HandleFrequency)

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
