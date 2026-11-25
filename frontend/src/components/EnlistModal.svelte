<script lang="ts">
  import { ApiError } from '../lib/api'
  import { me, prompt, register, settlePrompt } from '../lib/identity.svelte'
  import Modal from './Modal.svelte'

  let name = $state('')
  let error = $state('')
  let busy = $state(false)

  $effect(() => {
    if (prompt.open) {
      name = me.callsign ?? ''
      error = ''
    }
  })

  async function submit(e: SubmitEvent) {
    e.preventDefault()
    const callsign = name.trim()
    if (callsign.length < 2) return (error = 'At least 2 characters, please.')
    if (callsign.length > 32) return (error = 'Keep it under 32 characters.')
    busy = true
    error = ''
    try {
      await register(callsign)
      settlePrompt(true)
    } catch (err) {
      error =
        err instanceof ApiError && err.status === 409
          ? `"${callsign}" is already taken. Try another one.`
          : (err as Error).message
    } finally {
      busy = false
    }
  }
</script>

<Modal
  bind:open={prompt.open}
  title="Pick your callsign"
  subtitle={prompt.reason || 'This is how other players will see you in lobbies.'}
  onclose={() => settlePrompt(false)}
>
  <form onsubmit={submit} class="stack">
    <label class="field">
      <span class="label">Callsign</span>
      <!-- svelte-ignore a11y_autofocus -->
      <input
        class="input"
        bind:value={name}
        placeholder="e.g. Feldmarschall_Kaiserschmarrn"
        maxlength="32"
        autocomplete="nickname"
        autofocus
        aria-invalid={!!error}
      />
      {#if error}<span class="error">{error}</span>
      {:else}<span class="hint">Use the same name you go by in Steam or Discord so the host recognises you.</span>{/if}
    </label>
    <div class="row end">
      <button type="button" class="btn ghost" onclick={() => (prompt.open = false)}>Cancel</button>
      <button class="btn primary" disabled={busy}>
        {busy ? 'Enlisting…' : 'Enlist'}
      </button>
    </div>
  </form>
</Modal>
