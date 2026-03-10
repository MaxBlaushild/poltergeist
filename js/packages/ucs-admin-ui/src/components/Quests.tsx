import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useAPI, useTagContext, useZoneContext } from '@poltergeist/contexts';
import { Candidate, Character, InventoryItem, LocationArchetype, PointOfInterest, Quest, QuestArchetype, QuestArchetypeChallenge, QuestArchetypeNode, QuestNode, QuestNodeChallenge, QuestNodeSubmissionType, Spell, Tag } from '@poltergeist/types';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import * as wellknown from 'wellknown';
import { useQuestArchetypes } from '../contexts/questArchetypes.tsx';
import { useCandidates } from '@poltergeist/hooks';
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

type ScenarioNodeOption = {
  id: string;
  zoneId: string;
  latitude: number;
  longitude: number;
  prompt: string;
};

type MonsterNodeOption = {
  id: string;
  zoneId: string;
  latitude: number;
  longitude: number;
  name: string;
  monsterCount?: number;
  members?: { slot: number; monster: { id: string; name: string } }[];
};

type ChallengeNodeOption = {
  id: string;
  zoneId: string;
  latitude: number;
  longitude: number;
  question: string;
};

type MonsterRecord = {
  id: string;
  zoneId: string;
  name: string;
  level?: number;
};

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
  rewardMode: 'random' as 'explicit' | 'random',
  randomRewardSize: 'small' as 'small' | 'medium' | 'large',
  rewardExperience: 0,
  gold: 0,
  itemRewards: [] as { inventoryItemId: string; quantity: number }[],
  spellRewards: [] as { spellId: string }[],
};

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

