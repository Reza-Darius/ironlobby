<script lang="ts">
  // Searchable combobox: type a tag or name, arrow keys + Enter to pick.
  import { countries } from '../lib/state.svelte'
  import Icon from './Icon.svelte'
  import Flag from './Flag.svelte'

  let {
    exclude = [],
    onpick,
    placeholder = 'Search nations: Germany, Japan, Soviet…',
  }: { exclude?: string[]; onpick: (tag: string) => void; placeholder?: string } = $props()

  let query = $state('')
  let open = $state(false)
  let active = $state(0)
  let listEl = $state<HTMLUListElement>()

  const options = $derived.by(() => {
    const q = query.trim().toLowerCase()
    const taken = new Set(exclude)
    const avail = countries.list.filter((c) => !taken.has(c.country_tag))
    if (!q) return avail
    // exact tag matches first, then prefix matches, then substring
    const score = (tag: string, name: string) =>
      tag.toLowerCase() === q ? 0 : name.toLowerCase().startsWith(q) || tag.toLowerCase().startsWith(q) ? 1 : 2
    return avail
      .filter((c) => c.country_tag.toLowerCase().includes(q) || c.country_name.toLowerCase().includes(q))
      .sort((a, b) => score(a.country_tag, a.country_name) - score(b.country_tag, b.country_name))
  })

  $effect(() => {
    void query
    active = 0
  })

  function pick(tag: string) {
    onpick(tag)
    query = ''
    active = 0
    // close after a pick; typing, clicking the field or ArrowDown reopens it
    open = false
  }

  function onkeydown(e: KeyboardEvent) {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      open = true
      active = Math.min(active + 1, options.length - 1)
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      active = Math.max(active - 1, 0)
    } else if (e.key === 'Enter') {
      e.preventDefault()
      if (open && options[active]) pick(options[active].country_tag)
    } else if (e.key === 'Escape') {
      open = false
    }
  }

  $effect(() => {
    listEl?.querySelector(`[data-i="${active}"]`)?.scrollIntoView({ block: 'nearest' })
  })
</script>

<div class="picker">
  <div class="input-wrap">
    <Icon name="search" />
    <input
      class="input"
      role="combobox"
      aria-expanded={open}
      aria-controls="nation-options"
      aria-autocomplete="list"
      bind:value={query}
      {placeholder}
      onfocus={() => (open = true)}
      onclick={() => (open = true)}
      onblur={() => setTimeout(() => (open = false), 120)}
      oninput={() => (open = true)}
      {onkeydown}
    />
  </div>
  {#if open}
    <ul class="options" id="nation-options" role="listbox" bind:this={listEl}>
      {#each options as c, i (c.country_tag)}
        <li
          role="option"
          aria-selected={i === active}
          data-i={i}
          class:active={i === active}
          onmousedown={(e) => {
            e.preventDefault()
            pick(c.country_tag)
          }}
          onmousemove={() => (active = i)}
        >
          <Flag tag={c.country_tag} size="sm" />
          <span>{c.country_name}</span>
        </li>
      {:else}
        <li class="empty">
          {countries.error || (countries.list.length ? 'No nations match.' : 'Loading nations…')}
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .picker {
    position: relative;
  }
  .input-wrap {
    position: relative;
  }
  .input-wrap :global(svg) {
    position: absolute;
    left: 11px;
    top: 50%;
    transform: translateY(-50%);
    color: var(--dim);
    pointer-events: none;
  }
  .input-wrap .input {
    padding-left: 34px;
  }
  .options {
    position: absolute;
    z-index: 20;
    top: calc(100% + 4px);
    left: 0;
    right: 0;
    max-height: 260px;
    overflow-y: auto;
    list-style: none;
    margin: 0;
    padding: 4px;
    background: var(--panel-2);
    border: 1px solid var(--line-2);
    border-radius: var(--radius);
    box-shadow: var(--shadow-lg);
  }
  li {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 7px 8px;
    border-radius: 5px;
    cursor: pointer;
    font-size: 14px;
  }
  li.active {
    background: var(--line);
  }
  li.empty {
    color: var(--muted);
    cursor: default;
  }
</style>
