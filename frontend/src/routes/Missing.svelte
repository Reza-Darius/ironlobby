<script lang="ts">
  import Icon from '../components/Icon.svelte'
</script>

<div class="missing">
  <p class="kicker">404 · Off the map</p>
  <h1>Nothing here</h1>
  <p class="muted">This page doesn't exist.</p>
  <a href="/" class="btn"><Icon name="back" size={15} />Back to base</a>
</div>

<style>
  .missing {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
    padding: 64px 0;
  }
  h1 {
    font-size: 40px;
    text-transform: uppercase;
  }
</style>
