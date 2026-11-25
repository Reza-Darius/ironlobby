<script lang="ts">
  // Plain text with clickable links. Trusted sites (lib/links.ts) open
  // directly; anything else asks first and shows exactly where it leads.
  import { linkify } from '../lib/links'
  import Icon from './Icon.svelte'
  import Modal from './Modal.svelte'

  let { text }: { text: string } = $props()

  const parts = $derived(linkify(text))

  let pending = $state<URL | null>(null)
  let warnOpen = $state(false)

  function ask(url: URL) {
    pending = url
    warnOpen = true
  }

  function proceed() {
    if (pending) window.open(pending.href, '_blank', 'noopener,noreferrer')
    warnOpen = false
  }
</script>

{#each parts as part, i (i)}
  {#if part.kind === 'text'}{part.text}{:else if part.trusted}<a
      class="link"
      href={part.url.href}
      target="_blank"
      rel="noopener noreferrer nofollow ugc">{part.text}</a
    >{:else}<!-- a button, not an <a>: middle-click / "open in new tab" can't skip the warning --><button
      type="button"
      class="link untrusted"
      title="External link: you'll be asked before it opens"
      onclick={() => ask(part.url)}
      >{part.text.slice(0, -1)}<span class="tail">{part.text.slice(-1)}<Icon name="link" size={12} /></span></button
    >{/if}
{/each}

<Modal bind:open={warnOpen} title="Leaving Iron Lobby">
  {#if pending}
    <div class="stack">
      <p class="muted">This link was added by the lobby host and goes to a site we don't recognise. Only open it if you trust it.</p>
      <div class="target">
        <span class="host">{pending.hostname}</span>
        <span class="full">{pending.href}</span>
      </div>
      <div class="row end">
        <button type="button" class="btn ghost" onclick={() => (warnOpen = false)}>Cancel</button>
        <button type="button" class="btn primary" onclick={proceed}>Open link</button>
      </div>
    </div>
  {/if}
</Modal>

<style>
  .link {
    color: var(--accent-2);
    text-decoration: underline;
    text-decoration-color: rgb(232 192 106 / 0.4);
    text-underline-offset: 2px;
    overflow-wrap: anywhere;
  }
  .link:hover {
    text-decoration-color: currentColor;
  }
  button.link {
    display: inline;
    padding: 0;
    border: 0;
    background: none;
    font: inherit;
    text-align: left;
    cursor: pointer;
  }
  /* last character + icon never split, so the icon can't wrap onto its own line */
  .tail {
    white-space: nowrap;
  }
  button.link :global(svg) {
    display: inline-block;
    margin-left: 3px;
    vertical-align: -1px;
    opacity: 0.8;
  }
  .target {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 12px 14px;
    background: var(--bg-2);
    border: 1px solid var(--line-2);
    border-radius: var(--radius);
  }
  .host {
    font-weight: 600;
    font-size: 16px;
    color: var(--text);
  }
  .full {
    font-family: var(--mono);
    font-size: 13px;
    color: var(--muted);
    overflow-wrap: anywhere;
  }
</style>
