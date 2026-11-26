<script lang="ts">
  import EnlistModal from './components/EnlistModal.svelte'
  import Icon from './components/Icon.svelte'
  import Logo from './components/Logo.svelte'
  import Toasts from './components/Toasts.svelte'
  import { askForCallsign, me } from './lib/identity.svelte'
  import { location_, match } from './lib/router.svelte'
  import { loadCountries } from './lib/state.svelte'
  import CreateLobby from './routes/CreateLobby.svelte'
  import Home from './routes/Home.svelte'
  import Lobby from './routes/Lobby.svelte'
  import Missing from './routes/Missing.svelte'

  const route = $derived(match(location_.path))

  loadCountries()
</script>

<header class="topbar">
  <div class="container inner">
    <a href="/" class="brand" aria-label="Iron Lobby home">
      <Logo />
      <span class="word">Iron Lobby</span>
    </a>
    <div class="row">
      {#if me.callsign}
        <button class="callsign" onclick={() => askForCallsign('Register a new name for this browser.')} title="Change name">
          {me.callsign}
        </button>
      {:else}
        <button class="btn sm" onclick={() => askForCallsign()}>Enlist</button>
      {/if}
    </div>
  </div>
</header>

<main class="container">
  {#if route.name === 'home'}
    <Home />
  {:else if route.name === 'new'}
    <CreateLobby />
  {:else if route.name === 'lobby'}
    {#key route.id}
      <Lobby id={route.id} />
    {/key}
  {:else}
    <Missing />
  {/if}
</main>

<footer class="container foot dim">
  <span>Iron Lobby · a lobby organiser for Hearts of Iron IV multiplayer. Not affiliated with Paradox Interactive.</span>
  <a class="repo" href="https://github.com/Reza-Darius/ironlobby" target="_blank" rel="noopener noreferrer">
    <Icon name="github" size={15} />Source on GitHub
  </a>
</footer>

<EnlistModal />
<Toasts />

<style>
  .topbar {
    position: sticky;
    top: 0;
    z-index: 10;
    background: var(--bg-2);
    border-bottom: 1px solid var(--line);
  }
  .inner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 58px;
    gap: 12px;
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 10px;
    color: var(--text);
  }
  .word {
    font-family: var(--logo);
    font-weight: 400;
    font-size: 18px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }
  .callsign {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    max-width: 200px;
    height: 30px;
    padding: 0 12px;
    border-radius: 999px;
    border: 1px solid var(--line-2);
    background: var(--panel);
    color: var(--text);
    font: inherit;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .callsign:hover {
    border-color: var(--accent);
  }
  main {
    padding-top: 28px;
    padding-bottom: 48px;
    min-height: calc(100dvh - 58px - 72px);
  }
  .foot {
    padding-top: 24px;
    padding-bottom: 28px;
    font-size: 13.5px;
    border-top: 1px solid var(--line);
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px 24px;
    flex-wrap: wrap;
  }
  .repo {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: var(--muted);
    transition: color 0.15s;
  }
  .repo:hover {
    color: var(--accent-2);
  }
</style>
