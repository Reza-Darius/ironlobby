<script lang="ts">
  import { nameOf } from '../lib/state.svelte'
  import type { LobbyPlayer } from '../lib/types'
  import Icon from './Icon.svelte'
  import Stepper from './Stepper.svelte'
  import Flag from './Flag.svelte'
  import NoteBubble from './NoteBubble.svelte'

  let {
    tag,
    slots,
    occupants,
    myName,
    myTag,
    host,
    busy,
    onclaim,
    onleave,
    onnote,
    onslots,
    onremove,
    dragging = null,
    ondragseat,
    ondragdone,
    ondropseat,
    onkick,
  }: {
    tag: string
    slots: number
    occupants: LobbyPlayer[]
    myName: string | null
    myTag: string | null
    host: boolean
    busy: boolean
    onclaim: () => void
    onleave: () => void
    onnote: () => void
    onslots: (v: number) => void
    onremove: () => void
    /** host: the player currently being dragged (from any card) */
    dragging?: LobbyPlayer | null
    ondragseat?: (p: LobbyPlayer) => void
    ondragdone?: () => void
    /** host: the dragged player was dropped on this card */
    ondropseat?: () => void
    onkick?: (p: LobbyPlayer) => void
  } = $props()

  const isMine = $derived(myTag === tag)
  const full = $derived(occupants.length >= slots)
  const open = $derived(Math.max(0, slots - occupants.length))

  // ---- host: drag a player onto another card to move them -----------------
  // (touch screens get the per-seat Move button instead; HTML drag and drop
  // doesn't work with fingers on most mobile browsers)
  const dragFromHere = $derived(!!dragging && dragging.country_tag === tag)
  const canDrop = $derived(host && !busy && !!dragging && !dragFromHere && !full)
  let over = $state(false)
  let card = $state<HTMLElement>()

  function onseatdragstart(e: DragEvent, p: LobbyPlayer) {
    if (!e.dataTransfer) return
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', p.player_name) // Firefox won't start a drag without data
    ondragseat?.(p)
  }

  function ondragover(e: DragEvent) {
    if (!dragging || dragFromHere) return
    over = true
    if (!canDrop) {
      if (e.dataTransfer) e.dataTransfer.dropEffect = 'none'
      return
    }
    e.preventDefault() // allows the drop
    if (e.dataTransfer) e.dataTransfer.dropEffect = 'move'
  }

  function ondragleave(e: DragEvent) {
    // leaving into a child of the card isn't leaving the card
    if (!card?.contains(e.relatedTarget as Node | null)) over = false
  }

  function ondrop(e: DragEvent) {
    e.preventDefault()
    over = false
    if (canDrop) ondropseat?.()
  }

  $effect(() => {
    if (!dragging) over = false
  })
</script>

<article
  class="card"
  class:mine={isMine}
  class:drop-ok={over && canDrop}
  class:drop-no={over && !canDrop}
  class:drag-source={dragFromHere}
  bind:this={card}
  {ondragover}
  {ondragleave}
  {ondrop}
  role="group"
  aria-label={nameOf(tag)}
