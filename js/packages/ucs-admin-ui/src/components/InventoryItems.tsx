import { useAPI, useMediaContext } from '@poltergeist/contexts';
import {
  InventoryItem,
  InventoryItemSuggestionDraft,
  InventoryItemSuggestionJob,
  Rarity,
  Spell,
} from '@poltergeist/types';
import React, {
  useCallback,
  useMemo,
  useState,
  useEffect,
  useRef,
} from 'react';
import { useUsers } from '../hooks/useUsers.ts';

type SelectOption = {
  value: string;
  label: string;
  secondary?: string;
};

type InventoryConsumeStatus = {
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
  physicalDamageBonusPercent: number;
  piercingDamageBonusPercent: number;
  slashingDamageBonusPercent: number;
  bludgeoningDamageBonusPercent: number;
  fireDamageBonusPercent: number;
  iceDamageBonusPercent: number;
  lightningDamageBonusPercent: number;
  poisonDamageBonusPercent: number;
  arcaneDamageBonusPercent: number;
  holyDamageBonusPercent: number;
  shadowDamageBonusPercent: number;
  physicalResistancePercent: number;
  piercingResistancePercent: number;
  slashingResistancePercent: number;
  bludgeoningResistancePercent: number;
  fireResistancePercent: number;
  iceResistancePercent: number;
  lightningResistancePercent: number;
  poisonResistancePercent: number;
  arcaneResistancePercent: number;
  holyResistancePercent: number;
  shadowResistancePercent: number;
};

type InventoryRecipeIngredientDraft = {
  itemId: number;
  quantity: number;
};

type InventoryRecipeDraft = {
  id: string;
  tier: number;
  isPublic: boolean;
  ingredients: InventoryRecipeIngredientDraft[];
};

type InventoryItemRecord = InventoryItem & {
  consumeHealthDelta?: number;
  consumeManaDelta?: number;
  consumeRevivePartyMemberHealth?: number;
  consumeReviveAllDownedPartyMembersHealth?: number;
  consumeCreateBase?: boolean;
  consumeStatusesToAdd?: InventoryConsumeStatus[];
  consumeStatusesToRemove?: string[];
  consumeSpellIds?: string[];
  consumeTeachRecipeIds?: string[];
  alchemyRecipes?: InventoryRecipeDraft[];
  workshopRecipes?: InventoryRecipeDraft[];
};

type InventorySetGenerationResponse = {
  sourceItemId?: number;
  setTheme: string;
  targetLevel?: number;
  majorStat?: string;
  minorStat?: string;
  rarityTier?: string;
  createdItems: InventoryItemRecord[];
  skippedSlots: string[];
  enqueueWarnings?: string[];
  message: string;
};

type ConsumableQualitiesResponse = {
  sourceItemId: number;
  baseName: string;
  createdItems: InventoryItemRecord[];
  skippedQualities: string[];
  enqueueWarnings?: string[];
  message: string;
};

type InventorySeedPackResponse = {
  processedCount: number;
  createdCount: number;
  updatedCount: number;
};

type InventorySetFamiliesResponse = {
  requestedCount: number;
  createdFamilyCount: number;
  createdItemCount: number;
  skippedFamilyCount: number;
  families: InventorySetGenerationResponse[];
  skippedReasons: string[];
};

type QueueMissingInventoryImagesResponse = {
  totalCount: number;
  missingCount: number;
  queuedCount: number;
  skippedCount: number;
  failedCount: number;
};

type BulkTagAction = 'none' | 'replace' | 'add' | 'remove' | 'clear';

type InventorySuggestionFormState = {
  count: string;
  themePrompt: string;
  categoriesText: string;
  rarityTiersText: string;
  equipSlotsText: string;
  statTagsText: string;
  benefitTagsText: string;
  statusNamesText: string;
  internalTagsText: string;
  minItemLevel: string;
  maxItemLevel: string;
};

const emptyConsumeStatus = (): InventoryConsumeStatus => ({
  name: '',
  description: '',
  effect: '',
  positive: true,
  durationSeconds: 60,
  strengthMod: 0,
  dexterityMod: 0,
  constitutionMod: 0,
  intelligenceMod: 0,
  wisdomMod: 0,
  charismaMod: 0,
  physicalDamageBonusPercent: 0,
  piercingDamageBonusPercent: 0,
  slashingDamageBonusPercent: 0,
  bludgeoningDamageBonusPercent: 0,
  fireDamageBonusPercent: 0,
  iceDamageBonusPercent: 0,
  lightningDamageBonusPercent: 0,
  poisonDamageBonusPercent: 0,
  arcaneDamageBonusPercent: 0,
  holyDamageBonusPercent: 0,
  shadowDamageBonusPercent: 0,
  physicalResistancePercent: 0,
  piercingResistancePercent: 0,
  slashingResistancePercent: 0,
  bludgeoningResistancePercent: 0,
  fireResistancePercent: 0,
  iceResistancePercent: 0,
  lightningResistancePercent: 0,
  poisonResistancePercent: 0,
  arcaneResistancePercent: 0,
  holyResistancePercent: 0,
  shadowResistancePercent: 0,
});

const normalizeConsumeStatus = (
  status?: Partial<InventoryConsumeStatus> | null
): InventoryConsumeStatus => {
  const base = emptyConsumeStatus();
  if (!status) return base;
  return {
    ...base,
    ...status,
    name: (status.name ?? '').trim(),
    description: (status.description ?? '').trim(),
    effect: (status.effect ?? '').trim(),
    durationSeconds: Number.isFinite(status.durationSeconds)
      ? Number(status.durationSeconds)
      : base.durationSeconds,
    strengthMod: Number.isFinite(status.strengthMod)
      ? Number(status.strengthMod)
      : 0,
    dexterityMod: Number.isFinite(status.dexterityMod)
      ? Number(status.dexterityMod)
      : 0,
    constitutionMod: Number.isFinite(status.constitutionMod)
      ? Number(status.constitutionMod)
      : 0,
    intelligenceMod: Number.isFinite(status.intelligenceMod)
      ? Number(status.intelligenceMod)
      : 0,
    wisdomMod: Number.isFinite(status.wisdomMod) ? Number(status.wisdomMod) : 0,
    charismaMod: Number.isFinite(status.charismaMod)
      ? Number(status.charismaMod)
      : 0,
    physicalDamageBonusPercent: Number.isFinite(
      status.physicalDamageBonusPercent
    )
      ? Number(status.physicalDamageBonusPercent)
      : 0,
    piercingDamageBonusPercent: Number.isFinite(
      status.piercingDamageBonusPercent
    )
      ? Number(status.piercingDamageBonusPercent)
      : 0,
    slashingDamageBonusPercent: Number.isFinite(
      status.slashingDamageBonusPercent
    )
      ? Number(status.slashingDamageBonusPercent)
      : 0,
    bludgeoningDamageBonusPercent: Number.isFinite(
      status.bludgeoningDamageBonusPercent
    )
      ? Number(status.bludgeoningDamageBonusPercent)
      : 0,
    fireDamageBonusPercent: Number.isFinite(status.fireDamageBonusPercent)
      ? Number(status.fireDamageBonusPercent)
      : 0,
    iceDamageBonusPercent: Number.isFinite(status.iceDamageBonusPercent)
      ? Number(status.iceDamageBonusPercent)
      : 0,
    lightningDamageBonusPercent: Number.isFinite(
      status.lightningDamageBonusPercent
    )
      ? Number(status.lightningDamageBonusPercent)
      : 0,
    poisonDamageBonusPercent: Number.isFinite(status.poisonDamageBonusPercent)
      ? Number(status.poisonDamageBonusPercent)
      : 0,
    arcaneDamageBonusPercent: Number.isFinite(status.arcaneDamageBonusPercent)
      ? Number(status.arcaneDamageBonusPercent)
      : 0,
    holyDamageBonusPercent: Number.isFinite(status.holyDamageBonusPercent)
      ? Number(status.holyDamageBonusPercent)
      : 0,
    shadowDamageBonusPercent: Number.isFinite(status.shadowDamageBonusPercent)
      ? Number(status.shadowDamageBonusPercent)
      : 0,
    physicalResistancePercent: Number.isFinite(status.physicalResistancePercent)
      ? Number(status.physicalResistancePercent)
      : 0,
    piercingResistancePercent: Number.isFinite(status.piercingResistancePercent)
      ? Number(status.piercingResistancePercent)
      : 0,
    slashingResistancePercent: Number.isFinite(status.slashingResistancePercent)
      ? Number(status.slashingResistancePercent)
      : 0,
    bludgeoningResistancePercent: Number.isFinite(
      status.bludgeoningResistancePercent
    )
      ? Number(status.bludgeoningResistancePercent)
      : 0,
    fireResistancePercent: Number.isFinite(status.fireResistancePercent)
      ? Number(status.fireResistancePercent)
      : 0,
    iceResistancePercent: Number.isFinite(status.iceResistancePercent)
      ? Number(status.iceResistancePercent)
      : 0,
    lightningResistancePercent: Number.isFinite(
      status.lightningResistancePercent
    )
      ? Number(status.lightningResistancePercent)
      : 0,
    poisonResistancePercent: Number.isFinite(status.poisonResistancePercent)
      ? Number(status.poisonResistancePercent)
      : 0,
    arcaneResistancePercent: Number.isFinite(status.arcaneResistancePercent)
      ? Number(status.arcaneResistancePercent)
      : 0,
    holyResistancePercent: Number.isFinite(status.holyResistancePercent)
      ? Number(status.holyResistancePercent)
      : 0,
    shadowResistancePercent: Number.isFinite(status.shadowResistancePercent)
      ? Number(status.shadowResistancePercent)
      : 0,
    positive: status.positive ?? true,
  };
};

const emptyRecipeIngredient = (): InventoryRecipeIngredientDraft => ({
  itemId: 0,
  quantity: 1,
});

const emptyRecipe = (): InventoryRecipeDraft => ({
  id: '',
  tier: 1,
  isPublic: true,
  ingredients: [emptyRecipeIngredient()],
});

const normalizeRecipeIngredient = (
  ingredient?: Partial<InventoryRecipeIngredientDraft> | null
): InventoryRecipeIngredientDraft => ({
  itemId: Number.isFinite(ingredient?.itemId) ? Number(ingredient?.itemId) : 0,
  quantity:
    Number.isFinite(ingredient?.quantity) && Number(ingredient?.quantity) > 0
      ? Number(ingredient?.quantity)
      : 1,
});

const normalizeRecipe = (
  recipe?: Partial<InventoryRecipeDraft> | null
): InventoryRecipeDraft => {
  const base = emptyRecipe();
  const ingredients = (recipe?.ingredients ?? [])
    .map((ingredient) => normalizeRecipeIngredient(ingredient))
    .filter((ingredient) => ingredient.itemId > 0 && ingredient.quantity > 0);

  return {
    ...base,
    ...recipe,
    id: (recipe?.id ?? '').trim(),
    tier:
      Number.isFinite(recipe?.tier) && Number(recipe?.tier) > 0
        ? Number(recipe?.tier)
        : 1,
    isPublic: recipe?.isPublic ?? true,
    ingredients:
      ingredients.length > 0 ? ingredients : [emptyRecipeIngredient()],
  };
};

const parseInternalTagsInput = (value: string): string[] =>
  Array.from(
    new Set(
      value
        .split(',')
        .map((tag) => tag.trim().toLowerCase())
        .filter((tag) => tag !== '')
    )
  );

const normalizeInternalTags = (tags?: string[] | null): string[] =>
  parseInternalTagsInput((tags ?? []).join(','));

const parseCommaValues = (value: string): string[] =>
  Array.from(
    new Set(
      value
        .split(',')
        .map((entry) => entry.trim())
        .filter((entry) => entry !== '')
    )
  );

const renderSuggestionFacetLabel = (
  label: string,
  values?: string[] | null
) => {
  if (!values || values.length === 0) return null;
  return `${label}: ${values.join(', ')}`;
};

const emptyInventorySuggestionForm = (): InventorySuggestionFormState => ({
  count: '12',
  themePrompt: '',
  categoriesText: 'equippable, consumable, material',
  rarityTiersText: 'Common, Uncommon, Epic',
  equipSlotsText: '',
  statTagsText: '',
  benefitTagsText: '',
  statusNamesText: '',
  internalTagsText: '',
  minItemLevel: '1',
  maxItemLevel: '40',
});

const isSuggestionJobPending = (status?: string | null) =>
  status === 'queued' || status === 'in_progress';

const extractApiErrorMessage = (error: unknown, fallback: string) => {
  if (
    error &&
    typeof error === 'object' &&
    'response' in error &&
    error.response &&
    typeof error.response === 'object' &&
    'data' in error.response &&
    error.response.data &&
    typeof error.response.data === 'object' &&
    'error' in error.response.data &&
    typeof (error.response.data as { error?: unknown }).error === 'string'
  ) {
    return (error.response.data as { error: string }).error;
  }
  return fallback;
};

const applyBulkTagAction = (
  currentTags: string[] | undefined,
  nextTags: string[],
  action: BulkTagAction
): string[] => {
  const normalizedCurrent = normalizeInternalTags(currentTags);

  switch (action) {
    case 'replace':
      return nextTags;
    case 'add':
      return Array.from(new Set([...normalizedCurrent, ...nextTags]));
    case 'remove': {
      const tagsToRemove = new Set(nextTags);
      return normalizedCurrent.filter((tag) => !tagsToRemove.has(tag));
    }
    case 'clear':
      return [];
    case 'none':
    default:
      return normalizedCurrent;
  }
};

