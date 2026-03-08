package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Location struct {
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
	City string  `json:"city"`
}

type SpotifyConfig struct {
	ClientID    string `json:"client_id"`
	RedirectURI string `json:"redirect_uri"`
}

type Config struct {
	Location             Location      `json:"location"`
	WeatherAPIKey        string        `json:"weather_api_key"`
	Spotify              SpotifyConfig `json:"spotify"`
	PollingIntervalIdle  int           `json:"polling_interval_idle_ms"`
	PollingIntervalPlay  int           `json:"polling_interval_playing_ms"`
	MinSessionDuration   int           `json:"min_session_duration_sec"`
	SessionGap           int           `json:"session_gap_sec"`
	FirstRun             bool          `json:"first_run"`
}

func Default() *Config {
	return &Config{
		Location: Location{
			Lat:  0,
			Lon:  0,
			City: "",
		},
		Spotify: SpotifyConfig{
			RedirectURI: "http://localhost:27750/callback",
		},
		PollingIntervalIdle: 5000,
		PollingIntervalPlay: 2000,
		MinSessionDuration:  60,
		SessionGap:          30,
		FirstRun:            true,
	}
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".vinyl"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := Default()
			if saveErr := Save(cfg); saveErr != nil {
				return cfg, nil
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
