<script lang="ts">
  // Date + time picker. The value uses the <input type="datetime-local"> format
  // ("YYYY-MM-DDTHH:mm", viewer's timezone), so it's a drop-in replacement.
  // The calendar is a native popover: it sits in the top layer (never clipped by
  // a surrounding dialog) and the browser closes it on outside click / Escape.
  import Icon from './Icon.svelte'

  let {
    value = $bindable(''),
    invalid = false,
    allowPast = false,
  }: { value?: string; invalid?: boolean; allowPast?: boolean } = $props()

  const pad = (n: number) => String(n).padStart(2, '0')
  const dayKey = (d: Date) => `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`

  const current = $derived(value ? new Date(value) : null)
  const valid = $derived(!!current && !isNaN(current.getTime()))

  const uid = $props.id()
  const panelId = `dtp-${uid}`
  let open = $state(false)
  let trigger = $state<HTMLButtonElement>()
  let panel = $state<HTMLDivElement>()
  // month on screen, independent of the selected date
  let view = $state(new Date())

  const GAP = 6
  const EDGE = 8

  /** Below the field, or above it when there's no room below; kept on screen. */
  function place(measure = true) {
    if (!trigger || !panel) return
    const r = trigger.getBoundingClientRect()
    let top = r.bottom + GAP
    let left = r.left
    if (measure) {
      const h = panel.offsetHeight
      const w = panel.offsetWidth
      if (top + h > innerHeight - EDGE && r.top - GAP - h > EDGE) top = r.top - GAP - h
      left = Math.max(EDGE, Math.min(left, innerWidth - w - EDGE))
    }
    panel.style.top = `${top}px`
    panel.style.left = `${left}px`
  }

  function onbeforetoggle(e: Event) {
    if ((e as ToggleEvent).newState !== 'open') return
    view = valid ? new Date(current!) : new Date()
    place(false) // can't measure while hidden; refined in ontoggle
  }

  function ontoggle(e: Event) {
    open = (e as ToggleEvent).newState === 'open'
    if (open) place()
  }

  // follow the field while the page (or a dialog) scrolls
  $effect(() => {
    if (!open) return
    const follow = () => place()
    addEventListener('scroll', follow, true)
    addEventListener('resize', follow)
    return () => {
      removeEventListener('scroll', follow, true)
      removeEventListener('resize', follow)
    }
  })

  // ---- calendar grid -------------------------------------------------------

  const locale = navigator.language
  // first day of the week from the locale where supported, else Monday
  const firstDay = (() => {
    try {
      const l = new Intl.Locale(locale) as Intl.Locale & { getWeekInfo?: () => { firstDay: number } }
      return (l.getWeekInfo?.().firstDay ?? 1) % 7
    } catch {
      return 1
    }
  })()

  const monthFmt = new Intl.DateTimeFormat(locale, { month: 'long', year: 'numeric' })
  const fieldFmt = new Intl.DateTimeFormat(locale, { weekday: 'short', day: 'numeric', month: 'short', year: 'numeric' })
  // 24h to match the hour select
  const timeFmt = new Intl.DateTimeFormat(locale, { hour: '2-digit', minute: '2-digit', hourCycle: 'h23' })
  const weekdays = Array.from({ length: 7 }, (_, i) =>
    new Intl.DateTimeFormat(locale, { weekday: 'narrow' }).format(new Date(2024, 0, 7 + ((firstDay + i) % 7))),
  )

  const todayKey = dayKey(new Date())

  const days = $derived.by(() => {
    const y = view.getFullYear()
    const m = view.getMonth()
    const lead = (new Date(y, m, 1).getDay() - firstDay + 7) % 7
    const start = new Date(y, m, 1 - lead)
    // always 6 rows so the panel height doesn't jump between months
    return Array.from({ length: 42 }, (_, i) => new Date(start.getFullYear(), start.getMonth(), start.getDate() + i))
  })

  function shiftMonth(by: number) {
    view = new Date(view.getFullYear(), view.getMonth() + by, 1)
  }

  // ---- time ------------------------------------------------------------------

  const hours = Array.from({ length: 24 }, (_, i) => i)
  const minutes = $derived.by(() => {
    const steps = Array.from({ length: 12 }, (_, i) => i * 5)
    // keep an odd existing minute (e.g. 19:07) selectable
    const m = valid ? current!.getMinutes() : 0
    return steps.includes(m) ? steps : [...steps, m].sort((a, b) => a - b)
  })

  function set(date: Date, h: number, min: number) {
    value = `${dayKey(date)}T${pad(h)}:${pad(min)}`
  }

  function pickDay(d: Date) {
    // new dates default to 19:00, a typical game night start
    set(d, valid ? current!.getHours() : 19, valid ? current!.getMinutes() : 0)
    if (d.getMonth() !== view.getMonth()) view = new Date(d.getFullYear(), d.getMonth(), 1)
  }

  function setHour(h: number) {
    set(valid ? current! : new Date(), h, valid ? current!.getMinutes() : 0)
  }
  function setMinute(min: number) {
    set(valid ? current! : new Date(), valid ? current!.getHours() : 19, min)
  }
