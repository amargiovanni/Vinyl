<script>
  import { onMount } from 'svelte';
  import { GetHeatmapData } from '../../../wailsjs/go/main/App.js';
  import { navigateToDate } from '../stores/navigation.js';

  let heatmapData = [];
  let grid = [];
  let monthLabels = [];
  let hoveredCell = null;
  let tooltipX = 0;
  let tooltipY = 0;

  const dayLabels = ['M', '', 'W', '', 'F', '', 'S'];
  const cellSize = 11;
  const gap = 3;

  onMount(loadData);

  async function loadData() {
    try {
      const data = await GetHeatmapData();
      heatmapData = data || [];
      buildGrid();
    } catch (e) {
      console.error('Failed to load heatmap data:', e);
    }
  }

  function buildGrid() {
    // Build 52 weeks x 7 days grid
    const today = new Date();
    const dataMap = {};
    heatmapData.forEach(d => { dataMap[d.date_local] = d; });

    grid = [];
    monthLabels = [];

    // Start from 51 weeks ago
    const start = new Date(today);
    start.setDate(start.getDate() - 363);
    // Align to Monday
    const dayOfWeek = (start.getDay() + 6) % 7; // Convert to Mon=0
    start.setDate(start.getDate() - dayOfWeek);

    let lastMonth = -1;

    for (let week = 0; week < 52; week++) {
      const col = [];
      for (let day = 0; day < 7; day++) {
        const d = new Date(start);
        d.setDate(d.getDate() + week * 7 + day);
        const dateStr = d.toISOString().split('T')[0];
        const data = dataMap[dateStr];
        const totalMin = data ? Math.floor(data.total_listen_sec / 60) : 0;

        const month = d.getMonth();
        if (month !== lastMonth && day === 0) {
          monthLabels.push({
            label: d.toLocaleDateString('en-US', { month: 'short' }),
            week
          });
          lastMonth = month;
        }

        col.push({
          date: dateStr,
          minutes: totalMin,
          level: getLevel(totalMin),
          isToday: dateStr === today.toISOString().split('T')[0],
          isFuture: d > today,
        });
      }
      grid.push(col);
    }
  }

  function getLevel(minutes) {
    if (minutes <= 0) return 0;
    if (minutes <= 30) return 1;
    if (minutes <= 90) return 2;
    if (minutes <= 180) return 3;
    return 4;
  }

  const colors = ['#2A2018', '#5C4A3A', '#8B6D4F', '#C49A6C', '#E8C496'];

  function formatMinutes(min) {
    if (min <= 0) return 'No listening';
    const h = Math.floor(min / 60);
    const m = min % 60;
    if (h > 0) return `${h}h ${m}m listened`;
    return `${m}m listened`;
  }

  function onCellHover(e, cell) {
    if (cell.isFuture) return;
    hoveredCell = cell;
    const rect = e.target.getBoundingClientRect();
    tooltipX = rect.left + rect.width / 2;
    tooltipY = rect.top - 4;
  }

  function onCellLeave() {
    hoveredCell = null;
  }

  function onCellClick(cell) {
    if (!cell.isFuture && cell.minutes > 0) {
      navigateToDate(cell.date);
    }
  }
</script>

<div class="heatmap">
  <div class="month-labels">
    <div class="day-label-spacer"></div>
    {#each monthLabels as ml}
      <span class="month-label" style="left: {ml.week * (cellSize + gap) + 16}px">{ml.label}</span>
    {/each}
  </div>

  <div class="grid-container">
    <div class="day-labels">
      {#each dayLabels as label}
        <span class="day-label" style="height: {cellSize}px; line-height: {cellSize}px">{label}</span>
      {/each}
    </div>

    <div class="grid" style="grid-template-columns: repeat({grid.length}, {cellSize}px); gap: {gap}px">
      {#each grid as week, wi}
        {#each week as cell, di}
          <div
            class="cell"
            class:today={cell.isToday}
            class:future={cell.isFuture}
            class:clickable={cell.minutes > 0}
            style="width:{cellSize}px; height:{cellSize}px; background:{cell.isFuture ? 'transparent' : colors[cell.level]}; grid-row:{di + 1}; grid-column:{wi + 1}"
            role="button"
            tabindex={cell.isFuture ? -1 : 0}
            aria-label="{cell.date}: {formatMinutes(cell.minutes)}"
            on:mouseenter={(e) => onCellHover(e, cell)}
            on:mouseleave={onCellLeave}
            on:click={() => onCellClick(cell)}
            on:keydown={(e) => e.key === 'Enter' && onCellClick(cell)}
          ></div>
        {/each}
      {/each}
    </div>
  </div>

  {#if hoveredCell}
    <div class="tooltip" style="left:{tooltipX}px; top:{tooltipY}px">
      <strong>{hoveredCell.date}</strong><br>
      {formatMinutes(hoveredCell.minutes)}
    </div>
  {/if}
</div>

<style>
  .heatmap {
    width: 100%;
    padding: 8px;
    position: relative;
  }

  .month-labels {
    position: relative;
    height: 16px;
    margin-bottom: 4px;
    margin-left: 16px;
  }

  .month-label {
    position: absolute;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.6rem;
    color: var(--text-faint);
  }

  .day-label-spacer {
    width: 16px;
  }

  .grid-container {
    display: flex;
    gap: 4px;
    overflow-x: auto;
    padding-bottom: 4px;
  }

  .day-labels {
    display: flex;
    flex-direction: column;
    gap: 3px;
    flex-shrink: 0;
  }

  .day-label {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.55rem;
    color: var(--text-faint);
    display: flex;
    align-items: center;
    width: 12px;
  }

  .grid {
    display: grid;
    grid-template-rows: repeat(7, auto);
    grid-auto-flow: column;
  }

  .cell {
    border-radius: 2px;
    transition: opacity 100ms ease-out;
  }

  .cell:hover:not(.future) {
    opacity: 0.8;
    outline: 1px solid var(--amber-glow);
  }

  .cell.today {
    outline: 1.5px solid var(--amber-glow);
  }

  .cell.future {
    opacity: 0;
  }

  .cell.clickable {
    cursor: pointer;
  }

  .tooltip {
    position: fixed;
    transform: translate(-50%, -100%);
    background: var(--cardboard);
    border: 1px solid var(--worn-wood);
    padding: 4px 8px;
    border-radius: 4px;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.65rem;
    color: var(--cream);
    pointer-events: none;
    z-index: 100;
    white-space: nowrap;
    box-shadow: 0 2px 8px rgba(0,0,0,0.3);
  }
</style>
