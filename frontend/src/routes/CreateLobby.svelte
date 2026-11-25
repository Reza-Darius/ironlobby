<script lang="ts">
  import { flip } from 'svelte/animate'
  import { fade } from 'svelte/transition'
  import Icon from '../components/Icon.svelte'
  import NationPicker from '../components/NationPicker.svelte'
  import DateTimePicker from '../components/DateTimePicker.svelte'
  import Stepper from '../components/Stepper.svelte'
  import Flag from '../components/Flag.svelte'
  import { api } from '../lib/api'
  import { authed, markHosted, rememberLobby } from '../lib/identity.svelte'
  import { PRESETS } from '../lib/nations'
  import { navigate } from '../lib/router.svelte'
  import { countries, nameOf, toast, toastError } from '../lib/state.svelte'
  import { defaultStart, fromLocalInput, timezoneName } from '../lib/time'
  import { GAMEMODES, type Gamemode } from '../lib/types'

  let name = $state('')
  let description = $state('')
  let gamemode = $state<Gamemode>('vanilla')
  let startsAt = $state(defaultStart())
  let nations = $state<{ tag: string; slots: number }[]>([])

  let errors = $state<Record<string, string>>({})
  let busy = $state(false)

  const totalSlots = $derived(nations.reduce((n, c) => n + c.slots, 0))
  const known = $derived(new Set(countries.list.map((c) => c.country_tag)))

  function add(tag: string) {
    if (!nations.some((n) => n.tag === tag)) nations.push({ tag, slots: 1 })
    delete errors.nations
  }

  function addPreset(tags: string[]) {
    for (const t of tags) if (known.size === 0 || known.has(t)) add(t)
  }

  function validate(): boolean {
    const e: Record<string, string> = {}
    if (name.trim().length < 3) e.name = 'Give the lobby a name (3+ characters).'
    if (!startsAt) e.startsAt = 'Pick a start time.'
    else if (new Date(startsAt).getTime() < Date.now()) e.startsAt = 'The start time is in the past.'
    if (description.length > 1000) e.description = 'Keep the description under 1000 characters.'
    if (nations.length === 0) e.nations = 'Open at least one nation for players to claim.'
    errors = e
    return Object.keys(e).length === 0
  }

  async function submit(e: SubmitEvent) {
    e.preventDefault()
    if (!validate()) {
      document.querySelector<HTMLElement>('[aria-invalid="true"], .nations-error')?.focus()
      return
    }
    busy = true
    try {
      const res = await authed(
        () =>
          api.createLobby({
            lobby: {
              lobby_name: name.trim(),
              description: description.trim(),
              gamemode,
              starts_at: fromLocalInput(startsAt),
            },
            countries: nations.map((n) => ({ country_tag: n.tag, max_slots: n.slots })),
          }),
        'Hosts need a callsign so players know who is running the game.',
      )
      if (!res) return
      markHosted(res.lobby_id)
      rememberLobby({ id: res.lobby_id, name: name.trim(), starts_at: fromLocalInput(startsAt), role: 'host' })
      toast('Lobby created. Share the invite link with your players.')
      navigate(`/lobby/${res.lobby_id}`, { replace: true })
    } catch (err) {
      toastError(err)
    } finally {
      busy = false
    }
  }
</script>

<a href="/" class="back btn ghost sm"><Icon name="back" size={14} />Back</a>

<header class="head">
  <h1>Host a game</h1>
</header>

