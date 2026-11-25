import type { Country, Gamemode, LobbyInfo } from './types'

const BASE: string = import.meta.env.VITE_API_BASE ?? '/api'

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
    /** true when the server sent no explanation (e.g. auth middleware rejections) */
    public bare = false,
  ) {
    super(message)
  }
}

function fallbackMessage(status: number): string {
  if (status === 401) return 'You need a callsign to do that.'
  if (status === 404) return 'Not found.'
  if (status === 409) return 'That already exists.'
  if (status >= 500) return 'The server ran into a problem. Try again in a moment.'
  return `Request failed (${status}).`
}

/** The API answers with plain-text errors; turn them into readable sentences. */
function humanize(text: string): string {
  const t = text.trim()
  return t.charAt(0).toUpperCase() + t.slice(1).replace(/doesnt/g, "doesn't").replace(/couldnt/g, "couldn't")
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  let res: Response
  try {
    res = await fetch(BASE + path, {
      method,
      credentials: 'include',
      headers: body === undefined ? undefined : { 'Content-Type': 'application/json' },
      body: body === undefined ? undefined : JSON.stringify(body),
    })
  } catch {
    throw new ApiError(0, "Can't reach the Iron Lobby server.")
  }

  const text = await res.text()
  if (!res.ok) {
    const bare = text.trim() === ''
    throw new ApiError(res.status, bare ? fallbackMessage(res.status) : humanize(text), bare)
  }
  if (!text) return undefined as T
  try {
    return JSON.parse(text) as T
  } catch {
    return text as T
  }
}

const lobbyPath = (id: number) => `/lobby/${id}`

export const api = {
  register: (username: string) => request<void>('POST', '/user', { username }),

  countries: () => request<Country[]>('GET', '/countries'),

  lobby: (id: number) => request<LobbyInfo>('GET', lobbyPath(id)),

  createLobby: (input: {
    lobby: { lobby_name: string; description: string; starts_at: string; gamemode: Gamemode }
    countries: { country_tag: string; max_slots: number }[]
  }) => request<{ lobby_id: number }>('POST', '/lobby', input),

  /** null / omitted fields are left unchanged by the server */
  updateLobby: (
    id: number,
    patch: Partial<{ lobby_name: string; description: string; starts_at: string; ingame_id: string }>,
  ) => request<void>('PATCH', lobbyPath(id), patch),

  addCountry: (id: number, country_tag: string, max_slots: number) =>
    request<void>('POST', `${lobbyPath(id)}/country`, { country_tag, max_slots }),

  setSlots: (id: number, tag: string, max_slots: number) =>
    request<void>('PATCH', `${lobbyPath(id)}/country/${tag}`, { max_slots }),

  /** Host only. Spec: docs/delete-lobby.md */
  deleteLobby: (id: number) => request<void>('DELETE', lobbyPath(id)),

  removeCountry: (id: number, tag: string) => request<void>('DELETE', `${lobbyPath(id)}/country/${tag}`),

  /** Idempotent: also used to switch nation or change the note. */
  claim: (id: number, country_tag: string, note: string) =>
    request<void>('POST', `${lobbyPath(id)}/player`, { country_tag, note: note.trim() || null }),

  leave: (id: number) => request<void>('DELETE', `${lobbyPath(id)}/player`),
}
