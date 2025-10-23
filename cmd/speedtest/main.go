package speedtest

import (
	"fmt"
	"log"

	"github.com/aneelatwal/gospeed/internal/librespeed"
)

func PrintServerList(servers []librespeed.Server) {
	for _, server := range servers {
		fmt.Printf("Server ID: %d, Name: %s, URL: %s\n", server.ID, server.Name, server.ServerURL)
	}
}

func main() {
	// Get server list
	servers, err := librespeed.FetchServerList()
	if err != nil {
		log.Fatalf("Error fetching server list: %v", err)
	}

	// Find fastest server
	fastestServer := librespeed.PingServers(servers)
	fmt.Printf("Fastest server is %s with average latency %v\n", fastestServer.Server.Name, fastestServer.Latency)

	// Run download and upload tests
	downloadSpeed, data, err := librespeed.RunDownloadTest(fastestServer)
	if err != nil {
		log.Fatalf("Error during download test: %v", err)
	}
	fmt.Printf("Download speed: %.2f Mbps\n", downloadSpeed)

	uploadSpeed, err := librespeed.RunUploadTest(fastestServer, data)
	if err != nil {
		log.Fatalf("Error during upload test: %v", err)
	}
	fmt.Printf("Upload speed: %.2f Mbps\n", uploadSpeed)
}
