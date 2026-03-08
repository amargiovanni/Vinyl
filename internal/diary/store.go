package diary

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Store manages all database operations.
type Store struct {
	db   *sql.DB
	path string
}

// NewStore opens (or creates) the SQLite database.
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}

	dir := filepath.Join(home, "Library", "Application Support", "Vinyl")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	dbPath := filepath.Join(dir, "vinyl.db")
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=ON&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Store{db: db, path: dbPath}, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// CreateSession inserts a new listening session.
func (s *Store) CreateSession(session *Session) error {
	_, err := s.db.Exec(`INSERT INTO sessions (id, started_at, source, track_count, day_of_week, hour_of_day, month, year, date_local, time_of_day, weather_temp, weather_cond, weather_desc, weather_humid, weather_icon)
		VALUES (?, ?, ?, 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		session.ID, session.StartedAt, session.Source,
		session.DayOfWeek, session.HourOfDay, session.Month, session.Year,
		session.DateLocal, session.TimeOfDay,
		nilIfZero(session.WeatherTemp), nilIfEmpty(session.WeatherCond),
		nilIfEmpty(session.WeatherDesc), nilIfZeroInt(session.WeatherHumid),
		nilIfEmpty(session.WeatherIcon))
	return err
}

// EndSession finalizes a listening session.
func (s *Store) EndSession(id string, endedAt string, durationSec int, trackCount int) error {
	_, err := s.db.Exec(`UPDATE sessions SET ended_at = ?, duration_sec = ?, track_count = ?, updated_at = datetime('now') WHERE id = ?`,
		endedAt, durationSec, trackCount, id)
	if err != nil {
		return err
	}

	// Update daily stats
	var dateLocal string
	s.db.QueryRow("SELECT date_local FROM sessions WHERE id = ?", id).Scan(&dateLocal)
	if dateLocal != "" {
		return s.updateDailyStats(dateLocal)
	}
	return nil
}

// DeleteSession removes a session if it was too short.
func (s *Store) DeleteSession(id string) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE id = ?", id)
	return err
}

// AddTrack inserts a track into a session.
func (s *Store) AddTrack(track *SessionTrack) error {
	_, err := s.db.Exec(`INSERT INTO session_tracks (id, session_id, title, artist, album, duration_ms, listened_ms, album_art_url, album_art_path, spotify_id, genres, energy, valence, tempo, source, played_at, position_in_session)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		track.ID, track.SessionID, track.Title, track.Artist, track.Album,
		track.DurationMS, track.ListenedMS, nilIfEmpty(track.AlbumArtURL),
		nilIfEmpty(track.AlbumArtPath), nilIfEmpty(track.SpotifyID),
		nilIfEmpty(track.Genres), nilIfZeroF(track.Energy), nilIfZeroF(track.Valence),
		nilIfZeroF(track.Tempo), track.Source, track.PlayedAt, track.PositionInSession)
	if err != nil {
		return err
	}

	// Increment track count
	_, err = s.db.Exec("UPDATE sessions SET track_count = track_count + 1, updated_at = datetime('now') WHERE id = ?", track.SessionID)
	return err
}

// UpdateSessionMood sets or updates the mood for a session.
func (s *Store) UpdateSessionMood(sessionID string, mood string) error {
	if mood == "" {
		_, err := s.db.Exec("UPDATE sessions SET mood = NULL, mood_set_at = NULL, updated_at = datetime('now') WHERE id = ?", sessionID)
		return err
	}
	_, err := s.db.Exec("UPDATE sessions SET mood = ?, mood_set_at = datetime('now'), updated_at = datetime('now') WHERE id = ?", mood, sessionID)
	return err
}

