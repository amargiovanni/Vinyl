package diary

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *Store {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		t.Fatal(err)
	}

	store := &Store{db: db, path: dbPath}
	if err := store.Migrate(); err != nil {
		t.Fatal(err)
	}

	return store
}

func TestMigrate(t *testing.T) {
	store := setupTestDB(t)
	defer store.Close()

	// Verify tables exist
	tables := []string{"sessions", "session_tracks", "artists", "daily_stats", "weather_cache", "art_cache"}
	for _, table := range tables {
		var name string
		err := store.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %s not found: %v", table, err)
		}
	}
}

func TestCreateAndEndSession(t *testing.T) {
	store := setupTestDB(t)
	defer store.Close()

	now := time.Now()
	session := &Session{
		ID:        NewULID(),
		StartedAt: now.UTC().Format(time.RFC3339),
		Source:    "spotify",
		DayOfWeek: 0,
		HourOfDay: 14,
		Month:     3,
		Year:      2026,
		DateLocal: "2026-03-08",
		TimeOfDay: "afternoon",
	}

	if err := store.CreateSession(session); err != nil {
		t.Fatal(err)
	}

	// Verify session exists
	sessions, err := store.GetSessionsByDate("2026-03-08")
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].Source != "spotify" {
		t.Errorf("expected source spotify, got %s", sessions[0].Source)
	}

	// End the session
	endedAt := now.Add(2 * time.Hour).UTC().Format(time.RFC3339)
	if err := store.EndSession(session.ID, endedAt, 7200, 15); err != nil {
		t.Fatal(err)
	}

	// Verify ended
	sessions, err = store.GetSessionsByDate("2026-03-08")
	if err != nil {
		t.Fatal(err)
	}
	if sessions[0].DurationSec != 7200 {
		t.Errorf("expected duration 7200, got %d", sessions[0].DurationSec)
	}
	if sessions[0].TrackCount != 15 {
		t.Errorf("expected 15 tracks, got %d", sessions[0].TrackCount)
	}
}

func TestAddTrack(t *testing.T) {
	store := setupTestDB(t)
	defer store.Close()

	session := &Session{
		ID:        NewULID(),
		StartedAt: time.Now().UTC().Format(time.RFC3339),
		Source:    "apple_music",
		DayOfWeek: 0,
		HourOfDay: 10,
		Month:     3,
		Year:      2026,
		DateLocal: "2026-03-08",
		TimeOfDay: "morning",
	}
	store.CreateSession(session)

	track := &SessionTrack{
		ID:                NewULID(),
		SessionID:         session.ID,
		Title:             "Karma Police",
		Artist:            "Radiohead",
		Album:             "OK Computer",
		DurationMS:        261000,
		ListenedMS:        261000,
		Source:            "apple_music",
		PlayedAt:          time.Now().UTC().Format(time.RFC3339),
		PositionInSession: 0,
	}

	if err := store.AddTrack(track); err != nil {
		t.Fatal(err)
	}

	tracks, err := store.GetSessionTracks(session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(tracks))
	}
	if tracks[0].Title != "Karma Police" {
		t.Errorf("expected Karma Police, got %s", tracks[0].Title)
	}
	if tracks[0].Artist != "Radiohead" {
		t.Errorf("expected Radiohead, got %s", tracks[0].Artist)
	}
}

func TestUpdateSessionMood(t *testing.T) {
	store := setupTestDB(t)
	defer store.Close()

	session := &Session{
		ID:        NewULID(),
		StartedAt: time.Now().UTC().Format(time.RFC3339),
		Source:    "spotify",
		DayOfWeek: 0,
		HourOfDay: 14,
		Month:     3,
		Year:      2026,
		DateLocal: "2026-03-08",
		TimeOfDay: "afternoon",
	}
	store.CreateSession(session)

	// Set mood
	if err := store.UpdateSessionMood(session.ID, "happy"); err != nil {
		t.Fatal(err)
	}

	sessions, _ := store.GetSessionsByDate("2026-03-08")
	if sessions[0].Mood != "happy" {
		t.Errorf("expected mood happy, got %s", sessions[0].Mood)
	}

	// Clear mood
	if err := store.UpdateSessionMood(session.ID, ""); err != nil {
		t.Fatal(err)
	}
	sessions, _ = store.GetSessionsByDate("2026-03-08")
	if sessions[0].Mood != "" {
		t.Errorf("expected empty mood, got %s", sessions[0].Mood)
	}
}

func TestHeatmapData(t *testing.T) {
	store := setupTestDB(t)
	defer store.Close()

	// Create a session and end it to trigger daily stats
	session := &Session{
		ID:        NewULID(),
		StartedAt: time.Now().UTC().Format(time.RFC3339),
		Source:    "spotify",
		DayOfWeek: 0,
		HourOfDay: 14,
		Month:     int(time.Now().Month()),
		Year:      time.Now().Year(),
		DateLocal: time.Now().Format("2006-01-02"),
		TimeOfDay: "afternoon",
	}
	store.CreateSession(session)
	store.EndSession(session.ID, time.Now().Add(time.Hour).UTC().Format(time.RFC3339), 3600, 10)

	data, err := store.GetHeatmapData(52)
	if err != nil {
		t.Fatal(err)
	}

	// Should have at least one entry for today
	found := false
	today := time.Now().Format("2006-01-02")
	for _, d := range data {
		if d.DateLocal == today {
			found = true
			if d.TotalListenSec != 3600 {
				t.Errorf("expected 3600 sec, got %d", d.TotalListenSec)
			}
		}
	}
	if !found {
		t.Error("today not found in heatmap data")
	}
}

func TestDeleteSession(t *testing.T) {
	store := setupTestDB(t)
	defer store.Close()

	session := &Session{
		ID:        NewULID(),
		StartedAt: time.Now().UTC().Format(time.RFC3339),
		Source:    "spotify",
		DayOfWeek: 0,
		HourOfDay: 14,
		Month:     3,
		Year:      2026,
		DateLocal: "2026-03-08",
		TimeOfDay: "afternoon",
	}
	store.CreateSession(session)
	store.DeleteSession(session.ID)

	sessions, _ := store.GetSessionsByDate("2026-03-08")
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions after delete, got %d", len(sessions))
	}
}

func TestGetDatabaseInfo(t *testing.T) {
	store := setupTestDB(t)
	defer store.Close()

	info, err := store.GetDatabaseInfo()
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := info["path"]; !ok {
		t.Error("expected path in database info")
	}
	if _, ok := info["session_count"]; !ok {
		t.Error("expected session_count in database info")
	}
}
