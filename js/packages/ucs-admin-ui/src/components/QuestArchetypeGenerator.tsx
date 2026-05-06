import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAPI } from '@poltergeist/contexts';
import {
  QuestArchetypeSuggestionDraft,
  QuestArchetypeSuggestionJob,
  QuestArchetypeSuggestionNode,
  QuestArchetypeSuggestionPreset,
} from '@poltergeist/types';
import { useQuestArchetypes } from '../contexts/questArchetypes.tsx';
import { useZoneKinds, zoneKindLabel } from './zoneKindHelpers.ts';
import './questArchetypeTheme.css';

type GeneratorFormState = {
  count: string;
  yeetIt: boolean;
  zoneKind: string;
  themePrompt: string;
  familyTagsText: string;
  familyMixTargets: Record<string, string>;
  characterTagsText: string;
  internalTagsText: string;
  requiredLocationArchetypeIds: string[];
  requiredLocationMetadataTagsText: string;
};

type QuestArchetypeSuggestionJobPayload = {
  count: number;
  yeetIt: boolean;
  zoneKind: string;
  themePrompt: string;
  familyTags: string[];
  familyMixTargets: Record<string, number>;
  characterTags: string[];
  internalTags: string[];
  requiredLocationArchetypeIds: string[];
  requiredLocationMetadataTags: string[];
};

type QuestPresetZoneKind = {
  slug: string;
  name: string;
  description?: string | null;
};

type QuestPresetLocationArchetype = {
  id: string;
  name: string;
  includedTypes?: string[] | null;
};

type QuestPresetSeed = {
  label: string;
  themePrompts: string[];
  familyTags: string[];
  familyFocuses: string[];
  characterTags: string[];
  internalTags: string[];
  metadataTags: string[];
  locationKeywords: string[];
  zoneKindKeywords: string[];
};

const QUEST_FAMILY_OPTIONS = [
  { slug: 'investigation', label: 'Investigation' },
  { slug: 'delivery', label: 'Delivery' },
  { slug: 'negotiation', label: 'Negotiation' },
  { slug: 'pursuit', label: 'Pursuit' },
  { slug: 'containment', label: 'Containment' },
  { slug: 'omen_chasing', label: 'Omen Chasing' },
  { slug: 'ritual_interruption', label: 'Ritual Interruption' },
  { slug: 'survival', label: 'Survival' },
  { slug: 'rescue', label: 'Rescue' },
  { slug: 'combat_finale', label: 'Combat Finale' },
];

const QUEST_PRESET_LIBRARY: QuestPresetSeed[] = [
  {
    label: 'Smuggler Web',
    themePrompts: [
      'A contraband route is under pressure from rival crews, civic inspectors, and nervous couriers.',
      'Street-level smugglers need deniable help moving sensitive cargo through a district full of watchers.',
    ],
    familyTags: ['criminal', 'trade', 'civic'],
    familyFocuses: ['delivery', 'investigation', 'negotiation'],
    characterTags: ['courier', 'fence', 'dockhand', 'customs_clerk', 'fixer'],
    internalTags: ['smuggling', 'handoff', 'contraband', 'route_pressure', 'black_market'],
    metadataTags: ['market', 'warehouse', 'alley', 'waterfront', 'checkpoint'],
    locationKeywords: ['market', 'warehouse', 'harbor', 'gate', 'bridge', 'station'],
    zoneKindKeywords: ['harbor', 'market', 'river', 'canal', 'urban'],
  },
  {
    label: 'Occult Blackout',
    themePrompts: [
      'Protect a district during a magical blackout caused by unstable wards and opportunistic spirits.',
      'A surge of omen-laden failures is spreading through ritual infrastructure that should have been stable.',
    ],
    familyTags: ['occult', 'civic', 'storm'],
    familyFocuses: ['investigation', 'ritual_interruption', 'containment'],
    characterTags: ['medium', 'lamplighter', 'groundskeeper', 'apprentice_mage', 'sacristan'],
    internalTags: ['wards', 'blackout', 'omens', 'sigils', 'spirit_leak'],
    metadataTags: ['shrine', 'plaza', 'rooftop', 'substation', 'memorial'],
    locationKeywords: ['shrine', 'chapel', 'tower', 'plaza', 'park', 'cemetery'],
    zoneKindKeywords: ['ruin', 'cemetery', 'old', 'temple', 'urban'],
  },
  {
    label: 'Rooftop Chase',
    themePrompts: [
      'A dangerous messenger has taken to the rooftops with something too important to lose.',
      'Someone fast, desperate, and magically assisted is turning the skyline into a pursuit corridor.',
    ],
    familyTags: ['criminal', 'speed', 'surveillance'],
    familyFocuses: ['pursuit', 'delivery', 'combat_finale'],
    characterTags: ['runner', 'lookout', 'messenger', 'bounty_hunter', 'acrobat'],
    internalTags: ['rooftops', 'chase', 'signal_flares', 'tight_deadlines', 'crossings'],
    metadataTags: ['rooftop', 'stairwell', 'balcony', 'skybridge', 'billboard'],
    locationKeywords: ['tower', 'bridge', 'stairs', 'roof', 'balcony', 'high'],
    zoneKindKeywords: ['downtown', 'highrise', 'urban', 'industrial'],
  },
  {
    label: 'Floodwall Breach',
    themePrompts: [
      'A district barrier is failing, and local crews need help containing the breach before panic spreads.',
      'Something has compromised the floodwall, forcing street-level responders to improvise under pressure.',
    ],
    familyTags: ['civic', 'disaster', 'infrastructure'],
    familyFocuses: ['containment', 'survival', 'rescue'],
    characterTags: ['engineer', 'boatman', 'first_responder', 'maintenance_crew', 'lookout'],
    internalTags: ['barrier', 'breach', 'stabilization', 'evacuation', 'emergency_routes'],
    metadataTags: ['waterfront', 'pump', 'levee', 'floodgate', 'service_tunnel'],
    locationKeywords: ['water', 'tunnel', 'bridge', 'station', 'service', 'gate'],
    zoneKindKeywords: ['coast', 'harbor', 'river', 'swamp', 'canal'],
  },
  {
    label: 'Missing Envoy',
    themePrompts: [
      'A politically sensitive envoy vanished between two supposedly safe meeting points.',
      'Multiple factions want a missing emissary found first, but each one wants a different outcome.',
    ],
    familyTags: ['diplomacy', 'intrigue', 'civic'],
    familyFocuses: ['investigation', 'negotiation', 'rescue'],
    characterTags: ['envoy', 'bodyguard', 'clerk', 'mediator', 'fixer'],
    internalTags: ['missing_person', 'faction_pressure', 'quiet_recovery', 'back_channels', 'escort'],
    metadataTags: ['courtyard', 'meeting_hall', 'garden', 'checkpoint', 'arcade'],
    locationKeywords: ['hall', 'court', 'garden', 'market', 'embassy', 'office'],
    zoneKindKeywords: ['court', 'garden', 'urban', 'old', 'plaza'],
  },
  {
    label: 'Market Infestation',
    themePrompts: [
      'Predatory creatures are feeding on the district economy by nesting inside supply routes and vendor spaces.',
      'A lively market is being slowly strangled by an infestation nobody can fully pin down yet.',
    ],
    familyTags: ['trade', 'creatures', 'civic'],
    familyFocuses: ['investigation', 'containment', 'combat_finale'],
    characterTags: ['vendor', 'ratcatcher', 'porter', 'apothecary', 'market_guard'],
    internalTags: ['infestation', 'supply_chain', 'vendor_panic', 'nests', 'scarcity'],
    metadataTags: ['market', 'stall', 'storage', 'drain', 'loading_dock'],
    locationKeywords: ['market', 'storage', 'drain', 'dock', 'alley', 'basement'],
    zoneKindKeywords: ['market', 'industrial', 'urban', 'sewer'],
  },
  {
    label: 'Star Omen Trail',
    themePrompts: [
      'A chain of uncanny signs is pointing toward a district-scale event that only a few people can read in time.',
      'Celestial omens are surfacing in public spaces, and someone needs to chase the pattern before others weaponize it.',
    ],
    familyTags: ['occult', 'scholarly', 'omens'],
    familyFocuses: ['omen_chasing', 'investigation', 'ritual_interruption'],
    characterTags: ['astrologer', 'librarian', 'scribe', 'watcher', 'ritualist'],
    internalTags: ['omens', 'sky_signs', 'pattern_reading', 'star_charts', 'prophecy'],
    metadataTags: ['observatory', 'library', 'courtyard', 'monument', 'rooftop'],
    locationKeywords: ['library', 'tower', 'observatory', 'monument', 'roof', 'plaza'],
    zoneKindKeywords: ['arcane', 'temple', 'old', 'mountain', 'ruin'],
  },
  {
    label: 'Strike Rescue',
    themePrompts: [
      'A labor action has frozen a district, leaving vulnerable people and vital goods stranded between factions.',
      'A transport strike turned ugly, and now relief, diplomacy, and extraction are tangled together.',
    ],
    familyTags: ['labor', 'civic', 'trade'],
    familyFocuses: ['negotiation', 'rescue', 'delivery'],
    characterTags: ['organizer', 'porter', 'medic', 'quartermaster', 'teamster'],
    internalTags: ['strike', 'mutual_aid', 'relief_runs', 'picket_lines', 'tense_standoffs'],
    metadataTags: ['depot', 'station', 'clinic', 'queue', 'warehouse'],
    locationKeywords: ['station', 'depot', 'clinic', 'warehouse', 'market', 'yard'],
    zoneKindKeywords: ['industrial', 'urban', 'harbor', 'market'],
  },
];

