// Who is using this browser.
//
// The API identifies players by a UUID cookie set on POST /api/user, but offers
// no "who am I" endpoint, and in production the cookie is HttpOnly. So we keep
// the callsign and the lobbies this browser hosts in localStorage.

import { api, ApiError } from './api'
import type { Lobby } from './types'

const KEY = 'ironlobby.v1'

export type Role = 'host' | 'player' | 'visitor'

export interface RecentLobby {
  id: number
  name: string
  starts_at: string
  role: Role
}

interface Stored {
  callsign: string | null
  hosted: number[]
  recent: RecentLobby[]
}

function load(): Stored {
  try {
    const raw = localStorage.getItem(KEY)
    if (raw) return { callsign: null, hosted: [], recent: [], ...JSON.parse(raw) }
  } catch {
    /* storage unavailable */
  }
  return { callsign: null, hosted: [], recent: [] }
}

export const me = $state<Stored>(load())

function save() {
  try {
    localStorage.setItem(KEY, JSON.stringify(me))
  } catch {
    /* storage unavailable: state still works for this tab */
  }
}

/** The player UUID cookie. Only readable when the API runs with DEBUG_CORS=true. */
function cookiePlayerId(): string | null {
  const m = document.cookie.match(/(?:^|;\s*)username=([^;]+)/)
  return m ? decodeURIComponent(m[1]) : null
}

export function isHost(lobby: Lobby): boolean {
  return me.hosted.includes(lobby.id) || cookiePlayerId() === lobby.host_player
}

export function markHosted(id: number) {
  if (!me.hosted.includes(id)) me.hosted.push(id)
  save()
}

export function rememberLobby(entry: RecentLobby) {
  const rest = me.recent.filter((r) => r.id !== entry.id)
  me.recent = [entry, ...rest].slice(0, 12)
  save()
}

export function forgetLobby(id: number) {
  me.recent = me.recent.filter((r) => r.id !== id)
  save()
}

export async function register(callsign: string) {
  await api.register(callsign)
  me.callsign = callsign
  save()
}

// ---- callsign prompt -------------------------------------------------------
// Actions that need a player call `authed(fn)`. If we have no callsign, or the
// server rejects our cookie, the enlist dialog opens and the action resumes
// once the user has picked a name.

export const prompt = $state({ open: false, reason: '' })
let pending: ((ok: boolean) => void) | null = null

export function askForCallsign(reason = ''): Promise<boolean> {
  pending?.(false)
  prompt.reason = reason
  prompt.open = true
  return new Promise((resolve) => (pending = resolve))
}

export function settlePrompt(ok: boolean) {
  prompt.open = false
  pending?.(ok)
  pending = null
}

/** 401 = no cookie; a bare 400 = cookie points at a player the DB doesn't know. */
const isAuthFailure = (e: unknown) =>
  e instanceof ApiError && (e.status === 401 || (e.status === 400 && e.bare))

export async function authed<T>(action: () => Promise<T>, reason: string): Promise<T | undefined> {
  if (!me.callsign && !(await askForCallsign(reason))) return undefined
  try {
    return await action()
  } catch (e) {
    if (!isAuthFailure(e)) throw e
    me.callsign = null
    save()
    if (!(await askForCallsign('This browser lost its session. Pick a callsign to continue.'))) return undefined
    return await action()
  }
}
