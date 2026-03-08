<script>
  import { currentTrack, isPlaying, currentSession } from '../stores/currentTrack.js';
  import { formatDurationMS, formatDuration, weatherEmoji } from '../utils/formatters.js';
  import VinylDisc from './VinylDisc.svelte';
  import MoodPicker from './MoodPicker.svelte';

  $: track = $currentTrack;
  $: session = $currentSession;
  $: playing = $isPlaying;
  $: progress = track ? (track.position_ms / track.duration_ms) * 100 : 0;
  $: positionStr = track ? formatDurationMS(track.position_ms) : '0:00';
  $: durationStr = track ? formatDurationMS(track.duration_ms) : '0:00';
</script>

<div class="now-playing">
  {#if track}
    <div class="art-section">
      <VinylDisc albumArtUrl={track.album_art_url} spinning={playing} />
    </div>

    <div class="track-info">
      <h2 class="track-title">{track.title}</h2>
      <p class="track-artist">{track.artist} — {track.album}</p>
    </div>

    <div class="progress-section">
      <div class="progress-bar">
        <div class="progress-fill" style="width: {progress}%"></div>
        <div class="progress-dot" style="left: {progress}%"></div>
      </div>
      <div class="progress-times">
        <span class="time">{positionStr}</span>
        <span class="time">{durationStr}</span>
      </div>
    </div>

    {#if session}
      <MoodPicker sessionId={session.id} currentMood={session.mood} />
    {/if}

    <div class="session-info">
      {#if session}
        {#if session.weather_cond}
          <span class="weather">
            {weatherEmoji(session.weather_cond)} {Math.round(session.weather_temp)}°C · {session.weather_cond}
          </span>
        {/if}
        <span class="meta">
          Session: {formatDuration(session.duration_sec || Math.floor((Date.now() - new Date(session.started_at).getTime()) / 1000))} · {session.track_count || 0} tracks
        </span>
      {/if}
    </div>
  {:else}
    <div class="idle-state">
      <VinylDisc spinning={false} />
      <div class="idle-text">
        <h2>No music playing</h2>
        <p>Start playing a song in Spotify or Apple Music</p>
      </div>
    </div>
  {/if}
</div>

<style>
  .now-playing {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 16px 20px;
    height: 100%;
    gap: 12px;
  }

  .art-section {
    margin-top: 8px;
  }

  .track-info {
    text-align: center;
    width: 100%;
  }

  .track-title {
    font-family: 'Playfair Display', Georgia, serif;
    font-size: 1.05rem;
    color: var(--cream);
    font-weight: 400;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .track-artist {
    font-size: 0.8rem;
    color: var(--text-muted);
    margin-top: 2px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .progress-section {
    width: 100%;
  }

  .progress-bar {
    width: 100%;
    height: 3px;
    background: var(--groove-dark);
    border-radius: 2px;
    position: relative;
    cursor: default;
  }

  .progress-fill {
    height: 100%;
    background: var(--amber-glow);
    border-radius: 2px;
    transition: width 1s linear;
  }

  .progress-dot {
    position: absolute;
    top: 50%;
    width: 8px;
    height: 8px;
    background: var(--gold-bright);
    border-radius: 50%;
    transform: translate(-50%, -50%);
    transition: left 1s linear;
  }

  .progress-times {
    display: flex;
    justify-content: space-between;
    margin-top: 4px;
  }

  .time {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.7rem;
    color: var(--text-faint);
  }

  .session-info {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    width: 100%;
  }

  .weather, .meta {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.7rem;
    color: var(--text-muted);
  }

  .idle-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 24px;
    opacity: 0.7;
  }

  .idle-text {
    text-align: center;
  }

  .idle-text h2 {
    font-family: 'Playfair Display', Georgia, serif;
    font-size: 1.05rem;
    color: var(--cream);
    font-weight: 400;
  }

  .idle-text p {
    font-size: 0.8rem;
    color: var(--text-muted);
    margin-top: 6px;
  }
</style>
