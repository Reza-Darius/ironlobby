// Flag images per HoI4 tag, matching what the game shows at the 1936 start.
//
// Flags whose design hasn't changed since 1936 come from `flag-icons` (MIT);
// only the files in the glob are bundled. Everything that looked different in
// 1936 is drawn below as simplified SVG. Germany follows HoI4's own substitute
// design (Kriegsmarine layout with Iron Crosses instead of the swastika).

const modern = import.meta.glob(
  '/node_modules/flag-icons/flags/4x3/{jp,gb,fr,pl,au,nz,nl,be,lu,tw,no,dk,se,fi,pt,tr,cz,mx,br,ar,cl,th,ph,ro,bg,cn,mn}.svg',
  { eager: true, query: '?url', import: 'default' },
) as Record<string, string>

const SAME_SINCE_1936: Record<string, string> = {
  JAP: 'jp', ENG: 'gb', FRA: 'fr', POL: 'pl', AST: 'au', NZL: 'nz', HOL: 'nl',
  BEL: 'be', LUX: 'lu', CHI: 'tw', NOR: 'no', DEN: 'dk', SWE: 'se', FIN: 'fi',
  POR: 'pt', TUR: 'tr', CZE: 'cz', MEX: 'mx', BRA: 'br', ARG: 'ar', CHL: 'cl',
  THA: 'th', PHI: 'ph', ROM: 'ro', BUL: 'bg',
  PRC: 'cn', // unverified against the game
  MON: 'mn', // unverified against the game
}

// ---- drawing helpers (640×480 canvas) --------------------------------------

function star(cx: number, cy: number, r: number, fill: string, points = 5, inner = 0.382) {
  const pts: string[] = []
  for (let i = 0; i < points * 2; i++) {
    const rad = i % 2 ? r * inner : r
    const a = -Math.PI / 2 + (i * Math.PI) / points
    pts.push(`${(cx + rad * Math.cos(a)).toFixed(1)},${(cy + rad * Math.sin(a)).toFixed(1)}`)
  }
  return `<polygon points="${pts.join(' ')}" fill="${fill}"/>`
}

const hStripes = (...colors: string[]) => {
  const h = 480 / colors.length
  return colors.map((c, i) => `<rect y="${i * h}" width="640" height="${h}" fill="${c}"/>`).join('')
}

const vStripes = (...colors: string[]) => {
  const w = 640 / colors.length
  return colors.map((c, i) => `<rect x="${i * w}" width="${w}" height="480" fill="${c}"/>`).join('')
}

/** Union Jack filling a 320×240 box at the origin. */
const unionJack = `
  <rect width="320" height="240" fill="#012169"/>
  <path d="M0 0L320 240M320 0L0 240" stroke="#fff" stroke-width="48"/>
  <path d="M0 0L320 240M320 0L0 240" stroke="#c8102e" stroke-width="16"/>
  <path d="M160 0v240M0 120h320" stroke="#fff" stroke-width="80"/>
  <path d="M160 0v240M0 120h320" stroke="#c8102e" stroke-width="48"/>`

/** Iron Cross (cross pattée), centred on the origin, ~140px across. */
const ironCross = `<path d="M-12-12L-30-70H30L12-12L70-30V30L12 12L30 70H-30L-12 12L-70 30V-30Z"
  fill="#000" stroke="#fff" stroke-width="7" paint-order="stroke" stroke-linejoin="miter"/>`

const usaStars = (() => {
  let s = ''
  for (let row = 0; row < 6; row++)
    for (let col = 0; col < 8; col++) s += star(18 + col * 31, 20 + row * 44, 11, '#fff')
  return s
})()

// ---- 1936 flags ------------------------------------------------------------

