// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package consul

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	cs "github.com/hashicorp/hcp-sdk-go/clients/cloud-consul-service/preview/2021-02-04/models"
	"github.com/hashicorp/terraform-provider-hcs/internal/clients/hcs-ama-api-spec/models"
	"github.com/hashicorp/terraform-provider-hcs/internal/helper"
)

// hcpConsulAPIVersion is the version of the HCP Consul API we are using to retrieve versions.
const hcpConsulAPIVersion = "2021-02-04"

// platform_type is used by the request for versions to determine the subset of versions for HCS
var platform_type = string(cs.HashicorpCloudConsul20210204PlatformTypeHCS)

// Version represents a Consul version and the status of that version in regards to availability on HCP.
type Version struct {
	// Version is the Consul product version.
	Version string `json:"version"`

	// Status denotes the availability (and whether the version is recommended) of the associated Consul version.
	Status string `json:"status"`
}

// availableVersionsResponse is the body of the HCP Consul versions response.
type availableVersionsResponse struct {
	// Versions is a slice of available Consul versions and their statuses.
	Versions []Version `json:"versions"`
}

// GetAvailableHCPConsulVersions retrieves a slice of supported/available Consul versions from the HCP Consul API.
func GetAvailableHCPConsulVersions(ctx context.Context, hcpApiDomain string) ([]Version, error) {
	apiDomain := strings.TrimPrefix(hcpApiDomain, "https://")
	apiDomain = strings.TrimSuffix(hcpApiDomain, "/")

	url := fmt.Sprintf("https://%s/consul/%s/versions?platform_type=%s", apiDomain, hcpConsulAPIVersion, platform_type)

	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "terraform-provider-hcs")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve available Consul versions from HCP: %v", err)
	}
	var availableVersionsBody availableVersionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&availableVersionsBody); err != nil {
		return nil, fmt.Errorf("unable to deserialize versions JSON from HCP Consul service: %v", err)
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

// NormalizeVersion ensures the version starts with a 'v'
func NormalizeVersion(version string) string {
	return "v" + strings.TrimPrefix(version, "v")
}

// IsValidVersion determines that a given version string is contained within the slice of
// available Consul versions.
func IsValidVersion(version string, versions []Version) bool {
	for _, v := range versions {
		if version == v.Version {
			return true
		}
	}

	return false
}

// FromAMAVersions converts a slice of *HashicorpCloudConsulamaAmaVersion to a slice of
// Version.
func FromAMAVersions(amaVersions []*models.HashicorpCloudConsulamaAmaVersion) []Version {
	if amaVersions == nil {
		return nil
	}

	var versions []Version
	for _, v := range amaVersions {
		if v == nil {
			continue
		}

		versions = append(versions, Version{
			Version: v.Version,
			Status:  helper.AMAVersionStatusToString(v.Status),
		})
	}

	return versions
}
