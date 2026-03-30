import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useAPI, useTagContext, useZoneContext } from '@poltergeist/contexts';
import {
  Candidate,
  Character,
  InventoryItem,
  LocationArchetype,
  PointOfInterest,
  Quest,
  QuestArchetype,
  QuestArchetypeChallenge,
  QuestArchetypeNode,
  QuestDifficultyMode,
  QuestNode,
  QuestNodeSubmissionType,
  Spell,
  Tag,
} from '@poltergeist/types';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import * as wellknown from 'wellknown';
import { useQuestArchetypes } from '../contexts/questArchetypes.tsx';
import { useCandidates } from '@poltergeist/hooks';
import { Link } from 'react-router-dom';
import {
  MaterialRewardsEditor,
  emptyMaterialReward,
  normalizeMaterialRewards,
} from './MaterialRewardsEditor.tsx';
import './questArchetypeTheme.css';
import './questsTheme.css';

type PointOfInterestImport = {
  id: string;
  placeId: string;
  zoneId: string;
  status: string;
  errorMessage?: string | null;
  pointOfInterestId?: string | null;
  createdAt: string;
  updatedAt: string;
};

type QuestNodeType = 'poi' | 'polygon' | 'scenario' | 'monster' | 'challenge';

type LegacyQuestNodePrompt = {
  id: string;
  tier?: number;
  question?: string;
  reward?: number;
  inventoryItemId?: number | null;
  submissionType?: QuestNodeSubmissionType;
  difficulty?: number;
  statTags?: string[];
  proficiency?: string | null;
  challengeShuffleStatus?: string;
  challengeShuffleError?: string | null;
};

type ScenarioNodeOption = {
  id: string;
  zoneId: string;
  pointOfInterestId?: string | null;
  pointOfInterest?: PointOfInterest | null;
  latitude: number;
  longitude: number;
  prompt: string;
  difficulty?: number;
  openEnded?: boolean;
  options?: { id: string }[];
};

type MonsterNodeOption = {
  id: string;
  zoneId: string;
  pointOfInterestId?: string | null;
  pointOfInterest?: PointOfInterest | null;
  latitude: number;
  longitude: number;
  name: string;
  description?: string;
  encounterType?: string;
  scaleWithUserLevel?: boolean;
  monsterCount?: number;
  members?: { slot: number; monster: { id: string; name: string } }[];
};

type ChallengeNodeOption = {
  id: string;
  zoneId: string;
  pointOfInterestId?: string | null;
  pointOfInterest?: PointOfInterest | null;
  latitude: number;
  longitude: number;
  question: string;
  description?: string;
  submissionType?: QuestNodeSubmissionType;
  difficulty?: number;
  statTags?: string[];
  proficiency?: string | null;
  reward?: number;
  rewardExperience?: number;
};

type SelectOption = {
  value: string;
  label: string;
  secondary?: string;
};

type MonsterRecord = {
  id: string;
  zoneId: string;
  name: string;
  level?: number;
};

type ResolvedQuestNodeScenario = ScenarioNodeOption;
type ResolvedQuestNodeMonsterEncounter = MonsterNodeOption;
type ResolvedQuestNodeChallenge = ChallengeNodeOption;

type QuickCreateScenarioOptionForm = {
  optionText: string;
  statTag: string;
  difficulty: string;
  proficiencies: string;
  successText: string;
  failureText: string;
};

type QuickCreateScenarioForm = {
  prompt: string;
  imageUrl: string;
  thumbnailUrl: string;
  latitude: string;
  longitude: string;
  options: QuickCreateScenarioOptionForm[];
};

type QuickCreateChallengeForm = {
  pointOfInterestId: string;
  question: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  latitude: string;
  longitude: string;
  rewardExperience: string;
  rewardGold: string;
  submissionType: QuestNodeSubmissionType;
  statTags: string[];
  difficulty: string;
  proficiency: string;
};

type QuickCreateMonsterEncounterForm = {
  name: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  latitude: string;
  longitude: string;
  scaleWithUserLevel: boolean;
  monsterIds: string[];
};

mapboxgl.accessToken = process.env.REACT_APP_MAPBOX_ACCESS_TOKEN || '';

const emptyQuestForm = {
  name: '',
  description: '',
  acceptanceDialogue: [] as string[],
  imageUrl: '',
  zoneId: '',
  questGiverCharacterId: '',
  questArchetypeId: '',
  recurrenceFrequency: '',
  difficultyMode: 'scale' as QuestDifficultyMode,
  difficulty: 1,
  monsterEncounterTargetLevel: 1,
  rewardMode: 'random' as 'explicit' | 'random',
  randomRewardSize: 'small' as 'small' | 'medium' | 'large',
  rewardExperience: 0,
  gold: 0,
  materialRewards: [] as ReturnType<typeof emptyMaterialReward>[],
  itemRewards: [] as { inventoryItemId: string; quantity: number }[],
  spellRewards: [] as { spellId: string }[],
};

const buildQuestFormFromQuest = (quest: Quest) => {
  const rewardMode = getQuestRewardMode(quest);
  return {
    name: quest.name ?? '',
    description: quest.description ?? '',
    acceptanceDialogue: quest.acceptanceDialogue ?? [],
    imageUrl: quest.imageUrl ?? '',
    zoneId: quest.zoneId ?? '',
    questGiverCharacterId: quest.questGiverCharacterId ?? '',
    questArchetypeId: quest.questArchetypeId ?? '',
    recurrenceFrequency: quest.recurrenceFrequency ?? '',
    difficultyMode: quest.difficultyMode === 'fixed' ? 'fixed' : 'scale',
    difficulty: quest.difficulty ?? 1,
    monsterEncounterTargetLevel: quest.monsterEncounterTargetLevel ?? 1,
    rewardMode,
    randomRewardSize:
      (quest.randomRewardSize as 'small' | 'medium' | 'large') ?? 'small',
    rewardExperience: quest.rewardExperience ?? 0,
    gold: quest.gold ?? 0,
    materialRewards: (quest.materialRewards ?? []).map((reward) => ({
      resourceKey: reward.resourceKey,
      amount: reward.amount ?? 1,
    })),
    itemRewards: (quest.itemRewards ?? []).map((reward) => ({
      inventoryItemId: reward.inventoryItemId
        ? String(reward.inventoryItemId)
        : '',
      quantity: reward.quantity ?? 1,
    })),
    spellRewards: (quest.spellRewards ?? []).map((reward) => ({
      spellId: reward.spellId ?? '',
    })),
  };
};

const normalizeQuestSummary = (quest: Quest): Quest => ({
  ...quest,
  nodeCount: quest.nodeCount ?? quest.nodes?.length ?? 0,
  detailLoaded: false,
  nodes: undefined,
});

const normalizeQuestDetail = (quest: Quest): Quest => ({
  ...quest,
  nodeCount: quest.nodes?.length ?? quest.nodeCount ?? 0,
  detailLoaded: true,
});

const questStatOptions = [
  { id: 'strength', label: 'Strength' },
  { id: 'dexterity', label: 'Dexterity' },
  { id: 'constitution', label: 'Constitution' },
  { id: 'intelligence', label: 'Intelligence' },
  { id: 'wisdom', label: 'Wisdom' },
  { id: 'charisma', label: 'Charisma' },
];

const emptyNodeForm = {
  orderIndex: 1,
  nodeType: 'scenario' as QuestNodeType,
  submissionType: 'photo' as QuestNodeSubmissionType,
  pointOfInterestId: '',
  scenarioId: '',
  monsterEncounterId: '',
  challengeId: '',
  polygonPoints: '',
};

const questNodeSubmissionOptions: {
  value: QuestNodeSubmissionType;
  label: string;
}[] = [
  { value: 'text', label: 'Text' },
  { value: 'photo', label: 'Photo' },
  { value: 'video', label: 'Video' },
];

const emptyChallengeForm = {
  tier: 1,
  question: '',
  reward: 0,
  inventoryItemId: '',
  locationArchetypeId: '',
  locationChallenge: '',
  submissionType: 'photo' as QuestNodeSubmissionType,
  statTags: [] as string[],
  difficulty: 25,
  proficiency: '',
};

const createEmptyQuickScenarioOption = (): QuickCreateScenarioOptionForm => ({
  optionText: '',
  statTag: 'strength',
  difficulty: '25',
  proficiencies: '',
  successText: '',
  failureText: '',
});

const emptyQuickCreateScenarioForm = (): QuickCreateScenarioForm => ({
  prompt: '',
  imageUrl: '',
  thumbnailUrl: '',
  latitude: '',
  longitude: '',
  options: [createEmptyQuickScenarioOption()],
});

const emptyQuickCreateChallengeForm = (): QuickCreateChallengeForm => ({
  pointOfInterestId: '',
  question: '',
  description: '',
  imageUrl: '',
  thumbnailUrl: '',
  latitude: '',
  longitude: '',
  rewardExperience: '0',
  rewardGold: '0',
  submissionType: 'photo',
  statTags: [],
  difficulty: '25',
  proficiency: '',
});

const emptyQuickCreateMonsterEncounterForm =
  (): QuickCreateMonsterEncounterForm => ({
    name: '',
    description: '',
    imageUrl: '',
    thumbnailUrl: '',
    latitude: '',
    longitude: '',
    scaleWithUserLevel: false,
    monsterIds: [],
  });

const parseIntSafe = (value: string, fallback = 0) => {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const SearchableSelect = ({
  label,
  placeholder,
  options,
  value,
  onChange,
  disabled = false,
  noMatchesLabel = 'No matches found',
}: {
  label: string;
  placeholder: string;
  options: SelectOption[];
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
  noMatchesLabel?: string;
}) => {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');

  const selected = options.find((option) => option.value === value);
  const filtered = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    if (!normalized) return options;
    return options.filter((option) => {
      const haystack =
        `${option.label} ${option.secondary ?? ''}`.toLowerCase();
      return haystack.includes(normalized);
    });
  }, [options, query]);

  const displayValue = open ? query : selected?.label ?? '';

  return (
    <div className="relative">
      <label className="block text-sm">{label}</label>
      <input
        value={displayValue}
        onChange={(event) => {
          setQuery(event.target.value);
          setOpen(true);
        }}
        onFocus={() => {
          if (disabled) return;
          setOpen(true);
          setQuery('');
        }}
        onBlur={() => {
          window.setTimeout(() => setOpen(false), 150);
        }}
        placeholder={placeholder}
        disabled={disabled}
        className="w-full border rounded-md p-2 disabled:bg-gray-100 disabled:text-gray-500"
      />
      {open && !disabled ? (
        <div className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md border border-gray-200 bg-white shadow-lg">
          {filtered.length === 0 ? (
            <div className="px-3 py-2 text-sm text-gray-500">
              {noMatchesLabel}
            </div>
          ) : (
            filtered.map((option) => (
              <button
                type="button"
                key={option.value || '__empty-option__'}
                onMouseDown={(event) => event.preventDefault()}
                onClick={() => {
                  onChange(option.value);
                  setOpen(false);
                  setQuery('');
                }}
                className="flex w-full flex-col items-start px-3 py-2 text-left text-sm hover:bg-indigo-50"
              >
                <span className="font-medium text-gray-900">
                  {option.label}
                </span>
                {option.secondary ? (
                  <span className="text-xs text-gray-500">
                    {option.secondary}
                  </span>
                ) : null}
              </button>
            ))
          )}
        </div>
      ) : null}
    </div>
  );
};

const getQuestRewardMode = (quest: Quest): 'explicit' | 'random' => {
  if (quest.rewardMode === 'explicit' || quest.rewardMode === 'random') {
    return quest.rewardMode;
  }
  const hasExplicitRewards =
    (quest.rewardExperience ?? 0) > 0 ||
    (quest.gold ?? 0) > 0 ||
    (quest.itemRewards?.length ?? 0) > 0 ||
    (quest.spellRewards?.length ?? 0) > 0;
  return hasExplicitRewards ? 'explicit' : 'random';
};

