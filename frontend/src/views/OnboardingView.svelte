<script>
  import { createEventDispatcher } from 'svelte';
  import { SaveConfig, GetConfig, CompleteOnboarding, ConnectSpotify, UpdateLocationFromCoords, DetectLocation } from '../../wailsjs/go/main/App.js';

  const dispatch = createEventDispatcher();

  let step = 0;
  let config = {
    location: { lat: 0, lon: 0, city: '' },
    weather_api_key: '',
    spotify: { client_id: '', redirect_uri: 'http://localhost:27750/callback' },
    polling_interval_idle_ms: 5000,
    polling_interval_playing_ms: 2000,
    min_session_duration_sec: 60,
    session_gap_sec: 30,
    first_run: true,
  };

  let locationStatus = '';
  let locationDetecting = false;

  async function requestLocation() {
    locationStatus = 'Detecting your location...';
    locationDetecting = true;
    try {
      const loc = await DetectLocation();
      config.location.lat = loc.lat;
      config.location.lon = loc.lon;
      config.location.city = loc.city || '';
      locationStatus = `${loc.city}, ${loc.country} (${loc.lat.toFixed(4)}, ${loc.lon.toFixed(4)})`;
      await UpdateLocationFromCoords(loc.lat, loc.lon);
    } catch (e) {
      locationStatus = `Could not detect location: ${e}`;
    }
    locationDetecting = false;
  }

  async function nextStep() {
    if (step < 3) {
      step++;
    }
  }

  async function finish() {
    try {
      await SaveConfig(config);
      await CompleteOnboarding();
    } catch (e) {
      console.error('Failed to save config:', e);
    }
    dispatch('complete');
  }

  async function handleConnectSpotify() {
    if (!config.spotify.client_id) return;
    try {
      await ConnectSpotify(config.spotify.client_id);
    } catch (e) {
      console.error('Spotify connect failed:', e);
    }
  }
</script>

