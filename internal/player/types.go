package player

import "time"

// TrackInfo represents a currently playing track.
type TrackInfo struct {
	Title       string   `json:"title"`
	Artist      string   `json:"artist"`
	Album       string   `json:"album"`
	AlbumArtURL string   `json:"album_art_url,omitempty"`
	DurationMS  int      `json:"duration_ms"`
	PositionMS  int      `json:"position_ms"`
	Genres      []string `json:"genres,omitempty"`
	Source      string   `json:"source"`
	SpotifyID   string   `json:"spotify_id,omitempty"`
	IsPlaying   bool     `json:"is_playing"`
	Energy      float64  `json:"energy,omitempty"`
	Valence     float64  `json:"valence,omitempty"`
	Tempo       float64  `json:"tempo,omitempty"`
	DetectedAt  time.Time `json:"detected_at"`
}

// PlayerState represents the state of the detection engine.
type PlayerState int

const (
	StateIdle         PlayerState = iota
	StateActive
	StateEnding
)

func (s PlayerState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateActive:
		return "active"
	case StateEnding:
		return "ending"
	default:
		return "unknown"
	}
}

// SameTrack returns true if two track infos represent the same track.
func SameTrack(a, b *TrackInfo) bool {
	if a == nil || b == nil {
		return false
	}
	return a.Artist == b.Artist && a.Title == b.Title && a.Album == b.Album
}
