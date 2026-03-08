# TASKS.md — Vinyl Implementation Plan

## Overview

This document breaks the Vinyl project into ordered implementation tasks. Each task is self-contained, testable, and builds on previous ones. Designed for sequential execution with Claude Code.

---

## Phase 1: Skeleton (Tasks 1–4)

### Task 1: Project Scaffold
**Goal**: Working Wails app with tray icon and empty popup.
- Initialize Wails v2 project: `wails init -n vinyl -t svelte`
- Configure `wails.json` with app metadata
- Set up Go module with dependencies (sqlite3, ulid, go-keyring)
- Set up frontend with Svelte 4 + Tailwind CSS 3
- Create tray icon (static) and popup window (380×520, empty)
- Verify: `wails dev` shows tray icon, click opens empty popup
- **Files**: `main.go`, `app.go`, `wails.json`, `frontend/package.json`, `frontend/tailwind.config.js`

### Task 2: Configuration System
**Goal**: Load/save config from `~/.vinyl/config.json`.
- Implement `internal/config/config.go` with Config struct
- Default values (Pescara location, polling intervals)
- Load on startup, create default if missing
- Expose `GetConfig()` and `SaveConfig()` as Wails bindings
- **Files**: `internal/config/config.go`

### Task 3: Database Setup
**Goal**: SQLite database with schema migration on startup.
- Implement `internal/diary/store.go` with Store struct
- Implement `internal/diary/migrations.go` — Migration 001 (full schema from DATA_MODEL.md)
- Auto-create database at `~/Library/Application Support/Vinyl/vinyl.db`
- Auto-migrate on startup
- Verify: app starts, database file created with all tables
- **Files**: `internal/diary/store.go`, `internal/diary/migrations.go`, `internal/diary/models.go`

### Task 4: Frontend Shell
**Goal**: Popup with tab navigation and placeholder views.
- Implement `App.svelte` with view router (simple reactive variable, no router library)
- Create 4 views as empty placeholders: NowPlayingView, DiaryView, InsightsView, SettingsView
- Implement tab bar at bottom (icons, active state with amber underline)
- Apply base vintage theme: colors, fonts, noise overlay, vignette
- Verify: popup opens, tabs switch between views, vintage aesthetic visible
- **Files**: All frontend `src/` files as per CLAUDE.md structure

---

## Phase 2: Music Detection (Tasks 5–8)

### Task 5: Apple Music Detection via AppleScript
**Goal**: Detect Apple Music playback state and current track.
- Implement `internal/player/applemusic.go`
- Execute AppleScript via `os/exec` to query Music app
- Parse response into `TrackInfo` struct
- Handle: not running, not playing, playing states
- Album art extraction to temp file
- **Files**: `internal/player/applemusic.go`, `internal/player/types.go`

### Task 6: Spotify Detection via AppleScript
**Goal**: Detect Spotify playback as fallback (no API token yet).
- Implement `internal/player/spotify.go` — AppleScript detection method
- Same pattern as Apple Music but for Spotify app
- Parse artwork URL from Spotify's AppleScript interface
- **Files**: `internal/player/spotify.go`

### Task 7: Player Detection Orchestrator
**Goal**: Unified polling system that checks both players.
- Implement `internal/player/detector.go`
- Polling loop with configurable intervals (5s idle, 2s playing)
- Priority: Spotify → Apple Music
- Track change detection by comparing `(artist, title, album)` tuple
- Emit events to frontend via Wails runtime: `track:changed`, `playback:started`, `playback:stopped`
- Expose `GetCurrentTrack()` as Wails binding
- **Files**: `internal/player/detector.go`

### Task 8: Now Playing UI
**Goal**: Show current track in the popup with vinyl animation.
- Implement `VinylDisc.svelte` — spinning vinyl record with album art center
- Implement `NowPlaying.svelte` — album art (Polaroid frame), track info, progress bar
- Implement `NowPlayingView.svelte` — compose components
- Create Svelte store `currentTrack` that listens to Wails events
- Vinyl spins when playing, stops when paused
- Progress bar updates via polling
- Verify: play a song in Spotify/Apple Music, see it in Vinyl popup
- **Files**: Frontend components as listed

---

## Phase 3: Session Tracking (Tasks 9–12)

### Task 9: Session Lifecycle Engine
**Goal**: Automatically create, update, and persist listening sessions.
- Implement session state machine in `internal/player/detector.go`:
  - IDLE → SESSION_ACTIVE → SESSION_ENDING → IDLE
  - 30-second gap detection
  - 60-second minimum duration filter
