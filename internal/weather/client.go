package weather

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const (
	cacheTTL   = 30 * time.Minute
	apiBaseURL = "https://api.openweathermap.org/data/2.5/weather"
)

type Client struct {
	mu     sync.RWMutex
	apiKey string
	lat    float64
	lon    float64
	cache  *cachedWeather
	client *http.Client
}

type cachedWeather struct {
	Data      Data
	ExpiresAt time.Time
}

type owmResponse struct {
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
}

func NewClient(apiKey string, lat, lon float64) *Client {
	return &Client{
		apiKey: apiKey,
		lat:    lat,
		lon:    lon,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) UpdateLocation(apiKey string, lat, lon float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.apiKey = apiKey
	c.lat = lat
	c.lon = lon
	c.cache = nil
}

func (c *Client) GetCurrent() (*Data, error) {
	c.mu.RLock()
	if c.cache != nil && time.Now().Before(c.cache.ExpiresAt) {
		data := c.cache.Data
		c.mu.RUnlock()
		return &data, nil
	}
	c.mu.RUnlock()

	return c.fetchFromAPI()
}

func (c *Client) fetchFromAPI() (*Data, error) {
	c.mu.RLock()
	apiKey := c.apiKey
	lat := c.lat
	lon := c.lon
	c.mu.RUnlock()

	if apiKey == "" {
		return nil, fmt.Errorf("weather API key not configured")
	}
	if lat == 0 && lon == 0 {
		return nil, fmt.Errorf("location not configured")
	}

	url := fmt.Sprintf("%s?lat=%.4f&lon=%.4f&appid=%s&units=metric&lang=en",
		apiBaseURL, lat, lon, apiKey)

	resp, err := c.client.Get(url)
	if err != nil {
		return c.staleCache(), fmt.Errorf("weather API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.staleCache(), fmt.Errorf("weather API returned %d", resp.StatusCode)
	}

	var owm owmResponse
	if err := json.NewDecoder(resp.Body).Decode(&owm); err != nil {
		return c.staleCache(), fmt.Errorf("decode weather response: %w", err)
	}

	data := &Data{
		TempCelsius: owm.Main.Temp,
		Humidity:    owm.Main.Humidity,
		FetchedAt:   time.Now(),
	}
	if len(owm.Weather) > 0 {
		data.Condition = owm.Weather[0].Main
		data.Description = owm.Weather[0].Description
		data.IconCode = owm.Weather[0].Icon
	}

	c.mu.Lock()
	c.cache = &cachedWeather{Data: *data, ExpiresAt: time.Now().Add(cacheTTL)}
	c.mu.Unlock()

	slog.Info("weather fetched", "temp", data.TempCelsius, "condition", data.Condition)
	return data, nil
}

func (c *Client) staleCache() *Data {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cache != nil {
		data := c.cache.Data
		return &data
	}
	return nil
}
