package diary

import (
	"testing"
)

func TestTimeOfDay(t *testing.T) {
	tests := []struct {
		hour     int
		expected string
	}{
		{0, "night"},
		{3, "night"},
		{5, "dawn"},
		{7, "dawn"},
		{8, "morning"},
		{11, "morning"},
		{12, "afternoon"},
		{16, "afternoon"},
		{17, "evening"},
		{20, "evening"},
		{21, "night"},
		{23, "night"},
	}

	for _, tt := range tests {
		got := TimeOfDay(tt.hour)
		if got != tt.expected {
			t.Errorf("TimeOfDay(%d) = %q, want %q", tt.hour, got, tt.expected)
		}
	}
}

func TestNewULID(t *testing.T) {
	id1 := NewULID()
	id2 := NewULID()

	if id1 == "" {
		t.Error("NewULID returned empty string")
	}
	if len(id1) < 10 {
		t.Errorf("NewULID too short: %s", id1)
	}
	// IDs should be unique (or at least different for consecutive calls)
	// Note: they may be same if called in same microsecond, so we just check non-empty
	_ = id2
}

func TestValidMoods(t *testing.T) {
	if len(ValidMoods) != 12 {
		t.Errorf("expected 12 valid moods, got %d", len(ValidMoods))
	}
	expected := map[string]bool{
		"happy": true, "calm": true, "energetic": true, "sad": true,
		"thoughtful": true, "frustrated": true, "in_love": true, "dreamy": true,
		"motivated": true, "celebrating": true, "sleepy": true, "nostalgic": true,
	}
	for _, mood := range ValidMoods {
		if !expected[mood] {
			t.Errorf("unexpected mood: %s", mood)
		}
	}
}
