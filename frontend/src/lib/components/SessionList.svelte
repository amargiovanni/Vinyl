<script>
  import { GetSessionTracks } from '../../../wailsjs/go/main/App.js';
  import { formatTimeRange, formatDuration, formatDurationMS, weatherEmoji, moodEmoji, sourceLabel } from '../utils/formatters.js';

  export let sessions = [];

  let expandedSession = null;
  let expandedTracks = [];
  let loading = false;

  async function toggleSession(session) {
    if (expandedSession === session.id) {
      expandedSession = null;
      expandedTracks = [];
      return;
    }

    loading = true;
    expandedSession = session.id;
    try {
      expandedTracks = await GetSessionTracks(session.id) || [];
    } catch (e) {
      console.error('Failed to load tracks:', e);
      expandedTracks = [];
    }
    loading = false;
  }
</script>

<div class="session-list">
  {#each sessions as session (session.id)}
    <button class="session-card" on:click={() => toggleSession(session)}>
      <div class="session-header">
        <span class="time-range">
          {formatTimeRange(session.started_at, session.ended_at || new Date().toISOString())}
        </span>
        <span class="duration">{formatDuration(session.duration_sec)}</span>
      </div>
      <div class="session-meta">
        <span class="track-count">{session.track_count} tracks</span>
        {#if session.mood}
          <span class="mood">{moodEmoji(session.mood)}</span>
        {/if}
        {#if session.weather_cond}
          <span class="weather">{weatherEmoji(session.weather_cond)} {Math.round(session.weather_temp)}°</span>
        {/if}
        <span class="source" class:spotify={session.source === 'spotify'} class:apple={session.source === 'apple_music'}>
          {sourceLabel(session.source)}
        </span>
      </div>

      {#if expandedSession === session.id}
        <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-static-element-interactions -->
        <div class="track-list" on:click|stopPropagation>
          {#if loading}
            <p class="loading">Loading tracks...</p>
          {:else}
            {#each expandedTracks as track, i}
              <div class="track-item">
                <span class="track-num">{i + 1}.</span>
                <span class="track-name">{track.title}</span>
                <span class="track-sep">—</span>
                <span class="track-artist">{track.artist}</span>
                <span class="track-duration">{formatDurationMS(track.duration_ms)}</span>
              </div>
            {/each}
            {#if expandedTracks.length === 0}
              <p class="no-tracks">No tracks recorded</p>
            {/if}
          {/if}
        </div>
      {/if}
    </button>
  {/each}

  {#if sessions.length === 0}
    <div class="empty">
      <p>No sessions for this period</p>
    </div>
  {/if}
</div>

<style>
  .session-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .session-card {
    display: block;
    width: 100%;
    text-align: left;
    background: var(--cardboard);
    border: 1px solid var(--worn-wood);
    border-radius: 8px;
    padding: 10px 12px;
    cursor: pointer;
    transition: all 200ms ease-out;
    color: inherit;
    font-family: inherit;
    box-shadow: 0 1px 3px rgba(0,0,0,0.3), inset 0 1px 0 rgba(255,255,255,0.03);
  }

  .session-card:hover {
    border-color: var(--amber-glow);
  }

  .session-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 6px;
  }

  .time-range {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.8rem;
    color: var(--cream);
  }

  .duration {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.75rem;
    color: var(--amber-glow);
    font-weight: 600;
  }

  .session-meta {
    display: flex;
    gap: 10px;
    align-items: center;
    flex-wrap: wrap;
  }

  .track-count, .weather {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.7rem;
    color: var(--text-muted);
  }

  .mood {
    font-size: 0.9rem;
  }

  .source {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.6rem;
    padding: 1px 6px;
    border-radius: 4px;
    background: var(--groove-dark);
    color: var(--text-muted);
  }

  .source.spotify { color: #7BA67B; }
  .source.apple { color: #B87A8E; }

  .track-list {
    margin-top: 10px;
    padding-top: 8px;
    border-top: 1px solid var(--leather);
    max-height: 200px;
    overflow-y: auto;
  }

  .track-item {
    display: flex;
    align-items: baseline;
    gap: 6px;
    padding: 3px 0;
    font-size: 0.75rem;
  }

  .track-num {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.65rem;
    color: var(--text-faint);
    min-width: 18px;
    text-align: right;
  }

  .track-name {
    color: var(--cream);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    flex: 1;
    min-width: 0;
  }

  .track-sep {
    color: var(--text-faint);
    flex-shrink: 0;
  }

  .track-artist {
    color: var(--text-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 100px;
  }

  .track-duration {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.65rem;
    color: var(--text-faint);
    flex-shrink: 0;
  }

  .loading, .no-tracks {
    font-size: 0.75rem;
    color: var(--text-muted);
    text-align: center;
    padding: 8px;
  }

  .empty {
    text-align: center;
    padding: 24px;
    color: var(--text-muted);
    font-size: 0.85rem;
  }
</style>
