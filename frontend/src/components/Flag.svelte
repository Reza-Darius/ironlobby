<script lang="ts">
  import { flagUrl } from '../lib/flags'
  import { nationColor } from '../lib/nations'
  import { nameOf } from '../lib/state.svelte'

  let { tag, size = 'md' }: { tag: string; size?: 'sm' | 'md' | 'lg' } = $props()

  const url = $derived(flagUrl(tag))
</script>

{#if url}
  <img class="flag {size}" src={url} alt="Flag of {nameOf(tag)}" title={nameOf(tag)} loading="lazy" />
{:else}
  <!-- unknown tag (e.g. one added to the DB later): plain nation-coloured swatch -->
  <span class="flag {size}" style:background={nationColor(tag)} title={nameOf(tag)}></span>
{/if}

<style>
  .flag {
    display: block;
    flex-shrink: 0;
    aspect-ratio: 3 / 2; /* HoI4 flags are ~3:2; 4:3 fallbacks crop slightly */
    height: 18px;
    object-fit: cover;
    border-radius: 3px;
    /* thin edge so white-heavy flags (JAP, FIN) don't melt into light UI */
    box-shadow:
      0 0 0 1px rgb(255 255 255 / 0.12),
      0 1px 3px rgb(0 0 0 / 0.4);
  }
  .sm {
    height: 14px;
    border-radius: 2px;
  }
  .lg {
    height: 27px;
    border-radius: 4px;
  }
</style>