<form class="layout" onsubmit={submit} novalidate>
  <section class="panel card">
    <h2 class="card-title"><span class="num">1</span>Briefing</h2>

    <label class="field">
      <span class="label">Lobby name</span>
      <input
        class="input"
        bind:value={name}
        placeholder="Sunday Historical: 1936 start"
        maxlength="80"
        aria-invalid={!!errors.name}
      />
      {#if errors.name}<span class="error">{errors.name}</span>{/if}
    </label>

    <fieldset class="field">
      <legend class="label">Game mode</legend>
      <div class="segmented" role="radiogroup">
        {#each GAMEMODES as m (m.value)}
          <label class="seg" class:on={gamemode === m.value}>
            <input type="radio" name="gamemode" value={m.value} bind:group={gamemode} />
            <span class="seg-label">{m.label}</span>
            <span class="seg-blurb">{m.blurb}</span>
          </label>
        {/each}
      </div>
    </fieldset>

    <div class="field">
      <span class="label">Start time</span>
      <DateTimePicker bind:value={startsAt} invalid={!!errors.startsAt} />
      {#if errors.startsAt}<span class="error">{errors.startsAt}</span>
      {:else}<span class="hint">Your local time ({timezoneName()}). Players see it in theirs.</span>{/if}
    </div>

    <label class="field">
      <span class="label">Description <span class="dim">(optional)</span></span>
      <textarea
        class="textarea"
        bind:value={description}
        rows="5"
        placeholder="Historical focuses on, no gentleman's agreement breaks before 1939, Discord voice required, mods: …"
        aria-invalid={!!errors.description}
      ></textarea>
      <span class="hint counter" class:over={description.length > 1000}>{description.length} / 1000</span>
      {#if errors.description}<span class="error">{errors.description}</span>{/if}
    </label>
  </section>

  <section class="panel card">
    <h2 class="card-title"><span class="num">2</span>Nations</h2>
    <p class="muted small">Players can only claim the nations you open. Use slots for co-op (e.g. 2 players on SOV).</p>

    <div class="presets">
      <span class="dim small">Quick add:</span>
      {#each PRESETS as p (p.label)}
        <button type="button" class="btn sm" onclick={() => addPreset(p.tags)}>
          <Icon name="plus" size={13} />{p.label}
        </button>
      {/each}
    </div>

    <NationPicker exclude={nations.map((n) => n.tag)} onpick={add} />

    {#if nations.length}
      <ul class="nation-list">
        {#each nations as n, i (n.tag)}
          <li animate:flip={{ duration: 180 }} transition:fade={{ duration: 120 }}>
            <Flag tag={n.tag} />
            <span class="nation-name">{nameOf(n.tag)}</span>
            <Stepper value={n.slots} onchange={(v) => (nations[i].slots = v)} />
            <button
              type="button"
              class="btn icon sm ghost"
              aria-label="Remove {nameOf(n.tag)}"
              onclick={() => nations.splice(i, 1)}
            >
              <Icon name="x" size={14} />
            </button>
          </li>
        {/each}
      </ul>
      <div class="list-footer">
        <button type="button" class="btn ghost sm" onclick={() => (nations = [])}>Clear all</button>
        <span class="badge">{nations.length} nations · {totalSlots} slots</span>
      </div>
    {:else}
      <div class="empty" class:invalid={!!errors.nations}>
        <Icon name="flag" size={20} />
        <span>{errors.nations ?? 'No nations yet. Search above or use a quick-add preset.'}</span>
      </div>
    {/if}
  </section>

  <div class="submit-bar">
    <button class="btn primary lg" disabled={busy}>
      {busy ? 'Creating…' : 'Create lobby'}
      {#if !busy}<Icon name="arrow" size={16} />{/if}
    </button>
  </div>
</form>

<style>
  .back {
    margin-left: -10px;
    margin-bottom: 12px;
  }
  .head {
    margin-bottom: 24px;
  }
  .head h1 {
    font-size: clamp(30px, 5vw, 42px);
    text-transform: uppercase;
    margin-top: 6px;
  }
  .layout {
    display: grid;
    grid-template-columns: 1fr;
    gap: 16px;
    align-items: start;
  }
  @media (min-width: 900px) {
    .layout {
      grid-template-columns: 1fr 1fr;
    }
    .submit-bar {
      grid-column: 1 / -1;
    }
  }
  .card {
    padding: 22px;
    display: flex;
    flex-direction: column;
    gap: 18px;
  }
  .card-title {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 19px;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }
  .num {
    display: grid;
    place-items: center;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    border: 1px solid var(--accent);
    color: var(--accent);
    font-family: var(--mono);
    font-size: 13px;
  }
  .small {
    font-size: 14.5px;
  }
  fieldset {
    border: 0;
    padding: 0;
    margin: 0;
  }
  legend {
    margin-bottom: 6px;
  }
  .segmented {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 6px;
  }
  .seg {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 10px 12px;
    border: 1px solid var(--line-2);
    border-radius: var(--radius);
    background: var(--bg-2);
    cursor: pointer;
    transition:
      border-color 0.12s,
      background 0.12s;
  }
  .seg:hover {
    border-color: var(--line-hover);
  }
  .seg.on {
    border-color: var(--accent);
    background: rgb(214 170 76 / 0.08);
  }
  .seg:has(input:focus-visible) {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .seg input {
    position: absolute;
    opacity: 0;
    pointer-events: none;
  }
  .seg-label {
    font-weight: 600;
    font-size: 15px;
  }
  .seg.on .seg-label {
    color: var(--accent-2);
  }
  .seg-blurb {
    font-size: 13px;
    color: var(--dim);
    line-height: 1.3;
  }
  @media (max-width: 480px) {
    .seg-blurb {
      display: none;
    }
  }
  .counter {
    align-self: flex-end;
    font-family: var(--mono);
  }
  .counter.over {
    color: var(--danger);
  }
  .presets {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
  }
  .nation-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .nation-list li {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 6px 6px 8px;
    border: 1px solid var(--line);
    border-radius: var(--radius);
    background: var(--bg-2);
  }
  .nation-name {
    flex: 1;
    min-width: 0;
    font-size: 15.5px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .list-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    margin-top: -8px;
  }
  .empty {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 18px;
    border: 1px dashed var(--line-2);
    border-radius: var(--radius);
    color: var(--dim);
    font-size: 15px;
  }
  .empty.invalid {
    border-color: var(--danger);
    color: var(--danger);
  }
  .submit-bar {
    position: sticky;
    bottom: 0;
    display: flex;
    justify-content: flex-end;
    padding: 14px 0;
  }
  @media (max-width: 600px) {
    .submit-bar .btn {
      width: 100%;
    }
  }
</style>
