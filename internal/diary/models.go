package diary

import (
	"fmt"
	"time"
)

// Session represents a listening session.
type Session struct {
	ID           string  `json:"id"`
	StartedAt    string  `json:"started_at"`
	EndedAt      string  `json:"ended_at,omitempty"`
	DurationSec  int     `json:"duration_sec"`
	TrackCount   int     `json:"track_count"`
	Source       string  `json:"source"`
	Mood         string  `json:"mood,omitempty"`
	MoodSetAt    string  `json:"mood_set_at,omitempty"`
	WeatherTemp  float64 `json:"weather_temp,omitempty"`
	WeatherCond  string  `json:"weather_cond,omitempty"`
	WeatherDesc  string  `json:"weather_desc,omitempty"`
	WeatherHumid int     `json:"weather_humid,omitempty"`
	WeatherIcon  string  `json:"weather_icon,omitempty"`
	DayOfWeek    int     `json:"day_of_week"`
	HourOfDay    int     `json:"hour_of_day"`
	Month        int     `json:"month"`
	Year         int     `json:"year"`
	DateLocal    string  `json:"date_local"`
	TimeOfDay    string  `json:"time_of_day"`
}

// SessionTrack represents an individual track within a session.
type SessionTrack struct {
	ID                string  `json:"id"`
	SessionID         string  `json:"session_id"`
	Title             string  `json:"title"`
	Artist            string  `json:"artist"`
	Album             string  `json:"album"`
	DurationMS        int     `json:"duration_ms"`
	ListenedMS        int     `json:"listened_ms"`
	AlbumArtURL       string  `json:"album_art_url,omitempty"`
	AlbumArtPath      string  `json:"album_art_path,omitempty"`
	SpotifyID         string  `json:"spotify_id,omitempty"`
	Genres            string  `json:"genres,omitempty"`
	Energy            float64 `json:"energy,omitempty"`
	Valence           float64 `json:"valence,omitempty"`
	Tempo             float64 `json:"tempo,omitempty"`
	Source            string  `json:"source"`
	PlayedAt          string  `json:"played_at"`
	PositionInSession int     `json:"position_in_session"`
}

// DailyStats holds pre-aggregated daily statistics for the heatmap.
type DailyStats struct {
	DateLocal       string  `json:"date_local"`
	TotalListenSec  int     `json:"total_listen_sec"`
	SessionCount    int     `json:"session_count"`
	TrackCount      int     `json:"track_count"`
	DominantMood    string  `json:"dominant_mood,omitempty"`
	AvgTemp         float64 `json:"avg_temp,omitempty"`
	DominantWeather string  `json:"dominant_weather,omitempty"`
}

// ArtistStats holds aggregate stats for an artist.
type ArtistStats struct {
	Name         string `json:"name"`
	TotalMinutes int    `json:"total_minutes"`
	TrackPlays   int    `json:"track_plays"`
	Sessions     int    `json:"sessions"`
}

// TrackStats holds aggregate stats for a track.
type TrackStats struct {
	Title        string `json:"title"`
	Artist       string `json:"artist"`
	Album        string `json:"album"`
	PlayCount    int    `json:"play_count"`
	TotalMinutes int    `json:"total_minutes"`
}

// MoodStat represents mood distribution data.
type MoodStat struct {
	Mood       string  `json:"mood"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// WeatherStat represents listening by weather condition.
type WeatherStat struct {
	Condition    string  `json:"condition"`
	SessionCount int     `json:"session_count"`
	TotalHours   float64 `json:"total_hours"`
}

// HourStat represents listening by hour of day.
type HourStat struct {
	Hour       int     `json:"hour"`
	TotalHours float64 `json:"total_hours"`
	Sessions   int     `json:"sessions"`
}

// DayOfWeekStat represents listening by day of week.
type DayOfWeekStat struct {
	Day        int     `json:"day"`
	TotalHours float64 `json:"total_hours"`
	Sessions   int     `json:"sessions"`
}

// Insight represents a generated pattern insight.
type Insight struct {
	Icon        string `json:"icon"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Stat        string `json:"stat"`
}

// ValidMoods lists all valid mood values.
var ValidMoods = []string{
	"happy", "calm", "energetic", "sad", "thoughtful", "frustrated",
	"in_love", "dreamy", "motivated", "celebrating", "sleepy", "nostalgic",
}

// TimeOfDay derives time-of-day category from hour.
func TimeOfDay(hour int) string {
	switch {
	case hour >= 5 && hour < 8:
		return "dawn"
	case hour >= 8 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 17:
		return "afternoon"
	case hour >= 17 && hour < 21:
		return "evening"
	default:
		return "night"
	}
}

// NewULID generates a new ULID string.
func NewULID() string {
	// Use time-based UUID as simple unique ID
	return time.Now().Format("20060102150405") + fmt.Sprintf("%06d", time.Now().Nanosecond()/1000)
}
