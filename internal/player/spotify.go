package player

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/browser"
	"github.com/zalando/go-keyring"
)

const (
	keychainService     = "com.vinyl.app"
	keychainAccount     = "spotify_refresh_token"
	spotifyAuthURL      = "https://accounts.spotify.com/authorize"
	spotifyTokenURL     = "https://accounts.spotify.com/api/token"
	spotifyAPIBase      = "https://api.spotify.com/v1"
	spotifyCallbackPort = 27750
)

// SpotifyClient handles Spotify API interactions.
type SpotifyClient struct {
	mu           sync.RWMutex
	clientID     string
	accessToken  string
	refreshToken string
	tokenExpiry  time.Time
	httpClient   *http.Client
	connected    bool
}

func NewSpotifyClient(clientID string) *SpotifyClient {
	sc := &SpotifyClient{
		clientID:   clientID,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	// Try to load existing refresh token from keychain
	if clientID != "" {
		if token, err := keyring.Get(keychainService, keychainAccount); err == nil && token != "" {
			sc.refreshToken = token
			sc.connected = true
			slog.Info("spotify: loaded refresh token from keychain")
		}
	}

	return sc
}

// IsConnected returns whether Spotify is authenticated.
func (sc *SpotifyClient) IsConnected() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.connected && sc.refreshToken != ""
}

// Connect starts the OAuth2 PKCE flow.
func (sc *SpotifyClient) Connect(ctx context.Context, clientID string) error {
	sc.mu.Lock()
	sc.clientID = clientID
	sc.mu.Unlock()

	verifier := generateCodeVerifier()
	challenge := generateCodeChallenge(verifier)
	state := generateState()

	resultCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", spotifyCallbackPort),
		Handler: mux,
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			errCh <- fmt.Errorf("state mismatch")
			fmt.Fprint(w, "<html><body><h1>Error: State mismatch</h1></body></html>")
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no code in callback")
			fmt.Fprint(w, "<html><body><h1>Error: No authorization code</h1></body></html>")
			return
		}
		resultCh <- code
		fmt.Fprint(w, "<html><body style='font-family:sans-serif;text-align:center;padding:3em;'><h1>Connected to Spotify</h1><p>You can close this window and return to Vinyl.</p></body></html>")
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	authURL := fmt.Sprintf("%s?client_id=%s&response_type=code&redirect_uri=%s&scope=%s&code_challenge_method=S256&code_challenge=%s&state=%s",
		spotifyAuthURL,
		url.QueryEscape(clientID),
		url.QueryEscape(fmt.Sprintf("http://localhost:%d/callback", spotifyCallbackPort)),
		url.QueryEscape("user-read-currently-playing user-read-playback-state"),
		url.QueryEscape(challenge),
		url.QueryEscape(state))

	if err := browser.OpenURL(authURL); err != nil {
		server.Shutdown(ctx)
		return fmt.Errorf("open browser: %w", err)
	}

	select {
	case code := <-resultCh:
		server.Shutdown(ctx)
		return sc.exchangeCode(code, verifier, clientID)
	case err := <-errCh:
		server.Shutdown(ctx)
		return err
	case <-time.After(5 * time.Minute):
		server.Shutdown(ctx)
		return fmt.Errorf("spotify auth timeout")
	case <-ctx.Done():
		server.Shutdown(ctx)
		return ctx.Err()
	}
}

// Disconnect removes the Spotify connection.
func (sc *SpotifyClient) Disconnect() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.accessToken = ""
	sc.refreshToken = ""
	sc.connected = false
	keyring.Delete(keychainService, keychainAccount)
	return nil
}

func (sc *SpotifyClient) exchangeCode(code, verifier, clientID string) error {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {fmt.Sprintf("http://localhost:%d/callback", spotifyCallbackPort)},
		"client_id":     {clientID},
		"code_verifier": {verifier},
	}

	resp, err := sc.httpClient.PostForm(spotifyTokenURL, data)
	if err != nil {
		return fmt.Errorf("token exchange: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("decode token response: %w", err)
	}

	sc.mu.Lock()
	sc.accessToken = tokenResp.AccessToken
	sc.refreshToken = tokenResp.RefreshToken
	sc.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	sc.connected = true
	sc.mu.Unlock()

	if err := keyring.Set(keychainService, keychainAccount, tokenResp.RefreshToken); err != nil {
		slog.Warn("failed to store spotify token in keychain", "err", err)
	}

	slog.Info("spotify: connected successfully")
	return nil
}

