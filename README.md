<p align="center">
  <br>
  <img src="build/appicon.png" width="120" alt="Vinyl icon">
  <br><br>
</p>

<h1 align="center">V I N Y L</h1>

<p align="center">
  <em>Every song is a memory. Vinyl remembers them all.</em>
</p>

<p align="center">
  A macOS menubar app that turns your music listening into an emotional diary.<br>
  It detects what you play, logs every session with weather and mood,<br>
  and over time reveals patterns you never knew you had.
</p>

---

## What it does

Vinyl lives quietly in your menubar. When you play music in **Spotify** or **Apple Music**, it:

- Detects the track in real-time (title, artist, album, album art)
- Groups continuous playback into **listening sessions**
- Captures the **weather** at the start of each session
- Lets you tag your **mood** with a single tap
- Builds a **heatmap calendar** of your listening activity
- Generates **insights** about your patterns ("You listen to jazz when it rains")
- Exports a beautiful **annual book** of your musical year

All data stays **100% local** in a SQLite database. No cloud, no telemetry, no accounts.

## Screenshots

```
+-----------------------------------+    +-----------------------------------+
|         NOW PLAYING               |    |           DIARY                   |
|                                   |    |                                   |
|        [Spinning Vinyl]           |    |   [52-week heatmap calendar]      |
|                                   |    |                                   |
|    "Karma Police"                 |    |   Today                           |
|    Radiohead -- OK Computer       |    |   +---------------------------+   |
|                                   |    |   | 14:30-16:45 . 2h 15m     |   |
|    ====[====*==========] 3:42     |    |   | 12 tracks . Nostalgic     |   |
|                                   |    |   | Cloudy 17C . Spotify      |   |
|    MOOD                           |    |   +---------------------------+   |
|    [Happy] [Calm] [Energetic]     |    |                                   |
|    [Sad] [Thoughtful] [Dreamy]    |    |   Yesterday                      |
|                                   |    |   +---------------------------+   |
|    Clouds 18C . Session: 47m      |    |   | 09:00-10:15 . 1h 15m     |   |
|                                   |    |   | 7 tracks . Calm           |   |
|   [Now] [Diary] [Stats] [Gear]   |    |   +---------------------------+   |
+-----------------------------------+    +-----------------------------------+
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.22+ |
| Framework | [Wails v2](https://wails.io) |
| Frontend | Svelte 4 + Tailwind CSS 3 |
| Database | SQLite (WAL mode) via go-sqlite3 |
| Platform | macOS (arm64 + amd64) |

**Zero external runtime dependencies** -- single self-contained `.app` bundle.

## Getting Started

### Prerequisites

- Go 1.22+
- Node.js 18+
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Xcode Command Line Tools

### Development

```bash
wails dev
```

This starts the app with hot-reload for both Go and Svelte code.

### Production Build

```bash
wails build
```

Output: `build/bin/Vinyl.app`

### Run Tests

```bash
go test ./...
```

## Setup

On first launch, Vinyl guides you through a quick setup:

1. **Location** -- click "Allow Location Access" to auto-detect your city via IP, or type it manually
2. **Weather API Key** -- get a free key at [openweathermap.org/api](https://openweathermap.org/api) (1000 calls/day)
3. **Spotify** *(optional)* -- connect your Spotify account for richer metadata (genres, audio features, hi-res art). Requires a [Spotify Developer](https://developer.spotify.com) app with redirect URI `http://localhost:27750/callback`

Apple Music works automatically via AppleScript -- no setup needed.

## Architecture

```
vinyl/
|-- main.go, app.go, location.go    # Wails app entry + bindings
|-- internal/
|   |-- config/                      # JSON config (~/.vinyl/config.json)
|   |-- player/
|   |   |-- detector.go              # Polling orchestrator + session state machine
|   |   |-- spotify.go               # Spotify API (OAuth2 PKCE) + AppleScript fallback
|   |   +-- applemusic.go            # Apple Music via AppleScript
|   |-- weather/                     # OpenWeatherMap client with 30-min cache
|   +-- diary/
|       |-- store.go                 # SQLite repository (CRUD + analytics)
|       |-- migrations.go            # Schema versioning
|       +-- models.go                # Domain types
|-- frontend/
|   +-- src/
|       |-- App.svelte               # Root with tab router
|       |-- views/                   # NowPlaying, Diary, Insights, Settings, Onboarding
|       +-- lib/
|           |-- components/           # VinylDisc, Heatmap, MoodPicker, SessionList, ...
|           |-- stores/               # Svelte stores for track, session, navigation
|           +-- utils/                # Formatters, emoji maps
+-- build/                           # App icons, macOS plist
```

### Session Lifecycle

```
IDLE --> track detected --> ACTIVE --> pause/stop --> ENDING --> 30s gap --> persist --> IDLE
                              |                        |
                              |   track changes        |   play resumes
                              +-- update track list    +-- back to ACTIVE

                              Session < 60s at end --> discard
```

### Music Detection Priority

1. Spotify Web API (if OAuth token available)
2. Spotify via AppleScript (fallback)
3. Apple Music via AppleScript

### Data Storage

| What | Where |
|------|-------|
| Config | `~/.vinyl/config.json` |
| Database | `~/Library/Application Support/Vinyl/vinyl.db` |
| Spotify token | macOS Keychain (`com.vinyl.app`) |

## Design

The UI evokes the warmth of a 1970s hi-fi listening room:

- **Colors**: warm browns, ambers, and golds on deep warm black
- **Typography**: Playfair Display (headings), Source Serif 4 (body), JetBrains Mono (data)
- **Textures**: SVG noise overlay, radial vignette, vinyl groove patterns
- **Animation**: vinyl record spinning at ~33 RPM with independent light reflection

## Privacy

- All data stored locally in SQLite. Nothing leaves your machine.
- The only external API calls are to OpenWeatherMap (weather) and Spotify (if connected).
- Spotify refresh tokens are stored in macOS Keychain, never in plaintext.
- No analytics, no telemetry, no tracking.

## Annual Book Export

Vinyl can generate a beautiful HTML annual report of your listening year, structured by month with:

- Top artists and tracks
- Mood distribution charts
- Weather correlations
- Listening by hour and day of week
- Notable sessions

The output is a single self-contained HTML file optimized for printing as PDF via your browser.

## License

MIT

---

<p align="center">
  <sub>Made with music in Pescara.</sub>
</p>
