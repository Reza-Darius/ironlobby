<script lang="ts">
  import { nationColor } from '../lib/nations'
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
    onmoveseat,
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
    onmoveseat?: (p: LobbyPlayer) => void
    onkick?: (p: LobbyPlayer) => void
  } = $props()

  const isMine = $derived(myTag === tag)
  const full = $derived(occupants.length >= slots)
  const open = $derived(Math.max(0, slots - occupants.length))
  const pct = $derived(Math.min(100, (occupants.length / slots) * 100))

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
  class:full={full && !isMine}
  class:drop-ok={over && canDrop}
  class:drop-no={over && !canDrop}
  class:drag-source={dragFromHere}
  style:--nation={nationColor(tag)}
  bind:this={card}
  {ondragover}
  {ondragleave}
  {ondrop}
  role="group"
  aria-label={nameOf(tag)}
>
  <div class="fill" style:width="{pct}%"></div>

  <header>
    <Flag {tag} size="lg" />
    <h3>{nameOf(tag)}</h3>  </header>

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
            {p.player_name}
            {#if p.is_host}<span class="host-mark" title="Host" aria-label="Host"><Icon name="crown" size={13} /></span>{/if}
            {#if p.player_name === myName}<span class="you">you</span>{/if}
            {#if p.note}<NoteBubble note={p.note} name={p.player_name} />{/if}
          </span>
        </div>
        {#if host}
          <div class="seat-actions">
            <button
              type="button"
              class="seat-btn"
              aria-label="Move {p.player_name} to another nation"
              title="Move to another nation"
              disabled={busy}
              onclick={() => onmoveseat?.(p)}
            >
              <Icon name="move" size={14} />
            </button>
            {#if p.player_name !== myName}
              <!-- your own seat has Leave in the footer -->
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

  {#if isMine || host}
  <footer>
    {#if isMine}
      <button class="btn sm" onclick={onnote} disabled={busy}>
        <Icon name="note" size={13} />{occupants.find((p) => p.player_name === myName)?.note ? 'Edit note' : 'Add note'}
      </button>
      <button class="btn sm danger" onclick={onleave} disabled={busy}><Icon name="leave" size={13} />Leave</button>
    {/if}

    {#if host}
      <span class="spacer"></span>
      <Stepper value={slots} min={Math.max(1, occupants.length)} disabled={busy} onchange={onslots} />
      <button class="btn icon sm ghost danger" aria-label="Remove {nameOf(tag)} from lobby" title="Remove nation" onclick={onremove} disabled={busy}>
        <Icon name="trash" size={14} />
      </button>
    {/if}
  </footer>
  {/if}
</article>

<style>
  .card {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 14px;
    background: var(--panel);
    border: 1px solid var(--line);
    border-radius: var(--radius-lg);
    overflow: hidden;
    transition: border-color 0.15s;
  }
  .card:hover {
    border-color: var(--line-2);
  }
  .card.mine {
    border-color: var(--accent);
    box-shadow: 0 0 0 1px rgb(214 170 76 / 0.25) inset;
  }
  .card.full {
    background: var(--bg-2);
  }
  /* host drag and drop */
  .card.drop-ok {
    border-color: var(--accent);
    border-style: dashed;
    background: rgb(214 170 76 / 0.06);
  }
  .card.drop-no {
    border-color: var(--danger);
    border-style: dashed;
  }
  .card.drag-source {
    border-style: dashed;
  }
  .fill {
    position: absolute;
    top: 0;
    left: 0;
    height: 3px;
    background: var(--nation);
    transition: width 0.3s ease;
  }
  header {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
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
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
    flex: 1;
  }
  .seat {
    display: flex;
    align-items: center;
    gap: 9px;
    min-height: 38px;
    padding: 6px 10px;
    border-radius: 6px;
    font-size: 15px;
  }
  .seat.taken {
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
  .seat.me {
    background: rgb(214 170 76 / 0.1);
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
  /* no hover on touch screens: always show them */
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
    min-width: 0;
  }
  .pname {
    display: inline-flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0 6px;
    font-size: 17px;
    font-weight: 500;
    overflow-wrap: anywhere;
  }
  .host-mark {
    display: inline-flex;
    color: var(--accent);
  }
  .you {
    font-size: 12px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--accent);
  }
  footer {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
    min-height: 30px;
  }
  .spacer {
    flex: 1;
  }
</style>
