<script>
  import { SetSessionMood } from '../../../wailsjs/go/main/App.js';

  export let sessionId = '';
  export let currentMood = '';

  const moods = [
    { key: 'happy', emoji: '😊', label: 'Happy' },
    { key: 'calm', emoji: '😌', label: 'Calm' },
    { key: 'energetic', emoji: '🔥', label: 'Energetic' },
    { key: 'sad', emoji: '😢', label: 'Sad' },
    { key: 'thoughtful', emoji: '🤔', label: 'Thoughtful' },
    { key: 'frustrated', emoji: '😤', label: 'Frustrated' },
    { key: 'in_love', emoji: '🥰', label: 'In Love' },
    { key: 'dreamy', emoji: '🌙', label: 'Dreamy' },
    { key: 'motivated', emoji: '💪', label: 'Motivated' },
    { key: 'celebrating', emoji: '🎉', label: 'Celebrating' },
    { key: 'sleepy', emoji: '😴', label: 'Sleepy' },
    { key: 'nostalgic', emoji: '🌊', label: 'Nostalgic' },
  ];

  async function selectMood(mood) {
    const newMood = currentMood === mood ? '' : mood;
    currentMood = newMood;
    try {
      await SetSessionMood(sessionId, newMood);
    } catch (e) {
      console.error('Failed to set mood:', e);
    }
  }
</script>

<div class="mood-picker">
  <span class="label">Mood</span>
  <div class="mood-grid">
    {#each moods as mood}
      <button
        class="mood-btn"
        class:selected={currentMood === mood.key}
        on:click={() => selectMood(mood.key)}
        aria-label={mood.label}
        title={mood.label}
      >
        <span class="emoji">{mood.emoji}</span>
      </button>
    {/each}
  </div>
</div>

<style>
  .mood-picker {
    width: 100%;
    padding: 8px 0;
  }

  .label {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.65rem;
    color: var(--text-faint);
    text-transform: uppercase;
    letter-spacing: 0.1em;
    display: block;
    margin-bottom: 6px;
  }

  .mood-grid {
    display: grid;
    grid-template-columns: repeat(6, 1fr);
    gap: 4px;
  }

  .mood-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    aspect-ratio: 1;
    border: 1.5px solid transparent;
    border-radius: 8px;
    background: var(--groove-dark);
    cursor: pointer;
    transition: all 150ms ease-out;
    padding: 0;
    min-width: 32px;
    min-height: 32px;
  }

  .mood-btn:hover {
    background: var(--cardboard);
    border-color: var(--leather);
  }

  .mood-btn.selected {
    border-color: var(--amber-glow);
    background: var(--cardboard);
    transform: scale(1.1);
    box-shadow: 0 0 8px rgba(196, 154, 108, 0.3);
  }

  .emoji {
    font-size: 1.1rem;
    line-height: 1;
  }

  .mood-btn:focus-visible {
    outline: 2px solid var(--amber-glow);
    outline-offset: 2px;
  }

  @media (prefers-reduced-motion: reduce) {
    .mood-btn.selected {
      transform: none;
    }
  }
</style>
