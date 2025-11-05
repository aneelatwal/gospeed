package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Result represents one saved test run
type Result struct {
	Timestamp    time.Time
	PingMs       float64
	DownloadMbps float64
	UploadMbps   float64
}

const csvFile = "gospeed_results.csv"

// SaveResult appends a new test result to the CSV
func SaveResult(r Result) error {
	existing, _ := LoadLastResults(1000)
	existing = append(existing, r)

	if len(existing) > 5 {
		existing = existing[len(existing)-5:]
	}

	file, err := os.Create(csvFile)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, res := range existing {
		record := []string{
			res.Timestamp.Format(time.RFC3339),
			fmt.Sprintf("%.2f", res.PingMs),
			fmt.Sprintf("%.2f", res.DownloadMbps),
			fmt.Sprintf("%.2f", res.UploadMbps),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// LoadLastResults returns the N most recent results (up to 5)
func LoadLastResults(n int) ([]Result, error) {
	file, err := os.Open(csvFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Result{}, nil // No history yet
		}
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	results := []Result{}
	for _, row := range rows {
		if len(row) < 4 {
			continue
		}
		t, _ := time.Parse(time.RFC3339, row[0])
		ping, _ := strconv.ParseFloat(row[1], 64)
		down, _ := strconv.ParseFloat(row[2], 64)
		up, _ := strconv.ParseFloat(row[3], 64)
		results = append(results, Result{t, ping, down, up})
	}

	// Only return last n entries
	if len(results) > n {
		results = results[len(results)-n:]
	}
	return results, nil
}