</script>

<div class="dtp">
  <button
    type="button"
    class="input trigger"
    class:open
    class:invalid
    aria-haspopup="dialog"
    popovertarget={panelId}
    bind:this={trigger}
  >
    <Icon name="calendar" size={16} />
    {#if valid}
      <span>{fieldFmt.format(current!)}</span>
      <span class="sep">·</span>
      <span>{timeFmt.format(current!)}</span>
    {:else}
      <span class="placeholder">Pick a date and time</span>
    {/if}
  </button>

  <div
    class="panel"
    id={panelId}
    popover="auto"
    role="dialog"
    aria-label="Choose start date and time"
    bind:this={panel}
    {onbeforetoggle}
    {ontoggle}
  >
      <div class="month">
        <button type="button" class="btn icon sm ghost" aria-label="Previous month" onclick={() => shiftMonth(-1)}>
          <Icon name="back" size={15} />
        </button>
        <span class="month-name">{monthFmt.format(view)}</span>
        <button type="button" class="btn icon sm ghost" aria-label="Next month" onclick={() => shiftMonth(1)}>
          <Icon name="arrow" size={15} />
        </button>
      </div>

      <div class="grid" role="grid">
        {#each weekdays as w, i (i)}<span class="wd" aria-hidden="true">{w}</span>{/each}
        {#each days as d (d.getTime())}
          {@const key = dayKey(d)}
          <button
            type="button"
            class="day"
            class:outside={d.getMonth() !== view.getMonth()}
            class:today={key === todayKey}
            class:selected={valid && key === dayKey(current!)}
            disabled={!allowPast && key < todayKey}
            aria-label={fieldFmt.format(d)}
            aria-pressed={valid && key === dayKey(current!)}
            onclick={() => pickDay(d)}
          >
            {d.getDate()}
          </button>
        {/each}
      </div>

      <div class="time">
        <Icon name="clock" size={15} />
        <select
          class="input"
          aria-label="Hour"
          value={valid ? current!.getHours() : 19}
          onchange={(e) => setHour(+e.currentTarget.value)}
        >
          {#each hours as h (h)}<option value={h}>{pad(h)}</option>{/each}
        </select>
        <span class="colon">:</span>
        <select
          class="input"
          aria-label="Minute"
          value={valid ? current!.getMinutes() : 0}
          onchange={(e) => setMinute(+e.currentTarget.value)}
        >
          {#each minutes as m (m)}<option value={m}>{pad(m)}</option>{/each}
        </select>
        <button type="button" class="btn sm done" onclick={() => panel?.hidePopover()}>Done</button>
      </div>
  </div>
</div>

<style>
  .dtp {
    display: flex;
    flex-direction: column;
  }
  .trigger {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    text-align: left;
    cursor: pointer;
    color: var(--text);
  }
  .trigger :global(svg) {
    color: var(--muted);
  }
  .trigger.open {
    border-color: var(--accent);
  }
  .trigger.invalid {
    border-color: var(--danger);
  }
  .sep {
    color: var(--dim);
  }
  .placeholder {
    color: var(--dim);
  }

  .panel {
    /* undo the UA popover centring; place() sets top/left */
    position: fixed;
    inset: auto;
    margin: 0;
    width: min(320px, calc(100vw - 16px));
    padding: 14px;
    background-color: var(--panel);
    color: var(--text);
    border: 1px solid var(--line-2);
    border-radius: var(--radius-lg);
    box-shadow: var(--frame), var(--shadow-lg);
  }
  .month {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 10px;
  }
  .month-name {
    font-family: var(--display);
    font-weight: 600;
    font-size: 16px;
    letter-spacing: 0.02em;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(7, 1fr);
    gap: 2px;
  }
  .wd {
    text-align: center;
    font-size: 13px;
    font-weight: 600;
    color: var(--muted);
    padding: 4px 0;
  }
  .day {
    aspect-ratio: 1;
    display: grid;
    place-items: center;
    border: 1px solid transparent;
    border-radius: var(--radius);
    background: transparent;
    color: var(--text);
    font: inherit;
    font-size: 15px;
    cursor: pointer;
    transition:
      background 0.12s,
      border-color 0.12s;
  }
  .day:hover:not(:disabled):not(.selected) {
    background: var(--panel-2);
  }
  .day.outside {
    color: var(--muted);
  }
  .day.today {
    border-color: var(--line-2);
  }
  .day.selected {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--accent-ink);
    font-weight: 600;
  }
  .day:disabled {
    color: var(--dim);
    opacity: 0.45;
    cursor: not-allowed;
  }
  .day:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }

  .time {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid var(--line);
  }
  .time :global(svg) {
    color: var(--muted);
    margin-right: 4px;
  }
  .time select {
    width: auto;
    padding-right: 8px;
    font-family: var(--mono);
  }
  .colon {
    color: var(--muted);
  }
  .done {
    margin-left: auto;
  }
</style>
