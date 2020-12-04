package consul

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Version struct {
	Version string `json:"version"`

	Status string `json:"status"`
}

type availableVersionsResponse struct {
	Versions []Version `json:"versions"`
}

func GetAvailableHCPConsulVersions(ctx context.Context, hcpApiDomain string) ([]Version, error) {
	apiDomain := strings.TrimPrefix(hcpApiDomain, "https://")
	apiDomain = strings.TrimSuffix(hcpApiDomain, "/")

	url := fmt.Sprintf("https://%s/consul/2020-08-26/versions", apiDomain)
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve avaialable Consul versions from HCP: %+v", err)
	}
	var availableVersionsBody availableVersionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&availableVersionsBody); err != nil {
		return nil, fmt.Errorf("unable to deserialize versions JSON from HCP Consul service: %+v", err)
	}

	return availableVersionsBody.Versions, nil
}

// RecommendedVersion returns the recommended version of Consul
func RecommendedVersion(versions []Version) string {
	var defaultVersion string

	for _, v := range versions {
		defaultVersion = v.Version

		if v.Status == "RECOMMENDED" {
			return defaultVersion
		}
	}

	return defaultVersion
}

func NormalizeVersion(version string) string {
	return "v" + strings.TrimPrefix(version, "v")
}

func IsValidVersion(version string, versions []Version) bool {
	for _, v := range versions {
		if version == v.Version {
			return true
		}
	}

	return false
}