const questNodeSubmissionOptions: { value: QuestNodeSubmissionType; label: string }[] = [
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

const emptyQuickCreateMonsterEncounterForm = (): QuickCreateMonsterEncounterForm => ({
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

const questRecurrenceOptions = [
  { value: '', label: 'No Recurrence' },
  { value: 'daily', label: 'Daily' },
  { value: 'weekly', label: 'Weekly' },
  { value: 'monthly', label: 'Monthly' },
];

const buildChallengeFormFromChallenge = (
  challenge: QuestNodeChallenge,
  fallbackSubmissionType?: QuestNodeSubmissionType
) => ({
  ...emptyChallengeForm,
  tier: challenge.tier ?? 1,
  question: challenge.question ?? '',
  reward: challenge.reward ?? 0,
  inventoryItemId: challenge.inventoryItemId ? String(challenge.inventoryItemId) : '',
  statTags: challenge.statTags ?? [],
  difficulty: challenge.difficulty ?? 0,
  proficiency: challenge.proficiency ?? '',
  submissionType: (challenge.submissionType ?? fallbackSubmissionType ?? 'photo') as QuestNodeSubmissionType,
});

const emptyQuestReward = {
  inventoryItemId: '',
  quantity: 1,
};

const emptyQuestSpellReward = {
  spellId: '',
};

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

const normalizeText = (value: string) => value.trim().toLowerCase();

const normalizeAcceptanceDialogue = (lines: string[]) =>
  lines.map((line) => line.trim()).filter((line) => line.length > 0);

const normalizeStatTags = (tags: string[]) =>
  Array.from(new Set(tags.map((tag) => tag.trim().toLowerCase()).filter((tag) => tag.length > 0)));

const resolveChallengeSubmissionType = (challenge: QuestNodeChallenge, node?: QuestNode) =>
  (challenge.submissionType || node?.submissionType || 'photo') as QuestNodeSubmissionType;

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
  const [pointsOfInterest, setPointsOfInterest] = useState<PointOfInterest[]>([]);
  const [characters, setCharacters] = useState<Character[]>([]);
  const [inventoryItems, setInventoryItems] = useState<InventoryItem[]>([]);
  const [spells, setSpells] = useState<Spell[]>([]);
  const [scenarios, setScenarios] = useState<ScenarioNodeOption[]>([]);
  const [monsterRecords, setMonsterRecords] = useState<MonsterRecord[]>([]);
  const [monsterEncounters, setMonsterEncounters] = useState<MonsterNodeOption[]>([]);
  const [challenges, setChallenges] = useState<ChallengeNodeOption[]>([]);
  const [loading, setLoading] = useState(true);
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
  const [zoneDetailsById, setZoneDetailsById] = useState<Record<string, { boundary?: number[][]; boundaryCoords?: { latitude: number; longitude: number }[]; latitude?: number; longitude?: number }>>({});
  const [selectedQuestId, setSelectedQuestId] = useState<string>('');
  const [showCreateQuest, setShowCreateQuest] = useState(false);
  const [questForm, setQuestForm] = useState({ ...emptyQuestForm });
  const [nodeForm, setNodeForm] = useState({ ...emptyNodeForm });
  const [polygonDraftPoints, setPolygonDraftPoints] = useState<[number, number][]>([]);
  const [challengeDrafts, setChallengeDrafts] = useState<Record<string, typeof emptyChallengeForm>>({});
  const [challengeEdits, setChallengeEdits] = useState<Record<string, typeof emptyChallengeForm>>({});
  const [quickCreateOpen, setQuickCreateOpen] = useState<Record<'scenario' | 'monster' | 'challenge', boolean>>({
    scenario: false,
    monster: false,
    challenge: false,
  });
  const [quickCreateScenarioForm, setQuickCreateScenarioForm] = useState<QuickCreateScenarioForm>(
    emptyQuickCreateScenarioForm()
  );
  const [quickCreateChallengeForm, setQuickCreateChallengeForm] = useState<QuickCreateChallengeForm>(
    emptyQuickCreateChallengeForm()
  );
  const [quickCreateMonsterEncounterForm, setQuickCreateMonsterEncounterForm] =
    useState<QuickCreateMonsterEncounterForm>(emptyQuickCreateMonsterEncounterForm());
  const [quickCreateSubmitting, setQuickCreateSubmitting] = useState<
    null | 'scenario' | 'monster' | 'challenge'
  >(null);
  const [proficiencySearch, setProficiencySearch] = useState('');
  const [proficiencyOptions, setProficiencyOptions] = useState<string[]>([]);
  const [selectedPoiForModal, setSelectedPoiForModal] = useState<PointOfInterest | null>(null);
  const [characterLocationsOpen, setCharacterLocationsOpen] = useState(false);
  const [selectedCharacterLocations, setSelectedCharacterLocations] = useState<{ latitude: number; longitude: number }[]>([]);
  const [characterLocationsLoading, setCharacterLocationsLoading] = useState(false);
  const [showImportModal, setShowImportModal] = useState(false);
  const [importQuery, setImportQuery] = useState('');
  const [selectedCandidate, setSelectedCandidate] = useState<Candidate | null>(null);
  const [importZoneId, setImportZoneId] = useState('');
  const [importError, setImportError] = useState<string | null>(null);
  const [importJobs, setImportJobs] = useState<PointOfInterestImport[]>([]);
  const [importPolling, setImportPolling] = useState(false);
  const { candidates } = useCandidates(importQuery);
  const [importToasts, setImportToasts] = useState<string[]>([]);
  const [notifiedImportIds, setNotifiedImportIds] = useState<Set<string>>(new Set());
  const [polygonRefreshNonce, setPolygonRefreshNonce] = useState(0);
  const [deletingQuestId, setDeletingQuestId] = useState<string | null>(null);
  const [bulkDeletingQuests, setBulkDeletingQuests] = useState(false);
  const [selectedQuestIds, setSelectedQuestIds] = useState<Set<string>>(new Set());
  const [shufflingChallengeId, setShufflingChallengeId] = useState<string | null>(null);
  const [creatingArchetype, setCreatingArchetype] = useState(false);
  const questMapContainer = useRef<HTMLDivElement>(null);
  const questMap = useRef<mapboxgl.Map | null>(null);
  const [questMapLoaded, setQuestMapLoaded] = useState(false);
  const questNodeMarkers = useRef<mapboxgl.Marker[]>([]);
  const poiMarkers = useRef<mapboxgl.Marker[]>([]);
  const characterLocationMarkers = useRef<mapboxgl.Marker[]>([]);

  const selectedQuest = useMemo(() => quests.find((quest) => quest.id === selectedQuestId) ?? null, [quests, selectedQuestId]);
  const selectedQuestIdSet = useMemo(() => selectedQuestIds, [selectedQuestIds]);

  useEffect(() => {
    let isMounted = true;
    const load = async () => {
      const results = await Promise.allSettled([
        apiClient.get<Quest[]>('/sonar/quests'),
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
        questsResult,
        poiResult,
        charactersResult,
        inventoryResult,
        spellsResult,
        scenariosResult,
        monsterRecordsResult,
        monstersResult,
        challengesResult,
      ] = results;
      if (questsResult.status === 'fulfilled') {
        setQuests(questsResult.value);
      } else {
        console.error('Failed to load quests', questsResult.reason);
        setLoadError('Failed to load quests. Check console for details.');
      }

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
        setScenarios(Array.isArray(scenariosResult.value) ? scenariosResult.value : []);
      } else {
        console.error('Failed to load scenarios', scenariosResult.reason);
      }

      if (monsterRecordsResult.status === 'fulfilled') {
        setMonsterRecords(Array.isArray(monsterRecordsResult.value) ? monsterRecordsResult.value : []);
      } else {
        console.error('Failed to load monsters', monsterRecordsResult.reason);
      }

      if (monstersResult.status === 'fulfilled') {
        setMonsterEncounters(Array.isArray(monstersResult.value) ? monstersResult.value : []);
      } else {
        console.error('Failed to load monster encounters', monstersResult.reason);
      }

      if (challengesResult.status === 'fulfilled') {
        setChallenges(Array.isArray(challengesResult.value) ? challengesResult.value : []);
      } else {
        console.error('Failed to load challenges', challengesResult.reason);
      }

      setLoading(false);
    };

    load();
    return () => {
      isMounted = false;
    };
  }, []);

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
      const response = await apiClient.get<PointOfInterest[]>('/sonar/pointsOfInterest');
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
        const response = await apiClient.get<{ latitude: number; longitude: number }[]>(
          `/sonar/characters/${questForm.questGiverCharacterId}/locations`
        );
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
        const zone = await apiClient.get<{ boundary?: number[][]; boundaryCoords?: { latitude: number; longitude: number }[]; latitude?: number; longitude?: number }>(
          `/sonar/zones/${questForm.zoneId}`
        );
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
    if (!poiFiltersOpen || zonePoiMapLoaded || zonePoiMapLoading || zones.length === 0) return;
    let isMounted = true;
    const loadZonePoiMap = async () => {
      setZonePoiMapLoading(true);
      const results = await Promise.allSettled(
        zones.map((zone) => apiClient.get<PointOfInterest[]>(`/sonar/zones/${zone.id}/pointsOfInterest`))
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
  }, [apiClient, poiFiltersOpen, zonePoiMapLoaded, zonePoiMapLoading, zones]);

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
      .filter((node) => node.polygon || (node.polygonPoints && node.polygonPoints.length >= 3))
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
    const lineSource = map.getSource('quest-node-draft-line') as mapboxgl.GeoJSONSource | undefined;
    const polygonSource = map.getSource('quest-node-draft-polygon') as mapboxgl.GeoJSONSource | undefined;

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
      let polygonSource = map.getSource('quest-node-polygons') as mapboxgl.GeoJSONSource | undefined;
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
        polygonSource = map.getSource('quest-node-polygons') as mapboxgl.GeoJSONSource | undefined;
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
  }, [questPolygons, questMapLoaded, polygonRefreshNonce, selectedQuest?.nodes?.length]);

  useEffect(() => {
    if (!questMap.current || !questMapLoaded) return;
    const map = questMap.current;
    const handleClick = (event: mapboxgl.MapMouseEvent & mapboxgl.EventData) => {
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
    return characters.filter((character) => character.name?.toLowerCase().includes(term));
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
        return name.includes(term) || googleName.includes(term) || originalName.includes(term);
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
  }, [pointsOfInterest, poiSearch, poiZoneFilterId, zonePoiMap, poiTagFilterIds]);

  const filteredScenarios = useMemo(() => {
    let filtered = scenarios;
    if (questForm.zoneId) {
      filtered = filtered.filter((scenario) => scenario.zoneId === questForm.zoneId);
    }
    return filtered;
  }, [questForm.zoneId, scenarios]);

  const filteredMonsters = useMemo(() => {
    let filtered = monsterEncounters;
    if (questForm.zoneId) {
      filtered = filtered.filter((monster) => monster.zoneId === questForm.zoneId);
    }
    return filtered;
  }, [monsterEncounters, questForm.zoneId]);

  const filteredChallenges = useMemo(() => {
    let filtered = challenges;
    if (questForm.zoneId) {
      filtered = filtered.filter((challenge) => challenge.zoneId === questForm.zoneId);
    }
    return filtered;
  }, [challenges, questForm.zoneId]);

  const availableMonstersForQuickCreate = useMemo(() => {
    if (!questForm.zoneId) return monsterRecords;
    return monsterRecords.filter((monster) => monster.zoneId === questForm.zoneId);
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

  useEffect(() => {
    if (!selectedQuest?.nodes?.length) return;
    if (!locationArchetypes.length) return;

    setChallengeDrafts((prev) => {
      let changed = false;
      const next = { ...prev };
      selectedQuest.nodes?.forEach((node) => {
        if (!node.pointOfInterestId) return;
        const match = archetypeByPoiId[node.pointOfInterestId];
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
  }, [archetypeByPoiId, locationArchetypes.length, selectedQuest?.nodes]);

  const questNodePoints = useMemo(() => {
    if (!selectedQuest?.nodes?.length) return [];
    return selectedQuest.nodes
      .map((node) => {
        if (node.pointOfInterestId) {
          const poi = pointsOfInterest.find((item) => item.id === node.pointOfInterestId);
          if (!poi) return null;
          const lng = Number(poi.lng);
          const lat = Number(poi.lat);
          if (Number.isNaN(lng) || Number.isNaN(lat)) return null;
          return {
            id: node.id,
            name: poi.name,
            orderIndex: node.orderIndex,
            lng,
            lat,
            nodeType: 'poi' as QuestNodeType,
          };
        }
        if (node.scenarioId) {
          const scenario = scenarios.find((item) => item.id === node.scenarioId);
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
          const encounterId = node.monsterEncounterId ?? node.monsterId ?? '';
          const encounter = monsterEncounters.find((item) => item.id === encounterId);
          const fallbackEncounter = monsterEncounters.find((item) =>
            (item.members ?? []).some((member) => member.monster.id === node.monsterId)
          );
          const resolved = encounter ?? fallbackEncounter;
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
          const challenge = challenges.find((item) => item.id === node.challengeId);
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
      .filter((entry): entry is { id: string; name: string; orderIndex: number; lng: number; lat: number; nodeType: QuestNodeType } => Boolean(entry));
  }, [challenges, monsterEncounters, pointsOfInterest, scenarios, selectedQuest?.nodes]);

  const questNodePoiIdSet = useMemo(() => {
    if (!selectedQuest?.nodes?.length) return new Set<string>();
    return new Set(
      selectedQuest.nodes
        .map((node) => node.pointOfInterestId)
        .filter((id): id is string => Boolean(id))
    );
  }, [selectedQuest?.nodes]);

  const snapToZone = () => {
    if (!questMap.current || !questForm.zoneId) return;
    const zone =
      zoneDetailsById[questForm.zoneId] ||
      zones.find((z) => z.id === questForm.zoneId);
    if (!zone) return;
    const boundaryCoords =
      zone.boundaryCoords?.map((coord) => [coord.longitude, coord.latitude]) ??
      (Array.isArray(zone.boundary) ? zone.boundary : []);
    if (boundaryCoords.length > 0) {
      const [lng, lat] = boundaryCoords[0];
      questMap.current.setCenter([lng, lat]);
      questMap.current.setZoom(14);
      return;
    }
    const lng = typeof zone.longitude === 'number' ? zone.longitude : Number(zone.longitude);
    const lat = typeof zone.latitude === 'number' ? zone.latitude : Number(zone.latitude);
    if (!Number.isNaN(lng) && !Number.isNaN(lat)) {
      questMap.current.setCenter([lng, lat]);
      questMap.current.setZoom(12);
    }
  };

  useEffect(() => {
    if (!questMap.current || !questMapLoaded) return;

    questNodeMarkers.current.forEach((marker) => marker.remove());
    poiMarkers.current.forEach((marker) => marker.remove());
    characterLocationMarkers.current.forEach((marker) => marker.remove());
    questNodeMarkers.current = [];
    poiMarkers.current = [];
    characterLocationMarkers.current = [];

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
            : 'POI';
      const marker = new mapboxgl.Marker({ element: el })
        .setLngLat([point.lng, point.lat])
        .setPopup(new mapboxgl.Popup({ offset: 12 }).setText(`Node ${point.orderIndex} (${labelPrefix}): ${point.name}`))
        .addTo(questMap.current!);
      questNodeMarkers.current.push(marker);
      bounds.extend([point.lng, point.lat]);
      hasBounds = true;
    });

    const characterLocations = selectedCharacterLocations || [];
    const poiNearCharacter = (poi: PointOfInterest) => {
      const lng = Number(poi.lng);
      const lat = Number(poi.lat);
      if (Number.isNaN(lng) || Number.isNaN(lat)) return false;
      return characterLocations.some((loc) => {
        const dLng = Math.abs(lng - loc.longitude);
        const dLat = Math.abs(lat - loc.latitude);
        return dLng < 0.0001 && dLat < 0.0001;
      });
    };

    characterLocations.forEach((loc) => {
      if (Number.isNaN(loc.latitude) || Number.isNaN(loc.longitude)) return;
      const el = document.createElement('div');
      el.style.width = '14px';
      el.style.height = '14px';
      el.style.borderRadius = '9999px';
      el.style.background = '#10b981';
      el.style.border = '2px solid #065f46';
      const marker = new mapboxgl.Marker({ element: el })
        .setLngLat([loc.longitude, loc.latitude])
        .addTo(questMap.current!);
      characterLocationMarkers.current.push(marker);
      bounds.extend([loc.longitude, loc.latitude]);
      hasBounds = true;
    });

    filteredPointsOfInterest.forEach((poi) => {
      if (questNodePoiIdSet.has(poi.id)) return;
      const lng = Number(poi.lng);
      const lat = Number(poi.lat);
      if (Number.isNaN(lng) || Number.isNaN(lat)) return;
      const el = document.createElement('div');
      el.style.width = '10px';
      el.style.height = '10px';
      el.style.borderRadius = '9999px';
      const hasCharacter = poiNearCharacter(poi);
      el.style.background = hasCharacter ? '#a855f7' : '#3b82f6';
      el.style.border = hasCharacter ? '2px solid #6b21a8' : '2px solid #1e40af';
      el.style.cursor = 'pointer';
      el.addEventListener('click', () => {
        setSelectedPoiForModal(poi);
      });
      const marker = new mapboxgl.Marker({ element: el })
        .setLngLat([lng, lat])
        .addTo(questMap.current!);
      poiMarkers.current.push(marker);
      bounds.extend([lng, lat]);
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

    const zone =
      questForm.zoneId
        ? zoneDetailsById[questForm.zoneId] || zones.find((z) => z.id === questForm.zoneId)
        : null;
    const zoneBoundaryCoords =
      zone?.boundaryCoords?.map((coord) => [coord.longitude, coord.latitude]) ??
      (Array.isArray(zone?.boundary) ? zone?.boundary : []);

    if (zoneBoundaryCoords.length > 0) {
      console.log('Quest Map: fitting to zone boundary', zoneBoundaryCoords);
      const map = questMap.current;
      map.resize();
      zoneBoundaryCoords.forEach((coord) => {
        bounds.extend([coord[0], coord[1]]);
      });
      const fit = () => map.fitBounds(bounds, { padding: 40, maxZoom: 15 });
      requestAnimationFrame(fit);
      setTimeout(fit, 100);
      return;
    }

    if (questNodePoints.length > 0) {
      const firstNode = [...questNodePoints].sort((a, b) => a.orderIndex - b.orderIndex)[0];
      questMap.current.setCenter([firstNode.lng, firstNode.lat]);
      questMap.current.setZoom(14);
      return;
    }

    if (hasBounds) {
      questMap.current.fitBounds(bounds, { padding: 40, maxZoom: 15 });
    } else if (questForm.zoneId && zone) {
      if (zone) {
        const lng = typeof zone.longitude === 'number' ? zone.longitude : Number(zone.longitude);
        const lat = typeof zone.latitude === 'number' ? zone.latitude : Number(zone.latitude);
        if (!Number.isNaN(lng) && !Number.isNaN(lat)) {
          questMap.current.setCenter([lng, lat]);
          questMap.current.setZoom(12);
        }
      }
    }
  }, [filteredPointsOfInterest, questForm.zoneId, questMapLoaded, questNodePoiIdSet, questNodePoints, questPolygons, selectedCharacterLocations, zones, zoneDetailsById]);

  const updateQuestState = (questId: string, updater: (quest: Quest) => Quest) => {
    setQuests((prev) => prev.map((quest) => (quest.id === questId ? updater(quest) : quest)));
  };

  const refreshQuestById = async (questId: string) => {
    const latest = await apiClient.get<Quest>(`/sonar/quests/${questId}`);
    setQuests((prev) => {
      const found = prev.some((quest) => quest.id === questId);
      if (!found) return [latest, ...prev];
      return prev.map((quest) => (quest.id === questId ? latest : quest));
    });
    return latest;
  };

  const pollChallengeShuffleStatus = async (
    questId: string,
    challengeId: string,
    maxAttempts = 15
  ) => {
    for (let attempt = 0; attempt < maxAttempts; attempt += 1) {
      await new Promise((resolve) => window.setTimeout(resolve, 1200));
      try {
        const latest = await refreshQuestById(questId);
        const latestChallenge = (latest.nodes ?? [])
          .flatMap((node) => node.challenges ?? [])
          .find((challenge) => challenge.id === challengeId);
        const status = (latestChallenge?.challengeShuffleStatus || '').toLowerCase();
        if (status !== 'queued' && status !== 'in_progress') {
          break;
        }
      } catch (error) {
        console.error('Failed to poll challenge shuffle status', error);
        break;
      }
    }
  };

  const handleCreateQuest = async () => {
    try {
      const payload = {
        name: questForm.name,
        description: questForm.description,
        acceptanceDialogue: normalizeAcceptanceDialogue(questForm.acceptanceDialogue),
        zoneId: questForm.zoneId || null,
        questGiverCharacterId: questForm.questGiverCharacterId || null,
        questArchetypeId: questForm.questArchetypeId || null,
        recurrenceFrequency: questForm.recurrenceFrequency || '',
        rewardMode: questForm.rewardMode,
        randomRewardSize: questForm.randomRewardSize,
        rewardExperience: questForm.rewardMode === 'explicit' ? Number(questForm.rewardExperience) || 0 : 0,
        gold: questForm.rewardMode === 'explicit' ? Number(questForm.gold) || 0 : 0,
        itemRewards:
          questForm.rewardMode === 'explicit'
            ? questForm.itemRewards
                .map((reward) => ({
                  inventoryItemId: Number(reward.inventoryItemId) || 0,
                  quantity: Number(reward.quantity) || 0,
                }))
                .filter((reward) => reward.inventoryItemId > 0 && reward.quantity > 0)
            : [],
        spellRewards:
          questForm.rewardMode === 'explicit'
            ? questForm.spellRewards
                .map((reward) => ({ spellId: reward.spellId.trim() }))
                .filter((reward) => reward.spellId.length > 0)
            : [],
      };
      const created = await apiClient.post<Quest>('/sonar/quests', payload);
      setQuests((prev) => [created, ...prev]);
      setSelectedQuestId(created.id);
      setQuestForm({ ...emptyQuestForm });
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
        acceptanceDialogue: normalizeAcceptanceDialogue(questForm.acceptanceDialogue),
        zoneId: questForm.zoneId || null,
        questGiverCharacterId: questForm.questGiverCharacterId || null,
        questArchetypeId: questForm.questArchetypeId || null,
        recurrenceFrequency: questForm.recurrenceFrequency || '',
        rewardMode: questForm.rewardMode,
        randomRewardSize: questForm.randomRewardSize,
        rewardExperience: questForm.rewardMode === 'explicit' ? Number(questForm.rewardExperience) || 0 : 0,
        gold: questForm.rewardMode === 'explicit' ? Number(questForm.gold) || 0 : 0,
        itemRewards:
          questForm.rewardMode === 'explicit'
            ? questForm.itemRewards
                .map((reward) => ({
                  inventoryItemId: Number(reward.inventoryItemId) || 0,
                  quantity: Number(reward.quantity) || 0,
                }))
                .filter((reward) => reward.inventoryItemId > 0 && reward.quantity > 0)
            : [],
        spellRewards:
          questForm.rewardMode === 'explicit'
            ? questForm.spellRewards
                .map((reward) => ({ spellId: reward.spellId.trim() }))
                .filter((reward) => reward.spellId.length > 0)
            : [],
      };
      const updated = await apiClient.patch<Quest>(`/sonar/quests/${selectedQuest.id}`, payload);
      updateQuestState(selectedQuest.id, () => updated);
    } catch (error) {
      console.error('Failed to update quest', error);
      alert('Failed to update quest.');
    }
  };

  const handleCreateQuestArchetypeFromQuest = async () => {
    if (!selectedQuest || creatingArchetype) return;

    const nodes = (selectedQuest.nodes ?? []).slice().sort((a, b) => a.orderIndex - b.orderIndex);
    if (nodes.length === 0) {
      alert('Quest has no nodes to convert into an archetype.');
      return;
    }

    if (!locationArchetypes.length) {
      alert('Location archetypes are still loading. Please try again in a moment.');
      return;
    }

    const missing: string[] = [];
    const locationArchetypeIds: string[] = [];
    nodes.forEach((node) => {
      if (!node.pointOfInterestId) {
        if (node.scenarioId) {
          missing.push(`Node ${node.orderIndex}: scenario node`);
        } else if (node.monsterEncounterId || node.monsterId) {
          missing.push(`Node ${node.orderIndex}: monster node`);
        } else if (node.challengeId) {
          missing.push(`Node ${node.orderIndex}: challenge node`);
        } else {
          missing.push(`Node ${node.orderIndex}: polygon node`);
        }
        return;
      }
      const match = archetypeByPoiId[node.pointOfInterestId];
      if (!match) {
        const poiName = pointsOfInterest.find((poi) => poi.id === node.pointOfInterestId)?.name;
        missing.push(`Node ${node.orderIndex}: ${poiName ?? node.pointOfInterestId}`);
        return;
      }
      locationArchetypeIds.push(match.id);
    });

    if (missing.length > 0) {
      alert(`Cannot create quest archetype. Missing location archetypes for:\n${missing.join('\n')}`);
      return;
    }

    const name = `${selectedQuest.name} (Archetype)`;
    const itemRewards = (selectedQuest.itemRewards ?? [])
      .map((reward) => ({
        inventoryItemId: reward.inventoryItemId,
        quantity: reward.quantity ?? 0,
      }))
      .filter((reward) => reward.inventoryItemId && reward.quantity > 0);

    setCreatingArchetype(true);
    try {
      const rootNode = await apiClient.post<QuestArchetypeNode>('/sonar/questArchetypeNodes', {
        locationArchetypeID: locationArchetypeIds[0],
      });

      const archetype = await apiClient.post<QuestArchetype>('/sonar/questArchetypes', {
        name,
        rootId: rootNode.id,
        defaultGold: selectedQuest.gold ?? 0,
        itemRewards: itemRewards.length > 0 ? itemRewards : undefined,
      });

      let currentNodeId = rootNode.id;

      for (let index = 0; index < nodes.length; index += 1) {
        const node = nodes[index];
        const hasNext = index < nodes.length - 1;
        const nextLocationArchetypeId = hasNext ? locationArchetypeIds[index + 1] : null;
        const challenges = (node.challenges ?? []).slice().sort((a, b) => (a.tier ?? 0) - (b.tier ?? 0));

        if (hasNext && challenges.length === 0) {
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
          continue;
        }

        for (let challengeIndex = 0; challengeIndex < challenges.length; challengeIndex += 1) {
          const challenge = challenges[challengeIndex];
          const shouldUnlock = hasNext && challengeIndex === challenges.length - 1;
          const payload: {
            reward: number;
            inventoryItemId?: number;
            proficiency?: string;
            difficulty?: number;
            locationArchetypeID?: string;
          } = {
            reward: challenge.reward ?? 0,
          };

          if (challenge.inventoryItemId) {
            payload.inventoryItemId = challenge.inventoryItemId;
          }
          if (challenge.proficiency && challenge.proficiency.trim()) {
            payload.proficiency = challenge.proficiency.trim();
          }
          if (challenge.difficulty !== undefined && challenge.difficulty !== null) {
            payload.difficulty = challenge.difficulty;
          }
          if (shouldUnlock && nextLocationArchetypeId) {
            payload.locationArchetypeID = nextLocationArchetypeId;
          }

          const created = await apiClient.post<QuestArchetypeChallenge>(
            `/sonar/questArchetypes/${currentNodeId}/challenges`,
            payload
          );

          if (shouldUnlock) {
            if (!created.unlockedNodeId) {
              throw new Error('Failed to create next archetype node.');
            }
            currentNodeId = created.unlockedNodeId;
          }
        }
      }

      setQuestForm((prev) => ({ ...prev, questArchetypeId: archetype.id }));
      updateQuestState(selectedQuest.id, (quest) => ({ ...quest, questArchetypeId: archetype.id }));
      alert('Quest archetype created. Click Save Changes to link it to this quest.');
    } catch (error) {
      console.error('Failed to create quest archetype from quest', error);
      alert('Failed to create quest archetype from quest.');
    } finally {
      setCreatingArchetype(false);
    }
  };

  const handleDeleteQuest = async () => {
    if (!selectedQuest || bulkDeletingQuests) return;
    const confirmDelete = window.confirm(`Delete quest "${selectedQuest.name}"? This cannot be undone.`);
    if (!confirmDelete) return;

    setDeletingQuestId(selectedQuest.id);
    try {
      await apiClient.delete(`/sonar/quests/${selectedQuest.id}`);
      setQuests((prev) => prev.filter((quest) => quest.id !== selectedQuest.id));
      setSelectedQuestIds((prev) => {
        const next = new Set(prev);
        next.delete(selectedQuest.id);
        return next;
      });
      setSelectedQuestId('');
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
    const confirmDelete = window.confirm(`Delete quest "${quest.name}"? This cannot be undone.`);
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
        setSelectedQuestId('');
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
    if (bulkDeletingQuests || selectedQuestIds.size === 0 || deletingQuestId) return;

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
            preview ? ` (${preview}${moreCount > 0 ? ` +${moreCount} more` : ''})` : ''
          }? This cannot be undone.`;

    if (!window.confirm(confirmMessage)) return;

    setBulkDeletingQuests(true);
    try {
      const results = await Promise.allSettled(
        selectedIds.map((questId) => apiClient.delete(`/sonar/quests/${questId}`))
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
          setSelectedQuestId('');
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
    setSelectedQuestId(quest.id);
    setQuestForm({
      name: quest.name ?? '',
      description: quest.description ?? '',
      acceptanceDialogue: quest.acceptanceDialogue ?? [],
      imageUrl: quest.imageUrl ?? '',
      zoneId: quest.zoneId ?? '',
      questGiverCharacterId: quest.questGiverCharacterId ?? '',
      questArchetypeId: quest.questArchetypeId ?? '',
      recurrenceFrequency: quest.recurrenceFrequency ?? '',
      rewardMode: (quest.rewardMode as 'explicit' | 'random') ?? 'random',
      randomRewardSize: (quest.randomRewardSize as 'small' | 'medium' | 'large') ?? 'small',
      rewardExperience: quest.rewardExperience ?? 0,
      gold: quest.gold ?? 0,
      itemRewards: (quest.itemRewards ?? []).map((reward) => ({
        inventoryItemId: reward.inventoryItemId ? String(reward.inventoryItemId) : '',
        quantity: reward.quantity ?? 1,
      })),
      spellRewards: (quest.spellRewards ?? []).map((reward) => ({
        spellId: reward.spellId ?? '',
      })),
    });
  };

  const handleAddQuestReward = () => {
    setQuestForm((prev) => ({
      ...prev,
      itemRewards: [...prev.itemRewards, { ...emptyQuestReward }],
    }));
  };

  const handleUpdateQuestReward = (index: number, updates: Partial<{ inventoryItemId: string; quantity: number }>) => {
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
      itemRewards: prev.itemRewards.filter((_, rewardIndex) => rewardIndex !== index),
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
      spellRewards: prev.spellRewards.filter((_, rewardIndex) => rewardIndex !== index),
    }));
  };

  const handleCreateNode = async () => {
    if (!selectedQuest) return;
    try {
      if (nodeForm.nodeType === 'poi' || nodeForm.nodeType === 'polygon') {
        alert('Quest nodes now must be Scenario, Monster, or Challenge.');
        return;
      }
      const polygonPoints = nodeForm.nodeType === 'polygon' ? parsePolygonPoints(nodeForm.polygonPoints) : null;
      if (nodeForm.nodeType === 'polygon' && !polygonPoints) {
        alert('Please enter polygon points as JSON: [[lng,lat],[lng,lat],...]');
        return;
      }
      const payload = {
        orderIndex: Number(nodeForm.orderIndex) || 1,
        pointOfInterestId: nodeForm.nodeType === 'poi' ? nodeForm.pointOfInterestId || null : null,
        scenarioId: nodeForm.nodeType === 'scenario' ? nodeForm.scenarioId || null : null,
        monsterId: null,
        monsterEncounterId: nodeForm.nodeType === 'monster' ? nodeForm.monsterEncounterId || null : null,
        challengeId: nodeForm.nodeType === 'challenge' ? nodeForm.challengeId || null : null,
        polygonPoints: nodeForm.nodeType === 'polygon' ? polygonPoints : undefined,
        submissionType: nodeForm.submissionType,
      };
      const created = await apiClient.post<QuestNode>('/sonar/questNodes', {
        ...payload,
        questId: selectedQuest.id,
      });
      updateQuestState(selectedQuest.id, (quest) => ({
        ...quest,
        nodes: [...(quest.nodes ?? []), created].sort((a, b) => a.orderIndex - b.orderIndex),
      }));
      setNodeForm({ ...emptyNodeForm, orderIndex: (selectedQuest.nodes?.length ?? 0) + 2 });
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
    if (!quickCreateScenarioForm.prompt.trim() || !quickCreateScenarioForm.imageUrl.trim()) {
      alert('Scenario prompt and image URL are required.');
      return;
    }
    if (options.length === 0) {
      alert('Add at least one scenario option.');
      return;
    }

    setQuickCreateSubmitting('scenario');
    try {
      const created = await apiClient.post<ScenarioNodeOption & { attemptedByUser?: boolean }>(
        '/sonar/scenarios',
        {
          zoneId: questForm.zoneId,
          latitude,
          longitude,
          prompt: quickCreateScenarioForm.prompt.trim(),
          imageUrl: quickCreateScenarioForm.imageUrl.trim(),
          thumbnailUrl:
            quickCreateScenarioForm.thumbnailUrl.trim() || quickCreateScenarioForm.imageUrl.trim(),
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
        }
      );
      setScenarios((prev) => [created, ...prev]);
      setNodeForm((prev) => ({ ...prev, scenarioId: created.id }));
      setQuickCreateScenarioForm(emptyQuickCreateScenarioForm());
      setQuickCreateOpen((prev) => ({ ...prev, scenario: false }));
    } catch (error) {
      console.error('Failed to create scenario', error);
      alert(error instanceof Error ? error.message : 'Failed to create scenario.');
    } finally {
      setQuickCreateSubmitting(null);
    }
  };

  const handleCreateStandaloneChallenge = async () => {
    if (!questForm.zoneId) {
      alert('Select a zone for the quest before creating a challenge.');
      return;
    }
    const latitude = Number.parseFloat(quickCreateChallengeForm.latitude);
    const longitude = Number.parseFloat(quickCreateChallengeForm.longitude);
    if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
      alert('Challenge latitude and longitude are required.');
      return;
    }
    if (!quickCreateChallengeForm.question.trim()) {
      alert('Challenge question is required.');
      return;
    }

    setQuickCreateSubmitting('challenge');
    try {
      const created = await apiClient.post<ChallengeNodeOption>('/sonar/challenges', {
        zoneId: questForm.zoneId,
        latitude,
        longitude,
        question: quickCreateChallengeForm.question.trim(),
        description: quickCreateChallengeForm.description.trim(),
        imageUrl: quickCreateChallengeForm.imageUrl.trim(),
        thumbnailUrl:
          quickCreateChallengeForm.thumbnailUrl.trim() || quickCreateChallengeForm.imageUrl.trim(),
        rewardMode: 'explicit',
        randomRewardSize: 'small',
        rewardExperience: parseIntSafe(quickCreateChallengeForm.rewardExperience, 0),
        reward: parseIntSafe(quickCreateChallengeForm.rewardGold, 0),
        submissionType: quickCreateChallengeForm.submissionType,
        difficulty: parseIntSafe(quickCreateChallengeForm.difficulty, 0),
        scaleWithUserLevel: false,
        recurrenceFrequency: '',
        statTags: quickCreateChallengeForm.statTags,
        proficiency: quickCreateChallengeForm.proficiency.trim(),
      });
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
      alert(error instanceof Error ? error.message : 'Failed to create challenge.');
    } finally {
      setQuickCreateSubmitting(null);
    }
  };

  const handleCreateMonsterEncounter = async () => {
    if (!questForm.zoneId) {
      alert('Select a zone for the quest before creating a monster encounter.');
      return;
    }
    const latitude = Number.parseFloat(quickCreateMonsterEncounterForm.latitude);
    const longitude = Number.parseFloat(quickCreateMonsterEncounterForm.longitude);
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
      const created = await apiClient.post<MonsterNodeOption & { members?: { slot: number; monster: MonsterRecord }[] }>(
        '/sonar/monster-encounters',
        {
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
        }
      );
      setMonsterEncounters((prev) => [created, ...prev]);
      setNodeForm((prev) => ({ ...prev, monsterEncounterId: created.id }));
      setQuickCreateMonsterEncounterForm(emptyQuickCreateMonsterEncounterForm());
      setQuickCreateOpen((prev) => ({ ...prev, monster: false }));
    } catch (error) {
      console.error('Failed to create monster encounter', error);
      alert(error instanceof Error ? error.message : 'Failed to create monster encounter.');
    } finally {
      setQuickCreateSubmitting(null);
    }
  };

  const handleChallengeDraftChange = (nodeId: string, updates: Partial<typeof emptyChallengeForm>) => {
    setChallengeDrafts((prev) => ({
      ...prev,
      [nodeId]: { ...emptyChallengeForm, ...prev[nodeId], ...updates },
    }));
  };

  const handleEditChallengeDraftChange = (challengeId: string, updates: Partial<typeof emptyChallengeForm>) => {
    setChallengeEdits((prev) => ({
      ...prev,
      [challengeId]: { ...emptyChallengeForm, ...prev[challengeId], ...updates },
    }));
  };

  const handleProficiencyInputChange = (value: string) => {
    setProficiencySearch(value);
  };

  const handleStartEditChallenge = (node: QuestNode, challenge: QuestNodeChallenge) => {
    setChallengeEdits((prev) => ({
      ...prev,
      [challenge.id]: buildChallengeFormFromChallenge(challenge, node.submissionType),
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
    const draft = challengeDrafts[node.id] ?? emptyChallengeForm;
    if (!draft.question.trim()) {
      alert('Please enter a question for this challenge.');
      return;
    }
    try {
      const payload = {
        tier: Number(draft.tier) || 1,
        question: draft.question,
        reward: Number(draft.reward) || 0,
        inventoryItemId: draft.inventoryItemId ? Number(draft.inventoryItemId) : null,
        submissionType: draft.submissionType || node.submissionType || 'photo',
        statTags: normalizeStatTags(draft.statTags),
        difficulty: Number(draft.difficulty) || 0,
        proficiency: draft.proficiency.trim(),
      };
      const created = await apiClient.post<QuestNodeChallenge>(`/sonar/questNodes/${node.id}/challenges`, payload);
      updateQuestState(node.questId, (quest) => ({
        ...quest,
        nodes: (quest.nodes ?? []).map((n) =>
          n.id === node.id ? { ...n, challenges: [...(n.challenges ?? []), created] } : n
        ),
      }));
      setChallengeDrafts((prev) => ({ ...prev, [node.id]: { ...emptyChallengeForm } }));
    } catch (error) {
      console.error('Failed to create challenge', error);
      alert('Failed to create challenge.');
    }
  };

  const handleUpdateChallenge = async (node: QuestNode, challenge: QuestNodeChallenge) => {
    const draft = challengeEdits[challenge.id];
    if (!draft) return;
    if (!draft.question.trim()) {
      alert('Please enter a question for this challenge.');
      return;
    }
    try {
      const payload = {
        tier: Number(draft.tier) || 1,
        question: draft.question,
        reward: Number(draft.reward) || 0,
        inventoryItemId: draft.inventoryItemId ? Number(draft.inventoryItemId) : null,
        submissionType: draft.submissionType || node.submissionType || 'photo',
        statTags: normalizeStatTags(draft.statTags),
        difficulty: Number(draft.difficulty) || 0,
        proficiency: draft.proficiency.trim(),
      };
      const updated = await apiClient.patch<QuestNodeChallenge>(
        `/sonar/questNodes/${node.id}/challenges/${challenge.id}`,
        payload
      );
      updateQuestState(node.questId, (quest) => ({
        ...quest,
        nodes: (quest.nodes ?? []).map((n) =>
          n.id === node.id
            ? { ...n, challenges: (n.challenges ?? []).map((c) => (c.id === challenge.id ? updated : c)) }
            : n
        ),
      }));
      handleCancelEditChallenge(challenge.id);
    } catch (error) {
      console.error('Failed to update challenge', error);
      alert('Failed to update challenge.');
    }
  };

  const handleShuffleSavedChallenge = async (node: QuestNode, challenge: QuestNodeChallenge) => {
    if (!selectedQuest) return;
    if (challenge.challengeShuffleStatus === 'queued' || challenge.challengeShuffleStatus === 'in_progress') {
      return;
    }

    setShufflingChallengeId(challenge.id);
    try {
      const queued = await apiClient.post<QuestNodeChallenge>(
        `/sonar/questNodeChallenges/${challenge.id}/shuffle`,
        {}
      );

      updateQuestState(node.questId, (quest) => ({
        ...quest,
        nodes: (quest.nodes ?? []).map((n) =>
          n.id === node.id
            ? { ...n, challenges: (n.challenges ?? []).map((c) => (c.id === challenge.id ? queued : c)) }
            : n
        ),
      }));

      await pollChallengeShuffleStatus(selectedQuest.id, challenge.id);
    } catch (error) {
      console.error('Failed to shuffle challenge', error);
      alert('Failed to queue challenge shuffle.');
    } finally {
      setShufflingChallengeId(null);
    }
  };

  const handleDeleteNode = async (node: QuestNode) => {
    if (!selectedQuest) return;
    const confirmDelete = window.confirm(`Delete quest node ${node.orderIndex}? This cannot be undone.`);
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
      const importItem = await apiClient.post<PointOfInterestImport>('/sonar/pointOfInterest/import', {
        placeID: selectedCandidate.place_id,
        zoneID: zoneId,
      });
      setImportJobs((prev) => [importItem, ...prev]);
      setImportPolling(true);
    } catch (error) {
      console.error('Error importing point of interest:', error);
      setImportError('Failed to import point of interest.');
    }
  };

  const handleRetryImport = async (placeId: string, zoneId: string) => {
    try {
      const importItem = await apiClient.post<PointOfInterestImport>('/sonar/pointOfInterest/import', {
        placeID: placeId,
        zoneID: zoneId,
      });
      setImportJobs((prev) => [importItem, ...prev]);
      setImportPolling(true);
    } catch (error) {
      console.error('Failed to retry import', error);
      setImportError('Failed to retry import.');
    }
  };

  const fetchImportJobs = async (zoneId?: string) => {
    try {
      const url = zoneId ? `/sonar/pointOfInterest/imports?zoneId=${zoneId}` : '/sonar/pointOfInterest/imports';
      const response = await apiClient.get<PointOfInterestImport[]>(url);
      setImportJobs(response);
      const hasPending = response.some((item) => item.status === 'queued' || item.status === 'in_progress');
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
    const completed = importJobs.filter((job) => job.status === 'completed' && job.pointOfInterestId);
    if (completed.length === 0) return;

    setNotifiedImportIds((prev) => {
      const next = new Set(prev);
      let hasNew = false;
      completed.forEach((job) => {
        if (!next.has(job.id)) {
          next.add(job.id);
          hasNew = true;
          setImportToasts((existing) => [`Import complete: ${job.placeId}`, ...existing].slice(0, 3));
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
      const response = await apiClient.get<{ latitude: number; longitude: number }[]>(
        `/sonar/characters/${questForm.questGiverCharacterId}/locations`
      );
      setSelectedCharacterLocations(response);
    } catch (error) {
      console.error('Failed to load character locations', error);
    } finally {
      setCharacterLocationsLoading(false);
    }
  };

  const handleAddCharacterLocation = () => {
    setSelectedCharacterLocations((prev) => [...prev, { latitude: 0, longitude: 0 }]);
  };

  const handleUpdateCharacterLocation = (index: number, key: 'latitude' | 'longitude', value: number) => {
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
      await apiClient.put(`/sonar/characters/${questForm.questGiverCharacterId}/locations`, {
        locations: selectedCharacterLocations,
      });
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
            Build quests, manage nodes, and tune challenge inputs with the same archetype-focused UI language.
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
          <h2 className="qa-card-title">Create Quest</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Name</label>
              <input
                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                value={questForm.name}
                onChange={(e) => setQuestForm((prev) => ({ ...prev, name: e.target.value }))}
              />
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700">Description</label>
              <textarea
                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                rows={3}
                value={questForm.description}
                onChange={(e) => setQuestForm((prev) => ({ ...prev, description: e.target.value }))}
              />
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700">Quest Acceptance Dialogue</label>
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
              <p className="mt-1 text-xs text-gray-500">Each line becomes a separate dialogue line in the quest acceptance prompt.</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Zone</label>
              <input
                className="mt-1 mb-2 block w-full border border-gray-300 rounded-md p-2"
                placeholder="Filter zones..."
                value={zoneSearch}
                onChange={(e) => setZoneSearch(e.target.value)}
              />
              <select
                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                value={questForm.zoneId}
                onChange={(e) => setQuestForm((prev) => ({ ...prev, zoneId: e.target.value }))}
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
              <label className="block text-sm font-medium text-gray-700">Quest Giver Character</label>
              <input
                className="mt-1 mb-2 block w-full border border-gray-300 rounded-md p-2"
                placeholder="Filter characters..."
                value={characterSearch}
                onChange={(e) => setCharacterSearch(e.target.value)}
              />
              <select
                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                value={questForm.questGiverCharacterId}
                onChange={(e) => setQuestForm((prev) => ({ ...prev, questGiverCharacterId: e.target.value }))}
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
              <label className="block text-sm font-medium text-gray-700">Quest Archetype ID (optional)</label>
              <input
                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                value={questForm.questArchetypeId}
                onChange={(e) => setQuestForm((prev) => ({ ...prev, questArchetypeId: e.target.value }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Recurrence</label>
              <select
                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                value={questForm.recurrenceFrequency}
                onChange={(e) => setQuestForm((prev) => ({ ...prev, recurrenceFrequency: e.target.value }))}
              >
                {questRecurrenceOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Reward Mode</label>
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
              <label className="block text-sm font-medium text-gray-700">Random Reward Size</label>
              <select
                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                value={questForm.randomRewardSize}
                disabled={questForm.rewardMode !== 'random'}
                onChange={(e) =>
                  setQuestForm((prev) => ({
                    ...prev,
                    randomRewardSize: e.target.value as 'small' | 'medium' | 'large',
                  }))
                }
              >
                <option value="small">Small</option>
                <option value="medium">Medium</option>
                <option value="large">Large</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Experience Reward</label>
              <input
                type="number"
                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                value={questForm.rewardExperience}
                disabled={questForm.rewardMode !== 'explicit'}
                onChange={(e) =>
                  setQuestForm((prev) => ({ ...prev, rewardExperience: Number(e.target.value) }))
                }
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">Gold Reward</label>
              <input
                type="number"
                className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                value={questForm.gold}
                disabled={questForm.rewardMode !== 'explicit'}
                onChange={(e) => setQuestForm((prev) => ({ ...prev, gold: Number(e.target.value) }))}
              />
            </div>
            {questForm.rewardMode === 'random' && (
              <div className="md:col-span-2 text-xs text-gray-500">
                Random rewards ignore explicit gold/item/spell fields.
              </div>
            )}
            <div className="md:col-span-2">
              <div className="flex items-center justify-between">
                <label className="block text-sm font-medium text-gray-700">Item Rewards</label>
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
                <div className="mt-2 text-xs text-gray-500">No item rewards yet.</div>
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
                        onChange={(e) => handleUpdateQuestReward(index, { inventoryItemId: e.target.value })}
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
                        onChange={(e) => handleUpdateQuestReward(index, { quantity: Number(e.target.value) })}
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
                <label className="block text-sm font-medium text-gray-700">Spell Rewards</label>
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
                <div className="mt-2 text-xs text-gray-500">No spell rewards yet.</div>
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
                          handleUpdateQuestSpellReward(index, { spellId: e.target.value })
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

      <div className="grid grid-cols-1 lg:grid-cols-[340px_1fr] gap-6">
        <div className="qa-card">
          <h2 className="qa-card-title">Quest List</h2>
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
              {allFilteredQuestsSelected ? 'Unselect Visible' : 'Select Visible'}
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
              disabled={selectedQuestIds.size === 0 || bulkDeletingQuests || deletingQuestId !== null}
            >
              {bulkDeletingQuests
                ? `Deleting ${selectedQuestIds.size}...`
                : `Delete Selected (${selectedQuestIds.size})`}
            </button>
          </div>
          <div className="space-y-2 max-h-[520px] overflow-y-auto">
            {filteredQuests.map((quest) => (
              <div
                key={quest.id}
                className={`flex items-center justify-between gap-2 p-3 rounded-md border ${selectedQuestId === quest.id ? 'border-blue-500 bg-blue-50' : 'border-gray-200'}`}
              >
                <input
                  type="checkbox"
                  className="h-4 w-4"
                  checked={selectedQuestIdSet.has(quest.id)}
                  disabled={bulkDeletingQuests}
                  onChange={() => toggleQuestSelection(quest.id)}
                />
                <button
                  className="flex-1 text-left"
                  onClick={() => handleSelectQuest(quest)}
                >
                  <div className="font-semibold">{quest.name}</div>
                  <div className="text-xs text-gray-500">Nodes: {quest.nodes?.length ?? 0}</div>
                </button>
                <button
                  className="qa-btn qa-btn-danger"
                  onClick={() => handleDeleteQuestById(quest)}
                  disabled={deletingQuestId === quest.id || bulkDeletingQuests}
                >
                  {deletingQuestId === quest.id ? 'Deleting...' : 'Delete'}
                </button>
              </div>
            ))}
          </div>
        </div>

        <div className="qa-card">
          {!selectedQuest ? (
            <div className="text-gray-500">Select a quest to edit details and add nodes.</div>
          ) : (
            <>
              <div className="flex items-center justify-between mb-4">
                <h2 className="qa-card-title">Quest Details</h2>
                <div className="flex items-center gap-2">
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
                    {creatingArchetype ? 'Creating Archetype...' : 'Create Archetype'}
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
                    disabled={deletingQuestId === selectedQuest.id || bulkDeletingQuests}
                  >
                    {deletingQuestId === selectedQuest.id ? 'Deleting...' : 'Delete Quest'}
                  </button>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700">Name</label>
                  <input
                    className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                    value={questForm.name}
                    onChange={(e) => setQuestForm((prev) => ({ ...prev, name: e.target.value }))}
                  />
                </div>
                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700">Description</label>
                  <textarea
                    className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                    rows={3}
                    value={questForm.description}
                    onChange={(e) => setQuestForm((prev) => ({ ...prev, description: e.target.value }))}
                  />
                </div>
                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700">Quest Acceptance Dialogue</label>
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
                  <p className="mt-1 text-xs text-gray-500">Each line becomes a separate dialogue line in the quest acceptance prompt.</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Zone</label>
                  <select
                    className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                    value={questForm.zoneId}
                    onChange={(e) => setQuestForm((prev) => ({ ...prev, zoneId: e.target.value }))}
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
                  <label className="block text-sm font-medium text-gray-700">Quest Giver Character</label>
                <select
                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                  value={questForm.questGiverCharacterId}
                  onChange={(e) => setQuestForm((prev) => ({ ...prev, questGiverCharacterId: e.target.value }))}
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
                  <label className="block text-sm font-medium text-gray-700">Quest Archetype ID</label>
                  <input
                    className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                    value={questForm.questArchetypeId}
                    onChange={(e) => setQuestForm((prev) => ({ ...prev, questArchetypeId: e.target.value }))}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Recurrence</label>
                  <select
                    className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                    value={questForm.recurrenceFrequency}
                    onChange={(e) => setQuestForm((prev) => ({ ...prev, recurrenceFrequency: e.target.value }))}
                  >
                    {questRecurrenceOptions.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">Gold Reward</label>
                  <input
                    type="number"
                    className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                    value={questForm.gold}
                    onChange={(e) => setQuestForm((prev) => ({ ...prev, gold: Number(e.target.value) }))}
                  />
                </div>
                <div className="md:col-span-2">
                  <div className="flex items-center justify-between">
                    <label className="block text-sm font-medium text-gray-700">Item Rewards</label>
                    <button
                      type="button"
                      className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                      onClick={handleAddQuestReward}
                    >
                      Add Item Reward
                    </button>
                  </div>
                  {questForm.itemRewards.length === 0 ? (
                    <div className="mt-2 text-xs text-gray-500">No item rewards yet.</div>
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
                            onChange={(e) => handleUpdateQuestReward(index, { inventoryItemId: e.target.value })}
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
                            onChange={(e) => handleUpdateQuestReward(index, { quantity: Number(e.target.value) })}
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
                    <label className="block text-sm font-medium text-gray-700">Spell Rewards</label>
                    <button
                      type="button"
                      className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                      onClick={handleAddQuestSpellReward}
                    >
                      Add Spell Reward
                    </button>
                  </div>
                  {questForm.spellRewards.length === 0 ? (
                    <div className="mt-2 text-xs text-gray-500">No spell rewards yet.</div>
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
                              handleUpdateQuestSpellReward(index, { spellId: e.target.value })
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

              <div className="border-t pt-4">
                <h3 className="text-lg font-semibold mb-3">Quest Nodes</h3>
                <div className="bg-gray-50 border border-gray-200 rounded-md p-4 mb-4">
                  <h4 className="font-semibold mb-3">Add Node</h4>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700">Order Index</label>
                      <input
                        type="number"
                        className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                        value={nodeForm.orderIndex}
                        onChange={(e) => setNodeForm((prev) => ({ ...prev, orderIndex: Number(e.target.value) }))}
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700">Node Type</label>
                      <select
                        className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                        value={nodeForm.nodeType}
                        onChange={(e) => {
                          const nextNodeType = e.target.value as QuestNodeType;
                          setNodeForm((prev) => ({
                            ...prev,
                            nodeType: nextNodeType,
                            pointOfInterestId: nextNodeType === 'poi' ? prev.pointOfInterestId : '',
                            scenarioId: nextNodeType === 'scenario' ? prev.scenarioId : '',
                            monsterEncounterId: nextNodeType === 'monster' ? prev.monsterEncounterId : '',
                            challengeId: nextNodeType === 'challenge' ? prev.challengeId : '',
                            polygonPoints: nextNodeType === 'polygon' ? prev.polygonPoints : '',
                          }));
                        }}
                      >
                        <option value="scenario">Scenario</option>
                        <option value="monster">Monster</option>
                        <option value="challenge">Challenge</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700">Submission Type</label>
                      <select
                        className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                        value={nodeForm.submissionType}
                        onChange={(e) =>
                          setNodeForm((prev) => ({
                            ...prev,
                            submissionType: e.target.value as QuestNodeSubmissionType,
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
                        <label className="block text-sm font-medium text-gray-700">Point of Interest</label>
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
                              <label className="block text-xs font-medium text-gray-700">Zone</label>
                              <select
                                className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                value={poiZoneFilterId}
                                onChange={(e) => setPoiZoneFilterId(e.target.value)}
                              >
                                <option value="">All zones</option>
                                {zones.map((zone) => (
                                  <option key={zone.id} value={zone.id}>
                                    {zone.name}
                                  </option>
                                ))}
                              </select>
                              {poiZoneFilterId && zonePoiMapLoading && (
                                <p className="mt-1 text-xs text-gray-500">Loading zone points of interest…</p>
                              )}
                            </div>
                            <div>
                              <label className="block text-xs font-medium text-gray-700">Tags</label>
                              <input
                                className="mt-1 mb-2 block w-full rounded-md border border-gray-300 p-2"
                                placeholder="Search tags..."
                                value={poiTagSearch}
                                onChange={(e) => setPoiTagSearch(e.target.value)}
                              />
                              <div className="max-h-40 overflow-y-auto rounded-md border border-gray-200 bg-white p-2">
                                {filteredTags.length === 0 && (
                                  <div className="text-xs text-gray-500">No tags found.</div>
                                )}
                                {filteredTags.map((tag) => {
                                  const isSelected = poiTagFilterIds.includes(tag.id);
                                  return (
                                    <label key={tag.id} className="flex items-center gap-2 text-xs text-gray-700">
                                      <input
                                        type="checkbox"
                                        checked={isSelected}
                                        onChange={(e) => {
                                          if (e.target.checked) {
                                            setPoiTagFilterIds((prev) => [...prev, tag.id]);
                                          } else {
                                            setPoiTagFilterIds((prev) => prev.filter((id) => id !== tag.id));
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
                                Showing {filteredPointsOfInterest.length} / {pointsOfInterest.length}
                              </span>
                            </div>
                          </div>
                        )}
                        <select
                          className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                          value={nodeForm.pointOfInterestId}
                          onChange={(e) => setNodeForm((prev) => ({ ...prev, pointOfInterestId: e.target.value }))}
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
                          <label className="block text-sm font-medium text-gray-700">Scenario</label>
                          <button
                            type="button"
                            className="rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                            onClick={() => toggleQuickCreate('scenario')}
                          >
                            {quickCreateOpen.scenario ? 'Hide Quick Create' : 'Create New Scenario'}
                          </button>
                        </div>
                        <select
                          className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                          value={nodeForm.scenarioId}
                          onChange={(e) => setNodeForm((prev) => ({ ...prev, scenarioId: e.target.value }))}
                        >
                          <option value="">Select a scenario</option>
                          {filteredScenarios.map((scenario) => (
                            <option key={scenario.id} value={scenario.id}>
                              {summarizeScenarioPrompt(scenario.prompt)}
                            </option>
                          ))}
                        </select>
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
                                    setQuickCreateScenarioForm((prev) => ({ ...prev, prompt: e.target.value }))
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
                                      setQuickCreateScenarioForm((prev) => ({ ...prev, imageUrl: e.target.value }))
                                    }
                                  />
                                </label>
                                <label className="text-sm">
                                  Thumbnail URL
                                  <input
                                    className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                    value={quickCreateScenarioForm.thumbnailUrl}
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
                                    setQuickCreateScenarioForm((prev) => ({ ...prev, latitude: e.target.value }))
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
                                    setQuickCreateScenarioForm((prev) => ({ ...prev, longitude: e.target.value }))
                                  }
                                />
                              </label>
                            </div>

                            <div className="rounded-md border border-gray-200 p-3">
                              <div className="mb-2 flex items-center justify-between">
                                <div className="text-sm font-medium text-gray-700">Options</div>
                                <button
                                  type="button"
                                  className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                                  onClick={handleAddQuickScenarioOption}
                                >
                                  Add Option
                                </button>
                              </div>
                              <div className="space-y-3">
                                {quickCreateScenarioForm.options.map((option, index) => (
                                  <div key={`quick-scenario-option-${index}`} className="rounded-md border border-gray-200 p-3">
                                    <div className="mb-2 flex items-center justify-between">
                                      <div className="text-xs font-semibold uppercase tracking-wide text-gray-500">
                                        Option {index + 1}
                                      </div>
                                      {quickCreateScenarioForm.options.length > 1 && (
                                        <button
                                          type="button"
                                          className="rounded-md border border-red-200 px-2 py-1 text-xs text-red-600 hover:bg-red-50"
                                          onClick={() => handleRemoveQuickScenarioOption(index)}
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
                                            handleUpdateQuickScenarioOption(index, { optionText: e.target.value })
                                          }
                                        />
                                      </label>
                                      <label className="text-sm">
                                        Stat
                                        <select
                                          className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                          value={option.statTag}
                                          onChange={(e) =>
                                            handleUpdateQuickScenarioOption(index, { statTag: e.target.value })
                                          }
                                        >
                                          {questStatOptions.map((stat) => (
                                            <option key={stat.id} value={stat.id}>
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
                                            handleUpdateQuickScenarioOption(index, { difficulty: e.target.value })
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
                                            handleUpdateQuickScenarioOption(index, { proficiencies: e.target.value })
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
                                            handleUpdateQuickScenarioOption(index, { successText: e.target.value })
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
                                            handleUpdateQuickScenarioOption(index, { failureText: e.target.value })
                                          }
                                        />
                                      </label>
                                    </div>
                                  </div>
                                ))}
                              </div>
                            </div>

                            <button
                              type="button"
                              className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:bg-blue-300"
                              onClick={handleCreateStandaloneScenario}
                              disabled={quickCreateSubmitting === 'scenario'}
                            >
                              {quickCreateSubmitting === 'scenario' ? 'Creating Scenario...' : 'Create and Select Scenario'}
                            </button>
                          </div>
                        )}
                      </div>
                    ) : nodeForm.nodeType === 'monster' ? (
                      <div className="md:col-span-2">
                        <div className="flex items-center justify-between gap-3">
                          <label className="block text-sm font-medium text-gray-700">Monster Encounter</label>
                          <button
                            type="button"
                            className="rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                            onClick={() => toggleQuickCreate('monster')}
                          >
                            {quickCreateOpen.monster ? 'Hide Quick Create' : 'Create New Encounter'}
                          </button>
                        </div>
                        <select
                          className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                          value={nodeForm.monsterEncounterId}
                          onChange={(e) =>
                            setNodeForm((prev) => ({ ...prev, monsterEncounterId: e.target.value }))
                          }
                        >
                          <option value="">Select a monster encounter</option>
                          {filteredMonsters.map((monster) => (
                            <option key={monster.id} value={monster.id}>
                              {monster.name}
                              {monster.monsterCount && monster.monsterCount > 1
                                ? ` (${monster.monsterCount} monsters)`
                                : ''}
                            </option>
                          ))}
                        </select>
                        {quickCreateOpen.monster && (
                          <div className="mt-3 rounded-md border border-gray-200 bg-white p-4 space-y-3">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                              <label className="text-sm">
                                Name
                                <input
                                  className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                  value={quickCreateMonsterEncounterForm.name}
                                  onChange={(e) =>
                                    setQuickCreateMonsterEncounterForm((prev) => ({ ...prev, name: e.target.value }))
                                  }
                                />
                              </label>
                              <label className="text-sm">
                                Description
                                <input
                                  className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                  value={quickCreateMonsterEncounterForm.description}
                                  onChange={(e) =>
                                    setQuickCreateMonsterEncounterForm((prev) => ({
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
                                  value={quickCreateMonsterEncounterForm.imageUrl}
                                  onChange={(e) =>
                                    setQuickCreateMonsterEncounterForm((prev) => ({
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
                                  value={quickCreateMonsterEncounterForm.thumbnailUrl}
                                  onChange={(e) =>
                                    setQuickCreateMonsterEncounterForm((prev) => ({
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
                                  value={quickCreateMonsterEncounterForm.latitude}
                                  onChange={(e) =>
                                    setQuickCreateMonsterEncounterForm((prev) => ({
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
                                  value={quickCreateMonsterEncounterForm.longitude}
                                  onChange={(e) =>
                                    setQuickCreateMonsterEncounterForm((prev) => ({
                                      ...prev,
                                      longitude: e.target.value,
                                    }))
                                  }
                                />
                              </label>
                            </div>
                            <label className="flex items-center gap-2 text-sm text-gray-700">
                              <input
                                type="checkbox"
                                checked={quickCreateMonsterEncounterForm.scaleWithUserLevel}
                                onChange={(e) =>
                                  setQuickCreateMonsterEncounterForm((prev) => ({
                                    ...prev,
                                    scaleWithUserLevel: e.target.checked,
                                  }))
                                }
                              />
                              Scale encounter with user level
                            </label>
                            <div>
                              <div className="mb-2 text-sm font-medium text-gray-700">Monsters</div>
                              <div className="max-h-48 space-y-2 overflow-y-auto rounded-md border border-gray-200 p-3">
                                {availableMonstersForQuickCreate.length === 0 ? (
                                  <div className="text-sm text-gray-500">No monsters available in this quest zone.</div>
                                ) : (
                                  availableMonstersForQuickCreate.map((monster) => {
                                    const checked = quickCreateMonsterEncounterForm.monsterIds.includes(monster.id);
                                    return (
                                      <label key={monster.id} className="flex items-center gap-2 text-sm text-gray-700">
                                        <input
                                          type="checkbox"
                                          checked={checked}
                                          onChange={(e) => {
                                            setQuickCreateMonsterEncounterForm((prev) => ({
                                              ...prev,
                                              monsterIds: e.target.checked
                                                ? [...prev.monsterIds, monster.id]
                                                : prev.monsterIds.filter((id) => id !== monster.id),
                                            }));
                                          }}
                                        />
                                        <span>{monster.name}</span>
                                        {typeof monster.level === 'number' && (
                                          <span className="text-xs text-gray-500">Lvl {monster.level}</span>
                                        )}
                                      </label>
                                    );
                                  })
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
                          <label className="block text-sm font-medium text-gray-700">Challenge</label>
                          <button
                            type="button"
                            className="rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                            onClick={() => toggleQuickCreate('challenge')}
                          >
                            {quickCreateOpen.challenge ? 'Hide Quick Create' : 'Create New Challenge'}
                          </button>
                        </div>
                        <select
                          className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                          value={nodeForm.challengeId}
                          onChange={(e) => setNodeForm((prev) => ({ ...prev, challengeId: e.target.value }))}
                        >
                          <option value="">Select a challenge</option>
                          {filteredChallenges.map((challenge) => (
                            <option key={challenge.id} value={challenge.id}>
                              {challenge.question}
                            </option>
                          ))}
                        </select>
                        {quickCreateOpen.challenge && (
                          <div className="mt-3 rounded-md border border-gray-200 bg-white p-4 space-y-3">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                              <label className="text-sm md:col-span-2">
                                Question
                                <textarea
                                  className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                  rows={2}
                                  value={quickCreateChallengeForm.question}
                                  onChange={(e) =>
                                    setQuickCreateChallengeForm((prev) => ({ ...prev, question: e.target.value }))
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
                                    setQuickCreateChallengeForm((prev) => ({ ...prev, imageUrl: e.target.value }))
                                  }
                                />
                              </label>
                              <label className="text-sm">
                                Thumbnail URL
                                <input
                                  className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                  value={quickCreateChallengeForm.thumbnailUrl}
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
                                    setQuickCreateChallengeForm((prev) => ({ ...prev, latitude: e.target.value }))
                                  }
                                />
                              </label>
                              <label className="text-sm">
                                Longitude
                                <input
                                  type="number"
                                  className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                  value={quickCreateChallengeForm.longitude}
                                  onChange={(e) =>
                                    setQuickCreateChallengeForm((prev) => ({ ...prev, longitude: e.target.value }))
                                  }
                                />
                              </label>
                              <label className="text-sm">
                                Submission Type
                                <select
                                  className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                  value={quickCreateChallengeForm.submissionType}
                                  onChange={(e) =>
                                    setQuickCreateChallengeForm((prev) => ({
                                      ...prev,
                                      submissionType: e.target.value as QuestNodeSubmissionType,
                                    }))
                                  }
                                >
                                  {questNodeSubmissionOptions.map((option) => (
                                    <option key={option.value} value={option.value}>
                                      {option.label}
                                    </option>
                                  ))}
                                </select>
                              </label>
                              <label className="text-sm">
                                Difficulty
                                <input
                                  type="number"
                                  className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                  value={quickCreateChallengeForm.difficulty}
                                  onChange={(e) =>
                                    setQuickCreateChallengeForm((prev) => ({ ...prev, difficulty: e.target.value }))
                                  }
                                />
                              </label>
                              <label className="text-sm">
                                Reward XP
                                <input
                                  type="number"
                                  className="mt-1 block w-full rounded-md border border-gray-300 p-2"
                                  value={quickCreateChallengeForm.rewardExperience}
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
                              <div className="mb-2 text-sm font-medium text-gray-700">Stat Tags</div>
                              <div className="flex flex-wrap gap-3">
                                {questStatOptions.map((stat) => (
                                  <label key={`quick-challenge-stat-${stat.id}`} className="flex items-center gap-2 text-sm text-gray-700">
                                    <input
                                      type="checkbox"
                                      checked={quickCreateChallengeForm.statTags.includes(stat.id)}
                                      onChange={(e) =>
                                        setQuickCreateChallengeForm((prev) => ({
                                          ...prev,
                                          statTags: e.target.checked
                                            ? [...prev.statTags, stat.id]
                                            : prev.statTags.filter((tag) => tag !== stat.id),
                                        }))
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
                        <label className="block text-sm font-medium text-gray-700">Polygon Points</label>
                        <textarea
                          className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                          rows={3}
                          placeholder='[[lng,lat],[lng,lat],[lng,lat]]'
                          value={nodeForm.polygonPoints}
                          onChange={(e) => {
                            const value = e.target.value;
                            setNodeForm((prev) => ({ ...prev, polygonPoints: value }));
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
                              setNodeForm((prev) => ({ ...prev, polygonPoints: '' }));
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
                      (nodeForm.nodeType === 'poi' && !nodeForm.pointOfInterestId) ||
                      (nodeForm.nodeType === 'scenario' && !nodeForm.scenarioId) ||
                      (nodeForm.nodeType === 'monster' && !nodeForm.monsterEncounterId) ||
                      (nodeForm.nodeType === 'challenge' && !nodeForm.challengeId)
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
                        onClick={snapToZone}
                      >
                        Snap to Zone
                      </button>
                      <button
                        type="button"
                        className="rounded-md border border-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-50"
                        onClick={() => setPolygonRefreshNonce((prev) => prev + 1)}
                      >
                        Add Polygons to Map
                      </button>
                    </div>
                    <div className="flex items-center gap-3 text-xs text-gray-600">
                      <span className="flex items-center gap-1">
                        <span className="inline-block h-2.5 w-2.5 rounded-full bg-amber-500 border border-amber-800" />
                        POI nodes
                      </span>
                      <span className="flex items-center gap-1">
                        <span className="inline-block h-2.5 w-2.5 rounded-full bg-teal-500 border border-teal-800" />
                        Scenario nodes
                      </span>
                      <span className="flex items-center gap-1">
                        <span className="inline-block h-2.5 w-2.5 rounded-full bg-red-500 border border-red-900" />
                        Monster nodes
                      </span>
                      <span className="flex items-center gap-1">
                        <span className="inline-block h-2.5 w-2.5 rounded-full bg-blue-500 border border-blue-800" />
                        Filtered POIs
                      </span>
                      <span className="flex items-center gap-1">
                        <span className="inline-block h-2.5 w-2.5 rounded-full bg-purple-500 border border-purple-800" />
                        POIs with character
                      </span>
                      <span className="flex items-center gap-1">
                        <span className="inline-block h-2.5 w-2.5 rounded-full bg-emerald-500 border border-emerald-800" />
                        Character locations
                      </span>
                    </div>
                  </div>
                  <p className="mt-1 text-sm text-gray-600">
                    Use the POI filters above to narrow the map to locations that fit this quest.
                  </p>
                  <div className="mt-3 h-80 w-full overflow-hidden rounded-md border border-gray-200 relative">
                    <div ref={questMapContainer} className="h-full w-full" />
                  </div>
                </div>

                {selectedPoiForModal && (
                  <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
                    <div className="w-full max-w-lg rounded-lg bg-white p-6 shadow-lg">
                      <div className="flex items-start justify-between">
                        <h3 className="text-lg font-semibold">{selectedPoiForModal.name}</h3>
                        <button
                          className="text-gray-500 hover:text-gray-700"
                          onClick={() => setSelectedPoiForModal(null)}
                        >
                          Close
                        </button>
                      </div>
                      {selectedPoiForModal.imageURL && (
                        <img
                          src={selectedPoiForModal.imageURL}
                          alt={selectedPoiForModal.name}
                          className="mt-3 w-full rounded-md"
                        />
                      )}
                      {selectedPoiForModal.description && (
                        <p className="mt-3 text-sm text-gray-600">{selectedPoiForModal.description}</p>
                      )}
                      {selectedPoiForModal.tags && selectedPoiForModal.tags.length > 0 && (
                        <div className="mt-3 flex flex-wrap gap-2">
                          {selectedPoiForModal.tags.map((tag) => (
                            <span
                              key={tag.id}
                              className="rounded-full border border-gray-200 bg-gray-50 px-2 py-1 text-xs text-gray-700"
                            >
                              {tag.name}
                            </span>
                          ))}
                        </div>
                      )}
                      <div className="mt-5 flex items-center justify-end gap-3">
                        <button
                          className="rounded-md border border-gray-300 px-4 py-2 text-sm"
                          onClick={() => setSelectedPoiForModal(null)}
                        >
                          Cancel
                        </button>
                        <button
                          className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white"
                          onClick={() => {
                            window.alert(
                              'Quest nodes now support only Scenario, Monster, or Challenge targets.'
                            );
                            setSelectedPoiForModal(null);
                          }}
                        >
                          Select for Node
                        </button>
                      </div>
                    </div>
                  </div>
                )}

                {showImportModal && (
                  <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
                    <div className="w-full max-w-2xl max-h-[90vh] overflow-y-auto rounded-lg bg-white p-6 shadow-lg">
                      <div className="flex items-start justify-between">
                        <h3 className="text-lg font-semibold">Import Point of Interest</h3>
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
                        <div className="mt-3 text-sm text-red-600">{importError}</div>
                      )}

                      <div className="mt-4">
                        <label className="block text-sm font-medium text-gray-700 mb-1">Zone</label>
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
                        <label className="block text-sm font-medium text-gray-700 mb-1">Search Google Maps</label>
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
                          <div className="p-4 text-sm text-gray-500">No results yet.</div>
                        )}
                        {candidates.map((candidate) => (
                          <button
                            key={candidate.place_id}
                            type="button"
                            className={`w-full text-left px-4 py-3 border-b border-gray-100 hover:bg-gray-50 ${
                              selectedCandidate?.place_id === candidate.place_id ? 'bg-blue-50' : ''
                            }`}
                            onClick={() => setSelectedCandidate(candidate)}
                          >
                            <div className="font-medium">{candidate.name}</div>
                            <div className="text-xs text-gray-500">{candidate.formatted_address}</div>
                          </button>
                        ))}
                      </div>

                      <div className="mt-6">
                        <div className="flex items-center justify-between mb-2">
                          <h4 className="text-sm font-semibold">Import Status</h4>
                          <button
                            type="button"
                            className="text-xs text-blue-600"
                            onClick={() => fetchImportJobs(importZoneId || questForm.zoneId || undefined)}
                          >
                            Refresh
                          </button>
                        </div>
                        <div className="border border-gray-200 rounded-md max-h-40 overflow-y-auto">
                          {importJobs.length === 0 && (
                            <div className="p-3 text-xs text-gray-500">No import activity yet.</div>
                          )}
                          {importJobs.map((job) => (
                            <div key={job.id} className="flex items-center justify-between px-3 py-2 border-b border-gray-100 text-xs">
                              <div>
                                <div className="font-medium">{job.placeId}</div>
                                {job.errorMessage && (
                                  <div className="text-red-600">{job.errorMessage}</div>
                                )}
                              </div>
                              <div className="flex items-center gap-2">
                                <div className="uppercase text-[10px] text-gray-600">{job.status}</div>
                                {job.status === 'failed' && (
                                  <button
                                    type="button"
                                    className="rounded-md border border-gray-300 px-2 py-1 text-[10px] text-gray-700"
                                    onClick={() => handleRetryImport(job.placeId, job.zoneId)}
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
                        <h3 className="text-lg font-semibold">Character Locations</h3>
                        <button
                          className="text-gray-500 hover:text-gray-700"
                          onClick={() => setCharacterLocationsOpen(false)}
                        >
                          Close
                        </button>
                      </div>
                      {characterLocationsLoading ? (
                        <div className="mt-4 text-sm text-gray-600">Loading locations...</div>
                      ) : (
                        <div className="mt-4 space-y-3">
                          {selectedCharacterLocations.length === 0 && (
                            <div className="text-sm text-gray-500">No locations yet.</div>
                          )}
                          {selectedCharacterLocations.map((location, index) => (
                            <div key={`${location.latitude}-${location.longitude}-${index}`} className="flex items-center gap-2">
                              <input
                                type="number"
                                className="w-1/2 rounded-md border border-gray-300 p-2 text-sm"
                                value={location.latitude}
                                onChange={(e) => handleUpdateCharacterLocation(index, 'latitude', Number(e.target.value))}
                                placeholder="Latitude"
                              />
                              <input
                                type="number"
                                className="w-1/2 rounded-md border border-gray-300 p-2 text-sm"
                                value={location.longitude}
                                onChange={(e) => handleUpdateCharacterLocation(index, 'longitude', Number(e.target.value))}
                                placeholder="Longitude"
                              />
                              <button
                                className="rounded-md border border-red-200 bg-red-50 px-2 py-1 text-xs text-red-700"
                                onClick={() => handleRemoveCharacterLocation(index)}
                              >
                                Remove
                              </button>
                            </div>
                          ))}
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

                <div className="space-y-4">
                  {(selectedQuest.nodes ?? [])
                    .slice()
                    .sort((a, b) => a.orderIndex - b.orderIndex)
                    .map((node) => (
                      <div key={node.id} className="border border-gray-200 rounded-md p-4">
                        <div className="flex items-center justify-between">
                          <div>
                            <div className="font-semibold">Node {node.orderIndex}</div>
                            <div className="text-sm text-gray-600">
                              {node.pointOfInterestId
                                ? `POI: ${pointsOfInterest.find((poi) => poi.id === node.pointOfInterestId)?.name ?? node.pointOfInterestId}`
                                : node.scenarioId
                                  ? `Scenario: ${summarizeScenarioPrompt(
                                      scenarios.find((scenario) => scenario.id === node.scenarioId)?.prompt ?? ''
                                    )}`
                                : node.monsterEncounterId || node.monsterId
                                    ? `Monster Encounter: ${
                                        monsterEncounters.find((monster) => monster.id === (node.monsterEncounterId ?? node.monsterId))?.name ??
                                        node.monsterEncounterId ??
                                        node.monsterId
                                      }`
                                    : node.challengeId
                                      ? `Challenge: ${challenges.find((challenge) => challenge.id === node.challengeId)?.question ?? node.challengeId}`
                                    : 'Polygon'}
                            </div>
                          </div>
                          <button
                            className="rounded-md border border-red-200 bg-red-50 px-3 py-1 text-xs text-red-700 hover:bg-red-100"
                            onClick={() => handleDeleteNode(node)}
                          >
                            Remove Node
                          </button>
                        </div>

                        <div className="mt-3">
                          <h4 className="font-semibold mb-2">Challenges</h4>
                          <div className="space-y-2 mb-3">
                            {(node.challenges ?? []).map((challenge) => {
                              const editDraft = challengeEdits[challenge.id] ?? emptyChallengeForm;
                              const isEditing = Boolean(challengeEdits[challenge.id]);
                              return (
                                <div key={challenge.id} className="border border-gray-200 rounded-md p-2 text-sm">
                                  <div className="flex items-start justify-between gap-3">
                                    <div>
                                      <div>
                                        Tier {challenge.tier} · Difficulty {challenge.difficulty ?? 0} · Reward {challenge.reward} · Input{' '}
                                        {resolveChallengeSubmissionType(challenge, node).toUpperCase()}
                                      </div>
                                      <div className="text-xs text-gray-500">
                                        Shuffle: {formatChallengeShuffleStatus(challenge.challengeShuffleStatus)}
                                      </div>
                                      {!isEditing && (
                                        <>
                                          <div className="text-gray-600">{challenge.question}</div>
                                          {challenge.statTags && challenge.statTags.length > 0 && (
                                            <div className="text-xs text-gray-500">
                                              Stats:{' '}
                                              {challenge.statTags
                                                .map((tag) => tag.charAt(0).toUpperCase() + tag.slice(1))
                                                .join(', ')}
                                            </div>
                                          )}
                                          {challenge.proficiency && (
                                            <div className="text-xs text-gray-500">Proficiency: {challenge.proficiency}</div>
                                          )}
                                          {challenge.challengeShuffleError && (
                                            <div className="text-xs text-red-600">
                                              Shuffle error: {challenge.challengeShuffleError}
                                            </div>
                                          )}
                                        </>
                                      )}
                                    </div>
                                    <div className="flex items-center gap-2">
                                      <button
                                        type="button"
                                        className="rounded-md border border-indigo-200 bg-indigo-50 px-2 py-1 text-xs text-indigo-700 hover:bg-indigo-100 disabled:opacity-60"
                                        onClick={() => handleShuffleSavedChallenge(node, challenge)}
                                        disabled={
                                          isEditing ||
                                          shufflingChallengeId === challenge.id ||
                                          challenge.challengeShuffleStatus === 'queued' ||
                                          challenge.challengeShuffleStatus === 'in_progress'
                                        }
                                      >
                                        {shufflingChallengeId === challenge.id ||
                                        challenge.challengeShuffleStatus === 'queued' ||
                                        challenge.challengeShuffleStatus === 'in_progress'
                                          ? 'Shuffling...'
                                          : 'Shuffle'}
                                      </button>
                                      <button
                                        type="button"
                                        className="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 hover:bg-gray-50"
                                        onClick={() =>
                                          isEditing ? handleCancelEditChallenge(challenge.id) : handleStartEditChallenge(node, challenge)
                                        }
                                      >
                                        {isEditing ? 'Cancel' : 'Edit'}
                                      </button>
                                    </div>
                                  </div>
                                  {isEditing && (
                                    <div className="mt-3">
                                      <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
                                        <div>
                                          <label className="block text-xs font-medium text-gray-700">Tier</label>
                                          <input
                                            type="number"
                                            className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                            value={editDraft.tier}
                                            onChange={(e) =>
                                              handleEditChallengeDraftChange(challenge.id, { tier: Number(e.target.value) })
                                            }
                                          />
                                        </div>
                                        <div>
                                          <label className="block text-xs font-medium text-gray-700">Difficulty</label>
                                          <input
                                            type="number"
                                            min={0}
                                            className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                            value={editDraft.difficulty}
                                            onChange={(e) =>
                                              handleEditChallengeDraftChange(challenge.id, {
                                                difficulty: Number(e.target.value),
                                              })
                                            }
                                          />
                                        </div>
                                        <div>
                                          <label className="block text-xs font-medium text-gray-700">Input Type</label>
                                          <select
                                            className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                            value={editDraft.submissionType}
                                            onChange={(e) =>
                                              handleEditChallengeDraftChange(challenge.id, {
                                                submissionType: e.target.value as QuestNodeSubmissionType,
                                              })
                                            }
                                          >
                                            {questNodeSubmissionOptions.map((option) => (
                                              <option key={option.value} value={option.value}>
                                                {option.label}
                                              </option>
                                            ))}
                                          </select>
                                        </div>
                                        <div>
                                          <label className="block text-xs font-medium text-gray-700">Reward</label>
                                          <input
                                            type="number"
                                            className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                            value={editDraft.reward}
                                            onChange={(e) =>
                                              handleEditChallengeDraftChange(challenge.id, { reward: Number(e.target.value) })
                                            }
                                          />
                                        </div>
                                        <div>
                                          <label className="block text-xs font-medium text-gray-700">Inventory Item</label>
                                          <select
                                            className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                            value={editDraft.inventoryItemId}
                                            onChange={(e) =>
                                              handleEditChallengeDraftChange(challenge.id, { inventoryItemId: e.target.value })
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
                                          <label className="block text-xs font-medium text-gray-700">Stat Tags</label>
                                          <div className="mt-2 grid grid-cols-2 gap-2 sm:grid-cols-3">
                                            {questStatOptions.map((stat) => (
                                              <label key={stat.id} className="flex items-center gap-2 text-xs text-gray-700">
                                                <input
                                                  type="checkbox"
                                                  checked={editDraft.statTags.includes(stat.id)}
                                                  onChange={(e) => {
                                                    const current = editDraft.statTags;
                                                    const next = e.target.checked
                                                      ? [...current, stat.id]
                                                      : current.filter((tag) => tag !== stat.id);
                                                    handleEditChallengeDraftChange(challenge.id, { statTags: next });
                                                  }}
                                                />
                                                {stat.label}
                                              </label>
                                            ))}
                                          </div>
                                        </div>
                                        <div className="md:col-span-2">
                                          <label className="block text-xs font-medium text-gray-700">Proficiency</label>
                                          <input
                                            className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                            value={editDraft.proficiency}
                                            list="proficiency-options"
                                            onChange={(e) => {
                                              handleEditChallengeDraftChange(challenge.id, { proficiency: e.target.value });
                                              handleProficiencyInputChange(e.target.value);
                                            }}
                                            placeholder="Drawing"
                                          />
                                        </div>
                                        <div className="md:col-span-4">
                                          <label className="block text-xs font-medium text-gray-700">Question</label>
                                          <textarea
                                            className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                            rows={2}
                                            value={editDraft.question}
                                            onChange={(e) =>
                                              handleEditChallengeDraftChange(challenge.id, { question: e.target.value })
                                            }
                                          />
                                        </div>
                                      </div>
                                      <button
                                        type="button"
                                        className="mt-3 bg-blue-600 text-white px-3 py-2 rounded-md"
                                        onClick={() => handleUpdateChallenge(node, challenge)}
                                      >
                                        Save Changes
                                      </button>
                                    </div>
                                  )}
                                </div>
                              );
                            })}
                            {(node.challenges ?? []).length === 0 && (
                              <div className="text-sm text-gray-500">No challenges yet.</div>
                            )}
                          </div>

                          <div className="bg-gray-50 border border-gray-200 rounded-md p-3">
                            <h5 className="font-semibold mb-2">Add Challenge</h5>
                            <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
                              {node.pointOfInterestId && (
                                <div className="md:col-span-4 rounded-md border border-amber-200 bg-amber-50 p-3">
                                  <div className="text-xs font-semibold text-amber-900">Location Archetype Challenge</div>
                                  <div className="mt-2 grid grid-cols-1 md:grid-cols-2 gap-3">
                                    <div>
                                      <label className="block text-xs font-medium text-gray-700">Archetype</label>
                                      <select
                                        className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                        value={(challengeDrafts[node.id] ?? emptyChallengeForm).locationArchetypeId}
                                        onChange={(e) =>
                                          handleChallengeDraftChange(node.id, {
                                            locationArchetypeId: e.target.value,
                                            locationChallenge: '',
                                            question: '',
                                            submissionType: 'photo',
                                            proficiency: '',
                                          })
                                        }
                                      >
                                        <option value="">Select archetype</option>
                                        {locationArchetypes.map((archetype) => (
                                          <option key={archetype.id} value={archetype.id}>
                                            {archetype.name}
                                          </option>
                                        ))}
                                      </select>
                                    </div>
                                    <div>
                                      <label className="block text-xs font-medium text-gray-700">Challenge</label>
                                      {(() => {
                                        const selectedArchetype = locationArchetypes.find(
                                          (archetype) =>
                                            archetype.id ===
                                            (challengeDrafts[node.id] ?? emptyChallengeForm).locationArchetypeId
                                        );
                                        const challenges = selectedArchetype?.challenges ?? [];
                                        return (
                                      <select
                                        className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                        value={(challengeDrafts[node.id] ?? emptyChallengeForm).locationChallenge}
                                        onChange={(e) =>
                                          (() => {
                                            const value = e.target.value;
                                            const index = value === '' ? NaN : Number(value);
                                            const selected = Number.isFinite(index) ? challenges[index] : undefined;
                                            handleChallengeDraftChange(node.id, {
                                              locationChallenge: value,
                                              question: selected?.question ?? '',
                                              submissionType: (selected?.submissionType ?? 'photo') as QuestNodeSubmissionType,
                                              proficiency: selected?.proficiency ?? '',
                                              difficulty:
                                                selected?.difficulty ??
                                                (challengeDrafts[node.id] ?? emptyChallengeForm).difficulty,
                                            });
                                          })()
                                        }
                                      >
                                        <option value="">Select challenge</option>
                                        {challenges.map((challenge, index) => (
                                          <option key={`${challenge.question}-${index}`} value={index}>
                                            {challenge.question} · {challenge.submissionType.toUpperCase()}
                                            {challenge.proficiency ? ` · ${challenge.proficiency}` : ''}
                                          </option>
                                        ))}
                                      </select>
                                        );
                                      })()}
                                    </div>
                                  </div>
                                  <p className="mt-2 text-xs text-amber-800">
                                    Selecting a challenge will auto-fill the question field and input type.
                                  </p>
                                </div>
                              )}
                              <div>
                                <label className="block text-xs font-medium text-gray-700">Tier</label>
                                <input
                                  type="number"
                                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                  value={(challengeDrafts[node.id] ?? emptyChallengeForm).tier}
                                  onChange={(e) => handleChallengeDraftChange(node.id, { tier: Number(e.target.value) })}
                                />
                              </div>
                              <div>
                                <label className="block text-xs font-medium text-gray-700">Difficulty</label>
                                <input
                                  type="number"
                                  min={0}
                                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                  value={(challengeDrafts[node.id] ?? emptyChallengeForm).difficulty}
                                  onChange={(e) =>
                                    handleChallengeDraftChange(node.id, { difficulty: Number(e.target.value) })
                                  }
                                />
                              </div>
                              <div>
                                <label className="block text-xs font-medium text-gray-700">Input Type</label>
                                <select
                                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                  value={(challengeDrafts[node.id] ?? emptyChallengeForm).submissionType}
                                  onChange={(e) =>
                                    handleChallengeDraftChange(node.id, {
                                      submissionType: e.target.value as QuestNodeSubmissionType,
                                    })
                                  }
                                >
                                  {questNodeSubmissionOptions.map((option) => (
                                    <option key={option.value} value={option.value}>
                                      {option.label}
                                    </option>
                                  ))}
                                </select>
                              </div>
                              <div>
                                <label className="block text-xs font-medium text-gray-700">Reward</label>
                                <input
                                  type="number"
                                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                  value={(challengeDrafts[node.id] ?? emptyChallengeForm).reward}
                                  onChange={(e) => handleChallengeDraftChange(node.id, { reward: Number(e.target.value) })}
                                />
                              </div>
                              <div>
                                <label className="block text-xs font-medium text-gray-700">Inventory Item</label>
                                <select
                                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                  value={(challengeDrafts[node.id] ?? emptyChallengeForm).inventoryItemId}
                                  onChange={(e) => handleChallengeDraftChange(node.id, { inventoryItemId: e.target.value })}
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
                                <label className="block text-xs font-medium text-gray-700">Stat Tags</label>
                                <div className="mt-2 grid grid-cols-2 gap-2 sm:grid-cols-3">
                                  {questStatOptions.map((stat) => (
                                    <label key={stat.id} className="flex items-center gap-2 text-xs text-gray-700">
                                      <input
                                        type="checkbox"
                                        checked={(challengeDrafts[node.id] ?? emptyChallengeForm).statTags.includes(stat.id)}
                                        onChange={(e) => {
                                          const current = (challengeDrafts[node.id] ?? emptyChallengeForm).statTags;
                                          const next = e.target.checked
                                            ? [...current, stat.id]
                                            : current.filter((tag) => tag !== stat.id);
                                          handleChallengeDraftChange(node.id, { statTags: next });
                                        }}
                                      />
                                      {stat.label}
                                    </label>
                                  ))}
                                </div>
                              </div>
                              <div className="md:col-span-2">
                                <label className="block text-xs font-medium text-gray-700">Proficiency</label>
                                <input
                                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                  value={(challengeDrafts[node.id] ?? emptyChallengeForm).proficiency}
                                  list="proficiency-options"
                                  onChange={(e) => {
                                    handleChallengeDraftChange(node.id, { proficiency: e.target.value });
                                    handleProficiencyInputChange(e.target.value);
                                  }}
                                  placeholder="Drawing"
                                />
                              </div>
                              <div className="md:col-span-4">
                                <label className="block text-xs font-medium text-gray-700">Question</label>
                                <textarea
                                  className="mt-1 block w-full border border-gray-300 rounded-md p-2"
                                  rows={2}
                                  value={(challengeDrafts[node.id] ?? emptyChallengeForm).question}
                                  onChange={(e) => handleChallengeDraftChange(node.id, { question: e.target.value })}
                                />
                              </div>
                            </div>
                            <button
                              className="mt-3 bg-blue-600 text-white px-3 py-2 rounded-md"
                              onClick={() => handleCreateChallenge(node)}
                            >
                              Add Challenge
                            </button>
                          </div>
                        </div>
                      </div>
                    ))}
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
