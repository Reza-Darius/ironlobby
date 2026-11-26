<script lang="ts">
  // A player's note, folded into a small speech-bubble icon next to their name.
  // Hover shows it; click (or tap) pins it open until you click elsewhere.
  // The bubble is a native popover, so it sits in the top layer and is never
  // clipped by the card it belongs to.
  import Icon from './Icon.svelte'

  let { note, name }: { note: string; name: string } = $props()

  const uid = $props.id()
  const bubbleId = `note-${uid}`
  let trigger = $state<HTMLButtonElement>()
  let bubble = $state<HTMLDivElement>()
  let open = $state(false)
  // opened by hovering (closes on leave) vs. pinned by a click (stays)
  let byHover = false

  const GAP = 8
  const EDGE = 8

  /** Above the icon, or below when there's no room; kept on screen. */
  function place() {
    if (!trigger || !bubble) return
    const r = trigger.getBoundingClientRect()
    const w = bubble.offsetWidth
    const h = bubble.offsetHeight
    let top = r.top - GAP - h
    if (top < EDGE) top = r.bottom + GAP
    const left = Math.max(EDGE, Math.min(r.left + r.width / 2 - w / 2, innerWidth - w - EDGE))
    bubble.style.top = `${top}px`
    bubble.style.left = `${left}px`
  }

  function show(hover: boolean) {
    if (!bubble) return
    byHover = hover
    if (!bubble.matches(':popover-open')) bubble.showPopover()
    place()
  }

  function hide() {
    if (bubble?.matches(':popover-open')) bubble.hidePopover()
  }

  function onclick() {
    // a hover-opened bubble gets pinned rather than closed (touch fires both)
    if (open && !byHover) hide()
    else show(false)
  }

  function ontoggle(e: Event) {
    open = (e as ToggleEvent).newState === 'open'
  }

  // follow the icon while the page scrolls
  $effect(() => {
    if (!open) return
    const follow = () => place()
    addEventListener('scroll', follow, true)
    addEventListener('resize', follow)
    return () => {
      removeEventListener('scroll', follow, true)
      removeEventListener('resize', follow)
    }
  })
</script>

<button
  type="button"
  class="note-btn"
  class:open
  aria-label="Note from {name}"
  aria-expanded={open}
  aria-controls={bubbleId}
  bind:this={trigger}
  {onclick}
  onmouseenter={() => !open && show(true)}
  onmouseleave={() => byHover && hide()}
  onfocus={() => !open && show(true)}
  onblur={() => byHover && hide()}
>
  <Icon name="note" size={13} />
</button>

<div class="bubble" id={bubbleId} popover="auto" role="tooltip" bind:this={bubble} {ontoggle}>
  {note}
</div>

<style>
  .note-btn {
    display: inline-grid;
    place-items: center;
    width: 20px;
    height: 20px;
    padding: 0;
    border: 0;
    border-radius: var(--radius);
    background: transparent;
    color: var(--muted);
    cursor: pointer;
    transition:
      color 0.12s,
      background 0.12s;
  }
  .note-btn:hover,
  .note-btn.open {
    color: var(--accent);
    background: rgb(214 170 76 / 0.12);
  }
  .note-btn:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }

  .bubble {
    /* undo the UA popover centring; place() sets top/left */
    position: fixed;
    inset: auto;
    margin: 0;
    max-width: min(280px, calc(100vw - 16px));
    padding: 8px 11px;
    background: var(--panel-2);
    color: var(--text);
    border: 1px solid var(--line-2);
    border-radius: var(--radius);
    box-shadow: var(--shadow-lg);
    font-size: 14px;
    line-height: 1.45;
    overflow-wrap: anywhere;
    white-space: normal;
  }
</style>
