import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { ZoneAdminSummary, ZoneKind } from '@poltergeist/types';
import { Link } from 'react-router-dom';

type ZoneKindRatioField = {
  key:
    | 'placeCountRatio'
    | 'monsterCountRatio'
    | 'bossEncounterCountRatio'
    | 'raidEncounterCountRatio'
    | 'inputEncounterCountRatio'
    | 'optionEncounterCountRatio'
    | 'treasureChestCountRatio'
    | 'healingFountainCountRatio'
    | 'herbalismResourceCountRatio'
    | 'miningResourceCountRatio';
  label: string;
  description: string;
};

type ZoneKindFormState = {
  name: string;
  slug: string;
  description: string;
  overlayColor: string;
  placeCountRatio: string;
  monsterCountRatio: string;
  bossEncounterCountRatio: string;
  raidEncounterCountRatio: string;
  inputEncounterCountRatio: string;
  optionEncounterCountRatio: string;
  treasureChestCountRatio: string;
  healingFountainCountRatio: string;
  herbalismResourceCountRatio: string;
  miningResourceCountRatio: string;
};

type ZoneKindBackfillResult = {
  contentType: string;
  missingCount: number;
  assignedCount: number;
  ambiguousCount: number;
  skippedCount: number;
};

type ZoneKindBackfillSummary = {
  results: ZoneKindBackfillResult[];
  missingCount: number;
  assignedCount: number;
  ambiguousCount: number;
  skippedCount: number;
};

type ZoneKindBackfillStatus = {
  jobId: string;
  status: 'queued' | 'in_progress' | 'completed' | 'failed';
  summary: ZoneKindBackfillSummary;
  error?: string;
  queuedAt?: string;
  startedAt?: string;
  completedAt?: string;
  updatedAt: string;
};

type ZoneKindPayload = Omit<
  ZoneKind,
  | 'id'
  | 'createdAt'
  | 'updatedAt'
  | 'patternTileUrl'
  | 'patternTilePrompt'
  | 'patternTileGenerationStatus'
  | 'patternTileGenerationError'
>;

const ratioFields: ZoneKindRatioField[] = [
  {
    key: 'placeCountRatio',
    label: 'Places',
    description: 'POIs and place-led encounters',
  },
  {
    key: 'monsterCountRatio',
    label: 'Monsters',
    description: 'Standard encounters',
  },
  {
    key: 'bossEncounterCountRatio',
    label: 'Bosses',
    description: 'Boss encounter density',
  },
  {
    key: 'raidEncounterCountRatio',
    label: 'Raids',
    description: 'Group raid encounter density',
  },
  {
    key: 'inputEncounterCountRatio',
    label: 'Input scenarios',
    description: 'Open-ended scenario prompts',
  },
  {
    key: 'optionEncounterCountRatio',
    label: 'Option scenarios',
    description: 'Choice-driven scenarios',
  },
  {
    key: 'treasureChestCountRatio',
    label: 'Treasure chests',
    description: 'Chest reward density',
  },
  {
    key: 'healingFountainCountRatio',
    label: 'Healing fountains',
    description: 'Restorative nodes',
  },
  {
    key: 'herbalismResourceCountRatio',
    label: 'Herbalism',
    description: 'Herb and forage node density',
  },
  {
    key: 'miningResourceCountRatio',
    label: 'Mining',
    description: 'Ore and mineral node density',
  },
];

const defaultZoneKindOverlayColor = '#5f7d68';

const zoneKindBackfillContentLabels: Record<string, string> = {
  quests: 'Quests',
  challenges: 'Challenges',
  scenarios: 'Scenarios',
  expositions: 'Expositions',
  monsters: 'Monsters',
  monster_encounters: 'Monster Encounters',
  treasure_chests: 'Treasure Chests',
  healing_fountains: 'Healing Fountains',
  resources: 'Resources',
  movement_patterns: 'Movement Patterns',
  points_of_interest: 'Points of Interest',
  inventory_items: 'Inventory Items',
};

const isZoneKindBackfillPending = (status?: string | null) =>
  status === 'queued' || status === 'in_progress';

const formatZoneKindBackfillStatus = (status?: string | null) => {
  switch (status) {
    case 'queued':
      return 'Queued';
    case 'in_progress':
      return 'Running';
    case 'completed':
      return 'Completed';
    case 'failed':
      return 'Failed';
    default:
      return 'Idle';
  }
};

const isZoneKindPatternTilePending = (status?: string | null) =>
  status === 'queued' || status === 'in_progress';

const formatZoneKindPatternTileStatus = (status?: string | null) => {
  switch (status) {
    case 'queued':
      return 'Queued';
    case 'in_progress':
      return 'Generating';
    case 'complete':
      return 'Ready';
    case 'failed':
      return 'Failed';
    default:
      return 'Idle';
  }
};

const zoneKindPatternMotifs = (slug: string) => {
  switch (normalizeSlugDraft(slug)) {
    case 'forest':
      return 'leaf clusters, canopy blotches, branch forks, trail scratches';
    case 'swamp':
      return 'reed strokes, puddle curves, marsh ripples, hanging moss';
    case 'volcanic':
      return 'lava cracks, ember seams, ash flecks, broken magma lines';
    case 'graveyard':
      return 'worn stone hash marks, grave glyphs, crosshatched weathering';
    case 'desert':
      return 'wind-carved dune lines, grit speckle, drifting wave contours';
    case 'temple-grounds':
      return 'sacred rings, shrine geometry, halo lines, ceremonial inlay';
    case 'city':
      return 'street grids, masonry blocks, alley runs, civic linework';
    case 'industrial':
      return 'rivet grids, pipe runs, hazard striping, forged plate seams';
    case 'farmland':
      return 'furrow stripes, field parcels, stitched paths, crop rows';
    case 'academy':
      return 'arcane circles, star-compass marks, inked diagram lines';
    case 'village':
      return 'cottage roof rhythms, lantern marks, fence lines, paths';
    case 'badlands':
      return 'sun-baked fractures, mesa ridges, eroded gullies';
    case 'highlands':
      return 'wind-swept contours, cairn-like marks, ridge bands';
    case 'mountain':
      return 'rock strata, sharp peak silhouettes, mineral seams';
    case 'ruins':
      return 'cracked tiles, broken arch fragments, chipped stone patterns';
    default:
      return 'subtle fantasy map texture marks, organic linework, exploratory symbols';
  }
};

