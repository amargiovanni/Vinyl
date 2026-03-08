# DATA_MODEL.md — Vinyl Database Schema

## Overview

SQLite database stored at `~/Library/Application Support/Vinyl/vinyl.db`.
Use WAL journal mode for concurrent read/write. Foreign keys enforced.

## Schema

### Migrations Strategy

Use a `schema_version` table. On startup, `diary.Store.Migrate()` checks current version and applies pending migrations sequentially. Each migration is a Go function, not a SQL file.

```sql
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);
```

### Migration 001 — Initial Schema

```sql
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

-- ============================================================
-- SESSIONS: A continuous listening session
-- ============================================================
CREATE TABLE sessions (
    id              TEXT PRIMARY KEY,           -- ULID (sortable, unique)
    started_at      TEXT NOT NULL,              -- ISO 8601 UTC
    ended_at        TEXT,                       -- ISO 8601 UTC, NULL if active
    duration_sec    INTEGER,                    -- Computed on session end
    track_count     INTEGER DEFAULT 0,
    source          TEXT NOT NULL,              -- 'spotify' | 'apple_music' | 'mixed'
    mood            TEXT,                       -- Enum: 'happy','calm','energetic','sad','thoughtful','frustrated','in_love','dreamy','motivated','celebrating','sleepy','nostalgic'
    mood_set_at     TEXT,                       -- When mood was tagged
    
    -- Weather snapshot at session start
    weather_temp    REAL,                       -- Celsius
    weather_cond    TEXT,                       -- 'Clear','Clouds','Rain','Drizzle','Snow','Thunderstorm','Mist','Fog'
    weather_desc    TEXT,                       -- 'light rain', 'scattered clouds', etc.
    weather_humid   INTEGER,                    -- Percentage
    weather_icon    TEXT,                       -- OpenWeatherMap icon code
    
    -- Time context (denormalized for fast queries)
    day_of_week     INTEGER,                   -- 0=Monday, 6=Sunday
    hour_of_day     INTEGER,                   -- 0-23, start hour
    month           INTEGER,                   -- 1-12
    year            INTEGER,
    date_local      TEXT,                      -- 'YYYY-MM-DD' in local timezone
    time_of_day     TEXT,                      -- 'dawn','morning','afternoon','evening','night' (derived from hour)
    
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_sessions_date ON sessions(date_local);
CREATE INDEX idx_sessions_year_month ON sessions(year, month);
CREATE INDEX idx_sessions_mood ON sessions(mood);
CREATE INDEX idx_sessions_started ON sessions(started_at);

-- ============================================================
-- TRACKS: Individual tracks within a session
-- ============================================================
CREATE TABLE session_tracks (
    id              TEXT PRIMARY KEY,           -- ULID
    session_id      TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    
    title           TEXT NOT NULL,
    artist          TEXT NOT NULL,
    album           TEXT NOT NULL,
    
    duration_ms     INTEGER,                   -- Track total duration
    listened_ms     INTEGER,                   -- How long user actually listened
    
    album_art_url   TEXT,                      -- URL for cover art
    album_art_path  TEXT,                      -- Local cached path
    
    -- Spotify-specific metadata (NULL for Apple Music)
    spotify_id      TEXT,
    genres          TEXT,                       -- JSON array: ["jazz","soul"]
    energy          REAL,                      -- Spotify audio feature 0.0-1.0
    valence         REAL,                      -- Spotify audio feature 0.0-1.0 (happiness)
    tempo           REAL,                      -- BPM
    
    source          TEXT NOT NULL,              -- 'spotify' | 'apple_music'
    played_at       TEXT NOT NULL,              -- ISO 8601 UTC, when this track started
    position_in_session INTEGER NOT NULL,       -- 0-indexed order in session
    
    created_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_tracks_session ON session_tracks(session_id);
CREATE INDEX idx_tracks_artist ON session_tracks(artist);
CREATE INDEX idx_tracks_played ON session_tracks(played_at);
CREATE INDEX idx_tracks_spotify ON session_tracks(spotify_id);

-- ============================================================
-- ARTISTS: Denormalized artist stats (updated on session end)
-- ============================================================
CREATE TABLE artists (
    name            TEXT PRIMARY KEY,
    total_listen_ms INTEGER NOT NULL DEFAULT 0,
    session_count   INTEGER NOT NULL DEFAULT 0,
    track_count     INTEGER NOT NULL DEFAULT 0,
    first_heard     TEXT NOT NULL,              -- ISO 8601
    last_heard      TEXT NOT NULL,              -- ISO 8601
    genres          TEXT,                       -- JSON array, aggregated
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ============================================================
-- DAILY_STATS: Pre-aggregated daily statistics for heatmap
-- ============================================================
CREATE TABLE daily_stats (
    date_local      TEXT PRIMARY KEY,           -- 'YYYY-MM-DD'
    total_listen_sec INTEGER NOT NULL DEFAULT 0,
    session_count   INTEGER NOT NULL DEFAULT 0,
    track_count     INTEGER NOT NULL DEFAULT 0,
    dominant_mood   TEXT,                       -- Most frequent mood of the day
    avg_temp        REAL,
    dominant_weather TEXT,
    updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ============================================================
-- WEATHER CACHE: Avoid redundant API calls
-- ============================================================
CREATE TABLE weather_cache (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    lat             REAL NOT NULL,
    lon             REAL NOT NULL,
    temp            REAL NOT NULL,
    condition       TEXT NOT NULL,
    description     TEXT NOT NULL,
    humidity        INTEGER NOT NULL,
    icon            TEXT NOT NULL,
    fetched_at      TEXT NOT NULL,
    expires_at      TEXT NOT NULL               -- fetched_at + 30 minutes
);

-- ============================================================
-- ART CACHE: Track cached album art
-- ============================================================
CREATE TABLE art_cache (
    url             TEXT PRIMARY KEY,
    local_path      TEXT NOT NULL,
    last_accessed   TEXT NOT NULL DEFAULT (datetime('now')),
    size_bytes      INTEGER
);
```

