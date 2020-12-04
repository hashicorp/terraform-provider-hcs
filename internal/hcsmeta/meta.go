package hcsmeta

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type PlanDefaults struct {
	Name string `json:"name"`

	Version string `json:"version"`

	ManagedAppApiVersion string `json:"ama_api_version"`
}

func GetPlanDefaults(ctx context.Context) (PlanDefaults, error) {
	var planDefaults PlanDefaults

	url := "https://raw.githubusercontent.com/hashicorp/cloud-hcs-meta/master/ama-plans/defaults.json"
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
		return planDefaults, fmt.Errorf("unable to retrieve avaialable Consul versions from HCP: %+v", err)
	}

	if err := json.NewDecoder(resp.Body).Decode(&planDefaults); err != nil {
		return planDefaults, fmt.Errorf("unable to deserialize versions JSON from HCP Consul service: %+v", err)
	}

	return planDefaults, nil
}