const zoneKindPatternCueLabels = (zoneKind: ZoneKind) => {
  const cues = [
    { label: 'place-rich', value: zoneKind.placeCountRatio },
    { label: 'monster-heavy', value: zoneKind.monsterCountRatio },
    { label: 'boss-dangerous', value: zoneKind.bossEncounterCountRatio },
    { label: 'raid-heavy', value: zoneKind.raidEncounterCountRatio },
    {
      label: 'scenario-rich',
      value: Math.max(
        zoneKind.inputEncounterCountRatio,
        zoneKind.optionEncounterCountRatio
      ),
    },
    { label: 'treasure-rich', value: zoneKind.treasureChestCountRatio },
    { label: 'restorative', value: zoneKind.healingFountainCountRatio },
    {
      label: 'herbalism-rich',
      value: zoneKind.herbalismResourceCountRatio,
    },
    { label: 'mining-rich', value: zoneKind.miningResourceCountRatio },
  ]
    .sort((left, right) => right.value - left.value)
    .filter((entry) => entry.value > 1.05)
    .slice(0, 3)
    .map((entry) => entry.label);

  return cues.length > 0 ? cues : ['mixed frontier'];
};

const buildDefaultZoneKindPatternPrompt = (zoneKind: ZoneKind) => {
  const name = zoneKind.name.trim() || 'Frontier';
  const slug = normalizeSlugDraft(zoneKind.slug || zoneKind.name);
  const description =
    zoneKind.description.trim() ||
    'A fantasy zone texture used as a subtle map overlay.';
  const overlayColor =
    normalizeOverlayColorDraft(zoneKind.overlayColor) ||
    defaultZoneKindOverlayColor;

  return `Create a seamless repeating square texture tile for a fantasy RPG world map overlay.

Zone kind:
- name: ${name}
- slug: ${slug}
- description: ${description}
- dominant gameplay cues: ${zoneKindPatternCueLabels(zoneKind).join(', ')}
- palette anchor: ${overlayColor}
- motif direction: ${zoneKindPatternMotifs(slug)}

Requirements:
- The tile must repeat seamlessly on all four edges.
- This is not a full scene, landscape illustration, or diorama. It should read like an ornamental map texture.
- Use a subtle, game-ready pattern that can sit on top of a watercolor fantasy map without overwhelming it.
- Prefer transparent or near-transparent negative space between marks so the basemap can still show through.
- Keep the motifs medium-scale and legible when repeated across a polygon.
- No border, no frame, no text, no logos, no single centered subject.
- Square composition only.
- Top-down graphic texture language, never perspective or isometric.
- Fantasy RPG tone, handcrafted, slightly stylized, polished, tasteful.
- Avoid photorealism.`;
};

const resolveZoneKindPatternPrompt = (
  zoneKind: ZoneKind,
  draft?: string | null
) => {
  const nextDraft = draft?.trim();
  if (nextDraft) {
    return nextDraft;
  }
  const savedPrompt = zoneKind.patternTilePrompt?.trim();
  if (savedPrompt) {
    return savedPrompt;
  }
  return buildDefaultZoneKindPatternPrompt(zoneKind);
};

const emptyForm = (): ZoneKindFormState => ({
  name: '',
  slug: '',
  description: '',
  overlayColor: defaultZoneKindOverlayColor,
  placeCountRatio: '1',
  monsterCountRatio: '1',
  bossEncounterCountRatio: '1',
  raidEncounterCountRatio: '1',
  inputEncounterCountRatio: '1',
  optionEncounterCountRatio: '1',
  treasureChestCountRatio: '1',
  healingFountainCountRatio: '1',
  herbalismResourceCountRatio: '1',
  miningResourceCountRatio: '1',
});

