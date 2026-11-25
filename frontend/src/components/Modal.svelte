<script lang="ts">
  import type { Snippet } from 'svelte'

  let {
    open = $bindable(false),
    title,
    subtitle = '',
    onclose,
    children,
  }: {
    open?: boolean
    title: string
    subtitle?: string
    onclose?: () => void
    children: Snippet
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
      <!-- no close button: every prompt ends in a Cancel action; Escape and
           clicking outside also close it -->
      <header>
        <h2>{title}</h2>
        {#if subtitle}<p class="muted">{subtitle}</p>{/if}
      </header>
      <div class="body">{@render children()}</div>
    </div>
  {/if}
</dialog>

<style>
  /* the dialog itself is the surface: a shadow on an inner element gets
     clipped at the dialog's edge */
  dialog {
    padding: 0;
    background: var(--panel);
    border: 1px solid var(--line-2);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
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
  header {
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
  @keyframes rise {
    from {
      opacity: 0;
      transform: translateY(8px) scale(0.98);
    }
  }
</style>
