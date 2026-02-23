import React, { useEffect, useMemo, useState } from 'react';
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
};

type ZoneSeedQuestDraft = {
  draftId: string;
  name: string;
  description: string;
  acceptanceDialogue?: string[];
  placeId: string;
  questGiverDraftId: string;
  challengeQuestion?: string;
  gold?: number;
};

type ZoneSeedDraft = {
  fantasyName?: string;
  zoneDescription?: string;
  pointsOfInterest?: ZoneSeedPointOfInterestDraft[];
  characters?: ZoneSeedCharacterDraft[];
  quests?: ZoneSeedQuestDraft[];
};

type ZoneSeedJob = {
  id: string;
  zoneId: string;
  status: string;
  errorMessage?: string;
  placeCount: number;
  characterCount: number;
  questCount: number;
  createdAt?: string;
  updatedAt?: string;
  draft?: ZoneSeedDraft;
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

export const ZoneSeedJobs = () => {
  const { apiClient } = useAPI();
  const { zones, refreshZones } = useZoneContext();
  const [draftZoneId, setDraftZoneId] = useState<string>('');
  const [jobFilterZoneId, setJobFilterZoneId] = useState<string>('');
  const [jobs, setJobs] = useState<ZoneSeedJob[]>([]);
  const [loadingJobs, setLoadingJobs] = useState(false);
  const [creatingDraft, setCreatingDraft] = useState(false);
  const [approvingId, setApprovingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [placeCount, setPlaceCount] = useState('8');
  const [characterCount, setCharacterCount] = useState('4');
  const [questCount, setQuestCount] = useState('4');
  const [draftZoneQuery, setDraftZoneQuery] = useState('');
  const [showDraftZoneSuggestions, setShowDraftZoneSuggestions] = useState(false);
  const [filterZoneQuery, setFilterZoneQuery] = useState('');
  const [showFilterZoneSuggestions, setShowFilterZoneSuggestions] = useState(false);

  const selectedZone = useMemo<Zone | undefined>(() => {
    return zones.find((zone) => zone.id === draftZoneId);
  }, [zones, draftZoneId]);

  useEffect(() => {
    if (zones.length === 0) {
      refreshZones();
      return;
    }
    if (!draftZoneId && zones.length > 0) {
      setDraftZoneId(zones[0].id);
    }
  }, [zones, draftZoneId, refreshZones]);

  useEffect(() => {
    fetchJobs(jobFilterZoneId || undefined);
  }, [jobFilterZoneId]);

  useEffect(() => {
    if (selectedZone?.name) {
      setDraftZoneQuery(selectedZone.name);
    }
  }, [selectedZone]);

  const fetchJobs = async (zoneId?: string) => {
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
  };

  const handleCreateDraft = async () => {
    if (!draftZoneId) {
      setError('Please select a zone.');
      return;
    }
    const places = Number.parseInt(placeCount, 10);
    const characters = Number.parseInt(characterCount, 10);
    const quests = Number.parseInt(questCount, 10);
    if (Number.isNaN(places) || Number.isNaN(characters) || Number.isNaN(quests)) {
      setError('Counts must be integers.');
      return;
    }
    setCreatingDraft(true);
    setError(null);
    setSuccess(null);
    try {
      const created = await apiClient.post<ZoneSeedJob>(
        `/sonar/admin/zones/${draftZoneId}/seed-draft`,
        {
          placeCount: places,
          characterCount: characters,
          questCount: quests,
        }
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

  return (
    <div className="container mx-auto px-6 py-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Zone Seeding</h1>
          <p className="text-sm text-gray-500">
            Create fantasy rebrands, characters, and quests as drafts before approval.
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
            {showDraftZoneSuggestions && zones.length > 0 && (
              <div className="absolute z-20 mt-1 max-h-60 w-full overflow-y-auto rounded border border-gray-200 bg-white shadow">
                {zones
                  .filter((zone) =>
                    zone.name.toLowerCase().includes(draftZoneQuery.toLowerCase())
                  )
                  .map((zone) => (
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
          <div className="mt-4 grid grid-cols-3 gap-3">
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
                Characters
              </label>
              <input
                className="w-full rounded border border-gray-300 px-2 py-2 text-sm"
                value={characterCount}
                onChange={(e) => setCharacterCount(e.target.value)}
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">
                Quests
              </label>
              <input
                className="w-full rounded border border-gray-300 px-2 py-2 text-sm"
                value={questCount}
                onChange={(e) => setQuestCount(e.target.value)}
              />
            </div>
          </div>
          <button
            className="mt-5 w-full rounded bg-indigo-600 px-4 py-2 text-white hover:bg-indigo-500 disabled:opacity-60"
            onClick={handleCreateDraft}
            disabled={creatingDraft || !draftZoneId}
          >
            {creatingDraft ? 'Queuing...' : 'Create draft'}
          </button>
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
              {showFilterZoneSuggestions && zones.length > 0 && (
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
                  {zones
                    .filter((zone) =>
                      zone.name.toLowerCase().includes(filterZoneQuery.toLowerCase())
                    )
                    .map((zone) => (
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
              {jobs.map((job) => (
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
                        Counts: {job.placeCount} places, {job.characterCount} characters, {job.questCount} quests
                      </p>
                    </div>
                    <span
                      className={`inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold text-white ${statusBadgeClass(
                        job.status
                      )}`}
                    >
                      {job.status.replace(/_/g, ' ')}
                    </span>
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
                          <div className="font-semibold">Quests</div>
                          <div className="mt-2 space-y-3 text-xs text-gray-600">
                            {(job.draft.quests || []).map((quest) => (
                              <div
                                key={quest.draftId}
                                className="rounded border border-gray-100 bg-gray-50 p-3"
                              >
                                <div className="text-sm font-semibold text-gray-800">
                                  {quest.name || 'Untitled quest'}
                                </div>
                                <div>Place ID: {quest.placeId || 'n/a'}</div>
                                <div>
                                  Quest giver draft ID:{' '}
                                  {quest.questGiverDraftId || 'n/a'}
                                </div>
                                {typeof quest.gold === 'number' && (
                                  <div>Gold: {quest.gold}</div>
                                )}
                                {quest.description && (
                                  <div className="mt-1 text-gray-500 whitespace-pre-wrap">
                                    {quest.description}
                                  </div>
                                )}
                                {quest.challengeQuestion && (
                                  <div className="mt-2 text-gray-500">
                                    Challenge: {quest.challengeQuestion}
                                  </div>
                                )}
                                {quest.acceptanceDialogue &&
                                  quest.acceptanceDialogue.length > 0 && (
                                    <div className="mt-2">
                                      <div className="font-semibold text-gray-600">
                                        Acceptance dialogue
                                      </div>
                                      <div className="mt-1 space-y-1 text-gray-500">
                                        {quest.acceptanceDialogue.map((line, idx) => (
                                          <div key={`${quest.draftId}-line-${idx}`}>
                                            {line}
                                          </div>
                                        ))}
                                      </div>
                                    </div>
                                  )}
                              </div>
                            ))}
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
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ZoneSeedJobs;
