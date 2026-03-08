<script>
  import { onMount } from 'svelte';
  import { currentView } from './lib/stores/navigation.js';
  import { initTrackEvents } from './lib/stores/currentTrack.js';
  import { IsFirstRun } from '../wailsjs/go/main/App.js';

  import NowPlayingView from './views/NowPlayingView.svelte';
  import DiaryView from './views/DiaryView.svelte';
  import InsightsView from './views/InsightsView.svelte';
  import SettingsView from './views/SettingsView.svelte';
  import OnboardingView from './views/OnboardingView.svelte';

  let showOnboarding = false;
  let loaded = false;
  let direction = 0;

  const tabs = [
    { id: 'nowplaying', icon: '♪', label: 'Now Playing' },
    { id: 'diary', icon: '◉', label: 'Diary' },
    { id: 'insights', icon: '◆', label: 'Insights' },
    { id: 'settings', icon: '⚙', label: 'Settings' },
  ];

  const tabOrder = { nowplaying: 0, diary: 1, insights: 2, settings: 3 };

  onMount(async () => {
    try {
      showOnboarding = await IsFirstRun();
    } catch (e) {
      showOnboarding = false;
    }
    initTrackEvents();
    loaded = true;
  });

  function switchTab(tabId) {
    const oldIdx = tabOrder[$currentView] || 0;
    const newIdx = tabOrder[tabId] || 0;
    direction = newIdx > oldIdx ? 1 : -1;
    currentView.set(tabId);
  }

  function onOnboardingComplete() {
    showOnboarding = false;
  }
</script>

<div class="app vinyl-noise vinyl-vignette" class:loaded>
  {#if showOnboarding}
    <OnboardingView on:complete={onOnboardingComplete} />
  {:else}
    <div class="content">
      {#if $currentView === 'nowplaying'}
        <NowPlayingView />
      {:else if $currentView === 'diary'}
        <DiaryView />
      {:else if $currentView === 'insights'}
        <InsightsView />
      {:else if $currentView === 'settings'}
        <SettingsView />
      {/if}
    </div>

    <nav class="tab-bar">
      {#each tabs as tab}
        <button
          class="tab"
          class:active={$currentView === tab.id}
          on:click={() => switchTab(tab.id)}
          aria-label={tab.label}
          title={tab.label}
        >
          <span class="tab-icon">{tab.icon}</span>
          <div class="tab-indicator" class:active={$currentView === tab.id}></div>
        </button>
      {/each}
    </nav>
  {/if}
</div>

<style>
  .app {
    width: 380px;
    height: 520px;
    display: flex;
    flex-direction: column;
    background: var(--vinyl-black);
    overflow: hidden;
    opacity: 0;
    transition: opacity 400ms ease-in;
    border-radius: 10px;
  }

  .app.loaded {
    opacity: 1;
  }

  .content {
    flex: 1;
    overflow: hidden;
  }

  .tab-bar {
    display: flex;
    justify-content: space-around;
    align-items: center;
    height: 44px;
    background: var(--groove-dark);
    border-top: 1px solid rgba(255,255,255,0.04);
    flex-shrink: 0;
    padding: 0 8px;
  }

  .tab {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 3px;
    padding: 6px 16px;
    background: transparent;
    border: none;
    cursor: pointer;
    color: var(--text-faint, #6E5E4E);
    transition: color 200ms;
    position: relative;
  }

  .tab:hover {
    color: var(--text-muted, #8E7E6E);
  }

  .tab.active {
    color: var(--amber-glow);
  }

  .tab-icon {
    font-size: 1.1rem;
    line-height: 1;
  }

  .tab-indicator {
    width: 0;
    height: 2px;
    background: var(--amber-glow);
    border-radius: 1px;
    transition: width 200ms ease-in-out;
  }

  .tab-indicator.active {
    width: 20px;
  }

  .tab:focus-visible {
    outline: 2px solid var(--amber-glow);
    outline-offset: -2px;
    border-radius: 4px;
  }
</style>
