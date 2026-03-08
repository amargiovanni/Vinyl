package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.PollingIntervalIdle != 5000 {
		t.Errorf("expected idle interval 5000, got %d", cfg.PollingIntervalIdle)
	}
	if cfg.PollingIntervalPlay != 2000 {
		t.Errorf("expected play interval 2000, got %d", cfg.PollingIntervalPlay)
	}
	if cfg.MinSessionDuration != 60 {
		t.Errorf("expected min session 60, got %d", cfg.MinSessionDuration)
	}
	if cfg.SessionGap != 30 {
		t.Errorf("expected session gap 30, got %d", cfg.SessionGap)
	}
	if !cfg.FirstRun {
		t.Error("expected first_run to be true")
	}
	if cfg.Spotify.RedirectURI != "http://localhost:27750/callback" {
		t.Errorf("expected redirect URI, got %s", cfg.Spotify.RedirectURI)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-config.json")

	cfg := Default()
	cfg.Location.City = "TestCity"
	cfg.WeatherAPIKey = "test-key-123"

	// Write manually since Save uses fixed path
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		t.Fatal(err)
	}

	// Read back
	readData, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	loaded := Default()
	if err := json.Unmarshal(readData, loaded); err != nil {
		t.Fatal(err)
	}

	if loaded.Location.City != "TestCity" {
		t.Errorf("expected city TestCity, got %s", loaded.Location.City)
	}
	if loaded.WeatherAPIKey != "test-key-123" {
		t.Errorf("expected API key, got %s", loaded.WeatherAPIKey)
	}
}
