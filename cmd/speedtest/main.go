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

	for _, server := range servers {
		fmt.Printf("Server ID: %d, Name: %s, URL: %s\n", server.ID, server.Name, server.Server)
	}

	fastestServer := librespeed.PingServers(servers)
	fmt.Printf("Fastest Server - ID: %d, Name: %s, URL: %s\n", fastestServer.ID, fastestServer.Name, fastestServer.Server)
}
