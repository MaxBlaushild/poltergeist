import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { District, QuestArchetype } from '@poltergeist/types';
import { Link } from 'react-router-dom';
import { useQuestArchetypes } from '../contexts/questArchetypes.tsx';

type DistrictSeedResult = {
  questArchetypeId: string;
  questArchetypeName?: string;
  status: string;
  errorMessage?: string | null;
  zoneId?: string | null;
  zoneName?: string;
  matchCount: number;
  questId?: string | null;
  questGiverCharacterId?: string | null;
  questGiverCharacterName?: string;
  generatedCharacterId?: string | null;
  generatedCharacterName?: string;
};

type DistrictSeedJob = {
  id: string;
  districtId: string;
  status: string;
  errorMessage?: string | null;
  questArchetypeIds: string[];
  zoneSeedSettings?: {
    placeCount?: number;
    monsterCount?: number;
    bossEncounterCount?: number;
    raidEncounterCount?: number;
    inputEncounterCount?: number;
    optionEncounterCount?: number;
    treasureChestCount?: number;
    healingFountainCount?: number;
    requiredPlaceTags?: string[];
    shopkeeperItemTags?: string[];
  };
  zoneSeedJobIds?: string[];
  results: DistrictSeedResult[];
  createdAt: string;
  updatedAt: string;
};

const activeStatuses = new Set(['queued', 'in_progress']);

const statusClasses: Record<string, string> = {
  queued: 'bg-slate-100 text-slate-700',
  in_progress: 'bg-blue-100 text-blue-700',
  completed: 'bg-emerald-100 text-emerald-700',
  failed: 'bg-red-100 text-red-700',
};

const formatTimestamp = (value?: string) => {
  if (!value) {
    return 'Unknown time';
  }
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }
  return parsed.toLocaleString();
};

const templateSearchText = (questArchetype: QuestArchetype) =>
  [
    questArchetype.name,
    questArchetype.description,
    ...(questArchetype.internalTags || []),
    ...(questArchetype.characterTags || []),
  ]
    .join(' ')
    .toLowerCase();

const renderQuestArchetypeTags = (questArchetype: QuestArchetype) => {
  const tags = [
    ...(questArchetype.internalTags || []).map((tag) => ({
      key: `internal-${tag}`,
      label: tag,
      className: 'bg-slate-100 text-slate-700',
    })),
    ...(questArchetype.characterTags || []).map((tag) => ({
      key: `character-${tag}`,
      label: `giver:${tag}`,
      className: 'bg-emerald-100 text-emerald-700',
    })),
  ];

  if (tags.length === 0) {
    return (
      <span className="rounded-full bg-gray-100 px-2 py-0.5 text-[11px] text-gray-500">
        No internal tags
      </span>
    );
  }

  return (
    <>
      {tags.map((tag) => (
        <span
          key={tag.key}
          className={`rounded-full px-2 py-0.5 text-[11px] font-medium ${tag.className}`}
        >
          {tag.label}
        </span>
      ))}
    </>
  );
};

