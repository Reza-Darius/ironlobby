<script lang="ts">
  import Icon from '../components/Icon.svelte'
  import { forgetLobby, me } from '../lib/identity.svelte'
  import { navigate, parseLobbyRef } from '../lib/router.svelte'
  import { clock } from '../lib/state.svelte'
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

  const ROLE_LABEL = { host: 'Hosting', player: 'Playing', visitor: 'Viewed' } as const
</script>

<section class="hero">
  <p class="kicker">Hearts of Iron IV · Multiplayer</p>
  <h1>Organise your game night<br /><span>before the war starts.</span></h1>
  <p class="lede muted">
    Host a lobby, set the start time and rules, open up nations, and share one link. Players claim their country, and
    the in-game ID goes out when it's time to launch.
  </p>
</section>

<section class="modes">
  <a href="/new" class="mode panel host">
    <div class="mode-icon"><Icon name="crown" size={22} /></div>
    <div class="mode-body">
      <h2>Host a game</h2>
      <p class="muted">Pick the nations, the slots per nation, and when it starts. You get a link to send to everyone.</p>
    </div>
    <span class="btn primary">Create lobby <Icon name="arrow" size={15} /></span>
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
      <button class="btn">Open</button>
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
          <a href="/lobby/{r.id}" class="recent-link">
            <span class="recent-name">{r.name}</span>
            <span class="recent-meta muted">
              <Icon name="calendar" size={13} />
              {formatDate(r.starts_at)}
              <span class="dim">·</span>
              <span class:soon={cd.state === 'soon' || cd.state === 'live'}>{cd.text}</span>
            </span>
          </a>
          <span class="badge" class:accent={r.role === 'host'} class:ok={r.role === 'player'}>{ROLE_LABEL[r.role]}</span>
          <button class="btn icon sm ghost" aria-label="Remove from list" title="Remove from list" onclick={() => forgetLobby(r.id)}>
            <Icon name="x" size={14} />
          </button>
        </li>
      {/each}
    </ul>
  </section>
{/if}

<style>
  .hero {
    padding: 28px 0 32px;
    max-width: 720px;
  }
  h1 {
    font-size: clamp(32px, 6vw, 52px);
    text-transform: uppercase;
    margin-top: 12px;
    letter-spacing: 0.005em;
  }
  h1 span {
    color: var(--accent);
  }
  .lede {
    margin-top: 16px;
    font-size: 16.5px;
    max-width: 600px;
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
    transition:
      border-color 0.15s,
      transform 0.15s;
  }
  .mode.host:hover {
    border-color: rgb(214 170 76 / 0.5);
    transform: translateY(-2px);
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
    transition: border-color 0.12s;
  }
  .recent li:hover {
    border-color: var(--line-2);
  }
  .recent li.past {
    opacity: 0.6;
  }
  .recent-link {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    padding: 6px 10px;
    color: inherit;
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
    font-size: 13px;
    flex-wrap: wrap;
  }
  .soon {
    color: var(--live);
    font-weight: 600;
  }
</style>
