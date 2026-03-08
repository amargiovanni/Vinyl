package player

import (
	"testing"
	"time"
)

func TestSameTrack(t *testing.T) {
	a := &TrackInfo{Title: "Karma Police", Artist: "Radiohead", Album: "OK Computer"}
	b := &TrackInfo{Title: "Karma Police", Artist: "Radiohead", Album: "OK Computer"}
	c := &TrackInfo{Title: "Lucky", Artist: "Radiohead", Album: "OK Computer"}

	if !SameTrack(a, b) {
		t.Error("expected a and b to be the same track")
	}
	if SameTrack(a, c) {
		t.Error("expected a and c to be different tracks")
	}
	if SameTrack(nil, a) {
		t.Error("expected nil and a to be different")
	}
	if SameTrack(a, nil) {
		t.Error("expected a and nil to be different")
	}
	if SameTrack(nil, nil) {
		t.Error("expected nil and nil to be different")
	}
}

func TestPlayerStateString(t *testing.T) {
	tests := []struct {
		state    PlayerState
		expected string
	}{
		{StateIdle, "idle"},
		{StateActive, "active"},
		{StateEnding, "ending"},
		{PlayerState(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("PlayerState(%d).String() = %q, want %q", tt.state, got, tt.expected)
		}
	}
}

func TestTrackInfoFields(t *testing.T) {
	track := TrackInfo{
		Title:      "Test Song",
		Artist:     "Test Artist",
		Album:      "Test Album",
		DurationMS: 240000,
		PositionMS: 120000,
		Source:     "spotify",
		IsPlaying:  true,
		DetectedAt: time.Now(),
	}

	if track.Title != "Test Song" {
		t.Errorf("expected title Test Song, got %s", track.Title)
	}
	if track.DurationMS != 240000 {
		t.Errorf("expected duration 240000, got %d", track.DurationMS)
	}
	if track.Source != "spotify" {
		t.Errorf("expected source spotify, got %s", track.Source)
	}
}
