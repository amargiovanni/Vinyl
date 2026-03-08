# SPEC.md — Vinyl Product Specification

## 1. Vision

Vinyl is a macOS menubar app that turns your music listening into an emotional diary. It lives quietly in your menubar as a spinning vinyl record, observes what you listen to, and weaves together music, weather, time, and mood into a rich personal archive. After months of silent observation, Vinyl reveals patterns you never knew you had.

**Tagline**: _"Every song is a memory. Vinyl remembers them all."_

## 2. User Stories

### Core (P0 — Must Have for v1)

| ID | Story | Acceptance Criteria |
|----|-------|-------------------|
| US-01 | As a user, I see a vinyl disc icon in my menubar that spins when music plays | Icon animates at ~33rpm equivalent. Stops when no music detected. Works with Spotify and Apple Music. |
| US-02 | As a user, I click the icon and see what's currently playing | Popup shows: album art, track name, artist, album, progress bar. Updates in real-time. Vintage aesthetic. |
| US-03 | As a user, my listening sessions are automatically logged | Session starts on play, ends after 30s gap. Stores: tracks, duration, start/end time, weather, source (Spotify/AM). Min 60s to persist. |
| US-04 | As a user, I can tag my current mood during a session | MoodPicker shows 8-12 emoji/word pairs. One tap to tag. Optional — session saves without mood if skipped. Mood can be edited after session ends. |
| US-05 | As a user, I see a heatmap of my listening history | GitHub-style grid, 52 weeks rolling. Color intensity = minutes listened. Click day to see sessions. |
| US-06 | As a user, I can browse my past sessions as a diary | Scrollable timeline grouped by day. Each session shows: time range, tracks played, mood, weather icon, total duration. |
| US-07 | As a user, I can export my annual listening diary as a printable book | HTML export structured by month. Includes: stats, top artists/tracks, mood patterns, weather correlations, cover page. Printable via browser as PDF. |
| US-08 | As a user, I can configure my location for weather data | Settings view with city name or lat/lon input. Used for OpenWeatherMap queries. |
| US-09 | As a user, I can connect my Spotify account | OAuth2 PKCE flow launched from settings. Token stored securely. Enables rich metadata (genres, audio features, hi-res art). |

### Nice-to-Have (P1 — Target for v1 if time allows)

| ID | Story |
|----|-------|
| US-10 | As a user, I see AI-generated insights about my listening patterns ("You listen to jazz when it rains") |
| US-11 | As a user, I can see listening stats for any custom time range |
| US-12 | As a user, the app launches automatically at login |
| US-13 | As a user, I get a gentle weekly summary notification |

### Future (P2 — Post v1)

| ID | Story |
|----|-------|
| US-20 | Last.fm scrobble integration |
| US-21 | Plex integration |
| US-22 | Shared social features ("listening alongside") |
| US-23 | iOS companion app |

## 3. Feature Specifications

### 3.1 Music Detection Engine

**Behavior**:
- Polls every 5s when idle, every 2s when active playback detected
- Detects player: Spotify (preferred) → Apple Music → None
- If Spotify API token available: use Web API for metadata (genres, audio features, 640px art)
- If Spotify API token unavailable: use AppleScript for basic info
- Apple Music: always via AppleScript (no API)
- Track change detection: compare `(artist, track_name, album)` tuple

**Session lifecycle**:
```
IDLE → track detected → SESSION_ACTIVE
SESSION_ACTIVE → track changes → update session track list
SESSION_ACTIVE → pause/stop → start 30s gap timer → SESSION_ENDING
SESSION_ENDING → play resumes within 30s → SESSION_ACTIVE
SESSION_ENDING → 30s elapsed → persist session → IDLE
SESSION_ACTIVE → duration < 60s at end → discard (don't persist)
```

**TrackInfo fields**:
- `title` (string, required)
- `artist` (string, required)
- `album` (string, required)
- `album_art_url` (string, optional — Spotify only for hi-res)
- `duration_ms` (int)
- `genres` ([]string, optional — Spotify API only)
- `source` ("spotify" | "apple_music")
- `spotify_id` (string, optional)
- `started_at` (timestamp)
- `ended_at` (timestamp)

### 3.2 Weather Context

**On session start**:
1. Check cache (valid for 30 min)
2. If stale: call OpenWeatherMap Current Weather API
3. Store with session

**WeatherData fields**:
- `temp_celsius` (float)
- `condition` (string: "Clear", "Clouds", "Rain", "Drizzle", "Snow", "Thunderstorm", "Mist", "Fog")
- `description` (string: "light rain", "scattered clouds", etc.)
- `humidity` (int, percentage)
- `icon_code` (string: OpenWeatherMap icon code, e.g., "10d")
- `fetched_at` (timestamp)

### 3.3 Mood Tagging

**Mood vocabulary** (emoji + label):

| Emoji | Label |
|-------|-------|
| 😊 | Happy |
| 😌 | Calm |
| 🔥 | Energetic |
| 😢 | Sad |
| 🤔 | Thoughtful |
| 😤 | Frustrated |
| 🥰 | In Love |
| 🌙 | Dreamy |
| 💪 | Motivated |
| 🎉 | Celebrating |
| 😴 | Sleepy |
| 🌊 | Nostalgic |

**Behavior**:
- MoodPicker visible in NowPlaying view during active session
- Single selection (one mood per session)
- Tapping same mood again deselects it
- Mood can be changed/added retrospectively from session detail in diary view
- Mood is stored as enum string, not emoji (for portability)

