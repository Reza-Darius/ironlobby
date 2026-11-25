// Finds URLs in user-written text (lobby descriptions) so they can be shown as
// links. The text is never rendered as HTML: links are always shown as the
// URL itself, so a description can't disguise where a link goes.

/**
 * Sites that open without a warning. A host matches when it equals an entry or
 * is a subdomain of it (www.discord.com matches discord.com).
 */
export const TRUSTED_HOSTS = [
  // voice / community
  'discord.gg',
  'discord.com',
  'discordapp.com',
  // Steam: workshop mods + collections (steamcommunity.com), store, s.team short links
  'steamcommunity.com',
  'steampowered.com',
  's.team',
  // Google: rules docs, sheets, forms (docs.google.com), Drive files, form short links
  'docs.google.com',
  'drive.google.com',
  'forms.gle',
  // Paradox
  'paradoxplaza.com',
  'paradoxwikis.com',
  // general
  'github.com',
  'reddit.com',
  'youtube.com',
  'youtu.be',
  'twitch.tv',
]

export type Part = { kind: 'text'; text: string } | { kind: 'link'; text: string; url: URL; trusted: boolean }

// http(s)://, www. or a bare Discord invite (discord.gg/abc), up to the next
// whitespace; trailing punctuation is trimmed below
const URL_RE = /\b(?:https?:\/\/|www\.|discord\.gg\/)[^\s<>"]*/gi
const TRAILING = /[.,;:!?'")\]}>]+$/

export function isTrusted(url: URL): boolean {
  // "https://evil.io@discord.com" reads like evil.io; never skip the warning for those
  if (url.username || url.password) return false
  const host = url.hostname.toLowerCase()
  return TRUSTED_HOSTS.some((t) => host === t || host.endsWith('.' + t))
}

const count = (s: string, ch: string) => s.split(ch).length - 1

function toUrl(raw: string): URL | null {
  try {
    const url = new URL(/^https?:\/\//i.test(raw) ? raw : `https://${raw}`)
    return url.protocol === 'http:' || url.protocol === 'https:' ? url : null
  } catch {
    return null
  }
}

/** Splits text into plain-text and link parts, in order. */
export function linkify(text: string): Part[] {
  const parts: Part[] = []
  let last = 0
  for (const m of text.matchAll(URL_RE)) {
    let raw = m[0]
    // "see discord.gg/x." / "(https://x.com)": the punctuation isn't part of the URL,
    // but keep a closing bracket when the URL itself opened one (wiki links)
    const trail = raw.match(TRAILING)?.[0] ?? ''
    if (trail) {
      raw = raw.slice(0, -trail.length)
      if (trail[0] === ')' && count(raw, '(') > count(raw, ')')) raw += ')'
    }
    const url = toUrl(raw)
    if (!url) continue
    const start = m.index!
    if (start > last) parts.push({ kind: 'text', text: text.slice(last, start) })
    parts.push({ kind: 'link', text: raw, url, trusted: isTrusted(url) })
    last = start + raw.length
  }
  if (last < text.length) parts.push({ kind: 'text', text: text.slice(last) })
  return parts
}
