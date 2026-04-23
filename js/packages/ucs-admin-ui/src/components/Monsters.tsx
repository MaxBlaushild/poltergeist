import React, {
  useCallback,
  useDeferredValue,
  useEffect,
  useMemo,
  useState,
} from 'react';
import { useAPI, useZoneContext } from '@poltergeist/contexts';
import { Spell } from '@poltergeist/types';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import {
  MaterialRewardsEditor,
  MaterialRewardForm,
  summarizeMaterialRewards,
} from './MaterialRewardsEditor.tsx';
import ContentDashboard from './ContentDashboard.tsx';
import {
  countBy,
  useAdminAggregateDataset,
} from './contentDashboardUtils.ts';
import type { ZoneGenre } from '@poltergeist/types';
import {
  useZoneKinds,
  zoneKindDescription,
  zoneKindLabel,
  zoneKindSelectPlaceholderLabel,
  zoneKindSummaryLabel,
} from './zoneKindHelpers.ts';

type MonsterEncounterType = 'monster' | 'boss' | 'raid';
type MonsterTemplateType = 'monster' | 'boss' | 'raid';

type MonsterTemplateRecord = {
  id: string;
  createdAt: string;
  updatedAt: string;
  archived?: boolean;
  monsterType: MonsterTemplateType;
  zoneKind?: string;
  genreId: string;
  genre?: ZoneGenre;
  name: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  baseStrength: number;
  baseDexterity: number;
  baseConstitution: number;
  baseIntelligence: number;
  baseWisdom: number;
  baseCharisma: number;
  affinityDamageBonuses?: Record<string, number>;
  affinityResistances?: Record<string, number>;
  progressions?: MonsterTemplateProgressionRecord[];
  spells: Spell[];
  imageGenerationStatus?: string;
  imageGenerationError?: string | null;
};

const formatTemplateSpellLoadoutLabel = (
  template: MonsterTemplateRecord
): string => {
  const spellCount = template.spells?.length ?? 0;
  if (spellCount <= 0) {
    return '0 spells';
  }
  if (spellCount === 1) {
    return '1 spell';
  }
  if (spellCount === 2) {
    return '2 spells';
  }
  if (spellCount <= 4) {
    return '3-4 spells';
  }
  return '5+ spells';
};

const formatTemplateProgressionCoverageLabel = (
  template: MonsterTemplateRecord
): string => {
  const progressionCount = template.progressions?.length ?? 0;
  if (progressionCount <= 0) {
    return 'No progressions';
  }
  if (progressionCount === 1) {
    return '1 progression';
  }
  return '2+ progressions';
};

type MonsterTemplateProgressionRecord = {
  id: string;
  name: string;
  abilityType?: 'spell' | 'technique' | string;
};

type MonsterAbilityProgressionOption = {
  id: string;
  name: string;
  abilityType: 'spell' | 'technique';
  memberCount: number;
};

type MonsterRewardItem = {
  inventoryItemId: number;
  quantity: number;
  inventoryItem?: {
    id: number;
    name: string;
    imageUrl?: string;
  };
};

type MonsterRecord = {
  id: string;
  createdAt: string;
  updatedAt: string;
  name: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  zoneId: string;
  zoneKind?: string;
  genreId: string;
  genre?: ZoneGenre;
  latitude: number;
  longitude: number;
  templateId?: string;
  template?: MonsterTemplateRecord;
  dominantHandInventoryItemId?: number;
  dominantHandInventoryItem?: InventoryItemLite;
  offHandInventoryItemId?: number;
  offHandInventoryItem?: InventoryItemLite;
  weaponInventoryItemId?: number;
  weaponInventoryItem?: InventoryItemLite;
  level: number;
  strength: number;
  dexterity: number;
  constitution: number;
  intelligence: number;
  wisdom: number;
  charisma: number;
  health: number;
  maxHealth: number;
  mana: number;
  maxMana: number;
  attackDamageMin: number;
  attackDamageMax: number;
  attackSwipesPerAttack: number;
  affinityDamageBonuses?: Record<string, number>;
  affinityResistances?: Record<string, number>;
  spells: Spell[];
  rewardMode?: 'explicit' | 'random';
  randomRewardSize?: 'small' | 'medium' | 'large';
  rewardExperience: number;
  rewardGold: number;
  materialRewards?: MaterialRewardForm[];
  imageGenerationStatus?: string;
  imageGenerationError?: string | null;
  itemRewards: MonsterRewardItem[];
};

type MonsterEncounterMemberRecord = {
  slot: number;
  monster: MonsterRecord;
};

type MonsterEncounterRecord = {
  id: string;
  createdAt: string;
  updatedAt: string;
  name: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  encounterType: MonsterEncounterType;
  rewardMode?: 'explicit' | 'random';
  randomRewardSize?: 'small' | 'medium' | 'large';
  rewardExperience: number;
  rewardGold: number;
  materialRewards?: MaterialRewardForm[];
  itemRewards: MonsterRewardItem[];
  scaleWithUserLevel: boolean;
  recurrenceFrequency?: string | null;
  nextRecurrenceAt?: string | null;
  zoneId: string;
  zoneKind?: string;
  latitude: number;
  longitude: number;
  monsterCount: number;
  members: MonsterEncounterMemberRecord[];
  monsters: MonsterRecord[];
};

type InventoryItemLite = {
  id: number;
  name: string;
  imageUrl?: string;
  equipSlot?: string | null;
  handItemCategory?: string | null;
  handedness?: string | null;
  damageMin?: number | null;
  damageMax?: number | null;
  swipesPerAttack?: number | null;
  blockPercentage?: number | null;
  damageBlocked?: number | null;
  spellDamageBonusPercent?: number | null;
};

type ImagePreviewState = {
  url: string;
  alt: string;
};

type StaticThumbnailResponse = {
  thumbnailUrl?: string;
  status?: string;
  exists?: boolean;
  requestedAt?: string;
  lastModified?: string;
  prompt?: string;
};

type BulkMonsterTemplateStatus = {
  jobId: string;
  status: string;
  source?: string;
  monsterType?: MonsterTemplateType;
  genreId?: string;
  zoneKind?: string;
  yeetIt?: boolean;
  totalCount: number;
  createdCount: number;
  error?: string;
  queuedAt?: string;
  startedAt?: string;
  completedAt?: string;
  updatedAt?: string;
};