const normalizeSlugDraft = (value: string) =>
  value
    .trim()
    .toLowerCase()
    .replace(/[_\s]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');

const normalizeOverlayColorDraft = (value?: string | null) => {
  const trimmed = (value || '').trim().toLowerCase();
  if (!trimmed) {
    return '';
  }
  const normalized = trimmed.startsWith('#') ? trimmed : `#${trimmed}`;
  return /^#[0-9a-f]{6}$/.test(normalized) ? normalized : '';
};

const formatRatio = (value: number) =>
  `${value.toFixed(Number.isInteger(value) ? 1 : 2)}x`;

const zoneSearchText = (zone: ZoneAdminSummary) =>
  [
    zone.name,
    zone.description,
    zone.kind,
    zone.importMetroName || '',
    ...(zone.internalTags || []),
  ]
    .join(' ')
    .toLowerCase();

const formFromZoneKind = (zoneKind: ZoneKind): ZoneKindFormState => ({
  name: zoneKind.name,
  slug: zoneKind.slug,
  description: zoneKind.description || '',
  overlayColor:
    normalizeOverlayColorDraft(zoneKind.overlayColor) ||
    defaultZoneKindOverlayColor,
  placeCountRatio: String(zoneKind.placeCountRatio ?? 1),
  monsterCountRatio: String(zoneKind.monsterCountRatio ?? 1),
  bossEncounterCountRatio: String(zoneKind.bossEncounterCountRatio ?? 1),
  raidEncounterCountRatio: String(zoneKind.raidEncounterCountRatio ?? 1),
  inputEncounterCountRatio: String(zoneKind.inputEncounterCountRatio ?? 1),
  optionEncounterCountRatio: String(zoneKind.optionEncounterCountRatio ?? 1),
  treasureChestCountRatio: String(zoneKind.treasureChestCountRatio ?? 1),
  healingFountainCountRatio: String(zoneKind.healingFountainCountRatio ?? 1),
  herbalismResourceCountRatio: String(
    zoneKind.herbalismResourceCountRatio ?? zoneKind.resourceCountRatio ?? 1
  ),
  miningResourceCountRatio: String(
    zoneKind.miningResourceCountRatio ?? zoneKind.resourceCountRatio ?? 1
  ),
});

const parseZoneKindForm = (
  form: ZoneKindFormState
): {
  payload?: ZoneKindPayload;
  error?: string;
} => {
  const name = form.name.trim();
  if (!name) {
    return { error: 'Name is required.' };
  }

  const payload: ZoneKindPayload = {
    name,
    slug: normalizeSlugDraft(form.slug || name),
    description: form.description.trim(),
    overlayColor: '',
    placeCountRatio: 1,
    monsterCountRatio: 1,
    bossEncounterCountRatio: 1,
    raidEncounterCountRatio: 1,
    inputEncounterCountRatio: 1,
    optionEncounterCountRatio: 1,
    treasureChestCountRatio: 1,
    healingFountainCountRatio: 1,
    herbalismResourceCountRatio: 1,
    miningResourceCountRatio: 1,
    resourceCountRatio: 1,
  };

  const overlayColor =
    normalizeOverlayColorDraft(form.overlayColor) || defaultZoneKindOverlayColor;
  if (!overlayColor) {
    return { error: 'Overlay color must be a valid hex color.' };
  }
  payload.overlayColor = overlayColor;

  for (const field of ratioFields) {
    const parsed = Number.parseFloat(form[field.key]);
    if (!Number.isFinite(parsed) || parsed < 0) {
      return {
        error: `${field.label} ratio must be a number greater than or equal to 0.`,
      };
    }
    payload[field.key] = parsed;
  }
  payload.resourceCountRatio =
    (payload.herbalismResourceCountRatio + payload.miningResourceCountRatio) /
    2;

  return { payload };
};

export const ZoneKinds = () => {
  const { apiClient } = useAPI();
  const [zoneKinds, setZoneKinds] = useState<ZoneKind[]>([]);
  const [zones, setZones] = useState<ZoneAdminSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [assigningZoneId, setAssigningZoneId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [zoneSearch, setZoneSearch] = useState('');
  const [createForm, setCreateForm] = useState<ZoneKindFormState>(emptyForm());
  const [createSlugTouched, setCreateSlugTouched] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editForm, setEditForm] = useState<ZoneKindFormState>(emptyForm());
  const [editSlugTouched, setEditSlugTouched] = useState(false);
  const [backfillStarting, setBackfillStarting] = useState(false);
  const [backfillStatus, setBackfillStatus] =
    useState<ZoneKindBackfillStatus | null>(null);
  const [patternPromptDrafts, setPatternPromptDrafts] = useState<
    Record<string, string>
  >({});
  const [generatingPatternKindId, setGeneratingPatternKindId] = useState<
    string | null
  >(null);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [zoneKindsResponse, zonesResponse] = await Promise.all([
        apiClient.get<ZoneKind[]>('/sonar/zoneKinds'),
        apiClient.get<ZoneAdminSummary[]>('/sonar/admin/zones'),
      ]);
      setZoneKinds(zoneKindsResponse);
      setZones(zonesResponse);
      setError(null);
    } catch (err) {
      console.error('Failed to load zone kinds page data', err);
      setError('Unable to load zone kinds right now.');
    } finally {
      setLoading(false);
    }
  }, [apiClient]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const zoneKindBySlug = useMemo(() => {
    const next = new Map<string, ZoneKind>();
    zoneKinds.forEach((zoneKind) => next.set(zoneKind.slug, zoneKind));
    return next;
  }, [zoneKinds]);

  const assignedZonesByKind = useMemo(() => {
    const next = new Map<string, ZoneAdminSummary[]>();
    zoneKinds.forEach((zoneKind) => next.set(zoneKind.slug, []));
    zones.forEach((zone) => {
      if (!zone.kind) {
        return;
      }
      const current = next.get(zone.kind) || [];
      current.push(zone);
      next.set(zone.kind, current);
    });
    return next;
  }, [zoneKinds, zones]);

  const filteredZones = useMemo(() => {
    const normalizedQuery = zoneSearch.trim().toLowerCase();
    if (!normalizedQuery) {
      return zones;
    }
    return zones.filter((zone) =>
      zoneSearchText(zone).includes(normalizedQuery)
    );
  }, [zoneSearch, zones]);

  const setCreateField = <K extends keyof ZoneKindFormState>(
    key: K,
    value: ZoneKindFormState[K]
  ) => {
    setCreateForm((prev) => ({ ...prev, [key]: value }));
  };

  const setEditField = <K extends keyof ZoneKindFormState>(
    key: K,
    value: ZoneKindFormState[K]
  ) => {
    setEditForm((prev) => ({ ...prev, [key]: value }));
  };

  const handleCreate = async () => {
    const { payload, error: parseError } = parseZoneKindForm(createForm);
    if (!payload) {
      setError(parseError || 'Unable to build zone kind payload.');
      return;
    }

    setSaving(true);
    setError(null);
    setSuccess(null);
    try {
      const created = await apiClient.post<ZoneKind>(
        '/sonar/zoneKinds',
        payload
      );
      setZoneKinds((prev) =>
        [...prev, created].sort((a, b) => a.name.localeCompare(b.name))
      );
      setCreateForm(emptyForm());
      setCreateSlugTouched(false);
      setSuccess(`Created ${created.name}.`);
    } catch (err) {
      console.error('Failed to create zone kind', err);
      setError('Unable to create this zone kind.');
    } finally {
      setSaving(false);
    }
  };

  const startEditing = (zoneKind: ZoneKind) => {
    setEditingId(zoneKind.id);
    setEditForm(formFromZoneKind(zoneKind));
    setEditSlugTouched(true);
    setError(null);
    setSuccess(null);
  };

  const cancelEditing = () => {
    setEditingId(null);
    setEditForm(emptyForm());
    setEditSlugTouched(false);
  };

  const handleUpdate = async () => {
    if (!editingId) {
      return;
    }
    const editingZoneKind = zoneKinds.find(
      (zoneKind) => zoneKind.id === editingId
    );
    const { payload, error: parseError } = parseZoneKindForm(editForm);
    if (!payload) {
      setError(parseError || 'Unable to build zone kind payload.');
      return;
    }

    setSaving(true);
    setError(null);
    setSuccess(null);
    try {
      const updated = await apiClient.patch<ZoneKind>(
        `/sonar/zoneKinds/${editingId}`,
        payload
      );
      setZoneKinds((prev) =>
        prev
          .map((zoneKind) => (zoneKind.id === updated.id ? updated : zoneKind))
          .sort((a, b) => a.name.localeCompare(b.name))
      );
      setZones((prev) =>
        prev.map((zone) =>
          zone.kind === editingZoneKind?.slug && zone.kind !== updated.slug
            ? { ...zone, kind: updated.slug }
            : zone
        )
      );
      setSuccess(`Updated ${updated.name}.`);
      cancelEditing();
    } catch (err) {
      console.error('Failed to update zone kind', err);
      setError('Unable to update this zone kind.');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (zoneKind: ZoneKind) => {
    const confirmed = window.confirm(
      `Delete ${zoneKind.name}? Any zones using ${zoneKind.slug} will be cleared.`
    );
    if (!confirmed) {
      return;
    }

    setSaving(true);
    setError(null);
    setSuccess(null);
    try {
      await apiClient.delete(`/sonar/zoneKinds/${zoneKind.id}`);
      setZoneKinds((prev) => prev.filter((entry) => entry.id !== zoneKind.id));
      setZones((prev) =>
        prev.map((zone) =>
          zone.kind === zoneKind.slug ? { ...zone, kind: '' } : zone
        )
      );
      if (editingId === zoneKind.id) {
        cancelEditing();
      }
      setSuccess(`Deleted ${zoneKind.name}.`);
    } catch (err) {
      console.error('Failed to delete zone kind', err);
      setError('Unable to delete this zone kind.');
    } finally {
      setSaving(false);
    }
  };

  const assignZoneKind = async (zoneId: string, kind: string) => {
    setAssigningZoneId(zoneId);
    setError(null);
    setSuccess(null);
    try {
      await apiClient.post('/sonar/zoneKinds/assign-zones', {
        zoneIds: [zoneId],
        kind,
      });
      setZones((prev) =>
        prev.map((zone) => (zone.id === zoneId ? { ...zone, kind } : zone))
      );
      const assignedName =
        kind === ''
          ? 'Cleared zone kind.'
          : `Assigned ${zoneKindBySlug.get(kind)?.name ?? kind}.`;
      setSuccess(assignedName);
    } catch (err) {
      console.error('Failed to assign zone kind', err);
      setError('Unable to update that zone assignment.');
    } finally {
      setAssigningZoneId(null);
    }
  };

  const pollBackfillStatus = useCallback(
    async (jobId: string) => {
      try {
        const nextStatus = await apiClient.get<ZoneKindBackfillStatus>(
          `/sonar/zoneKinds/backfill-content-kinds/${jobId}/status`
        );
        setBackfillStatus(nextStatus);
      } catch (err) {
        console.error('Failed to poll zone kind backfill status', err);
      }
    },
    [apiClient]
  );

  useEffect(() => {
    if (
      !backfillStatus?.jobId ||
      !isZoneKindBackfillPending(backfillStatus.status)
    ) {
      return;
    }

    const timeoutId = window.setTimeout(() => {
      void pollBackfillStatus(backfillStatus.jobId);
    }, 2000);

    return () => window.clearTimeout(timeoutId);
  }, [backfillStatus, pollBackfillStatus]);

  const pendingPatternGenerationKey = useMemo(
    () =>
      zoneKinds
        .filter((zoneKind) =>
          isZoneKindPatternTilePending(zoneKind.patternTileGenerationStatus)
        )
        .map(
          (zoneKind) =>
            `${zoneKind.id}:${zoneKind.patternTileGenerationStatus}:${zoneKind.updatedAt}`
        )
        .join('|'),
    [zoneKinds]
  );

  useEffect(() => {
    if (!pendingPatternGenerationKey) {
      return;
    }

    const timeoutId = window.setTimeout(() => {
      void loadData();
    }, 2200);

    return () => window.clearTimeout(timeoutId);
  }, [loadData, pendingPatternGenerationKey]);

  const handleBackfillContentKinds = async () => {
    setBackfillStarting(true);
    setError(null);
    setSuccess(null);
    try {
      const status = await apiClient.post<ZoneKindBackfillStatus>(
        '/sonar/zoneKinds/backfill-content-kinds',
        {}
      );
      setBackfillStatus(status);
      setSuccess('Queued the missing content zone kind backfill job.');
    } catch (err) {
      console.error('Failed to queue zone kind backfill job', err);
      setError('Unable to queue the content zone kind backfill.');
    } finally {
      setBackfillStarting(false);
    }
  };

  const handleGeneratePatternTile = async (zoneKind: ZoneKind) => {
    const prompt = resolveZoneKindPatternPrompt(
      zoneKind,
      patternPromptDrafts[zoneKind.id]
    );
    if (prompt.length < 24) {
      setError('Pattern tile prompts must be at least 24 characters.');
      return;
    }
    if (prompt.length > 8000) {
      setError('Pattern tile prompts must be at most 8000 characters.');
      return;
    }

    setGeneratingPatternKindId(zoneKind.id);
    setError(null);
    setSuccess(null);
    try {
      const updated = await apiClient.post<ZoneKind>(
        `/sonar/zoneKinds/${zoneKind.id}/generate-pattern-tile`,
        { prompt }
      );
      setZoneKinds((prev) =>
        prev
          .map((entry) => (entry.id === updated.id ? updated : entry))
          .sort((a, b) => a.name.localeCompare(b.name))
      );
      setPatternPromptDrafts((prev) => ({
        ...prev,
        [zoneKind.id]: updated.patternTilePrompt || prompt,
      }));
      setSuccess(`Queued pattern tile generation for ${zoneKind.name}.`);
    } catch (err) {
      console.error('Failed to queue zone kind pattern tile job', err);
      setError('Unable to queue that zone kind pattern tile job.');
    } finally {
      setGeneratingPatternKindId(null);
    }
  };

  if (loading) {
    return <div className="p-6 text-gray-500">Loading zone kinds...</div>;
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">Zone Kinds</h1>
          <p className="mt-1 max-w-3xl text-sm text-gray-500">
            Create reusable zone presets, tune the auto-seed ratios for every
            generated content type, and assign those presets directly to zones.
            A ratio of 1.0 keeps the baseline area recommendation, 2.0 doubles
            it, and 0.5 halves it.
          </p>
        </div>
        <div className="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-900">
          <div className="font-medium">{zoneKinds.length} zone kinds</div>
          <div>{zones.filter((zone) => zone.kind).length} assigned zones</div>
        </div>
      </div>

      {(error || success) && (
        <div className="space-y-2">
          {error && (
            <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
              {error}
            </div>
          )}
          {success && (
            <div className="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
              {success}
            </div>
          )}
        </div>
      )}

      <section className="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">
              Fill missing content kinds
            </h2>
            <p className="mt-1 max-w-3xl text-sm text-gray-500">
              Backfill empty content and inventory item zone kinds from their
              linked zones and reward relationships. Existing assignments stay
              untouched.
            </p>
          </div>
          <button
            type="button"
            className="rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
            onClick={handleBackfillContentKinds}
            disabled={
              backfillStarting ||
              isZoneKindBackfillPending(backfillStatus?.status)
            }
          >
            {backfillStarting || isZoneKindBackfillPending(backfillStatus?.status)
              ? 'Running...'
              : 'Backfill Missing Kinds'}
          </button>
        </div>

        {backfillStatus && (
          <div className="mt-4 rounded-2xl border border-slate-200 bg-slate-50 p-4">
            <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
              <div>
                <div className="text-xs font-medium uppercase tracking-wide text-slate-500">
                  Latest job
                </div>
                <div className="mt-1 text-sm text-slate-700">
                  {backfillStatus.jobId}
                </div>
              </div>
              <span
                className={`inline-flex w-fit rounded-full px-3 py-1 text-xs font-medium ${
                  backfillStatus.status === 'failed'
                    ? 'bg-red-100 text-red-700'
                    : backfillStatus.status === 'completed'
                      ? 'bg-emerald-100 text-emerald-700'
                      : 'bg-amber-100 text-amber-700'
                }`}
              >
                {formatZoneKindBackfillStatus(backfillStatus.status)}
              </span>
            </div>

            <div className="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
              <div className="rounded-xl border border-white bg-white px-3 py-3 shadow-sm">
                <div className="text-xs font-medium uppercase tracking-wide text-gray-500">
                  Missing
                </div>
                <div className="mt-1 text-lg font-semibold text-gray-900">
                  {backfillStatus.summary.missingCount ?? 0}
                </div>
              </div>
              <div className="rounded-xl border border-white bg-white px-3 py-3 shadow-sm">
                <div className="text-xs font-medium uppercase tracking-wide text-gray-500">
                  Assigned
                </div>
                <div className="mt-1 text-lg font-semibold text-emerald-700">
                  {backfillStatus.summary.assignedCount ?? 0}
                </div>
              </div>
              <div className="rounded-xl border border-white bg-white px-3 py-3 shadow-sm">
                <div className="text-xs font-medium uppercase tracking-wide text-gray-500">
                  Ambiguous
                </div>
                <div className="mt-1 text-lg font-semibold text-amber-700">
                  {backfillStatus.summary.ambiguousCount ?? 0}
                </div>
              </div>
              <div className="rounded-xl border border-white bg-white px-3 py-3 shadow-sm">
                <div className="text-xs font-medium uppercase tracking-wide text-gray-500">
                  Skipped
                </div>
                <div className="mt-1 text-lg font-semibold text-slate-700">
                  {backfillStatus.summary.skippedCount ?? 0}
                </div>
              </div>
            </div>

            {backfillStatus.error && (
              <div className="mt-4 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
                {backfillStatus.error}
              </div>
            )}

            {(backfillStatus.summary.results?.length ?? 0) > 0 && (
              <div className="mt-4 overflow-hidden rounded-2xl border border-gray-200">
                <div className="grid grid-cols-[minmax(0,1.4fr)_repeat(4,minmax(0,0.8fr))] gap-4 bg-white px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500">
                  <div>Content type</div>
                  <div>Missing</div>
                  <div>Assigned</div>
                  <div>Ambiguous</div>
                  <div>Skipped</div>
                </div>
                {backfillStatus.summary.results.map((result) => (
                  <div
                    key={result.contentType}
                    className="grid grid-cols-[minmax(0,1.4fr)_repeat(4,minmax(0,0.8fr))] gap-4 border-t border-gray-200 bg-slate-50 px-4 py-3 text-sm text-slate-700"
                  >
                    <div className="font-medium text-slate-900">
                      {zoneKindBackfillContentLabels[result.contentType] ??
                        result.contentType}
                    </div>
                    <div>{result.missingCount}</div>
                    <div>{result.assignedCount}</div>
                    <div>{result.ambiguousCount}</div>
                    <div>{result.skippedCount}</div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </section>

      <div className="grid gap-6 xl:grid-cols-[minmax(320px,420px)_minmax(0,1fr)]">
        <section className="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm">
          <div className="flex items-start justify-between gap-3">
            <div>
              <h2 className="text-lg font-semibold text-gray-900">
                Create zone kind
              </h2>
              <p className="mt-1 text-sm text-gray-500">
                Start with a readable label, then tune how heavily each content
                category should be represented in auto-seeding.
              </p>
            </div>
          </div>

          <div className="mt-4 space-y-4">
            <div>
              <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                Name
              </label>
              <input
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                value={createForm.name}
                onChange={(event) => {
                  const nextName = event.target.value;
                  setCreateField('name', nextName);
                  if (!createSlugTouched) {
                    setCreateField('slug', normalizeSlugDraft(nextName));
                  }
                }}
                placeholder="Forest"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                Slug
              </label>
              <input
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                value={createForm.slug}
                onChange={(event) => {
                  setCreateSlugTouched(true);
                  setCreateField(
                    'slug',
                    normalizeSlugDraft(event.target.value)
                  );
                }}
                placeholder="forest"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                Description
              </label>
              <textarea
                className="min-h-[88px] w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                value={createForm.description}
                onChange={(event) =>
                  setCreateField('description', event.target.value)
                }
                placeholder="High-herbalism wilderness with more beasts, shrines, and restorative nodes."
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                Map overlay color
              </label>
              <div className="flex items-center gap-3 rounded-xl border border-gray-200 bg-slate-50 px-3 py-3">
                <input
                  type="color"
                  className="h-11 w-14 cursor-pointer rounded-md border border-gray-300 bg-white p-1"
                  value={
                    normalizeOverlayColorDraft(createForm.overlayColor) ||
                    defaultZoneKindOverlayColor
                  }
                  onChange={(event) =>
                    setCreateField('overlayColor', event.target.value)
                  }
                />
                <div className="min-w-0">
                  <div className="text-sm font-medium text-gray-900">
                    {normalizeOverlayColorDraft(createForm.overlayColor) ||
                      defaultZoneKindOverlayColor}
                  </div>
                  <div className="text-xs text-gray-500">
                    Used for the player map zone overlay. New kinds start with a
                    grounded default tone.
                  </div>
                </div>
              </div>
            </div>

            <div className="grid gap-3 sm:grid-cols-2">
              {ratioFields.map((field) => (
                <label key={field.key} className="block">
                  <span className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                    {field.label}
                  </span>
                  <input
                    type="number"
                    min="0"
                    step="0.1"
                    className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                    value={createForm[field.key]}
                    onChange={(event) =>
                      setCreateField(field.key, event.target.value)
                    }
                  />
                  <span className="mt-1 block text-[11px] text-gray-500">
                    {field.description}
                  </span>
                </label>
              ))}
            </div>

            <button
              type="button"
              className="w-full rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
              onClick={handleCreate}
              disabled={saving}
            >
              {saving ? 'Saving...' : 'Create Zone Kind'}
            </button>
          </div>
        </section>

        <section className="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm">
          <div className="flex items-start justify-between gap-3">
            <div>
              <h2 className="text-lg font-semibold text-gray-900">
                Zone kind library
              </h2>
              <p className="mt-1 text-sm text-gray-500">
                These presets are available anywhere we assign or seed a zone.
              </p>
            </div>
          </div>

          {zoneKinds.length === 0 ? (
            <div className="mt-4 rounded-xl border border-dashed border-gray-300 bg-gray-50 px-4 py-10 text-center text-sm text-gray-500">
              No zone kinds yet. Create one to start defining reusable seeding
              identities.
            </div>
          ) : (
            <div className="mt-4 space-y-4">
              {zoneKinds.map((zoneKind) => {
                const assignedZones =
                  assignedZonesByKind.get(zoneKind.slug) || [];
                const isEditing = editingId === zoneKind.id;

                return (
                  <article
                    key={zoneKind.id}
                    className="rounded-2xl border border-gray-200 bg-gray-50 p-4"
                  >
                    <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
                      <div>
                        <div className="flex flex-wrap items-center gap-2">
                          <span
                            className="h-4 w-4 rounded-full border border-black/10 shadow-sm"
                            style={{
                              backgroundColor:
                                normalizeOverlayColorDraft(
                                  zoneKind.overlayColor
                                ) || defaultZoneKindOverlayColor,
                            }}
                            aria-hidden="true"
                          />
                          <h3 className="text-base font-semibold text-gray-900">
                            {zoneKind.name}
                          </h3>
                          <span className="rounded-full bg-slate-900 px-2.5 py-1 text-[11px] font-medium uppercase tracking-wide text-white">
                            {zoneKind.slug}
                          </span>
                          <span className="rounded-full bg-white px-2.5 py-1 text-[11px] font-medium text-slate-600">
                            {assignedZones.length} assigned zone
                            {assignedZones.length === 1 ? '' : 's'}
                          </span>
                          <span className="rounded-full bg-white px-2.5 py-1 text-[11px] font-medium text-slate-600">
                            Overlay{' '}
                            {normalizeOverlayColorDraft(zoneKind.overlayColor) ||
                              defaultZoneKindOverlayColor}
                          </span>
                        </div>
                        {zoneKind.description && (
                          <p className="mt-2 text-sm text-gray-600">
                            {zoneKind.description}
                          </p>
                        )}
                      </div>
                      <div className="flex gap-2">
                        <button
                          type="button"
                          className="rounded-lg border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 transition hover:border-slate-400 hover:text-slate-900"
                          onClick={() =>
                            isEditing ? cancelEditing() : startEditing(zoneKind)
                          }
                        >
                          {isEditing ? 'Cancel' : 'Edit'}
                        </button>
                        <button
                          type="button"
                          className="rounded-lg border border-red-200 bg-red-50 px-3 py-1.5 text-sm font-medium text-red-700 transition hover:border-red-300 hover:bg-red-100"
                          onClick={() => handleDelete(zoneKind)}
                          disabled={saving}
                        >
                          Delete
                        </button>
                      </div>
                    </div>

                    <div className="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
                      {ratioFields.map((field) => (
                        <div
                          key={`${zoneKind.id}-${field.key}`}
                          className="rounded-xl border border-white bg-white px-3 py-3 shadow-sm"
                        >
                          <div className="text-xs font-medium uppercase tracking-wide text-gray-500">
                            {field.label}
                          </div>
                          <div className="mt-1 text-lg font-semibold text-gray-900">
                            {formatRatio(zoneKind[field.key])}
                          </div>
                          <div className="mt-1 text-xs text-gray-500">
                            {field.description}
                          </div>
                        </div>
                      ))}
                    </div>

                    <div className="mt-4 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
                      <div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
                        <div>
                          <div className="flex flex-wrap items-center gap-2">
                            <h4 className="text-sm font-semibold text-slate-900">
                              Pattern Tile
                            </h4>
                            <span
                              className={`rounded-full px-2.5 py-1 text-[11px] font-medium ${
                                zoneKind.patternTileGenerationStatus === 'failed'
                                  ? 'bg-red-100 text-red-700'
                                  : zoneKind.patternTileGenerationStatus === 'complete'
                                    ? 'bg-emerald-100 text-emerald-700'
                                    : isZoneKindPatternTilePending(
                                          zoneKind.patternTileGenerationStatus
                                        )
                                      ? 'bg-amber-100 text-amber-700'
                                      : 'bg-slate-100 text-slate-600'
                              }`}
                            >
                              {formatZoneKindPatternTileStatus(
                                zoneKind.patternTileGenerationStatus
                              )}
                            </span>
                          </div>
                          <p className="mt-1 max-w-2xl text-sm text-slate-600">
                            Generate a seamless, S3-hosted pattern tile that
                            the mobile client can use as the zone fill texture.
                          </p>
                          {zoneKind.patternTileUrl ? (
                            <a
                              href={zoneKind.patternTileUrl}
                              target="_blank"
                              rel="noreferrer"
                              className="mt-2 inline-flex text-xs font-medium text-slate-700 underline decoration-slate-300 underline-offset-4 transition hover:text-slate-900"
                            >
                              Open current tile asset
                            </a>
                          ) : null}
                          {zoneKind.patternTileGenerationError ? (
                            <div className="mt-3 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-700">
                              {zoneKind.patternTileGenerationError}
                            </div>
                          ) : null}
                        </div>

                        <div className="flex items-start gap-4">
                          <div
                            className="relative h-28 w-28 overflow-hidden rounded-2xl border border-slate-200 shadow-inner"
                            style={{
                              backgroundColor:
                                normalizeOverlayColorDraft(
                                  zoneKind.overlayColor
                                ) || defaultZoneKindOverlayColor,
                            }}
                          >
                            {zoneKind.patternTileUrl ? (
                              <div
                                className="absolute inset-0"
                                style={{
                                  backgroundImage: `url(${zoneKind.patternTileUrl})`,
                                  backgroundRepeat: 'repeat',
                                  backgroundSize: '96px 96px',
                                  opacity: 0.7,
                                  mixBlendMode: 'multiply',
                                }}
                              />
                            ) : (
                              <div className="absolute inset-0 flex items-center justify-center px-3 text-center text-[11px] font-medium uppercase tracking-wide text-white/80">
                                No tile yet
                              </div>
                            )}
                          </div>
                        </div>
                      </div>

                      <div className="mt-4">
                        <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-slate-500">
                          Boilerplate prompt
                        </label>
                        <textarea
                          className="min-h-[184px] w-full rounded-xl border border-slate-300 px-3 py-2 text-sm text-slate-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                          value={resolveZoneKindPatternPrompt(
                            zoneKind,
                            patternPromptDrafts[zoneKind.id]
                          )}
                          onChange={(event) =>
                            setPatternPromptDrafts((prev) => ({
                              ...prev,
                              [zoneKind.id]: event.target.value,
                            }))
                          }
                          spellCheck={false}
                        />
                        <div className="mt-3 flex flex-wrap items-center gap-2">
                          <button
                            type="button"
                            className="rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
                            onClick={() => void handleGeneratePatternTile(zoneKind)}
                            disabled={
                              saving ||
                              generatingPatternKindId === zoneKind.id ||
                              isZoneKindPatternTilePending(
                                zoneKind.patternTileGenerationStatus
                              )
                            }
                          >
                            {generatingPatternKindId === zoneKind.id ||
                            isZoneKindPatternTilePending(
                              zoneKind.patternTileGenerationStatus
                            )
                              ? 'Queueing...'
                              : 'Generate Tile Job'}
                          </button>
                          <button
                            type="button"
                            className="rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm font-medium text-slate-700 transition hover:border-slate-400 hover:text-slate-900"
                            onClick={() =>
                              setPatternPromptDrafts((prev) => ({
                                ...prev,
                                [zoneKind.id]:
                                  buildDefaultZoneKindPatternPrompt(zoneKind),
                              }))
                            }
                          >
                            Use Boilerplate
                          </button>
                        </div>
                      </div>
                    </div>

                    {assignedZones.length > 0 && (
                      <div className="mt-4 text-sm text-gray-600">
                        Assigned zones:{' '}
                        {assignedZones
                          .slice(0, 4)
                          .map((zone) => zone.name)
                          .join(', ')}
                        {assignedZones.length > 4
                          ? ` +${assignedZones.length - 4} more`
                          : ''}
                      </div>
                    )}

                    {isEditing && (
                      <div className="mt-4 rounded-2xl border border-slate-200 bg-white p-4">
                        <div className="grid gap-4">
                          <div className="grid gap-4 md:grid-cols-2">
                            <div>
                              <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                                Name
                              </label>
                              <input
                                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                                value={editForm.name}
                                onChange={(event) => {
                                  const nextName = event.target.value;
                                  setEditField('name', nextName);
                                  if (!editSlugTouched) {
                                    setEditField(
                                      'slug',
                                      normalizeSlugDraft(nextName)
                                    );
                                  }
                                }}
                              />
                            </div>
                            <div>
                              <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                                Slug
                              </label>
                              <input
                                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                                value={editForm.slug}
                                onChange={(event) => {
                                  setEditSlugTouched(true);
                                  setEditField(
                                    'slug',
                                    normalizeSlugDraft(event.target.value)
                                  );
                                }}
                              />
                            </div>
                          </div>

                          <div>
                            <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                              Description
                            </label>
                            <textarea
                              className="min-h-[88px] w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                              value={editForm.description}
                              onChange={(event) =>
                                setEditField('description', event.target.value)
                              }
                            />
                          </div>

                          <div>
                            <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                              Map overlay color
                            </label>
                            <div className="flex items-center gap-3 rounded-xl border border-gray-200 bg-slate-50 px-3 py-3">
                              <input
                                type="color"
                                className="h-11 w-14 cursor-pointer rounded-md border border-gray-300 bg-white p-1"
                                value={
                                  normalizeOverlayColorDraft(
                                    editForm.overlayColor
                                  ) || defaultZoneKindOverlayColor
                                }
                                onChange={(event) =>
                                  setEditField('overlayColor', event.target.value)
                                }
                              />
                              <div className="min-w-0">
                                <div className="text-sm font-medium text-gray-900">
                                  {normalizeOverlayColorDraft(
                                    editForm.overlayColor
                                  ) || defaultZoneKindOverlayColor}
                                </div>
                                <div className="text-xs text-gray-500">
                                  Applied to the player map overlay for zones
                                  using this kind.
                                </div>
                              </div>
                            </div>
                          </div>

                          <div className="grid gap-3 sm:grid-cols-2">
                            {ratioFields.map((field) => (
                              <label
                                key={`${zoneKind.id}-edit-${field.key}`}
                                className="block"
                              >
                                <span className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
                                  {field.label}
                                </span>
                                <input
                                  type="number"
                                  min="0"
                                  step="0.1"
                                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                                  value={editForm[field.key]}
                                  onChange={(event) =>
                                    setEditField(field.key, event.target.value)
                                  }
                                />
                              </label>
                            ))}
                          </div>

                          <div className="flex justify-end gap-2">
                            <button
                              type="button"
                              className="rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 transition hover:border-slate-400 hover:text-slate-900"
                              onClick={cancelEditing}
                            >
                              Cancel
                            </button>
                            <button
                              type="button"
                              className="rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-slate-800"
                              onClick={handleUpdate}
                              disabled={saving}
                            >
                              Save changes
                            </button>
                          </div>
                        </div>
                      </div>
                    )}
                  </article>
                );
              })}
            </div>
          )}
        </section>
      </div>

      <section className="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">
              Assign kinds to zones
            </h2>
            <p className="mt-1 text-sm text-gray-500">
              Pick a zone kind per zone. Seed jobs will use the assigned kind by
              default whenever no explicit override is provided.
            </p>
          </div>
          <div className="w-full max-w-sm">
            <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-gray-500">
              Search zones
            </label>
            <input
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
              value={zoneSearch}
              onChange={(event) => setZoneSearch(event.target.value)}
              placeholder="Forest, downtown, tag, metro..."
            />
          </div>
        </div>

        <div className="mt-4 overflow-hidden rounded-2xl border border-gray-200">
          <div className="grid grid-cols-[minmax(0,1.4fr)_minmax(0,0.9fr)_minmax(180px,0.9fr)] gap-4 bg-slate-50 px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-500">
            <div>Zone</div>
            <div>Current kind</div>
            <div>Assignment</div>
          </div>

          {filteredZones.length === 0 ? (
            <div className="px-4 py-10 text-center text-sm text-gray-500">
              No zones match this search.
            </div>
          ) : (
            filteredZones.map((zone) => {
              const matchedKind = zone.kind
                ? zoneKindBySlug.get(zone.kind)
                : null;
              const currentKindLabel = zone.kind
                ? matchedKind?.name ?? `Unknown (${zone.kind})`
                : 'Unassigned';

              return (
                <div
                  key={zone.id}
                  className="grid grid-cols-[minmax(0,1.4fr)_minmax(0,0.9fr)_minmax(180px,0.9fr)] gap-4 border-t border-gray-200 px-4 py-4 text-sm text-gray-700"
                >
                  <div>
                    <div className="font-medium text-gray-900">
                      <Link
                        to={`/zones/${zone.id}`}
                        className="transition hover:text-slate-700 hover:underline"
                      >
                        {zone.name}
                      </Link>
                    </div>
                    <div className="mt-1 text-xs text-gray-500">
                      {zone.importMetroName || 'Custom zone'}
                    </div>
                  </div>
                  <div className="flex items-center">
                    <span
                      className={`rounded-full px-2.5 py-1 text-xs font-medium ${
                        zone.kind && !matchedKind
                          ? 'bg-amber-100 text-amber-800'
                          : zone.kind
                            ? 'bg-slate-100 text-slate-700'
                          : 'bg-gray-100 text-gray-500'
                      }`}
                    >
                      {matchedKind && (
                        <span
                          className="mr-1.5 inline-block h-2 w-2 rounded-full align-middle"
                          style={{
                            backgroundColor:
                              normalizeOverlayColorDraft(
                                matchedKind.overlayColor
                              ) || defaultZoneKindOverlayColor,
                          }}
                        />
                      )}
                      {currentKindLabel}
                    </span>
                  </div>
                  <div>
                    <select
                      className="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm focus:border-slate-500 focus:outline-none focus:ring-2 focus:ring-slate-200"
                      value={zone.kind || ''}
                      onChange={(event) =>
                        void assignZoneKind(zone.id, event.target.value)
                      }
                      disabled={assigningZoneId === zone.id}
                    >
                      <option value="">Unassigned</option>
                      {zoneKinds.map((zoneKind) => (
                        <option
                          key={`${zone.id}-${zoneKind.id}`}
                          value={zoneKind.slug}
                        >
                          {zoneKind.name}
                        </option>
                      ))}
                    </select>
                    {assigningZoneId === zone.id && (
                      <div className="mt-1 text-xs text-slate-500">
                        Saving...
                      </div>
                    )}
                  </div>
                </div>
              );
            })
          )}
        </div>
      </section>
    </div>
  );
};

export default ZoneKinds;
