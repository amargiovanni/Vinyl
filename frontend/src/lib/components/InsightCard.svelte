<script>
  export let icon = '';
  export let title = '';
  export let description = '';
  export let stat = '';
</script>

<div class="insight-card">
  <div class="icon">{icon}</div>
  <div class="content">
    <h4 class="title">{title}</h4>
    <p class="description">{description}</p>
  </div>
  {#if stat}
    <div class="stat">{stat}</div>
  {/if}
</div>

<style>
  .insight-card {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    background: var(--cardboard);
    border: 1px solid var(--worn-wood);
    border-radius: 8px;
    padding: 12px;
    box-shadow: 0 1px 3px rgba(0,0,0,0.3), inset 0 1px 0 rgba(255,255,255,0.03);
  }

  .icon {
    font-size: 1.5rem;
    flex-shrink: 0;
    line-height: 1;
  }

  .content {
    flex: 1;
    min-width: 0;
  }

  .title {
    font-family: 'Playfair Display', Georgia, serif;
    font-size: 0.85rem;
    color: var(--cream);
    font-weight: 400;
    line-height: 1.3;
  }

  .description {
    font-size: 0.75rem;
    color: var(--text-muted);
    margin-top: 3px;
    line-height: 1.4;
  }

  .stat {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.7rem;
    color: var(--amber-glow);
    flex-shrink: 0;
    text-align: right;
  }
</style>