- Create session in DB on start, update on track change, finalize on end
- Implement CRUD in `internal/diary/store.go`: `CreateSession`, `AddTrack`, `EndSession`
- Emit `session:started`, `session:ended` events
- **Files**: Updated `detector.go`, updated `store.go`

### Task 10: Weather Integration
**Goal**: Fetch weather on session start, attach to session.
- Implement `internal/weather/client.go` — OpenWeatherMap client
- Implement `internal/weather/types.go`
- Caching: in-memory + SQLite (30-min TTL)
- Fetch on session start, write weather fields to session record
- Graceful degradation: if API fails, session saves without weather
- Expose weather status in frontend for NowPlaying view
- **Files**: `internal/weather/client.go`, `internal/weather/types.go`

### Task 11: Mood Tagging
**Goal**: User can tag mood during or after a session.
- Implement `MoodPicker.svelte` — grid of 12 emoji/label pairs
- Single selection, tap to toggle
- Call Wails binding `SetSessionMood(sessionID, mood)` on selection
- Implement `UpdateSessionMood` in `diary/store.go`
- Mood editable from session detail in diary view (later task)
- **Files**: `MoodPicker.svelte`, updated `store.go`

### Task 12: Spotify OAuth & Rich Metadata
**Goal**: Connect Spotify account for genres, audio features, hi-res art.
- Implement OAuth2 PKCE flow in `internal/player/spotify.go`:
  - Local HTTP server on port 27750
  - Browser redirect for auth
  - Token exchange and refresh
  - Store refresh token in macOS Keychain
- When token available: use Spotify Web API for `currently-playing`
- Fetch audio features (energy, valence, tempo) for each track
- Fetch artist genres
- Settings UI: Connect/Disconnect Spotify button
- **Files**: `internal/player/spotify.go` (extended), `SettingsView.svelte`

---

## Phase 4: Diary & History (Tasks 13–16)

### Task 13: Daily Stats Aggregation
**Goal**: Pre-compute daily stats for heatmap performance.
- Implement `internal/diary/analytics.go`:
  - `UpdateDailyStats(date)` — aggregate from sessions table
  - Called on every session end
  - Updates `daily_stats` table
- Implement `GetHeatmapData(weeks int)` — returns 52-week grid data
- Implement `GetDailySessions(date string)` — returns sessions for a day
- **Files**: `internal/diary/analytics.go`

### Task 14: Heatmap Calendar Component
**Goal**: GitHub-style heatmap showing listening activity.
- Implement `Heatmap.svelte`:
  - 52 × 7 grid with configurable color scale
  - Month labels, day labels
  - Hover tooltip with date + hours listened
  - Click emits event with date (for diary navigation)
- Fetch data via Wails binding `GetHeatmapData()`
- **Files**: `Heatmap.svelte`

### Task 15: Session History / Diary View
**Goal**: Scrollable timeline of past sessions grouped by day.
- Implement `SessionList.svelte` — renders sessions grouped by date
- Implement `SessionCard` — compact session summary (time, tracks, mood, weather)
- Implement expand behavior — click card to see full track list
- Implement `DiaryView.svelte` — compose Heatmap + SessionList
- Month navigation (◀ ▶) to browse history
- Heatmap click → scroll to that day in session list
- Fetch via Wails bindings: `GetSessionsByMonth(year, month)`, `GetSessionTracks(sessionID)`
- **Files**: `SessionList.svelte`, `DiaryView.svelte`

### Task 16: Artist Stats Table
**Goal**: Update artist aggregate stats on every session end.
- Implement `UpdateArtistStats` in `diary/store.go`
- On session end: for each unique artist in session, update totals
- `GetTopArtists(year, month, limit)` query
- `GetTopTracks(year, month, limit)` query
- **Files**: Updated `store.go`, updated `analytics.go`

---

## Phase 5: Export & Polish (Tasks 17–20)

### Task 17: Annual Book — Data Gathering
**Goal**: Collect all data needed for annual export.
- Implement `internal/export/book.go`:
  - `GatherAnnualData(year int) *AnnualReport`
  - Collects: monthly stats, top artists/tracks per month, mood distributions, weather correlations, notable sessions, hourly/weekly patterns
- Define `AnnualReport`, `MonthReport` structs
- **Files**: `internal/export/book.go`