export const DistrictSeedJobsPanel = ({
  district,
}: {
  district: District | null;
}) => {
  const { apiClient } = useAPI();
  const { questArchetypes } = useQuestArchetypes();
  const [jobs, setJobs] = useState<DistrictSeedJob[]>([]);
  const [loadingJobs, setLoadingJobs] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [templateSearch, setTemplateSearch] = useState('');
  const [placeCount, setPlaceCount] = useState('0');
  const [monsterCount, setMonsterCount] = useState('0');
  const [bossEncounterCount, setBossEncounterCount] = useState('0');
  const [raidEncounterCount, setRaidEncounterCount] = useState('0');
  const [inputEncounterCount, setInputEncounterCount] = useState('0');
  const [optionEncounterCount, setOptionEncounterCount] = useState('0');
  const [treasureChestCount, setTreasureChestCount] = useState('0');
  const [healingFountainCount, setHealingFountainCount] = useState('0');
  const [requiredPlaceTags, setRequiredPlaceTags] = useState<string[]>([]);
  const [requiredTagQuery, setRequiredTagQuery] = useState('');
  const [shopkeeperItemTags, setShopkeeperItemTags] = useState<string[]>([]);
  const [shopkeeperTagQuery, setShopkeeperTagQuery] = useState('');
  const [selectedQuestArchetypeIds, setSelectedQuestArchetypeIds] = useState<
    Set<string>
  >(new Set());

  const districtId = district?.id || '';
  const hasZones = (district?.zones?.length || 0) > 0;
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

  const loadJobs = useCallback(
    async ({ silent = false }: { silent?: boolean } = {}) => {
      if (!districtId) {
        setJobs([]);
        return;
      }
      if (!silent) {
        setLoadingJobs(true);
      }
      try {
        const response = await apiClient.get<DistrictSeedJob[]>(
          `/sonar/admin/district-seed-jobs?districtId=${encodeURIComponent(districtId)}&limit=25`
        );
        setJobs(response);
      } catch (err) {
        console.error('Failed to load district seed jobs', err);
        if (!silent) {
          setError('Unable to load district seed jobs right now.');
        }
      } finally {
        if (!silent) {
          setLoadingJobs(false);
        }
      }
    },
    [apiClient, districtId]
  );

  useEffect(() => {
    void loadJobs();
  }, [districtId, loadJobs]);

  useEffect(() => {
    if (!districtId) {
      return;
    }
    const interval = window.setInterval(() => {
      if (jobs.some((job) => activeStatuses.has(job.status))) {
        void loadJobs({ silent: true });
      }
    }, 10000);
    return () => window.clearInterval(interval);
  }, [districtId, jobs, loadJobs]);

  const filteredQuestArchetypes = useMemo(() => {
    const query = templateSearch.trim().toLowerCase();
    if (!query) {
      return questArchetypes;
    }
    return questArchetypes.filter((questArchetype) =>
      templateSearchText(questArchetype).includes(query)
    );
  }, [questArchetypes, templateSearch]);

  const selectedCount = selectedQuestArchetypeIds.size;
  const hasZoneSeedSettings =
    requiredPlaceTags.length > 0 ||
    shopkeeperItemTags.length > 0 ||
    [
      placeCount,
      monsterCount,
      bossEncounterCount,
      raidEncounterCount,
      inputEncounterCount,
      optionEncounterCount,
      treasureChestCount,
      healingFountainCount,
    ].some((value) => Number.parseInt(value, 10) > 0);

  const addTag = (
    rawValue: string,
    current: string[],
    setter: React.Dispatch<React.SetStateAction<string[]>>
  ) => {
    const normalized = rawValue.trim().toLowerCase();
    if (!normalized || current.includes(normalized)) {
      return;
    }
    setter([...current, normalized]);
  };

  const removeTag = (
    tag: string,
    setter: React.Dispatch<React.SetStateAction<string[]>>
  ) => {
    setter((current) => current.filter((entry) => entry !== tag));
  };

  const toggleQuestArchetype = (questArchetypeId: string) => {
    setSelectedQuestArchetypeIds((current) => {
      const next = new Set(current);
      if (next.has(questArchetypeId)) {
        next.delete(questArchetypeId);
      } else {
        next.add(questArchetypeId);
      }
      return next;
    });
  };

  const handleQueueJob = async () => {
    if (
      !districtId ||
      (!hasZoneSeedSettings && selectedCount === 0) ||
      !hasZones
    ) {
      return;
    }

    const numericValues = {
      placeCount: Number.parseInt(placeCount, 10),
      monsterCount: Number.parseInt(monsterCount, 10),
      bossEncounterCount: Number.parseInt(bossEncounterCount, 10),
      raidEncounterCount: Number.parseInt(raidEncounterCount, 10),
      inputEncounterCount: Number.parseInt(inputEncounterCount, 10),
      optionEncounterCount: Number.parseInt(optionEncounterCount, 10),
      treasureChestCount: Number.parseInt(treasureChestCount, 10),
      healingFountainCount: Number.parseInt(healingFountainCount, 10),
    };
    if (Object.values(numericValues).some((value) => Number.isNaN(value))) {
      setError('Counts must be integers.');
      return;
    }

    setSubmitting(true);
    setError(null);
    setSuccess(null);
    try {
      const created = await apiClient.post<DistrictSeedJob>(
        '/sonar/admin/district-seed-jobs',
        {
          districtId,
          questArchetypeIds: Array.from(selectedQuestArchetypeIds),
          ...numericValues,
          requiredPlaceTags,
          shopkeeperItemTags,
        }
      );
      setJobs((current) => [created, ...current]);
      setSelectedQuestArchetypeIds(new Set());
      setRequiredTagQuery('');
      setShopkeeperTagQuery('');
      setSuccess(
        `Queued district seed job for ${created.questArchetypeIds.length} quest templates and ${created.zoneSeedJobIds?.length || 0} child-zone seed jobs.`
      );
    } catch (err) {
      console.error('Failed to queue district seed job', err);
      setError('Unable to queue that district seed job right now.');
    } finally {
      setSubmitting(false);
    }
  };

  const handleRetryJob = async (job: DistrictSeedJob) => {
    setError(null);
    setSuccess(null);
    try {
      const updated = await apiClient.post<DistrictSeedJob>(
        `/sonar/admin/district-seed-jobs/${job.id}/retry`,
        {}
      );
      setJobs((current) =>
        current.map((item) => (item.id === updated.id ? updated : item))
      );
      setSuccess('Re-queued district seed job.');
    } catch (err) {
      console.error('Failed to retry district seed job', err);
      setError('Unable to retry that district seed job right now.');
    }
  };

  const handleDeleteJob = async (job: DistrictSeedJob) => {
    const confirmed = window.confirm(
      `Delete district seed job ${job.id.slice(0, 8)}?`
    );
    if (!confirmed) {
      return;
    }

    setError(null);
    setSuccess(null);
    try {
      await apiClient.delete(`/sonar/admin/district-seed-jobs/${job.id}`);
      setJobs((current) => current.filter((item) => item.id !== job.id));
      setSuccess('Deleted district seed job.');
    } catch (err) {
      console.error('Failed to delete district seed job', err);
      setError('Unable to delete that district seed job right now.');
    }
  };

  return (
    <div className="grid gap-6 xl:grid-cols-[minmax(0,420px)_minmax(0,1fr)]">
      <div className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">
              District seeding
            </h2>
            <p className="text-sm text-gray-500">
              Queue full child-zone seed drafts and optionally layer quest
              templates across the district. Each quest template lands in the
              child zone with the strongest internal-tag overlap.
            </p>
          </div>
          <div className="rounded-lg bg-slate-50 px-3 py-2 text-right">
            <div className="text-xs uppercase tracking-wide text-slate-500">
              Selected
            </div>
            <div className="text-lg font-semibold text-slate-900">
              {selectedCount}
            </div>
          </div>
        </div>

        <div className="mt-4 rounded-lg bg-gray-50 px-3 py-3 text-sm text-gray-600">
          {hasZones
            ? `${district?.zones?.length || 0} district zones are available for matching.`
            : 'Add at least one zone to this district before queueing a seed job.'}
        </div>

        <div className="mt-4 grid gap-3 sm:grid-cols-2">
          {[
            ['Places', placeCount, setPlaceCount],
            ['Monsters', monsterCount, setMonsterCount],
            ['Bosses', bossEncounterCount, setBossEncounterCount],
            ['Raids', raidEncounterCount, setRaidEncounterCount],
            ['Input Encounters', inputEncounterCount, setInputEncounterCount],
            [
              'Option Encounters',
              optionEncounterCount,
              setOptionEncounterCount,
            ],
            ['Treasure Chests', treasureChestCount, setTreasureChestCount],
            [
              'Healing Fountains',
              healingFountainCount,
              setHealingFountainCount,
            ],
          ].map(([label, value, setter]) => (
            <label
              key={label}
              className="flex flex-col gap-1 text-xs font-medium uppercase tracking-wide text-gray-500"
            >
              {label}
              <input
                className="rounded-lg border border-gray-300 px-3 py-2 text-sm font-normal normal-case tracking-normal text-gray-900"
                value={value as string}
                onChange={(event) =>
                  (setter as (value: string) => void)(event.target.value)
                }
              />
            </label>
          ))}
        </div>

        <div className="mt-4 rounded-lg border border-gray-200 bg-gray-50 p-3">
          <div className="text-xs font-semibold uppercase tracking-wide text-gray-500">
            Required Place Tags
          </div>
          <div className="mt-2 flex flex-wrap gap-2">
            {requiredPlaceTags.map((tag) => (
              <button
                key={tag}
                type="button"
                onClick={() => removeTag(tag, setRequiredPlaceTags)}
                className="rounded-full bg-blue-100 px-2.5 py-1 text-xs font-medium text-blue-700"
              >
                {tag} x
              </button>
            ))}
          </div>
          <div className="mt-2 flex gap-2">
            <input
              className="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm"
              placeholder="Add required place tag"
              value={requiredTagQuery}
              onChange={(event) => setRequiredTagQuery(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === 'Enter' || event.key === ',') {
                  event.preventDefault();
                  addTag(
                    requiredTagQuery,
                    requiredPlaceTags,
                    setRequiredPlaceTags
                  );
                  setRequiredTagQuery('');
                }
              }}
            />
            <button
              type="button"
              onClick={() => {
                addTag(
                  requiredTagQuery,
                  requiredPlaceTags,
                  setRequiredPlaceTags
                );
                setRequiredTagQuery('');
              }}
              className="rounded-lg border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-white"
            >
              Add
            </button>
          </div>
          <div className="mt-2 flex flex-wrap gap-2">
            {knownPlaceTags
              .filter(
                (tag) =>
                  !requiredPlaceTags.includes(tag) &&
                  (!requiredTagQuery.trim() ||
                    tag.includes(requiredTagQuery.trim().toLowerCase()))
              )
              .slice(0, 12)
              .map((tag) => (
                <button
                  key={tag}
                  type="button"
                  onClick={() =>
                    addTag(tag, requiredPlaceTags, setRequiredPlaceTags)
                  }
                  className="rounded-full bg-white px-2 py-1 text-xs text-gray-600 hover:bg-gray-100"
                >
                  {tag}
                </button>
              ))}
          </div>
        </div>

        <div className="mt-4 rounded-lg border border-gray-200 bg-gray-50 p-3">
          <div className="text-xs font-semibold uppercase tracking-wide text-gray-500">
            Shopkeeper Item Tags
          </div>
          <div className="mt-2 flex flex-wrap gap-2">
            {shopkeeperItemTags.map((tag) => (
              <button
                key={tag}
                type="button"
                onClick={() => removeTag(tag, setShopkeeperItemTags)}
                className="rounded-full bg-emerald-100 px-2.5 py-1 text-xs font-medium text-emerald-700"
              >
                {tag} x
              </button>
            ))}
          </div>
          <div className="mt-2 flex gap-2">
            <input
              className="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm"
              placeholder="Add shopkeeper item tag"
              value={shopkeeperTagQuery}
              onChange={(event) => setShopkeeperTagQuery(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === 'Enter' || event.key === ',') {
                  event.preventDefault();
                  addTag(
                    shopkeeperTagQuery,
                    shopkeeperItemTags,
                    setShopkeeperItemTags
                  );
                  setShopkeeperTagQuery('');
                }
              }}
            />
            <button
              type="button"
              onClick={() => {
                addTag(
                  shopkeeperTagQuery,
                  shopkeeperItemTags,
                  setShopkeeperItemTags
                );
                setShopkeeperTagQuery('');
              }}
              className="rounded-lg border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-white"
            >
              Add
            </button>
          </div>
        </div>

        <input
          className="mt-4 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm"
          placeholder="Search quest templates by name or tag"
          value={templateSearch}
          onChange={(event) => setTemplateSearch(event.target.value)}
        />

        <div className="mt-4 max-h-[520px] space-y-2 overflow-y-auto pr-1">
          {filteredQuestArchetypes.map((questArchetype) => {
            const checked = selectedQuestArchetypeIds.has(questArchetype.id);
            return (
              <label
                key={questArchetype.id}
                className={`flex cursor-pointer items-start gap-3 rounded-lg border px-3 py-3 text-sm transition ${
                  checked
                    ? 'border-blue-300 bg-blue-50'
                    : 'border-gray-200 bg-gray-50 hover:border-gray-300 hover:bg-white'
                }`}
              >
                <input
                  type="checkbox"
                  checked={checked}
                  onChange={() => toggleQuestArchetype(questArchetype.id)}
                  className="mt-1 h-4 w-4"
                />
                <div className="min-w-0 flex-1">
                  <div className="font-medium text-gray-900">
                    {questArchetype.name || 'Untitled template'}
                  </div>
                  <div className="mt-1 text-xs text-gray-500">
                    {questArchetype.description || 'No description'}
                  </div>
                  <div className="mt-2 flex flex-wrap gap-1.5">
                    {renderQuestArchetypeTags(questArchetype)}
                  </div>
                </div>
              </label>
            );
          })}
          {filteredQuestArchetypes.length === 0 && (
            <div className="rounded-lg border border-dashed border-gray-300 px-4 py-8 text-center text-sm text-gray-500">
              No quest templates match that search.
            </div>
          )}
        </div>

        <div className="mt-4 flex items-center justify-between gap-3">
          <div className="text-xs text-gray-500">
            Zone seed settings queue regular zone-seed drafts for every child
            zone. Missing quest-giver characters are generated automatically,
            and quest templates still match to the strongest child-zone tag fit.
          </div>
          <button
            type="button"
            onClick={() => void handleQueueJob()}
            disabled={
              !hasZones ||
              (!hasZoneSeedSettings && selectedCount === 0) ||
              submitting
            }
            className="rounded-lg bg-slate-900 px-4 py-2 text-sm font-semibold text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
          >
            {submitting ? 'Queueing...' : 'Queue seed job'}
          </button>
        </div>
      </div>

      <div className="rounded-xl border border-gray-200 bg-white p-5 shadow-sm">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">Recent jobs</h2>
            <p className="text-sm text-gray-500">
              Track which template landed in which zone, and whether a quest
              giver had to be created.
            </p>
          </div>
          <button
            type="button"
            onClick={() => void loadJobs()}
            className="rounded-lg border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          >
            Refresh
          </button>
        </div>

        {(error || success) && (
          <div className="mt-4 space-y-2">
            {error && (
              <div className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                {error}
              </div>
            )}
            {success && (
              <div className="rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-700">
                {success}
              </div>
            )}
          </div>
        )}

        <div className="mt-4 space-y-4">
          {loadingJobs && jobs.length === 0 && (
            <div className="text-sm text-gray-500">
              Loading district seed jobs...
            </div>
          )}

          {!loadingJobs && jobs.length === 0 && (
            <div className="rounded-lg border border-dashed border-gray-300 px-4 py-10 text-center text-sm text-gray-500">
              No district seed jobs yet.
            </div>
          )}

          {jobs.map((job) => (
            <div
              key={job.id}
              className="rounded-xl border border-gray-200 bg-gray-50 p-4"
            >
              <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <span
                      className={`rounded-full px-2.5 py-1 text-xs font-semibold uppercase tracking-wide ${
                        statusClasses[job.status] || 'bg-gray-100 text-gray-700'
                      }`}
                    >
                      {job.status.replace('_', ' ')}
                    </span>
                    <span className="text-xs text-gray-500">
                      {formatTimestamp(job.createdAt)}
                    </span>
                  </div>
                  <div className="mt-2 text-sm text-gray-700">
                    {job.questArchetypeIds.length} quest templates and{' '}
                    {job.zoneSeedJobIds?.length || 0} child-zone seed jobs
                  </div>
                  <div className="mt-1 text-xs text-gray-500">
                    Job ID: {job.id}
                  </div>
                </div>
                <div className="flex gap-2">
                  {job.status === 'failed' && (
                    <button
                      type="button"
                      onClick={() => void handleRetryJob(job)}
                      className="rounded-lg border border-blue-200 px-3 py-2 text-sm font-medium text-blue-700 hover:bg-blue-50"
                    >
                      Retry
                    </button>
                  )}
                  {job.status !== 'in_progress' && (
                    <button
                      type="button"
                      onClick={() => void handleDeleteJob(job)}
                      className="rounded-lg border border-red-200 px-3 py-2 text-sm font-medium text-red-700 hover:bg-red-50"
                    >
                      Delete
                    </button>
                  )}
                </div>
              </div>

              {job.errorMessage && (
                <div className="mt-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                  {job.errorMessage}
                </div>
              )}

              <div className="mt-4 space-y-3">
                {job.results.map((result, index) => (
                  <div
                    key={`${job.id}-${result.questArchetypeId}-${index}`}
                    className="rounded-lg border border-white bg-white px-3 py-3"
                  >
                    <div className="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
                      <div>
                        <div className="font-medium text-gray-900">
                          {result.questArchetypeName || result.questArchetypeId}
                        </div>
                        <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-gray-500">
                          <span
                            className={`rounded-full px-2 py-0.5 font-semibold uppercase tracking-wide ${
                              statusClasses[result.status] ||
                              'bg-gray-100 text-gray-700'
                            }`}
                          >
                            {result.status}
                          </span>
                          {result.zoneId && (
                            <Link
                              to={`/zones/${result.zoneId}`}
                              className="font-medium text-blue-700 hover:text-blue-900"
                            >
                              {result.zoneName || result.zoneId}
                            </Link>
                          )}
                          <span>Tag overlap: {result.matchCount}</span>
                        </div>
                      </div>
                      {result.questId && (
                        <div className="text-xs text-gray-500">
                          Quest ID:{' '}
                          <span className="font-mono">{result.questId}</span>
                        </div>
                      )}
                    </div>

                    {result.errorMessage && (
                      <div className="mt-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                        {result.errorMessage}
                      </div>
                    )}

                    <div className="mt-3 grid gap-2 text-sm text-gray-600 md:grid-cols-2">
                      <div>
                        <span className="font-medium text-gray-900">
                          Quest giver:
                        </span>{' '}
                        {result.questGiverCharacterName || 'None'}
                      </div>
                      <div>
                        <span className="font-medium text-gray-900">
                          Generated:
                        </span>{' '}
                        {result.generatedCharacterName || 'No'}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
