package hcsmeta

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// MetaURLPrefix is the URL prefix for the HCS meta repository.
const MetaURLPrefix = "https://raw.githubusercontent.com/hashicorp/cloud-hcs-meta/master"

// PlanDefaults represents the default values of the current HCS Meta AMA plan.
type PlanDefaults struct {
	// Name is the name of the default HCS plan.
	Name string `json:"name"`

	// Version is the version of the default HCS plan.
	Version string `json:"version"`

	// ManagedAppApiVersion is the current version of the HCS AMA API.
	ManagedAppApiVersion string `json:"ama_api_version"`
}

// SupportedRegion represents a valid Azure region (location) for an HCS Cluster.
type SupportedRegion struct {
	// ShortName is the short-name the Azure region.
	ShortName string `json:"short"`

	// FriendlyName is the display name of the Azure region.
	FriendlyName string `json:"friendly"`
}

// supportedRegionsResponse is the body of the HCS Meta supported regions response.
type supportedRegionsResponse struct {
	// Regions is a slice of HCS supported Azure regions.
	Regions []SupportedRegion
}

// GetPlanDefaults gets the current HCS plan defaults from the HCS Meta repository.
func GetPlanDefaults(ctx context.Context) (PlanDefaults, error) {
	var planDefaults PlanDefaults

	url := MetaURLPrefix + "/ama-plans/defaults.json"
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return planDefaults, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return planDefaults, fmt.Errorf("unable to retrieve HCS plan defaults: %v", err)
	}

	if err := json.NewDecoder(resp.Body).Decode(&planDefaults); err != nil {
		return planDefaults, fmt.Errorf("unable to deserialize HCS plan defaults: %v", err)
	}

	return planDefaults, nil
}

// GetSupportedRegions gets the currently supported Azure regions from the HCS Meta repository.
func GetSupportedRegions(ctx context.Context) ([]SupportedRegion, error) {
	url := MetaURLPrefix + "/regions/regions.json"
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
		return nil, fmt.Errorf("unable to retrieve supported HCS regions: %v", err)
	}

	var supportedRegionsBody supportedRegionsResponse

	if err := json.NewDecoder(resp.Body).Decode(&supportedRegionsBody); err != nil {
		return nil, fmt.Errorf("unable to deserialize supported HCS regions JSON: %v", err)
	}

	return supportedRegionsBody.Regions, nil
}

// RegionIsSupported determines that a given region is supported by HCS.
func RegionIsSupported(region string, supportedRegions []SupportedRegion) bool {
	// Default to allowing the region if we have none to check against
	if supportedRegions == nil {
		return true
	}

	for _, s := range supportedRegions {
		if s.ShortName == region {
			return true
		}
	}

	return false
}