const emptyFamilyMixTargets = () =>
  QUEST_FAMILY_OPTIONS.reduce<Record<string, string>>((accumulator, family) => {
    accumulator[family.slug] = '0';
    return accumulator;
  }, {});

const emptyGeneratorForm = (): GeneratorFormState => ({
  count: '2',
  yeetIt: false,
  zoneKind: '',
  themePrompt: '',
  familyTagsText: '',
  familyMixTargets: emptyFamilyMixTargets(),
  characterTagsText: '',
  internalTagsText: '',
  requiredLocationArchetypeIds: [],
  requiredLocationMetadataTagsText: '',
});

const isPendingStatus = (status?: string | null) =>
  status === 'queued' || status === 'in_progress';

const statusChipClass = (status?: string | null) => {
  switch (status) {
    case 'completed':
    case 'converted':
      return 'qa-chip success';
    case 'failed':
      return 'qa-chip danger';
    case 'in_progress':
      return 'qa-chip accent';
    case 'queued':
      return 'qa-chip muted';
    default:
      return 'qa-chip muted';
  }
};

const formatStatus = (status?: string | null) =>
  (status || 'unknown').replace(/_/g, ' ');

const formatDate = (value?: string | null) => {
  if (!value) return 'n/a';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

const parseTags = (value: string) =>
  value
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean);

const requestedQuestArchetypeSuggestionBatchCount = (value: string) =>
  Math.max(1, parseInt(value, 10) || 1);

const buildQuestArchetypeSuggestionJobPayload = (
  form: GeneratorFormState
): QuestArchetypeSuggestionJobPayload => ({
  count: requestedQuestArchetypeSuggestionBatchCount(form.count),
  yeetIt: form.yeetIt,
  zoneKind: form.zoneKind.trim(),
  themePrompt: form.themePrompt.trim(),
  familyTags: parseTags(form.familyTagsText),
  familyMixTargets: buildFamilyMixTargetsPayload(form.familyMixTargets),
  characterTags: parseTags(form.characterTagsText),
  internalTags: parseTags(form.internalTagsText),
  requiredLocationArchetypeIds: form.requiredLocationArchetypeIds,
  requiredLocationMetadataTags: parseTags(
    form.requiredLocationMetadataTagsText
  ),
});

const buildGeneratorFormFromPreset = (
  preset: QuestArchetypeSuggestionPreset,
  requestedCount: number,
  zoneKindHint: string,
  yeetIt: boolean
): GeneratorFormState => ({
  count: String(requestedCount),
  yeetIt,
  zoneKind: zoneKindHint || preset.zoneKind?.trim() || '',
  themePrompt: preset.themePrompt ?? '',
  familyTagsText: (preset.familyTags ?? []).join(', '),
  familyMixTargets: buildFamilyMixTargetFormState(preset.familyMixTargets),
  characterTagsText: (preset.characterTags ?? []).join(', '),
  internalTagsText: (preset.internalTags ?? []).join(', '),
  requiredLocationArchetypeIds: preset.requiredLocationArchetypeIds ?? [],
  requiredLocationMetadataTagsText: (
    preset.requiredLocationMetadataTags ?? []
  ).join(', '),
});

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
    typeof error.response.error !== 'string' &&
    typeof (error.response.data as { error?: unknown }).error === 'string'
  ) {
    return (error.response.data as { error: string }).error;
  }
  return fallback;
};

const draftNodesForReview = (
  draft: QuestArchetypeSuggestionDraft
): QuestArchetypeSuggestionNode[] => {
  const usingGraphNodes = !!(draft.nodes && draft.nodes.length > 0);
  const rawNodes = usingGraphNodes
    ? draft.nodes!
    : (draft.steps ?? []).map((step, index) => ({
          ...step,
          nodeKey: `node_${index + 1}`,
          outcomes:
            index + 1 < (draft.steps?.length ?? 0)
              ? [{ outcome: 'success' as const, nextNodeKey: `node_${index + 2}` }]
              : [],
        }));

  return rawNodes.map((node, index) => ({
    ...node,
    nodeKey: node.nodeKey?.trim() || `node_${index + 1}`,
    outcomes:
      node.outcomes && node.outcomes.length > 0
        ? node.outcomes
        : !usingGraphNodes && index + 1 < rawNodes.length
          ? [
              {
                outcome: 'success' as const,
                nextNodeKey: `node_${index + 2}`,
              },
            ]
          : [],
  }));
};

const draftFailureBranchCount = (nodes: QuestArchetypeSuggestionNode[]) =>
  nodes.reduce(
    (count, node) =>
      count +
      (node.outcomes ?? []).filter((outcome) => outcome.outcome === 'failure')
        .length,
    0
  );

const formatOutcomeLabel = (outcome: string) => outcome.replace(/_/g, ' ');

const buildFamilyMixTargetsPayload = (values: Record<string, string>) =>
  Object.entries(values).reduce<Record<string, number>>(
    (accumulator, [slug, rawValue]) => {
      const count = Math.max(0, parseInt(rawValue, 10) || 0);
      if (count > 0) {
        accumulator[slug] = count;
      }
      return accumulator;
    },
    {}
  );

const familyMixTargetCount = (values: Record<string, string>) =>
  Object.values(buildFamilyMixTargetsPayload(values)).reduce(
    (sum, count) => sum + count,
    0
  );

const formatFamilyMixTargets = (
  targets?: Record<string, number> | null
): string => {
  if (!targets) return 'none';
  const parts = QUEST_FAMILY_OPTIONS.map((family) => {
    const count = targets[family.slug];
    if (!count || count <= 0) return null;
    return `${family.label} x${count}`;
  }).filter(Boolean);
  return parts.length > 0 ? parts.join(', ') : 'none';
};

const buildFamilyMixTargetFormState = (
  targets?: Record<string, number> | null
): Record<string, string> => {
  const next = emptyFamilyMixTargets();
  Object.entries(targets ?? {}).forEach(([slug, count]) => {
    if (!(slug in next)) {
      return;
    }
    next[slug] = String(Math.max(0, count || 0));
  });
  return next;
};

const randomInt = (min: number, max: number) =>
  Math.floor(Math.random() * (max - min + 1)) + min;

const shuffleArray = <T,>(items: T[]): T[] => {
  const next = [...items];
  for (let index = next.length - 1; index > 0; index -= 1) {
    const swapIndex = randomInt(0, index);
    [next[index], next[swapIndex]] = [next[swapIndex], next[index]];
  }
  return next;
};

