package librespeed

import (
	"time"
)

type ServerResult struct {
	Server        Server
	Latency       time.Duration
	DownloadSpeed float64
	UploadSpeed   float64
}
