<script lang="ts">
  import type { Snippet } from 'svelte'
  import Icon from './Icon.svelte'

  let {
    open = $bindable(false),
    title,
    subtitle = '',
    onclose,
    children,
    footer,
  }: {
    open?: boolean
    title: string
    subtitle?: string
    onclose?: () => void
    children: Snippet
    footer?: Snippet
  } = $props()

  let dialog: HTMLDialogElement

  $effect(() => {
    if (open && !dialog.open) dialog.showModal()
    else if (!open && dialog.open) dialog.close()
  })
</script>

<!-- Native <dialog>: focus trapping, Escape and top-layer for free. -->
<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
<dialog
  bind:this={dialog}
  onclose={() => {
    open = false
    onclose?.()
  }}
  onclick={(e) => e.target === dialog && (open = false)}
>
  {#if open}
    <div class="sheet">
      <header>
        <div>
          <h2>{title}</h2>
          {#if subtitle}<p class="muted">{subtitle}</p>{/if}
        </div>
        <button class="btn icon ghost" onclick={() => (open = false)} aria-label="Close">
          <Icon name="x" />
        </button>
      </header>
      <div class="body">{@render children()}</div>
      {#if footer}<footer>{@render footer()}</footer>{/if}
    </div>
  {/if}
</dialog>

<style>
  dialog {
    padding: 0;
    border: 0;
    background: transparent;
    color: inherit;
    max-width: min(520px, calc(100vw - 32px));
    width: 100%;
    max-height: calc(100dvh - 32px);
  }
  dialog::backdrop {
    background: rgb(5 6 4 / 0.7);
    backdrop-filter: blur(3px);
  }
  /* animate the dialog itself, not .sheet: moving the sheet inside the
     dialog overflows it for a frame and flashes a scrollbar */
  dialog[open] {
    animation: rise 0.18s ease-out;
  }
  .sheet {
    background: var(--panel);
    border: 1px solid var(--line-2);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
  }
  header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
    padding: 20px 20px 0;
  }
  h2 {
    font-size: 20px;
  }
  header p {
    margin-top: 4px;
    font-size: 14px;
  }
  .body {
    padding: 16px 20px 20px;
  }
  footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 14px 20px;
    border-top: 1px solid var(--line);
    background: var(--bg-2);
    border-radius: 0 0 var(--radius-lg) var(--radius-lg);
  }
  @keyframes rise {
    from {
      opacity: 0;
      transform: translateY(8px) scale(0.98);
    }
  }
</style>