type MonsterTemplateSuggestionPayload = {
  monsterType: MonsterTemplateType;
  genreId: string;
  zoneKind?: string;
  name: string;
  description: string;
  baseStrength: number;
  baseDexterity: number;
  baseConstitution: number;
  baseIntelligence: number;
  baseWisdom: number;
  baseCharisma: number;
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

type MonsterTemplateSuggestionDraft = {
  id: string;
  createdAt: string;
  updatedAt: string;
  jobId: string;
  status: string;
  monsterType: MonsterTemplateType;
  genreId: string;
  genre?: ZoneGenre;
  zoneKind?: string;
  name: string;
  description: string;
  payload: MonsterTemplateSuggestionPayload;
  monsterTemplateId?: string;
  monsterTemplate?: MonsterTemplateRecord;
  convertedAt?: string;
};

type MonsterTemplateAffinityRefreshStatus = {
  jobId: string;
  status: string;
  totalCount: number;
  updatedCount: number;
  templateIds?: string[];
  error?: string;
  queuedAt?: string;
  startedAt?: string;
  completedAt?: string;
  updatedAt?: string;
};

type MonsterTemplateProgressionResetStatus = {
  jobId: string;
  status: string;
  totalCount: number;
  updatedCount: number;
  templateIds?: string[];
  error?: string;
  queuedAt?: string;
  startedAt?: string;
  completedAt?: string;
  updatedAt?: string;
};

type PaginatedResponse<T> = {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  activeCount?: number;
  archivedCount?: number;
};

const defaultMonsterUndiscoveredIconPrompt =
  'A retro 16-bit RPG map marker icon for an undiscovered monster. Hidden beast silhouette and warning rune motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette.';
const defaultBossUndiscoveredIconPrompt =
  'A retro 16-bit RPG map marker icon for an undiscovered boss encounter. Hidden crown-horned beast silhouette with elite warning sigil motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette.';
const defaultRaidUndiscoveredIconPrompt =
  'A retro 16-bit RPG map marker icon for an undiscovered raid encounter. Hidden multi-creature threat silhouette with party danger rune motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette.';

type EncounterIconState = {
  url: string;
  status: string;
  exists: boolean;
  requestedAt: string | null;
  lastModified: string | null;
  previewNonce: number;
  prompt: string;
  message: string | null;
  error: string | null;
  busy: boolean;
  statusLoading: boolean;
};

type MonsterTemplateFormState = {
  monsterType: MonsterTemplateType;
  zoneKind: string;
  genreId: string;
  name: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  baseStrength: string;
  baseDexterity: string;
  baseConstitution: string;
  baseIntelligence: string;
  baseWisdom: string;
  baseCharisma: string;
  physicalDamageBonusPercent: string;
  piercingDamageBonusPercent: string;
  slashingDamageBonusPercent: string;
  bludgeoningDamageBonusPercent: string;
  fireDamageBonusPercent: string;
  iceDamageBonusPercent: string;
  lightningDamageBonusPercent: string;
  poisonDamageBonusPercent: string;
  arcaneDamageBonusPercent: string;
  holyDamageBonusPercent: string;
  shadowDamageBonusPercent: string;
  physicalResistancePercent: string;
  piercingResistancePercent: string;
  slashingResistancePercent: string;
  bludgeoningResistancePercent: string;
  fireResistancePercent: string;
  iceResistancePercent: string;
  lightningResistancePercent: string;
  poisonResistancePercent: string;
  arcaneResistancePercent: string;
  holyResistancePercent: string;
  shadowResistancePercent: string;
  spellProgressionIds: string[];
  techniqueProgressionIds: string[];
};

type MonsterFormItem = {
  inventoryItemId: string;
  quantity: string;
};

type MonsterFormState = {
  name: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  zoneId: string;
  zoneKind: string;
  latitude: string;
  longitude: string;
  templateId: string;
  dominantHandInventoryItemId: string;
  offHandInventoryItemId: string;
  level: string;
  rewardMode: 'explicit' | 'random';
  randomRewardSize: 'small' | 'medium' | 'large';
  rewardExperience: string;
  rewardGold: string;
  materialRewards: MaterialRewardForm[];
  itemRewards: MonsterFormItem[];
};

type MonsterEncounterFormState = {
  name: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  encounterType: MonsterEncounterType;
  rewardMode: 'explicit' | 'random';
  randomRewardSize: 'small' | 'medium' | 'large';
  rewardExperience: string;
  rewardGold: string;
  materialRewards: MaterialRewardForm[];
  itemRewards: MonsterFormItem[];
  scaleWithUserLevel: boolean;
  recurrenceFrequency: string;
  zoneId: string;
  zoneKind: string;
  latitude: string;
  longitude: string;
  monsterIds: string[];
};

const recurrenceOptions = [
  { value: '', label: 'No Recurrence' },
  { value: 'daily', label: 'Daily' },
  { value: 'weekly', label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
];

const monsterEncounterTypeOptions: Array<{
  value: MonsterEncounterType;
  label: string;
}> = [
  { value: 'monster', label: 'Monster Encounter' },
  { value: 'boss', label: 'Boss Encounter' },
  { value: 'raid', label: 'Raid Encounter' },
];

const monsterTemplateTypeOptions: Array<{
  value: MonsterTemplateType;
  label: string;
}> = [
  { value: 'monster', label: 'Standard Monster' },
  { value: 'boss', label: 'Boss Monster' },
  { value: 'raid', label: 'Raid Monster' },
];

const encounterIconDefaults: Record<
  MonsterEncounterType,
  { label: string; url: string; prompt: string }
> = {
  monster: {
    label: 'Monster',
    url: 'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/monster-undiscovered.png',
    prompt: defaultMonsterUndiscoveredIconPrompt,
  },
  boss: {
    label: 'Boss',
    url: 'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/boss-undiscovered.png',
    prompt: defaultBossUndiscoveredIconPrompt,
  },
  raid: {
    label: 'Raid',
    url: 'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/raid-undiscovered.png',
    prompt: defaultRaidUndiscoveredIconPrompt,
  },
};

const damageBonusFieldOptions = [
  {
    key: 'physicalDamageBonusPercent',
    label: 'Physical Dmg %',
    affinity: 'physical',
  },
  {
    key: 'piercingDamageBonusPercent',
    label: 'Piercing Dmg %',
    affinity: 'piercing',
  },
  {
    key: 'slashingDamageBonusPercent',
    label: 'Slashing Dmg %',
    affinity: 'slashing',
  },
  {
    key: 'bludgeoningDamageBonusPercent',
    label: 'Bludgeoning Dmg %',
    affinity: 'bludgeoning',
  },
  { key: 'fireDamageBonusPercent', label: 'Fire Dmg %', affinity: 'fire' },
  { key: 'iceDamageBonusPercent', label: 'Ice Dmg %', affinity: 'ice' },
  {
    key: 'lightningDamageBonusPercent',
    label: 'Lightning Dmg %',
    affinity: 'lightning',
  },
  {
    key: 'poisonDamageBonusPercent',
    label: 'Poison Dmg %',
    affinity: 'poison',
  },
  {
    key: 'arcaneDamageBonusPercent',
    label: 'Arcane Dmg %',
    affinity: 'arcane',
  },
  { key: 'holyDamageBonusPercent', label: 'Holy Dmg %', affinity: 'holy' },
  {
    key: 'shadowDamageBonusPercent',
    label: 'Shadow Dmg %',
    affinity: 'shadow',
  },
] as const;

const resistanceFieldOptions = [
  {
    key: 'physicalResistancePercent',
    label: 'Physical Res %',
    affinity: 'physical',
  },
  {
    key: 'piercingResistancePercent',
    label: 'Piercing Res %',
    affinity: 'piercing',
  },
  {
    key: 'slashingResistancePercent',
    label: 'Slashing Res %',
    affinity: 'slashing',
  },
  {
    key: 'bludgeoningResistancePercent',
    label: 'Bludgeoning Res %',
    affinity: 'bludgeoning',
  },
  { key: 'fireResistancePercent', label: 'Fire Res %', affinity: 'fire' },
  { key: 'iceResistancePercent', label: 'Ice Res %', affinity: 'ice' },
  {
    key: 'lightningResistancePercent',
    label: 'Lightning Res %',
    affinity: 'lightning',
  },
  { key: 'poisonResistancePercent', label: 'Poison Res %', affinity: 'poison' },
  { key: 'arcaneResistancePercent', label: 'Arcane Res %', affinity: 'arcane' },
  { key: 'holyResistancePercent', label: 'Holy Res %', affinity: 'holy' },
  { key: 'shadowResistancePercent', label: 'Shadow Res %', affinity: 'shadow' },
] as const;

const parseIntSafe = (value: string, fallback = 0): number => {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parseFloatSafe = (value: string, fallback = 0): number => {
  const parsed = Number.parseFloat(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parseOptionalInt = (value: string): number | undefined => {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  const parsed = Number.parseInt(trimmed, 10);
  return Number.isFinite(parsed) ? parsed : undefined;
};

const defaultGenreIdFromList = (genres: ZoneGenre[]): string => {
  const fantasy = genres.find(
    (genre) => (genre.name || '').trim().toLowerCase() === 'fantasy'
  );
  return fantasy?.id ?? genres[0]?.id ?? '';
};

const formatGenreLabel = (genre?: ZoneGenre | null): string =>
  genre?.name?.trim() || 'Fantasy';

const monsterListPageSize = 25;
const monsterTemplateSuggestionDraftsPerPage = 12;
const monsterTemplateSuggestionJobsPerPage = 10;

type PaginationControlsProps = {
  page: number;
  pageSize: number;
  total: number;
  label: string;
  onPageChange: (page: number) => void;
};

const PaginationControls = ({
  page,
  pageSize,
  total,
  label,
  onPageChange,
}: PaginationControlsProps) => {
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const start = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const end = total === 0 ? 0 : Math.min(total, page * pageSize);

  return (
    <div className="mt-4 flex flex-wrap items-center justify-between gap-3 border-t border-gray-200 pt-3">
      <p className="text-sm text-gray-600">
        {total === 0
          ? `No ${label}.`
          : `Showing ${start}-${end} of ${total} ${label}`}
      </p>
      <div className="flex items-center gap-2">
        <button
          type="button"
          className="rounded-md border border-gray-300 px-3 py-1.5 text-sm text-gray-700 disabled:cursor-not-allowed disabled:opacity-50"
          onClick={() => onPageChange(page - 1)}
          disabled={page <= 1}
        >
          Previous
        </button>
        <span className="text-sm text-gray-600">
          Page {page} of {totalPages}
        </span>
        <button
          type="button"
          className="rounded-md border border-gray-300 px-3 py-1.5 text-sm text-gray-700 disabled:cursor-not-allowed disabled:opacity-50"
          onClick={() => onPageChange(page + 1)}
          disabled={page >= totalPages}
        >
          Next
        </button>
      </div>
    </div>
  );
};

const emptyTemplateForm = (genreId = ''): MonsterTemplateFormState => ({
  monsterType: 'monster',
  zoneKind: '',
  genreId,
  name: '',
  description: '',
  imageUrl: '',
  thumbnailUrl: '',
  baseStrength: '10',
  baseDexterity: '10',
  baseConstitution: '10',
  baseIntelligence: '10',
  baseWisdom: '10',
  baseCharisma: '10',
  physicalDamageBonusPercent: '0',
  piercingDamageBonusPercent: '0',
  slashingDamageBonusPercent: '0',
  bludgeoningDamageBonusPercent: '0',
  fireDamageBonusPercent: '0',
  iceDamageBonusPercent: '0',
  lightningDamageBonusPercent: '0',
  poisonDamageBonusPercent: '0',
  arcaneDamageBonusPercent: '0',
  holyDamageBonusPercent: '0',
  shadowDamageBonusPercent: '0',
  physicalResistancePercent: '0',
  piercingResistancePercent: '0',
  slashingResistancePercent: '0',
  bludgeoningResistancePercent: '0',
  fireResistancePercent: '0',
  iceResistancePercent: '0',
  lightningResistancePercent: '0',
  poisonResistancePercent: '0',
  arcaneResistancePercent: '0',
  holyResistancePercent: '0',
  shadowResistancePercent: '0',
  spellProgressionIds: [],
  techniqueProgressionIds: [],
});

const templateFormFromRecord = (
  template: MonsterTemplateRecord
): MonsterTemplateFormState => ({
  monsterType: template.monsterType ?? 'monster',
  zoneKind: template.zoneKind ?? '',
  genreId: template.genreId ?? template.genre?.id ?? '',
  name: template.name ?? '',
  description: template.description ?? '',
  imageUrl: template.imageUrl ?? '',
  thumbnailUrl: template.thumbnailUrl ?? '',
  baseStrength: String(template.baseStrength ?? 10),
  baseDexterity: String(template.baseDexterity ?? 10),
  baseConstitution: String(template.baseConstitution ?? 10),
  baseIntelligence: String(template.baseIntelligence ?? 10),
  baseWisdom: String(template.baseWisdom ?? 10),
  baseCharisma: String(template.baseCharisma ?? 10),
  physicalDamageBonusPercent: String(
    template.affinityDamageBonuses?.physical ?? 0
  ),
  piercingDamageBonusPercent: String(
    template.affinityDamageBonuses?.piercing ?? 0
  ),
  slashingDamageBonusPercent: String(
    template.affinityDamageBonuses?.slashing ?? 0
  ),
  bludgeoningDamageBonusPercent: String(
    template.affinityDamageBonuses?.bludgeoning ?? 0
  ),
  fireDamageBonusPercent: String(template.affinityDamageBonuses?.fire ?? 0),
  iceDamageBonusPercent: String(template.affinityDamageBonuses?.ice ?? 0),
  lightningDamageBonusPercent: String(
    template.affinityDamageBonuses?.lightning ?? 0
  ),
  poisonDamageBonusPercent: String(template.affinityDamageBonuses?.poison ?? 0),
  arcaneDamageBonusPercent: String(template.affinityDamageBonuses?.arcane ?? 0),
  holyDamageBonusPercent: String(template.affinityDamageBonuses?.holy ?? 0),
  shadowDamageBonusPercent: String(template.affinityDamageBonuses?.shadow ?? 0),
  physicalResistancePercent: String(
    template.affinityResistances?.physical ?? 0
  ),
  piercingResistancePercent: String(
    template.affinityResistances?.piercing ?? 0
  ),
  slashingResistancePercent: String(
    template.affinityResistances?.slashing ?? 0
  ),
  bludgeoningResistancePercent: String(
    template.affinityResistances?.bludgeoning ?? 0
  ),
  fireResistancePercent: String(template.affinityResistances?.fire ?? 0),
  iceResistancePercent: String(template.affinityResistances?.ice ?? 0),
  lightningResistancePercent: String(
    template.affinityResistances?.lightning ?? 0
  ),
  poisonResistancePercent: String(template.affinityResistances?.poison ?? 0),
  arcaneResistancePercent: String(template.affinityResistances?.arcane ?? 0),
  holyResistancePercent: String(template.affinityResistances?.holy ?? 0),
  shadowResistancePercent: String(template.affinityResistances?.shadow ?? 0),
  spellProgressionIds: (template.progressions ?? [])
    .filter(
      (progression) => (progression.abilityType ?? 'spell') !== 'technique'
    )
    .map((progression) => progression.id),
  techniqueProgressionIds: (template.progressions ?? [])
    .filter(
      (progression) => (progression.abilityType ?? 'spell') === 'technique'
    )
    .map((progression) => progression.id),
});

const templatePayloadFromForm = (form: MonsterTemplateFormState) => ({
  monsterType: form.monsterType,
  zoneKind: form.zoneKind,
  genreId: form.genreId.trim(),
  name: form.name.trim(),
  description: form.description.trim(),
  imageUrl: form.imageUrl.trim(),
  thumbnailUrl: form.thumbnailUrl.trim(),
  baseStrength: parseIntSafe(form.baseStrength, 10),
  baseDexterity: parseIntSafe(form.baseDexterity, 10),
  baseConstitution: parseIntSafe(form.baseConstitution, 10),
  baseIntelligence: parseIntSafe(form.baseIntelligence, 10),
  baseWisdom: parseIntSafe(form.baseWisdom, 10),
  baseCharisma: parseIntSafe(form.baseCharisma, 10),
  physicalDamageBonusPercent: parseIntSafe(form.physicalDamageBonusPercent, 0),
  piercingDamageBonusPercent: parseIntSafe(form.piercingDamageBonusPercent, 0),
  slashingDamageBonusPercent: parseIntSafe(form.slashingDamageBonusPercent, 0),
  bludgeoningDamageBonusPercent: parseIntSafe(
    form.bludgeoningDamageBonusPercent,
    0
  ),
  fireDamageBonusPercent: parseIntSafe(form.fireDamageBonusPercent, 0),
  iceDamageBonusPercent: parseIntSafe(form.iceDamageBonusPercent, 0),
  lightningDamageBonusPercent: parseIntSafe(
    form.lightningDamageBonusPercent,
    0
  ),
  poisonDamageBonusPercent: parseIntSafe(form.poisonDamageBonusPercent, 0),
  arcaneDamageBonusPercent: parseIntSafe(form.arcaneDamageBonusPercent, 0),
  holyDamageBonusPercent: parseIntSafe(form.holyDamageBonusPercent, 0),
  shadowDamageBonusPercent: parseIntSafe(form.shadowDamageBonusPercent, 0),
  physicalResistancePercent: parseIntSafe(form.physicalResistancePercent, 0),
  piercingResistancePercent: parseIntSafe(form.piercingResistancePercent, 0),
  slashingResistancePercent: parseIntSafe(form.slashingResistancePercent, 0),
  bludgeoningResistancePercent: parseIntSafe(
    form.bludgeoningResistancePercent,
    0
  ),
  fireResistancePercent: parseIntSafe(form.fireResistancePercent, 0),
  iceResistancePercent: parseIntSafe(form.iceResistancePercent, 0),
  lightningResistancePercent: parseIntSafe(form.lightningResistancePercent, 0),
  poisonResistancePercent: parseIntSafe(form.poisonResistancePercent, 0),
  arcaneResistancePercent: parseIntSafe(form.arcaneResistancePercent, 0),
  holyResistancePercent: parseIntSafe(form.holyResistancePercent, 0),
  shadowResistancePercent: parseIntSafe(form.shadowResistancePercent, 0),
  progressionIds: Array.from(
    new Set([...form.spellProgressionIds, ...form.techniqueProgressionIds])
  ),
});

const emptyMonsterForm = (): MonsterFormState => ({
  name: '',
  description: '',
  imageUrl: '',
  thumbnailUrl: '',
  zoneId: '',
  zoneKind: '',
  latitude: '',
  longitude: '',
  templateId: '',
  dominantHandInventoryItemId: '',
  offHandInventoryItemId: '',
  level: '1',
  rewardMode: 'random',
  randomRewardSize: 'small',
  rewardExperience: '0',
  rewardGold: '0',
  materialRewards: [],
  itemRewards: [],
});

const monsterFormFromRecord = (monster: MonsterRecord): MonsterFormState => ({
  name: monster.name ?? '',
  description: monster.description ?? '',
  imageUrl: monster.imageUrl ?? '',
  thumbnailUrl: monster.thumbnailUrl ?? '',
  zoneId: monster.zoneId ?? '',
  zoneKind: monster.zoneKind ?? '',
  latitude: String(monster.latitude ?? ''),
  longitude: String(monster.longitude ?? ''),
  templateId: monster.templateId ?? monster.template?.id ?? '',
  dominantHandInventoryItemId:
    monster.dominantHandInventoryItemId !== undefined &&
    monster.dominantHandInventoryItemId !== null
      ? String(monster.dominantHandInventoryItemId)
      : monster.weaponInventoryItemId !== undefined &&
          monster.weaponInventoryItemId !== null
        ? String(monster.weaponInventoryItemId)
        : '',
  offHandInventoryItemId:
    monster.offHandInventoryItemId !== undefined &&
    monster.offHandInventoryItemId !== null
      ? String(monster.offHandInventoryItemId)
      : '',
  level: String(monster.level ?? 1),
  rewardMode: monster.rewardMode === 'explicit' ? 'explicit' : 'random',
  randomRewardSize:
    monster.randomRewardSize === 'medium' ||
    monster.randomRewardSize === 'large'
      ? monster.randomRewardSize
      : 'small',
  rewardExperience: String(monster.rewardExperience ?? 0),
  rewardGold: String(monster.rewardGold ?? 0),
  materialRewards: (monster.materialRewards ?? []).map((reward) => ({
    resourceKey: reward.resourceKey,
    amount: reward.amount ?? 1,
  })),
  itemRewards: (monster.itemRewards ?? []).map((reward) => ({
    inventoryItemId: String(reward.inventoryItemId),
    quantity: String(reward.quantity),
  })),
});

const monsterPayloadFromForm = (form: MonsterFormState) => ({
  name: form.name.trim(),
  description: form.description.trim(),
  imageUrl: form.imageUrl.trim(),
  thumbnailUrl: form.thumbnailUrl.trim(),
  zoneId: form.zoneId.trim(),
  zoneKind: form.zoneKind,
  latitude: parseFloatSafe(form.latitude, 0),
  longitude: parseFloatSafe(form.longitude, 0),
  templateId: form.templateId.trim(),
  dominantHandInventoryItemId: parseIntSafe(
    form.dominantHandInventoryItemId,
    0
  ),
  offHandInventoryItemId: parseOptionalInt(form.offHandInventoryItemId),
  weaponInventoryItemId: parseIntSafe(form.dominantHandInventoryItemId, 0),
  level: parseIntSafe(form.level, 1),
  rewardMode: form.rewardMode,
  randomRewardSize: form.randomRewardSize,
  rewardExperience:
    form.rewardMode === 'explicit' ? parseIntSafe(form.rewardExperience, 0) : 0,
  rewardGold:
    form.rewardMode === 'explicit' ? parseIntSafe(form.rewardGold, 0) : 0,
  materialRewards: form.rewardMode === 'explicit' ? form.materialRewards : [],
  itemRewards: form.itemRewards
    .map((reward) => ({
      inventoryItemId: parseIntSafe(reward.inventoryItemId, 0),
      quantity: parseIntSafe(reward.quantity, 0),
    }))
    .filter((reward) => reward.inventoryItemId > 0 && reward.quantity > 0),
});

const emptyMonsterEncounterForm = (): MonsterEncounterFormState => ({
  name: '',
  description: '',
  imageUrl: '',
  thumbnailUrl: '',
  encounterType: 'monster',
  rewardMode: 'random',
  randomRewardSize: 'small',
  rewardExperience: '0',
  rewardGold: '0',
  materialRewards: [],
  itemRewards: [],
  scaleWithUserLevel: false,
  recurrenceFrequency: '',
  zoneId: '',
  zoneKind: '',
  latitude: '',
  longitude: '',
  monsterIds: [],
});

const monsterEncounterFormFromRecord = (
  encounter: MonsterEncounterRecord
): MonsterEncounterFormState => ({
  name: encounter.name ?? '',
  description: encounter.description ?? '',
  imageUrl: encounter.imageUrl ?? '',
  thumbnailUrl: encounter.thumbnailUrl ?? '',
  encounterType: encounter.encounterType ?? 'monster',
  rewardMode: encounter.rewardMode === 'explicit' ? 'explicit' : 'random',
  randomRewardSize:
    encounter.randomRewardSize === 'medium' ||
    encounter.randomRewardSize === 'large'
      ? encounter.randomRewardSize
      : 'small',
  rewardExperience: String(encounter.rewardExperience ?? 0),
  rewardGold: String(encounter.rewardGold ?? 0),
  materialRewards: (encounter.materialRewards ?? []).map((reward) => ({
    resourceKey: reward.resourceKey,
    amount: reward.amount ?? 1,
  })),
  itemRewards: (encounter.itemRewards ?? []).map((reward) => ({
    inventoryItemId: String(reward.inventoryItemId),
    quantity: String(reward.quantity),
  })),
  scaleWithUserLevel: Boolean(encounter.scaleWithUserLevel),
  recurrenceFrequency: encounter.recurrenceFrequency ?? '',
  zoneId: encounter.zoneId ?? '',
  zoneKind: encounter.zoneKind ?? '',
  latitude: String(encounter.latitude ?? ''),
  longitude: String(encounter.longitude ?? ''),
  monsterIds: (encounter.members ?? [])
    .slice()
    .sort((a, b) => (a.slot ?? 0) - (b.slot ?? 0))
    .map((member) => member.monster?.id)
    .filter((id): id is string => Boolean(id)),
});

const monsterEncounterPayloadFromForm = (form: MonsterEncounterFormState) => ({
  name: form.name.trim(),
  description: form.description.trim(),
  imageUrl: form.imageUrl.trim(),
  thumbnailUrl: form.thumbnailUrl.trim(),
  encounterType: form.encounterType,
  rewardMode: form.rewardMode,
  randomRewardSize: form.randomRewardSize,
  rewardExperience:
    form.rewardMode === 'explicit' ? parseIntSafe(form.rewardExperience, 0) : 0,
  rewardGold:
    form.rewardMode === 'explicit' ? parseIntSafe(form.rewardGold, 0) : 0,
  materialRewards: form.rewardMode === 'explicit' ? form.materialRewards : [],
  itemRewards: form.itemRewards
    .map((reward) => ({
      inventoryItemId: parseIntSafe(reward.inventoryItemId, 0),
      quantity: parseIntSafe(reward.quantity, 0),
    }))
    .filter((reward) => reward.inventoryItemId > 0 && reward.quantity > 0),
  scaleWithUserLevel: form.scaleWithUserLevel,
  recurrenceFrequency: form.recurrenceFrequency,
  zoneId: form.zoneId.trim(),
  zoneKind: form.zoneKind,
  latitude: parseFloatSafe(form.latitude, 0),
  longitude: parseFloatSafe(form.longitude, 0),
  monsterIds: Array.from(
    new Set(
      form.monsterIds.map((id) => id.trim()).filter((id) => id.length > 0)
    )
  ).slice(0, 9),
});

const formatGenerationStatus = (status?: string) => {
  switch ((status || '').trim()) {
    case 'queued':
      return 'Queued';
    case 'in_progress':
      return 'In Progress';
    case 'complete':
      return 'Complete';
    case 'failed':
      return 'Failed';
    default:
      return 'None';
  }
};

const summarizeAffinityMap = (
  values: Record<string, number> | undefined,
  options: ReadonlyArray<{ affinity: string; label: string }>
): string => {
  if (!values) return 'None';
  const parts = options
    .map(({ affinity, label }) => ({
      label,
      value: values[affinity] ?? 0,
    }))
    .filter(({ value }) => value !== 0)
    .map(({ label, value }) => `${label} ${value > 0 ? '+' : ''}${value}`);
  return parts.length > 0 ? parts.join(' · ') : 'None';
};

const monsterTemplateSuggestionDamageMap = (
  payload?: MonsterTemplateSuggestionPayload
): Record<string, number> => ({
  physical: payload?.physicalDamageBonusPercent ?? 0,
  piercing: payload?.piercingDamageBonusPercent ?? 0,
  slashing: payload?.slashingDamageBonusPercent ?? 0,
  bludgeoning: payload?.bludgeoningDamageBonusPercent ?? 0,
  fire: payload?.fireDamageBonusPercent ?? 0,
  ice: payload?.iceDamageBonusPercent ?? 0,
  lightning: payload?.lightningDamageBonusPercent ?? 0,
  poison: payload?.poisonDamageBonusPercent ?? 0,
  arcane: payload?.arcaneDamageBonusPercent ?? 0,
  holy: payload?.holyDamageBonusPercent ?? 0,
  shadow: payload?.shadowDamageBonusPercent ?? 0,
});

const monsterTemplateSuggestionResistanceMap = (
  payload?: MonsterTemplateSuggestionPayload
): Record<string, number> => ({
  physical: payload?.physicalResistancePercent ?? 0,
  piercing: payload?.piercingResistancePercent ?? 0,
  slashing: payload?.slashingResistancePercent ?? 0,
  bludgeoning: payload?.bludgeoningResistancePercent ?? 0,
  fire: payload?.fireResistancePercent ?? 0,
  ice: payload?.iceResistancePercent ?? 0,
  lightning: payload?.lightningResistancePercent ?? 0,
  poison: payload?.poisonResistancePercent ?? 0,
  arcane: payload?.arcaneResistancePercent ?? 0,
  holy: payload?.holyResistancePercent ?? 0,
  shadow: payload?.shadowResistancePercent ?? 0,
});

const formatMonsterEncounterTypeLabel = (
  encounterType?: string | null
): string => {
  switch ((encounterType || '').trim()) {
    case 'boss':
      return 'Boss Encounter';
    case 'raid':
      return 'Raid Encounter';
    default:
      return 'Monster Encounter';
  }
};

const formatMonsterTemplateTypeLabel = (
  monsterType?: string | null
): string => {
  switch ((monsterType || '').trim()) {
    case 'boss':
      return 'Boss Template';
    case 'raid':
      return 'Raid Template';
    default:
      return 'Standard Template';
  }
};

const formatBulkTemplateStatus = (status?: string): string => {
  switch ((status || '').trim()) {
    case 'queued':
      return 'Queued';
    case 'in_progress':
      return 'In Progress';
    case 'completed':
      return 'Completed';
    case 'failed':
      return 'Failed';
    default:
      return 'Unknown';
  }
};

const staticStatusClassName = (status: string): string => {
  switch ((status || '').trim()) {
    case 'queued':
      return 'bg-slate-600';
    case 'in_progress':
      return 'bg-amber-600';
    case 'completed':
      return 'bg-emerald-600';
    case 'failed':
      return 'bg-red-600';
    case 'missing':
      return 'bg-gray-500';
    default:
      return 'bg-gray-400';
  }
};

const formatDate = (value?: string): string => {
  if (!value) return 'n/a';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

const emptyEncounterIconState = (
  encounterType: MonsterEncounterType
): EncounterIconState => ({
  url: encounterIconDefaults[encounterType].url,
  status: 'unknown',
  exists: false,
  requestedAt: null,
  lastModified: null,
  previewNonce: Date.now(),
  prompt: encounterIconDefaults[encounterType].prompt,
  message: null,
  error: null,
  busy: false,
  statusLoading: false,
});

mapboxgl.accessToken = process.env.REACT_APP_MAPBOX_ACCESS_TOKEN || '';

export const Monsters = () => {
  const { apiClient } = useAPI();
  const { zones } = useZoneContext();
  const { zoneKinds, zoneKindBySlug } = useZoneKinds();

  const [loading, setLoading] = useState(true);
  const [referenceLoading, setReferenceLoading] = useState(true);
  const [templates, setTemplates] = useState<MonsterTemplateRecord[]>([]);
  const [records, setRecords] = useState<MonsterRecord[]>([]);
  const [encounters, setEncounters] = useState<MonsterEncounterRecord[]>([]);
  const [templateOptions, setTemplateOptions] = useState<
    MonsterTemplateRecord[]
  >([]);
  const [monsterOptions, setMonsterOptions] = useState<MonsterRecord[]>([]);
  const [spells, setSpells] = useState<Spell[]>([]);
  const [genres, setGenres] = useState<ZoneGenre[]>([]);
  const [inventoryItems, setInventoryItems] = useState<InventoryItemLite[]>([]);
  const [query, setQuery] = useState('');
  const [zoneQuery, setZoneQuery] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [bulkDeletingMonsters, setBulkDeletingMonsters] = useState(false);
  const [bulkDeletingEncounters, setBulkDeletingEncounters] = useState(false);
  const [selectedMonsterIds, setSelectedMonsterIds] = useState<Set<string>>(
    new Set()
  );
  const [selectedEncounterIds, setSelectedEncounterIds] = useState<Set<string>>(
    new Set()
  );
  const [templatePage, setTemplatePage] = useState(1);
  const [monsterPage, setMonsterPage] = useState(1);
  const [encounterPage, setEncounterPage] = useState(1);
  const [templateTotal, setTemplateTotal] = useState(0);
  const [monsterTotal, setMonsterTotal] = useState(0);
  const [encounterTotal, setEncounterTotal] = useState(0);
  const [activeTemplateCount, setActiveTemplateCount] = useState(0);
  const [archivedTemplateCount, setArchivedTemplateCount] = useState(0);
  const [templateOptionsLoaded, setTemplateOptionsLoaded] = useState(false);
  const [monsterOptionsLoaded, setMonsterOptionsLoaded] = useState(false);

  const [showTemplateModal, setShowTemplateModal] = useState(false);
  const [editingTemplate, setEditingTemplate] =
    useState<MonsterTemplateRecord | null>(null);
  const [templateTab, setTemplateTab] = useState<'active' | 'archived'>(
    'active'
  );
  const [selectedTemplateIds, setSelectedTemplateIds] = useState<Set<string>>(
    new Set()
  );
  const [templateForm, setTemplateForm] =
    useState<MonsterTemplateFormState>(emptyTemplateForm());

  const [showMonsterModal, setShowMonsterModal] = useState(false);
  const [editingMonster, setEditingMonster] = useState<MonsterRecord | null>(
    null
  );
  const [monsterForm, setMonsterForm] =
    useState<MonsterFormState>(emptyMonsterForm());
  const [showEncounterModal, setShowEncounterModal] = useState(false);
  const [editingEncounter, setEditingEncounter] =
    useState<MonsterEncounterRecord | null>(null);
  const [encounterForm, setEncounterForm] = useState<MonsterEncounterFormState>(
    emptyMonsterEncounterForm()
  );
  const [imagePreview, setImagePreview] = useState<ImagePreviewState | null>(
    null
  );
  const [geoLoading, setGeoLoading] = useState(false);
  const [encounterIconStates, setEncounterIconStates] = useState<
    Record<MonsterEncounterType, EncounterIconState>
  >({
    monster: emptyEncounterIconState('monster'),
    boss: emptyEncounterIconState('boss'),
    raid: emptyEncounterIconState('raid'),
  });

  const [generatingMonsterId, setGeneratingMonsterId] = useState<string | null>(
    null
  );
  const [generatingTemplateId, setGeneratingTemplateId] = useState<
    string | null
  >(null);
  const [bulkTemplateCount, setBulkTemplateCount] = useState('8');
  const [bulkTemplateType, setBulkTemplateType] =
    useState<MonsterTemplateType>('monster');
  const [bulkTemplateGenreId, setBulkTemplateGenreId] = useState('');
  const [bulkTemplateZoneKind, setBulkTemplateZoneKind] = useState('');
  const [bulkTemplateYeetIt, setBulkTemplateYeetIt] = useState(false);
  const [genreFilter, setGenreFilter] = useState('all');
  const [templateTypeFilter, setTemplateTypeFilter] = useState<
    'all' | MonsterTemplateType
  >('all');
  const [bulkTemplateBusy, setBulkTemplateBusy] = useState(false);
  const [bulkTemplateJob, setBulkTemplateJob] =
    useState<BulkMonsterTemplateStatus | null>(null);
  const [bulkTemplateError, setBulkTemplateError] = useState<string | null>(
    null
  );
  const [bulkTemplateMessage, setBulkTemplateMessage] = useState<string | null>(
    null
  );
  const [templateSuggestionJobs, setTemplateSuggestionJobs] = useState<
    BulkMonsterTemplateStatus[]
  >([]);
  const [templateSuggestionJobsPage, setTemplateSuggestionJobsPage] =
    useState(1);
  const [templateSuggestionJobsHasMore, setTemplateSuggestionJobsHasMore] =
    useState(false);
  const [templateSuggestionJobsLoadedLimit, setTemplateSuggestionJobsLoadedLimit] =
    useState(monsterTemplateSuggestionJobsPerPage);
  const [selectedTemplateSuggestionJobId, setSelectedTemplateSuggestionJobId] =
    useState('');
  const [templateSuggestionDrafts, setTemplateSuggestionDrafts] = useState<
    MonsterTemplateSuggestionDraft[]
  >([]);
  const [templateSuggestionDraftPage, setTemplateSuggestionDraftPage] =
    useState(1);
  const [loadingTemplateSuggestionJobs, setLoadingTemplateSuggestionJobs] =
    useState(false);
  const [loadingTemplateSuggestionDrafts, setLoadingTemplateSuggestionDrafts] =
    useState(false);
  const [
    convertingTemplateSuggestionDraftId,
    setConvertingTemplateSuggestionDraftId,
  ] = useState<string | null>(null);
  const [
    deletingTemplateSuggestionDraftId,
    setDeletingTemplateSuggestionDraftId,
  ] = useState<string | null>(null);
  const [
    convertingAllTemplateSuggestionDrafts,
    setConvertingAllTemplateSuggestionDrafts,
  ] = useState(false);
  const [affinityRefreshBusy, setAffinityRefreshBusy] = useState(false);
  const [affinityRefreshJob, setAffinityRefreshJob] =
    useState<MonsterTemplateAffinityRefreshStatus | null>(null);
  const [affinityRefreshError, setAffinityRefreshError] = useState<
    string | null
  >(null);
  const [affinityRefreshMessage, setAffinityRefreshMessage] = useState<
    string | null
  >(null);
  const [progressionResetBusy, setProgressionResetBusy] = useState(false);
  const [progressionResetJob, setProgressionResetJob] =
    useState<MonsterTemplateProgressionResetStatus | null>(null);
  const [progressionResetError, setProgressionResetError] = useState<
    string | null
  >(null);
  const [progressionResetMessage, setProgressionResetMessage] = useState<
    string | null
  >(null);
  const mapContainerRef = React.useRef<HTMLDivElement | null>(null);
  const mapRef = React.useRef<mapboxgl.Map | null>(null);
  const markerRef = React.useRef<mapboxgl.Marker | null>(null);
  const formLatitudeRef = React.useRef(monsterForm.latitude);
  const formLongitudeRef = React.useRef(monsterForm.longitude);
  const encounterMapContainerRef = React.useRef<HTMLDivElement | null>(null);
  const encounterMapRef = React.useRef<mapboxgl.Map | null>(null);
  const encounterMarkerRef = React.useRef<mapboxgl.Marker | null>(null);
  const encounterLatitudeRef = React.useRef(encounterForm.latitude);
  const encounterLongitudeRef = React.useRef(encounterForm.longitude);
  const deferredQuery = useDeferredValue(query);
  const deferredZoneQuery = useDeferredValue(zoneQuery);
  const defaultGenreId = useMemo(
    () => defaultGenreIdFromList(genres),
    [genres]
  );

  const loadReferenceData = useCallback(async () => {
    try {
      setReferenceLoading(true);
      const [spellResp, inventoryResp, genreResp] = await Promise.all([
        apiClient.get<Spell[]>('/sonar/spells'),
        apiClient.get<InventoryItemLite[]>('/sonar/inventory-items'),
        apiClient.get<ZoneGenre[]>('/sonar/zone-genres?includeInactive=true'),
      ]);
      setSpells(Array.isArray(spellResp) ? spellResp : []);
      setInventoryItems(Array.isArray(inventoryResp) ? inventoryResp : []);
      setGenres(Array.isArray(genreResp) ? genreResp : []);
    } catch (err) {
      console.error('Failed to load monster admin reference data', err);
      setError('Failed to load monster admin reference data.');
    } finally {
      setReferenceLoading(false);
    }
  }, [apiClient]);

  const loadTemplateOptions = useCallback(async () => {
    const response = await apiClient.get<MonsterTemplateRecord[]>(
      '/sonar/monster-templates'
    );
    const nextTemplates = Array.isArray(response) ? response : [];
    setTemplateOptions(nextTemplates);
    setTemplateOptionsLoaded(true);
    return nextTemplates;
  }, [apiClient]);

  const loadMonsterOptions = useCallback(async () => {
    const response = await apiClient.get<MonsterRecord[]>('/sonar/monsters');
    const nextMonsters = Array.isArray(response) ? response : [];
    setMonsterOptions(nextMonsters);
    setMonsterOptionsLoaded(true);
    return nextMonsters;
  }, [apiClient]);

  const loadPagedData = useCallback(
    async (suppressLoading = false) => {
      try {
        if (!suppressLoading) {
          setLoading(true);
        }
        setError(null);
        const [templateResp, monsterResp, encounterResp] = await Promise.all([
          apiClient.get<PaginatedResponse<MonsterTemplateRecord>>(
            '/sonar/admin/monster-templates',
            {
              page: templatePage,
              pageSize: monsterListPageSize,
              query: deferredQuery.trim(),
              zoneQuery: deferredZoneQuery.trim(),
              genreId: genreFilter === 'all' ? '' : genreFilter,
              archived: templateTab === 'archived',
              monsterType:
                templateTypeFilter === 'all' ? '' : templateTypeFilter,
            }
          ),
          apiClient.get<PaginatedResponse<MonsterRecord>>(
            '/sonar/admin/monsters',
            {
              page: monsterPage,
              pageSize: monsterListPageSize,
              query: deferredQuery.trim(),
              zoneQuery: deferredZoneQuery.trim(),
              genreId: genreFilter === 'all' ? '' : genreFilter,
            }
          ),
          apiClient.get<PaginatedResponse<MonsterEncounterRecord>>(
            '/sonar/admin/monster-encounters',
            {
              page: encounterPage,
              pageSize: monsterListPageSize,
              query: deferredQuery.trim(),
              zoneQuery: deferredZoneQuery.trim(),
              genreId: genreFilter === 'all' ? '' : genreFilter,
            }
          ),
        ]);

        setTemplates(
          Array.isArray(templateResp?.items) ? templateResp.items : []
        );
        setRecords(Array.isArray(monsterResp?.items) ? monsterResp.items : []);
        setEncounters(
          Array.isArray(encounterResp?.items) ? encounterResp.items : []
        );
        setTemplateTotal(templateResp?.total ?? 0);
        setMonsterTotal(monsterResp?.total ?? 0);
        setEncounterTotal(encounterResp?.total ?? 0);
        setActiveTemplateCount(templateResp?.activeCount ?? 0);
        setArchivedTemplateCount(templateResp?.archivedCount ?? 0);
      } catch (err) {
        console.error('Failed to load monster admin lists', err);
        setError('Failed to load monster admin lists.');
      } finally {
        if (!suppressLoading) {
          setLoading(false);
        }
      }
    },
    [
      apiClient,
      deferredQuery,
      deferredZoneQuery,
      encounterPage,
      genreFilter,
      monsterPage,
      templatePage,
      templateTab,
      templateTypeFilter,
    ]
  );

  const fetchTemplateSuggestionJobs = useCallback(async () => {
    try {
      setLoadingTemplateSuggestionJobs(true);
      const requestedLimit =
        Math.max(
          monsterTemplateSuggestionJobsPerPage,
          templateSuggestionJobsLoadedLimit
        ) + 1;
      const response = await apiClient.get<BulkMonsterTemplateStatus[]>(
        '/sonar/monster-template-suggestion-jobs',
        { limit: requestedLimit }
      );
      const jobs = Array.isArray(response) ? response : [];
      const hasMore = jobs.length > templateSuggestionJobsLoadedLimit;
      const loadedJobs = hasMore
        ? jobs.slice(0, templateSuggestionJobsLoadedLimit)
        : jobs;
      setTemplateSuggestionJobs(loadedJobs);
      setTemplateSuggestionJobsHasMore(hasMore);
      setSelectedTemplateSuggestionJobId((current) => {
        if (current && loadedJobs.some((job) => job.jobId === current)) {
          return current;
        }
        return loadedJobs[0]?.jobId ?? '';
      });
    } catch (err) {
      console.error('Failed to load monster template suggestion jobs', err);
    } finally {
      setLoadingTemplateSuggestionJobs(false);
    }
  }, [apiClient, templateSuggestionJobsLoadedLimit]);

  const fetchTemplateSuggestionDrafts = useCallback(
    async (jobId: string) => {
      const trimmedJobId = (jobId || '').trim();
      if (!trimmedJobId) {
        setTemplateSuggestionDrafts([]);
        return;
      }
      try {
        setLoadingTemplateSuggestionDrafts(true);
        const response = await apiClient.get<MonsterTemplateSuggestionDraft[]>(
          `/sonar/monster-template-suggestion-jobs/${trimmedJobId}/drafts`
        );
        setTemplateSuggestionDrafts(Array.isArray(response) ? response : []);
      } catch (err) {
        console.error('Failed to load monster template suggestion drafts', err);
        setTemplateSuggestionDrafts([]);
      } finally {
        setLoadingTemplateSuggestionDrafts(false);
      }
    },
    [apiClient]
  );

  const refreshBulkTemplateJobStatus = useCallback(
    async (jobId: string) => {
      try {
        const status = await apiClient.get<BulkMonsterTemplateStatus>(
          `/sonar/monster-templates/bulk-generate/${jobId}/status`
        );
        setBulkTemplateJob(status);
        if (status.status === 'completed') {
          setBulkTemplateBusy(false);
          setBulkTemplateError(null);
          const typeLabel = formatMonsterTemplateTypeLabel(
            status.monsterType ?? bulkTemplateType
          );
          setBulkTemplateMessage(
            `${status.yeetIt ? 'Created' : 'Prepared'} ${status.createdCount} ${typeLabel} ${
              status.yeetIt ? 'template' : 'draft'
            }${
              status.createdCount === 1 ? '' : 's'
            }${status.yeetIt ? ' live.' : '.'}`
          );
          setSelectedTemplateSuggestionJobId(status.jobId);
          await fetchTemplateSuggestionJobs();
          if (status.yeetIt) {
            setTemplateSuggestionDrafts([]);
          } else {
            await fetchTemplateSuggestionDrafts(status.jobId);
          }
        } else if (status.status === 'failed') {
          setBulkTemplateBusy(false);
          setBulkTemplateMessage(null);
          setBulkTemplateError(
            status.error || 'Bulk template generation failed.'
          );
        }
      } catch (err) {
        console.error('Failed to refresh bulk monster template status', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to refresh bulk template generation status.';
        setBulkTemplateError(message);
        setBulkTemplateBusy(false);
      }
    },
    [
      apiClient,
      bulkTemplateType,
      fetchTemplateSuggestionDrafts,
      fetchTemplateSuggestionJobs,
    ]
  );

  useEffect(() => {
    void loadReferenceData();
  }, [loadReferenceData]);

  useEffect(() => {
    if (!bulkTemplateGenreId && defaultGenreId) {
      setBulkTemplateGenreId(defaultGenreId);
    }
  }, [bulkTemplateGenreId, defaultGenreId]);

  useEffect(() => {
    if (!showTemplateModal) return;
    if (!templateForm.genreId && defaultGenreId) {
      setTemplateForm((prev) =>
        prev.genreId ? prev : { ...prev, genreId: defaultGenreId }
      );
    }
  }, [defaultGenreId, showTemplateModal, templateForm.genreId]);

  useEffect(() => {
    void loadPagedData();
  }, [loadPagedData]);

  useEffect(() => {
    void fetchTemplateSuggestionJobs();
  }, [fetchTemplateSuggestionJobs]);

  useEffect(() => {
    if (!selectedTemplateSuggestionJobId) {
      setTemplateSuggestionDrafts([]);
      return;
    }
    void fetchTemplateSuggestionDrafts(selectedTemplateSuggestionJobId);
  }, [fetchTemplateSuggestionDrafts, selectedTemplateSuggestionJobId]);

  useEffect(() => {
    setTemplateSuggestionDraftPage(1);
  }, [selectedTemplateSuggestionJobId]);

  useEffect(() => {
    const requiredJobCount =
      templateSuggestionJobsPage * monsterTemplateSuggestionJobsPerPage;
    if (
      requiredJobCount > templateSuggestionJobsLoadedLimit &&
      (templateSuggestionJobsHasMore ||
        templateSuggestionJobs.length < requiredJobCount)
    ) {
      setTemplateSuggestionJobsLoadedLimit(requiredJobCount);
    }
  }, [
    templateSuggestionJobs.length,
    templateSuggestionJobsHasMore,
    templateSuggestionJobsLoadedLimit,
    templateSuggestionJobsPage,
  ]);

  useEffect(() => {
    const totalPages = Math.max(
      1,
      Math.ceil(templateSuggestionJobs.length / monsterTemplateSuggestionJobsPerPage)
    );
    setTemplateSuggestionJobsPage((current) => Math.min(current, totalPages));
  }, [templateSuggestionJobs.length]);

  useEffect(() => {
    const totalPages = Math.max(
      1,
      Math.ceil(
        templateSuggestionDrafts.length / monsterTemplateSuggestionDraftsPerPage
      )
    );
    setTemplateSuggestionDraftPage((current) => Math.min(current, totalPages));
  }, [templateSuggestionDrafts.length]);

  useEffect(() => {
    const hasPendingMonsterGeneration = records.some((record) =>
      ['queued', 'in_progress'].includes(record.imageGenerationStatus || '')
    );
    const hasPendingTemplateGeneration = templates.some((template) =>
      ['queued', 'in_progress'].includes(template.imageGenerationStatus || '')
    );
    const hasPendingGeneration =
      hasPendingMonsterGeneration || hasPendingTemplateGeneration;
    if (!hasPendingGeneration) return;

    const interval = window.setInterval(() => {
      void loadPagedData(true);
    }, 5000);

    return () => clearInterval(interval);
  }, [loadPagedData, records, templates]);

  useEffect(() => {
    setTemplatePage(1);
    setMonsterPage(1);
    setEncounterPage(1);
  }, [genreFilter, query, zoneQuery]);

  useEffect(() => {
    setTemplatePage(1);
  }, [templateTab, templateTypeFilter]);

  useEffect(() => {
    const totalPages = Math.max(
      1,
      Math.ceil(templateTotal / monsterListPageSize)
    );
    if (templatePage > totalPages) {
      setTemplatePage(totalPages);
    }
  }, [templatePage, templateTotal]);

  useEffect(() => {
    const totalPages = Math.max(
      1,
      Math.ceil(monsterTotal / monsterListPageSize)
    );
    if (monsterPage > totalPages) {
      setMonsterPage(totalPages);
    }
  }, [monsterPage, monsterTotal]);

  useEffect(() => {
    const totalPages = Math.max(
      1,
      Math.ceil(encounterTotal / monsterListPageSize)
    );
    if (encounterPage > totalPages) {
      setEncounterPage(totalPages);
    }
  }, [encounterPage, encounterTotal]);

  useEffect(() => {
    setSelectedTemplateIds((prev) => {
      if (prev.size === 0) return prev;
      const validIDs = new Set(templates.map((template) => template.id));
      const next = new Set<string>();
      prev.forEach((id) => {
        if (validIDs.has(id)) {
          next.add(id);
        }
      });
      return next.size === prev.size ? prev : next;
    });
  }, [templates]);

  useEffect(() => {
    setSelectedTemplateIds(new Set());
  }, [templateTab]);

  useEffect(() => {
    setSelectedMonsterIds((prev) => {
      if (prev.size === 0) return prev;
      const validIDs = new Set(records.map((monster) => monster.id));
      const next = new Set<string>();
      prev.forEach((id) => {
        if (validIDs.has(id)) {
          next.add(id);
        }
      });
      return next.size === prev.size ? prev : next;
    });
  }, [records]);

  useEffect(() => {
    setSelectedEncounterIds((prev) => {
      if (prev.size === 0) return prev;
      const validIDs = new Set(encounters.map((encounter) => encounter.id));
      const next = new Set<string>();
      prev.forEach((id) => {
        if (validIDs.has(id)) {
          next.add(id);
        }
      });
      return next.size === prev.size ? prev : next;
    });
  }, [encounters]);

  useEffect(() => {
    formLatitudeRef.current = monsterForm.latitude;
    formLongitudeRef.current = monsterForm.longitude;
  }, [monsterForm.latitude, monsterForm.longitude]);

  useEffect(() => {
    encounterLatitudeRef.current = encounterForm.latitude;
    encounterLongitudeRef.current = encounterForm.longitude;
  }, [encounterForm.latitude, encounterForm.longitude]);

  const zoneNameById = useMemo(() => {
    const map = new Map<string, string>();
    for (const zone of zones) {
      map.set(zone.id, zone.name);
    }
    return map;
  }, [zones]);
  const zoneDefaultKindById = useMemo(() => {
    const map = new Map<string, string>();
    for (const zone of zones) {
      map.set(zone.id, zone.kind?.trim() ?? '');
    }
    return map;
  }, [zones]);
  const selectedMonsterZoneDefaultKind = useMemo(
    () => zoneDefaultKindById.get(monsterForm.zoneId) ?? '',
    [monsterForm.zoneId, zoneDefaultKindById]
  );
  const selectedEncounterZoneDefaultKind = useMemo(
    () => zoneDefaultKindById.get(encounterForm.zoneId) ?? '',
    [encounterForm.zoneId, zoneDefaultKindById]
  );
  const sharedDashboardParams = useMemo(
    () => ({
      query: deferredQuery.trim(),
      zoneQuery: deferredZoneQuery.trim(),
      genreId: genreFilter === 'all' ? '' : genreFilter,
    }),
    [deferredQuery, deferredZoneQuery, genreFilter]
  );
  const templateDashboardParams = useMemo(
    () => ({
      ...sharedDashboardParams,
      archived: templateTab === 'archived' ? 'true' : 'false',
      monsterType: templateTypeFilter === 'all' ? '' : templateTypeFilter,
    }),
    [sharedDashboardParams, templateTab, templateTypeFilter]
  );
  const {
    items: dashboardTemplates,
    loading: dashboardTemplatesLoading,
    error: dashboardTemplatesError,
  } = useAdminAggregateDataset<MonsterTemplateRecord>(
    '/sonar/admin/monster-templates',
    templateDashboardParams
  );
  const dashboardLoading = dashboardTemplatesLoading;
  const dashboardError = dashboardTemplatesError;
  const dashboardMetrics = useMemo(() => {
    const templatesWithSpells = dashboardTemplates.filter(
      (template) => (template.spells?.length ?? 0) > 0
    ).length;
    const templatesWithProgressions = dashboardTemplates.filter(
      (template) => (template.progressions?.length ?? 0) > 0
    ).length;
    return [
      { label: 'Visible Templates', value: dashboardTemplates.length },
      { label: 'Active Templates', value: activeTemplateCount },
      { label: 'Archived Templates', value: archivedTemplateCount },
      { label: 'With Spells', value: templatesWithSpells },
      { label: 'With Progressions', value: templatesWithProgressions },
    ];
  }, [
    activeTemplateCount,
    archivedTemplateCount,
    dashboardTemplates,
  ]);
  const dashboardSections = useMemo(
    () => [
      {
        title: 'Template Genres',
        note: 'Genre mix for the currently visible template pool.',
        buckets: countBy(
          dashboardTemplates,
          (template) =>
            formatGenreLabel(
              template.genre ??
                genres.find((genre) => genre.id === template.genreId) ??
                null
            ),
          { emptyLabel: 'Fantasy' }
        ),
      },
      {
        title: 'Template Types',
        note: 'How the template pool splits across monster roles.',
        buckets: countBy(dashboardTemplates, (template) =>
          formatMonsterTemplateTypeLabel(template.monsterType)
        ),
      },
      {
        title: 'Template Zone Kinds',
        note: 'Likely habitat classification across the current template pool.',
        buckets: countBy(dashboardTemplates, (template) =>
          zoneKindLabel(template.zoneKind, zoneKindBySlug)
        ),
      },
      {
        title: 'Spell Loadout',
        note: 'How many direct spells are configured per template.',
        buckets: countBy(dashboardTemplates, (template) =>
          formatTemplateSpellLoadoutLabel(template)
        ),
      },
      {
        title: 'Progression Coverage',
        note: 'How many progression tracks each template references.',
        buckets: countBy(dashboardTemplates, (template) =>
          formatTemplateProgressionCoverageLabel(template)
        ),
      },
    ],
    [
      dashboardTemplates,
      genres,
      zoneKindBySlug,
    ]
  );

  const templateNameById = useMemo(() => {
    const map = new Map<string, string>();
    for (const template of templateOptions) {
      map.set(template.id, template.name);
    }
    return map;
  }, [templateOptions]);

  const dominantHandItems = useMemo(() => {
    return inventoryItems.filter((item) => {
      const equipSlot = (item.equipSlot ?? '').trim().toLowerCase();
      const category = (item.handItemCategory ?? '').trim().toLowerCase();
      return (
        equipSlot === 'dominant_hand' &&
        (category === 'weapon' || category === 'staff') &&
        item.damageMin !== undefined &&
        item.damageMin !== null &&
        item.damageMax !== undefined &&
        item.damageMax !== null
      );
    });
  }, [inventoryItems]);

  const offHandItems = useMemo(() => {
    return inventoryItems.filter((item) => {
      const equipSlot = (item.equipSlot ?? '').trim().toLowerCase();
      const category = (item.handItemCategory ?? '').trim().toLowerCase();
      const handedness = (item.handedness ?? '').trim().toLowerCase();
      const isOffhandUtility =
        equipSlot === 'off_hand' &&
        handedness === 'one_handed' &&
        (category === 'shield' || category === 'orb');
      const isOneHandedWeapon =
        equipSlot === 'dominant_hand' &&
        handedness === 'one_handed' &&
        category === 'weapon' &&
        item.damageMin !== undefined &&
        item.damageMin !== null &&
        item.damageMax !== undefined &&
        item.damageMax !== null;
      return isOffhandUtility || isOneHandedWeapon;
    });
  }, [inventoryItems]);

  const dominantHandItemById = useMemo(() => {
    const map = new Map<number, InventoryItemLite>();
    for (const item of dominantHandItems) {
      map.set(item.id, item);
    }
    return map;
  }, [dominantHandItems]);

  const selectedDominantHandItem = useMemo(() => {
    const selectedID = parseIntSafe(monsterForm.dominantHandInventoryItemId, 0);
    return selectedID > 0 ? dominantHandItemById.get(selectedID) : undefined;
  }, [monsterForm.dominantHandInventoryItemId, dominantHandItemById]);

  const dominantIsTwoHanded =
    (selectedDominantHandItem?.handedness ?? '').trim().toLowerCase() ===
    'two_handed';

  const progressionOptions = useMemo(() => {
    const groups = new Map<string, MonsterAbilityProgressionOption>();
    for (const spell of spells) {
      const link = spell.progressionLinks?.[0];
      if (!link?.progressionId) {
        continue;
      }
      const abilityType =
        (spell.abilityType ?? 'spell') === 'technique' ? 'technique' : 'spell';
      const existing = groups.get(link.progressionId);
      if (existing) {
        existing.memberCount += 1;
        if (
          link.progression?.name &&
          link.progression.name.trim() &&
          existing.name.startsWith('Progression ')
        ) {
          existing.name = link.progression.name.trim();
        }
        continue;
      }
      groups.set(link.progressionId, {
        id: link.progressionId,
        name:
          link.progression?.name?.trim() || `Progression ${groups.size + 1}`,
        abilityType,
        memberCount: 1,
      });
    }
    return Array.from(groups.values()).sort((left, right) =>
      left.name.localeCompare(right.name)
    );
  }, [spells]);
  const spellProgressionOptions = useMemo(
    () =>
      progressionOptions.filter((option) => option.abilityType !== 'technique'),
    [progressionOptions]
  );
  const techniqueProgressionOptions = useMemo(
    () =>
      progressionOptions.filter((option) => option.abilityType === 'technique'),
    [progressionOptions]
  );

  const filteredTemplates = templates;
  const filteredMonsters = records;
  const filteredEncounters = encounters;
  const allFilteredTemplatesSelected = useMemo(() => {
    if (filteredTemplates.length === 0) return false;
    return filteredTemplates.every((template) =>
      selectedTemplateIds.has(template.id)
    );
  }, [filteredTemplates, selectedTemplateIds]);
  const selectedMonsterIdSet = useMemo(
    () => selectedMonsterIds,
    [selectedMonsterIds]
  );
  const allFilteredMonstersSelected = useMemo(() => {
    if (filteredMonsters.length === 0) return false;
    return filteredMonsters.every((monster) =>
      selectedMonsterIds.has(monster.id)
    );
  }, [filteredMonsters, selectedMonsterIds]);
  const selectedEncounterIdSet = useMemo(
    () => selectedEncounterIds,
    [selectedEncounterIds]
  );
  const allFilteredEncountersSelected = useMemo(() => {
    if (filteredEncounters.length === 0) return false;
    return filteredEncounters.every((encounter) =>
      selectedEncounterIds.has(encounter.id)
    );
  }, [filteredEncounters, selectedEncounterIds]);
  const selectedTemplateSuggestionJob = useMemo(
    () =>
      templateSuggestionJobs.find(
        (job) => job.jobId === selectedTemplateSuggestionJobId
      ) ?? null,
    [selectedTemplateSuggestionJobId, templateSuggestionJobs]
  );
  const visibleTemplateSuggestionJobs = useMemo(() => {
    const startIndex =
      (templateSuggestionJobsPage - 1) * monsterTemplateSuggestionJobsPerPage;
    return templateSuggestionJobs.slice(
      startIndex,
      startIndex + monsterTemplateSuggestionJobsPerPage
    );
  }, [templateSuggestionJobs, templateSuggestionJobsPage]);
  const templateSuggestionJobsLoadedPages = useMemo(
    () =>
      Math.max(
        1,
        Math.ceil(
          templateSuggestionJobs.length / monsterTemplateSuggestionJobsPerPage
        )
      ),
    [templateSuggestionJobs.length]
  );
  const templateSuggestionJobsCanAdvance =
    templateSuggestionJobsPage < templateSuggestionJobsLoadedPages ||
    templateSuggestionJobsHasMore;
  const visibleTemplateSuggestionJobsRangeStart =
    templateSuggestionJobs.length === 0
      ? 0
      : (templateSuggestionJobsPage - 1) * monsterTemplateSuggestionJobsPerPage +
        1;
  const visibleTemplateSuggestionJobsRangeEnd =
    templateSuggestionJobs.length === 0
      ? 0
      : Math.min(
          templateSuggestionJobs.length,
          templateSuggestionJobsPage * monsterTemplateSuggestionJobsPerPage
        );
  const templateSuggestionDraftTotalPages = useMemo(
    () =>
      Math.max(
        1,
        Math.ceil(
          templateSuggestionDrafts.length / monsterTemplateSuggestionDraftsPerPage
        )
      ),
    [templateSuggestionDrafts.length]
  );
  const paginatedTemplateSuggestionDrafts = useMemo(() => {
    const startIndex =
      (templateSuggestionDraftPage - 1) * monsterTemplateSuggestionDraftsPerPage;
    return templateSuggestionDrafts.slice(
      startIndex,
      startIndex + monsterTemplateSuggestionDraftsPerPage
    );
  }, [templateSuggestionDraftPage, templateSuggestionDrafts]);
  const templateSuggestionDraftRangeStart =
    templateSuggestionDrafts.length === 0
      ? 0
      : (templateSuggestionDraftPage - 1) *
          monsterTemplateSuggestionDraftsPerPage +
        1;
  const templateSuggestionDraftRangeEnd =
    templateSuggestionDrafts.length === 0
      ? 0
      : Math.min(
          templateSuggestionDrafts.length,
          templateSuggestionDraftPage * monsterTemplateSuggestionDraftsPerPage
        );
  const unconvertedTemplateSuggestionDrafts = useMemo(
    () =>
      templateSuggestionDrafts.filter(
        (draft) => (draft.status || '').trim().toLowerCase() !== 'converted'
      ),
    [templateSuggestionDrafts]
  );

  const encounterMonsterOptions = useMemo(() => {
    if (!encounterForm.zoneId) {
      return monsterOptions;
    }
    return monsterOptions.filter(
      (monster) => monster.zoneId === encounterForm.zoneId
    );
  }, [encounterForm.zoneId, monsterOptions]);

  const openCreateTemplate = () => {
    setEditingTemplate(null);
    setTemplateForm(emptyTemplateForm(defaultGenreId));
    setShowTemplateModal(true);
  };

  const openEditTemplate = (template: MonsterTemplateRecord) => {
    setEditingTemplate(template);
    setTemplateForm(templateFormFromRecord(template));
    setShowTemplateModal(true);
  };

  const closeTemplateModal = () => {
    setShowTemplateModal(false);
    setEditingTemplate(null);
    setTemplateForm(emptyTemplateForm(defaultGenreId));
  };

  const saveTemplate = async () => {
    try {
      const payload = templatePayloadFromForm(templateForm);
      if (!payload.name) {
        alert('Template name is required.');
        return;
      }
      if (
        payload.baseStrength <= 0 ||
        payload.baseDexterity <= 0 ||
        payload.baseConstitution <= 0 ||
        payload.baseIntelligence <= 0 ||
        payload.baseWisdom <= 0 ||
        payload.baseCharisma <= 0
      ) {
        alert('All base stats must be positive.');
        return;
      }
      if (editingTemplate) {
        const updated = await apiClient.put<MonsterTemplateRecord>(
          `/sonar/monster-templates/${editingTemplate.id}`,
          payload
        );
        if (templateOptionsLoaded) {
          setTemplateOptions((prev) =>
            prev.map((template) =>
              template.id === updated.id ? updated : template
            )
          );
        }
      } else {
        const created = await apiClient.post<MonsterTemplateRecord>(
          '/sonar/monster-templates',
          payload
        );
        if (templateOptionsLoaded) {
          setTemplateOptions((prev) => [created, ...prev]);
        }
      }
      await loadPagedData(true);
      closeTemplateModal();
    } catch (err) {
      console.error('Failed to save template', err);
      const message =
        err instanceof Error ? err.message : 'Failed to save template.';
      alert(message);
    }
  };

  const deleteTemplate = async (template: MonsterTemplateRecord) => {
    if (!window.confirm(`Delete template "${template.name}"?`)) return;
    try {
      await apiClient.delete(`/sonar/monster-templates/${template.id}`);
      if (templateOptionsLoaded) {
        setTemplateOptions((prev) =>
          prev.filter((entry) => entry.id !== template.id)
        );
      }
      await loadPagedData(true);
    } catch (err) {
      console.error('Failed to delete template', err);
      const message =
        err instanceof Error ? err.message : 'Failed to delete template.';
      alert(message);
    }
  };

  const toggleTemplateSelection = (templateId: string, checked: boolean) => {
    setSelectedTemplateIds((prev) => {
      const next = new Set(prev);
      if (checked) {
        next.add(templateId);
      } else {
        next.delete(templateId);
      }
      return next;
    });
  };

  const setTemplatesArchived = async (ids: string[], archived: boolean) => {
    if (ids.length === 0) return;
    try {
      await apiClient.post('/sonar/monster-templates/bulk-archive', {
        ids,
        archived,
      });
      if (templateOptionsLoaded) {
        const idSet = new Set(ids);
        setTemplateOptions((prev) =>
          prev.map((template) =>
            idSet.has(template.id) ? { ...template, archived } : template
          )
        );
      }
      setSelectedTemplateIds((prev) => {
        const next = new Set(prev);
        ids.forEach((id) => next.delete(id));
        return next;
      });
      await loadPagedData(true);
    } catch (err) {
      console.error('Failed to update monster template archive state', err);
      alert(
        `Failed to ${archived ? 'archive' : 'restore'} monster template(s).`
      );
    }
  };

  const handleBulkGenerateTemplates = async () => {
    const count = Number.parseInt(bulkTemplateCount, 10);
    if (!Number.isFinite(count) || count < 1 || count > 100) {
      alert('Bulk template count must be between 1 and 100.');
      return;
    }
    try {
      setBulkTemplateBusy(true);
      setBulkTemplateError(null);
      setBulkTemplateMessage(null);
      setBulkTemplateJob(null);
      const response = await apiClient.post<BulkMonsterTemplateStatus>(
        '/sonar/monster-templates/bulk-generate',
        {
          count,
          monsterType: bulkTemplateType,
          genreId: bulkTemplateGenreId,
          zoneKind: bulkTemplateZoneKind,
          yeetIt: bulkTemplateYeetIt,
        }
      );
      setBulkTemplateJob(response);
      if (response.status === 'completed') {
        setBulkTemplateBusy(false);
        const typeLabel = formatMonsterTemplateTypeLabel(
          response.monsterType ?? bulkTemplateType
        );
        setBulkTemplateMessage(
          `${response.yeetIt ? 'Created' : 'Prepared'} ${response.createdCount} ${typeLabel} ${
            response.yeetIt ? 'template' : 'draft'
          }${
            response.createdCount === 1 ? '' : 's'
          }${response.yeetIt ? ' live.' : '.'}`
        );
        setSelectedTemplateSuggestionJobId(response.jobId);
        setTemplateSuggestionJobsPage(1);
        await fetchTemplateSuggestionJobs();
        if (response.yeetIt) {
          setTemplateSuggestionDrafts([]);
        } else {
          await fetchTemplateSuggestionDrafts(response.jobId);
        }
      } else if (response.status === 'failed') {
        setBulkTemplateBusy(false);
        setBulkTemplateError(
          response.error || 'Bulk template generation failed.'
        );
      }
    } catch (err) {
      console.error('Failed to bulk generate monster templates', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to bulk generate monster templates.';
      setBulkTemplateError(message);
      setBulkTemplateBusy(false);
    }
  };

  const handleGenerateSingleTypedTemplate = async (
    monsterType: MonsterTemplateType
  ) => {
    try {
      setBulkTemplateBusy(true);
      setBulkTemplateError(null);
      setBulkTemplateMessage(null);
      setBulkTemplateJob(null);
      const response = await apiClient.post<BulkMonsterTemplateStatus>(
        '/sonar/monster-templates/bulk-generate',
        {
          count: 1,
          monsterType,
          genreId: bulkTemplateGenreId,
          zoneKind: bulkTemplateZoneKind,
          yeetIt: bulkTemplateYeetIt,
        }
      );
      setBulkTemplateJob(response);
      setBulkTemplateType(monsterType);
      if (response.status === 'completed') {
        setBulkTemplateBusy(false);
        setBulkTemplateMessage(
          `${response.yeetIt ? 'Created' : 'Prepared'} 1 ${formatMonsterTemplateTypeLabel(monsterType)} ${
            response.yeetIt ? 'template' : 'draft'
          }${response.yeetIt ? ' live.' : '.'}`
        );
        setSelectedTemplateSuggestionJobId(response.jobId);
        setTemplateSuggestionJobsPage(1);
        await fetchTemplateSuggestionJobs();
        if (response.yeetIt) {
          setTemplateSuggestionDrafts([]);
        } else {
          await fetchTemplateSuggestionDrafts(response.jobId);
        }
      } else if (response.status === 'failed') {
        setBulkTemplateBusy(false);
        setBulkTemplateError(
          response.error || 'Bulk template generation failed.'
        );
      }
    } catch (err) {
      console.error(`Failed to generate ${monsterType} template`, err);
      const message =
        err instanceof Error
          ? err.message
          : `Failed to generate ${monsterType} template.`;
      setBulkTemplateError(message);
      setBulkTemplateBusy(false);
    }
  };

  const handleConvertTemplateSuggestionDraft = async (draftId: string) => {
    try {
      setConvertingTemplateSuggestionDraftId(draftId);
      await apiClient.post<MonsterTemplateRecord>(
        `/sonar/monster-template-suggestion-drafts/${draftId}/convert`,
        {}
      );
      if (selectedTemplateSuggestionJobId) {
        await fetchTemplateSuggestionDrafts(selectedTemplateSuggestionJobId);
      }
      await loadPagedData(true);
    } catch (err) {
      console.error('Failed to convert monster template suggestion draft', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to convert monster template suggestion draft.';
      alert(message);
    } finally {
      setConvertingTemplateSuggestionDraftId(null);
    }
  };

  const handleDeleteTemplateSuggestionDraft = async (draftId: string) => {
    try {
      setDeletingTemplateSuggestionDraftId(draftId);
      await apiClient.delete(`/sonar/monster-template-suggestion-drafts/${draftId}`);
      if (selectedTemplateSuggestionJobId) {
        await fetchTemplateSuggestionDrafts(selectedTemplateSuggestionJobId);
      }
    } catch (err) {
      console.error('Failed to delete monster template suggestion draft', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to delete monster template suggestion draft.';
      alert(message);
    } finally {
      setDeletingTemplateSuggestionDraftId(null);
    }
  };

  const handleConvertAllTemplateSuggestionDrafts = async () => {
    if (
      !selectedTemplateSuggestionJobId ||
      unconvertedTemplateSuggestionDrafts.length === 0
    ) {
      return;
    }
    try {
      setConvertingAllTemplateSuggestionDrafts(true);
      for (const draft of unconvertedTemplateSuggestionDrafts) {
        await apiClient.post<MonsterTemplateRecord>(
          `/sonar/monster-template-suggestion-drafts/${draft.id}/convert`,
          {}
        );
      }
      await fetchTemplateSuggestionDrafts(selectedTemplateSuggestionJobId);
      await loadPagedData(true);
    } catch (err) {
      console.error(
        'Failed to bulk convert monster template suggestion drafts',
        err
      );
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to convert all monster template suggestion drafts.';
      alert(message);
    } finally {
      setConvertingAllTemplateSuggestionDrafts(false);
    }
  };

  const refreshAffinityJobStatus = useCallback(
    async (jobId: string) => {
      try {
        const status =
          await apiClient.get<MonsterTemplateAffinityRefreshStatus>(
            `/sonar/admin/monster-templates/refresh-affinities/${jobId}/status`
          );
        setAffinityRefreshJob(status);
        if (status.status === 'completed') {
          setAffinityRefreshBusy(false);
          setAffinityRefreshError(null);
          setAffinityRefreshMessage(
            `Refreshed affinities and zone kinds for ${status.updatedCount} template${
              status.updatedCount === 1 ? '' : 's'
            }.`
          );
          await loadPagedData(true);
        } else if (status.status === 'failed') {
          setAffinityRefreshBusy(false);
          setAffinityRefreshMessage(null);
          setAffinityRefreshError(
            status.error || 'Affinity refresh generation failed.'
          );
        }
      } catch (err) {
        console.error('Failed to refresh monster affinity job status', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to refresh affinity job status.';
        setAffinityRefreshError(message);
        setAffinityRefreshBusy(false);
      }
    },
    [apiClient, loadPagedData]
  );

  const handleRefreshTemplateAffinities = async () => {
    try {
      setAffinityRefreshBusy(true);
      setAffinityRefreshError(null);
      setAffinityRefreshMessage(null);
      setAffinityRefreshJob(null);
      const ids = Array.from(selectedTemplateIds);
      const response =
        await apiClient.post<MonsterTemplateAffinityRefreshStatus>(
          '/sonar/admin/monster-templates/refresh-affinities',
          ids.length > 0 ? { ids } : {}
        );
      setAffinityRefreshJob(response);
      if (response.status === 'completed') {
        setAffinityRefreshBusy(false);
        setAffinityRefreshMessage(
          `Refreshed affinities and zone kinds for ${response.updatedCount} template${
            response.updatedCount === 1 ? '' : 's'
          }.`
        );
        await loadPagedData(true);
      } else if (response.status === 'failed') {
        setAffinityRefreshBusy(false);
        setAffinityRefreshError(
          response.error || 'Affinity refresh generation failed.'
        );
      }
    } catch (err) {
      console.error('Failed to refresh monster template affinities', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to refresh monster template affinities.';
      setAffinityRefreshError(message);
      setAffinityRefreshBusy(false);
    }
  };

  const refreshProgressionResetJobStatus = useCallback(
    async (jobId: string) => {
      try {
        const status =
          await apiClient.get<MonsterTemplateProgressionResetStatus>(
            `/sonar/admin/monster-templates/reset-progressions/${jobId}/status`
          );
        setProgressionResetJob(status);
        if (status.status === 'completed') {
          setProgressionResetBusy(false);
          setProgressionResetError(null);
          setProgressionResetMessage(
            `Reset progressions for ${status.updatedCount} template${
              status.updatedCount === 1 ? '' : 's'
            }.`
          );
          await loadPagedData(true);
        } else if (status.status === 'failed') {
          setProgressionResetBusy(false);
          setProgressionResetMessage(null);
          setProgressionResetError(status.error || 'Progression reset failed.');
        }
      } catch (err) {
        console.error(
          'Failed to refresh monster progression reset job status',
          err
        );
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to refresh progression reset job status.';
        setProgressionResetError(message);
        setProgressionResetBusy(false);
      }
    },
    [apiClient, loadPagedData]
  );

  const handleResetTemplateProgressions = async () => {
    try {
      setProgressionResetBusy(true);
      setProgressionResetError(null);
      setProgressionResetMessage(null);
      setProgressionResetJob(null);
      const ids = Array.from(selectedTemplateIds);
      const response =
        await apiClient.post<MonsterTemplateProgressionResetStatus>(
          '/sonar/admin/monster-templates/reset-progressions',
          ids.length > 0 ? { ids } : {}
        );
      setProgressionResetJob(response);
      if (response.status === 'completed') {
        setProgressionResetBusy(false);
        setProgressionResetMessage(
          `Reset progressions for ${response.updatedCount} template${
            response.updatedCount === 1 ? '' : 's'
          }.`
        );
        await loadPagedData(true);
      } else if (response.status === 'failed') {
        setProgressionResetBusy(false);
        setProgressionResetError(response.error || 'Progression reset failed.');
      }
    } catch (err) {
      console.error('Failed to reset monster template progressions', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to reset monster template progressions.';
      setProgressionResetError(message);
      setProgressionResetBusy(false);
    }
  };

  const ensureTemplateOptionsLoaded = useCallback(async () => {
    if (templateOptionsLoaded && templateOptions.length > 0) {
      return templateOptions;
    }
    return loadTemplateOptions();
  }, [loadTemplateOptions, templateOptions, templateOptionsLoaded]);

  const ensureMonsterOptionsLoaded = useCallback(async () => {
    if (monsterOptionsLoaded && monsterOptions.length > 0) {
      return monsterOptions;
    }
    return loadMonsterOptions();
  }, [loadMonsterOptions, monsterOptions, monsterOptionsLoaded]);

  const openCreateMonster = async () => {
    try {
      const availableTemplates = await ensureTemplateOptionsLoaded();
      setEditingMonster(null);
      setMonsterForm({
        ...emptyMonsterForm(),
        zoneId: zones[0]?.id ?? '',
        templateId: availableTemplates[0]?.id ?? '',
        dominantHandInventoryItemId: dominantHandItems[0]
          ? String(dominantHandItems[0].id)
          : '',
      });
      setShowMonsterModal(true);
    } catch (err) {
      console.error('Failed to load template options', err);
      alert('Failed to load monster template options.');
    }
  };

  const openEditMonster = async (monster: MonsterRecord) => {
    try {
      await ensureTemplateOptionsLoaded();
      setEditingMonster(monster);
      setMonsterForm(monsterFormFromRecord(monster));
      setShowMonsterModal(true);
    } catch (err) {
      console.error('Failed to load template options', err);
      alert('Failed to load monster template options.');
    }
  };

  const closeMonsterModal = () => {
    setShowMonsterModal(false);
    setEditingMonster(null);
    setMonsterForm(emptyMonsterForm());
  };

  const openCreateEncounter = async () => {
    try {
      const availableMonsters = await ensureMonsterOptionsLoaded();
      const defaultZoneId = zones[0]?.id ?? '';
      const defaultMonsters = availableMonsters
        .filter((monster) =>
          defaultZoneId ? monster.zoneId === defaultZoneId : true
        )
        .slice(0, 3)
        .map((monster) => monster.id);
      setEditingEncounter(null);
      setEncounterForm({
        ...emptyMonsterEncounterForm(),
        zoneId: defaultZoneId,
        monsterIds: defaultMonsters,
      });
      setShowEncounterModal(true);
    } catch (err) {
      console.error('Failed to load monster options', err);
      alert('Failed to load monsters for encounter editing.');
    }
  };

  const openEditEncounter = async (encounter: MonsterEncounterRecord) => {
    try {
      await ensureMonsterOptionsLoaded();
      setEditingEncounter(encounter);
      setEncounterForm(monsterEncounterFormFromRecord(encounter));
      setShowEncounterModal(true);
    } catch (err) {
      console.error('Failed to load monster options', err);
      alert('Failed to load monsters for encounter editing.');
    }
  };

  const closeEncounterModal = () => {
    setShowEncounterModal(false);
    setEditingEncounter(null);
    setEncounterForm(emptyMonsterEncounterForm());
  };

  const saveEncounter = async () => {
    try {
      const payload = monsterEncounterPayloadFromForm(encounterForm);
      if (!payload.zoneId) {
        alert('Zone is required.');
        return;
      }
      if (payload.monsterIds.length < 1 || payload.monsterIds.length > 9) {
        alert('Encounter must include between 1 and 9 monsters.');
        return;
      }

      if (editingEncounter) {
        await apiClient.put<MonsterEncounterRecord>(
          `/sonar/monster-encounters/${editingEncounter.id}`,
          payload
        );
      } else {
        await apiClient.post<MonsterEncounterRecord>(
          '/sonar/monster-encounters',
          payload
        );
      }
      await loadPagedData(true);
      closeEncounterModal();
    } catch (err) {
      console.error('Failed to save monster encounter', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to save monster encounter.';
      alert(message);
    }
  };

  const deleteEncounter = async (encounter: MonsterEncounterRecord) => {
    if (bulkDeletingEncounters) return;
    if (!window.confirm(`Delete encounter "${encounter.name}"?`)) return;
    try {
      await apiClient.delete(`/sonar/monster-encounters/${encounter.id}`);
      setSelectedEncounterIds((prev) => {
        if (!prev.has(encounter.id)) return prev;
        const next = new Set(prev);
        next.delete(encounter.id);
        return next;
      });
      await loadPagedData(true);
    } catch (err) {
      console.error('Failed to delete monster encounter', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to delete monster encounter.';
      alert(message);
    }
  };

  const toggleEncounterSelection = (encounterId: string) => {
    setSelectedEncounterIds((prev) => {
      const next = new Set(prev);
      if (next.has(encounterId)) {
        next.delete(encounterId);
      } else {
        next.add(encounterId);
      }
      return next;
    });
  };

  const toggleSelectVisibleEncounters = () => {
    if (filteredEncounters.length === 0) return;
    setSelectedEncounterIds((prev) => {
      const next = new Set(prev);
      if (allFilteredEncountersSelected) {
        filteredEncounters.forEach((encounter) => next.delete(encounter.id));
      } else {
        filteredEncounters.forEach((encounter) => next.add(encounter.id));
      }
      return next;
    });
  };

  const clearEncounterSelection = () => {
    setSelectedEncounterIds(new Set());
  };

  const handleBulkDeleteEncounters = async () => {
    if (bulkDeletingEncounters || selectedEncounterIds.size === 0) return;

    const selectedIds = Array.from(selectedEncounterIds);
    const selectedNames = encounters
      .filter((encounter) => selectedEncounterIds.has(encounter.id))
      .map((encounter) => encounter.name);
    const preview = selectedNames.slice(0, 5).join(', ');
    const moreCount = Math.max(0, selectedNames.length - 5);
    const confirmMessage =
      selectedIds.length === 1
        ? `Delete 1 selected monster encounter (${preview})? This cannot be undone.`
        : `Delete ${selectedIds.length} selected monster encounters${
            preview
              ? ` (${preview}${moreCount > 0 ? ` +${moreCount} more` : ''})`
              : ''
          }? This cannot be undone.`;
    if (!window.confirm(confirmMessage)) return;

    setBulkDeletingEncounters(true);
    try {
      const results = await Promise.allSettled(
        selectedIds.map((encounterId) =>
          apiClient.delete(`/sonar/monster-encounters/${encounterId}`)
        )
      );
      const deletedIds = new Set<string>();
      const failedIds: string[] = [];
      results.forEach((result, index) => {
        const encounterId = selectedIds[index];
        if (result.status === 'fulfilled') {
          deletedIds.add(encounterId);
        } else {
          console.error(
            `Failed to delete encounter ${encounterId}`,
            result.reason
          );
          failedIds.push(encounterId);
        }
      });

      if (deletedIds.size > 0) {
        setSelectedEncounterIds((prev) => {
          const next = new Set(prev);
          deletedIds.forEach((encounterId) => next.delete(encounterId));
          return next;
        });
        if (editingEncounter && deletedIds.has(editingEncounter.id)) {
          closeEncounterModal();
        }
        await loadPagedData(true);
      }

      if (failedIds.length > 0) {
        alert(
          `Deleted ${deletedIds.size} encounter${deletedIds.size === 1 ? '' : 's'}, but failed to delete ${
            failedIds.length
          }. Check console for details.`
        );
      }
    } catch (err) {
      console.error('Failed to bulk delete encounters', err);
      alert('Failed to delete selected monster encounters.');
    } finally {
      setBulkDeletingEncounters(false);
    }
  };

  const saveMonster = async () => {
    try {
      const payload = monsterPayloadFromForm(monsterForm);
      if (!payload.zoneId || !payload.templateId) {
        alert('Zone and template are required.');
        return;
      }
      if (
        !payload.dominantHandInventoryItemId ||
        payload.dominantHandInventoryItemId <= 0
      ) {
        alert('A dominant hand weapon is required.');
        return;
      }
      if (payload.level <= 0) {
        alert('Level must be positive.');
        return;
      }
      if (dominantIsTwoHanded && payload.offHandInventoryItemId) {
        alert(
          'Two-handed dominant weapons cannot be combined with an off-hand item.'
        );
        return;
      }

      if (editingMonster) {
        await apiClient.put<MonsterRecord>(
          `/sonar/monsters/${editingMonster.id}`,
          payload
        );
      } else {
        await apiClient.post<MonsterRecord>('/sonar/monsters', payload);
      }
      if (monsterOptionsLoaded) {
        await loadMonsterOptions();
      }
      await loadPagedData(true);
      closeMonsterModal();
    } catch (err) {
      console.error('Failed to save monster', err);
      const message =
        err instanceof Error ? err.message : 'Failed to save monster.';
      alert(message);
    }
  };

  const deleteMonster = async (monster: MonsterRecord) => {
    if (!window.confirm(`Delete monster "${monster.name}"?`)) return;
    try {
      await apiClient.delete(`/sonar/monsters/${monster.id}`);
      setSelectedMonsterIds((prev) => {
        if (!prev.has(monster.id)) return prev;
        const next = new Set(prev);
        next.delete(monster.id);
        return next;
      });
      if (monsterOptionsLoaded) {
        setMonsterOptions((prev) =>
          prev.filter((record) => record.id !== monster.id)
        );
      }
      await loadPagedData(true);
    } catch (err) {
      console.error('Failed to delete monster', err);
      alert('Failed to delete monster.');
    }
  };

  const toggleMonsterSelection = (monsterId: string) => {
    setSelectedMonsterIds((prev) => {
      const next = new Set(prev);
      if (next.has(monsterId)) {
        next.delete(monsterId);
      } else {
        next.add(monsterId);
      }
      return next;
    });
  };

  const selectAllVisibleMonsters = () => {
    if (filteredMonsters.length === 0) return;
    setSelectedMonsterIds((prev) => {
      const next = new Set(prev);
      filteredMonsters.forEach((monster) => next.add(monster.id));
      return next;
    });
  };

  const clearMonsterSelection = () => {
    setSelectedMonsterIds(new Set());
  };

  const handleBulkDeleteMonsters = async () => {
    if (bulkDeletingMonsters || selectedMonsterIds.size === 0) return;

    const selectedIds = Array.from(selectedMonsterIds);
    const selectedNames = records
      .filter((monster) => selectedMonsterIds.has(monster.id))
      .map((monster) => monster.name);
    const preview = selectedNames.slice(0, 5).join(', ');
    const moreCount = Math.max(0, selectedNames.length - 5);
    const confirmMessage =
      selectedIds.length === 1
        ? `Delete 1 selected monster (${preview})? This cannot be undone.`
        : `Delete ${selectedIds.length} selected monsters${
            preview
              ? ` (${preview}${moreCount > 0 ? ` +${moreCount} more` : ''})`
              : ''
          }? This cannot be undone.`;
    if (!window.confirm(confirmMessage)) return;

    setBulkDeletingMonsters(true);
    try {
      const results = await Promise.allSettled(
        selectedIds.map((monsterId) =>
          apiClient.delete(`/sonar/monsters/${monsterId}`)
        )
      );
      const deletedIds = new Set<string>();
      const failedIds: string[] = [];
      results.forEach((result, index) => {
        const monsterId = selectedIds[index];
        if (result.status === 'fulfilled') {
          deletedIds.add(monsterId);
        } else {
          console.error(`Failed to delete monster ${monsterId}`, result.reason);
          failedIds.push(monsterId);
        }
      });

      if (deletedIds.size > 0) {
        setSelectedMonsterIds((prev) => {
          const next = new Set(prev);
          deletedIds.forEach((monsterId) => next.delete(monsterId));
          return next;
        });
        if (monsterOptionsLoaded) {
          setMonsterOptions((prev) =>
            prev.filter((monster) => !deletedIds.has(monster.id))
          );
        }
        if (editingMonster && deletedIds.has(editingMonster.id)) {
          closeMonsterModal();
        }
        await loadPagedData(true);
      }

      if (failedIds.length > 0) {
        alert(
          `Deleted ${deletedIds.size} monster${deletedIds.size === 1 ? '' : 's'}, but failed to delete ${
            failedIds.length
          }. Check console for details.`
        );
      }
    } catch (err) {
      console.error('Failed to bulk delete monsters', err);
      alert('Failed to delete selected monsters.');
    } finally {
      setBulkDeletingMonsters(false);
    }
  };

  const handleGenerateImage = async (monster: MonsterRecord) => {
    try {
      setGeneratingMonsterId(monster.id);
      const updated = await apiClient.post<MonsterRecord>(
        `/sonar/monsters/${monster.id}/generate-image`,
        {}
      );
      setRecords((prev) =>
        prev.map((record) => (record.id === monster.id ? updated : record))
      );
    } catch (err) {
      console.error('Failed to queue monster image generation', err);
      alert('Failed to queue monster image generation.');
    } finally {
      setGeneratingMonsterId(null);
    }
  };

  const handleGenerateTemplateImage = async (
    template: MonsterTemplateRecord
  ) => {
    try {
      setGeneratingTemplateId(template.id);
      const updated = await apiClient.post<MonsterTemplateRecord>(
        `/sonar/monster-templates/${template.id}/generate-image`,
        {}
      );
      setTemplates((prev) =>
        prev.map((record) => (record.id === template.id ? updated : record))
      );
    } catch (err) {
      console.error('Failed to queue monster template image generation', err);
      alert('Failed to queue monster template image generation.');
    } finally {
      setGeneratingTemplateId(null);
    }
  };

  const updateMonsterItemReward = (
    index: number,
    nextReward: Partial<MonsterFormItem>
  ) => {
    setMonsterForm((prev) => {
      const next = [...prev.itemRewards];
      next[index] = { ...next[index], ...nextReward };
      return { ...prev, itemRewards: next };
    });
  };

  const updateEncounterItemReward = (
    index: number,
    nextReward: Partial<MonsterFormItem>
  ) => {
    setEncounterForm((prev) => {
      const next = [...prev.itemRewards];
      next[index] = { ...next[index], ...nextReward };
      return { ...prev, itemRewards: next };
    });
  };

  const updateTemplateSpellIds = (selected: HTMLSelectElement) => {
    const spellProgressionIds = Array.from(selected.selectedOptions).map(
      (option) => option.value
    );
    setTemplateForm((prev) => ({ ...prev, spellProgressionIds }));
  };

  const updateTemplateTechniqueIds = (selected: HTMLSelectElement) => {
    const techniqueProgressionIds = Array.from(selected.selectedOptions).map(
      (option) => option.value
    );
    setTemplateForm((prev) => ({ ...prev, techniqueProgressionIds }));
  };

  const updateDominantHandSelection = (value: string) => {
    setMonsterForm((prev) => {
      const selectedID = parseIntSafe(value, 0);
      const selected =
        selectedID > 0 ? dominantHandItemById.get(selectedID) : undefined;
      const handedness = (selected?.handedness ?? '').trim().toLowerCase();
      return {
        ...prev,
        dominantHandInventoryItemId: value,
        offHandInventoryItemId:
          handedness === 'two_handed' ? '' : prev.offHandInventoryItemId,
      };
    });
  };

  const setMonsterLocation = useCallback(
    (latitude: number, longitude: number) => {
      setMonsterForm((prev) => ({
        ...prev,
        latitude: latitude.toFixed(6),
        longitude: longitude.toFixed(6),
      }));
    },
    []
  );

  const setEncounterLocation = useCallback(
    (latitude: number, longitude: number) => {
      setEncounterForm((prev) => ({
        ...prev,
        latitude: latitude.toFixed(6),
        longitude: longitude.toFixed(6),
      }));
    },
    []
  );

  const handleUseCurrentLocation = useCallback(() => {
    if (!navigator.geolocation) {
      alert('Geolocation is not supported in this browser.');
      return;
    }
    setGeoLoading(true);
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setGeoLoading(false);
        setMonsterLocation(position.coords.latitude, position.coords.longitude);
      },
      (error) => {
        setGeoLoading(false);
        alert(`Unable to get current location: ${error.message}`);
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 30000 }
    );
  }, [setMonsterLocation]);

  const handleUseCurrentEncounterLocation = useCallback(() => {
    if (!navigator.geolocation) {
      alert('Geolocation is not supported in this browser.');
      return;
    }
    setGeoLoading(true);
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setGeoLoading(false);
        setEncounterLocation(
          position.coords.latitude,
          position.coords.longitude
        );
      },
      (error) => {
        setGeoLoading(false);
        alert(`Unable to get current location: ${error.message}`);
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 30000 }
    );
  }, [setEncounterLocation]);

  useEffect(() => {
    if (!showMonsterModal) return;
    if (!mapContainerRef.current) return;
    if (!mapboxgl.accessToken) return;
    if (mapRef.current) return;

    const parsedLat = Number.parseFloat(formLatitudeRef.current);
    const parsedLng = Number.parseFloat(formLongitudeRef.current);
    const selectedZone = zones.find((zone) => zone.id === monsterForm.zoneId);
    const zoneLat = selectedZone
      ? Number.parseFloat(String(selectedZone.latitude ?? ''))
      : Number.NaN;
    const zoneLng = selectedZone
      ? Number.parseFloat(String(selectedZone.longitude ?? ''))
      : Number.NaN;

    const center: [number, number] =
      Number.isFinite(parsedLat) && Number.isFinite(parsedLng)
        ? [parsedLng, parsedLat]
        : Number.isFinite(zoneLat) && Number.isFinite(zoneLng)
          ? [zoneLng, zoneLat]
          : [-73.98513, 40.7589];

    const map = new mapboxgl.Map({
      container: mapContainerRef.current,
      style: 'mapbox://styles/mapbox/streets-v12',
      center,
      zoom: 13,
    });

    map.on('click', (event) => {
      setMonsterLocation(event.lngLat.lat, event.lngLat.lng);
    });

    mapRef.current = map;

    return () => {
      markerRef.current?.remove();
      markerRef.current = null;
      map.remove();
      mapRef.current = null;
    };
  }, [monsterForm.zoneId, setMonsterLocation, showMonsterModal, zones]);

  useEffect(() => {
    if (!showMonsterModal) return;
    if (!mapRef.current) return;

    const lat = Number.parseFloat(monsterForm.latitude);
    const lng = Number.parseFloat(monsterForm.longitude);
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) {
      markerRef.current?.remove();
      markerRef.current = null;
      return;
    }

    if (!markerRef.current) {
      markerRef.current = new mapboxgl.Marker({ color: '#dc2626' })
        .setLngLat([lng, lat])
        .addTo(mapRef.current);
    } else {
      markerRef.current.setLngLat([lng, lat]);
    }

    mapRef.current.easeTo({ center: [lng, lat], duration: 350 });
  }, [monsterForm.latitude, monsterForm.longitude, showMonsterModal]);

  useEffect(() => {
    if (!showEncounterModal) return;
    if (!encounterMapContainerRef.current) return;
    if (!mapboxgl.accessToken) return;
    if (encounterMapRef.current) return;

    const parsedLat = Number.parseFloat(encounterLatitudeRef.current);
    const parsedLng = Number.parseFloat(encounterLongitudeRef.current);
    const selectedZone = zones.find((zone) => zone.id === encounterForm.zoneId);
    const zoneLat = selectedZone
      ? Number.parseFloat(String(selectedZone.latitude ?? ''))
      : Number.NaN;
    const zoneLng = selectedZone
      ? Number.parseFloat(String(selectedZone.longitude ?? ''))
      : Number.NaN;

    const center: [number, number] =
      Number.isFinite(parsedLat) && Number.isFinite(parsedLng)
        ? [parsedLng, parsedLat]
        : Number.isFinite(zoneLat) && Number.isFinite(zoneLng)
          ? [zoneLng, zoneLat]
          : [-73.98513, 40.7589];

    const map = new mapboxgl.Map({
      container: encounterMapContainerRef.current,
      style: 'mapbox://styles/mapbox/streets-v12',
      center,
      zoom: 13,
    });

    map.on('click', (event) => {
      setEncounterLocation(event.lngLat.lat, event.lngLat.lng);
    });

    encounterMapRef.current = map;

    return () => {
      encounterMarkerRef.current?.remove();
      encounterMarkerRef.current = null;
      map.remove();
      encounterMapRef.current = null;
    };
  }, [encounterForm.zoneId, setEncounterLocation, showEncounterModal, zones]);

  useEffect(() => {
    if (!showEncounterModal) return;
    if (!encounterMapRef.current) return;

    const lat = Number.parseFloat(encounterForm.latitude);
    const lng = Number.parseFloat(encounterForm.longitude);
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) {
      encounterMarkerRef.current?.remove();
      encounterMarkerRef.current = null;
      return;
    }

    if (!encounterMarkerRef.current) {
      encounterMarkerRef.current = new mapboxgl.Marker({ color: '#b91c1c' })
        .setLngLat([lng, lat])
        .addTo(encounterMapRef.current);
    } else {
      encounterMarkerRef.current.setLngLat([lng, lat]);
    }

    encounterMapRef.current.easeTo({ center: [lng, lat], duration: 350 });
  }, [encounterForm.latitude, encounterForm.longitude, showEncounterModal]);

  const openMonsterImagePreview = (monster: MonsterRecord) => {
    const url = monster.imageUrl || monster.thumbnailUrl;
    if (!url) return;
    setImagePreview({
      url,
      alt: `${monster.name || 'Monster'} image`,
    });
  };

  const openTemplateImagePreview = (template: MonsterTemplateRecord) => {
    const url = template.imageUrl || template.thumbnailUrl;
    if (!url) return;
    setImagePreview({
      url,
      alt: `${template.name || 'Monster template'} image`,
    });
  };

  const closeImagePreview = () => {
    setImagePreview(null);
  };

  const setEncounterIconState = useCallback(
    (
      encounterType: MonsterEncounterType,
      updates:
        | Partial<EncounterIconState>
        | ((previous: EncounterIconState) => Partial<EncounterIconState>)
    ) => {
      setEncounterIconStates((prev) => {
        const current = prev[encounterType];
        const nextUpdates =
          typeof updates === 'function' ? updates(current) : updates;
        return {
          ...prev,
          [encounterType]: {
            ...current,
            ...nextUpdates,
          },
        };
      });
    },
    []
  );

  const refreshEncounterIconStatus = useCallback(
    async (encounterType: MonsterEncounterType, showMessage = false) => {
      try {
        setEncounterIconState(encounterType, {
          statusLoading: true,
          error: null,
        });
        const response = await apiClient.get<StaticThumbnailResponse>(
          `/sonar/admin/thumbnails/monster-undiscovered/${encounterType}/status`
        );
        const url = (response?.thumbnailUrl || '').trim();
        setEncounterIconState(encounterType, {
          url: url || encounterIconDefaults[encounterType].url,
          status: (response?.status || 'unknown').trim() || 'unknown',
          exists: Boolean(response?.exists),
          requestedAt: response?.requestedAt ? response.requestedAt : null,
          lastModified: response?.lastModified ? response.lastModified : null,
          previewNonce: Date.now(),
          message: showMessage
            ? `${encounterIconDefaults[encounterType].label} icon status refreshed.`
            : null,
        });
      } catch (err) {
        console.error(
          `Failed to load ${encounterType} undiscovered monster icon status`,
          err
        );
        const message =
          err instanceof Error
            ? err.message
            : `Failed to load ${encounterType} undiscovered monster icon status.`;
        setEncounterIconState(encounterType, { error: message });
      } finally {
        setEncounterIconState(encounterType, { statusLoading: false });
      }
    },
    [apiClient, setEncounterIconState]
  );

  const handleGenerateEncounterIcon = useCallback(
    async (encounterType: MonsterEncounterType) => {
      const prompt = encounterIconStates[encounterType].prompt.trim();
      if (!prompt) {
        setEncounterIconState(encounterType, { error: 'Prompt is required.' });
        return;
      }
      try {
        setEncounterIconState(encounterType, {
          busy: true,
          error: null,
          message: null,
        });
        await apiClient.post<StaticThumbnailResponse>(
          `/sonar/admin/thumbnails/monster-undiscovered/${encounterType}`,
          { prompt }
        );
        setEncounterIconState(encounterType, {
          message: `${encounterIconDefaults[encounterType].label} icon queued for generation.`,
        });
        await refreshEncounterIconStatus(encounterType);
      } catch (err) {
        console.error(
          `Failed to generate ${encounterType} undiscovered monster icon`,
          err
        );
        const message =
          err instanceof Error
            ? err.message
            : `Failed to generate ${encounterType} undiscovered monster icon.`;
        setEncounterIconState(encounterType, { error: message });
      } finally {
        setEncounterIconState(encounterType, { busy: false });
      }
    },
    [
      apiClient,
      encounterIconStates,
      refreshEncounterIconStatus,
      setEncounterIconState,
    ]
  );

  const handleDeleteEncounterIcon = useCallback(
    async (encounterType: MonsterEncounterType) => {
      try {
        setEncounterIconState(encounterType, {
          busy: true,
          error: null,
          message: null,
        });
        await apiClient.delete<StaticThumbnailResponse>(
          `/sonar/admin/thumbnails/monster-undiscovered/${encounterType}`
        );
        setEncounterIconState(encounterType, {
          message: `${encounterIconDefaults[encounterType].label} icon deleted.`,
        });
        await refreshEncounterIconStatus(encounterType);
      } catch (err) {
        console.error(
          `Failed to delete ${encounterType} undiscovered monster icon`,
          err
        );
        const message =
          err instanceof Error
            ? err.message
            : `Failed to delete ${encounterType} undiscovered monster icon.`;
        setEncounterIconState(encounterType, { error: message });
      } finally {
        setEncounterIconState(encounterType, { busy: false });
      }
    },
    [apiClient, refreshEncounterIconStatus, setEncounterIconState]
  );

  useEffect(() => {
    void Promise.all(
      (Object.keys(encounterIconDefaults) as MonsterEncounterType[]).map(
        (encounterType) => refreshEncounterIconStatus(encounterType)
      )
    );
  }, [refreshEncounterIconStatus]);

  useEffect(() => {
    const hasPendingIcon = (
      Object.keys(encounterIconDefaults) as MonsterEncounterType[]
    ).some((encounterType) =>
      ['queued', 'in_progress'].includes(
        encounterIconStates[encounterType].status
      )
    );
    if (!hasPendingIcon) {
      return;
    }
    const interval = window.setInterval(() => {
      void Promise.all(
        (Object.keys(encounterIconDefaults) as MonsterEncounterType[]).map(
          (encounterType) => refreshEncounterIconStatus(encounterType)
        )
      );
    }, 4000);
    return () => window.clearInterval(interval);
  }, [encounterIconStates, refreshEncounterIconStatus]);

  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => {
    if (!bulkTemplateJob?.jobId) {
      return;
    }
    if (
      bulkTemplateJob.status !== 'queued' &&
      bulkTemplateJob.status !== 'in_progress'
    ) {
      return;
    }
    const interval = window.setInterval(() => {
      void refreshBulkTemplateJobStatus(bulkTemplateJob.jobId);
    }, 3000);
    return () => window.clearInterval(interval);
  }, [bulkTemplateJob, refreshBulkTemplateJobStatus]);

  useEffect(() => {
    if (!affinityRefreshJob?.jobId) {
      return;
    }
    if (
      affinityRefreshJob.status !== 'queued' &&
      affinityRefreshJob.status !== 'in_progress'
    ) {
      return;
    }
    const interval = window.setInterval(() => {
      void refreshAffinityJobStatus(affinityRefreshJob.jobId);
    }, 3000);
    return () => window.clearInterval(interval);
  }, [affinityRefreshJob, refreshAffinityJobStatus]);

  useEffect(() => {
    if (!progressionResetJob?.jobId) {
      return;
    }
    if (
      progressionResetJob.status !== 'queued' &&
      progressionResetJob.status !== 'in_progress'
    ) {
      return;
    }
    const interval = window.setInterval(() => {
      void refreshProgressionResetJobStatus(progressionResetJob.jobId);
    }, 3000);
    return () => window.clearInterval(interval);
  }, [progressionResetJob, refreshProgressionResetJobStatus]);

  return (
    <div className="p-6 bg-gray-100 min-h-screen">
      <div className="max-w-7xl mx-auto space-y-6">
        <ContentDashboard
          title="Monster Dashboard"
          subtitle="Aggregate monster template coverage for the current query, zone, genre, and template filters."
          status={
            query.trim() ||
            zoneQuery.trim() ||
            genreFilter !== 'all' ||
            templateTypeFilter !== 'all' ||
            templateTab === 'archived'
              ? 'Reflects current filters and template view'
              : 'All monster templates in the active view'
          }
          loading={dashboardLoading}
          error={dashboardError}
          metrics={dashboardMetrics}
          sections={dashboardSections}
        />

        <div className="qa-card">
          <div className="flex items-center justify-between gap-3">
            <div>
              <h1 className="qa-card-title">Monsters</h1>
              <p className="text-sm text-gray-600">
                Configure monster templates (base stats + abilities), then place
                monsters with level and weapon loadouts.
              </p>
            </div>
            <div className="flex gap-2">
              <select
                value={bulkTemplateType}
                onChange={(event) =>
                  setBulkTemplateType(event.target.value as MonsterTemplateType)
                }
                className="rounded-md border border-gray-300 px-2 py-2 text-sm"
                aria-label="Bulk template type"
              >
                {monsterTemplateTypeOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
              <select
                value={bulkTemplateGenreId}
                onChange={(event) => setBulkTemplateGenreId(event.target.value)}
                className="min-w-[180px] rounded-md border border-gray-300 px-2 py-2 text-sm"
                aria-label="Bulk template genre"
              >
                {genres.length === 0 ? (
                  <option value="">Fantasy</option>
                ) : (
                  genres.map((genre) => (
                    <option key={genre.id} value={genre.id}>
                      {formatGenreLabel(genre)}
                      {genre.active === false ? ' (inactive)' : ''}
                    </option>
                  ))
                )}
              </select>
              <select
                value={bulkTemplateZoneKind}
                onChange={(event) =>
                  setBulkTemplateZoneKind(event.target.value)
                }
                className="min-w-[180px] rounded-md border border-gray-300 px-2 py-2 text-sm"
                aria-label="Bulk template zone kind"
              >
                <option value="">Any zone kind</option>
                {zoneKinds.map((zoneKind) => (
                  <option key={zoneKind.id} value={zoneKind.slug}>
                    {zoneKind.name}
                  </option>
                ))}
              </select>
              <input
                type="number"
                min={1}
                max={100}
                value={bulkTemplateCount}
                onChange={(event) => setBulkTemplateCount(event.target.value)}
                className="w-24 rounded-md border border-gray-300 px-2 py-2 text-sm"
                aria-label="Bulk template count"
              />
              <label className="inline-flex items-center gap-2 rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-700">
                <input
                  type="checkbox"
                  checked={bulkTemplateYeetIt}
                  onChange={(event) =>
                    setBulkTemplateYeetIt(event.target.checked)
                  }
                />
                Yeet it
              </label>
              <button
                className="qa-btn qa-btn-secondary"
                onClick={handleBulkGenerateTemplates}
                disabled={bulkTemplateBusy}
              >
                {bulkTemplateBusy
                  ? 'Generating...'
                  : bulkTemplateYeetIt
                    ? 'Generate Live Templates'
                    : 'Generate Drafts'}
              </button>
              <button
                className="qa-btn qa-btn-secondary"
                onClick={() => handleGenerateSingleTypedTemplate('boss')}
                disabled={bulkTemplateBusy}
              >
                {bulkTemplateYeetIt ? 'Generate Live Boss' : 'Generate Boss Draft'}
              </button>
              <button
                className="qa-btn qa-btn-secondary"
                onClick={() => handleGenerateSingleTypedTemplate('raid')}
                disabled={bulkTemplateBusy}
              >
                {bulkTemplateYeetIt ? 'Generate Live Raid' : 'Generate Raid Draft'}
              </button>
              <button
                className="qa-btn qa-btn-secondary"
                onClick={openCreateTemplate}
              >
                Create Template
              </button>
              <button
                className="qa-btn qa-btn-secondary"
                onClick={() => void openCreateEncounter()}
              >
                Create Encounter
              </button>
              <button
                className="qa-btn qa-btn-primary"
                onClick={() => void openCreateMonster()}
              >
                Create Monster
              </button>
            </div>
          </div>
          {bulkTemplateJob && (
            <div className="mt-3 flex flex-wrap items-center gap-3 text-sm text-gray-700">
              <span
                className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold text-white ${staticStatusClassName(
                  bulkTemplateJob.status
                )}`}
              >
                {formatBulkTemplateStatus(bulkTemplateJob.status)}
              </span>
              <span>
                {bulkTemplateJob.yeetIt ? 'Templates' : 'Drafts'}:{' '}
                {bulkTemplateJob.createdCount}/
                {bulkTemplateJob.totalCount}
              </span>
              <span>
                Type:{' '}
                {formatMonsterTemplateTypeLabel(
                  bulkTemplateJob.monsterType ?? bulkTemplateType
                )}
              </span>
              <span>
                Genre:{' '}
                {formatGenreLabel(
                  genres.find(
                    (genre) =>
                      genre.id ===
                      (bulkTemplateJob.genreId ?? bulkTemplateGenreId)
                  )
                )}
              </span>
              {bulkTemplateJob.zoneKind ? (
                <span>
                  Zone Kind:{' '}
                  {zoneKindLabel(bulkTemplateJob.zoneKind, zoneKindBySlug)}
                </span>
              ) : null}
              <span>{bulkTemplateJob.yeetIt ? 'Yeet mode' : 'Draft mode'}</span>
              <span>Job: {bulkTemplateJob.jobId}</span>
              <span>Updated: {formatDate(bulkTemplateJob.updatedAt)}</span>
            </div>
          )}
          {bulkTemplateMessage && (
            <p className="mt-2 text-sm text-emerald-700">
              {bulkTemplateMessage}
            </p>
          )}
          {bulkTemplateError && (
            <p className="mt-2 text-sm text-red-700">{bulkTemplateError}</p>
          )}
        </div>

        <div className="qa-card">
          <div className="mb-4 flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
            <div>
              <h2 className="text-lg font-semibold">Template Drafts</h2>
              <p className="text-sm text-gray-600">
                Review generated monster template drafts, approve the keepers,
                and only materialize real templates once they look worth
                keeping. Yeeted jobs skip this review step and go live
                immediately.
              </p>
            </div>
            <div className="rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-600">
              Recent jobs: {templateSuggestionJobs.length}
            </div>
          </div>

          <div className="grid grid-cols-1 gap-4 xl:grid-cols-[0.9fr_1.1fr]">
            <div className="rounded-lg border border-gray-200 bg-white p-4">
              <div className="mb-3 flex items-center justify-between gap-3">
                <div>
                  <div className="text-sm font-semibold text-slate-800">
                    Recent Draft Jobs
                  </div>
                  {templateSuggestionJobs.length > 0 && (
                    <div className="mt-1 text-xs text-slate-500">
                      Showing {visibleTemplateSuggestionJobsRangeStart}-
                      {visibleTemplateSuggestionJobsRangeEnd} of{' '}
                      {templateSuggestionJobs.length}
                      {templateSuggestionJobsHasMore ? '+' : ''} loaded jobs ·
                      {' '}Page {templateSuggestionJobsPage}
                    </div>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  {loadingTemplateSuggestionJobs && (
                    <div className="text-xs text-slate-500">Refreshing...</div>
                  )}
                  <button
                    type="button"
                    onClick={() =>
                      setTemplateSuggestionJobsPage((current) =>
                        Math.max(1, current - 1)
                      )
                    }
                    disabled={
                      loadingTemplateSuggestionJobs ||
                      templateSuggestionJobsPage <= 1
                    }
                    className="rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm text-slate-700 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-400"
                  >
                    Previous
                  </button>
                  <button
                    type="button"
                    onClick={() =>
                      setTemplateSuggestionJobsPage((current) => current + 1)
                    }
                    disabled={
                      loadingTemplateSuggestionJobs ||
                      !templateSuggestionJobsCanAdvance
                    }
                    className="rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm text-slate-700 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-400"
                  >
                    Next
                  </button>
                </div>
              </div>
              <div className="space-y-2">
                {templateSuggestionJobs.length === 0 &&
                  !loadingTemplateSuggestionJobs && (
                    <div className="rounded-md border border-dashed border-slate-300 px-3 py-4 text-sm text-slate-500">
                      No monster template draft jobs yet.
                    </div>
                  )}
                {visibleTemplateSuggestionJobs.map((job) => {
                  const selected = job.jobId === selectedTemplateSuggestionJobId;
                  return (
                    <button
                      key={job.jobId}
                      type="button"
                      onClick={() => setSelectedTemplateSuggestionJobId(job.jobId)}
                      className={`w-full rounded-lg border px-3 py-3 text-left transition ${
                        selected
                          ? 'border-indigo-500 bg-indigo-50'
                          : 'border-slate-200 bg-slate-50 hover:border-slate-300'
                      }`}
                    >
                      <div className="flex items-center justify-between gap-3">
                        <div className="font-medium text-slate-900">
                          {formatMonsterTemplateTypeLabel(job.monsterType)} ·{' '}
                          {job.totalCount} {job.yeetIt ? 'template' : 'draft'}
                          {job.totalCount === 1 ? '' : 's'}
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
                        Genre:{' '}
                        {formatGenreLabel(
                          genres.find((genre) => genre.id === job.genreId)
                        )}{' '}
                        {job.zoneKind
                          ? `· ${zoneKindLabel(job.zoneKind, zoneKindBySlug)} `
                          : ''}
                        · {job.yeetIt ? 'yeet mode' : 'draft mode'} 
                        · {job.createdCount}/{job.totalCount} ready
                      </div>
                      <div className="mt-2 flex flex-wrap gap-2 text-[11px] text-slate-500">
                        {[
                          job.source
                            ? job.source.replace(/_/g, ' ')
                            : 'generated',
                          job.updatedAt
                            ? `Updated ${formatDate(job.updatedAt)}`
                            : null,
                        ]
                          .filter((entry): entry is string => Boolean(entry))
                          .map((entry) => (
                            <span
                              key={`${job.jobId}-${entry}`}
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

            <div className="rounded-lg border border-gray-200 bg-white p-4">
              <div className="mb-3 flex flex-col gap-1 md:flex-row md:items-center md:justify-between">
                <div>
                  <div className="text-sm font-semibold text-slate-800">
                    {selectedTemplateSuggestionJob
                      ? `${selectedTemplateSuggestionJob.yeetIt ? 'Results' : 'Drafts'} for ${formatMonsterTemplateTypeLabel(
                          selectedTemplateSuggestionJob.monsterType
                        )}`
                      : 'Generated Drafts'}
                  </div>
                  {selectedTemplateSuggestionJob && (
                    <div className="space-y-1 text-xs text-slate-500">
                      <div>
                        Genre:{' '}
                        {formatGenreLabel(
                          genres.find(
                            (genre) =>
                              genre.id === selectedTemplateSuggestionJob.genreId
                          )
                        )}{' '}
                        {selectedTemplateSuggestionJob.zoneKind
                          ? `· ${zoneKindLabel(
                              selectedTemplateSuggestionJob.zoneKind,
                              zoneKindBySlug
                            )} `
                          : ''}
                        · {selectedTemplateSuggestionJob.createdCount}{' '}
                        {selectedTemplateSuggestionJob.yeetIt
                          ? 'template'
                          : 'draft'}
                        {selectedTemplateSuggestionJob.createdCount === 1
                          ? ''
                          : 's'}
                        {' '}·{' '}
                        {selectedTemplateSuggestionJob.yeetIt
                          ? 'yeet mode'
                          : 'draft mode'}
                      </div>
                    </div>
                  )}
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  {selectedTemplateSuggestionJob &&
                    !selectedTemplateSuggestionJob.yeetIt && (
                    <button
                      type="button"
                      onClick={() =>
                        void handleConvertAllTemplateSuggestionDrafts()
                      }
                      disabled={
                        loadingTemplateSuggestionDrafts ||
                        convertingAllTemplateSuggestionDrafts ||
                        convertingTemplateSuggestionDraftId !== null ||
                        unconvertedTemplateSuggestionDrafts.length === 0
                      }
                      className="rounded-md bg-emerald-600 px-3 py-2 text-sm font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
                    >
                      {convertingAllTemplateSuggestionDrafts
                        ? 'Converting all...'
                        : `Convert All to Templates${
                            unconvertedTemplateSuggestionDrafts.length > 0
                              ? ` (${unconvertedTemplateSuggestionDrafts.length})`
                              : ''
                          }`}
                    </button>
                  )}
                  {loadingTemplateSuggestionDrafts && (
                    <div className="text-xs text-slate-500">
                      Loading drafts...
                    </div>
                  )}
                </div>
              </div>

              {templateSuggestionDrafts.length > 0 && (
                <div className="mb-3 flex flex-col gap-2 rounded-md border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-600 md:flex-row md:items-center md:justify-between">
                  <div>
                    Showing {templateSuggestionDraftRangeStart}-
                    {templateSuggestionDraftRangeEnd} of{' '}
                    {templateSuggestionDrafts.length} drafts · Page{' '}
                    {templateSuggestionDraftPage} of{' '}
                    {templateSuggestionDraftTotalPages}
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      onClick={() =>
                        setTemplateSuggestionDraftPage((current) =>
                          Math.max(1, current - 1)
                        )
                      }
                      disabled={templateSuggestionDraftPage <= 1}
                      className="rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm text-slate-700 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-400"
                    >
                      Previous
                    </button>
                    <button
                      type="button"
                      onClick={() =>
                        setTemplateSuggestionDraftPage((current) =>
                          Math.min(templateSuggestionDraftTotalPages, current + 1)
                        )
                      }
                      disabled={
                        templateSuggestionDraftPage >=
                        templateSuggestionDraftTotalPages
                      }
                      className="rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm text-slate-700 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-400"
                    >
                      Next
                    </button>
                  </div>
                </div>
              )}

              <div className="grid grid-cols-1 gap-3 xl:grid-cols-2">
                {templateSuggestionDrafts.length === 0 &&
                  !loadingTemplateSuggestionDrafts && (
                    <div className="rounded-md border border-dashed border-slate-300 px-3 py-6 text-sm text-slate-500 xl:col-span-2">
                      {selectedTemplateSuggestionJob
                        ? selectedTemplateSuggestionJob.yeetIt
                          ? 'This job yeeted its generated monsters straight into live templates, so there are no drafts to review here.'
                          : 'This job has not produced any drafts yet.'
                        : 'Select a draft job to review its generated templates.'}
                    </div>
                  )}
                {paginatedTemplateSuggestionDrafts.map((draft) => (
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
                            {formatMonsterTemplateTypeLabel(
                              draft.monsterType || draft.payload.monsterType
                            )}
                          </span>
                          <span className="rounded-full bg-slate-200 px-2 py-1 text-slate-700">
                            {formatGenreLabel(
                              draft.genre ??
                                genres.find(
                                  (genre) => genre.id === draft.genreId
                                ) ??
                                null
                            )}
                          </span>
                          {draft.zoneKind && (
                            <span className="rounded-full bg-slate-200 px-2 py-1 text-slate-700">
                              {zoneKindLabel(draft.zoneKind, zoneKindBySlug)}
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
                    {draft.description && (
                      <p className="mt-3 text-sm text-slate-700">
                        {draft.description}
                      </p>
                    )}
                    <p className="mt-3 text-sm text-slate-600">
                      STR {draft.payload.baseStrength} · DEX{' '}
                      {draft.payload.baseDexterity} · CON{' '}
                      {draft.payload.baseConstitution} · INT{' '}
                      {draft.payload.baseIntelligence} · WIS{' '}
                      {draft.payload.baseWisdom} · CHA{' '}
                      {draft.payload.baseCharisma}
                    </p>
                    <p className="mt-2 text-sm text-slate-500">
                      Damage Bonuses:{' '}
                      {summarizeAffinityMap(
                        monsterTemplateSuggestionDamageMap(draft.payload),
                        damageBonusFieldOptions
                      )}
                    </p>
                    <p className="mt-1 text-sm text-slate-500">
                      Resistances:{' '}
                      {summarizeAffinityMap(
                        monsterTemplateSuggestionResistanceMap(draft.payload),
                        resistanceFieldOptions
                      )}
                    </p>
                    <div className="mt-4 flex flex-wrap gap-2">
                      <button
                        type="button"
                        onClick={() =>
                          void handleConvertTemplateSuggestionDraft(draft.id)
                        }
                        disabled={
                          draft.status === 'converted' ||
                          convertingTemplateSuggestionDraftId === draft.id ||
                          convertingAllTemplateSuggestionDrafts
                        }
                        className="rounded-md bg-emerald-600 px-3 py-2 text-sm font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-300"
                      >
                        {draft.status === 'converted'
                          ? 'Converted'
                          : convertingAllTemplateSuggestionDrafts
                            ? 'Bulk converting...'
                            : convertingTemplateSuggestionDraftId === draft.id
                              ? 'Converting...'
                              : 'Convert to Template'}
                      </button>
                      <button
                        type="button"
                        onClick={() =>
                          void handleDeleteTemplateSuggestionDraft(draft.id)
                        }
                        disabled={
                          draft.status === 'converted' ||
                          deletingTemplateSuggestionDraftId === draft.id ||
                          convertingAllTemplateSuggestionDrafts
                        }
                        className="rounded-md border border-slate-300 bg-white px-3 py-2 text-sm text-slate-700 disabled:cursor-not-allowed disabled:bg-slate-100"
                      >
                        {deletingTemplateSuggestionDraftId === draft.id
                          ? 'Deleting...'
                          : 'Delete Draft'}
                      </button>
                    </div>
                  </div>
                ))}
              </div>

              {templateSuggestionDrafts.length > 0 &&
                templateSuggestionDraftTotalPages > 1 && (
                  <div className="mt-3 flex flex-col gap-2 rounded-md border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-600 md:flex-row md:items-center md:justify-between">
                    <div>
                      Showing {templateSuggestionDraftRangeStart}-
                      {templateSuggestionDraftRangeEnd} of{' '}
                      {templateSuggestionDrafts.length} drafts · Page{' '}
                      {templateSuggestionDraftPage} of{' '}
                      {templateSuggestionDraftTotalPages}
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() =>
                          setTemplateSuggestionDraftPage((current) =>
                            Math.max(1, current - 1)
                          )
                        }
                        disabled={templateSuggestionDraftPage <= 1}
                        className="rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm text-slate-700 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-400"
                      >
                        Previous
                      </button>
                      <button
                        type="button"
                        onClick={() =>
                          setTemplateSuggestionDraftPage((current) =>
                            Math.min(templateSuggestionDraftTotalPages, current + 1)
                          )
                        }
                        disabled={
                          templateSuggestionDraftPage >=
                          templateSuggestionDraftTotalPages
                        }
                        className="rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm text-slate-700 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-400"
                      >
                        Next
                      </button>
                    </div>
                  </div>
                )}
            </div>
          </div>
        </div>

        <div className="qa-card">
          <div className="mb-3">
            <h2 className="text-lg font-semibold">
              Undiscovered Encounter Icons
            </h2>
            <p className="text-sm text-gray-600">
              Standard, boss, and raid encounters each have their own mystery
              pin.
            </p>
          </div>
          <div className="grid gap-4 lg:grid-cols-3">
            {(Object.keys(encounterIconDefaults) as MonsterEncounterType[]).map(
              (encounterType) => {
                const iconState = encounterIconStates[encounterType];
                const meta = encounterIconDefaults[encounterType];
                return (
                  <div
                    key={encounterType}
                    className="rounded-lg border border-gray-200 bg-white p-4"
                  >
                    <div className="flex flex-wrap items-start justify-between gap-3">
                      <div>
                        <h3 className="text-base font-semibold">
                          {meta.label} Icon
                        </h3>
                        <p className="text-xs text-gray-600 break-all">
                          URL: {iconState.url}
                        </p>
                        <p className="text-xs text-gray-600 mt-1">
                          Requested:{' '}
                          {formatDate(iconState.requestedAt ?? undefined)}
                          {' · '}
                          Last updated:{' '}
                          {formatDate(iconState.lastModified ?? undefined)}
                        </p>
                      </div>
                      <span
                        className={`inline-flex text-white text-xs px-2 py-0.5 rounded ${staticStatusClassName(
                          iconState.status
                        )}`}
                      >
                        {iconState.status || 'unknown'}
                      </span>
                    </div>
                    <div className="mt-3 flex flex-wrap gap-2">
                      <button
                        className="qa-btn qa-btn-secondary"
                        onClick={() =>
                          void refreshEncounterIconStatus(encounterType, true)
                        }
                        disabled={iconState.statusLoading}
                      >
                        {iconState.statusLoading
                          ? 'Refreshing...'
                          : 'Refresh Status'}
                      </button>
                      <button
                        className="qa-btn qa-btn-secondary"
                        onClick={() =>
                          void handleGenerateEncounterIcon(encounterType)
                        }
                        disabled={iconState.busy || iconState.statusLoading}
                      >
                        {iconState.busy ? 'Working...' : 'Generate Icon'}
                      </button>
                      <button
                        className="qa-btn qa-btn-danger"
                        onClick={() =>
                          void handleDeleteEncounterIcon(encounterType)
                        }
                        disabled={iconState.busy || iconState.statusLoading}
                      >
                        {iconState.busy ? 'Working...' : 'Delete Icon'}
                      </button>
                    </div>
                    <label className="block text-sm mt-3">
                      Generation Prompt
                      <textarea
                        className="block w-full border border-gray-300 rounded-md p-2 mt-1 min-h-[88px]"
                        value={iconState.prompt}
                        onChange={(event) =>
                          setEncounterIconState(encounterType, {
                            prompt: event.target.value,
                          })
                        }
                        placeholder={`Prompt used to generate the undiscovered ${meta.label.toLowerCase()} icon.`}
                      />
                    </label>
                    {iconState.exists ? (
                      <div className="mt-3">
                        <img
                          src={`${iconState.url}?v=${iconState.previewNonce}`}
                          alt={`Undiscovered ${meta.label.toLowerCase()} icon preview`}
                          className="w-24 h-24 object-cover border rounded-md bg-gray-50"
                        />
                      </div>
                    ) : (
                      <p className="text-xs text-gray-500 mt-2">
                        No icon currently found at this URL.
                      </p>
                    )}
                    {iconState.message ? (
                      <p className="text-sm text-emerald-700 mt-2">
                        {iconState.message}
                      </p>
                    ) : null}
                    {iconState.error ? (
                      <p className="text-sm text-red-600 mt-2">
                        {iconState.error}
                      </p>
                    ) : null}
                  </div>
                );
              }
            )}
          </div>
        </div>

        <div className="qa-card">
          <div className="flex flex-wrap gap-3">
            <input
              className="min-w-[280px] flex-1 border border-gray-300 rounded-md p-2"
              placeholder="Search names, descriptions, templates..."
              value={query}
              onChange={(event) => setQuery(event.target.value)}
            />
            <input
              className="min-w-[220px] flex-1 border border-gray-300 rounded-md p-2"
              placeholder="Search by zone..."
              value={zoneQuery}
              onChange={(event) => setZoneQuery(event.target.value)}
            />
            <select
              className="rounded-md border border-gray-300 p-2"
              value={templateTypeFilter}
              onChange={(event) =>
                setTemplateTypeFilter(
                  event.target.value as 'all' | MonsterTemplateType
                )
              }
              aria-label="Filter monster templates by type"
            >
              <option value="all">All Template Types</option>
              {monsterTemplateTypeOptions.map((option) => (
                <option key={`filter-${option.value}`} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            <select
              className="rounded-md border border-gray-300 p-2"
              value={genreFilter}
              onChange={(event) => setGenreFilter(event.target.value)}
              aria-label="Filter monsters by genre"
            >
              <option value="all">All Genres</option>
              {genres.map((genre) => (
                <option key={`filter-genre-${genre.id}`} value={genre.id}>
                  {formatGenreLabel(genre)}
                </option>
              ))}
            </select>
          </div>
        </div>

        {error && <div className="qa-card text-red-600">{error}</div>}

        {loading || referenceLoading ? (
          <div className="qa-card">Loading...</div>
        ) : null}

        {!loading && !referenceLoading ? (
          <>
            <div className="qa-card">
              <h2 className="text-lg font-semibold mb-3">Monster Templates</h2>
              <div className="mb-4 flex flex-wrap items-center gap-2">
                <button
                  className={`rounded-md border px-3 py-2 text-sm ${
                    templateTab === 'active'
                      ? 'border-slate-900 bg-slate-900 text-white'
                      : 'border-gray-300 bg-white text-gray-700'
                  }`}
                  onClick={() => setTemplateTab('active')}
                >
                  Active ({activeTemplateCount})
                </button>
                <button
                  className={`rounded-md border px-3 py-2 text-sm ${
                    templateTab === 'archived'
                      ? 'border-slate-900 bg-slate-900 text-white'
                      : 'border-gray-300 bg-white text-gray-700'
                  }`}
                  onClick={() => setTemplateTab('archived')}
                >
                  Archived ({archivedTemplateCount})
                </button>
                <button
                  className="qa-btn qa-btn-secondary"
                  disabled={
                    filteredTemplates.length === 0 ||
                    allFilteredTemplatesSelected
                  }
                  onClick={() =>
                    setSelectedTemplateIds(
                      new Set(filteredTemplates.map((template) => template.id))
                    )
                  }
                >
                  Select All
                </button>
                <button
                  className="qa-btn qa-btn-secondary"
                  disabled={selectedTemplateIds.size === 0}
                  onClick={() => setSelectedTemplateIds(new Set())}
                >
                  Clear Selection
                </button>
                <button
                  className="qa-btn qa-btn-secondary"
                  disabled={selectedTemplateIds.size === 0}
                  onClick={() =>
                    void setTemplatesArchived(
                      Array.from(selectedTemplateIds),
                      templateTab === 'active'
                    )
                  }
                >
                  {templateTab === 'active'
                    ? `Archive Selected (${selectedTemplateIds.size})`
                    : `Restore Selected (${selectedTemplateIds.size})`}
                </button>
                <button
                  className="qa-btn qa-btn-secondary"
                  disabled={affinityRefreshBusy}
                  onClick={() => void handleRefreshTemplateAffinities()}
                >
                  {affinityRefreshBusy
                    ? 'Refreshing Classification...'
                    : selectedTemplateIds.size > 0
                      ? `Refresh Selected Affinities + Zone Kinds (${selectedTemplateIds.size})`
                      : 'Refresh All Template Affinities + Zone Kinds'}
                </button>
                <button
                  className="qa-btn qa-btn-secondary"
                  disabled={progressionResetBusy}
                  onClick={() => void handleResetTemplateProgressions()}
                >
                  {progressionResetBusy
                    ? 'Resetting Progressions...'
                    : selectedTemplateIds.size > 0
                      ? `Reset Selected Progressions (${selectedTemplateIds.size})`
                      : 'Reset All Template Progressions'}
                </button>
              </div>
              {affinityRefreshJob ? (
                <div className="mb-4 flex flex-wrap items-center gap-3 text-sm text-gray-700">
                  <span
                    className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold text-white ${staticStatusClassName(
                      affinityRefreshJob.status
                    )}`}
                  >
                    {formatBulkTemplateStatus(affinityRefreshJob.status)}
                  </span>
                  <span>
                    Progress: {affinityRefreshJob.updatedCount}/
                    {affinityRefreshJob.totalCount}
                  </span>
                  <span>Job: {affinityRefreshJob.jobId}</span>
                  <span>
                    Updated: {formatDate(affinityRefreshJob.updatedAt)}
                  </span>
                </div>
              ) : null}
              {affinityRefreshMessage ? (
                <p className="mb-4 text-sm text-emerald-700">
                  {affinityRefreshMessage}
                </p>
              ) : null}
              {affinityRefreshError ? (
                <p className="mb-4 text-sm text-red-700">
                  {affinityRefreshError}
                </p>
              ) : null}
              {progressionResetJob ? (
                <div className="mb-4 flex flex-wrap items-center gap-3 text-sm text-gray-700">
                  <span
                    className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold text-white ${staticStatusClassName(
                      progressionResetJob.status
                    )}`}
                  >
                    {formatBulkTemplateStatus(progressionResetJob.status)}
                  </span>
                  <span>
                    Progress: {progressionResetJob.updatedCount}/
                    {progressionResetJob.totalCount}
                  </span>
                  <span>Job: {progressionResetJob.jobId}</span>
                  <span>
                    Updated: {formatDate(progressionResetJob.updatedAt)}
                  </span>
                </div>
              ) : null}
              {progressionResetMessage ? (
                <p className="mb-4 text-sm text-emerald-700">
                  {progressionResetMessage}
                </p>
              ) : null}
              {progressionResetError ? (
                <p className="mb-4 text-sm text-red-700">
                  {progressionResetError}
                </p>
              ) : null}
              {filteredTemplates.length === 0 ? (
                <p className="text-sm text-gray-600">No templates found.</p>
              ) : (
                <div className="grid gap-3">
                  {filteredTemplates.map((template) => (
                    <div
                      key={template.id}
                      className="border border-gray-200 rounded-md p-3 bg-white"
                    >
                      <div className="flex items-start justify-between gap-3">
                        <div className="flex gap-3 min-w-0">
                          {template.thumbnailUrl || template.imageUrl ? (
                            <button
                              type="button"
                              className="block"
                              onClick={() => openTemplateImagePreview(template)}
                              title="Open image preview"
                            >
                              <img
                                src={template.thumbnailUrl || template.imageUrl}
                                alt={template.name}
                                className="w-14 h-14 rounded object-cover border border-gray-200 cursor-zoom-in"
                              />
                            </button>
                          ) : (
                            <div className="w-14 h-14 rounded bg-gray-200 flex items-center justify-center text-gray-500">
                              ?
                            </div>
                          )}
                          <div>
                            <div className="flex items-center gap-2">
                              <input
                                type="checkbox"
                                checked={selectedTemplateIds.has(template.id)}
                                onChange={(event) =>
                                  toggleTemplateSelection(
                                    template.id,
                                    event.target.checked
                                  )
                                }
                                aria-label={`Select ${template.name}`}
                              />
                              <div className="font-semibold text-base">
                                {template.name}
                              </div>
                              <span
                                className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                                  template.archived
                                    ? 'bg-amber-100 text-amber-800'
                                    : 'bg-emerald-100 text-emerald-700'
                                }`}
                              >
                                {template.archived ? 'Archived' : 'Active'}
                              </span>
                            </div>
                            <p className="text-xs uppercase tracking-wide text-gray-500 mt-1">
                              {formatMonsterTemplateTypeLabel(
                                template.monsterType
                              )}
                            </p>
                            <p className="text-sm text-gray-500 mt-1">
                              Genre: {formatGenreLabel(template.genre)}
                            </p>
                            <p className="text-sm text-gray-500 mt-1">
                              Likely Zone Kind:{' '}
                              {zoneKindLabel(template.zoneKind, zoneKindBySlug)}
                            </p>
                            {template.description ? (
                              <p className="text-sm text-gray-600 mt-1">
                                {template.description}
                              </p>
                            ) : null}
                            <p className="text-sm text-gray-600 mt-1">
                              STR {template.baseStrength} · DEX{' '}
                              {template.baseDexterity} · CON{' '}
                              {template.baseConstitution} · INT{' '}
                              {template.baseIntelligence} · WIS{' '}
                              {template.baseWisdom} · CHA{' '}
                              {template.baseCharisma}
                            </p>
                            <p className="text-sm text-gray-500 mt-1">
                              Damage Bonuses:{' '}
                              {summarizeAffinityMap(
                                template.affinityDamageBonuses,
                                damageBonusFieldOptions
                              )}
                            </p>
                            <p className="text-sm text-gray-500 mt-1">
                              Resistances:{' '}
                              {summarizeAffinityMap(
                                template.affinityResistances,
                                resistanceFieldOptions
                              )}
                            </p>
                            <p className="text-sm text-gray-500 mt-1">
                              Spell Progressions:{' '}
                              {template.progressions?.filter(
                                (progression) =>
                                  (progression.abilityType ?? 'spell') !==
                                  'technique'
                              ).length ?? 0}
                              {' · '}
                              Technique Progressions:{' '}
                              {template.progressions?.filter(
                                (progression) =>
                                  (progression.abilityType ?? 'spell') ===
                                  'technique'
                              ).length ?? 0}
                            </p>
                            <p className="text-xs text-gray-500 mt-1">
                              Image generation:{' '}
                              {formatGenerationStatus(
                                template.imageGenerationStatus
                              )}
                            </p>
                            {template.imageGenerationError ? (
                              <p className="text-xs text-red-600 mt-1">
                                {template.imageGenerationError}
                              </p>
                            ) : null}
                          </div>
                        </div>
                        <div className="flex gap-2">
                          <button
                            className="qa-btn qa-btn-secondary"
                            onClick={() =>
                              handleGenerateTemplateImage(template)
                            }
                            disabled={
                              generatingTemplateId === template.id ||
                              ['queued', 'in_progress'].includes(
                                template.imageGenerationStatus || ''
                              )
                            }
                          >
                            {generatingTemplateId === template.id
                              ? 'Queueing...'
                              : 'Generate Image'}
                          </button>
                          <button
                            className="qa-btn qa-btn-secondary"
                            onClick={() => openEditTemplate(template)}
                          >
                            Edit
                          </button>
                          <button
                            className="qa-btn qa-btn-secondary"
                            onClick={() =>
                              void setTemplatesArchived(
                                [template.id],
                                !template.archived
                              )
                            }
                          >
                            {template.archived ? 'Restore' : 'Archive'}
                          </button>
                          <button
                            className="qa-btn qa-btn-danger"
                            onClick={() => deleteTemplate(template)}
                          >
                            Delete
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
              <PaginationControls
                page={templatePage}
                pageSize={monsterListPageSize}
                total={templateTotal}
                label="templates"
                onPageChange={setTemplatePage}
              />
            </div>

            <div className="qa-card">
              <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
                <h2 className="text-lg font-semibold">Monsters</h2>
                <div className="flex flex-wrap items-center gap-2">
                  <button
                    type="button"
                    className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                    onClick={selectAllVisibleMonsters}
                    disabled={
                      filteredMonsters.length === 0 ||
                      allFilteredMonstersSelected ||
                      bulkDeletingMonsters
                    }
                  >
                    Select All
                  </button>
                  <button
                    type="button"
                    className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                    onClick={clearMonsterSelection}
                    disabled={
                      selectedMonsterIds.size === 0 || bulkDeletingMonsters
                    }
                  >
                    Clear Selection
                  </button>
                  <button
                    type="button"
                    className="qa-btn qa-btn-danger"
                    onClick={handleBulkDeleteMonsters}
                    disabled={
                      selectedMonsterIds.size === 0 || bulkDeletingMonsters
                    }
                  >
                    {bulkDeletingMonsters
                      ? `Deleting ${selectedMonsterIds.size}...`
                      : `Delete Selected (${selectedMonsterIds.size})`}
                  </button>
                </div>
              </div>
              {filteredMonsters.length === 0 ? (
                <p className="text-sm text-gray-600">No monsters found.</p>
              ) : (
                <div className="grid gap-4">
                  {filteredMonsters.map((monster) => (
                    <div
                      key={monster.id}
                      className="border border-gray-200 rounded-md p-3 bg-white"
                    >
                      <div className="flex flex-col md:flex-row md:items-start md:justify-between gap-4">
                        <div className="flex gap-3 min-w-0">
                          <input
                            type="checkbox"
                            className="h-4 w-4 mt-1"
                            checked={selectedMonsterIdSet.has(monster.id)}
                            disabled={bulkDeletingMonsters}
                            onChange={() => toggleMonsterSelection(monster.id)}
                          />
                          {monster.thumbnailUrl || monster.imageUrl ? (
                            <button
                              type="button"
                              className="block"
                              onClick={() => openMonsterImagePreview(monster)}
                              title="Open image preview"
                            >
                              <img
                                src={monster.thumbnailUrl || monster.imageUrl}
                                alt={monster.name}
                                className="w-16 h-16 rounded object-cover border border-gray-200 cursor-zoom-in"
                              />
                            </button>
                          ) : (
                            <div className="w-16 h-16 rounded bg-gray-200 flex items-center justify-center text-gray-500">
                              ?
                            </div>
                          )}
                          <div className="min-w-0">
                            <h3 className="font-semibold text-lg truncate">
                              {monster.name}
                            </h3>
                            <p className="text-sm text-gray-600">
                              Zone:{' '}
                              {zoneNameById.get(monster.zoneId) ??
                                monster.zoneId}{' '}
                              · Level {monster.level}
                            </p>
                            <p className="text-sm text-gray-600">
                              Zone Kind:{' '}
                              {zoneKindSummaryLabel(
                                monster.zoneKind,
                                zoneDefaultKindById.get(monster.zoneId) ?? '',
                                zoneKindBySlug
                              )}
                            </p>
                            {monster.zoneKind?.trim() &&
                            (zoneDefaultKindById.get(monster.zoneId) ?? '') &&
                            monster.zoneKind.trim() !==
                              (zoneDefaultKindById.get(monster.zoneId) ??
                                '') ? (
                              <p className="text-xs text-gray-500">
                                Zone default:{' '}
                                {zoneKindLabel(
                                  zoneDefaultKindById.get(monster.zoneId) ?? '',
                                  zoneKindBySlug
                                )}
                              </p>
                            ) : null}
                            <p className="text-sm text-gray-600">
                              Template:{' '}
                              {monster.template?.name ??
                                (monster.templateId
                                  ? templateNameById.get(monster.templateId)
                                  : 'N/A')}
                            </p>
                            <p className="text-sm text-gray-600">
                              Genre:{' '}
                              {formatGenreLabel(
                                monster.genre ?? monster.template?.genre
                              )}
                            </p>
                            <p className="text-sm text-gray-600">
                              Dominant Hand:{' '}
                              {monster.dominantHandInventoryItem?.name ??
                                monster.weaponInventoryItem?.name ??
                                ((monster.dominantHandInventoryItemId ??
                                monster.weaponInventoryItemId)
                                  ? `#${monster.dominantHandInventoryItemId ?? monster.weaponInventoryItemId}`
                                  : 'N/A')}
                            </p>
                            <p className="text-sm text-gray-600">
                              Off Hand:{' '}
                              {monster.offHandInventoryItem?.name ??
                                (monster.offHandInventoryItemId
                                  ? `#${monster.offHandInventoryItemId}`
                                  : 'N/A')}
                            </p>
                            <p className="text-sm text-gray-600">
                              Damage {monster.attackDamageMin}-
                              {monster.attackDamageMax} · Swipes{' '}
                              {monster.attackSwipesPerAttack}
                            </p>
                            <p className="text-sm text-gray-600">
                              Health {monster.health}/{monster.maxHealth} · Mana{' '}
                              {monster.mana}/{monster.maxMana}
                            </p>
                            <p className="text-sm text-gray-600">
                              Damage Bonuses:{' '}
                              {summarizeAffinityMap(
                                monster.affinityDamageBonuses,
                                damageBonusFieldOptions
                              )}
                            </p>
                            <p className="text-sm text-gray-600">
                              Resistances:{' '}
                              {summarizeAffinityMap(
                                monster.affinityResistances,
                                resistanceFieldOptions
                              )}
                            </p>
                            <p className="text-sm text-gray-600">
                              STR {monster.strength} · DEX {monster.dexterity} ·
                              CON {monster.constitution} · INT{' '}
                              {monster.intelligence} · WIS {monster.wisdom} ·
                              CHA {monster.charisma}
                            </p>
                            <p className="text-sm text-gray-600">
                              {(monster.rewardMode ?? 'random') === 'random'
                                ? `Random ${(monster.randomRewardSize ?? 'small').toUpperCase()} reward${
                                    (monster.itemRewards?.length ?? 0) > 0
                                      ? ` + ${monster.itemRewards?.length ?? 0} guaranteed item reward${
                                          (monster.itemRewards?.length ?? 0) ===
                                          1
                                            ? ''
                                            : 's'
                                        }`
                                      : ''
                                  }`
                                : `XP ${monster.rewardExperience} · Gold ${monster.rewardGold} · Item rewards ${
                                    monster.itemRewards?.length ?? 0
                                  }`}
                            </p>
                            {monster.materialRewards &&
                            monster.materialRewards.length > 0 ? (
                              <p className="text-xs text-gray-500">
                                {summarizeMaterialRewards(
                                  monster.materialRewards
                                )}
                              </p>
                            ) : null}
                            <p className="text-xs text-gray-500 mt-1">
                              Image generation:{' '}
                              {formatGenerationStatus(
                                monster.imageGenerationStatus
                              )}
                            </p>
                            {monster.imageGenerationError ? (
                              <p className="text-xs text-red-600 mt-1">
                                {monster.imageGenerationError}
                              </p>
                            ) : null}
                          </div>
                        </div>
                        <div className="flex flex-wrap gap-2">
                          <button
                            className="qa-btn qa-btn-secondary"
                            onClick={() => handleGenerateImage(monster)}
                            disabled={generatingMonsterId === monster.id}
                          >
                            {generatingMonsterId === monster.id
                              ? 'Queueing...'
                              : 'Generate Image'}
                          </button>
                          <button
                            className="qa-btn qa-btn-secondary"
                            onClick={() => void openEditMonster(monster)}
                          >
                            Edit
                          </button>
                          <button
                            className="qa-btn qa-btn-danger"
                            onClick={() => deleteMonster(monster)}
                            disabled={bulkDeletingMonsters}
                          >
                            Delete
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
              <PaginationControls
                page={monsterPage}
                pageSize={monsterListPageSize}
                total={monsterTotal}
                label="monsters"
                onPageChange={setMonsterPage}
              />
            </div>

            <div className="qa-card">
              <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
                <h2 className="text-lg font-semibold">Monster Encounters</h2>
                <div className="flex flex-wrap items-center gap-2">
                  <button
                    type="button"
                    className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                    onClick={toggleSelectVisibleEncounters}
                    disabled={
                      filteredEncounters.length === 0 || bulkDeletingEncounters
                    }
                  >
                    {allFilteredEncountersSelected
                      ? 'Unselect Visible'
                      : 'Select Visible'}
                  </button>
                  <button
                    type="button"
                    className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                    onClick={clearEncounterSelection}
                    disabled={
                      selectedEncounterIds.size === 0 || bulkDeletingEncounters
                    }
                  >
                    Clear Selection
                  </button>
                  <button
                    type="button"
                    className="qa-btn qa-btn-danger"
                    onClick={handleBulkDeleteEncounters}
                    disabled={
                      selectedEncounterIds.size === 0 || bulkDeletingEncounters
                    }
                  >
                    {bulkDeletingEncounters
                      ? `Deleting ${selectedEncounterIds.size}...`
                      : `Delete Selected (${selectedEncounterIds.size})`}
                  </button>
                </div>
              </div>
              {filteredEncounters.length === 0 ? (
                <p className="text-sm text-gray-600">No encounters found.</p>
              ) : (
                <div className="grid gap-4">
                  {filteredEncounters.map((encounter) => (
                    <div
                      key={encounter.id}
                      className="border border-gray-200 rounded-md p-3 bg-white"
                    >
                      <div className="flex flex-col md:flex-row md:items-start md:justify-between gap-4">
                        <div className="flex gap-3 min-w-0">
                          <input
                            type="checkbox"
                            className="h-4 w-4 mt-1"
                            checked={selectedEncounterIdSet.has(encounter.id)}
                            disabled={bulkDeletingEncounters}
                            onChange={() =>
                              toggleEncounterSelection(encounter.id)
                            }
                          />
                          {encounter.thumbnailUrl || encounter.imageUrl ? (
                            <button
                              type="button"
                              className="block"
                              onClick={() =>
                                setImagePreview({
                                  url:
                                    encounter.thumbnailUrl ||
                                    encounter.imageUrl,
                                  alt: `${encounter.name || 'Encounter'} image`,
                                })
                              }
                              title="Open image preview"
                            >
                              <img
                                src={
                                  encounter.thumbnailUrl || encounter.imageUrl
                                }
                                alt={encounter.name}
                                className="w-16 h-16 rounded object-cover border border-gray-200 cursor-zoom-in"
                              />
                            </button>
                          ) : (
                            <div className="w-16 h-16 rounded bg-gray-200 flex items-center justify-center text-gray-500">
                              ?
                            </div>
                          )}
                          <div className="min-w-0">
                            <h3 className="font-semibold text-lg truncate">
                              {encounter.name}
                            </h3>
                            <p className="text-sm text-gray-600">
                              Type:{' '}
                              {formatMonsterEncounterTypeLabel(
                                encounter.encounterType
                              )}
                            </p>
                            <p className="text-sm text-gray-600">
                              Zone:{' '}
                              {zoneNameById.get(encounter.zoneId) ??
                                encounter.zoneId}
                            </p>
                            <p className="text-sm text-gray-600">
                              Zone Kind:{' '}
                              {zoneKindSummaryLabel(
                                encounter.zoneKind,
                                zoneDefaultKindById.get(encounter.zoneId) ?? '',
                                zoneKindBySlug
                              )}
                            </p>
                            {encounter.zoneKind?.trim() &&
                            (zoneDefaultKindById.get(encounter.zoneId) ?? '') &&
                            encounter.zoneKind.trim() !==
                              (zoneDefaultKindById.get(encounter.zoneId) ??
                                '') ? (
                              <p className="text-xs text-gray-500">
                                Zone default:{' '}
                                {zoneKindLabel(
                                  zoneDefaultKindById.get(encounter.zoneId) ??
                                    '',
                                  zoneKindBySlug
                                )}
                              </p>
                            ) : null}
                            <p className="text-sm text-gray-600">
                              Monsters:{' '}
                              {encounter.monsterCount ||
                                encounter.members?.length ||
                                0}
                            </p>
                            <p className="text-sm text-gray-600">
                              Scaling:{' '}
                              {encounter.scaleWithUserLevel
                                ? 'Scales with user level'
                                : 'Fixed monster levels'}
                            </p>
                            <p className="text-sm text-gray-600">
                              Rewards:{' '}
                              {encounter.rewardMode === 'random'
                                ? `Random ${(encounter.randomRewardSize ?? 'small').toUpperCase()} reward${
                                    (encounter.itemRewards?.length ?? 0) > 0
                                      ? ` + ${encounter.itemRewards?.length ?? 0} guaranteed item reward${
                                          (encounter.itemRewards?.length ??
                                            0) === 1
                                            ? ''
                                            : 's'
                                        }`
                                      : ''
                                  }`
                                : `XP ${encounter.rewardExperience} · Gold ${encounter.rewardGold} · Item rewards ${
                                    encounter.itemRewards?.length ?? 0
                                  }`}
                            </p>
                            {encounter.materialRewards &&
                            encounter.materialRewards.length > 0 ? (
                              <p className="text-xs text-gray-500">
                                {summarizeMaterialRewards(
                                  encounter.materialRewards
                                )}
                              </p>
                            ) : null}
                            {encounter.recurrenceFrequency ? (
                              <p className="text-sm text-indigo-700">
                                Recurs {encounter.recurrenceFrequency}
                              </p>
                            ) : null}
                            {encounter.description ? (
                              <p className="text-sm text-gray-600 mt-1">
                                {encounter.description}
                              </p>
                            ) : null}
                            <p className="text-xs text-gray-500 mt-1">
                              Members:{' '}
                              {(encounter.members ?? [])
                                .slice()
                                .sort((a, b) => (a.slot ?? 0) - (b.slot ?? 0))
                                .map(
                                  (member) =>
                                    member.monster?.name ||
                                    member.monster?.id ||
                                    'Unknown'
                                )
                                .join(', ') || 'None'}
                            </p>
                          </div>
                        </div>
                        <div className="flex flex-wrap gap-2">
                          <button
                            className="qa-btn qa-btn-secondary"
                            onClick={() => void openEditEncounter(encounter)}
                          >
                            Edit
                          </button>
                          <button
                            className="qa-btn qa-btn-danger"
                            onClick={() => deleteEncounter(encounter)}
                            disabled={bulkDeletingEncounters}
                          >
                            Delete
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
              <PaginationControls
                page={encounterPage}
                pageSize={monsterListPageSize}
                total={encounterTotal}
                label="encounters"
                onPageChange={setEncounterPage}
              />
            </div>
          </>
        ) : null}
      </div>

      {showTemplateModal ? (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-3xl max-h-[92vh] overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <h2 className="text-lg font-semibold">
                {editingTemplate ? 'Edit Template' : 'Create Template'}
              </h2>
              <button
                className="qa-btn qa-btn-secondary"
                onClick={closeTemplateModal}
              >
                Close
              </button>
            </div>
            <div className="p-6 overflow-y-auto max-h-[calc(92vh-72px)] space-y-4">
              <div className="grid md:grid-cols-2 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Template Type</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.monsterType}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        monsterType: event.target.value as MonsterTemplateType,
                      }))
                    }
                  >
                    {monsterTemplateTypeOptions.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Genre</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.genreId}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        genreId: event.target.value,
                      }))
                    }
                  >
                    {genres.length === 0 ? (
                      <option value="">Fantasy</option>
                    ) : (
                      genres.map((genre) => (
                        <option key={genre.id} value={genre.id}>
                          {formatGenreLabel(genre)}
                          {genre.active === false ? ' (inactive)' : ''}
                        </option>
                      ))
                    )}
                  </select>
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Likely Zone Kind</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.zoneKind}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        zoneKind: event.target.value,
                      }))
                    }
                  >
                    <option value="">
                      {zoneKindSelectPlaceholderLabel('', zoneKindBySlug)}
                    </option>
                    {zoneKinds.map((zoneKind) => (
                      <option key={zoneKind.id} value={zoneKind.slug}>
                        {zoneKind.name}
                      </option>
                    ))}
                  </select>
                  {zoneKindDescription(
                    templateForm.zoneKind,
                    '',
                    zoneKindBySlug
                  ) ? (
                    <p className="mt-1 text-xs text-gray-500">
                      {zoneKindDescription(
                        templateForm.zoneKind,
                        '',
                        zoneKindBySlug
                      )}
                    </p>
                  ) : null}
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Name</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.name}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        name: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Description</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.description}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        description: event.target.value,
                      }))
                    }
                  />
                </label>
              </div>

              <div className="grid md:grid-cols-2 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Image URL</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.imageUrl}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        imageUrl: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Thumbnail URL</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.thumbnailUrl}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        thumbnailUrl: event.target.value,
                      }))
                    }
                  />
                </label>
              </div>

              <div className="grid md:grid-cols-3 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Base STR</span>
                  <input
                    type="number"
                    min={1}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.baseStrength}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        baseStrength: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Base DEX</span>
                  <input
                    type="number"
                    min={1}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.baseDexterity}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        baseDexterity: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Base CON</span>
                  <input
                    type="number"
                    min={1}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.baseConstitution}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        baseConstitution: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Base INT</span>
                  <input
                    type="number"
                    min={1}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.baseIntelligence}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        baseIntelligence: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Base WIS</span>
                  <input
                    type="number"
                    min={1}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.baseWisdom}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        baseWisdom: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Base CHA</span>
                  <input
                    type="number"
                    min={1}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.baseCharisma}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        baseCharisma: event.target.value,
                      }))
                    }
                  />
                </label>
              </div>

              <div className="mt-4">
                <div className="text-sm font-medium mb-2">
                  Affinity Damage Bonuses
                </div>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                  {damageBonusFieldOptions.map(({ key, label }) => (
                    <label className="block" key={key}>
                      <span className="block text-sm mb-1">{label}</span>
                      <input
                        type="number"
                        className="w-full border border-gray-300 rounded-md p-2"
                        value={templateForm[key]}
                        onChange={(event) =>
                          setTemplateForm((prev) => ({
                            ...prev,
                            [key]: event.target.value,
                          }))
                        }
                      />
                    </label>
                  ))}
                </div>
              </div>

              <div className="mt-4">
                <div className="text-sm font-medium mb-2">
                  Affinity Resistances
                </div>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                  {resistanceFieldOptions.map(({ key, label }) => (
                    <label className="block" key={key}>
                      <span className="block text-sm mb-1">{label}</span>
                      <input
                        type="number"
                        className="w-full border border-gray-300 rounded-md p-2"
                        value={templateForm[key]}
                        onChange={(event) =>
                          setTemplateForm((prev) => ({
                            ...prev,
                            [key]: event.target.value,
                          }))
                        }
                      />
                    </label>
                  ))}
                </div>
              </div>

              <label className="block">
                <span className="block text-sm mb-1">Spell Progressions</span>
                <select
                  multiple
                  className="w-full border border-gray-300 rounded-md p-2 min-h-[180px]"
                  value={templateForm.spellProgressionIds}
                  onChange={(event) => updateTemplateSpellIds(event.target)}
                >
                  {spellProgressionOptions.map((progression) => (
                    <option key={progression.id} value={progression.id}>
                      {progression.name} ({progression.memberCount})
                    </option>
                  ))}
                </select>
              </label>

              <label className="block">
                <span className="block text-sm mb-1">
                  Technique Progressions
                </span>
                <select
                  multiple
                  className="w-full border border-gray-300 rounded-md p-2 min-h-[180px]"
                  value={templateForm.techniqueProgressionIds}
                  onChange={(event) => updateTemplateTechniqueIds(event.target)}
                >
                  {techniqueProgressionOptions.map((progression) => (
                    <option key={progression.id} value={progression.id}>
                      {progression.name} ({progression.memberCount})
                    </option>
                  ))}
                </select>
              </label>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  className="qa-btn qa-btn-secondary"
                  onClick={closeTemplateModal}
                >
                  Cancel
                </button>
                <button
                  className="qa-btn qa-btn-primary"
                  onClick={saveTemplate}
                >
                  Save Template
                </button>
              </div>
            </div>
          </div>
        </div>
      ) : null}

      {showMonsterModal ? (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-4xl max-h-[92vh] overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <h2 className="text-lg font-semibold">
                {editingMonster ? 'Edit Monster' : 'Create Monster'}
              </h2>
              <button
                className="qa-btn qa-btn-secondary"
                onClick={closeMonsterModal}
              >
                Close
              </button>
            </div>
            <div className="p-6 overflow-y-auto max-h-[calc(92vh-72px)] space-y-4">
              <div className="grid md:grid-cols-3 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Name (optional)</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.name}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        name: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Template</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.templateId}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        templateId: event.target.value,
                      }))
                    }
                  >
                    <option value="">Select template</option>
                    {templateOptions.map((template) => (
                      <option key={template.id} value={template.id}>
                        {template.name}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Zone Kind</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={encounterForm.zoneKind}
                    onChange={(event) =>
                      setEncounterForm((prev) => ({
                        ...prev,
                        zoneKind: event.target.value,
                      }))
                    }
                  >
                    <option value="">
                      {zoneKindSelectPlaceholderLabel(
                        selectedEncounterZoneDefaultKind,
                        zoneKindBySlug
                      )}
                    </option>
                    {zoneKinds.map((zoneKind) => (
                      <option
                        key={`monster-encounter-zone-kind-${zoneKind.id}`}
                        value={zoneKind.slug}
                      >
                        {zoneKind.name}
                      </option>
                    ))}
                  </select>
                  {zoneKindDescription(
                    encounterForm.zoneKind,
                    selectedEncounterZoneDefaultKind,
                    zoneKindBySlug
                  ) ? (
                    <div className="mt-1 text-xs text-gray-500">
                      {zoneKindDescription(
                        encounterForm.zoneKind,
                        selectedEncounterZoneDefaultKind,
                        zoneKindBySlug
                      )}
                    </div>
                  ) : null}
                </label>
              </div>

              <label className="block">
                <span className="block text-sm mb-1">
                  Description (optional)
                </span>
                <textarea
                  className="w-full border border-gray-300 rounded-md p-2"
                  rows={3}
                  value={monsterForm.description}
                  onChange={(event) =>
                    setMonsterForm((prev) => ({
                      ...prev,
                      description: event.target.value,
                    }))
                  }
                />
              </label>

              <div className="grid md:grid-cols-3 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Zone</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.zoneId}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        zoneId: event.target.value,
                      }))
                    }
                  >
                    <option value="">Select zone</option>
                    {zones.map((zone) => (
                      <option key={zone.id} value={zone.id}>
                        {zone.name}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Zone Kind</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.zoneKind}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        zoneKind: event.target.value,
                      }))
                    }
                  >
                    <option value="">
                      {zoneKindSelectPlaceholderLabel(
                        selectedMonsterZoneDefaultKind,
                        zoneKindBySlug
                      )}
                    </option>
                    {zoneKinds.map((zoneKind) => (
                      <option
                        key={`monster-zone-kind-${zoneKind.id}`}
                        value={zoneKind.slug}
                      >
                        {zoneKind.name}
                      </option>
                    ))}
                  </select>
                  {zoneKindDescription(
                    monsterForm.zoneKind,
                    selectedMonsterZoneDefaultKind,
                    zoneKindBySlug
                  ) ? (
                    <div className="mt-1 text-xs text-gray-500">
                      {zoneKindDescription(
                        monsterForm.zoneKind,
                        selectedMonsterZoneDefaultKind,
                        zoneKindBySlug
                      )}
                    </div>
                  ) : null}
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Dominant Hand</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.dominantHandInventoryItemId}
                    onChange={(event) =>
                      updateDominantHandSelection(event.target.value)
                    }
                  >
                    <option value="">Select dominant hand item</option>
                    {dominantHandItems.map((item) => (
                      <option key={item.id} value={String(item.id)}>
                        {item.name} ({item.damageMin}-{item.damageMax} ·{' '}
                        {item.swipesPerAttack ?? 1} swipes)
                      </option>
                    ))}
                  </select>
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Off Hand</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.offHandInventoryItemId}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        offHandInventoryItemId: event.target.value,
                      }))
                    }
                    disabled={dominantIsTwoHanded}
                  >
                    <option value="">None</option>
                    {offHandItems.map((item) => (
                      <option key={item.id} value={String(item.id)}>
                        {item.name}
                      </option>
                    ))}
                  </select>
                </label>
              </div>

              <div className="grid md:grid-cols-3 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Latitude</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.latitude}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        latitude: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Longitude</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.longitude}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        longitude: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Level</span>
                  <input
                    type="number"
                    min={1}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.level}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        level: event.target.value,
                      }))
                    }
                  />
                </label>
              </div>

              <div className="text-sm">
                <button
                  type="button"
                  className="qa-btn qa-btn-secondary"
                  onClick={handleUseCurrentLocation}
                  disabled={geoLoading}
                >
                  {geoLoading ? 'Locating...' : 'Use Current Browser Location'}
                </button>
              </div>

              <div className="text-sm">
                <div className="flex items-center justify-between mb-1">
                  <span>Map Location Picker</span>
                  <span className="text-xs text-gray-500">
                    Click map to set latitude/longitude
                  </span>
                </div>
                {mapboxgl.accessToken ? (
                  <div
                    ref={mapContainerRef}
                    className="w-full h-64 border border-gray-300 rounded-md"
                  />
                ) : (
                  <div className="w-full border border-gray-300 rounded-md p-3 text-sm text-gray-600 bg-gray-50">
                    Missing `REACT_APP_MAPBOX_ACCESS_TOKEN`; map picker is
                    unavailable.
                  </div>
                )}
              </div>

              <div className="grid md:grid-cols-2 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Image URL</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.imageUrl}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        imageUrl: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Thumbnail URL</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.thumbnailUrl}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        thumbnailUrl: event.target.value,
                      }))
                    }
                  />
                </label>
              </div>

              <div className="grid md:grid-cols-2 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Reward Mode</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.rewardMode}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        rewardMode: event.target.value as 'explicit' | 'random',
                      }))
                    }
                  >
                    <option value="random">Random Reward</option>
                    <option value="explicit">Explicit Reward</option>
                  </select>
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Random Reward Size</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.randomRewardSize}
                    disabled={monsterForm.rewardMode !== 'random'}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        randomRewardSize: event.target.value as
                          | 'small'
                          | 'medium'
                          | 'large',
                      }))
                    }
                  >
                    <option value="small">Small</option>
                    <option value="medium">Medium</option>
                    <option value="large">Large</option>
                  </select>
                </label>
              </div>

              <div className="grid md:grid-cols-2 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Reward Experience</span>
                  <input
                    type="number"
                    min={0}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.rewardExperience}
                    disabled={monsterForm.rewardMode !== 'explicit'}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        rewardExperience: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Reward Gold</span>
                  <input
                    type="number"
                    min={0}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={monsterForm.rewardGold}
                    disabled={monsterForm.rewardMode !== 'explicit'}
                    onChange={(event) =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        rewardGold: event.target.value,
                      }))
                    }
                  />
                </label>
              </div>

              <MaterialRewardsEditor
                value={monsterForm.materialRewards}
                onChange={(materialRewards) =>
                  setMonsterForm((prev) => ({ ...prev, materialRewards }))
                }
                disabled={monsterForm.rewardMode !== 'explicit'}
              />

              <div className="border border-gray-200 rounded-md p-3 space-y-2">
                <div className="flex items-center justify-between">
                  <h3 className="font-medium">Guaranteed Item Rewards</h3>
                  <button
                    type="button"
                    className="qa-btn qa-btn-secondary"
                    onClick={() =>
                      setMonsterForm((prev) => ({
                        ...prev,
                        itemRewards: [
                          ...prev.itemRewards,
                          { inventoryItemId: '', quantity: '1' },
                        ],
                      }))
                    }
                  >
                    Add Item
                  </button>
                </div>
                {monsterForm.itemRewards.length === 0 ? (
                  <p className="text-sm text-gray-500">
                    No item rewards configured.
                  </p>
                ) : (
                  monsterForm.itemRewards.map((reward, index) => (
                    <div
                      key={`${reward.inventoryItemId}-${index}`}
                      className="grid grid-cols-[1fr_120px_auto] gap-2"
                    >
                      <select
                        className="border border-gray-300 rounded-md p-2"
                        value={reward.inventoryItemId}
                        onChange={(event) =>
                          updateMonsterItemReward(index, {
                            inventoryItemId: event.target.value,
                          })
                        }
                      >
                        <option value="">Select item</option>
                        {inventoryItems.map((item) => (
                          <option key={item.id} value={String(item.id)}>
                            {item.name}
                          </option>
                        ))}
                      </select>
                      <input
                        type="number"
                        min={1}
                        className="border border-gray-300 rounded-md p-2"
                        value={reward.quantity}
                        onChange={(event) =>
                          updateMonsterItemReward(index, {
                            quantity: event.target.value,
                          })
                        }
                      />
                      <button
                        type="button"
                        className="qa-btn qa-btn-danger"
                        onClick={() =>
                          setMonsterForm((prev) => ({
                            ...prev,
                            itemRewards: prev.itemRewards.filter(
                              (_, i) => i !== index
                            ),
                          }))
                        }
                      >
                        Remove
                      </button>
                    </div>
                  ))
                )}
              </div>

              {monsterForm.rewardMode === 'random' ? (
                <p className="text-xs text-gray-500">
                  Random rewards still grant scaled XP and gold. Guaranteed
                  items above are awarded too; materials stay explicit-only.
                </p>
              ) : null}

              <div className="flex justify-end gap-2 pt-2">
                <button
                  className="qa-btn qa-btn-secondary"
                  onClick={closeMonsterModal}
                >
                  Cancel
                </button>
                <button className="qa-btn qa-btn-primary" onClick={saveMonster}>
                  Save Monster
                </button>
              </div>
            </div>
          </div>
        </div>
      ) : null}

      {showEncounterModal ? (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-4xl max-h-[92vh] overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <h2 className="text-lg font-semibold">
                {editingEncounter
                  ? 'Edit Monster Encounter'
                  : 'Create Monster Encounter'}
              </h2>
              <button
                className="qa-btn qa-btn-secondary"
                onClick={closeEncounterModal}
              >
                Close
              </button>
            </div>
            <div className="p-6 overflow-y-auto max-h-[calc(92vh-72px)] space-y-4">
              <div className="grid md:grid-cols-2 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Name (optional)</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={encounterForm.name}
                    onChange={(event) =>
                      setEncounterForm((prev) => ({
                        ...prev,
                        name: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Zone</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={encounterForm.zoneId}
                    onChange={(event) =>
                      setEncounterForm((prev) => {
                        const zoneId = event.target.value;
                        const zoneMonsters = monsterOptions
                          .filter((monster) => monster.zoneId === zoneId)
                          .slice(0, 9)
                          .map((monster) => monster.id);
                        const selectedInZone = prev.monsterIds.filter((id) =>
                          monsterOptions.some(
                            (monster) =>
                              monster.id === id && monster.zoneId === zoneId
                          )
                        );
                        return {
                          ...prev,
                          zoneId,
                          monsterIds:
                            selectedInZone.length > 0
                              ? selectedInZone
                              : zoneMonsters.slice(0, 3),
                        };
                      })
                    }
                  >
                    <option value="">Select zone</option>
                    {zones.map((zone) => (
                      <option key={zone.id} value={zone.id}>
                        {zone.name}
                      </option>
                    ))}
                  </select>
                </label>
              </div>

              <label className="block">
                <span className="block text-sm mb-1">Description</span>
                <textarea
                  className="w-full border border-gray-300 rounded-md p-2"
                  rows={3}
                  value={encounterForm.description}
                  onChange={(event) =>
                    setEncounterForm((prev) => ({
                      ...prev,
                      description: event.target.value,
                    }))
                  }
                />
              </label>

              <label className="block">
                <span className="block text-sm mb-1">Encounter Type</span>
                <select
                  className="w-full border border-gray-300 rounded-md p-2"
                  value={encounterForm.encounterType}
                  onChange={(event) =>
                    setEncounterForm((prev) => ({
                      ...prev,
                      encounterType: event.target.value as MonsterEncounterType,
                    }))
                  }
                >
                  {monsterEncounterTypeOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
                <p className="mt-1 text-xs text-gray-500">
                  Boss encounters scale each monster from a player level +5
                  baseline. Raid encounters are tuned for a five-player party.
                </p>
              </label>

              <label className="block text-sm">
                <span className="inline-flex items-center gap-2">
                  <input
                    type="checkbox"
                    checked={encounterForm.scaleWithUserLevel}
                    onChange={(event) =>
                      setEncounterForm((prev) => ({
                        ...prev,
                        scaleWithUserLevel: event.target.checked,
                      }))
                    }
                  />
                  Scale included monster levels with user level
                </span>
              </label>

              <div className="grid md:grid-cols-2 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Reward Mode</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={encounterForm.rewardMode}
                    onChange={(event) =>
                      setEncounterForm((prev) => ({
                        ...prev,
                        rewardMode: event.target.value as 'explicit' | 'random',
                      }))
                    }
                  >
                    <option value="random">Random Reward</option>
                    <option value="explicit">Explicit Reward</option>
                  </select>
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Random Reward Size</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={encounterForm.randomRewardSize}
                    disabled={encounterForm.rewardMode !== 'random'}
                    onChange={(event) =>
                      setEncounterForm((prev) => ({
                        ...prev,
                        randomRewardSize: event.target.value as
                          | 'small'
                          | 'medium'
                          | 'large',
                      }))
                    }
                  >
                    <option value="small">Small</option>
                    <option value="medium">Medium</option>
                    <option value="large">Large</option>
                  </select>
                </label>
              </div>

              <div className="grid md:grid-cols-2 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Reward Experience</span>
                  <input
                    type="number"
                    min={0}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={encounterForm.rewardExperience}
                    disabled={encounterForm.rewardMode !== 'explicit'}
                    onChange={(event) =>
                      setEncounterForm((prev) => ({
                        ...prev,
                        rewardExperience: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Reward Gold</span>
                  <input
                    type="number"
                    min={0}
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={encounterForm.rewardGold}
                    disabled={encounterForm.rewardMode !== 'explicit'}
                    onChange={(event) =>
                      setEncounterForm((prev) => ({
                        ...prev,
                        rewardGold: event.target.value,
                      }))
                    }
                  />
                </label>
              </div>

              <MaterialRewardsEditor
                value={encounterForm.materialRewards}
                onChange={(materialRewards) =>
                  setEncounterForm((prev) => ({ ...prev, materialRewards }))
                }
                disabled={encounterForm.rewardMode !== 'explicit'}
                title="Encounter Material Rewards"
              />

              <div className="border border-gray-200 rounded-md p-3 space-y-2">
                <div className="flex items-center justify-between">
                  <h3 className="font-medium">
                    Guaranteed Encounter Item Rewards
                  </h3>
                  <button
                    type="button"
                    className="qa-btn qa-btn-secondary"
                    onClick={() =>
                      setEncounterForm((prev) => ({
                        ...prev,
                        itemRewards: [
                          ...prev.itemRewards,
                          { inventoryItemId: '', quantity: '1' },
                        ],
                      }))
                    }
                  >
                    Add Item
                  </button>
                </div>
                {encounterForm.itemRewards.length === 0 ? (
                  <p className="text-sm text-gray-500">
                    No encounter item rewards configured.
                  </p>
                ) : (
                  encounterForm.itemRewards.map((reward, index) => (
                    <div
                      key={`${reward.inventoryItemId}-${index}`}
                      className="grid grid-cols-[1fr_120px_auto] gap-2"
                    >
                      <select
                        className="border border-gray-300 rounded-md p-2"
                        value={reward.inventoryItemId}
                        onChange={(event) =>
                          updateEncounterItemReward(index, {
                            inventoryItemId: event.target.value,
                          })
                        }
                      >
                        <option value="">Select item</option>
                        {inventoryItems.map((item) => (
                          <option key={item.id} value={String(item.id)}>
                            {item.name}
                          </option>
                        ))}
                      </select>
                      <input
                        type="number"
                        min={1}
                        className="border border-gray-300 rounded-md p-2"
                        value={reward.quantity}
                        onChange={(event) =>
                          updateEncounterItemReward(index, {
                            quantity: event.target.value,
                          })
                        }
                      />
                      <button
                        type="button"
                        className="qa-btn qa-btn-danger"
                        onClick={() =>
                          setEncounterForm((prev) => ({
                            ...prev,
                            itemRewards: prev.itemRewards.filter(
                              (_, i) => i !== index
                            ),
                          }))
                        }
                      >
                        Remove
                      </button>
                    </div>
                  ))
                )}
              </div>

              <label className="block">
                <span className="block text-sm mb-1">Recurrence</span>
                <select
                  className="w-full border border-gray-300 rounded-md p-2"
                  value={encounterForm.recurrenceFrequency}
                  onChange={(event) =>
                    setEncounterForm((prev) => ({
                      ...prev,
                      recurrenceFrequency: event.target.value,
                    }))
                  }
                >
                  {recurrenceOptions.map((option) => (
                    <option key={option.value || 'none'} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </label>

              <div className="grid md:grid-cols-3 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Latitude</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={encounterForm.latitude}
                    onChange={(event) =>
                      setEncounterForm((prev) => ({
                        ...prev,
                        latitude: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Longitude</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={encounterForm.longitude}
                    onChange={(event) =>
                      setEncounterForm((prev) => ({
                        ...prev,
                        longitude: event.target.value,
                      }))
                    }
                  />
                </label>
                <div className="block">
                  <span className="block text-sm mb-1">Tools</span>
                  <button
                    type="button"
                    className="qa-btn qa-btn-secondary"
                    onClick={handleUseCurrentEncounterLocation}
                    disabled={geoLoading}
                  >
                    {geoLoading
                      ? 'Locating...'
                      : 'Use Current Browser Location'}
                  </button>
                </div>
              </div>

              <div className="text-sm">
                <div className="flex items-center justify-between mb-1">
                  <span>Map Location Picker</span>
                  <span className="text-xs text-gray-500">
                    Click map to set latitude/longitude
                  </span>
                </div>
                {mapboxgl.accessToken ? (
                  <div
                    ref={encounterMapContainerRef}
                    className="w-full h-64 border border-gray-300 rounded-md"
                  />
                ) : (
                  <div className="w-full border border-gray-300 rounded-md p-3 text-sm text-gray-600 bg-gray-50">
                    Missing `REACT_APP_MAPBOX_ACCESS_TOKEN`; map picker is
                    unavailable.
                  </div>
                )}
              </div>

              <div className="grid md:grid-cols-2 gap-4">
                <label className="block">
                  <span className="block text-sm mb-1">Image URL</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={encounterForm.imageUrl}
                    onChange={(event) =>
                      setEncounterForm((prev) => ({
                        ...prev,
                        imageUrl: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Thumbnail URL</span>
                  <input
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={encounterForm.thumbnailUrl}
                    onChange={(event) =>
                      setEncounterForm((prev) => ({
                        ...prev,
                        thumbnailUrl: event.target.value,
                      }))
                    }
                  />
                </label>
              </div>

              <div className="border border-gray-200 rounded-md p-3">
                <div className="flex items-center justify-between">
                  <h3 className="font-medium">Encounter Monsters (1-9)</h3>
                  <span className="text-xs text-gray-500">
                    Selected: {encounterForm.monsterIds.length}
                  </span>
                </div>
                {encounterForm.zoneId ? null : (
                  <p className="mt-2 text-sm text-amber-700">
                    Select a zone first.
                  </p>
                )}
                {encounterForm.zoneId &&
                encounterMonsterOptions.length === 0 ? (
                  <p className="mt-2 text-sm text-gray-500">
                    No monsters available for this zone.
                  </p>
                ) : null}
                {encounterForm.zoneId && encounterMonsterOptions.length > 0 ? (
                  <div className="mt-2 max-h-56 overflow-y-auto space-y-2">
                    {encounterMonsterOptions.map((monster) => {
                      const checked = encounterForm.monsterIds.includes(
                        monster.id
                      );
                      return (
                        <label
                          key={monster.id}
                          className="flex items-center justify-between rounded border border-gray-200 px-3 py-2 text-sm"
                        >
                          <span>
                            {monster.name} (Lv {monster.level})
                          </span>
                          <input
                            type="checkbox"
                            checked={checked}
                            onChange={(event) =>
                              setEncounterForm((prev) => {
                                if (event.target.checked) {
                                  if (prev.monsterIds.length >= 9) {
                                    alert(
                                      'An encounter can include at most 9 monsters.'
                                    );
                                    return prev;
                                  }
                                  return {
                                    ...prev,
                                    monsterIds: [
                                      ...prev.monsterIds,
                                      monster.id,
                                    ],
                                  };
                                }
                                return {
                                  ...prev,
                                  monsterIds: prev.monsterIds.filter(
                                    (id) => id !== monster.id
                                  ),
                                };
                              })
                            }
                          />
                        </label>
                      );
                    })}
                  </div>
                ) : null}
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  className="qa-btn qa-btn-secondary"
                  onClick={closeEncounterModal}
                >
                  Cancel
                </button>
                <button
                  className="qa-btn qa-btn-primary"
                  onClick={saveEncounter}
                >
                  Save Encounter
                </button>
              </div>
            </div>
          </div>
        </div>
      ) : null}

      {imagePreview ? (
        <div
          className="fixed inset-0 bg-black/70 flex items-center justify-center p-4 z-[60]"
          onClick={closeImagePreview}
        >
          <div
            className="relative bg-white rounded-lg shadow-xl p-3 max-w-5xl w-full max-h-[92vh] flex items-center justify-center"
            onClick={(event) => event.stopPropagation()}
          >
            <button
              type="button"
              className="absolute top-2 right-2 qa-btn qa-btn-secondary"
              onClick={closeImagePreview}
            >
              Close
            </button>
            <img
              src={imagePreview.url}
              alt={imagePreview.alt}
              className="max-w-full max-h-[84vh] rounded object-contain"
            />
          </div>
        </div>
      ) : null}
    </div>
  );
};

export default Monsters;