### Task 18: Annual Book — HTML Generation
**Goal**: Render annual data into a beautiful printable HTML document.
- Create Go `html/template` templates in `internal/export/templates/`:
  - `book.html` — main wrapper with print CSS
  - `cover.html` — cover page
  - `month.html` — month spread
  - `yearend.html` — year in review final page
- Inline all CSS (Google Fonts, vintage typography, print layout)
- Output single self-contained HTML file
- Use CSS `page-break-before` for print pagination
- Charts: pure CSS horizontal bars (no JS charting library)
- **Files**: `internal/export/book.go`, `internal/export/templates/*`

### Task 19: Export UI
**Goal**: User can trigger and save annual export from settings.
- Implement `ExportWizard.svelte`:
  - Select year to export
  - Preview stats summary before export
  - "Generate Book" button
  - Progress indicator
  - Save dialog (via Wails `runtime.SaveFileDialog`)
- Wire to Wails binding `ExportAnnualBook(year int) (string, error)`
- **Files**: `ExportWizard.svelte`, `SettingsView.svelte` (updated)

### Task 20: Settings View Complete
**Goal**: Full settings UI with all configuration options.
- Location input with geocoding (simple: user types city, we use it)
- Weather API key input (masked)
- Spotify connect/disconnect
- Apple Music status (always detected)
- Export section
- Database info (path, size, session count)
- About section with version
- Implement `SettingsView.svelte` fully
- **Files**: `SettingsView.svelte`

---

## Phase 6: System Integration (Tasks 21–23)

### Task 21: Context Menu
**Goal**: Right-click tray icon shows context menu.
- Implement via Wails system tray API
- Menu items: Now Playing, Open Diary, separator, listening status, session duration, separator, Settings, Quit
- Dynamic update: listening status and session duration refresh
- **Files**: `app.go` (tray menu setup)

### Task 22: Tray Icon Animation
**Goal**: Animated spinning vinyl in the actual menubar.
- Create 4 template PNG frames (16×16 @2x) of vinyl at 0°, 90°, 180°, 270°
- Cycle frames at ~200ms when playing
- Static frame when idle
- Use Wails/macOS tray icon API for frame cycling
- **Files**: `app.go`, `build/` assets

### Task 23: First Run & Onboarding
**Goal**: Welcome flow on first launch.
- Detect first run (no config file exists)
- Show onboarding wizard in popup:
  1. Welcome screen
  2. Location setup
  3. Weather API key
  4. Spotify connect (optional)
  5. Done
- Save config on completion
- **Files**: `OnboardingView.svelte`, `App.svelte` (routing)

---

## Phase 7: Quality & Reliability (Tasks 24–26)

### Task 24: Error Handling & Logging
**Goal**: Robust error handling throughout.
- Implement structured logging with `log/slog`
- Log file at `~/Library/Logs/Vinyl/vinyl.log` with 10MB rotation
- Graceful degradation for all API failures
- Exponential backoff for rate-limited APIs
- User-visible error states in UI (e.g., "Weather unavailable")
- **Files**: Across all Go files

### Task 25: Album Art Cache
**Goal**: Persistent art cache with auto-pruning.
- Cache directory: `~/Library/Caches/Vinyl/art/`
- Download art on new track (URL → sha256-named file)
- Track in `art_cache` SQLite table
- Serve to frontend via Wails static serving or base64
- Prune files not accessed in 90 days (on startup)
- **Files**: New helper in `internal/player/`, updated `store.go`

### Task 26: Testing
**Goal**: Core logic is tested.
- Table-driven tests for:
  - `diary/analytics.go` — daily stats, top artists/tracks queries
  - `player/detector.go` — session state machine transitions
  - `weather/client.go` — caching logic
  - `config/config.go` — load/save/defaults
- Run with `go test ./...`
- **Files**: `*_test.go` files alongside each package

---

## Execution Notes for Claude Code

1. **Read all spec files first**: CLAUDE.md, SPEC.md, DATA_MODEL.md, API_INTEGRATIONS.md, UI_SPEC.md
2. **Execute tasks sequentially**: Each task builds on previous ones
3. **Test after each task**: `wails dev` to verify functionality
4. **Commit after each task**: One commit per task, descriptive message
5. **Frontend hot reload**: Wails dev mode supports Svelte hot reload
6. **macOS only**: Some features (AppleScript, Keychain) are macOS-specific — no cross-platform abstractions needed
7. **Offline-first**: The app must work without internet (just no weather data and no Spotify API enrichment)