## Key Queries

### Heatmap data (last 52 weeks)

```sql
SELECT date_local, total_listen_sec
FROM daily_stats
WHERE date_local >= date('now', '-364 days')
ORDER BY date_local ASC;
```

### Sessions for a specific day

```sql
SELECT s.*, COUNT(t.id) as actual_tracks
FROM sessions s
LEFT JOIN session_tracks t ON t.session_id = s.id
WHERE s.date_local = ?
GROUP BY s.id
ORDER BY s.started_at DESC;
```

### Top artists for a month

```sql
SELECT t.artist, 
       COUNT(DISTINCT t.id) as track_plays,
       SUM(t.listened_ms) / 60000 as total_minutes,
       COUNT(DISTINCT s.id) as sessions
FROM session_tracks t
JOIN sessions s ON s.id = t.session_id
WHERE s.year = ? AND s.month = ?
GROUP BY t.artist
ORDER BY total_minutes DESC
LIMIT 10;
```

### Top tracks for a month

```sql
SELECT t.title, t.artist, t.album,
       COUNT(*) as play_count,
       SUM(t.listened_ms) / 60000 as total_minutes
FROM session_tracks t
JOIN sessions s ON s.id = t.session_id
WHERE s.year = ? AND s.month = ?
GROUP BY t.title, t.artist
ORDER BY play_count DESC
LIMIT 10;
```

### Mood distribution for a time range

```sql
SELECT mood, COUNT(*) as count,
       ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 1) as percentage
FROM sessions
WHERE mood IS NOT NULL
  AND started_at BETWEEN ? AND ?
GROUP BY mood
ORDER BY count DESC;
```

### Weather correlation

```sql
SELECT weather_cond,
       COUNT(*) as session_count,
       SUM(duration_sec) / 3600.0 as total_hours
FROM sessions
WHERE weather_cond IS NOT NULL
  AND year = ?
GROUP BY weather_cond
ORDER BY total_hours DESC;
```

### Listening by hour of day

```sql
SELECT hour_of_day,
       SUM(duration_sec) / 3600.0 as total_hours,
       COUNT(*) as sessions
FROM sessions
WHERE year = ?
GROUP BY hour_of_day
ORDER BY hour_of_day;
```

### Listening by day of week

```sql
SELECT day_of_week,
       SUM(duration_sec) / 3600.0 as total_hours,
       COUNT(*) as sessions
FROM sessions
WHERE year = ?
GROUP BY day_of_week
ORDER BY day_of_week;
```

## ULID Generation

Use `github.com/oklog/ulid/v2` for sortable, unique IDs:

```go
import (
    "math/rand"
    "time"
    "github.com/oklog/ulid/v2"
)

func NewULID() string {
    t := time.Now()
    entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
    return ulid.MustNew(ulid.Timestamp(t), entropy).String()
}
```

## Time of Day Derivation

```go
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
```

## Data Lifecycle

- **Session creation**: INSERT into `sessions` with `ended_at = NULL`
- **Track detected**: INSERT into `session_tracks`, UPDATE `sessions.track_count`
- **Session end**: UPDATE `sessions` with `ended_at`, `duration_sec`. UPDATE/INSERT `daily_stats`. UPDATE `artists` stats.
- **Album art**: Download to `~/Library/Caches/Vinyl/art/{sha256_of_url}.jpg`. INSERT into `art_cache`.
- **Art pruning**: On startup, DELETE from `art_cache` WHERE `last_accessed < date('now', '-90 days')`. Remove corresponding files.
- **Weather cache**: DELETE expired entries on each weather fetch.