const sampleMany = <T,>(items: T[], count: number): T[] =>
  shuffleArray(items).slice(0, Math.max(0, Math.min(count, items.length)));

const sampleOne = <T,>(items: T[]): T | null =>
  items.length > 0 ? items[randomInt(0, items.length - 1)] : null;

const uniqueStrings = (items: string[]) => Array.from(new Set(items));

const buildRandomPresetFamilyMixTargets = (
  familyFocuses: string[],
  maxTotal: number
): Record<string, string> => {
  const next = emptyFamilyMixTargets();
  const cappedTotal = Math.max(1, maxTotal);
  const focusCount = Math.max(
    1,
    Math.min(cappedTotal, Math.min(3, familyFocuses.length))
  );
  const selectedFamilies = sampleMany(familyFocuses, focusCount);
  let remaining = cappedTotal;
  selectedFamilies.forEach((slug, index) => {
    if (remaining <= 0) {
      return;
    }
    const remainingSlots = selectedFamilies.length - index - 1;
    const minimum = 1;
    const familyMaximum = slug === 'combat_finale' ? 1 : 2;
    const maximum = Math.max(
      minimum,
      Math.min(familyMaximum, remaining-remainingSlots)
    );
    const assigned = Math.min(
      remaining,
      index === selectedFamilies.length - 1 ? remaining : randomInt(minimum, maximum)
    );
    next[slug] = String(Math.max(minimum, assigned));
    remaining -= Math.max(minimum, assigned);
  });
  return next;
};

const choosePresetZoneKindSlug = (
  preset: QuestPresetSeed,
  zoneKinds: QuestPresetZoneKind[]
): string => {
  if (zoneKinds.length === 0) return '';
  const matching = zoneKinds.filter((zoneKind) => {
    const haystack = [
      zoneKind.slug,
      zoneKind.name,
      zoneKind.description ?? '',
    ]
      .join(' ')
      .toLowerCase();
    return preset.zoneKindKeywords.some((keyword) =>
      haystack.includes(keyword.toLowerCase())
    );
  });
  const selected = sampleOne(matching.length > 0 ? matching : zoneKinds);
  return selected?.slug?.trim() || '';
};

const scorePresetLocationArchetype = (
  archetype: QuestPresetLocationArchetype,
  keywords: string[]
) => {
  const haystack = [
    archetype.name,
    ...(archetype.includedTypes ?? []),
  ]
    .join(' ')
    .toLowerCase();
  return keywords.reduce(
    (score, keyword) => (haystack.includes(keyword.toLowerCase()) ? score + 1 : score),
    0
  );
};

const choosePresetLocationArchetypeIds = (
  preset: QuestPresetSeed,
  locationArchetypes: QuestPresetLocationArchetype[]
): string[] => {
  if (locationArchetypes.length === 0) return [];
  const ranked = locationArchetypes
    .map((archetype) => ({
      archetype,
      score: scorePresetLocationArchetype(archetype, preset.locationKeywords),
    }))
    .sort((left, right) => right.score - left.score);
  const pool = ranked.filter((entry) => entry.score > 0);
  const sourcePool = (pool.length > 0 ? pool : ranked)
    .slice(0, Math.min(8, Math.max(3, pool.length || ranked.length)))
    .map((entry) => entry.archetype);
  if (sourcePool.length === 0) {
    return [];
  }
  const desiredCount = Math.min(
    sourcePool.length,
    randomInt(0, Math.min(2, sourcePool.length))
  );
  if (desiredCount <= 0) {
    return [];
  }
  return sampleMany(sourcePool, desiredCount).map((archetype) => archetype.id);
};

const buildRandomPresetForm = (
  zoneKinds: QuestPresetZoneKind[],
  locationArchetypes: QuestPresetLocationArchetype[],
  options?: {
    preferredCount?: number;
    preferredZoneKind?: string;
    yeetIt?: boolean;
  }
): GeneratorFormState => {
  const preset = sampleOne(QUEST_PRESET_LIBRARY) ?? QUEST_PRESET_LIBRARY[0];
  const selectedCount = Math.max(
    1,
    Math.min(12, Math.trunc(options?.preferredCount ?? 2) || 2)
  );
  const familyMixTargets = buildRandomPresetFamilyMixTargets(
    preset.familyFocuses,
    selectedCount
  );
  const preferredZoneKind = options?.preferredZoneKind?.trim() ?? '';
  const resolvedZoneKind =
    preferredZoneKind ||
    choosePresetZoneKindSlug(preset, zoneKinds);

  return {
    count: String(selectedCount),
    yeetIt: options?.yeetIt ?? false,
    zoneKind: resolvedZoneKind,
    themePrompt: sampleOne(preset.themePrompts) ?? '',
    familyTagsText: uniqueStrings(
      sampleMany(
        uniqueStrings([...preset.familyTags, ...preset.familyFocuses]),
        randomInt(3, Math.min(5, preset.familyTags.length + preset.familyFocuses.length))
      )
    ).join(', '),
    familyMixTargets,
    characterTagsText: sampleMany(
      uniqueStrings(preset.characterTags),
      Math.min(preset.characterTags.length, randomInt(2, 4))
    ).join(', '),
    internalTagsText: uniqueStrings(
      [
        ...sampleMany(
          uniqueStrings(preset.internalTags),
          Math.min(preset.internalTags.length, randomInt(3, 5))
        ),
        ...sampleMany(
          uniqueStrings(preset.metadataTags),
          Math.min(preset.metadataTags.length, randomInt(1, 2))
        ),
      ]
    ).join(', '),
    requiredLocationArchetypeIds: choosePresetLocationArchetypeIds(
      preset,
      locationArchetypes
    ),
    requiredLocationMetadataTagsText: sampleMany(
      uniqueStrings(preset.metadataTags),
      Math.min(preset.metadataTags.length, randomInt(3, 4))
    ).join(', '),
  };
};

