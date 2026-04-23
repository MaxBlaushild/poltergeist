import { useAPI, useInventory, useZoneContext } from '@poltergeist/contexts';
import React, {
  useCallback,
  useDeferredValue,
  useEffect,
  useMemo,
  useState,
} from 'react';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import {
  PointOfInterest,
  Spell,
  ZoneGenre,
  ZoneKind,
} from '@poltergeist/types';
import { useSearchParams } from 'react-router-dom';
import ContentDashboard from './ContentDashboard.tsx';
import {
  countBy,
  difficultyBandLabel,
  useAdminAggregateDataset,
} from './contentDashboardUtils.ts';
import {
  MaterialRewardsEditor,
  MaterialRewardForm,
} from './MaterialRewardsEditor.tsx';

type ScenarioRewardItem = {
  inventoryItemId: number;
  quantity: number;
};

type ScenarioRewardSpell = {
  spellId: string;
};

type ScenarioFailurePenaltyMode = 'shared' | 'individual';
type ScenarioSuccessRewardMode = 'shared' | 'individual';
type ScenarioFailureDrainType = 'none' | 'flat' | 'percent';

type ScenarioFailureStatus = {
  name: string;
  description: string;
  effect: string;
  effectType: string;
  positive: boolean;
  damagePerTick: number;
  healthPerTick: number;
  manaPerTick: number;
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

type ScenarioOption = {
  id?: string;
  optionText: string;
  successText: string;
  failureText: string;
  successHandoffText: string;
  failureHandoffText: string;
  statTag: string;
  proficiencies: string[];
  difficulty?: number | null;
  rewardExperience: number;
  rewardGold: number;
  materialRewards: MaterialRewardForm[];
  failureHealthDrainType: ScenarioFailureDrainType;
  failureHealthDrainValue: number;
  failureManaDrainType: ScenarioFailureDrainType;
  failureManaDrainValue: number;
  failureStatuses: ScenarioFailureStatus[];
  successHealthRestoreType: ScenarioFailureDrainType;
  successHealthRestoreValue: number;
  successManaRestoreType: ScenarioFailureDrainType;
  successManaRestoreValue: number;
  successStatuses: ScenarioFailureStatus[];
  itemRewards: ScenarioRewardItem[];
  spellRewards: ScenarioRewardSpell[];
};

type ScenarioRecord = {
  id: string;
  zoneId: string;
  zoneKind?: string;
  genreId: string;
  genre?: ZoneGenre;
  pointOfInterestId?: string | null;
  latitude: number;
  longitude: number;
  prompt: string;
  internalTags?: string[];
  imageUrl: string;
  thumbnailUrl: string;
  difficulty: number;
  scaleWithUserLevel: boolean;
  recurrenceFrequency?: string | null;
  nextRecurrenceAt?: string | null;
  rewardMode?: 'explicit' | 'random';
  randomRewardSize?: 'small' | 'medium' | 'large';
  rewardExperience: number;
  rewardGold: number;
  materialRewards: MaterialRewardForm[];
  openEnded: boolean;
  successHandoffText: string;
  failureHandoffText: string;
  failurePenaltyMode: ScenarioFailurePenaltyMode;
  failureHealthDrainType: ScenarioFailureDrainType;
  failureHealthDrainValue: number;
  failureManaDrainType: ScenarioFailureDrainType;
  failureManaDrainValue: number;
  failureStatuses: ScenarioFailureStatus[];
  successRewardMode: ScenarioSuccessRewardMode;
  successHealthRestoreType: ScenarioFailureDrainType;
  successHealthRestoreValue: number;
  successManaRestoreType: ScenarioFailureDrainType;
  successManaRestoreValue: number;
  successStatuses: ScenarioFailureStatus[];
  options: ScenarioOption[];
  itemRewards: ScenarioRewardItem[];
  spellRewards: ScenarioRewardSpell[];
  attemptedByUser?: boolean;
};

type ScenarioTemplateDashboardRecord = {
  id: string;
  genreId: string;
  genre?: ZoneGenre;
  zoneKind?: string;
  difficulty: number;
  scaleWithUserLevel: boolean;
  rewardMode?: 'explicit' | 'random';
  openEnded: boolean;
};

type ScenarioFormState = {
  zoneId: string;
  zoneKind: string;
  genreId: string;
  pointOfInterestId: string;
  latitude: string;
  longitude: string;
  prompt: string;
  internalTagsInput: string;
  imageUrl: string;
  thumbnailUrl: string;
  difficulty: string;
  scaleWithUserLevel: boolean;
  recurrenceFrequency: string;
  openEnded: boolean;
  rewardMode: 'explicit' | 'random';
  randomRewardSize: 'small' | 'medium' | 'large';
  rewardExperience: string;
  rewardGold: string;
  materialRewards: MaterialRewardForm[];
  successHandoffText: string;
  failureHandoffText: string;
  failurePenaltyMode: ScenarioFailurePenaltyMode;
  failureHealthDrainType: ScenarioFailureDrainType;
  failureHealthDrainValue: string;
  failureManaDrainType: ScenarioFailureDrainType;
  failureManaDrainValue: string;
  failureStatuses: ScenarioFailureStatus[];
  successRewardMode: ScenarioSuccessRewardMode;
  successHealthRestoreType: ScenarioFailureDrainType;
  successHealthRestoreValue: string;
  successManaRestoreType: ScenarioFailureDrainType;
  successManaRestoreValue: string;
  successStatuses: ScenarioFailureStatus[];
  options: ScenarioOption[];
  itemRewards: ScenarioRewardItem[];
  spellRewards: ScenarioRewardSpell[];
};

type ScenarioGenerationJob = {
  id: string;
  zoneId: string;
  genreId: string;
  genre?: ZoneGenre;
  status: string;
  openEnded: boolean;
  latitude?: number | null;
  longitude?: number | null;
  generatedScenarioId?: string | null;
  errorMessage?: string | null;
  createdAt?: string;
  updatedAt?: string;
};

type ScenarioGenerationFormState = {
  zoneId: string;
  genreId: string;
  openEnded: boolean;
  includeLocation: boolean;
  latitude: string;
  longitude: string;
};

type PointOfInterestOption = {
  id: string;
  name: string;
  latitude: number;
  longitude: number;
};

type StaticThumbnailResponse = {
  thumbnailUrl?: string;
  status?: string;
  exists?: boolean;
  requestedAt?: string;
  lastModified?: string;
  prompt?: string;
};

type PaginatedResponse<T> = {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
};

const scenarioListPageSize = 25;

const PaginationControls = ({
  page,
  pageSize,
  total,
  label,
  onPageChange,
}: {
  page: number;
  pageSize: number;
  total: number;
  label: string;
  onPageChange: (page: number) => void;
}) => {
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const start = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const end = total === 0 ? 0 : Math.min(total, page * pageSize);

  return (
    <div className="mt-4 flex flex-wrap items-center justify-between gap-3 border-t pt-3">
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

const defaultScenarioUndiscoveredIconPrompt =
  'A retro 16-bit RPG map marker icon for an undiscovered scenario. Mysterious parchment sigil, subtle compass motif, no text, no logos, transparent or clean background, centered composition, crisp outlines, limited palette.';

const statTags = [
  'strength',
  'dexterity',
  'constitution',
  'intelligence',
  'wisdom',
  'charisma',
] as const;

const failureDrainTypes: ScenarioFailureDrainType[] = [
  'none',
  'flat',
  'percent',
];

const statusEffectTypes = [
  'stat_modifier',
  'damage_over_time',
  'health_over_time',
  'mana_over_time',
] as const;

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

const emptyFailureStatus = (): ScenarioFailureStatus => ({
  name: '',
  description: '',
  effect: '',
  effectType: 'stat_modifier',
  positive: true,
  damagePerTick: 0,
  healthPerTick: 0,
  manaPerTick: 0,
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

const emptyOption = (): ScenarioOption => ({
  optionText: '',
  successText: 'Your approach works, and momentum turns in your favor.',
  failureText: 'The attempt falls short, and the moment slips away.',
  successHandoffText: '',
  failureHandoffText: '',
  statTag: 'charisma',
  proficiencies: [],
  difficulty: null,
  rewardExperience: 0,
  rewardGold: 0,
  materialRewards: [],
  failureHealthDrainType: 'none',
  failureHealthDrainValue: 0,
  failureManaDrainType: 'none',
  failureManaDrainValue: 0,
  failureStatuses: [],
  successHealthRestoreType: 'none',
  successHealthRestoreValue: 0,
  successManaRestoreType: 'none',
  successManaRestoreValue: 0,
  successStatuses: [],
  itemRewards: [],
  spellRewards: [],
});

const emptyFormState = (): ScenarioFormState => ({
  zoneId: '',
  zoneKind: '',
  genreId: '',
  pointOfInterestId: '',
  latitude: '',
  longitude: '',
  prompt: '',
  internalTagsInput: '',
  imageUrl: '',
  thumbnailUrl: '',
  difficulty: '24',
  scaleWithUserLevel: false,
  recurrenceFrequency: '',
  openEnded: false,
  rewardMode: 'random',
  randomRewardSize: 'small',
  rewardExperience: '0',
  rewardGold: '0',
  materialRewards: [],
  successHandoffText: '',
  failureHandoffText: '',
  failurePenaltyMode: 'shared',
  failureHealthDrainType: 'none',
  failureHealthDrainValue: '0',
  failureManaDrainType: 'none',
  failureManaDrainValue: '0',
  failureStatuses: [],
  successRewardMode: 'shared',
  successHealthRestoreType: 'none',
  successHealthRestoreValue: '0',
  successManaRestoreType: 'none',
  successManaRestoreValue: '0',
  successStatuses: [],
  options: [emptyOption()],
  itemRewards: [],
  spellRewards: [],
});

const emptyGenerationFormState = (): ScenarioGenerationFormState => ({
  zoneId: '',
  genreId: '',
  openEnded: false,
  includeLocation: false,
  latitude: '',
  longitude: '',
});

const defaultGenreIdFromList = (genres: ZoneGenre[]): string => {
  const fantasy = genres.find(
    (genre) => (genre.name || '').trim().toLowerCase() === 'fantasy'
  );
  return fantasy?.id ?? genres[0]?.id ?? '';
};

const formatGenreLabel = (genre?: ZoneGenre | null): string =>
  genre?.name?.trim() || 'Fantasy';

const zoneKindLabel = (
  slug: string | null | undefined,
  zoneKindBySlug: Map<string, ZoneKind>
): string => {
  const normalizedSlug = (slug ?? '').trim();
  if (!normalizedSlug) return 'Unassigned';
  return zoneKindBySlug.get(normalizedSlug)?.name?.trim() || normalizedSlug;
};

const recurrenceOptions = [
  { value: '', label: 'No Recurrence' },
  { value: 'daily', label: 'Daily' },
  { value: 'weekly', label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
];

const parseIntValue = (value: string, fallback = 0): number => {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parseFloatValue = (value: string, fallback = 0): number => {
  const parsed = Number.parseFloat(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parseCsv = (value: string): string[] => {
  return value
    .split(',')
    .map((part) => part.trim())
    .filter(Boolean);
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

const normalizeFailureDrainType = (
  value?: string | null
): ScenarioFailureDrainType => {
  if (value === 'flat' || value === 'percent') return value;
  return 'none';
};

const normalizeFailurePenaltyMode = (
  value?: string | null
): ScenarioFailurePenaltyMode => {
  if (value === 'individual') return 'individual';
  return 'shared';
};

const normalizeSuccessRewardMode = (
  value?: string | null
): ScenarioSuccessRewardMode => {
  if (value === 'individual') return 'individual';
  return 'shared';
};

const normalizeFailureStatus = (
  status?: Partial<ScenarioFailureStatus> | null
): ScenarioFailureStatus => {
  const base = emptyFailureStatus();
  if (!status) return base;
  return {
    ...base,
    ...status,
    name: (status.name ?? '').trim(),
    description: (status.description ?? '').trim(),
    effect: (status.effect ?? '').trim(),
    effectType:
      typeof status.effectType === 'string' && status.effectType.trim() !== ''
        ? status.effectType.trim().toLowerCase()
        : base.effectType,
    damagePerTick: Number.isFinite(status.damagePerTick)
      ? Number(status.damagePerTick)
      : 0,
    healthPerTick: Number.isFinite(status.healthPerTick)
      ? Number(status.healthPerTick)
      : 0,
    manaPerTick: Number.isFinite(status.manaPerTick)
      ? Number(status.manaPerTick)
      : 0,
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

const scenarioGenerationStatusBadgeClass = (status: string): string => {
  switch (status) {
    case 'queued':
      return 'bg-slate-600';
    case 'in_progress':
      return 'bg-amber-600';
    case 'completed':
      return 'bg-emerald-600';
    case 'failed':
      return 'bg-red-600';
    default:
      return 'bg-gray-600';
  }
};

const formatDate = (value?: string): string => {
  if (!value) return 'n/a';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
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

mapboxgl.accessToken = process.env.REACT_APP_MAPBOX_ACCESS_TOKEN || '';

export const Scenarios = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const { apiClient } = useAPI();
  const { zones } = useZoneContext();
  const { inventoryItems } = useInventory();
  const [spells, setSpells] = useState<Spell[]>([]);
  const [genres, setGenres] = useState<ZoneGenre[]>([]);
  const [zoneKinds, setZoneKinds] = useState<ZoneKind[]>([]);
  const [zonePointOfInterestMap, setZonePointOfInterestMap] = useState<
    Record<string, PointOfInterestOption[]>
  >({});
  const [pointOfInterestLoadingByZone, setPointOfInterestLoadingByZone] =
    useState<Record<string, boolean>>({});

  const [loading, setLoading] = useState(true);
  const [records, setRecords] = useState<ScenarioRecord[]>([]);
  const [query, setQuery] = useState('');
  const [zoneQuery, setZoneQuery] = useState('');
  const [genreFilter, setGenreFilter] = useState('all');
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [generationForm, setGenerationForm] =
    useState<ScenarioGenerationFormState>(emptyGenerationFormState);
  const [generationJobs, setGenerationJobs] = useState<ScenarioGenerationJob[]>(
    []
  );
  const [generationJobsLoading, setGenerationJobsLoading] = useState(false);
  const [generationSubmitting, setGenerationSubmitting] = useState(false);
  const [generationError, setGenerationError] = useState<string | null>(null);
  const [generationGeoLoading, setGenerationGeoLoading] = useState(false);
  const [generationZoneQuery, setGenerationZoneQuery] = useState('');
  const [showGenerationZoneSuggestions, setShowGenerationZoneSuggestions] =
    useState(false);
  const [scenarioUndiscoveredBusy, setScenarioUndiscoveredBusy] =
    useState(false);
  const [
    scenarioUndiscoveredStatusLoading,
    setScenarioUndiscoveredStatusLoading,
  ] = useState(false);
  const [scenarioUndiscoveredError, setScenarioUndiscoveredError] = useState<
    string | null
  >(null);
  const [scenarioUndiscoveredMessage, setScenarioUndiscoveredMessage] =
    useState<string | null>(null);
  const [scenarioUndiscoveredUrl, setScenarioUndiscoveredUrl] = useState(
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/scenario-undiscovered.png'
  );
  const [scenarioUndiscoveredStatus, setScenarioUndiscoveredStatus] =
    useState('unknown');
  const [scenarioUndiscoveredExists, setScenarioUndiscoveredExists] =
    useState(false);
  const [scenarioUndiscoveredRequestedAt, setScenarioUndiscoveredRequestedAt] =
    useState<string | null>(null);
  const [
    scenarioUndiscoveredLastModified,
    setScenarioUndiscoveredLastModified,
  ] = useState<string | null>(null);
  const [
    scenarioUndiscoveredPreviewNonce,
    setScenarioUndiscoveredPreviewNonce,
  ] = useState<number>(Date.now());
  const [scenarioUndiscoveredPrompt, setScenarioUndiscoveredPrompt] = useState(
    defaultScenarioUndiscoveredIconPrompt
  );
  const deferredQuery = useDeferredValue(query);
  const deferredZoneQuery = useDeferredValue(zoneQuery);
  const defaultGenreId = useMemo(
    () => defaultGenreIdFromList(genres),
    [genres]
  );
  const zoneKindBySlug = useMemo(() => {
    const next = new Map<string, ZoneKind>();
    zoneKinds.forEach((zoneKind) => {
      next.set(zoneKind.slug, zoneKind);
    });
    return next;
  }, [zoneKinds]);
  const dashboardParams = useMemo(
    () => ({
      query: deferredQuery.trim(),
      genreId: genreFilter === 'all' ? '' : genreFilter,
    }),
    [deferredQuery, genreFilter]
  );
  const {
    items: dashboardRecords,
    loading: dashboardLoading,
    error: dashboardError,
  } = useAdminAggregateDataset<ScenarioTemplateDashboardRecord>(
    '/sonar/admin/scenario-templates',
    dashboardParams
  );
  const dashboardMetrics = useMemo(() => {
    const totalScenarios = dashboardRecords.length;
    const openEndedCount = dashboardRecords.filter(
      (record) => record.openEnded
    ).length;
    const explicitRewardCount = dashboardRecords.filter(
      (record) => record.rewardMode === 'explicit'
    ).length;
    const scaledCount = dashboardRecords.filter((record) =>
      record.scaleWithUserLevel
    ).length;

    return [
      { label: 'Templates', value: totalScenarios },
      {
        label: 'Open-ended',
        value: openEndedCount,
        note: `${Math.max(0, totalScenarios - openEndedCount)} choice-based`,
      },
      { label: 'Explicit Rewards', value: explicitRewardCount },
      { label: 'Scaled Difficulty', value: scaledCount },
    ];
  }, [dashboardRecords]);
  const dashboardSections = useMemo(
    () => [
      {
        title: 'Genre Mix',
        note: 'Reusable scenario templates grouped by story genre.',
        buckets: countBy(
          dashboardRecords,
          (record) =>
            formatGenreLabel(
              record.genre ??
                genres.find((genre) => genre.id === record.genreId) ??
                null
            ),
          { emptyLabel: 'Fantasy' }
        ),
      },
      {
        title: 'Zone Kind Coverage',
        note: 'Which environments the current template pool is tagged to support.',
        buckets: countBy(
          dashboardRecords,
          (record) =>
            record.zoneKind?.trim()
              ? zoneKindLabel(record.zoneKind, zoneKindBySlug)
              : 'Unassigned',
          { emptyLabel: 'Unassigned' }
        ),
      },
      {
        title: 'Difficulty Bands',
        note: 'How hard the current template pool skews.',
        buckets: countBy(
          dashboardRecords,
          (record) => difficultyBandLabel(record.difficulty),
          { limit: 4 }
        ),
      },
      {
        title: 'Reward Model',
        note: 'Whether templates carry explicit or randomized rewards.',
        buckets: countBy(dashboardRecords, (record) =>
          record.rewardMode === 'explicit' ? 'Explicit rewards' : 'Randomized'
        ),
      },
    ],
    [dashboardRecords, genres, zoneKindBySlug]
  );

  const [showModal, setShowModal] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<ScenarioFormState>(emptyFormState);
  const [generatingScenarioId, setGeneratingScenarioId] = useState<
    string | null
  >(null);
  const [bulkDeletingScenarios, setBulkDeletingScenarios] = useState(false);
  const [selectedScenarioIds, setSelectedScenarioIds] = useState<Set<string>>(
    new Set()
  );
  const [geoLoading, setGeoLoading] = useState(false);
  const [expandedScenarioImage, setExpandedScenarioImage] = useState<{
    url: string;
    title: string;
  } | null>(null);

  const [deleteId, setDeleteId] = useState<string | null>(null);
  const mapContainerRef = React.useRef<HTMLDivElement | null>(null);
  const mapRef = React.useRef<mapboxgl.Map | null>(null);
  const markerRef = React.useRef<mapboxgl.Marker | null>(null);
  const formLatitudeRef = React.useRef(form.latitude);
  const formLongitudeRef = React.useRef(form.longitude);
  const generationMapContainerRef = React.useRef<HTMLDivElement | null>(null);
  const generationMapRef = React.useRef<mapboxgl.Map | null>(null);
  const generationMarkerRef = React.useRef<mapboxgl.Marker | null>(null);
  const generationLatitudeRef = React.useRef(generationForm.latitude);
  const generationLongitudeRef = React.useRef(generationForm.longitude);
  const seenCompletedGenerationJobsRef = React.useRef<Set<string>>(new Set());
  const didHydrateDeepLinkedScenarioRef = React.useRef(false);
  const deepLinkedScenarioId = searchParams.get('id')?.trim() ?? '';
  const replaceDeepLinkedScenarioId = useCallback(
    (scenarioId?: string | null) => {
      const normalizedScenarioId = (scenarioId ?? '').trim();
      const currentScenarioId = searchParams.get('id')?.trim() ?? '';
      if (normalizedScenarioId === currentScenarioId) {
        return;
      }
      const next = new URLSearchParams(searchParams);
      if (normalizedScenarioId) {
        next.set('id', normalizedScenarioId);
      } else {
        next.delete('id');
      }
      setSearchParams(next, { replace: true });
    },
    [searchParams, setSearchParams]
  );

  const loadPointsOfInterestForZone = useCallback(
    async (zoneId: string) => {
      const trimmedZoneId = zoneId.trim();
      if (!trimmedZoneId) return;
      if (zonePointOfInterestMap[trimmedZoneId]) return;
      setPointOfInterestLoadingByZone((prev) => ({
        ...prev,
        [trimmedZoneId]: true,
      }));
      try {
        const points = await apiClient.get<PointOfInterest[]>(
          `/sonar/zones/${trimmedZoneId}/pointsOfInterest`
        );
        const mapped = (Array.isArray(points) ? points : [])
          .map((point) => {
            const lat = Number.parseFloat(String(point.lat ?? ''));
            const lng = Number.parseFloat(String(point.lng ?? ''));
            if (!Number.isFinite(lat) || !Number.isFinite(lng)) return null;
            return {
              id: point.id,
              name: point.name || point.id,
              latitude: lat,
              longitude: lng,
            };
          })
          .filter((point): point is PointOfInterestOption => point !== null);
        setZonePointOfInterestMap((prev) => ({
          ...prev,
          [trimmedZoneId]: mapped,
        }));
      } catch (err) {
        console.error(
          `Error loading points of interest for zone ${trimmedZoneId}:`,
          err
        );
      } finally {
        setPointOfInterestLoadingByZone((prev) => ({
          ...prev,
          [trimmedZoneId]: false,
        }));
      }
    },
    [apiClient, zonePointOfInterestMap]
  );

  const load = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.get<PaginatedResponse<ScenarioRecord>>(
        '/sonar/admin/scenarios',
        {
          page,
          pageSize: scenarioListPageSize,
          query: deferredQuery.trim(),
          zoneQuery: deferredZoneQuery.trim(),
          genreId: genreFilter === 'all' ? '' : genreFilter,
        }
      );
      setRecords(Array.isArray(response?.items) ? response.items : []);
      setTotal(response?.total ?? 0);
    } catch (err) {
      console.error('Error loading scenarios:', err);
      setError('Failed to load scenarios.');
    } finally {
      setLoading(false);
    }
  }, [apiClient, deferredQuery, deferredZoneQuery, genreFilter, page]);

  const refreshScenarioById = useCallback(
    async (scenarioId: string) => {
      const latest = await apiClient.get<ScenarioRecord>(
        `/sonar/scenarios/${scenarioId}`
      );
      setRecords((prev) =>
        prev.map((record) => (record.id === scenarioId ? latest : record))
      );
      return latest;
    },
    [apiClient]
  );

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    setPage(1);
  }, [genreFilter, query, zoneQuery]);

  useEffect(() => {
    const totalPages = Math.max(1, Math.ceil(total / scenarioListPageSize));
    if (page > totalPages) {
      setPage(totalPages);
    }
  }, [page, total]);

  useEffect(() => {
    let active = true;
    const loadSpells = async () => {
      try {
        const response = await apiClient.get<Spell[]>('/sonar/spells');
        if (!active) return;
        setSpells(Array.isArray(response) ? response : []);
      } catch (err) {
        console.error('Error loading spells:', err);
        if (active) setSpells([]);
      }
    };
    void loadSpells();
    return () => {
      active = false;
    };
  }, [apiClient]);

  useEffect(() => {
    let active = true;
    const loadGenres = async () => {
      try {
        const response = await apiClient.get<ZoneGenre[]>(
          '/sonar/zone-genres?includeInactive=true'
        );
        if (!active) return;
        setGenres(Array.isArray(response) ? response : []);
      } catch (err) {
        console.error('Error loading scenario genres:', err);
        if (active) setGenres([]);
      }
    };
    void loadGenres();
    return () => {
      active = false;
    };
  }, [apiClient]);

  useEffect(() => {
    let active = true;
    const loadZoneKinds = async () => {
      try {
        const response = await apiClient.get<ZoneKind[]>('/sonar/zoneKinds');
        if (!active) return;
        setZoneKinds(Array.isArray(response) ? response : []);
      } catch (err) {
        console.error('Error loading zone kinds:', err);
        if (active) setZoneKinds([]);
      }
    };
    void loadZoneKinds();
    return () => {
      active = false;
    };
  }, [apiClient]);

  const loadGenerationJobs = useCallback(async () => {
    try {
      setGenerationJobsLoading(true);
      const response = await apiClient.get<ScenarioGenerationJob[]>(
        '/sonar/admin/scenario-generation-jobs',
        {
          limit: 25,
        }
      );
      setGenerationJobs(response);
      setGenerationError(null);
    } catch (err) {
      console.error('Error loading scenario generation jobs:', err);
      setGenerationError('Failed to load scenario generation jobs.');
    } finally {
      setGenerationJobsLoading(false);
    }
  }, [apiClient]);

  useEffect(() => {
    void loadGenerationJobs();
  }, [loadGenerationJobs]);

  useEffect(() => {
    if (!generationForm.zoneId && zones.length > 0) {
      setGenerationForm((prev) => ({ ...prev, zoneId: zones[0].id }));
    }
  }, [generationForm.zoneId, zones]);

  useEffect(() => {
    if (!generationForm.genreId && defaultGenreId) {
      setGenerationForm((prev) =>
        prev.genreId ? prev : { ...prev, genreId: defaultGenreId }
      );
    }
  }, [defaultGenreId, generationForm.genreId]);

  const selectedGenerationZone = useMemo(
    () => zones.find((zone) => zone.id === generationForm.zoneId),
    [generationForm.zoneId, zones]
  );

  useEffect(() => {
    if (selectedGenerationZone?.name) {
      setGenerationZoneQuery(selectedGenerationZone.name);
      return;
    }
    if (!generationForm.zoneId) {
      setGenerationZoneQuery('');
    }
  }, [generationForm.zoneId, selectedGenerationZone]);

  useEffect(() => {
    generationLatitudeRef.current = generationForm.latitude;
    generationLongitudeRef.current = generationForm.longitude;
  }, [generationForm.latitude, generationForm.longitude]);

  const hasActiveGenerationJobs = useMemo(
    () =>
      generationJobs.some(
        (job) => job.status === 'queued' || job.status === 'in_progress'
      ),
    [generationJobs]
  );

  useEffect(() => {
    if (!hasActiveGenerationJobs) return;
    const interval = window.setInterval(() => {
      void loadGenerationJobs();
    }, 3000);
    return () => {
      window.clearInterval(interval);
    };
  }, [hasActiveGenerationJobs, loadGenerationJobs]);

  useEffect(() => {
    const completedWithScenario = generationJobs.filter(
      (job) => job.status === 'completed' && !!job.generatedScenarioId
    );
    let shouldReloadScenarios = false;
    for (const job of completedWithScenario) {
      if (!seenCompletedGenerationJobsRef.current.has(job.id)) {
        seenCompletedGenerationJobsRef.current.add(job.id);
        shouldReloadScenarios = true;
      }
    }
    if (shouldReloadScenarios) {
      void load();
    }
  }, [generationJobs, load]);

  useEffect(() => {
    setSelectedScenarioIds((prev) => {
      if (prev.size === 0) return prev;
      const validIDs = new Set(records.map((record) => record.id));
      const next = new Set<string>();
      prev.forEach((id) => {
        if (validIDs.has(id)) {
          next.add(id);
        }
      });
      return next.size === prev.size ? prev : next;
    });
  }, [records]);

  const setGenerationLocation = useCallback(
    (latitude: number, longitude: number) => {
      setGenerationForm((prev) => ({
        ...prev,
        includeLocation: true,
        latitude: latitude.toFixed(6),
        longitude: longitude.toFixed(6),
      }));
    },
    []
  );

  const handleQueueScenarioGeneration = async () => {
    if (!generationForm.zoneId) {
      setGenerationError('Please select a zone.');
      return;
    }
    const payload: {
      zoneId: string;
      genreId: string;
      openEnded: boolean;
      latitude?: number;
      longitude?: number;
    } = {
      zoneId: generationForm.zoneId,
      genreId: generationForm.genreId.trim(),
      openEnded: generationForm.openEnded,
    };

    if (generationForm.includeLocation) {
      const latitude = parseFloatValue(generationForm.latitude, Number.NaN);
      const longitude = parseFloatValue(generationForm.longitude, Number.NaN);
      if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
        setGenerationError(
          'Location is enabled, so latitude and longitude are required.'
        );
        return;
      }
      payload.latitude = latitude;
      payload.longitude = longitude;
    }

    try {
      setGenerationSubmitting(true);
      setGenerationError(null);
      const created = await apiClient.post<ScenarioGenerationJob>(
        '/sonar/admin/scenario-generation-jobs',
        payload
      );
      setGenerationJobs((prev) => [created, ...prev]);
      setGenerationForm((prev) => ({
        ...prev,
        includeLocation: false,
        latitude: '',
        longitude: '',
      }));
    } catch (err) {
      console.error('Error queueing scenario generation job:', err);
      setGenerationError('Failed to queue scenario generation job.');
    } finally {
      setGenerationSubmitting(false);
    }
  };

  const handleUseCurrentGenerationLocation = useCallback(() => {
    if (!navigator.geolocation) {
      setGenerationError('Geolocation is not supported in this browser.');
      return;
    }
    setGenerationGeoLoading(true);
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setGenerationGeoLoading(false);
        setGenerationError(null);
        setGenerationLocation(
          position.coords.latitude,
          position.coords.longitude
        );
      },
      (geoError) => {
        setGenerationGeoLoading(false);
        setGenerationError(
          `Unable to get current location: ${geoError.message}`
        );
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 30000 }
    );
  }, [setGenerationLocation]);

  const refreshUndiscoveredScenarioIconStatus = useCallback(
    async (showMessage = false) => {
      try {
        setScenarioUndiscoveredStatusLoading(true);
        setScenarioUndiscoveredError(null);
        const response = await apiClient.get<StaticThumbnailResponse>(
          '/sonar/admin/thumbnails/scenario-undiscovered/status'
        );
        const url = (response?.thumbnailUrl || '').trim();
        if (url) {
          setScenarioUndiscoveredUrl(url);
        }
        setScenarioUndiscoveredStatus(
          (response?.status || 'unknown').trim() || 'unknown'
        );
        setScenarioUndiscoveredExists(Boolean(response?.exists));
        setScenarioUndiscoveredRequestedAt(
          response?.requestedAt ? response.requestedAt : null
        );
        setScenarioUndiscoveredLastModified(
          response?.lastModified ? response.lastModified : null
        );
        setScenarioUndiscoveredPreviewNonce(Date.now());
        if (showMessage) {
          setScenarioUndiscoveredMessage(
            'Undiscovered scenario icon status refreshed.'
          );
        }
      } catch (err) {
        console.error('Failed to load undiscovered scenario icon status', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to load undiscovered scenario icon status.';
        setScenarioUndiscoveredError(message);
      } finally {
        setScenarioUndiscoveredStatusLoading(false);
      }
    },
    [apiClient]
  );

  const handleGenerateUndiscoveredScenarioIcon = useCallback(async () => {
    const prompt = scenarioUndiscoveredPrompt.trim();
    if (!prompt) {
      setScenarioUndiscoveredError('Prompt is required.');
      return;
    }
    try {
      setScenarioUndiscoveredBusy(true);
      setScenarioUndiscoveredError(null);
      setScenarioUndiscoveredMessage(null);
      await apiClient.post<StaticThumbnailResponse>(
        '/sonar/admin/thumbnails/scenario-undiscovered',
        { prompt }
      );
      setScenarioUndiscoveredMessage(
        'Undiscovered scenario icon queued for generation.'
      );
      await refreshUndiscoveredScenarioIconStatus();
    } catch (err) {
      console.error('Failed to generate undiscovered scenario icon', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to generate undiscovered scenario icon.';
      setScenarioUndiscoveredError(message);
    } finally {
      setScenarioUndiscoveredBusy(false);
    }
  }, [
    apiClient,
    refreshUndiscoveredScenarioIconStatus,
    scenarioUndiscoveredPrompt,
  ]);

  const handleDeleteUndiscoveredScenarioIcon = useCallback(async () => {
    try {
      setScenarioUndiscoveredBusy(true);
      setScenarioUndiscoveredError(null);
      setScenarioUndiscoveredMessage(null);
      await apiClient.delete<StaticThumbnailResponse>(
        '/sonar/admin/thumbnails/scenario-undiscovered'
      );
      setScenarioUndiscoveredMessage('Undiscovered scenario icon deleted.');
      await refreshUndiscoveredScenarioIconStatus();
    } catch (err) {
      console.error('Failed to delete undiscovered scenario icon', err);
      const message =
        err instanceof Error
          ? err.message
          : 'Failed to delete undiscovered scenario icon.';
      setScenarioUndiscoveredError(message);
    } finally {
      setScenarioUndiscoveredBusy(false);
    }
  }, [apiClient, refreshUndiscoveredScenarioIconStatus]);

  useEffect(() => {
    void refreshUndiscoveredScenarioIconStatus();
  }, [refreshUndiscoveredScenarioIconStatus]);

  useEffect(() => {
    if (
      scenarioUndiscoveredStatus !== 'queued' &&
      scenarioUndiscoveredStatus !== 'in_progress'
    ) {
      return;
    }
    const interval = window.setInterval(() => {
      void refreshUndiscoveredScenarioIconStatus();
    }, 4000);
    return () => window.clearInterval(interval);
  }, [scenarioUndiscoveredStatus, refreshUndiscoveredScenarioIconStatus]);

  useEffect(() => {
    if (!generationForm.includeLocation) {
      generationMarkerRef.current?.remove();
      generationMarkerRef.current = null;
      generationMapRef.current?.remove();
      generationMapRef.current = null;
      return;
    }
    if (!generationMapContainerRef.current) return;
    if (!mapboxgl.accessToken) return;
    if (generationMapRef.current) return;

    const parsedLat = Number.parseFloat(generationLatitudeRef.current);
    const parsedLng = Number.parseFloat(generationLongitudeRef.current);
    const selectedZone = zones.find(
      (zone) => zone.id === generationForm.zoneId
    );
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
      container: generationMapContainerRef.current,
      style: 'mapbox://styles/mapbox/streets-v12',
      center,
      zoom: 13,
    });

    map.on('click', (event) => {
      setGenerationLocation(event.lngLat.lat, event.lngLat.lng);
    });

    generationMapRef.current = map;

    return () => {
      generationMarkerRef.current?.remove();
      generationMarkerRef.current = null;
      map.remove();
      generationMapRef.current = null;
    };
  }, [
    generationForm.includeLocation,
    generationForm.zoneId,
    setGenerationLocation,
    zones,
  ]);

  useEffect(() => {
    if (!generationForm.includeLocation) return;
    if (!generationMapRef.current) return;

    const latitude = Number.parseFloat(generationForm.latitude);
    const longitude = Number.parseFloat(generationForm.longitude);
    if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
      generationMarkerRef.current?.remove();
      generationMarkerRef.current = null;
      return;
    }

    if (!generationMarkerRef.current) {
      generationMarkerRef.current = new mapboxgl.Marker({ color: '#2563eb' })
        .setLngLat([longitude, latitude])
        .addTo(generationMapRef.current);
    } else {
      generationMarkerRef.current.setLngLat([longitude, latitude]);
    }

    generationMapRef.current.easeTo({
      center: [longitude, latitude],
      duration: 350,
    });
  }, [
    generationForm.includeLocation,
    generationForm.latitude,
    generationForm.longitude,
  ]);

  const filtered = records;
  const selectedScenarioIdSet = useMemo(
    () => selectedScenarioIds,
    [selectedScenarioIds]
  );
  const allFilteredScenariosSelected = useMemo(() => {
    if (filtered.length === 0) return false;
    return filtered.every((record) => selectedScenarioIds.has(record.id));
  }, [filtered, selectedScenarioIds]);

  const allPointOfInterestNamesById = useMemo(() => {
    const byId = new Map<string, string>();
    Object.values(zonePointOfInterestMap).forEach((points) => {
      points.forEach((point) => {
        if (!byId.has(point.id)) {
          byId.set(point.id, point.name);
        }
      });
    });
    return byId;
  }, [zonePointOfInterestMap]);

  const pointsOfInterestForFormZone = useMemo(() => {
    return zonePointOfInterestMap[form.zoneId] ?? [];
  }, [form.zoneId, zonePointOfInterestMap]);
  const selectedFormZone = useMemo(
    () => zones.find((zone) => zone.id === form.zoneId) ?? null,
    [form.zoneId, zones]
  );
  const selectedFormZoneDefaultKind = selectedFormZone?.kind?.trim() ?? '';
  const effectiveFormZoneKind =
    form.zoneKind.trim() || selectedFormZoneDefaultKind;
  const effectiveFormZoneKindDetails = effectiveFormZoneKind
    ? (zoneKindBySlug.get(effectiveFormZoneKind) ?? null)
    : null;
  const hasSelectedPointOfInterest = form.pointOfInterestId.trim().length > 0;

  useEffect(() => {
    if (!showModal) return;
    if (!form.zoneId) return;
    void loadPointsOfInterestForZone(form.zoneId);
  }, [form.zoneId, loadPointsOfInterestForZone, showModal]);

  useEffect(() => {
    if (!showModal) return;
    if (!form.genreId && defaultGenreId) {
      setForm((prev) =>
        prev.genreId ? prev : { ...prev, genreId: defaultGenreId }
      );
    }
  }, [defaultGenreId, form.genreId, showModal]);

  useEffect(() => {
    if (!records.length) return;
    const zoneIds = Array.from(new Set(records.map((record) => record.zoneId)));
    zoneIds.forEach((zoneId) => {
      if (zoneId && !zonePointOfInterestMap[zoneId]) {
        void loadPointsOfInterestForZone(zoneId);
      }
    });
  }, [loadPointsOfInterestForZone, records, zonePointOfInterestMap]);

  const openCreate = () => {
    setEditingId(null);
    setForm(emptyFormState());
    setShowModal(true);
  };

  const openEdit = (record: ScenarioRecord) => {
    setEditingId(record.id);
    setForm({
      zoneId: record.zoneId,
      zoneKind: record.zoneKind?.trim() ?? '',
      genreId: record.genreId ?? record.genre?.id ?? '',
      pointOfInterestId: record.pointOfInterestId ?? '',
      latitude: record.latitude.toString(),
      longitude: record.longitude.toString(),
      prompt: record.prompt,
      internalTagsInput: (record.internalTags ?? []).join(', '),
      imageUrl: record.imageUrl,
      thumbnailUrl: record.thumbnailUrl ?? '',
      difficulty: record.difficulty.toString(),
      scaleWithUserLevel: Boolean(record.scaleWithUserLevel),
      recurrenceFrequency: record.recurrenceFrequency ?? '',
      openEnded: record.openEnded,
      rewardMode: record.rewardMode === 'explicit' ? 'explicit' : 'random',
      randomRewardSize:
        record.randomRewardSize === 'medium' ||
        record.randomRewardSize === 'large'
          ? record.randomRewardSize
          : 'small',
      rewardExperience: record.rewardExperience.toString(),
      rewardGold: record.rewardGold.toString(),
      materialRewards: (record.materialRewards ?? []).map((reward) => ({
        resourceKey: reward.resourceKey,
        amount: reward.amount ?? 1,
      })),
      successHandoffText: record.successHandoffText?.trim() ?? '',
      failureHandoffText: record.failureHandoffText?.trim() ?? '',
      failurePenaltyMode: normalizeFailurePenaltyMode(
        record.failurePenaltyMode
      ),
      failureHealthDrainType: normalizeFailureDrainType(
        record.failureHealthDrainType
      ),
      failureHealthDrainValue: String(record.failureHealthDrainValue ?? 0),
      failureManaDrainType: normalizeFailureDrainType(
        record.failureManaDrainType
      ),
      failureManaDrainValue: String(record.failureManaDrainValue ?? 0),
      failureStatuses: (record.failureStatuses ?? []).map((status) =>
        normalizeFailureStatus(status)
      ),
      successRewardMode: normalizeSuccessRewardMode(record.successRewardMode),
      successHealthRestoreType: normalizeFailureDrainType(
        record.successHealthRestoreType
      ),
      successHealthRestoreValue: String(record.successHealthRestoreValue ?? 0),
      successManaRestoreType: normalizeFailureDrainType(
        record.successManaRestoreType
      ),
      successManaRestoreValue: String(record.successManaRestoreValue ?? 0),
      successStatuses: (record.successStatuses ?? []).map((status) =>
        normalizeFailureStatus(status)
      ),
      options:
        record.options.length > 0
          ? record.options.map((option) => ({
              ...option,
              materialRewards: (option.materialRewards ?? []).map((reward) => ({
                resourceKey: reward.resourceKey,
                amount: reward.amount ?? 1,
              })),
              successText:
                option.successText?.trim() ||
                'Your approach works, and momentum turns in your favor.',
              failureText:
                option.failureText?.trim() ||
                'The attempt falls short, and the moment slips away.',
              successHandoffText: option.successHandoffText?.trim() ?? '',
              failureHandoffText: option.failureHandoffText?.trim() ?? '',
              failureHealthDrainType: normalizeFailureDrainType(
                option.failureHealthDrainType
              ),
              failureHealthDrainValue: option.failureHealthDrainValue ?? 0,
              failureManaDrainType: normalizeFailureDrainType(
                option.failureManaDrainType
              ),
              failureManaDrainValue: option.failureManaDrainValue ?? 0,
              failureStatuses: (option.failureStatuses ?? []).map((status) =>
                normalizeFailureStatus(status)
              ),
              successHealthRestoreType: normalizeFailureDrainType(
                option.successHealthRestoreType
              ),
              successHealthRestoreValue: option.successHealthRestoreValue ?? 0,
              successManaRestoreType: normalizeFailureDrainType(
                option.successManaRestoreType
              ),
              successManaRestoreValue: option.successManaRestoreValue ?? 0,
              successStatuses: (option.successStatuses ?? []).map((status) =>
                normalizeFailureStatus(status)
              ),
              spellRewards: (option.spellRewards ?? []).map((reward) => ({
                spellId: reward.spellId,
              })),
            }))
          : [emptyOption()],
      itemRewards: record.itemRewards,
      spellRewards: (record.spellRewards ?? []).map((reward) => ({
        spellId: reward.spellId,
      })),
    });
    setShowModal(true);
  };

  useEffect(() => {
    if (didHydrateDeepLinkedScenarioRef.current) {
      return;
    }
    if (!deepLinkedScenarioId) {
      didHydrateDeepLinkedScenarioRef.current = true;
      return;
    }
    if (editingId === deepLinkedScenarioId && showModal) {
      didHydrateDeepLinkedScenarioRef.current = true;
      return;
    }
    void apiClient
      .get<ScenarioRecord>(`/sonar/scenarios/${deepLinkedScenarioId}`)
      .then((record) => {
        if (!record) {
          didHydrateDeepLinkedScenarioRef.current = true;
          return;
        }
        setRecords((prev) => {
          const withoutExisting = prev.filter(
            (entry) => entry.id !== record.id
          );
          return [record, ...withoutExisting];
        });
        openEdit(record);
      })
      .catch((error) => {
        console.error('Failed to deep link scenario', error);
        didHydrateDeepLinkedScenarioRef.current = true;
      });
  }, [apiClient, deepLinkedScenarioId, editingId, showModal]);

  useEffect(() => {
    if (!didHydrateDeepLinkedScenarioRef.current) {
      return;
    }
    replaceDeepLinkedScenarioId(showModal ? editingId : null);
  }, [editingId, replaceDeepLinkedScenarioId, showModal]);

  const closeModal = () => {
    setShowModal(false);
    setEditingId(null);
    setForm(emptyFormState());
  };

  useEffect(() => {
    formLatitudeRef.current = form.latitude;
    formLongitudeRef.current = form.longitude;
  }, [form.latitude, form.longitude]);

  const setFormLocation = useCallback((latitude: number, longitude: number) => {
    setForm((prev) => ({
      ...prev,
      pointOfInterestId: '',
      latitude: latitude.toFixed(6),
      longitude: longitude.toFixed(6),
    }));
  }, []);

  const formPayload = () => {
    const toStatusPayload = (statuses: ScenarioFailureStatus[]) =>
      statuses
        .map((status) => ({
          ...status,
          name: status.name.trim(),
          description: status.description.trim(),
          effect: status.effect.trim(),
          durationSeconds: parseIntValue(String(status.durationSeconds), 0),
        }))
        .filter((status) => status.name !== '' && status.durationSeconds > 0);

    const scenarioPenaltyMode: ScenarioFailurePenaltyMode = form.openEnded
      ? 'shared'
      : form.failurePenaltyMode;
    const useOptionPenalties =
      !form.openEnded && scenarioPenaltyMode === 'individual';
    const successRewardMode: ScenarioSuccessRewardMode = form.openEnded
      ? 'shared'
      : form.successRewardMode;
    const useOptionSuccessRewards =
      !form.openEnded && successRewardMode === 'individual';

    return {
      zoneId: form.zoneId,
      zoneKind: form.zoneKind.trim(),
      genreId: form.genreId.trim(),
      pointOfInterestId: form.pointOfInterestId.trim(),
      latitude: parseFloatValue(form.latitude),
      longitude: parseFloatValue(form.longitude),
      prompt: form.prompt.trim(),
      internalTags: parseInternalTagsInput(form.internalTagsInput),
      imageUrl: form.imageUrl.trim(),
      thumbnailUrl: form.thumbnailUrl.trim(),
      difficulty: parseIntValue(form.difficulty, 24),
      scaleWithUserLevel: form.scaleWithUserLevel,
      recurrenceFrequency: form.recurrenceFrequency,
      openEnded: form.openEnded,
      rewardMode: form.rewardMode,
      randomRewardSize: form.randomRewardSize,
      rewardExperience:
        form.openEnded && form.rewardMode === 'explicit'
          ? parseIntValue(form.rewardExperience)
          : 0,
      rewardGold:
        form.openEnded && form.rewardMode === 'explicit'
          ? parseIntValue(form.rewardGold)
          : 0,
      materialRewards:
        form.openEnded && form.rewardMode === 'explicit'
          ? form.materialRewards
          : [],
      successHandoffText: form.successHandoffText.trim(),
      failureHandoffText: form.failureHandoffText.trim(),
      failurePenaltyMode: scenarioPenaltyMode,
      successRewardMode,
      failureHealthDrainType: form.failureHealthDrainType,
      failureHealthDrainValue:
        form.failureHealthDrainType === 'none'
          ? 0
          : parseIntValue(form.failureHealthDrainValue, 0),
      failureManaDrainType: form.failureManaDrainType,
      failureManaDrainValue:
        form.failureManaDrainType === 'none'
          ? 0
          : parseIntValue(form.failureManaDrainValue, 0),
      failureStatuses: toStatusPayload(form.failureStatuses),
      successHealthRestoreType: form.successHealthRestoreType,
      successHealthRestoreValue:
        form.successHealthRestoreType === 'none'
          ? 0
          : parseIntValue(form.successHealthRestoreValue, 0),
      successManaRestoreType: form.successManaRestoreType,
      successManaRestoreValue:
        form.successManaRestoreType === 'none'
          ? 0
          : parseIntValue(form.successManaRestoreValue, 0),
      successStatuses: toStatusPayload(form.successStatuses),
      options: form.openEnded
        ? []
        : form.options.map((option) => ({
            optionText: option.optionText.trim(),
            successText: option.successText.trim(),
            failureText: option.failureText.trim(),
            successHandoffText: option.successHandoffText.trim(),
            failureHandoffText: option.failureHandoffText.trim(),
            statTag: option.statTag,
            proficiencies: option.proficiencies,
            difficulty: option.difficulty,
            rewardExperience: option.rewardExperience,
            rewardGold: option.rewardGold,
            materialRewards: option.materialRewards,
            failureHealthDrainType: useOptionPenalties
              ? option.failureHealthDrainType
              : 'none',
            failureHealthDrainValue:
              useOptionPenalties && option.failureHealthDrainType !== 'none'
                ? option.failureHealthDrainValue
                : 0,
            failureManaDrainType: useOptionPenalties
              ? option.failureManaDrainType
              : 'none',
            failureManaDrainValue:
              useOptionPenalties && option.failureManaDrainType !== 'none'
                ? option.failureManaDrainValue
                : 0,
            failureStatuses: useOptionPenalties
              ? toStatusPayload(option.failureStatuses)
              : [],
            successHealthRestoreType: useOptionSuccessRewards
              ? option.successHealthRestoreType
              : 'none',
            successHealthRestoreValue:
              useOptionSuccessRewards &&
              option.successHealthRestoreType !== 'none'
                ? option.successHealthRestoreValue
                : 0,
            successManaRestoreType: useOptionSuccessRewards
              ? option.successManaRestoreType
              : 'none',
            successManaRestoreValue:
              useOptionSuccessRewards &&
              option.successManaRestoreType !== 'none'
                ? option.successManaRestoreValue
                : 0,
            successStatuses: useOptionSuccessRewards
              ? toStatusPayload(option.successStatuses)
              : [],
            itemRewards: option.itemRewards,
            spellRewards: option.spellRewards
              .map((reward) => ({ spellId: reward.spellId.trim() }))
              .filter((reward) => reward.spellId.length > 0),
          })),
      itemRewards:
        form.rewardMode === 'random' || form.openEnded ? form.itemRewards : [],
      spellRewards:
        form.openEnded && form.rewardMode === 'explicit'
          ? form.spellRewards
              .map((reward) => ({ spellId: reward.spellId.trim() }))
              .filter((reward) => reward.spellId.length > 0)
          : [],
    };
  };

  const save = async () => {
    try {
      const payload = formPayload();
      if (
        !payload.zoneId ||
        !payload.prompt ||
        !payload.imageUrl ||
        !payload.thumbnailUrl
      ) {
        alert('Zone, prompt, image URL, and thumbnail URL are required.');
        return;
      }
      if (payload.openEnded === false && payload.options.length === 0) {
        alert('Non-open-ended scenarios need at least one option.');
        return;
      }
      const hasPointOfInterest = payload.pointOfInterestId.length > 0;
      if (!hasPointOfInterest) {
        const latitude = Number.parseFloat(form.latitude);
        const longitude = Number.parseFloat(form.longitude);
        if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
          alert(
            'Choose a point of interest or provide valid latitude/longitude.'
          );
          return;
        }
      }

      if (editingId) {
        await apiClient.put<ScenarioRecord>(
          `/sonar/scenarios/${editingId}`,
          payload
        );
      } else {
        await apiClient.post<ScenarioRecord>('/sonar/scenarios', payload);
      }
      await load();
      closeModal();
    } catch (err) {
      console.error('Error saving scenario:', err);
      alert('Failed to save scenario. Check required fields and try again.');
    }
  };

  const confirmDelete = async () => {
    if (!deleteId) return;
    try {
      await apiClient.delete(`/sonar/scenarios/${deleteId}`);
      setSelectedScenarioIds((prev) => {
        if (!prev.has(deleteId)) return prev;
        const next = new Set(prev);
        next.delete(deleteId);
        return next;
      });
      setDeleteId(null);
      await load();
    } catch (err) {
      console.error('Error deleting scenario:', err);
      alert('Failed to delete scenario.');
    }
  };

  const toggleScenarioSelection = (scenarioId: string) => {
    setSelectedScenarioIds((prev) => {
      const next = new Set(prev);
      if (next.has(scenarioId)) {
        next.delete(scenarioId);
      } else {
        next.add(scenarioId);
      }
      return next;
    });
  };

  const toggleSelectVisibleScenarios = () => {
    if (filtered.length === 0) return;
    setSelectedScenarioIds((prev) => {
      const next = new Set(prev);
      if (allFilteredScenariosSelected) {
        filtered.forEach((record) => next.delete(record.id));
      } else {
        filtered.forEach((record) => next.add(record.id));
      }
      return next;
    });
  };

  const clearScenarioSelection = () => {
    setSelectedScenarioIds(new Set());
  };

  const handleBulkDeleteScenarios = async () => {
    if (
      bulkDeletingScenarios ||
      selectedScenarioIds.size === 0 ||
      deleteId !== null
    ) {
      return;
    }

    const selectedIds = Array.from(selectedScenarioIds);
    const selectedNames = records
      .filter((record) => selectedScenarioIds.has(record.id))
      .map((record) => record.prompt);
    const preview = selectedNames.slice(0, 5).join(', ');
    const moreCount = Math.max(0, selectedNames.length - 5);
    const confirmMessage =
      selectedIds.length === 1
        ? `Delete 1 selected scenario (${preview})? This cannot be undone.`
        : `Delete ${selectedIds.length} selected scenarios${
            preview
              ? ` (${preview}${moreCount > 0 ? ` +${moreCount} more` : ''})`
              : ''
          }? This cannot be undone.`;

    if (!window.confirm(confirmMessage)) return;

    setBulkDeletingScenarios(true);
    try {
      await apiClient.post('/sonar/scenarios/bulk-delete', {
        ids: selectedIds,
      });
      const deletedIds = new Set(selectedIds);
      setSelectedScenarioIds(new Set());
      if (editingId && deletedIds.has(editingId)) {
        closeModal();
      }
      await load();
    } catch (err) {
      console.error('Failed to bulk delete scenarios', err);
      alert('Failed to delete selected scenarios.');
    } finally {
      setBulkDeletingScenarios(false);
    }
  };

  const handleGenerateScenarioImage = async (record: ScenarioRecord) => {
    if (generatingScenarioId) return;
    setGeneratingScenarioId(record.id);
    const previousImageURL = (record.imageUrl || '').trim();
    try {
      await apiClient.post(`/sonar/scenarios/${record.id}/generate-image`, {});

      for (let attempt = 0; attempt < 18; attempt += 1) {
        await new Promise((resolve) => window.setTimeout(resolve, 1200));
        const latest = await refreshScenarioById(record.id);
        const nextImageURL = (latest.imageUrl || '').trim();
        if (nextImageURL && nextImageURL !== previousImageURL) {
          break;
        }
      }
    } catch (err) {
      console.error('Error generating scenario image:', err);
      alert('Failed to queue scenario image generation.');
    } finally {
      setGeneratingScenarioId(null);
    }
  };

  const handleUseCurrentLocation = useCallback(() => {
    if (!navigator.geolocation) {
      alert('Geolocation is not supported in this browser.');
      return;
    }
    setGeoLoading(true);
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setGeoLoading(false);
        setFormLocation(position.coords.latitude, position.coords.longitude);
      },
      (error) => {
        setGeoLoading(false);
        alert(`Unable to get current location: ${error.message}`);
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 30000 }
    );
  }, [setFormLocation]);

  useEffect(() => {
    if (!showModal) return;
    if (hasSelectedPointOfInterest) return;
    if (!mapContainerRef.current) return;
    if (!mapboxgl.accessToken) return;
    if (mapRef.current) return;

    const parsedLat = Number.parseFloat(formLatitudeRef.current);
    const parsedLng = Number.parseFloat(formLongitudeRef.current);
    const selectedZone = zones.find((zone) => zone.id === form.zoneId);
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
      setFormLocation(event.lngLat.lat, event.lngLat.lng);
    });

    mapRef.current = map;

    return () => {
      markerRef.current?.remove();
      markerRef.current = null;
      map.remove();
      mapRef.current = null;
    };
  }, [
    form.zoneId,
    hasSelectedPointOfInterest,
    setFormLocation,
    showModal,
    zones,
  ]);

  useEffect(() => {
    if (!showModal) return;
    if (hasSelectedPointOfInterest) return;
    if (!mapRef.current) return;

    const lat = Number.parseFloat(form.latitude);
    const lng = Number.parseFloat(form.longitude);
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
  }, [form.latitude, form.longitude, hasSelectedPointOfInterest, showModal]);

  const updateOption = (index: number, next: Partial<ScenarioOption>) => {
    setForm((prev) => {
      const options = [...prev.options];
      options[index] = { ...options[index], ...next };
      return { ...prev, options };
    });
  };

  const addOption = () => {
    setForm((prev) => ({ ...prev, options: [...prev.options, emptyOption()] }));
  };

  const removeOption = (index: number) => {
    setForm((prev) => ({
      ...prev,
      options: prev.options.filter((_, i) => i !== index),
    }));
  };

  const addOptionItemReward = (optionIndex: number) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.itemRewards = [
        ...option.itemRewards,
        { inventoryItemId: 0, quantity: 1 },
      ];
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const updateOptionItemReward = (
    optionIndex: number,
    rewardIndex: number,
    next: Partial<ScenarioRewardItem>
  ) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      const rewards = [...option.itemRewards];
      rewards[rewardIndex] = { ...rewards[rewardIndex], ...next };
      option.itemRewards = rewards;
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const removeOptionItemReward = (optionIndex: number, rewardIndex: number) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.itemRewards = option.itemRewards.filter(
        (_, i) => i !== rewardIndex
      );
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const addOptionSpellReward = (optionIndex: number) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.spellRewards = [...option.spellRewards, { spellId: '' }];
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const updateOptionSpellReward = (
    optionIndex: number,
    rewardIndex: number,
    next: Partial<ScenarioRewardSpell>
  ) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      const rewards = [...option.spellRewards];
      rewards[rewardIndex] = { ...rewards[rewardIndex], ...next };
      option.spellRewards = rewards;
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const removeOptionSpellReward = (
    optionIndex: number,
    rewardIndex: number
  ) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.spellRewards = option.spellRewards.filter(
        (_, i) => i !== rewardIndex
      );
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const addScenarioItemReward = () => {
    setForm((prev) => ({
      ...prev,
      itemRewards: [...prev.itemRewards, { inventoryItemId: 0, quantity: 1 }],
    }));
  };

  const updateScenarioItemReward = (
    index: number,
    next: Partial<ScenarioRewardItem>
  ) => {
    setForm((prev) => {
      const rewards = [...prev.itemRewards];
      rewards[index] = { ...rewards[index], ...next };
      return { ...prev, itemRewards: rewards };
    });
  };

  const removeScenarioItemReward = (index: number) => {
    setForm((prev) => ({
      ...prev,
      itemRewards: prev.itemRewards.filter((_, i) => i !== index),
    }));
  };

  const addScenarioSpellReward = () => {
    setForm((prev) => ({
      ...prev,
      spellRewards: [...prev.spellRewards, { spellId: '' }],
    }));
  };

  const updateScenarioSpellReward = (
    index: number,
    next: Partial<ScenarioRewardSpell>
  ) => {
    setForm((prev) => {
      const rewards = [...prev.spellRewards];
      rewards[index] = { ...rewards[index], ...next };
      return { ...prev, spellRewards: rewards };
    });
  };

  const removeScenarioSpellReward = (index: number) => {
    setForm((prev) => ({
      ...prev,
      spellRewards: prev.spellRewards.filter((_, i) => i !== index),
    }));
  };

  const addScenarioFailureStatus = () => {
    setForm((prev) => ({
      ...prev,
      failureStatuses: [...prev.failureStatuses, emptyFailureStatus()],
    }));
  };

  const updateScenarioFailureStatus = (
    index: number,
    next: Partial<ScenarioFailureStatus>
  ) => {
    setForm((prev) => {
      const statuses = [...prev.failureStatuses];
      statuses[index] = { ...statuses[index], ...next };
      return { ...prev, failureStatuses: statuses };
    });
  };

  const removeScenarioFailureStatus = (index: number) => {
    setForm((prev) => ({
      ...prev,
      failureStatuses: prev.failureStatuses.filter((_, i) => i !== index),
    }));
  };

  const addOptionFailureStatus = (optionIndex: number) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.failureStatuses = [
        ...option.failureStatuses,
        emptyFailureStatus(),
      ];
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const updateOptionFailureStatus = (
    optionIndex: number,
    statusIndex: number,
    next: Partial<ScenarioFailureStatus>
  ) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      const statuses = [...option.failureStatuses];
      statuses[statusIndex] = { ...statuses[statusIndex], ...next };
      option.failureStatuses = statuses;
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const removeOptionFailureStatus = (
    optionIndex: number,
    statusIndex: number
  ) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.failureStatuses = option.failureStatuses.filter(
        (_, i) => i !== statusIndex
      );
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const addScenarioSuccessStatus = () => {
    setForm((prev) => ({
      ...prev,
      successStatuses: [...prev.successStatuses, emptyFailureStatus()],
    }));
  };

  const updateScenarioSuccessStatus = (
    index: number,
    next: Partial<ScenarioFailureStatus>
  ) => {
    setForm((prev) => {
      const statuses = [...prev.successStatuses];
      statuses[index] = { ...statuses[index], ...next };
      return { ...prev, successStatuses: statuses };
    });
  };

  const removeScenarioSuccessStatus = (index: number) => {
    setForm((prev) => ({
      ...prev,
      successStatuses: prev.successStatuses.filter((_, i) => i !== index),
    }));
  };

  const addOptionSuccessStatus = (optionIndex: number) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.successStatuses = [
        ...option.successStatuses,
        emptyFailureStatus(),
      ];
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const updateOptionSuccessStatus = (
    optionIndex: number,
    statusIndex: number,
    next: Partial<ScenarioFailureStatus>
  ) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      const statuses = [...option.successStatuses];
      statuses[statusIndex] = { ...statuses[statusIndex], ...next };
      option.successStatuses = statuses;
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const removeOptionSuccessStatus = (
    optionIndex: number,
    statusIndex: number
  ) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.successStatuses = option.successStatuses.filter(
        (_, i) => i !== statusIndex
      );
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const renderFailurePenaltyEditor = (config: {
    title: string;
    healthDrainType: ScenarioFailureDrainType;
    healthDrainValue: number;
    manaDrainType: ScenarioFailureDrainType;
    manaDrainValue: number;
    statuses: ScenarioFailureStatus[];
    onHealthDrainTypeChange: (value: ScenarioFailureDrainType) => void;
    onHealthDrainValueChange: (value: number) => void;
    onManaDrainTypeChange: (value: ScenarioFailureDrainType) => void;
    onManaDrainValueChange: (value: number) => void;
    onAddStatus: () => void;
    onUpdateStatus: (
      index: number,
      next: Partial<ScenarioFailureStatus>
    ) => void;
    onRemoveStatus: (index: number) => void;
  }) => (
    <div className="border rounded-md p-3 mt-3">
      <div className="font-medium mb-2">{config.title}</div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-3 mb-3">
        <label className="text-sm">
          Health Drain Type
          <select
            value={config.healthDrainType}
            onChange={(e) =>
              config.onHealthDrainTypeChange(
                e.target.value as ScenarioFailureDrainType
              )
            }
            className="w-full border rounded-md p-2"
          >
            {failureDrainTypes.map((type) => (
              <option key={type} value={type}>
                {type}
              </option>
            ))}
          </select>
        </label>
        <label className="text-sm">
          Health Drain Value
          <input
            value={config.healthDrainValue}
            onChange={(e) =>
              config.onHealthDrainValueChange(parseIntValue(e.target.value, 0))
            }
            className="w-full border rounded-md p-2"
            type="number"
            min={0}
            max={config.healthDrainType === 'percent' ? 100 : undefined}
            disabled={config.healthDrainType === 'none'}
          />
        </label>
        <label className="text-sm">
          Mana Drain Type
          <select
            value={config.manaDrainType}
            onChange={(e) =>
              config.onManaDrainTypeChange(
                e.target.value as ScenarioFailureDrainType
              )
            }
            className="w-full border rounded-md p-2"
          >
            {failureDrainTypes.map((type) => (
              <option key={type} value={type}>
                {type}
              </option>
            ))}
          </select>
        </label>
        <label className="text-sm">
          Mana Drain Value
          <input
            value={config.manaDrainValue}
            onChange={(e) =>
              config.onManaDrainValueChange(parseIntValue(e.target.value, 0))
            }
            className="w-full border rounded-md p-2"
            type="number"
            min={0}
            max={config.manaDrainType === 'percent' ? 100 : undefined}
            disabled={config.manaDrainType === 'none'}
          />
        </label>
      </div>

      <div className="flex items-center justify-between mb-2">
        <div className="font-medium text-sm">Failure Statuses</div>
        <button
          type="button"
          className="bg-green-600 text-white px-2 py-1 rounded-md text-xs"
          onClick={config.onAddStatus}
        >
          Add Status
        </button>
      </div>

      {config.statuses.length === 0 && (
        <div className="text-sm text-gray-600">
          No failure statuses configured.
        </div>
      )}

      {config.statuses.map((status, statusIndex) => (
        <div key={statusIndex} className="border rounded-md p-3 mb-2">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-2 mb-2">
            <label className="text-sm">
              Name
              <input
                value={status.name}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, { name: e.target.value })
                }
                className="w-full border rounded-md p-2"
              />
            </label>
            <label className="text-sm">
              Duration (seconds)
              <input
                value={status.durationSeconds}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    durationSeconds: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-2"
                type="number"
                min={1}
              />
            </label>
            <label className="text-sm">
              Status Effect Type
              <select
                value={status.effectType}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    effectType: e.target.value,
                  })
                }
                className="w-full border rounded-md p-2"
              >
                {statusEffectTypes.map((effectType) => (
                  <option key={effectType} value={effectType}>
                    {effectType}
                  </option>
                ))}
              </select>
            </label>
            <label className="text-sm md:col-span-2">
              Description
              <input
                value={status.description}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    description: e.target.value,
                  })
                }
                className="w-full border rounded-md p-2"
              />
            </label>
            <label className="text-sm md:col-span-2">
              Effect
              <input
                value={status.effect}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, { effect: e.target.value })
                }
                className="w-full border rounded-md p-2"
              />
            </label>
          </div>

          <label className="inline-flex items-center gap-2 text-sm mb-2">
            <input
              type="checkbox"
              checked={status.positive}
              onChange={(e) =>
                config.onUpdateStatus(statusIndex, {
                  positive: e.target.checked,
                })
              }
            />
            Positive status
          </label>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-2 mb-2">
            <label className="text-sm">
              Damage Per Tick
              <input
                value={status.damagePerTick}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    damagePerTick: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-2"
                type="number"
              />
            </label>
            <label className="text-sm">
              Health Per Tick
              <input
                value={status.healthPerTick}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    healthPerTick: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-2"
                type="number"
              />
            </label>
            <label className="text-sm">
              Mana Per Tick
              <input
                value={status.manaPerTick}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    manaPerTick: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-2"
                type="number"
              />
            </label>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-6 gap-2 mb-2">
            <label className="text-xs">
              STR
              <input
                value={status.strengthMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    strengthMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              DEX
              <input
                value={status.dexterityMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    dexterityMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              CON
              <input
                value={status.constitutionMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    constitutionMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              INT
              <input
                value={status.intelligenceMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    intelligenceMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              WIS
              <input
                value={status.wisdomMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    wisdomMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              CHA
              <input
                value={status.charismaMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    charismaMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-2 mb-2">
            {damageBonusFieldOptions.map(({ key, label }) => (
              <label className="text-xs" key={key}>
                {label}
                <input
                  value={status[key]}
                  onChange={(e) =>
                    config.onUpdateStatus(statusIndex, {
                      [key]: parseIntValue(e.target.value, 0),
                    })
                  }
                  className="w-full border rounded-md p-1"
                  type="number"
                />
              </label>
            ))}
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-2 mb-2">
            {resistanceFieldOptions.map(({ key, label }) => (
              <label className="text-xs" key={key}>
                {label}
                <input
                  value={status[key]}
                  onChange={(e) =>
                    config.onUpdateStatus(statusIndex, {
                      [key]: parseIntValue(e.target.value, 0),
                    })
                  }
                  className="w-full border rounded-md p-1"
                  type="number"
                />
              </label>
            ))}
          </div>

          <button
            type="button"
            className="bg-red-500 text-white px-3 py-1 rounded-md text-xs"
            onClick={() => config.onRemoveStatus(statusIndex)}
          >
            Remove Status
          </button>
        </div>
      ))}
    </div>
  );

  const renderSuccessRewardEditor = (config: {
    title: string;
    healthRestoreType: ScenarioFailureDrainType;
    healthRestoreValue: number;
    manaRestoreType: ScenarioFailureDrainType;
    manaRestoreValue: number;
    statuses: ScenarioFailureStatus[];
    onHealthRestoreTypeChange: (value: ScenarioFailureDrainType) => void;
    onHealthRestoreValueChange: (value: number) => void;
    onManaRestoreTypeChange: (value: ScenarioFailureDrainType) => void;
    onManaRestoreValueChange: (value: number) => void;
    onAddStatus: () => void;
    onUpdateStatus: (
      index: number,
      next: Partial<ScenarioFailureStatus>
    ) => void;
    onRemoveStatus: (index: number) => void;
  }) => (
    <div className="border rounded-md p-3 mt-3">
      <div className="font-medium mb-2">{config.title}</div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-3 mb-3">
        <label className="text-sm">
          Health Restore Type
          <select
            value={config.healthRestoreType}
            onChange={(e) =>
              config.onHealthRestoreTypeChange(
                e.target.value as ScenarioFailureDrainType
              )
            }
            className="w-full border rounded-md p-2"
          >
            {failureDrainTypes.map((type) => (
              <option key={type} value={type}>
                {type}
              </option>
            ))}
          </select>
        </label>
        <label className="text-sm">
          Health Restore Value
          <input
            value={config.healthRestoreValue}
            onChange={(e) =>
              config.onHealthRestoreValueChange(
                parseIntValue(e.target.value, 0)
              )
            }
            className="w-full border rounded-md p-2"
            type="number"
            min={0}
            max={config.healthRestoreType === 'percent' ? 100 : undefined}
            disabled={config.healthRestoreType === 'none'}
          />
        </label>
        <label className="text-sm">
          Mana Restore Type
          <select
            value={config.manaRestoreType}
            onChange={(e) =>
              config.onManaRestoreTypeChange(
                e.target.value as ScenarioFailureDrainType
              )
            }
            className="w-full border rounded-md p-2"
          >
            {failureDrainTypes.map((type) => (
              <option key={type} value={type}>
                {type}
              </option>
            ))}
          </select>
        </label>
        <label className="text-sm">
          Mana Restore Value
          <input
            value={config.manaRestoreValue}
            onChange={(e) =>
              config.onManaRestoreValueChange(parseIntValue(e.target.value, 0))
            }
            className="w-full border rounded-md p-2"
            type="number"
            min={0}
            max={config.manaRestoreType === 'percent' ? 100 : undefined}
            disabled={config.manaRestoreType === 'none'}
          />
        </label>
      </div>

      <div className="flex items-center justify-between mb-2">
        <div className="font-medium text-sm">Success Statuses</div>
        <button
          type="button"
          className="bg-green-600 text-white px-2 py-1 rounded-md text-xs"
          onClick={config.onAddStatus}
        >
          Add Status
        </button>
      </div>

      {config.statuses.length === 0 && (
        <div className="text-sm text-gray-600">
          No success statuses configured.
        </div>
      )}

      {config.statuses.map((status, statusIndex) => (
        <div key={statusIndex} className="border rounded-md p-3 mb-2">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-2 mb-2">
            <label className="text-sm">
              Name
              <input
                value={status.name}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, { name: e.target.value })
                }
                className="w-full border rounded-md p-2"
              />
            </label>
            <label className="text-sm">
              Duration (seconds)
              <input
                value={status.durationSeconds}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    durationSeconds: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-2"
                type="number"
                min={1}
              />
            </label>
            <label className="text-sm">
              Status Effect Type
              <select
                value={status.effectType}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    effectType: e.target.value,
                  })
                }
                className="w-full border rounded-md p-2"
              >
                {statusEffectTypes.map((effectType) => (
                  <option key={effectType} value={effectType}>
                    {effectType}
                  </option>
                ))}
              </select>
            </label>
            <label className="text-sm md:col-span-2">
              Description
              <input
                value={status.description}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    description: e.target.value,
                  })
                }
                className="w-full border rounded-md p-2"
              />
            </label>
            <label className="text-sm md:col-span-2">
              Effect
              <input
                value={status.effect}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, { effect: e.target.value })
                }
                className="w-full border rounded-md p-2"
              />
            </label>
          </div>

          <label className="inline-flex items-center gap-2 text-sm mb-2">
            <input
              type="checkbox"
              checked={status.positive}
              onChange={(e) =>
                config.onUpdateStatus(statusIndex, {
                  positive: e.target.checked,
                })
              }
            />
            Positive status
          </label>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-2 mb-2">
            <label className="text-sm">
              Damage Per Tick
              <input
                value={status.damagePerTick}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    damagePerTick: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-2"
                type="number"
              />
            </label>
            <label className="text-sm">
              Health Per Tick
              <input
                value={status.healthPerTick}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    healthPerTick: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-2"
                type="number"
              />
            </label>
            <label className="text-sm">
              Mana Per Tick
              <input
                value={status.manaPerTick}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    manaPerTick: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-2"
                type="number"
              />
            </label>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-6 gap-2 mb-2">
            <label className="text-xs">
              STR
              <input
                value={status.strengthMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    strengthMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              DEX
              <input
                value={status.dexterityMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    dexterityMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              CON
              <input
                value={status.constitutionMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    constitutionMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              INT
              <input
                value={status.intelligenceMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    intelligenceMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              WIS
              <input
                value={status.wisdomMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    wisdomMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
            <label className="text-xs">
              CHA
              <input
                value={status.charismaMod}
                onChange={(e) =>
                  config.onUpdateStatus(statusIndex, {
                    charismaMod: parseIntValue(e.target.value, 0),
                  })
                }
                className="w-full border rounded-md p-1"
                type="number"
              />
            </label>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-2 mb-2">
            {damageBonusFieldOptions.map(({ key, label }) => (
              <label className="text-xs" key={key}>
                {label}
                <input
                  value={status[key]}
                  onChange={(e) =>
                    config.onUpdateStatus(statusIndex, {
                      [key]: parseIntValue(e.target.value, 0),
                    })
                  }
                  className="w-full border rounded-md p-1"
                  type="number"
                />
              </label>
            ))}
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-2 mb-2">
            {resistanceFieldOptions.map(({ key, label }) => (
              <label className="text-xs" key={key}>
                {label}
                <input
                  value={status[key]}
                  onChange={(e) =>
                    config.onUpdateStatus(statusIndex, {
                      [key]: parseIntValue(e.target.value, 0),
                    })
                  }
                  className="w-full border rounded-md p-1"
                  type="number"
                />
              </label>
            ))}
          </div>

          <button
            type="button"
            className="bg-red-500 text-white px-3 py-1 rounded-md text-xs"
            onClick={() => config.onRemoveStatus(statusIndex)}
          >
            Remove Status
          </button>
        </div>
      ))}
    </div>
  );

  if (loading) {
    return <div className="m-10">Loading scenarios...</div>;
  }

  return (
    <div className="m-10">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Scenarios</h1>
        <button
          className="bg-blue-500 text-white px-4 py-2 rounded-md"
          onClick={openCreate}
        >
          Create Scenario
        </button>
      </div>

      <div className="mb-6">
        <ContentDashboard
          title="Scenario Dashboard"
          subtitle="Aggregate scenario template coverage for the current search and genre filters."
          status={
            zoneQuery.trim()
              ? 'Template counts reflect search and genre filters; zone search still applies to the scenario list below'
              : query.trim() || genreFilter !== 'all'
                ? 'Reflects current template-oriented search and genre filters'
                : 'All reusable scenario templates'
          }
          loading={dashboardLoading}
          error={dashboardError}
          metrics={dashboardMetrics}
          sections={dashboardSections}
        />
      </div>

      <div className="mb-6 border rounded-md p-4 bg-white shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-2 mb-3">
          <h2 className="text-lg font-semibold">Undiscovered Scenario Icon</h2>
          <div className="flex gap-2">
            <button
              type="button"
              className="bg-gray-700 text-white px-3 py-1 rounded-md disabled:opacity-60"
              onClick={() => void refreshUndiscoveredScenarioIconStatus(true)}
              disabled={scenarioUndiscoveredStatusLoading}
            >
              {scenarioUndiscoveredStatusLoading
                ? 'Refreshing…'
                : 'Refresh Status'}
            </button>
            <button
              type="button"
              className="bg-indigo-600 text-white px-3 py-1 rounded-md disabled:opacity-60"
              onClick={handleGenerateUndiscoveredScenarioIcon}
              disabled={
                scenarioUndiscoveredBusy || scenarioUndiscoveredStatusLoading
              }
            >
              {scenarioUndiscoveredBusy ? 'Working…' : 'Generate Icon'}
            </button>
            <button
              type="button"
              className="bg-red-600 text-white px-3 py-1 rounded-md disabled:opacity-60"
              onClick={handleDeleteUndiscoveredScenarioIcon}
              disabled={
                scenarioUndiscoveredBusy || scenarioUndiscoveredStatusLoading
              }
            >
              {scenarioUndiscoveredBusy ? 'Working…' : 'Delete Icon'}
            </button>
          </div>
        </div>
        <div className="mb-2">
          <span
            className={`inline-flex text-white text-xs px-2 py-0.5 rounded ${staticStatusClassName(
              scenarioUndiscoveredStatus
            )}`}
          >
            {scenarioUndiscoveredStatus || 'unknown'}
          </span>
        </div>
        <div className="text-xs text-gray-600 break-all">
          URL: {scenarioUndiscoveredUrl}
        </div>
        <div className="text-xs text-gray-600 mt-1">
          Requested: {formatDate(scenarioUndiscoveredRequestedAt ?? undefined)}
          {' · '}
          Last updated:{' '}
          {formatDate(scenarioUndiscoveredLastModified ?? undefined)}
        </div>
        <label className="block text-sm mt-3">
          Generation Prompt
          <textarea
            className="w-full border rounded-md p-2 mt-1 min-h-[88px]"
            value={scenarioUndiscoveredPrompt}
            onChange={(event) =>
              setScenarioUndiscoveredPrompt(event.target.value)
            }
            placeholder="Prompt used to generate the undiscovered scenario icon."
          />
        </label>
        {scenarioUndiscoveredExists ? (
          <div className="mt-3">
            <img
              src={`${scenarioUndiscoveredUrl}?v=${scenarioUndiscoveredPreviewNonce}`}
              alt="Undiscovered scenario icon preview"
              className="w-24 h-24 object-cover border rounded-md bg-gray-50"
            />
          </div>
        ) : (
          <div className="text-xs text-gray-500 mt-2">
            No icon currently found at this URL.
          </div>
        )}
        {scenarioUndiscoveredMessage ? (
          <div className="text-sm text-emerald-700 mt-2">
            {scenarioUndiscoveredMessage}
          </div>
        ) : null}
        {scenarioUndiscoveredError ? (
          <div className="text-sm text-red-600 mt-2">
            {scenarioUndiscoveredError}
          </div>
        ) : null}
      </div>

      <div className="mb-6 border rounded-md p-4 bg-white shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-2 mb-3">
          <h2 className="text-lg font-semibold">Generate Scenario (Async)</h2>
          <button
            type="button"
            className="bg-gray-700 text-white px-3 py-1 rounded-md disabled:opacity-60"
            onClick={() => void loadGenerationJobs()}
            disabled={generationJobsLoading}
          >
            {generationJobsLoading ? 'Refreshing…' : 'Refresh Jobs'}
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-3 mb-3">
          <label className="text-sm">
            Zone
            <div className="relative">
              <input
                className="w-full border rounded-md p-2"
                value={generationZoneQuery}
                onChange={(event) => {
                  const value = event.target.value;
                  setGenerationZoneQuery(value);
                  setShowGenerationZoneSuggestions(true);
                  if (value.trim() === '') {
                    setGenerationForm((prev) => ({
                      ...prev,
                      zoneId: '',
                    }));
                  }
                }}
                onFocus={() => setShowGenerationZoneSuggestions(true)}
                onBlur={() => {
                  window.setTimeout(
                    () => setShowGenerationZoneSuggestions(false),
                    120
                  );
                }}
                placeholder="Type to filter zones..."
              />
              {showGenerationZoneSuggestions && zones.length > 0 ? (
                <div className="absolute z-20 mt-1 max-h-60 w-full overflow-y-auto rounded border border-gray-200 bg-white shadow">
                  {zones
                    .filter((zone) =>
                      zone.name
                        .toLowerCase()
                        .includes(generationZoneQuery.toLowerCase())
                    )
                    .map((zone) => (
                      <button
                        type="button"
                        key={zone.id}
                        onClick={() => {
                          setGenerationForm((prev) => ({
                            ...prev,
                            zoneId: zone.id,
                          }));
                          setGenerationZoneQuery(zone.name);
                          setShowGenerationZoneSuggestions(false);
                        }}
                        className="block w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                      >
                        {zone.name}
                      </button>
                    ))}
                </div>
              ) : null}
            </div>
          </label>
          <label className="text-sm">
            Genre
            <select
              value={generationForm.genreId}
              onChange={(e) =>
                setGenerationForm((prev) => ({
                  ...prev,
                  genreId: e.target.value,
                }))
              }
              className="w-full border rounded-md p-2"
            >
              {genres.length === 0 ? (
                <option value="">Fantasy</option>
              ) : (
                genres.map((genre) => (
                  <option key={`generation-genre-${genre.id}`} value={genre.id}>
                    {formatGenreLabel(genre)}
                    {genre.active === false ? ' (inactive)' : ''}
                  </option>
                ))
              )}
            </select>
          </label>
          <label className="text-sm">
            Scenario Type
            <select
              value={generationForm.openEnded ? 'open_ended' : 'choice'}
              onChange={(e) =>
                setGenerationForm((prev) => ({
                  ...prev,
                  openEnded: e.target.value === 'open_ended',
                }))
              }
              className="w-full border rounded-md p-2"
            >
              <option value="choice">Choice (Options)</option>
              <option value="open_ended">Open-Ended (Free Text)</option>
            </select>
          </label>
        </div>

        <label className="inline-flex items-center gap-2 mb-3 text-sm">
          <input
            type="checkbox"
            checked={generationForm.includeLocation}
            onChange={(e) =>
              setGenerationForm((prev) => ({
                ...prev,
                includeLocation: e.target.checked,
                latitude: e.target.checked ? prev.latitude : '',
                longitude: e.target.checked ? prev.longitude : '',
              }))
            }
          />
          Provide a specific location (optional)
        </label>

        {generationForm.includeLocation && (
          <div className="border rounded-md p-3 mb-3">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-2">
              <label className="text-sm">
                Latitude
                <input
                  value={generationForm.latitude}
                  onChange={(e) =>
                    setGenerationForm((prev) => ({
                      ...prev,
                      latitude: e.target.value,
                    }))
                  }
                  className="w-full border rounded-md p-2"
                  type="number"
                  step="any"
                />
              </label>
              <label className="text-sm">
                Longitude
                <input
                  value={generationForm.longitude}
                  onChange={(e) =>
                    setGenerationForm((prev) => ({
                      ...prev,
                      longitude: e.target.value,
                    }))
                  }
                  className="w-full border rounded-md p-2"
                  type="number"
                  step="any"
                />
              </label>
            </div>
            <div className="mb-2">
              <button
                type="button"
                className="bg-gray-700 text-white px-3 py-2 rounded-md disabled:opacity-60"
                onClick={handleUseCurrentGenerationLocation}
                disabled={generationGeoLoading}
              >
                {generationGeoLoading
                  ? 'Locating…'
                  : 'Use Current Browser Location'}
              </button>
            </div>
            <div>
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm">Map Location Picker</span>
                <span className="text-xs text-gray-500">
                  Click map to set latitude/longitude
                </span>
              </div>
              {mapboxgl.accessToken ? (
                <div
                  ref={generationMapContainerRef}
                  className="w-full h-56 border rounded-md"
                />
              ) : (
                <div className="w-full border rounded-md p-3 text-sm text-gray-600 bg-gray-50">
                  Missing `REACT_APP_MAPBOX_ACCESS_TOKEN`; map picker is
                  unavailable.
                </div>
              )}
            </div>
          </div>
        )}

        <div className="flex flex-wrap items-center gap-2 mb-3">
          <button
            type="button"
            className="bg-indigo-600 text-white px-4 py-2 rounded-md disabled:opacity-60"
            onClick={handleQueueScenarioGeneration}
            disabled={generationSubmitting}
          >
            {generationSubmitting ? 'Queueing…' : 'Queue Scenario Generation'}
          </button>
        </div>

        {generationError && (
          <div className="mb-3 text-red-600 text-sm">{generationError}</div>
        )}

        <div className="font-medium mb-2">Recent Generation Jobs</div>
        {generationJobsLoading && generationJobs.length === 0 ? (
          <div className="text-sm text-gray-600">Loading jobs...</div>
        ) : generationJobs.length === 0 ? (
          <div className="text-sm text-gray-600">No jobs yet.</div>
        ) : (
          <div className="grid gap-2">
            {generationJobs.map((job) => {
              const zoneName =
                zones.find((zone) => zone.id === job.zoneId)?.name ??
                job.zoneId;
              const record = job.generatedScenarioId
                ? records.find(
                    (scenario) => scenario.id === job.generatedScenarioId
                  )
                : undefined;
              return (
                <div
                  key={job.id}
                  className="border rounded-md p-2 text-sm bg-gray-50"
                >
                  <div className="flex flex-wrap items-center gap-2 mb-1">
                    <span className="font-mono text-xs">{job.id}</span>
                    <span
                      className={`text-white text-xs px-2 py-0.5 rounded ${scenarioGenerationStatusBadgeClass(job.status)}`}
                    >
                      {job.status}
                    </span>
                    <span>{job.openEnded ? 'Open-Ended' : 'Choice'}</span>
                  </div>
                  <div className="text-gray-700">Zone: {zoneName}</div>
                  <div className="text-gray-700">
                    Genre:{' '}
                    {formatGenreLabel(
                      job.genre ??
                        genres.find((genre) => genre.id === job.genreId)
                    )}
                  </div>
                  <div className="text-gray-700">
                    Location:{' '}
                    {typeof job.latitude === 'number' &&
                    typeof job.longitude === 'number'
                      ? `${job.latitude.toFixed(5)}, ${job.longitude.toFixed(5)}`
                      : 'Auto-selected'}
                  </div>
                  <div className="text-gray-600 text-xs">
                    Created: {formatDate(job.createdAt)}
                  </div>
                  {job.generatedScenarioId && (
                    <div className="text-gray-700">
                      Scenario ID:{' '}
                      <span className="font-mono text-xs">
                        {job.generatedScenarioId}
                      </span>
                    </div>
                  )}
                  {job.errorMessage && (
                    <div className="text-red-600 text-xs mt-1">
                      {job.errorMessage}
                    </div>
                  )}
                  {record && (
                    <div className="mt-2">
                      <button
                        type="button"
                        className="bg-blue-500 text-white px-2 py-1 rounded-md text-xs"
                        onClick={() => openEdit(record)}
                      >
                        Open Scenario
                      </button>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>

      {error && <div className="mb-3 text-red-600">{error}</div>}

      <div className="mb-4">
        <div className="flex flex-wrap gap-3">
          <input
            type="text"
            placeholder="Search scenarios by prompt, ID, or internal tag..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="w-full max-w-xl p-2 border rounded-md"
          />
          <input
            type="text"
            placeholder="Search by zone..."
            value={zoneQuery}
            onChange={(e) => setZoneQuery(e.target.value)}
            className="w-full max-w-sm p-2 border rounded-md"
          />
          <select
            value={genreFilter}
            onChange={(e) => setGenreFilter(e.target.value)}
            className="w-full max-w-xs p-2 border rounded-md"
            aria-label="Filter scenarios by genre"
          >
            <option value="all">All Genres</option>
            {genres.map((genre) => (
              <option key={`scenario-filter-${genre.id}`} value={genre.id}>
                {formatGenreLabel(genre)}
              </option>
            ))}
          </select>
        </div>
      </div>
      <div className="mb-4 flex flex-wrap items-center gap-2">
        <button
          type="button"
          className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
          onClick={toggleSelectVisibleScenarios}
          disabled={filtered.length === 0 || bulkDeletingScenarios}
        >
          {allFilteredScenariosSelected ? 'Unselect Visible' : 'Select Visible'}
        </button>
        <button
          type="button"
          className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
          onClick={clearScenarioSelection}
          disabled={selectedScenarioIds.size === 0 || bulkDeletingScenarios}
        >
          Clear Selection
        </button>
        <button
          type="button"
          className="qa-btn qa-btn-danger"
          onClick={handleBulkDeleteScenarios}
          disabled={
            selectedScenarioIds.size === 0 ||
            bulkDeletingScenarios ||
            deleteId !== null
          }
        >
          {bulkDeletingScenarios
            ? `Deleting ${selectedScenarioIds.size}...`
            : `Delete Selected (${selectedScenarioIds.size})`}
        </button>
      </div>

      <div
        className="grid gap-4"
        style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))' }}
      >
        {filtered.map((record) => {
          const zoneName =
            zones.find((zone) => zone.id === record.zoneId)?.name ??
            record.zoneId;
          const zoneDefaultKind =
            zones.find((zone) => zone.id === record.zoneId)?.kind?.trim() ?? '';
          const explicitZoneKind = record.zoneKind?.trim() ?? '';
          const zoneKindSummary = explicitZoneKind
            ? zoneKindLabel(explicitZoneKind, zoneKindBySlug)
            : zoneDefaultKind
              ? `${zoneKindLabel(zoneDefaultKind, zoneKindBySlug)} (zone default)`
              : 'Unassigned';
          return (
            <div
              key={record.id}
              className="border rounded-md p-4 bg-white shadow-sm"
            >
              <div className="flex items-start justify-between gap-2 mb-1">
                <div className="text-xs text-gray-500 break-all">
                  {record.id}
                </div>
                <input
                  type="checkbox"
                  className="h-4 w-4"
                  checked={selectedScenarioIdSet.has(record.id)}
                  disabled={bulkDeletingScenarios}
                  onChange={() => toggleScenarioSelection(record.id)}
                />
              </div>
              <div className="font-semibold mb-2">
                {record.openEnded ? 'Open-Ended' : 'Choice'} Scenario
              </div>
              <div className="text-sm text-gray-700 mb-1">Zone: {zoneName}</div>
              <div className="text-sm text-gray-700 mb-1">
                Zone Kind: {zoneKindSummary}
              </div>
              {explicitZoneKind &&
              zoneDefaultKind &&
              explicitZoneKind !== zoneDefaultKind ? (
                <div className="text-xs text-gray-500 mb-1">
                  Zone default: {zoneKindLabel(zoneDefaultKind, zoneKindBySlug)}
                </div>
              ) : null}
              <div className="text-sm text-gray-700 mb-1">
                Genre:{' '}
                {formatGenreLabel(
                  record.genre ??
                    genres.find((genre) => genre.id === record.genreId)
                )}
              </div>
              <div className="text-sm text-gray-700 mb-1">
                Location:{' '}
                {record.pointOfInterestId
                  ? `POI: ${
                      allPointOfInterestNamesById.get(
                        record.pointOfInterestId
                      ) ?? record.pointOfInterestId
                    }`
                  : `${record.latitude.toFixed(5)}, ${record.longitude.toFixed(5)}`}
              </div>
              <div className="text-sm text-gray-700 mb-2">
                Difficulty: {record.difficulty}
                {record.scaleWithUserLevel ? ' (scales with user level)' : ''}
              </div>
              {record.recurrenceFrequency ? (
                <div className="text-xs text-indigo-700 mb-2">
                  Repeats {record.recurrenceFrequency}
                  {record.nextRecurrenceAt
                    ? ` (next ${formatDate(record.nextRecurrenceAt)})`
                    : ''}
                </div>
              ) : null}
              <div className="text-sm text-gray-800 mb-3 line-clamp-3">
                {record.prompt}
              </div>
              <div className="text-xs text-gray-600 mb-3">
                Internal Tags:{' '}
                {(record.internalTags ?? []).join(', ') || 'None'}
              </div>
              {(record.thumbnailUrl || record.imageUrl) && (
                <button
                  type="button"
                  className="w-full mb-3"
                  onClick={() =>
                    setExpandedScenarioImage({
                      url: record.imageUrl || record.thumbnailUrl,
                      title: record.prompt,
                    })
                  }
                >
                  <img
                    src={record.thumbnailUrl || record.imageUrl}
                    alt="Scenario"
                    className="w-full aspect-square object-cover rounded-md border cursor-zoom-in"
                  />
                </button>
              )}
              <div className="flex gap-2">
                <button
                  className="bg-blue-500 text-white px-3 py-1 rounded-md"
                  onClick={() => openEdit(record)}
                >
                  Edit
                </button>
                <button
                  className="bg-purple-600 text-white px-3 py-1 rounded-md disabled:opacity-60"
                  onClick={() => handleGenerateScenarioImage(record)}
                  disabled={generatingScenarioId === record.id}
                >
                  {generatingScenarioId === record.id
                    ? 'Generating…'
                    : 'Generate Image'}
                </button>
                <button
                  className="bg-red-500 text-white px-3 py-1 rounded-md"
                  onClick={() => setDeleteId(record.id)}
                  disabled={bulkDeletingScenarios}
                >
                  Delete
                </button>
              </div>
            </div>
          );
        })}
      </div>
      <PaginationControls
        page={page}
        pageSize={scenarioListPageSize}
        total={total}
        label="scenarios"
        onPageChange={setPage}
      />

      {showModal && (
        <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-md p-6 w-full max-w-5xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-semibold mb-4">
              {editingId ? 'Edit Scenario' : 'Create Scenario'}
            </h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-4">
              <label className="text-sm">
                Zone
                <select
                  value={form.zoneId}
                  onChange={(e) =>
                    setForm((prev) => ({
                      ...prev,
                      zoneId: e.target.value,
                      pointOfInterestId: '',
                    }))
                  }
                  className="w-full border rounded-md p-2"
                >
                  <option value="">Select zone</option>
                  {zones.map((zone) => (
                    <option key={zone.id} value={zone.id}>
                      {zone.name}
                    </option>
                  ))}
                </select>
              </label>
              <label className="text-sm">
                Zone Kind
                <select
                  value={form.zoneKind}
                  onChange={(e) =>
                    setForm((prev) => ({
                      ...prev,
                      zoneKind: e.target.value,
                    }))
                  }
                  className="w-full border rounded-md p-2"
                >
                  <option value="">
                    {selectedFormZoneDefaultKind
                      ? `Use zone default (${zoneKindLabel(
                          selectedFormZoneDefaultKind,
                          zoneKindBySlug
                        )})`
                      : 'No zone kind'}
                  </option>
                  {zoneKinds.map((zoneKind) => (
                    <option
                      key={`scenario-zone-kind-${zoneKind.id}`}
                      value={zoneKind.slug}
                    >
                      {zoneKind.name}
                    </option>
                  ))}
                </select>
                {effectiveFormZoneKindDetails?.description?.trim() ? (
                  <div className="mt-1 text-xs text-gray-500">
                    {form.zoneKind.trim()
                      ? effectiveFormZoneKindDetails.description
                      : `Using zone default: ${effectiveFormZoneKindDetails.description}`}
                  </div>
                ) : null}
              </label>
              <label className="text-sm">
                Genre
                <select
                  value={form.genreId}
                  onChange={(e) =>
                    setForm((prev) => ({
                      ...prev,
                      genreId: e.target.value,
                    }))
                  }
                  className="w-full border rounded-md p-2"
                >
                  {genres.length === 0 ? (
                    <option value="">Fantasy</option>
                  ) : (
                    genres.map((genre) => (
                      <option
                        key={`scenario-genre-${genre.id}`}
                        value={genre.id}
                      >
                        {formatGenreLabel(genre)}
                        {genre.active === false ? ' (inactive)' : ''}
                      </option>
                    ))
                  )}
                </select>
              </label>
              <label className="text-sm">
                Difficulty
                <input
                  value={form.difficulty}
                  onChange={(e) =>
                    setForm((prev) => ({ ...prev, difficulty: e.target.value }))
                  }
                  className="w-full border rounded-md p-2"
                  type="number"
                  min={0}
                />
              </label>
              <label className="text-sm flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={form.scaleWithUserLevel}
                  onChange={(e) =>
                    setForm((prev) => ({
                      ...prev,
                      scaleWithUserLevel: e.target.checked,
                    }))
                  }
                />
                Scale difficulty with user level
              </label>
              <label className="text-sm">
                Recurrence
                <select
                  value={form.recurrenceFrequency}
                  onChange={(e) =>
                    setForm((prev) => ({
                      ...prev,
                      recurrenceFrequency: e.target.value,
                    }))
                  }
                  className="w-full border rounded-md p-2"
                >
                  {recurrenceOptions.map((option) => (
                    <option key={option.value || 'none'} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </label>
              <label className="text-sm md:col-span-2">
                Point of Interest (Optional)
                <select
                  value={form.pointOfInterestId}
                  onChange={(e) => {
                    const nextPointOfInterestId = e.target.value;
                    if (!nextPointOfInterestId) {
                      setForm((prev) => ({ ...prev, pointOfInterestId: '' }));
                      return;
                    }
                    const selectedPoint = pointsOfInterestForFormZone.find(
                      (point) => point.id === nextPointOfInterestId
                    );
                    setForm((prev) => ({
                      ...prev,
                      pointOfInterestId: nextPointOfInterestId,
                      latitude:
                        selectedPoint?.latitude.toFixed(6) ?? prev.latitude,
                      longitude:
                        selectedPoint?.longitude.toFixed(6) ?? prev.longitude,
                    }));
                  }}
                  className="w-full border rounded-md p-2"
                >
                  <option value="">Use standalone coordinates</option>
                  {pointsOfInterestForFormZone.map((point) => (
                    <option key={point.id} value={point.id}>
                      {point.name}
                    </option>
                  ))}
                </select>
                {pointOfInterestLoadingByZone[form.zoneId] ? (
                  <div className="text-xs text-gray-500 mt-1">
                    Loading points of interest...
                  </div>
                ) : null}
              </label>
              <label className="text-sm">
                Latitude
                <input
                  value={form.latitude}
                  onChange={(e) =>
                    setForm((prev) => ({
                      ...prev,
                      pointOfInterestId: '',
                      latitude: e.target.value,
                    }))
                  }
                  className="w-full border rounded-md p-2"
                  type="number"
                  step="any"
                  disabled={hasSelectedPointOfInterest}
                />
              </label>
              <label className="text-sm">
                Longitude
                <input
                  value={form.longitude}
                  onChange={(e) =>
                    setForm((prev) => ({
                      ...prev,
                      pointOfInterestId: '',
                      longitude: e.target.value,
                    }))
                  }
                  className="w-full border rounded-md p-2"
                  type="number"
                  step="any"
                  disabled={hasSelectedPointOfInterest}
                />
              </label>
              <div className="text-sm md:col-span-2">
                <button
                  type="button"
                  className="bg-gray-700 text-white px-3 py-2 rounded-md disabled:opacity-60"
                  onClick={handleUseCurrentLocation}
                  disabled={geoLoading || hasSelectedPointOfInterest}
                >
                  {geoLoading ? 'Locating…' : 'Use Current Browser Location'}
                </button>
                {hasSelectedPointOfInterest ? (
                  <div className="text-xs text-gray-500 mt-1">
                    Clear point of interest selection to set manual coordinates.
                  </div>
                ) : null}
              </div>
              {!hasSelectedPointOfInterest && (
                <div className="text-sm md:col-span-2">
                  <div className="flex items-center justify-between mb-1">
                    <span>Map Location Picker</span>
                    <span className="text-xs text-gray-500">
                      Click map to set latitude/longitude
                    </span>
                  </div>
                  {mapboxgl.accessToken ? (
                    <div
                      ref={mapContainerRef}
                      className="w-full h-64 border rounded-md"
                    />
                  ) : (
                    <div className="w-full border rounded-md p-3 text-sm text-gray-600 bg-gray-50">
                      Missing `REACT_APP_MAPBOX_ACCESS_TOKEN`; map picker is
                      unavailable.
                    </div>
                  )}
                </div>
              )}
              <label className="text-sm md:col-span-2">
                Image URL
                <input
                  value={form.imageUrl}
                  onChange={(e) =>
                    setForm((prev) => ({ ...prev, imageUrl: e.target.value }))
                  }
                  className="w-full border rounded-md p-2"
                />
              </label>
              <label className="text-sm md:col-span-2">
                Thumbnail URL
                <input
                  value={form.thumbnailUrl}
                  onChange={(e) =>
                    setForm((prev) => ({
                      ...prev,
                      thumbnailUrl: e.target.value,
                    }))
                  }
                  className="w-full border rounded-md p-2"
                />
              </label>
            </div>

            <label className="text-sm block mb-4">
              Prompt
              <textarea
                value={form.prompt}
                onChange={(e) =>
                  setForm((prev) => ({ ...prev, prompt: e.target.value }))
                }
                className="w-full border rounded-md p-2 min-h-[90px]"
              />
            </label>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-4">
              <label className="text-sm">
                Success Handoff
                <textarea
                  value={form.successHandoffText}
                  onChange={(e) =>
                    setForm((prev) => ({
                      ...prev,
                      successHandoffText: e.target.value,
                    }))
                  }
                  className="w-full border rounded-md p-2 min-h-[72px]"
                />
                <span className="mt-1 block text-xs text-gray-500">
                  Optional story bridge for what this success points to next.
                </span>
              </label>
              <label className="text-sm">
                Failure Handoff
                <textarea
                  value={form.failureHandoffText}
                  onChange={(e) =>
                    setForm((prev) => ({
                      ...prev,
                      failureHandoffText: e.target.value,
                    }))
                  }
                  className="w-full border rounded-md p-2 min-h-[72px]"
                />
                <span className="mt-1 block text-xs text-gray-500">
                  Optional story bridge for how the quest thread continues after
                  failure.
                </span>
              </label>
            </div>

            <label className="text-sm block mb-4">
              Internal Tags
              <input
                value={form.internalTagsInput}
                onChange={(e) =>
                  setForm((prev) => ({
                    ...prev,
                    internalTagsInput: e.target.value,
                  }))
                }
                placeholder="tutorial, boss_intro, social_check"
                className="w-full border rounded-md p-2"
              />
              <span className="mt-1 block text-xs text-gray-500">
                Internal-only metadata tags used for admin classification and
                tooling.
              </span>
            </label>

            <label className="inline-flex items-center gap-2 mb-4">
              <input
                type="checkbox"
                checked={form.openEnded}
                onChange={(e) =>
                  setForm((prev) => ({
                    ...prev,
                    openEnded: e.target.checked,
                    failurePenaltyMode: e.target.checked
                      ? 'shared'
                      : prev.failurePenaltyMode,
                    successRewardMode: e.target.checked
                      ? 'shared'
                      : prev.successRewardMode,
                    options: e.target.checked
                      ? prev.options
                      : prev.options.length > 0
                        ? prev.options
                        : [emptyOption()],
                  }))
                }
              />
              Open-ended scenario (freeform response)
            </label>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-4">
              <label className="text-sm">
                Reward Mode
                <select
                  value={form.rewardMode}
                  onChange={(e) =>
                    setForm((prev) => ({
                      ...prev,
                      rewardMode: e.target.value as 'explicit' | 'random',
                    }))
                  }
                  className="w-full border rounded-md p-2"
                >
                  <option value="random">Random Reward</option>
                  <option value="explicit">Explicit Reward</option>
                </select>
              </label>
              <label className="text-sm">
                Random Reward Size
                <select
                  value={form.randomRewardSize}
                  disabled={form.rewardMode !== 'random'}
                  onChange={(e) =>
                    setForm((prev) => ({
                      ...prev,
                      randomRewardSize: e.target.value as
                        | 'small'
                        | 'medium'
                        | 'large',
                    }))
                  }
                  className="w-full border rounded-md p-2"
                >
                  <option value="small">Small</option>
                  <option value="medium">Medium</option>
                  <option value="large">Large</option>
                </select>
              </label>
            </div>
            {form.rewardMode === 'random' ? (
              <div className="text-xs text-gray-500 mb-4">
                Random rewards still grant scaled XP and gold. Guaranteed items
                below are awarded too; material and spell rewards stay
                explicit-only.
              </div>
            ) : null}

            {!form.openEnded && form.rewardMode === 'random' ? (
              <div className="border rounded-md p-3 mb-4">
                <div className="flex items-center justify-between mb-2">
                  <div className="font-medium">Guaranteed Item Rewards</div>
                  <button
                    className="bg-green-600 text-white px-3 py-1 rounded-md"
                    type="button"
                    onClick={addScenarioItemReward}
                  >
                    Add Item
                  </button>
                </div>
                {form.itemRewards.length === 0 ? (
                  <div className="text-sm text-gray-500 mb-2">
                    No guaranteed item rewards yet.
                  </div>
                ) : null}
                {form.itemRewards.map((reward, index) => (
                  <div
                    key={`shared-random-item-${index}`}
                    className="grid grid-cols-1 md:grid-cols-3 gap-2 mb-2"
                  >
                    <select
                      value={reward.inventoryItemId}
                      onChange={(e) =>
                        updateScenarioItemReward(index, {
                          inventoryItemId: parseIntValue(e.target.value, 0),
                        })
                      }
                      className="border rounded-md p-2"
                    >
                      <option value={0}>Select item</option>
                      {inventoryItems.map((item) => (
                        <option key={item.id} value={item.id}>
                          {item.name}
                        </option>
                      ))}
                    </select>
                    <input
                      value={reward.quantity}
                      onChange={(e) =>
                        updateScenarioItemReward(index, {
                          quantity: parseIntValue(e.target.value, 1),
                        })
                      }
                      className="border rounded-md p-2"
                      type="number"
                      min={1}
                    />
                    <button
                      type="button"
                      className="bg-red-500 text-white px-3 py-1 rounded-md"
                      onClick={() => removeScenarioItemReward(index)}
                    >
                      Remove
                    </button>
                  </div>
                ))}
              </div>
            ) : null}

            {form.openEnded ? (
              <div className="border rounded-md p-3 mb-4">
                <div className="font-medium mb-2">Scenario Rewards</div>
                {form.openEnded ? (
                  <>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-3">
                      <label className="text-sm">
                        Reward Experience
                        <input
                          value={form.rewardExperience}
                          disabled={form.rewardMode !== 'explicit'}
                          onChange={(e) =>
                            setForm((prev) => ({
                              ...prev,
                              rewardExperience: e.target.value,
                            }))
                          }
                          className="w-full border rounded-md p-2"
                          type="number"
                          min={0}
                        />
                      </label>
                      <label className="text-sm">
                        Reward Gold
                        <input
                          value={form.rewardGold}
                          disabled={form.rewardMode !== 'explicit'}
                          onChange={(e) =>
                            setForm((prev) => ({
                              ...prev,
                              rewardGold: e.target.value,
                            }))
                          }
                          className="w-full border rounded-md p-2"
                          type="number"
                          min={0}
                        />
                      </label>
                    </div>
                    <div className="mb-4">
                      <MaterialRewardsEditor
                        value={form.materialRewards}
                        onChange={(materialRewards) =>
                          setForm((prev) => ({ ...prev, materialRewards }))
                        }
                        disabled={form.rewardMode !== 'explicit'}
                      />
                    </div>
                  </>
                ) : null}
                <div className="flex items-center justify-between mb-2">
                  <div className="font-medium">Guaranteed Item Rewards</div>
                  <button
                    className="bg-green-600 text-white px-3 py-1 rounded-md"
                    type="button"
                    onClick={addScenarioItemReward}
                  >
                    Add Item
                  </button>
                </div>
                {form.itemRewards.map((reward, index) => (
                  <div
                    key={index}
                    className="grid grid-cols-1 md:grid-cols-3 gap-2 mb-2"
                  >
                    <select
                      value={reward.inventoryItemId}
                      onChange={(e) =>
                        updateScenarioItemReward(index, {
                          inventoryItemId: parseIntValue(e.target.value, 0),
                        })
                      }
                      className="border rounded-md p-2"
                    >
                      <option value={0}>Select item</option>
                      {inventoryItems.map((item) => (
                        <option key={item.id} value={item.id}>
                          {item.name}
                        </option>
                      ))}
                    </select>
                    <input
                      value={reward.quantity}
                      onChange={(e) =>
                        updateScenarioItemReward(index, {
                          quantity: parseIntValue(e.target.value, 1),
                        })
                      }
                      className="border rounded-md p-2"
                      type="number"
                      min={1}
                    />
                    <button
                      type="button"
                      className="bg-red-500 text-white px-3 py-1 rounded-md"
                      onClick={() => removeScenarioItemReward(index)}
                    >
                      Remove
                    </button>
                  </div>
                ))}

                <div className="flex items-center justify-between mb-2 mt-4">
                  <div className="font-medium">Spell Rewards</div>
                  <button
                    className="bg-green-600 text-white px-3 py-1 rounded-md"
                    type="button"
                    disabled={form.rewardMode !== 'explicit'}
                    onClick={addScenarioSpellReward}
                  >
                    Add Spell
                  </button>
                </div>
                {form.spellRewards.map((reward, index) => (
                  <div
                    key={`scenario-spell-${index}`}
                    className="grid grid-cols-1 md:grid-cols-2 gap-2 mb-2"
                  >
                    <select
                      value={reward.spellId}
                      disabled={form.rewardMode !== 'explicit'}
                      onChange={(e) =>
                        updateScenarioSpellReward(index, {
                          spellId: e.target.value,
                        })
                      }
                      className="border rounded-md p-2"
                    >
                      <option value="">Select spell</option>
                      {spells.map((spell) => (
                        <option key={spell.id} value={spell.id}>
                          {spell.name}
                        </option>
                      ))}
                    </select>
                    <button
                      type="button"
                      className="bg-red-500 text-white px-3 py-1 rounded-md"
                      disabled={form.rewardMode !== 'explicit'}
                      onClick={() => removeScenarioSpellReward(index)}
                    >
                      Remove
                    </button>
                  </div>
                ))}

                {renderFailurePenaltyEditor({
                  title: 'Failure Penalty (Open-Ended)',
                  healthDrainType: form.failureHealthDrainType,
                  healthDrainValue: parseIntValue(
                    form.failureHealthDrainValue,
                    0
                  ),
                  manaDrainType: form.failureManaDrainType,
                  manaDrainValue: parseIntValue(form.failureManaDrainValue, 0),
                  onHealthDrainTypeChange: (value) =>
                    setForm((prev) => ({
                      ...prev,
                      failureHealthDrainType: value,
                    })),
                  onHealthDrainValueChange: (value) =>
                    setForm((prev) => ({
                      ...prev,
                      failureHealthDrainValue: String(value),
                    })),
                  onManaDrainTypeChange: (value) =>
                    setForm((prev) => ({
                      ...prev,
                      failureManaDrainType: value,
                    })),
                  onManaDrainValueChange: (value) =>
                    setForm((prev) => ({
                      ...prev,
                      failureManaDrainValue: String(value),
                    })),
                  statuses: form.failureStatuses,
                  onAddStatus: addScenarioFailureStatus,
                  onUpdateStatus: updateScenarioFailureStatus,
                  onRemoveStatus: removeScenarioFailureStatus,
                })}

                {renderSuccessRewardEditor({
                  title: 'Success Rewards (Open-Ended)',
                  healthRestoreType: form.successHealthRestoreType,
                  healthRestoreValue: parseIntValue(
                    form.successHealthRestoreValue,
                    0
                  ),
                  manaRestoreType: form.successManaRestoreType,
                  manaRestoreValue: parseIntValue(
                    form.successManaRestoreValue,
                    0
                  ),
                  onHealthRestoreTypeChange: (value) =>
                    setForm((prev) => ({
                      ...prev,
                      successHealthRestoreType: value,
                    })),
                  onHealthRestoreValueChange: (value) =>
                    setForm((prev) => ({
                      ...prev,
                      successHealthRestoreValue: String(value),
                    })),
                  onManaRestoreTypeChange: (value) =>
                    setForm((prev) => ({
                      ...prev,
                      successManaRestoreType: value,
                    })),
                  onManaRestoreValueChange: (value) =>
                    setForm((prev) => ({
                      ...prev,
                      successManaRestoreValue: String(value),
                    })),
                  statuses: form.successStatuses,
                  onAddStatus: addScenarioSuccessStatus,
                  onUpdateStatus: updateScenarioSuccessStatus,
                  onRemoveStatus: removeScenarioSuccessStatus,
                })}
              </div>
            ) : (
              <div className="border rounded-md p-3 mb-4">
                <div className="flex items-center justify-between mb-3">
                  <div className="font-medium">Response Options</div>
                  <button
                    className="bg-green-600 text-white px-3 py-1 rounded-md"
                    type="button"
                    onClick={addOption}
                  >
                    Add Option
                  </button>
                </div>

                <label className="text-sm block mb-3">
                  Failure Penalty Mode
                  <select
                    value={form.failurePenaltyMode}
                    onChange={(e) =>
                      setForm((prev) => ({
                        ...prev,
                        failurePenaltyMode: e.target
                          .value as ScenarioFailurePenaltyMode,
                      }))
                    }
                    className="w-full border rounded-md p-2"
                  >
                    <option value="shared">Shared (scenario-level)</option>
                    <option value="individual">Individual (per option)</option>
                  </select>
                </label>

                {form.failurePenaltyMode === 'shared' &&
                  renderFailurePenaltyEditor({
                    title: 'Shared Failure Penalty',
                    healthDrainType: form.failureHealthDrainType,
                    healthDrainValue: parseIntValue(
                      form.failureHealthDrainValue,
                      0
                    ),
                    manaDrainType: form.failureManaDrainType,
                    manaDrainValue: parseIntValue(
                      form.failureManaDrainValue,
                      0
                    ),
                    onHealthDrainTypeChange: (value) =>
                      setForm((prev) => ({
                        ...prev,
                        failureHealthDrainType: value,
                      })),
                    onHealthDrainValueChange: (value) =>
                      setForm((prev) => ({
                        ...prev,
                        failureHealthDrainValue: String(value),
                      })),
                    onManaDrainTypeChange: (value) =>
                      setForm((prev) => ({
                        ...prev,
                        failureManaDrainType: value,
                      })),
                    onManaDrainValueChange: (value) =>
                      setForm((prev) => ({
                        ...prev,
                        failureManaDrainValue: String(value),
                      })),
                    statuses: form.failureStatuses,
                    onAddStatus: addScenarioFailureStatus,
                    onUpdateStatus: updateScenarioFailureStatus,
                    onRemoveStatus: removeScenarioFailureStatus,
                  })}

                <label className="text-sm block mb-3">
                  Success Reward Mode
                  <select
                    value={form.successRewardMode}
                    onChange={(e) =>
                      setForm((prev) => ({
                        ...prev,
                        successRewardMode: e.target
                          .value as ScenarioSuccessRewardMode,
                      }))
                    }
                    className="w-full border rounded-md p-2"
                  >
                    <option value="shared">Shared (scenario-level)</option>
                    <option value="individual">Individual (per option)</option>
                  </select>
                </label>

                {form.successRewardMode === 'shared' &&
                  renderSuccessRewardEditor({
                    title: 'Shared Success Rewards',
                    healthRestoreType: form.successHealthRestoreType,
                    healthRestoreValue: parseIntValue(
                      form.successHealthRestoreValue,
                      0
                    ),
                    manaRestoreType: form.successManaRestoreType,
                    manaRestoreValue: parseIntValue(
                      form.successManaRestoreValue,
                      0
                    ),
                    onHealthRestoreTypeChange: (value) =>
                      setForm((prev) => ({
                        ...prev,
                        successHealthRestoreType: value,
                      })),
                    onHealthRestoreValueChange: (value) =>
                      setForm((prev) => ({
                        ...prev,
                        successHealthRestoreValue: String(value),
                      })),
                    onManaRestoreTypeChange: (value) =>
                      setForm((prev) => ({
                        ...prev,
                        successManaRestoreType: value,
                      })),
                    onManaRestoreValueChange: (value) =>
                      setForm((prev) => ({
                        ...prev,
                        successManaRestoreValue: String(value),
                      })),
                    statuses: form.successStatuses,
                    onAddStatus: addScenarioSuccessStatus,
                    onUpdateStatus: updateScenarioSuccessStatus,
                    onRemoveStatus: removeScenarioSuccessStatus,
                  })}

                {form.options.map((option, optionIndex) => (
                  <div key={optionIndex} className="border rounded-md p-3 mb-3">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      <label className="text-sm md:col-span-2">
                        Option Text
                        <input
                          value={option.optionText}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              optionText: e.target.value,
                            })
                          }
                          className="w-full border rounded-md p-2"
                        />
                      </label>
                      <label className="text-sm md:col-span-2">
                        Success Text
                        <textarea
                          value={option.successText}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              successText: e.target.value,
                            })
                          }
                          className="w-full border rounded-md p-2 min-h-[64px]"
                        />
                      </label>
                      <label className="text-sm md:col-span-2">
                        Failure Text
                        <textarea
                          value={option.failureText}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              failureText: e.target.value,
                            })
                          }
                          className="w-full border rounded-md p-2 min-h-[64px]"
                        />
                      </label>
                      <label className="text-sm md:col-span-2">
                        Success Handoff
                        <textarea
                          value={option.successHandoffText}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              successHandoffText: e.target.value,
                            })
                          }
                          className="w-full border rounded-md p-2 min-h-[64px]"
                        />
                        <span className="mt-1 block text-xs text-gray-500">
                          Optional story bridge describing what this success
                          reveals or leads to next.
                        </span>
                      </label>
                      <label className="text-sm md:col-span-2">
                        Failure Handoff
                        <textarea
                          value={option.failureHandoffText}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              failureHandoffText: e.target.value,
                            })
                          }
                          className="w-full border rounded-md p-2 min-h-[64px]"
                        />
                        <span className="mt-1 block text-xs text-gray-500">
                          Optional story bridge describing how the quest thread
                          continues after this choice fails.
                        </span>
                      </label>
                      <label className="text-sm">
                        Stat Tag
                        <select
                          value={option.statTag}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              statTag: e.target.value,
                            })
                          }
                          className="w-full border rounded-md p-2"
                        >
                          {statTags.map((tag) => (
                            <option key={tag} value={tag}>
                              {tag}
                            </option>
                          ))}
                        </select>
                      </label>
                      <label className="text-sm">
                        Difficulty Override (optional)
                        <input
                          value={option.difficulty ?? ''}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              difficulty:
                                e.target.value === ''
                                  ? null
                                  : parseIntValue(e.target.value, 0),
                            })
                          }
                          className="w-full border rounded-md p-2"
                          type="number"
                          min={0}
                        />
                      </label>
                      <label className="text-sm md:col-span-2">
                        Proficiencies (comma separated)
                        <input
                          value={option.proficiencies.join(', ')}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              proficiencies: parseCsv(e.target.value),
                            })
                          }
                          className="w-full border rounded-md p-2"
                        />
                      </label>
                      <label className="text-sm">
                        Reward Experience
                        <input
                          value={option.rewardExperience}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              rewardExperience: parseIntValue(
                                e.target.value,
                                0
                              ),
                            })
                          }
                          className="w-full border rounded-md p-2"
                          type="number"
                          min={0}
                        />
                      </label>
                      <label className="text-sm">
                        Reward Gold
                        <input
                          value={option.rewardGold}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              rewardGold: parseIntValue(e.target.value, 0),
                            })
                          }
                          className="w-full border rounded-md p-2"
                          type="number"
                          min={0}
                        />
                      </label>
                    </div>

                    <div className="mt-3">
                      <MaterialRewardsEditor
                        value={option.materialRewards}
                        onChange={(materialRewards) =>
                          updateOption(optionIndex, { materialRewards })
                        }
                        title="Option Material Rewards"
                      />
                    </div>

                    <div className="mt-3">
                      <div className="flex items-center justify-between mb-2">
                        <div className="font-medium text-sm">
                          Option Item Rewards
                        </div>
                        <button
                          type="button"
                          className="bg-green-600 text-white px-2 py-1 rounded-md text-xs"
                          onClick={() => addOptionItemReward(optionIndex)}
                        >
                          Add Item
                        </button>
                      </div>
                      {option.itemRewards.map((reward, rewardIndex) => (
                        <div
                          key={rewardIndex}
                          className="grid grid-cols-1 md:grid-cols-3 gap-2 mb-2"
                        >
                          <select
                            value={reward.inventoryItemId}
                            onChange={(e) =>
                              updateOptionItemReward(optionIndex, rewardIndex, {
                                inventoryItemId: parseIntValue(
                                  e.target.value,
                                  0
                                ),
                              })
                            }
                            className="border rounded-md p-2"
                          >
                            <option value={0}>Select item</option>
                            {inventoryItems.map((item) => (
                              <option key={item.id} value={item.id}>
                                {item.name}
                              </option>
                            ))}
                          </select>
                          <input
                            value={reward.quantity}
                            onChange={(e) =>
                              updateOptionItemReward(optionIndex, rewardIndex, {
                                quantity: parseIntValue(e.target.value, 1),
                              })
                            }
                            className="border rounded-md p-2"
                            type="number"
                            min={1}
                          />
                          <button
                            type="button"
                            className="bg-red-500 text-white px-3 py-1 rounded-md"
                            onClick={() =>
                              removeOptionItemReward(optionIndex, rewardIndex)
                            }
                          >
                            Remove
                          </button>
                        </div>
                      ))}
                    </div>

                    <div className="mt-3">
                      <div className="flex items-center justify-between mb-2">
                        <div className="font-medium text-sm">
                          Option Spell Rewards
                        </div>
                        <button
                          type="button"
                          className="bg-green-600 text-white px-2 py-1 rounded-md text-xs"
                          onClick={() => addOptionSpellReward(optionIndex)}
                        >
                          Add Spell
                        </button>
                      </div>
                      {option.spellRewards.map((reward, rewardIndex) => (
                        <div
                          key={rewardIndex}
                          className="grid grid-cols-1 md:grid-cols-2 gap-2 mb-2"
                        >
                          <select
                            value={reward.spellId}
                            onChange={(e) =>
                              updateOptionSpellReward(
                                optionIndex,
                                rewardIndex,
                                {
                                  spellId: e.target.value,
                                }
                              )
                            }
                            className="border rounded-md p-2"
                          >
                            <option value="">Select spell</option>
                            {spells.map((spell) => (
                              <option key={spell.id} value={spell.id}>
                                {spell.name}
                              </option>
                            ))}
                          </select>
                          <button
                            type="button"
                            className="bg-red-500 text-white px-3 py-1 rounded-md"
                            onClick={() =>
                              removeOptionSpellReward(optionIndex, rewardIndex)
                            }
                          >
                            Remove
                          </button>
                        </div>
                      ))}
                    </div>

                    {form.failurePenaltyMode === 'individual' &&
                      renderFailurePenaltyEditor({
                        title: 'Option Failure Penalty',
                        healthDrainType: option.failureHealthDrainType,
                        healthDrainValue: option.failureHealthDrainValue,
                        manaDrainType: option.failureManaDrainType,
                        manaDrainValue: option.failureManaDrainValue,
                        onHealthDrainTypeChange: (value) =>
                          updateOption(optionIndex, {
                            failureHealthDrainType: value,
                          }),
                        onHealthDrainValueChange: (value) =>
                          updateOption(optionIndex, {
                            failureHealthDrainValue: value,
                          }),
                        onManaDrainTypeChange: (value) =>
                          updateOption(optionIndex, {
                            failureManaDrainType: value,
                          }),
                        onManaDrainValueChange: (value) =>
                          updateOption(optionIndex, {
                            failureManaDrainValue: value,
                          }),
                        statuses: option.failureStatuses,
                        onAddStatus: () => addOptionFailureStatus(optionIndex),
                        onUpdateStatus: (statusIndex, next) =>
                          updateOptionFailureStatus(
                            optionIndex,
                            statusIndex,
                            next
                          ),
                        onRemoveStatus: (statusIndex) =>
                          removeOptionFailureStatus(optionIndex, statusIndex),
                      })}

                    {form.successRewardMode === 'individual' &&
                      renderSuccessRewardEditor({
                        title: 'Option Success Rewards',
                        healthRestoreType: option.successHealthRestoreType,
                        healthRestoreValue: option.successHealthRestoreValue,
                        manaRestoreType: option.successManaRestoreType,
                        manaRestoreValue: option.successManaRestoreValue,
                        onHealthRestoreTypeChange: (value) =>
                          updateOption(optionIndex, {
                            successHealthRestoreType: value,
                          }),
                        onHealthRestoreValueChange: (value) =>
                          updateOption(optionIndex, {
                            successHealthRestoreValue: value,
                          }),
                        onManaRestoreTypeChange: (value) =>
                          updateOption(optionIndex, {
                            successManaRestoreType: value,
                          }),
                        onManaRestoreValueChange: (value) =>
                          updateOption(optionIndex, {
                            successManaRestoreValue: value,
                          }),
                        statuses: option.successStatuses,
                        onAddStatus: () => addOptionSuccessStatus(optionIndex),
                        onUpdateStatus: (statusIndex, next) =>
                          updateOptionSuccessStatus(
                            optionIndex,
                            statusIndex,
                            next
                          ),
                        onRemoveStatus: (statusIndex) =>
                          removeOptionSuccessStatus(optionIndex, statusIndex),
                      })}

                    <div className="mt-3">
                      <button
                        type="button"
                        className="bg-red-500 text-white px-3 py-1 rounded-md"
                        onClick={() => removeOption(optionIndex)}
                      >
                        Remove Option
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}

            <div className="flex gap-2">
              <button
                className="bg-blue-500 text-white px-4 py-2 rounded-md"
                onClick={save}
              >
                {editingId ? 'Update Scenario' : 'Create Scenario'}
              </button>
              <button
                className="bg-gray-500 text-white px-4 py-2 rounded-md"
                onClick={closeModal}
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {deleteId && (
        <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-md p-6 max-w-sm w-full">
            <h3 className="text-lg font-semibold mb-2">Delete Scenario</h3>
            <p className="text-sm text-gray-700 mb-4">
              This action cannot be undone.
            </p>
            <div className="flex gap-2">
              <button
                className="bg-red-500 text-white px-4 py-2 rounded-md"
                onClick={confirmDelete}
              >
                Delete
              </button>
              <button
                className="bg-gray-500 text-white px-4 py-2 rounded-md"
                onClick={() => setDeleteId(null)}
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {expandedScenarioImage && (
        <div
          className="fixed inset-0 bg-black/80 z-[60] flex items-center justify-center p-4"
          onClick={() => setExpandedScenarioImage(null)}
        >
          <div
            className="max-w-6xl max-h-[90vh] w-full"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex justify-end mb-2">
              <button
                type="button"
                className="bg-white text-gray-900 px-3 py-1 rounded-md"
                onClick={() => setExpandedScenarioImage(null)}
              >
                Close
              </button>
            </div>
            <img
              src={expandedScenarioImage.url}
              alt={expandedScenarioImage.title}
              className="w-full max-h-[80vh] object-contain rounded-md bg-black"
            />
          </div>
        </div>
      )}
    </div>
  );
};
