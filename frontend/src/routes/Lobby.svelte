<script lang="ts">
  import Icon from '../components/Icon.svelte'
  import Modal from '../components/Modal.svelte'
  import NationCard from '../components/NationCard.svelte'
  import NationPicker from '../components/NationPicker.svelte'
  import DateTimePicker from '../components/DateTimePicker.svelte'
  import Flag from '../components/Flag.svelte'
  import LinkedText from '../components/LinkedText.svelte'
  import { api, ApiError } from '../lib/api'
  import { authed, forgetLobby, isHost, me, rememberLobby, type Role } from '../lib/identity.svelte'
  import { BLOCS, isMajor, nationBloc, type Bloc } from '../lib/nations'
  import { lobbyUrl, navigate } from '../lib/router.svelte'
  import { clock, copy, nameOf, toast, toastError } from '../lib/state.svelte'
  import { countdown, formatDate, fromLocalInput, timezoneName, toLocalInput } from '../lib/time'
  import { gamemodeLabel, type LobbyInfo, type LobbyPlayer } from '../lib/types'

  let { id }: { id: number } = $props()

  // ---- data --------------------------------------------------------------

  let info = $state<LobbyInfo | null>(null)
  let loadError = $state<{ status: number; message: string } | null>(null)
  let busy = $state(false)

  async function refresh(silent = false) {
    try {
      info = await api.lobby(id)
      loadError = null
    } catch (e) {
      // keep showing stale data on background-poll failures
      if (!silent || !info) loadError = { status: e instanceof ApiError ? e.status : 0, message: (e as Error).message }
    }
  }

  $effect(() => {
    refresh()
    // Lobbies are planned ahead and change slowly; poll while the tab is visible.
    const timer = setInterval(() => document.visibilityState === 'visible' && refresh(true), 10_000)
    const onVisible = () => document.visibilityState === 'visible' && refresh(true)
    document.addEventListener('visibilitychange', onVisible)
    return () => {
      clearInterval(timer)
      document.removeEventListener('visibilitychange', onVisible)
    }
  })

  const lobby = $derived(info?.lobby)
  const host = $derived(lobby ? isHost(lobby) : false)
  const players = $derived(info?.players ?? [])
  const mine = $derived(me.callsign ? players.find((p) => p.player_name === me.callsign) : undefined)
  const gameId = $derived(lobby?.ingame_id?.trim() ?? '')
  const cd = $derived(lobby ? countdown(lobby.starts_at, clock.now) : null)

  const seats = $derived.by(() => {
    const m = new Map<string, LobbyPlayer[]>()
    for (const p of [...players].sort((a, b) => a.joined_at.localeCompare(b.joined_at))) {
      m.set(p.country_tag, [...(m.get(p.country_tag) ?? []), p])
    }
    return m
  })

  const BLOC_ORDER: Bloc[] = ['allies', 'axis', 'comintern', 'other']
  const groups = $derived.by(() => {
    const list = info?.countries ?? []
    return BLOC_ORDER.map((bloc) => ({
      bloc,
      players: list
        .filter((c) => nationBloc(c.country_tag) === bloc)
        .reduce((n, c) => n + (seats.get(c.country_tag)?.length ?? 0), 0),
      slots: list.filter((c) => nationBloc(c.country_tag) === bloc).reduce((n, c) => n + c.max_slots, 0),
      nations: list
        .filter((c) => nationBloc(c.country_tag) === bloc)
        // majors first, then alphabetical. Only static facts: cards must not
        // jump around when slots or players change
        .sort(
          (a, b) =>
            Number(isMajor(b.country_tag)) - Number(isMajor(a.country_tag)) ||
            nameOf(a.country_tag).localeCompare(nameOf(b.country_tag)),
        ),
    })).filter((g) => g.nations.length)
  })

  const openNations = $derived((info?.countries ?? []).map((c) => c.country_tag))

  // Keep "Your lobbies" on the home page current.
  $effect(() => {
    if (!lobby) return
    const role: Role = host ? 'host' : mine ? 'player' : 'visitor'
    const tag = mine?.country_tag ?? null
    const prev = me.recent.find((r) => r.id === lobby.id)
    const count = players.length
    const slots = (info?.countries ?? []).reduce((n, c) => n + c.max_slots, 0)
    if (
      !prev ||
      prev.role !== role ||
      prev.tag !== tag ||
      prev.name !== lobby.lobby_name ||
      prev.starts_at !== lobby.starts_at ||
      prev.players !== count ||
      prev.slots !== slots
    ) {
      rememberLobby({ id: lobby.id, name: lobby.lobby_name, starts_at: lobby.starts_at, role, tag, players: count, slots })
    }
  })

  // ---- actions -----------------------------------------------------------

  async function run(action: () => Promise<unknown>, reason: string, success?: string) {
    busy = true
    try {
      const done = await authed(async () => {
        await action()
        return true
      }, reason)
      if (done && success) toast(success)
    } catch (e) {
      toastError(e)
    } finally {
      await refresh(true)
      busy = false
    }
  }

  // claim / switch: one click, no dialog. The API upserts tag + note together,
  // so switching carries the current note over; it can be edited afterwards.
  function claimSlot(tag: string) {
    run(
      () => api.claim(id, tag, mine?.note ?? ''),
      'Pick a name so the host knows who claimed this nation.',
      `You're now playing ${nameOf(tag)}.`,
    )
  }

  // note: opt-in, edited from your own nation's card
  let noteOpen = $state(false)
  let noteDraft = $state('')

  function openNote() {
    noteDraft = mine?.note ?? ''
    noteOpen = true
  }

  async function saveNote(e: SubmitEvent) {
    e.preventDefault()
    if (!mine) return
    const tag = mine.country_tag
    noteOpen = false
    await run(() => api.claim(id, tag, noteDraft), '', noteDraft.trim() ? 'Note saved.' : 'Note removed.')
  }

  // confirmations
  let confirm = $state<{ title: string; body: string; label: string; action: () => void } | null>(null)
  let confirmOpen = $state(false)
  function ask(c: NonNullable<typeof confirm>) {
    confirm = c
    confirmOpen = true
  }

  function leave() {
    if (!mine) return
    ask({
      title: `Leave ${nameOf(mine.country_tag)}?`,
      body: 'Your slot opens up for someone else. You can claim a nation again any time before the game.',
      label: 'Leave slot',
      action: () => run(() => api.leave(id), '', 'You left your slot.'),
    })
  }

  // host: nations
  let addingNation = $state(false)

  function setSlots(tag: string, slots: number) {
    const c = info?.countries?.find((c) => c.country_tag === tag)
    if (c) c.max_slots = slots // optimistic
    run(() => api.setSlots(id, tag, slots), '')
  }

  function removeNation(tag: string) {
    const taken = seats.get(tag)?.length ?? 0
    ask({
      title: `Remove ${nameOf(tag)}?`,
      body: taken
        ? `${taken === 1 ? '1 player is' : `${taken} players are`} on this nation and will be removed from the lobby.`
        : 'Players will no longer be able to claim it.',
      label: 'Remove nation',
      action: () => run(() => api.removeCountry(id, tag), '', `${nameOf(tag)} removed.`),
    })
  }

  function addNation(tag: string) {
    addingNation = false // back to the "Add nation" button after each pick
    run(() => api.addCountry(id, tag, 1), '', `${nameOf(tag)} added.`)
  }

  // host: players. Move by drag and drop (or the Move button), kick with ×.
  let dragging = $state<LobbyPlayer | null>(null)
  let moveTarget = $state<LobbyPlayer | null>(null)
  let moveOpen = $state(false)

  // every open nation in page order, with what's free on it
  const moveChoices = $derived(
    groups.flatMap((g) =>
      g.nations.map((c) => ({
        tag: c.country_tag,
        taken: seats.get(c.country_tag)?.length ?? 0,
        slots: c.max_slots,
      })),
    ),
  )

  function movePlayer(p: LobbyPlayer, tag: string) {
    dragging = null
    moveOpen = false
    if (p.country_tag === tag) return
    const name = p.player_name
    // optimistic: the seat jumps right away; a refused move snaps back on refresh
    const live = info?.players?.find((x) => x.player_name === name)
    if (live) live.country_tag = tag
    run(() => api.movePlayer(id, name, tag), '', `${name} moved to ${nameOf(tag)}.`)
  }

  function openMove(p: LobbyPlayer) {
    moveTarget = p
    moveOpen = true
  }

  function kickPlayer(p: LobbyPlayer) {
    const name = p.player_name
    ask({
      title: `Kick ${name}?`,
      body: `${name} loses their slot on ${nameOf(p.country_tag)}. They can still claim an open slot again.`,
      label: 'Kick player',
      action: () => run(() => api.kickPlayer(id, name), '', `${name} was kicked.`),
    })
  }

  // host: game id
  let editingGameId = $state(false)
  let gameIdDraft = $state('')

  function startGameIdEdit() {
    gameIdDraft = gameId
    editingGameId = true
  }

  async function saveGameId(e: SubmitEvent) {
    e.preventDefault()
    const v = gameIdDraft.trim()
    if (!v) return
    editingGameId = false
    await run(() => api.updateLobby(id, { ingame_id: v }), '', 'Game ID posted. Players can see it now.')
  }

  // host: edit details
  let editOpen = $state(false)
  let draft = $state({ name: '', description: '', startsAt: '' })
  let draftError = $state('')

  function openEdit() {
    if (!lobby) return
    draft = { name: lobby.lobby_name, description: lobby.description, startsAt: toLocalInput(lobby.starts_at) }
    draftError = ''
    editOpen = true
  }

  async function saveEdit(e: SubmitEvent) {
    e.preventDefault()
    if (!lobby) return
    if (draft.name.trim().length < 3) return (draftError = 'The name needs at least 3 characters.')
    if (draft.description.length > 1000) return (draftError = 'Keep the description under 1000 characters.')
    const patch: Parameters<typeof api.updateLobby>[1] = {}
    if (draft.name.trim() !== lobby.lobby_name) patch.lobby_name = draft.name.trim()
    if (draft.description.trim() !== lobby.description) patch.description = draft.description.trim()
    if (draft.startsAt && draft.startsAt !== toLocalInput(lobby.starts_at)) patch.starts_at = fromLocalInput(draft.startsAt)
    editOpen = false
    if (Object.keys(patch).length) await run(() => api.updateLobby(id, patch), '', 'Lobby updated.')
  }

  // host: delete. Not via run(): there's no lobby left to refresh afterwards.
  function deleteLobby() {
    if (!lobby) return
    editOpen = false
    const name = lobby.lobby_name
    ask({
      title: `Delete ${name}?`,
      body: "Every player loses their slot and the invite link stops working. This can't be undone.",
      label: 'Delete lobby',
      action: async () => {
        busy = true
        try {
          const done = await authed(async () => {
            await api.deleteLobby(id)
            return true
          }, '')
          if (!done) return
          forgetLobby(id)
          toast(`${name} deleted.`)
          navigate('/', { replace: true })
        } catch (e) {
          toastError(e)
        } finally {
          busy = false
        }
      },
    })
  }
