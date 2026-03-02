export type SpellEffectType = 'deal_damage' | 'restore_life_party_member' | 'restore_life_all_party_members' | 'apply_beneficial_statuses' | 'remove_detrimental_statuses' | string;
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
    statusesToApply?: SpellStatusTemplate[];
    statusesToRemove?: string[];
    effectData?: Record<string, unknown>;
}
export interface Spell {
    id: string;
    createdAt: string;
    updatedAt: string;
    name: string;
    description: string;
    iconUrl: string;
    abilityType?: 'spell' | 'technique' | string;
    imageGenerationStatus?: string;
    imageGenerationError?: string | null;
    effectText: string;
    schoolOfMagic: string;
    manaCost: number;
    effects: SpellEffect[];
}
