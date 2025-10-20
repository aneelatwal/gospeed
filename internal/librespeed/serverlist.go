package librespeed

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type Server struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Server   string `json:"server"`
	DlURL    string `json:"dlURL"`
	GetIpURL string `json:"getIpURL"`
	PingURL  string `json:"pingURL"`
	UlURL    string `json:"ulURL"`
}

func FetchServerList() ([]Server, error) {
	resp, err := http.Get("https://librespeed.org/backend-servers/servers.php")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch server list: " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var servers []Server
	if err := json.Unmarshal(body, &servers); err != nil {
		return nil, err
	}

	return servers, nil
}
