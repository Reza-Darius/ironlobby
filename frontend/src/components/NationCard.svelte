<script lang="ts">
  import { nationColor } from '../lib/nations'
  import { nameOf } from '../lib/state.svelte'
  import type { LobbyPlayer } from '../lib/types'
  import Icon from './Icon.svelte'
  import Stepper from './Stepper.svelte'
  import Flag from './Flag.svelte'

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
  } = $props()

  const isMine = $derived(myTag === tag)
  const full = $derived(occupants.length >= slots)
  const open = $derived(Math.max(0, slots - occupants.length))
  const pct = $derived(Math.min(100, (occupants.length / slots) * 100))
</script>

<article class="card" class:mine={isMine} class:full={full && !isMine} style:--nation={nationColor(tag)}>
  <div class="fill" style:width="{pct}%"></div>

  <header>
    <Flag {tag} size="lg" />
    <h3>{nameOf(tag)}</h3>  </header>

  <ul class="seats">
    {#each occupants as p (p.player_name)}
      <li class="seat taken" class:me={p.player_name === myName}>
        <span class="pip"></span>
        <div class="who">
          <span class="pname">
            {p.player_name}
            {#if p.is_host}<span class="host-mark" title="Host" aria-label="Host"><Icon name="crown" size={13} /></span>{/if}
            {#if p.player_name === myName}<span class="you">you</span>{/if}
          </span>
          {#if p.note}<span class="note">“{p.note}”</span>{/if}
        </div>
      </li>
    {/each}
    {#each { length: open } as _, i (i)}
      <li>
        {#if isMine}
          <!-- you already hold this nation; the spare co-op slot is for someone else -->
          <span class="seat empty"><span class="pip"></span><span>Open slot</span></span>
        {:else}
          <button
            class="seat empty claimable"
            onclick={onclaim}
            disabled={busy}
            aria-label="{myTag ? 'Switch to' : 'Claim'} {nameOf(tag)}"
          >
            <span class="pip"></span>
            <span class="slot-label">Open slot</span>
            <span class="cta"><Icon name="back" size={13} />{myTag ? 'Switch here' : 'Claim'}</span>
          </button>
        {/if}
      </li>
    {/each}
  </ul>

  {#if isMine || full || host}
  <footer>
    {#if isMine}
      <button class="btn sm" onclick={onnote} disabled={busy}>
        <Icon name="note" size={13} />{occupants.find((p) => p.player_name === myName)?.note ? 'Edit note' : 'Add note'}
      </button>
      <button class="btn sm danger" onclick={onleave} disabled={busy}><Icon name="leave" size={13} />Leave</button>
    {:else if full}
      <span class="full-label dim">Full</span>
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
    align-items: flex-start;
    gap: 9px;
    padding: 6px 8px;
    border-radius: 6px;
    font-size: 15px;
  }
  .seat.taken {
    background: var(--panel-2);
  }
  .seat.me {
    background: rgb(214 170 76 / 0.1);
  }
  .seat.empty {
    color: var(--dim);
    border: 1px dashed var(--line);
  }
  button.claimable {
    width: 100%;
    align-items: center;
    background: transparent;
    font: inherit;
    font-size: 15px;
    text-align: left;
    cursor: pointer;
    transition:
      border-color 0.12s,
      background 0.12s,
      color 0.12s;
  }
  button.claimable .pip {
    margin-top: 0;
  }
  .slot-label {
    flex: 1;
  }
  .cta {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 13.5px;
    font-weight: 600;
    color: var(--accent);
    opacity: 0;
    transform: translateX(-4px);
    transition:
      opacity 0.12s,
      transform 0.12s;
  }
  button.claimable:hover:not(:disabled),
  button.claimable:focus-visible {
    border-style: solid;
    border-color: var(--accent);
    background: rgb(214 170 76 / 0.08);
    color: var(--text);
  }
  button.claimable:hover:not(:disabled) .pip {
    border-color: var(--accent);
  }
  button.claimable:hover:not(:disabled) .cta,
  button.claimable:focus-visible .cta {
    opacity: 1;
    transform: none;
  }
  button.claimable:disabled {
    cursor: progress;
  }
  /* touch screens have no hover: always show the hint */
  @media (hover: none) {
    .cta {
      opacity: 1;
      transform: none;
    }
  }
  .pip {
    width: 8px;
    height: 8px;
    margin-top: 6px;
    border-radius: 50%;
    flex-shrink: 0;
    background: var(--nation);
  }
  .empty .pip {
    background: transparent;
    border: 1px solid var(--line-2);
  }
  .who {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .pname {
    font-weight: 500;
    overflow-wrap: anywhere;
  }
  .host-mark {
    display: inline-flex;
    vertical-align: -2px;
    margin-left: 4px;
    color: var(--accent);
  }
  .you {
    margin-left: 4px;
    font-size: 12px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--accent);
  }
  .note {
    font-size: 14px;
    color: var(--muted);
    font-style: italic;
    overflow-wrap: anywhere;
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
  .full-label {
    font-size: 14px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }
</style>