### 3.4 Heatmap Calendar

**Visual spec**:
- Grid: 52 columns (weeks) × 7 rows (Mon-Sun)
- Cell size: 12×12px with 3px gap
- Color scale: 5 levels based on minutes listened
  - 0 min: `#2a2018` (dark brown, empty)
  - 1-30 min: `#5c4a3a` (warm brown)
  - 31-90 min: `#8b6d4f` (medium amber)
  - 91-180 min: `#c49a6c` (warm gold)
  - 180+ min: `#e8c496` (bright gold)
- Today highlighted with subtle border
- Month labels above grid
- Day labels (M, W, F) on left

**Interaction**:
- Hover: tooltip with date and "Xh Ym listened"
- Click: scrolls diary view to that day's sessions

### 3.5 Diary / Session History

**Layout**:
- Reverse chronological (newest first)
- Grouped by day with date header
- Each session card shows:
  - Time range: "14:30 — 16:45"
  - Duration: "2h 15m"
  - Track count: "12 tracks"
  - Top track (most played during session)
  - Mood emoji (if set)
  - Weather icon + temp
  - Source badge (Spotify/Apple Music icon)
- Click session card: expands to show full track list with individual durations

### 3.6 Annual Book Export

**Structure**:

```
COVER PAGE
├── Year (large, vintage typography)
├── Total listening hours
├── Total sessions
├── Total unique artists
├── Top 3 artists (with listen counts)
└── Decorative element (generative based on dominant mood)

FOR EACH MONTH:
├── Month title page
│   ├── Month name + year
│   ├── Total hours listened
│   ├── Sessions count
│   └── Dominant mood
├── Top Artists (max 10)
│   ├── Artist name
│   ├── Listen count
│   └── Total minutes
├── Top Tracks (max 10)
│   ├── Track name — Artist
│   ├── Times played
│   └── Total minutes
├── Mood Distribution
│   └── Horizontal bar chart (emoji + percentage)
├── Weather Correlation
│   └── "You listened X hours during rain, Y during sun"
├── Notable Sessions
│   ├── Longest session
│   ├── Most diverse session (most genres)
│   └── Late night sessions (after midnight)
└── Monthly calendar mini-heatmap

YEAR IN REVIEW (final page)
├── Listening by hour of day (bar chart)
├── Listening by day of week
├── Seasonal breakdown
├── Mood journey (month-by-month mood shifts)
└── "Your year in one sentence" (generated summary)
```

**Output**: Single HTML file with inline CSS, optimized for A4 print. User prints via browser's Print → Save as PDF.

**Typography for book**: Use Google Fonts loaded inline — `Playfair Display` for headings, `Lora` for body text.

## 4. System Tray & Popup

### Menubar Icon

- **Idle**: Static vinyl disc icon (16×16 template image for macOS dark/light)
- **Playing**: Animated spinning disc. Rotation via frontend CSS when popup is open; for the tray icon itself, use 4-frame animation (0°, 90°, 180°, 270°) cycled at ~200ms interval
- **Icon style**: Simple, recognizable at 16×16 — concentric circles suggesting vinyl grooves

### Popup Window

- **Size**: 380px × 520px
- **Position**: Anchored below menubar icon (standard macOS popover position)
- **Behavior**: Toggle on left-click. Dismiss on click outside or Escape key.
- **Navigation**: Tab bar at bottom with icons: 🎵 Now Playing | 📖 Diary | 📊 Insights | ⚙️ Settings
- **Transitions**: Slide left/right between views (200ms ease-out)

### Context Menu (Right-click)

```
Now Playing          ▶ (opens popup on Now Playing view)
Open Diary           ▶ (opens popup on Diary view)
─────────────────
Listening: Yes/No    (status indicator, not clickable)
Session: 45m         (current session duration, not clickable)
─────────────────
Settings...
Quit Vinyl
```

## 5. Non-Functional Requirements

| Requirement | Target |
|-------------|--------|
| Memory usage | < 50MB RSS when idle, < 80MB during active session |
| CPU (idle) | < 0.5% |
| CPU (polling) | < 2% |
| Database size | < 100MB after 1 year of heavy use |
| Startup time | < 2 seconds to tray icon visible |
| Popup open time | < 200ms |
| Binary size | < 30MB |
| macOS support | 12.0 (Monterey) and above |
| Data privacy | All data stored locally. No telemetry. Weather API is the only external call. |

## 6. Privacy & Data

- **Zero cloud**: All data in local SQLite. No sync, no upload, no analytics.
- **Spotify token**: Stored in macOS Keychain (via `go-keyring` or similar). Never written to plaintext files.
- **Weather API key**: Stored in config.json (user provides their own key).
- **Export**: HTML file generated locally. User decides what to do with it.
- **Album art**: Cached locally in `~/Library/Caches/Vinyl/art/`. Auto-pruned after 90 days of no access.

## 7. First Run Experience

1. App opens popup automatically on first launch
2. Welcome screen: "Welcome to Vinyl" with brief description
3. Step 1: Set your location (city search or manual lat/lon)
4. Step 2: Enter OpenWeatherMap API key (with link to free signup)
5. Step 3 (optional): Connect Spotify (skip for Apple Music only)
6. Done: "Start playing music. Vinyl will take care of the rest."
7. Popup closes, vinyl icon appears in menubar
