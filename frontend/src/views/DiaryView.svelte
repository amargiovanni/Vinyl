<script>
  import { onMount } from 'svelte';
  import { selectedDate } from '../lib/stores/navigation.js';
  import { GetSessionsByMonth, GetSessionsByDate } from '../../wailsjs/go/main/App.js';
  import Heatmap from '../lib/components/Heatmap.svelte';
  import SessionList from '../lib/components/SessionList.svelte';
  import { formatDate, isToday, isYesterday } from '../lib/utils/formatters.js';

  let year = new Date().getFullYear();
  let month = new Date().getMonth() + 1;
  let sessions = [];
  let loading = false;
  let groupedSessions = [];

  onMount(() => {
    loadSessions();
  });

  // React to selectedDate changes (from heatmap click)
  $: if ($selectedDate) {
    const d = new Date($selectedDate + 'T00:00:00');
    year = d.getFullYear();
    month = d.getMonth() + 1;
    loadSessions();
  }

  async function loadSessions() {
    loading = true;
    try {
      sessions = await GetSessionsByMonth(year, month) || [];
      groupSessions();
    } catch (e) {
      console.error('Failed to load sessions:', e);
      sessions = [];
    }
    loading = false;
  }

  function groupSessions() {
    const groups = {};
    sessions.forEach(s => {
      const date = s.date_local;
      if (!groups[date]) groups[date] = [];
      groups[date].push(s);
    });
    groupedSessions = Object.entries(groups)
      .sort(([a], [b]) => b.localeCompare(a))
      .map(([date, sessions]) => ({
        date,
        label: getDateLabel(date),
        sessions
      }));
  }

  function getDateLabel(dateStr) {
    if (isToday(dateStr)) return 'Today';
    if (isYesterday(dateStr)) return 'Yesterday';
    return formatDate(dateStr);
  }

  function prevMonth() {
    month--;
    if (month < 1) { month = 12; year--; }
    selectedDate.set(null);
    loadSessions();
  }

  function nextMonth() {
    month++;
    if (month > 12) { month = 1; year++; }
    selectedDate.set(null);
    loadSessions();
  }

  const monthNames = ['', 'January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December'];
</script>

<div class="view">
  <div class="header">
    <h2 class="title">Diary</h2>
    <div class="month-nav">
      <button class="nav-btn" on:click={prevMonth} aria-label="Previous month">&larr;</button>
      <span class="month-label">{monthNames[month]} {year}</span>
      <button class="nav-btn" on:click={nextMonth} aria-label="Next month">&rarr;</button>
    </div>
  </div>

  <Heatmap />

  <div class="sessions-container">
    {#if loading}
      <p class="loading">Loading sessions...</p>
    {:else}
      {#each groupedSessions as group (group.date)}
        <div class="date-group">
          <h3 class="date-header">{group.label}</h3>
          <SessionList sessions={group.sessions} />
        </div>
      {/each}
      {#if groupedSessions.length === 0}
        <p class="empty">No listening sessions this month</p>
      {/if}
    {/if}
  </div>
</div>

<style>
  .view {
    height: 100%;
    overflow-y: auto;
    padding: 12px 16px;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .title {
    font-family: 'Playfair Display', Georgia, serif;
    font-size: 1.3rem;
    color: var(--warm-white, #FAF3E8);
    font-weight: 400;
  }

  .month-nav {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .nav-btn {
    background: var(--groove-dark);
    border: 1px solid var(--worn-wood);
    border-radius: 4px;
    color: var(--cream);
    padding: 2px 8px;
    cursor: pointer;
    font-size: 0.8rem;
    transition: background 150ms;
  }

  .nav-btn:hover {
    background: var(--cardboard);
  }

  .month-label {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.75rem;
    color: var(--text-muted, #8E7E6E);
    min-width: 100px;
    text-align: center;
  }

  .sessions-container {
    margin-top: 8px;
  }

  .date-group {
    margin-bottom: 16px;
  }

  .date-header {
    font-family: 'Playfair Display', Georgia, serif;
    font-size: 0.9rem;
    color: var(--amber-glow);
    font-weight: 400;
    margin-bottom: 8px;
    padding-bottom: 4px;
    border-bottom: 1px solid var(--groove-dark);
  }

  .loading, .empty {
    text-align: center;
    padding: 32px;
    color: var(--text-muted, #8E7E6E);
    font-size: 0.85rem;
  }
</style>
