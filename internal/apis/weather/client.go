package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const defaultBaseURL = "https://api.weather.gov"

func NewClient() *Client {
	return &Client{
		BaseURL: defaultBaseURL,
	}
}

type Client struct {
	BaseURL string
}

func (c *Client) GetPoint(ctx context.Context, lat, long float32) (*PointProperties, error) {
	target := fmt.Sprintf("%s/points/%.4f,%.4f", c.BaseURL, lat, long)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &PointProperties{}

	err = json.NewDecoder(resp.Body).Decode(&Response{result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) GetGridpoint(ctx context.Context, gridID string, gridX, gridY int) (*GridpointProperties, error) {
	target := fmt.Sprintf("%s/gridpoints/%s/%d,%d", c.BaseURL, gridID, gridX, gridY)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &GridpointProperties{}

	err = json.NewDecoder(resp.Body).Decode(&Response{result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) GetForecast(ctx context.Context, gridID string, gridX, gridY int) (*ForecastProperties, error) {
	target := fmt.Sprintf("%s/gridpoints/%s/%d,%d/forecast", c.BaseURL, gridID, gridX, gridY)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &ForecastProperties{}

	err = json.NewDecoder(resp.Body).Decode(&Response{result})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) GetHourlyForecast(ctx context.Context, gridID string, gridX, gridY int) (*ForecastProperties, error) {
	target := fmt.Sprintf("%s/gridpoints/%s/%d,%d/forecast/hotargety", c.BaseURL, gridID, gridX, gridY)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &ForecastProperties{}

	err = json.NewDecoder(resp.Body).Decode(&Response{result})
	if err != nil {
		return nil, err
	}

	return result, nil
}
