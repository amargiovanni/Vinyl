package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// detectLocationViaIP queries a free IP geolocation service to approximate
// the user's location without requiring browser geolocation permissions.
func detectLocationViaIP() (map[string]interface{}, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get("http://ip-api.com/json/?fields=status,message,city,lat,lon,country")
	if err != nil {
		return nil, fmt.Errorf("ip geolocation request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ip geolocation returned %d", resp.StatusCode)
	}

	var result struct {
		Status  string  `json:"status"`
		Message string  `json:"message"`
		City    string  `json:"city"`
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
		Country string  `json:"country"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode geolocation response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("geolocation failed: %s", result.Message)
	}

	return map[string]interface{}{
		"lat":     result.Lat,
		"lon":     result.Lon,
		"city":    result.City,
		"country": result.Country,
	}, nil
}
