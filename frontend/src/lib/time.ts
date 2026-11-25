const dateFmt = new Intl.DateTimeFormat(undefined, {
  weekday: 'short',
  day: 'numeric',
  month: 'short',
  hour: '2-digit',
  minute: '2-digit',
})

export const formatDate = (iso: string) => dateFmt.format(new Date(iso))

/** "in 2d 4h", "in 12m 30s", "started 3h ago" */
export function countdown(iso: string, now: number): { text: string; state: 'future' | 'soon' | 'live' | 'past' } {
  const diff = new Date(iso).getTime() - now
  const abs = Math.abs(diff)
  const s = Math.floor(abs / 1000)
  const d = Math.floor(s / 86400)
  const h = Math.floor((s % 86400) / 3600)
  const m = Math.floor((s % 3600) / 60)
  const sec = s % 60

  const span = d > 0 ? `${d}d ${h}h` : h > 0 ? `${h}h ${m}m` : m > 0 ? `${m}m ${sec}s` : `${sec}s`

  if (diff > 0) return { text: `in ${span}`, state: diff < 3600_000 ? 'soon' : 'future' }
  // HoI4 MP sessions run for hours; treat the first 6h as "in progress".
  if (abs < 6 * 3600_000) return { text: `started ${span} ago`, state: 'live' }
  return { text: `${span} ago`, state: 'past' }
}

const pad = (n: number) => String(n).padStart(2, '0')

/** ISO -> value for <input type="datetime-local"> in the viewer's timezone */
export function toLocalInput(iso: string | Date): string {
  const d = typeof iso === 'string' ? new Date(iso) : iso
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export const fromLocalInput = (v: string) => new Date(v).toISOString()

/** Next Saturday 19:00 local time: a sensible default for a game night. */
export function defaultStart(): string {
  const d = new Date()
  d.setDate(d.getDate() + ((6 - d.getDay() + 7) % 7 || 7))
  d.setHours(19, 0, 0, 0)
  return toLocalInput(d)
}

export const timezoneName = () => Intl.DateTimeFormat().resolvedOptions().timeZone
