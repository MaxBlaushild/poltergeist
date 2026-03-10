import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI, useZoneContext } from '@poltergeist/contexts';
import { Zone } from '@poltergeist/types';

type ZoneSeedPointOfInterestDraft = {
  draftId: string;
  placeId: string;
  name: string;
  address?: string;
  types?: string[];
  latitude?: number;
  longitude?: number;
  rating?: number;
  userRatingCount?: number;
  editorialSummary?: string;
};

type ZoneSeedCharacterDraft = {
  draftId: string;
  name: string;
  description: string;
  placeId: string;
  latitude?: number;
  longitude?: number;
  shopItemTags?: string[];
};

type ZoneSeedDraft = {
  fantasyName?: string;
  zoneDescription?: string;
  pointsOfInterest?: ZoneSeedPointOfInterestDraft[];
  characters?: ZoneSeedCharacterDraft[];
};

type ZoneSeedJob = {
  id: string;
  zoneId: string;
  status: string;
  errorMessage?: string;
  placeCount: number;
  characterCount: number;
  questCount: number;
  mainQuestCount: number;
  monsterCount: number;
  inputEncounterCount: number;
  optionEncounterCount: number;
  treasureChestCount?: number;
  healingFountainCount?: number;
  requiredPlaceTags?: string[];
  shopkeeperItemTags?: string[];
  createdAt?: string;
  updatedAt?: string;
  draft?: ZoneSeedDraft;
};

type ZoneSeedDraftPayload = {
  placeCount: number;
  monsterCount: number;
  inputEncounterCount: number;
  optionEncounterCount: number;
  treasureChestCount: number;
  healingFountainCount: number;
  requiredPlaceTags: string[];
  shopkeeperItemTags: string[];
};

type BulkQueueZoneSeedJobsResponse = {
  queuedCount: number;
  requestedZoneCount: number;
  jobs: ZoneSeedJob[];
};

const statusBadgeClass = (status: string) => {
  switch (status) {
    case 'queued':
      return 'bg-slate-600';
    case 'in_progress':
      return 'bg-amber-600';
    case 'awaiting_approval':
      return 'bg-indigo-600';
    case 'approved':
      return 'bg-indigo-700';
    case 'applying':
      return 'bg-amber-700';
    case 'applied':
      return 'bg-emerald-600';
    case 'failed':
      return 'bg-red-600';
    default:
      return 'bg-gray-600';
  }
};

