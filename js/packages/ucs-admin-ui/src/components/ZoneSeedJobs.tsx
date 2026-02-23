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
  const [selectedZoneId, setSelectedZoneId] = useState<string>('');
  const [jobs, setJobs] = useState<ZoneSeedJob[]>([]);
  const [loadingJobs, setLoadingJobs] = useState(false);
  const [creatingDraft, setCreatingDraft] = useState(false);
  const [approvingId, setApprovingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [placeCount, setPlaceCount] = useState('8');
  const [characterCount, setCharacterCount] = useState('4');
  const [questCount, setQuestCount] = useState('4');

  const selectedZone = useMemo<Zone | undefined>(() => {
    return zones.find((zone) => zone.id === selectedZoneId);
  }, [zones, selectedZoneId]);

  useEffect(() => {
    if (zones.length === 0) {
      refreshZones();
      return;
    }
    if (!selectedZoneId && zones.length > 0) {
      setSelectedZoneId(zones[0].id);
    }
  }, [zones, selectedZoneId, refreshZones]);

  useEffect(() => {
    if (!selectedZoneId) return;
    fetchJobs(selectedZoneId);
  }, [selectedZoneId]);

  const fetchJobs = async (zoneId: string) => {
    setLoadingJobs(true);
    setError(null);
    try {
      const response = await apiClient.get<ZoneSeedJob[]>(
        '/sonar/admin/zone-seed-jobs',
        { zoneId, limit: 25 }
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
    if (!selectedZoneId) {
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
        `/sonar/admin/zones/${selectedZoneId}/seed-draft`,
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
          onClick={() => selectedZoneId && fetchJobs(selectedZoneId)}
          disabled={loadingJobs || !selectedZoneId}
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
          <select
            className="w-full rounded border border-gray-300 px-3 py-2 text-sm"
            value={selectedZoneId}
            onChange={(e) => setSelectedZoneId(e.target.value)}
          >
            <option value="">Select a zone</option>
            {zones.map((zone) => (
              <option key={zone.id} value={zone.id}>
                {zone.name}
              </option>
            ))}
          </select>
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
            disabled={creatingDraft || !selectedZoneId}
          >
            {creatingDraft ? 'Queuing...' : 'Create draft'}
          </button>
        </div>

        <div className="lg:col-span-2 rounded-lg border border-gray-200 bg-white p-5 shadow-sm">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Draft jobs</h2>
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
                      <div className="mt-3 space-y-3 text-sm text-gray-700">
                        <div>
                          <div className="font-semibold">Fantasy branding</div>
                          <div className="text-sm text-gray-600">
                            {job.draft.fantasyName || 'Untitled district'}
                          </div>
                          {job.draft.zoneDescription && (
                            <p className="mt-1 text-xs text-gray-500">
                              {job.draft.zoneDescription}
                            </p>
                          )}
                        </div>
                        <div>
                          <div className="font-semibold">Points of interest</div>
                          <ul className="mt-1 list-disc list-inside text-xs text-gray-500">
                            {(job.draft.pointsOfInterest || []).slice(0, 6).map((poi) => (
                              <li key={poi.draftId}>
                                {poi.name} ({poi.placeId})
                              </li>
                            ))}
                          </ul>
                        </div>
                        <div>
                          <div className="font-semibold">Characters</div>
                          <ul className="mt-1 list-disc list-inside text-xs text-gray-500">
                            {(job.draft.characters || []).slice(0, 6).map((character) => (
                              <li key={character.draftId}>
                                {character.name} ({character.placeId})
                              </li>
                            ))}
                          </ul>
                        </div>
                        <div>
                          <div className="font-semibold">Quests</div>
                          <ul className="mt-1 list-disc list-inside text-xs text-gray-500">
                            {(job.draft.quests || []).slice(0, 6).map((quest) => (
                              <li key={quest.draftId}>
                                {quest.name} (giver {quest.questGiverDraftId.slice(0, 6)})
                              </li>
                            ))}
                          </ul>
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
