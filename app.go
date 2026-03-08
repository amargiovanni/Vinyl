package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"vinyl/internal/config"
	"vinyl/internal/diary"
	"vinyl/internal/player"
	"vinyl/internal/weather"
)

// App is the main application struct exposed to the frontend via Wails bindings.
type App struct {
	ctx      context.Context
	cancel   context.CancelFunc
	cfg      *config.Config
	store    *diary.Store
	detector *player.Detector
	weather  *weather.Client
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	a.ctx = ctx
	a.cancel = cancel

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		cfg = config.Default()
	}
	a.cfg = cfg

	store, err := diary.NewStore()
	if err != nil {
		slog.Error("failed to open database", "err", err)
		return
	}
	a.store = store

	if err := store.Migrate(); err != nil {
		slog.Error("failed to run migrations", "err", err)
	}

	a.weather = weather.NewClient(cfg.WeatherAPIKey, cfg.Location.Lat, cfg.Location.Lon)

	a.detector = player.NewDetector(a.store, a.weather, cfg, func(event string, data interface{}) {
		runtime.EventsEmit(a.ctx, event, data)
	})

	go a.detector.Start(ctx)

	slog.Info("Vinyl started")
}

func (a *App) shutdown(ctx context.Context) {
	if a.cancel != nil {
		a.cancel()
	}
	if a.detector != nil {
		a.detector.Stop()
	}
	if a.store != nil {
		a.store.Close()
	}
	slog.Info("Vinyl shutdown")
}

// --- Wails bindings (exposed to frontend) ---

// GetCurrentTrack returns the currently playing track info.
func (a *App) GetCurrentTrack() *player.TrackInfo {
	if a.detector == nil {
		return nil
	}
	return a.detector.GetCurrentTrack()
}

// GetCurrentSession returns the active listening session.
func (a *App) GetCurrentSession() *diary.Session {
	if a.detector == nil {
		return nil
	}
	return a.detector.GetCurrentSession()
}

// SetSessionMood sets the mood for a session.
func (a *App) SetSessionMood(sessionID string, mood string) error {
	if a.store == nil {
		return fmt.Errorf("store not initialized")
	}
	return a.store.UpdateSessionMood(sessionID, mood)
}

// GetConfig returns the current configuration.
func (a *App) GetConfig() *config.Config {
	return a.cfg
}

// SaveConfig persists the configuration.
func (a *App) SaveConfig(cfg config.Config) error {
	a.cfg = &cfg
	a.weather.UpdateLocation(cfg.WeatherAPIKey, cfg.Location.Lat, cfg.Location.Lon)
	return config.Save(&cfg)
}

// IsFirstRun checks if this is the first time the app is launched.
func (a *App) IsFirstRun() bool {
	return a.cfg.FirstRun
}

// CompleteOnboarding marks the first-run onboarding as complete.
func (a *App) CompleteOnboarding() error {
	a.cfg.FirstRun = false
	return config.Save(a.cfg)
}

// GetHeatmapData returns listening data for the last 52 weeks.
func (a *App) GetHeatmapData() ([]diary.DailyStats, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return a.store.GetHeatmapData(52)
}

// GetSessionsByMonth returns sessions for a given year/month.
func (a *App) GetSessionsByMonth(year, month int) ([]diary.Session, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return a.store.GetSessionsByMonth(year, month)
}

// GetSessionTracks returns all tracks for a given session.
func (a *App) GetSessionTracks(sessionID string) ([]diary.SessionTrack, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return a.store.GetSessionTracks(sessionID)
}

// GetTopArtists returns top artists for a given year/month.
func (a *App) GetTopArtists(year, month, limit int) ([]diary.ArtistStats, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return a.store.GetTopArtists(year, month, limit)
}

// GetTopTracks returns top tracks for a given year/month.
func (a *App) GetTopTracks(year, month, limit int) ([]diary.TrackStats, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return a.store.GetTopTracks(year, month, limit)
}

// GetMoodDistribution returns mood stats for a time range.
func (a *App) GetMoodDistribution(startDate, endDate string) ([]diary.MoodStat, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return a.store.GetMoodDistribution(startDate, endDate)
}

