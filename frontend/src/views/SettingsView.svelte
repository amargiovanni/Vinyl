<script>
  import { onMount } from 'svelte';
  import {
    GetConfig, SaveConfig, GetDatabaseInfo,
    IsFirstRun, IsSpotifyConnected, ConnectSpotify, DisconnectSpotify,
    GetYearsWithData, ExportAnnualBook, UpdateLocationFromCoords, DetectLocation
  } from '../../wailsjs/go/main/App.js';

  let config = null;
  let dbInfo = {};
  let spotifyConnected = false;
  let years = [];
  let exporting = false;
  let exportMessage = '';
  let locationStatus = '';
  let saving = false;

  onMount(async () => {
    try {
      config = await GetConfig();
      dbInfo = await GetDatabaseInfo();
      spotifyConnected = await IsSpotifyConnected();
      years = await GetYearsWithData() || [];
    } catch (e) {
      console.error('Failed to load settings:', e);
    }
  });

  async function saveConfig() {
    if (!config) return;
    saving = true;
    try {
      await SaveConfig(config);
    } catch (e) {
      console.error('Failed to save config:', e);
    }
    saving = false;
  }

  async function requestLocation() {
    locationStatus = 'Detecting your location...';
    try {
      const loc = await DetectLocation();
      config.location.lat = loc.lat;
      config.location.lon = loc.lon;
      config.location.city = loc.city || '';
      await UpdateLocationFromCoords(loc.lat, loc.lon);
      locationStatus = `${loc.city}, ${loc.country} (${loc.lat.toFixed(4)}, ${loc.lon.toFixed(4)})`;
      await saveConfig();
    } catch (e) {
      locationStatus = `Could not detect location: ${e}`;
    }
  }

  async function handleConnectSpotify() {
    if (!config.spotify.client_id) {
      alert('Please enter your Spotify Client ID first');
      return;
    }
    try {
      await ConnectSpotify(config.spotify.client_id);
      spotifyConnected = true;
    } catch (e) {
      console.error('Spotify connect failed:', e);
    }
  }

  async function handleDisconnectSpotify() {
    try {
      await DisconnectSpotify();
      spotifyConnected = false;
    } catch (e) {
      console.error('Spotify disconnect failed:', e);
    }
  }

  async function handleExport(year) {
    exporting = true;
    exportMessage = '';
    try {
      const path = await ExportAnnualBook(year);
      if (path) {
        exportMessage = `Exported to ${path}`;
      }
    } catch (e) {
      exportMessage = `Export failed: ${e}`;
    }
    exporting = false;
  }
</script>

