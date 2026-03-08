<script>
  export let albumArtUrl = '';
  export let spinning = false;
</script>

<div class="vinyl-container">
  <div class="vinyl-disc" class:spinning>
    <div class="grooves"></div>
    <div class="label">
      {#if albumArtUrl}
        <img src={albumArtUrl} alt="Album art" class="album-art" />
      {:else}
        <div class="album-art-placeholder"></div>
      {/if}
      <div class="center-hole"></div>
    </div>
    <div class="reflection"></div>
  </div>
  <div class="vinyl-shadow"></div>
</div>

<style>
  .vinyl-container {
    position: relative;
    width: 120px;
    height: 120px;
    margin: 0 auto;
  }

  .vinyl-disc {
    width: 120px;
    height: 120px;
    border-radius: 50%;
    background: #1a1410;
    position: relative;
    overflow: hidden;
    box-shadow: 0 0 0 2px rgba(255,255,255,0.05), 0 4px 12px rgba(0,0,0,0.5);
  }

  .vinyl-disc.spinning {
    animation: spin 1.8s linear infinite;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  .grooves {
    position: absolute;
    top: 0; left: 0; right: 0; bottom: 0;
    border-radius: 50%;
    background: repeating-radial-gradient(
      circle at center,
      transparent 20px,
      rgba(255,255,255,0.03) 20.5px,
      transparent 21px,
      transparent 23px
    );
  }

  .label {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background: #3D3024;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .album-art {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    object-fit: cover;
  }

  .album-art-placeholder {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: linear-gradient(135deg, #5C4A3A, #3D3024);
  }

  .center-hole {
    position: absolute;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: #1a1410;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
  }

  .reflection {
    position: absolute;
    top: 0; left: 0; right: 0; bottom: 0;
    border-radius: 50%;
    background: conic-gradient(
      from 0deg,
      transparent 0deg,
      rgba(255,255,255,0.06) 30deg,
      transparent 60deg,
      transparent 180deg,
      rgba(255,255,255,0.04) 210deg,
      transparent 240deg
    );
    animation: spin 8s linear infinite;
  }

  .vinyl-shadow {
    position: absolute;
    bottom: -6px;
    left: 10%;
    width: 80%;
    height: 8px;
    background: rgba(0,0,0,0.3);
    border-radius: 50%;
    filter: blur(4px);
  }

  @media (prefers-reduced-motion: reduce) {
    .vinyl-disc.spinning, .reflection {
      animation: none;
    }
  }
</style>