// GetWeatherCorrelation returns listening hours grouped by weather condition.
func (a *App) GetWeatherCorrelation(year int) ([]diary.WeatherStat, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return a.store.GetWeatherCorrelation(year)
}

// GetListeningByHour returns listening hours grouped by hour of day.
func (a *App) GetListeningByHour(year int) ([]diary.HourStat, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return a.store.GetListeningByHour(year)
}

// GetListeningByDayOfWeek returns listening hours grouped by day of week.
func (a *App) GetListeningByDayOfWeek(year int) ([]diary.DayOfWeekStat, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return a.store.GetListeningByDayOfWeek(year)
}

// GetDatabaseInfo returns database file path and stats.
func (a *App) GetDatabaseInfo() map[string]interface{} {
	if a.store == nil {
		return map[string]interface{}{"error": "store not initialized"}
	}
	info, err := a.store.GetDatabaseInfo()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	return info
}

// ExportAnnualBook generates an HTML annual report and returns the file path.
func (a *App) ExportAnnualBook(year int) (string, error) {
	if a.store == nil {
		return "", fmt.Errorf("store not initialized")
	}

	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           fmt.Sprintf("Save Vinyl %d Annual Book", year),
		DefaultFilename: fmt.Sprintf("vinyl-%d.html", year),
		Filters: []runtime.FileFilter{
			{DisplayName: "HTML Files", Pattern: "*.html"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("save dialog: %w", err)
	}
	if savePath == "" {
		return "", nil
	}

	return savePath, a.store.ExportAnnualBook(year, savePath)
}

// GetCurrentWeather returns the current weather data.
func (a *App) GetCurrentWeather() *weather.Data {
	if a.weather == nil {
		return nil
	}
	data, err := a.weather.GetCurrent()
	if err != nil {
		slog.Warn("weather fetch failed", "err", err)
		return nil
	}
	return data
}

// UpdateLocationFromCoords updates weather location from given coordinates.
func (a *App) UpdateLocationFromCoords(lat, lon float64) error {
	a.cfg.Location.Lat = lat
	a.cfg.Location.Lon = lon
	a.cfg.Location.City = ""
	a.weather.UpdateLocation(a.cfg.WeatherAPIKey, lat, lon)
	return config.Save(a.cfg)
}

// DetectLocation uses IP-based geolocation to find the user's approximate location.
func (a *App) DetectLocation() (map[string]interface{}, error) {
	return detectLocationViaIP()
}

// GetYearsWithData returns the years that have listening data.
func (a *App) GetYearsWithData() ([]int, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return a.store.GetYearsWithData()
}

// GetYearStats returns aggregate stats for a given year (for export preview).
func (a *App) GetYearStats(year int) (map[string]interface{}, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return a.store.GetYearStats(year)
}

// GetSessionsByDate returns sessions for a specific date.
func (a *App) GetSessionsByDate(date string) ([]diary.Session, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return a.store.GetSessionsByDate(date)
}

// GetInsights returns generated insights about listening patterns.
func (a *App) GetInsights() ([]diary.Insight, error) {
	if a.store == nil {
		return nil, fmt.Errorf("store not initialized")
	}
	return a.store.GetInsights(time.Now().Year())
}

// ConnectSpotify starts the Spotify OAuth flow.
func (a *App) ConnectSpotify(clientID string) error {
	if a.detector == nil {
		return fmt.Errorf("detector not initialized")
	}
	a.cfg.Spotify.ClientID = clientID
	if err := config.Save(a.cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	return a.detector.ConnectSpotify(a.ctx, clientID)
}

// DisconnectSpotify removes the Spotify connection.
func (a *App) DisconnectSpotify() error {
	if a.detector == nil {
		return fmt.Errorf("detector not initialized")
	}
	a.cfg.Spotify.ClientID = ""
	if err := config.Save(a.cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	return a.detector.DisconnectSpotify()
}

// IsSpotifyConnected checks if Spotify is authenticated.
func (a *App) IsSpotifyConnected() bool {
	if a.detector == nil {
		return false
	}
	return a.detector.IsSpotifyConnected()
}