<div class="view">
  <h2 class="title">Settings</h2>

  {#if config}
    <div class="section">
      <h3 class="section-title">Location</h3>
      <div class="field">
        <button class="btn primary" on:click={requestLocation}>
          Use My Location
        </button>
        {#if locationStatus}
          <p class="field-hint">{locationStatus}</p>
        {/if}
        {#if config.location.lat}
          <p class="field-hint">{config.location.lat.toFixed(4)}° N, {config.location.lon.toFixed(4)}° E</p>
        {/if}
      </div>
      <div class="field">
        <label class="field-label" for="settings-city">City (optional)</label>
        <input id="settings-city" type="text" class="input" bind:value={config.location.city} on:blur={saveConfig} placeholder="e.g. Pescara" />
      </div>
    </div>

    <div class="section">
      <h3 class="section-title">Weather API Key</h3>
      <div class="field">
        <input type="password" class="input" bind:value={config.weather_api_key} on:blur={saveConfig} placeholder="OpenWeatherMap API key" />
        <p class="field-hint">Get a free key at openweathermap.org/api</p>
      </div>
    </div>

    <div class="section">
      <h3 class="section-title">Spotify</h3>
      {#if spotifyConnected}
        <div class="connected-badge">Connected</div>
        <button class="btn danger" on:click={handleDisconnectSpotify}>Disconnect</button>
      {:else}
        <div class="field">
          <label class="field-label" for="settings-spotify">Client ID</label>
          <input id="settings-spotify" type="text" class="input" bind:value={config.spotify.client_id} on:blur={saveConfig} placeholder="From developer.spotify.com" />
        </div>
        <button class="btn primary" on:click={handleConnectSpotify}>Connect Spotify</button>
      {/if}
    </div>

    <div class="section">
      <h3 class="section-title">Apple Music</h3>
      <div class="connected-badge">Detected automatically</div>
    </div>

    <div class="section">
      <h3 class="section-title">Export</h3>
      {#each years as yr}
        <button class="btn secondary" on:click={() => handleExport(yr)} disabled={exporting}>
          Export {yr} Annual Book
        </button>
      {/each}
      {#if years.length === 0}
        <p class="field-hint">No listening data yet</p>
      {/if}
      {#if exportMessage}
        <p class="field-hint">{exportMessage}</p>
      {/if}
    </div>

    <div class="section">
      <h3 class="section-title">Data</h3>
      <p class="data-info">Database: {dbInfo.path || '...'}</p>
      <p class="data-info">Size: {dbInfo.size_mb || '0'} MB · {dbInfo.session_count || 0} sessions</p>
    </div>
  {/if}

  <div class="footer">
    <p>Vinyl v1.0.0</p>
    <p>Made with music in Pescara</p>
  </div>
</div>

<style>
  .view {
    height: 100%;
    overflow-y: auto;
    padding: 12px 16px;
  }

  .title {
    font-family: 'Playfair Display', Georgia, serif;
    font-size: 1.3rem;
    color: var(--warm-white, #FAF3E8);
    font-weight: 400;
    margin-bottom: 16px;
  }

  .section {
    margin-bottom: 20px;
  }

  .section-title {
    font-family: 'Source Serif 4', serif;
    font-size: 0.85rem;
    color: var(--amber-glow);
    margin-bottom: 8px;
  }

  .field {
    margin-bottom: 8px;
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
    padding: 6px 10px;
    background: var(--groove-dark);
    border: 1px solid var(--worn-wood);
    border-radius: 6px;
    color: var(--cream);
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.8rem;
    outline: none;
    transition: border-color 150ms;
  }

  .input:focus {
    border-color: var(--amber-glow);
  }

  .input::placeholder {
    color: var(--text-faint, #6E5E4E);
  }

  .field-hint {
    font-size: 0.7rem;
    color: var(--text-faint, #6E5E4E);
    margin-top: 4px;
  }

  .btn {
    padding: 6px 14px;
    border: 1px solid var(--worn-wood);
    border-radius: 6px;
    font-family: 'Source Serif 4', serif;
    font-size: 0.8rem;
    cursor: pointer;
    transition: all 150ms;
    margin-top: 4px;
    margin-right: 6px;
  }

  .btn.primary {
    background: var(--amber-glow);
    color: var(--vinyl-black, #1A1410);
    border-color: var(--amber-glow);
  }

  .btn.primary:hover {
    background: var(--gold-bright);
  }

  .btn.secondary {
    background: var(--cardboard);
    color: var(--cream);
  }

  .btn.secondary:hover {
    background: var(--leather);
  }

  .btn.danger {
    background: transparent;
    color: #C47A5A;
    border-color: #C47A5A;
  }

  .btn.danger:hover {
    background: rgba(196, 122, 90, 0.1);
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .connected-badge {
    display: inline-block;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.7rem;
    color: #7BA67B;
    padding: 3px 8px;
    background: rgba(123, 166, 123, 0.1);
    border-radius: 4px;
    margin-bottom: 6px;
  }

  .data-info {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.7rem;
    color: var(--text-muted, #8E7E6E);
    margin-bottom: 2px;
    word-break: break-all;
  }

  .footer {
    text-align: center;
    padding: 16px 0 8px;
    border-top: 1px solid var(--groove-dark);
    margin-top: 16px;
  }

  .footer p {
    font-size: 0.7rem;
    color: var(--text-faint, #6E5E4E);
    margin: 2px 0;
  }
</style>
