# CLAUDE.md — Vinyl

## Project Overview

**Vinyl** is a macOS menubar application that creates an emotional music diary. It detects currently playing music from Spotify and Apple Music, displays a rotating vinyl record animation in the menubar, and silently logs every listening session with contextual metadata (weather, time of day, optional mood tags). Over time it builds a rich emotional map of the user's musical life.

**Stack**: Go 1.22+ backend, Wails v2, Svelte 4 frontend, SQLite (via go-sqlite3), Tailwind CSS 3.

## Tech Stack & Constraints

- **Language**: Go 1.22+
- **Framework**: Wails v2 (https://wails.io)
- **Frontend**: Svelte 4 + Tailwind CSS 3 + vanilla JS for animations
- **Database**: SQLite via `github.com/mattn/go-sqlite3` — single file `~/Library/Application Support/Vinyl/vinyl.db`
- **Platform**: macOS only (arm64 + amd64)
- **Build**: `wails build` for production, `wails dev` for development
- **No external runtime dependencies** — fully self-contained binary
- **All code, comments, variable names, commit messages in English**
- **Config file**: `~/.vinyl/config.json`

## Architecture Pattern

- Backend (Go): music detection, weather API, database, system tray, scheduling
- Frontend (Svelte): popup UI rendered in Wails webview, all views
- Communication: Wails bindings (Go functions exposed to JS and vice versa via `runtime.EventsEmit` / `runtime.EventsOn`)

## Key Directories

```
vinyl/
├── CLAUDE.md
├── wails.json
├── main.go                    # Entry point
├── app.go                     # Wails app struct, lifecycle, bindings
├── internal/
│   ├── player/
│   │   ├── detector.go        # Unified player detection orchestrator
│   │   ├── spotify.go         # Spotify integration (API + AppleScript fallback)
│   │   ├── applemusic.go      # Apple Music via AppleScript/MusicKit
│   │   └── types.go           # TrackInfo, PlayerState types
│   ├── weather/
│   │   ├── client.go          # OpenWeatherMap API client
│   │   └── types.go           # WeatherData type
│   ├── diary/
│   │   ├── store.go           # SQLite repository (CRUD for sessions, moods)
│   │   ├── migrations.go      # Schema migrations
│   │   ├── models.go          # Session, MoodTag, DailySummary models
│   │   └── analytics.go       # Pattern detection, heatmap data, insights
│   ├── export/
│   │   ├── book.go            # Annual PDF/HTML book export
│   │   └── templates/         # Go templates for book pages
│   └── config/
│       └── config.go          # Configuration management
├── frontend/
│   ├── src/
│   │   ├── App.svelte         # Root with view router
│   │   ├── lib/
│   │   │   ├── stores/        # Svelte stores (currentTrack, sessions, etc.)
│   │   │   ├── components/
│   │   │   │   ├── VinylDisc.svelte      # Animated rotating vinyl
│   │   │   │   ├── NowPlaying.svelte     # Current track display
│   │   │   │   ├── MoodPicker.svelte     # Emoji/word mood selector
│   │   │   │   ├── Heatmap.svelte        # Calendar heatmap
│   │   │   │   ├── SessionList.svelte    # Recent sessions timeline
│   │   │   │   ├── InsightCard.svelte    # Pattern insight display
│   │   │   │   └── ExportWizard.svelte   # Annual book export
│   │   │   └── utils/
│   │   │       ├── animations.js          # Vinyl rotation, transitions
│   │   │       └── formatters.js          # Time, duration helpers
│   │   └── views/
│   │       ├── NowPlayingView.svelte     # Main popup view
│   │       ├── DiaryView.svelte          # Session history + heatmap
│   │       ├── InsightsView.svelte       # Pattern analysis
│   │       ├── ExportView.svelte         # Book export
│   │       └── SettingsView.svelte       # Configuration
│   ├── index.html
│   ├── package.json
│   └── tailwind.config.js
└── build/
    └── appicon.png
```

## Critical Implementation Notes

### Music Detection

1. **Spotify**: Use Spotify Web API for rich metadata (album art, genres, audio features). Requires OAuth2 PKCE flow. Fallback to AppleScript (`osascript`) for basic now-playing if no API token.
2. **Apple Music**: Use AppleScript via `osascript` to query the Music app. No API key needed. Returns track name, artist, album, player state, position.
3. **Polling interval**: Every 5 seconds when idle, every 2 seconds when playing.
4. **Session logic**: A "listening session" starts when playback begins and ends after 30 seconds of silence/pause. Minimum session duration to persist: 60 seconds.

### AppleScript Examples

```applescript
-- Spotify
tell application "Spotify"
    if player state is playing then
        set trackName to name of current track
        set artistName to artist of current track
        set albumName to album of current track
        set artUrl to artwork url of current track
        set trackDuration to duration of current track
        set playerPosition to player position
    end if
end tell

-- Apple Music
tell application "Music"
    if player state is playing then
        set trackName to name of current track
        set artistName to artist of current track
        set albumName to album of current track
        set trackDuration to duration of current track
        set playerPosition to player position
    end if
end tell
```

### Weather Integration

- Use OpenWeatherMap free tier (1000 calls/day)
- Cache weather data for 30 minutes
- Fetch on session start, attach to session record
- Store: temp (°C), condition (clear/clouds/rain/snow/etc.), humidity, description
- Default location: configurable in settings (lat/lon or city name)

### System Tray Behavior

- **Icon**: Custom vinyl disc icon (static when idle, animated spinning when music plays)
- **Left click**: Opens/toggles popup window (Wails webview)
- **Right click**: Context menu with: "Now Playing", "Open Diary", "Settings", separator, "Quit Vinyl"
- **Popup size**: 380px wide × 520px tall, anchored to menubar icon
- **Popup should feel native**: no title bar, rounded corners, shadow, vibrancy (if possible via Wails)

### Session Persistence

- Every detected track change during a session updates the session's track list
- On session end: calculate total duration, dominant genre, track count
- Weather is fetched once at session start
- Mood tag is optional — user can add it during or after session via popup

### Heatmap

- GitHub-style contribution heatmap showing listening activity
- X axis: weeks, Y axis: days of week
- Color intensity: minutes listened per day
- Covers rolling 52 weeks
- Click on a day: shows sessions for that day

### Annual Export ("Il Libro")

- Generates a beautiful HTML document (printable as PDF)
- Structured by months
- Each month: top artists, top tracks, total listening time, mood distribution, weather correlation, notable sessions (longest, most diverse, etc.)
- Cover page with year, total stats, and generative art based on listening data
- Use Go `html/template` for rendering

## Code Style

- Go: standard `gofmt`, error wrapping with `fmt.Errorf("context: %w", err)`
- Svelte: single-file components, minimal external dependencies
- CSS: Tailwind utilities + custom CSS for vinyl animation and vintage textures
- No global state mutations — use Svelte stores exclusively
- All database access through `diary.Store` interface
- All player access through `player.Detector` interface

## Testing

- Go: table-driven tests for `diary/analytics.go` and `player/` detection logic
- Frontend: manual testing via `wails dev` (no unit test framework required for v1)

## Configuration (config.json)

```json
{
  "location": {
    "lat": 42.4612,
    "lon": 14.2111,
    "city": "Pescara"
  },
  "weather_api_key": "",
  "spotify": {
    "client_id": "",
    "redirect_uri": "http://localhost:27750/callback"
  },
  "polling_interval_idle_ms": 5000,
  "polling_interval_playing_ms": 2000,
  "min_session_duration_sec": 60,
  "session_gap_sec": 30
}
```