const questRecurrenceOptions = [
  { value: '', label: 'No Recurrence' },
  { value: 'daily', label: 'Daily' },
  { value: 'weekly', label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
];

const getQuestRecurrenceLabel = (value?: string | null) => {
  const match = questRecurrenceOptions.find(
    (option) => option.value === (value ?? '')
  );
  if (match) return match.label;
  if (!value) return 'No Recurrence';
  return value.charAt(0).toUpperCase() + value.slice(1);
};

const getQuestNodeKind = (node: QuestNode): QuestNodeType => {
  if (node.pointOfInterestId) return 'poi';
  if (node.scenarioId) return 'scenario';
  if (node.monsterEncounterId || node.monsterId) return 'monster';
  if (node.challengeId) return 'challenge';
  return 'polygon';
};

const getQuestNodeKindLabel = (nodeType: QuestNodeType) => {
  switch (nodeType) {
    case 'poi':
      return 'Location';
    case 'scenario':
      return 'Scenario';
    case 'monster':
      return 'Monster';
    case 'challenge':
      return 'Challenge Objective';
    case 'polygon':
    default:
      return 'Polygon';
  }
};

const formatStatTagLabel = (tag: string) =>
  tag.charAt(0).toUpperCase() + tag.slice(1);

const buildChallengeFormFromChallenge = (
  challenge: LegacyQuestNodePrompt,
  fallbackSubmissionType?: QuestNodeSubmissionType
) => ({
  ...emptyChallengeForm,
  tier: challenge.tier ?? 1,
  question: challenge.question ?? '',
  reward: challenge.reward ?? 0,
  inventoryItemId: challenge.inventoryItemId
    ? String(challenge.inventoryItemId)
    : '',
  statTags: challenge.statTags ?? [],
  difficulty: challenge.difficulty ?? 0,
  proficiency: challenge.proficiency ?? '',
  submissionType: (challenge.submissionType ??
    fallbackSubmissionType ??
    'photo') as QuestNodeSubmissionType,
});

const emptyQuestReward = {
  inventoryItemId: '',
  quantity: 1,
};

const emptyQuestSpellReward = {
  spellId: '',
};

const adminEntityLinkClass =
  'inline-flex rounded-md border border-indigo-200 bg-indigo-50 px-2 py-1 text-xs text-indigo-700 hover:bg-indigo-100';

const parsePolygonPoints = (input: string): [number, number][] | null => {
  if (!input.trim()) return null;
  try {
    const parsed = JSON.parse(input);
    if (!Array.isArray(parsed)) return null;
    const points: [number, number][] = [];
    for (const entry of parsed) {
      if (!Array.isArray(entry) || entry.length < 2) return null;
      const lng = Number(entry[0]);
      const lat = Number(entry[1]);
      if (Number.isNaN(lng) || Number.isNaN(lat)) return null;
      points.push([lng, lat]);
    }
    return points.length ? points : null;
  } catch {
    return null;
  }
};

const parsePolygonWkt = (raw: string): number[][][] | null => {
  if (!raw) return null;
  let trimmed = raw.trim();
  if (!trimmed) return null;
  if (trimmed.toUpperCase().startsWith('SRID=')) {
    const parts = trimmed.split(';');
    trimmed = parts[parts.length - 1] || trimmed;
  }
  try {
    const geometry = wellknown.parse(trimmed);
    if (!geometry || geometry.type !== 'Polygon') return null;
    return geometry.coordinates as number[][][];
  } catch {
    return null;
  }
};

const summarizeScenarioPrompt = (prompt: string) => {
  const normalized = prompt.replace(/\s+/g, ' ').trim();
  if (!normalized) return '(Untitled scenario)';
  return normalized.length > 80 ? `${normalized.slice(0, 80)}...` : normalized;
};

const summarizePoiText = (value?: string | null, maxLength = 160) => {
  const normalized = (value ?? '').replace(/\s+/g, ' ').trim();
  if (!normalized) return '';
  return normalized.length > maxLength
    ? `${normalized.slice(0, maxLength)}...`
    : normalized;
};

const getPointOfInterestAliases = (poi?: PointOfInterest | null) => {
  if (!poi) return [];
  return [poi.googleMapsPlaceName, poi.originalName]
    .map((value) => (value ?? '').trim())
    .filter((value, index, items) => value.length > 0 && items.indexOf(value) === index);
};

const formatPoiCoordinate = (value?: string | number | null) => {
  const numeric = typeof value === 'number' ? value : Number(value);
  if (Number.isNaN(numeric)) return null;
  return numeric.toFixed(5);
};

const formatNodeCoordinatePair = (
  latitude?: string | number | null,
  longitude?: string | number | null
) => {
  const formattedLatitude = formatPoiCoordinate(latitude);
  const formattedLongitude = formatPoiCoordinate(longitude);
  if (!formattedLatitude || !formattedLongitude) return null;
  return `${formattedLatitude}, ${formattedLongitude}`;
};

const resolveLinkedQuestScenario = (
  node: QuestNode,
  scenarios: ScenarioNodeOption[]
): ResolvedQuestNodeScenario | null =>
  (node.scenario as ResolvedQuestNodeScenario | null | undefined) ??
  (node.scenarioId
    ? scenarios.find((scenario) => scenario.id === node.scenarioId) ?? null
    : null);

const resolveLinkedQuestMonsterEncounter = (
  node: QuestNode,
  monsterEncounters: MonsterNodeOption[]
): ResolvedQuestNodeMonsterEncounter | null => {
  const embeddedEncounter = node.monsterEncounter as
    | ResolvedQuestNodeMonsterEncounter
    | null
    | undefined;
  if (embeddedEncounter) {
    return embeddedEncounter;
  }
  if (node.monsterEncounterId) {
    const directEncounter = monsterEncounters.find(
      (encounter) => encounter.id === node.monsterEncounterId
    );
    if (directEncounter) {
      return directEncounter;
    }
  }
  if (node.monsterId) {
    const directEncounter = monsterEncounters.find(
      (encounter) => encounter.id === node.monsterId
    );
    if (directEncounter) {
      return directEncounter;
    }
    const encounterByMember = monsterEncounters.find((encounter) =>
      (encounter.members ?? []).some(
        (member) => member.monster.id === node.monsterId
      )
    );
    if (encounterByMember) {
      return encounterByMember;
    }
  }
  if (node.monster) {
    return {
      id: node.monster.id,
      zoneId: node.monster.zoneId ?? '',
      latitude: node.monster.latitude,
      longitude: node.monster.longitude,
      name: node.monster.name,
      description: node.monster.description,
      encounterType: 'monster',
      scaleWithUserLevel: false,
      monsterCount: 1,
      members: [
        {
          slot: 1,
          monster: {
            id: node.monster.id,
            name: node.monster.name,
          },
        },
      ],
    };
  }
  return null;
};

const resolveLinkedQuestChallenge = (
  node: QuestNode,
  challenges: ChallengeNodeOption[]
): ResolvedQuestNodeChallenge | null =>
  (node.challenge as ResolvedQuestNodeChallenge | null | undefined) ??
  (node.challengeId
    ? challenges.find((challenge) => challenge.id === node.challengeId) ?? null
    : null);

const getLinkedQuestNodePoiId = (
  node: QuestNode,
  linkedScenario?: ResolvedQuestNodeScenario | null,
  linkedMonsterEncounter?: ResolvedQuestNodeMonsterEncounter | null,
  linkedChallenge?: ResolvedQuestNodeChallenge | null
) =>
  node.pointOfInterestId ??
  linkedChallenge?.pointOfInterestId ??
  linkedChallenge?.pointOfInterest?.id ??
  linkedScenario?.pointOfInterestId ??
  linkedScenario?.pointOfInterest?.id ??
  linkedMonsterEncounter?.pointOfInterestId ??
  linkedMonsterEncounter?.pointOfInterest?.id ??
  null;

const resolveLinkedQuestNodePoi = (
  node: QuestNode,
  pointsOfInterest: PointOfInterest[],
  linkedScenario?: ResolvedQuestNodeScenario | null,
  linkedMonsterEncounter?: ResolvedQuestNodeMonsterEncounter | null,
  linkedChallenge?: ResolvedQuestNodeChallenge | null
) => {
  const linkedPoiId = getLinkedQuestNodePoiId(
    node,
    linkedScenario,
    linkedMonsterEncounter,
    linkedChallenge
  );
  const linkedPoi =
    (linkedPoiId
      ? pointsOfInterest.find((poi) => poi.id === linkedPoiId) ?? null
      : null) ??
    linkedChallenge?.pointOfInterest ??
    linkedScenario?.pointOfInterest ??
    linkedMonsterEncounter?.pointOfInterest ??
    null;
  return linkedPoi;
};

const normalizeText = (value: string) => value.trim().toLowerCase();

const normalizeAcceptanceDialogue = (lines: string[]) =>
  lines.map((line) => line.trim()).filter((line) => line.length > 0);

const resolveChallengeSubmissionType = (
  challenge: LegacyQuestNodePrompt,
  node?: QuestNode
) =>
  (challenge.submissionType ||
    node?.submissionType ||
    'photo') as QuestNodeSubmissionType;

const getLegacyQuestNodePrompts = (node?: QuestNode | null) =>
  (((node ?? null) as (QuestNode & {
    challenges?: LegacyQuestNodePrompt[];
  }) | null)?.challenges ?? []) as LegacyQuestNodePrompt[];

const questNodeUsesLinkedObjective = (node?: QuestNode | null) =>
  Boolean(
    node?.challengeId ||
      node?.scenarioId ||
      node?.monsterEncounterId ||
      node?.monsterId
  );

const formatChallengeShuffleStatus = (status?: string | null) => {
  switch ((status || '').toLowerCase()) {
    case 'queued':
      return 'Queued';
    case 'in_progress':
      return 'In progress';
    case 'completed':
      return 'Completed';
    case 'failed':
      return 'Failed';
    default:
      return 'Idle';
  }
};

const closePolygonRing = (ring: [number, number][]) => {
  if (ring.length === 0) return ring;
  const [firstLng, firstLat] = ring[0];
  const [lastLng, lastLat] = ring[ring.length - 1];
  if (firstLng === lastLng && firstLat === lastLat) return ring;
  return [...ring, ring[0]];
};

const normalizePolygonCoordinates = (coords: number[][][] | null) => {
  if (!coords || coords.length === 0) return null;
  const ring = coords[0] ?? [];
  if (ring.length < 3) return null;
  return [closePolygonRing(ring as [number, number][])] as number[][][];
};

const matchLocationArchetypeForPoi = (
  poi: PointOfInterest,
  archetypes: LocationArchetype[]
): LocationArchetype | null => {
  if (!poi.tags || poi.tags.length === 0) return null;
  const tagSet = new Set(poi.tags.map((tag) => normalizeText(tag.name)));
  let best: LocationArchetype | null = null;
  let bestScore = 0;

  archetypes.forEach((archetype) => {
    const included = (archetype.includedTypes || []).map(normalizeText);
    let score = 0;
    included.forEach((type) => {
      if (tagSet.has(type)) score += 1;
    });
    if (score > bestScore) {
      bestScore = score;
      best = archetype;
    }
  });

  return bestScore > 0 ? best : null;
};

export const Quests = () => {
  const { apiClient } = useAPI();
  const { zones } = useZoneContext();
  const { tagGroups } = useTagContext();
  const { locationArchetypes } = useQuestArchetypes();
  const [quests, setQuests] = useState<Quest[]>([]);
  const [pointsOfInterest, setPointsOfInterest] = useState<PointOfInterest[]>(
    []
  );
  const [characters, setCharacters] = useState<Character[]>([]);
  const [inventoryItems, setInventoryItems] = useState<InventoryItem[]>([]);
  const [spells, setSpells] = useState<Spell[]>([]);
  const [scenarios, setScenarios] = useState<ScenarioNodeOption[]>([]);
  const [monsterRecords, setMonsterRecords] = useState<MonsterRecord[]>([]);
  const [monsterEncounters, setMonsterEncounters] = useState<
    MonsterNodeOption[]
  >([]);
  const [challenges, setChallenges] = useState<ChallengeNodeOption[]>([]);
  const [loading, setLoading] = useState(true);
  const [questDetailLoadingId, setQuestDetailLoadingId] = useState<
    string | null
  >(null);
  const [questDetailErrorId, setQuestDetailErrorId] = useState<string | null>(
    null
  );
  const [loadError, setLoadError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [zoneSearch, setZoneSearch] = useState('');
  const [characterSearch, setCharacterSearch] = useState('');
  const [poiSearch, setPoiSearch] = useState('');
  const [poiFiltersOpen, setPoiFiltersOpen] = useState(false);
  const [poiZoneFilterId, setPoiZoneFilterId] = useState('');
  const [poiTagSearch, setPoiTagSearch] = useState('');
  const [poiTagFilterIds, setPoiTagFilterIds] = useState<string[]>([]);
  const [zonePoiMap, setZonePoiMap] = useState<Record<string, Set<string>>>({});
  const [zonePoiMapLoaded, setZonePoiMapLoaded] = useState(false);
  const [zonePoiMapLoading, setZonePoiMapLoading] = useState(false);
  const [zoneDetailsById, setZoneDetailsById] = useState<
    Record<
      string,
      {
        boundary?: number[][];
        boundaryCoords?: { latitude: number; longitude: number }[];
        latitude?: number;
        longitude?: number;
      }
    >
  >({});
  const [selectedQuestId, setSelectedQuestId] = useState<string>('');
  const [showCreateQuest, setShowCreateQuest] = useState(false);
  const [questForm, setQuestForm] = useState({ ...emptyQuestForm });
  const [nodeForm, setNodeForm] = useState({ ...emptyNodeForm });
  const [polygonDraftPoints, setPolygonDraftPoints] = useState<
    [number, number][]
  >([]);
  const [challengeDrafts, setChallengeDrafts] = useState<
    Record<string, typeof emptyChallengeForm>
  >({});
  const [challengeEdits, setChallengeEdits] = useState<
    Record<string, typeof emptyChallengeForm>
  >({});
  const [quickCreateOpen, setQuickCreateOpen] = useState<
    Record<'scenario' | 'monster' | 'challenge', boolean>
  >({
    scenario: false,
    monster: false,
    challenge: false,
  });
  const [quickCreateScenarioForm, setQuickCreateScenarioForm] =
    useState<QuickCreateScenarioForm>(emptyQuickCreateScenarioForm());
  const [quickCreateChallengeForm, setQuickCreateChallengeForm] =
    useState<QuickCreateChallengeForm>(emptyQuickCreateChallengeForm());
  const [quickCreateMonsterEncounterForm, setQuickCreateMonsterEncounterForm] =
    useState<QuickCreateMonsterEncounterForm>(
      emptyQuickCreateMonsterEncounterForm()
    );
  const [quickCreateSubmitting, setQuickCreateSubmitting] = useState<
    null | 'scenario' | 'monster' | 'challenge'
  >(null);
  const [proficiencySearch, setProficiencySearch] = useState('');
  const [proficiencyOptions, setProficiencyOptions] = useState<string[]>([]);
  const [characterLocationsOpen, setCharacterLocationsOpen] = useState(false);
  const [selectedCharacterLocations, setSelectedCharacterLocations] = useState<
    { latitude: number; longitude: number }[]
  >([]);
  const [characterLocationsLoading, setCharacterLocationsLoading] =
    useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [importQuery, setImportQuery] = useState('');
  const [selectedCandidate, setSelectedCandidate] = useState<Candidate | null>(
    null
  );
  const [importZoneId, setImportZoneId] = useState('');
  const [importError, setImportError] = useState<string | null>(null);
  const [importJobs, setImportJobs] = useState<PointOfInterestImport[]>([]);
  const [importPolling, setImportPolling] = useState(false);
  const { candidates } = useCandidates(importQuery);
  const [importToasts, setImportToasts] = useState<string[]>([]);
  const [notifiedImportIds, setNotifiedImportIds] = useState<Set<string>>(
    new Set()
  );
  const [polygonRefreshNonce, setPolygonRefreshNonce] = useState(0);
  const [deletingQuestId, setDeletingQuestId] = useState<string | null>(null);
  const [bulkDeletingQuests, setBulkDeletingQuests] = useState(false);
  const [selectedQuestIds, setSelectedQuestIds] = useState<Set<string>>(
    new Set()
  );
  const shufflingChallengeId: string | null = null;
  const [creatingArchetype, setCreatingArchetype] = useState(false);
  const questMapContainer = useRef<HTMLDivElement>(null);
  const questMap = useRef<mapboxgl.Map | null>(null);
  const [questMapLoaded, setQuestMapLoaded] = useState(false);
  const questNodeMarkers = useRef<mapboxgl.Marker[]>([]);
  const selectedQuestIdRef = useRef('');

  const setActiveQuestId = (questId: string) => {
    selectedQuestIdRef.current = questId;
    setSelectedQuestId(questId);
  };

  const selectedQuest = useMemo(
    () => quests.find((quest) => quest.id === selectedQuestId) ?? null,
    [quests, selectedQuestId]
  );
  const selectedQuestIsHydrating = Boolean(
    selectedQuest && questDetailLoadingId === selectedQuest.id
  );
  const selectedQuestNeedsHydration = Boolean(
    selectedQuest && !selectedQuest.detailLoaded
  );
  const selectedQuestIdSet = useMemo(
    () => selectedQuestIds,
    [selectedQuestIds]
  );
  const orderedQuestNodes = useMemo(
    () =>
      (selectedQuest?.nodes ?? [])
        .slice()
        .sort((a, b) => a.orderIndex - b.orderIndex),
    [selectedQuest?.nodes]
  );
  const selectedQuestZone = useMemo(
    () => zones.find((zone) => zone.id === questForm.zoneId) ?? null,
    [questForm.zoneId, zones]
  );
  const selectedQuestGiver = useMemo(
    () =>
      characters.find(
        (character) => character.id === questForm.questGiverCharacterId
      ) ?? null,
    [characters, questForm.questGiverCharacterId]
  );
  const selectedQuestNodeCounts = useMemo(() => {
    const counts: Record<QuestNodeType, number> = {
      poi: 0,
      polygon: 0,
      scenario: 0,
      monster: 0,
      challenge: 0,
    };
    orderedQuestNodes.forEach((node) => {
      counts[getQuestNodeKind(node)] += 1;
    });
    return counts;
  }, [orderedQuestNodes]);
  const selectedQuestNodePreview = useMemo(
    () =>
      orderedQuestNodes.map((node) => {
        const nodeType = getQuestNodeKind(node);
        if (node.pointOfInterestId) {
          const poi = pointsOfInterest.find(
            (item) => item.id === node.pointOfInterestId
          );
          return {
            id: node.id,
            orderIndex: node.orderIndex,
            nodeType,
            label: poi?.name ?? node.pointOfInterestId,
          };
        }
        if (node.scenarioId) {
          const scenario = resolveLinkedQuestScenario(node, scenarios);
          return {
            id: node.id,
            orderIndex: node.orderIndex,
            nodeType,
            label: summarizeScenarioPrompt(scenario?.prompt ?? ''),
          };
        }
        if (node.monsterEncounterId || node.monsterId) {
          const encounterId = node.monsterEncounterId ?? node.monsterId ?? '';
          const resolved = resolveLinkedQuestMonsterEncounter(
            node,
            monsterEncounters
          );
          return {
            id: node.id,
            orderIndex: node.orderIndex,
            nodeType,
            label: resolved?.name ?? encounterId,
          };
        }
        if (node.challengeId) {
          const challenge = resolveLinkedQuestChallenge(node, challenges);
          return {
            id: node.id,
            orderIndex: node.orderIndex,
            nodeType,
            label: challenge?.question ?? node.challengeId,
          };
        }
        return {
          id: node.id,
          orderIndex: node.orderIndex,
          nodeType,
          label: 'Drawn Area',
        };
      }),
    [challenges, monsterEncounters, orderedQuestNodes, pointsOfInterest, scenarios]
  );

  useEffect(() => {
    let isMounted = true;
    const loadQuestSummaries = async () => {
      try {
        const result = await apiClient.get<Quest[]>('/sonar/admin/quests');
        if (!isMounted) return;
        setQuests(
          Array.isArray(result) ? result.map(normalizeQuestSummary) : []
        );
        setLoadError(null);
      } catch (error) {
        if (!isMounted) return;
        console.error('Failed to load quests', error);
        setLoadError('Failed to load quests. Check console for details.');
      } finally {
        if (isMounted) {
          setLoading(false);
        }
      }
    };

    loadQuestSummaries();
    return () => {
      isMounted = false;
    };
  }, [apiClient]);

  useEffect(() => {
    let isMounted = true;
    const loadSupportingData = async () => {
      const results = await Promise.allSettled([
        apiClient.get<PointOfInterest[]>('/sonar/pointsOfInterest'),
        apiClient.get<Character[]>('/sonar/characters'),
        apiClient.get<InventoryItem[]>('/sonar/inventory-items'),
        apiClient.get<Spell[]>('/sonar/spells'),
        apiClient.get<ScenarioNodeOption[]>('/sonar/scenarios'),
        apiClient.get<MonsterRecord[]>('/sonar/monsters'),
        apiClient.get<MonsterNodeOption[]>('/sonar/monster-encounters'),
        apiClient.get<ChallengeNodeOption[]>('/sonar/challenges'),
      ]);

      if (!isMounted) return;

      const [
        poiResult,
        charactersResult,
        inventoryResult,
        spellsResult,
        scenariosResult,
        monsterRecordsResult,
        monstersResult,
        challengesResult,
      ] = results;

      if (poiResult.status === 'fulfilled') {
        setPointsOfInterest(poiResult.value);
      } else {
        console.error('Failed to load points of interest', poiResult.reason);
      }

      if (charactersResult.status === 'fulfilled') {
        setCharacters(charactersResult.value);
      } else {
        console.error('Failed to load characters', charactersResult.reason);
      }

      if (inventoryResult.status === 'fulfilled') {
        setInventoryItems(inventoryResult.value);
      } else {
        console.error('Failed to load inventory items', inventoryResult.reason);
      }

      if (spellsResult.status === 'fulfilled') {
        setSpells(spellsResult.value);
      } else {
        console.error('Failed to load spells', spellsResult.reason);
      }

      if (scenariosResult.status === 'fulfilled') {
        setScenarios(
          Array.isArray(scenariosResult.value) ? scenariosResult.value : []
        );
      } else {
        console.error('Failed to load scenarios', scenariosResult.reason);
      }

      if (monsterRecordsResult.status === 'fulfilled') {
        setMonsterRecords(
          Array.isArray(monsterRecordsResult.value)
            ? monsterRecordsResult.value
            : []
        );
      } else {
        console.error('Failed to load monsters', monsterRecordsResult.reason);
      }

      if (monstersResult.status === 'fulfilled') {
        setMonsterEncounters(
          Array.isArray(monstersResult.value) ? monstersResult.value : []
        );
      } else {
        console.error(
          'Failed to load monster encounters',
          monstersResult.reason
        );
      }

      if (challengesResult.status === 'fulfilled') {
        setChallenges(
          Array.isArray(challengesResult.value) ? challengesResult.value : []
        );
      } else {
        console.error('Failed to load challenges', challengesResult.reason);
      }
    };

    loadSupportingData();
    return () => {
      isMounted = false;
    };
  }, [apiClient]);

  useEffect(() => {
    const query = proficiencySearch.trim();
    let active = true;
    const handle = window.setTimeout(async () => {
      try {
        const results = await apiClient.get<string[]>(
          `/sonar/proficiencies?query=${encodeURIComponent(query)}&limit=25`
        );
        if (!active) return;
        setProficiencyOptions(Array.isArray(results) ? results : []);
      } catch (error) {
        if (active) {
          console.error('Failed to load proficiencies', error);
          setProficiencyOptions([]);
        }
      }
    }, 200);
    return () => {
      active = false;
      window.clearTimeout(handle);
    };
  }, [apiClient, proficiencySearch]);

  const refreshPointsOfInterest = async () => {
    try {
      const response = await apiClient.get<PointOfInterest[]>(
        '/sonar/pointsOfInterest'
      );
      setPointsOfInterest(response);
    } catch (error) {
      console.error('Failed to refresh points of interest', error);
    }
  };

  useEffect(() => {
    if (!selectedQuest) return;
    const nextIndex = (selectedQuest.nodes?.length ?? 0) + 1;
    setNodeForm((prev) => ({ ...prev, orderIndex: nextIndex }));
  }, [selectedQuest]);

  useEffect(() => {
    if (!questForm.questGiverCharacterId) {
      setSelectedCharacterLocations([]);
      return;
    }
    let isMounted = true;
    const loadLocations = async () => {
      try {
        const response = await apiClient.get<
          { latitude: number; longitude: number }[]
        >(`/sonar/characters/${questForm.questGiverCharacterId}/locations`);
        if (!isMounted) return;
        setSelectedCharacterLocations(response);
      } catch (error) {
        console.error('Failed to load character locations for map', error);
      }
    };
    loadLocations();
    return () => {
      isMounted = false;
    };
  }, [apiClient, questForm.questGiverCharacterId]);

  useEffect(() => {
    if (!questForm.zoneId) return;
    if (zoneDetailsById[questForm.zoneId]) return;
    let isMounted = true;
    const loadZoneDetail = async () => {
      try {
        const zone = await apiClient.get<{
          boundary?: number[][];
          boundaryCoords?: { latitude: number; longitude: number }[];
          latitude?: number;
          longitude?: number;
        }>(`/sonar/zones/${questForm.zoneId}`);
        if (!isMounted) return;
        console.log('Quest Map: loaded zone details', zone);
        setZoneDetailsById((prev) => ({ ...prev, [questForm.zoneId]: zone }));
      } catch (error) {
        console.error('Failed to load zone details', error);
      }
    };
    loadZoneDetail();
    return () => {
      isMounted = false;
    };
  }, [apiClient, questForm.zoneId, zoneDetailsById]);

  useEffect(() => {
    if (
      (!poiFiltersOpen && !quickCreateOpen.challenge) ||
      zonePoiMapLoaded ||
      zonePoiMapLoading ||
      zones.length === 0
    ) {
      return;
    }
    let isMounted = true;
    const loadZonePoiMap = async () => {
      setZonePoiMapLoading(true);
      const results = await Promise.allSettled(
        zones.map((zone) =>
          apiClient.get<PointOfInterest[]>(
            `/sonar/zones/${zone.id}/pointsOfInterest`
          )
        )
      );
      if (!isMounted) return;
      const nextMap: Record<string, Set<string>> = {};
      results.forEach((result, index) => {
        if (result.status === 'fulfilled') {
          nextMap[zones[index].id] = new Set(result.value.map((poi) => poi.id));
        }
      });
      setZonePoiMap(nextMap);
      setZonePoiMapLoaded(true);
      setZonePoiMapLoading(false);
    };

    loadZonePoiMap();
    return () => {
      isMounted = false;
    };
  }, [
    apiClient,
    poiFiltersOpen,
    quickCreateOpen.challenge,
    zonePoiMapLoaded,
    zonePoiMapLoading,
    zones,
  ]);

  useEffect(() => {
    if (questMapContainer.current && !questMap.current) {
      questMap.current = new mapboxgl.Map({
        container: questMapContainer.current,
        style: 'mapbox://styles/mapbox/streets-v12',
        center: [0, 0],
        zoom: 2,
        interactive: true,
      });

      questMap.current.on('load', () => {
        setQuestMapLoaded(true);
      });
    }

    return () => {
      if (questMap.current) {
        questMap.current.remove();
        questMap.current = null;
      }
    };
  }, [selectedQuest]);

  useEffect(() => {
    if (questMap.current && questMapLoaded) {
      questMap.current.resize();
    }
  }, [questMapLoaded, selectedQuest]);

  const questPolygons = useMemo(() => {
    if (!selectedQuest?.nodes?.length) return [];
    return selectedQuest.nodes
      .filter(
        (node) =>
          node.polygon || (node.polygonPoints && node.polygonPoints.length >= 3)
      )
      .map((node) => ({
        id: node.id,
        orderIndex: node.orderIndex,
        coordinates: normalizePolygonCoordinates(
          node.polygonPoints && node.polygonPoints.length >= 3
            ? [node.polygonPoints]
            : parsePolygonWkt(node.polygon ?? '')
        ),
      }))
      .filter((entry) => entry.coordinates);
  }, [selectedQuest?.nodes]);

  useEffect(() => {
    if (!questMap.current || !questMapLoaded) return;
    const map = questMap.current;
    if (!map.getSource('quest-node-draft-line')) {
      map.addSource('quest-node-draft-line', {
        type: 'geojson',
        data: {
          type: 'Feature',
          geometry: { type: 'LineString', coordinates: [] },
          properties: {},
        },
      });
      map.addLayer({
        id: 'quest-node-draft-line',
        type: 'line',
        source: 'quest-node-draft-line',
        paint: {
          'line-color': '#0f766e',
          'line-width': 2,
          'line-dasharray': [2, 2],
        },
      });
    }

    if (!map.getSource('quest-node-draft-polygon')) {
      map.addSource('quest-node-draft-polygon', {
        type: 'geojson',
        data: {
          type: 'Feature',
          geometry: { type: 'Polygon', coordinates: [] },
          properties: {},
        },
      });
      map.addLayer({
        id: 'quest-node-draft-polygon',
        type: 'fill',
        source: 'quest-node-draft-polygon',
        paint: {
          'fill-color': '#14b8a6',
          'fill-opacity': 0.15,
        },
      });
      map.addLayer({
        id: 'quest-node-draft-polygon-outline',
        type: 'line',
        source: 'quest-node-draft-polygon',
        paint: {
          'line-color': '#0f766e',
          'line-width': 2,
        },
      });
    }

    if (!map.getSource('quest-node-polygons')) {
      map.addSource('quest-node-polygons', {
        type: 'geojson',
        data: {
          type: 'FeatureCollection',
          features: [],
        },
      });
      map.addLayer({
        id: 'quest-node-polygons-fill',
        type: 'fill',
        source: 'quest-node-polygons',
        paint: {
          'fill-color': '#f59e0b',
          'fill-opacity': 0.18,
        },
      });
      map.addLayer({
        id: 'quest-node-polygons-outline',
        type: 'line',
        source: 'quest-node-polygons',
        paint: {
          'line-color': '#b45309',
          'line-width': 2,
        },
      });
    }
  }, [questMapLoaded]);

  useEffect(() => {
    if (!questMap.current || !questMapLoaded) return;
    const map = questMap.current;
    const lineSource = map.getSource('quest-node-draft-line') as
      | mapboxgl.GeoJSONSource
      | undefined;
    const polygonSource = map.getSource('quest-node-draft-polygon') as
      | mapboxgl.GeoJSONSource
      | undefined;

    const lineCoords = polygonDraftPoints;
    if (lineSource) {
      lineSource.setData({
        type: 'Feature',
        geometry: { type: 'LineString', coordinates: lineCoords },
        properties: {},
      });
    }

    if (polygonSource) {
      if (polygonDraftPoints.length >= 3) {
        const ring = [...polygonDraftPoints, polygonDraftPoints[0]];
        polygonSource.setData({
          type: 'Feature',
          geometry: { type: 'Polygon', coordinates: [ring] },
          properties: {},
        });
      } else {
        polygonSource.setData({
          type: 'Feature',
          geometry: { type: 'Polygon', coordinates: [] },
          properties: {},
        });
      }
    }
  }, [polygonDraftPoints, questMapLoaded]);

  useEffect(() => {
    if (!questMap.current || !questMapLoaded) return;
    const map = questMap.current;
    const ensurePolygonSource = () => {
      if (!map.isStyleLoaded()) {
        return undefined;
      }
      let polygonSource = map.getSource('quest-node-polygons') as
        | mapboxgl.GeoJSONSource
        | undefined;
      if (!polygonSource) {
        map.addSource('quest-node-polygons', {
          type: 'geojson',
          data: {
            type: 'FeatureCollection',
            features: [],
          },
        });
        map.addLayer({
          id: 'quest-node-polygons-fill',
          type: 'fill',
          source: 'quest-node-polygons',
          paint: {
            'fill-color': '#f59e0b',
            'fill-opacity': 0.18,
          },
        });
        map.addLayer({
          id: 'quest-node-polygons-outline',
          type: 'line',
          source: 'quest-node-polygons',
          paint: {
            'line-color': '#b45309',
            'line-width': 2,
          },
        });
        polygonSource = map.getSource('quest-node-polygons') as
          | mapboxgl.GeoJSONSource
          | undefined;
      }
      return polygonSource;
    };

    if (!map.isStyleLoaded()) {
      const handleStyleLoad = () => {
        setPolygonRefreshNonce((prev) => prev + 1);
      };
      map.once('style.load', handleStyleLoad);
      return () => {
        map.off('style.load', handleStyleLoad);
      };
    }

    const polygonSource = ensurePolygonSource();
    if (!polygonSource) return;

    const features = questPolygons
      .filter((entry) => entry.coordinates && entry.coordinates.length > 0)
      .map((entry) => ({
        type: 'Feature' as const,
        properties: {
          id: entry.id,
          orderIndex: entry.orderIndex,
        },
        geometry: {
          type: 'Polygon' as const,
          coordinates: entry.coordinates ?? [],
        },
      }));

    console.log('Quest Map: polygon refresh', {
      totalNodes: selectedQuest?.nodes?.length ?? 0,
      polygonCount: questPolygons.length,
      features,
    });

    polygonSource.setData({
      type: 'FeatureCollection',
      features,
    });
  }, [
    questPolygons,
    questMapLoaded,
    polygonRefreshNonce,
    selectedQuest?.nodes?.length,
  ]);

  useEffect(() => {
    if (!questMap.current || !questMapLoaded) return;
    const map = questMap.current;
    const handleClick = (
      event: mapboxgl.MapMouseEvent & mapboxgl.EventData
    ) => {
      if (nodeForm.nodeType !== 'polygon') return;
      const { lng, lat } = event.lngLat;
      setPolygonDraftPoints((prev) => {
        const next = [...prev, [lng, lat]] as [number, number][];
        setNodeForm((formPrev) => ({
          ...formPrev,
          polygonPoints: JSON.stringify(next),
        }));
        return next;
      });
    };
    map.on('click', handleClick);
    return () => {
      map.off('click', handleClick);
    };
  }, [nodeForm.nodeType, questMapLoaded]);

  useEffect(() => {
    if (nodeForm.nodeType !== 'polygon') {
      setPolygonDraftPoints([]);
      setNodeForm((prev) => ({ ...prev, polygonPoints: '' }));
    }
  }, [nodeForm.nodeType]);

  useEffect(() => {
    setSelectedQuestIds((prev) => {
      if (prev.size === 0) return prev;
      const available = new Set(quests.map((quest) => quest.id));
      let changed = false;
      const next = new Set<string>();
      prev.forEach((questId) => {
        if (available.has(questId)) {
          next.add(questId);
        } else {
          changed = true;
        }
      });
      return changed ? next : prev;
    });
  }, [quests]);

  const filteredQuests = useMemo(() => {
    if (!searchQuery.trim()) return quests;
    const term = searchQuery.toLowerCase();
    return quests.filter((quest) => quest.name.toLowerCase().includes(term));
  }, [quests, searchQuery]);

  const allFilteredQuestsSelected = useMemo(
    () =>
      filteredQuests.length > 0 &&
      filteredQuests.every((quest) => selectedQuestIdSet.has(quest.id)),
    [filteredQuests, selectedQuestIdSet]
  );

  const filteredZones = useMemo(() => {
    if (!zoneSearch.trim()) return zones;
    const term = zoneSearch.toLowerCase();
    return zones.filter((zone) => zone.name?.toLowerCase().includes(term));
  }, [zones, zoneSearch]);

  const filteredCharacters = useMemo(() => {
    if (!characterSearch.trim()) return characters;
    const term = characterSearch.toLowerCase();
    return characters.filter((character) =>
      character.name?.toLowerCase().includes(term)
    );
  }, [characters, characterSearch]);

  const allTags = useMemo(() => {
    const tags: Tag[] = [];
    const seen = new Set<string>();
    tagGroups.forEach((group) => {
      group.tags?.forEach((tag) => {
        if (!seen.has(tag.id)) {
          seen.add(tag.id);
          tags.push(tag);
        }
      });
    });
    return tags;
  }, [tagGroups]);

  const filteredTags = useMemo(() => {
    if (!poiTagSearch.trim()) return allTags;
    const term = poiTagSearch.toLowerCase();
    return allTags.filter((tag) => tag.name?.toLowerCase().includes(term));
  }, [allTags, poiTagSearch]);

  const filteredPointsOfInterest = useMemo(() => {
    let filtered = pointsOfInterest;
    if (poiSearch.trim()) {
      const term = poiSearch.toLowerCase();
      filtered = filtered.filter((poi) => {
        const name = poi.name?.toLowerCase() ?? '';
        const googleName = poi.googleMapsPlaceName?.toLowerCase() ?? '';
        const originalName = poi.originalName?.toLowerCase() ?? '';
        return (
          name.includes(term) ||
          googleName.includes(term) ||
          originalName.includes(term)
        );
      });
    }
    if (poiZoneFilterId && zonePoiMap[poiZoneFilterId]) {
      const allowed = zonePoiMap[poiZoneFilterId];
      filtered = filtered.filter((poi) => allowed.has(poi.id));
    }
    if (poiTagFilterIds.length > 0) {
      filtered = filtered.filter((poi) =>
        poi.tags?.some((tag) => poiTagFilterIds.includes(tag.id))
      );
    }
    return filtered;
  }, [
    pointsOfInterest,
    poiSearch,
    poiZoneFilterId,
    zonePoiMap,
    poiTagFilterIds,
  ]);

  const quickCreateChallengePointsOfInterest = useMemo(() => {
    if (!questForm.zoneId) return pointsOfInterest;
    const allowedPoiIds = zonePoiMap[questForm.zoneId];
    if (!allowedPoiIds) return pointsOfInterest;
    return pointsOfInterest.filter((poi) => allowedPoiIds.has(poi.id));
  }, [pointsOfInterest, questForm.zoneId, zonePoiMap]);

  const quickCreateChallengePoiOptions = useMemo(
    () =>
      quickCreateChallengePointsOfInterest.map((poi) => ({
        value: poi.id,
        label: poi.name,
        secondary: [poi.googleMapsPlaceName, poi.originalName]
          .filter(Boolean)
          .join(' · '),
      })),
    [quickCreateChallengePointsOfInterest]
  );

  const filteredScenarios = useMemo(() => {
    let filtered = scenarios;
    if (questForm.zoneId) {
      filtered = filtered.filter(
        (scenario) => scenario.zoneId === questForm.zoneId
      );
    }
    return filtered;
  }, [questForm.zoneId, scenarios]);

  const filteredMonsters = useMemo(() => {
    let filtered = monsterEncounters;
    if (questForm.zoneId) {
      filtered = filtered.filter(
        (monster) => monster.zoneId === questForm.zoneId
      );
    }
    return filtered;
  }, [monsterEncounters, questForm.zoneId]);

  const filteredChallenges = useMemo(() => {
    let filtered = challenges;
    if (questForm.zoneId) {
      filtered = filtered.filter(
        (challenge) => challenge.zoneId === questForm.zoneId
      );
    }
    return filtered;
  }, [challenges, questForm.zoneId]);

  const adminEntityHref = (
    type: 'scenario' | 'monster' | 'challenge',
    id: string
  ) => {
    const basePath =
      type === 'scenario'
        ? '/scenarios'
        : type === 'challenge'
          ? '/challenges'
          : '/monsters';
    return `${basePath}?focus=${encodeURIComponent(id)}`;
  };

  const availableMonstersForQuickCreate = useMemo(() => {
    if (!questForm.zoneId) return monsterRecords;
    return monsterRecords.filter(
      (monster) => monster.zoneId === questForm.zoneId
    );
  }, [monsterRecords, questForm.zoneId]);

  const archetypeByPoiId = useMemo(() => {
    const result: Record<string, LocationArchetype> = {};
    if (!pointsOfInterest.length || !locationArchetypes.length) return result;
    pointsOfInterest.forEach((poi) => {
      const match = matchLocationArchetypeForPoi(poi, locationArchetypes);
      if (match) {
        result[poi.id] = match;
      }
    });
    return result;
  }, [pointsOfInterest, locationArchetypes]);

  const selectedQuestLocationDetails = useMemo(
    () =>
      orderedQuestNodes
        .map((node) => {
          const linkedScenario = resolveLinkedQuestScenario(node, scenarios);
          const linkedMonsterEncounter = resolveLinkedQuestMonsterEncounter(
            node,
            monsterEncounters
          );
          const linkedChallenge = resolveLinkedQuestChallenge(node, challenges);
          const poi = resolveLinkedQuestNodePoi(
            node,
            pointsOfInterest,
            linkedScenario,
            linkedMonsterEncounter,
            linkedChallenge
          );
          if (!poi) return null;
          const archetype = archetypeByPoiId[poi.id] ?? null;
          const aliases = getPointOfInterestAliases(poi);
          const summary =
            summarizePoiText(poi.description) || summarizePoiText(poi.clue);
          const tagNames = (poi.tags ?? []).map((tag) => tag.name).filter(Boolean);
          return {
            nodeId: node.id,
            orderIndex: node.orderIndex,
            poi,
            archetype,
            aliases,
            summary,
            tagNames,
            latitude: formatPoiCoordinate(poi.lat),
            longitude: formatPoiCoordinate(poi.lng),
          };
        })
        .filter(
          (
            entry
          ): entry is {
            nodeId: string;
            orderIndex: number;
            poi: PointOfInterest;
            archetype: LocationArchetype | null;
            aliases: string[];
            summary: string;
            tagNames: string[];
            latitude: string | null;
            longitude: string | null;
          } => Boolean(entry)
        ),
    [
      archetypeByPoiId,
      challenges,
      monsterEncounters,
      orderedQuestNodes,
      pointsOfInterest,
      scenarios,
    ]
  );

  useEffect(() => {
    if (!selectedQuest?.nodes?.length) return;
    if (!locationArchetypes.length) return;

    setChallengeDrafts((prev) => {
      let changed = false;
      const next = { ...prev };
      selectedQuest.nodes?.forEach((node) => {
        const linkedScenario = resolveLinkedQuestScenario(node, scenarios);
        const linkedMonsterEncounter = resolveLinkedQuestMonsterEncounter(
          node,
          monsterEncounters
        );
        const linkedChallenge = resolveLinkedQuestChallenge(node, challenges);
        const linkedPoiId = getLinkedQuestNodePoiId(
          node,
          linkedScenario,
          linkedMonsterEncounter,
          linkedChallenge
        );
        if (!linkedPoiId) return;
        const match = archetypeByPoiId[linkedPoiId];
        if (!match) return;
        const existing = next[node.id];
        if (existing?.locationArchetypeId) return;
        next[node.id] = {
          ...emptyChallengeForm,
          ...existing,
          locationArchetypeId: match.id,
        };
        changed = true;
      });
      return changed ? next : prev;
    });
  }, [
    archetypeByPoiId,
    challenges,
    locationArchetypes.length,
    monsterEncounters,
    scenarios,
    selectedQuest?.nodes,
  ]);

  const questNodePoints = useMemo(() => {
    if (!selectedQuest?.nodes?.length) return [];
    return selectedQuest.nodes
      .map((node) => {
        const linkedScenario = resolveLinkedQuestScenario(node, scenarios);
        const linkedMonsterEncounter = resolveLinkedQuestMonsterEncounter(
          node,
          monsterEncounters
        );
        const linkedChallenge = resolveLinkedQuestChallenge(node, challenges);
        const linkedPoi = resolveLinkedQuestNodePoi(
          node,
          pointsOfInterest,
          linkedScenario,
          linkedMonsterEncounter,
          linkedChallenge
        );

        if (linkedPoi) {
          const poi = linkedPoi;
          const lng = Number(poi.lng);
          const lat = Number(poi.lat);
          if (Number.isNaN(lng) || Number.isNaN(lat)) return null;
          let name = poi.name;
          let nodeType: QuestNodeType = getQuestNodeKind(node);
          if (node.scenarioId) {
            name = summarizeScenarioPrompt(linkedScenario?.prompt ?? '');
            nodeType = 'scenario';
          } else if (node.monsterEncounterId || node.monsterId) {
            name =
              linkedMonsterEncounter?.name ??
              node.monsterEncounterId ??
              node.monsterId ??
              poi.name;
            nodeType = 'monster';
          } else if (node.challengeId) {
            name = linkedChallenge?.question || node.challengeId || poi.name;
            nodeType = 'challenge';
          }
          return {
            id: node.id,
            name,
            orderIndex: node.orderIndex,
            lng,
            lat,
            nodeType,
          };
        }
        if (node.scenarioId) {
          const scenario = linkedScenario;
          if (!scenario) return null;
          const lng = Number(scenario.longitude);
          const lat = Number(scenario.latitude);
          if (Number.isNaN(lng) || Number.isNaN(lat)) return null;
          return {
            id: node.id,
            name: summarizeScenarioPrompt(scenario.prompt),
            orderIndex: node.orderIndex,
            lng,
            lat,
            nodeType: 'scenario' as QuestNodeType,
          };
        }
        if (node.monsterEncounterId || node.monsterId) {
          const resolved = linkedMonsterEncounter;
          if (!resolved) return null;
          const lng = Number(resolved.longitude);
          const lat = Number(resolved.latitude);
          if (Number.isNaN(lng) || Number.isNaN(lat)) return null;
          return {
            id: node.id,
            name: resolved.name || resolved.id,
            orderIndex: node.orderIndex,
            lng,
            lat,
            nodeType: 'monster' as QuestNodeType,
          };
        }
        if (node.challengeId) {
          const challenge = linkedChallenge;
          if (!challenge) return null;
          const lng = Number(challenge.longitude);
          const lat = Number(challenge.latitude);
          if (Number.isNaN(lng) || Number.isNaN(lat)) return null;
          return {
            id: node.id,
            name: challenge.question || challenge.id,
            orderIndex: node.orderIndex,
            lng,
            lat,
            nodeType: 'challenge' as QuestNodeType,
          };
        }
        return null;
      })
      .filter(
        (
          entry
        ): entry is {
          id: string;
          name: string;
          orderIndex: number;
          lng: number;
          lat: number;
          nodeType: QuestNodeType;
        } => Boolean(entry)
      );
  }, [
    challenges,
    monsterEncounters,
    pointsOfInterest,
    scenarios,
    selectedQuest?.nodes,
  ]);

  const focusQuestArea = useCallback(() => {
    if (!questMap.current) return;

    const bounds = new mapboxgl.LngLatBounds();
    let hasBounds = false;

    questNodePoints.forEach((point) => {
      bounds.extend([point.lng, point.lat]);
      hasBounds = true;
    });

    questPolygons.forEach((polygon) => {
      polygon.coordinates?.forEach((ring) => {
        ring.forEach((coord) => {
          bounds.extend([coord[0], coord[1]]);
          hasBounds = true;
        });
      });
    });

    if (!hasBounds && questForm.zoneId) {
      const zone =
        zoneDetailsById[questForm.zoneId] ||
        zones.find((z) => z.id === questForm.zoneId);
      if (zone) {
        const lng =
          typeof zone.longitude === 'number'
            ? zone.longitude
            : Number(zone.longitude);
        const lat =
          typeof zone.latitude === 'number'
            ? zone.latitude
            : Number(zone.latitude);
        if (!Number.isNaN(lng) && !Number.isNaN(lat)) {
          questMap.current.setCenter([lng, lat]);
          questMap.current.setZoom(13);
        }
      }
      return;
    }

    if (!hasBounds) return;

    if (
      questNodePoints.length === 1 &&
      questPolygons.length === 0
    ) {
      const point = questNodePoints[0];
      questMap.current.easeTo({
        center: [point.lng, point.lat],
        zoom: 16,
        duration: 600,
      });
      return;
    }

    questMap.current.fitBounds(bounds, {
      padding: 64,
      maxZoom: 16,
      duration: 600,
    });
  }, [
    questForm.zoneId,
    questNodePoints,
    questPolygons,
    zoneDetailsById,
    zones,
  ]);

  useEffect(() => {
    if (!questMap.current || !questMapLoaded) return;

    questNodeMarkers.current.forEach((marker) => marker.remove());
    questNodeMarkers.current = [];

    const bounds = new mapboxgl.LngLatBounds();
    let hasBounds = false;

    questNodePoints.forEach((point) => {
      const el = document.createElement('div');
      el.style.width = '12px';
      el.style.height = '12px';
      el.style.borderRadius = '9999px';
      if (point.nodeType === 'scenario') {
        el.style.background = '#14b8a6';
        el.style.border = '2px solid #115e59';
      } else if (point.nodeType === 'monster') {
        el.style.background = '#ef4444';
        el.style.border = '2px solid #7f1d1d';
      } else {
        el.style.background = '#f59e0b';
        el.style.border = '2px solid #92400e';
      }
      const labelPrefix =
        point.nodeType === 'scenario'
          ? 'Scenario'
          : point.nodeType === 'monster'
            ? 'Monster'
            : point.nodeType === 'challenge'
              ? 'Challenge'
              : 'Location';
      const marker = new mapboxgl.Marker({ element: el })
        .setLngLat([point.lng, point.lat])
        .setPopup(
          new mapboxgl.Popup({ offset: 12 }).setText(
            `Node ${point.orderIndex} (${labelPrefix}): ${point.name}`
          )
        )
        .addTo(questMap.current!);
      questNodeMarkers.current.push(marker);
      bounds.extend([point.lng, point.lat]);
      hasBounds = true;
    });

    const polygonFeatures = questPolygons
      .map((polygon) => {
        if (!polygon.coordinates) return null;
        polygon.coordinates.forEach((ring) => {
          ring.forEach((coord) => {
            bounds.extend([coord[0], coord[1]]);
            hasBounds = true;
          });
        });
        return {
          type: 'Feature' as const,
          geometry: {
            type: 'Polygon' as const,
            coordinates: polygon.coordinates,
          },
          properties: {
            id: polygon.id,
          },
        };
      })
      .filter(Boolean);

    const sourceId = 'quest-polygons';
    const fillLayerId = 'quest-polygons-fill';
    const lineLayerId = 'quest-polygons-line';
    const existingSource = questMap.current.getSource(sourceId);
    if (existingSource) {
      (existingSource as mapboxgl.GeoJSONSource).setData({
        type: 'FeatureCollection',
        features: polygonFeatures as GeoJSON.Feature[],
      });
    } else {
      questMap.current.addSource(sourceId, {
        type: 'geojson',
        data: {
          type: 'FeatureCollection',
          features: polygonFeatures as GeoJSON.Feature[],
        },
      });
      questMap.current.addLayer({
        id: fillLayerId,
        type: 'fill',
        source: sourceId,
        paint: {
          'fill-color': '#f59e0b',
          'fill-opacity': 0.15,
        },
      });
      questMap.current.addLayer({
        id: lineLayerId,
        type: 'line',
        source: sourceId,
        paint: {
          'line-color': '#f59e0b',
          'line-width': 2,
        },
      });
    }

    if (hasBounds) {
      if (questNodePoints.length === 1 && questPolygons.length === 0) {
        const point = questNodePoints[0];
        questMap.current.easeTo({
          center: [point.lng, point.lat],
          zoom: 16,
          duration: 600,
        });
      } else {
        questMap.current.fitBounds(bounds, {
          padding: 64,
          maxZoom: 16,
          duration: 600,
        });
      }
    } else if (questForm.zoneId) {
      focusQuestArea();
    }
  }, [
    focusQuestArea,
    questForm.zoneId,
    questMapLoaded,
    questNodePoints,
    questPolygons,
    zones,
    zoneDetailsById,
  ]);

  const updateQuestState = (
    questId: string,
    updater: (quest: Quest) => Quest
  ) => {
    setQuests((prev) =>
      prev.map((quest) => (quest.id === questId ? updater(quest) : quest))
    );
  };

  const handleCreateQuest = async () => {
    try {
      const payload = {
        name: questForm.name,
        description: questForm.description,
        acceptanceDialogue: normalizeAcceptanceDialogue(
          questForm.acceptanceDialogue
        ),
        zoneId: questForm.zoneId || null,
        questGiverCharacterId: questForm.questGiverCharacterId || null,
        questArchetypeId: questForm.questArchetypeId || null,
        recurrenceFrequency: questForm.recurrenceFrequency || '',
        difficultyMode: questForm.difficultyMode,
        difficulty: Math.max(1, Number(questForm.difficulty) || 1),
        monsterEncounterTargetLevel: Math.max(
          1,
          Number(questForm.monsterEncounterTargetLevel) || 1
        ),
        rewardMode: questForm.rewardMode,
        randomRewardSize: questForm.randomRewardSize,
        rewardExperience:
          questForm.rewardMode === 'explicit'
            ? Number(questForm.rewardExperience) || 0
            : 0,
        gold:
          questForm.rewardMode === 'explicit' ? Number(questForm.gold) || 0 : 0,
        materialRewards:
          questForm.rewardMode === 'explicit'
            ? normalizeMaterialRewards(questForm.materialRewards)
            : [],
        itemRewards:
          questForm.rewardMode === 'explicit'
            ? questForm.itemRewards
                .map((reward) => ({
                  inventoryItemId: Number(reward.inventoryItemId) || 0,
                  quantity: Number(reward.quantity) || 0,
                }))
                .filter(
                  (reward) => reward.inventoryItemId > 0 && reward.quantity > 0
                )
            : [],
        spellRewards:
          questForm.rewardMode === 'explicit'
            ? questForm.spellRewards
                .map((reward) => ({ spellId: reward.spellId.trim() }))
                .filter((reward) => reward.spellId.length > 0)
            : [],
      };
      const created = await apiClient.post<Quest>('/sonar/quests', payload);
      const normalizedQuest = normalizeQuestDetail(created);
      setQuests((prev) => [normalizedQuest, ...prev]);
      setActiveQuestId(normalizedQuest.id);
      setQuestForm(buildQuestFormFromQuest(normalizedQuest));
      setShowCreateQuest(false);
    } catch (error) {
      console.error('Failed to create quest', error);
      alert('Failed to create quest. Please check the required fields.');
    }
  };

  const handleUpdateQuest = async () => {
    if (!selectedQuest) return;
    try {
      const payload = {
        name: questForm.name,
        description: questForm.description,
        acceptanceDialogue: normalizeAcceptanceDialogue(
          questForm.acceptanceDialogue
        ),
        zoneId: questForm.zoneId || null,
        questGiverCharacterId: questForm.questGiverCharacterId || null,
        questArchetypeId: questForm.questArchetypeId || null,
        recurrenceFrequency: questForm.recurrenceFrequency || '',
        difficultyMode: questForm.difficultyMode,
        difficulty: Math.max(1, Number(questForm.difficulty) || 1),
        monsterEncounterTargetLevel: Math.max(
          1,
          Number(questForm.monsterEncounterTargetLevel) || 1
        ),
        rewardMode: questForm.rewardMode,
        randomRewardSize: questForm.randomRewardSize,
        rewardExperience:
          questForm.rewardMode === 'explicit'
            ? Number(questForm.rewardExperience) || 0
            : 0,
        gold:
          questForm.rewardMode === 'explicit' ? Number(questForm.gold) || 0 : 0,
        materialRewards:
          questForm.rewardMode === 'explicit'
            ? normalizeMaterialRewards(questForm.materialRewards)
            : [],
        itemRewards:
          questForm.rewardMode === 'explicit'
            ? questForm.itemRewards
                .map((reward) => ({
                  inventoryItemId: Number(reward.inventoryItemId) || 0,
                  quantity: Number(reward.quantity) || 0,
                }))
                .filter(
                  (reward) => reward.inventoryItemId > 0 && reward.quantity > 0
                )
            : [],
        spellRewards:
          questForm.rewardMode === 'explicit'
            ? questForm.spellRewards
                .map((reward) => ({ spellId: reward.spellId.trim() }))
                .filter((reward) => reward.spellId.length > 0)
            : [],
      };
      const updated = await apiClient.patch<Quest>(
        `/sonar/quests/${selectedQuest.id}`,
        payload
      );
      const normalizedQuest = normalizeQuestDetail(updated);
      updateQuestState(selectedQuest.id, () => normalizedQuest);
      setQuestForm(buildQuestFormFromQuest(normalizedQuest));
    } catch (error) {
      console.error('Failed to update quest', error);
      alert('Failed to update quest.');
    }
  };

  const handleCreateQuestArchetypeFromQuest = async () => {
    if (!selectedQuest || creatingArchetype) return;

    const nodes = (selectedQuest.nodes ?? [])
      .slice()
      .sort((a, b) => a.orderIndex - b.orderIndex);
    if (nodes.length === 0) {
      alert('Quest has no nodes to convert into an archetype.');
      return;
    }

    if (!locationArchetypes.length) {
      alert(
        'Location archetypes are still loading. Please try again in a moment.'
      );
      return;
    }

    const missing: string[] = [];
    const nodeLocationArchetypeIds: string[] = [];
    const linkedChallengesForNodes: (ChallengeNodeOption | undefined)[] = [];
    nodes.forEach((node) => {
      const linkedScenario = resolveLinkedQuestScenario(node, scenarios) ?? undefined;
      const linkedMonsterEncounter =
        resolveLinkedQuestMonsterEncounter(node, monsterEncounters) ?? undefined;
      const linkedChallenge = resolveLinkedQuestChallenge(node, challenges) ?? undefined;
      linkedChallengesForNodes.push(linkedChallenge ?? undefined);
      const sourcePointOfInterestId = getLinkedQuestNodePoiId(
        node,
        linkedScenario,
        linkedMonsterEncounter,
        linkedChallenge
      );
      if (!sourcePointOfInterestId) {
        if (node.scenarioId) {
          missing.push(`Node ${node.orderIndex}: scenario node`);
        } else if (node.monsterEncounterId || node.monsterId) {
          missing.push(`Node ${node.orderIndex}: monster node`);
        } else if (node.challengeId) {
          missing.push(
            `Node ${node.orderIndex}: ${linkedChallenge?.question ?? 'challenge objective'}`
          );
        } else {
          missing.push(`Node ${node.orderIndex}: polygon node`);
        }
        return;
      }
      const match = archetypeByPoiId[sourcePointOfInterestId];
      if (!match) {
        const poiName = pointsOfInterest.find(
          (poi) => poi.id === sourcePointOfInterestId
        )?.name;
        missing.push(
          `Node ${node.orderIndex}: ${poiName ?? sourcePointOfInterestId}`
        );
        return;
      }
      nodeLocationArchetypeIds.push(match.id);
    });

    if (missing.length > 0) {
      alert(
        `Cannot create quest archetype. Missing location archetypes for:\n${missing.join('\n')}`
      );
      return;
    }

    const name = selectedQuest.name;
    const rewardMode = getQuestRewardMode(selectedQuest);
    const itemRewards = (selectedQuest.itemRewards ?? [])
      .map((reward) => ({
        inventoryItemId: reward.inventoryItemId,
        quantity: reward.quantity ?? 0,
      }))
      .filter((reward) => reward.inventoryItemId && reward.quantity > 0);
    const spellRewards = (selectedQuest.spellRewards ?? [])
      .map((reward) => ({
        spellId: reward.spellId,
      }))
      .filter((reward) => reward.spellId);

    setCreatingArchetype(true);
    try {
      const rootNode = await apiClient.post<QuestArchetypeNode>(
        '/sonar/questArchetypeNodes',
        {
          locationArchetypeID: nodeLocationArchetypeIds[0],
        }
      );

      const archetype = await apiClient.post<QuestArchetype>(
        '/sonar/questArchetypes',
        {
          name,
          description: selectedQuest.description ?? '',
          acceptanceDialogue: selectedQuest.acceptanceDialogue ?? [],
          imageUrl: selectedQuest.imageUrl ?? '',
          rootId: rootNode.id,
          difficultyMode:
            selectedQuest.difficultyMode === 'fixed' ? 'fixed' : 'scale',
          difficulty: selectedQuest.difficulty ?? 1,
          monsterEncounterTargetLevel:
            selectedQuest.monsterEncounterTargetLevel ?? 1,
          defaultGold: selectedQuest.gold ?? 0,
          rewardMode: rewardMode,
          randomRewardSize: selectedQuest.randomRewardSize ?? 'small',
          rewardExperience: selectedQuest.rewardExperience ?? 0,
          recurrenceFrequency: selectedQuest.recurrenceFrequency ?? null,
          materialRewards: selectedQuest.materialRewards ?? [],
          itemRewards: itemRewards.length > 0 ? itemRewards : undefined,
          spellRewards: spellRewards.length > 0 ? spellRewards : undefined,
        }
      );

      let currentNodeId = rootNode.id;

      for (let index = 0; index < nodes.length; index += 1) {
        const node = nodes[index];
        const linkedChallenge = linkedChallengesForNodes[index];
        const hasNext = index < nodes.length - 1;
        const nextLocationArchetypeId = hasNext
          ? nodeLocationArchetypeIds[index + 1]
          : null;
        const currentLocationArchetypeId = nodeLocationArchetypeIds[index];
        if (node.challengeId) {
          if (!linkedChallenge) {
            throw new Error(
              `Missing linked challenge data for node ${node.orderIndex}.`
            );
          }
          if (!currentLocationArchetypeId) {
            throw new Error(
              `Missing location archetype for challenge node ${node.orderIndex}.`
            );
          }
          const template = await apiClient.post<{ id: string }>(
            '/sonar/challenge-templates',
            {
              locationArchetypeId: currentLocationArchetypeId,
              question: linkedChallenge.question,
              description: linkedChallenge.description ?? '',
              imageUrl: '',
              thumbnailUrl: '',
              scaleWithUserLevel: false,
              rewardMode: 'random',
              randomRewardSize: 'small',
              rewardExperience: 0,
              reward: 0,
              inventoryItemId: null,
              itemChoiceRewards: [],
              submissionType: linkedChallenge.submissionType ?? 'photo',
              difficulty: linkedChallenge.difficulty ?? 0,
              statTags: linkedChallenge.statTags ?? [],
              proficiency: linkedChallenge.proficiency ?? '',
            }
          );
          const created = await apiClient.post<QuestArchetypeChallenge>(
            `/sonar/questArchetypes/${currentNodeId}/challenges`,
            {
              challengeTemplateId: template.id,
              locationArchetypeID: nextLocationArchetypeId ?? undefined,
            }
          );
          if (hasNext) {
            if (!created.unlockedNodeId) {
              throw new Error('Failed to create next archetype node.');
            }
            currentNodeId = created.unlockedNodeId;
          }
          continue;
        }

        if (hasNext) {
          const created = await apiClient.post<QuestArchetypeChallenge>(
            `/sonar/questArchetypes/${currentNodeId}/challenges`,
            {
              reward: 0,
              difficulty: 0,
              locationArchetypeID: nextLocationArchetypeId ?? undefined,
            }
          );
          if (!created.unlockedNodeId) {
            throw new Error('Failed to create next archetype node.');
          }
          currentNodeId = created.unlockedNodeId;
        }
      }

      setQuestForm((prev) => ({ ...prev, questArchetypeId: archetype.id }));
      updateQuestState(selectedQuest.id, (quest) => ({
        ...quest,
        questArchetypeId: archetype.id,
      }));
      alert(
        'Quest archetype created. Click Save Changes to link it to this quest.'
      );
    } catch (error) {
      console.error('Failed to create quest archetype from quest', error);
      alert('Failed to create quest archetype from quest.');
    } finally {
      setCreatingArchetype(false);
    }
  };

  const handleDeleteQuest = async () => {
    if (!selectedQuest || bulkDeletingQuests) return;
    const confirmDelete = window.confirm(
      `Delete quest "${selectedQuest.name}"? This cannot be undone.`
    );
    if (!confirmDelete) return;

    setDeletingQuestId(selectedQuest.id);
    try {
      await apiClient.delete(`/sonar/quests/${selectedQuest.id}`);
      setQuests((prev) =>
        prev.filter((quest) => quest.id !== selectedQuest.id)
      );
      setSelectedQuestIds((prev) => {
        const next = new Set(prev);
        next.delete(selectedQuest.id);
        return next;
      });
      setActiveQuestId('');
      setQuestForm({ ...emptyQuestForm });
    } catch (error) {
      console.error('Failed to delete quest', error);
      alert('Failed to delete quest.');
    } finally {
      setDeletingQuestId(null);
    }
  };

  const handleDeleteQuestById = async (quest: Quest) => {
    if (bulkDeletingQuests) return;
    const confirmDelete = window.confirm(
      `Delete quest "${quest.name}"? This cannot be undone.`
    );
    if (!confirmDelete) return;

    setDeletingQuestId(quest.id);
    try {
      await apiClient.delete(`/sonar/quests/${quest.id}`);
      setQuests((prev) => prev.filter((item) => item.id !== quest.id));
      setSelectedQuestIds((prev) => {
        const next = new Set(prev);
        next.delete(quest.id);
        return next;
      });
      if (selectedQuestId === quest.id) {
        setActiveQuestId('');
        setQuestForm({ ...emptyQuestForm });
      }
    } catch (error) {
      console.error('Failed to delete quest', error);
      alert('Failed to delete quest.');
    } finally {
      setDeletingQuestId(null);
    }
  };

  const toggleQuestSelection = (questId: string) => {
    setSelectedQuestIds((prev) => {
      const next = new Set(prev);
      if (next.has(questId)) {
        next.delete(questId);
      } else {
        next.add(questId);
      }
      return next;
    });
  };

  const toggleSelectVisibleQuests = () => {
    if (filteredQuests.length === 0) return;
    setSelectedQuestIds((prev) => {
      const next = new Set(prev);
      if (allFilteredQuestsSelected) {
        filteredQuests.forEach((quest) => next.delete(quest.id));
      } else {
        filteredQuests.forEach((quest) => next.add(quest.id));
      }
      return next;
    });
  };

  const clearQuestSelection = () => {
    setSelectedQuestIds(new Set());
  };

  const handleBulkDeleteQuests = async () => {
    if (bulkDeletingQuests || selectedQuestIds.size === 0 || deletingQuestId)
      return;

    const selectedIds = Array.from(selectedQuestIds);
    const selectedNames = quests
      .filter((quest) => selectedQuestIds.has(quest.id))
      .map((quest) => quest.name);
    const preview = selectedNames.slice(0, 5).join(', ');
    const moreCount = Math.max(0, selectedNames.length - 5);
    const confirmMessage =
      selectedIds.length === 1
        ? `Delete 1 selected quest (${preview})? This cannot be undone.`
        : `Delete ${selectedIds.length} selected quests${
            preview
              ? ` (${preview}${moreCount > 0 ? ` +${moreCount} more` : ''})`
              : ''
          }? This cannot be undone.`;

    if (!window.confirm(confirmMessage)) return;

    setBulkDeletingQuests(true);
    try {
      const results = await Promise.allSettled(
        selectedIds.map((questId) =>
          apiClient.delete(`/sonar/quests/${questId}`)
        )
      );
      const deletedIds = new Set<string>();
      const failedIds: string[] = [];
      results.forEach((result, index) => {
        const questId = selectedIds[index];
        if (result.status === 'fulfilled') {
          deletedIds.add(questId);
        } else {
          console.error(`Failed to delete quest ${questId}`, result.reason);
          failedIds.push(questId);
        }
      });

      if (deletedIds.size > 0) {
        setQuests((prev) => prev.filter((quest) => !deletedIds.has(quest.id)));
        setSelectedQuestIds((prev) => {
          const next = new Set(prev);
          deletedIds.forEach((questId) => next.delete(questId));
          return next;
        });
        if (selectedQuestId && deletedIds.has(selectedQuestId)) {
          setActiveQuestId('');
          setQuestForm({ ...emptyQuestForm });
        }
      }

      if (failedIds.length > 0) {
        alert(
          `Deleted ${deletedIds.size} quest${deletedIds.size === 1 ? '' : 's'}, but failed to delete ${
            failedIds.length
          }. Check console for details.`
        );
      }
    } catch (error) {
      console.error('Failed to bulk delete quests', error);
      alert('Failed to delete selected quests.');
    } finally {
      setBulkDeletingQuests(false);
    }
  };

  const handleSelectQuest = (quest: Quest) => {
    setActiveQuestId(quest.id);
    setQuestDetailErrorId(null);
    if (quest.detailLoaded) {
      setQuestDetailLoadingId(null);
      setQuestForm(buildQuestFormFromQuest(quest));
      return;
    }

    setQuestDetailLoadingId(quest.id);
    void (async () => {
      try {
        const detail = normalizeQuestDetail(
          await apiClient.get<Quest>(`/sonar/quests/${quest.id}`)
        );
        setQuests((prev) =>
          prev.map((item) => (item.id === detail.id ? detail : item))
        );
        if (selectedQuestIdRef.current === detail.id) {
          setQuestDetailErrorId(null);
          setQuestForm(buildQuestFormFromQuest(detail));
        }
      } catch (error) {
        console.error('Failed to load quest details', error);
        if (selectedQuestIdRef.current === quest.id) {
          setQuestDetailErrorId(quest.id);
        }
      } finally {
        if (selectedQuestIdRef.current === quest.id) {
          setQuestDetailLoadingId(null);
        }
      }
    })();
  };

  const handleAddQuestReward = () => {
    setQuestForm((prev) => ({
      ...prev,
      itemRewards: [...prev.itemRewards, { ...emptyQuestReward }],
    }));
  };

  const handleUpdateQuestReward = (
    index: number,
    updates: Partial<{ inventoryItemId: string; quantity: number }>
  ) => {
    setQuestForm((prev) => ({
      ...prev,
      itemRewards: prev.itemRewards.map((reward, rewardIndex) =>
        rewardIndex === index ? { ...reward, ...updates } : reward
      ),
    }));
  };

  const handleRemoveQuestReward = (index: number) => {
    setQuestForm((prev) => ({
      ...prev,
      itemRewards: prev.itemRewards.filter(
        (_, rewardIndex) => rewardIndex !== index
      ),
    }));
  };

  const handleAddQuestSpellReward = () => {
    setQuestForm((prev) => ({
      ...prev,
      spellRewards: [...prev.spellRewards, { ...emptyQuestSpellReward }],
    }));
  };

  const handleUpdateQuestSpellReward = (
    index: number,
    updates: Partial<{ spellId: string }>
  ) => {
    setQuestForm((prev) => ({
      ...prev,
      spellRewards: prev.spellRewards.map((reward, rewardIndex) =>
        rewardIndex === index ? { ...reward, ...updates } : reward
      ),
    }));
  };

  const handleRemoveQuestSpellReward = (index: number) => {
    setQuestForm((prev) => ({
      ...prev,
      spellRewards: prev.spellRewards.filter(
        (_, rewardIndex) => rewardIndex !== index
      ),
    }));
  };

  const handleCreateNode = async () => {
    if (!selectedQuest) return;
    try {
      if (nodeForm.nodeType === 'poi' || nodeForm.nodeType === 'polygon') {
        alert('Quest nodes now must be Scenario, Monster, or Challenge.');
        return;
      }
      const polygonPoints =
        nodeForm.nodeType === 'polygon'
          ? parsePolygonPoints(nodeForm.polygonPoints)
          : null;
      if (nodeForm.nodeType === 'polygon' && !polygonPoints) {
        alert('Please enter polygon points as JSON: [[lng,lat],[lng,lat],...]');
        return;
      }
      const payload = {
        orderIndex: Number(nodeForm.orderIndex) || 1,
        pointOfInterestId:
          nodeForm.nodeType === 'poi'
            ? nodeForm.pointOfInterestId || null
            : null,
        scenarioId:
          nodeForm.nodeType === 'scenario' ? nodeForm.scenarioId || null : null,
        monsterId: null,
        monsterEncounterId:
          nodeForm.nodeType === 'monster'
            ? nodeForm.monsterEncounterId || null
            : null,
        challengeId:
          nodeForm.nodeType === 'challenge'
            ? nodeForm.challengeId || null
            : null,
        polygonPoints:
          nodeForm.nodeType === 'polygon' ? polygonPoints : undefined,
        submissionType: nodeForm.submissionType,
      };
      const created = await apiClient.post<QuestNode>('/sonar/questNodes', {
        ...payload,
        questId: selectedQuest.id,
      });
      updateQuestState(selectedQuest.id, (quest) => ({
        ...quest,
        nodes: [...(quest.nodes ?? []), created].sort(
          (a, b) => a.orderIndex - b.orderIndex
        ),
      }));
      setNodeForm({
        ...emptyNodeForm,
        orderIndex: (selectedQuest.nodes?.length ?? 0) + 2,
      });
    } catch (error) {
      console.error('Failed to create quest node', error);
      alert('Failed to create quest node.');
    }
  };

  const toggleQuickCreate = (type: 'scenario' | 'monster' | 'challenge') => {
    setQuickCreateOpen((prev) => ({ ...prev, [type]: !prev[type] }));
  };

  const handleAddQuickScenarioOption = () => {
    setQuickCreateScenarioForm((prev) => ({
      ...prev,
      options: [...prev.options, createEmptyQuickScenarioOption()],
    }));
  };

  const handleUpdateQuickScenarioOption = (
    index: number,
    updates: Partial<QuickCreateScenarioOptionForm>
  ) => {
    setQuickCreateScenarioForm((prev) => ({
      ...prev,
      options: prev.options.map((option, optionIndex) =>
        optionIndex === index ? { ...option, ...updates } : option
      ),
    }));
  };

  const handleRemoveQuickScenarioOption = (index: number) => {
    setQuickCreateScenarioForm((prev) => ({
      ...prev,
      options:
        prev.options.length <= 1
          ? prev.options
          : prev.options.filter((_, optionIndex) => optionIndex !== index),
    }));
  };

  const handleCreateStandaloneScenario = async () => {
    if (!questForm.zoneId) {
      alert('Select a zone for the quest before creating a scenario.');
      return;
    }
    const latitude = Number.parseFloat(quickCreateScenarioForm.latitude);
    const longitude = Number.parseFloat(quickCreateScenarioForm.longitude);
    if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
      alert('Scenario latitude and longitude are required.');
      return;
    }
    const options = quickCreateScenarioForm.options
      .map((option) => ({
        optionText: option.optionText.trim(),
        statTag: option.statTag,
        proficiencies: option.proficiencies
          .split(',')
          .map((entry) => entry.trim())
          .filter((entry) => entry.length > 0),
        difficulty: parseIntSafe(option.difficulty, 0),
        successText: option.successText.trim(),
        failureText: option.failureText.trim(),
        rewardExperience: 0,
        rewardGold: 0,
        itemRewards: [],
        itemChoiceRewards: [],
        spellRewards: [],
      }))
      .filter((option) => option.optionText.length > 0);
    if (
      !quickCreateScenarioForm.prompt.trim() ||
      !quickCreateScenarioForm.imageUrl.trim()
    ) {
      alert('Scenario prompt and image URL are required.');
      return;
    }
    if (options.length === 0) {
      alert('Add at least one scenario option.');
      return;
    }

    setQuickCreateSubmitting('scenario');
    try {
      const created = await apiClient.post<
        ScenarioNodeOption & { attemptedByUser?: boolean }
      >('/sonar/scenarios', {
        zoneId: questForm.zoneId,
        latitude,
        longitude,
        prompt: quickCreateScenarioForm.prompt.trim(),
        imageUrl: quickCreateScenarioForm.imageUrl.trim(),
        thumbnailUrl:
          quickCreateScenarioForm.thumbnailUrl.trim() ||
          quickCreateScenarioForm.imageUrl.trim(),
        rewardMode: 'random',
        randomRewardSize: 'small',
        openEnded: false,
        scaleWithUserLevel: false,
        failurePenaltyMode: 'shared',
        failureHealthDrainType: 'flat',
        failureHealthDrainValue: 0,
        failureManaDrainType: 'flat',
        failureManaDrainValue: 0,
        failureStatuses: [],
        successRewardMode: 'shared',
        successHealthRestoreType: 'flat',
        successHealthRestoreValue: 0,
        successManaRestoreType: 'flat',
        successManaRestoreValue: 0,
        successStatuses: [],
        options,
        itemRewards: [],
        itemChoiceRewards: [],
        spellRewards: [],
      });
      setScenarios((prev) => [created, ...prev]);
      setNodeForm((prev) => ({ ...prev, scenarioId: created.id }));
      setQuickCreateScenarioForm(emptyQuickCreateScenarioForm());
      setQuickCreateOpen((prev) => ({ ...prev, scenario: false }));
    } catch (error) {
      console.error('Failed to create scenario', error);
      alert(
        error instanceof Error ? error.message : 'Failed to create scenario.'
      );
    } finally {
      setQuickCreateSubmitting(null);
    }
  };

  const handleCreateStandaloneChallenge = async () => {
    if (!questForm.zoneId) {
      alert('Select a zone for the quest before creating a challenge.');
      return;
    }
    const pointOfInterestId = quickCreateChallengeForm.pointOfInterestId.trim();
    const latitude = Number.parseFloat(quickCreateChallengeForm.latitude);
    const longitude = Number.parseFloat(quickCreateChallengeForm.longitude);
    if (
      !pointOfInterestId &&
      (!Number.isFinite(latitude) || !Number.isFinite(longitude))
    ) {
      alert(
        'Select a point of interest or provide challenge latitude and longitude.'
      );
      return;
    }
    if (!quickCreateChallengeForm.question.trim()) {
      alert('Challenge question is required.');
      return;
    }

    setQuickCreateSubmitting('challenge');
    try {
      const created = await apiClient.post<ChallengeNodeOption>(
        '/sonar/challenges',
        {
          zoneId: questForm.zoneId,
          pointOfInterestId: pointOfInterestId || null,
          latitude,
          longitude,
          question: quickCreateChallengeForm.question.trim(),
          description: quickCreateChallengeForm.description.trim(),
          imageUrl: quickCreateChallengeForm.imageUrl.trim(),
          thumbnailUrl:
            quickCreateChallengeForm.thumbnailUrl.trim() ||
            quickCreateChallengeForm.imageUrl.trim(),
          rewardMode: 'explicit',
          randomRewardSize: 'small',
          rewardExperience: parseIntSafe(
            quickCreateChallengeForm.rewardExperience,
            0
          ),
          reward: parseIntSafe(quickCreateChallengeForm.rewardGold, 0),
          submissionType: quickCreateChallengeForm.submissionType,
          difficulty: parseIntSafe(quickCreateChallengeForm.difficulty, 0),
          scaleWithUserLevel: false,
          recurrenceFrequency: '',
          statTags: quickCreateChallengeForm.statTags,
          proficiency: quickCreateChallengeForm.proficiency.trim(),
        }
      );
      setChallenges((prev) => [created, ...prev]);
      setNodeForm((prev) => ({
        ...prev,
        challengeId: created.id,
        submissionType: quickCreateChallengeForm.submissionType,
      }));
      setQuickCreateChallengeForm(emptyQuickCreateChallengeForm());
      setQuickCreateOpen((prev) => ({ ...prev, challenge: false }));
    } catch (error) {
      console.error('Failed to create challenge', error);
      alert(
        error instanceof Error ? error.message : 'Failed to create challenge.'
      );
    } finally {
      setQuickCreateSubmitting(null);
    }
  };

  const handleCreateMonsterEncounter = async () => {
    if (!questForm.zoneId) {
      alert('Select a zone for the quest before creating a monster encounter.');
      return;
    }
    const latitude = Number.parseFloat(
      quickCreateMonsterEncounterForm.latitude
    );
    const longitude = Number.parseFloat(
      quickCreateMonsterEncounterForm.longitude
    );
    if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
      alert('Monster encounter latitude and longitude are required.');
      return;
    }
    if (!quickCreateMonsterEncounterForm.name.trim()) {
      alert('Monster encounter name is required.');
      return;
    }
    if (quickCreateMonsterEncounterForm.monsterIds.length === 0) {
      alert('Select at least one monster for the encounter.');
      return;
    }

    setQuickCreateSubmitting('monster');
    try {
      const created = await apiClient.post<
        MonsterNodeOption & {
          members?: { slot: number; monster: MonsterRecord }[];
        }
      >('/sonar/monster-encounters', {
        name: quickCreateMonsterEncounterForm.name.trim(),
        description: quickCreateMonsterEncounterForm.description.trim(),
        imageUrl: quickCreateMonsterEncounterForm.imageUrl.trim(),
        thumbnailUrl:
          quickCreateMonsterEncounterForm.thumbnailUrl.trim() ||
          quickCreateMonsterEncounterForm.imageUrl.trim(),
        scaleWithUserLevel: quickCreateMonsterEncounterForm.scaleWithUserLevel,
        recurrenceFrequency: '',
        zoneId: questForm.zoneId,
        latitude,
        longitude,
        monsterIds: quickCreateMonsterEncounterForm.monsterIds,
      });
      setMonsterEncounters((prev) => [created, ...prev]);
      setNodeForm((prev) => ({ ...prev, monsterEncounterId: created.id }));
      setQuickCreateMonsterEncounterForm(
        emptyQuickCreateMonsterEncounterForm()
      );
      setQuickCreateOpen((prev) => ({ ...prev, monster: false }));
    } catch (error) {
      console.error('Failed to create monster encounter', error);
      alert(
        error instanceof Error
          ? error.message
          : 'Failed to create monster encounter.'
      );
    } finally {
      setQuickCreateSubmitting(null);
    }
  };

  const handleChallengeDraftChange = (
    nodeId: string,
    updates: Partial<typeof emptyChallengeForm>
  ) => {
    setChallengeDrafts((prev) => ({
      ...prev,
      [nodeId]: { ...emptyChallengeForm, ...prev[nodeId], ...updates },
    }));
  };

  const handleEditChallengeDraftChange = (
    challengeId: string,
    updates: Partial<typeof emptyChallengeForm>
  ) => {
    setChallengeEdits((prev) => ({
      ...prev,
      [challengeId]: {
        ...emptyChallengeForm,
        ...prev[challengeId],
        ...updates,
      },
    }));
  };

  const handleProficiencyInputChange = (value: string) => {
    setProficiencySearch(value);
  };

  const handleStartEditChallenge = (
    node: QuestNode,
    challenge: LegacyQuestNodePrompt
  ) => {
    setChallengeEdits((prev) => ({
      ...prev,
      [challenge.id]: buildChallengeFormFromChallenge(
        challenge,
        node.submissionType
      ),
    }));
  };

  const handleCancelEditChallenge = (challengeId: string) => {
    setChallengeEdits((prev) => {
      const next = { ...prev };
      delete next[challengeId];
      return next;
    });
  };

  const handleCreateChallenge = async (node: QuestNode) => {
    alert(
      'Quest nodes now use their linked objective directly. Nested objective prompts are no longer supported.'
    );
  };

  const handleUpdateChallenge = async (
    node: QuestNode,
    challenge: LegacyQuestNodePrompt
  ) => {
    alert(
      'Quest nodes now use their linked objective directly. Nested objective prompts are no longer supported.'
    );
  };

  const handleShuffleSavedChallenge = async (
    node: QuestNode,
    challenge: LegacyQuestNodePrompt
  ) => {
    alert(
      'Quest nodes now use their linked objective directly. Nested objective prompts are no longer supported.'
    );
  };

  const handleDeleteNode = async (node: QuestNode) => {
    if (!selectedQuest) return;
    const confirmDelete = window.confirm(
      `Delete quest node ${node.orderIndex}? This cannot be undone.`
    );
    if (!confirmDelete) return;
    try {
      await apiClient.delete(`/sonar/questNodes/${node.id}`);
      updateQuestState(selectedQuest.id, (quest) => ({
        ...quest,
        nodes: (quest.nodes ?? []).filter((n) => n.id !== node.id),
      }));
    } catch (error) {
      console.error('Failed to delete quest node', error);
      alert('Failed to delete quest node.');
    }
  };

  const resetImportForm = () => {
    setImportQuery('');
    setSelectedCandidate(null);
    setImportError(null);
    setImportJobs([]);
    setImportPolling(false);
    setImportZoneId(questForm.zoneId || '');
  };

  const handleImportPointOfInterest = async () => {
    setImportError(null);
    if (!selectedCandidate) {
      setImportError('Please select a Google Maps location.');
      return;
    }
    const zoneId = importZoneId || questForm.zoneId;
    if (!zoneId) {
      setImportError('Please select a zone.');
      return;
    }
    try {
      const importItem = await apiClient.post<PointOfInterestImport>(
        '/sonar/pointOfInterest/import',
        {
          placeID: selectedCandidate.place_id,
          zoneID: zoneId,
        }
      );
      setImportJobs((prev) => [importItem, ...prev]);
      setImportPolling(true);
    } catch (error) {
      console.error('Error importing point of interest:', error);
      setImportError('Failed to import point of interest.');
    }
  };

  const handleRetryImport = async (placeId: string, zoneId: string) => {
    try {
      const importItem = await apiClient.post<PointOfInterestImport>(
        '/sonar/pointOfInterest/import',
        {
          placeID: placeId,
          zoneID: zoneId,
        }
      );
      setImportJobs((prev) => [importItem, ...prev]);
      setImportPolling(true);
    } catch (error) {
      console.error('Failed to retry import', error);
      setImportError('Failed to retry import.');
    }
  };

  const fetchImportJobs = async (zoneId?: string) => {
    try {
      const url = zoneId
        ? `/sonar/pointOfInterest/imports?zoneId=${zoneId}`
        : '/sonar/pointOfInterest/imports';
      const response = await apiClient.get<PointOfInterestImport[]>(url);
      setImportJobs(response);
      const hasPending = response.some(
        (item) => item.status === 'queued' || item.status === 'in_progress'
      );
      setImportPolling(hasPending);
    } catch (error) {
      console.error('Failed to fetch import status', error);
    }
  };

  useEffect(() => {
    if (!showImportModal) return;
    fetchImportJobs(importZoneId || questForm.zoneId || undefined);
  }, [showImportModal, importZoneId, questForm.zoneId]);

  useEffect(() => {
    if (!importPolling) return;
    const interval = setInterval(() => {
      fetchImportJobs(importZoneId || questForm.zoneId || undefined);
    }, 3000);
    return () => clearInterval(interval);
  }, [importPolling, importZoneId, questForm.zoneId]);

  useEffect(() => {
    if (importJobs.length === 0) return;
    const completed = importJobs.filter(
      (job) => job.status === 'completed' && job.pointOfInterestId
    );
    if (completed.length === 0) return;

    setNotifiedImportIds((prev) => {
      const next = new Set(prev);
      let hasNew = false;
      completed.forEach((job) => {
        if (!next.has(job.id)) {
          next.add(job.id);
          hasNew = true;
          setImportToasts((existing) =>
            [`Import complete: ${job.placeId}`, ...existing].slice(0, 3)
          );
        }
      });
      return hasNew ? next : prev;
    });

    refreshPointsOfInterest();
  }, [importJobs]);

  const openCharacterLocations = async () => {
    if (!questForm.questGiverCharacterId) return;
    setCharacterLocationsOpen(true);
    setCharacterLocationsLoading(true);
    try {
      const response = await apiClient.get<
        { latitude: number; longitude: number }[]
      >(`/sonar/characters/${questForm.questGiverCharacterId}/locations`);
      setSelectedCharacterLocations(response);
    } catch (error) {
      console.error('Failed to load character locations', error);
    } finally {
      setCharacterLocationsLoading(false);
    }
  };

  const handleAddCharacterLocation = () => {
    setSelectedCharacterLocations((prev) => [
      ...prev,
      { latitude: 0, longitude: 0 },
    ]);
  };

  const handleUpdateCharacterLocation = (
    index: number,
    key: 'latitude' | 'longitude',
    value: number
  ) => {
    setSelectedCharacterLocations((prev) =>
      prev.map((loc, i) => (i === index ? { ...loc, [key]: value } : loc))
    );
  };

  const handleRemoveCharacterLocation = (index: number) => {
    setSelectedCharacterLocations((prev) => prev.filter((_, i) => i !== index));
  };

  const handleSaveCharacterLocations = async () => {
    if (!questForm.questGiverCharacterId) return;
    try {
      await apiClient.put(
        `/sonar/characters/${questForm.questGiverCharacterId}/locations`,
        {
          locations: selectedCharacterLocations,
        }
      );
      setCharacterLocationsOpen(false);
    } catch (error) {
      console.error('Failed to save character locations', error);
      alert('Failed to save character locations.');
    }
  };

  if (loading) {
    return (
      <div className="qa-theme qa-quests">
        <div className="qa-shell">
          <div className="qa-panel">Loading quests...</div>
        </div>
      </div>
    );
  }

  return (
    <div className="qa-theme qa-quests">
      <div className="qa-shell">
        <datalist id="proficiency-options">
          {proficiencyOptions.map((option) => (
            <option key={option} value={option} />
          ))}
        </datalist>
        {importToasts.length > 0 && (
          <div className="fixed right-4 top-4 z-50 space-y-2">
            {importToasts.map((toast, index) => (
              <div
                key={`${toast}-${index}`}
                className="rounded-md bg-emerald-600 px-4 py-2 text-sm text-white shadow"
              >
                {toast}
              </div>
            ))}
          </div>
        )}
        {loadError && (
          <div className="mb-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">
            {loadError}
          </div>
        )}
        <header className="qa-hero">
          <div>
            <div className="qa-kicker">Quest Operations</div>
            <h1 className="qa-title">Quests</h1>
            <p className="qa-subtitle">
              Build quests, manage nodes, and tune challenge inputs with the
              same archetype-focused UI language.
            </p>
          </div>
          <div className="qa-hero-actions">
            <button
              className="qa-btn qa-btn-primary"
              onClick={() => setShowCreateQuest((prev) => !prev)}
            >
              {showCreateQuest ? 'Close' : 'Create Quest'}
            </button>
          </div>
        </header>

        {showCreateQuest && (
          <div className="qa-card">
            <div className="qa-card-header">
              <div>
                <div className="qa-kicker">New Quest</div>
                <h2 className="qa-card-title">Create Quest</h2>
                <p className="qa-meta">
                  Start with the player-facing narrative, then assign the quest
                  to a zone and define the reward behavior.
                </p>
              </div>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="md:col-span-2 qa-form-section">
                <div className="qa-card-kicker">Narrative</div>
                <div className="qa-section-title">Overview</div>
                <p className="qa-section-copy">
                  Set the quest identity that players will see before you wire
                  it into the world.
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  Name
                </label>
                <input
                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                  value={questForm.name}
                  onChange={(e) =>
                    setQuestForm((prev) => ({ ...prev, name: e.target.value }))
                  }
                />
              </div>
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-gray-700">
                  Description
                </label>
                <textarea
                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                  rows={3}
                  value={questForm.description}
                  onChange={(e) =>
                    setQuestForm((prev) => ({
                      ...prev,
                      description: e.target.value,
                    }))
                  }
                />
              </div>
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-gray-700">
                  Quest Acceptance Dialogue
                </label>
                <textarea
                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                  rows={4}
                  placeholder="One line per dialogue message shown before accepting the quest."
                  value={questForm.acceptanceDialogue.join('\n')}
                  onChange={(e) =>
                    setQuestForm((prev) => ({
                      ...prev,
                      acceptanceDialogue: e.target.value.split('\n'),
                    }))
                  }
                />
                <p className="mt-1 text-xs text-gray-500">
                  Each line becomes a separate dialogue line in the quest
                  acceptance prompt.
                </p>
              </div>
              <div className="md:col-span-2 qa-form-section">
                <div className="qa-card-kicker">Assignment</div>
                <div className="qa-section-title">World Placement</div>
                <p className="qa-section-copy">
                  Choose the zone, quest giver, and pacing rules that frame how
                  the route should behave.
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  Zone
                </label>
                <input
                  className="mt-1 mb-2 block w-full border border-gray-300 rounded-md p-2"
                  placeholder="Filter zones..."
                  value={zoneSearch}
                  onChange={(e) => setZoneSearch(e.target.value)}
                />
                <select
                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                  value={questForm.zoneId}
                  onChange={(e) =>
                    setQuestForm((prev) => ({
                      ...prev,
                      zoneId: e.target.value,
                    }))
                  }
                >
                  <option value="">No Zone</option>
                  {filteredZones.map((zone) => (
                    <option key={zone.id} value={zone.id}>
                      {zone.name}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  Quest Giver Character
                </label>
                <input
                  className="mt-1 mb-2 block w-full border border-gray-300 rounded-md p-2"
                  placeholder="Filter characters..."
                  value={characterSearch}
                  onChange={(e) => setCharacterSearch(e.target.value)}
                />
                <select
                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                  value={questForm.questGiverCharacterId}
                  onChange={(e) =>
                    setQuestForm((prev) => ({
                      ...prev,
                      questGiverCharacterId: e.target.value,
                    }))
                  }
                >
                  <option value="">None</option>
                  {filteredCharacters.map((character) => (
                    <option key={character.id} value={character.id}>
                      {character.name}
                    </option>
                  ))}
                </select>
                {questForm.questGiverCharacterId && (
                  <button
                    type="button"
                    className="mt-2 rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                    onClick={openCharacterLocations}
                  >
                    Edit Character Locations
                  </button>
                )}
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  Quest Archetype ID (optional)
                </label>
                <input
                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                  value={questForm.questArchetypeId}
                  onChange={(e) =>
                    setQuestForm((prev) => ({
                      ...prev,
                      questArchetypeId: e.target.value,
                    }))
                  }
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  Recurrence
                </label>
                <select
                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                  value={questForm.recurrenceFrequency}
                  onChange={(e) =>
                    setQuestForm((prev) => ({
                      ...prev,
                      recurrenceFrequency: e.target.value,
                    }))
                  }
                >
                  {questRecurrenceOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  Difficulty Mode
                </label>
                <select
                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                  value={questForm.difficultyMode}
                  onChange={(e) =>
                    setQuestForm((prev) => ({
                      ...prev,
                      difficultyMode: e.target.value as QuestDifficultyMode,
                    }))
                  }
                >
                  <option value="scale">Scale With User Level</option>
                  <option value="fixed">Fixed Difficulty</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  Difficulty
                </label>
                <input
                  type="number"
                  min={1}
                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                  value={questForm.difficulty}
                  disabled={questForm.difficultyMode !== 'fixed'}
                  onChange={(e) =>
                    setQuestForm((prev) => ({
                      ...prev,
                      difficulty: Math.max(1, Number(e.target.value) || 1),
                    }))
                  }
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  Monster Encounter Target Level
                </label>
                <input
                  type="number"
                  min={1}
                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                  value={questForm.monsterEncounterTargetLevel}
                  onChange={(e) =>
                    setQuestForm((prev) => ({
                      ...prev,
                      monsterEncounterTargetLevel: Math.max(
                        1,
                        Number(e.target.value) || 1
                      ),
                    }))
                  }
                />
              </div>
              <div className="md:col-span-2 qa-form-section">
                <div className="qa-card-kicker">Rewards</div>
                <div className="qa-section-title">Quest Payoff</div>
                <p className="qa-section-copy">
                  Configure the quest-level reward package before adding any
                  route nodes.
                </p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  Reward Mode
                </label>
                <select
                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                  value={questForm.rewardMode}
                  onChange={(e) =>
                    setQuestForm((prev) => ({
                      ...prev,
                      rewardMode: e.target.value as 'explicit' | 'random',
                    }))
                  }
                >
                  <option value="random">Random Reward</option>
                  <option value="explicit">Explicit Reward</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  Random Reward Size
                </label>
                <select
                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                  value={questForm.randomRewardSize}
                  disabled={questForm.rewardMode !== 'random'}
                  onChange={(e) =>
                    setQuestForm((prev) => ({
                      ...prev,
                      randomRewardSize: e.target.value as
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
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  Experience Reward
                </label>
                <input
                  type="number"
                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                  value={questForm.rewardExperience}
                  disabled={questForm.rewardMode !== 'explicit'}
                  onChange={(e) =>
                    setQuestForm((prev) => ({
                      ...prev,
                      rewardExperience: Number(e.target.value),
                    }))
                  }
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  Gold Reward
                </label>
                <input
                  type="number"
                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                  value={questForm.gold}
                  disabled={questForm.rewardMode !== 'explicit'}
                  onChange={(e) =>
                    setQuestForm((prev) => ({
                      ...prev,
                      gold: Number(e.target.value),
                    }))
                  }
                />
              </div>
              {questForm.rewardMode === 'random' && (
                <div className="md:col-span-2 text-xs text-gray-500">
                  Random rewards ignore explicit gold/material/item/spell
                  fields.
                </div>
              )}
              <div className="md:col-span-2">
                <MaterialRewardsEditor
                  value={questForm.materialRewards}
                  onChange={(materialRewards) =>
                    setQuestForm((prev) => ({ ...prev, materialRewards }))
                  }
                  disabled={questForm.rewardMode !== 'explicit'}
                />
              </div>
              <div className="md:col-span-2">
                <div className="flex items-center justify-between">
                  <label className="block text-sm font-medium text-gray-700">
                    Item Rewards
                  </label>
                  <button
                    type="button"
                    className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                    onClick={handleAddQuestReward}
                    disabled={questForm.rewardMode !== 'explicit'}
                  >
                    Add Item Reward
                  </button>
                </div>
                {questForm.itemRewards.length === 0 ? (
                  <div className="mt-2 text-xs text-gray-500">
                    No item rewards yet.
                  </div>
                ) : (
                  <div className="mt-2 space-y-2">
                    {questForm.itemRewards.map((reward, index) => (
                      <div
                        key={`create-reward-${index}`}
                        className="grid grid-cols-[1fr_120px_auto] gap-2 items-center"
                      >
                        <select
                          className="block w-full border border-gray-300 rounded-md p-2"
                          value={reward.inventoryItemId}
                          disabled={questForm.rewardMode !== 'explicit'}
                          onChange={(e) =>
                            handleUpdateQuestReward(index, {
                              inventoryItemId: e.target.value,
                            })
                          }
                        >
                          <option value="">Select item</option>
                          {inventoryItems.map((item) => (
                            <option key={item.id} value={item.id}>
                              {item.name}
                            </option>
                          ))}
                        </select>
                        <input
                          type="number"
                          className="block w-full border border-gray-300 rounded-md p-2"
                          min={1}
                          value={reward.quantity}
                          disabled={questForm.rewardMode !== 'explicit'}
                          onChange={(e) =>
                            handleUpdateQuestReward(index, {
                              quantity: Number(e.target.value),
                            })
                          }
                        />
                        <button
                          type="button"
                          className="rounded-md border border-red-200 px-2 py-1 text-xs text-red-600 hover:bg-red-50"
                          disabled={questForm.rewardMode !== 'explicit'}
                          onClick={() => handleRemoveQuestReward(index)}
                        >
                          Remove
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>
              <div className="md:col-span-2">
                <div className="flex items-center justify-between">
                  <label className="block text-sm font-medium text-gray-700">
                    Spell Rewards
                  </label>
                  <button
                    type="button"
                    className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                    onClick={handleAddQuestSpellReward}
                    disabled={questForm.rewardMode !== 'explicit'}
                  >
                    Add Spell Reward
                  </button>
                </div>
                {questForm.spellRewards.length === 0 ? (
                  <div className="mt-2 text-xs text-gray-500">
                    No spell rewards yet.
                  </div>
                ) : (
                  <div className="mt-2 space-y-2">
                    {questForm.spellRewards.map((reward, index) => (
                      <div
                        key={`create-spell-reward-${index}`}
                        className="grid grid-cols-[1fr_auto] gap-2 items-center"
                      >
                        <select
                          className="block w-full border border-gray-300 rounded-md p-2"
                          value={reward.spellId}
                          disabled={questForm.rewardMode !== 'explicit'}
                          onChange={(e) =>
                            handleUpdateQuestSpellReward(index, {
                              spellId: e.target.value,
                            })
                          }
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
                          className="rounded-md border border-red-200 px-2 py-1 text-xs text-red-600 hover:bg-red-50"
                          disabled={questForm.rewardMode !== 'explicit'}
                          onClick={() => handleRemoveQuestSpellReward(index)}
                        >
                          Remove
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
            <div className="mt-4">
              <button
                className="qa-btn qa-btn-primary"
                onClick={handleCreateQuest}
                disabled={!questForm.name.trim()}
              >
                Create Quest
              </button>
            </div>
          </div>
        )}

        <div className="qa-workspace">
          <aside className="qa-sidebar">
            <div className="qa-card qa-library-card">
              <div className="qa-library-header">
                <div>
                  <div className="qa-kicker">Library</div>
                  <h2 className="qa-card-title">Quest Stack</h2>
                  <p className="qa-meta">
                    {filteredQuests.length} visible of {quests.length} quests
                  </p>
                </div>
                <div className="qa-chip accent">
                  Selected {selectedQuestIds.size}
                </div>
              </div>
              <input
                className="mb-3 block w-full border border-gray-300 rounded-md p-2"
                placeholder="Search quests..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
              <div className="mb-3 flex flex-wrap items-center gap-2">
                <button
                  type="button"
                  className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                  onClick={toggleSelectVisibleQuests}
                  disabled={filteredQuests.length === 0 || bulkDeletingQuests}
                >
                  {allFilteredQuestsSelected
                    ? 'Unselect Visible'
                    : 'Select Visible'}
                </button>
                <button
                  type="button"
                  className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                  onClick={clearQuestSelection}
                  disabled={selectedQuestIds.size === 0 || bulkDeletingQuests}
                >
                  Clear Selection
                </button>
                <button
                  type="button"
                  className="qa-btn qa-btn-danger"
                  onClick={handleBulkDeleteQuests}
                  disabled={
                    selectedQuestIds.size === 0 ||
                    bulkDeletingQuests ||
                    deletingQuestId !== null
                  }
                >
                  {bulkDeletingQuests
                    ? `Deleting ${selectedQuestIds.size}...`
                    : `Delete Selected (${selectedQuestIds.size})`}
                </button>
              </div>
              <div className="qa-library-list">
                {filteredQuests.map((quest) => {
                  const isActive = selectedQuestId === quest.id;
                  return (
                    <div
                      key={quest.id}
                      className={`qa-library-item ${isActive ? 'active' : ''}`}
                    >
                      <input
                        type="checkbox"
                        className="h-4 w-4"
                        checked={selectedQuestIdSet.has(quest.id)}
                        disabled={bulkDeletingQuests}
                        onChange={() => toggleQuestSelection(quest.id)}
                      />
                      <button
                        className="qa-library-item-body"
                        onClick={() => handleSelectQuest(quest)}
                      >
                        <div className="qa-library-item-title">{quest.name}</div>
                        <div className="qa-library-item-meta">
                          <span>
                            {quest.nodeCount ?? quest.nodes?.length ?? 0} nodes
                          </span>
                          <span>
                            {getQuestRecurrenceLabel(
                              quest.recurrenceFrequency ?? ''
                            )}
                          </span>
                          <span>{getQuestRewardMode(quest)} rewards</span>
                        </div>
                      </button>
                      <button
                        className="qa-btn qa-btn-danger"
                        onClick={() => handleDeleteQuestById(quest)}
                        disabled={
                          deletingQuestId === quest.id || bulkDeletingQuests
                        }
                      >
                        {deletingQuestId === quest.id ? 'Deleting...' : 'Delete'}
                      </button>
                    </div>
                  );
                })}
              </div>
            </div>
          </aside>

          <div className="qa-workbench">
            {!selectedQuest ? (
              <div className="qa-card qa-empty-state">
                <div className="qa-kicker">Workbench</div>
                <h2 className="qa-card-title">Pick a quest to shape</h2>
                <p className="qa-meta">
                  The quest editor now centers the active quest as a route. Pick
                  one from the library to tune the narrative, rewards, and node
                  flow in separate spaces.
                </p>
              </div>
            ) : selectedQuestIsHydrating ? (
              <div className="qa-card qa-empty-state">
                <div className="qa-kicker">Quest Workbench</div>
                <h2 className="qa-card-title">Loading quest details</h2>
                <p className="qa-meta">
                  Pulling the full node graph and reward package for{' '}
                  {selectedQuest.name}.
                </p>
              </div>
            ) : selectedQuestNeedsHydration ? (
              <div className="qa-card qa-empty-state">
                <div className="qa-kicker">Quest Workbench</div>
                <h2 className="qa-card-title">Quest details unavailable</h2>
                <p className="qa-meta">
                  {questDetailErrorId === selectedQuest.id
                    ? `We couldn't load the full quest payload for ${selectedQuest.name}.`
                    : `Select ${selectedQuest.name} to load its full node graph.`}
                </p>
                <div className="mt-4">
                  <button
                    className="qa-btn qa-btn-primary"
                    onClick={() => handleSelectQuest(selectedQuest)}
                  >
                    {questDetailErrorId === selectedQuest.id
                      ? 'Retry Load'
                      : 'Load Quest'}
                  </button>
                </div>
              </div>
            ) : (
              <>
                <div className="qa-card qa-focus-card">
                  <div className="qa-card-header">
                    <div>
                      <div className="qa-kicker">Quest Workbench</div>
                      <h2 className="qa-focus-title">
                        {questForm.name.trim() || 'Untitled Quest'}
                      </h2>
                      <p className="qa-focus-copy">
                        {questForm.description.trim() ||
                          'Add a description to frame the route and player intent.'}
                      </p>
                    </div>
                    <div className="qa-actions">
                      <button
                        className="qa-btn qa-btn-outline"
                        onClick={() => {
                          resetImportForm();
                          setShowImportModal(true);
                        }}
                      >
                        Import POI
                      </button>
                      <button
                        className="qa-btn qa-btn-outline"
                        onClick={handleCreateQuestArchetypeFromQuest}
                        disabled={creatingArchetype}
                      >
                        {creatingArchetype
                          ? 'Creating Archetype...'
                          : 'Create Archetype'}
                      </button>
                      <button
                        className="qa-btn qa-btn-primary"
                        onClick={handleUpdateQuest}
                      >
                        Save Changes
                      </button>
                      <button
                        className="qa-btn qa-btn-danger"
                        onClick={handleDeleteQuest}
                        disabled={
                          deletingQuestId === selectedQuest.id ||
                          bulkDeletingQuests
                        }
                      >
                        {deletingQuestId === selectedQuest.id
                          ? 'Deleting...'
                          : 'Delete Quest'}
                      </button>
                    </div>
                  </div>

                  <div className="qa-stat-grid">
                    <div className="qa-stat">
                      <div className="qa-stat-label">Zone</div>
                      <div className="qa-stat-value">
                        {selectedQuestZone?.name ?? 'No zone'}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Quest Giver</div>
                      <div className="qa-stat-value">
                        {selectedQuestGiver?.name ?? 'Unassigned'}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Difficulty</div>
                      <div className="qa-stat-value">
                        {questForm.difficultyMode === 'fixed'
                          ? `Fixed ${questForm.difficulty}`
                          : 'Scales with player'}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Rewards</div>
                      <div className="qa-stat-value">
                        {questForm.rewardMode === 'explicit'
                          ? 'Explicit'
                          : 'Random'}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Recurrence</div>
                      <div className="qa-stat-value">
                        {getQuestRecurrenceLabel(
                          questForm.recurrenceFrequency
                        )}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Encounter Level</div>
                      <div className="qa-stat-value">
                        {questForm.monsterEncounterTargetLevel}
                      </div>
                    </div>
                  </div>

                  <div className="qa-node-strip">
                    {selectedQuestNodePreview.length === 0 ? (
                      <div className="qa-chip muted">
                        No nodes yet. Start building the route below.
                      </div>
                    ) : (
                      selectedQuestNodePreview.map((node) => (
                        <div
                          key={node.id}
                          className={`qa-node-pill ${node.nodeType}`}
                        >
                          <span className="qa-node-pill-index">
                            {node.orderIndex}
                          </span>
                          <span className="qa-node-pill-label">
                            {getQuestNodeKindLabel(node.nodeType)}
                          </span>
                          <span className="qa-node-pill-copy">
                            {node.label}
                          </span>
                        </div>
                      ))
                    )}
                  </div>

                  {selectedQuestLocationDetails.length > 0 && (
                    <div className="qa-panel" style={{ marginTop: 16 }}>
                      <div className="qa-card-header" style={{ marginBottom: 12 }}>
                        <div>
                          <div className="qa-meta">Quest Locations</div>
                          <div className="qa-card-title" style={{ fontSize: 16 }}>
                            Linked Points of Interest
                          </div>
                        </div>
                        <div className="qa-chip muted">
                          {selectedQuestLocationDetails.length} linked
                        </div>
                      </div>
                      <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                        {selectedQuestLocationDetails.map((location) => (
                          <div
                            key={location.nodeId}
                            className="rounded-md border border-gray-200 bg-white p-3"
                          >
                            <div className="flex gap-3">
                              {location.poi.thumbnailUrl || location.poi.imageURL ? (
                                <img
                                  src={
                                    location.poi.thumbnailUrl ||
                                    location.poi.imageURL
                                  }
                                  alt={location.poi.name}
                                  className="h-20 w-20 rounded-md object-cover"
                                />
                              ) : null}
                              <div className="min-w-0 flex-1">
                                <div className="flex items-start justify-between gap-3">
                                  <div>
                                    <div className="text-xs font-medium text-gray-500">
                                      Node {location.orderIndex}
                                    </div>
                                    <div className="font-semibold text-gray-900">
                                      {location.poi.name}
                                    </div>
                                  </div>
                                  <Link
                                    to={`/points-of-interest/${location.poi.id}`}
                                    className={adminEntityLinkClass}
                                  >
                                    Open POI
                                  </Link>
                                </div>
                                {location.aliases.length > 0 && (
                                  <div className="mt-1 text-xs text-gray-500">
                                    {location.aliases.join(' · ')}
                                  </div>
                                )}
                                {location.summary && (
                                  <p className="mt-2 text-sm text-gray-600">
                                    {location.summary}
                                  </p>
                                )}
                                <div className="mt-2 flex flex-wrap gap-2">
                                  {location.archetype && (
                                    <span className="qa-chip muted">
                                      Archetype: {location.archetype.name}
                                    </span>
                                  )}
                                  {(location.latitude || location.longitude) && (
                                    <span className="qa-chip muted">
                                      {location.latitude ?? '?'}, {location.longitude ?? '?'}
                                    </span>
                                  )}
                                </div>
                                {location.tagNames.length > 0 && (
                                  <div className="mt-2 flex flex-wrap gap-2">
                                    {location.tagNames.slice(0, 5).map((tagName) => (
                                      <span
                                        key={`${location.poi.id}-${tagName}`}
                                        className="rounded-full border border-gray-200 bg-gray-50 px-2 py-1 text-xs text-gray-700"
                                      >
                                        {tagName}
                                      </span>
                                    ))}
                                    {location.tagNames.length > 5 && (
                                      <span className="rounded-full border border-gray-200 bg-gray-50 px-2 py-1 text-xs text-gray-500">
                                        +{location.tagNames.length - 5} more
                                      </span>
                                    )}
                                  </div>
                                )}
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>

                <div className="qa-card qa-detail-card">
                  <div className="qa-card-header">
                    <div>
                      <div className="qa-kicker">Configuration</div>
                      <h3 className="qa-card-title">Quest Settings</h3>
                      <p className="qa-meta">
                        Narrative copy, assignment, pacing, and reward behavior
                        all live here.
                      </p>
                    </div>
                    <div className="qa-chip muted">
                      {orderedQuestNodes.length} route steps
                    </div>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
                  <div className="md:col-span-2 qa-form-section">
                    <div className="qa-card-kicker">Narrative</div>
                    <div className="qa-section-title">Overview</div>
                    <p className="qa-section-copy">
                      Define the player-facing quest identity before tuning the
                      route itself.
                    </p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Name
                    </label>
                    <input
                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                      value={questForm.name}
                      onChange={(e) =>
                        setQuestForm((prev) => ({
                          ...prev,
                          name: e.target.value,
                        }))
                      }
                    />
                  </div>
                  <div className="md:col-span-2">
                    <label className="block text-sm font-medium text-gray-700">
                      Description
                    </label>
                    <textarea
                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                      rows={3}
                      value={questForm.description}
                      onChange={(e) =>
                        setQuestForm((prev) => ({
                          ...prev,
                          description: e.target.value,
                        }))
                      }
                    />
                  </div>
                  <div className="md:col-span-2">
                    <label className="block text-sm font-medium text-gray-700">
                      Quest Acceptance Dialogue
                    </label>
                    <textarea
                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                      rows={4}
                      placeholder="One line per dialogue message shown before accepting the quest."
                      value={questForm.acceptanceDialogue.join('\n')}
                      onChange={(e) =>
                        setQuestForm((prev) => ({
                          ...prev,
                          acceptanceDialogue: e.target.value.split('\n'),
                        }))
                      }
                    />
                    <p className="mt-1 text-xs text-gray-500">
                      Each line becomes a separate dialogue line in the quest
                      acceptance prompt.
                    </p>
                  </div>
                  <div className="md:col-span-2 qa-form-section">
                    <div className="qa-card-kicker">Assignment</div>
                    <div className="qa-section-title">Quest Framing</div>
                    <p className="qa-section-copy">
                      Set where the quest lives, who gives it, and how it
                      should scale over time.
                    </p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Zone
                    </label>
                    <select
                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                      value={questForm.zoneId}
                      onChange={(e) =>
                        setQuestForm((prev) => ({
                          ...prev,
                          zoneId: e.target.value,
                        }))
                      }
                    >
                      <option value="">No Zone</option>
                      {zones.map((zone) => (
                        <option key={zone.id} value={zone.id}>
                          {zone.name}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Quest Giver Character
                    </label>
                    <select
                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                      value={questForm.questGiverCharacterId}
                      onChange={(e) =>
                        setQuestForm((prev) => ({
                          ...prev,
                          questGiverCharacterId: e.target.value,
                        }))
                      }
                    >
                      <option value="">None</option>
                      {characters.map((character) => (
                        <option key={character.id} value={character.id}>
                          {character.name}
                        </option>
                      ))}
                    </select>
                    {questForm.questGiverCharacterId && (
                      <button
                        type="button"
                        className="mt-2 rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                        onClick={openCharacterLocations}
                      >
                        Edit Character Locations
                      </button>
                    )}
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Quest Archetype ID
                    </label>
                    <input
                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                      value={questForm.questArchetypeId}
                      onChange={(e) =>
                        setQuestForm((prev) => ({
                          ...prev,
                          questArchetypeId: e.target.value,
                        }))
                      }
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Recurrence
                    </label>
                    <select
                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                      value={questForm.recurrenceFrequency}
                      onChange={(e) =>
                        setQuestForm((prev) => ({
                          ...prev,
                          recurrenceFrequency: e.target.value,
                        }))
                      }
                    >
                      {questRecurrenceOptions.map((option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                        ))}
                      </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Difficulty Mode
                    </label>
                    <select
                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                      value={questForm.difficultyMode}
                      onChange={(e) =>
                        setQuestForm((prev) => ({
                          ...prev,
                          difficultyMode: e.target.value as QuestDifficultyMode,
                        }))
                      }
                    >
                      <option value="scale">Scale With User Level</option>
                      <option value="fixed">Fixed Difficulty</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Difficulty
                    </label>
                    <input
                      type="number"
                      min={1}
                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                      value={questForm.difficulty}
                      disabled={questForm.difficultyMode !== 'fixed'}
                      onChange={(e) =>
                        setQuestForm((prev) => ({
                          ...prev,
                          difficulty: Math.max(1, Number(e.target.value) || 1),
                        }))
                      }
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Monster Encounter Target Level
                    </label>
                    <input
                      type="number"
                      min={1}
                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                      value={questForm.monsterEncounterTargetLevel}
                      onChange={(e) =>
                        setQuestForm((prev) => ({
                          ...prev,
                          monsterEncounterTargetLevel: Math.max(
                            1,
                            Number(e.target.value) || 1
                          ),
                        }))
                      }
                    />
                  </div>
                  <div className="md:col-span-2 qa-form-section">
                    <div className="qa-card-kicker">Rewards</div>
                    <div className="qa-section-title">Quest Payoff</div>
                    <p className="qa-section-copy">
                      Rewards are configured at the quest level so every node
                      stays focused on structure rather than loot.
                    </p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Reward Mode
                    </label>
                    <select
                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                      value={questForm.rewardMode}
                      onChange={(e) =>
                        setQuestForm((prev) => ({
                          ...prev,
                          rewardMode: e.target.value as 'explicit' | 'random',
                        }))
                      }
                    >
                      <option value="random">Random Reward</option>
                      <option value="explicit">Explicit Reward</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Random Reward Size
                    </label>
                    <select
                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                      value={questForm.randomRewardSize}
                      disabled={questForm.rewardMode !== 'random'}
                      onChange={(e) =>
                        setQuestForm((prev) => ({
                          ...prev,
                          randomRewardSize: e.target.value as
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
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Experience Reward
                    </label>
                    <input
                      type="number"
                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                      value={questForm.rewardExperience}
                      disabled={questForm.rewardMode !== 'explicit'}
                      onChange={(e) =>
                        setQuestForm((prev) => ({
                          ...prev,
                          rewardExperience: Number(e.target.value),
                        }))
                      }
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Gold Reward
                    </label>
                    <input
                      type="number"
                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                      value={questForm.gold}
                      disabled={questForm.rewardMode !== 'explicit'}
                      onChange={(e) =>
                        setQuestForm((prev) => ({
                          ...prev,
                          gold: Number(e.target.value),
                        }))
                      }
                    />
                  </div>
                  {questForm.rewardMode === 'random' && (
                    <div className="md:col-span-2 text-xs text-gray-500">
                      Random rewards ignore explicit gold/material/item/spell
                      fields.
                    </div>
                  )}
                  <div className="md:col-span-2">
                    <MaterialRewardsEditor
                      value={questForm.materialRewards}
                      onChange={(materialRewards) =>
                        setQuestForm((prev) => ({ ...prev, materialRewards }))
                      }
                      disabled={questForm.rewardMode !== 'explicit'}
                    />
                  </div>
                  <div className="md:col-span-2">
                    <div className="flex items-center justify-between">
                      <label className="block text-sm font-medium text-gray-700">
                        Item Rewards
                      </label>
                      <button
                        type="button"
                        className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                        onClick={handleAddQuestReward}
                      >
                        Add Item Reward
                      </button>
                    </div>
                    {questForm.itemRewards.length === 0 ? (
                      <div className="mt-2 text-xs text-gray-500">
                        No item rewards yet.
                      </div>
                    ) : (
                      <div className="mt-2 space-y-2">
                        {questForm.itemRewards.map((reward, index) => (
                          <div
                            key={`edit-reward-${index}`}
                            className="grid grid-cols-[1fr_120px_auto] gap-2 items-center"
                          >
                            <select
                              className="block w-full border border-gray-300 rounded-md p-2"
                              value={reward.inventoryItemId}
                              onChange={(e) =>
                                handleUpdateQuestReward(index, {
                                  inventoryItemId: e.target.value,
                                })
                              }
                            >
                              <option value="">Select item</option>
                              {inventoryItems.map((item) => (
                                <option key={item.id} value={item.id}>
                                  {item.name}
                                </option>
                              ))}
                            </select>
                            <input
                              type="number"
                              className="block w-full border border-gray-300 rounded-md p-2"
                              min={1}
                              value={reward.quantity}
                              onChange={(e) =>
                                handleUpdateQuestReward(index, {
                                  quantity: Number(e.target.value),
                                })
                              }
                            />
                            <button
                              type="button"
                              className="rounded-md border border-red-200 px-2 py-1 text-xs text-red-600 hover:bg-red-50"
                              onClick={() => handleRemoveQuestReward(index)}
                            >
                              Remove
                            </button>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                  <div className="md:col-span-2">
                    <div className="flex items-center justify-between">
                      <label className="block text-sm font-medium text-gray-700">
                        Spell Rewards
                      </label>
                      <button
                        type="button"
                        className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                        onClick={handleAddQuestSpellReward}
                      >
                        Add Spell Reward
                      </button>
                    </div>
                    {questForm.spellRewards.length === 0 ? (
                      <div className="mt-2 text-xs text-gray-500">
                        No spell rewards yet.
                      </div>
                    ) : (
                      <div className="mt-2 space-y-2">
                        {questForm.spellRewards.map((reward, index) => (
                          <div
                            key={`edit-spell-reward-${index}`}
                            className="grid grid-cols-[1fr_auto] gap-2 items-center"
                          >
                            <select
                              className="block w-full border border-gray-300 rounded-md p-2"
                              value={reward.spellId}
                              onChange={(e) =>
                                handleUpdateQuestSpellReward(index, {
                                  spellId: e.target.value,
                                })
                              }
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
                              className="rounded-md border border-red-200 px-2 py-1 text-xs text-red-600 hover:bg-red-50"
                              onClick={() =>
                                handleRemoveQuestSpellReward(index)
                              }
                            >
                              Remove
                            </button>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>

                </div>

                <div className="qa-card qa-route-card">
                  <div className="qa-card-header">
                    <div>
                      <div className="qa-kicker">Route Studio</div>
                      <h3 className="qa-card-title">Quest Nodes</h3>
                      <p className="qa-meta">
                        Build the route, preview its composition, and attach the
                        concrete world entities that bring each step to life.
                      </p>
                    </div>
                    <div className="qa-route-summary">
                      <div className="qa-chip muted">
                        {selectedQuestNodeCounts.poi} locations
                      </div>
                      <div className="qa-chip muted">
                        {selectedQuestNodeCounts.scenario} scenarios
                      </div>
                      <div className="qa-chip muted">
                        {selectedQuestNodeCounts.monster} monsters
                      </div>
                      <div className="qa-chip muted">
                        {selectedQuestNodeCounts.challenge} challenges
                      </div>
                      <div className="qa-chip muted">
                        {selectedQuestNodeCounts.polygon} polygons
                      </div>
                    </div>
                  </div>
                  <div className="qa-divider" />
                  <div className="qa-route-builder">
                  <div className="bg-gray-50 border border-gray-200 rounded-md p-4 mb-4">
                    <h4 className="font-semibold mb-3">Add Node</h4>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700">
                          Order Index
                        </label>
                        <input
                          type="number"
                          className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                          value={nodeForm.orderIndex}
                          onChange={(e) =>
                            setNodeForm((prev) => ({
                              ...prev,
                              orderIndex: Number(e.target.value),
                            }))
                          }
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700">
                          Node Type
                        </label>
                        <select
                          className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                          value={nodeForm.nodeType}
                          onChange={(e) => {
                            const nextNodeType = e.target
                              .value as QuestNodeType;
                            setNodeForm((prev) => ({
                              ...prev,
                              nodeType: nextNodeType,
                              pointOfInterestId:
                                nextNodeType === 'poi'
                                  ? prev.pointOfInterestId
                                  : '',
                              scenarioId:
                                nextNodeType === 'scenario'
                                  ? prev.scenarioId
                                  : '',
                              monsterEncounterId:
                                nextNodeType === 'monster'
                                  ? prev.monsterEncounterId
                                  : '',
                              challengeId:
                                nextNodeType === 'challenge'
                                  ? prev.challengeId
                                  : '',
                              polygonPoints:
                                nextNodeType === 'polygon'
                                  ? prev.polygonPoints
                                  : '',
                            }));
                          }}
                        >
                          <option value="scenario">Scenario</option>
                          <option value="monster">Monster</option>
                          <option value="challenge">Challenge Objective</option>
                        </select>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700">
                          Submission Type
                        </label>
                        <select
                          className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                          value={nodeForm.submissionType}
                          onChange={(e) =>
                            setNodeForm((prev) => ({
                              ...prev,
                              submissionType: e.target
                                .value as QuestNodeSubmissionType,
                            }))
                          }
                        >
                          {questNodeSubmissionOptions.map((option) => (
                            <option key={option.value} value={option.value}>
                              {option.label}
                            </option>
                          ))}
                        </select>
                      </div>
                      {nodeForm.nodeType === 'poi' ? (
                        <div className="md:col-span-2">
                          <label className="block text-sm font-medium text-gray-700">
                            Point of Interest
                          </label>
                          <input
                            className="mt-1 mb-2 block w-full border border-gray-300 rounded-md p-2"
                            placeholder="Search points of interest..."
                            value={poiSearch}
                            onChange={(e) => setPoiSearch(e.target.value)}
                          />
                          <button
                            type="button"
                            className="mb-2 flex w-full items-center justify-between rounded-md border border-gray-300 bg-white px-3 py-2 text-sm"
                            onClick={() => setPoiFiltersOpen((prev) => !prev)}
                          >
                            <span>Filters</span>
                            <span>{poiFiltersOpen ? 'Hide' : 'Show'}</span>
                          </button>
                          {poiFiltersOpen && (
                            <div className="mb-3 rounded-md border border-gray-200 bg-gray-50 p-3">
                              <div className="mb-3">
                                <label className="block text-xs font-medium text-gray-700">
                                  Zone
                                </label>
                                <select
                                  className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                  value={poiZoneFilterId}
                                  onChange={(e) =>
                                    setPoiZoneFilterId(e.target.value)
                                  }
                                >
                                  <option value="">All zones</option>
                                  {zones.map((zone) => (
                                    <option key={zone.id} value={zone.id}>
                                      {zone.name}
                                    </option>
                                  ))}
                                </select>
                                {poiZoneFilterId && zonePoiMapLoading && (
                                  <p className="mt-1 text-xs text-gray-500">
                                    Loading zone points of interest…
                                  </p>
                                )}
                              </div>
                              <div>
                                <label className="block text-xs font-medium text-gray-700">
                                  Tags
                                </label>
                                <input
                                  className="mt-1 mb-2 block w-full rounded-md border border-gray-300 p-2"
                                  placeholder="Search tags..."
                                  value={poiTagSearch}
                                  onChange={(e) =>
                                    setPoiTagSearch(e.target.value)
                                  }
                                />
                                <div className="max-h-40 overflow-y-auto rounded-md border border-gray-200 bg-white p-2">
                                  {filteredTags.length === 0 && (
                                    <div className="text-xs text-gray-500">
                                      No tags found.
                                    </div>
                                  )}
                                  {filteredTags.map((tag) => {
                                    const isSelected = poiTagFilterIds.includes(
                                      tag.id
                                    );
                                    return (
                                      <label
                                        key={tag.id}
                                        className="flex items-center gap-2 text-xs text-gray-700"
                                      >
                                        <input
                                          type="checkbox"
                                          checked={isSelected}
                                          onChange={(e) => {
                                            if (e.target.checked) {
                                              setPoiTagFilterIds((prev) => [
                                                ...prev,
                                                tag.id,
                                              ]);
                                            } else {
                                              setPoiTagFilterIds((prev) =>
                                                prev.filter(
                                                  (id) => id !== tag.id
                                                )
                                              );
                                            }
                                          }}
                                        />
                                        {tag.name}
                                      </label>
                                    );
                                  })}
                                </div>
                              </div>
                              <div className="mt-3 flex items-center gap-2 text-xs">
                                <button
                                  type="button"
                                  className="rounded-md border border-gray-300 bg-white px-2 py-1"
                                  onClick={() => {
                                    setPoiZoneFilterId('');
                                    setPoiTagFilterIds([]);
                                    setPoiTagSearch('');
                                  }}
                                >
                                  Clear filters
                                </button>
                                <span className="text-gray-500">
                                  Showing {filteredPointsOfInterest.length} /{' '}
                                  {pointsOfInterest.length}
                                </span>
                              </div>
                            </div>
                          )}
                          <select
                            className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                            value={nodeForm.pointOfInterestId}
                            onChange={(e) =>
                              setNodeForm((prev) => ({
                                ...prev,
                                pointOfInterestId: e.target.value,
                              }))
                            }
                          >
                            <option value="">Select a POI</option>
                            {filteredPointsOfInterest.map((poi) => (
                              <option key={poi.id} value={poi.id}>
                                {poi.name}
                              </option>
                            ))}
                          </select>
                        </div>
                      ) : nodeForm.nodeType === 'scenario' ? (
                        <div className="md:col-span-2">
                          <div className="flex items-center justify-between gap-3">
                            <label className="block text-sm font-medium text-gray-700">
                              Scenario
                            </label>
                            <button
                              type="button"
                              className="rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                              onClick={() => toggleQuickCreate('scenario')}
                            >
                              {quickCreateOpen.scenario
                                ? 'Hide Quick Create'
                                : 'Create New Scenario'}
                            </button>
                          </div>
                          <select
                            className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                            value={nodeForm.scenarioId}
                            onChange={(e) =>
                              setNodeForm((prev) => ({
                                ...prev,
                                scenarioId: e.target.value,
                              }))
                            }
                          >
                            <option value="">Select a scenario</option>
                            {filteredScenarios.map((scenario) => (
                              <option key={scenario.id} value={scenario.id}>
                                {summarizeScenarioPrompt(scenario.prompt)}
                              </option>
                            ))}
                          </select>
                          {nodeForm.scenarioId ? (
                            <div className="mt-2">
                              <Link
                                to={adminEntityHref(
                                  'scenario',
                                  nodeForm.scenarioId
                                )}
                                className={adminEntityLinkClass}
                              >
                                Open Scenario Page
                              </Link>
                            </div>
                          ) : null}
                          {quickCreateOpen.scenario && (
                            <div className="mt-3 rounded-md border border-gray-200 bg-white p-4 space-y-3">
                              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                                <label className="text-sm">
                                  Prompt
                                  <textarea
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    rows={3}
                                    value={quickCreateScenarioForm.prompt}
                                    onChange={(e) =>
                                      setQuickCreateScenarioForm((prev) => ({
                                        ...prev,
                                        prompt: e.target.value,
                                      }))
                                    }
                                  />
                                </label>
                                <div className="grid grid-cols-1 gap-3">
                                  <label className="text-sm">
                                    Image URL
                                    <input
                                      className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                      value={quickCreateScenarioForm.imageUrl}
                                      onChange={(e) =>
                                        setQuickCreateScenarioForm((prev) => ({
                                          ...prev,
                                          imageUrl: e.target.value,
                                        }))
                                      }
                                    />
                                  </label>
                                  <label className="text-sm">
                                    Thumbnail URL
                                    <input
                                      className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                      value={
                                        quickCreateScenarioForm.thumbnailUrl
                                      }
                                      onChange={(e) =>
                                        setQuickCreateScenarioForm((prev) => ({
                                          ...prev,
                                          thumbnailUrl: e.target.value,
                                        }))
                                      }
                                    />
                                  </label>
                                </div>
                                <label className="text-sm">
                                  Latitude
                                  <input
                                    type="number"
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={quickCreateScenarioForm.latitude}
                                    onChange={(e) =>
                                      setQuickCreateScenarioForm((prev) => ({
                                        ...prev,
                                        latitude: e.target.value,
                                      }))
                                    }
                                  />
                                </label>
                                <label className="text-sm">
                                  Longitude
                                  <input
                                    type="number"
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={quickCreateScenarioForm.longitude}
                                    onChange={(e) =>
                                      setQuickCreateScenarioForm((prev) => ({
                                        ...prev,
                                        longitude: e.target.value,
                                      }))
                                    }
                                  />
                                </label>
                              </div>

                              <div className="rounded-md border border-gray-200 p-3">
                                <div className="mb-2 flex items-center justify-between">
                                  <div className="text-sm font-medium text-gray-700">
                                    Options
                                  </div>
                                  <button
                                    type="button"
                                    className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                                    onClick={handleAddQuickScenarioOption}
                                  >
                                    Add Option
                                  </button>
                                </div>
                                <div className="space-y-3">
                                  {quickCreateScenarioForm.options.map(
                                    (option, index) => (
                                      <div
                                        key={`quick-scenario-option-${index}`}
                                        className="rounded-md border border-gray-200 p-3"
                                      >
                                        <div className="mb-2 flex items-center justify-between">
                                          <div className="text-xs font-semibold uppercase tracking-wide text-gray-500">
                                            Option {index + 1}
                                          </div>
                                          {quickCreateScenarioForm.options
                                            .length > 1 && (
                                            <button
                                              type="button"
                                              className="rounded-md border border-red-200 px-2 py-1 text-xs text-red-600 hover:bg-red-50"
                                              onClick={() =>
                                                handleRemoveQuickScenarioOption(
                                                  index
                                                )
                                              }
                                            >
                                              Remove
                                            </button>
                                          )}
                                        </div>
                                        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                                          <label className="text-sm md:col-span-2">
                                            Option Text
                                            <input
                                              className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                              value={option.optionText}
                                              onChange={(e) =>
                                                handleUpdateQuickScenarioOption(
                                                  index,
                                                  { optionText: e.target.value }
                                                )
                                              }
                                            />
                                          </label>
                                          <label className="text-sm">
                                            Stat
                                            <select
                                              className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                              value={option.statTag}
                                              onChange={(e) =>
                                                handleUpdateQuickScenarioOption(
                                                  index,
                                                  { statTag: e.target.value }
                                                )
                                              }
                                            >
                                              {questStatOptions.map((stat) => (
                                                <option
                                                  key={stat.id}
                                                  value={stat.id}
                                                >
                                                  {stat.label}
                                                </option>
                                              ))}
                                            </select>
                                          </label>
                                          <label className="text-sm">
                                            Difficulty
                                            <input
                                              type="number"
                                              className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                              value={option.difficulty}
                                              onChange={(e) =>
                                                handleUpdateQuickScenarioOption(
                                                  index,
                                                  { difficulty: e.target.value }
                                                )
                                              }
                                            />
                                          </label>
                                          <label className="text-sm md:col-span-2">
                                            Proficiencies
                                            <input
                                              className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                              placeholder="comma, separated, proficiencies"
                                              value={option.proficiencies}
                                              onChange={(e) =>
                                                handleUpdateQuickScenarioOption(
                                                  index,
                                                  {
                                                    proficiencies:
                                                      e.target.value,
                                                  }
                                                )
                                              }
                                            />
                                          </label>
                                          <label className="text-sm">
                                            Success Text
                                            <textarea
                                              className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                              rows={2}
                                              value={option.successText}
                                              onChange={(e) =>
                                                handleUpdateQuickScenarioOption(
                                                  index,
                                                  {
                                                    successText: e.target.value,
                                                  }
                                                )
                                              }
                                            />
                                          </label>
                                          <label className="text-sm">
                                            Failure Text
                                            <textarea
                                              className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                              rows={2}
                                              value={option.failureText}
                                              onChange={(e) =>
                                                handleUpdateQuickScenarioOption(
                                                  index,
                                                  {
                                                    failureText: e.target.value,
                                                  }
                                                )
                                              }
                                            />
                                          </label>
                                        </div>
                                      </div>
                                    )
                                  )}
                                </div>
                              </div>

                              <button
                                type="button"
                                className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:bg-blue-300"
                                onClick={handleCreateStandaloneScenario}
                                disabled={quickCreateSubmitting === 'scenario'}
                              >
                                {quickCreateSubmitting === 'scenario'
                                  ? 'Creating Scenario...'
                                  : 'Create and Select Scenario'}
                              </button>
                            </div>
                          )}
                        </div>
                      ) : nodeForm.nodeType === 'monster' ? (
                        <div className="md:col-span-2">
                          <div className="flex items-center justify-between gap-3">
                            <label className="block text-sm font-medium text-gray-700">
                              Monster Encounter
                            </label>
                            <button
                              type="button"
                              className="rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                              onClick={() => toggleQuickCreate('monster')}
                            >
                              {quickCreateOpen.monster
                                ? 'Hide Quick Create'
                                : 'Create New Encounter'}
                            </button>
                          </div>
                          <select
                            className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                            value={nodeForm.monsterEncounterId}
                            onChange={(e) =>
                              setNodeForm((prev) => ({
                                ...prev,
                                monsterEncounterId: e.target.value,
                              }))
                            }
                          >
                            <option value="">Select a monster encounter</option>
                            {filteredMonsters.map((monster) => (
                              <option key={monster.id} value={monster.id}>
                                {monster.name}
                                {monster.monsterCount &&
                                monster.monsterCount > 1
                                  ? ` (${monster.monsterCount} monsters)`
                                  : ''}
                              </option>
                            ))}
                          </select>
                          {nodeForm.monsterEncounterId ? (
                            <div className="mt-2">
                              <Link
                                to={adminEntityHref(
                                  'monster',
                                  nodeForm.monsterEncounterId
                                )}
                                className={adminEntityLinkClass}
                              >
                                Open Monster Encounters Page
                              </Link>
                            </div>
                          ) : null}
                          {quickCreateOpen.monster && (
                            <div className="mt-3 rounded-md border border-gray-200 bg-white p-4 space-y-3">
                              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                                <label className="text-sm">
                                  Name
                                  <input
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={quickCreateMonsterEncounterForm.name}
                                    onChange={(e) =>
                                      setQuickCreateMonsterEncounterForm(
                                        (prev) => ({
                                          ...prev,
                                          name: e.target.value,
                                        })
                                      )
                                    }
                                  />
                                </label>
                                <label className="text-sm">
                                  Description
                                  <input
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={
                                      quickCreateMonsterEncounterForm.description
                                    }
                                    onChange={(e) =>
                                      setQuickCreateMonsterEncounterForm(
                                        (prev) => ({
                                          ...prev,
                                          description: e.target.value,
                                        })
                                      )
                                    }
                                  />
                                </label>
                                <label className="text-sm">
                                  Image URL
                                  <input
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={
                                      quickCreateMonsterEncounterForm.imageUrl
                                    }
                                    onChange={(e) =>
                                      setQuickCreateMonsterEncounterForm(
                                        (prev) => ({
                                          ...prev,
                                          imageUrl: e.target.value,
                                        })
                                      )
                                    }
                                  />
                                </label>
                                <label className="text-sm">
                                  Thumbnail URL
                                  <input
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={
                                      quickCreateMonsterEncounterForm.thumbnailUrl
                                    }
                                    onChange={(e) =>
                                      setQuickCreateMonsterEncounterForm(
                                        (prev) => ({
                                          ...prev,
                                          thumbnailUrl: e.target.value,
                                        })
                                      )
                                    }
                                  />
                                </label>
                                <label className="text-sm">
                                  Latitude
                                  <input
                                    type="number"
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={
                                      quickCreateMonsterEncounterForm.latitude
                                    }
                                    onChange={(e) =>
                                      setQuickCreateMonsterEncounterForm(
                                        (prev) => ({
                                          ...prev,
                                          latitude: e.target.value,
                                        })
                                      )
                                    }
                                  />
                                </label>
                                <label className="text-sm">
                                  Longitude
                                  <input
                                    type="number"
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={
                                      quickCreateMonsterEncounterForm.longitude
                                    }
                                    onChange={(e) =>
                                      setQuickCreateMonsterEncounterForm(
                                        (prev) => ({
                                          ...prev,
                                          longitude: e.target.value,
                                        })
                                      )
                                    }
                                  />
                                </label>
                              </div>
                              <label className="flex items-center gap-2 text-sm text-gray-700">
                                <input
                                  type="checkbox"
                                  checked={
                                    quickCreateMonsterEncounterForm.scaleWithUserLevel
                                  }
                                  onChange={(e) =>
                                    setQuickCreateMonsterEncounterForm(
                                      (prev) => ({
                                        ...prev,
                                        scaleWithUserLevel: e.target.checked,
                                      })
                                    )
                                  }
                                />
                                Scale encounter with user level
                              </label>
                              <div>
                                <div className="mb-2 text-sm font-medium text-gray-700">
                                  Monsters
                                </div>
                                <div className="max-h-48 space-y-2 overflow-y-auto rounded-md border border-gray-200 p-3">
                                  {availableMonstersForQuickCreate.length ===
                                  0 ? (
                                    <div className="text-sm text-gray-500">
                                      No monsters available in this quest zone.
                                    </div>
                                  ) : (
                                    availableMonstersForQuickCreate.map(
                                      (monster) => {
                                        const checked =
                                          quickCreateMonsterEncounterForm.monsterIds.includes(
                                            monster.id
                                          );
                                        return (
                                          <label
                                            key={monster.id}
                                            className="flex items-center gap-2 text-sm text-gray-700"
                                          >
                                            <input
                                              type="checkbox"
                                              checked={checked}
                                              onChange={(e) => {
                                                setQuickCreateMonsterEncounterForm(
                                                  (prev) => ({
                                                    ...prev,
                                                    monsterIds: e.target.checked
                                                      ? [
                                                          ...prev.monsterIds,
                                                          monster.id,
                                                        ]
                                                      : prev.monsterIds.filter(
                                                          (id) =>
                                                            id !== monster.id
                                                        ),
                                                  })
                                                );
                                              }}
                                            />
                                            <span>{monster.name}</span>
                                            {typeof monster.level ===
                                              'number' && (
                                              <span className="text-xs text-gray-500">
                                                Lvl {monster.level}
                                              </span>
                                            )}
                                          </label>
                                        );
                                      }
                                    )
                                  )}
                                </div>
                              </div>
                              <button
                                type="button"
                                className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:bg-blue-300"
                                onClick={handleCreateMonsterEncounter}
                                disabled={quickCreateSubmitting === 'monster'}
                              >
                                {quickCreateSubmitting === 'monster'
                                  ? 'Creating Encounter...'
                                  : 'Create and Select Encounter'}
                              </button>
                            </div>
                          )}
                        </div>
                      ) : nodeForm.nodeType === 'challenge' ? (
                        <div className="md:col-span-2">
                          <div className="flex items-center justify-between gap-3">
                            <label className="block text-sm font-medium text-gray-700">
                              Challenge Objective
                            </label>
                            <button
                              type="button"
                              className="rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                              onClick={() => toggleQuickCreate('challenge')}
                            >
                              {quickCreateOpen.challenge
                                ? 'Hide Quick Create'
                                : 'Create New Challenge Objective'}
                            </button>
                          </div>
                          <select
                            className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                            value={nodeForm.challengeId}
                            onChange={(e) =>
                              setNodeForm((prev) => ({
                                ...prev,
                                challengeId: e.target.value,
                              }))
                            }
                          >
                            <option value="">Select a challenge objective</option>
                            {filteredChallenges.map((challenge) => (
                              <option key={challenge.id} value={challenge.id}>
                                {challenge.question}
                              </option>
                            ))}
                          </select>
                          {nodeForm.challengeId ? (
                            <div className="mt-2">
                              <Link
                                to={adminEntityHref(
                                  'challenge',
                                  nodeForm.challengeId
                                )}
                                className={adminEntityLinkClass}
                              >
                                Open Challenge Page
                              </Link>
                            </div>
                          ) : null}
                          {quickCreateOpen.challenge && (
                            <div className="mt-3 rounded-md border border-gray-200 bg-white p-4 space-y-3">
                              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                                <div className="md:col-span-2">
                                  <SearchableSelect
                                    label="Point of Interest (Optional)"
                                    placeholder="Search points of interest..."
                                    options={quickCreateChallengePoiOptions}
                                    value={
                                      quickCreateChallengeForm.pointOfInterestId
                                    }
                                    onChange={(pointOfInterestId) => {
                                      const selectedPoint =
                                        quickCreateChallengePointsOfInterest.find(
                                          (point) =>
                                            point.id === pointOfInterestId
                                        );
                                      setQuickCreateChallengeForm((prev) => ({
                                        ...prev,
                                        pointOfInterestId,
                                        latitude:
                                          selectedPoint?.lat !== undefined
                                            ? selectedPoint.lat
                                            : prev.latitude,
                                        longitude:
                                          selectedPoint?.lng !== undefined
                                            ? selectedPoint.lng
                                            : prev.longitude,
                                      }));
                                    }}
                                    disabled={!questForm.zoneId}
                                    noMatchesLabel={
                                      questForm.zoneId
                                        ? 'No matching points of interest.'
                                        : 'Select a quest zone first.'
                                    }
                                  />
                                  {zonePoiMapLoading && questForm.zoneId ? (
                                    <div className="mt-1 text-xs text-gray-500">
                                      Loading points of interest...
                                    </div>
                                  ) : null}
                                </div>
                                <label className="text-sm md:col-span-2">
                                  Question
                                  <textarea
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    rows={2}
                                    value={quickCreateChallengeForm.question}
                                    onChange={(e) =>
                                      setQuickCreateChallengeForm((prev) => ({
                                        ...prev,
                                        question: e.target.value,
                                      }))
                                    }
                                  />
                                </label>
                                <label className="text-sm md:col-span-2">
                                  Description
                                  <textarea
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    rows={2}
                                    value={quickCreateChallengeForm.description}
                                    onChange={(e) =>
                                      setQuickCreateChallengeForm((prev) => ({
                                        ...prev,
                                        description: e.target.value,
                                      }))
                                    }
                                  />
                                </label>
                                <label className="text-sm">
                                  Image URL
                                  <input
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={quickCreateChallengeForm.imageUrl}
                                    onChange={(e) =>
                                      setQuickCreateChallengeForm((prev) => ({
                                        ...prev,
                                        imageUrl: e.target.value,
                                      }))
                                    }
                                  />
                                </label>
                                <label className="text-sm">
                                  Thumbnail URL
                                  <input
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={
                                      quickCreateChallengeForm.thumbnailUrl
                                    }
                                    onChange={(e) =>
                                      setQuickCreateChallengeForm((prev) => ({
                                        ...prev,
                                        thumbnailUrl: e.target.value,
                                      }))
                                    }
                                  />
                                </label>
                                <label className="text-sm">
                                  Latitude
                                  <input
                                    type="number"
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={quickCreateChallengeForm.latitude}
                                    onChange={(e) =>
                                      setQuickCreateChallengeForm((prev) => ({
                                        ...prev,
                                        pointOfInterestId: '',
                                        latitude: e.target.value,
                                      }))
                                    }
                                    disabled={Boolean(
                                      quickCreateChallengeForm.pointOfInterestId
                                    )}
                                  />
                                </label>
                                <label className="text-sm">
                                  Longitude
                                  <input
                                    type="number"
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={quickCreateChallengeForm.longitude}
                                    onChange={(e) =>
                                      setQuickCreateChallengeForm((prev) => ({
                                        ...prev,
                                        pointOfInterestId: '',
                                        longitude: e.target.value,
                                      }))
                                    }
                                    disabled={Boolean(
                                      quickCreateChallengeForm.pointOfInterestId
                                    )}
                                  />
                                </label>
                                <label className="text-sm">
                                  Submission Type
                                  <select
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={
                                      quickCreateChallengeForm.submissionType
                                    }
                                    onChange={(e) =>
                                      setQuickCreateChallengeForm((prev) => ({
                                        ...prev,
                                        submissionType: e.target
                                          .value as QuestNodeSubmissionType,
                                      }))
                                    }
                                  >
                                    {questNodeSubmissionOptions.map(
                                      (option) => (
                                        <option
                                          key={option.value}
                                          value={option.value}
                                        >
                                          {option.label}
                                        </option>
                                      )
                                    )}
                                  </select>
                                </label>
                                <label className="text-sm">
                                  Difficulty
                                  <input
                                    type="number"
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={quickCreateChallengeForm.difficulty}
                                    onChange={(e) =>
                                      setQuickCreateChallengeForm((prev) => ({
                                        ...prev,
                                        difficulty: e.target.value,
                                      }))
                                    }
                                  />
                                </label>
                                <label className="text-sm">
                                  Reward XP
                                  <input
                                    type="number"
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={
                                      quickCreateChallengeForm.rewardExperience
                                    }
                                    onChange={(e) =>
                                      setQuickCreateChallengeForm((prev) => ({
                                        ...prev,
                                        rewardExperience: e.target.value,
                                      }))
                                    }
                                  />
                                </label>
                                <label className="text-sm">
                                  Reward Gold
                                  <input
                                    type="number"
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={quickCreateChallengeForm.rewardGold}
                                    onChange={(e) =>
                                      setQuickCreateChallengeForm((prev) => ({
                                        ...prev,
                                        rewardGold: e.target.value,
                                      }))
                                    }
                                  />
                                </label>
                                <label className="text-sm md:col-span-2">
                                  Proficiency
                                  <input
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={quickCreateChallengeForm.proficiency}
                                    onChange={(e) =>
                                      setQuickCreateChallengeForm((prev) => ({
                                        ...prev,
                                        proficiency: e.target.value,
                                      }))
                                    }
                                  />
                                </label>
                              </div>
                              <div>
                                <div className="mb-2 text-sm font-medium text-gray-700">
                                  Stat Tags
                                </div>
                                <div className="flex flex-wrap gap-3">
                                  {questStatOptions.map((stat) => (
                                    <label
                                      key={`quick-challenge-stat-${stat.id}`}
                                      className="flex items-center gap-2 text-sm text-gray-700"
                                    >
                                      <input
                                        type="checkbox"
                                        checked={quickCreateChallengeForm.statTags.includes(
                                          stat.id
                                        )}
                                        onChange={(e) =>
                                          setQuickCreateChallengeForm(
                                            (prev) => ({
                                              ...prev,
                                              statTags: e.target.checked
                                                ? [...prev.statTags, stat.id]
                                                : prev.statTags.filter(
                                                    (tag) => tag !== stat.id
                                                  ),
                                            })
                                          )
                                        }
                                      />
                                      {stat.label}
                                    </label>
                                  ))}
                                </div>
                              </div>
                              <button
                                type="button"
                                className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:bg-blue-300"
                                onClick={handleCreateStandaloneChallenge}
                                disabled={quickCreateSubmitting === 'challenge'}
                              >
                                {quickCreateSubmitting === 'challenge'
                                  ? 'Creating Challenge...'
                                  : 'Create and Select Challenge'}
                              </button>
                            </div>
                          )}
                        </div>
                      ) : (
                        <div className="md:col-span-2">
                          <label className="block text-sm font-medium text-gray-700">
                            Polygon Points
                          </label>
                          <textarea
                            className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                            rows={3}
                            placeholder="[[lng,lat],[lng,lat],[lng,lat]]"
                            value={nodeForm.polygonPoints}
                            onChange={(e) => {
                              const value = e.target.value;
                              setNodeForm((prev) => ({
                                ...prev,
                                polygonPoints: value,
                              }));
                              const parsed = parsePolygonPoints(value);
                              if (parsed) {
                                setPolygonDraftPoints(parsed);
                              }
                            }}
                          />
                          <div className="mt-2 flex flex-wrap gap-2 text-xs text-gray-600">
                            <span className="rounded-md bg-teal-50 px-2 py-1 text-teal-700">
                              Click on the map to add polygon points.
                            </span>
                            <button
                              type="button"
                              className="rounded-md border border-gray-300 bg-white px-2 py-1 text-gray-700 hover:bg-gray-50"
                              onClick={() => {
                                setPolygonDraftPoints([]);
                                setNodeForm((prev) => ({
                                  ...prev,
                                  polygonPoints: '',
                                }));
                              }}
                            >
                              Clear polygon
                            </button>
                            <button
                              type="button"
                              className="rounded-md border border-gray-300 bg-white px-2 py-1 text-gray-700 hover:bg-gray-50"
                              onClick={() => {
                                setPolygonDraftPoints((prev) => {
                                  if (prev.length === 0) return prev;
                                  const next = prev.slice(0, -1);
                                  setNodeForm((formPrev) => ({
                                    ...formPrev,
                                    polygonPoints: JSON.stringify(next),
                                  }));
                                  return next;
                                });
                              }}
                            >
                              Undo last point
                            </button>
                          </div>
                        </div>
                      )}
                    </div>
                    <button
                      className="mt-4 bg-green-600 text-white px-4 py-2 rounded-md"
                      onClick={handleCreateNode}
                      disabled={
                        nodeForm.nodeType === 'poi' ||
                        nodeForm.nodeType === 'polygon' ||
                        (nodeForm.nodeType === 'poi' &&
                          !nodeForm.pointOfInterestId) ||
                        (nodeForm.nodeType === 'scenario' &&
                          !nodeForm.scenarioId) ||
                        (nodeForm.nodeType === 'monster' &&
                          !nodeForm.monsterEncounterId) ||
                        (nodeForm.nodeType === 'challenge' &&
                          !nodeForm.challengeId)
                      }
                    >
                      Add Node
                    </button>
                  </div>

                  <div className="mb-6 rounded-lg border border-gray-200 bg-white p-4">
                    <div className="flex items-center justify-between">
                      <h4 className="font-semibold">Quest Map</h4>
                      <div className="flex items-center gap-2">
                        <button
                          type="button"
                          className="rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                          onClick={focusQuestArea}
                        >
                          Focus Quest Area
                        </button>
                        <button
                          type="button"
                          className="rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                          onClick={() =>
                            setPolygonRefreshNonce((prev) => prev + 1)
                          }
                        >
                          Add Polygons to Map
                        </button>
                      </div>
                      <div className="flex items-center gap-3 text-xs text-gray-600">
                        <span className="flex items-center gap-1">
                          <span className="inline-block h-2.5 w-2.5 rounded-full bg-amber-500 border border-amber-800" />
                          Quest locations
                        </span>
                        <span className="flex items-center gap-1">
                          <span className="inline-block h-2.5 w-2.5 rounded-full bg-teal-500 border border-teal-800" />
                          Scenario nodes
                        </span>
                        <span className="flex items-center gap-1">
                          <span className="inline-block h-2.5 w-2.5 rounded-full bg-red-500 border border-red-900" />
                          Monster nodes
                        </span>
                      </div>
                    </div>
                    <p className="mt-1 text-sm text-gray-600">
                      The map now shows only this quest&apos;s nodes and polygons,
                      then automatically zooms to the quest footprint.
                    </p>
                    <div className="mt-3 h-80 w-full overflow-hidden rounded-md border border-gray-200 relative">
                      <div ref={questMapContainer} className="h-full w-full" />
                    </div>
                  </div>

                  {showImportModal && (
                    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
                      <div className="w-full max-w-2xl max-h-[90vh] overflow-y-auto rounded-lg bg-white p-6 shadow-lg">
                        <div className="flex items-start justify-between">
                          <h3 className="text-lg font-semibold">
                            Import Point of Interest
                          </h3>
                          <button
                            className="text-gray-500 hover:text-gray-700"
                            onClick={() => {
                              setShowImportModal(false);
                              resetImportForm();
                            }}
                          >
                            Close
                          </button>
                        </div>

                        {importError && (
                          <div className="mt-3 text-sm text-red-600">
                            {importError}
                          </div>
                        )}

                        <div className="mt-4">
                          <label className="block text-sm font-medium text-gray-700 mb-1">
                            Zone
                          </label>
                          <select
                            className="w-full border border-gray-300 rounded-md px-3 py-2"
                            value={importZoneId || questForm.zoneId}
                            onChange={(e) => setImportZoneId(e.target.value)}
                          >
                            <option value="">Select a zone</option>
                            {zones.map((zone) => (
                              <option key={zone.id} value={zone.id}>
                                {zone.name}
                              </option>
                            ))}
                          </select>
                        </div>

                        <div className="mt-4">
                          <label className="block text-sm font-medium text-gray-700 mb-1">
                            Search Google Maps
                          </label>
                          <input
                            type="text"
                            className="w-full border border-gray-300 rounded-md px-3 py-2"
                            value={importQuery}
                            onChange={(e) => setImportQuery(e.target.value)}
                            placeholder="Search for a place..."
                          />
                        </div>

                        <div className="mt-4 border border-gray-200 rounded-md max-h-64 overflow-y-auto">
                          {candidates.length === 0 && (
                            <div className="p-4 text-sm text-gray-500">
                              No results yet.
                            </div>
                          )}
                          {candidates.map((candidate) => (
                            <button
                              key={candidate.place_id}
                              type="button"
                              className={`w-full text-left px-4 py-3 border-b border-gray-100 hover:bg-gray-50 ${
                                selectedCandidate?.place_id ===
                                candidate.place_id
                                  ? 'bg-blue-50'
                                  : ''
                              }`}
                              onClick={() => setSelectedCandidate(candidate)}
                            >
                              <div className="font-medium">
                                {candidate.name}
                              </div>
                              <div className="text-xs text-gray-500">
                                {candidate.formatted_address}
                              </div>
                            </button>
                          ))}
                        </div>

                        <div className="mt-6">
                          <div className="flex items-center justify-between mb-2">
                            <h4 className="text-sm font-semibold">
                              Import Status
                            </h4>
                            <button
                              type="button"
                              className="text-xs text-blue-600"
                              onClick={() =>
                                fetchImportJobs(
                                  importZoneId || questForm.zoneId || undefined
                                )
                              }
                            >
                              Refresh
                            </button>
                          </div>
                          <div className="border border-gray-200 rounded-md max-h-40 overflow-y-auto">
                            {importJobs.length === 0 && (
                              <div className="p-3 text-xs text-gray-500">
                                No import activity yet.
                              </div>
                            )}
                            {importJobs.map((job) => (
                              <div
                                key={job.id}
                                className="flex items-center justify-between px-3 py-2 border-b border-gray-100 text-xs"
                              >
                                <div>
                                  <div className="font-medium">
                                    {job.placeId}
                                  </div>
                                  {job.errorMessage && (
                                    <div className="text-red-600">
                                      {job.errorMessage}
                                    </div>
                                  )}
                                </div>
                                <div className="flex items-center gap-2">
                                  <div className="uppercase text-[10px] text-gray-600">
                                    {job.status}
                                  </div>
                                  {job.status === 'failed' && (
                                    <button
                                      type="button"
                                      className="rounded-md border border-gray-300 px-2 py-1 text-[10px] text-gray-700"
                                      onClick={() =>
                                        handleRetryImport(
                                          job.placeId,
                                          job.zoneId
                                        )
                                      }
                                    >
                                      Retry
                                    </button>
                                  )}
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>

                        <div className="mt-6 flex justify-end gap-2">
                          <button
                            type="button"
                            className="px-4 py-2 rounded-md border border-gray-300"
                            onClick={() => {
                              setShowImportModal(false);
                              resetImportForm();
                            }}
                          >
                            Cancel
                          </button>
                          <button
                            type="button"
                            className="px-4 py-2 rounded-md bg-green-600 text-white"
                            onClick={handleImportPointOfInterest}
                          >
                            Import
                          </button>
                        </div>
                      </div>
                    </div>
                  )}

                  {characterLocationsOpen && (
                    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
                      <div className="w-full max-w-lg rounded-lg bg-white p-6 shadow-lg">
                        <div className="flex items-start justify-between">
                          <h3 className="text-lg font-semibold">
                            Character Locations
                          </h3>
                          <button
                            className="text-gray-500 hover:text-gray-700"
                            onClick={() => setCharacterLocationsOpen(false)}
                          >
                            Close
                          </button>
                        </div>
                        {characterLocationsLoading ? (
                          <div className="mt-4 text-sm text-gray-600">
                            Loading locations...
                          </div>
                        ) : (
                          <div className="mt-4 space-y-3">
                            {selectedCharacterLocations.length === 0 && (
                              <div className="text-sm text-gray-500">
                                No locations yet.
                              </div>
                            )}
                            {selectedCharacterLocations.map(
                              (location, index) => (
                                <div
                                  key={`${location.latitude}-${location.longitude}-${index}`}
                                  className="flex items-center gap-2"
                                >
                                  <input
                                    type="number"
                                    className="w-1/2 rounded-md border border-gray-300 p-2 text-sm"
                                    value={location.latitude}
                                    onChange={(e) =>
                                      handleUpdateCharacterLocation(
                                        index,
                                        'latitude',
                                        Number(e.target.value)
                                      )
                                    }
                                    placeholder="Latitude"
                                  />
                                  <input
                                    type="number"
                                    className="w-1/2 rounded-md border border-gray-300 p-2 text-sm"
                                    value={location.longitude}
                                    onChange={(e) =>
                                      handleUpdateCharacterLocation(
                                        index,
                                        'longitude',
                                        Number(e.target.value)
                                      )
                                    }
                                    placeholder="Longitude"
                                  />
                                  <button
                                    className="rounded-md border border-red-200 bg-red-50 px-2 py-1 text-xs text-red-700"
                                    onClick={() =>
                                      handleRemoveCharacterLocation(index)
                                    }
                                  >
                                    Remove
                                  </button>
                                </div>
                              )
                            )}
                          </div>
                        )}
                        <div className="mt-4 flex items-center justify-between">
                          <button
                            className="rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                            onClick={handleAddCharacterLocation}
                          >
                            Add Location
                          </button>
                          <button
                            className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white"
                            onClick={handleSaveCharacterLocations}
                            disabled={characterLocationsLoading}
                          >
                            Save Locations
                          </button>
                        </div>
                      </div>
                    </div>
                  )}

                  <div className="qa-route-list-header">
                    <div>
                      <div className="qa-card-kicker">Current Route</div>
                      <div className="qa-section-title">Node Breakdown</div>
                    </div>
                    <div className="qa-section-copy">
                      Modern quest nodes use their linked objective directly.
                      Any nested prompt rows shown here are legacy-only data.
                    </div>
                  </div>

                  <div className="space-y-4">
                    {orderedQuestNodes.map((node) => {
                        const linkedScenario = resolveLinkedQuestScenario(
                          node,
                          scenarios
                        );
                        const linkedMonsterEncounter =
                          resolveLinkedQuestMonsterEncounter(
                            node,
                            monsterEncounters
                          );
                        const linkedChallenge = resolveLinkedQuestChallenge(
                          node,
                          challenges
                        );
                        const isChallengeTargetNode = Boolean(node.challengeId);
                        const isScenarioTargetNode = Boolean(node.scenarioId);
                        const isMonsterTargetNode = Boolean(
                          node.monsterEncounterId || node.monsterId
                        );
                        const usesLinkedObjective =
                          questNodeUsesLinkedObjective(node);
                        const nodeObjectivePrompts = getLegacyQuestNodePrompts(node);
                        const challengeObjectiveSubmissionType = (
                          linkedChallenge?.submissionType ||
                          node.submissionType ||
                          'photo'
                        ) as QuestNodeSubmissionType;
                        const linkedPoi = resolveLinkedQuestNodePoi(
                          node,
                          pointsOfInterest,
                          linkedScenario,
                          linkedMonsterEncounter,
                          linkedChallenge
                        );
                        const linkedPoiArchetype = linkedPoi
                          ? archetypeByPoiId[linkedPoi.id] ?? null
                          : null;
                        const linkedPoiAliases = getPointOfInterestAliases(
                          linkedPoi
                        );
                        const linkedPoiSummary = linkedPoi
                          ? summarizePoiText(linkedPoi.description) ||
                            summarizePoiText(linkedPoi.clue)
                          : '';
                        const linkedPoiTagNames = linkedPoi
                          ? (linkedPoi.tags ?? [])
                              .map((tag) => tag.name)
                              .filter(Boolean)
                          : [];
                        const linkedPoiLatitude = linkedPoi
                          ? formatPoiCoordinate(linkedPoi.lat)
                          : null;
                        const linkedPoiLongitude = linkedPoi
                          ? formatPoiCoordinate(linkedPoi.lng)
                          : null;
                        const linkedObjectiveCoordinates =
                          linkedPoi
                            ? null
                            : linkedScenario
                              ? formatNodeCoordinatePair(
                                  linkedScenario.latitude,
                                  linkedScenario.longitude
                                )
                              : linkedMonsterEncounter
                                ? formatNodeCoordinatePair(
                                    linkedMonsterEncounter.latitude,
                                    linkedMonsterEncounter.longitude
                                  )
                                : linkedChallenge
                                  ? formatNodeCoordinatePair(
                                      linkedChallenge.latitude,
                                      linkedChallenge.longitude
                                    )
                                  : null;
                        const hasLegacyChallengeOverrides =
                          usesLinkedObjective &&
                          nodeObjectivePrompts.length > 0;

                        return (
                          <div
                            key={node.id}
                            className="border border-gray-200 rounded-md p-4"
                          >
                            <div className="flex items-center justify-between">
                              <div>
                                <div className="font-semibold">
                                  Node {node.orderIndex}
                                </div>
                                <div className="mt-1 flex flex-wrap items-center gap-2 text-sm text-gray-600">
                                  <span>
                                    {node.pointOfInterestId
                                      ? `POI: ${pointsOfInterest.find((poi) => poi.id === node.pointOfInterestId)?.name ?? node.pointOfInterestId}`
                                      : node.scenarioId
                                        ? `Scenario: ${summarizeScenarioPrompt(linkedScenario?.prompt ?? '')}`
                                        : node.monsterEncounterId ||
                                            node.monsterId
                                          ? `Monster Encounter: ${
                                              linkedMonsterEncounter?.name ??
                                              node.monsterEncounterId ??
                                              node.monsterId
                                            }`
                                          : node.challengeId
                                            ? `Challenge Objective: ${linkedChallenge?.question ?? node.challengeId}`
                                            : 'Polygon'}
                                  </span>
                                  {node.scenarioId ? (
                                    <Link
                                      to={adminEntityHref(
                                        'scenario',
                                        node.scenarioId
                                      )}
                                      className={adminEntityLinkClass}
                                    >
                                      Open
                                    </Link>
                                  ) : node.monsterEncounterId ||
                                    node.monsterId ? (
                                    <Link
                                      to={adminEntityHref(
                                        'monster',
                                        node.monsterEncounterId ??
                                          node.monsterId ??
                                          ''
                                      )}
                                      className={adminEntityLinkClass}
                                    >
                                      Open
                                    </Link>
                                  ) : node.challengeId ? (
                                    <Link
                                      to={adminEntityHref(
                                        'challenge',
                                        node.challengeId
                                      )}
                                      className={adminEntityLinkClass}
                                    >
                                      Open
                                    </Link>
                                  ) : null}
                                </div>
                              </div>
                              <button
                                className="rounded-md border border-red-200 bg-red-50 px-3 py-1 text-xs text-red-700 hover:bg-red-100"
                                onClick={() => handleDeleteNode(node)}
                              >
                                Remove Node
                              </button>
                            </div>

                            <div className="mt-3 rounded-md border border-gray-200 bg-white p-3">
                              <div className="text-xs font-semibold uppercase tracking-wide text-gray-500">
                                Location
                              </div>
                              {linkedPoi ? (
                                <div className="mt-2 flex items-center gap-2 text-sm text-gray-700">
                                  <span>
                                    Point of Interest: <strong>{linkedPoi.name}</strong>
                                  </span>
                                  <Link
                                    to={`/points-of-interest/${linkedPoi.id}`}
                                    className={adminEntityLinkClass}
                                  >
                                    Open POI
                                  </Link>
                                </div>
                              ) : linkedObjectiveCoordinates ? (
                                <div className="mt-2 text-sm text-gray-700">
                                  Coordinates: <strong>{linkedObjectiveCoordinates}</strong>
                                </div>
                              ) : node.polygon ? (
                                <div className="mt-2 text-sm text-gray-700">
                                  Custom polygon area
                                </div>
                              ) : (
                                <div className="mt-2 text-sm text-gray-500">
                                  No linked location details available.
                                </div>
                              )}
                            </div>

                            {linkedPoi && (
                              <div className="mt-3 rounded-md border border-amber-200 bg-amber-50 p-3">
                                <div className="flex items-start justify-between gap-3">
                                  <div>
                                    <div className="text-xs font-semibold text-amber-900">
                                      Point of Interest
                                    </div>
                                    <div className="text-sm font-semibold text-gray-900">
                                      {linkedPoi.name}
                                    </div>
                                    {linkedPoiAliases.length > 0 && (
                                      <div className="mt-1 text-xs text-gray-600">
                                        {linkedPoiAliases.join(' · ')}
                                      </div>
                                    )}
                                  </div>
                                  <Link
                                    to={`/points-of-interest/${linkedPoi.id}`}
                                    className={adminEntityLinkClass}
                                  >
                                    Open POI
                                  </Link>
                                </div>
                                {linkedPoiSummary && (
                                  <p className="mt-2 text-sm text-gray-700">
                                    {linkedPoiSummary}
                                  </p>
                                )}
                                <div className="mt-2 flex flex-wrap gap-2">
                                  {linkedPoiArchetype && (
                                    <span className="qa-chip muted">
                                      Archetype: {linkedPoiArchetype.name}
                                    </span>
                                  )}
                                  {(linkedPoiLatitude || linkedPoiLongitude) && (
                                    <span className="qa-chip muted">
                                      {linkedPoiLatitude ?? '?'}, {linkedPoiLongitude ?? '?'}
                                    </span>
                                  )}
                                </div>
                                {linkedPoiTagNames.length > 0 && (
                                  <div className="mt-2 flex flex-wrap gap-2">
                                    {linkedPoiTagNames
                                      .slice(0, 6)
                                      .map((tagName) => (
                                        <span
                                          key={`${linkedPoi.id}-${tagName}`}
                                          className="rounded-full border border-amber-200 bg-white px-2 py-1 text-xs text-amber-900"
                                        >
                                          {tagName}
                                        </span>
                                      ))}
                                    {linkedPoiTagNames.length > 6 && (
                                      <span className="rounded-full border border-amber-200 bg-white px-2 py-1 text-xs text-gray-500">
                                        +{linkedPoiTagNames.length - 6} more
                                      </span>
                                    )}
                                  </div>
                                )}
                              </div>
                            )}

                            <div className="mt-3">
                              {usesLinkedObjective ? (
                                <div className="mb-3 rounded-md border border-gray-200 bg-gray-50 p-3">
                                  <div className="text-xs font-semibold uppercase tracking-wide text-gray-500">
                                    Linked Objective
                                  </div>
                                  {isChallengeTargetNode ? (
                                    <>
                                      <div className="mt-2 text-sm font-semibold text-gray-900">
                                        {linkedChallenge?.question ??
                                          'Challenge details unavailable'}
                                      </div>
                                      {linkedChallenge?.description ? (
                                        <p className="mt-2 text-sm text-gray-600">
                                          {linkedChallenge.description}
                                        </p>
                                      ) : null}
                                      <div className="mt-2 flex flex-wrap items-center gap-2 text-xs text-gray-500">
                                        <span>
                                          Input{' '}
                                          {challengeObjectiveSubmissionType.toUpperCase()}
                                        </span>
                                        <span>
                                          Difficulty{' '}
                                          {linkedChallenge?.difficulty ?? 0}
                                        </span>
                                        {(linkedChallenge?.statTags?.length ??
                                          0) > 0 && (
                                          <span>
                                            Stats:{' '}
                                            {linkedChallenge?.statTags
                                              ?.map(formatStatTagLabel)
                                              .join(', ')}
                                          </span>
                                        )}
                                        {linkedChallenge?.proficiency && (
                                          <span>
                                            Proficiency:{' '}
                                            {linkedChallenge.proficiency}
                                          </span>
                                        )}
                                      </div>
                                    </>
                                  ) : isScenarioTargetNode ? (
                                    <>
                                      <div className="mt-2 text-sm font-semibold text-gray-900">
                                        {linkedScenario?.prompt
                                          ? summarizeScenarioPrompt(
                                              linkedScenario.prompt
                                            )
                                          : 'Scenario details unavailable'}
                                      </div>
                                      <div className="mt-2 flex flex-wrap items-center gap-2 text-xs text-gray-500">
                                        <span>
                                          {linkedScenario?.openEnded
                                            ? 'Open-ended'
                                            : 'Choice scenario'}
                                        </span>
                                        <span>
                                          Difficulty{' '}
                                          {linkedScenario?.difficulty ?? 0}
                                        </span>
                                        <span>
                                          {(linkedScenario?.options?.length ??
                                            0) > 0
                                            ? `${linkedScenario?.options?.length ?? 0} options`
                                            : 'Single resolution path'}
                                        </span>
                                      </div>
                                    </>
                                  ) : isMonsterTargetNode ? (
                                    <>
                                      <div className="mt-2 text-sm font-semibold text-gray-900">
                                        {linkedMonsterEncounter?.name ??
                                          'Monster encounter details unavailable'}
                                      </div>
                                      {linkedMonsterEncounter?.description ? (
                                        <p className="mt-2 text-sm text-gray-600">
                                          {linkedMonsterEncounter.description}
                                        </p>
                                      ) : null}
                                      <div className="mt-2 flex flex-wrap items-center gap-2 text-xs text-gray-500">
                                        <span>
                                          Type{' '}
                                          {linkedMonsterEncounter?.encounterType ??
                                            'monster'}
                                        </span>
                                        <span>
                                          {(linkedMonsterEncounter?.members
                                            ?.length ??
                                            linkedMonsterEncounter?.monsterCount ??
                                            0) || 0}{' '}
                                          monsters
                                        </span>
                                        <span>
                                          {linkedMonsterEncounter?.scaleWithUserLevel
                                            ? 'Scales with player'
                                            : 'Fixed encounter'}
                                        </span>
                                      </div>
                                    </>
                                  ) : null}
                                  <p className="mt-2 text-xs text-gray-500">
                                    Target-backed quest nodes now use their
                                    linked objective directly. Nested
                                    quest-node prompt rows are legacy-only.
                                  </p>
                                </div>
                              ) : (
                                <>
                                  <h4 className="font-semibold mb-1">
                                    Objective Prompts
                                  </h4>
                                  <p className="mb-2 text-xs text-gray-500">
                                    These prompts define what the player
                                    actually submits or proves at this target.
                                  </p>
                                </>
                              )}
                              {hasLegacyChallengeOverrides && (
                                <div className="mb-3 rounded-md border border-amber-200 bg-amber-50 p-3">
                                  <div className="text-xs font-semibold text-amber-900">
                                    Legacy Prompt Overrides
                                  </div>
                                  <p className="mt-1 text-xs text-amber-800">
                                    This target-backed node still has older
                                    quest-specific prompt overrides attached.
                                    New challenge, scenario, and monster nodes
                                    no longer use these duplicates.
                                  </p>
                                </div>
                              )}
                              {(nodeObjectivePrompts.length > 0 ||
                                !usesLinkedObjective) && (
                                <div className="space-y-2 mb-3">
                                  {nodeObjectivePrompts.map((challenge) => {
                                  const editDraft =
                                    challengeEdits[challenge.id] ??
                                    emptyChallengeForm;
                                  const isEditing = Boolean(
                                    challengeEdits[challenge.id]
                                  );
                                  return (
                                    <div
                                      key={challenge.id}
                                      className="border border-gray-200 rounded-md p-2 text-sm"
                                    >
                                      <div className="flex items-start justify-between gap-3">
                                        <div>
                                          <div>
                                            Tier {challenge.tier} · Difficulty{' '}
                                            {challenge.difficulty ?? 0} · Reward{' '}
                                            {challenge.reward} · Input{' '}
                                            {resolveChallengeSubmissionType(
                                              challenge,
                                              node
                                            ).toUpperCase()}
                                          </div>
                                          <div className="text-xs text-gray-500">
                                            Shuffle:{' '}
                                            {formatChallengeShuffleStatus(
                                              challenge.challengeShuffleStatus
                                            )}
                                          </div>
                                          {!isEditing && (
                                            <>
                                              <div className="text-gray-600">
                                                {challenge.question}
                                              </div>
                                              {challenge.statTags &&
                                                challenge.statTags.length >
                                                  0 && (
                                                  <div className="text-xs text-gray-500">
                                                    Stats:{' '}
                                                    {challenge.statTags
                                                      .map(formatStatTagLabel)
                                                      .join(', ')}
                                                  </div>
                                                )}
                                              {challenge.proficiency && (
                                                <div className="text-xs text-gray-500">
                                                  Proficiency:{' '}
                                                  {challenge.proficiency}
                                                </div>
                                              )}
                                              {challenge.challengeShuffleError && (
                                                <div className="text-xs text-red-600">
                                                  Shuffle error:{' '}
                                                  {
                                                    challenge.challengeShuffleError
                                                  }
                                                </div>
                                              )}
                                            </>
                                          )}
                                        </div>
                                        {!usesLinkedObjective && (
                                          <div className="flex items-center gap-2">
                                            <button
                                              type="button"
                                              className="rounded-md border border-indigo-200 bg-indigo-50 px-2 py-1 text-xs text-indigo-700 hover:bg-indigo-100 disabled:opacity-60"
                                              onClick={() =>
                                                handleShuffleSavedChallenge(
                                                  node,
                                                  challenge
                                                )
                                              }
                                              disabled={
                                                isEditing ||
                                                shufflingChallengeId ===
                                                  challenge.id ||
                                                challenge.challengeShuffleStatus ===
                                                  'queued' ||
                                                challenge.challengeShuffleStatus ===
                                                  'in_progress'
                                              }
                                            >
                                              {shufflingChallengeId ===
                                                challenge.id ||
                                              challenge.challengeShuffleStatus ===
                                                'queued' ||
                                              challenge.challengeShuffleStatus ===
                                                'in_progress'
                                                ? 'Shuffling...'
                                                : 'Shuffle'}
                                            </button>
                                            <button
                                              type="button"
                                              className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                                              onClick={() =>
                                                isEditing
                                                  ? handleCancelEditChallenge(
                                                      challenge.id
                                                    )
                                                  : handleStartEditChallenge(
                                                      node,
                                                      challenge
                                                    )
                                              }
                                            >
                                              {isEditing ? 'Cancel' : 'Edit'}
                                            </button>
                                          </div>
                                        )}
                                      </div>
                                      {isEditing && (
                                        <div className="mt-3">
                                          <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
                                            <div>
                                              <label className="block text-xs font-medium text-gray-700">
                                                Tier
                                              </label>
                                              <input
                                                type="number"
                                                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                                value={editDraft.tier}
                                                onChange={(e) =>
                                                  handleEditChallengeDraftChange(
                                                    challenge.id,
                                                    {
                                                      tier: Number(
                                                        e.target.value
                                                      ),
                                                    }
                                                  )
                                                }
                                              />
                                            </div>
                                            <div>
                                              <label className="block text-xs font-medium text-gray-700">
                                                Difficulty
                                              </label>
                                              <input
                                                type="number"
                                                min={0}
                                                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                                value={editDraft.difficulty}
                                                onChange={(e) =>
                                                  handleEditChallengeDraftChange(
                                                    challenge.id,
                                                    {
                                                      difficulty: Number(
                                                        e.target.value
                                                      ),
                                                    }
                                                  )
                                                }
                                              />
                                            </div>
                                            <div>
                                              <label className="block text-xs font-medium text-gray-700">
                                                Input Type
                                              </label>
                                              <select
                                                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                                value={editDraft.submissionType}
                                                onChange={(e) =>
                                                  handleEditChallengeDraftChange(
                                                    challenge.id,
                                                    {
                                                      submissionType: e.target
                                                        .value as QuestNodeSubmissionType,
                                                    }
                                                  )
                                                }
                                              >
                                                {questNodeSubmissionOptions.map(
                                                  (option) => (
                                                    <option
                                                      key={option.value}
                                                      value={option.value}
                                                    >
                                                      {option.label}
                                                    </option>
                                                  )
                                                )}
                                              </select>
                                            </div>
                                            <div>
                                              <label className="block text-xs font-medium text-gray-700">
                                                Reward
                                              </label>
                                              <input
                                                type="number"
                                                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                                value={editDraft.reward}
                                                onChange={(e) =>
                                                  handleEditChallengeDraftChange(
                                                    challenge.id,
                                                    {
                                                      reward: Number(
                                                        e.target.value
                                                      ),
                                                    }
                                                  )
                                                }
                                              />
                                            </div>
                                            <div>
                                              <label className="block text-xs font-medium text-gray-700">
                                                Inventory Item
                                              </label>
                                              <select
                                                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                                value={
                                                  editDraft.inventoryItemId
                                                }
                                                onChange={(e) =>
                                                  handleEditChallengeDraftChange(
                                                    challenge.id,
                                                    {
                                                      inventoryItemId:
                                                        e.target.value,
                                                    }
                                                  )
                                                }
                                              >
                                                <option value="">None</option>
                                                {inventoryItems.map((item) => (
                                                  <option
                                                    key={item.id}
                                                    value={item.id}
                                                  >
                                                    {item.name}
                                                  </option>
                                                ))}
                                              </select>
                                            </div>
                                            <div className="md:col-span-4">
                                              <label className="block text-xs font-medium text-gray-700">
                                                Stat Tags
                                              </label>
                                              <div className="mt-2 grid grid-cols-2 gap-2 sm:grid-cols-3">
                                                {questStatOptions.map(
                                                  (stat) => (
                                                    <label
                                                      key={stat.id}
                                                      className="flex items-center gap-2 text-xs text-gray-700"
                                                    >
                                                      <input
                                                        type="checkbox"
                                                        checked={editDraft.statTags.includes(
                                                          stat.id
                                                        )}
                                                        onChange={(e) => {
                                                          const current =
                                                            editDraft.statTags;
                                                          const next = e.target
                                                            .checked
                                                            ? [
                                                                ...current,
                                                                stat.id,
                                                              ]
                                                            : current.filter(
                                                                (tag) =>
                                                                  tag !==
                                                                  stat.id
                                                              );
                                                          handleEditChallengeDraftChange(
                                                            challenge.id,
                                                            { statTags: next }
                                                          );
                                                        }}
                                                      />
                                                      {stat.label}
                                                    </label>
                                                  )
                                                )}
                                              </div>
                                            </div>
                                            <div className="md:col-span-2">
                                              <label className="block text-xs font-medium text-gray-700">
                                                Proficiency
                                              </label>
                                              <input
                                                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                                value={editDraft.proficiency}
                                                list="proficiency-options"
                                                onChange={(e) => {
                                                  handleEditChallengeDraftChange(
                                                    challenge.id,
                                                    {
                                                      proficiency:
                                                        e.target.value,
                                                    }
                                                  );
                                                  handleProficiencyInputChange(
                                                    e.target.value
                                                  );
                                                }}
                                                placeholder="Drawing"
                                              />
                                            </div>
                                            <div className="md:col-span-4">
                                              <label className="block text-xs font-medium text-gray-700">
                                                Question
                                              </label>
                                              <textarea
                                                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                                rows={2}
                                                value={editDraft.question}
                                                onChange={(e) =>
                                                  handleEditChallengeDraftChange(
                                                    challenge.id,
                                                    { question: e.target.value }
                                                  )
                                                }
                                              />
                                            </div>
                                          </div>
                                          <button
                                            type="button"
                                            className="mt-3 bg-blue-600 text-white px-3 py-2 rounded-md"
                                            onClick={() =>
                                              handleUpdateChallenge(
                                                node,
                                                challenge
                                              )
                                            }
                                          >
                                            Save Changes
                                          </button>
                                        </div>
                                      )}
                                    </div>
                                  );
                                })}
                                {nodeObjectivePrompts.length === 0 &&
                                  !usesLinkedObjective && (
                                  <div className="text-sm text-gray-500">
                                    No objective prompts yet.
                                  </div>
                                )}
                                </div>
                              )}

                              {!usesLinkedObjective && (
                                <div className="bg-gray-50 border border-gray-200 rounded-md p-3">
                                <h5 className="font-semibold mb-2">
                                  Add Objective Prompt
                                </h5>
                                <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
                                  {node.pointOfInterestId && (
                                    <div className="md:col-span-4 rounded-md border border-amber-200 bg-amber-50 p-3">
                                      <div className="text-xs font-semibold text-amber-900">
                                        Location Archetype Prompt
                                      </div>
                                      <div className="mt-2 grid grid-cols-1 md:grid-cols-2 gap-3">
                                        <div>
                                          <label className="block text-xs font-medium text-gray-700">
                                            Archetype
                                          </label>
                                          <select
                                            className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                            value={
                                              (
                                                challengeDrafts[node.id] ??
                                                emptyChallengeForm
                                              ).locationArchetypeId
                                            }
                                            onChange={(e) =>
                                              handleChallengeDraftChange(
                                                node.id,
                                                {
                                                  locationArchetypeId:
                                                    e.target.value,
                                                  locationChallenge: '',
                                                  question: '',
                                                  submissionType: 'photo',
                                                  proficiency: '',
                                                }
                                              )
                                            }
                                          >
                                            <option value="">
                                              Select archetype
                                            </option>
                                            {locationArchetypes.map(
                                              (archetype) => (
                                                <option
                                                  key={archetype.id}
                                                  value={archetype.id}
                                                >
                                                  {archetype.name}
                                                </option>
                                              )
                                            )}
                                          </select>
                                        </div>
                                        <div>
                                          <label className="block text-xs font-medium text-gray-700">
                                            Prompt Template
                                          </label>
                                          {(() => {
                                            const selectedArchetype =
                                              locationArchetypes.find(
                                                (archetype) =>
                                                  archetype.id ===
                                                  (
                                                    challengeDrafts[node.id] ??
                                                    emptyChallengeForm
                                                  ).locationArchetypeId
                                              );
                                            const challenges =
                                              selectedArchetype?.challenges ??
                                              [];
                                            return (
                                              <select
                                                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                                value={
                                                  (
                                                    challengeDrafts[node.id] ??
                                                    emptyChallengeForm
                                                  ).locationChallenge
                                                }
                                                onChange={(e) =>
                                                  (() => {
                                                    const value =
                                                      e.target.value;
                                                    const index =
                                                      value === ''
                                                        ? NaN
                                                        : Number(value);
                                                    const selected =
                                                      Number.isFinite(index)
                                                        ? challenges[index]
                                                        : undefined;
                                                    handleChallengeDraftChange(
                                                      node.id,
                                                      {
                                                        locationChallenge:
                                                          value,
                                                        question:
                                                          selected?.question ??
                                                          '',
                                                        submissionType:
                                                          (selected?.submissionType ??
                                                            'photo') as QuestNodeSubmissionType,
                                                        proficiency:
                                                          selected?.proficiency ??
                                                          '',
                                                        difficulty:
                                                          selected?.difficulty ??
                                                          (
                                                            challengeDrafts[
                                                              node.id
                                                            ] ??
                                                            emptyChallengeForm
                                                          ).difficulty,
                                                      }
                                                    );
                                                  })()
                                                }
                                              >
                                                <option value="">
                                                  Select prompt template
                                                </option>
                                                {challenges.map(
                                                  (challenge, index) => (
                                                    <option
                                                      key={`${challenge.question}-${index}`}
                                                      value={index}
                                                    >
                                                      {challenge.question} ·{' '}
                                                      {challenge.submissionType.toUpperCase()}
                                                      {challenge.proficiency
                                                        ? ` · ${challenge.proficiency}`
                                                        : ''}
                                                    </option>
                                                  )
                                                )}
                                              </select>
                                            );
                                          })()}
                                        </div>
                                      </div>
                                      <p className="mt-2 text-xs text-amber-800">
                                        Selecting a prompt template will
                                        auto-fill the question field and input
                                        type.
                                      </p>
                                    </div>
                                  )}
                                  <div>
                                    <label className="block text-xs font-medium text-gray-700">
                                      Tier
                                    </label>
                                    <input
                                      type="number"
                                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                      value={
                                        (
                                          challengeDrafts[node.id] ??
                                          emptyChallengeForm
                                        ).tier
                                      }
                                      onChange={(e) =>
                                        handleChallengeDraftChange(node.id, {
                                          tier: Number(e.target.value),
                                        })
                                      }
                                    />
                                  </div>
                                  <div>
                                    <label className="block text-xs font-medium text-gray-700">
                                      Difficulty
                                    </label>
                                    <input
                                      type="number"
                                      min={0}
                                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                      value={
                                        (
                                          challengeDrafts[node.id] ??
                                          emptyChallengeForm
                                        ).difficulty
                                      }
                                      onChange={(e) =>
                                        handleChallengeDraftChange(node.id, {
                                          difficulty: Number(e.target.value),
                                        })
                                      }
                                    />
                                  </div>
                                  <div>
                                    <label className="block text-xs font-medium text-gray-700">
                                      Input Type
                                    </label>
                                    <select
                                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                      value={
                                        (
                                          challengeDrafts[node.id] ??
                                          emptyChallengeForm
                                        ).submissionType
                                      }
                                      onChange={(e) =>
                                        handleChallengeDraftChange(node.id, {
                                          submissionType: e.target
                                            .value as QuestNodeSubmissionType,
                                        })
                                      }
                                    >
                                      {questNodeSubmissionOptions.map(
                                        (option) => (
                                          <option
                                            key={option.value}
                                            value={option.value}
                                          >
                                            {option.label}
                                          </option>
                                        )
                                      )}
                                    </select>
                                  </div>
                                  <div>
                                    <label className="block text-xs font-medium text-gray-700">
                                      Reward
                                    </label>
                                    <input
                                      type="number"
                                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                      value={
                                        (
                                          challengeDrafts[node.id] ??
                                          emptyChallengeForm
                                        ).reward
                                      }
                                      onChange={(e) =>
                                        handleChallengeDraftChange(node.id, {
                                          reward: Number(e.target.value),
                                        })
                                      }
                                    />
                                  </div>
                                  <div>
                                    <label className="block text-xs font-medium text-gray-700">
                                      Inventory Item
                                    </label>
                                    <select
                                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                      value={
                                        (
                                          challengeDrafts[node.id] ??
                                          emptyChallengeForm
                                        ).inventoryItemId
                                      }
                                      onChange={(e) =>
                                        handleChallengeDraftChange(node.id, {
                                          inventoryItemId: e.target.value,
                                        })
                                      }
                                    >
                                      <option value="">None</option>
                                      {inventoryItems.map((item) => (
                                        <option key={item.id} value={item.id}>
                                          {item.name}
                                        </option>
                                      ))}
                                    </select>
                                  </div>
                                  <div className="md:col-span-4">
                                    <label className="block text-xs font-medium text-gray-700">
                                      Stat Tags
                                    </label>
                                    <div className="mt-2 grid grid-cols-2 gap-2 sm:grid-cols-3">
                                      {questStatOptions.map((stat) => (
                                        <label
                                          key={stat.id}
                                          className="flex items-center gap-2 text-xs text-gray-700"
                                        >
                                          <input
                                            type="checkbox"
                                            checked={(
                                              challengeDrafts[node.id] ??
                                              emptyChallengeForm
                                            ).statTags.includes(stat.id)}
                                            onChange={(e) => {
                                              const current = (
                                                challengeDrafts[node.id] ??
                                                emptyChallengeForm
                                              ).statTags;
                                              const next = e.target.checked
                                                ? [...current, stat.id]
                                                : current.filter(
                                                    (tag) => tag !== stat.id
                                                  );
                                              handleChallengeDraftChange(
                                                node.id,
                                                { statTags: next }
                                              );
                                            }}
                                          />
                                          {stat.label}
                                        </label>
                                      ))}
                                    </div>
                                  </div>
                                  <div className="md:col-span-2">
                                    <label className="block text-xs font-medium text-gray-700">
                                      Proficiency
                                    </label>
                                    <input
                                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                      value={
                                        (
                                          challengeDrafts[node.id] ??
                                          emptyChallengeForm
                                        ).proficiency
                                      }
                                      list="proficiency-options"
                                      onChange={(e) => {
                                        handleChallengeDraftChange(node.id, {
                                          proficiency: e.target.value,
                                        });
                                        handleProficiencyInputChange(
                                          e.target.value
                                        );
                                      }}
                                      placeholder="Drawing"
                                    />
                                  </div>
                                  <div className="md:col-span-4">
                                    <label className="block text-xs font-medium text-gray-700">
                                      Question
                                    </label>
                                    <textarea
                                      className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                      rows={2}
                                      value={
                                        (
                                          challengeDrafts[node.id] ??
                                          emptyChallengeForm
                                        ).question
                                      }
                                      onChange={(e) =>
                                        handleChallengeDraftChange(node.id, {
                                          question: e.target.value,
                                        })
                                      }
                                    />
                                  </div>
                                </div>
                                <button
                                  className="mt-3 bg-blue-600 text-white px-3 py-2 rounded-md"
                                  onClick={() => handleCreateChallenge(node)}
                                >
                                  Add Objective Prompt
                                </button>
                              </div>
                              )}
                            </div>
                          </div>
                        );
                      })}
                  </div>
                </div>
              </div>
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
