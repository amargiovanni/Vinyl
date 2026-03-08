package player

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"vinyl/internal/config"
	"vinyl/internal/diary"
	"vinyl/internal/weather"
)

// Detector orchestrates music detection from multiple players.
type Detector struct {
	mu             sync.RWMutex
	store          *diary.Store
	weather        *weather.Client
	cfg            *config.Config
	emit           func(string, interface{})
	spotify        *SpotifyClient
	state          PlayerState
	currentTrack   *TrackInfo
	currentSession *diary.Session
	trackCount     int
	lastTrackTime  time.Time
	gapTimer       *time.Timer
	stopCh         chan struct{}
}

func NewDetector(store *diary.Store, weather *weather.Client, cfg *config.Config, emit func(string, interface{})) *Detector {
	d := &Detector{
		store:   store,
		weather: weather,
		cfg:     cfg,
		emit:    emit,
		state:   StateIdle,
		stopCh:  make(chan struct{}),
	}

	if cfg.Spotify.ClientID != "" {
		d.spotify = NewSpotifyClient(cfg.Spotify.ClientID)
	}

	return d
}

// Start begins the polling loop.
func (d *Detector) Start(ctx context.Context) {
	slog.Info("player detector started")

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		default:
		}

		track := d.detect()
		d.processTrack(track)

		interval := time.Duration(d.cfg.PollingIntervalIdle) * time.Millisecond
		if d.state == StateActive {
			interval = time.Duration(d.cfg.PollingIntervalPlay) * time.Millisecond
		}

		select {
		case <-time.After(interval):
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		}
	}
}

// Stop stops the detector.
func (d *Detector) Stop() {
	close(d.stopCh)
	d.finalizeSession()
}

// GetCurrentTrack returns the currently detected track.
func (d *Detector) GetCurrentTrack() *TrackInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.currentTrack
}

// GetCurrentSession returns the active session.
func (d *Detector) GetCurrentSession() *diary.Session {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.currentSession
}

// ConnectSpotify starts the Spotify OAuth flow.
func (d *Detector) ConnectSpotify(ctx context.Context, clientID string) error {
	d.mu.Lock()
	if d.spotify == nil {
		d.spotify = NewSpotifyClient(clientID)
	}
	d.mu.Unlock()
	return d.spotify.Connect(ctx, clientID)
}

// DisconnectSpotify removes the Spotify connection.
func (d *Detector) DisconnectSpotify() error {
	d.mu.RLock()
	s := d.spotify
	d.mu.RUnlock()
	if s == nil {
		return nil
	}
	return s.Disconnect()
}

// IsSpotifyConnected returns whether Spotify is authenticated.
func (d *Detector) IsSpotifyConnected() bool {
	d.mu.RLock()
	s := d.spotify
	d.mu.RUnlock()
	if s == nil {
		return false
	}
	return s.IsConnected()
}

func (d *Detector) detect() *TrackInfo {
	// Priority: Spotify API > Spotify AppleScript > Apple Music
	d.mu.RLock()
	sc := d.spotify
	d.mu.RUnlock()

	if sc != nil && sc.IsConnected() {
		track, err := sc.GetCurrentTrack()
		if err != nil {
			slog.Debug("spotify API failed, trying applescript", "err", err)
		}
		if track != nil {
			return track
		}
	}

	// Spotify AppleScript fallback
	track, err := GetSpotifyViaAppleScript()
	if err != nil {
		slog.Debug("spotify applescript failed", "err", err)
	}
	if track != nil {
		return track
	}

	// Apple Music
	track, err = GetAppleMusicTrack()
	if err != nil {
		slog.Debug("apple music failed", "err", err)
	}
	return track
}

func (d *Detector) processTrack(track *TrackInfo) {
	d.mu.Lock()
	defer d.mu.Unlock()

	switch d.state {
	case StateIdle:
		if track != nil {
			d.startSession(track)
		}

	case StateActive:
		if track == nil {
			d.state = StateEnding
			d.gapTimer = time.AfterFunc(time.Duration(d.cfg.SessionGap)*time.Second, func() {
				d.mu.Lock()
				defer d.mu.Unlock()
				if d.state == StateEnding {
					d.endSession()
				}
			})
			d.emit("playback:stopped", nil)
		} else if !SameTrack(d.currentTrack, track) {
			d.onTrackChange(track)
		} else {
			// Same track, update position
			d.currentTrack.PositionMS = track.PositionMS
			d.currentTrack.DetectedAt = track.DetectedAt
			d.emit("track:updated", track)
		}

	case StateEnding:
		if track != nil {
			// Music resumed within gap
			if d.gapTimer != nil {
				d.gapTimer.Stop()
			}
			d.state = StateActive
			if !SameTrack(d.currentTrack, track) {
				d.onTrackChange(track)
			} else {
				d.currentTrack.PositionMS = track.PositionMS
			}
			d.emit("playback:started", track)
		}
	}
}

