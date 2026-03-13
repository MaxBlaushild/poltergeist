export type SpellEffectType =
  | 'deal_damage'
  | 'deal_damage_all_enemies'
  | 'restore_life_party_member'
  | 'restore_life_all_party_members'
  | 'revive_party_member'
  | 'revive_all_downed_party_members'
  | 'apply_beneficial_statuses'
  | 'apply_detrimental_statuses'
  | 'apply_detrimental_statuses_all_enemies'
  | 'remove_detrimental_statuses'
  | string;

export type DamageAffinity =
  | 'physical'
  | 'piercing'
  | 'slashing'
  | 'bludgeoning'
  | 'fire'
  | 'ice'
  | 'lightning'
  | 'poison'
  | 'arcane'
  | 'holy'
  | 'shadow'
  | string;

export interface SpellStatusTemplate {
  name: string;
  description: string;
  effect: string;
  positive: boolean;
  durationSeconds: number;
  strengthMod: number;
  dexterityMod: number;
  constitutionMod: number;
  intelligenceMod: number;
  wisdomMod: number;
  charismaMod: number;
}

export interface SpellEffect {
  type: SpellEffectType;
  amount?: number;
  damageAffinity?: DamageAffinity;
  statusesToApply?: SpellStatusTemplate[];
  statusesToRemove?: string[];
  effectData?: Record<string, unknown>;
}

export interface SpellProgression {
  id: string;
  createdAt: string;
  updatedAt: string;
  name: string;
  abilityType?: 'spell' | 'technique' | string;
}

export interface SpellProgressionLink {
  id: string;
  createdAt: string;
  updatedAt: string;
  progressionId: string;
  spellId: string;
  levelBand: number;
  progression?: SpellProgression;
}

export interface Spell {
  id: string;
  createdAt: string;
  updatedAt: string;
  name: string;
  description: string;
  iconUrl: string;
  abilityType?: 'spell' | 'technique' | string;
  abilityLevel?: number;
  cooldownTurns?: number;
  imageGenerationStatus?: string;
  imageGenerationError?: string | null;
  effectText: string;
  schoolOfMagic: string;
  manaCost: number;
  effects: SpellEffect[];
  progressionLinks?: SpellProgressionLink[];
}
