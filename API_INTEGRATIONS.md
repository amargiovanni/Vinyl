# API_INTEGRATIONS.md — Vinyl External Service Integrations

## 1. Spotify Integration

### 1.1 Authentication — OAuth 2.0 PKCE Flow

Vinyl uses the Authorization Code with PKCE flow (no client secret needed for desktop apps).

**Spotify App Setup** (user must create at https://developer.spotify.com):
- App name: "Vinyl"
- Redirect URI: `http://localhost:27750/callback`
- Scopes needed: `user-read-currently-playing`, `user-read-playback-state`

**Flow**:

```
1. User clicks "Connect Spotify" in settings
2. App generates:
   - code_verifier: 43-128 char random string (A-Z, a-z, 0-9, -._~)
   - code_challenge: base64url(sha256(code_verifier))
   - state: random 16-byte hex string
3. App starts local HTTP server on port 27750
4. App opens browser to:
   https://accounts.spotify.com/authorize?
     client_id={CLIENT_ID}
     &response_type=code
     &redirect_uri=http://localhost:27750/callback
     &scope=user-read-currently-playing%20user-read-playback-state
     &code_challenge_method=S256
     &code_challenge={code_challenge}
     &state={state}
5. User authorizes in browser
6. Spotify redirects to http://localhost:27750/callback?code={code}&state={state}
7. App validates state, exchanges code for tokens:
   POST https://accounts.spotify.com/api/token
   Content-Type: application/x-www-form-urlencoded
   Body:
     grant_type=authorization_code
     &code={code}
     &redirect_uri=http://localhost:27750/callback
     &client_id={CLIENT_ID}
     &code_verifier={code_verifier}
8. Response: { access_token, refresh_token, expires_in, token_type }
9. Store refresh_token in macOS Keychain
10. Close local HTTP server
```

**Token Refresh** (before each API call if expired):

```
POST https://accounts.spotify.com/api/token
Content-Type: application/x-www-form-urlencoded
Body:
  grant_type=refresh_token
  &refresh_token={refresh_token}
  &client_id={CLIENT_ID}
```

### 1.2 Currently Playing — Poll Endpoint

```
GET https://api.spotify.com/v1/me/player/currently-playing
Authorization: Bearer {access_token}
```

**Response** (200 OK, playing):
```json
{
  "is_playing": true,
  "progress_ms": 120000,
  "item": {
    "id": "spotify_track_id",
    "name": "Track Name",
    "duration_ms": 240000,
    "artists": [{"name": "Artist Name", "id": "artist_id"}],
    "album": {
      "name": "Album Name",
      "images": [
        {"url": "https://i.scdn.co/image/...", "width": 640, "height": 640},
        {"url": "https://i.scdn.co/image/...", "width": 300, "height": 300},
        {"url": "https://i.scdn.co/image/...", "width": 64, "height": 64}
      ]
    }
  }
}
```

**Response** (204 No Content): Nothing playing.
**Response** (401 Unauthorized): Token expired, refresh and retry.

**Rate Limits**: Spotify allows ~180 requests/minute. At 2s polling = 30 req/min. Safe.

### 1.3 Audio Features (optional enrichment)

After detecting a new track, optionally fetch audio features:

```
GET https://api.spotify.com/v1/audio-features/{track_id}
Authorization: Bearer {access_token}
```

**Relevant fields**:
- `energy` (0.0 - 1.0): intensity/activity
- `valence` (0.0 - 1.0): musical positiveness (happy vs sad)
- `tempo` (BPM)
- `danceability` (0.0 - 1.0)

**NOTE**: Audio features endpoint may be deprecated. Check availability. If unavailable, skip gracefully — these fields are optional enrichment only.

### 1.4 Artist Genres

```
GET https://api.spotify.com/v1/artists/{artist_id}
Authorization: Bearer {access_token}
```

Response includes `genres: ["jazz", "soul", "neo-soul"]`. Cache per artist (genres don't change often). Store in `session_tracks.genres` as JSON array.

### 1.5 Fallback: AppleScript Detection for Spotify

If no API token configured, detect Spotify playback via AppleScript:

```go
func getSpotifyViaAppleScript() (*TrackInfo, error) {
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
    // Execute via exec.Command("osascript", "-e", script)
    // Parse result by splitting on "||"
}
```

**Limitations of AppleScript fallback**:
- No genres, no audio features, no Spotify track ID
- Art URL available but lower reliability
- Sufficient for basic session tracking

---

## 2. Apple Music Integration

Apple Music on macOS has no public REST API for local playback detection. Use AppleScript exclusively.

### 2.1 AppleScript Detection

```go
func getAppleMusicTrack() (*TrackInfo, error) {
    script := `
    if application "Music" is running then
        tell application "Music"
            if player state is playing then
                set t to current track
                set trackName to name of t
                set artistName to artist of t
                set albumName to album of t
                set trackDuration to duration of t
                set playerPos to player position
                -- Album art: extract from track as raw data
                try
                    set artData to raw data of artwork 1 of t
                    -- Save to temp file for frontend display
                    set artPath to (POSIX path of (path to temporary items)) & "vinyl_art.jpg"
                    set fileRef to open for access artPath with write permission
                    write artData to fileRef
                    close access fileRef
                    return trackName & "||" & artistName & "||" & albumName & "||" & (trackDuration as string) & "||" & (playerPos as string) & "||" & artPath
                on error
                    return trackName & "||" & artistName & "||" & albumName & "||" & (trackDuration as string) & "||" & (playerPos as string) & "||NO_ART"
                end try
            else
                return "NOT_PLAYING"
            end if
        end tell
    else
        return "NOT_RUNNING"
    end if
    `
    // Execute and parse
}
```

### 2.2 Album Art for Apple Music

Apple Music's AppleScript provides raw artwork data. Strategy:
1. Extract raw bytes via AppleScript (see above)
2. Save to temp file
3. On new track: copy to art cache directory with content-hash filename
4. Serve to frontend via Wails static file serving or base64 inline

### 2.3 Player Priority

Detection order on each poll:
1. Check Spotify (API if token available, else AppleScript)
2. If Spotify not playing, check Apple Music (AppleScript)
3. If both playing (rare): prefer Spotify (richer metadata)
4. If neither: return `nil` (no active playback)

---

## 3. OpenWeatherMap Integration

### 3.1 API Setup

- Free tier: 1,000 calls/day, 60 calls/minute
- Signup: https://openweathermap.org/api
- API key provided by user in settings

### 3.2 Current Weather Endpoint

```
GET https://api.openweathermap.org/data/2.5/weather?lat={lat}&lon={lon}&appid={API_KEY}&units=metric&lang=en
```

**Response** (relevant fields):
```json
{
  "main": {
    "temp": 18.5,
    "humidity": 72
  },
  "weather": [
    {
      "main": "Rain",
      "description": "light rain",
      "icon": "10d"
    }
  ]
}
```

### 3.3 Caching Strategy

```go
type WeatherCache struct {
    Data      WeatherData
    FetchedAt time.Time
    ExpiresAt time.Time // FetchedAt + 30 minutes
}

func (c *WeatherClient) GetCurrent() (*WeatherData, error) {
    // 1. Check in-memory cache
    if c.cache != nil && time.Now().Before(c.cache.ExpiresAt) {
        return &c.cache.Data, nil
    }
    // 2. Check SQLite cache
    cached, err := c.store.GetLatestWeather(c.lat, c.lon)
    if err == nil && time.Now().Before(cached.ExpiresAt) {
        c.cache = cached
        return &cached.Data, nil
    }
    // 3. Fetch from API
    data, err := c.fetchFromAPI()
    if err != nil {
        // If API fails, return stale cache if available
        if cached != nil {
            return &cached.Data, nil
        }
        return nil, err
    }
    // 4. Update caches
    c.cache = &WeatherCache{Data: *data, FetchedAt: time.Now(), ExpiresAt: time.Now().Add(30 * time.Minute)}
    c.store.SaveWeather(c.lat, c.lon, data)
    return data, nil
}
```

### 3.4 Weather Icon Mapping

Map OpenWeatherMap icons to vintage-style emoji/icons in the UI:

| OWM Condition | Icon for UI | Color accent |
|---------------|-------------|--------------|
| Clear (day) | ☀️ | #E8C496 (gold) |
| Clear (night) | 🌙 | #8B7D6B (silver) |
| Clouds | ☁️ | #9E9083 (grey-brown) |
| Rain/Drizzle | 🌧️ | #6B8E9E (steel blue) |
| Thunderstorm | ⛈️ | #5A5A6E (dark blue-grey) |
| Snow | ❄️ | #C8C0B8 (off-white) |
| Mist/Fog | 🌫️ | #A89E92 (warm grey) |

---

## 4. macOS Keychain Integration

Store Spotify tokens securely using `github.com/zalando/go-keyring`:

```go
import "github.com/zalando/go-keyring"

const keychainService = "com.vinyl.app"

func SaveSpotifyToken(refreshToken string) error {
    return keyring.Set(keychainService, "spotify_refresh_token", refreshToken)
}

func GetSpotifyToken() (string, error) {
    return keyring.Get(keychainService, "spotify_refresh_token")
}

func DeleteSpotifyToken() error {
    return keyring.Delete(keychainService, "spotify_refresh_token")
}
```

---

## 5. Error Handling & Resilience

### General Principles

- **Never crash on API failure**: Log error, use fallback/cache, continue.
- **Exponential backoff**: On repeated API failures (Spotify 429, OWM 429), back off: 5s → 10s → 20s → 60s → 5min max.
- **Graceful degradation**:
  - Spotify API down → fall back to AppleScript
  - Weather API down → session saved without weather data
  - No internet → AppleScript detection still works, weather fields NULL
- **Logging**: Use `log/slog` structured logging. Log file at `~/Library/Logs/Vinyl/vinyl.log`. Rotate at 10MB.

### Specific Error Scenarios

| Scenario | Behavior |
|----------|----------|
| Spotify 401 | Refresh token. If refresh fails, mark as disconnected. Notify user. |
| Spotify 429 | Back off. Use AppleScript fallback. |
| Spotify 403 | Scope changed. Prompt re-authorization. |
| OWM 401 | Invalid API key. Show error in settings. |
| OWM 429 | Cache stale data. Retry in 5 min. |
| AppleScript timeout | Skip this poll cycle. Retry next interval. |
| Music app not running | Return nil. No error. Expected state. |
| SQLite write error | Log error. Retry once. If persistent, surface to user. |
| Disk full | Detect on write. Alert user. Stop art caching. |

---

## 6. Go Dependencies

```
require (
    github.com/wailsapp/wails/v2 v2.x.x
    github.com/mattn/go-sqlite3 v1.14.x
    github.com/oklog/ulid/v2 v2.1.x
    github.com/zalando/go-keyring v0.2.x
)
```

No other external dependencies. HTTP client is `net/http`. JSON is `encoding/json`. AppleScript is `os/exec`.
