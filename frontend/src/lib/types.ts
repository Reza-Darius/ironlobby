// Mirrors the JSON shapes produced by the Go API (sqlc structs in api/internal/database).

export type Gamemode = 'vanilla' | 'modded' | 'rp'

export interface Country {
  country_tag: string
  country_name: string
}

export interface Lobby {
  id: number
  host_player: string
  lobby_name: string
  starts_at: string
  gamemode: Gamemode
  /** CHAR(17) in postgres, so it may come back space-padded */
  ingame_id: string | null
  description: string
}

export interface LobbyCountry {
  country_tag: string
  max_slots: number
}

export interface LobbyPlayer {
  player_name: string
  country_tag: string
  note: string | null
  joined_at: string
  is_host: boolean
}

/** Go encodes empty slices as null, so both lists are nullable. */
export interface LobbyInfo {
  lobby: Lobby
  countries: LobbyCountry[] | null
  players: LobbyPlayer[] | null
}

export const GAMEMODES: { value: Gamemode; label: string; blurb: string }[] = [
  { value: 'vanilla', label: 'Vanilla', blurb: 'Base game, standard rules' },
  { value: 'modded', label: 'Modded', blurb: 'Mods required to join' },
  { value: 'rp', label: 'Roleplay', blurb: 'Historical / RP rules' },
]

export const gamemodeLabel = (g: Gamemode) => GAMEMODES.find((m) => m.value === g)?.label ?? g