>

  <header>
    <Flag {tag} size="lg" />
    <h3>{nameOf(tag)}</h3>
    <!-- host: slots + remove. With a mouse they replace the name while the row
         is hovered, so host rows keep the same 50px-per-slot rhythm -->
    {#if host}
      <div class="host-controls">
        <Stepper value={slots} min={Math.max(1, occupants.length)} disabled={busy} onchange={onslots} />
        <button class="btn icon sm ghost danger" aria-label="Remove {nameOf(tag)} from lobby" title="Remove nation" onclick={onremove} disabled={busy}>
          <Icon name="trash" size={14} />
        </button>
      </div>
    {/if}
  </header>

  <ul class="seats">
    {#each occupants as p (p.player_name)}
      <li
        class="seat taken"
        class:me={p.player_name === myName}
        class:movable={host}
        class:dragged={dragging?.player_name === p.player_name}
        draggable={host && !busy}
        ondragstart={(e) => onseatdragstart(e, p)}
        ondragend={() => ondragdone?.()}
        title={host ? 'Drag onto another nation to move this player' : undefined}
      >
        <div class="who">
          <span class="pname">
            <span class="pname-text" title={p.player_name}>{p.player_name}</span>
            {#if p.is_host}<span class="host-mark" title="Host" aria-label="Host"><Icon name="crown" size={13} /></span>{/if}
            {#if p.note}<NoteBubble note={p.note} name={p.player_name} />{/if}
          </span>
        </div>
        {#if host || p.player_name === myName}
          <!-- your own seat always shows its buttons; host tools on others appear on hover -->
          <div class="seat-actions" class:own={p.player_name === myName}>
            {#if p.player_name === myName}
              <button
                type="button"
                class="seat-btn"
                aria-label={p.note ? 'Edit your note' : 'Add a note'}
                title={p.note ? 'Edit note' : 'Add note'}
                disabled={busy}
                onclick={onnote}
              >
                <Icon name="pencil" size={13} />
              </button>
            {/if}
            {#if p.player_name === myName}
              <button
                type="button"
                class="seat-btn kick"
                aria-label="Leave {nameOf(tag)}"
                title="Leave this nation"
                disabled={busy}
                onclick={onleave}
              >
                <Icon name="leave" size={14} />
              </button>
            {:else if host}
              <button
                type="button"
                class="seat-btn kick"
                aria-label="Kick {p.player_name}"
                title="Kick from lobby"
                disabled={busy}
                onclick={() => onkick?.(p)}
              >
                <Icon name="x" size={14} />
              </button>
            {/if}
          </div>
        {/if}
      </li>
    {/each}
    {#each { length: open } as _, i (i)}
      <li>
        {#if isMine}
          <!-- you already hold this nation; the spare co-op slot is for someone else -->
          <span class="seat empty"><span>Open slot</span></span>
        {:else}
          <button
            class="seat empty claimable"
            onclick={onclaim}
            disabled={busy}
            aria-label="{myTag ? 'Switch to' : 'Claim'} {nameOf(tag)}"
          >
            <span class="slot-label">Open slot</span>
          </button>
        {/if}
      </li>
    {/each}
  </ul>


</article>

<style>
  /* one nation = one row of its faction's list: name left, seats right */
  .card {
    position: relative;
    display: grid;
    grid-template-columns: 190px minmax(0, 1fr);
    gap: 0 16px;
    /* seats sit 5px inside the list on the right, top and bottom: with their
       6px corners that matches the list's 11px inner corner (12px - 1px border) */
    padding: 5px 5px 5px 14px;
    height: 100%;
    /* every nation row sits a little darker than the list panel */
    background: rgb(0 0 0 / 0.14);
  }
  /* your nation: gold edge with a soft inner glow, on a warmer tint */
  .card.mine {
    background: #2f2f21;
    box-shadow:
      inset 0 0 0 1px rgb(214 170 76 / 0.75),
      inset 0 0 18px rgb(214 170 76 / 0.16);
  }
  /* host drag and drop */
  .card.drop-ok {
    outline: 2px dashed var(--accent);
    outline-offset: -3px;
    background: rgb(214 170 76 / 0.08);
  }
  .card.drop-no {
    outline: 2px dashed var(--danger);
    outline-offset: -3px;
  }
  .card.drag-source {
    outline: 1px dashed var(--line-2);
    outline-offset: -3px;
  }
  header {
    grid-column: 1;
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
    /* centred on the first seat (40px tall) */
    height: 40px;
    align-self: start;
  }
  h3 {
    flex: 1;
    min-width: 0;
    font-size: 17px;
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .seats {
    grid-column: 2;
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 10px;
    flex: 1;
  }
  /* fixed height: rows are exactly 50px per slot, so the lists line up */
  .seat {
    display: flex;
    align-items: center;
    gap: 9px;
    height: 40px;
    padding: 0 10px;
    border-radius: 6px;
    font-size: 15px;
  }
  .seat.taken {
    position: relative;
    background: var(--panel-2);
  }
  .seat.movable {
    cursor: grab;
  }
  .seat.movable:active {
    cursor: grabbing;
  }
  .seat.dragged {
    opacity: 0.4;
  }
  /* solid (10% gold over the card) so the hover actions can match it */
  .seat.me {
    background: #383826;
  }
  .seat.empty {
    color: var(--dim);
    border: 1px dashed var(--line);
  }
  /* free seats you can take: gold, with Claim always shown, so they stand
     out when scanning a busy lobby (beats the grey .seat.empty look) */
  button.seat.claimable {
    width: 100%;
    align-items: center;
    color: var(--accent);
    border-color: rgb(214 170 76 / 0.5);
    background: rgb(214 170 76 / 0.04);
    font: inherit;
    font-size: 15px;
    text-align: left;
    cursor: pointer;
    transition:
      border-color 0.12s,
      background 0.12s,
      color 0.12s;
  }
  .slot-label {
    flex: 1;
  }
  /* plain highlight: the dashed outline fades out, the slot fills softly */
  button.claimable:hover:not(:disabled),
  button.claimable:focus-visible {
    border-color: transparent;
    background: rgb(214 170 76 / 0.16);
    color: var(--text);
  }
  button.claimable:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }
  button.claimable:disabled {
    cursor: progress;
  }
  .seat-actions {
    display: flex;
    gap: 2px;
    margin: -2px -4px -2px auto;
    flex-shrink: 0;
    opacity: 0;
    transition: opacity 0.12s;
  }
  .seat:hover .seat-actions,
  .seat:focus-within .seat-actions {
    opacity: 1;
  }
  /* with a mouse they take no room until the row is hovered or focused, so
     names aren't cut short by buttons you can't see (still tabbable) */
  @media (hover: hover) {
    .seat-actions {
      width: 0;
      overflow: hidden;
    }
    .seat:hover .seat-actions,
    .seat:focus-within .seat-actions {
      width: auto;
      overflow: visible;
    }
  }
  .seat .seat-actions.own {
    width: auto;
    overflow: visible;
    opacity: 1;
  }
  /* no hover on touch screens: always show them, in the row */
  @media (hover: none) {
    .seat-actions {
      opacity: 1;
    }
  }
  .seat-btn {
    display: grid;
    place-items: center;
    width: 26px;
    height: 26px;
    padding: 0;
    border: 1px solid transparent;
    border-radius: 5px;
    background: transparent;
    color: var(--muted);
    cursor: pointer;
    transition:
      color 0.12s,
      background 0.12s,
      border-color 0.12s;
  }
  .seat-btn:hover:not(:disabled) {
    color: var(--text);
    background: var(--panel);
    border-color: var(--line-2);
  }
  .seat-btn.kick:hover:not(:disabled) {
    color: var(--danger);
  }
  .seat-btn:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }
  .seat-btn:disabled {
    cursor: progress;
    opacity: 0.5;
  }
  .who {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
  }
  /* one line: the name truncates, the crown / YOU / note markers never wrap */
  .pname {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    font-size: 17px;
    font-weight: 500;
  }
  .pname-text {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .pname > :not(.pname-text) {
    flex-shrink: 0;
  }
  .host-mark {
    display: inline-flex;
    color: var(--accent);
  }
  .host-controls {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  @media (hover: hover) {
    .host-controls {
      display: none;
    }
    .card:hover .host-controls,
    .card:focus-within .host-controls {
      display: flex;
    }
    /* only where there are host controls to show instead */
    .card:hover header:has(.host-controls) h3,
    .card:focus-within header:has(.host-controls) h3 {
      display: none;
    }
  }
  /* touch: no hover, so the controls sit on their own line under the name */
  @media (hover: none) {
    header:has(.host-controls) {
      flex-wrap: wrap;
      height: auto;
      min-height: 40px;
      row-gap: 6px;
    }
    .host-controls {
      flex-basis: 100%;
    }
  }
  /* narrow: name above the seats */
  @media (max-width: 520px) {
    .card {
      grid-template-columns: minmax(0, 1fr);
    }
    header,
    .seats {
      grid-column: 1;
    }
  }
</style>
