<script>
  import { onMount } from 'svelte';
  import { GetInsights, GetListeningByDayOfWeek } from '../../wailsjs/go/main/App.js';
  import InsightCard from '../lib/components/InsightCard.svelte';

  let insights = [];
  let dayStats = [];
  let loading = true;
  let year = new Date().getFullYear();

  const dayNames = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

  onMount(async () => {
    try {
      const [insightData, dayData] = await Promise.all([
        GetInsights(),
        GetListeningByDayOfWeek(year)
      ]);
      insights = insightData || [];
      dayStats = dayData || [];
    } catch (e) {
      console.error('Failed to load insights:', e);
    }
    loading = false;
  });

  $: maxHours = Math.max(...dayStats.map(d => d.total_hours), 1);
</script>

<div class="view">
  <div class="header">
    <h2 class="title">Insights</h2>
    <span class="year">{year}</span>
  </div>

  {#if loading}
    <p class="loading">Analyzing your listening patterns...</p>
  {:else}
    {#if insights.length > 0}
      <div class="insights-list">
        {#each insights as insight}
          <InsightCard
            icon={insight.icon}
            title={insight.title}
            description={insight.description}
            stat={insight.stat}
          />
        {/each}
      </div>
    {:else}
      <div class="empty-insights">
        <p>Keep listening! Insights appear after a few sessions.</p>
      </div>
    {/if}

    {#if dayStats.length > 0}
      <div class="section">
        <h3 class="section-title">Your Week</h3>
        <div class="day-chart">
          {#each dayStats as stat}
            <div class="day-row">
              <span class="day-name">{dayNames[stat.day] || ''}</span>
              <div class="bar-bg">
                <div class="bar-fill" style="width: {(stat.total_hours / maxHours) * 100}%"></div>
              </div>
              <span class="day-hours">{stat.total_hours.toFixed(1)}h</span>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  {/if}
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
    margin-bottom: 16px;
  }

  .title {
    font-family: 'Playfair Display', Georgia, serif;
    font-size: 1.3rem;
    color: var(--warm-white, #FAF3E8);
    font-weight: 400;
  }

  .year {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.8rem;
    color: var(--text-muted, #8E7E6E);
  }

  .insights-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-bottom: 20px;
  }

  .section {
    margin-top: 16px;
  }

  .section-title {
    font-family: 'Playfair Display', Georgia, serif;
    font-size: 0.9rem;
    color: var(--amber-glow);
    font-weight: 400;
    margin-bottom: 10px;
    padding-bottom: 4px;
    border-bottom: 1px solid var(--groove-dark);
  }

  .day-chart {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .day-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .day-name {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.7rem;
    color: var(--text-muted, #8E7E6E);
    width: 28px;
    flex-shrink: 0;
  }

  .bar-bg {
    flex: 1;
    height: 14px;
    background: var(--groove-dark);
    border-radius: 3px;
    overflow: hidden;
  }

  .bar-fill {
    height: 100%;
    background: var(--amber-glow);
    border-radius: 3px;
    transition: width 500ms ease-out;
    min-width: 2px;
  }

  .day-hours {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.65rem;
    color: var(--text-faint, #6E5E4E);
    width: 32px;
    text-align: right;
    flex-shrink: 0;
  }

  .loading, .empty-insights {
    text-align: center;
    padding: 32px;
    color: var(--text-muted, #8E7E6E);
    font-size: 0.85rem;
  }
</style>