export const QuestArchetypeGenerator = () => {
  const { apiClient } = useAPI();
  const { locationArchetypes } = useQuestArchetypes();
  const { zoneKinds, zoneKindBySlug } = useZoneKinds();
  const [form, setForm] = useState<GeneratorFormState>(emptyGeneratorForm);
  const [jobs, setJobs] = useState<QuestArchetypeSuggestionJob[]>([]);
  const [selectedJobId, setSelectedJobId] = useState<string>('');
  const [drafts, setDrafts] = useState<QuestArchetypeSuggestionDraft[]>([]);
  const [loadingJobs, setLoadingJobs] = useState(false);
  const [loadingDrafts, setLoadingDrafts] = useState(false);
  const [queueing, setQueueing] = useState(false);
  const [generatingPreset, setGeneratingPreset] = useState(false);
  const [queueingPresetZoneKinds, setQueueingPresetZoneKinds] = useState(false);
  const [presetZoneKind, setPresetZoneKind] = useState<string>('');
  const [presetQueueZoneKinds, setPresetQueueZoneKinds] = useState<string[]>([]);
  const [pageError, setPageError] = useState<string | null>(null);
  const [jobActionError, setJobActionError] = useState<string | null>(null);
  const [presetStatusMessage, setPresetStatusMessage] = useState<string | null>(
    null
  );
  const [convertingDraftId, setConvertingDraftId] = useState<string | null>(
    null
  );
  const [deletingDraftId, setDeletingDraftId] = useState<string | null>(null);

  const selectedJob = useMemo(
    () => jobs.find((job) => job.id === selectedJobId) ?? null,
    [jobs, selectedJobId]
  );
  const locationArchetypeNamesById = useMemo(() => {
    const map = new Map<string, string>();
    locationArchetypes.forEach((archetype) => {
      map.set(archetype.id, archetype.name);
    });
    return map;
  }, [locationArchetypes]);
  const selectedJobRequiredLocationArchetypes = useMemo(() => {
    if (!selectedJob?.requiredLocationArchetypeIds?.length) {
      return [];
    }
    return selectedJob.requiredLocationArchetypeIds.map(
      (id) =>
        locationArchetypeNamesById.get(id) ??
        `Unknown archetype (${id.slice(0, 8)}...)`
    );
  }, [locationArchetypeNamesById, selectedJob]);
  const requestedFamilyMixCount = useMemo(
    () => familyMixTargetCount(form.familyMixTargets),
    [form.familyMixTargets]
  );
  const requestedBatchCount = useMemo(
    () => requestedQuestArchetypeSuggestionBatchCount(form.count),
    [form.count]
  );
  const presetQueueZoneKindLabels = useMemo(
    () =>
      presetQueueZoneKinds
        .map((slug) => zoneKindLabel(slug, zoneKindBySlug))
        .filter(Boolean),
    [presetQueueZoneKinds, zoneKindBySlug]
  );

  const fetchJobs = useCallback(async () => {
    setLoadingJobs(true);
    try {
      const response = await apiClient.get<QuestArchetypeSuggestionJob[]>(
        '/sonar/questArchetypeSuggestionJobs?limit=30'
      );
      setJobs(response);
      setPageError(null);
      setSelectedJobId((current) => {
        if (current && response.some((job) => job.id === current)) {
          return current;
        }
        return response[0]?.id ?? '';
      });
    } catch (error) {
      console.error('Failed to load quest archetype suggestion jobs', error);
      setPageError(
        extractApiErrorMessage(
          error,
          'Failed to load quest archetype generator jobs.'
        )
      );
    } finally {
      setLoadingJobs(false);
    }
  }, [apiClient]);

  const fetchDrafts = useCallback(
    async (jobId: string) => {
      if (!jobId) {
        setDrafts([]);
        return;
      }
      setLoadingDrafts(true);
      try {
        const response = await apiClient.get<QuestArchetypeSuggestionDraft[]>(
          `/sonar/questArchetypeSuggestionJobs/${jobId}/drafts`
        );
        setDrafts(response);
        setJobActionError(null);
      } catch (error) {
        console.error(
          'Failed to load quest archetype suggestion drafts',
          error
        );
        setJobActionError(
          extractApiErrorMessage(error, 'Failed to load generated drafts.')
        );
      } finally {
        setLoadingDrafts(false);
      }
    },
    [apiClient]
  );

  useEffect(() => {
    void fetchJobs();
  }, [fetchJobs]);

  useEffect(() => {
    if (!selectedJobId) {
      setDrafts([]);
      return;
    }
    void fetchDrafts(selectedJobId);
  }, [fetchDrafts, selectedJobId]);

  useEffect(() => {
    const hasPending = jobs.some((job) => isPendingStatus(job.status));
    if (!hasPending) {
      return;
    }
    const interval = window.setInterval(() => {
      void fetchJobs();
      if (selectedJobId) {
        void fetchDrafts(selectedJobId);
      }
    }, 5000);
    return () => window.clearInterval(interval);
  }, [fetchDrafts, fetchJobs, jobs, selectedJobId]);

  const handleQueueJob = async () => {
    setQueueing(true);
    setJobActionError(null);
    try {
      const created = await apiClient.post<QuestArchetypeSuggestionJob>(
        '/sonar/questArchetypeSuggestionJobs',
        buildQuestArchetypeSuggestionJobPayload(form)
      );
      setJobs((current) => [
        created,
        ...current.filter((job) => job.id !== created.id),
      ]);
      setSelectedJobId(created.id);
      setDrafts([]);
    } catch (error) {
      console.error('Failed to queue quest archetype suggestion job', error);
      setJobActionError(
        extractApiErrorMessage(
          error,
          'Failed to queue quest archetype suggestion job.'
        )
      );
    } finally {
      setQueueing(false);
    }
  };

  const handleConvertDraft = async (draftId: string) => {
    setConvertingDraftId(draftId);
    setJobActionError(null);
    try {
      await apiClient.post(
        `/sonar/questArchetypeSuggestionDrafts/${draftId}/convert`,
        {}
      );
      if (selectedJobId) {
        await fetchDrafts(selectedJobId);
      }
    } catch (error) {
      console.error(
        'Failed to convert quest archetype suggestion draft',
        error
      );
      setJobActionError(
        extractApiErrorMessage(error, 'Failed to convert draft into archetype.')
      );
    } finally {
      setConvertingDraftId(null);
    }
  };

  const handleDeleteDraft = async (draftId: string) => {
    setDeletingDraftId(draftId);
    setJobActionError(null);
    try {
      await apiClient.delete(
        `/sonar/questArchetypeSuggestionDrafts/${draftId}`
      );
      setDrafts((current) => current.filter((draft) => draft.id !== draftId));
    } catch (error) {
      console.error('Failed to delete quest archetype suggestion draft', error);
      setJobActionError(
        extractApiErrorMessage(error, 'Failed to delete generated draft.')
      );
    } finally {
      setDeletingDraftId(null);
    }
  };

  const handleGeneratePreset = async () => {
    setGeneratingPreset(true);
    setJobActionError(null);
    setPresetStatusMessage(null);
    const requestedCount = Math.max(1, parseInt(form.count, 10) || 2);
    const presetZoneKindHint = presetZoneKind.trim() || form.zoneKind.trim();
    try {
      const preset = await apiClient.post<QuestArchetypeSuggestionPreset>(
        '/sonar/questArchetypeSuggestionJobs/generatePreset',
        {
          count: requestedCount,
          zoneKind: presetZoneKindHint,
          themePrompt: form.themePrompt.trim(),
          familyTags: parseTags(form.familyTagsText),
          familyMixTargets: buildFamilyMixTargetsPayload(form.familyMixTargets),
          characterTags: parseTags(form.characterTagsText),
          internalTags: parseTags(form.internalTagsText),
          requiredLocationArchetypeIds: form.requiredLocationArchetypeIds,
          requiredLocationMetadataTags: parseTags(
            form.requiredLocationMetadataTagsText
          ),
        }
      );
      const nextForm = buildGeneratorFormFromPreset(
        preset,
        requestedCount,
        presetZoneKindHint,
        form.yeetIt
      );
      setForm(nextForm);
      setPresetZoneKind(nextForm.zoneKind);
      setPresetStatusMessage('Loaded an LLM-generated preset.');
    } catch (error) {
      console.error('Failed to generate quest archetype preset', error);
      const fallbackForm = buildRandomPresetForm(
          zoneKinds as QuestPresetZoneKind[],
          locationArchetypes as QuestPresetLocationArchetype[],
          {
            preferredCount: requestedCount,
            preferredZoneKind: presetZoneKindHint,
            yeetIt: form.yeetIt,
          }
        );
      setForm(fallbackForm);
      setPresetZoneKind(fallbackForm.zoneKind);
      setPresetStatusMessage(
        'LLM preset generation was unavailable, so a local fallback preset was loaded.'
      );
    } finally {
      setGeneratingPreset(false);
    }
  };

  const handleQueuePresetJobsForZoneKinds = async () => {
    if (presetQueueZoneKinds.length === 0) {
      return;
    }
    setQueueingPresetZoneKinds(true);
    setJobActionError(null);
    setPresetStatusMessage(null);

    const createdJobs: QuestArchetypeSuggestionJob[] = [];
    const fallbackZoneKinds: string[] = [];
    const failedZoneKinds: string[] = [];

    for (const zoneKindSlug of presetQueueZoneKinds) {
      let queuedForm: GeneratorFormState;
      try {
        const preset = await apiClient.post<QuestArchetypeSuggestionPreset>(
          '/sonar/questArchetypeSuggestionJobs/generatePreset',
          {
            ...buildQuestArchetypeSuggestionJobPayload(form),
            count: requestedBatchCount,
            zoneKind: zoneKindSlug,
          }
        );
        queuedForm = buildGeneratorFormFromPreset(
          preset,
          requestedBatchCount,
          zoneKindSlug,
          form.yeetIt
        );
      } catch (error) {
        console.error(
          `Failed to generate LLM preset for zone kind ${zoneKindSlug}`,
          error
        );
        queuedForm = buildRandomPresetForm(
          zoneKinds as QuestPresetZoneKind[],
          locationArchetypes as QuestPresetLocationArchetype[],
          {
            preferredCount: requestedBatchCount,
            preferredZoneKind: zoneKindSlug,
            yeetIt: form.yeetIt,
          }
        );
        fallbackZoneKinds.push(zoneKindLabel(zoneKindSlug, zoneKindBySlug));
      }

      try {
        const created = await apiClient.post<QuestArchetypeSuggestionJob>(
          '/sonar/questArchetypeSuggestionJobs',
          buildQuestArchetypeSuggestionJobPayload(queuedForm)
        );
        createdJobs.push(created);
      } catch (error) {
        console.error(
          `Failed to queue preset quest archetype suggestion job for ${zoneKindSlug}`,
          error
        );
        failedZoneKinds.push(zoneKindLabel(zoneKindSlug, zoneKindBySlug));
      }
    }

    if (createdJobs.length > 0) {
      const newestFirst = [...createdJobs].reverse();
      setJobs((current) => [
        ...newestFirst,
        ...current.filter(
          (job) => !createdJobs.some((created) => created.id === job.id)
        ),
      ]);
      setSelectedJobId(newestFirst[0].id);
      setDrafts([]);
    }

    if (failedZoneKinds.length > 0 && createdJobs.length === 0) {
      setJobActionError(
        `Failed to queue preset jobs for ${failedZoneKinds.join(', ')}.`
      );
    } else {
      const messageParts = [
        `Queued ${createdJobs.length} preset ${form.yeetIt ? 'yeet' : 'draft'} job${
          createdJobs.length === 1 ? '' : 's'
        }.`,
      ];
      if (fallbackZoneKinds.length > 0) {
        messageParts.push(
          `Local fallback presets were used for ${fallbackZoneKinds.join(', ')}.`
        );
      }
      if (failedZoneKinds.length > 0) {
        messageParts.push(
          `Queueing still failed for ${failedZoneKinds.join(', ')}.`
        );
      }
      setPresetStatusMessage(messageParts.join(' '));
    }

    setQueueingPresetZoneKinds(false);
  };

  return (
    <div className="qa-theme">
      <div className="qa-shell">
        <header className="qa-hero">
          <div>
            <div className="qa-kicker">Questing</div>
            <h1 className="qa-title">Quest Archetype Generator</h1>
            <p className="qa-subtitle">
              Generate batches of draft archetype bundles, review node-by-node
              content, and convert the strongest ones into live quest
              archetypes.
            </p>
          </div>
        </header>

        <section
          className="qa-grid"
          style={{ gridTemplateColumns: '1.2fr 1fr' }}
        >
          <div className="qa-panel">
            <div className="qa-card-title">Queue Suggestion Job</div>
            <p className="qa-muted" style={{ marginTop: 8 }}>
              Use this to generate reusable quest archetype bundles with
              explicit node content, location metadata tags, and suggested
              template copy.
            </p>
            <div className="qa-form-grid" style={{ marginTop: 18 }}>
              <div className="qa-field">
                <div className="qa-label">Batch Size</div>
                <input
                  className="qa-input"
                  value={form.count}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      count: event.target.value,
                    }))
                  }
                  type="number"
                  min={1}
                  max={100}
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Zone Kind</div>
                <select
                  className="qa-input"
                  value={form.zoneKind}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      zoneKind: event.target.value,
                    }))
                  }
                >
                  <option value="">Any zone kind</option>
                  {zoneKinds.map((zoneKind) => (
                    <option key={zoneKind.id} value={zoneKind.slug}>
                      {zoneKind.name}
                    </option>
                  ))}
                </select>
              </div>
              <label
                className="qa-field"
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  justifyContent: 'flex-start',
                  paddingTop: 30,
                }}
              >
                <input
                  type="checkbox"
                  checked={form.yeetIt}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      yeetIt: event.target.checked,
                    }))
                  }
                />
                <div>
                  <div className="qa-label" style={{ marginBottom: 0 }}>
                    Yeet it
                  </div>
                  <div className="qa-muted" style={{ marginTop: 4 }}>
                    Auto-convert generated drafts into live quest archetypes
                    when the job completes.
                  </div>
                </div>
              </label>
              <div className="qa-field" style={{ gridColumn: '1 / -1' }}>
                <div className="qa-label">Theme Prompt</div>
                <textarea
                  className="qa-textarea"
                  value={form.themePrompt}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      themePrompt: event.target.value,
                    }))
                  }
                  rows={4}
                  placeholder="Urban food-logistics quests with criminal and civic variants."
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Family Tags</div>
                <input
                  className="qa-input"
                  value={form.familyTagsText}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      familyTagsText: event.target.value,
                    }))
                  }
                  placeholder="civic, criminal, occult"
                />
              </div>
              <div className="qa-field" style={{ gridColumn: '1 / -1' }}>
                <div className="qa-label">Family Mix Targets</div>
                <div
                  className="qa-muted"
                  style={{ marginTop: 6, marginBottom: 10 }}
                >
                  Set explicit minimum counts for the batch. Leave a family at 0
                  to make it optional.
                </div>
                <div
                  className="qa-form-grid"
                  style={{ gridTemplateColumns: 'repeat(2, minmax(0, 1fr))' }}
                >
                  {QUEST_FAMILY_OPTIONS.map((family) => (
                    <label key={family.slug} className="qa-field">
                      <div className="qa-meta" style={{ marginBottom: 6 }}>
                        {family.label}
                      </div>
                      <input
                        className="qa-input"
                        type="number"
                        min={0}
                        max={100}
                        value={form.familyMixTargets[family.slug] ?? '0'}
                        onChange={(event) =>
                          setForm((current) => ({
                            ...current,
                            familyMixTargets: {
                              ...current.familyMixTargets,
                              [family.slug]: event.target.value,
                            },
                          }))
                        }
                      />
                    </label>
                  ))}
                </div>
                <div className="qa-meta" style={{ marginTop: 10 }}>
                  Requested family slots: {requestedFamilyMixCount}
                </div>
              </div>
              <div className="qa-field">
                <div className="qa-label">Character Tags</div>
                <input
                  className="qa-input"
                  value={form.characterTagsText}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      characterTagsText: event.target.value,
                    }))
                  }
                  placeholder="quartermaster, merchant"
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Internal Tags</div>
                <input
                  className="qa-input"
                  value={form.internalTagsText}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      internalTagsText: event.target.value,
                    }))
                  }
                  placeholder="market, trade, food"
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Required Location Metadata Tags</div>
                <input
                  className="qa-input"
                  value={form.requiredLocationMetadataTagsText}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      requiredLocationMetadataTagsText: event.target.value,
                    }))
                  }
                  placeholder="market, alley, warehouse"
                />
              </div>
              <div className="qa-field" style={{ gridColumn: '1 / -1' }}>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    gap: 12,
                  }}
                >
                  <div className="qa-label">Required Location Archetypes</div>
                  {form.requiredLocationArchetypeIds.length > 0 && (
                    <button
                      type="button"
                      className="qa-button secondary"
                      onClick={() =>
                        setForm((current) => ({
                          ...current,
                          requiredLocationArchetypeIds: [],
                        }))
                      }
                    >
                      Clear All
                    </button>
                  )}
                </div>
                <div
                  className="qa-muted"
                  style={{ marginTop: 6, marginBottom: 10 }}
                >
                  Leave everything unchecked to allow flexible routing. If you
                  check archetypes here, every generated draft should include
                  each one at least once.
                </div>
                <div
                  className="qa-tree"
                  style={{ maxHeight: 220, overflowY: 'auto', padding: 8 }}
                >
                  {locationArchetypes.length === 0 ? (
                    <div className="qa-node-card">
                      <div className="qa-muted">
                        No location archetypes available.
                      </div>
                    </div>
                  ) : (
                    locationArchetypes
                      .slice()
                      .sort((left, right) =>
                        left.name.localeCompare(right.name)
                      )
                      .map((archetype) => {
                        const checked =
                          form.requiredLocationArchetypeIds.includes(
                            archetype.id
                          );
                        return (
                          <label
                            key={archetype.id}
                            className="qa-node-card"
                            style={{
                              display: 'flex',
                              alignItems: 'center',
                              gap: 10,
                              cursor: 'pointer',
                            }}
                          >
                            <input
                              type="checkbox"
                              checked={checked}
                              onChange={(event) =>
                                setForm((current) => ({
                                  ...current,
                                  requiredLocationArchetypeIds: event.target
                                    .checked
                                    ? [
                                        ...current.requiredLocationArchetypeIds,
                                        archetype.id,
                                      ]
                                    : current.requiredLocationArchetypeIds.filter(
                                        (id) => id !== archetype.id
                                      ),
                                }))
                              }
                            />
                            <div>
                              <div
                                className="qa-meta"
                                style={{ fontWeight: 600 }}
                              >
                                {archetype.name}
                              </div>
                              <div
                                className="qa-muted"
                                style={{ marginTop: 4 }}
                              >
                                {(archetype.includedTypes ?? [])
                                  .slice(0, 4)
                                  .join(', ') || 'No included place types'}
                              </div>
                            </div>
                          </label>
                        );
                      })
                  )}
                </div>
              </div>
            </div>

            <div
              className="qa-form-grid"
              style={{
                marginTop: 18,
                gridTemplateColumns: 'minmax(0, 260px) auto auto auto',
                alignItems: 'end',
              }}
            >
              <label className="qa-field" style={{ marginBottom: 0 }}>
                <div className="qa-label">Preset Zone Kind</div>
                <select
                  className="qa-input"
                  value={presetZoneKind}
                  onChange={(event) => setPresetZoneKind(event.target.value)}
                >
                  <option value="">Use current form zone kind</option>
                  {zoneKinds.map((zoneKind) => (
                    <option key={`preset-zone-kind-${zoneKind.id}`} value={zoneKind.slug}>
                      {zoneKind.name}
                    </option>
                  ))}
                </select>
              </label>
              <button
                className="qa-btn qa-btn-outline"
                onClick={() => void handleGeneratePreset()}
                type="button"
                disabled={generatingPreset}
              >
                {generatingPreset ? 'Generating Preset...' : 'Suggest Preset'}
              </button>
              <button
                className="qa-btn qa-btn-primary"
                onClick={() => void handleQueueJob()}
                disabled={
                  queueing ||
                  requestedFamilyMixCount >
                    Math.max(1, parseInt(form.count, 10) || 1)
                }
              >
                {queueing
                  ? 'Queueing...'
                  : form.yeetIt
                    ? 'Queue Yeet Job'
                    : 'Queue Draft Job'}
              </button>
              <button
                className="qa-btn qa-btn-outline"
                onClick={() => void fetchJobs()}
                disabled={loadingJobs}
              >
                {loadingJobs ? 'Refreshing...' : 'Refresh Jobs'}
              </button>
            </div>

            <div className="qa-field" style={{ marginTop: 18 }}>
              <div className="qa-label">Preset Queue Zone Kinds</div>
              <div
                className="qa-muted"
                style={{ marginTop: 6, marginBottom: 10 }}
              >
                Select multiple zone kinds to generate one preset-backed{' '}
                {form.yeetIt ? 'yeet' : 'draft'} job per zone kind using the
                current form as hints.
              </div>
              <div
                className="qa-tree"
                style={{ maxHeight: 180, overflowY: 'auto', padding: 8 }}
              >
                {zoneKinds.length === 0 ? (
                  <div className="qa-node-card">
                    <div className="qa-muted">No zone kinds available.</div>
                  </div>
                ) : (
                  zoneKinds.map((zoneKind) => {
                    const checked = presetQueueZoneKinds.includes(zoneKind.slug);
                    return (
                      <label
                        key={`preset-queue-zone-kind-${zoneKind.id}`}
                        className="qa-node-card"
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: 10,
                          cursor: 'pointer',
                        }}
                      >
                        <input
                          type="checkbox"
                          checked={checked}
                          onChange={(event) =>
                            setPresetQueueZoneKinds((current) =>
                              event.target.checked
                                ? [...current, zoneKind.slug]
                                : current.filter((slug) => slug !== zoneKind.slug)
                            )
                          }
                        />
                        <div>
                          <div
                            className="qa-meta"
                            style={{ fontWeight: 600 }}
                          >
                            {zoneKind.name}
                          </div>
                          {zoneKind.description ? (
                            <div
                              className="qa-muted"
                              style={{ marginTop: 4 }}
                            >
                              {zoneKind.description}
                            </div>
                          ) : null}
                        </div>
                      </label>
                    );
                  })
                )}
              </div>
              <div
                className="qa-actions"
                style={{ marginTop: 12, alignItems: 'center' }}
              >
                <span className="qa-muted">
                  Selected: {presetQueueZoneKindLabels.join(', ') || 'none'}
                </span>
                {presetQueueZoneKinds.length > 0 && (
                  <button
                    type="button"
                    className="qa-button secondary"
                    onClick={() => setPresetQueueZoneKinds([])}
                  >
                    Clear Selection
                  </button>
                )}
                {zoneKinds.length > 0 &&
                  presetQueueZoneKinds.length !== zoneKinds.length && (
                    <button
                      type="button"
                      className="qa-button secondary"
                      onClick={() =>
                        setPresetQueueZoneKinds(zoneKinds.map((zoneKind) => zoneKind.slug))
                      }
                    >
                      Select All
                    </button>
                  )}
                <button
                  className="qa-btn qa-btn-primary"
                  type="button"
                  onClick={() => void handleQueuePresetJobsForZoneKinds()}
                  disabled={
                    queueingPresetZoneKinds ||
                    presetQueueZoneKinds.length === 0 ||
                    requestedFamilyMixCount > requestedBatchCount
                  }
                >
                  {queueingPresetZoneKinds
                    ? 'Queueing Preset Jobs...'
                    : `Queue ${presetQueueZoneKinds.length || ''} Preset ${form.yeetIt ? 'Yeet' : 'Draft'} Job${
                        presetQueueZoneKinds.length === 1 ? '' : 's'
                      }`}
                </button>
              </div>
            </div>

            <div className="qa-actions" style={{ marginTop: 18 }}>
              <span className="qa-muted">
                Preset batch size follows the current form count.
              </span>
            </div>

            {presetStatusMessage && (
              <div className="qa-chip muted" style={{ marginTop: 14 }}>
                {presetStatusMessage}
              </div>
            )}

            {(jobActionError || pageError) && (
              <div className="qa-chip danger" style={{ marginTop: 14 }}>
                {jobActionError || pageError}
              </div>
            )}
            {requestedFamilyMixCount >
              Math.max(1, parseInt(form.count, 10) || 1) && (
              <div className="qa-chip danger" style={{ marginTop: 14 }}>
                Family mix targets cannot exceed the requested batch size.
              </div>
            )}
          </div>

          <div className="qa-panel">
            <div className="qa-card-title">Recent Jobs</div>
            <p className="qa-muted" style={{ marginTop: 8 }}>
              Choose a job to inspect its generated archetype bundles.
            </p>
            <div className="qa-tree" style={{ marginTop: 16 }}>
              {jobs.length === 0 ? (
                <div className="qa-node-card">
                  <div className="qa-muted">No suggestion jobs yet.</div>
                </div>
              ) : (
                jobs.map((job) => (
                  <button
                    key={job.id}
                    type="button"
                    className="qa-node-card"
                    onClick={() => setSelectedJobId(job.id)}
                    style={{
                      textAlign: 'left',
                      border:
                        job.id === selectedJobId
                          ? '1px solid rgba(244, 180, 26, 0.8)'
                          : undefined,
                    }}
                  >
                    <div className="qa-card-header" style={{ marginBottom: 8 }}>
                      <div>
                        <div className="qa-card-title" style={{ fontSize: 16 }}>
                          Job {job.id.slice(0, 8)}...
                        </div>
                        <div className="qa-meta">
                          {formatDate(job.createdAt)}
                        </div>
                      </div>
                      <div className={statusChipClass(job.status)}>
                        {formatStatus(job.status)}
                      </div>
                    </div>
                    <div className="qa-meta">
                      {job.yeetIt ? 'Live archetypes' : 'Drafts'}:{' '}
                      {job.createdCount}/{job.count}
                    </div>
                    <div className="qa-meta" style={{ marginTop: 8 }}>
                      Mode: {job.yeetIt ? 'Yeet' : 'Draft'}
                    </div>
                    {job.zoneKind?.trim() && (
                      <div className="qa-meta" style={{ marginTop: 8 }}>
                        Zone kind: {zoneKindLabel(job.zoneKind, zoneKindBySlug)}
                      </div>
                    )}
                    {job.familyMixTargets &&
                      Object.keys(job.familyMixTargets).length > 0 && (
                        <div className="qa-meta" style={{ marginTop: 8 }}>
                          Family mix: {formatFamilyMixTargets(job.familyMixTargets)}
                        </div>
                      )}
                    {job.requiredLocationArchetypeIds?.length > 0 && (
                      <div className="qa-meta" style={{ marginTop: 8 }}>
                        Required archetypes:{' '}
                        {job.requiredLocationArchetypeIds
                          .map(
                            (id) =>
                              locationArchetypeNamesById.get(id) ??
                              `Unknown archetype (${id.slice(0, 8)}...)`
                          )
                          .join(', ')}
                      </div>
                    )}
                    {job.themePrompt && (
                      <div className="qa-muted" style={{ marginTop: 8 }}>
                        {job.themePrompt}
                      </div>
                    )}
                    {job.errorMessage?.trim() && (
                      <div
                        className="qa-chip danger"
                        style={{
                          marginTop: 10,
                          display: 'block',
                          whiteSpace: 'normal',
                        }}
                      >
                        Failed reason: {job.errorMessage}
                      </div>
                    )}
                  </button>
                ))
              )}
            </div>
          </div>
        </section>

        <section className="qa-panel" style={{ marginTop: 24 }}>
          <div className="qa-card-header">
            <div>
              <div className="qa-card-title">Generated Drafts</div>
              <div className="qa-meta">
                {selectedJob
                  ? `Showing ${selectedJob.yeetIt ? 'generated results' : 'drafts'} for job ${selectedJob.id.slice(0, 8)}...`
                  : 'Select a job to inspect its generated drafts.'}
              </div>
            </div>
            {selectedJob && (
              <div className={statusChipClass(selectedJob.status)}>
                {formatStatus(selectedJob.status)}
              </div>
            )}
          </div>

          {loadingDrafts && (
            <div className="qa-muted" style={{ marginTop: 16 }}>
              Loading drafts...
            </div>
          )}

          {!loadingDrafts && selectedJob?.errorMessage?.trim() && (
            <div
              className="qa-chip danger"
              style={{
                marginTop: 16,
                display: 'block',
                whiteSpace: 'normal',
              }}
            >
              Failed reason: {selectedJob.errorMessage}
            </div>
          )}

          {!loadingDrafts && selectedJob && drafts.length === 0 && (
            <div className="qa-muted" style={{ marginTop: 16 }}>
              {isPendingStatus(selectedJob.status)
                ? selectedJob.yeetIt
                  ? 'This yeet job is still generating and will auto-convert into live quest archetypes when it finishes.'
                  : 'This job is still generating drafts.'
                : selectedJob.yeetIt
                  ? 'This yeet job did not leave any draft history behind.'
                  : 'No drafts were generated for this job.'}
            </div>
          )}

          {!loadingDrafts && drafts.length > 0 && (
            <div className="qa-tree" style={{ marginTop: 18 }}>
              {selectedJob?.yeetIt && (
                <div className="qa-chip success" style={{ marginBottom: 12 }}>
                  This yeet job auto-converted its generated drafts into live
                  quest archetypes. The converted records are shown below for
                  review.
                </div>
              )}
              {drafts.map((draft) => {
                const draftNodes = draftNodesForReview(draft);
                const failureBranchCount = draftFailureBranchCount(draftNodes);

                return (
                  <article key={draft.id} className="qa-node-card">
                    <div className="qa-card-header">
                      <div>
                        <div className="qa-card-title">{draft.name}</div>
                        <div className="qa-meta" style={{ marginTop: 4 }}>
                          {draft.hook || 'No hook provided'}
                        </div>
                      </div>
                      <div className={statusChipClass(draft.status)}>
                        {formatStatus(draft.status)}
                      </div>
                    </div>

                    <div className="qa-stat-grid" style={{ marginTop: 16 }}>
                      <div className="qa-stat">
                        <div className="qa-stat-label">Difficulty</div>
                        <div className="qa-stat-value">
                          {draft.difficultyMode} / {draft.difficulty}
                        </div>
                      </div>
                      <div className="qa-stat">
                        <div className="qa-stat-label">Encounter Level</div>
                        <div className="qa-stat-value">
                          {draft.monsterEncounterTargetLevel}
                        </div>
                      </div>
                      <div className="qa-stat">
                        <div className="qa-stat-label">Quest Nodes</div>
                        <div className="qa-stat-value">
                          {draftNodes.length}
                        </div>
                      </div>
                      <div className="qa-stat">
                        <div className="qa-stat-label">Failure Paths</div>
                        <div className="qa-stat-value">
                          {failureBranchCount}
                        </div>
                      </div>
                    </div>

                    <div style={{ marginTop: 18 }}>
                      {draft.zoneKind?.trim() && (
                        <div className="qa-meta" style={{ marginBottom: 8 }}>
                          Zone kind:{' '}
                          {zoneKindLabel(draft.zoneKind, zoneKindBySlug)}
                        </div>
                      )}
                      {selectedJobRequiredLocationArchetypes.length > 0 && (
                        <div className="qa-meta" style={{ marginBottom: 8 }}>
                          Required archetypes:{' '}
                          {selectedJobRequiredLocationArchetypes.join(', ')}
                        </div>
                      )}
                      {selectedJob?.familyMixTargets &&
                        Object.keys(selectedJob.familyMixTargets).length > 0 && (
                          <div className="qa-meta" style={{ marginBottom: 8 }}>
                            Family mix:{' '}
                            {formatFamilyMixTargets(selectedJob.familyMixTargets)}
                          </div>
                        )}
                      <div className="qa-meta">
                        Character tags:{' '}
                        {(draft.characterTags ?? []).join(', ') || 'none'}
                      </div>
                      <div className="qa-meta" style={{ marginTop: 6 }}>
                        Internal tags:{' '}
                        {(draft.internalTags ?? []).join(', ') || 'none'}
                      </div>
                      <p style={{ marginTop: 12 }}>{draft.description}</p>
                    </div>

                    {draft.acceptanceDialogue?.length > 0 && (
                      <div style={{ marginTop: 18 }}>
                        <div className="qa-stat-label">Acceptance Dialogue</div>
                        <div className="qa-tree" style={{ marginTop: 8 }}>
                          {draft.acceptanceDialogue.map((line, index) => (
                            <div
                              key={`${draft.id}-dialogue-${index}`}
                              className="qa-node-card"
                            >
                              {line}
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    <div style={{ marginTop: 18 }}>
                      <div className="qa-stat-label">Node Plan</div>
                      <div className="qa-tree" style={{ marginTop: 10 }}>
                        {draftNodes.map((node, index) => (
                          <div
                            key={`${draft.id}-node-${node.nodeKey}-${index}`}
                            className="qa-node-card"
                          >
                            <div
                              className="qa-card-title"
                              style={{ fontSize: 15 }}
                            >
                              Node {index + 1}: {node.source} {node.content}
                            </div>
                            <div className="qa-meta" style={{ marginTop: 8 }}>
                              Node key: {node.nodeKey}
                            </div>
                            <div className="qa-meta" style={{ marginTop: 6 }}>
                              Location concept: {node.locationConcept}
                            </div>
                            {node.locationArchetypeName && (
                              <div className="qa-meta" style={{ marginTop: 6 }}>
                                Location archetype: {node.locationArchetypeName}
                              </div>
                            )}
                            {node.distanceMeters != null && (
                              <div className="qa-meta" style={{ marginTop: 6 }}>
                                Distance: {node.distanceMeters}m
                              </div>
                            )}
                            <div className="qa-meta" style={{ marginTop: 6 }}>
                              Metadata tags:{' '}
                              {(node.locationMetadataTags ?? []).join(', ')}
                            </div>
                            <div className="qa-meta" style={{ marginTop: 6 }}>
                              Template concept: {node.templateConcept}
                            </div>

                            {node.content === 'challenge' && (
                              <div style={{ marginTop: 12 }}>
                                <div className="qa-stat-label">
                                  Challenge Template Draft
                                </div>
                                <div className="qa-meta" style={{ marginTop: 6 }}>
                                  Question: {node.challengeQuestion}
                                </div>
                                <div className="qa-meta" style={{ marginTop: 6 }}>
                                  Description: {node.challengeDescription}
                                </div>
                                <div className="qa-meta" style={{ marginTop: 6 }}>
                                  Submission:{' '}
                                  {node.challengeSubmissionType || 'photo'}
                                </div>
                              </div>
                            )}

                            {node.content === 'scenario' && (
                              <div style={{ marginTop: 12 }}>
                                <div className="qa-stat-label">
                                  Scenario Template Draft
                                </div>
                                <div className="qa-meta" style={{ marginTop: 6 }}>
                                  Prompt: {node.scenarioPrompt}
                                </div>
                                <div className="qa-meta" style={{ marginTop: 6 }}>
                                  Beats:{' '}
                                  {(node.scenarioBeats ?? []).join(', ') ||
                                    'none'}
                                </div>
                              </div>
                            )}

                            {node.content === 'exposition' && (
                              <div style={{ marginTop: 12 }}>
                                <div className="qa-stat-label">
                                  Exposition Template Draft
                                </div>
                                <div className="qa-meta" style={{ marginTop: 6 }}>
                                  Title: {node.expositionTitle || 'none'}
                                </div>
                                <div className="qa-meta" style={{ marginTop: 6 }}>
                                  Description:{' '}
                                  {node.expositionDescription || 'none'}
                                </div>
                                <div className="qa-meta" style={{ marginTop: 6 }}>
                                  Speaker:{' '}
                                  {node.expositionSpeakerName || 'none'}
                                </div>
                                <div className="qa-meta" style={{ marginTop: 6 }}>
                                  Portrait URL:{' '}
                                  {node.expositionPortraitUrl || 'none'}
                                </div>
                                <div className="qa-meta" style={{ marginTop: 6 }}>
                                  Dialogue:{' '}
                                  {(node.expositionDialogue ?? []).length > 0
                                    ? (node.expositionDialogue ?? []).join(' / ')
                                    : 'none'}
                                </div>
                              </div>
                            )}

                            {node.content === 'monster' && (
                              <div style={{ marginTop: 12 }}>
                                <div className="qa-stat-label">
                                  Monster Encounter Draft
                                </div>
                                <div className="qa-meta" style={{ marginTop: 6 }}>
                                  Templates:{' '}
                                  {(node.monsterTemplateNames ?? []).join(
                                    ', '
                                  ) || 'none'}
                                </div>
                                <div className="qa-meta" style={{ marginTop: 6 }}>
                                  Tone:{' '}
                                  {(node.encounterTone ?? []).join(', ') ||
                                    'none'}
                                </div>
                              </div>
                            )}

                            {(node.outcomes ?? []).length > 0 ? (
                              <div style={{ marginTop: 12 }}>
                                <div className="qa-stat-label">Transitions</div>
                                <div className="qa-tree" style={{ marginTop: 8 }}>
                                  {(node.outcomes ?? []).map(
                                    (outcome, outcomeIndex) => (
                                      <div
                                        key={`${draft.id}-${node.nodeKey}-outcome-${outcomeIndex}`}
                                        className="qa-chip accent"
                                      >
                                        {formatOutcomeLabel(outcome.outcome)}{' '}
                                        -&gt; {outcome.nextNodeKey || 'end'}
                                      </div>
                                    )
                                  )}
                                </div>
                              </div>
                            ) : (
                              <div className="qa-meta" style={{ marginTop: 12 }}>
                                Transitions: this node can end the quest path.
                              </div>
                            )}

                            {(node.potentialContent ?? []).length > 0 && (
                              <div style={{ marginTop: 12 }}>
                                <div className="qa-stat-label">
                                  Potential Content
                                </div>
                                <div className="qa-tree" style={{ marginTop: 8 }}>
                                  {node.potentialContent.map(
                                    (item, ideaIndex) => (
                                      <div
                                        key={`${draft.id}-${node.nodeKey}-idea-${ideaIndex}`}
                                        className="qa-node-card"
                                      >
                                        {item}
                                      </div>
                                    )
                                  )}
                                </div>
                              </div>
                            )}
                          </div>
                        ))}
                      </div>
                    </div>

                    {draft.whyThisScales && (
                      <div style={{ marginTop: 18 }}>
                        <div className="qa-stat-label">Why This Scales</div>
                        <p style={{ marginTop: 8 }}>{draft.whyThisScales}</p>
                      </div>
                    )}

                    {(draft.warnings ?? []).length > 0 && (
                      <div style={{ marginTop: 18 }}>
                        <div className="qa-stat-label">Warnings</div>
                        <div className="qa-tree" style={{ marginTop: 8 }}>
                          {draft.warnings.map((warning, index) => (
                            <div
                              key={`${draft.id}-warning-${index}`}
                              className="qa-chip danger"
                            >
                              {warning}
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                    {(draft.challengeTemplateSeeds?.length ||
                      draft.scenarioTemplateSeeds?.length ||
                      draft.monsterTemplateSeeds?.length) && (
                      <div style={{ marginTop: 18 }}>
                        <div className="qa-stat-label">Seed Notes</div>
                        {draft.challengeTemplateSeeds?.length ? (
                          <div className="qa-meta" style={{ marginTop: 8 }}>
                            Challenge seeds:{' '}
                            {draft.challengeTemplateSeeds.join(' | ')}
                          </div>
                        ) : null}
                        {draft.scenarioTemplateSeeds?.length ? (
                          <div className="qa-meta" style={{ marginTop: 8 }}>
                            Scenario seeds:{' '}
                            {draft.scenarioTemplateSeeds.join(' | ')}
                          </div>
                        ) : null}
                        {draft.monsterTemplateSeeds?.length ? (
                          <div className="qa-meta" style={{ marginTop: 8 }}>
                            Monster seeds:{' '}
                            {draft.monsterTemplateSeeds.join(' | ')}
                          </div>
                        ) : null}
                      </div>
                    )}

                    <div className="qa-actions" style={{ marginTop: 18 }}>
                      {draft.questArchetypeId ? (
                        <Link
                          to="/quest-archetypes"
                          className="qa-btn qa-btn-outline"
                        >
                          Open Quest Archetypes
                        </Link>
                      ) : (
                        <button
                          type="button"
                          className="qa-btn qa-btn-primary"
                          onClick={() => void handleConvertDraft(draft.id)}
                          disabled={convertingDraftId === draft.id}
                        >
                          {convertingDraftId === draft.id
                            ? 'Converting...'
                            : 'Convert to Archetype'}
                        </button>
                      )}
                      {!draft.questArchetypeId && (
                        <button
                          type="button"
                          className="qa-btn qa-btn-danger"
                          onClick={() => void handleDeleteDraft(draft.id)}
                          disabled={deletingDraftId === draft.id}
                        >
                          {deletingDraftId === draft.id
                            ? 'Deleting...'
                            : 'Delete Draft'}
                        </button>
                      )}
                    </div>
                  </article>
                );
              })}
            </div>
          )}
        </section>
      </div>
    </div>
  );
};

export default QuestArchetypeGenerator;
