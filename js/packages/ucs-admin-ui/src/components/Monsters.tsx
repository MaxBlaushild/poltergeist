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

type DamageAffinity =
  | 'physical'
  | 'fire'
  | 'ice'
  | 'lightning'
  | 'poison'
  | 'arcane'
  | 'holy'
  | 'shadow';

type MonsterEncounterType = 'monster' | 'boss' | 'raid';
type MonsterTemplateType = 'monster' | 'boss' | 'raid';

type MonsterTemplateRecord = {
  id: string;
  createdAt: string;
  updatedAt: string;
  archived?: boolean;
  monsterType: MonsterTemplateType;
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
  strongAgainstAffinity?: DamageAffinity | null;
  weakAgainstAffinity?: DamageAffinity | null;
  spells: Spell[];
  imageGenerationStatus?: string;
  imageGenerationError?: string | null;
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
  strongAgainstAffinity?: DamageAffinity | null;
  weakAgainstAffinity?: DamageAffinity | null;
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
  totalCount: number;
  createdCount: number;
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
  strongAgainstAffinity: string;
  weakAgainstAffinity: string;
  spellIds: string[];
  techniqueIds: string[];
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

const damageAffinityOptions: DamageAffinity[] = [
  'physical',
  'fire',
  'ice',
  'lightning',
  'poison',
  'arcane',
  'holy',
  'shadow',
];

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

const monsterListPageSize = 25;

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

const emptyTemplateForm = (): MonsterTemplateFormState => ({
  monsterType: 'monster',
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
  strongAgainstAffinity: '',
  weakAgainstAffinity: '',
  spellIds: [],
  techniqueIds: [],
});

const templateFormFromRecord = (
  template: MonsterTemplateRecord
): MonsterTemplateFormState => ({
  monsterType: template.monsterType ?? 'monster',
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
  strongAgainstAffinity: template.strongAgainstAffinity ?? '',
  weakAgainstAffinity: template.weakAgainstAffinity ?? '',
  spellIds: (template.spells ?? [])
    .filter((spell) => (spell.abilityType ?? 'spell') !== 'technique')
    .map((spell) => spell.id),
  techniqueIds: (template.spells ?? [])
    .filter((spell) => (spell.abilityType ?? 'spell') === 'technique')
    .map((spell) => spell.id),
});

const templatePayloadFromForm = (form: MonsterTemplateFormState) => ({
  monsterType: form.monsterType,
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
  strongAgainstAffinity: form.strongAgainstAffinity.trim(),
  weakAgainstAffinity: form.weakAgainstAffinity.trim(),
  spellIds: Array.from(new Set([...form.spellIds, ...form.techniqueIds])),
});

const emptyMonsterForm = (): MonsterFormState => ({
  name: '',
  description: '',
  imageUrl: '',
  thumbnailUrl: '',
  zoneId: '',
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
  itemRewards:
    form.rewardMode === 'explicit'
      ? form.itemRewards
          .map((reward) => ({
            inventoryItemId: parseIntSafe(reward.inventoryItemId, 0),
            quantity: parseIntSafe(reward.quantity, 0),
          }))
          .filter((reward) => reward.inventoryItemId > 0 && reward.quantity > 0)
      : [],
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
  itemRewards:
    form.rewardMode === 'explicit'
      ? form.itemRewards
          .map((reward) => ({
            inventoryItemId: parseIntSafe(reward.inventoryItemId, 0),
            quantity: parseIntSafe(reward.quantity, 0),
          }))
          .filter((reward) => reward.inventoryItemId > 0 && reward.quantity > 0)
      : [],
  scaleWithUserLevel: form.scaleWithUserLevel,
  recurrenceFrequency: form.recurrenceFrequency,
  zoneId: form.zoneId.trim(),
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

const formatAffinityLabel = (affinity?: string | null): string => {
  if (!affinity) return 'None';
  return affinity.charAt(0).toUpperCase() + affinity.slice(1);
};

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
  const [inventoryItems, setInventoryItems] = useState<InventoryItemLite[]>([]);
  const [query, setQuery] = useState('');
  const [zoneQuery, setZoneQuery] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [bulkDeletingEncounters, setBulkDeletingEncounters] = useState(false);
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

  const loadReferenceData = useCallback(async () => {
    try {
      setReferenceLoading(true);
      const [spellResp, inventoryResp] = await Promise.all([
        apiClient.get<Spell[]>('/sonar/spells'),
        apiClient.get<InventoryItemLite[]>('/sonar/inventory-items'),
      ]);
      setSpells(Array.isArray(spellResp) ? spellResp : []);
      setInventoryItems(Array.isArray(inventoryResp) ? inventoryResp : []);
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
            }
          ),
          apiClient.get<PaginatedResponse<MonsterEncounterRecord>>(
            '/sonar/admin/monster-encounters',
            {
              page: encounterPage,
              pageSize: monsterListPageSize,
              query: deferredQuery.trim(),
              zoneQuery: deferredZoneQuery.trim(),
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
      monsterPage,
      templatePage,
      templateTab,
      templateTypeFilter,
    ]
  );

  useEffect(() => {
    void loadReferenceData();
  }, [loadReferenceData]);

  useEffect(() => {
    void loadPagedData();
  }, [loadPagedData]);

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
  }, [query, zoneQuery]);

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

  const spellAbilities = useMemo(
    () =>
      spells.filter((spell) => (spell.abilityType ?? 'spell') !== 'technique'),
    [spells]
  );
  const techniqueAbilities = useMemo(
    () =>
      spells.filter((spell) => (spell.abilityType ?? 'spell') === 'technique'),
    [spells]
  );

  const filteredTemplates = templates;
  const filteredMonsters = records;
  const filteredEncounters = encounters;
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
    setTemplateForm(emptyTemplateForm());
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
    setTemplateForm(emptyTemplateForm());
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
      if (
        payload.strongAgainstAffinity &&
        payload.strongAgainstAffinity === payload.weakAgainstAffinity
      ) {
        alert('Strong against and weak against affinities must be different.');
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
            `Created ${status.createdCount} ${typeLabel}${
              status.createdCount === 1 ? '' : 's'
            }.`
          );
          await loadPagedData(true);
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
    [apiClient, bulkTemplateType, loadPagedData]
  );

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
        { count, monsterType: bulkTemplateType }
      );
      setBulkTemplateJob(response);
      if (response.status === 'completed') {
        setBulkTemplateBusy(false);
        const typeLabel = formatMonsterTemplateTypeLabel(
          response.monsterType ?? bulkTemplateType
        );
        setBulkTemplateMessage(
          `Created ${response.createdCount} ${typeLabel}${
            response.createdCount === 1 ? '' : 's'
          }.`
        );
        await loadPagedData(true);
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
        { count: 1, monsterType }
      );
      setBulkTemplateJob(response);
      setBulkTemplateType(monsterType);
      if (response.status === 'completed') {
        setBulkTemplateBusy(false);
        setBulkTemplateMessage(
          `Created 1 ${formatMonsterTemplateTypeLabel(monsterType)}.`
        );
        await loadPagedData(true);
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
    const spellIds = Array.from(selected.selectedOptions).map(
      (option) => option.value
    );
    setTemplateForm((prev) => ({ ...prev, spellIds }));
  };

  const updateTemplateTechniqueIds = (selected: HTMLSelectElement) => {
    const techniqueIds = Array.from(selected.selectedOptions).map(
      (option) => option.value
    );
    setTemplateForm((prev) => ({ ...prev, techniqueIds }));
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

  return (
    <div className="p-6 bg-gray-100 min-h-screen">
      <div className="max-w-7xl mx-auto space-y-6">
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
              <input
                type="number"
                min={1}
                max={100}
                value={bulkTemplateCount}
                onChange={(event) => setBulkTemplateCount(event.target.value)}
                className="w-24 rounded-md border border-gray-300 px-2 py-2 text-sm"
                aria-label="Bulk template count"
              />
              <button
                className="qa-btn qa-btn-secondary"
                onClick={handleBulkGenerateTemplates}
                disabled={bulkTemplateBusy}
              >
                {bulkTemplateBusy ? 'Generating...' : 'Generate Templates'}
              </button>
              <button
                className="qa-btn qa-btn-secondary"
                onClick={() => handleGenerateSingleTypedTemplate('boss')}
                disabled={bulkTemplateBusy}
              >
                Generate Boss Template
              </button>
              <button
                className="qa-btn qa-btn-secondary"
                onClick={() => handleGenerateSingleTypedTemplate('raid')}
                disabled={bulkTemplateBusy}
              >
                Generate Raid Template
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
                Progress: {bulkTemplateJob.createdCount}/
                {bulkTemplateJob.totalCount}
              </span>
              <span>
                Type:{' '}
                {formatMonsterTemplateTypeLabel(
                  bulkTemplateJob.monsterType ?? bulkTemplateType
                )}
              </span>
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
              </div>
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
                              Strong vs{' '}
                              {formatAffinityLabel(
                                template.strongAgainstAffinity
                              )}
                              {' · '}
                              Weak vs{' '}
                              {formatAffinityLabel(
                                template.weakAgainstAffinity
                              )}
                            </p>
                            <p className="text-sm text-gray-500 mt-1">
                              Spells:{' '}
                              {template.spells?.filter(
                                (spell) =>
                                  (spell.abilityType ?? 'spell') !== 'technique'
                              ).length ?? 0}
                              {' · '}
                              Techniques:{' '}
                              {template.spells?.filter(
                                (spell) =>
                                  (spell.abilityType ?? 'spell') === 'technique'
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
              <h2 className="text-lg font-semibold mb-3">Monsters</h2>
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
                              Template:{' '}
                              {monster.template?.name ??
                                (monster.templateId
                                  ? templateNameById.get(monster.templateId)
                                  : 'N/A')}
                            </p>
                            <p className="text-sm text-gray-600">
                              Dominant Hand:{' '}
                              {monster.dominantHandInventoryItem?.name ??
                                monster.weaponInventoryItem?.name ??
                                (monster.dominantHandInventoryItemId ??
                                monster.weaponInventoryItemId
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
                              Strong vs{' '}
                              {formatAffinityLabel(
                                monster.strongAgainstAffinity ??
                                  monster.template?.strongAgainstAffinity
                              )}
                              {' · '}
                              Weak vs{' '}
                              {formatAffinityLabel(
                                monster.weakAgainstAffinity ??
                                  monster.template?.weakAgainstAffinity
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
                                ? `Random ${(monster.randomRewardSize ?? 'small').toUpperCase()} reward`
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
                                ? `Random ${(encounter.randomRewardSize ?? 'small').toUpperCase()} reward`
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
                <label className="block">
                  <span className="block text-sm mb-1">Strong Against</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.strongAgainstAffinity}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        strongAgainstAffinity: event.target.value,
                      }))
                    }
                  >
                    <option value="">None</option>
                    {damageAffinityOptions.map((affinity) => (
                      <option key={`strong-${affinity}`} value={affinity}>
                        {formatAffinityLabel(affinity)}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="block">
                  <span className="block text-sm mb-1">Weak Against</span>
                  <select
                    className="w-full border border-gray-300 rounded-md p-2"
                    value={templateForm.weakAgainstAffinity}
                    onChange={(event) =>
                      setTemplateForm((prev) => ({
                        ...prev,
                        weakAgainstAffinity: event.target.value,
                      }))
                    }
                  >
                    <option value="">None</option>
                    {damageAffinityOptions.map((affinity) => (
                      <option key={`weak-${affinity}`} value={affinity}>
                        {formatAffinityLabel(affinity)}
                      </option>
                    ))}
                  </select>
                </label>
              </div>

              <label className="block">
                <span className="block text-sm mb-1">Spells</span>
                <select
                  multiple
                  className="w-full border border-gray-300 rounded-md p-2 min-h-[180px]"
                  value={templateForm.spellIds}
                  onChange={(event) => updateTemplateSpellIds(event.target)}
                >
                  {spellAbilities.map((spell) => (
                    <option key={spell.id} value={spell.id}>
                      {spell.name}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block">
                <span className="block text-sm mb-1">Techniques</span>
                <select
                  multiple
                  className="w-full border border-gray-300 rounded-md p-2 min-h-[180px]"
                  value={templateForm.techniqueIds}
                  onChange={(event) => updateTemplateTechniqueIds(event.target)}
                >
                  {techniqueAbilities.map((technique) => (
                    <option key={technique.id} value={technique.id}>
                      {technique.name}
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
              <div className="grid md:grid-cols-2 gap-4">
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
                  <h3 className="font-medium">Item Rewards</h3>
                  <button
                    type="button"
                    className="qa-btn qa-btn-secondary"
                    disabled={monsterForm.rewardMode !== 'explicit'}
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
                        disabled={monsterForm.rewardMode !== 'explicit'}
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
                        disabled={monsterForm.rewardMode !== 'explicit'}
                        onChange={(event) =>
                          updateMonsterItemReward(index, {
                            quantity: event.target.value,
                          })
                        }
                      />
                      <button
                        type="button"
                        className="qa-btn qa-btn-danger"
                        disabled={monsterForm.rewardMode !== 'explicit'}
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
                  Random rewards ignore explicit XP, gold, material, and item
                  rewards.
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
                  <h3 className="font-medium">Encounter Item Rewards</h3>
                  <button
                    type="button"
                    className="qa-btn qa-btn-secondary"
                    disabled={encounterForm.rewardMode !== 'explicit'}
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
                        disabled={encounterForm.rewardMode !== 'explicit'}
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
                        disabled={encounterForm.rewardMode !== 'explicit'}
                        onChange={(event) =>
                          updateEncounterItemReward(index, {
                            quantity: event.target.value,
                          })
                        }
                      />
                      <button
                        type="button"
                        className="qa-btn qa-btn-danger"
                        disabled={encounterForm.rewardMode !== 'explicit'}
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
