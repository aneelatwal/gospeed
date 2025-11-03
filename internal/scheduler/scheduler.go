package scheduler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/aneelatwal/gospeed/internal/librespeed"
	"github.com/aneelatwal/gospeed/internal/storage"
)

var (
	currentFrequency int
	mutex            sync.Mutex
	stopChan         chan bool
)

// Start starts or restarts the background scheduler with the given frequency
func Start(frequencyHours int) {
	mutex.Lock()
	defer mutex.Unlock()

	// Stop existing scheduler if running
	if stopChan != nil {
		close(stopChan)
		stopChan = nil
	}

	currentFrequency = frequencyHours

	// If frequency is 0, don't start scheduler
	if frequencyHours == 0 {
		fmt.Println("Automatic speedtests disabled")
		return
	}

	// Create new stop channel
	stopChan = make(chan bool)

	// Start scheduler goroutine
	go func() {
		duration := time.Duration(frequencyHours) * time.Hour
		ticker := time.NewTicker(duration)
		defer ticker.Stop()

		fmt.Printf("Scheduler started: running speedtests every %d hours\n", frequencyHours)

		for {
			select {
			case <-ticker.C:
				if _, err := librespeed.RunSpeedtest(false); err != nil {
					fmt.Printf("Error running scheduled speedtest: %v\n", err)
				}
			case <-stopChan:
				fmt.Println("Scheduler stopped")
				return
			}
		}
	}()
}

// GetFrequency returns the current frequency setting
func GetFrequency() int {
	mutex.Lock()
	defer mutex.Unlock()
	return currentFrequency
}

// HandleFrequency handles GET and POST requests for frequency settings
func HandleFrequency(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		frequency := GetFrequency()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"frequency_hours": frequency})
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			FrequencyHours int `json:"frequency_hours"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate frequency (must be >= 0)
		if req.FrequencyHours < 0 {
			http.Error(w, "Frequency must be >= 0", http.StatusBadRequest)
			return
		}

		// Save config
		config := storage.Config{FrequencyHours: req.FrequencyHours}
		if err := storage.SaveConfig(config); err != nil {
			http.Error(w, "Failed to save config", http.StatusInternalServerError)
			return
		}

		// Restart scheduler with new frequency
		Start(req.FrequencyHours)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":         true,
			"frequency_hours": req.FrequencyHours,
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
