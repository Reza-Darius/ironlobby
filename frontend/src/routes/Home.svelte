<script lang="ts">
  import { onMount } from 'svelte'
  import Flag from '../components/Flag.svelte'
  import Icon from '../components/Icon.svelte'
  import Modal from '../components/Modal.svelte'
  import { api, ApiError } from '../lib/api'
  import { forgetLobby, isHost, me, patchLobby, type RecentLobby } from '../lib/identity.svelte'
  import { navigate, parseLobbyRef } from '../lib/router.svelte'
  import { clock, nameOf, toast, toastError } from '../lib/state.svelte'
  import type { LobbyPlayer } from '../lib/types'
  import { countdown, formatDate } from '../lib/time'

  let ref = $state('')
  let error = $state('')

  function join(e: SubmitEvent) {
    e.preventDefault()
    const id = parseLobbyRef(ref)
    if (id === null) return (error = 'Paste an invite link or a lobby number.')
    navigate(`/lobby/${id}`)
  }

  const upcoming = $derived(
    [...me.recent].sort((a, b) => {
      const now = clock.now
      const fa = new Date(a.starts_at).getTime() < now - 6 * 3600_000
      const fb = new Date(b.starts_at).getTime() < now - 6 * 3600_000
      if (fa !== fb) return fa ? 1 : -1
      return new Date(a.starts_at).getTime() - new Date(b.starts_at).getTime()
    }),
  )

  // Removing a lobby from the list also gives up your slot in it. We ask the
  // server whether you're actually seated rather than trusting the stored role.
  let removing = $state<{ lobby: RecentLobby; seat: LobbyPlayer | null } | null>(null)
  let removeOpen = $state(false)
  let removeBusy = $state(false)

  async function askRemove(lobby: RecentLobby) {
    if (lobby.role === 'visitor') return forgetLobby(lobby.id)
    let seat: LobbyPlayer | null = null
    try {
      const info = await api.lobby(lobby.id)
      seat = info.players?.find((p) => p.player_name === me.callsign) ?? null
    } catch {
      // lobby is gone (or unreachable): nothing to leave
      return forgetLobby(lobby.id)
    }
    removing = { lobby, seat }
    removeOpen = true
  }

  async function confirmRemove() {
    if (!removing) return
    const { lobby, seat } = removing
    removeBusy = true
    try {
      if (seat) {
        try {
          await api.leave(lobby.id)
        } catch (e) {
          // 404: already gone from the lobby, which is what we want anyway
          if (!(e instanceof ApiError && e.status === 404)) throw e
        }
      }
      forgetLobby(lobby.id)
      removeOpen = false
      toast(seat ? `You left ${lobby.name}.` : `${lobby.name} removed from your list.`)
    } catch (e) {
      toastError(e)
    } finally {
      removeBusy = false
    }
  }

  // Refresh the list from the server once on load: picks up nation switches,
  // renames and time changes made elsewhere since the lobby was last opened.
  onMount(() => {
    for (const r of me.recent) {
      if (r.role === 'visitor') continue
      api
        .lobby(r.id)
        .then((info) => {
          const seat = info.players?.find((p) => p.player_name === me.callsign) ?? null
          patchLobby(r.id, {
            name: info.lobby.lobby_name,
            starts_at: info.lobby.starts_at,
            tag: seat?.country_tag ?? null,
            role: isHost(info.lobby) ? 'host' : seat ? 'player' : 'visitor',
          })
        })
        .catch(() => {})
    }
  })

  const ROLE_LABEL = { host: 'Hosting', player: 'Playing', visitor: 'Viewed' } as const
</script>

