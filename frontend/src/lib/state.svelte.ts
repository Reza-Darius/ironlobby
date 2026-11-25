import { api } from './api'
import type { Country } from './types'

// ---- countries (static list, fetched once) ---------------------------------

export const countries = $state({
  list: [] as Country[],
  names: {} as Record<string, string>,
  error: '',
})

let loading: Promise<void> | null = null

export function loadCountries(): Promise<void> {
  loading ??= api
    .countries()
    .then((list) => {
      countries.list = [...list].sort((a, b) => a.country_name.localeCompare(b.country_name))
      countries.names = Object.fromEntries(list.map((c) => [c.country_tag, c.country_name]))
      countries.error = ''
    })
    .catch((e: Error) => {
      loading = null
      countries.error = e.message
    })
  return loading
}

export const nameOf = (tag: string) => countries.names[tag] ?? tag

// ---- clock (drives countdowns) ---------------------------------------------

export const clock = $state({ now: Date.now() })
setInterval(() => (clock.now = Date.now()), 1000)

// ---- toasts ----------------------------------------------------------------

export type ToastKind = 'ok' | 'error' | 'info'
export const toasts = $state<{ id: number; kind: ToastKind; text: string }[]>([])
let toastId = 0

export function toast(text: string, kind: ToastKind = 'ok') {
  const id = ++toastId
  toasts.push({ id, kind, text })
  setTimeout(() => dismiss(id), kind === 'error' ? 6000 : 3500)
}

export function dismiss(id: number) {
  const i = toasts.findIndex((t) => t.id === id)
  if (i >= 0) toasts.splice(i, 1)
}

export function toastError(e: unknown) {
  toast(e instanceof Error ? e.message : String(e), 'error')
}

export async function copy(text: string, what = 'Copied') {
  try {
    await navigator.clipboard.writeText(text)
    toast(`${what} to clipboard`, 'info')
  } catch {
    toast("Couldn't access the clipboard", 'error')
  }
}
