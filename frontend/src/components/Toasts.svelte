<script lang="ts">
  import { fly } from 'svelte/transition'
  import { dismiss, toasts } from '../lib/state.svelte'
  import Icon from './Icon.svelte'
</script>

<div class="toasts" aria-live="polite">
  {#each toasts as t (t.id)}
    <button class="toast {t.kind}" onclick={() => dismiss(t.id)} transition:fly={{ y: 12, duration: 180 }}>
      <Icon name={t.kind === 'error' ? 'x' : 'check'} />
      <span>{t.text}</span>
    </button>
  {/each}
</div>

<style>
  .toasts {
    position: fixed;
    bottom: 16px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    flex-direction: column;
    gap: 8px;
    z-index: 100;
    width: min(420px, calc(100vw - 32px));
    pointer-events: none;
  }
  .toast {
    pointer-events: auto;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 11px 14px;
    background: var(--panel-2);
    border: 1px solid var(--line-2);
    border-left: 3px solid var(--ok);
    border-radius: var(--radius);
    box-shadow: var(--shadow-lg);
    color: var(--text);
    font: inherit;
    font-size: 14px;
    text-align: left;
    cursor: pointer;
  }
  .toast.error {
    border-left-color: var(--danger);
  }
  .toast.error :global(svg) {
    color: var(--danger);
  }
  .toast.info {
    border-left-color: var(--accent);
  }
  .toast :global(svg) {
    color: var(--ok);
  }
  .toast.info :global(svg) {
    color: var(--accent);
  }
</style>