<section class="modes">
  <a href="/new" class="mode panel host">
    <div class="mode-icon"><Icon name="crown" size={22} /></div>
    <div class="mode-body">
      <h2>Host a game</h2>
      <p class="muted">Pick the nations, the slots per nation, and when it starts. You get a link to send to everyone.</p>
    </div>
    <span class="btn primary">Create lobby</span>
  </a>

  <form class="mode panel" onsubmit={join}>
    <div class="mode-icon"><Icon name="flag" size={22} /></div>
    <div class="mode-body">
      <h2>Join a game</h2>
      <p class="muted">Got an invite? Paste the link or lobby number to see the open nations and claim one.</p>
    </div>
    <div class="join-row">
      <input
        class="input"
        bind:value={ref}
        oninput={() => (error = '')}
        placeholder="https://…/lobby/42 or 42"
        aria-label="Invite link or lobby number"
        aria-invalid={!!error}
      />
      <button class="btn primary">Join</button>
    </div>
    {#if error}<span class="error">{error}</span>{/if}
  </form>
</section>

{#if upcoming.length}
  <section class="recent">
    <h2 class="section-title">Your lobbies</h2>
    <ul>
      {#each upcoming as r (r.id)}
        {@const cd = countdown(r.starts_at, clock.now)}
        <li class="panel" class:past={cd.state === 'past'}>
          <span class="recent-flag">
            {#if r.tag}<Flag tag={r.tag} size="lg" />{/if}
          </span>
          <a href="/lobby/{r.id}" class="recent-link">
            <span class="recent-name">{r.name}</span>
            {#if r.tag}<span class="recent-nation">Playing {nameOf(r.tag)}</span>{/if}
            <span class="recent-meta muted">
              <Icon name="calendar" size={13} />
              {formatDate(r.starts_at)}
              <span class="dim">·</span>
              <span class:soon={cd.state === 'soon' || cd.state === 'live'}>{cd.text}</span>
            </span>
          </a>
          <span class="badge" class:accent={r.role === 'host'} class:ok={r.role === 'player'}>{ROLE_LABEL[r.role]}</span>
          <button
            class="btn icon sm ghost"
            aria-label={r.role === 'visitor' ? 'Remove from list' : 'Leave lobby'}
            title={r.role === 'visitor' ? 'Remove from list' : 'Leave lobby'}
            onclick={() => askRemove(r)}
          >
            <Icon name="x" size={14} />
          </button>
        </li>
      {/each}
    </ul>
  </section>
{/if}

<Modal
  bind:open={removeOpen}
  title={removing ? (removing.seat ? `Leave ${removing.lobby.name}?` : `Remove ${removing.lobby.name}?`) : ''}
>
  {#if removing}
    <div class="stack">
      <p class="muted">
        {#if removing.seat}
          You'll give up your slot on <strong>{nameOf(removing.seat.country_tag)}</strong> and the lobby disappears from
          your list.
        {/if}
        {#if removing.lobby.role === 'host'}
          {removing.seat ? 'As host, the lobby itself stays up for the other players.' : 'This only removes it from your list. The lobby stays up for the other players.'}
          To delete it for everyone, use Delete lobby under Edit on the lobby page.
        {/if}
      </p>
      <div class="row end">
        <button class="btn ghost" onclick={() => (removeOpen = false)}>Cancel</button>
        <button class="btn solid-danger" onclick={confirmRemove} disabled={removeBusy}>
          {removing.seat ? 'Leave lobby' : 'Remove'}
        </button>
      </div>
    </div>
  {/if}
</Modal>

<style>
  .modes {
    margin-top: 32px;
  }
  .modes {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 16px;
  }
  .mode {
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 22px;
    color: inherit;
    transition: border-color 0.15s;
  }
  .mode.host:hover {
    border-color: rgb(214 170 76 / 0.5);
  }
  .mode-icon {
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    border-radius: 10px;
    background: var(--panel-2);
    border: 1px solid var(--line-2);
    color: var(--accent);
  }
  .mode-body {
    flex: 1;
  }
  .mode h2 {
    font-size: 24px;
    text-transform: uppercase;
    margin-bottom: 6px;
  }
  .mode .btn {
    align-self: flex-start;
  }
  .join-row {
    display: flex;
    gap: 8px;
  }
  .section-title {
    font-size: 15px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--muted);
    margin-bottom: 12px;
  }
  .recent {
    margin-top: 40px;
  }
  .recent ul {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .recent li {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 8px 6px 4px;
    transition: border-color 0.15s;
  }
  .recent li:hover {
    border-color: rgb(214 170 76 / 0.5);
  }
  .recent li.past {
    opacity: 0.75;
  }
  .recent-link {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    padding: 6px 10px;
    color: inherit;
  }
  .recent-flag {
    display: grid;
    place-items: center;
    width: 36px;
    margin-left: 10px;
    flex-shrink: 0;
  }
  .recent-flag:empty {
    /* no nation: a faint empty frame keeps names aligned */
    height: 27px;
    border: 1px dashed var(--line-2);
    border-radius: 4px;
  }
  .recent-nation {
    font-size: 14px;
    color: var(--accent-2);
    font-weight: 500;
  }
  .recent-name {
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .recent-meta {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 14px;
    flex-wrap: wrap;
  }
  .soon {
    color: var(--live);
    font-weight: 600;
  }
</style>
