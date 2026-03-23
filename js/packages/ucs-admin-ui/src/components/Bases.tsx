import React, { useCallback, useEffect, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';

type BaseRecord = {
  id: string;
  userId: string;
  latitude: number;
  longitude: number;
  description?: string;
  imageUrl?: string;
  thumbnailUrl?: string;
  createdAt?: string;
  updatedAt?: string;
  owner?: {
    id?: string;
    name?: string;
    username?: string;
    profilePictureUrl?: string;
  };
};

type StaticThumbnailResponse = {
  thumbnailUrl?: string;
  status?: string;
  exists?: boolean;
  requestedAt?: string;
  lastModified?: string;
};

type BaseDescriptionGenerationJob = {
  id: string;
  baseId: string;
  status?: string;
  generatedDescription?: string;
  generatedImageUrl?: string;
  errorMessage?: string;
  createdAt?: string;
  updatedAt?: string;
};

type BaseStructureLevelVisual = {
  id?: string | null;
  level: number;
  imageUrl?: string;
  thumbnailUrl?: string;
  imageGenerationStatus?: string;
  imageGenerationError?: string | null;
  topDownImageUrl?: string;
  topDownThumbnailUrl?: string;
  topDownImageGenerationStatus?: string;
  topDownImageGenerationError?: string | null;
};

type BaseStructureDefinition = {
  id: string;
  key: string;
  name: string;
  description?: string;
  category?: string;
  maxLevel: number;
  active?: boolean;
  imagePrompt?: string;
  topDownImagePrompt?: string;
  resolvedImagePrompt?: string;
  resolvedTopDownImagePrompt?: string;
  effectConfig?: Record<string, unknown>;
  levelVisuals?: BaseStructureLevelVisual[];
};

type HearthRecoveryStatusTemplate = {
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
};

type HearthRecoveryDraft = {
  level2Statuses: HearthRecoveryStatusTemplate[];
  level3Statuses: HearthRecoveryStatusTemplate[];
};

type BaseGrassTileStatus = {
  gridX: number;
  gridY: number;
  status?: string;
  exists?: boolean;
  thumbnailUrl?: string;
  requestedAt?: string;
  lastModified?: string;
  prompt?: string;
};

const defaultBaseIconPrompt =
  'A discovered adventurer base marker in a retro 16-bit fantasy MMORPG style. Top-down map-ready icon art, sturdy camp or homestead sigil, welcoming hearth glow, no text, no logos, centered composition, crisp outlines, limited palette.';

const defaultBaseGrassPrompt =
  'A seamless top-down grass terrain tile for a retro 16-bit fantasy MMORPG base builder. Overhead view, softly varied green blades, subtle earth patches, crisp pixel edges, no structures, no text, no logos, tileable and clean.';

const staticStatusClassName = (status?: string) => {
  switch ((status || '').toLowerCase()) {
    case 'complete':
    case 'completed':
      return 'bg-emerald-600';
    case 'in_progress':
      return 'bg-amber-600';
    case 'queued':
      return 'bg-blue-600';
    case 'failed':
      return 'bg-red-600';
    case 'missing':
      return 'bg-slate-500';
    default:
      return 'bg-slate-500';
  }
};

const formatDate = (value?: string) => {
  if (!value) return 'n/a';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

const ownerLabel = (record: BaseRecord) => {
  const username = record.owner?.username?.trim();
  if (username) return `@${username}`;
  const name = record.owner?.name?.trim();
  if (name) return name;
  return record.userId;
};

const statusEffectTypes = [
  'stat_modifier',
  'damage_over_time',
  'health_over_time',
  'mana_over_time',
] as const;

const emptyHearthRecoveryStatus = (): HearthRecoveryStatusTemplate => ({
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
});

const parseIntValue = (value: string, fallback = 0) => {
  const parsed = parseInt(value, 10);
  return Number.isNaN(parsed) ? fallback : parsed;
};

const coerceString = (value: unknown) =>
  typeof value === 'string' ? value : '';

const coerceNumber = (value: unknown, fallback = 0) =>
  typeof value === 'number' && Number.isFinite(value) ? value : fallback;

const coerceBoolean = (value: unknown, fallback = true) =>
  typeof value === 'boolean' ? value : fallback;

const normalizeHearthRecoveryStatus = (
  value: unknown
): HearthRecoveryStatusTemplate => {
  if (!value || typeof value !== 'object') {
    return emptyHearthRecoveryStatus();
  }
  const status = value as Record<string, unknown>;
  return {
    name: coerceString(status.name),
    description: coerceString(status.description),
    effect: coerceString(status.effect),
    effectType: coerceString(status.effectType) || 'stat_modifier',
    positive: coerceBoolean(status.positive, true),
    damagePerTick: coerceNumber(status.damagePerTick),
    healthPerTick: coerceNumber(status.healthPerTick),
    manaPerTick: coerceNumber(status.manaPerTick),
    durationSeconds: coerceNumber(status.durationSeconds, 60),
    strengthMod: coerceNumber(status.strengthMod),
    dexterityMod: coerceNumber(status.dexterityMod),
    constitutionMod: coerceNumber(status.constitutionMod),
    intelligenceMod: coerceNumber(status.intelligenceMod),
    wisdomMod: coerceNumber(status.wisdomMod),
    charismaMod: coerceNumber(status.charismaMod),
  };
};

const normalizeHearthRecoveryStatuses = (
  value: unknown
): HearthRecoveryStatusTemplate[] =>
  Array.isArray(value) ? value.map(normalizeHearthRecoveryStatus) : [];

const secondaryOwnerLabel = (record: BaseRecord) => {
  const username = record.owner?.username?.trim();
  const name = record.owner?.name?.trim();
  if (!username || !name || name === username) return '';
  return name;
};

const grassTileKey = (gridX: number, gridY: number) => `${gridX}:${gridY}`;

const defaultGrassPromptForCell = (gridX: number, gridY: number) =>
  `${defaultBaseGrassPrompt} Subtle variation for base grid coordinate (${gridX},${gridY}), so neighboring tiles feel related but not identical.`;

const buildGrassPromptForCell = (
  basePrompt: string,
  gridX: number,
  gridY: number
) => {
  const trimmed = basePrompt.trim();
  if (!trimmed) {
    return defaultGrassPromptForCell(gridX, gridY);
  }
  return `${trimmed} Subtle variation for base grid coordinate (${gridX},${gridY}), so neighboring tiles feel related but not identical.`;
};

const hearthRecoveryDraftFromStructure = (
  structure: BaseStructureDefinition
): HearthRecoveryDraft => {
  const effectConfig =
    structure.effectConfig && typeof structure.effectConfig === 'object'
      ? structure.effectConfig
      : {};
  const byLevel =
    effectConfig.hearthRecoveryStatusesByLevel &&
    typeof effectConfig.hearthRecoveryStatusesByLevel === 'object'
      ? (effectConfig.hearthRecoveryStatusesByLevel as Record<string, unknown>)
      : {};
  return {
    level2Statuses: normalizeHearthRecoveryStatuses(byLevel['2']),
    level3Statuses: normalizeHearthRecoveryStatuses(byLevel['3']),
  };
};

export const Bases = () => {
  const { apiClient } = useAPI();
  const [records, setRecords] = useState<BaseRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [iconPrompt, setIconPrompt] = useState(defaultBaseIconPrompt);
  const [iconUrl, setIconUrl] = useState(
    'https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/base-discovered.png'
  );
  const [iconStatus, setIconStatus] = useState<string>('unknown');
  const [iconExists, setIconExists] = useState(false);
  const [iconRequestedAt, setIconRequestedAt] = useState<string | null>(null);
  const [iconLastModified, setIconLastModified] = useState<string | null>(null);
  const [iconStatusLoading, setIconStatusLoading] = useState(false);
  const [iconBusy, setIconBusy] = useState(false);
  const [iconMessage, setIconMessage] = useState<string | null>(null);
  const [iconError, setIconError] = useState<string | null>(null);
  const [iconPreviewNonce, setIconPreviewNonce] = useState(Date.now());
  const [isIconLightboxOpen, setIsIconLightboxOpen] = useState(false);
  const [grassTilesByKey, setGrassTilesByKey] = useState<
    Record<string, BaseGrassTileStatus>
  >({});
  const [selectedGrassKey, setSelectedGrassKey] = useState('0:0');
  const [selectedGrassKeys, setSelectedGrassKeys] = useState<string[]>(['0:0']);
  const [grassPrompt, setGrassPrompt] = useState(defaultBaseGrassPrompt);
  const [grassStatusLoading, setGrassStatusLoading] = useState(false);
  const [grassBusyKeys, setGrassBusyKeys] = useState<string[]>([]);
  const [grassMessage, setGrassMessage] = useState<string | null>(null);
  const [grassError, setGrassError] = useState<string | null>(null);
  const [grassPreviewNonce, setGrassPreviewNonce] = useState(Date.now());
  const [isGrassLightboxOpen, setIsGrassLightboxOpen] = useState(false);
  const [baseImageLightbox, setBaseImageLightbox] = useState<{
    src: string;
    alt: string;
  } | null>(null);
  const [deletingBaseId, setDeletingBaseId] = useState<string | null>(null);
  const [regeneratingBaseId, setRegeneratingBaseId] = useState<string | null>(
    null
  );
  const [baseMessage, setBaseMessage] = useState<string | null>(null);
  const [descriptionJobsByBaseId, setDescriptionJobsByBaseId] = useState<
    Record<string, BaseDescriptionGenerationJob>
  >({});
  const [structures, setStructures] = useState<BaseStructureDefinition[]>([]);
  const [structureLoading, setStructureLoading] = useState(true);
  const [generatingRoomImageKey, setGeneratingRoomImageKey] = useState<
    string | null
  >(null);
  const [generatingTopDownRoomImageKey, setGeneratingTopDownRoomImageKey] =
    useState<string | null>(null);
  const [savingStructurePromptId, setSavingStructurePromptId] = useState<
    string | null
  >(null);
  const [structurePromptDrafts, setStructurePromptDrafts] = useState<
    Record<string, { imagePrompt: string; topDownImagePrompt: string }>
  >({});
  const [savingHearthRecoveryId, setSavingHearthRecoveryId] = useState<
    string | null
  >(null);
  const [hearthRecoveryDrafts, setHearthRecoveryDrafts] = useState<
    Record<string, HearthRecoveryDraft>
  >({});

  const updateHearthRecoveryDraft = useCallback(
    (
      structure: BaseStructureDefinition,
      updater: (draft: HearthRecoveryDraft) => HearthRecoveryDraft
    ) => {
      setHearthRecoveryDrafts((prev) => {
        const current =
          prev[structure.id] ?? hearthRecoveryDraftFromStructure(structure);
        return {
          ...prev,
          [structure.id]: updater(current),
        };
      });
    },
    []
  );

  const fetchBases = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.get<BaseRecord[]>('/sonar/admin/bases');
      setRecords(Array.isArray(response) ? response : []);
    } catch (err) {
      console.error('Failed to load bases', err);
      setError('Failed to load bases.');
    } finally {
      setLoading(false);
    }
  }, [apiClient]);

  const refreshIconStatus = useCallback(
    async (showMessage = false) => {
      try {
        setIconStatusLoading(true);
        setIconError(null);
        const response = await apiClient.get<StaticThumbnailResponse>(
          '/sonar/admin/thumbnails/base/status'
        );
        const url = (response?.thumbnailUrl || '').trim();
        if (url) {
          setIconUrl(url);
        }
        setIconStatus((response?.status || 'unknown').trim() || 'unknown');
        setIconExists(Boolean(response?.exists));
        setIconRequestedAt(response?.requestedAt ? response.requestedAt : null);
        setIconLastModified(
          response?.lastModified ? response.lastModified : null
        );
        setIconPreviewNonce(Date.now());
        if (showMessage) {
          setIconMessage('Base icon status refreshed.');
        }
      } catch (err) {
        console.error('Failed to load base icon status', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to load base icon status.';
        setIconError(message);
      } finally {
        setIconStatusLoading(false);
      }
    },
    [apiClient]
  );

  const handleGenerateIcon = useCallback(async () => {
    const prompt = iconPrompt.trim();
    if (!prompt) {
      setIconError('Prompt is required.');
      return;
    }
    try {
      setIconBusy(true);
      setIconError(null);
      setIconMessage(null);
      await apiClient.post('/sonar/admin/thumbnails/base', { prompt });
      setIconMessage('Base icon queued for generation.');
      await refreshIconStatus();
    } catch (err) {
      console.error('Failed to generate base icon', err);
      const message =
        err instanceof Error ? err.message : 'Failed to generate base icon.';
      setIconError(message);
    } finally {
      setIconBusy(false);
    }
  }, [apiClient, iconPrompt, refreshIconStatus]);

  const handleDeleteIcon = useCallback(async () => {
    try {
      setIconBusy(true);
      setIconError(null);
      setIconMessage(null);
      await apiClient.delete('/sonar/admin/thumbnails/base');
      setIconMessage('Base icon deleted.');
      await refreshIconStatus();
    } catch (err) {
      console.error('Failed to delete base icon', err);
      const message =
        err instanceof Error ? err.message : 'Failed to delete base icon.';
      setIconError(message);
    } finally {
      setIconBusy(false);
    }
  }, [apiClient, refreshIconStatus]);

  const refreshGrassStatus = useCallback(
    async (showMessage = false) => {
      try {
        setGrassStatusLoading(true);
        setGrassError(null);
        const response = await apiClient.get<{ tiles?: BaseGrassTileStatus[] }>(
          '/sonar/admin/thumbnails/base-grass'
        );
        const tiles = Array.isArray(response?.tiles) ? response.tiles : [];
        const next: Record<string, BaseGrassTileStatus> = {};
        tiles.forEach((tile) => {
          next[grassTileKey(tile.gridX, tile.gridY)] = tile;
        });
        setGrassTilesByKey(next);
        setGrassPreviewNonce(Date.now());
        if (showMessage) {
          setGrassMessage('Base grass tiles refreshed.');
        }
      } catch (err) {
        console.error('Failed to load base grass tiles', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to load base grass tiles.';
        setGrassError(message);
      } finally {
        setGrassStatusLoading(false);
      }
    },
    [apiClient]
  );

  const handleGenerateGrass = useCallback(
    async (keys?: string[]) => {
      const targetKeys = Array.from(
        new Set(
          (keys && keys.length > 0 ? keys : selectedGrassKeys).filter(Boolean)
        )
      );
      if (targetKeys.length === 0) {
        setGrassError('Select at least one tile.');
        return;
      }
      const runnableKeys = targetKeys.filter(
        (key) => !grassBusyKeys.includes(key)
      );
      if (runnableKeys.length === 0) {
        return;
      }
      try {
        setGrassError(null);
        setGrassMessage(null);
        setGrassBusyKeys((prev) =>
          Array.from(new Set([...prev, ...runnableKeys]))
        );
        const results = await Promise.allSettled(
          runnableKeys.map(async (key) => {
            const [gridXText, gridYText] = key.split(':');
            const gridX = Number(gridXText);
            const gridY = Number(gridYText);
            await apiClient.post(
              `/sonar/admin/thumbnails/base-grass/${gridX}/${gridY}`,
              {
                prompt: buildGrassPromptForCell(grassPrompt, gridX, gridY),
              }
            );
            return key;
          })
        );
        const successCount = results.filter(
          (result) => result.status === 'fulfilled'
        ).length;
        const failedCount = results.length - successCount;
        if (failedCount === 0) {
          setGrassMessage(
            successCount === 1
              ? '1 grass tile queued for generation.'
              : `${successCount} grass tiles queued for generation.`
          );
        } else {
          setGrassMessage(
            `${successCount} grass ${successCount === 1 ? 'tile' : 'tiles'} queued, ${failedCount} failed.`
          );
        }
        await refreshGrassStatus();
      } catch (err) {
        console.error('Failed to generate base grass tile', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to generate base grass tile.';
        setGrassError(message);
      } finally {
        setGrassBusyKeys((prev) =>
          prev.filter((key) => !runnableKeys.includes(key))
        );
      }
    },
    [
      apiClient,
      grassBusyKeys,
      grassPrompt,
      refreshGrassStatus,
      selectedGrassKeys,
    ]
  );

  const handleDeleteGrass = useCallback(
    async (keys?: string[]) => {
      const targetKeys = Array.from(
        new Set(
          (keys && keys.length > 0 ? keys : selectedGrassKeys).filter(Boolean)
        )
      );
      if (targetKeys.length === 0) {
        setGrassError('Select at least one tile.');
        return;
      }
      const runnableKeys = targetKeys.filter(
        (key) => !grassBusyKeys.includes(key)
      );
      if (runnableKeys.length === 0) {
        return;
      }
      try {
        setGrassError(null);
        setGrassMessage(null);
        setGrassBusyKeys((prev) =>
          Array.from(new Set([...prev, ...runnableKeys]))
        );
        const results = await Promise.allSettled(
          runnableKeys.map(async (key) => {
            const [gridXText, gridYText] = key.split(':');
            const gridX = Number(gridXText);
            const gridY = Number(gridYText);
            await apiClient.delete(
              `/sonar/admin/thumbnails/base-grass/${gridX}/${gridY}`
            );
            return key;
          })
        );
        const successCount = results.filter(
          (result) => result.status === 'fulfilled'
        ).length;
        const failedCount = results.length - successCount;
        if (failedCount === 0) {
          setGrassMessage(
            successCount === 1
              ? '1 grass tile deleted.'
              : `${successCount} grass tiles deleted.`
          );
        } else {
          setGrassMessage(
            `${successCount} grass ${successCount === 1 ? 'tile' : 'tiles'} deleted, ${failedCount} failed.`
          );
        }
        await refreshGrassStatus();
      } catch (err) {
        console.error('Failed to delete base grass tile', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to delete base grass tile.';
        setGrassError(message);
      } finally {
        setGrassBusyKeys((prev) =>
          prev.filter((key) => !runnableKeys.includes(key))
        );
      }
    },
    [apiClient, grassBusyKeys, refreshGrassStatus, selectedGrassKeys]
  );

  const handleDeleteBase = useCallback(
    async (record: BaseRecord) => {
      const owner = ownerLabel(record);
      const confirmed = window.confirm(
        `Delete ${owner}'s base? This will remove the base pin from the map.`
      );
      if (!confirmed) return;

      try {
        setDeletingBaseId(record.id);
        setError(null);
        setBaseMessage(null);
        await apiClient.delete(`/sonar/admin/bases/${record.id}`);
        setRecords((prev) => prev.filter((base) => base.id !== record.id));
        setBaseMessage(`Deleted ${owner}'s base.`);
      } catch (err) {
        console.error('Failed to delete base', err);
        const message =
          err instanceof Error ? err.message : 'Failed to delete base.';
        setError(message);
      } finally {
        setDeletingBaseId(null);
      }
    },
    [apiClient]
  );

  const fetchDescriptionJobs = useCallback(async () => {
    try {
      const response = await apiClient.get<BaseDescriptionGenerationJob[]>(
        '/sonar/admin/base-description-jobs?limit=100'
      );
      const jobs = Array.isArray(response) ? response : [];
      const next: Record<string, BaseDescriptionGenerationJob> = {};
      jobs.forEach((job) => {
        if (!job.baseId) return;
        if (!next[job.baseId]) {
          next[job.baseId] = job;
        }
      });
      setDescriptionJobsByBaseId(next);
    } catch (err) {
      console.error('Failed to load base description jobs', err);
    }
  }, [apiClient]);

  const handleRegenerateDescription = useCallback(
    async (record: BaseRecord) => {
      try {
        setRegeneratingBaseId(record.id);
        setError(null);
        setBaseMessage(null);
        await apiClient.post(
          `/sonar/admin/bases/${record.id}/generate-description`
        );
        setBaseMessage(
          `Queued base flavor generation for ${ownerLabel(record)}.`
        );
        await fetchDescriptionJobs();
      } catch (err) {
        console.error('Failed to queue base description generation', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to queue base description generation.';
        setError(message);
      } finally {
        setRegeneratingBaseId(null);
      }
    },
    [apiClient, fetchDescriptionJobs]
  );

  const fetchStructures = useCallback(async () => {
    try {
      setStructureLoading(true);
      const response = await apiClient.get<{
        structures?: BaseStructureDefinition[];
      }>('/sonar/admin/base-structures');
      setStructures(
        Array.isArray(response?.structures) ? response.structures : []
      );
    } catch (err) {
      console.error('Failed to load base structures', err);
      setError('Failed to load base structures.');
    } finally {
      setStructureLoading(false);
    }
  }, [apiClient]);

  const handleGenerateRoomImage = useCallback(
    async (structure: BaseStructureDefinition, level: number) => {
      const jobKey = `${structure.id}:${level}`;
      try {
        setGeneratingRoomImageKey(jobKey);
        setError(null);
        setBaseMessage(null);
        const updated = await apiClient.post<BaseStructureDefinition>(
          `/sonar/admin/base-structures/${structure.id}/levels/${level}/generate-image`
        );
        setStructures((prev) =>
          prev.map((record) => (record.id === structure.id ? updated : record))
        );
        setBaseMessage(
          `Queued ${structure.name} level ${level} image generation.`
        );
      } catch (err) {
        console.error('Failed to queue base room image generation', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to queue base room image generation.';
        setError(message);
      } finally {
        setGeneratingRoomImageKey(null);
      }
    },
    [apiClient]
  );

  const handleGenerateTopDownRoomImage = useCallback(
    async (structure: BaseStructureDefinition, level: number) => {
      const jobKey = `${structure.id}:${level}`;
      try {
        setGeneratingTopDownRoomImageKey(jobKey);
        setError(null);
        setBaseMessage(null);
        const updated = await apiClient.post<BaseStructureDefinition>(
          `/sonar/admin/base-structures/${structure.id}/levels/${level}/generate-top-down-image`
        );
        setStructures((prev) =>
          prev.map((record) => (record.id === structure.id ? updated : record))
        );
        setBaseMessage(
          `Queued ${structure.name} level ${level} top-down image generation.`
        );
      } catch (err) {
        console.error(
          'Failed to queue base room top-down image generation',
          err
        );
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to queue base room top-down image generation.';
        setError(message);
      } finally {
        setGeneratingTopDownRoomImageKey(null);
      }
    },
    [apiClient]
  );

  const handleSaveStructurePrompts = useCallback(
    async (structure: BaseStructureDefinition) => {
      const draft = structurePromptDrafts[structure.id] || {
        imagePrompt:
          structure.imagePrompt || structure.resolvedImagePrompt || '',
        topDownImagePrompt:
          structure.topDownImagePrompt ||
          structure.resolvedTopDownImagePrompt ||
          '',
      };
      try {
        setSavingStructurePromptId(structure.id);
        setError(null);
        setBaseMessage(null);
        const updated = await apiClient.put<BaseStructureDefinition>(
          `/sonar/admin/base-structures/${structure.id}/prompts`,
          {
            imagePrompt: draft.imagePrompt,
            topDownImagePrompt: draft.topDownImagePrompt,
          }
        );
        setStructures((prev) =>
          prev.map((record) => (record.id === structure.id ? updated : record))
        );
        setStructurePromptDrafts((prev) => ({
          ...prev,
          [structure.id]: {
            imagePrompt: updated.imagePrompt || '',
            topDownImagePrompt: updated.topDownImagePrompt || '',
          },
        }));
        setBaseMessage(`Saved prompt overrides for ${structure.name}.`);
      } catch (err) {
        console.error('Failed to save base structure prompts', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to save base structure prompts.';
        setError(message);
      } finally {
        setSavingStructurePromptId(null);
      }
    },
    [apiClient, structurePromptDrafts]
  );

  const handleSaveHearthRecovery = useCallback(
    async (structure: BaseStructureDefinition) => {
      const draft =
        hearthRecoveryDrafts[structure.id] ||
        hearthRecoveryDraftFromStructure(structure);

      try {
        setSavingHearthRecoveryId(structure.id);
        setError(null);
        setBaseMessage(null);
        const updated = await apiClient.put<BaseStructureDefinition>(
          `/sonar/admin/base-structures/${structure.id}/hearth-recovery-config`,
          {
            level2Statuses: draft.level2Statuses,
            level3Statuses: draft.level3Statuses,
          }
        );
        setStructures((prev) =>
          prev.map((record) => (record.id === structure.id ? updated : record))
        );
        setHearthRecoveryDrafts((prev) => ({
          ...prev,
          [structure.id]: hearthRecoveryDraftFromStructure(updated),
        }));
        setBaseMessage(`Saved hearth recovery effects for ${structure.name}.`);
      } catch (err) {
        console.error('Failed to save hearth recovery config', err);
        const message =
          err instanceof Error
            ? err.message
            : 'Failed to save hearth recovery config.';
        setError(message);
      } finally {
        setSavingHearthRecoveryId(null);
      }
    },
    [apiClient, hearthRecoveryDrafts]
  );

  const renderHearthStatusEditor = (
    structure: BaseStructureDefinition,
    rankLabel: string,
    statuses: HearthRecoveryStatusTemplate[],
    statusKey: 'level2Statuses' | 'level3Statuses'
  ) => (
    <div className="rounded border border-gray-200 bg-gray-50 p-3">
      <div className="mb-3 flex items-center justify-between gap-3">
        <div>
          <div className="text-sm font-semibold text-gray-900">{rankLabel}</div>
          <div className="text-xs text-gray-600">
            {statusKey === 'level2Statuses'
              ? 'Applied starting at hearth rank 2.'
              : 'Added on top at hearth rank 3.'}
          </div>
        </div>
        <button
          type="button"
          onClick={() =>
            updateHearthRecoveryDraft(structure, (draft) => ({
              ...draft,
              [statusKey]: [...draft[statusKey], emptyHearthRecoveryStatus()],
            }))
          }
          className="rounded bg-emerald-600 px-3 py-2 text-xs font-medium text-white hover:bg-emerald-700"
        >
          Add Status
        </button>
      </div>

      {statuses.length === 0 ? (
        <div className="rounded border border-dashed border-gray-300 bg-white px-3 py-4 text-sm text-gray-500">
          No statuses configured.
        </div>
      ) : (
        <div className="space-y-3">
          {statuses.map((status, statusIndex) => (
            <div
              key={`${statusKey}-${statusIndex}`}
              className="rounded border border-gray-200 bg-white p-3"
            >
              <div className="mb-3 flex items-center justify-between gap-3">
                <div className="text-sm font-semibold text-gray-900">
                  {status.name.trim() || `Status ${statusIndex + 1}`}
                </div>
                <button
                  type="button"
                  onClick={() =>
                    updateHearthRecoveryDraft(structure, (draft) => ({
                      ...draft,
                      [statusKey]: draft[statusKey].filter(
                        (_, index) => index !== statusIndex
                      ),
                    }))
                  }
                  className="rounded border border-rose-300 px-3 py-1 text-xs font-medium text-rose-700 hover:bg-rose-50"
                >
                  Remove
                </button>
              </div>

              <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                <label className="text-sm font-medium text-gray-700">
                  Name
                  <input
                    value={status.name}
                    onChange={(event) =>
                      updateHearthRecoveryDraft(structure, (draft) => {
                        const nextStatuses = [...draft[statusKey]];
                        nextStatuses[statusIndex] = {
                          ...nextStatuses[statusIndex],
                          name: event.target.value,
                        };
                        return { ...draft, [statusKey]: nextStatuses };
                      })
                    }
                    className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
                  />
                </label>
                <label className="text-sm font-medium text-gray-700">
                  Duration (seconds)
                  <input
                    value={status.durationSeconds}
                    onChange={(event) =>
                      updateHearthRecoveryDraft(structure, (draft) => {
                        const nextStatuses = [...draft[statusKey]];
                        nextStatuses[statusIndex] = {
                          ...nextStatuses[statusIndex],
                          durationSeconds: parseIntValue(event.target.value, 0),
                        };
                        return { ...draft, [statusKey]: nextStatuses };
                      })
                    }
                    className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
                    type="number"
                    min={1}
                  />
                </label>
                <label className="text-sm font-medium text-gray-700">
                  Effect Type
                  <select
                    value={status.effectType}
                    onChange={(event) =>
                      updateHearthRecoveryDraft(structure, (draft) => {
                        const nextStatuses = [...draft[statusKey]];
                        nextStatuses[statusIndex] = {
                          ...nextStatuses[statusIndex],
                          effectType: event.target.value,
                        };
                        return { ...draft, [statusKey]: nextStatuses };
                      })
                    }
                    className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
                  >
                    {statusEffectTypes.map((effectType) => (
                      <option key={effectType} value={effectType}>
                        {effectType}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="flex items-center gap-2 pt-7 text-sm font-medium text-gray-700">
                  <input
                    type="checkbox"
                    checked={status.positive}
                    onChange={(event) =>
                      updateHearthRecoveryDraft(structure, (draft) => {
                        const nextStatuses = [...draft[statusKey]];
                        nextStatuses[statusIndex] = {
                          ...nextStatuses[statusIndex],
                          positive: event.target.checked,
                        };
                        return { ...draft, [statusKey]: nextStatuses };
                      })
                    }
                  />
                  Positive status
                </label>
                <label className="text-sm font-medium text-gray-700 md:col-span-2">
                  Description
                  <input
                    value={status.description}
                    onChange={(event) =>
                      updateHearthRecoveryDraft(structure, (draft) => {
                        const nextStatuses = [...draft[statusKey]];
                        nextStatuses[statusIndex] = {
                          ...nextStatuses[statusIndex],
                          description: event.target.value,
                        };
                        return { ...draft, [statusKey]: nextStatuses };
                      })
                    }
                    className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
                  />
                </label>
                <label className="text-sm font-medium text-gray-700 md:col-span-2">
                  Effect Text
                  <input
                    value={status.effect}
                    onChange={(event) =>
                      updateHearthRecoveryDraft(structure, (draft) => {
                        const nextStatuses = [...draft[statusKey]];
                        nextStatuses[statusIndex] = {
                          ...nextStatuses[statusIndex],
                          effect: event.target.value,
                        };
                        return { ...draft, [statusKey]: nextStatuses };
                      })
                    }
                    className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
                  />
                </label>
              </div>

              <div className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-3">
                <label className="text-sm font-medium text-gray-700">
                  Damage / Tick
                  <input
                    value={status.damagePerTick}
                    onChange={(event) =>
                      updateHearthRecoveryDraft(structure, (draft) => {
                        const nextStatuses = [...draft[statusKey]];
                        nextStatuses[statusIndex] = {
                          ...nextStatuses[statusIndex],
                          damagePerTick: parseIntValue(event.target.value, 0),
                        };
                        return { ...draft, [statusKey]: nextStatuses };
                      })
                    }
                    className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
                    type="number"
                  />
                </label>
                <label className="text-sm font-medium text-gray-700">
                  Health / Tick
                  <input
                    value={status.healthPerTick}
                    onChange={(event) =>
                      updateHearthRecoveryDraft(structure, (draft) => {
                        const nextStatuses = [...draft[statusKey]];
                        nextStatuses[statusIndex] = {
                          ...nextStatuses[statusIndex],
                          healthPerTick: parseIntValue(event.target.value, 0),
                        };
                        return { ...draft, [statusKey]: nextStatuses };
                      })
                    }
                    className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
                    type="number"
                  />
                </label>
                <label className="text-sm font-medium text-gray-700">
                  Mana / Tick
                  <input
                    value={status.manaPerTick}
                    onChange={(event) =>
                      updateHearthRecoveryDraft(structure, (draft) => {
                        const nextStatuses = [...draft[statusKey]];
                        nextStatuses[statusIndex] = {
                          ...nextStatuses[statusIndex],
                          manaPerTick: parseIntValue(event.target.value, 0),
                        };
                        return { ...draft, [statusKey]: nextStatuses };
                      })
                    }
                    className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
                    type="number"
                  />
                </label>
              </div>

              <div className="mt-3 grid grid-cols-2 gap-3 md:grid-cols-6">
                {[
                  ['strengthMod', 'STR'],
                  ['dexterityMod', 'DEX'],
                  ['constitutionMod', 'CON'],
                  ['intelligenceMod', 'INT'],
                  ['wisdomMod', 'WIS'],
                  ['charismaMod', 'CHA'],
                ].map(([field, label]) => (
                  <label
                    key={`${statusKey}-${statusIndex}-${field}`}
                    className="text-xs font-medium text-gray-700"
                  >
                    {label}
                    <input
                      value={
                        status[
                          field as keyof HearthRecoveryStatusTemplate
                        ] as number
                      }
                      onChange={(event) =>
                        updateHearthRecoveryDraft(structure, (draft) => {
                          const nextStatuses = [...draft[statusKey]];
                          nextStatuses[statusIndex] = {
                            ...nextStatuses[statusIndex],
                            [field]: parseIntValue(event.target.value, 0),
                          };
                          return { ...draft, [statusKey]: nextStatuses };
                        })
                      }
                      className="mt-1 w-full rounded border border-gray-300 px-2 py-2 text-sm"
                      type="number"
                    />
                  </label>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );

  useEffect(() => {
    void fetchBases();
  }, [fetchBases]);

  useEffect(() => {
    void fetchStructures();
  }, [fetchStructures]);

  useEffect(() => {
    void refreshIconStatus();
  }, [refreshIconStatus]);

  useEffect(() => {
    void refreshGrassStatus();
  }, [refreshGrassStatus]);

  useEffect(() => {
    const selected = grassTilesByKey[selectedGrassKey];
    if (selected?.prompt?.trim()) {
      setGrassPrompt(selected.prompt.trim());
      return;
    }
    const [gridXText, gridYText] = selectedGrassKey.split(':');
    setGrassPrompt(
      defaultGrassPromptForCell(Number(gridXText), Number(gridYText))
    );
  }, [grassTilesByKey, selectedGrassKey]);

  useEffect(() => {
    void fetchDescriptionJobs();
  }, [fetchDescriptionJobs]);

  useEffect(() => {
    if (iconStatus !== 'queued' && iconStatus !== 'in_progress') {
      return;
    }
    const interval = window.setInterval(() => {
      void refreshIconStatus();
    }, 4000);
    return () => window.clearInterval(interval);
  }, [iconStatus, refreshIconStatus]);

  useEffect(() => {
    const hasPendingGrassTiles = Object.values(grassTilesByKey).some((tile) =>
      ['queued', 'in_progress'].includes((tile.status || '').toLowerCase())
    );
    if (!hasPendingGrassTiles) {
      return;
    }
    const interval = window.setInterval(() => {
      void refreshGrassStatus();
    }, 4000);
    return () => window.clearInterval(interval);
  }, [grassTilesByKey, refreshGrassStatus]);

  useEffect(() => {
    const hasPendingJobs = Object.values(descriptionJobsByBaseId).some((job) =>
      ['queued', 'in_progress'].includes((job.status || '').toLowerCase())
    );
    const hasPendingRoomImages = structures.some((structure) =>
      (structure.levelVisuals || []).some(
        (visual) =>
          ['queued', 'in_progress'].includes(
            (visual.imageGenerationStatus || '').toLowerCase()
          ) ||
          ['queued', 'in_progress'].includes(
            (visual.topDownImageGenerationStatus || '').toLowerCase()
          )
      )
    );
    if (!hasPendingJobs && !hasPendingRoomImages) {
      return;
    }
    const interval = window.setInterval(() => {
      void fetchDescriptionJobs();
      void fetchBases();
      void fetchStructures();
    }, 4000);
    return () => window.clearInterval(interval);
  }, [
    descriptionJobsByBaseId,
    fetchBases,
    fetchDescriptionJobs,
    fetchStructures,
    structures,
  ]);

  if (loading) {
    return <div className="m-10">Loading bases...</div>;
  }

  const selectedGrassTile = grassTilesByKey[selectedGrassKey];
  const selectedGrassCount = selectedGrassKeys.length;
  const selectedGrassSummary = selectedGrassKeys
    .map((key) => {
      const [gridXText, gridYText] = key.split(':');
      return `(${Number(gridXText) + 1}, ${Number(gridYText) + 1})`;
    })
    .join(', ');
  const selectedGrassUrl =
    selectedGrassTile?.thumbnailUrl?.trim() ||
    `https://crew-profile-icons.s3.amazonaws.com/thumbnails/placeholders/base-grass-${selectedGrassKey.replace(
      ':',
      '-'
    )}.png`;
  const selectedGrassStatus =
    (selectedGrassTile?.status || 'unknown').trim() || 'unknown';
  const selectedGrassExists = Boolean(selectedGrassTile?.exists);
  const selectedGrassRequestedAt = selectedGrassTile?.requestedAt || null;
  const selectedGrassLastModified = selectedGrassTile?.lastModified || null;

  return (
    <div className="m-10 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Bases</h1>
        <button
          type="button"
          onClick={() => void fetchBases()}
          className="rounded bg-blue-600 px-3 py-2 text-white hover:bg-blue-700"
        >
          Refresh Bases
        </button>
      </div>

      <p className="text-sm text-gray-600">
        Bases are player-owned map pins created in the app. Shared generated art
        assets for the map icon and the base board grass tile are managed here.
      </p>

      <section className="rounded border border-gray-200 bg-white p-4 shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-sm font-semibold text-gray-900">
              Base Pin Icon
            </h2>
            <p className="mt-1 text-xs text-gray-600">
              Requested: {formatDate(iconRequestedAt ?? undefined)}
            </p>
            <p className="text-xs text-gray-600">
              Last updated: {formatDate(iconLastModified ?? undefined)}
            </p>
          </div>
          <span
            className={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide text-white ${staticStatusClassName(
              iconStatus
            )}`}
          >
            {iconStatus || 'unknown'}
          </span>
        </div>

        <div className="mt-4 flex flex-wrap gap-3">
          <button
            type="button"
            onClick={() => void refreshIconStatus(true)}
            disabled={iconStatusLoading}
            className="rounded bg-slate-700 px-3 py-2 text-sm text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {iconStatusLoading ? 'Refreshing...' : 'Refresh Status'}
          </button>
          <button
            type="button"
            onClick={() => void handleGenerateIcon()}
            disabled={iconBusy || iconStatusLoading}
            className="rounded bg-emerald-600 px-3 py-2 text-sm text-white hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {iconBusy ? 'Working...' : 'Generate Icon'}
          </button>
          <button
            type="button"
            onClick={() => void handleDeleteIcon()}
            disabled={iconBusy || iconStatusLoading}
            className="rounded bg-red-600 px-3 py-2 text-sm text-white hover:bg-red-700 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {iconBusy ? 'Working...' : 'Delete Icon'}
          </button>
        </div>

        <label className="mt-4 block text-sm font-medium text-gray-700">
          Prompt
          <textarea
            value={iconPrompt}
            onChange={(e) => setIconPrompt(e.target.value)}
            rows={4}
            className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
          />
        </label>

        <p className="mt-3 text-xs text-gray-600 break-all">URL: {iconUrl}</p>

        {iconExists ? (
          <div className="mt-4 flex justify-center rounded border border-dashed border-gray-300 bg-gray-50 p-4">
            <button
              type="button"
              onClick={() => setIsIconLightboxOpen(true)}
              className="rounded focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
              title="Open large preview"
            >
              <img
                src={`${iconUrl}?v=${iconPreviewNonce}`}
                alt="Base icon preview"
                className="h-28 w-28 rounded object-contain"
              />
            </button>
          </div>
        ) : (
          <div className="mt-4 rounded border border-dashed border-gray-300 bg-gray-50 p-4 text-sm text-gray-500">
            No generated icon found yet.
          </div>
        )}

        {iconMessage ? (
          <p className="mt-3 text-sm text-emerald-700">{iconMessage}</p>
        ) : null}
        {iconError ? (
          <p className="mt-3 text-sm text-red-700">{iconError}</p>
        ) : null}
      </section>

      <section className="rounded border border-gray-200 bg-white p-4 shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-sm font-semibold text-gray-900">
              Base Grass Tiles
            </h2>
            <p className="mt-1 text-xs text-gray-600">
              Click tiles to build a batch. The last tile you click is the one
              shown in the detail panel.
            </p>
          </div>
          <button
            type="button"
            onClick={() => void refreshGrassStatus(true)}
            disabled={grassStatusLoading}
            className="rounded bg-slate-700 px-3 py-2 text-sm text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {grassStatusLoading ? 'Refreshing...' : 'Refresh Grid'}
          </button>
        </div>

        <div className="mt-4 grid gap-1 rounded border border-gray-200 bg-gray-50 p-3 md:w-fit">
          {Array.from({ length: 5 }).map((_, gridY) => (
            <div key={gridY} className="flex gap-1">
              {Array.from({ length: 5 }).map((_, gridX) => {
                const key = grassTileKey(gridX, gridY);
                const tile = grassTilesByKey[key];
                const status = (tile?.status || 'unknown').trim() || 'unknown';
                const isSelected = selectedGrassKey === key;
                const isInBatch = selectedGrassKeys.includes(key);
                const isBusy = grassBusyKeys.includes(key);
                return (
                  <button
                    key={key}
                    type="button"
                    onClick={() => {
                      setSelectedGrassKey(key);
                      setSelectedGrassKeys((prev) =>
                        prev.includes(key)
                          ? prev.filter((entry) => entry !== key)
                          : [...prev, key]
                      );
                    }}
                    className={`flex h-14 w-14 flex-col items-center justify-center rounded border text-[10px] font-semibold ${
                      isSelected && isInBatch
                        ? 'border-blue-600 bg-blue-50 text-blue-700'
                        : isSelected
                          ? 'border-sky-500 bg-sky-50 text-sky-700'
                          : isInBatch
                            ? 'border-emerald-500 bg-emerald-50 text-emerald-700'
                            : 'border-gray-200 bg-white text-gray-700 hover:border-gray-300'
                    }`}
                    title={`Tile (${gridX + 1}, ${gridY + 1})`}
                  >
                    <span>
                      {gridX + 1},{gridY + 1}
                    </span>
                    <div className="mt-1 flex items-center gap-1">
                      <span
                        className={`h-2 w-2 rounded-full ${staticStatusClassName(
                          status
                        )}`}
                      />
                      {isBusy ? (
                        <span className="text-[8px] uppercase">Q</span>
                      ) : null}
                    </div>
                  </button>
                );
              })}
            </div>
          ))}
        </div>

        <div className="mt-4 rounded border border-gray-200 bg-gray-50 p-4">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <p className="text-xs font-semibold uppercase tracking-wide text-gray-500">
                Batch selection
              </p>
              <p className="mt-1 text-sm text-gray-900">
                {selectedGrassCount === 0
                  ? 'No tiles selected.'
                  : `${selectedGrassCount} tile${selectedGrassCount === 1 ? '' : 's'} selected`}
              </p>
              {selectedGrassCount > 0 ? (
                <p className="mt-1 max-w-xl text-xs text-gray-600">
                  {selectedGrassSummary}
                </p>
              ) : null}
              <div className="mt-3 flex flex-wrap gap-3">
                <button
                  type="button"
                  onClick={() => void handleGenerateGrass(selectedGrassKeys)}
                  disabled={selectedGrassCount === 0}
                  className="rounded bg-emerald-600 px-3 py-2 text-sm text-white hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {selectedGrassCount <= 1
                    ? 'Generate Selected'
                    : `Generate ${selectedGrassCount} Tiles`}
                </button>
                <button
                  type="button"
                  onClick={() => void handleDeleteGrass(selectedGrassKeys)}
                  disabled={selectedGrassCount === 0}
                  className="rounded bg-red-600 px-3 py-2 text-sm text-white hover:bg-red-700 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {selectedGrassCount <= 1
                    ? 'Delete Selected'
                    : `Delete ${selectedGrassCount} Tiles`}
                </button>
                <button
                  type="button"
                  onClick={() =>
                    setSelectedGrassKeys(
                      selectedGrassKey ? [selectedGrassKey] : []
                    )
                  }
                  disabled={selectedGrassCount <= 1}
                  className="rounded bg-white px-3 py-2 text-sm text-gray-700 ring-1 ring-gray-300 hover:bg-gray-100 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  Keep Only Focused Tile
                </button>
                <button
                  type="button"
                  onClick={() => setSelectedGrassKeys([])}
                  disabled={selectedGrassCount === 0}
                  className="rounded bg-white px-3 py-2 text-sm text-gray-700 ring-1 ring-gray-300 hover:bg-gray-100 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  Clear Selection
                </button>
              </div>
            </div>
            <div>
              <h3 className="text-sm font-semibold text-gray-900">
                Tile ({Number(selectedGrassKey.split(':')[0]) + 1},{' '}
                {Number(selectedGrassKey.split(':')[1]) + 1})
              </h3>
              <p className="mt-1 text-xs text-gray-600">
                Requested: {formatDate(selectedGrassRequestedAt ?? undefined)}
              </p>
              <p className="text-xs text-gray-600">
                Last updated:{' '}
                {formatDate(selectedGrassLastModified ?? undefined)}
              </p>
            </div>
            <span
              className={`rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide text-white ${staticStatusClassName(
                selectedGrassStatus
              )}`}
            >
              {selectedGrassStatus}
            </span>
          </div>

          <div className="mt-4 flex flex-wrap gap-3">
            <button
              type="button"
              onClick={() => void handleGenerateGrass()}
              disabled={grassBusyKeys.includes(selectedGrassKey)}
              className="rounded bg-emerald-600 px-3 py-2 text-sm text-white hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {grassBusyKeys.includes(selectedGrassKey)
                ? 'Working...'
                : 'Generate Focused Tile'}
            </button>
            <button
              type="button"
              onClick={() => void handleDeleteGrass()}
              disabled={grassBusyKeys.includes(selectedGrassKey)}
              className="rounded bg-red-600 px-3 py-2 text-sm text-white hover:bg-red-700 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {grassBusyKeys.includes(selectedGrassKey)
                ? 'Working...'
                : 'Delete Focused Tile'}
            </button>
          </div>

          <label className="mt-4 block text-sm font-medium text-gray-700">
            Prompt
            <textarea
              value={grassPrompt}
              onChange={(e) => setGrassPrompt(e.target.value)}
              rows={4}
              className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
            />
          </label>

          <p className="mt-3 break-all text-xs text-gray-600">
            URL: {selectedGrassUrl}
          </p>

          {selectedGrassExists ? (
            <div className="mt-4 flex justify-center rounded border border-dashed border-gray-300 bg-white p-4">
              <button
                type="button"
                onClick={() => setIsGrassLightboxOpen(true)}
                className="rounded focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                title="Open large preview"
              >
                <img
                  src={`${selectedGrassUrl}?v=${grassPreviewNonce}`}
                  alt="Base grass tile preview"
                  className="h-28 w-28 rounded object-cover"
                />
              </button>
            </div>
          ) : (
            <div className="mt-4 rounded border border-dashed border-gray-300 bg-white p-4 text-sm text-gray-500">
              No generated grass tile found for this coordinate yet.
            </div>
          )}
        </div>

        {grassMessage ? (
          <p className="mt-3 text-sm text-emerald-700">{grassMessage}</p>
        ) : null}
        {grassError ? (
          <p className="mt-3 text-sm text-red-700">{grassError}</p>
        ) : null}
      </section>

      {isIconLightboxOpen ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-6"
          onClick={() => setIsIconLightboxOpen(false)}
        >
          <div
            className="relative max-h-[90vh] max-w-[90vw] rounded-lg bg-white p-4 shadow-2xl"
            onClick={(event) => event.stopPropagation()}
          >
            <button
              type="button"
              onClick={() => setIsIconLightboxOpen(false)}
              className="absolute right-3 top-3 rounded bg-black/70 px-2 py-1 text-xs font-semibold text-white hover:bg-black/80"
            >
              Close
            </button>
            <img
              src={`${iconUrl}?v=${iconPreviewNonce}`}
              alt="Large base icon preview"
              className="max-h-[80vh] max-w-[80vw] rounded object-contain"
            />
          </div>
        </div>
      ) : null}

      {isGrassLightboxOpen ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-6"
          onClick={() => setIsGrassLightboxOpen(false)}
        >
          <div
            className="relative max-h-[90vh] max-w-[90vw] rounded-lg bg-white p-4 shadow-2xl"
            onClick={(event) => event.stopPropagation()}
          >
            <button
              type="button"
              onClick={() => setIsGrassLightboxOpen(false)}
              className="absolute right-3 top-3 rounded bg-black/70 px-2 py-1 text-xs font-semibold text-white hover:bg-black/80"
            >
              Close
            </button>
            <img
              src={`${selectedGrassUrl}?v=${grassPreviewNonce}`}
              alt="Large base grass tile preview"
              className="max-h-[80vh] max-w-[80vw] rounded object-contain"
            />
          </div>
        </div>
      ) : null}

      {baseImageLightbox ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-6"
          onClick={() => setBaseImageLightbox(null)}
        >
          <div
            className="relative max-h-[90vh] max-w-[90vw] rounded-lg bg-white p-4 shadow-2xl"
            onClick={(event) => event.stopPropagation()}
          >
            <button
              type="button"
              onClick={() => setBaseImageLightbox(null)}
              className="absolute right-3 top-3 rounded bg-black/70 px-2 py-1 text-xs font-semibold text-white hover:bg-black/80"
            >
              Close
            </button>
            <img
              src={baseImageLightbox.src}
              alt={baseImageLightbox.alt}
              className="max-h-[80vh] max-w-[80vw] rounded object-contain"
            />
          </div>
        </div>
      ) : null}

      {error ? (
        <div className="rounded border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      ) : null}
      {baseMessage ? (
        <div className="rounded border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
          {baseMessage}
        </div>
      ) : null}

      <section className="rounded border border-gray-200 bg-white p-4 shadow-sm">
        <div className="flex items-center justify-between gap-3">
          <div>
            <h2 className="text-sm font-semibold text-gray-900">
              Base Room Images
            </h2>
            <p className="mt-1 text-xs text-gray-600">
              Queue level-by-level room art for the base management screen.
            </p>
          </div>
          <button
            type="button"
            onClick={() => void fetchStructures()}
            disabled={structureLoading}
            className="rounded bg-slate-700 px-3 py-2 text-sm text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {structureLoading ? 'Refreshing...' : 'Refresh Rooms'}
          </button>
        </div>

        {structureLoading ? (
          <div className="mt-4 text-sm text-gray-500">
            Loading room images...
          </div>
        ) : structures.length === 0 ? (
          <div className="mt-4 text-sm text-gray-500">No base rooms found.</div>
        ) : (
          <div className="mt-4 space-y-4">
            {structures.map((structure) => (
              <div
                key={structure.id}
                className="rounded border border-gray-200 bg-gray-50 p-4"
              >
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <h3 className="text-sm font-semibold text-gray-900">
                      {structure.name}
                    </h3>
                    <p className="mt-1 text-xs uppercase tracking-wide text-gray-500">
                      {(structure.category || 'room').trim() || 'room'}
                    </p>
                    {structure.description?.trim() ? (
                      <p className="mt-2 max-w-3xl text-sm text-gray-600">
                        {structure.description.trim()}
                      </p>
                    ) : null}
                  </div>
                  <span
                    className={`rounded-full px-2 py-1 text-xs font-semibold ${
                      structure.active === false
                        ? 'bg-amber-100 text-amber-800'
                        : 'bg-emerald-100 text-emerald-700'
                    }`}
                  >
                    {structure.active === false ? 'Inactive' : 'Active'}
                  </span>
                </div>

                <div className="mt-4 rounded border border-gray-200 bg-white p-3">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <h4 className="text-sm font-semibold text-gray-900">
                        Prompt Overrides
                      </h4>
                      <p className="mt-1 text-xs text-gray-600">
                        These saved prompts will be used for future room image
                        jobs.
                      </p>
                    </div>
                    <button
                      type="button"
                      onClick={() => void handleSaveStructurePrompts(structure)}
                      disabled={savingStructurePromptId === structure.id}
                      className="rounded bg-slate-800 px-3 py-2 text-sm text-white hover:bg-slate-900 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      {savingStructurePromptId === structure.id
                        ? 'Saving...'
                        : 'Save Prompts'}
                    </button>
                  </div>
                  <div className="mt-3 grid gap-3 lg:grid-cols-2">
                    <label className="block text-sm font-medium text-gray-700">
                      Card View Prompt
                      <textarea
                        value={
                          structurePromptDrafts[structure.id]?.imagePrompt ??
                          structure.resolvedImagePrompt ??
                          structure.imagePrompt ??
                          ''
                        }
                        onChange={(event) =>
                          setStructurePromptDrafts((prev) => ({
                            ...prev,
                            [structure.id]: {
                              imagePrompt: event.target.value,
                              topDownImagePrompt:
                                prev[structure.id]?.topDownImagePrompt ??
                                structure.resolvedTopDownImagePrompt ??
                                structure.topDownImagePrompt ??
                                '',
                            },
                          }))
                        }
                        rows={5}
                        className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
                      />
                    </label>
                    <label className="block text-sm font-medium text-gray-700">
                      Top-Down Prompt
                      <textarea
                        value={
                          structurePromptDrafts[structure.id]
                            ?.topDownImagePrompt ??
                          structure.resolvedTopDownImagePrompt ??
                          structure.topDownImagePrompt ??
                          ''
                        }
                        onChange={(event) =>
                          setStructurePromptDrafts((prev) => ({
                            ...prev,
                            [structure.id]: {
                              imagePrompt:
                                prev[structure.id]?.imagePrompt ??
                                structure.resolvedImagePrompt ??
                                structure.imagePrompt ??
                                '',
                              topDownImagePrompt: event.target.value,
                            },
                          }))
                        }
                        rows={7}
                        className="mt-1 w-full rounded border border-gray-300 px-3 py-2 text-sm"
                      />
                    </label>
                  </div>
                </div>

                {structure.key === 'hearth' ? (
                  <div className="mt-4 rounded border border-gray-200 bg-white p-3">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <h4 className="text-sm font-semibold text-gray-900">
                          Hearth Recovery Effects
                        </h4>
                        <p className="mt-1 text-xs text-gray-600">
                          Rank 2 applies these statuses. Rank 3 applies both
                          rank 2 and rank 3 statuses.
                        </p>
                      </div>
                      <button
                        type="button"
                        onClick={() => void handleSaveHearthRecovery(structure)}
                        disabled={savingHearthRecoveryId === structure.id}
                        className="rounded bg-slate-800 px-3 py-2 text-sm text-white hover:bg-slate-900 disabled:cursor-not-allowed disabled:opacity-60"
                      >
                        {savingHearthRecoveryId === structure.id
                          ? 'Saving...'
                          : 'Save Hearth Effects'}
                      </button>
                    </div>
                    <div className="mt-3 grid gap-3 lg:grid-cols-2">
                      {renderHearthStatusEditor(
                        structure,
                        'Rank 2 Statuses',
                        hearthRecoveryDrafts[structure.id]?.level2Statuses ??
                          hearthRecoveryDraftFromStructure(structure)
                            .level2Statuses,
                        'level2Statuses'
                      )}
                      {renderHearthStatusEditor(
                        structure,
                        'Rank 3 Extra Statuses',
                        hearthRecoveryDrafts[structure.id]?.level3Statuses ??
                          hearthRecoveryDraftFromStructure(structure)
                            .level3Statuses,
                        'level3Statuses'
                      )}
                    </div>
                  </div>
                ) : null}

                <div className="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                  {(structure.levelVisuals || []).map((visual) => {
                    const previewUrl =
                      visual.thumbnailUrl?.trim() ||
                      visual.imageUrl?.trim() ||
                      '';
                    const topDownPreviewUrl =
                      visual.topDownThumbnailUrl?.trim() ||
                      visual.topDownImageUrl?.trim() ||
                      '';
                    const visualKey = `${structure.id}:${visual.level}`;
                    const isPending = ['queued', 'in_progress'].includes(
                      (visual.imageGenerationStatus || '').toLowerCase()
                    );
                    const isTopDownPending = ['queued', 'in_progress'].includes(
                      (visual.topDownImageGenerationStatus || '').toLowerCase()
                    );
                    return (
                      <div
                        key={visualKey}
                        className="rounded border border-gray-200 bg-white p-3 shadow-sm"
                      >
                        <div className="flex items-center justify-between gap-3">
                          <div className="text-sm font-semibold text-gray-900">
                            Level {visual.level}
                          </div>
                          <span
                            className={`rounded-full px-2 py-1 text-[11px] font-semibold uppercase tracking-wide text-white ${staticStatusClassName(
                              visual.imageGenerationStatus
                            )}`}
                          >
                            {visual.imageGenerationStatus || 'none'}
                          </span>
                        </div>
                        <div className="mt-3 grid gap-3 md:grid-cols-2">
                          <div>
                            <div className="mb-2 flex items-center justify-between gap-2">
                              <div className="text-xs font-semibold uppercase tracking-wide text-gray-500">
                                Card View
                              </div>
                              <span
                                className={`rounded-full px-2 py-1 text-[10px] font-semibold uppercase tracking-wide text-white ${staticStatusClassName(
                                  visual.imageGenerationStatus
                                )}`}
                              >
                                {visual.imageGenerationStatus || 'none'}
                              </span>
                            </div>
                            {previewUrl ? (
                              <button
                                type="button"
                                onClick={() =>
                                  setBaseImageLightbox({
                                    src: visual.imageUrl?.trim() || previewUrl,
                                    alt: `${structure.name} level ${visual.level} card view`,
                                  })
                                }
                                className="rounded focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                                title="Open large room image"
                              >
                                <img
                                  src={previewUrl}
                                  alt={`${structure.name} level ${visual.level} card view`}
                                  className="h-36 w-full rounded object-cover"
                                />
                              </button>
                            ) : (
                              <div className="flex h-36 w-full items-center justify-center rounded border border-dashed border-gray-300 bg-gray-50 text-sm text-gray-500">
                                No image yet
                              </div>
                            )}
                            {visual.imageGenerationError ? (
                              <p className="mt-3 text-xs text-red-700">
                                {visual.imageGenerationError}
                              </p>
                            ) : null}
                            <div className="mt-3 flex justify-end">
                              <button
                                type="button"
                                onClick={() =>
                                  void handleGenerateRoomImage(
                                    structure,
                                    visual.level
                                  )
                                }
                                disabled={
                                  generatingRoomImageKey === visualKey ||
                                  isPending
                                }
                                className="rounded bg-emerald-600 px-3 py-2 text-sm text-white hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-60"
                              >
                                {generatingRoomImageKey === visualKey
                                  ? 'Queueing...'
                                  : 'Generate Card View'}
                              </button>
                            </div>
                          </div>
                          <div>
                            <div className="mb-2 flex items-center justify-between gap-2">
                              <div className="text-xs font-semibold uppercase tracking-wide text-gray-500">
                                Top-Down View
                              </div>
                              <span
                                className={`rounded-full px-2 py-1 text-[10px] font-semibold uppercase tracking-wide text-white ${staticStatusClassName(
                                  visual.topDownImageGenerationStatus
                                )}`}
                              >
                                {visual.topDownImageGenerationStatus || 'none'}
                              </span>
                            </div>
                            {topDownPreviewUrl ? (
                              <button
                                type="button"
                                onClick={() =>
                                  setBaseImageLightbox({
                                    src:
                                      visual.topDownImageUrl?.trim() ||
                                      topDownPreviewUrl,
                                    alt: `${structure.name} level ${visual.level} top-down view`,
                                  })
                                }
                                className="rounded focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                                title="Open large top-down room image"
                              >
                                <img
                                  src={topDownPreviewUrl}
                                  alt={`${structure.name} level ${visual.level} top-down view`}
                                  className="h-36 w-full rounded object-cover"
                                />
                              </button>
                            ) : (
                              <div className="flex h-36 w-full items-center justify-center rounded border border-dashed border-gray-300 bg-gray-50 text-sm text-gray-500">
                                No top-down image yet
                              </div>
                            )}
                            {visual.topDownImageGenerationError ? (
                              <p className="mt-3 text-xs text-red-700">
                                {visual.topDownImageGenerationError}
                              </p>
                            ) : null}
                            <div className="mt-3 flex justify-end">
                              <button
                                type="button"
                                onClick={() =>
                                  void handleGenerateTopDownRoomImage(
                                    structure,
                                    visual.level
                                  )
                                }
                                disabled={
                                  generatingTopDownRoomImageKey === visualKey ||
                                  isTopDownPending
                                }
                                className="rounded bg-sky-600 px-3 py-2 text-sm text-white hover:bg-sky-700 disabled:cursor-not-allowed disabled:opacity-60"
                              >
                                {generatingTopDownRoomImageKey === visualKey
                                  ? 'Queueing...'
                                  : 'Generate Top-Down'}
                              </button>
                            </div>
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      <section className="space-y-3">
        {records.length === 0 ? (
          <div className="rounded border border-gray-200 bg-white px-4 py-6 text-sm text-gray-500 shadow-sm">
            No bases created yet.
          </div>
        ) : (
          records.map((record) => {
            const latestJob = descriptionJobsByBaseId[record.id];
            const generatedImageUrl =
              record.imageUrl?.trim() ||
              latestJob?.generatedImageUrl?.trim() ||
              '';
            return (
              <div
                key={record.id}
                className="rounded border border-gray-200 bg-white p-4 shadow-sm"
              >
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <h3 className="text-sm font-semibold text-gray-900">
                      {ownerLabel(record)}
                    </h3>
                    {secondaryOwnerLabel(record) ? (
                      <p className="text-xs text-gray-500">
                        {secondaryOwnerLabel(record)}
                      </p>
                    ) : null}
                  </div>
                  <div className="text-xs text-gray-500">
                    Updated {formatDate(record.updatedAt)}
                  </div>
                </div>
                {generatedImageUrl ? (
                  <div className="mt-3">
                    <button
                      type="button"
                      onClick={() =>
                        setBaseImageLightbox({
                          src: generatedImageUrl,
                          alt: `${ownerLabel(record)} base`,
                        })
                      }
                      className="rounded focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                      title="Open large base image"
                    >
                      <img
                        src={generatedImageUrl}
                        alt={`${ownerLabel(record)} base`}
                        className="h-40 w-full max-w-xs rounded object-cover"
                      />
                    </button>
                  </div>
                ) : record.thumbnailUrl?.trim() ? (
                  <div className="mt-3">
                    <img
                      src={record.thumbnailUrl}
                      alt={`${ownerLabel(record)} base`}
                      className="h-28 w-28 rounded object-cover"
                    />
                  </div>
                ) : null}
                {record.description?.trim() ? (
                  <p className="mt-3 text-sm leading-6 text-gray-700">
                    {record.description.trim()}
                  </p>
                ) : (
                  <p className="mt-3 text-sm italic text-gray-500">
                    No description generated yet.
                  </p>
                )}
                {latestJob ? (
                  <div className="mt-3 flex flex-wrap items-center gap-2 text-xs">
                    <span
                      className={`rounded-full px-2 py-1 font-semibold uppercase tracking-wide text-white ${staticStatusClassName(
                        latestJob.status
                      )}`}
                    >
                      {latestJob.status || 'unknown'}
                    </span>
                    <span className="text-gray-500">
                      {formatDate(latestJob.updatedAt)}
                    </span>
                    {latestJob.errorMessage ? (
                      <span className="text-red-700">
                        {latestJob.errorMessage}
                      </span>
                    ) : null}
                  </div>
                ) : null}
                <div className="mt-3 grid gap-2 text-sm text-gray-700 md:grid-cols-3">
                  <div>Latitude: {record.latitude.toFixed(6)}</div>
                  <div>Longitude: {record.longitude.toFixed(6)}</div>
                  <div>User ID: {record.userId}</div>
                </div>
                <div className="mt-4 flex justify-end gap-3">
                  <button
                    type="button"
                    onClick={() => void handleRegenerateDescription(record)}
                    disabled={regeneratingBaseId === record.id}
                    className="rounded bg-emerald-600 px-3 py-2 text-sm text-white hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {regeneratingBaseId === record.id
                      ? 'Queueing...'
                      : 'Regenerate Flavor'}
                  </button>
                  <button
                    type="button"
                    onClick={() => void handleDeleteBase(record)}
                    disabled={deletingBaseId === record.id}
                    className="rounded bg-red-600 px-3 py-2 text-sm text-white hover:bg-red-700 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {deletingBaseId === record.id
                      ? 'Deleting...'
                      : 'Delete Base'}
                  </button>
                </div>
              </div>
            );
          })
        )}
      </section>
    </div>
  );
};

export default Bases;