func (d *Detector) startSession(track *TrackInfo) {
	now := time.Now()

	session := &diary.Session{
		ID:        diary.NewULID(),
		StartedAt: now.UTC().Format(time.RFC3339),
		Source:    track.Source,
		DayOfWeek: (int(now.Weekday()) + 6) % 7, // Convert Sunday=0 to Monday=0
		HourOfDay: now.Hour(),
		Month:     int(now.Month()),
		Year:      now.Year(),
		DateLocal: now.Format("2006-01-02"),
		TimeOfDay: diary.TimeOfDay(now.Hour()),
	}

	// Fetch weather
	if weatherData, err := d.weather.GetCurrent(); err == nil && weatherData != nil {
		session.WeatherTemp = weatherData.TempCelsius
		session.WeatherCond = weatherData.Condition
		session.WeatherDesc = weatherData.Description
		session.WeatherHumid = weatherData.Humidity
		session.WeatherIcon = weatherData.IconCode
	}

	if err := d.store.CreateSession(session); err != nil {
		slog.Error("failed to create session", "err", err)
		return
	}

	d.state = StateActive
	d.currentSession = session
	d.currentTrack = track
	d.trackCount = 0
	d.lastTrackTime = now

	d.addTrackToSession(track)

	d.emit("session:started", session)
	d.emit("track:changed", track)
	slog.Info("session started", "id", session.ID, "track", track.Title)
}

func (d *Detector) onTrackChange(track *TrackInfo) {
	d.currentTrack = track
	d.lastTrackTime = time.Now()

	// Update session source if mixed
	if d.currentSession != nil && d.currentSession.Source != track.Source {
		d.currentSession.Source = "mixed"
	}

	d.addTrackToSession(track)

	d.emit("track:changed", track)
	slog.Info("track changed", "title", track.Title, "artist", track.Artist)
}

func (d *Detector) addTrackToSession(track *TrackInfo) {
	if d.currentSession == nil {
		return
	}

	st := &diary.SessionTrack{
		ID:                diary.NewULID(),
		SessionID:         d.currentSession.ID,
		Title:             track.Title,
		Artist:            track.Artist,
		Album:             track.Album,
		DurationMS:        track.DurationMS,
		ListenedMS:        track.DurationMS, // Approximate
		AlbumArtURL:       track.AlbumArtURL,
		SpotifyID:         track.SpotifyID,
		Source:            track.Source,
		PlayedAt:          time.Now().UTC().Format(time.RFC3339),
		PositionInSession: d.trackCount,
	}

	if len(track.Genres) > 0 {
		st.Genres = `["` + joinGenres(track.Genres) + `"]`
	}

	if track.Energy > 0 {
		st.Energy = track.Energy
	}
	if track.Valence > 0 {
		st.Valence = track.Valence
	}
	if track.Tempo > 0 {
		st.Tempo = track.Tempo
	}

	if err := d.store.AddTrack(st); err != nil {
		slog.Error("failed to add track", "err", err)
	}

	d.trackCount++
}

func joinGenres(genres []string) string {
	result := ""
	for i, g := range genres {
		if i > 0 {
			result += `","`
		}
		result += g
	}
	return result
}

func (d *Detector) endSession() {
	if d.currentSession == nil {
		d.state = StateIdle
		return
	}

	now := time.Now()
	startedAt, _ := time.Parse(time.RFC3339, d.currentSession.StartedAt)
	durationSec := int(now.Sub(startedAt).Seconds())

	// Discard sessions shorter than minimum duration
	if durationSec < d.cfg.MinSessionDuration {
		slog.Info("session too short, discarding", "duration_sec", durationSec)
		d.store.DeleteSession(d.currentSession.ID)
		d.currentSession = nil
		d.currentTrack = nil
		d.state = StateIdle
		d.emit("session:ended", nil)
		return
	}

	endedAt := now.UTC().Format(time.RFC3339)
	if err := d.store.EndSession(d.currentSession.ID, endedAt, durationSec, d.trackCount); err != nil {
		slog.Error("failed to end session", "err", err)
	}

	// Update artist stats
	d.store.UpdateArtistStats(d.currentSession.ID)

	d.currentSession.EndedAt = endedAt
	d.currentSession.DurationSec = durationSec
	d.currentSession.TrackCount = d.trackCount

	d.emit("session:ended", d.currentSession)
	slog.Info("session ended", "id", d.currentSession.ID, "duration", durationSec, "tracks", d.trackCount)

	d.currentSession = nil
	d.currentTrack = nil
	d.state = StateIdle
}

func (d *Detector) finalizeSession() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.state != StateIdle {
		d.endSession()
	}
}