const formatDate = (value?: string) => {
  if (!value) return 'n/a';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

const buildSeedDraftPayload = (params: {
  placeCount: string;
  monsterCount: string;
  inputEncounterCount: string;
  optionEncounterCount: string;
  treasureChestCount: string;
  healingFountainCount: string;
  requiredPlaceTags: string[];
  shopkeeperItemTags: string[];
}): { payload?: ZoneSeedDraftPayload; error?: string } => {
  const placeCount = Number.parseInt(params.placeCount, 10);
  const monsterCount = Number.parseInt(params.monsterCount, 10);
  const inputEncounterCount = Number.parseInt(params.inputEncounterCount, 10);
  const optionEncounterCount = Number.parseInt(params.optionEncounterCount, 10);
  const treasureChestCount = Number.parseInt(params.treasureChestCount, 10);
  const healingFountainCount = Number.parseInt(params.healingFountainCount, 10);

  if (
    Number.isNaN(placeCount) ||
    Number.isNaN(monsterCount) ||
    Number.isNaN(inputEncounterCount) ||
    Number.isNaN(optionEncounterCount) ||
    Number.isNaN(treasureChestCount) ||
    Number.isNaN(healingFountainCount)
  ) {
    return { error: 'Counts must be integers.' };
  }

  return {
    payload: {
      placeCount,
      monsterCount,
      inputEncounterCount,
      optionEncounterCount,
      treasureChestCount,
      healingFountainCount,
      requiredPlaceTags: params.requiredPlaceTags,
      shopkeeperItemTags: params.shopkeeperItemTags,
    },
  };
};

export const ZoneSeedJobs = () => {
  const { apiClient } = useAPI();
  const { zones, refreshZones } = useZoneContext();
  const [draftZoneId, setDraftZoneId] = useState<string>('');
  const [jobFilterZoneId, setJobFilterZoneId] = useState<string>('');
  const [jobs, setJobs] = useState<ZoneSeedJob[]>([]);
  const [loadingJobs, setLoadingJobs] = useState(false);
  const [creatingDraft, setCreatingDraft] = useState(false);
  const [creatingBulkDrafts, setCreatingBulkDrafts] = useState(false);
  const [approvingId, setApprovingId] = useState<string | null>(null);
  const [retryingId, setRetryingId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [placeCount, setPlaceCount] = useState('0');
  const [monsterCount, setMonsterCount] = useState('0');
  const [inputEncounterCount, setInputEncounterCount] = useState('0');
  const [optionEncounterCount, setOptionEncounterCount] = useState('0');
  const [treasureChestCount, setTreasureChestCount] = useState('0');
  const [healingFountainCount, setHealingFountainCount] = useState('0');
  const [requiredPlaceTags, setRequiredPlaceTags] = useState<string[]>([]);
  const [requiredTagQuery, setRequiredTagQuery] = useState('');
  const [showRequiredTagSuggestions, setShowRequiredTagSuggestions] = useState(false);
  const [shopkeeperItemTags, setShopkeeperItemTags] = useState<string[]>([]);
  const [shopkeeperTagQuery, setShopkeeperTagQuery] = useState('');
  const [bulkZoneQuery, setBulkZoneQuery] = useState('');
  const [bulkZoneCount, setBulkZoneCount] = useState('5');

  const knownPlaceTags = useMemo(
    () => [
      'cafe',
      'coffee_shop',
      'bakery',
      'restaurant',
      'bar',
      'ice_cream_shop',
      'dessert',
      'park',
      'garden',
      'playground',
      'trail',
      'hiking_area',
      'natural_feature',
      'beach',
      'plaza',
      'square',
      'bridge',
      'museum',
      'art_gallery',
      'gallery',
      'library',
      'book_store',
      'movie_theater',
      'theater',
      'music_venue',
      'stadium',
      'sports_complex',
      'amusement_park',
      'zoo',
      'aquarium',
      'market',
      'shopping_mall',
      'store',
      'clothing_store',
      'florist',
    ],
    []
  );
  const [draftZoneQuery, setDraftZoneQuery] = useState('');
  const [showDraftZoneSuggestions, setShowDraftZoneSuggestions] = useState(false);
  const [filterZoneQuery, setFilterZoneQuery] = useState('');
  const [showFilterZoneSuggestions, setShowFilterZoneSuggestions] = useState(false);
  const sortedZones = useMemo(
    () => [...zones].sort((left, right) => left.name.localeCompare(right.name)),
    [zones]
  );

  const selectedZone = useMemo<Zone | undefined>(() => {
    return sortedZones.find((zone) => zone.id === draftZoneId);
  }, [sortedZones, draftZoneId]);

  const draftZoneSuggestions = useMemo(() => {
    const query = draftZoneQuery.toLowerCase();
    return sortedZones.filter((zone) => zone.name.toLowerCase().includes(query));
  }, [sortedZones, draftZoneQuery]);

  const filterZoneSuggestions = useMemo(() => {
    const query = filterZoneQuery.toLowerCase();
    return sortedZones.filter((zone) => zone.name.toLowerCase().includes(query));
  }, [sortedZones, filterZoneQuery]);

  const bulkRequestedZoneCount = useMemo(() => {
    const parsed = Number.parseInt(bulkZoneCount, 10);
    if (Number.isNaN(parsed) || parsed < 0) {
      return 0;
    }
    return parsed;
  }, [bulkZoneCount]);

  const bulkMatchingZones = useMemo(() => {
    const query = bulkZoneQuery.trim().toLowerCase();
    if (!query) {
      return sortedZones;
    }
    return sortedZones.filter((zone) => zone.name.toLowerCase().includes(query));
  }, [sortedZones, bulkZoneQuery]);

  const bulkTargetZones = useMemo(() => {
    if (bulkRequestedZoneCount <= 0) {
      return [];
    }
    return bulkMatchingZones.slice(0, bulkRequestedZoneCount);
  }, [bulkMatchingZones, bulkRequestedZoneCount]);

  useEffect(() => {
    if (sortedZones.length === 0) {
      refreshZones();
      return;
    }
    if (!draftZoneId && sortedZones.length > 0) {
      setDraftZoneId(sortedZones[0].id);
    }
  }, [sortedZones, draftZoneId, refreshZones]);

  const fetchJobs = useCallback(async (zoneId?: string) => {
    setLoadingJobs(true);
    setError(null);
    try {
      const response = await apiClient.get<ZoneSeedJob[]>(
        '/sonar/admin/zone-seed-jobs',
        zoneId ? { zoneId, limit: 25 } : { limit: 25 }
      );
      setJobs(response);
    } catch (err) {
      console.error('Failed to load zone seed jobs', err);
      setError('Failed to load zone seed jobs.');
    } finally {
      setLoadingJobs(false);
    }
  }, [apiClient]);

  useEffect(() => {
    fetchJobs(jobFilterZoneId || undefined);
  }, [fetchJobs, jobFilterZoneId]);

  useEffect(() => {
    if (selectedZone?.name) {
      setDraftZoneQuery(selectedZone.name);
    }
  }, [selectedZone]);

  const handleCreateDraft = async () => {
    if (!draftZoneId) {
      setError('Please select a zone.');
      return;
    }
    const { payload, error: payloadError } = buildSeedDraftPayload({
      placeCount,
      monsterCount,
      inputEncounterCount,
      optionEncounterCount,
      treasureChestCount,
      healingFountainCount,
      requiredPlaceTags,
      shopkeeperItemTags,
    });
    if (!payload) {
      setError(payloadError ?? 'Counts must be integers.');
      return;
    }
    setCreatingDraft(true);
    setError(null);
    setSuccess(null);
    try {
      const created = await apiClient.post<ZoneSeedJob>(
        `/sonar/admin/zones/${draftZoneId}/seed-draft`,
        payload
      );
      setJobs((prev) => [created, ...prev]);
      setSuccess('Draft queued successfully.');
    } catch (err) {
      console.error('Failed to queue draft', err);
      setError('Failed to queue zone seed draft.');
    } finally {
      setCreatingDraft(false);
    }
  };

  const handleBulkCreateDrafts = async () => {
    if (bulkRequestedZoneCount <= 0) {
      setError('Bulk zone count must be greater than zero.');
      return;
    }
    if (bulkTargetZones.length === 0) {
      setError('No zones match the current bulk filter.');
      return;
    }

    const { payload, error: payloadError } = buildSeedDraftPayload({
      placeCount,
      monsterCount,
      inputEncounterCount,
      optionEncounterCount,
      treasureChestCount,
      healingFountainCount,
      requiredPlaceTags,
      shopkeeperItemTags,
    });
    if (!payload) {
      setError(payloadError ?? 'Counts must be integers.');
      return;
    }

    setCreatingBulkDrafts(true);
    setError(null);
    setSuccess(null);
    try {
      const response = await apiClient.post<BulkQueueZoneSeedJobsResponse>(
        '/sonar/admin/zone-seed-jobs/bulk-queue',
        {
          zoneIds: bulkTargetZones.map((zone) => zone.id),
          ...payload,
        }
      );
      setJobs((prev) => [...response.jobs, ...prev]);
      if (response.queuedCount === response.requestedZoneCount) {
        setSuccess(`Queued ${response.queuedCount} zone seed draft jobs.`);
      } else {
        setSuccess(
          `Queued ${response.queuedCount} zone seed draft jobs from ${response.requestedZoneCount} requested zones.`
        );
      }
    } catch (err) {
      console.error('Failed to bulk queue drafts', err);
      setError('Failed to bulk queue zone seed drafts.');
    } finally {
      setCreatingBulkDrafts(false);
    }
  };

  const handleApprove = async (job: ZoneSeedJob) => {
    if (approvingId) return;
    setApprovingId(job.id);
    setError(null);
    setSuccess(null);
    try {
      await apiClient.post(`/sonar/admin/zone-seed-jobs/${job.id}/approve`);
      setSuccess('Draft approved and applying.');
      await fetchJobs(job.zoneId);
    } catch (err) {
      console.error('Failed to approve draft', err);
      setError('Failed to approve draft.');
    } finally {
      setApprovingId(null);
    }
  };

  const handleDelete = async (job: ZoneSeedJob) => {
    if (deletingId) return;
    const confirmed = window.confirm(`Delete draft job ${job.id.slice(0, 8)}? This cannot be undone.`);
    if (!confirmed) return;
    setDeletingId(job.id);
    setError(null);
    setSuccess(null);
    try {
      await apiClient.delete(`/sonar/admin/zone-seed-jobs/${job.id}`);
      setJobs((prev) => prev.filter((existing) => existing.id !== job.id));
      setSuccess('Draft job deleted.');
    } catch (err) {
      console.error('Failed to delete draft job', err);
      setError('Failed to delete draft job.');
    } finally {
      setDeletingId(null);
    }
  };

  const handleRetry = async (job: ZoneSeedJob) => {
    if (retryingId) return;
    setRetryingId(job.id);
    setError(null);
    setSuccess(null);
    try {
      await apiClient.post(`/sonar/admin/zone-seed-jobs/${job.id}/retry`);
      setSuccess('Draft retry queued.');
      await fetchJobs(job.zoneId);
    } catch (err) {
      console.error('Failed to retry draft job', err);
      setError('Failed to retry draft job.');
    } finally {
      setRetryingId(null);
    }
  };

  const addRequiredTag = (value: string) => {
    const trimmed = value.trim().toLowerCase();
    if (!trimmed) return;
    if (requiredPlaceTags.includes(trimmed)) return;
    setRequiredPlaceTags((prev) => [...prev, trimmed]);
  };

  const removeRequiredTag = (value: string) => {
    setRequiredPlaceTags((prev) => prev.filter((tag) => tag !== value));
  };

  const addShopkeeperTag = (value: string) => {
    const trimmed = value.trim().toLowerCase();
    if (!trimmed) return;
    if (shopkeeperItemTags.includes(trimmed)) return;
    setShopkeeperItemTags((prev) => [...prev, trimmed]);
  };

  const removeShopkeeperTag = (value: string) => {
    setShopkeeperItemTags((prev) => prev.filter((tag) => tag !== value));
  };

  const filteredTagSuggestions = useMemo(() => {
    const query = requiredTagQuery.trim().toLowerCase();
    const available = knownPlaceTags.filter((tag) => !requiredPlaceTags.includes(tag));
    if (!query) return available;
    return available.filter((tag) => tag.includes(query));
  }, [knownPlaceTags, requiredPlaceTags, requiredTagQuery]);

  return (
    <div className="container mx-auto px-6 py-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Zone Seeding</h1>
          <p className="text-sm text-gray-500">
            Create fantasy zone drafts with POIs, standalone challenges, scalable encounters, and treasure.
          </p>
        </div>
        <button
          className="px-4 py-2 rounded bg-gray-800 text-white hover:bg-gray-700"
          onClick={() => fetchJobs(jobFilterZoneId || undefined)}
          disabled={loadingJobs}
        >
          {loadingJobs ? 'Refreshing...' : 'Refresh drafts'}
        </button>
      </div>

      {error && (
        <div className="mb-4 rounded border border-red-200 bg-red-50 px-4 py-3 text-red-700">
          {error}
        </div>
      )}
      {success && (
        <div className="mb-4 rounded border border-emerald-200 bg-emerald-50 px-4 py-3 text-emerald-700">
          {success}
        </div>
      )}

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Draft settings</h2>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Zone
          </label>
          <div className="relative">
            <input
              className="w-full rounded border border-gray-300 px-3 py-2 text-sm"
              value={draftZoneQuery}
              onChange={(e) => {
                const value = e.target.value;
                setDraftZoneQuery(value);
                setShowDraftZoneSuggestions(true);
                if (value.trim() === '') {
                  setDraftZoneId('');
                }
              }}
              onFocus={() => setShowDraftZoneSuggestions(true)}
              onBlur={() => {
                setTimeout(() => setShowDraftZoneSuggestions(false), 120);
              }}
              placeholder="Type to filter zones..."
            />
            {showDraftZoneSuggestions && draftZoneSuggestions.length > 0 && (
              <div className="absolute z-20 mt-1 max-h-60 w-full overflow-y-auto rounded border border-gray-200 bg-white shadow">
                {draftZoneSuggestions.map((zone) => (
                    <button
                      type="button"
                      key={zone.id}
                      onClick={() => {
                        setDraftZoneId(zone.id);
                        setDraftZoneQuery(zone.name);
                        setShowDraftZoneSuggestions(false);
                      }}
                      className="block w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                    >
                      {zone.name}
                    </button>
                ))}
              </div>
            )}
          </div>
          {selectedZone && (
            <p className="mt-2 text-xs text-gray-500">
              Selected: {selectedZone.name}
            </p>
          )}
          <div className="mt-4 grid grid-cols-2 gap-3 md:grid-cols-6">
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">
                Places
              </label>
              <input
                className="w-full rounded border border-gray-300 px-2 py-2 text-sm"
                value={placeCount}
                onChange={(e) => setPlaceCount(e.target.value)}
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">
                Monster encounters
              </label>
              <input
                className="w-full rounded border border-gray-300 px-2 py-2 text-sm"
                value={monsterCount}
                onChange={(e) => setMonsterCount(e.target.value)}
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">
                Input scenarios
              </label>
              <input
                className="w-full rounded border border-gray-300 px-2 py-2 text-sm"
                value={inputEncounterCount}
                onChange={(e) => setInputEncounterCount(e.target.value)}
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">
                Option scenarios
              </label>
              <input
                className="w-full rounded border border-gray-300 px-2 py-2 text-sm"
                value={optionEncounterCount}
                onChange={(e) => setOptionEncounterCount(e.target.value)}
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">
                Treasure chests
              </label>
              <input
                className="w-full rounded border border-gray-300 px-2 py-2 text-sm"
                value={treasureChestCount}
                onChange={(e) => setTreasureChestCount(e.target.value)}
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">
                Healing fountains
              </label>
              <input
                className="w-full rounded border border-gray-300 px-2 py-2 text-sm"
                value={healingFountainCount}
                onChange={(e) => setHealingFountainCount(e.target.value)}
              />
            </div>
          </div>
          <div className="mt-4">
            <label className="block text-xs font-medium text-gray-500 mb-1">
              Required POI tags
            </label>
            <div className="rounded border border-gray-300 px-2 py-2 text-sm">
              <div className="flex flex-wrap gap-2">
                {requiredPlaceTags.map((tag) => (
                  <span
                    key={tag}
                    className="inline-flex items-center rounded-full bg-indigo-50 px-2 py-1 text-xs text-indigo-700"
                  >
                    {tag}
                    <button
                      type="button"
                      className="ml-2 text-indigo-500 hover:text-indigo-700"
                      onClick={() => removeRequiredTag(tag)}
                    >
                      x
                    </button>
                  </span>
                ))}
                <div className="relative flex-1 min-w-[140px]">
                  <input
                    className="w-full border-0 px-2 py-1 text-sm focus:outline-none"
                    placeholder="Add tag..."
                    value={requiredTagQuery}
                    onChange={(e) => {
                      setRequiredTagQuery(e.target.value);
                      setShowRequiredTagSuggestions(true);
                    }}
                    onFocus={() => setShowRequiredTagSuggestions(true)}
                    onBlur={() => setTimeout(() => setShowRequiredTagSuggestions(false), 120)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ',') {
                        e.preventDefault();
                        addRequiredTag(requiredTagQuery);
                        setRequiredTagQuery('');
                      }
                      if (e.key === 'Backspace' && requiredTagQuery === '' && requiredPlaceTags.length > 0) {
                        removeRequiredTag(requiredPlaceTags[requiredPlaceTags.length - 1]);
                      }
                    }}
                  />
                  {showRequiredTagSuggestions && (filteredTagSuggestions.length > 0 || requiredTagQuery.trim()) && (
                    <div className="absolute z-20 mt-1 max-h-56 w-full overflow-y-auto rounded border border-gray-200 bg-white shadow">
                      {filteredTagSuggestions.map((tag) => (
                        <button
                          key={tag}
                          type="button"
                          className="block w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                          onClick={() => {
                            addRequiredTag(tag);
                            setRequiredTagQuery('');
                            setShowRequiredTagSuggestions(false);
                          }}
                        >
                          {tag}
                        </button>
                      ))}
                      {requiredTagQuery.trim() && !requiredPlaceTags.includes(requiredTagQuery.trim().toLowerCase()) && (
                        <button
                          type="button"
                          className="block w-full px-3 py-2 text-left text-sm text-indigo-600 hover:bg-indigo-50"
                          onClick={() => {
                            addRequiredTag(requiredTagQuery);
                            setRequiredTagQuery('');
                            setShowRequiredTagSuggestions(false);
                          }}
                        >
                          Add &quot;{requiredTagQuery.trim()}&quot;
                        </button>
                      )}
                    </div>
                  )}
                </div>
              </div>
            </div>
            <p className="mt-1 text-xs text-gray-400">
              We will ensure at least one POI matches each tag.
            </p>
          </div>
          <div className="mt-4">
            <label className="block text-xs font-medium text-gray-500 mb-1">
              Shopkeeper item tags
            </label>
            <div className="rounded border border-gray-300 px-2 py-2 text-sm">
              <div className="flex flex-wrap gap-2">
                {shopkeeperItemTags.map((tag) => (
                  <span
                    key={tag}
                    className="inline-flex items-center rounded-full bg-emerald-50 px-2 py-1 text-xs text-emerald-700"
                  >
                    {tag}
                    <button
                      type="button"
                      className="ml-2 text-emerald-600 hover:text-emerald-800"
                      onClick={() => removeShopkeeperTag(tag)}
                    >
                      x
                    </button>
                  </span>
                ))}
                <input
                  className="flex-1 min-w-[140px] border-0 px-2 py-1 text-sm focus:outline-none"
                  placeholder="Add item tag..."
                  value={shopkeeperTagQuery}
                  onChange={(e) => setShopkeeperTagQuery(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' || e.key === ',') {
                      e.preventDefault();
                      addShopkeeperTag(shopkeeperTagQuery);
                      setShopkeeperTagQuery('');
                    }
                    if (e.key === 'Backspace' && shopkeeperTagQuery === '' && shopkeeperItemTags.length > 0) {
                      removeShopkeeperTag(shopkeeperItemTags[shopkeeperItemTags.length - 1]);
                    }
                  }}
                />
              </div>
            </div>
            <p className="mt-1 text-xs text-gray-400">
              One shopkeeper will be generated per tag at a random location in the zone.
            </p>
          </div>
          <button
            className="mt-5 w-full rounded bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-500 disabled:opacity-60"
            onClick={handleCreateDraft}
            disabled={creatingDraft || !draftZoneId}
          >
            {creatingDraft ? 'Queuing...' : 'Create draft'}
          </button>
          <div className="mt-6 rounded-lg border border-dashed border-gray-300 bg-gray-50 p-4">
            <div className="flex items-center justify-between gap-3">
              <div>
                <h3 className="text-sm font-semibold text-gray-900">Bulk queue</h3>
                <p className="text-xs text-gray-500">
                  Queue this same seed configuration across the first N zones matching a filter.
                </p>
              </div>
              <div className="w-24">
                <label className="block text-xs font-medium text-gray-500 mb-1">
                  Zones
                </label>
                <input
                  className="w-full rounded border border-gray-300 px-2 py-2 text-sm"
                  value={bulkZoneCount}
                  onChange={(e) => setBulkZoneCount(e.target.value)}
                />
              </div>
            </div>
            <div className="mt-3">
              <label className="block text-xs font-medium text-gray-500 mb-1">
                Zone filter
              </label>
              <input
                className="w-full rounded border border-gray-300 px-3 py-2 text-sm"
                value={bulkZoneQuery}
                onChange={(e) => setBulkZoneQuery(e.target.value)}
                placeholder="Optional name filter..."
              />
            </div>
            <p className="mt-2 text-xs text-gray-500">
              Matching zones: {bulkMatchingZones.length}. Targeting {bulkTargetZones.length}
              {bulkRequestedZoneCount > 0 ? ` of requested ${bulkRequestedZoneCount}` : ''}.
            </p>
            {bulkTargetZones.length > 0 && (
              <p className="mt-2 text-xs text-gray-500">
                Preview: {bulkTargetZones.slice(0, 5).map((zone) => zone.name).join(', ')}
                {bulkTargetZones.length > 5 ? ` +${bulkTargetZones.length - 5} more` : ''}
              </p>
            )}
            <button
              className="mt-4 w-full rounded bg-slate-800 px-4 py-2 text-white hover:bg-slate-700 disabled:opacity-60"
              onClick={handleBulkCreateDrafts}
              disabled={creatingBulkDrafts || bulkTargetZones.length === 0}
            >
              {creatingBulkDrafts ? 'Queuing bulk drafts...' : `Queue for ${bulkTargetZones.length} zones`}
            </button>
          </div>
        </div>

        <div className="lg:col-span-2 rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
          <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900">Draft jobs</h2>
            <div className="relative w-full md:w-72">
              <input
                className="w-full rounded border border-gray-300 px-3 py-2 text-sm"
                value={filterZoneQuery}
                onChange={(e) => {
                  const value = e.target.value;
                  setFilterZoneQuery(value);
                  setShowFilterZoneSuggestions(true);
                  if (value.trim() === '') {
                    setJobFilterZoneId('');
                  }
                }}
                onFocus={() => setShowFilterZoneSuggestions(true)}
                onBlur={() => setTimeout(() => setShowFilterZoneSuggestions(false), 120)}
                placeholder="Filter by zone (optional)..."
              />
              {showFilterZoneSuggestions && filterZoneSuggestions.length > 0 && (
                <div className="absolute z-20 mt-1 max-h-60 w-full overflow-y-auto rounded border border-gray-200 bg-white shadow">
                  <button
                    type="button"
                    onClick={() => {
                      setJobFilterZoneId('');
                      setFilterZoneQuery('');
                      setShowFilterZoneSuggestions(false);
                    }}
                    className="block w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                  >
                    All zones
                  </button>
                  {filterZoneSuggestions.map((zone) => (
                      <button
                        type="button"
                        key={zone.id}
                        onClick={() => {
                          setJobFilterZoneId(zone.id);
                          setFilterZoneQuery(zone.name);
                          setShowFilterZoneSuggestions(false);
                        }}
                        className="block w-full px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                      >
                        {zone.name}
                      </button>
                  ))}
                </div>
              )}
            </div>
          </div>
          {loadingJobs ? (
            <p className="text-sm text-gray-500">Loading drafts...</p>
          ) : jobs.length === 0 ? (
            <p className="text-sm text-gray-500">No draft jobs for this zone yet.</p>
          ) : (
            <div className="space-y-4">
              {jobs.map((job) => {
                return (
                  <div
                    key={job.id}
                    className="rounded-lg border border-gray-200 p-4"
                  >
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <h3 className="text-sm font-semibold text-gray-900">
                        Job {job.id.slice(0, 8)}
                      </h3>
                      <p className="text-xs text-gray-500">
                        Created: {formatDate(job.createdAt)} | Updated: {formatDate(job.updatedAt)}
                      </p>
                      <p className="text-xs text-gray-500">
                        Counts: {job.placeCount} POIs/challenges, {job.monsterCount ?? 0} monster encounters,{' '}
                        {job.inputEncounterCount ?? 0} input scenarios, {job.optionEncounterCount ?? 0} option scenarios,{' '}
                        {job.treasureChestCount ?? 0} treasure chests, {job.healingFountainCount ?? 0} healing fountains,{' '}
                        {job.shopkeeperItemTags?.length ?? 0} shopkeepers
                      </p>
                      {job.requiredPlaceTags && job.requiredPlaceTags.length > 0 && (
                        <p className="text-xs text-gray-500">
                          Required tags: {job.requiredPlaceTags.join(', ')}
                        </p>
                      )}
                      {job.shopkeeperItemTags && job.shopkeeperItemTags.length > 0 && (
                        <p className="text-xs text-gray-500">
                          Shopkeeper tags: {job.shopkeeperItemTags.join(', ')}
                        </p>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      <span
                        className={`inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold text-white ${statusBadgeClass(
                          job.status
                        )}`}
                      >
                        {job.status.replace(/_/g, ' ')}
                      </span>
                      {job.status === 'failed' && (
                        <button
                          className="rounded border border-gray-200 px-2 py-1 text-xs text-indigo-700 hover:bg-indigo-50 disabled:opacity-50"
                          onClick={() => handleRetry(job)}
                          disabled={retryingId === job.id}
                          title="Retry draft job"
                        >
                          {retryingId === job.id ? 'Retrying...' : 'Retry'}
                        </button>
                      )}
                      <button
                        className="rounded border border-gray-200 px-2 py-1 text-xs text-gray-600 hover:bg-gray-50 disabled:opacity-50"
                        onClick={() => handleDelete(job)}
                        disabled={deletingId === job.id || job.status === 'in_progress' || job.status === 'applying'}
                        title={
                          job.status === 'in_progress' || job.status === 'applying'
                            ? 'Cannot delete while running'
                            : 'Delete draft job'
                        }
                      >
                        {deletingId === job.id ? 'Deleting...' : 'Delete'}
                      </button>
                    </div>
                  </div>

                  {job.errorMessage && (
                    <div className="mt-3 rounded border border-red-100 bg-red-50 px-3 py-2 text-xs text-red-700">
                      {job.errorMessage}
                    </div>
                  )}

                  {job.draft && (
                    <details className="mt-3">
                      <summary className="cursor-pointer text-sm font-medium text-gray-700">
                        Draft details
                      </summary>
                      <div className="mt-3 space-y-6 text-sm text-gray-700">
                        <div>
                          <div className="font-semibold">Fantasy branding</div>
                          <div className="text-sm text-gray-600">
                            {job.draft.fantasyName || 'Untitled district'}
                          </div>
                          {job.draft.zoneDescription && (
                            <p className="mt-2 text-sm text-gray-600 whitespace-pre-wrap">
                              {job.draft.zoneDescription}
                            </p>
                          )}
                        </div>
                        <div>
                          <div className="font-semibold">Points of interest</div>
                          <div className="mt-2 space-y-3 text-xs text-gray-600">
                            {(job.draft.pointsOfInterest || []).map((poi) => (
                              <div
                                key={poi.draftId}
                                className="rounded border border-gray-100 bg-gray-50 p-3"
                              >
                                <div className="text-sm font-semibold text-gray-800">
                                  {poi.name || 'Unnamed place'}
                                </div>
                                <div>Place ID: {poi.placeId || 'n/a'}</div>
                                {poi.address && <div>Address: {poi.address}</div>}
                                {typeof poi.latitude === 'number' &&
                                  typeof poi.longitude === 'number' && (
                                    <div>
                                      Coordinates: {poi.latitude}, {poi.longitude}
                                    </div>
                                  )}
                                {typeof poi.rating === 'number' && (
                                  <div>
                                    Rating: {poi.rating}
                                    {typeof poi.userRatingCount === 'number'
                                      ? ` (${poi.userRatingCount} reviews)`
                                      : ''}
                                  </div>
                                )}
                                {poi.types && poi.types.length > 0 && (
                                  <div>Types: {poi.types.join(', ')}</div>
                                )}
                                {poi.editorialSummary && (
                                  <div className="mt-1 text-gray-500">
                                    Summary: {poi.editorialSummary}
                                  </div>
                                )}
                              </div>
                            ))}
                          </div>
                        </div>
                        <div>
                          <div className="font-semibold">Characters</div>
                          <div className="mt-2 space-y-3 text-xs text-gray-600">
                            {(job.draft.characters || []).map((character) => (
                              <div
                                key={character.draftId}
                                className="rounded border border-gray-100 bg-gray-50 p-3"
                              >
                                <div className="text-sm font-semibold text-gray-800">
                                  {character.name || 'Unnamed character'}
                                </div>
                                <div>Place ID: {character.placeId || 'n/a'}</div>
                                {typeof character.latitude === 'number' &&
                                  typeof character.longitude === 'number' && (
                                    <div>
                                      Coordinates: {character.latitude}, {character.longitude}
                                    </div>
                                  )}
                                {character.shopItemTags && character.shopItemTags.length > 0 && (
                                  <div>Shopkeeper tags: {character.shopItemTags.join(', ')}</div>
                                )}
                                {character.description && (
                                  <div className="mt-1 text-gray-500 whitespace-pre-wrap">
                                    {character.description}
                                  </div>
                                )}
                              </div>
                            ))}
                          </div>
                        </div>
                        <div>
                          <div className="font-semibold">Seeding plan preview</div>
                          <div className="mt-2 rounded border border-gray-100 bg-gray-50 p-3 text-xs text-gray-600">
                            <div>{job.placeCount} POIs selected for challenge placement</div>
                            <div>{job.placeCount} standalone challenges at those POIs</div>
                            <div>{job.monsterCount ?? 0} random monster encounters (scalable)</div>
                            <div>{job.inputEncounterCount ?? 0} random input scenarios (scalable)</div>
                            <div>{job.optionEncounterCount ?? 0} random option scenarios (scalable)</div>
                            <div>{job.treasureChestCount ?? 0} random treasure chests (scalable rewards)</div>
                            <div>{job.healingFountainCount ?? 0} random healing fountains</div>
                            <div>{job.shopkeeperItemTags?.length ?? 0} shopkeepers generated at random zone locations</div>
                          </div>
                        </div>
                      </div>
                    </details>
                  )}

                  {job.status === 'awaiting_approval' && (
                    <div className="mt-4">
                      <button
                        className="rounded bg-emerald-600 px-4 py-2 text-sm text-white hover:bg-emerald-500 disabled:opacity-60"
                        onClick={() => handleApprove(job)}
                        disabled={approvingId === job.id}
                      >
                        {approvingId === job.id ? 'Approving...' : 'Approve and apply'}
                      </button>
                    </div>
                  )}
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ZoneSeedJobs;
