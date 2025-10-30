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
	file, err := os.OpenFile(csvFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{
		r.Timestamp.Format(time.RFC3339),
		fmt.Sprintf("%.2f", r.PingMs),
		fmt.Sprintf("%.2f", r.DownloadMbps),
		fmt.Sprintf("%.2f", r.UploadMbps),
	}

	return writer.Write(record)
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
