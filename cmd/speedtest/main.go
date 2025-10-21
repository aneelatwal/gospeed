package main

import (
	"fmt"
	"log"

	"github.com/aneelatwal/gospeed/internal/librespeed"
)

func main() {
	servers, err := librespeed.FetchServerList()
	if err != nil {
		log.Fatalf("Error fetching server list: %v", err)
	}

	// for _, server := range servers {
	// 	fmt.Printf("Server ID: %d, Name: %s, URL: %s\n", server.ID, server.Name, server.ServerURL)
	// }

	fastestServer := librespeed.PingServers(servers)
	fmt.Printf("Fastest server is %s with average latency %v\n", fastestServer.Server.Name, fastestServer.Latency)

	downloadSpeed, err := librespeed.RunDownloadTest(fastestServer)
	if err != nil {
		log.Fatalf("Error during download test: %v", err)
	}

	fmt.Printf("Download speed: %.2f Mbps\n", downloadSpeed)
}