</script>

{#if loadError && !info}
  <div class="state-msg">
    <p class="kicker">Lobby #{id}</p>
    <h1>{loadError.status === 404 || loadError.status === 500 ? 'Lobby not found' : "Couldn't load lobby"}</h1>
    <p class="muted">
      {loadError.status === 404 || loadError.status === 500
        ? 'Double-check the invite link. The lobby may have been removed.'
        : loadError.message}
    </p>
    <div class="row">
      <button class="btn" onclick={() => refresh()}>Try again</button>
      <a href="/" class="btn ghost">Back home</a>
    </div>
  </div>
{:else if !info || !lobby || !cd}
  <div class="loading" aria-busy="true" aria-label="Loading lobby">
    <div class="skeleton" style="height: 18px; width: 140px"></div>
    <div class="skeleton" style="height: 44px; width: min(480px, 90%)"></div>
    <div class="skeleton" style="height: 20px; width: 300px"></div>
    <div class="grid" style="margin-top: 24px">
      {#each { length: 6 } as _, i (i)}<div class="skeleton" style="height: 150px"></div>{/each}
    </div>
  </div>
{:else}
  <header class="lobby-head">
    <h1>{lobby.lobby_name}</h1>
    <div class="head-actions">
      <button class="btn" onclick={() => copy(lobbyUrl(id), 'Invite link copied')}>
        <Icon name="link" size={15} />Copy invite
      </button>
      {#if host}
        <button class="btn" onclick={openEdit}><Icon name="pencil" size={14} />Edit</button>
      {/if}
    </div>
  </header>

  <div class="main">
      {#if groups.length === 0}
        <div class="empty panel">
          <Icon name="flag" size={22} />
          <p>{host ? 'No nations open yet. Add some below so players can sign up.' : "The host hasn't opened any nations yet."}</p>
        </div>
      {/if}

      {#each groups as g (g.bloc)}
        <section class="bloc">
          <h2 class="bloc-title" style:--bloc={BLOCS[g.bloc].color}>
            <span class="bloc-dot"></span>{BLOCS[g.bloc].label}
            <span class="bloc-count" title="Players / slots">{g.players}/{g.slots}</span>
          </h2>
          <div class="grid">
            {#each g.nations as c (c.country_tag)}
              <div>
                <NationCard
                  tag={c.country_tag}
                  slots={c.max_slots}
                  occupants={seats.get(c.country_tag) ?? []}
                  myName={me.callsign}
                  myTag={mine?.country_tag ?? null}
                  {host}
                  {busy}
                  onclaim={() => claimSlot(c.country_tag)}
                  onnote={openNote}
                  onleave={leave}
                  onslots={(v) => setSlots(c.country_tag, v)}
                  onremove={() => removeNation(c.country_tag)}
                  {dragging}
                  ondragseat={(p) => (dragging = p)}
                  ondragdone={() => (dragging = null)}
                  ondropseat={() => dragging && movePlayer(dragging, c.country_tag)}
                  onmoveseat={openMove}
                  onkick={kickPlayer}
                />
              </div>
            {/each}
          </div>
        </section>
      {/each}

      {#if host}
        <section class="add-nation panel">
          {#if addingNation}
            <div class="add-row">
              <div class="add-picker"><NationPicker exclude={openNations} onpick={addNation} focusOnMount /></div>
              <button class="btn ghost" onclick={() => (addingNation = false)}>Cancel</button>
            </div>
            <p class="hint">New nations start with 1 slot. Use the +/− on a card to change it.</p>
          {:else}
            <button class="btn block ghost add-btn" onclick={() => (addingNation = true)}>
              <Icon name="plus" size={15} />Add nation
            </button>
          {/if}
        </section>
      {/if}
  </div>

  <!-- everything about the game, below the nations -->
  <div class="info">
    <div class="info-col">
      <section class="panel side-card">
        <h2 class="side-title"><Icon name="calendar" size={15} />Details</h2>
        <dl class="details">
          <div>
            <dt>Starts</dt>
            <dd>{formatDate(lobby.starts_at)}</dd>
          </div>
          <div>
            <dt>Countdown</dt>
            <dd>
              <span class="badge" class:live={cd.state === 'soon' || cd.state === 'live'}>
                {#if cd.state === 'live'}<span class="live-dot"></span>{:else}<Icon name="clock" size={12} />{/if}
                {cd.text}
              </span>
            </dd>
          </div>
          <div>
            <dt>Game mode</dt>
            <dd>{gamemodeLabel(lobby.gamemode)}</dd>
          </div>
          {#if host}
            <div>
              <dt>Host</dt>
              <dd><span class="badge accent"><Icon name="crown" size={12} />You're hosting</span></dd>
            </div>
          {/if}
        </dl>
      </section>

      <!-- in-game ID -->
      <section class="panel side-card launch" class:ready={!!gameId}>
        <h2 class="side-title"><Icon name="radio" size={15} />In-game lobby ID</h2>
        {#if editingGameId}
          <form class="stack" onsubmit={saveGameId}>
            <!-- svelte-ignore a11y_autofocus -->
            <input
              class="input mono"
              bind:value={gameIdDraft}
              maxlength="17"
              placeholder="Paste the game ID"
              aria-label="In-game lobby ID"
              autofocus
            />
            <div class="row">
              <button class="btn primary sm" disabled={!gameIdDraft.trim() || busy}>Post ID</button>
              <button type="button" class="btn ghost sm" onclick={() => (editingGameId = false)}>Cancel</button>
            </div>
          </form>
        {:else if gameId}
          <button class="game-id mono" onclick={() => copy(gameId, 'Game ID copied')} title="Copy game ID">
            {gameId}<Icon name="copy" size={16} />
          </button>
          <!-- button label from the game's own text (multiplayer_l_english: CONNECT_TO_IP) -->
          <p class="hint">In HoI4, go to Multiplayer → <strong>Connect to Server ID</strong> and paste it there.</p>
          {#if host}<button class="btn ghost sm" onclick={startGameIdEdit}><Icon name="pencil" size={13} />Change</button>{/if}
        {:else if host}
          <p class="muted small">
            Host the game in HoI4, then post its <strong>Server ID</strong> here so players can connect. Right-click
            the Server ID in the game lobby to copy it.
          </p>
          <button class="btn primary sm" onclick={startGameIdEdit}>Post game ID</button>
        {:else}
          <p class="muted small">The host will post the ID here when the game is up. Check back near the start time.</p>
        {/if}
      </section>
    </div>

    {#if lobby.description}
      <div class="info-col">
        <section class="panel side-card">
          <h2 class="side-title">Description</h2>
          <div class="description"><LinkedText text={lobby.description} /></div>
        </section>
      </div>
    {/if}
  </div>
{/if}

<!-- note -->
<Modal bind:open={noteOpen} title={mine?.note ? 'Edit your note' : 'Add a note'} subtitle="Visible to everyone in the lobby.">
  {#if mine}
    <form class="stack" onsubmit={saveNote}>
      <div class="claim-nation">
        <Flag tag={mine.country_tag} size="lg" />
        <span>{nameOf(mine.country_tag)}</span>
      </div>
      <label class="field">
        <span class="label">Note</span>
        <!-- svelte-ignore a11y_autofocus -->
        <input
          class="input"
          bind:value={noteDraft}
          maxlength="140"
          placeholder="e.g. first time as this nation, can play until 23:00"
          autofocus
        />
        <span class="hint">Leave empty to remove it.</span>
      </label>
      <div class="row end">
        <button type="button" class="btn ghost" onclick={() => (noteOpen = false)}>Cancel</button>
        <button class="btn primary">Save note</button>
      </div>
    </form>
  {/if}
</Modal>

<!-- host: move a player (also the touch / keyboard path for drag and drop) -->
<Modal
  bind:open={moveOpen}
  title={moveTarget ? `Move ${moveTarget.player_name}` : ''}
  subtitle="Pick a nation with a free slot. Their note stays."
>
  {#if moveTarget}
    {@const from = moveTarget.country_tag}
    <div class="stack">
      <ul class="move-list">
        {#each moveChoices as m (m.tag)}
          {@const current = m.tag === from}
          {@const isFull = m.taken >= m.slots}
          <li>
            <button
              type="button"
              class="move-opt"
              class:current
              disabled={current || isFull || busy}
              onclick={() => moveTarget && movePlayer(moveTarget, m.tag)}
            >
              <Flag tag={m.tag} size="sm" />
              <span class="m-name">{nameOf(m.tag)}</span>
              <span class="m-count">
                {#if current}Current{:else if isFull}Full{:else}{m.taken} / {m.slots}{/if}
              </span>
            </button>
          </li>
        {/each}
      </ul>
      <div class="row end">
        <button type="button" class="btn ghost" onclick={() => (moveOpen = false)}>Cancel</button>
      </div>
    </div>
  {/if}
</Modal>

<!-- confirm -->
<Modal bind:open={confirmOpen} title={confirm?.title ?? ''}>
  {#if confirm}
    <div class="stack">
      <p class="muted">{confirm.body}</p>
      <div class="row end">
        <button class="btn ghost" onclick={() => (confirmOpen = false)}>Cancel</button>
        <button
          class="btn solid-danger"
          onclick={() => {
            confirmOpen = false
            confirm?.action()
          }}>{confirm.label}</button
        >
      </div>
    </div>
  {/if}
</Modal>

<!-- edit lobby -->
<Modal bind:open={editOpen} title="Edit lobby">
  <form class="stack" onsubmit={saveEdit}>
    <label class="field">
      <span class="label">Lobby name</span>
      <input class="input" bind:value={draft.name} maxlength="80" />
    </label>
    <div class="field">
      <span class="label">Start time</span>
      <DateTimePicker bind:value={draft.startsAt} />
      <span class="hint">Your local time ({timezoneName()}).</span>
    </div>
    <label class="field">
      <span class="label">Description</span>
      <textarea class="textarea" rows="5" bind:value={draft.description}></textarea>
      <span class="hint">{draft.description.length} / 1000 · Game mode can't be changed after creation.</span>
    </label>
    {#if draftError}<span class="error">{draftError}</span>{/if}
    <div class="row end">
      <button type="button" class="btn ghost danger delete-lobby" onclick={deleteLobby}>
        <Icon name="trash" size={14} />Delete lobby
      </button>
      <button type="button" class="btn ghost" onclick={() => (editOpen = false)}>Cancel</button>
      <button class="btn primary">Save changes</button>
    </div>
  </form>
</Modal>

<style>
  /* the game's own UI labels stand out in the launch hints */
  .launch strong {
    color: var(--text);
    font-weight: 600;
  }
  .move-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
    max-height: min(420px, 55vh);
    overflow-y: auto;
  }
  .move-opt {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    padding: 9px 12px;
    border: 1px solid var(--line);
    border-radius: var(--radius);
    background: var(--panel-2);
    color: var(--text);
    font: inherit;
    font-size: 15px;
    text-align: left;
    cursor: pointer;
    transition:
      border-color 0.12s,
      background 0.12s;
  }
  .move-opt:hover:not(:disabled) {
    border-color: var(--accent);
    background: rgb(214 170 76 / 0.08);
  }
  .move-opt:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }
  .move-opt:disabled {
    cursor: default;
    color: var(--muted);
    background: transparent;
  }
  .move-opt.current {
    border-color: var(--line-2);
  }
  .m-name {
    flex: 1;
    min-width: 0;
  }
  .m-count {
    font-size: 13px;
    font-family: var(--mono);
    color: var(--muted);
  }
  .state-msg,
  .loading {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
    padding: 32px 0;
  }
  .state-msg h1 {
    font-size: 36px;
    text-transform: uppercase;
  }

  .lobby-head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 16px 24px;
    padding-bottom: 22px;
    margin-bottom: 22px;
    border-bottom: 1px solid var(--line);
  }
  h1 {
    font-size: clamp(28px, 5vw, 44px);
    text-transform: uppercase;
    overflow-wrap: anywhere;
  }
  .details {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin: 0;
  }
  .details > div {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }
  .details dt {
    color: var(--muted);
  }
  .details dd {
    margin: 0;
    font-weight: 500;
    text-align: right;
  }
  .live-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--live);
    animation: pulse 1.4s infinite;
  }
  @keyframes pulse {
    50% {
      opacity: 0.3;
    }
  }
  .head-actions {
    display: flex;
    gap: 8px;
  }

  /* below the nations: Details + Server ID | Description */
  .info {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 12px;
    align-items: start;
    margin-top: 32px;
    padding-top: 28px;
    border-top: 1px solid var(--line);
  }
  .info-col {
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-width: 0;
  }

  .main {
    display: flex;
    flex-direction: column;
    gap: 24px;
    min-width: 0;
  }

  .bloc-title {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 20px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    margin-bottom: 12px;
    color: var(--text);
  }
  /* same face as the heading, lighter weight so the label leads */
  .bloc-count {
    margin-left: 6px;
    font-weight: 500;
    letter-spacing: 0.04em;
    font-variant-numeric: tabular-nums;
    color: var(--muted);
  }
  .bloc-title .bloc-dot {
    width: 13px;
    height: 13px;
  }
  .bloc-dot {
    width: 10px;
    height: 10px;
    border-radius: 2px;
    background: var(--bloc);
  }
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 12px;
  }
  .grid > div {
    display: flex;
    flex-direction: column;
  }
  .grid > div > :global(*) {
    flex: 1;
  }

  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    padding: 32px 16px;
    color: var(--dim);
    text-align: center;
  }
  .add-nation {
    padding: 12px;
    border-style: dashed;
  }
  .add-btn {
    color: var(--muted);
  }
  .add-row {
    display: flex;
    gap: 8px;
  }
  .add-picker {
    flex: 1;
    min-width: 0;
  }
  .add-nation .hint {
    margin-top: 8px;
  }

  .side-card {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 16px;
  }
  .side-card .btn {
    align-self: flex-start;
  }
  .side-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    text-transform: uppercase;
    letter-spacing: 0.12em;
    color: var(--muted);
  }
  .launch.ready {
    border-color: rgb(134 180 111 / 0.45);
    background: #26301f;
  }
  .launch.ready .side-title {
    color: var(--ok);
  }
  .game-id {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    width: 100%;
    padding: 12px 14px;
    background: var(--bg-2);
    border: 1px solid var(--line-2);
    border-radius: var(--radius);
    color: var(--text);
    font-size: 18px;
    font-weight: 600;
    letter-spacing: 0.04em;
    cursor: pointer;
    overflow-wrap: anywhere;
    text-align: left;
  }
  .game-id:hover {
    border-color: var(--ok);
  }
  .game-id :global(svg) {
    color: var(--muted);
  }
  .small {
    font-size: 14.5px;
  }
  .description {
    white-space: pre-wrap;
    overflow-wrap: anywhere;
    font-size: 15px;
    color: var(--text);
  }

  .claim-nation {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 17px;
    font-weight: 500;
  }
  /* destructive action sits apart from Cancel / Save */
  .btn.delete-lobby {
    margin-right: auto;
  }
</style>
