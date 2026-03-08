package player

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// GetAppleMusicTrack detects the currently playing track from Apple Music via AppleScript.
func GetAppleMusicTrack() (*TrackInfo, error) {
	script := `
	if application "Music" is running then
		tell application "Music"
			if player state is playing then
				set t to current track
				set trackName to name of t
				set artistName to artist of t
				set albumName to album of t
				set trackDuration to duration of t
				set playerPos to player position
				return trackName & "||" & artistName & "||" & albumName & "||" & (trackDuration as string) & "||" & (playerPos as string)
			else
				return "NOT_PLAYING"
			end if
		end tell
	else
		return "NOT_RUNNING"
	end if
	`

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return nil, fmt.Errorf("apple music osascript: %w", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "NOT_RUNNING" || result == "NOT_PLAYING" {
		return nil, nil
	}

	parts := strings.Split(result, "||")
	if len(parts) < 5 {
		return nil, fmt.Errorf("unexpected apple music response: %s", result)
	}

	duration, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
	position, _ := strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)

	return &TrackInfo{
		Title:      strings.TrimSpace(parts[0]),
		Artist:     strings.TrimSpace(parts[1]),
		Album:      strings.TrimSpace(parts[2]),
		DurationMS: int(duration * 1000),
		PositionMS: int(position * 1000),
		Source:     "apple_music",
		IsPlaying:  true,
		DetectedAt: time.Now(),
	}, nil
}