const HISTORIC: Record<string, string> = {
  // HoI4's Germany: red field, white-edged black cross, white disc and Iron Crosses.
  GER: `<rect width="640" height="480" fill="#d40000"/>
    <path d="M230 0v480M0 240h640" stroke="#fff" stroke-width="54"/>
    <path d="M230 0v480M0 240h640" stroke="#000" stroke-width="26"/>
    <circle cx="230" cy="240" r="100" fill="#fff" stroke="#000" stroke-width="7"/>
    <g transform="translate(230 240)">${ironCross}</g>
    <g transform="translate(92 98) scale(.62)">${ironCross}</g>`,

  // Kingdom of Italy: tricolour with the crowned Savoy shield.
  ITA: `${vStripes('#009246', '#fff', '#ce2b37')}
    <path d="M276 186l9-34 17 20 18-30 18 30 17-20 9 34z" fill="#f4c430" stroke="#9a7b1c" stroke-width="3"/>
    <path d="M262 192h116v74c0 44-28 66-58 76c-30-10-58-32-58-76z" fill="#0047ab"/>
    <path d="M274 204h92v62c0 34-22 52-46 60c-24-8-46-26-46-60z" fill="#ce2b37"/>
    <path d="M311 204h18v122h-18zM274 246h92v18h-92z" fill="#fff"/>`,

  // 48-star flag (1912–1959).
  USA: `${hStripes(...Array.from({ length: 13 }, (_, i) => (i % 2 ? '#fff' : '#b22234')))}
    <rect width="256" height="258.5" fill="#3c3b6e"/>${usaStars}`,

  // Canadian Red Ensign (1921–1957).
  CAN: `<rect width="640" height="480" fill="#c8102e"/>${unionJack}
    <path d="M430 220h120v110c0 40-30 60-60 70c-30-10-60-30-60-70z" fill="#fff" stroke="#f4c430" stroke-width="6"/>
    <rect x="433" y="223" width="57" height="40" fill="#c8102e"/><rect x="490" y="223" width="57" height="40" fill="#f4c430"/>
    <rect x="433" y="263" width="57" height="38" fill="#012169"/><rect x="490" y="263" width="57" height="38" fill="#012169"/>
    <path d="M490 312l-22 30 14-2-4 20 12-10 12 10-4-20 14 2z" fill="#1e7b34"/>`,

  // Union of South Africa (1928–1994): Prinsenvlag with three small flags.
  SAF: `${hStripes('#ff7900', '#fff', '#002395')}
    <g transform="translate(212 196) scale(.3)">${unionJack}</g>
    <g transform="translate(320 196)">${[0, 1, 2, 3, 4, 5, 6].map((i) => `<rect y="${i * 10.3}" width="54" height="10.3" fill="${i % 2 ? '#fff' : '#ff7900'}"/>`).join('')}</g>
    <g transform="translate(386 196)"><rect width="66" height="24" fill="#c8102e"/><rect y="24" width="66" height="24" fill="#fff"/>
      <rect y="48" width="66" height="24" fill="#002395"/><rect width="20" height="72" fill="#007a4d"/></g>`,

  // Second Spanish Republic.
  SPR: hStripes('#da121a', '#f1bf00', '#6c3082'),

  // Kingdom of Greece land flag (blue with white cross).
  GRE: `<rect width="640" height="480" fill="#0d5eaf"/><path d="M288 0h64v480h-64zM0 208h640v64H0z" fill="#fff"/>`,

  // Kingdom of Hungary: tricolour with crowned arms.
  HUN: `${hStripes('#ce2939', '#fff', '#477050')}
    <path d="M296 170l6-22 12 12 6-16 6 16 12-12 6 22z" fill="#f4c430" stroke="#9a7b1c" stroke-width="2"/>
    <path d="M288 176h64v64c0 28-16 40-32 46c-16-6-32-18-32-46z" fill="#ce2939" stroke="#9a7b1c" stroke-width="3"/>
    ${[0, 1, 2, 3].map((i) => `<rect x="290" y="${186 + i * 22}" width="30" height="11" fill="#fff"/>`).join('')}
    <path d="M334 188v62M326 202h16M323 218h22" stroke="#fff" stroke-width="5"/>
    <path d="M320 268q14-22 32-12v6q-4 18-32 30z" fill="#477050"/>`,

  // Imperial State of Iran (Pahlavi): tricolour with lion and sun.
  PER: `${hStripes('#239f40', '#fff', '#da0000')}
    ${Array.from({ length: 16 }, (_, i) => {
      const a = (i * Math.PI) / 8
      return `<path d="M${320 + 30 * Math.cos(a)} ${232 + 30 * Math.sin(a)}L${320 + 52 * Math.cos(a)} ${232 + 52 * Math.sin(a)}" stroke="#e8a317" stroke-width="5"/>`
    }).join('')}
    <circle cx="320" cy="232" r="30" fill="#f4c430"/>
    <path d="M286 286c10-16 30-20 46-14l12-12 8 8-8 10c10 4 14 10 14 16h-12c-2-6-8-8-14-8l-6 10h-10l2-12c-8 0-16 2-20 8z" fill="#e8a317"/>`,

  // Kingdom of Iraq (1924–1958).
  IRQ: `${hStripes('#000', '#fff', '#007a3d')}
    <path d="M0 0L200 160V320L0 480z" fill="#ce1126"/>
    ${star(72, 200, 30, '#fff', 7, 0.5)}${star(72, 280, 30, '#fff', 7, 0.5)}`,

  // Kingdom of Egypt (1922–1953).
  EGY: `<rect width="640" height="480" fill="#0b7a3a"/>
    <circle cx="300" cy="240" r="140" fill="#fff"/><circle cx="345" cy="240" r="120" fill="#0b7a3a"/>
    ${star(372, 168, 30, '#fff')}${star(408, 240, 30, '#fff')}${star(372, 312, 30, '#fff')}`,

  // Ethiopian Empire (plain tricolour; the Lion of Judah variant is omitted).
  ETH: hStripes('#078930', '#fcdd09', '#da121a'),

  SOV: `<rect width="640" height="480" fill="#cc0000"/>
    <path d="M120 40l9 27h28l-23 17 9 27-23-17-23 17 9-27-23-17h28z" fill="none" stroke="#ffd700" stroke-width="6"/>
    <path d="M86 150a62 62 0 1 0 88-20a50 50 0 1 1-70 30z" fill="#ffd700"/>
    <path d="M90 216l70-70 12 12-70 70zM150 126l30 30 14-14-30-30z" fill="#ffd700"/>`,

  // Kingdom of Yugoslavia.
  YUG: hStripes('#0c4076', '#fff', '#c6363c'),

  MAN: `<rect width="640" height="480" fill="#ffd700"/>
    <rect width="240" height="40" fill="#e00"/><rect y="40" width="240" height="40" fill="#0047ab"/>
    <rect y="80" width="240" height="40" fill="#fff"/><rect y="120" width="240" height="40" fill="#000"/>`,

  // Star of India.
  RAJ: `<rect width="640" height="480" fill="#c8102e"/>${unionJack}
    <circle cx="480" cy="320" r="70" fill="#1d4f9a" stroke="#ffd700" stroke-width="10"/>
    ${star(480, 320, 50, '#fff')}`,

  // Tibetan snow-lion flag (simplified).
  TIB: `<rect width="640" height="480" fill="#ffd700"/>
    <rect x="20" y="20" width="600" height="440" fill="#c8102e"/>
    <path d="M320 300L20 20h120zM320 300L260 20h120zM320 300L500 20h120zM320 300L20 200v120zM320 300L620 200v120z" fill="#1d4f9a"/>
    <circle cx="320" cy="250" r="60" fill="#ffd700"/>
    <path d="M20 460L320 290 620 460z" fill="#fff"/>`,
}

const dataUrl = (inner: string) =>
  `data:image/svg+xml,${encodeURIComponent(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 640 480">${inner}</svg>`)}`

const cache: Record<string, string | null> = {}

/** URL of the flag image for a tag, or null if we have none. */
export function flagUrl(tag: string): string | null {
  if (tag in cache) return cache[tag]
  const iso = SAME_SINCE_1936[tag]
  const url = HISTORIC[tag]
    ? dataUrl(HISTORIC[tag])
    : iso
      ? (modern[`/node_modules/flag-icons/flags/4x3/${iso}.svg`] ?? null)
      : null
  return (cache[tag] = url)
}