// GetHeatmapData returns daily listening stats for the heatmap.
func (s *Store) GetHeatmapData(weeks int) ([]DailyStats, error) {
	days := weeks * 7
	rows, err := s.db.Query(`SELECT date_local, total_listen_sec, session_count, track_count, dominant_mood, avg_temp, dominant_weather
		FROM daily_stats WHERE date_local >= date('now', ? || ' days') ORDER BY date_local ASC`,
		fmt.Sprintf("-%d", days))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []DailyStats
	for rows.Next() {
		var ds DailyStats
		var mood, weather sql.NullString
		var temp sql.NullFloat64
		if err := rows.Scan(&ds.DateLocal, &ds.TotalListenSec, &ds.SessionCount, &ds.TrackCount, &mood, &temp, &weather); err != nil {
			return nil, err
		}
		if mood.Valid {
			ds.DominantMood = mood.String
		}
		if temp.Valid {
			ds.AvgTemp = temp.Float64
		}
		if weather.Valid {
			ds.DominantWeather = weather.String
		}
		stats = append(stats, ds)
	}
	return stats, nil
}

// GetSessionsByMonth returns sessions for a given year/month.
func (s *Store) GetSessionsByMonth(year, month int) ([]Session, error) {
	rows, err := s.db.Query(`SELECT id, started_at, COALESCE(ended_at, ''), COALESCE(duration_sec, 0), track_count, source,
		COALESCE(mood, ''), COALESCE(mood_set_at, ''), COALESCE(weather_temp, 0), COALESCE(weather_cond, ''),
		COALESCE(weather_desc, ''), COALESCE(weather_humid, 0), COALESCE(weather_icon, ''),
		day_of_week, hour_of_day, month, year, date_local, time_of_day
		FROM sessions WHERE year = ? AND month = ? ORDER BY started_at DESC`, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSessions(rows)
}

// GetSessionsByDate returns sessions for a specific date.
func (s *Store) GetSessionsByDate(date string) ([]Session, error) {
	rows, err := s.db.Query(`SELECT id, started_at, COALESCE(ended_at, ''), COALESCE(duration_sec, 0), track_count, source,
		COALESCE(mood, ''), COALESCE(mood_set_at, ''), COALESCE(weather_temp, 0), COALESCE(weather_cond, ''),
		COALESCE(weather_desc, ''), COALESCE(weather_humid, 0), COALESCE(weather_icon, ''),
		day_of_week, hour_of_day, month, year, date_local, time_of_day
		FROM sessions WHERE date_local = ? ORDER BY started_at DESC`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSessions(rows)
}

func scanSessions(rows *sql.Rows) ([]Session, error) {
	var sessions []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.StartedAt, &s.EndedAt, &s.DurationSec, &s.TrackCount, &s.Source,
			&s.Mood, &s.MoodSetAt, &s.WeatherTemp, &s.WeatherCond, &s.WeatherDesc, &s.WeatherHumid, &s.WeatherIcon,
			&s.DayOfWeek, &s.HourOfDay, &s.Month, &s.Year, &s.DateLocal, &s.TimeOfDay); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

// GetSessionTracks returns all tracks for a session.
func (s *Store) GetSessionTracks(sessionID string) ([]SessionTrack, error) {
	rows, err := s.db.Query(`SELECT id, session_id, title, artist, album, COALESCE(duration_ms, 0), COALESCE(listened_ms, 0),
		COALESCE(album_art_url, ''), COALESCE(album_art_path, ''), COALESCE(spotify_id, ''), COALESCE(genres, ''),
		COALESCE(energy, 0), COALESCE(valence, 0), COALESCE(tempo, 0), source, played_at, position_in_session
		FROM session_tracks WHERE session_id = ? ORDER BY position_in_session ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []SessionTrack
	for rows.Next() {
		var t SessionTrack
		if err := rows.Scan(&t.ID, &t.SessionID, &t.Title, &t.Artist, &t.Album, &t.DurationMS, &t.ListenedMS,
			&t.AlbumArtURL, &t.AlbumArtPath, &t.SpotifyID, &t.Genres, &t.Energy, &t.Valence, &t.Tempo,
			&t.Source, &t.PlayedAt, &t.PositionInSession); err != nil {
			return nil, err
		}
		tracks = append(tracks, t)
	}
	return tracks, nil
}

// GetTopArtists returns top artists for a given year/month.
func (s *Store) GetTopArtists(year, month, limit int) ([]ArtistStats, error) {
	rows, err := s.db.Query(`SELECT t.artist, COUNT(DISTINCT t.id) as track_plays,
		SUM(t.listened_ms) / 60000 as total_minutes, COUNT(DISTINCT s.id) as sessions
		FROM session_tracks t JOIN sessions s ON s.id = t.session_id
		WHERE s.year = ? AND s.month = ?
		GROUP BY t.artist ORDER BY total_minutes DESC LIMIT ?`, year, month, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artists []ArtistStats
	for rows.Next() {
		var a ArtistStats
		if err := rows.Scan(&a.Name, &a.TrackPlays, &a.TotalMinutes, &a.Sessions); err != nil {
			return nil, err
		}
		artists = append(artists, a)
	}
	return artists, nil
}

// GetTopTracks returns top tracks for a given year/month.
func (s *Store) GetTopTracks(year, month, limit int) ([]TrackStats, error) {
	rows, err := s.db.Query(`SELECT t.title, t.artist, t.album, COUNT(*) as play_count,
		SUM(t.listened_ms) / 60000 as total_minutes
		FROM session_tracks t JOIN sessions s ON s.id = t.session_id
		WHERE s.year = ? AND s.month = ?
		GROUP BY t.title, t.artist ORDER BY play_count DESC LIMIT ?`, year, month, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []TrackStats
	for rows.Next() {
		var t TrackStats
		if err := rows.Scan(&t.Title, &t.Artist, &t.Album, &t.PlayCount, &t.TotalMinutes); err != nil {
			return nil, err
		}
		tracks = append(tracks, t)
	}
	return tracks, nil
}

// GetMoodDistribution returns mood stats for a time range.
func (s *Store) GetMoodDistribution(startDate, endDate string) ([]MoodStat, error) {
	rows, err := s.db.Query(`SELECT mood, COUNT(*) as count,
		ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 1) as percentage
		FROM sessions WHERE mood IS NOT NULL AND mood != '' AND date_local BETWEEN ? AND ?
		GROUP BY mood ORDER BY count DESC`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var moods []MoodStat
	for rows.Next() {
		var m MoodStat
		if err := rows.Scan(&m.Mood, &m.Count, &m.Percentage); err != nil {
			return nil, err
		}
		moods = append(moods, m)
	}
	return moods, nil
}

// GetWeatherCorrelation returns listening by weather condition.
func (s *Store) GetWeatherCorrelation(year int) ([]WeatherStat, error) {
	rows, err := s.db.Query(`SELECT weather_cond, COUNT(*) as session_count,
		SUM(duration_sec) / 3600.0 as total_hours
		FROM sessions WHERE weather_cond IS NOT NULL AND weather_cond != '' AND year = ?
		GROUP BY weather_cond ORDER BY total_hours DESC`, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []WeatherStat
	for rows.Next() {
		var w WeatherStat
		if err := rows.Scan(&w.Condition, &w.SessionCount, &w.TotalHours); err != nil {
			return nil, err
		}
		stats = append(stats, w)
	}
	return stats, nil
}

// GetListeningByHour returns listening hours grouped by hour.
func (s *Store) GetListeningByHour(year int) ([]HourStat, error) {
	rows, err := s.db.Query(`SELECT hour_of_day, SUM(duration_sec) / 3600.0 as total_hours, COUNT(*) as sessions
		FROM sessions WHERE year = ? GROUP BY hour_of_day ORDER BY hour_of_day`, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []HourStat
	for rows.Next() {
		var h HourStat
		if err := rows.Scan(&h.Hour, &h.TotalHours, &h.Sessions); err != nil {
			return nil, err
		}
		stats = append(stats, h)
	}
	return stats, nil
}

// GetListeningByDayOfWeek returns listening hours grouped by day of week.
func (s *Store) GetListeningByDayOfWeek(year int) ([]DayOfWeekStat, error) {
	rows, err := s.db.Query(`SELECT day_of_week, SUM(duration_sec) / 3600.0 as total_hours, COUNT(*) as sessions
		FROM sessions WHERE year = ? GROUP BY day_of_week ORDER BY day_of_week`, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []DayOfWeekStat
	for rows.Next() {
		var d DayOfWeekStat
		if err := rows.Scan(&d.Day, &d.TotalHours, &d.Sessions); err != nil {
			return nil, err
		}
		stats = append(stats, d)
	}
	return stats, nil
}

// GetDatabaseInfo returns database path and statistics.
func (s *Store) GetDatabaseInfo() (map[string]interface{}, error) {
	info := map[string]interface{}{
		"path": s.path,
	}

	fi, err := os.Stat(s.path)
	if err == nil {
		info["size_bytes"] = fi.Size()
		info["size_mb"] = fmt.Sprintf("%.1f", float64(fi.Size())/(1024*1024))
	}

	var sessionCount int
	s.db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessionCount)
	info["session_count"] = sessionCount

	var trackCount int
	s.db.QueryRow("SELECT COUNT(*) FROM session_tracks").Scan(&trackCount)
	info["track_count"] = trackCount

	return info, nil
}

// GetYearsWithData returns distinct years that have sessions.
func (s *Store) GetYearsWithData() ([]int, error) {
	rows, err := s.db.Query("SELECT DISTINCT year FROM sessions ORDER BY year DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var years []int
	for rows.Next() {
		var y int
		if err := rows.Scan(&y); err != nil {
			return nil, err
		}
		years = append(years, y)
	}
	return years, nil
}

// GetYearStats returns aggregate stats for a year.
func (s *Store) GetYearStats(year int) (map[string]interface{}, error) {
	stats := map[string]interface{}{"year": year}

	var totalSec sql.NullInt64
	var sessionCount int
	s.db.QueryRow("SELECT COALESCE(SUM(duration_sec), 0), COUNT(*) FROM sessions WHERE year = ?", year).Scan(&totalSec, &sessionCount)
	stats["total_hours"] = fmt.Sprintf("%.0f", float64(totalSec.Int64)/3600)
	stats["session_count"] = sessionCount

	var artistCount int
	s.db.QueryRow("SELECT COUNT(DISTINCT artist) FROM session_tracks t JOIN sessions s ON s.id = t.session_id WHERE s.year = ?", year).Scan(&artistCount)
	stats["artist_count"] = artistCount

	var trackCount int
	s.db.QueryRow("SELECT COUNT(DISTINCT title || artist) FROM session_tracks t JOIN sessions s ON s.id = t.session_id WHERE s.year = ?", year).Scan(&trackCount)
	stats["track_count"] = trackCount

	return stats, nil
}

// GetInsights generates pattern-based insights.
func (s *Store) GetInsights(year int) ([]Insight, error) {
	var insights []Insight

	// Weather-genre correlation
	rows, err := s.db.Query(`SELECT s.weather_cond, t.genres, COUNT(*) as cnt
		FROM sessions s JOIN session_tracks t ON s.id = t.session_id
		WHERE s.year = ? AND s.weather_cond IS NOT NULL AND t.genres IS NOT NULL AND t.genres != ''
		GROUP BY s.weather_cond, t.genres ORDER BY cnt DESC LIMIT 5`, year)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cond, genres string
			var cnt int
			if rows.Scan(&cond, &genres, &cnt) == nil && cnt >= 3 {
				var genreList []string
				json.Unmarshal([]byte(genres), &genreList)
				if len(genreList) > 0 {
					insights = append(insights, Insight{
						Icon:        weatherEmoji(cond),
						Title:       fmt.Sprintf("You listen to %s when it's %s", genreList[0], strings.ToLower(cond)),
						Description: fmt.Sprintf("Found in %d listening sessions this year", cnt),
						Stat:        fmt.Sprintf("%d sessions", cnt),
					})
				}
			}
		}
	}

	// Most active day
	var bestDay string
	var bestDayHours float64
	s.db.QueryRow(`SELECT date_local, SUM(duration_sec)/3600.0 FROM sessions WHERE year = ? GROUP BY date_local ORDER BY SUM(duration_sec) DESC LIMIT 1`, year).Scan(&bestDay, &bestDayHours)
	if bestDay != "" {
		insights = append(insights, Insight{
			Icon:        "\U0001f4c5",
			Title:       "Your biggest listening day",
			Description: fmt.Sprintf("%s — %.1f hours of music", bestDay, bestDayHours),
			Stat:        fmt.Sprintf("%.1fh", bestDayHours),
		})
	}

	// Late night listening
	var lateNightCount int
	s.db.QueryRow("SELECT COUNT(*) FROM sessions WHERE year = ? AND hour_of_day >= 23 OR hour_of_day < 5", year).Scan(&lateNightCount)
	if lateNightCount > 0 {
		insights = append(insights, Insight{
			Icon:        "\U0001f319",
			Title:       "Night owl sessions",
			Description: fmt.Sprintf("You had %d late-night listening sessions", lateNightCount),
			Stat:        fmt.Sprintf("%d sessions", lateNightCount),
		})
	}

	// Most listened artist
	var topArtist string
	var topArtistMin int
	s.db.QueryRow(`SELECT artist, SUM(listened_ms)/60000 FROM session_tracks t JOIN sessions s ON s.id = t.session_id WHERE s.year = ? GROUP BY artist ORDER BY SUM(listened_ms) DESC LIMIT 1`, year).Scan(&topArtist, &topArtistMin)
	if topArtist != "" {
		insights = append(insights, Insight{
			Icon:        "\U0001f3a4",
			Title:       fmt.Sprintf("Your #1 artist: %s", topArtist),
			Description: fmt.Sprintf("%d minutes of listening this year", topArtistMin),
			Stat:        fmt.Sprintf("%dm", topArtistMin),
		})
	}

	return insights, nil
}

// updateDailyStats re-aggregates stats for a given date.
func (s *Store) updateDailyStats(dateLocal string) error {
	_, err := s.db.Exec(`INSERT OR REPLACE INTO daily_stats (date_local, total_listen_sec, session_count, track_count, dominant_mood, avg_temp, dominant_weather, updated_at)
		SELECT date_local,
			COALESCE(SUM(duration_sec), 0),
			COUNT(*),
			COALESCE(SUM(track_count), 0),
			(SELECT mood FROM sessions WHERE date_local = ? AND mood IS NOT NULL GROUP BY mood ORDER BY COUNT(*) DESC LIMIT 1),
			AVG(weather_temp),
			(SELECT weather_cond FROM sessions WHERE date_local = ? AND weather_cond IS NOT NULL GROUP BY weather_cond ORDER BY COUNT(*) DESC LIMIT 1),
			datetime('now')
		FROM sessions WHERE date_local = ?`, dateLocal, dateLocal, dateLocal)
	return err
}

// UpdateArtistStats updates aggregate artist data after a session ends.
func (s *Store) UpdateArtistStats(sessionID string) error {
	rows, err := s.db.Query(`SELECT artist, SUM(listened_ms), COUNT(*), genres FROM session_tracks WHERE session_id = ? GROUP BY artist`, sessionID)
	if err != nil {
		return err
	}
	defer rows.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	for rows.Next() {
		var name string
		var listenedMS int
		var trackCount int
		var genres sql.NullString
		if err := rows.Scan(&name, &listenedMS, &trackCount, &genres); err != nil {
			continue
		}
		_, err := s.db.Exec(`INSERT INTO artists (name, total_listen_ms, session_count, track_count, first_heard, last_heard, genres)
			VALUES (?, ?, 1, ?, ?, ?, ?)
			ON CONFLICT(name) DO UPDATE SET
				total_listen_ms = total_listen_ms + ?,
				session_count = session_count + 1,
				track_count = track_count + ?,
				last_heard = ?,
				genres = COALESCE(excluded.genres, genres)`,
			name, listenedMS, trackCount, now, now, genres,
			listenedMS, trackCount, now)
		if err != nil {
			slog.Warn("update artist stats", "artist", name, "err", err)
		}
	}
	return nil
}

// ExportAnnualBook generates an HTML annual report.
func (s *Store) ExportAnnualBook(year int, outputPath string) error {
	type MonthData struct {
		Month        int
		MonthName    string
		TotalHours   float64
		Sessions     int
		DominantMood string
		TopArtists   []ArtistStats
		TopTracks    []TrackStats
		MoodDist     []MoodStat
		WeatherCorr  []WeatherStat
	}

	type BookData struct {
		Year          int
		TotalHours    string
		TotalSessions int
		TotalArtists  int
		TotalTracks   int
		TopArtists    []ArtistStats
		Months        []MonthData
		HourStats     []HourStat
		DayStats      []DayOfWeekStat
	}

	yearStats, err := s.GetYearStats(year)
	if err != nil {
		return fmt.Errorf("get year stats: %w", err)
	}

	book := BookData{
		Year:          year,
		TotalHours:    yearStats["total_hours"].(string),
		TotalSessions: yearStats["session_count"].(int),
		TotalArtists:  yearStats["artist_count"].(int),
		TotalTracks:   yearStats["track_count"].(int),
	}

	// Top artists for the year
	for m := 1; m <= 12; m++ {
		artists, _ := s.GetTopArtists(year, m, 10)
		if m == 0 { // dummy to avoid unused
			_ = artists
		}
	}

	// Gather top 3 artists for the whole year
	rows, err := s.db.Query(`SELECT artist, SUM(listened_ms)/60000 as mins, COUNT(*) as plays
		FROM session_tracks t JOIN sessions s ON s.id = t.session_id
		WHERE s.year = ? GROUP BY artist ORDER BY mins DESC LIMIT 3`, year)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var a ArtistStats
			rows.Scan(&a.Name, &a.TotalMinutes, &a.TrackPlays)
			book.TopArtists = append(book.TopArtists, a)
		}
	}

	monthNames := []string{"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}

	for m := 1; m <= 12; m++ {
		md := MonthData{Month: m, MonthName: monthNames[m]}

		var totalSec sql.NullInt64
		var sessions int
		s.db.QueryRow("SELECT COALESCE(SUM(duration_sec),0), COUNT(*) FROM sessions WHERE year = ? AND month = ?", year, m).Scan(&totalSec, &sessions)
		md.TotalHours = float64(totalSec.Int64) / 3600
		md.Sessions = sessions

		if sessions == 0 {
			continue
		}

		md.TopArtists, _ = s.GetTopArtists(year, m, 10)
		md.TopTracks, _ = s.GetTopTracks(year, m, 10)

		startDate := fmt.Sprintf("%d-%02d-01", year, m)
		endDate := fmt.Sprintf("%d-%02d-31", year, m)
		md.MoodDist, _ = s.GetMoodDistribution(startDate, endDate)

		var mood sql.NullString
		s.db.QueryRow("SELECT mood FROM sessions WHERE year = ? AND month = ? AND mood IS NOT NULL GROUP BY mood ORDER BY COUNT(*) DESC LIMIT 1", year, m).Scan(&mood)
		if mood.Valid {
			md.DominantMood = mood.String
		}

		book.Months = append(book.Months, md)
	}

	book.HourStats, _ = s.GetListeningByHour(year)
	book.DayStats, _ = s.GetListeningByDayOfWeek(year)

	tmpl := template.Must(template.New("book").Funcs(template.FuncMap{
		"moodEmoji": moodEmoji,
		"dayName":   dayName,
		"formatHours": func(h float64) string {
			return fmt.Sprintf("%.1f", h)
		},
		"pct": func(v, max float64) float64 {
			if max == 0 {
				return 0
			}
			return (v / max) * 100
		},
	}).Parse(bookTemplate))

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	return tmpl.Execute(f, book)
}

func weatherEmoji(cond string) string {
	switch cond {
	case "Clear":
		return "\u2600\ufe0f"
	case "Clouds":
		return "\u2601\ufe0f"
	case "Rain", "Drizzle":
		return "\U0001f327\ufe0f"
	case "Thunderstorm":
		return "\u26c8\ufe0f"
	case "Snow":
		return "\u2744\ufe0f"
	case "Mist", "Fog":
		return "\U0001f32b\ufe0f"
	default:
		return "\U0001f324\ufe0f"
	}
}

func moodEmoji(mood string) string {
	switch mood {
	case "happy":
		return "\U0001f60a"
	case "calm":
		return "\U0001f60c"
	case "energetic":
		return "\U0001f525"
	case "sad":
		return "\U0001f622"
	case "thoughtful":
		return "\U0001f914"
	case "frustrated":
		return "\U0001f624"
	case "in_love":
		return "\U0001f970"
	case "dreamy":
		return "\U0001f319"
	case "motivated":
		return "\U0001f4aa"
	case "celebrating":
		return "\U0001f389"
	case "sleepy":
		return "\U0001f634"
	case "nostalgic":
		return "\U0001f30a"
	default:
		return ""
	}
}

func dayName(d int) string {
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	if d >= 0 && d < len(days) {
		return days[d]
	}
	return ""
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nilIfZero(f float64) interface{} {
	if f == 0 {
		return nil
	}
	return f
}

func nilIfZeroF(f float64) interface{} {
	if f == 0 {
		return nil
	}
	return f
}

func nilIfZeroInt(i int) interface{} {
	if i == 0 {
		return nil
	}
	return i
}

const bookTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Vinyl — {{.Year}}</title>
<link href="https://fonts.googleapis.com/css2?family=Playfair+Display:ital,wght@0,400;0,700;1,400&family=Lora:ital,wght@0,400;0,600;1,400&family=JetBrains+Mono:wght@400&display=swap" rel="stylesheet">
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: 'Lora', serif; background: #FAF3E8; color: #2A2018; line-height: 1.6; }
.page { page-break-before: always; padding: 2cm; min-height: 100vh; }
.page:first-child { page-break-before: auto; }
h1, h2, h3 { font-family: 'Playfair Display', serif; }
.cover { display: flex; flex-direction: column; align-items: center; justify-content: center; text-align: center; }
.cover h1 { font-size: 48pt; letter-spacing: 0.3em; color: #2A2018; margin-bottom: 0.2em; }
.cover .year { font-size: 72pt; color: #C49A6C; font-family: 'Playfair Display', serif; }
.cover .stats { margin-top: 2em; font-size: 14pt; color: #5C4A3A; }
.cover .top-artists { margin-top: 2em; text-align: left; display: inline-block; }
.cover .top-artists li { font-size: 16pt; margin: 0.3em 0; }
.month-header { font-size: 28pt; color: #C49A6C; border-bottom: 2px solid #C49A6C; padding-bottom: 0.3em; margin-bottom: 1em; }
.month-stats { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1em; margin-bottom: 1.5em; }
.stat-box { background: #F0E6D2; padding: 1em; border-radius: 8px; text-align: center; }
.stat-box .value { font-size: 24pt; font-family: 'Playfair Display', serif; color: #C49A6C; }
.stat-box .label { font-size: 10pt; color: #7A6450; text-transform: uppercase; letter-spacing: 0.1em; }
.top-list { margin-bottom: 1.5em; }
.top-list h3 { font-size: 14pt; margin-bottom: 0.5em; color: #5C4A3A; }
.top-list ol { padding-left: 1.5em; }
.top-list li { margin: 0.3em 0; font-size: 11pt; }
.top-list .minutes { font-family: 'JetBrains Mono', monospace; font-size: 9pt; color: #8E7E6E; }
.mood-bar { display: flex; align-items: center; margin: 0.3em 0; gap: 0.5em; }
.mood-bar .bar { height: 16px; background: #C49A6C; border-radius: 3px; }
.mood-bar .label { font-size: 10pt; min-width: 100px; }
.mood-bar .pct { font-family: 'JetBrains Mono', monospace; font-size: 9pt; color: #8E7E6E; }
.section { margin-bottom: 2em; }
.section h3 { font-size: 14pt; margin-bottom: 0.5em; color: #5C4A3A; border-bottom: 1px solid #D4C4A8; padding-bottom: 0.3em; }
.hour-bar { display: flex; align-items: center; gap: 0.5em; margin: 0.15em 0; }
.hour-bar .time { font-family: 'JetBrains Mono', monospace; font-size: 9pt; width: 3em; text-align: right; }
.hour-bar .bar { height: 14px; background: #C49A6C; border-radius: 2px; min-width: 2px; }
.hour-bar .val { font-family: 'JetBrains Mono', monospace; font-size: 9pt; color: #8E7E6E; }
.footer { text-align: center; margin-top: 3em; font-size: 10pt; color: #8E7E6E; font-style: italic; }
@media print {
  body { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
  .page { padding: 2cm; page-break-before: always; }
}
</style>
</head>
<body>
<div class="page cover">
  <h1>V I N Y L</h1>
  <div class="year">{{.Year}}</div>
  <div class="stats">
    {{.TotalHours}} hours · {{.TotalSessions}} sessions<br>
    {{.TotalArtists}} artists · {{.TotalTracks}} tracks
  </div>
  {{if .TopArtists}}
  <div class="top-artists">
    <h3 style="font-size:14pt; margin-bottom:0.5em;">Top Artists</h3>
    <ol>
    {{range .TopArtists}}<li>{{.Name}} <span class="minutes">{{.TotalMinutes}}m</span></li>{{end}}
    </ol>
  </div>
  {{end}}
</div>

{{range .Months}}
<div class="page">
  <h2 class="month-header">{{.MonthName}}</h2>
  <div class="month-stats">
    <div class="stat-box"><div class="value">{{formatHours .TotalHours}}h</div><div class="label">Hours</div></div>
    <div class="stat-box"><div class="value">{{.Sessions}}</div><div class="label">Sessions</div></div>
    <div class="stat-box"><div class="value">{{if .DominantMood}}{{moodEmoji .DominantMood}}{{else}}—{{end}}</div><div class="label">Mood</div></div>
  </div>

  {{if .TopArtists}}
  <div class="section top-list">
    <h3>Top Artists</h3>
    <ol>{{range .TopArtists}}<li>{{.Name}} <span class="minutes">{{.TotalMinutes}} min · {{.TrackPlays}} plays</span></li>{{end}}</ol>
  </div>
  {{end}}

  {{if .TopTracks}}
  <div class="section top-list">
    <h3>Top Tracks</h3>
    <ol>{{range .TopTracks}}<li>{{.Title}} — {{.Artist}} <span class="minutes">{{.PlayCount}}× · {{.TotalMinutes}} min</span></li>{{end}}</ol>
  </div>
  {{end}}

  {{if .MoodDist}}
  <div class="section">
    <h3>Mood Distribution</h3>
    {{range .MoodDist}}
    <div class="mood-bar">
      <span class="label">{{moodEmoji .Mood}} {{.Mood}}</span>
      <div class="bar" style="width: {{.Percentage}}%;"></div>
      <span class="pct">{{.Percentage}}%</span>
    </div>
    {{end}}
  </div>
  {{end}}
</div>
{{end}}

<div class="page">
  <h2 class="month-header">Year in Review</h2>

  {{if .HourStats}}
  <div class="section">
    <h3>Listening by Hour</h3>
    {{range .HourStats}}
    <div class="hour-bar">
      <span class="time">{{printf "%02d" .Hour}}:00</span>
      <div class="bar" style="width: {{formatHours .TotalHours}}%;"></div>
      <span class="val">{{formatHours .TotalHours}}h</span>
    </div>
    {{end}}
  </div>
  {{end}}

  {{if .DayStats}}
  <div class="section">
    <h3>Listening by Day</h3>
    {{range .DayStats}}
    <div class="hour-bar">
      <span class="time" style="width:6em;">{{dayName .Day}}</span>
      <div class="bar" style="width: {{formatHours .TotalHours}}%;"></div>
      <span class="val">{{formatHours .TotalHours}}h</span>
    </div>
    {{end}}
  </div>
  {{end}}

  <div class="footer">Generated by Vinyl · Every song is a memory</div>
</div>
</body>
</html>`