const SearchableSelect = ({
  label,
  placeholder,
  options,
  value,
  onChange,
}: {
  label: string;
  placeholder: string;
  options: SelectOption[];
  value: string;
  onChange: (value: string) => void;
}) => {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');

  const selected = options.find((o) => o.value === value);
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return options;
    return options.filter((o) => {
      const hay = `${o.label} ${o.secondary ?? ''}`.toLowerCase();
      return hay.includes(q);
    });
  }, [options, query]);

  const displayValue = open ? query : selected?.label ?? '';

  return (
    <div className="relative">
      <label className="block text-sm font-medium text-gray-700">{label}</label>
      <input
        value={displayValue}
        onChange={(e) => {
          setQuery(e.target.value);
          setOpen(true);
        }}
        onFocus={() => {
          setOpen(true);
          setQuery('');
        }}
        onBlur={() => {
          setTimeout(() => setOpen(false), 150);
        }}
        placeholder={placeholder}
        className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
      />
      {open && (
        <div className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md border border-gray-200 bg-white shadow-lg">
          {filtered.length === 0 && (
            <div className="px-3 py-2 text-sm text-gray-500">
              No matches found
            </div>
          )}
          {filtered.map((option) => (
            <button
              type="button"
              key={option.value}
              onMouseDown={(e) => e.preventDefault()}
              onClick={() => {
                onChange(option.value);
                setOpen(false);
                setQuery('');
              }}
              className="flex w-full flex-col items-start px-3 py-2 text-left text-sm hover:bg-indigo-50"
            >
              <span className="font-medium text-gray-900">{option.label}</span>
              {option.secondary && (
                <span className="text-xs text-gray-500">
                  {option.secondary}
                </span>
              )}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

const equipSlotOptions: SelectOption[] = [
  { value: '', label: 'Not equippable' },
  { value: 'hat', label: 'Hat' },
  { value: 'necklace', label: 'Necklace' },
  { value: 'chest', label: 'Chest' },
  { value: 'legs', label: 'Legs' },
  { value: 'shoes', label: 'Shoes' },
  { value: 'gloves', label: 'Gloves' },
  { value: 'dominant_hand', label: 'Dominant Hand' },
  { value: 'off_hand', label: 'Off-hand' },
  { value: 'ring', label: 'Ring (Either Hand)' },
  { value: 'ring_left', label: 'Ring (Left)' },
  { value: 'ring_right', label: 'Ring (Right)' },
];

const itemSetStatOptions: SelectOption[] = [
  { value: 'strength', label: 'Strength' },
  { value: 'dexterity', label: 'Dexterity' },
  { value: 'constitution', label: 'Constitution' },
  { value: 'intelligence', label: 'Intelligence' },
  { value: 'wisdom', label: 'Wisdom' },
  { value: 'charisma', label: 'Charisma' },
];

const itemSetRarityOptions: SelectOption[] = [
  { value: 'auto', label: 'Auto (Level-Based)' },
  { value: Rarity.Common, label: 'Common' },
  { value: Rarity.Uncommon, label: 'Uncommon' },
  { value: Rarity.Epic, label: 'Epic' },
  { value: Rarity.Mythic, label: 'Mythic' },
];

const inventorySetProfileOptions: SelectOption[] = [
  { value: 'martial', label: 'Martial' },
  { value: 'tank', label: 'Tank' },
  { value: 'caster', label: 'Caster' },
  { value: 'skirmisher', label: 'Skirmisher' },
  { value: 'support', label: 'Support' },
  { value: 'hybrid', label: 'Hybrid' },
];

const inventorySetPowerBiasOptions: SelectOption[] = [
  { value: 'balanced', label: 'Balanced' },
  { value: 'offense', label: 'Offense' },
  { value: 'defense', label: 'Defense' },
  { value: 'utility', label: 'Utility' },
];

const inventorySetNamingStyleOptions: SelectOption[] = [
  { value: 'grounded', label: 'Grounded' },
  { value: 'heroic', label: 'Heroic' },
  { value: 'occult', label: 'Occult' },
  { value: 'royal', label: 'Royal' },
  { value: 'wild', label: 'Wild' },
];

const inventorySetSlotScopeOptions: SelectOption[] = [
  { value: 'full_set', label: 'Full Set' },
  { value: 'armor_only', label: 'Armor Only' },
  { value: 'jewelry_only', label: 'Jewelry Only' },
  { value: 'hand_items_only', label: 'Hand Items Only' },
];

const damageAffinityOptions: SelectOption[] = [
  { value: 'physical', label: 'Physical' },
  { value: 'piercing', label: 'Piercing' },
  { value: 'slashing', label: 'Slashing' },
  { value: 'bludgeoning', label: 'Bludgeoning' },
  { value: 'fire', label: 'Fire' },
  { value: 'ice', label: 'Ice' },
  { value: 'lightning', label: 'Lightning' },
  { value: 'poison', label: 'Poison' },
  { value: 'arcane', label: 'Arcane' },
  { value: 'holy', label: 'Holy' },
  { value: 'shadow', label: 'Shadow' },
];

const resistanceFieldOptions = [
  { key: 'physicalResistancePercent', label: 'Physical %' },
  { key: 'piercingResistancePercent', label: 'Piercing %' },
  { key: 'slashingResistancePercent', label: 'Slashing %' },
  { key: 'bludgeoningResistancePercent', label: 'Bludgeoning %' },
  { key: 'fireResistancePercent', label: 'Fire %' },
  { key: 'iceResistancePercent', label: 'Ice %' },
  { key: 'lightningResistancePercent', label: 'Lightning %' },
  { key: 'poisonResistancePercent', label: 'Poison %' },
  { key: 'arcaneResistancePercent', label: 'Arcane %' },
  { key: 'holyResistancePercent', label: 'Holy %' },
  { key: 'shadowResistancePercent', label: 'Shadow %' },
] as const;

const damageBonusFieldOptions = [
  { key: 'physicalDamageBonusPercent', label: 'Physical Dmg %' },
  { key: 'piercingDamageBonusPercent', label: 'Piercing Dmg %' },
  { key: 'slashingDamageBonusPercent', label: 'Slashing Dmg %' },
  { key: 'bludgeoningDamageBonusPercent', label: 'Bludgeoning Dmg %' },
  { key: 'fireDamageBonusPercent', label: 'Fire Dmg %' },
  { key: 'iceDamageBonusPercent', label: 'Ice Dmg %' },
  { key: 'lightningDamageBonusPercent', label: 'Lightning Dmg %' },
  { key: 'poisonDamageBonusPercent', label: 'Poison Dmg %' },
  { key: 'arcaneDamageBonusPercent', label: 'Arcane Dmg %' },
  { key: 'holyDamageBonusPercent', label: 'Holy Dmg %' },
  { key: 'shadowDamageBonusPercent', label: 'Shadow Dmg %' },
] as const;

const equipSlotLabel = (slot?: string | null) => {
  if (!slot) return 'Not equippable';
  const found = equipSlotOptions.find((opt) => opt.value === slot);
  return found?.label || slot;
};

const isHandEquipSlot = (slot?: string | null) =>
  slot === 'dominant_hand' || slot === 'off_hand';

const handItemCategoryOptions: Record<string, SelectOption[]> = {
  dominant_hand: [
    { value: 'weapon', label: 'Weapon' },
    { value: 'staff', label: 'Staff' },
  ],
  off_hand: [
    { value: 'shield', label: 'Shield' },
    { value: 'orb', label: 'Magic Orb' },
  ],
};

const handednessOptions: SelectOption[] = [
  { value: 'one_handed', label: 'One-Handed' },
  { value: 'two_handed', label: 'Two-Handed' },
];

const handItemCategoryLabel = (category?: string | null) => {
  switch (category) {
    case 'weapon':
      return 'Weapon';
    case 'staff':
      return 'Staff';
    case 'shield':
      return 'Shield';
    case 'orb':
      return 'Magic Orb';
    default:
      return category || '';
  }
};

const handednessLabel = (handedness?: string | null) => {
  switch (handedness) {
    case 'one_handed':
      return 'One-Handed';
    case 'two_handed':
      return 'Two-Handed';
    default:
      return handedness || '';
  }
};

const statModSummary = (item: InventoryItemRecord) => {
  const mods: string[] = [];
  const push = (label: string, value?: number) => {
    if (!value || value === 0) return;
    mods.push(`${label} +${value}`);
  };
  push('STR', item.strengthMod);
  push('DEX', item.dexterityMod);
  push('CON', item.constitutionMod);
  push('INT', item.intelligenceMod);
  push('WIS', item.wisdomMod);
  push('CHA', item.charismaMod);
  for (const { key, label } of damageBonusFieldOptions) {
    const value = item[key];
    if (!value || value === 0) continue;
    mods.push(`${label} ${value > 0 ? '+' : ''}${value}`);
  }
  for (const { key, label } of resistanceFieldOptions) {
    const value = item[key];
    if (!value || value === 0) continue;
    mods.push(`${label} ${value > 0 ? '+' : ''}${value}`);
  }
  return mods.join(', ');
};

const handCombatSummary = (item: InventoryItemRecord) => {
  if (!isHandEquipSlot(item.equipSlot)) return [];
  const details: string[] = [];
  if (item.handItemCategory) {
    details.push(`Type: ${handItemCategoryLabel(item.handItemCategory)}`);
  }
  if (item.handedness) {
    details.push(`Usage: ${handednessLabel(item.handedness)}`);
  }
  if (
    item.damageMin !== undefined &&
    item.damageMin !== null &&
    item.damageMax !== undefined &&
    item.damageMax !== null
  ) {
    const swipes = item.swipesPerAttack ?? 0;
    details.push(
      `Damage: ${item.damageMin}-${item.damageMax} (${swipes} swipes)`
    );
  }
  if (item.damageAffinity) {
    details.push(`Affinity: ${item.damageAffinity}`);
  }
  if (
    item.blockPercentage !== undefined &&
    item.blockPercentage !== null &&
    item.damageBlocked !== undefined &&
    item.damageBlocked !== null
  ) {
    details.push(
      `Block: ${item.blockPercentage}% / ${item.damageBlocked} damage`
    );
  }
  if (
    item.spellDamageBonusPercent !== undefined &&
    item.spellDamageBonusPercent !== null
  ) {
    details.push(`Spell bonus: +${item.spellDamageBonusPercent}%`);
  }
  return details;
};

const consumeSummary = (
  item: InventoryItemRecord,
  spellNamesByID: Map<string, string>
) => {
  const details: string[] = [];
  if ((item.consumeHealthDelta ?? 0) !== 0) {
    const value = item.consumeHealthDelta ?? 0;
    details.push(`Health on use: ${value > 0 ? '+' : ''}${value}`);
  }
  if ((item.consumeManaDelta ?? 0) !== 0) {
    const value = item.consumeManaDelta ?? 0;
    details.push(`Mana on use: ${value > 0 ? '+' : ''}${value}`);
  }
  if ((item.consumeRevivePartyMemberHealth ?? 0) > 0) {
    details.push(
      `Revive one party member to ${item.consumeRevivePartyMemberHealth} HP`
    );
  }
  if ((item.consumeReviveAllDownedPartyMembersHealth ?? 0) > 0) {
    details.push(
      `Revive all downed party members to ${item.consumeReviveAllDownedPartyMembersHealth} HP`
    );
  }
  if ((item.consumeStatusesToAdd?.length ?? 0) > 0) {
    details.push(
      `Adds statuses: ${item.consumeStatusesToAdd?.map((status) => status.name).join(', ')}`
    );
  }
  if (item.consumeCreateBase) {
    details.push('Creates a player base on use');
  }
  if ((item.consumeStatusesToRemove?.length ?? 0) > 0) {
    details.push(
      `Removes statuses: ${item.consumeStatusesToRemove?.join(', ')}`
    );
  }
  if ((item.consumeSpellIds?.length ?? 0) > 0) {
    details.push(
      `Grants spells: ${item.consumeSpellIds
        ?.map((spellID) => spellNamesByID.get(spellID) ?? spellID)
        .join(', ')}`
    );
  }
  return details;
};

const hasConsumableEffects = (item: InventoryItemRecord) => {
  if ((item.consumeHealthDelta ?? 0) !== 0) return true;
  if ((item.consumeManaDelta ?? 0) !== 0) return true;
  if ((item.consumeRevivePartyMemberHealth ?? 0) > 0) return true;
  if ((item.consumeReviveAllDownedPartyMembersHealth ?? 0) > 0) return true;
  if (item.consumeCreateBase) return true;
  if ((item.consumeStatusesToAdd?.length ?? 0) > 0) return true;
  if ((item.consumeStatusesToRemove?.length ?? 0) > 0) return true;
  if ((item.consumeSpellIds?.length ?? 0) > 0) return true;
  return false;
};

const consumableQualityPrefixes = [
  'minor',
  'lesser',
  'greater',
  'major',
  'superior',
  'superb',
] as const;

const parseCommaSeparatedValues = (value: string) =>
  value
    .split(',')
    .map((entry) => entry.trim())
    .filter(Boolean);

const toggleStringSetValue = (
  current: Set<string>,
  value: string,
  nextChecked?: boolean
) => {
  const next = new Set(current);
  const shouldAdd = nextChecked ?? !next.has(value);
  if (shouldAdd) {
    next.add(value);
  } else {
    next.delete(value);
  }
  return next;
};

const hasConsumableQualityPrefix = (name?: string | null) => {
  const normalized = (name ?? '').trim().toLowerCase();
  if (!normalized) return false;
  return consumableQualityPrefixes.some((prefix) =>
    normalized.startsWith(`${prefix} `)
  );
};

const canGenerateConsumableQualities = (item: InventoryItemRecord) => {
  if (item.equipSlot) return false;
  if (!hasConsumableEffects(item)) return false;
  return hasConsumableQualityPrefix(item.name);
};

export const InventoryItems = () => {
  const { apiClient } = useAPI();
  const { uploadMedia, getPresignedUploadURL } = useMediaContext();
  const { users } = useUsers();
  const [items, setItems] = useState<InventoryItemRecord[]>([]);
  const [itemTab, setItemTab] = useState<'active' | 'archived'>('active');
  const [spells, setSpells] = useState<Spell[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [showCreateItem, setShowCreateItem] = useState(false);
  const [showGenerateItem, setShowGenerateItem] = useState(false);
  const [editingItem, setEditingItem] = useState<InventoryItemRecord | null>(
    null
  );
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [itemToDelete, setItemToDelete] = useState<InventoryItemRecord | null>(
    null
  );
  const [showBulkDeleteConfirm, setShowBulkDeleteConfirm] = useState(false);
  const [selectedItemIDs, setSelectedItemIDs] = useState<Set<number>>(
    new Set()
  );
  const [bulkDeleteBusy, setBulkDeleteBusy] = useState(false);
  const [bulkEditBusy, setBulkEditBusy] = useState(false);
  const [bulkBuyPriceInput, setBulkBuyPriceInput] = useState('');
  const [bulkTagsInput, setBulkTagsInput] = useState('');
  const [bulkTagAction, setBulkTagAction] = useState<BulkTagAction>('none');
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [useOutfitItem, setUseOutfitItem] =
    useState<InventoryItemRecord | null>(null);
  const [useOutfitUser, setUseOutfitUser] = useState('');
  const [useOutfitSelfieUrl, setUseOutfitSelfieUrl] = useState('');
  const [useOutfitStatus, setUseOutfitStatus] = useState<string | null>(null);
  const [useOutfitStatusKind, setUseOutfitStatusKind] = useState<
    'success' | 'error' | null
  >(null);
  const [useOutfitSubmitting, setUseOutfitSubmitting] = useState(false);
  const [setGenerationBusyIds, setSetGenerationBusyIds] = useState<Set<number>>(
    new Set()
  );
  const [progressionGenerationBusyIds, setProgressionGenerationBusyIds] =
    useState<Set<number>>(new Set());
  const [bulkSetTargetLevel, setBulkSetTargetLevel] = useState('25');
  const [bulkSetMajorStat, setBulkSetMajorStat] = useState('strength');
  const [bulkSetMinorStat, setBulkSetMinorStat] = useState('constitution');
  const [bulkSetRarityTier, setBulkSetRarityTier] = useState('auto');
  const [bulkSetGenerationBusy, setBulkSetGenerationBusy] = useState(false);
  const [setFamilyCount, setSetFamilyCount] = useState('6');
  const [setFamilyLevelMin, setSetFamilyLevelMin] = useState('10');
  const [setFamilyLevelMax, setSetFamilyLevelMax] = useState('60');
  const [setFamilyThemesInput, setSetFamilyThemesInput] = useState('');
  const [setFamilyRequiredTagsInput, setSetFamilyRequiredTagsInput] =
    useState('');
  const [setFamilyForbiddenTagsInput, setSetFamilyForbiddenTagsInput] =
    useState('');
  const [setFamilyRarityTiers, setSetFamilyRarityTiers] = useState<Set<string>>(
    new Set([Rarity.Common, Rarity.Uncommon, Rarity.Epic])
  );
  const [setFamilyProfiles, setSetFamilyProfiles] = useState<Set<string>>(
    new Set(['martial', 'tank', 'caster'])
  );
  const [setFamilyMajorStats, setSetFamilyMajorStats] = useState<Set<string>>(
    new Set()
  );
  const [setFamilyMinorStats, setSetFamilyMinorStats] = useState<Set<string>>(
    new Set()
  );
  const [setFamilyDamageAffinities, setSetFamilyDamageAffinities] = useState<
    Set<string>
  >(new Set(['physical', 'fire', 'ice', 'lightning', 'arcane', 'holy']));
  const [setFamilyResistanceAffinities, setSetFamilyResistanceAffinities] =
    useState<Set<string>>(
      new Set(['physical', 'fire', 'ice', 'shadow', 'holy', 'arcane'])
    );
  const [setFamilySlotScope, setSetFamilySlotScope] = useState('full_set');
  const [setFamilyPowerBias, setSetFamilyPowerBias] = useState('balanced');
  const [setFamilyNamingStyle, setSetFamilyNamingStyle] = useState('grounded');
  const [setFamilyAllowHybridAffinities, setSetFamilyAllowHybridAffinities] =
    useState(true);
  const [
    setFamilyAvoidExistingThemeOverlap,
    setSetFamilyAvoidExistingThemeOverlap,
  ] = useState(true);
  const [setFamilyQueueImages, setSetFamilyQueueImages] = useState(true);
  const [setFamilyGenerationBusy, setSetFamilyGenerationBusy] = useState(false);
  const [seedPackBusy, setSeedPackBusy] = useState(false);
  const [queueMissingImagesBusy, setQueueMissingImagesBusy] = useState(false);
  const [consumableGenerationBusyIds, setConsumableGenerationBusyIds] =
    useState<Set<number>>(new Set());
  const [sortField, setSortField] = useState('name');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');
  const [showFilters, setShowFilters] = useState(false);
  const [suggestionForm, setSuggestionForm] =
    useState<InventorySuggestionFormState>(emptyInventorySuggestionForm);
  const [suggestionJobs, setSuggestionJobs] = useState<
    InventoryItemSuggestionJob[]
  >([]);
  const [selectedSuggestionJobId, setSelectedSuggestionJobId] = useState('');
  const [suggestionDrafts, setSuggestionDrafts] = useState<
    InventoryItemSuggestionDraft[]
  >([]);
  const [loadingSuggestionJobs, setLoadingSuggestionJobs] = useState(false);
  const [loadingSuggestionDrafts, setLoadingSuggestionDrafts] = useState(false);
  const [queueingSuggestionJob, setQueueingSuggestionJob] = useState(false);
  const [suggestionError, setSuggestionError] = useState<string | null>(null);
  const [convertingSuggestionDraftId, setConvertingSuggestionDraftId] =
    useState<string | null>(null);
  const [deletingSuggestionDraftId, setDeletingSuggestionDraftId] = useState<
    string | null
  >(null);
  const [filters, setFilters] = useState({
    rarity: '',
    equipSlot: '',
    imageStatus: '',
    captureType: '',
    equippable: '',
    minId: '',
    maxId: '',
    minBuyPrice: '',
    maxBuyPrice: '',
    minUnlockTier: '',
    maxUnlockTier: '',
    minStrength: '',
    maxStrength: '',
    minDexterity: '',
    maxDexterity: '',
    minConstitution: '',
    maxConstitution: '',
    minIntelligence: '',
    maxIntelligence: '',
    minWisdom: '',
    maxWisdom: '',
    minCharisma: '',
    maxCharisma: '',
  });

  const [formData, setFormData] = useState({
    name: '',
    imageUrl: '',
    flavorText: '',
    effectText: '',
    rarityTier: 'Common' as string,
    isCaptureType: false,
    buyPrice: undefined as number | undefined,
    unlockTier: undefined as number | undefined,
    unlockLocksStrength: undefined as number | undefined,
    itemLevel: 1,
    equipSlot: '',
    strengthMod: 0,
    dexterityMod: 0,
    constitutionMod: 0,
    intelligenceMod: 0,
    wisdomMod: 0,
    charismaMod: 0,
    physicalDamageBonusPercent: 0,
    piercingDamageBonusPercent: 0,
    slashingDamageBonusPercent: 0,
    bludgeoningDamageBonusPercent: 0,
    fireDamageBonusPercent: 0,
    iceDamageBonusPercent: 0,
    lightningDamageBonusPercent: 0,
    poisonDamageBonusPercent: 0,
    arcaneDamageBonusPercent: 0,
    holyDamageBonusPercent: 0,
    shadowDamageBonusPercent: 0,
    physicalResistancePercent: 0,
    piercingResistancePercent: 0,
    slashingResistancePercent: 0,
    bludgeoningResistancePercent: 0,
    fireResistancePercent: 0,
    iceResistancePercent: 0,
    lightningResistancePercent: 0,
    poisonResistancePercent: 0,
    arcaneResistancePercent: 0,
    holyResistancePercent: 0,
    shadowResistancePercent: 0,
    handItemCategory: '',
    handedness: '',
    damageMin: undefined as number | undefined,
    damageMax: undefined as number | undefined,
    damageAffinity: 'slashing',
    swipesPerAttack: undefined as number | undefined,
    blockPercentage: undefined as number | undefined,
    damageBlocked: undefined as number | undefined,
    spellDamageBonusPercent: undefined as number | undefined,
    consumeHealthDelta: 0,
    consumeManaDelta: 0,
    consumeRevivePartyMemberHealth: 0,
    consumeReviveAllDownedPartyMembersHealth: 0,
    consumeCreateBase: false,
    consumeStatusesToAdd: [] as InventoryConsumeStatus[],
    consumeStatusesToRemove: [] as string[],
    consumeSpellIds: [] as string[],
    consumeTeachRecipeIds: [] as string[],
    alchemyRecipes: [] as InventoryRecipeDraft[],
    workshopRecipes: [] as InventoryRecipeDraft[],
    internalTags: [] as string[],
  });
  const [internalTagsInput, setInternalTagsInput] = useState('');

  const [generationData, setGenerationData] = useState({
    name: '',
    description: '',
    rarityTier: 'Common' as string,
    equipSlot: '',
    handItemCategory: '',
    handedness: '',
  });

  const userOptions = useMemo(() => {
    return (users ?? []).map((user) => {
      const username = user.username?.trim() ? `@${user.username}` : '';
      const display = username || user.name || user.phoneNumber;
      const secondary = username ? user.name : user.phoneNumber;
      return {
        value: user.id,
        label: display,
        secondary: secondary && secondary !== display ? secondary : undefined,
      };
    });
  }, [users]);

  const itemOptions = useMemo(() => {
    return items
      .filter((item) => !editingItem || item.id !== editingItem.id)
      .map((item) => ({
        value: String(item.id),
        label: item.name,
        secondary: `ID ${item.id}`,
      }));
  }, [editingItem, items]);

  const recipeOptions = useMemo(() => {
    const options = items.flatMap((item) => {
      const stationRecipes = [
        ...(item.alchemyRecipes ?? []).map((recipe) => ({
          recipe,
          stationLabel: 'Alchemy',
        })),
        ...(item.workshopRecipes ?? []).map((recipe) => ({
          recipe,
          stationLabel: 'Workshop',
        })),
      ];
      return stationRecipes
        .filter(({ recipe }) => recipe.id.trim() !== '')
        .map(({ recipe, stationLabel }) => ({
          value: recipe.id,
          label: `${item.name} (${stationLabel} T${recipe.tier})`,
          secondary: recipe.isPublic ? 'Public recipe' : 'Private recipe',
        }));
    });
    const seen = new Set(options.map((option) => option.value));
    const draftRecipes = [
      ...(formData.alchemyRecipes ?? []).map((recipe) => ({
        recipe,
        stationLabel: 'Alchemy',
      })),
      ...(formData.workshopRecipes ?? []).map((recipe) => ({
        recipe,
        stationLabel: 'Workshop',
      })),
    ];
    for (const { recipe, stationLabel } of draftRecipes) {
      if (!recipe.id || seen.has(recipe.id)) continue;
      options.push({
        value: recipe.id,
        label: `${formData.name || 'Current Item'} (${stationLabel} T${recipe.tier})`,
        secondary: recipe.isPublic ? 'Public recipe' : 'Private recipe',
      });
      seen.add(recipe.id);
    }
    return options;
  }, [formData.alchemyRecipes, formData.name, formData.workshopRecipes, items]);

  const selectedSuggestionJob = useMemo(
    () =>
      suggestionJobs.find((job) => job.id === selectedSuggestionJobId) ?? null,
    [selectedSuggestionJobId, suggestionJobs]
  );

  const fetchItems = useCallback(async () => {
    try {
      const response = await apiClient.get<InventoryItemRecord[]>(
        '/sonar/inventory-items'
      );
      setItems(response);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching inventory items:', error);
      setLoading(false);
    }
  }, [apiClient]);

  const fetchSpells = useCallback(async () => {
    try {
      const response = await apiClient.get<Spell[]>('/sonar/spells');
      setSpells(Array.isArray(response) ? response : []);
    } catch (error) {
      console.error('Error fetching spells:', error);
      setSpells([]);
    }
  }, [apiClient]);

  const fetchSuggestionJobs = useCallback(async () => {
    setLoadingSuggestionJobs(true);
    try {
      const response = await apiClient.get<InventoryItemSuggestionJob[]>(
        '/sonar/inventory-item-suggestion-jobs?limit=20'
      );
      const jobs = Array.isArray(response) ? response : [];
      setSuggestionJobs(jobs);
      setSuggestionError(null);
      setSelectedSuggestionJobId((current) => {
        if (current && jobs.some((job) => job.id === current)) {
          return current;
        }
        return jobs[0]?.id ?? '';
      });
    } catch (error) {
      console.error('Error fetching inventory item suggestion jobs:', error);
      setSuggestionError(
        extractApiErrorMessage(
          error,
          'Failed to load inventory draft generation jobs.'
        )
      );
    } finally {
      setLoadingSuggestionJobs(false);
    }
  }, [apiClient]);

  const fetchSuggestionDrafts = useCallback(
    async (jobId: string) => {
      if (!jobId) {
        setSuggestionDrafts([]);
        return;
      }
      setLoadingSuggestionDrafts(true);
      try {
        const response = await apiClient.get<InventoryItemSuggestionDraft[]>(
          `/sonar/inventory-item-suggestion-jobs/${jobId}/drafts`
        );
        setSuggestionDrafts(Array.isArray(response) ? response : []);
        setSuggestionError(null);
      } catch (error) {
        console.error(
          'Error fetching inventory item suggestion drafts:',
          error
        );
        setSuggestionError(
          extractApiErrorMessage(
            error,
            'Failed to load generated inventory drafts.'
          )
        );
      } finally {
        setLoadingSuggestionDrafts(false);
      }
    },
    [apiClient]
  );

  useEffect(() => {
    void fetchItems();
    void fetchSpells();
    void fetchSuggestionJobs();
  }, [fetchItems, fetchSpells, fetchSuggestionJobs]);

  useEffect(() => {
    const hasPending = items.some((item) =>
      ['queued', 'in_progress'].includes(item.imageGenerationStatus || '')
    );
    if (!hasPending) return;

    const interval = setInterval(() => {
      void fetchItems();
    }, 5000);

    return () => clearInterval(interval);
  }, [fetchItems, items]);

  useEffect(() => {
    setSelectedItemIDs((prev) => {
      if (prev.size === 0) return prev;
      const validIDs = new Set(items.map((item) => item.id));
      const next = new Set<number>();
      prev.forEach((id) => {
        if (validIDs.has(id)) {
          next.add(id);
        }
      });
      return next.size === prev.size ? prev : next;
    });
  }, [items]);

  useEffect(() => {
    setSelectedItemIDs(new Set());
  }, [itemTab]);

  useEffect(() => {
    if (!selectedSuggestionJobId) {
      setSuggestionDrafts([]);
      return;
    }
    void fetchSuggestionDrafts(selectedSuggestionJobId);
  }, [fetchSuggestionDrafts, selectedSuggestionJobId]);

  useEffect(() => {
    const hasPendingSuggestionJob = suggestionJobs.some((job) =>
      isSuggestionJobPending(job.status)
    );
    if (!hasPendingSuggestionJob) return;

    const interval = setInterval(() => {
      void fetchSuggestionJobs();
      if (selectedSuggestionJobId) {
        void fetchSuggestionDrafts(selectedSuggestionJobId);
      }
    }, 5000);

    return () => clearInterval(interval);
  }, [
    fetchSuggestionDrafts,
    fetchSuggestionJobs,
    selectedSuggestionJobId,
    suggestionJobs,
  ]);

  const handleQueueSuggestionJob = async () => {
    setQueueingSuggestionJob(true);
    setSuggestionError(null);
    try {
      const created = await apiClient.post<InventoryItemSuggestionJob>(
        '/sonar/inventory-item-suggestion-jobs',
        {
          count: Math.max(1, parseInt(suggestionForm.count, 10) || 1),
          themePrompt: suggestionForm.themePrompt.trim(),
          categories: parseCommaValues(suggestionForm.categoriesText),
          rarityTiers: parseCommaValues(suggestionForm.rarityTiersText),
          equipSlots: parseCommaValues(suggestionForm.equipSlotsText),
          statTags: parseCommaValues(suggestionForm.statTagsText),
          benefitTags: parseCommaValues(suggestionForm.benefitTagsText),
          statusNames: parseCommaValues(suggestionForm.statusNamesText),
          internalTags: parseInternalTagsInput(suggestionForm.internalTagsText),
          minItemLevel: Math.max(
            1,
            parseInt(suggestionForm.minItemLevel, 10) || 1
          ),
          maxItemLevel: Math.max(
            1,
            parseInt(suggestionForm.maxItemLevel, 10) || 1
          ),
        }
      );
      setSuggestionJobs((current) => [
        created,
        ...current.filter((job) => job.id !== created.id),
      ]);
      setSelectedSuggestionJobId(created.id);
      setSuggestionDrafts([]);
    } catch (error) {
      console.error('Failed to queue inventory item suggestion job:', error);
      setSuggestionError(
        extractApiErrorMessage(
          error,
          'Failed to queue inventory draft generation job.'
        )
      );
    } finally {
      setQueueingSuggestionJob(false);
    }
  };

  const handleConvertSuggestionDraft = async (draftId: string) => {
    setConvertingSuggestionDraftId(draftId);
    setSuggestionError(null);
    try {
      await apiClient.post(
        `/sonar/inventory-item-suggestion-drafts/${draftId}/convert`,
        {}
      );
      await fetchItems();
      if (selectedSuggestionJobId) {
        await fetchSuggestionDrafts(selectedSuggestionJobId);
      }
    } catch (error) {
      console.error(
        'Failed to convert inventory item suggestion draft:',
        error
      );
      setSuggestionError(
        extractApiErrorMessage(
          error,
          'Failed to convert inventory draft into a live item.'
        )
      );
    } finally {
      setConvertingSuggestionDraftId(null);
    }
  };

  const handleDeleteSuggestionDraft = async (draftId: string) => {
    setDeletingSuggestionDraftId(draftId);
    setSuggestionError(null);
    try {
      await apiClient.delete(
        `/sonar/inventory-item-suggestion-drafts/${draftId}`
      );
      setSuggestionDrafts((current) =>
        current.filter((draft) => draft.id !== draftId)
      );
    } catch (error) {
      console.error('Failed to delete inventory item suggestion draft:', error);
      setSuggestionError(
        extractApiErrorMessage(error, 'Failed to delete inventory draft.')
      );
    } finally {
      setDeletingSuggestionDraftId(null);
    }
  };

  const handleGenerateProgressionDrafts = async (item: InventoryItemRecord) => {
    setProgressionGenerationBusyIds((current) => new Set(current).add(item.id));
    setSuggestionError(null);
    try {
      const job = await apiClient.post<InventoryItemSuggestionJob>(
        `/sonar/inventory-items/${item.id}/generate-progression-drafts`,
        {}
      );
      await fetchSuggestionJobs();
      setSelectedSuggestionJobId(job.id);
      await fetchSuggestionDrafts(job.id);
    } catch (error) {
      console.error('Failed to generate progression drafts:', error);
      setSuggestionError(
        extractApiErrorMessage(
          error,
          'Failed to generate progression drafts for this item.'
        )
      );
    } finally {
      setProgressionGenerationBusyIds((current) => {
        const next = new Set(current);
        next.delete(item.id);
        return next;
      });
    }
  };

  const spellNamesByID = useMemo(() => {
    return new Map((spells ?? []).map((spell) => [spell.id, spell.name]));
  }, [spells]);

  const resetForm = () => {
    setFormData({
      name: '',
      imageUrl: '',
      flavorText: '',
      effectText: '',
      rarityTier: 'Common',
      isCaptureType: false,
      buyPrice: undefined,
      unlockTier: undefined,
      unlockLocksStrength: undefined,
      itemLevel: 1,
      equipSlot: '',
      strengthMod: 0,
      dexterityMod: 0,
      constitutionMod: 0,
      intelligenceMod: 0,
      wisdomMod: 0,
      charismaMod: 0,
      physicalDamageBonusPercent: 0,
      piercingDamageBonusPercent: 0,
      slashingDamageBonusPercent: 0,
      bludgeoningDamageBonusPercent: 0,
      fireDamageBonusPercent: 0,
      iceDamageBonusPercent: 0,
      lightningDamageBonusPercent: 0,
      poisonDamageBonusPercent: 0,
      arcaneDamageBonusPercent: 0,
      holyDamageBonusPercent: 0,
      shadowDamageBonusPercent: 0,
      physicalResistancePercent: 0,
      piercingResistancePercent: 0,
      slashingResistancePercent: 0,
      bludgeoningResistancePercent: 0,
      fireResistancePercent: 0,
      iceResistancePercent: 0,
      lightningResistancePercent: 0,
      poisonResistancePercent: 0,
      arcaneResistancePercent: 0,
      holyResistancePercent: 0,
      shadowResistancePercent: 0,
      handItemCategory: '',
      handedness: '',
      damageMin: undefined,
      damageMax: undefined,
      damageAffinity: 'slashing',
      swipesPerAttack: undefined,
      blockPercentage: undefined,
      damageBlocked: undefined,
      spellDamageBonusPercent: undefined,
      consumeHealthDelta: 0,
      consumeManaDelta: 0,
      consumeRevivePartyMemberHealth: 0,
      consumeReviveAllDownedPartyMembersHealth: 0,
      consumeCreateBase: false,
      consumeStatusesToAdd: [],
      consumeStatusesToRemove: [],
      consumeSpellIds: [],
      consumeTeachRecipeIds: [],
      alchemyRecipes: [],
      workshopRecipes: [],
      internalTags: [],
    });
    setInternalTagsInput('');
    setImageFile(null);
    setImagePreview(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const resetGenerationForm = () => {
    setGenerationData({
      name: '',
      description: '',
      rarityTier: 'Common',
      equipSlot: '',
      handItemCategory: '',
      handedness: '',
    });
  };

  const normalizeGenerationDataForSubmit = () => {
    const next = { ...generationData };
    if (!isHandEquipSlot(next.equipSlot)) {
      next.handItemCategory = '';
      next.handedness = '';
      return next;
    }
    if (next.equipSlot === 'dominant_hand') {
      if (
        next.handItemCategory !== 'weapon' &&
        next.handItemCategory !== 'staff'
      ) {
        next.handItemCategory = '';
      }
      if (next.handItemCategory === 'staff') {
        next.handedness = 'two_handed';
      }
    }
    if (next.equipSlot === 'off_hand') {
      if (
        next.handItemCategory !== 'shield' &&
        next.handItemCategory !== 'orb'
      ) {
        next.handItemCategory = '';
      }
      next.handedness = 'one_handed';
    }
    return next;
  };

  const handleGenerationEquipSlotChange = (slot: string) => {
    setGenerationData((prev) => {
      const next = { ...prev, equipSlot: slot };
      if (!isHandEquipSlot(slot)) {
        next.handItemCategory = '';
        next.handedness = '';
        return next;
      }
      if (slot === 'dominant_hand') {
        if (
          next.handItemCategory === 'shield' ||
          next.handItemCategory === 'orb'
        ) {
          next.handItemCategory = '';
        }
        if (next.handItemCategory === 'staff') {
          next.handedness = 'two_handed';
        }
      }
      if (slot === 'off_hand') {
        if (
          next.handItemCategory === 'weapon' ||
          next.handItemCategory === 'staff'
        ) {
          next.handItemCategory = '';
        }
        next.handedness = 'one_handed';
      }
      return next;
    });
  };

  const handleGenerationHandCategoryChange = (category: string) => {
    setGenerationData((prev) => {
      const next = { ...prev, handItemCategory: category };
      if (category === 'staff') {
        next.handedness = 'two_handed';
      } else if (prev.equipSlot === 'off_hand') {
        next.handedness = 'one_handed';
      }
      return next;
    });
  };

  const clearHandFields = () => ({
    handItemCategory: '',
    handedness: '',
    damageMin: undefined as number | undefined,
    damageMax: undefined as number | undefined,
    damageAffinity: undefined as string | undefined,
    swipesPerAttack: undefined as number | undefined,
    blockPercentage: undefined as number | undefined,
    damageBlocked: undefined as number | undefined,
    spellDamageBonusPercent: undefined as number | undefined,
  });

  const normalizeHandFieldsForSubmit = () => {
    const next = { ...formData };
    if (!isHandEquipSlot(next.equipSlot)) {
      return { ...next, ...clearHandFields() };
    }

    if (next.equipSlot === 'dominant_hand') {
      if (
        next.handItemCategory !== 'weapon' &&
        next.handItemCategory !== 'staff'
      ) {
        next.handItemCategory = '';
      }
      if (!next.damageAffinity) {
        next.damageAffinity =
          next.handItemCategory === 'staff' ? 'arcane' : 'physical';
      }
      next.blockPercentage = undefined;
      next.damageBlocked = undefined;
      if (next.handItemCategory === 'staff') {
        next.handedness = 'two_handed';
      }
      if (next.handItemCategory === 'weapon') {
        next.spellDamageBonusPercent = undefined;
      }
    }

    if (next.equipSlot === 'off_hand') {
      if (
        next.handItemCategory !== 'shield' &&
        next.handItemCategory !== 'orb'
      ) {
        next.handItemCategory = '';
      }
      next.handedness = 'one_handed';
      next.damageMin = undefined;
      next.damageMax = undefined;
      next.damageAffinity = undefined;
      next.swipesPerAttack = undefined;
      if (next.handItemCategory === 'shield') {
        next.spellDamageBonusPercent = undefined;
      }
      if (next.handItemCategory === 'orb') {
        next.blockPercentage = undefined;
        next.damageBlocked = undefined;
      }
    }

    next.consumeStatusesToAdd = (next.consumeStatusesToAdd ?? [])
      .map((status) => normalizeConsumeStatus(status))
      .filter((status) => status.name !== '' && status.durationSeconds > 0);
    next.consumeStatusesToRemove = Array.from(
      new Set(
        (next.consumeStatusesToRemove ?? [])
          .map((name) => name.trim())
          .filter((name) => name !== '')
      )
    );
    next.consumeSpellIds = Array.from(
      new Set(
        (next.consumeSpellIds ?? [])
          .map((spellID) => spellID.trim())
          .filter((spellID) => spellID !== '')
      )
    );
    next.alchemyRecipes = (next.alchemyRecipes ?? [])
      .map((recipe) => normalizeRecipe(recipe))
      .map((recipe) => ({
        ...recipe,
        ingredients: recipe.ingredients.filter(
          (ingredient) => ingredient.itemId > 0 && ingredient.quantity > 0
        ),
      }))
      .filter((recipe) => recipe.ingredients.length > 0);
    next.workshopRecipes = (next.workshopRecipes ?? [])
      .map((recipe) => normalizeRecipe(recipe))
      .map((recipe) => ({
        ...recipe,
        ingredients: recipe.ingredients.filter(
          (ingredient) => ingredient.itemId > 0 && ingredient.quantity > 0
        ),
      }))
      .filter((recipe) => recipe.ingredients.length > 0);
    const validRecipeIDs = new Set([
      ...next.alchemyRecipes
        .map((recipe) => recipe.id)
        .filter((recipeID) => recipeID !== ''),
      ...next.workshopRecipes
        .map((recipe) => recipe.id)
        .filter((recipeID) => recipeID !== ''),
      ...items.flatMap((item) => [
        ...(item.alchemyRecipes ?? [])
          .map((recipe) => recipe.id)
          .filter((recipeID) => recipeID !== ''),
        ...(item.workshopRecipes ?? [])
          .map((recipe) => recipe.id)
          .filter((recipeID) => recipeID !== ''),
      ]),
    ]);
    next.consumeTeachRecipeIds = Array.from(
      new Set(
        (next.consumeTeachRecipeIds ?? [])
          .map((recipeID) => recipeID.trim())
          .filter((recipeID) => recipeID !== '' && validRecipeIDs.has(recipeID))
      )
    );
    next.internalTags = parseInternalTagsInput(internalTagsInput);

    return next;
  };

  const addConsumeStatusToAdd = () => {
    setFormData((prev) => ({
      ...prev,
      consumeStatusesToAdd: [
        ...prev.consumeStatusesToAdd,
        emptyConsumeStatus(),
      ],
    }));
  };

  const updateConsumeStatusToAdd = (
    index: number,
    next: Partial<InventoryConsumeStatus>
  ) => {
    setFormData((prev) => {
      const statuses = [...prev.consumeStatusesToAdd];
      statuses[index] = { ...statuses[index], ...next };
      return { ...prev, consumeStatusesToAdd: statuses };
    });
  };

  const removeConsumeStatusToAdd = (index: number) => {
    setFormData((prev) => ({
      ...prev,
      consumeStatusesToAdd: prev.consumeStatusesToAdd.filter(
        (_, i) => i !== index
      ),
    }));
  };

  const addConsumeStatusToRemove = () => {
    setFormData((prev) => ({
      ...prev,
      consumeStatusesToRemove: [...prev.consumeStatusesToRemove, ''],
    }));
  };

  const updateConsumeStatusToRemove = (index: number, value: string) => {
    setFormData((prev) => {
      const names = [...prev.consumeStatusesToRemove];
      names[index] = value;
      return { ...prev, consumeStatusesToRemove: names };
    });
  };

  const removeConsumeStatusToRemove = (index: number) => {
    setFormData((prev) => ({
      ...prev,
      consumeStatusesToRemove: prev.consumeStatusesToRemove.filter(
        (_, i) => i !== index
      ),
    }));
  };

  const addConsumeSpellId = () => {
    setFormData((prev) => ({
      ...prev,
      consumeSpellIds: [...prev.consumeSpellIds, ''],
    }));
  };

  const updateConsumeSpellId = (index: number, value: string) => {
    setFormData((prev) => {
      const spellIDs = [...prev.consumeSpellIds];
      spellIDs[index] = value;
      return { ...prev, consumeSpellIds: spellIDs };
    });
  };

  const removeConsumeSpellId = (index: number) => {
    setFormData((prev) => ({
      ...prev,
      consumeSpellIds: prev.consumeSpellIds.filter((_, i) => i !== index),
    }));
  };

  const addTeachRecipeId = () => {
    setFormData((prev) => ({
      ...prev,
      consumeTeachRecipeIds: [...prev.consumeTeachRecipeIds, ''],
    }));
  };

  const updateTeachRecipeId = (index: number, value: string) => {
    setFormData((prev) => {
      const recipeIDs = [...prev.consumeTeachRecipeIds];
      recipeIDs[index] = value;
      return { ...prev, consumeTeachRecipeIds: recipeIDs };
    });
  };

  const removeTeachRecipeId = (index: number) => {
    setFormData((prev) => ({
      ...prev,
      consumeTeachRecipeIds: prev.consumeTeachRecipeIds.filter(
        (_, i) => i !== index
      ),
    }));
  };

  const addRecipe = (kind: 'alchemyRecipes' | 'workshopRecipes') => {
    setFormData((prev) => ({
      ...prev,
      [kind]: [...prev[kind], emptyRecipe()],
    }));
  };

  const updateRecipe = (
    kind: 'alchemyRecipes' | 'workshopRecipes',
    index: number,
    next: Partial<InventoryRecipeDraft>
  ) => {
    setFormData((prev) => {
      const recipes = [...prev[kind]];
      recipes[index] = { ...recipes[index], ...next };
      return { ...prev, [kind]: recipes };
    });
  };

  const removeRecipe = (
    kind: 'alchemyRecipes' | 'workshopRecipes',
    index: number
  ) => {
    setFormData((prev) => ({
      ...prev,
      [kind]: prev[kind].filter((_, i) => i !== index),
    }));
  };

  const addRecipeIngredient = (
    kind: 'alchemyRecipes' | 'workshopRecipes',
    recipeIndex: number
  ) => {
    setFormData((prev) => {
      const recipes = [...prev[kind]];
      recipes[recipeIndex] = {
        ...recipes[recipeIndex],
        ingredients: [
          ...recipes[recipeIndex].ingredients,
          emptyRecipeIngredient(),
        ],
      };
      return { ...prev, [kind]: recipes };
    });
  };

  const updateRecipeIngredient = (
    kind: 'alchemyRecipes' | 'workshopRecipes',
    recipeIndex: number,
    ingredientIndex: number,
    next: Partial<InventoryRecipeIngredientDraft>
  ) => {
    setFormData((prev) => {
      const recipes = [...prev[kind]];
      const recipe = recipes[recipeIndex];
      const ingredients = [...recipe.ingredients];
      ingredients[ingredientIndex] = {
        ...ingredients[ingredientIndex],
        ...next,
      };
      recipes[recipeIndex] = {
        ...recipe,
        ingredients,
      };
      return { ...prev, [kind]: recipes };
    });
  };

  const removeRecipeIngredient = (
    kind: 'alchemyRecipes' | 'workshopRecipes',
    recipeIndex: number,
    ingredientIndex: number
  ) => {
    setFormData((prev) => {
      const recipes = [...prev[kind]];
      const recipe = recipes[recipeIndex];
      const remaining = recipe.ingredients.filter(
        (_, i) => i !== ingredientIndex
      );
      recipes[recipeIndex] = {
        ...recipe,
        ingredients:
          remaining.length > 0 ? remaining : [emptyRecipeIngredient()],
      };
      return { ...prev, [kind]: recipes };
    });
  };

  const handleEquipSlotChange = (slot: string) => {
    setFormData((prev) => {
      const next = { ...prev, equipSlot: slot };
      if (!isHandEquipSlot(slot)) {
        return { ...next, ...clearHandFields() };
      }
      if (slot === 'dominant_hand') {
        if (
          next.handItemCategory === 'shield' ||
          next.handItemCategory === 'orb'
        ) {
          next.handItemCategory = '';
        }
        if (next.handItemCategory === 'staff') {
          next.handedness = 'two_handed';
        }
      }
      if (slot === 'off_hand') {
        if (
          next.handItemCategory === 'weapon' ||
          next.handItemCategory === 'staff'
        ) {
          next.handItemCategory = '';
        }
        next.handedness = 'one_handed';
      }
      return next;
    });
  };

  const handleHandCategoryChange = (category: string) => {
    setFormData((prev) => {
      const next = { ...prev, handItemCategory: category };
      if (category === 'staff') {
        next.handedness = 'two_handed';
        next.damageAffinity = 'arcane';
        next.blockPercentage = undefined;
        next.damageBlocked = undefined;
      } else if (category === 'weapon') {
        next.damageAffinity = 'slashing';
        next.spellDamageBonusPercent = undefined;
        next.blockPercentage = undefined;
        next.damageBlocked = undefined;
      } else if (category === 'shield') {
        next.handedness = 'one_handed';
        next.damageMin = undefined;
        next.damageMax = undefined;
        next.damageAffinity = undefined;
        next.swipesPerAttack = undefined;
        next.spellDamageBonusPercent = undefined;
      } else if (category === 'orb') {
        next.handedness = 'one_handed';
        next.damageMin = undefined;
        next.damageMax = undefined;
        next.damageAffinity = undefined;
        next.swipesPerAttack = undefined;
        next.blockPercentage = undefined;
        next.damageBlocked = undefined;
      }
      return next;
    });
  };

  const handleCreateItem = async () => {
    try {
      let imageUrl = formData.imageUrl;

      // Upload image to S3 if a file is selected
      if (imageFile) {
        const getExtension = (filename: string): string => {
          return filename.split('.').pop()?.toLowerCase() || 'jpg';
        };
        const extension = getExtension(imageFile.name);
        const timestamp = Date.now();
        const imageKey = `inventory-items/${timestamp}-${Math.random().toString(36).substring(2, 15)}.${extension}`;

        const presignedUrl = await getPresignedUploadURL(
          'crew-points-of-interest',
          imageKey
        );
        if (!presignedUrl) {
          alert('Failed to get upload URL. Please try again.');
          return;
        }

        const uploadSuccess = await uploadMedia(presignedUrl, imageFile);
        if (!uploadSuccess) {
          alert('Failed to upload image. Please try again.');
          return;
        }

        imageUrl = presignedUrl.split('?')[0];
      }

      const submitData = { ...normalizeHandFieldsForSubmit(), imageUrl };
      const newItem = await apiClient.post<InventoryItemRecord>(
        '/sonar/inventory-items',
        submitData
      );
      setItems([...items, newItem]);
      setShowCreateItem(false);
      resetForm();
    } catch (error) {
      console.error('Error creating inventory item:', error);
      alert('Error creating inventory item. Please check all required fields.');
    }
  };

  const buildItemSubmitData = (
    item: InventoryItemRecord,
    overrides: Partial<InventoryItemRecord> = {}
  ) => {
    const next: InventoryItemRecord = {
      ...item,
      ...overrides,
      consumeStatusesToAdd: (
        overrides.consumeStatusesToAdd ??
        item.consumeStatusesToAdd ??
        []
      )
        .map((status) => normalizeConsumeStatus(status))
        .filter((status) => status.name !== '' && status.durationSeconds > 0),
      consumeStatusesToRemove: Array.from(
        new Set(
          (
            overrides.consumeStatusesToRemove ??
            item.consumeStatusesToRemove ??
            []
          )
            .map((name) => name.trim())
            .filter((name) => name !== '')
        )
      ),
      consumeSpellIds: Array.from(
        new Set(
          (overrides.consumeSpellIds ?? item.consumeSpellIds ?? [])
            .map((spellID) => spellID.trim())
            .filter((spellID) => spellID !== '')
        )
      ),
      consumeTeachRecipeIds: Array.from(
        new Set(
          (overrides.consumeTeachRecipeIds ?? item.consumeTeachRecipeIds ?? [])
            .map((recipeID) => recipeID.trim())
            .filter((recipeID) => recipeID !== '')
        )
      ),
      alchemyRecipes: (overrides.alchemyRecipes ?? item.alchemyRecipes ?? [])
        .map((recipe) => normalizeRecipe(recipe))
        .map((recipe) => ({
          ...recipe,
          ingredients: recipe.ingredients.filter(
            (ingredient) => ingredient.itemId > 0 && ingredient.quantity > 0
          ),
        }))
        .filter((recipe) => recipe.ingredients.length > 0),
      workshopRecipes: (overrides.workshopRecipes ?? item.workshopRecipes ?? [])
        .map((recipe) => normalizeRecipe(recipe))
        .map((recipe) => ({
          ...recipe,
          ingredients: recipe.ingredients.filter(
            (ingredient) => ingredient.itemId > 0 && ingredient.quantity > 0
          ),
        }))
        .filter((recipe) => recipe.ingredients.length > 0),
      internalTags: normalizeInternalTags(
        overrides.internalTags ?? item.internalTags
      ),
    };

    if (!isHandEquipSlot(next.equipSlot ?? '')) {
      return { ...next, ...clearHandFields() };
    }

    if (next.equipSlot === 'dominant_hand') {
      if (
        next.handItemCategory !== 'weapon' &&
        next.handItemCategory !== 'staff'
      ) {
        next.handItemCategory = '';
      }
      if (!next.damageAffinity) {
        next.damageAffinity =
          next.handItemCategory === 'staff' ? 'arcane' : 'physical';
      }
      next.blockPercentage = undefined;
      next.damageBlocked = undefined;
      if (next.handItemCategory === 'staff') {
        next.handedness = 'two_handed';
      }
      if (next.handItemCategory === 'weapon') {
        next.spellDamageBonusPercent = undefined;
      }
    }

    if (next.equipSlot === 'off_hand') {
      if (
        next.handItemCategory !== 'shield' &&
        next.handItemCategory !== 'orb'
      ) {
        next.handItemCategory = '';
      }
      next.handedness = 'one_handed';
      next.damageMin = undefined;
      next.damageMax = undefined;
      next.damageAffinity = undefined;
      next.swipesPerAttack = undefined;
      if (next.handItemCategory === 'shield') {
        next.spellDamageBonusPercent = undefined;
      }
      if (next.handItemCategory === 'orb') {
        next.blockPercentage = undefined;
        next.damageBlocked = undefined;
      }
    }

    return next;
  };

  const handleUpdateItem = async () => {
    if (!editingItem) return;

    try {
      let imageUrl = formData.imageUrl;

      // Upload new image to S3 if a file is selected, otherwise keep existing imageUrl
      if (imageFile) {
        const getExtension = (filename: string): string => {
          return filename.split('.').pop()?.toLowerCase() || 'jpg';
        };
        const extension = getExtension(imageFile.name);
        const timestamp = Date.now();
        const imageKey = `inventory-items/${timestamp}-${Math.random().toString(36).substring(2, 15)}.${extension}`;

        const presignedUrl = await getPresignedUploadURL(
          'crew-points-of-interest',
          imageKey
        );
        if (!presignedUrl) {
          alert('Failed to get upload URL. Please try again.');
          return;
        }

        const uploadSuccess = await uploadMedia(presignedUrl, imageFile);
        if (!uploadSuccess) {
          alert('Failed to upload image. Please try again.');
          return;
        }

        imageUrl = presignedUrl.split('?')[0];
      }

      const submitData = { ...normalizeHandFieldsForSubmit(), imageUrl };
      const updatedItem = await apiClient.put<InventoryItemRecord>(
        `/sonar/inventory-items/${editingItem.id}`,
        submitData
      );
      setItems(items.map((i) => (i.id === editingItem.id ? updatedItem : i)));
      setEditingItem(null);
      resetForm();
    } catch (error) {
      console.error('Error updating inventory item:', error);
      alert('Error updating inventory item. Please check all required fields.');
    }
  };

  const handleGenerateItem = async () => {
    try {
      const normalized = normalizeGenerationDataForSubmit();
      if (
        isHandEquipSlot(normalized.equipSlot) &&
        (!normalized.handItemCategory || !normalized.handedness)
      ) {
        alert(
          'For hand equipment generation, select both hand item type and handedness.'
        );
        return;
      }
      const newItem = await apiClient.post<InventoryItemRecord>(
        '/sonar/inventory-items/generate',
        {
          name: normalized.name,
          description: normalized.description,
          rarityTier: normalized.rarityTier,
          equipSlot: normalized.equipSlot,
          handItemCategory: normalized.handItemCategory,
          handedness: normalized.handedness,
        }
      );
      setItems([...items, newItem]);
      setShowGenerateItem(false);
      resetGenerationForm();
    } catch (error) {
      console.error('Error generating inventory item:', error);
      alert(
        'Error generating inventory item. Please check all required fields.'
      );
    }
  };

  const handleRegenerateImage = async (item: InventoryItemRecord) => {
    try {
      const updated = await apiClient.post<InventoryItemRecord>(
        `/sonar/inventory-items/${item.id}/regenerate`,
        {}
      );
      setItems(items.map((i) => (i.id === item.id ? updated : i)));
    } catch (error) {
      console.error('Error regenerating inventory item image:', error);
      alert('Error regenerating inventory item image.');
    }
  };

  const handleUseOutfit = async () => {
    if (!useOutfitItem) return;
    try {
      setUseOutfitSubmitting(true);
      setUseOutfitStatus(null);
      setUseOutfitStatusKind(null);
      await apiClient.post('/sonar/admin/useOutfitItem', {
        userID: useOutfitUser,
        itemID: useOutfitItem.id,
        selfieUrl: useOutfitSelfieUrl,
      });
      setUseOutfitStatus('Outfit generation queued.');
      setUseOutfitStatusKind('success');
    } catch (error) {
      console.error('Error using outfit item:', error);
      setUseOutfitStatus('Failed to start outfit generation.');
      setUseOutfitStatusKind('error');
    } finally {
      setUseOutfitSubmitting(false);
    }
  };

  const isOutfitName = (name?: string) =>
    (name || '').trim().toLowerCase().endsWith('outfit');

  const handleDeleteItem = async (item: InventoryItemRecord) => {
    setItemToDelete(item);
    setShowDeleteConfirm(true);
  };

  const confirmDelete = async () => {
    if (!itemToDelete) return;

    try {
      await apiClient.delete(`/sonar/inventory-items/${itemToDelete.id}`);
      setItems(items.filter((i) => i.id !== itemToDelete.id));
      setSelectedItemIDs((prev) => {
        if (!prev.has(itemToDelete.id)) return prev;
        const next = new Set(prev);
        next.delete(itemToDelete.id);
        return next;
      });
      setShowDeleteConfirm(false);
      setItemToDelete(null);
    } catch (error) {
      console.error('Error deleting inventory item:', error);
      alert('Error deleting inventory item.');
    }
  };

  const toggleItemSelection = (itemID: number, checked: boolean) => {
    setSelectedItemIDs((prev) => {
      const next = new Set(prev);
      if (checked) {
        next.add(itemID);
      } else {
        next.delete(itemID);
      }
      return next;
    });
  };

  const toggleSelectAllVisible = (checked: boolean, itemIDs: number[]) => {
    setSelectedItemIDs((prev) => {
      const next = new Set(prev);
      for (const itemID of itemIDs) {
        if (checked) {
          next.add(itemID);
        } else {
          next.delete(itemID);
        }
      }
      return next;
    });
  };

  const confirmBulkDelete = async () => {
    const ids = Array.from(selectedItemIDs);
    if (ids.length === 0) return;

    try {
      setBulkDeleteBusy(true);
      await apiClient.post('/sonar/inventory-items/bulk-delete', { ids });
      const selectedIDSet = new Set(ids);
      setItems((prev) => prev.filter((item) => !selectedIDSet.has(item.id)));
      setSelectedItemIDs(new Set());
      setShowBulkDeleteConfirm(false);
    } catch (error) {
      console.error('Error bulk deleting inventory items:', error);
      alert('Error bulk deleting inventory items.');
    } finally {
      setBulkDeleteBusy(false);
    }
  };

  const handleSetItemsArchived = async (ids: number[], archived: boolean) => {
    if (ids.length === 0) return;

    try {
      await apiClient.post('/sonar/inventory-items/bulk-archive', {
        ids,
        archived,
      });
      const idSet = new Set(ids);
      setItems((prev) =>
        prev.map((item) => (idSet.has(item.id) ? { ...item, archived } : item))
      );
      setSelectedItemIDs((prev) => {
        const next = new Set(prev);
        ids.forEach((id) => next.delete(id));
        return next;
      });
    } catch (error) {
      console.error(
        'Error updating archived state for inventory items:',
        error
      );
      alert(`Error ${archived ? 'archiving' : 'restoring'} inventory items.`);
    }
  };

  const handleBulkEditSelected = async () => {
    const selectedItems = items.filter((item) => selectedItemIDs.has(item.id));
    if (selectedItems.length === 0) return;

    const trimmedPrice = bulkBuyPriceInput.trim();
    const hasPriceChange = trimmedPrice !== '';
    const hasTagChange = bulkTagAction !== 'none';

    if (!hasPriceChange && !hasTagChange) {
      alert(
        'Choose a buy price and/or a tag action before applying bulk edits.'
      );
      return;
    }

    let buyPrice: number | undefined;
    if (hasPriceChange) {
      buyPrice = Number(trimmedPrice);
      if (!Number.isFinite(buyPrice) || buyPrice < 0) {
        alert('Buy price must be a non-negative number.');
        return;
      }
    }

    const parsedTags =
      hasTagChange && bulkTagAction !== 'clear'
        ? parseInternalTagsInput(bulkTagsInput)
        : [];

    if (hasTagChange && bulkTagAction !== 'clear' && parsedTags.length === 0) {
      alert('Enter at least one tag for the selected tag action.');
      return;
    }

    setBulkEditBusy(true);
    try {
      const updatedItems = await Promise.all(
        selectedItems.map((item) => {
          const submitData = buildItemSubmitData(item, {
            ...(hasPriceChange ? { buyPrice } : {}),
            ...(hasTagChange
              ? {
                  internalTags: applyBulkTagAction(
                    item.internalTags,
                    parsedTags,
                    bulkTagAction
                  ),
                }
              : {}),
          });

          return apiClient.put<InventoryItemRecord>(
            `/sonar/inventory-items/${item.id}`,
            submitData
          );
        })
      );

      const updatedByID = new Map(updatedItems.map((item) => [item.id, item]));
      setItems((prev) => prev.map((item) => updatedByID.get(item.id) ?? item));
      setBulkBuyPriceInput('');
      setBulkTagsInput('');
      setBulkTagAction('none');
      alert(`Updated ${updatedItems.length} inventory item(s).`);
    } catch (error) {
      console.error('Error bulk updating inventory items:', error);
      alert('Error bulk updating inventory items.');
    } finally {
      setBulkEditBusy(false);
    }
  };

  const handleGenerateSet = async (item: InventoryItemRecord) => {
    if (!item.equipSlot) {
      alert('Only equippable items can generate a set.');
      return;
    }

    setSetGenerationBusyIds((prev) => {
      const next = new Set(prev);
      next.add(item.id);
      return next;
    });

    try {
      const response = await apiClient.post<InventorySetGenerationResponse>(
        `/sonar/inventory-items/${item.id}/generate-set`,
        {}
      );

      const createdItems = Array.isArray(response.createdItems)
        ? response.createdItems
        : [];
      setItems((prev) => {
        const byId = new Map(prev.map((entry) => [entry.id, entry]));
        createdItems.forEach((created) => {
          byId.set(created.id, created);
        });
        return Array.from(byId.values());
      });

      const skippedCount = Array.isArray(response.skippedSlots)
        ? response.skippedSlots.length
        : 0;
      const warningCount = Array.isArray(response.enqueueWarnings)
        ? response.enqueueWarnings.length
        : 0;
      alert(
        `Set generation complete. Created ${createdItems.length} item(s), skipped ${skippedCount} slot(s)` +
          (warningCount > 0
            ? `, with ${warningCount} image queue warning(s).`
            : '.')
      );
    } catch (error) {
      console.error('Error generating equipment set:', error);
      alert('Error generating item set.');
    } finally {
      setSetGenerationBusyIds((prev) => {
        const next = new Set(prev);
        next.delete(item.id);
        return next;
      });
    }
  };

  const handleGenerateSetFromStats = async () => {
    const targetLevel = Number.parseInt(bulkSetTargetLevel, 10);
    if (!Number.isFinite(targetLevel) || targetLevel < 1 || targetLevel > 100) {
      alert('Target level must be between 1 and 100.');
      return;
    }
    if (!bulkSetMajorStat || !bulkSetMinorStat) {
      alert('Major and minor stats are required.');
      return;
    }
    if (bulkSetMajorStat === bulkSetMinorStat) {
      alert('Major and minor stats must be different.');
      return;
    }

    setBulkSetGenerationBusy(true);
    try {
      const response = await apiClient.post<InventorySetGenerationResponse>(
        '/sonar/inventory-items/generate-equippable-set',
        {
          targetLevel,
          majorStat: bulkSetMajorStat,
          minorStat: bulkSetMinorStat,
          rarityTier:
            bulkSetRarityTier !== 'auto' ? bulkSetRarityTier : undefined,
        }
      );
      const createdItems = Array.isArray(response.createdItems)
        ? response.createdItems
        : [];
      setItems((prev) => {
        const byId = new Map(prev.map((entry) => [entry.id, entry]));
        createdItems.forEach((created) => {
          byId.set(created.id, created);
        });
        return Array.from(byId.values());
      });

      const skippedCount = Array.isArray(response.skippedSlots)
        ? response.skippedSlots.length
        : 0;
      const warningCount = Array.isArray(response.enqueueWarnings)
        ? response.enqueueWarnings.length
        : 0;
      const resolvedRarity = response.rarityTier ?? 'Unknown';
      alert(
        `Generated ${resolvedRarity} set "${response.setTheme}". Created ${createdItems.length} item(s), skipped ${skippedCount} slot(s)` +
          (warningCount > 0
            ? `, with ${warningCount} image queue warning(s).`
            : '.')
      );
    } catch (error) {
      console.error('Error generating stat-driven equipment set:', error);
      alert('Error generating equipment set.');
    } finally {
      setBulkSetGenerationBusy(false);
    }
  };

  const handleGenerateSetFamilies = async () => {
    const count = Number.parseInt(setFamilyCount, 10);
    const levelMin = Number.parseInt(setFamilyLevelMin, 10);
    const levelMax = Number.parseInt(setFamilyLevelMax, 10);

    if (!Number.isFinite(count) || count < 1 || count > 40) {
      alert('Family count must be between 1 and 40.');
      return;
    }
    if (
      !Number.isFinite(levelMin) ||
      !Number.isFinite(levelMax) ||
      levelMin < 1 ||
      levelMax > 100 ||
      levelMin > levelMax
    ) {
      alert('Level range must stay between 1 and 100, with min <= max.');
      return;
    }

    setSetFamilyGenerationBusy(true);
    try {
      const response = await apiClient.post<InventorySetFamiliesResponse>(
        '/sonar/inventory-items/generate-set-families',
        {
          count,
          levelMin,
          levelMax,
          rarityTiers: Array.from(setFamilyRarityTiers),
          themes: parseCommaSeparatedValues(setFamilyThemesInput),
          majorStats: Array.from(setFamilyMajorStats),
          minorStats: Array.from(setFamilyMinorStats),
          profiles: Array.from(setFamilyProfiles),
          damageAffinities: Array.from(setFamilyDamageAffinities),
          resistanceAffinities: Array.from(setFamilyResistanceAffinities),
          requiredInternalTags: parseCommaSeparatedValues(
            setFamilyRequiredTagsInput
          ),
          forbiddenTags: parseCommaSeparatedValues(setFamilyForbiddenTagsInput),
          slotScope: setFamilySlotScope,
          powerBias: setFamilyPowerBias,
          namingStyle: setFamilyNamingStyle,
          allowHybridAffinities: setFamilyAllowHybridAffinities,
          avoidExistingThemeOverlap: setFamilyAvoidExistingThemeOverlap,
          queueImages: setFamilyQueueImages,
        }
      );
      await fetchItems();

      const warningCount = (response.families ?? []).reduce(
        (countWarnings, family) =>
          countWarnings + (family.enqueueWarnings?.length ?? 0),
        0
      );
      alert(
        `Generated ${response.createdFamilyCount} set family(s) and ${response.createdItemCount} item(s). Skipped ${response.skippedFamilyCount} family attempt(s)` +
          (warningCount > 0
            ? `, with ${warningCount} image queue warning(s).`
            : '.')
      );
    } catch (error) {
      console.error('Error generating equipment set families:', error);
      alert('Error generating equipment set families.');
    } finally {
      setSetFamilyGenerationBusy(false);
    }
  };

  const handleSeedCorePack = async () => {
    setSeedPackBusy(true);
    try {
      const response = await apiClient.post<InventorySeedPackResponse>(
        '/sonar/inventory-items/seed-pack',
        {}
      );
      await fetchItems();
      alert(
        `Seeded core inventory pack. Processed ${response.processedCount} item(s) (${response.createdCount} created, ${response.updatedCount} updated).`
      );
    } catch (error) {
      console.error('Error seeding core inventory pack:', error);
      alert('Error seeding core inventory pack.');
    } finally {
      setSeedPackBusy(false);
    }
  };

  const handleQueueMissingImages = async () => {
    setQueueMissingImagesBusy(true);
    try {
      const response =
        await apiClient.post<QueueMissingInventoryImagesResponse>(
          '/sonar/inventory-items/queue-missing-images',
          {}
        );
      await fetchItems();
      alert(
        `Queued ${response.queuedCount} missing inventory image job(s). ${response.skippedCount} skipped, ${response.failedCount} failed.`
      );
    } catch (error) {
      console.error('Error queueing missing inventory item images:', error);
      alert('Error queueing missing inventory item images.');
    } finally {
      setQueueMissingImagesBusy(false);
    }
  };

  const handleGenerateConsumableQualities = async (
    item: InventoryItemRecord
  ) => {
    if (!canGenerateConsumableQualities(item)) {
      alert(
        'Only non-equippable consumables with effects and a quality prefix (Minor/Lesser/Greater/Major/Superior/Superb) can generate quality progression.'
      );
      return;
    }

    setConsumableGenerationBusyIds((prev) => {
      const next = new Set(prev);
      next.add(item.id);
      return next;
    });

    try {
      const response = await apiClient.post<ConsumableQualitiesResponse>(
        `/sonar/inventory-items/${item.id}/generate-consumable-qualities`,
        {}
      );

      const createdItems = Array.isArray(response.createdItems)
        ? response.createdItems
        : [];
      setItems((prev) => {
        const byId = new Map(prev.map((entry) => [entry.id, entry]));
        createdItems.forEach((created) => {
          byId.set(created.id, created);
        });
        return Array.from(byId.values());
      });

      const skippedCount = Array.isArray(response.skippedQualities)
        ? response.skippedQualities.length
        : 0;
      const warningCount = Array.isArray(response.enqueueWarnings)
        ? response.enqueueWarnings.length
        : 0;
      alert(
        `Consumable quality generation complete. Created ${createdItems.length} item(s), skipped ${skippedCount} quality tier(s)` +
          (warningCount > 0
            ? `, with ${warningCount} image queue warning(s).`
            : '.')
      );
    } catch (error) {
      console.error('Error generating consumable qualities:', error);
      alert('Error generating consumable qualities.');
    } finally {
      setConsumableGenerationBusyIds((prev) => {
        const next = new Set(prev);
        next.delete(item.id);
        return next;
      });
    }
  };

  const handleEditItem = (item: InventoryItemRecord) => {
    setEditingItem(item);
    setFormData({
      name: item.name,
      imageUrl: item.imageUrl,
      flavorText: item.flavorText,
      effectText: item.effectText,
      rarityTier: item.rarityTier,
      isCaptureType: item.isCaptureType,
      buyPrice: item.buyPrice,
      unlockTier: item.unlockTier,
      unlockLocksStrength: item.unlockLocksStrength,
      itemLevel: item.itemLevel ?? 1,
      equipSlot: item.equipSlot || '',
      strengthMod: item.strengthMod ?? 0,
      dexterityMod: item.dexterityMod ?? 0,
      constitutionMod: item.constitutionMod ?? 0,
      intelligenceMod: item.intelligenceMod ?? 0,
      wisdomMod: item.wisdomMod ?? 0,
      charismaMod: item.charismaMod ?? 0,
      physicalDamageBonusPercent: item.physicalDamageBonusPercent ?? 0,
      piercingDamageBonusPercent: item.piercingDamageBonusPercent ?? 0,
      slashingDamageBonusPercent: item.slashingDamageBonusPercent ?? 0,
      bludgeoningDamageBonusPercent: item.bludgeoningDamageBonusPercent ?? 0,
      fireDamageBonusPercent: item.fireDamageBonusPercent ?? 0,
      iceDamageBonusPercent: item.iceDamageBonusPercent ?? 0,
      lightningDamageBonusPercent: item.lightningDamageBonusPercent ?? 0,
      poisonDamageBonusPercent: item.poisonDamageBonusPercent ?? 0,
      arcaneDamageBonusPercent: item.arcaneDamageBonusPercent ?? 0,
      holyDamageBonusPercent: item.holyDamageBonusPercent ?? 0,
      shadowDamageBonusPercent: item.shadowDamageBonusPercent ?? 0,
      physicalResistancePercent: item.physicalResistancePercent ?? 0,
      piercingResistancePercent: item.piercingResistancePercent ?? 0,
      slashingResistancePercent: item.slashingResistancePercent ?? 0,
      bludgeoningResistancePercent: item.bludgeoningResistancePercent ?? 0,
      fireResistancePercent: item.fireResistancePercent ?? 0,
      iceResistancePercent: item.iceResistancePercent ?? 0,
      lightningResistancePercent: item.lightningResistancePercent ?? 0,
      poisonResistancePercent: item.poisonResistancePercent ?? 0,
      arcaneResistancePercent: item.arcaneResistancePercent ?? 0,
      holyResistancePercent: item.holyResistancePercent ?? 0,
      shadowResistancePercent: item.shadowResistancePercent ?? 0,
      handItemCategory: item.handItemCategory ?? '',
      handedness: item.handedness ?? '',
      damageMin: item.damageMin ?? undefined,
      damageMax: item.damageMax ?? undefined,
      damageAffinity: item.damageAffinity ?? 'physical',
      swipesPerAttack: item.swipesPerAttack ?? undefined,
      blockPercentage: item.blockPercentage ?? undefined,
      damageBlocked: item.damageBlocked ?? undefined,
      spellDamageBonusPercent: item.spellDamageBonusPercent ?? undefined,
      consumeHealthDelta: item.consumeHealthDelta ?? 0,
      consumeManaDelta: item.consumeManaDelta ?? 0,
      consumeRevivePartyMemberHealth: item.consumeRevivePartyMemberHealth ?? 0,
      consumeReviveAllDownedPartyMembersHealth:
        item.consumeReviveAllDownedPartyMembersHealth ?? 0,
      consumeCreateBase: item.consumeCreateBase ?? false,
      consumeStatusesToAdd: (item.consumeStatusesToAdd ?? []).map((status) =>
        normalizeConsumeStatus(status)
      ),
      consumeStatusesToRemove: [...(item.consumeStatusesToRemove ?? [])],
      consumeSpellIds: [...(item.consumeSpellIds ?? [])],
      consumeTeachRecipeIds: [...(item.consumeTeachRecipeIds ?? [])],
      alchemyRecipes: (item.alchemyRecipes ?? []).map((recipe) =>
        normalizeRecipe(recipe)
      ),
      workshopRecipes: (item.workshopRecipes ?? []).map((recipe) =>
        normalizeRecipe(recipe)
      ),
      internalTags: [...(item.internalTags ?? [])],
    });
    setInternalTagsInput((item.internalTags ?? []).join(', '));
    setImageFile(null);
    setImagePreview(item.imageUrl || null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const renderRecipeEditor = (
    label: string,
    kind: 'alchemyRecipes' | 'workshopRecipes',
    accentClass: string
  ) => {
    const recipes = formData[kind];
    return (
      <div style={{ marginBottom: '12px' }}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '8px',
          }}
        >
          <label style={{ fontSize: '13px', fontWeight: 600 }}>{label}</label>
          <button
            type="button"
            onClick={() => addRecipe(kind)}
            className={`${accentClass} text-white px-2 py-1 rounded-md text-xs`}
          >
            Add Recipe
          </button>
        </div>
        {recipes.length === 0 && (
          <small style={{ color: '#666', fontSize: '12px' }}>
            No recipes configured.
          </small>
        )}
        {recipes.map((recipe, recipeIndex) => (
          <div
            key={`${kind}-${recipe.id || recipeIndex}`}
            style={{
              border: '1px solid #e5e7eb',
              borderRadius: '8px',
              padding: '12px',
              marginBottom: '10px',
            }}
          >
            <div
              style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(3, 1fr)',
                gap: '8px',
                marginBottom: '10px',
              }}
            >
              <div>
                <label
                  style={{
                    display: 'block',
                    marginBottom: '4px',
                    fontSize: '12px',
                  }}
                >
                  Tier
                </label>
                <input
                  type="number"
                  min="1"
                  value={recipe.tier}
                  onChange={(e) =>
                    updateRecipe(kind, recipeIndex, {
                      tier: parseInt(e.target.value, 10) || 1,
                    })
                  }
                  style={{
                    width: '100%',
                    padding: '6px',
                    border: '1px solid #ccc',
                    borderRadius: '4px',
                  }}
                />
              </div>
              <div>
                <label
                  style={{
                    display: 'block',
                    marginBottom: '4px',
                    fontSize: '12px',
                  }}
                >
                  Visibility
                </label>
                <select
                  value={recipe.isPublic ? 'public' : 'private'}
                  onChange={(e) =>
                    updateRecipe(kind, recipeIndex, {
                      isPublic: e.target.value === 'public',
                    })
                  }
                  style={{
                    width: '100%',
                    padding: '6px',
                    border: '1px solid #ccc',
                    borderRadius: '4px',
                  }}
                >
                  <option value="public">Public</option>
                  <option value="private">Private</option>
                </select>
              </div>
              <div>
                <label
                  style={{
                    display: 'block',
                    marginBottom: '4px',
                    fontSize: '12px',
                  }}
                >
                  Recipe ID
                </label>
                <input
                  type="text"
                  value={recipe.id}
                  onChange={(e) =>
                    updateRecipe(kind, recipeIndex, { id: e.target.value })
                  }
                  placeholder="Auto-generated if blank"
                  style={{
                    width: '100%',
                    padding: '6px',
                    border: '1px solid #ccc',
                    borderRadius: '4px',
                  }}
                />
              </div>
              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(4, 1fr)',
                  gap: '10px',
                  marginTop: '10px',
                }}
              >
                {damageBonusFieldOptions.map(({ key, label }) => (
                  <div key={key}>
                    <label
                      style={{
                        display: 'block',
                        marginBottom: '4px',
                        fontSize: '12px',
                      }}
                    >
                      {label}
                    </label>
                    <input
                      type="number"
                      value={formData[key]}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          [key]: parseInt(e.target.value, 10) || 0,
                        })
                      }
                      style={{
                        width: '100%',
                        padding: '6px',
                        border: '1px solid #ccc',
                        borderRadius: '4px',
                      }}
                    />
                  </div>
                ))}
              </div>
              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(4, 1fr)',
                  gap: '10px',
                  marginTop: '10px',
                }}
              >
                {resistanceFieldOptions.map(({ key, label }) => (
                  <div key={key}>
                    <label
                      style={{
                        display: 'block',
                        marginBottom: '4px',
                        fontSize: '12px',
                      }}
                    >
                      {label}
                    </label>
                    <input
                      type="number"
                      value={formData[key]}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          [key]: parseInt(e.target.value, 10) || 0,
                        })
                      }
                      style={{
                        width: '100%',
                        padding: '6px',
                        border: '1px solid #ccc',
                        borderRadius: '4px',
                      }}
                    />
                  </div>
                ))}
              </div>
            </div>

            <div style={{ marginBottom: '8px' }}>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  marginBottom: '8px',
                }}
              >
                <label style={{ fontSize: '12px', fontWeight: 600 }}>
                  Ingredients
                </label>
                <button
                  type="button"
                  onClick={() => addRecipeIngredient(kind, recipeIndex)}
                  className="bg-slate-700 text-white px-2 py-1 rounded-md text-xs"
                >
                  Add Ingredient
                </button>
              </div>
              {recipe.ingredients.map((ingredient, ingredientIndex) => (
                <div
                  key={`${kind}-${recipeIndex}-ingredient-${ingredientIndex}`}
                  style={{
                    display: 'grid',
                    gridTemplateColumns: 'minmax(0, 1fr) 120px 86px',
                    gap: '8px',
                    marginBottom: '8px',
                  }}
                >
                  <select
                    value={
                      ingredient.itemId > 0 ? String(ingredient.itemId) : ''
                    }
                    onChange={(e) =>
                      updateRecipeIngredient(
                        kind,
                        recipeIndex,
                        ingredientIndex,
                        {
                          itemId: parseInt(e.target.value, 10) || 0,
                        }
                      )
                    }
                    style={{
                      width: '100%',
                      padding: '6px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                    }}
                  >
                    <option value="">Select ingredient item</option>
                    {itemOptions.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                  <input
                    type="number"
                    min="1"
                    value={ingredient.quantity}
                    onChange={(e) =>
                      updateRecipeIngredient(
                        kind,
                        recipeIndex,
                        ingredientIndex,
                        {
                          quantity: parseInt(e.target.value, 10) || 1,
                        }
                      )
                    }
                    style={{
                      width: '100%',
                      padding: '6px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                    }}
                  />
                  <button
                    type="button"
                    className="bg-red-600 text-white px-2 py-1 rounded-md text-xs"
                    onClick={() =>
                      removeRecipeIngredient(kind, recipeIndex, ingredientIndex)
                    }
                  >
                    Remove
                  </button>
                </div>
              ))}
            </div>

            <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
              <button
                type="button"
                className="bg-red-600 text-white px-2 py-1 rounded-md text-xs"
                onClick={() => removeRecipe(kind, recipeIndex)}
              >
                Remove Recipe
              </button>
            </div>
          </div>
        ))}
      </div>
    );
  };

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      setImageFile(file);

      // Create preview URL
      const reader = new FileReader();
      reader.onloadend = () => {
        setImagePreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const formatGenerationStatus = (status?: string) => {
    switch (status) {
      case 'queued':
        return 'Queued';
      case 'in_progress':
        return 'Generating';
      case 'complete':
        return 'Complete';
      case 'failed':
        return 'Failed';
      case 'none':
        return 'Not requested';
      default:
        return 'Unknown';
    }
  };

  const numericValue = (value: string) => {
    if (value.trim() === '') return undefined;
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : undefined;
  };

  const matchRange = (
    value: number | undefined,
    minValue: string,
    maxValue: string,
    defaultValue?: number
  ) => {
    const min = numericValue(minValue);
    const max = numericValue(maxValue);
    let actual = value;
    if (actual === undefined || actual === null) {
      actual = defaultValue;
    }
    if (min === undefined && max === undefined) return true;
    if (actual === undefined || actual === null) return false;
    if (min !== undefined && actual < min) return false;
    if (max !== undefined && actual > max) return false;
    return true;
  };

  const sortOptions = [
    { value: 'name', label: 'Name' },
    { value: 'id', label: 'ID' },
    { value: 'flavorText', label: 'Flavor Text' },
    { value: 'effectText', label: 'Effect Text' },
    { value: 'imageUrl', label: 'Image URL' },
    { value: 'rarityTier', label: 'Rarity' },
    { value: 'equipSlot', label: 'Equip Slot' },
    { value: 'imageGenerationStatus', label: 'Image Status' },
    { value: 'isCaptureType', label: 'Capture Type' },
    { value: 'buyPrice', label: 'Buy Price' },
    { value: 'unlockTier', label: 'Unlock Tier' },
    { value: 'strengthMod', label: 'STR' },
    { value: 'dexterityMod', label: 'DEX' },
    { value: 'constitutionMod', label: 'CON' },
    { value: 'intelligenceMod', label: 'INT' },
    { value: 'wisdomMod', label: 'WIS' },
    { value: 'charismaMod', label: 'CHA' },
    { value: 'handItemCategory', label: 'Hand Item Type' },
    { value: 'handedness', label: 'Handedness' },
    { value: 'damageMin', label: 'Damage Min' },
    { value: 'damageMax', label: 'Damage Max' },
    { value: 'swipesPerAttack', label: 'Swipes Per Attack' },
    { value: 'blockPercentage', label: 'Block %' },
    { value: 'damageBlocked', label: 'Damage Blocked' },
    { value: 'spellDamageBonusPercent', label: 'Spell Bonus %' },
    { value: 'consumeHealthDelta', label: 'Use Health Delta' },
    { value: 'consumeManaDelta', label: 'Use Mana Delta' },
    { value: 'consumeRevivePartyMemberHealth', label: 'Revive One HP' },
    {
      value: 'consumeReviveAllDownedPartyMembersHealth',
      label: 'Revive All HP',
    },
    { value: 'consumeSpellIds', label: 'Use Grants Spells' },
    { value: 'createdAt', label: 'Created At' },
    { value: 'updatedAt', label: 'Updated At' },
  ];

  const rarityRank: Record<string, number> = {
    [Rarity.Common]: 1,
    [Rarity.Uncommon]: 2,
    [Rarity.Epic]: 3,
    [Rarity.Mythic]: 4,
    [Rarity.NotDroppable]: 5,
  };

  const activeFilterCount = useMemo(() => {
    return Object.values(filters).filter((value) => value !== '').length;
  }, [filters]);

  const visibleItems = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    const filtered = items.filter((item) => {
      if (itemTab === 'active' && item.archived) return false;
      if (itemTab === 'archived' && !item.archived) return false;

      const haystack = [
        item.id?.toString(),
        item.name,
        item.flavorText,
        item.effectText,
        item.imageUrl,
        item.rarityTier,
        item.equipSlot,
        item.imageGenerationStatus,
        item.buyPrice?.toString(),
        item.unlockTier?.toString(),
        item.strengthMod?.toString(),
        item.dexterityMod?.toString(),
        item.constitutionMod?.toString(),
        item.intelligenceMod?.toString(),
        item.wisdomMod?.toString(),
        item.charismaMod?.toString(),
        item.handItemCategory,
        item.handedness,
        item.damageMin?.toString(),
        item.damageMax?.toString(),
        item.swipesPerAttack?.toString(),
        item.blockPercentage?.toString(),
        item.damageBlocked?.toString(),
        item.spellDamageBonusPercent?.toString(),
        item.consumeHealthDelta?.toString(),
        item.consumeManaDelta?.toString(),
        item.consumeRevivePartyMemberHealth?.toString(),
        item.consumeReviveAllDownedPartyMembersHealth?.toString(),
        item.consumeStatusesToAdd?.map((status) => status.name).join(' '),
        item.consumeStatusesToRemove?.join(' '),
        item.consumeSpellIds
          ?.map((spellID) => spellNamesByID.get(spellID) ?? spellID)
          .join(' '),
      ]
        .filter(Boolean)
        .join(' ')
        .toLowerCase();

      if (query && !haystack.includes(query)) return false;

      if (filters.rarity && item.rarityTier !== filters.rarity) return false;
      if (filters.equipSlot && (item.equipSlot ?? '') !== filters.equipSlot)
        return false;
      if (
        filters.imageStatus &&
        (item.imageGenerationStatus ?? '') !== filters.imageStatus
      )
        return false;
      if (filters.captureType === 'yes' && !item.isCaptureType) return false;
      if (filters.captureType === 'no' && item.isCaptureType) return false;
      if (filters.equippable === 'yes' && !item.equipSlot) return false;
      if (filters.equippable === 'no' && item.equipSlot) return false;

      if (!matchRange(item.id, filters.minId, filters.maxId)) return false;
      if (!matchRange(item.buyPrice, filters.minBuyPrice, filters.maxBuyPrice))
        return false;
      if (
        !matchRange(
          item.unlockTier,
          filters.minUnlockTier,
          filters.maxUnlockTier
        )
      )
        return false;

      if (
        !matchRange(
          item.strengthMod ?? 0,
          filters.minStrength,
          filters.maxStrength,
          0
        )
      )
        return false;
      if (
        !matchRange(
          item.dexterityMod ?? 0,
          filters.minDexterity,
          filters.maxDexterity,
          0
        )
      )
        return false;
      if (
        !matchRange(
          item.constitutionMod ?? 0,
          filters.minConstitution,
          filters.maxConstitution,
          0
        )
      )
        return false;
      if (
        !matchRange(
          item.intelligenceMod ?? 0,
          filters.minIntelligence,
          filters.maxIntelligence,
          0
        )
      )
        return false;
      if (
        !matchRange(
          item.wisdomMod ?? 0,
          filters.minWisdom,
          filters.maxWisdom,
          0
        )
      )
        return false;
      if (
        !matchRange(
          item.charismaMod ?? 0,
          filters.minCharisma,
          filters.maxCharisma,
          0
        )
      )
        return false;

      return true;
    });

    const sorted = [...filtered].sort((a, b) => {
      const direction = sortDirection === 'asc' ? 1 : -1;
      const field = sortField as keyof InventoryItemRecord;
      if (field === 'rarityTier') {
        const rankA = rarityRank[a.rarityTier] ?? 999;
        const rankB = rarityRank[b.rarityTier] ?? 999;
        return (rankA - rankB) * direction;
      }
      if (field === 'createdAt' || field === 'updatedAt') {
        const timeA = a[field] ? new Date(a[field] as string).getTime() : 0;
        const timeB = b[field] ? new Date(b[field] as string).getTime() : 0;
        return (timeA - timeB) * direction;
      }
      const valueA = a[field];
      const valueB = b[field];
      if (typeof valueA === 'number' || typeof valueB === 'number') {
        const numA = Number(valueA ?? 0);
        const numB = Number(valueB ?? 0);
        return (numA - numB) * direction;
      }
      const strA = (valueA ?? '').toString().toLowerCase();
      const strB = (valueB ?? '').toString().toLowerCase();
      return strA.localeCompare(strB) * direction;
    });

    return sorted;
  }, [
    items,
    itemTab,
    searchQuery,
    filters,
    sortField,
    sortDirection,
    spellNamesByID,
  ]);

  const activeItemCount = useMemo(
    () => items.filter((item) => !item.archived).length,
    [items]
  );
  const archivedItemCount = useMemo(
    () => items.filter((item) => item.archived).length,
    [items]
  );
  const missingImageCount = useMemo(
    () =>
      items.filter((item) => {
        const hasImage = (item.imageUrl ?? '').trim() !== '';
        const status = (item.imageGenerationStatus ?? '').trim().toLowerCase();
        return !hasImage && status !== 'queued' && status !== 'in_progress';
      }).length,
    [items]
  );

  const visibleItemIDs = useMemo(
    () => visibleItems.map((item) => item.id),
    [visibleItems]
  );
  const selectedVisibleCount = useMemo(
    () =>
      visibleItems.reduce(
        (count, item) => count + (selectedItemIDs.has(item.id) ? 1 : 0),
        0
      ),
    [visibleItems, selectedItemIDs]
  );
  const allVisibleSelected =
    visibleItems.length > 0 && selectedVisibleCount === visibleItems.length;
  const hasSelectedItems = selectedItemIDs.size > 0;

  if (loading) {
    return <div className="m-10">Loading inventory items...</div>;
  }

  return (
    <div className="m-10">
      <div className="mb-4 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <h1 className="text-2xl font-bold">Inventory Items</h1>
        <div className="flex flex-wrap gap-2">
          <button
            className={`px-4 py-2 rounded-md border ${
              itemTab === 'active'
                ? 'bg-slate-900 text-white border-slate-900'
                : 'bg-white text-slate-700 border-slate-300'
            }`}
            onClick={() => setItemTab('active')}
          >
            Active ({activeItemCount})
          </button>
          <button
            className={`px-4 py-2 rounded-md border ${
              itemTab === 'archived'
                ? 'bg-slate-900 text-white border-slate-900'
                : 'bg-white text-slate-700 border-slate-300'
            }`}
            onClick={() => setItemTab('archived')}
          >
            Archived ({archivedItemCount})
          </button>
          <button
            className="bg-blue-500 text-white px-4 py-2 rounded-md"
            onClick={() => setShowCreateItem(true)}
          >
            Create Inventory Item
          </button>
          <button
            className="bg-green-600 text-white px-4 py-2 rounded-md"
            onClick={() => setShowGenerateItem(true)}
          >
            Generate Inventory Item
          </button>
          <button
            className="bg-violet-700 text-white px-4 py-2 rounded-md disabled:bg-gray-300 disabled:cursor-not-allowed"
            onClick={() => void handleSeedCorePack()}
            disabled={seedPackBusy}
          >
            {seedPackBusy ? 'Seeding Core Pack...' : 'Seed Core Pack'}
          </button>
          <button
            className="bg-amber-600 text-white px-4 py-2 rounded-md disabled:bg-gray-300 disabled:cursor-not-allowed"
            onClick={() => void handleQueueMissingImages()}
            disabled={queueMissingImagesBusy || missingImageCount === 0}
          >
            {queueMissingImagesBusy
              ? 'Queueing Images...'
              : `Queue Missing Images (${missingImageCount})`}
          </button>
          <button
            className="bg-slate-700 text-white px-4 py-2 rounded-md disabled:bg-gray-300 disabled:cursor-not-allowed"
            onClick={() =>
              void handleSetItemsArchived(
                Array.from(selectedItemIDs),
                itemTab === 'active'
              )
            }
            disabled={!hasSelectedItems}
          >
            {itemTab === 'active'
              ? `Archive Selected (${selectedItemIDs.size})`
              : `Restore Selected (${selectedItemIDs.size})`}
          </button>
          <button
            className="bg-red-600 text-white px-4 py-2 rounded-md disabled:bg-gray-300 disabled:cursor-not-allowed"
            onClick={() => setShowBulkDeleteConfirm(true)}
            disabled={!hasSelectedItems || bulkDeleteBusy}
          >
            {bulkDeleteBusy
              ? 'Deleting...'
              : `Delete Selected (${selectedItemIDs.size})`}
          </button>
        </div>
      </div>

      <div className="mb-5 rounded-md border border-gray-200 bg-gray-50 p-4">
        <div className="mb-2 text-sm font-semibold text-gray-800">
          Generate Full Equippable Set
        </div>
        <div className="grid grid-cols-1 gap-3 md:grid-cols-5">
          <div>
            <label className="mb-1 block text-xs text-gray-600">
              Target Level
            </label>
            <input
              type="number"
              min={1}
              max={100}
              value={bulkSetTargetLevel}
              onChange={(e) => setBulkSetTargetLevel(e.target.value)}
              className="w-full rounded-md border border-gray-300 p-2 text-sm"
            />
          </div>
          <div>
            <label className="mb-1 block text-xs text-gray-600">
              Major Stat
            </label>
            <select
              value={bulkSetMajorStat}
              onChange={(e) => setBulkSetMajorStat(e.target.value)}
              className="w-full rounded-md border border-gray-300 p-2 text-sm"
            >
              {itemSetStatOptions.map((option) => (
                <option key={`major-${option.value}`} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="mb-1 block text-xs text-gray-600">
              Minor Stat
            </label>
            <select
              value={bulkSetMinorStat}
              onChange={(e) => setBulkSetMinorStat(e.target.value)}
              className="w-full rounded-md border border-gray-300 p-2 text-sm"
            >
              {itemSetStatOptions.map((option) => (
                <option key={`minor-${option.value}`} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="mb-1 block text-xs text-gray-600">Rarity</label>
            <select
              value={bulkSetRarityTier}
              onChange={(e) => setBulkSetRarityTier(e.target.value)}
              className="w-full rounded-md border border-gray-300 p-2 text-sm"
            >
              {itemSetRarityOptions.map((option) => (
                <option
                  key={`bulk-set-rarity-${option.value}`}
                  value={option.value}
                >
                  {option.label}
                </option>
              ))}
            </select>
          </div>
          <div className="flex items-end">
            <button
              type="button"
              onClick={handleGenerateSetFromStats}
              disabled={bulkSetGenerationBusy}
              className="w-full rounded-md bg-violet-700 px-4 py-2 text-white disabled:cursor-not-allowed disabled:bg-gray-300"
            >
              {bulkSetGenerationBusy
                ? 'Generating Set...'
                : 'Generate Full Set'}
            </button>
          </div>
        </div>
      </div>

      <div className="mb-6 rounded-xl border border-emerald-200 bg-emerald-50/60 p-4">
        <div className="mb-4 flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
          <div>
            <div className="text-sm font-semibold uppercase tracking-wide text-emerald-700">
              Equipment Families
            </div>
            <div className="mt-1 text-lg font-semibold text-slate-900">
              Generate more seeded-style set families
            </div>
            <p className="mt-1 max-w-3xl text-sm text-slate-600">
              This uses the same slot scaling and affinity identity logic as the
              core equipment seed pack, but lets you steer the level bands,
              profiles, stats, affinities, and naming lane.
            </p>
          </div>
          <button
            type="button"
            onClick={handleGenerateSetFamilies}
            disabled={setFamilyGenerationBusy}
            className="rounded-md bg-emerald-700 px-4 py-2 text-white disabled:cursor-not-allowed disabled:bg-gray-300"
          >
            {setFamilyGenerationBusy
              ? 'Generating Families...'
              : 'Generate Set Families'}
          </button>
        </div>

        <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
          <div className="rounded-lg border border-white/80 bg-white p-4 shadow-sm">
            <div className="mb-3 text-sm font-semibold text-slate-800">
              Identity
            </div>
            <div className="grid grid-cols-1 gap-3 md:grid-cols-3">
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Family Count
                </label>
                <input
                  type="number"
                  min={1}
                  max={40}
                  value={setFamilyCount}
                  onChange={(event) => setSetFamilyCount(event.target.value)}
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Level Min
                </label>
                <input
                  type="number"
                  min={1}
                  max={100}
                  value={setFamilyLevelMin}
                  onChange={(event) => setSetFamilyLevelMin(event.target.value)}
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Level Max
                </label>
                <input
                  type="number"
                  min={1}
                  max={100}
                  value={setFamilyLevelMax}
                  onChange={(event) => setSetFamilyLevelMax(event.target.value)}
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                />
              </div>
            </div>

            <div className="mt-3">
              <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                Theme Seeds
              </label>
              <input
                type="text"
                value={setFamilyThemesInput}
                onChange={(event) =>
                  setSetFamilyThemesInput(event.target.value)
                }
                placeholder="storm knight, fungal court, cathedral hunter"
                className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
              />
            </div>

            <div className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-3">
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Slot Scope
                </label>
                <select
                  value={setFamilySlotScope}
                  onChange={(event) =>
                    setSetFamilySlotScope(event.target.value)
                  }
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                >
                  {inventorySetSlotScopeOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Power Bias
                </label>
                <select
                  value={setFamilyPowerBias}
                  onChange={(event) =>
                    setSetFamilyPowerBias(event.target.value)
                  }
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                >
                  {inventorySetPowerBiasOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Naming Style
                </label>
                <select
                  value={setFamilyNamingStyle}
                  onChange={(event) =>
                    setSetFamilyNamingStyle(event.target.value)
                  }
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                >
                  {inventorySetNamingStyleOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            <div className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-2">
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Required Internal Tags
                </label>
                <input
                  type="text"
                  value={setFamilyRequiredTagsInput}
                  onChange={(event) =>
                    setSetFamilyRequiredTagsInput(event.target.value)
                  }
                  placeholder="frontline, cathedral, occult"
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Forbidden Tags
                </label>
                <input
                  type="text"
                  value={setFamilyForbiddenTagsInput}
                  onChange={(event) =>
                    setSetFamilyForbiddenTagsInput(event.target.value)
                  }
                  placeholder="street, nautical"
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                />
              </div>
            </div>

            <div className="mt-4 grid grid-cols-1 gap-2 md:grid-cols-3">
              <label className="flex items-center gap-2 rounded-md border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700">
                <input
                  type="checkbox"
                  checked={setFamilyAllowHybridAffinities}
                  onChange={(event) =>
                    setSetFamilyAllowHybridAffinities(event.target.checked)
                  }
                />
                Allow hybrid affinities
              </label>
              <label className="flex items-center gap-2 rounded-md border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700">
                <input
                  type="checkbox"
                  checked={setFamilyAvoidExistingThemeOverlap}
                  onChange={(event) =>
                    setSetFamilyAvoidExistingThemeOverlap(event.target.checked)
                  }
                />
                Avoid theme overlap
              </label>
              <label className="flex items-center gap-2 rounded-md border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700">
                <input
                  type="checkbox"
                  checked={setFamilyQueueImages}
                  onChange={(event) =>
                    setSetFamilyQueueImages(event.target.checked)
                  }
                />
                Queue image generation
              </label>
            </div>
          </div>

          <div className="rounded-lg border border-white/80 bg-white p-4 shadow-sm">
            <div className="mb-3 text-sm font-semibold text-slate-800">
              Mechanics
            </div>

            <div className="mb-3">
              <div className="mb-2 text-xs font-medium uppercase tracking-wide text-slate-500">
                Rarity Tiers
              </div>
              <div className="grid grid-cols-2 gap-2 md:grid-cols-3">
                {itemSetRarityOptions
                  .filter((option) => option.value !== 'auto')
                  .map((option) => (
                    <label
                      key={`family-rarity-${option.value}`}
                      className="flex items-center gap-2 rounded-md border border-slate-200 px-3 py-2 text-sm text-slate-700"
                    >
                      <input
                        type="checkbox"
                        checked={setFamilyRarityTiers.has(option.value)}
                        onChange={(event) =>
                          setSetFamilyRarityTiers((current) =>
                            toggleStringSetValue(
                              current,
                              option.value,
                              event.target.checked
                            )
                          )
                        }
                      />
                      {option.label}
                    </label>
                  ))}
              </div>
            </div>

            <div className="mb-3">
              <div className="mb-2 text-xs font-medium uppercase tracking-wide text-slate-500">
                Profiles
              </div>
              <div className="grid grid-cols-2 gap-2 md:grid-cols-3">
                {inventorySetProfileOptions.map((option) => (
                  <label
                    key={`family-profile-${option.value}`}
                    className="flex items-center gap-2 rounded-md border border-slate-200 px-3 py-2 text-sm text-slate-700"
                  >
                    <input
                      type="checkbox"
                      checked={setFamilyProfiles.has(option.value)}
                      onChange={(event) =>
                        setSetFamilyProfiles((current) =>
                          toggleStringSetValue(
                            current,
                            option.value,
                            event.target.checked
                          )
                        )
                      }
                    />
                    {option.label}
                  </label>
                ))}
              </div>
            </div>

            <div className="mb-3 grid grid-cols-1 gap-3 md:grid-cols-2">
              <div>
                <div className="mb-2 text-xs font-medium uppercase tracking-wide text-slate-500">
                  Major Stats
                </div>
                <div className="space-y-2">
                  {itemSetStatOptions.map((option) => (
                    <label
                      key={`family-major-${option.value}`}
                      className="flex items-center gap-2 rounded-md border border-slate-200 px-3 py-2 text-sm text-slate-700"
                    >
                      <input
                        type="checkbox"
                        checked={setFamilyMajorStats.has(option.value)}
                        onChange={(event) =>
                          setSetFamilyMajorStats((current) =>
                            toggleStringSetValue(
                              current,
                              option.value,
                              event.target.checked
                            )
                          )
                        }
                      />
                      {option.label}
                    </label>
                  ))}
                </div>
              </div>
              <div>
                <div className="mb-2 text-xs font-medium uppercase tracking-wide text-slate-500">
                  Minor Stats
                </div>
                <div className="space-y-2">
                  {itemSetStatOptions.map((option) => (
                    <label
                      key={`family-minor-${option.value}`}
                      className="flex items-center gap-2 rounded-md border border-slate-200 px-3 py-2 text-sm text-slate-700"
                    >
                      <input
                        type="checkbox"
                        checked={setFamilyMinorStats.has(option.value)}
                        onChange={(event) =>
                          setSetFamilyMinorStats((current) =>
                            toggleStringSetValue(
                              current,
                              option.value,
                              event.target.checked
                            )
                          )
                        }
                      />
                      {option.label}
                    </label>
                  ))}
                </div>
              </div>
            </div>

            <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
              <div>
                <div className="mb-2 text-xs font-medium uppercase tracking-wide text-slate-500">
                  Damage Affinities
                </div>
                <div className="grid grid-cols-2 gap-2">
                  {damageAffinityOptions.map((option) => (
                    <label
                      key={`family-damage-${option.value}`}
                      className="flex items-center gap-2 rounded-md border border-slate-200 px-3 py-2 text-sm text-slate-700"
                    >
                      <input
                        type="checkbox"
                        checked={setFamilyDamageAffinities.has(option.value)}
                        onChange={(event) =>
                          setSetFamilyDamageAffinities((current) =>
                            toggleStringSetValue(
                              current,
                              option.value,
                              event.target.checked
                            )
                          )
                        }
                      />
                      {option.label}
                    </label>
                  ))}
                </div>
              </div>
              <div>
                <div className="mb-2 text-xs font-medium uppercase tracking-wide text-slate-500">
                  Resistance Affinities
                </div>
                <div className="grid grid-cols-2 gap-2">
                  {damageAffinityOptions.map((option) => (
                    <label
                      key={`family-resistance-${option.value}`}
                      className="flex items-center gap-2 rounded-md border border-slate-200 px-3 py-2 text-sm text-slate-700"
                    >
                      <input
                        type="checkbox"
                        checked={setFamilyResistanceAffinities.has(
                          option.value
                        )}
                        onChange={(event) =>
                          setSetFamilyResistanceAffinities((current) =>
                            toggleStringSetValue(
                              current,
                              option.value,
                              event.target.checked
                            )
                          )
                        }
                      />
                      {option.label}
                    </label>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="mb-6 rounded-xl border border-indigo-200 bg-indigo-50/60 p-4">
        <div className="mb-4 flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
          <div>
            <div className="text-sm font-semibold uppercase tracking-wide text-indigo-700">
              Draft Generator
            </div>
            <div className="mt-1 text-lg font-semibold text-slate-900">
              Agent-first inventory ideation
            </div>
            <p className="mt-1 max-w-3xl text-sm text-slate-600">
              Queue batches of draft items, review the generated concepts, and
              convert only the ones that feel worth keeping.
            </p>
          </div>
          <div className="rounded-lg border border-indigo-200 bg-white px-3 py-2 text-sm text-slate-600">
            Recent jobs: {suggestionJobs.length}
          </div>
        </div>

        <div className="grid grid-cols-1 gap-4 xl:grid-cols-[1.15fr_0.85fr]">
          <div className="rounded-lg border border-white/70 bg-white p-4 shadow-sm">
            <div className="mb-3 text-sm font-semibold text-slate-800">
              Queue Draft Batch
            </div>
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Batch Size
                </label>
                <input
                  type="number"
                  min={1}
                  max={100}
                  value={suggestionForm.count}
                  onChange={(event) =>
                    setSuggestionForm((current) => ({
                      ...current,
                      count: event.target.value,
                    }))
                  }
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Level Band
                </label>
                <div className="grid grid-cols-2 gap-2">
                  <input
                    type="number"
                    min={1}
                    value={suggestionForm.minItemLevel}
                    onChange={(event) =>
                      setSuggestionForm((current) => ({
                        ...current,
                        minItemLevel: event.target.value,
                      }))
                    }
                    placeholder="Min"
                    className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                  />
                  <input
                    type="number"
                    min={1}
                    value={suggestionForm.maxItemLevel}
                    onChange={(event) =>
                      setSuggestionForm((current) => ({
                        ...current,
                        maxItemLevel: event.target.value,
                      }))
                    }
                    placeholder="Max"
                    className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                  />
                </div>
              </div>
              <div className="md:col-span-2">
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Theme Prompt
                </label>
                <textarea
                  rows={4}
                  value={suggestionForm.themePrompt}
                  onChange={(event) =>
                    setSuggestionForm((current) => ({
                      ...current,
                      themePrompt: event.target.value,
                    }))
                  }
                  placeholder="Low-level occult street gear, scavenger kits, and odd consumables for night exploration."
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Categories
                </label>
                <input
                  value={suggestionForm.categoriesText}
                  onChange={(event) =>
                    setSuggestionForm((current) => ({
                      ...current,
                      categoriesText: event.target.value,
                    }))
                  }
                  placeholder="equippable, consumable, material"
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Rarity Tiers
                </label>
                <input
                  value={suggestionForm.rarityTiersText}
                  onChange={(event) =>
                    setSuggestionForm((current) => ({
                      ...current,
                      rarityTiersText: event.target.value,
                    }))
                  }
                  placeholder="Common, Uncommon, Epic"
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Equip Slots
                </label>
                <input
                  value={suggestionForm.equipSlotsText}
                  onChange={(event) =>
                    setSuggestionForm((current) => ({
                      ...current,
                      equipSlotsText: event.target.value,
                    }))
                  }
                  placeholder="hat, necklace, dominant_hand"
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Stat Tags
                </label>
                <input
                  value={suggestionForm.statTagsText}
                  onChange={(event) =>
                    setSuggestionForm((current) => ({
                      ...current,
                      statTagsText: event.target.value,
                    }))
                  }
                  placeholder="strength, intelligence, wisdom"
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Benefit Tags
                </label>
                <input
                  value={suggestionForm.benefitTagsText}
                  onChange={(event) =>
                    setSuggestionForm((current) => ({
                      ...current,
                      benefitTagsText: event.target.value,
                    }))
                  }
                  placeholder="healing, mana, fire damage, resistances"
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Status Names
                </label>
                <input
                  value={suggestionForm.statusNamesText}
                  onChange={(event) =>
                    setSuggestionForm((current) => ({
                      ...current,
                      statusNamesText: event.target.value,
                    }))
                  }
                  placeholder="Blessed, Guarded, Regenerating"
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                  Internal Tags
                </label>
                <input
                  value={suggestionForm.internalTagsText}
                  onChange={(event) =>
                    setSuggestionForm((current) => ({
                      ...current,
                      internalTagsText: event.target.value,
                    }))
                  }
                  placeholder="occult, scavenger, downtown"
                  className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
                />
              </div>
            </div>

            <div className="mt-4 flex flex-wrap items-center gap-3">
              <button
                type="button"
                onClick={() => void handleQueueSuggestionJob()}
                disabled={queueingSuggestionJob}
                className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
              >
                {queueingSuggestionJob ? 'Queueing...' : 'Queue Draft Job'}
              </button>
              <button
                type="button"
                onClick={() =>
                  setSuggestionForm(emptyInventorySuggestionForm())
                }
                className="rounded-md border border-slate-300 bg-white px-4 py-2 text-sm text-slate-700"
              >
                Reset
              </button>
            </div>

            {suggestionError && (
              <div className="mt-3 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                {suggestionError}
              </div>
            )}
          </div>

          <div className="rounded-lg border border-white/70 bg-white p-4 shadow-sm">
            <div className="mb-3 flex items-center justify-between gap-3">
              <div className="text-sm font-semibold text-slate-800">
                Recent Draft Jobs
              </div>
              {loadingSuggestionJobs && (
                <div className="text-xs text-slate-500">Refreshing...</div>
              )}
            </div>
            <div className="space-y-2">
              {suggestionJobs.length === 0 && !loadingSuggestionJobs && (
                <div className="rounded-md border border-dashed border-slate-300 px-3 py-4 text-sm text-slate-500">
                  No draft jobs yet.
                </div>
              )}
              {suggestionJobs.map((job) => {
                const selected = job.id === selectedSuggestionJobId;
                return (
                  <button
                    key={job.id}
                    type="button"
                    onClick={() => setSelectedSuggestionJobId(job.id)}
                    className={`w-full rounded-lg border px-3 py-3 text-left transition ${
                      selected
                        ? 'border-indigo-500 bg-indigo-50'
                        : 'border-slate-200 bg-slate-50 hover:border-slate-300'
                    }`}
                  >
                    <div className="flex items-center justify-between gap-3">
                      <div className="font-medium text-slate-900">
                        {job.themePrompt?.trim() || 'Untitled draft job'}
                      </div>
                      <span
                        className={`rounded-full px-2 py-1 text-[11px] font-semibold uppercase tracking-wide ${
                          job.status === 'completed'
                            ? 'bg-emerald-100 text-emerald-700'
                            : job.status === 'failed'
                              ? 'bg-red-100 text-red-700'
                              : 'bg-amber-100 text-amber-700'
                        }`}
                      >
                        {job.status.replace(/_/g, ' ')}
                      </span>
                    </div>
                    <div className="mt-2 text-xs text-slate-500">
                      {job.createdCount}/{job.count} drafts · levels{' '}
                      {job.minItemLevel}-{job.maxItemLevel}
                    </div>
                    <div className="mt-2 flex flex-wrap gap-2 text-[11px] text-slate-500">
                      {[
                        renderSuggestionFacetLabel('stats', job.statTags),
                        renderSuggestionFacetLabel('benefits', job.benefitTags),
                        renderSuggestionFacetLabel('statuses', job.statusNames),
                      ]
                        .filter((entry): entry is string => Boolean(entry))
                        .map((entry) => (
                          <span
                            key={`${job.id}-${entry}`}
                            className="rounded-full bg-white/80 px-2 py-1"
                          >
                            {entry}
                          </span>
                        ))}
                    </div>
                  </button>
                );
              })}
            </div>
          </div>
        </div>

        <div className="mt-4 rounded-lg border border-white/70 bg-white p-4 shadow-sm">
          <div className="mb-3 flex flex-col gap-1 md:flex-row md:items-center md:justify-between">
            <div>
              <div className="text-sm font-semibold text-slate-800">
                {selectedSuggestionJob
                  ? `Drafts for ${selectedSuggestionJob.themePrompt || 'selected job'}`
                  : 'Generated Drafts'}
              </div>
              {selectedSuggestionJob && (
                <div className="space-y-1 text-xs text-slate-500">
                  <div>
                    {selectedSuggestionJob.createdCount} draft
                    {selectedSuggestionJob.createdCount === 1 ? '' : 's'} ·{' '}
                    {selectedSuggestionJob.categories.join(', ') ||
                      'all categories'}
                  </div>
                  <div className="flex flex-wrap gap-2">
                    {[
                      renderSuggestionFacetLabel(
                        'Stat focus',
                        selectedSuggestionJob.statTags
                      ),
                      renderSuggestionFacetLabel(
                        'Benefit focus',
                        selectedSuggestionJob.benefitTags
                      ),
                      renderSuggestionFacetLabel(
                        'Statuses',
                        selectedSuggestionJob.statusNames
                      ),
                    ]
                      .filter((entry): entry is string => Boolean(entry))
                      .map((entry) => (
                        <span
                          key={`${selectedSuggestionJob.id}-${entry}`}
                          className="rounded-full bg-slate-100 px-2 py-1 text-slate-600"
                        >
                          {entry}
                        </span>
                      ))}
                  </div>
                </div>
              )}
            </div>
            {loadingSuggestionDrafts && (
              <div className="text-xs text-slate-500">Loading drafts...</div>
            )}
          </div>

          <div className="grid grid-cols-1 gap-3 xl:grid-cols-2">
            {suggestionDrafts.length === 0 && !loadingSuggestionDrafts && (
              <div className="rounded-md border border-dashed border-slate-300 px-3 py-6 text-sm text-slate-500 xl:col-span-2">
                {selectedSuggestionJob
                  ? 'This job has not produced any drafts yet.'
                  : 'Select a draft job to review its generated items.'}
              </div>
            )}
            {suggestionDrafts.map((draft) => (
              <div
                key={draft.id}
                className="rounded-lg border border-slate-200 bg-slate-50 p-4"
              >
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <div className="text-lg font-semibold text-slate-900">
                      {draft.name}
                    </div>
                    <div className="mt-1 flex flex-wrap gap-2 text-xs">
                      <span className="rounded-full bg-slate-200 px-2 py-1 text-slate-700">
                        {draft.category}
                      </span>
                      <span className="rounded-full bg-slate-200 px-2 py-1 text-slate-700">
                        {draft.rarityTier}
                      </span>
                      <span className="rounded-full bg-slate-200 px-2 py-1 text-slate-700">
                        Lv {draft.itemLevel}
                      </span>
                      {draft.equipSlot && (
                        <span className="rounded-full bg-slate-200 px-2 py-1 text-slate-700">
                          {draft.equipSlot}
                        </span>
                      )}
                    </div>
                  </div>
                  <span
                    className={`rounded-full px-2 py-1 text-[11px] font-semibold uppercase tracking-wide ${
                      draft.status === 'converted'
                        ? 'bg-emerald-100 text-emerald-700'
                        : 'bg-indigo-100 text-indigo-700'
                    }`}
                  >
                    {draft.status}
                  </span>
                </div>

                <p className="mt-3 text-sm text-slate-700">{draft.whyItFits}</p>
                {draft.payload.item.effectText && (
                  <div className="mt-3 rounded-md bg-white px-3 py-2 text-sm text-slate-700">
                    <span className="font-medium text-slate-900">Effect:</span>{' '}
                    {draft.payload.item.effectText}
                  </div>
                )}
                {draft.payload.item.flavorText && (
                  <div className="mt-3 text-sm text-slate-600">
                    {draft.payload.item.flavorText}
                  </div>
                )}
                {(draft.internalTags?.length ?? 0) > 0 && (
                  <div className="mt-3 flex flex-wrap gap-2">
                    {draft.internalTags.map((tag) => (
                      <span
                        key={`${draft.id}-${tag}`}
                        className="rounded-full bg-indigo-100 px-2 py-1 text-xs text-indigo-700"
                      >
                        {tag}
                      </span>
                    ))}
                  </div>
                )}
                {(draft.warnings?.length ?? 0) > 0 && (
                  <div className="mt-3 rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">
                    {draft.warnings.join(' ')}
                  </div>
                )}

                <div className="mt-4 flex flex-wrap gap-2">
                  <button
                    type="button"
                    onClick={() => void handleConvertSuggestionDraft(draft.id)}
                    disabled={
                      draft.status === 'converted' ||
                      convertingSuggestionDraftId === draft.id
                    }
                    className="rounded-md bg-emerald-600 px-3 py-2 text-sm font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
                  >
                    {draft.status === 'converted'
                      ? 'Converted'
                      : convertingSuggestionDraftId === draft.id
                        ? 'Converting...'
                        : 'Convert to Item'}
                  </button>
                  <button
                    type="button"
                    onClick={() => void handleDeleteSuggestionDraft(draft.id)}
                    disabled={
                      draft.status === 'converted' ||
                      deletingSuggestionDraftId === draft.id
                    }
                    className="rounded-md border border-slate-300 bg-white px-3 py-2 text-sm text-slate-700 disabled:cursor-not-allowed disabled:bg-slate-100"
                  >
                    {deletingSuggestionDraftId === draft.id
                      ? 'Deleting...'
                      : 'Delete Draft'}
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Search + Sort */}
      <div className="mb-4 flex flex-col gap-3 md:flex-row md:items-center">
        <div className="flex-1">
          <input
            type="text"
            placeholder="Search inventory items..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full p-2 border rounded-md"
          />
        </div>
        <div className="flex w-full flex-col gap-3 md:w-auto md:flex-row">
          <select
            value={sortField}
            onChange={(e) => setSortField(e.target.value)}
            className="w-full p-2 border rounded-md md:w-56"
          >
            {sortOptions.map((option) => (
              <option key={option.value} value={option.value}>
                Sort: {option.label}
              </option>
            ))}
          </select>
          <button
            type="button"
            onClick={() =>
              setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
            }
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-700 md:w-44"
          >
            Direction: {sortDirection === 'asc' ? 'Ascending' : 'Descending'}
          </button>
          <button
            type="button"
            onClick={() => setShowFilters(!showFilters)}
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-700 md:w-36"
          >
            {showFilters ? 'Hide filters' : 'Show filters'}
            {activeFilterCount > 0 ? ` (${activeFilterCount})` : ''}
          </button>
        </div>
      </div>

      {showFilters && (
        <div className="mb-6 rounded-md border border-gray-200 bg-gray-50 p-4">
          <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
            <div className="text-sm font-semibold text-gray-700">
              Filters
              {activeFilterCount > 0 ? ` (${activeFilterCount} active)` : ''}
            </div>
            <button
              type="button"
              onClick={() =>
                setFilters({
                  rarity: '',
                  equipSlot: '',
                  imageStatus: '',
                  captureType: '',
                  equippable: '',
                  minId: '',
                  maxId: '',
                  minBuyPrice: '',
                  maxBuyPrice: '',
                  minUnlockTier: '',
                  maxUnlockTier: '',
                  minStrength: '',
                  maxStrength: '',
                  minDexterity: '',
                  maxDexterity: '',
                  minConstitution: '',
                  maxConstitution: '',
                  minIntelligence: '',
                  maxIntelligence: '',
                  minWisdom: '',
                  maxWisdom: '',
                  minCharisma: '',
                  maxCharisma: '',
                })
              }
              className="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700"
            >
              Clear filters
            </button>
          </div>
          <div className="space-y-3">
            <details className="rounded-md border border-gray-200 bg-white px-3 py-2">
              <summary className="cursor-pointer text-sm font-medium text-gray-700">
                Quick filters
              </summary>
              <div className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-4">
                <select
                  value={filters.rarity}
                  onChange={(e) =>
                    setFilters({ ...filters, rarity: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                >
                  <option value="">All rarities</option>
                  <option value={Rarity.Common}>Common</option>
                  <option value={Rarity.Uncommon}>Uncommon</option>
                  <option value={Rarity.Epic}>Epic</option>
                  <option value={Rarity.Mythic}>Mythic</option>
                  <option value={Rarity.NotDroppable}>Not Droppable</option>
                </select>
                <select
                  value={filters.equipSlot}
                  onChange={(e) =>
                    setFilters({ ...filters, equipSlot: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                >
                  <option value="">All equip slots</option>
                  {equipSlotOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
                <select
                  value={filters.equippable}
                  onChange={(e) =>
                    setFilters({ ...filters, equippable: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                >
                  <option value="">All items</option>
                  <option value="yes">Equippable</option>
                  <option value="no">Not equippable</option>
                </select>
                <select
                  value={filters.captureType}
                  onChange={(e) =>
                    setFilters({ ...filters, captureType: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                >
                  <option value="">All capture types</option>
                  <option value="yes">Capture items</option>
                  <option value="no">Non-capture</option>
                </select>
                <select
                  value={filters.imageStatus}
                  onChange={(e) =>
                    setFilters({ ...filters, imageStatus: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                >
                  <option value="">All image statuses</option>
                  <option value="queued">Queued</option>
                  <option value="in_progress">Generating</option>
                  <option value="complete">Complete</option>
                  <option value="failed">Failed</option>
                  <option value="none">Not requested</option>
                </select>
              </div>
            </details>
            <details className="rounded-md border border-gray-200 bg-white px-3 py-2">
              <summary className="cursor-pointer text-sm font-medium text-gray-700">
                IDs & values
              </summary>
              <div className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-4">
                <input
                  type="number"
                  placeholder="Min ID"
                  value={filters.minId}
                  onChange={(e) =>
                    setFilters({ ...filters, minId: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
                <input
                  type="number"
                  placeholder="Max ID"
                  value={filters.maxId}
                  onChange={(e) =>
                    setFilters({ ...filters, maxId: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
                <input
                  type="number"
                  placeholder="Min buy price"
                  value={filters.minBuyPrice}
                  onChange={(e) =>
                    setFilters({ ...filters, minBuyPrice: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
                <input
                  type="number"
                  placeholder="Max buy price"
                  value={filters.maxBuyPrice}
                  onChange={(e) =>
                    setFilters({ ...filters, maxBuyPrice: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
                <input
                  type="number"
                  placeholder="Min unlock tier"
                  value={filters.minUnlockTier}
                  onChange={(e) =>
                    setFilters({ ...filters, minUnlockTier: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
                <input
                  type="number"
                  placeholder="Max unlock tier"
                  value={filters.maxUnlockTier}
                  onChange={(e) =>
                    setFilters({ ...filters, maxUnlockTier: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
              </div>
            </details>
            <details className="rounded-md border border-gray-200 bg-white px-3 py-2">
              <summary className="cursor-pointer text-sm font-medium text-gray-700">
                Stat bonuses
              </summary>
              <div className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-6">
                <input
                  type="number"
                  placeholder="Min STR"
                  value={filters.minStrength}
                  onChange={(e) =>
                    setFilters({ ...filters, minStrength: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
                <input
                  type="number"
                  placeholder="Max STR"
                  value={filters.maxStrength}
                  onChange={(e) =>
                    setFilters({ ...filters, maxStrength: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
                <input
                  type="number"
                  placeholder="Min DEX"
                  value={filters.minDexterity}
                  onChange={(e) =>
                    setFilters({ ...filters, minDexterity: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
                <input
                  type="number"
                  placeholder="Max DEX"
                  value={filters.maxDexterity}
                  onChange={(e) =>
                    setFilters({ ...filters, maxDexterity: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
                <input
                  type="number"
                  placeholder="Min CON"
                  value={filters.minConstitution}
                  onChange={(e) =>
                    setFilters({ ...filters, minConstitution: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
                <input
                  type="number"
                  placeholder="Max CON"
                  value={filters.maxConstitution}
                  onChange={(e) =>
                    setFilters({ ...filters, maxConstitution: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
                <input
                  type="number"
                  placeholder="Min INT"
                  value={filters.minIntelligence}
                  onChange={(e) =>
                    setFilters({ ...filters, minIntelligence: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
                <input
                  type="number"
                  placeholder="Max INT"
                  value={filters.maxIntelligence}
                  onChange={(e) =>
                    setFilters({ ...filters, maxIntelligence: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
                <input
                  type="number"
                  placeholder="Min WIS"
                  value={filters.minWisdom}
                  onChange={(e) =>
                    setFilters({ ...filters, minWisdom: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
                <input
                  type="number"
                  placeholder="Max WIS"
                  value={filters.maxWisdom}
                  onChange={(e) =>
                    setFilters({ ...filters, maxWisdom: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
                <input
                  type="number"
                  placeholder="Min CHA"
                  value={filters.minCharisma}
                  onChange={(e) =>
                    setFilters({ ...filters, minCharisma: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
                <input
                  type="number"
                  placeholder="Max CHA"
                  value={filters.maxCharisma}
                  onChange={(e) =>
                    setFilters({ ...filters, maxCharisma: e.target.value })
                  }
                  className="w-full p-2 border rounded-md"
                />
              </div>
            </details>
          </div>
        </div>
      )}

      <div className="mb-4 flex flex-col gap-2 rounded-md border border-gray-200 bg-gray-50 p-3 md:flex-row md:items-center md:justify-between">
        <label className="inline-flex items-center gap-2 text-sm text-gray-700">
          <input
            type="checkbox"
            checked={allVisibleSelected}
            onChange={(e) =>
              toggleSelectAllVisible(e.target.checked, visibleItemIDs)
            }
            className="h-4 w-4 cursor-pointer"
          />
          Select all visible ({visibleItems.length})
        </label>
        <span className="text-sm text-gray-700">
          {selectedItemIDs.size} selected
        </span>
      </div>

      <div className="mb-6 rounded-md border border-amber-200 bg-amber-50 p-4">
        <div className="mb-2 text-sm font-semibold text-amber-950">
          Bulk Edit Selected Items
        </div>
        <div className="grid grid-cols-1 gap-3 md:grid-cols-4">
          <div>
            <label className="mb-1 block text-xs text-gray-600">
              Set Buy Price
            </label>
            <input
              type="number"
              min={0}
              step={1}
              value={bulkBuyPriceInput}
              onChange={(e) => setBulkBuyPriceInput(e.target.value)}
              placeholder="Leave blank to keep"
              disabled={bulkEditBusy || !hasSelectedItems}
              className="w-full rounded-md border border-gray-300 p-2 text-sm disabled:cursor-not-allowed disabled:bg-gray-100"
            />
          </div>
          <div>
            <label className="mb-1 block text-xs text-gray-600">
              Tag Action
            </label>
            <select
              value={bulkTagAction}
              onChange={(e) =>
                setBulkTagAction(e.target.value as BulkTagAction)
              }
              disabled={bulkEditBusy || !hasSelectedItems}
              className="w-full rounded-md border border-gray-300 p-2 text-sm disabled:cursor-not-allowed disabled:bg-gray-100"
            >
              <option value="none">No tag changes</option>
              <option value="replace">Replace tags</option>
              <option value="add">Add tags</option>
              <option value="remove">Remove tags</option>
              <option value="clear">Clear all tags</option>
            </select>
          </div>
          <div>
            <label className="mb-1 block text-xs text-gray-600">Tags</label>
            <input
              type="text"
              value={bulkTagsInput}
              onChange={(e) => setBulkTagsInput(e.target.value)}
              placeholder="comma,separated,tags"
              disabled={
                bulkEditBusy ||
                !hasSelectedItems ||
                bulkTagAction === 'none' ||
                bulkTagAction === 'clear'
              }
              className="w-full rounded-md border border-gray-300 p-2 text-sm disabled:cursor-not-allowed disabled:bg-gray-100"
            />
          </div>
          <div className="flex items-end">
            <button
              type="button"
              onClick={handleBulkEditSelected}
              disabled={!hasSelectedItems || bulkEditBusy}
              className="w-full rounded-md bg-amber-600 px-4 py-2 text-white disabled:cursor-not-allowed disabled:bg-gray-300"
            >
              {bulkEditBusy
                ? 'Applying...'
                : `Apply to ${selectedItemIDs.size} selected`}
            </button>
          </div>
        </div>
        <p className="mt-2 text-xs text-gray-600">
          Leave buy price blank to keep current values. Tags are normalized to
          lowercase.
        </p>
      </div>

      {/* Items Grid */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
          gap: '20px',
          padding: '20px',
        }}
      >
        {visibleItems.map((item) => (
          <div
            key={item.id}
            style={{
              padding: '20px',
              border: '1px solid #ccc',
              borderRadius: '8px',
              backgroundColor: '#fff',
              boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
            }}
          >
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'flex-start',
                gap: '12px',
              }}
            >
              <h2
                style={{
                  margin: '0 0 15px 0',
                  color: '#333',
                }}
              >
                {item.name}
              </h2>
              <input
                type="checkbox"
                checked={selectedItemIDs.has(item.id)}
                onChange={(e) => toggleItemSelection(item.id, e.target.checked)}
                style={{ width: 18, height: 18, cursor: 'pointer' }}
                aria-label={`Select ${item.name}`}
              />
            </div>

            <p style={{ margin: '5px 0', color: '#666' }}>ID: {item.id}</p>

            <p
              style={{
                margin: '5px 0',
                color: item.archived ? '#92400e' : '#166534',
              }}
            >
              Status: {item.archived ? 'Archived' : 'Active'}
            </p>

            <p style={{ margin: '5px 0', color: '#666' }}>
              Image Status: {formatGenerationStatus(item.imageGenerationStatus)}
            </p>
            {item.imageGenerationStatus === 'failed' &&
              item.imageGenerationError && (
                <p
                  style={{
                    margin: '5px 0',
                    color: '#b91c1c',
                    fontSize: '12px',
                  }}
                >
                  Error: {item.imageGenerationError}
                </p>
              )}

            <p style={{ margin: '5px 0', color: '#666' }}>
              Rarity: {item.rarityTier}
            </p>

            <p style={{ margin: '5px 0', color: '#666' }}>
              Item Level: {item.itemLevel ?? 1}
            </p>

            <p style={{ margin: '5px 0', color: '#666' }}>
              Capture Type: {item.isCaptureType ? 'Yes' : 'No'}
            </p>

            <p style={{ margin: '5px 0', color: '#666' }}>
              Equip Slot: {equipSlotLabel(item.equipSlot)}
            </p>
            {handCombatSummary(item).map((line) => (
              <p
                key={`${item.id}-${line}`}
                style={{ margin: '5px 0', color: '#666' }}
              >
                {line}
              </p>
            ))}
            {consumeSummary(item, spellNamesByID).map((line) => (
              <p
                key={`${item.id}-consume-${line}`}
                style={{ margin: '5px 0', color: '#666' }}
              >
                {line}
              </p>
            ))}
            {(item.internalTags?.length ?? 0) > 0 && (
              <p style={{ margin: '5px 0', color: '#666' }}>
                Internal Tags: {item.internalTags?.join(', ')}
              </p>
            )}

            {statModSummary(item) && (
              <p style={{ margin: '5px 0', color: '#666' }}>
                Stat Mods: {statModSummary(item)}
              </p>
            )}

            {item.buyPrice !== undefined && item.buyPrice !== null && (
              <p style={{ margin: '5px 0', color: '#666' }}>
                Buy Price: {item.buyPrice} gold
              </p>
            )}

            {item.imageUrl && (
              <img
                src={item.imageUrl}
                alt={item.name}
                style={{
                  maxWidth: '100%',
                  maxHeight: 120,
                  borderRadius: 4,
                  marginTop: '10px',
                }}
              />
            )}

            <p style={{ margin: '10px 0', color: '#666', fontSize: '14px' }}>
              <strong>Flavor:</strong> {item.flavorText || '—'}
            </p>

            <p style={{ margin: '10px 0', color: '#666', fontSize: '14px' }}>
              <strong>Effect:</strong> {item.effectText || '—'}
            </p>

            <div style={{ marginTop: '15px' }}>
              <button
                onClick={() => handleEditItem(item)}
                className="bg-blue-500 text-white px-4 py-2 rounded-md mr-2"
              >
                Edit
              </button>
              <button
                onClick={() =>
                  void handleSetItemsArchived([item.id], !item.archived)
                }
                className="bg-slate-700 text-white px-4 py-2 rounded-md mr-2"
              >
                {item.archived ? 'Restore' : 'Archive'}
              </button>
              {isOutfitName(item.name) && (
                <button
                  onClick={() => {
                    setUseOutfitItem(item);
                    setUseOutfitUser('');
                    setUseOutfitSelfieUrl('');
                    setUseOutfitStatus(null);
                    setUseOutfitStatusKind(null);
                  }}
                  className="bg-indigo-600 text-white px-4 py-2 rounded-md mr-2"
                >
                  Use Outfit
                </button>
              )}
              <button
                onClick={() => handleRegenerateImage(item)}
                className="bg-yellow-500 text-white px-4 py-2 rounded-md mr-2"
                disabled={['queued', 'in_progress'].includes(
                  item.imageGenerationStatus || ''
                )}
              >
                Regenerate Image
              </button>
              {canGenerateConsumableQualities(item) && (
                <button
                  onClick={() => handleGenerateConsumableQualities(item)}
                  className="bg-orange-600 text-white px-4 py-2 rounded-md mr-2 disabled:bg-gray-300 disabled:cursor-not-allowed"
                  disabled={consumableGenerationBusyIds.has(item.id)}
                >
                  {consumableGenerationBusyIds.has(item.id)
                    ? 'Generating Qualities...'
                    : 'Generate Qualities'}
                </button>
              )}
              {item.equipSlot && (
                <button
                  onClick={() => handleGenerateSet(item)}
                  className="bg-violet-600 text-white px-4 py-2 rounded-md mr-2 disabled:bg-gray-300 disabled:cursor-not-allowed"
                  disabled={setGenerationBusyIds.has(item.id)}
                >
                  {setGenerationBusyIds.has(item.id)
                    ? 'Generating Set...'
                    : 'Generate Set'}
                </button>
              )}
              <button
                onClick={() => void handleGenerateProgressionDrafts(item)}
                className="bg-indigo-700 text-white px-4 py-2 rounded-md mr-2 disabled:bg-gray-300 disabled:cursor-not-allowed"
                disabled={progressionGenerationBusyIds.has(item.id)}
              >
                {progressionGenerationBusyIds.has(item.id)
                  ? 'Generating Progression...'
                  : 'Generate Progression'}
              </button>
              <button
                onClick={() => handleDeleteItem(item)}
                className="bg-red-500 text-white px-4 py-2 rounded-md"
              >
                Delete
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Create/Edit Item Modal */}
      {(showCreateItem || editingItem) && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            backgroundColor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            zIndex: 1000,
          }}
        >
          <div
            style={{
              backgroundColor: '#fff',
              padding: '30px',
              borderRadius: '8px',
              width: '600px',
              maxHeight: '80vh',
              overflow: 'auto',
            }}
          >
            <h2>
              {editingItem ? 'Edit Inventory Item' : 'Create Inventory Item'}
            </h2>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Name *:
              </label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) =>
                  setFormData({ ...formData, name: e.target.value })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
                required
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Image:
              </label>
              <input
                type="file"
                accept="image/*"
                ref={fileInputRef}
                onChange={handleImageChange}
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
              />
              {imagePreview && (
                <img
                  src={imagePreview}
                  alt="Preview"
                  style={{
                    maxWidth: '100%',
                    maxHeight: 200,
                    borderRadius: 4,
                    marginTop: '10px',
                    objectFit: 'contain',
                  }}
                />
              )}
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Flavor Text:
              </label>
              <textarea
                value={formData.flavorText}
                onChange={(e) =>
                  setFormData({ ...formData, flavorText: e.target.value })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                  minHeight: '60px',
                }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Effect Text:
              </label>
              <textarea
                value={formData.effectText}
                onChange={(e) =>
                  setFormData({ ...formData, effectText: e.target.value })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                  minHeight: '60px',
                }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Internal Tags (comma-separated):
              </label>
              <input
                type="text"
                value={internalTagsInput}
                onChange={(e) => setInternalTagsInput(e.target.value)}
                placeholder="e.g. consumable, potion, healing, seed_drop_only"
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
              />
              <small style={{ color: '#666', fontSize: '12px' }}>
                Used only for internal classification; not shown in
                player-facing gameplay UI.
              </small>
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Rarity Tier *:
              </label>
              <select
                value={formData.rarityTier}
                onChange={(e) =>
                  setFormData({ ...formData, rarityTier: e.target.value })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
                required
              >
                <option value={Rarity.Common}>Common</option>
                <option value={Rarity.Uncommon}>Uncommon</option>
                <option value={Rarity.Epic}>Epic</option>
                <option value={Rarity.Mythic}>Mythic</option>
                <option value="Not Droppable">Not Droppable</option>
              </select>
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label
                style={{ display: 'flex', alignItems: 'center', gap: '8px' }}
              >
                <input
                  type="checkbox"
                  checked={formData.isCaptureType}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      isCaptureType: e.target.checked,
                    })
                  }
                />
                Is Capture Type
              </label>
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Buy Price (gold):
              </label>
              <input
                type="number"
                min="0"
                value={formData.buyPrice !== undefined ? formData.buyPrice : ''}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    buyPrice:
                      e.target.value === ''
                        ? undefined
                        : parseInt(e.target.value, 10),
                  })
                }
                placeholder="Leave empty if shops should not use a fixed buy price"
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
              />
              <small style={{ color: '#666', fontSize: '12px' }}>
                Base vendor price. Shops sell for this amount before charisma
                discounts, and buy from players for half before charisma
                bonuses.
              </small>
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Unlock Tier:
              </label>
              <input
                type="number"
                min="0"
                value={
                  formData.unlockTier !== undefined ? formData.unlockTier : ''
                }
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    unlockTier:
                      e.target.value === ''
                        ? undefined
                        : parseInt(e.target.value, 10),
                  })
                }
                placeholder="Leave empty if no unlock tier required"
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
              />
              <small style={{ color: '#666', fontSize: '12px' }}>
                Set the tier level required to unlock this item. Leave empty if
                no tier requirement.
              </small>
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Unlock Locks Strength:
              </label>
              <input
                type="number"
                min="1"
                max="100"
                value={
                  formData.unlockLocksStrength !== undefined
                    ? formData.unlockLocksStrength
                    : ''
                }
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    unlockLocksStrength:
                      e.target.value === ''
                        ? undefined
                        : parseInt(e.target.value, 10),
                  })
                }
                placeholder="Leave empty if this item cannot unlock locks"
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
              />
              <small style={{ color: '#666', fontSize: '12px' }}>
                Items with this effect can unlock chests or doors with lock
                strength less than or equal to this value.
              </small>
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Item Level *:
              </label>
              <input
                type="number"
                min="1"
                value={formData.itemLevel}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    itemLevel:
                      e.target.value === ''
                        ? 1
                        : Math.max(1, parseInt(e.target.value, 10) || 1),
                  })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
              />
              <small style={{ color: '#666', fontSize: '12px' }}>
                Used for balancing and progression. Must be at least 1.
              </small>
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Equip Slot:
              </label>
              <select
                value={formData.equipSlot}
                onChange={(e) => handleEquipSlotChange(e.target.value)}
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
              >
                {equipSlotOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
              <small style={{ color: '#666', fontSize: '12px' }}>
                Choose a slot to make the item equippable. Leave as not
                equippable for consumables.
              </small>
            </div>

            {isHandEquipSlot(formData.equipSlot) && (
              <div
                style={{
                  marginBottom: '15px',
                  padding: '12px',
                  border: '1px solid #e5e7eb',
                  borderRadius: '6px',
                }}
              >
                <label
                  style={{
                    display: 'block',
                    marginBottom: '8px',
                    fontWeight: 600,
                  }}
                >
                  Hand Equipment Settings
                </label>

                <div style={{ marginBottom: '10px' }}>
                  <label
                    style={{
                      display: 'block',
                      marginBottom: '4px',
                      fontSize: '12px',
                    }}
                  >
                    Hand Item Type *
                  </label>
                  <select
                    value={formData.handItemCategory}
                    onChange={(e) => handleHandCategoryChange(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '8px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                    }}
                  >
                    <option value="">Select hand item type</option>
                    {(handItemCategoryOptions[formData.equipSlot] || []).map(
                      (option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      )
                    )}
                  </select>
                </div>

                <div style={{ marginBottom: '10px' }}>
                  <label
                    style={{
                      display: 'block',
                      marginBottom: '4px',
                      fontSize: '12px',
                    }}
                  >
                    Handedness *
                  </label>
                  <select
                    value={formData.handedness}
                    onChange={(e) =>
                      setFormData({ ...formData, handedness: e.target.value })
                    }
                    disabled={
                      formData.equipSlot === 'off_hand' ||
                      formData.handItemCategory === 'staff'
                    }
                    style={{
                      width: '100%',
                      padding: '8px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                    }}
                  >
                    <option value="">Select handedness</option>
                    {handednessOptions.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </div>

                {formData.equipSlot === 'dominant_hand' && (
                  <div
                    style={{
                      display: 'grid',
                      gridTemplateColumns: 'repeat(4, 1fr)',
                      gap: '10px',
                      marginBottom: '10px',
                    }}
                  >
                    <div>
                      <label
                        style={{
                          display: 'block',
                          marginBottom: '4px',
                          fontSize: '12px',
                        }}
                      >
                        Damage Min *
                      </label>
                      <input
                        type="number"
                        min="1"
                        value={
                          formData.damageMin !== undefined
                            ? formData.damageMin
                            : ''
                        }
                        onChange={(e) =>
                          setFormData({
                            ...formData,
                            damageMin:
                              e.target.value === ''
                                ? undefined
                                : parseInt(e.target.value, 10),
                          })
                        }
                        style={{
                          width: '100%',
                          padding: '6px',
                          border: '1px solid #ccc',
                          borderRadius: '4px',
                        }}
                      />
                    </div>
                    <div>
                      <label
                        style={{
                          display: 'block',
                          marginBottom: '4px',
                          fontSize: '12px',
                        }}
                      >
                        Damage Max *
                      </label>
                      <input
                        type="number"
                        min="1"
                        value={
                          formData.damageMax !== undefined
                            ? formData.damageMax
                            : ''
                        }
                        onChange={(e) =>
                          setFormData({
                            ...formData,
                            damageMax:
                              e.target.value === ''
                                ? undefined
                                : parseInt(e.target.value, 10),
                          })
                        }
                        style={{
                          width: '100%',
                          padding: '6px',
                          border: '1px solid #ccc',
                          borderRadius: '4px',
                        }}
                      />
                    </div>
                    <div>
                      <label
                        style={{
                          display: 'block',
                          marginBottom: '4px',
                          fontSize: '12px',
                        }}
                      >
                        Swipes / Attack *
                      </label>
                      <input
                        type="number"
                        min="1"
                        value={
                          formData.swipesPerAttack !== undefined
                            ? formData.swipesPerAttack
                            : ''
                        }
                        onChange={(e) =>
                          setFormData({
                            ...formData,
                            swipesPerAttack:
                              e.target.value === ''
                                ? undefined
                                : parseInt(e.target.value, 10),
                          })
                        }
                        style={{
                          width: '100%',
                          padding: '6px',
                          border: '1px solid #ccc',
                          borderRadius: '4px',
                        }}
                      />
                    </div>
                    <div>
                      <label
                        style={{
                          display: 'block',
                          marginBottom: '4px',
                          fontSize: '12px',
                        }}
                      >
                        Damage Affinity *
                      </label>
                      <select
                        value={formData.damageAffinity ?? 'physical'}
                        onChange={(e) =>
                          setFormData({
                            ...formData,
                            damageAffinity: e.target.value,
                          })
                        }
                        style={{
                          width: '100%',
                          padding: '6px',
                          border: '1px solid #ccc',
                          borderRadius: '4px',
                        }}
                      >
                        {damageAffinityOptions.map((option) => (
                          <option key={option.value} value={option.value}>
                            {option.label}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>
                )}

                {formData.handItemCategory === 'shield' && (
                  <div
                    style={{
                      display: 'grid',
                      gridTemplateColumns: 'repeat(2, 1fr)',
                      gap: '10px',
                      marginBottom: '10px',
                    }}
                  >
                    <div>
                      <label
                        style={{
                          display: 'block',
                          marginBottom: '4px',
                          fontSize: '12px',
                        }}
                      >
                        Block Percentage *
                      </label>
                      <input
                        type="number"
                        min="1"
                        max="100"
                        value={
                          formData.blockPercentage !== undefined
                            ? formData.blockPercentage
                            : ''
                        }
                        onChange={(e) =>
                          setFormData({
                            ...formData,
                            blockPercentage:
                              e.target.value === ''
                                ? undefined
                                : parseInt(e.target.value, 10),
                          })
                        }
                        style={{
                          width: '100%',
                          padding: '6px',
                          border: '1px solid #ccc',
                          borderRadius: '4px',
                        }}
                      />
                    </div>
                    <div>
                      <label
                        style={{
                          display: 'block',
                          marginBottom: '4px',
                          fontSize: '12px',
                        }}
                      >
                        Damage Blocked *
                      </label>
                      <input
                        type="number"
                        min="1"
                        value={
                          formData.damageBlocked !== undefined
                            ? formData.damageBlocked
                            : ''
                        }
                        onChange={(e) =>
                          setFormData({
                            ...formData,
                            damageBlocked:
                              e.target.value === ''
                                ? undefined
                                : parseInt(e.target.value, 10),
                          })
                        }
                        style={{
                          width: '100%',
                          padding: '6px',
                          border: '1px solid #ccc',
                          borderRadius: '4px',
                        }}
                      />
                    </div>
                  </div>
                )}

                {(formData.handItemCategory === 'orb' ||
                  formData.handItemCategory === 'staff') && (
                  <div>
                    <label
                      style={{
                        display: 'block',
                        marginBottom: '4px',
                        fontSize: '12px',
                      }}
                    >
                      Spell Damage Bonus % *
                    </label>
                    <input
                      type="number"
                      min="1"
                      value={
                        formData.spellDamageBonusPercent !== undefined
                          ? formData.spellDamageBonusPercent
                          : ''
                      }
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          spellDamageBonusPercent:
                            e.target.value === ''
                              ? undefined
                              : parseInt(e.target.value, 10),
                        })
                      }
                      style={{
                        width: '100%',
                        padding: '6px',
                        border: '1px solid #ccc',
                        borderRadius: '4px',
                      }}
                    />
                  </div>
                )}
              </div>
            )}

            <div
              style={{
                marginBottom: '15px',
                padding: '12px',
                border: '1px solid #e5e7eb',
                borderRadius: '6px',
              }}
            >
              <label
                style={{
                  display: 'block',
                  marginBottom: '8px',
                  fontWeight: 600,
                }}
              >
                Consume Effects
              </label>
              <small
                style={{
                  color: '#666',
                  fontSize: '12px',
                  display: 'block',
                  marginBottom: '10px',
                }}
              >
                Positive deltas restore resources. Revive values set HP when
                reviving.
              </small>

              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(2, 1fr)',
                  gap: '10px',
                  marginBottom: '12px',
                }}
              >
                <div>
                  <label
                    style={{
                      display: 'block',
                      marginBottom: '4px',
                      fontSize: '12px',
                    }}
                  >
                    Health Delta
                  </label>
                  <input
                    type="number"
                    value={formData.consumeHealthDelta}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        consumeHealthDelta: parseInt(e.target.value, 10) || 0,
                      })
                    }
                    style={{
                      width: '100%',
                      padding: '6px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                    }}
                  />
                </div>
                <div>
                  <label
                    style={{
                      display: 'block',
                      marginBottom: '4px',
                      fontSize: '12px',
                    }}
                  >
                    Mana Delta
                  </label>
                  <input
                    type="number"
                    value={formData.consumeManaDelta}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        consumeManaDelta: parseInt(e.target.value, 10) || 0,
                      })
                    }
                    style={{
                      width: '100%',
                      padding: '6px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                    }}
                  />
                </div>
                <div>
                  <label
                    style={{
                      display: 'block',
                      marginBottom: '4px',
                      fontSize: '12px',
                    }}
                  >
                    Revive Party Member HP
                  </label>
                  <input
                    type="number"
                    min="0"
                    value={formData.consumeRevivePartyMemberHealth}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        consumeRevivePartyMemberHealth:
                          parseInt(e.target.value, 10) || 0,
                      })
                    }
                    style={{
                      width: '100%',
                      padding: '6px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                    }}
                  />
                </div>
                <div>
                  <label
                    style={{
                      display: 'block',
                      marginBottom: '4px',
                      fontSize: '12px',
                    }}
                  >
                    Revive All Downed HP
                  </label>
                  <input
                    type="number"
                    min="0"
                    value={formData.consumeReviveAllDownedPartyMembersHealth}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        consumeReviveAllDownedPartyMembersHealth:
                          parseInt(e.target.value, 10) || 0,
                      })
                    }
                    style={{
                      width: '100%',
                      padding: '6px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                    }}
                  />
                </div>
                <label
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '8px',
                    fontSize: '13px',
                    fontWeight: 600,
                  }}
                >
                  <input
                    type="checkbox"
                    checked={formData.consumeCreateBase}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        consumeCreateBase: e.target.checked,
                      })
                    }
                  />
                  Create base on use
                </label>
              </div>

              <div style={{ marginBottom: '12px' }}>
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    marginBottom: '8px',
                  }}
                >
                  <label style={{ fontSize: '13px', fontWeight: 600 }}>
                    Statuses Added On Consume
                  </label>
                  <button
                    type="button"
                    onClick={addConsumeStatusToAdd}
                    className="bg-green-600 text-white px-2 py-1 rounded-md text-xs"
                  >
                    Add Status
                  </button>
                </div>
                {formData.consumeStatusesToAdd.length === 0 && (
                  <small style={{ color: '#666', fontSize: '12px' }}>
                    No statuses will be added.
                  </small>
                )}
                {formData.consumeStatusesToAdd.map((status, statusIndex) => (
                  <div
                    key={`consume-add-${statusIndex}`}
                    style={{
                      border: '1px solid #e5e7eb',
                      borderRadius: '6px',
                      padding: '10px',
                      marginBottom: '8px',
                    }}
                  >
                    <div
                      style={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(2, 1fr)',
                        gap: '8px',
                        marginBottom: '8px',
                      }}
                    >
                      <div>
                        <label
                          style={{
                            display: 'block',
                            marginBottom: '4px',
                            fontSize: '12px',
                          }}
                        >
                          Name
                        </label>
                        <input
                          type="text"
                          value={status.name}
                          onChange={(e) =>
                            updateConsumeStatusToAdd(statusIndex, {
                              name: e.target.value,
                            })
                          }
                          style={{
                            width: '100%',
                            padding: '6px',
                            border: '1px solid #ccc',
                            borderRadius: '4px',
                          }}
                        />
                      </div>
                      <div>
                        <label
                          style={{
                            display: 'block',
                            marginBottom: '4px',
                            fontSize: '12px',
                          }}
                        >
                          Duration (seconds)
                        </label>
                        <input
                          type="number"
                          min="1"
                          value={status.durationSeconds}
                          onChange={(e) =>
                            updateConsumeStatusToAdd(statusIndex, {
                              durationSeconds:
                                parseInt(e.target.value, 10) || 0,
                            })
                          }
                          style={{
                            width: '100%',
                            padding: '6px',
                            border: '1px solid #ccc',
                            borderRadius: '4px',
                          }}
                        />
                      </div>
                      <div>
                        <label
                          style={{
                            display: 'block',
                            marginBottom: '4px',
                            fontSize: '12px',
                          }}
                        >
                          Description
                        </label>
                        <input
                          type="text"
                          value={status.description}
                          onChange={(e) =>
                            updateConsumeStatusToAdd(statusIndex, {
                              description: e.target.value,
                            })
                          }
                          style={{
                            width: '100%',
                            padding: '6px',
                            border: '1px solid #ccc',
                            borderRadius: '4px',
                          }}
                        />
                      </div>
                      <div>
                        <label
                          style={{
                            display: 'block',
                            marginBottom: '4px',
                            fontSize: '12px',
                          }}
                        >
                          Effect
                        </label>
                        <input
                          type="text"
                          value={status.effect}
                          onChange={(e) =>
                            updateConsumeStatusToAdd(statusIndex, {
                              effect: e.target.value,
                            })
                          }
                          style={{
                            width: '100%',
                            padding: '6px',
                            border: '1px solid #ccc',
                            borderRadius: '4px',
                          }}
                        />
                      </div>
                    </div>

                    <div
                      style={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(3, 1fr)',
                        gap: '8px',
                        marginBottom: '8px',
                      }}
                    >
                      <div>
                        <label
                          style={{
                            display: 'block',
                            marginBottom: '4px',
                            fontSize: '12px',
                          }}
                        >
                          STR
                        </label>
                        <input
                          type="number"
                          value={status.strengthMod}
                          onChange={(e) =>
                            updateConsumeStatusToAdd(statusIndex, {
                              strengthMod: parseInt(e.target.value, 10) || 0,
                            })
                          }
                          style={{
                            width: '100%',
                            padding: '6px',
                            border: '1px solid #ccc',
                            borderRadius: '4px',
                          }}
                        />
                      </div>
                      <div>
                        <label
                          style={{
                            display: 'block',
                            marginBottom: '4px',
                            fontSize: '12px',
                          }}
                        >
                          DEX
                        </label>
                        <input
                          type="number"
                          value={status.dexterityMod}
                          onChange={(e) =>
                            updateConsumeStatusToAdd(statusIndex, {
                              dexterityMod: parseInt(e.target.value, 10) || 0,
                            })
                          }
                          style={{
                            width: '100%',
                            padding: '6px',
                            border: '1px solid #ccc',
                            borderRadius: '4px',
                          }}
                        />
                      </div>
                      <div>
                        <label
                          style={{
                            display: 'block',
                            marginBottom: '4px',
                            fontSize: '12px',
                          }}
                        >
                          CON
                        </label>
                        <input
                          type="number"
                          value={status.constitutionMod}
                          onChange={(e) =>
                            updateConsumeStatusToAdd(statusIndex, {
                              constitutionMod:
                                parseInt(e.target.value, 10) || 0,
                            })
                          }
                          style={{
                            width: '100%',
                            padding: '6px',
                            border: '1px solid #ccc',
                            borderRadius: '4px',
                          }}
                        />
                      </div>
                      <div>
                        <label
                          style={{
                            display: 'block',
                            marginBottom: '4px',
                            fontSize: '12px',
                          }}
                        >
                          INT
                        </label>
                        <input
                          type="number"
                          value={status.intelligenceMod}
                          onChange={(e) =>
                            updateConsumeStatusToAdd(statusIndex, {
                              intelligenceMod:
                                parseInt(e.target.value, 10) || 0,
                            })
                          }
                          style={{
                            width: '100%',
                            padding: '6px',
                            border: '1px solid #ccc',
                            borderRadius: '4px',
                          }}
                        />
                      </div>
                      <div>
                        <label
                          style={{
                            display: 'block',
                            marginBottom: '4px',
                            fontSize: '12px',
                          }}
                        >
                          WIS
                        </label>
                        <input
                          type="number"
                          value={status.wisdomMod}
                          onChange={(e) =>
                            updateConsumeStatusToAdd(statusIndex, {
                              wisdomMod: parseInt(e.target.value, 10) || 0,
                            })
                          }
                          style={{
                            width: '100%',
                            padding: '6px',
                            border: '1px solid #ccc',
                            borderRadius: '4px',
                          }}
                        />
                      </div>
                      <div>
                        <label
                          style={{
                            display: 'block',
                            marginBottom: '4px',
                            fontSize: '12px',
                          }}
                        >
                          CHA
                        </label>
                        <input
                          type="number"
                          value={status.charismaMod}
                          onChange={(e) =>
                            updateConsumeStatusToAdd(statusIndex, {
                              charismaMod: parseInt(e.target.value, 10) || 0,
                            })
                          }
                          style={{
                            width: '100%',
                            padding: '6px',
                            border: '1px solid #ccc',
                            borderRadius: '4px',
                          }}
                        />
                      </div>
                    </div>

                    <div
                      style={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(4, 1fr)',
                        gap: '8px',
                        marginBottom: '8px',
                      }}
                    >
                      {damageBonusFieldOptions.map(({ key, label }) => (
                        <div key={key}>
                          <label
                            style={{
                              display: 'block',
                              marginBottom: '4px',
                              fontSize: '12px',
                            }}
                          >
                            {label}
                          </label>
                          <input
                            type="number"
                            value={status[key]}
                            onChange={(e) =>
                              updateConsumeStatusToAdd(statusIndex, {
                                [key]: parseInt(e.target.value, 10) || 0,
                              })
                            }
                            style={{
                              width: '100%',
                              padding: '6px',
                              border: '1px solid #ccc',
                              borderRadius: '4px',
                            }}
                          />
                        </div>
                      ))}
                    </div>

                    <div
                      style={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(4, 1fr)',
                        gap: '8px',
                        marginBottom: '8px',
                      }}
                    >
                      {resistanceFieldOptions.map(({ key, label }) => (
                        <div key={key}>
                          <label
                            style={{
                              display: 'block',
                              marginBottom: '4px',
                              fontSize: '12px',
                            }}
                          >
                            {label}
                          </label>
                          <input
                            type="number"
                            value={status[key]}
                            onChange={(e) =>
                              updateConsumeStatusToAdd(statusIndex, {
                                [key]: parseInt(e.target.value, 10) || 0,
                              })
                            }
                            style={{
                              width: '100%',
                              padding: '6px',
                              border: '1px solid #ccc',
                              borderRadius: '4px',
                            }}
                          />
                        </div>
                      ))}
                    </div>

                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                      }}
                    >
                      <label
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '6px',
                          fontSize: '12px',
                        }}
                      >
                        <input
                          type="checkbox"
                          checked={status.positive}
                          onChange={(e) =>
                            updateConsumeStatusToAdd(statusIndex, {
                              positive: e.target.checked,
                            })
                          }
                        />
                        Positive status
                      </label>
                      <button
                        type="button"
                        className="bg-red-600 text-white px-2 py-1 rounded-md text-xs"
                        onClick={() => removeConsumeStatusToAdd(statusIndex)}
                      >
                        Remove
                      </button>
                    </div>
                  </div>
                ))}
              </div>

              <div>
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    marginBottom: '8px',
                  }}
                >
                  <label style={{ fontSize: '13px', fontWeight: 600 }}>
                    Statuses Removed On Consume
                  </label>
                  <button
                    type="button"
                    onClick={addConsumeStatusToRemove}
                    className="bg-blue-600 text-white px-2 py-1 rounded-md text-xs"
                  >
                    Add Name
                  </button>
                </div>
                {formData.consumeStatusesToRemove.length === 0 && (
                  <small style={{ color: '#666', fontSize: '12px' }}>
                    No statuses will be removed.
                  </small>
                )}
                {formData.consumeStatusesToRemove.map((name, index) => (
                  <div
                    key={`consume-remove-${index}`}
                    style={{ display: 'flex', gap: '8px', marginBottom: '6px' }}
                  >
                    <input
                      type="text"
                      value={name}
                      placeholder="Status name (e.g. Poisoned)"
                      onChange={(e) =>
                        updateConsumeStatusToRemove(index, e.target.value)
                      }
                      style={{
                        flex: 1,
                        padding: '6px',
                        border: '1px solid #ccc',
                        borderRadius: '4px',
                      }}
                    />
                    <button
                      type="button"
                      className="bg-red-600 text-white px-2 py-1 rounded-md text-xs"
                      onClick={() => removeConsumeStatusToRemove(index)}
                    >
                      Remove
                    </button>
                  </div>
                ))}
              </div>

              <div style={{ marginTop: '12px' }}>
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    marginBottom: '8px',
                  }}
                >
                  <label style={{ fontSize: '13px', fontWeight: 600 }}>
                    Spells Granted On Consume
                  </label>
                  <button
                    type="button"
                    onClick={addConsumeSpellId}
                    className="bg-indigo-600 text-white px-2 py-1 rounded-md text-xs"
                  >
                    Add Spell
                  </button>
                </div>
                {formData.consumeSpellIds.length === 0 && (
                  <small style={{ color: '#666', fontSize: '12px' }}>
                    No spells will be granted.
                  </small>
                )}
                {formData.consumeSpellIds.map((spellID, index) => (
                  <div
                    key={`consume-spell-${index}`}
                    style={{ display: 'flex', gap: '8px', marginBottom: '6px' }}
                  >
                    <select
                      value={spellID}
                      onChange={(e) =>
                        updateConsumeSpellId(index, e.target.value)
                      }
                      style={{
                        flex: 1,
                        padding: '6px',
                        border: '1px solid #ccc',
                        borderRadius: '4px',
                      }}
                    >
                      <option value="">Select spell</option>
                      {spells.map((spell) => (
                        <option key={spell.id} value={spell.id}>
                          {spell.name} ({spell.schoolOfMagic})
                        </option>
                      ))}
                    </select>
                    <button
                      type="button"
                      className="bg-red-600 text-white px-2 py-1 rounded-md text-xs"
                      onClick={() => removeConsumeSpellId(index)}
                    >
                      Remove
                    </button>
                  </div>
                ))}
              </div>

              <div style={{ marginTop: '12px' }}>
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    marginBottom: '8px',
                  }}
                >
                  <label style={{ fontSize: '13px', fontWeight: 600 }}>
                    Recipes Learned On Consume
                  </label>
                  <button
                    type="button"
                    onClick={addTeachRecipeId}
                    className="bg-amber-600 text-white px-2 py-1 rounded-md text-xs"
                  >
                    Add Recipe
                  </button>
                </div>
                {formData.consumeTeachRecipeIds.length === 0 && (
                  <small style={{ color: '#666', fontSize: '12px' }}>
                    No recipes will be taught.
                  </small>
                )}
                {formData.consumeTeachRecipeIds.map((recipeID, index) => (
                  <div
                    key={`consume-teach-recipe-${index}`}
                    style={{ display: 'flex', gap: '8px', marginBottom: '6px' }}
                  >
                    <select
                      value={recipeID}
                      onChange={(e) =>
                        updateTeachRecipeId(index, e.target.value)
                      }
                      style={{
                        flex: 1,
                        padding: '6px',
                        border: '1px solid #ccc',
                        borderRadius: '4px',
                      }}
                    >
                      <option value="">Select recipe</option>
                      {recipeOptions.map((option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </select>
                    <button
                      type="button"
                      className="bg-red-600 text-white px-2 py-1 rounded-md text-xs"
                      onClick={() => removeTeachRecipeId(index)}
                    >
                      Remove
                    </button>
                  </div>
                ))}
              </div>
            </div>

            <div
              style={{
                marginBottom: '15px',
                padding: '12px',
                border: '1px solid #e5e7eb',
                borderRadius: '6px',
              }}
            >
              <label
                style={{
                  display: 'block',
                  marginBottom: '8px',
                  fontWeight: 600,
                }}
              >
                Crafting Recipes
              </label>
              <small
                style={{
                  color: '#666',
                  fontSize: '12px',
                  display: 'block',
                  marginBottom: '10px',
                }}
              >
                Public recipes unlock automatically at the required room tier.
                Private recipes must be learned from an item first.
              </small>

              {renderRecipeEditor(
                'Alchemy Recipes',
                'alchemyRecipes',
                'bg-emerald-600'
              )}
              {renderRecipeEditor(
                'Workshop Recipes',
                'workshopRecipes',
                'bg-orange-600'
              )}
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '8px' }}>
                Stat Modifiers (while equipped):
              </label>
              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(3, 1fr)',
                  gap: '10px',
                }}
              >
                <div>
                  <label
                    style={{
                      display: 'block',
                      marginBottom: '4px',
                      fontSize: '12px',
                    }}
                  >
                    Strength
                  </label>
                  <input
                    type="number"
                    min="0"
                    value={formData.strengthMod}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        strengthMod: parseInt(e.target.value, 10) || 0,
                      })
                    }
                    style={{
                      width: '100%',
                      padding: '6px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                    }}
                  />
                </div>
                <div>
                  <label
                    style={{
                      display: 'block',
                      marginBottom: '4px',
                      fontSize: '12px',
                    }}
                  >
                    Dexterity
                  </label>
                  <input
                    type="number"
                    min="0"
                    value={formData.dexterityMod}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        dexterityMod: parseInt(e.target.value, 10) || 0,
                      })
                    }
                    style={{
                      width: '100%',
                      padding: '6px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                    }}
                  />
                </div>
                <div>
                  <label
                    style={{
                      display: 'block',
                      marginBottom: '4px',
                      fontSize: '12px',
                    }}
                  >
                    Constitution
                  </label>
                  <input
                    type="number"
                    min="0"
                    value={formData.constitutionMod}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        constitutionMod: parseInt(e.target.value, 10) || 0,
                      })
                    }
                    style={{
                      width: '100%',
                      padding: '6px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                    }}
                  />
                </div>
                <div>
                  <label
                    style={{
                      display: 'block',
                      marginBottom: '4px',
                      fontSize: '12px',
                    }}
                  >
                    Intelligence
                  </label>
                  <input
                    type="number"
                    min="0"
                    value={formData.intelligenceMod}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        intelligenceMod: parseInt(e.target.value, 10) || 0,
                      })
                    }
                    style={{
                      width: '100%',
                      padding: '6px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                    }}
                  />
                </div>
                <div>
                  <label
                    style={{
                      display: 'block',
                      marginBottom: '4px',
                      fontSize: '12px',
                    }}
                  >
                    Wisdom
                  </label>
                  <input
                    type="number"
                    min="0"
                    value={formData.wisdomMod}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        wisdomMod: parseInt(e.target.value, 10) || 0,
                      })
                    }
                    style={{
                      width: '100%',
                      padding: '6px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                    }}
                  />
                </div>
                <div>
                  <label
                    style={{
                      display: 'block',
                      marginBottom: '4px',
                      fontSize: '12px',
                    }}
                  >
                    Charisma
                  </label>
                  <input
                    type="number"
                    min="0"
                    value={formData.charismaMod}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        charismaMod: parseInt(e.target.value, 10) || 0,
                      })
                    }
                    style={{
                      width: '100%',
                      padding: '6px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                    }}
                  />
                </div>
              </div>
            </div>

            <div style={{ marginTop: '20px', display: 'flex', gap: '10px' }}>
              <button
                onClick={() => {
                  if (editingItem) {
                    handleUpdateItem();
                  } else {
                    handleCreateItem();
                  }
                }}
                className="bg-blue-500 text-white px-4 py-2 rounded-md"
              >
                {editingItem ? 'Update' : 'Create'}
              </button>
              <button
                onClick={() => {
                  setShowCreateItem(false);
                  setEditingItem(null);
                  resetForm();
                }}
                className="bg-gray-500 text-white px-4 py-2 rounded-md"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {useOutfitItem && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            backgroundColor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            zIndex: 1000,
          }}
        >
          <div
            style={{
              backgroundColor: '#fff',
              padding: '30px',
              borderRadius: '8px',
              width: '520px',
              maxHeight: '80vh',
              overflow: 'auto',
            }}
          >
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold">Use Outfit</h2>
              <button
                onClick={() => setUseOutfitItem(null)}
                className="text-gray-500 hover:text-gray-700"
              >
                ✕
              </button>
            </div>

            <div className="mb-4 text-sm text-gray-600">
              Selected item:{' '}
              <span className="font-medium text-gray-900">
                {useOutfitItem.name}
              </span>
            </div>

            <div className="mb-4">
              <SearchableSelect
                label="User"
                placeholder="Search by username or name…"
                options={userOptions}
                value={useOutfitUser}
                onChange={setUseOutfitUser}
              />
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700">
                Selfie URL
              </label>
              <input
                type="text"
                value={useOutfitSelfieUrl}
                onChange={(e) => setUseOutfitSelfieUrl(e.target.value)}
                placeholder="https://..."
                className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
            </div>

            {useOutfitStatus && (
              <div
                className={`mb-4 rounded-md border px-3 py-2 text-sm ${
                  useOutfitStatusKind === 'error'
                    ? 'border-rose-200 bg-rose-50 text-rose-800'
                    : 'border-emerald-200 bg-emerald-50 text-emerald-800'
                }`}
              >
                {useOutfitStatus}
              </div>
            )}

            <div className="flex gap-2">
              <button
                onClick={handleUseOutfit}
                disabled={
                  !useOutfitUser || !useOutfitSelfieUrl || useOutfitSubmitting
                }
                className="bg-indigo-600 text-white px-4 py-2 rounded-md disabled:opacity-60"
              >
                {useOutfitSubmitting ? 'Starting…' : 'Start Generation'}
              </button>
              <button
                onClick={() => setUseOutfitItem(null)}
                className="bg-gray-100 text-gray-700 px-4 py-2 rounded-md"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Generate Item Modal */}
      {showGenerateItem && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            backgroundColor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            zIndex: 1000,
          }}
        >
          <div
            style={{
              backgroundColor: '#fff',
              padding: '30px',
              borderRadius: '8px',
              width: '500px',
              maxHeight: '80vh',
              overflow: 'auto',
            }}
          >
            <h2>Generate Inventory Item</h2>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Name *:
              </label>
              <input
                type="text"
                value={generationData.name}
                onChange={(e) =>
                  setGenerationData({ ...generationData, name: e.target.value })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
                required
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Description:
              </label>
              <textarea
                value={generationData.description}
                onChange={(e) =>
                  setGenerationData({
                    ...generationData,
                    description: e.target.value,
                  })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                  minHeight: '80px',
                }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Rarity Tier *:
              </label>
              <select
                value={generationData.rarityTier}
                onChange={(e) =>
                  setGenerationData({
                    ...generationData,
                    rarityTier: e.target.value,
                  })
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
                required
              >
                <option value={Rarity.Common}>Common</option>
                <option value={Rarity.Uncommon}>Uncommon</option>
                <option value={Rarity.Epic}>Epic</option>
                <option value={Rarity.Mythic}>Mythic</option>
                <option value="Not Droppable">Not Droppable</option>
              </select>
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>
                Equip Slot:
              </label>
              <select
                value={generationData.equipSlot}
                onChange={(e) =>
                  handleGenerationEquipSlotChange(e.target.value)
                }
                style={{
                  width: '100%',
                  padding: '8px',
                  border: '1px solid #ccc',
                  borderRadius: '4px',
                }}
              >
                {equipSlotOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>

            {isHandEquipSlot(generationData.equipSlot) && (
              <div
                style={{
                  marginBottom: '15px',
                  padding: '12px',
                  border: '1px solid #e5e7eb',
                  borderRadius: '6px',
                }}
              >
                <label
                  style={{
                    display: 'block',
                    marginBottom: '8px',
                    fontWeight: 600,
                  }}
                >
                  Generated Hand Equipment
                </label>
                <div style={{ marginBottom: '10px' }}>
                  <label
                    style={{
                      display: 'block',
                      marginBottom: '4px',
                      fontSize: '12px',
                    }}
                  >
                    Hand Item Type *
                  </label>
                  <select
                    value={generationData.handItemCategory}
                    onChange={(e) =>
                      handleGenerationHandCategoryChange(e.target.value)
                    }
                    style={{
                      width: '100%',
                      padding: '8px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                    }}
                  >
                    <option value="">Select hand item type</option>
                    {(
                      handItemCategoryOptions[generationData.equipSlot] || []
                    ).map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label
                    style={{
                      display: 'block',
                      marginBottom: '4px',
                      fontSize: '12px',
                    }}
                  >
                    Handedness *
                  </label>
                  <select
                    value={generationData.handedness}
                    onChange={(e) =>
                      setGenerationData({
                        ...generationData,
                        handedness: e.target.value,
                      })
                    }
                    disabled={
                      generationData.equipSlot === 'off_hand' ||
                      generationData.handItemCategory === 'staff'
                    }
                    style={{
                      width: '100%',
                      padding: '8px',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                    }}
                  >
                    <option value="">Select handedness</option>
                    {handednessOptions.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
            )}

            <div style={{ marginTop: '20px', display: 'flex', gap: '10px' }}>
              <button
                onClick={handleGenerateItem}
                className="bg-green-600 text-white px-4 py-2 rounded-md"
              >
                Generate
              </button>
              <button
                onClick={() => {
                  setShowGenerateItem(false);
                  resetGenerationForm();
                }}
                className="bg-gray-500 text-white px-4 py-2 rounded-md"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {showDeleteConfirm && itemToDelete && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            backgroundColor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            zIndex: 1000,
          }}
        >
          <div
            style={{
              backgroundColor: '#fff',
              padding: '30px',
              borderRadius: '8px',
              width: '400px',
            }}
          >
            <h2>Confirm Delete</h2>
            <p>
              Are you sure you want to delete "{itemToDelete.name}"? This action
              cannot be undone.
            </p>
            <div style={{ marginTop: '20px', display: 'flex', gap: '10px' }}>
              <button
                onClick={confirmDelete}
                className="bg-red-500 text-white px-4 py-2 rounded-md"
              >
                Delete
              </button>
              <button
                onClick={() => {
                  setShowDeleteConfirm(false);
                  setItemToDelete(null);
                }}
                className="bg-gray-500 text-white px-4 py-2 rounded-md"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Bulk Delete Confirmation Modal */}
      {showBulkDeleteConfirm && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            backgroundColor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            zIndex: 1000,
          }}
        >
          <div
            style={{
              backgroundColor: '#fff',
              padding: '30px',
              borderRadius: '8px',
              width: '420px',
            }}
          >
            <h2>Confirm Bulk Delete</h2>
            <p>
              Delete {selectedItemIDs.size} selected inventory item(s)? This
              action cannot be undone.
            </p>
            <div style={{ marginTop: '20px', display: 'flex', gap: '10px' }}>
              <button
                onClick={confirmBulkDelete}
                className="bg-red-600 text-white px-4 py-2 rounded-md disabled:bg-gray-300 disabled:cursor-not-allowed"
                disabled={selectedItemIDs.size === 0 || bulkDeleteBusy}
              >
                {bulkDeleteBusy ? 'Deleting...' : 'Delete Selected'}
              </button>
              <button
                onClick={() => setShowBulkDeleteConfirm(false)}
                className="bg-gray-500 text-white px-4 py-2 rounded-md"
                disabled={bulkDeleteBusy}
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