func (sc *SpotifyClient) refreshAccessToken() error {
	sc.mu.RLock()
	refreshToken := sc.refreshToken
	clientID := sc.clientID
	sc.mu.RUnlock()

	if refreshToken == "" {
		return fmt.Errorf("no refresh token")
	}

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
	}

	resp, err := sc.httpClient.PostForm(spotifyTokenURL, data)
	if err != nil {
		return fmt.Errorf("refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		sc.mu.Lock()
		sc.connected = false
		sc.mu.Unlock()
		return fmt.Errorf("refresh token failed with status %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("decode refresh response: %w", err)
	}

	sc.mu.Lock()
	sc.accessToken = tokenResp.AccessToken
	sc.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	if tokenResp.RefreshToken != "" {
		sc.refreshToken = tokenResp.RefreshToken
		keyring.Set(keychainService, keychainAccount, tokenResp.RefreshToken)
	}
	sc.mu.Unlock()

	return nil
}

func (sc *SpotifyClient) ensureToken() error {
	sc.mu.RLock()
	expired := time.Now().After(sc.tokenExpiry)
	hasToken := sc.accessToken != ""
	sc.mu.RUnlock()

	if !hasToken || expired {
		return sc.refreshAccessToken()
	}
	return nil
}

// GetCurrentTrack fetches the currently playing track from Spotify API.
func (sc *SpotifyClient) GetCurrentTrack() (*TrackInfo, error) {
	if !sc.IsConnected() {
		return nil, fmt.Errorf("not connected")
	}

	if err := sc.ensureToken(); err != nil {
		return nil, err
	}

	sc.mu.RLock()
	token := sc.accessToken
	sc.mu.RUnlock()

	req, err := http.NewRequest("GET", spotifyAPIBase+"/me/player/currently-playing", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("spotify API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if resp.StatusCode == http.StatusUnauthorized {
		if err := sc.refreshAccessToken(); err != nil {
			return nil, err
		}
		return sc.GetCurrentTrack()
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("spotify API returned %d", resp.StatusCode)
	}

	var result struct {
		IsPlaying  bool `json:"is_playing"`
		ProgressMS int  `json:"progress_ms"`
		Item       struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			DurationMS int    `json:"duration_ms"`
			Artists    []struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			} `json:"artists"`
			Album struct {
				Name   string `json:"name"`
				Images []struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"images"`
			} `json:"album"`
		} `json:"item"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode spotify response: %w", err)
	}

	if !result.IsPlaying {
		return nil, nil
	}

	artist := ""
	if len(result.Item.Artists) > 0 {
		artist = result.Item.Artists[0].Name
	}

	artURL := ""
	if len(result.Item.Album.Images) > 0 {
		artURL = result.Item.Album.Images[0].URL
	}

	return &TrackInfo{
		Title:       result.Item.Name,
		Artist:      artist,
		Album:       result.Item.Album.Name,
		AlbumArtURL: artURL,
		DurationMS:  result.Item.DurationMS,
		PositionMS:  result.ProgressMS,
		Source:      "spotify",
		SpotifyID:   result.Item.ID,
		IsPlaying:   true,
		DetectedAt:  time.Now(),
	}, nil
}

// GetSpotifyViaAppleScript detects Spotify playback via AppleScript (fallback).
func GetSpotifyViaAppleScript() (*TrackInfo, error) {
	script := `
	if application "Spotify" is running then
		tell application "Spotify"
			if player state is playing then
				set output to name of current track & "||" & artist of current track & "||" & album of current track & "||" & artwork url of current track & "||" & (duration of current track as string) & "||" & (player position as string)
				return output
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
		return nil, fmt.Errorf("spotify osascript: %w", err)
	}

	result := strings.TrimSpace(string(out))
	if result == "NOT_RUNNING" || result == "NOT_PLAYING" {
		return nil, nil
	}

	parts := strings.Split(result, "||")
	if len(parts) < 6 {
		return nil, fmt.Errorf("unexpected spotify response: %s", result)
	}

	duration, _ := strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)
	position, _ := strconv.ParseFloat(strings.TrimSpace(parts[5]), 64)

	// Spotify duration from AppleScript is in milliseconds
	durationMS := int(duration)
	if duration < 100000 {
		// Looks like seconds, convert
		durationMS = int(duration * 1000)
	}

	return &TrackInfo{
		Title:       strings.TrimSpace(parts[0]),
		Artist:      strings.TrimSpace(parts[1]),
		Album:       strings.TrimSpace(parts[2]),
		AlbumArtURL: strings.TrimSpace(parts[3]),
		DurationMS:  durationMS,
		PositionMS:  int(position * 1000),
		Source:      "spotify",
		IsPlaying:   true,
		DetectedAt:  time.Now(),
	}, nil
}

func generateCodeVerifier() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
