// Presentation-only metadata for the tags seeded by the API migration.
// Colours loosely follow the HoI4 map palette. Blocs follow common historical
// MP rulesets (Allies, Axis, Comintern, Co-Prosperity Sphere, Chinese United
// Front); minors without a usual side stay non-aligned. They only affect
// grouping and ordering in the UI.

export type Bloc = 'axis' | 'allies' | 'comintern' | 'geacps' | 'chinese' | 'other'

export const BLOCS: Record<Bloc, { label: string; color: string; full?: string }> = {
  axis: { label: 'Axis', color: '#8d8f86' },
  allies: { label: 'Allies', color: '#5f86c4' },
  comintern: { label: 'Comintern', color: '#c0463b' },
  // Japan's own faction in HoI4
  geacps: { label: 'GEACPS', color: '#d9c58a', full: 'Greater East Asia Co-Prosperity Sphere' },
  chinese: { label: 'United Front', color: '#5f9a78', full: 'Chinese United Front' },
  other: { label: 'Non-aligned', color: '#a3946b' },
}

const META: Record<string, [color: string, bloc: Bloc]> = {
  // Axis: Germany, Italy and the European co-belligerents. MP rulesets
  // usually put Spain and Finland here too (Spain often with a join delay).
  GER: ['#5b5f5c', 'axis'],
  ITA: ['#3f8a4a', 'axis'],
  HUN: ['#c98b4b', 'axis'],
  ROM: ['#c2a83e', 'axis'],
  BUL: ['#4f7a3a', 'axis'],
  SPR: ['#c9a64a', 'axis'],
  FIN: ['#d8d8d8', 'axis'],
  // Co-Prosperity Sphere: Japan, Manchukuo, and Siam where it is played
  JAP: ['#d9c58a', 'geacps'],
  MAN: ['#b9a05a', 'geacps'],
  THA: ['#8a6ab0', 'geacps'],
  // Comintern
  SOV: ['#a2302a', 'comintern'],
  MON: ['#b65a3e', 'comintern'],
  // Chinese United Front
  CHI: ['#5f7d4e', 'chinese'],
  PRC: ['#c75b5b', 'chinese'],
  // Allies: UK + Commonwealth, France, USA, the invaded democracies and the
  // later wartime Allies
  ENG: ['#c0473f', 'allies'],
  FRA: ['#3e61b0', 'allies'],
  USA: ['#4b77a8', 'allies'],
  CAN: ['#b94a4a', 'allies'],
  AST: ['#4e7fb8', 'allies'],
  NZL: ['#3b5b8c', 'allies'],
  SAF: ['#a8875a', 'allies'],
  RAJ: ['#c98f5a', 'allies'],
  POL: ['#c45a78', 'allies'],
  HOL: ['#d9853b', 'allies'],
  BEL: ['#b5a23b', 'allies'],
  LUX: ['#7ab0c8', 'allies'],
  NOR: ['#a24848', 'allies'],
  GRE: ['#5b8fc9', 'allies'],
  YUG: ['#6f5fa8', 'allies'],
  PHI: ['#5a7ac8', 'allies'],
  BRA: ['#55a05a', 'allies'],
  MEX: ['#4e8a6a', 'allies'],
  // non-aligned
  DEN: ['#b85c5c', 'other'],
  SWE: ['#3f78b5', 'other'],
  POR: ['#3e8b57', 'other'],
  TUR: ['#8f6a4e', 'other'],
  CZE: ['#5a8eb0', 'other'],
  ARG: ['#8fb4d8', 'other'],
  CHL: ['#b0566a', 'other'],
  TIB: ['#b88a5a', 'other'],
  PER: ['#6aa58a', 'other'],
  IRQ: ['#8a9a5a', 'other'],
  EGY: ['#c4b07a', 'other'],
  ETH: ['#6b9a4a', 'other'],
}

function hashColor(tag: string) {
  let h = 0
  for (const c of tag) h = (h * 31 + c.charCodeAt(0)) % 360
  return `hsl(${h} 35% 50%)`
}

export const nationColor = (tag: string) => META[tag]?.[0] ?? hashColor(tag)
export const nationBloc = (tag: string): Bloc => META[tag]?.[1] ?? 'other'

/** The seven 1936 major powers. */
export const MAJORS = ['GER', 'ITA', 'JAP', 'SOV', 'USA', 'ENG', 'FRA']
export const isMajor = (tag: string) => MAJORS.includes(tag)

export const PRESETS: { label: string; tags: string[] }[] = [
  { label: 'Majors', tags: MAJORS },
  { label: 'Axis', tags: ['GER', 'ITA', 'HUN', 'ROM', 'BUL', 'SPR', 'FIN'] },
  { label: 'GEACPS', tags: ['JAP', 'MAN'] },
  { label: 'Allies', tags: ['ENG', 'FRA', 'USA', 'CAN', 'AST', 'RAJ', 'SAF', 'POL', 'HOL'] },
  { label: 'Comintern', tags: ['SOV', 'MON'] },
  { label: 'United Front', tags: ['CHI', 'PRC'] },
]
