package diary

import (
	"database/sql"
	"fmt"
	"log/slog"
)

var migrations = []func(*sql.DB) error{
	migration001,
}

func (s *Store) Migrate() error {
	s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)

	var currentVersion int
	row := s.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version")
	if err := row.Scan(&currentVersion); err != nil {
		currentVersion = 0
	}

	for i := currentVersion; i < len(migrations); i++ {
		slog.Info("applying migration", "version", i+1)
		if err := migrations[i](s.db); err != nil {
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
		if _, err := s.db.Exec("INSERT INTO schema_version (version) VALUES (?)", i+1); err != nil {
			return fmt.Errorf("record migration %d: %w", i+1, err)
		}
	}
	return nil
}

func migration001(db *sql.DB) error {
	statements := []string{
		`PRAGMA journal_mode = WAL`,
		`PRAGMA foreign_keys = ON`,

		`CREATE TABLE IF NOT EXISTS sessions (
			id              TEXT PRIMARY KEY,
			started_at      TEXT NOT NULL,
			ended_at        TEXT,
			duration_sec    INTEGER,
			track_count     INTEGER DEFAULT 0,
			source          TEXT NOT NULL,
			mood            TEXT,
			mood_set_at     TEXT,
			weather_temp    REAL,
			weather_cond    TEXT,
			weather_desc    TEXT,
			weather_humid   INTEGER,
			weather_icon    TEXT,
			day_of_week     INTEGER,
			hour_of_day     INTEGER,
			month           INTEGER,
			year            INTEGER,
			date_local      TEXT,
			time_of_day     TEXT,
			created_at      TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE INDEX IF NOT EXISTS idx_sessions_date ON sessions(date_local)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_year_month ON sessions(year, month)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_mood ON sessions(mood)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_started ON sessions(started_at)`,

		`CREATE TABLE IF NOT EXISTS session_tracks (
			id              TEXT PRIMARY KEY,
			session_id      TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
			title           TEXT NOT NULL,
			artist          TEXT NOT NULL,
			album           TEXT NOT NULL,
			duration_ms     INTEGER,
			listened_ms     INTEGER,
			album_art_url   TEXT,
			album_art_path  TEXT,
			spotify_id      TEXT,
			genres          TEXT,
			energy          REAL,
			valence         REAL,
			tempo           REAL,
			source          TEXT NOT NULL,
			played_at       TEXT NOT NULL,
			position_in_session INTEGER NOT NULL,
			created_at      TEXT NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE INDEX IF NOT EXISTS idx_tracks_session ON session_tracks(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tracks_artist ON session_tracks(artist)`,
		`CREATE INDEX IF NOT EXISTS idx_tracks_played ON session_tracks(played_at)`,
		`CREATE INDEX IF NOT EXISTS idx_tracks_spotify ON session_tracks(spotify_id)`,

		`CREATE TABLE IF NOT EXISTS artists (
			name            TEXT PRIMARY KEY,
			total_listen_ms INTEGER NOT NULL DEFAULT 0,
			session_count   INTEGER NOT NULL DEFAULT 0,
			track_count     INTEGER NOT NULL DEFAULT 0,
			first_heard     TEXT NOT NULL,
			last_heard      TEXT NOT NULL,
			genres          TEXT,
			updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE TABLE IF NOT EXISTS daily_stats (
			date_local      TEXT PRIMARY KEY,
			total_listen_sec INTEGER NOT NULL DEFAULT 0,
			session_count   INTEGER NOT NULL DEFAULT 0,
			track_count     INTEGER NOT NULL DEFAULT 0,
			dominant_mood   TEXT,
			avg_temp        REAL,
			dominant_weather TEXT,
			updated_at      TEXT NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE TABLE IF NOT EXISTS weather_cache (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			lat             REAL NOT NULL,
			lon             REAL NOT NULL,
			temp            REAL NOT NULL,
			condition       TEXT NOT NULL,
			description     TEXT NOT NULL,
			humidity        INTEGER NOT NULL,
			icon            TEXT NOT NULL,
			fetched_at      TEXT NOT NULL,
			expires_at      TEXT NOT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS art_cache (
			url             TEXT PRIMARY KEY,
			local_path      TEXT NOT NULL,
			last_accessed   TEXT NOT NULL DEFAULT (datetime('now')),
			size_bytes      INTEGER
		)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("exec %q: %w", stmt[:50], err)
		}
	}
	return nil
}