<div class="onboarding">
  {#if step === 0}
    <div class="step welcome">
      <div class="vinyl-icon">
        <svg width="80" height="80" viewBox="0 0 80 80">
          <circle cx="40" cy="40" r="38" fill="#1A1410" stroke="#5C4A3A" stroke-width="1"/>
          <circle cx="40" cy="40" r="30" fill="none" stroke="rgba(255,255,255,0.05)" stroke-width="0.5"/>
          <circle cx="40" cy="40" r="24" fill="none" stroke="rgba(255,255,255,0.05)" stroke-width="0.5"/>
          <circle cx="40" cy="40" r="18" fill="none" stroke="rgba(255,255,255,0.05)" stroke-width="0.5"/>
          <circle cx="40" cy="40" r="12" fill="#3D3024"/>
          <circle cx="40" cy="40" r="3" fill="#1A1410"/>
        </svg>
      </div>
      <h1>Welcome to Vinyl</h1>
      <p class="subtitle">Every song is a memory.<br>Vinyl remembers them all.</p>
      <p class="desc">Vinyl quietly observes what you listen to and builds a rich emotional diary of your musical life.</p>
      <button class="btn primary" on:click={nextStep}>Get Started</button>
    </div>

  {:else if step === 1}
    <div class="step">
      <h2>Your Location</h2>
      <p class="desc">Vinyl uses your location to record weather conditions during listening sessions.</p>
      <button class="btn primary" on:click={requestLocation}>Allow Location Access</button>
      {#if locationStatus}
        <p class="status">{locationStatus}</p>
      {/if}
      <div class="field" style="margin-top: 16px">
        <label class="field-label" for="onboard-city">Or enter your city</label>
        <input id="onboard-city" type="text" class="input" bind:value={config.location.city} placeholder="e.g. Pescara" />
      </div>
      <button class="btn secondary" on:click={nextStep}>Next</button>
    </div>

  {:else if step === 2}
    <div class="step">
      <h2>Weather API Key</h2>
      <p class="desc">Vinyl uses OpenWeatherMap to record weather. The free tier is more than enough.</p>
      <div class="field">
        <input type="text" class="input" bind:value={config.weather_api_key} placeholder="Paste your API key" />
        <p class="hint">Get a free key at openweathermap.org/api</p>
      </div>
      <button class="btn secondary" on:click={nextStep}>
        {config.weather_api_key ? 'Next' : 'Skip for now'}
      </button>
    </div>

  {:else if step === 3}
    <div class="step">
      <h2>Spotify (Optional)</h2>
      <p class="desc">Connect Spotify for richer metadata: genres, audio features, and high-res album art. Apple Music works automatically.</p>
      <div class="field">
        <label class="field-label" for="onboard-spotify">Spotify Client ID</label>
        <input id="onboard-spotify" type="text" class="input" bind:value={config.spotify.client_id} placeholder="From developer.spotify.com" />
      </div>
      {#if config.spotify.client_id}
        <button class="btn primary" on:click={handleConnectSpotify}>Connect Spotify</button>
      {/if}
      <button class="btn finish" on:click={finish}>
        {config.spotify.client_id ? 'Done' : 'Skip & Start'}
      </button>
    </div>
  {/if}

  <div class="dots">
    {#each [0, 1, 2, 3] as i}
      <span class="dot" class:active={step === i}></span>
    {/each}
  </div>
</div>

<style>
  .onboarding {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    padding: 24px;
    text-align: center;
  }

  .step {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    width: 100%;
    max-width: 320px;
  }

  .welcome .vinyl-icon {
    margin-bottom: 8px;
    opacity: 0.9;
  }

  h1 {
    font-family: 'Playfair Display', Georgia, serif;
    font-size: 1.7rem;
    color: var(--warm-white, #FAF3E8);
    font-weight: 700;
  }

  h2 {
    font-family: 'Playfair Display', Georgia, serif;
    font-size: 1.3rem;
    color: var(--warm-white, #FAF3E8);
    font-weight: 400;
  }

  .subtitle {
    font-family: 'Source Serif 4', serif;
    font-size: 0.9rem;
    color: var(--amber-glow);
    font-style: italic;
    line-height: 1.4;
  }

  .desc {
    font-size: 0.8rem;
    color: var(--text-muted, #8E7E6E);
    line-height: 1.5;
    max-width: 280px;
  }

  .field {
    width: 100%;
    text-align: left;
  }

  .field-label {
    display: block;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.65rem;
    color: var(--text-faint, #6E5E4E);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 4px;
  }

  .input {
    width: 100%;
    padding: 8px 12px;
    background: var(--groove-dark);
    border: 1px solid var(--worn-wood);
    border-radius: 6px;
    color: var(--cream);
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.8rem;
    outline: none;
  }

  .input:focus {
    border-color: var(--amber-glow);
  }

  .input::placeholder {
    color: var(--text-faint, #6E5E4E);
  }

  .hint {
    font-size: 0.65rem;
    color: var(--text-faint, #6E5E4E);
    margin-top: 4px;
  }

  .status {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.7rem;
    color: var(--amber-glow);
  }

  .btn {
    padding: 8px 20px;
    border: 1px solid var(--worn-wood);
    border-radius: 8px;
    font-family: 'Source Serif 4', serif;
    font-size: 0.85rem;
    cursor: pointer;
    transition: all 150ms;
    margin-top: 8px;
  }

  .btn.primary {
    background: var(--amber-glow);
    color: var(--vinyl-black, #1A1410);
    border-color: var(--amber-glow);
  }

  .btn.primary:hover {
    background: var(--gold-bright);
  }

  .btn.secondary, .btn.finish {
    background: var(--cardboard);
    color: var(--cream);
  }

  .btn.secondary:hover, .btn.finish:hover {
    background: var(--leather);
  }

  .dots {
    display: flex;
    gap: 6px;
    margin-top: 24px;
  }

  .dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--leather);
    transition: background 200ms;
  }

  .dot.active {
    background: var(--amber-glow);
  }
</style>
