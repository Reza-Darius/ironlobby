// Presentation-only metadata for the tags seeded by the API migration.
// Colours loosely follow the HoI4 map palette; blocs are the usual 1936 MP
// team split and only affect grouping/ordering in the UI.

export type Bloc = 'axis' | 'allies' | 'comintern' | 'other'

export const BLOCS: Record<Bloc, { label: string; color: string }> = {
  axis: { label: 'Axis', color: '#8d8f86' },
  allies: { label: 'Allies', color: '#5f86c4' },
  comintern: { label: 'Comintern', color: '#c0463b' },
  other: { label: 'Non-aligned', color: '#a3946b' },
}

const META: Record<string, [color: string, bloc: Bloc]> = {
  GER: ['#5b5f5c', 'axis'],
  ITA: ['#3f8a4a', 'axis'],
  JAP: ['#d9c58a', 'axis'],
  HUN: ['#c98b4b', 'axis'],
  ROM: ['#c2a83e', 'axis'],
  BUL: ['#4f7a3a', 'axis'],
  MAN: ['#b9a05a', 'axis'],
  FIN: ['#d8d8d8', 'axis'],
  SOV: ['#a2302a', 'comintern'],
  MON: ['#b65a3e', 'comintern'],
  PRC: ['#c75b5b', 'comintern'],
  ENG: ['#c0473f', 'allies'],
  FRA: ['#3e61b0', 'allies'],
  USA: ['#4b77a8', 'allies'],
  POL: ['#c45a78', 'allies'],
  CAN: ['#b94a4a', 'allies'],
  AST: ['#4e7fb8', 'allies'],
  NZL: ['#3b5b8c', 'allies'],
  SAF: ['#a8875a', 'allies'],
  RAJ: ['#c98f5a', 'allies'],
  HOL: ['#d9853b', 'allies'],
  BEL: ['#b5a23b', 'allies'],
  LUX: ['#7ab0c8', 'allies'],
  CHI: ['#5f7d4e', 'allies'],
  NOR: ['#a24848', 'other'],
  DEN: ['#b85c5c', 'other'],
  SWE: ['#3f78b5', 'other'],
  SPR: ['#c9a64a', 'other'],
  POR: ['#3e8b57', 'other'],
  TUR: ['#8f6a4e', 'other'],
  GRE: ['#5b8fc9', 'other'],
  YUG: ['#6f5fa8', 'other'],
  CZE: ['#5a8eb0', 'other'],
  MEX: ['#4e8a6a', 'other'],
  BRA: ['#55a05a', 'other'],
  ARG: ['#8fb4d8', 'other'],
  CHL: ['#b0566a', 'other'],
  TIB: ['#b88a5a', 'other'],
  PER: ['#6aa58a', 'other'],
  IRQ: ['#8a9a5a', 'other'],
  EGY: ['#c4b07a', 'other'],
  ETH: ['#6b9a4a', 'other'],
  THA: ['#8a6ab0', 'other'],
  PHI: ['#5a7ac8', 'other'],
}

function hashColor(tag: string) {
  let h = 0
  for (const c of tag) h = (h * 31 + c.charCodeAt(0)) % 360
  return `hsl(${h} 35% 50%)`
}

export const nationColor = (tag: string) => META[tag]?.[0] ?? hashColor(tag)
export const nationBloc = (tag: string): Bloc => META[tag]?.[1] ?? 'other'

export const PRESETS: { label: string; tags: string[] }[] = [
  { label: 'Majors', tags: ['GER', 'ITA', 'JAP', 'SOV', 'USA', 'ENG', 'FRA'] },
  { label: 'Axis', tags: ['GER', 'ITA', 'JAP', 'HUN', 'ROM', 'BUL', 'MAN'] },
  { label: 'Allies', tags: ['ENG', 'FRA', 'USA', 'CAN', 'AST', 'RAJ', 'SAF', 'POL', 'HOL'] },
  { label: 'Comintern', tags: ['SOV', 'MON', 'PRC'] },
]
