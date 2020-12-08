package hcsmeta

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const MetaURLPrefix = "https://raw.githubusercontent.com/hashicorp/cloud-hcs-meta/master"

type PlanDefaults struct {
	Name string `json:"name"`

	Version string `json:"version"`

	ManagedAppApiVersion string `json:"ama_api_version"`
}

type SupportedRegion struct {
	ShortName string `json:"short"`

	FriendlyName string `json:"friendly"`
}

type supportedRegionsResponse struct {
	Regions []SupportedRegion
}

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
		return planDefaults, fmt.Errorf("unable to retrieve HCS plan defaults: %+v", err)
	}

	if err := json.NewDecoder(resp.Body).Decode(&planDefaults); err != nil {
		return planDefaults, fmt.Errorf("unable to deserialize HCS plan defaults: %+v", err)
	}

	return planDefaults, nil
}

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
		return nil, fmt.Errorf("unable to retrieve supported HCS regions: %+v", err)
	}

	var supportedRegionsBody supportedRegionsResponse

	if err := json.NewDecoder(resp.Body).Decode(&supportedRegionsBody); err != nil {
		return nil, fmt.Errorf("unable to deserialize supported HCS regions JSON: %+v", err)
	}

	return supportedRegionsBody.Regions, nil
}

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
