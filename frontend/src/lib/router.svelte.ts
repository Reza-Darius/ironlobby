// A tiny history-API router. Three routes don't need a dependency.

export type Route = { name: 'home' } | { name: 'new' } | { name: 'lobby'; id: number } | { name: 'missing' }

export const location_ = $state({ path: window.location.pathname })

export function match(path: string): Route {
  if (path === '/' || path === '') return { name: 'home' }
  if (path === '/new') return { name: 'new' }
  const m = path.match(/^\/lobby\/(\d+)\/?$/)
  if (m) return { name: 'lobby', id: Number(m[1]) }
  return { name: 'missing' }
}

export function navigate(to: string, { replace = false } = {}) {
  const url = new URL(to, window.location.origin)
  if (url.pathname === location_.path) return
  // inApp marks entries reached by navigating inside the app, so goBack() knows
  // the previous history entry is one of our pages (and not another site)
  const inApp = replace ? !!history.state?.inApp : true
  history[replace ? 'replaceState' : 'pushState']({ inApp }, '', url)
  location_.path = url.pathname
  window.scrollTo({ top: 0 })
}

/** Back to the page the user came from; `fallback` when they landed here directly. */
export function goBack(fallback = '/') {
  if (history.state?.inApp) history.back()
  else navigate(fallback, { replace: true })
}

window.addEventListener('popstate', () => (location_.path = window.location.pathname))

// Intercept plain left-clicks on same-origin <a href> so links stay client-side.
document.addEventListener('click', (e) => {
  if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return
  const a = (e.target as Element | null)?.closest('a')
  if (!a || a.target || a.hasAttribute('download') || a.origin !== window.location.origin) return
  e.preventDefault()
  navigate(a.pathname + a.search)
})

/** Accepts "42", "#42" or any URL containing /lobby/42. */
export function parseLobbyRef(input: string): number | null {
  const s = input.trim()
  const m = s.match(/\/lobby\/(\d+)/) ?? s.match(/^#?(\d+)$/)
  return m ? Number(m[1]) : null
}

export const lobbyUrl = (id: number) => `${window.location.origin}/lobby/${id}`
