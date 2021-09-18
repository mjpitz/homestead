package geocoding

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultBaseURL   = "https://geocoding.geo.census.gov/geocoder"
	defaultBenchmark = "4"
)

func NewClient() *Client {
	return &Client{
		BaseURL: defaultBaseURL,
	}
}

type Client struct {
	BaseURL string
}

func (c *Client) SearchByAddress(ctx context.Context, address *Address) (*SearchByAddressResponse, error) {
	target := fmt.Sprintf("%s/locations/address", c.BaseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = strings.Join([]string{
		"benchmark=" + defaultBenchmark,
		"format=json",
		// the following must be provided in the proper order
		"street=" + url.QueryEscape(address.Street),
		"city=" + url.QueryEscape(address.City),
		"state=" + url.QueryEscape(address.State),
		"zip=" + url.QueryEscape(address.Zip),
	}, "&")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &SearchByAddressResponse{}

	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
