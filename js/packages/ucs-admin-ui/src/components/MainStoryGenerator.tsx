import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAPI } from '@poltergeist/contexts';
import {
  MainStorySuggestionDraft,
  MainStorySuggestionJob,
  MainStoryTemplate,
} from '@poltergeist/types';
import { useQuestArchetypes } from '../contexts/questArchetypes.tsx';
import './questArchetypeTheme.css';

type GeneratorFormState = {
  count: string;
  questCount: string;
  themePrompt: string;
  districtFit: string;
  tone: string;
  familyTagsText: string;
  characterTagsText: string;
  internalTagsText: string;
  requiredLocationArchetypeIds: string[];
  requiredLocationMetadataTagsText: string;
};

const emptyGeneratorForm = (): GeneratorFormState => ({
  count: '2',
  questCount: '15',
  themePrompt: '',
  districtFit: '',
  tone: '',
  familyTagsText: '',
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
    typeof (error.response.data as { error?: unknown }).error === 'string'
  ) {
    return (error.response.data as { error: string }).error;
  }
  return fallback;
};

export const MainStoryGenerator = () => {
  const { apiClient } = useAPI();
  const { locationArchetypes } = useQuestArchetypes();
  const [form, setForm] = useState<GeneratorFormState>(emptyGeneratorForm);
  const [jobs, setJobs] = useState<MainStorySuggestionJob[]>([]);
  const [selectedJobId, setSelectedJobId] = useState<string>('');
  const [drafts, setDrafts] = useState<MainStorySuggestionDraft[]>([]);
  const [loadingJobs, setLoadingJobs] = useState(false);
  const [loadingDrafts, setLoadingDrafts] = useState(false);
  const [queueing, setQueueing] = useState(false);
  const [pageError, setPageError] = useState<string | null>(null);
  const [jobActionError, setJobActionError] = useState<string | null>(null);
  const [locationArchetypeSearch, setLocationArchetypeSearch] = useState('');
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
  const selectedLocationArchetypes = useMemo(
    () =>
      form.requiredLocationArchetypeIds
        .map((id) =>
          locationArchetypes.find((archetype) => archetype.id === id)
        )
        .filter(Boolean),
    [form.requiredLocationArchetypeIds, locationArchetypes]
  );
  const filteredLocationArchetypes = useMemo(() => {
    const query = locationArchetypeSearch.trim().toLowerCase();
    return locationArchetypes
      .filter(
        (archetype) =>
          !form.requiredLocationArchetypeIds.includes(archetype.id) &&
          (!query ||
            archetype.name.toLowerCase().includes(query) ||
            (archetype.description || '').toLowerCase().includes(query))
      )
      .slice(0, 12);
  }, [
    form.requiredLocationArchetypeIds,
    locationArchetypes,
    locationArchetypeSearch,
  ]);

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

  const fetchJobs = useCallback(async () => {
    setLoadingJobs(true);
    try {
      const response = await apiClient.get<MainStorySuggestionJob[]>(
        '/sonar/mainStorySuggestionJobs?limit=30'
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
      console.error('Failed to load main story suggestion jobs', error);
      setPageError(
        extractApiErrorMessage(
          error,
          'Failed to load main story generator jobs.'
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
        const response = await apiClient.get<MainStorySuggestionDraft[]>(
          `/sonar/mainStorySuggestionJobs/${jobId}/drafts`
        );
        setDrafts(response);
        setJobActionError(null);
      } catch (error) {
        console.error('Failed to load main story suggestion drafts', error);
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
      const created = await apiClient.post<MainStorySuggestionJob>(
        '/sonar/mainStorySuggestionJobs',
        {
          count: Math.max(1, parseInt(form.count, 10) || 1),
          questCount: Math.max(3, parseInt(form.questCount, 10) || 15),
          themePrompt: form.themePrompt.trim(),
          districtFit: form.districtFit.trim(),
          tone: form.tone.trim(),
          familyTags: parseTags(form.familyTagsText),
          characterTags: parseTags(form.characterTagsText),
          internalTags: parseTags(form.internalTagsText),
          requiredLocationArchetypeIds: form.requiredLocationArchetypeIds,
          requiredLocationMetadataTags: parseTags(
            form.requiredLocationMetadataTagsText
          ),
        }
      );
      setJobs((current) => [
        created,
        ...current.filter((job) => job.id !== created.id),
      ]);
      setSelectedJobId(created.id);
      setDrafts([]);
      setLocationArchetypeSearch('');
    } catch (error) {
      console.error('Failed to queue main story suggestion job', error);
      setJobActionError(
        extractApiErrorMessage(error, 'Failed to queue main story job.')
      );
    } finally {
      setQueueing(false);
    }
  };

  const handleConvertDraft = async (draftId: string) => {
    setConvertingDraftId(draftId);
    setJobActionError(null);
    try {
      await apiClient.post<MainStoryTemplate>(
        `/sonar/mainStorySuggestionDrafts/${draftId}/convert`,
        {}
      );
      if (selectedJobId) {
        await fetchDrafts(selectedJobId);
      }
    } catch (error) {
      console.error('Failed to convert main story suggestion draft', error);
      setJobActionError(
        extractApiErrorMessage(error, 'Failed to convert draft into template.')
      );
    } finally {
      setConvertingDraftId(null);
    }
  };

  const handleDeleteDraft = async (draftId: string) => {
    setDeletingDraftId(draftId);
    setJobActionError(null);
    try {
      await apiClient.delete(`/sonar/mainStorySuggestionDrafts/${draftId}`);
      setDrafts((current) => current.filter((draft) => draft.id !== draftId));
    } catch (error) {
      console.error('Failed to delete main story suggestion draft', error);
      setJobActionError(
        extractApiErrorMessage(error, 'Failed to delete generated draft.')
      );
    } finally {
      setDeletingDraftId(null);
    }
  };

  const addRequiredLocationArchetype = (archetypeId: string) => {
    setForm((current) => {
      if (current.requiredLocationArchetypeIds.includes(archetypeId)) {
        return current;
      }
      return {
        ...current,
        requiredLocationArchetypeIds: [
          ...current.requiredLocationArchetypeIds,
          archetypeId,
        ],
      };
    });
    setLocationArchetypeSearch('');
  };

  const removeRequiredLocationArchetype = (archetypeId: string) => {
    setForm((current) => ({
      ...current,
      requiredLocationArchetypeIds: current.requiredLocationArchetypeIds.filter(
        (id) => id !== archetypeId
      ),
    }));
  };

  return (
    <div className="qa-theme">
      <div className="qa-shell">
        <header className="qa-hero">
          <div>
            <div className="qa-kicker">Main Story</div>
            <h1 className="qa-title">District Campaign Generator</h1>
            <p className="qa-subtitle">
              Generate reviewable 15-quest campaign drafts with a shared story
              bible, ordered beats, and beat-level quest archetype packages.
            </p>
          </div>
        </header>

        <section
          className="qa-grid"
          style={{ gridTemplateColumns: '1.1fr 1fr' }}
        >
          <div className="qa-panel">
            <div className="qa-card-title">Queue Story Job</div>
            <p className="qa-muted" style={{ marginTop: 8 }}>
              Each draft is a complete main story outline with beat-by-beat
              quest content that can be converted into a reusable template.
            </p>
            <div className="qa-form-grid" style={{ marginTop: 18 }}>
              <div className="qa-field">
                <div className="qa-label">Campaign Draft Count</div>
                <input
                  className="qa-input"
                  value={form.count}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      count: event.target.value,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Quests Per Story</div>
                <input
                  className="qa-input"
                  value={form.questCount}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      questCount: event.target.value,
                    }))
                  }
                />
              </div>
              <div className="qa-field qa-field--full">
                <div className="qa-label">Theme Prompt</div>
                <textarea
                  className="qa-textarea"
                  rows={3}
                  value={form.themePrompt}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      themePrompt: event.target.value,
                    }))
                  }
                  placeholder="A district-scale occult conspiracy built around disappearing transit workers and a false saint."
                />
              </div>
              <div className="qa-field qa-field--full">
                <div className="qa-label">District Fit</div>
                <textarea
                  className="qa-textarea"
                  rows={2}
                  value={form.districtFit}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      districtFit: event.target.value,
                    }))
                  }
                  placeholder="Dense nightlife district with rail lines, old churches, food markets, and waterfront spillover."
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Tone</div>
                <input
                  className="qa-input"
                  value={form.tone}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      tone: event.target.value,
                    }))
                  }
                  placeholder="haunting, street-level, intimate"
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
                  placeholder="main_story, mystery, occult"
                />
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
                  placeholder="priest, fixer, transit_worker"
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
                  placeholder="district_arc, recurring_antagonist"
                />
              </div>
              <div className="qa-field qa-field--full">
                <div className="qa-label">Required Location Archetypes</div>
                <div className="qa-tag-row" style={{ marginBottom: 10 }}>
                  {selectedLocationArchetypes.length > 0 ? (
                    selectedLocationArchetypes.map((archetype) => (
                      <button
                        key={archetype?.id}
                        type="button"
                        className="qa-chip"
                        onClick={() =>
                          archetype &&
                          removeRequiredLocationArchetype(archetype.id)
                        }
                        style={{
                          display: 'inline-flex',
                          alignItems: 'center',
                          gap: 8,
                          border: 'none',
                          cursor: 'pointer',
                        }}
                        title="Remove archetype"
                      >
                        <span>{archetype?.name}</span>
                        <span aria-hidden="true">x</span>
                      </button>
                    ))
                  ) : (
                    <span className="qa-muted">
                      No required archetypes selected.
                    </span>
                  )}
                </div>
                <div style={{ position: 'relative' }}>
                  <input
                    className="qa-input"
                    value={locationArchetypeSearch}
                    onChange={(event) =>
                      setLocationArchetypeSearch(event.target.value)
                    }
                    placeholder="Search location archetypes..."
                  />
                  {(locationArchetypeSearch.trim() ||
                    filteredLocationArchetypes.length > 0) && (
                    <div
                      className="qa-panel"
                      style={{
                        marginTop: 8,
                        maxHeight: 240,
                        overflowY: 'auto',
                        padding: 8,
                      }}
                    >
                      {filteredLocationArchetypes.length > 0 ? (
                        filteredLocationArchetypes.map((archetype) => (
                          <button
                            key={archetype.id}
                            type="button"
                            className="qa-job-card"
                            style={{
                              width: '100%',
                              textAlign: 'left',
                              marginBottom: 8,
                            }}
                            onClick={() =>
                              addRequiredLocationArchetype(archetype.id)
                            }
                          >
                            <div className="qa-job-card__title">
                              {archetype.name}
                            </div>
                            <div className="qa-job-card__meta">
                              {archetype.description || 'No description'}
                            </div>
                          </button>
                        ))
                      ) : (
                        <div className="qa-empty">
                          No matching location archetypes.
                        </div>
                      )}
                    </div>
                  )}
                </div>
              </div>
              <div className="qa-field qa-field--full">
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
                  placeholder="nightlife, market, ritual, waterfront"
                />
              </div>
            </div>
            <div className="qa-actions" style={{ marginTop: 18 }}>
              <button
                className="qa-button qa-button--primary"
                disabled={queueing}
                onClick={() => void handleQueueJob()}
              >
                {queueing ? 'Queueing...' : 'Queue Main Story Job'}
              </button>
            </div>
          </div>

          <div className="qa-panel">
            <div className="qa-card-title">Recent Jobs</div>
            <p className="qa-muted" style={{ marginTop: 8 }}>
              Pick a job to inspect its campaign drafts and convert the ones
              that feel strong enough to become reusable main stories.
            </p>
            {pageError && (
              <div className="qa-alert qa-alert--danger">{pageError}</div>
            )}
            {jobActionError && (
              <div className="qa-alert qa-alert--danger">{jobActionError}</div>
            )}
            {loadingJobs ? (
              <div className="qa-empty">Loading jobs...</div>
            ) : jobs.length === 0 ? (
              <div className="qa-empty">No main story jobs yet.</div>
            ) : (
              <div className="qa-list" style={{ marginTop: 16 }}>
                {jobs.map((job) => {
                  const active = job.id === selectedJobId;
                  return (
                    <button
                      key={job.id}
                      className={`qa-job-card ${active ? 'is-active' : ''}`}
                      onClick={() => setSelectedJobId(job.id)}
                    >
                      <div className="qa-job-card__header">
                        <div>
                          <div className="qa-job-card__title">
                            {job.themePrompt || 'Untitled story prompt'}
                          </div>
                          <div className="qa-job-card__meta">
                            {job.count} draft(s) · {job.questCount} beats
                          </div>
                        </div>
                        <span className={statusChipClass(job.status)}>
                          {formatStatus(job.status)}
                        </span>
                      </div>
                      <div className="qa-job-card__body">
                        <div>
                          {job.districtFit || 'No district fit guidance.'}
                        </div>
                        <div className="qa-job-card__timestamp">
                          Updated {formatDate(job.updatedAt)}
                        </div>
                      </div>
                    </button>
                  );
                })}
              </div>
            )}
          </div>
        </section>

        <section className="qa-panel" style={{ marginTop: 24 }}>
          <div className="qa-card-title">Generated Campaign Drafts</div>
          {selectedJob && (
            <div className="qa-muted" style={{ marginTop: 8 }}>
              Required archetypes:{' '}
              {selectedJobRequiredLocationArchetypes.length > 0
                ? selectedJobRequiredLocationArchetypes.join(', ')
                : 'none'}
            </div>
          )}
          {loadingDrafts ? (
            <div className="qa-empty">Loading drafts...</div>
          ) : drafts.length === 0 ? (
            <div className="qa-empty">No campaign drafts for this job yet.</div>
          ) : (
            <div className="qa-stack" style={{ marginTop: 18 }}>
              {drafts.map((draft) => (
                <article key={draft.id} className="qa-draft-card">
                  <div className="qa-draft-card__header">
                    <div>
                      <div className="qa-draft-card__title">{draft.name}</div>
                      <div className="qa-draft-card__meta">
                        {draft.beats.length} beats · tone {draft.tone || 'n/a'}
                      </div>
                    </div>
                    <span className={statusChipClass(draft.status)}>
                      {formatStatus(draft.status)}
                    </span>
                  </div>
                  <div className="qa-draft-card__body">
                    <p className="qa-copy">{draft.premise}</p>
                    {draft.districtFit && (
                      <p className="qa-copy">
                        <strong>District Fit:</strong> {draft.districtFit}
                      </p>
                    )}
                    <p className="qa-copy">
                      <strong>Climax:</strong> {draft.climaxSummary || 'n/a'}
                    </p>
                    <p className="qa-copy">
                      <strong>Resolution:</strong>{' '}
                      {draft.resolutionSummary || 'n/a'}
                    </p>

                    <div className="qa-tag-row">
                      {draft.themeTags.map((tag) => (
                        <span
                          key={`${draft.id}-theme-${tag}`}
                          className="qa-chip"
                        >
                          {tag}
                        </span>
                      ))}
                    </div>

                    {draft.warnings.length > 0 && (
                      <div className="qa-alert qa-alert--warning">
                        {draft.warnings.join(' | ')}
                      </div>
                    )}

                    <div className="qa-stack" style={{ marginTop: 16 }}>
                      {draft.beats.map((beat) => (
                        <div
                          key={`${draft.id}-beat-${beat.orderIndex}`}
                          className="qa-step-card"
                        >
                          <div className="qa-step-card__header">
                            <div>
                              <div className="qa-step-card__title">
                                {beat.orderIndex}.{' '}
                                {beat.chapterTitle || beat.name}
                              </div>
                              <div className="qa-step-card__meta">
                                Act {beat.act} ·{' '}
                                {beat.storyRole || 'story_beat'} ·{' '}
                                {beat.questArchetypeName || beat.name}
                              </div>
                            </div>
                          </div>
                          <div className="qa-copy">{beat.chapterSummary}</div>
                          <div className="qa-copy">
                            <strong>Purpose:</strong> {beat.purpose || 'n/a'}
                          </div>
                          <div className="qa-copy">
                            <strong>What Changes:</strong>{' '}
                            {beat.whatChanges || 'n/a'}
                          </div>
                          {(beat.questGiverCharacterKey ||
                            beat.questGiverCharacterName) && (
                            <div className="qa-copy">
                              <strong>Quest Giver:</strong>{' '}
                              {beat.questGiverCharacterName ||
                                beat.questGiverCharacterKey}
                              {beat.questGiverCharacterName &&
                              beat.questGiverCharacterKey
                                ? ` (${beat.questGiverCharacterKey})`
                                : ''}
                            </div>
                          )}
                          {beat.requiredStoryFlags.length > 0 && (
                            <div className="qa-copy">
                              <strong>Requires:</strong>{' '}
                              {beat.requiredStoryFlags.join(', ')}
                            </div>
                          )}
                          {beat.setStoryFlags.length > 0 && (
                            <div className="qa-copy">
                              <strong>Sets:</strong>{' '}
                              {beat.setStoryFlags.join(', ')}
                            </div>
                          )}
                          {beat.clearStoryFlags.length > 0 && (
                            <div className="qa-copy">
                              <strong>Clears:</strong>{' '}
                              {beat.clearStoryFlags.join(', ')}
                            </div>
                          )}
                          {beat.questGiverRelationshipEffects &&
                            Object.values(
                              beat.questGiverRelationshipEffects
                            ).some((value) => (value ?? 0) !== 0) && (
                              <div className="qa-copy">
                                <strong>Quest Giver Relationship:</strong>{' '}
                                {[
                                  beat.questGiverRelationshipEffects.trust
                                    ? `Trust ${beat.questGiverRelationshipEffects.trust > 0 ? '+' : ''}${beat.questGiverRelationshipEffects.trust}`
                                    : null,
                                  beat.questGiverRelationshipEffects.respect
                                    ? `Respect ${beat.questGiverRelationshipEffects.respect > 0 ? '+' : ''}${beat.questGiverRelationshipEffects.respect}`
                                    : null,
                                  beat.questGiverRelationshipEffects.fear
                                    ? `Fear ${beat.questGiverRelationshipEffects.fear > 0 ? '+' : ''}${beat.questGiverRelationshipEffects.fear}`
                                    : null,
                                  beat.questGiverRelationshipEffects.debt
                                    ? `Debt ${beat.questGiverRelationshipEffects.debt > 0 ? '+' : ''}${beat.questGiverRelationshipEffects.debt}`
                                    : null,
                                ]
                                  .filter(Boolean)
                                  .join(', ')}
                              </div>
                            )}
                          {(beat.questGiverAfterDescription ||
                            beat.questGiverAfterDialogue.length > 0) && (
                            <div className="qa-copy">
                              <strong>Quest Giver Aftermath:</strong>{' '}
                              {beat.questGiverAfterDescription || 'n/a'}
                              {beat.questGiverAfterDialogue.length > 0
                                ? ` ${beat.questGiverAfterDialogue.join(' / ')}`
                                : ''}
                            </div>
                          )}
                          {(beat.worldChanges?.length ?? 0) > 0 && (
                            <div className="qa-copy">
                              <strong>World Changes:</strong>{' '}
                              {(beat.worldChanges ?? [])
                                .map((change) => {
                                  if (change.type === 'move_character') {
                                    return `Move ${change.targetKey || 'character'} to ${change.destinationHint || 'a new location'}`;
                                  }
                                  if (change.type === 'show_poi_text') {
                                    return `Update ${change.targetKey || 'poi'} text`;
                                  }
                                  return change.type;
                                })
                                .join(' / ')}
                            </div>
                          )}
                          {((beat.unlockedScenarios?.length ?? 0) > 0 ||
                            (beat.unlockedChallenges?.length ?? 0) > 0 ||
                            (beat.unlockedMonsterEncounters?.length ?? 0) >
                              0) && (
                            <div className="qa-copy">
                              <strong>Unlocked Content:</strong>{' '}
                              {[
                                ...(beat.unlockedScenarios ?? []).map(
                                  (scenario) =>
                                    `Scenario: ${scenario.name || scenario.prompt}`
                                ),
                                ...(beat.unlockedChallenges ?? []).map(
                                  (challenge) =>
                                    `Challenge: ${challenge.question}`
                                ),
                                ...(beat.unlockedMonsterEncounters ?? []).map(
                                  (encounter) => `Encounter: ${encounter.name}`
                                ),
                              ].join(' / ')}
                            </div>
                          )}
                          <div className="qa-tag-row">
                            {beat.requiredZoneTags.map((tag) => (
                              <span
                                key={`${draft.id}-beat-${beat.orderIndex}-tag-${tag}`}
                                className="qa-chip muted"
                              >
                                zone:{tag}
                              </span>
                            ))}
                            {beat.preferredContentMix.map((tag) => (
                              <span
                                key={`${draft.id}-beat-${beat.orderIndex}-mix-${tag}`}
                                className="qa-chip accent"
                              >
                                {tag}
                              </span>
                            ))}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                  <div className="qa-draft-card__footer">
                    {draft.mainStoryTemplateId ? (
                      <Link
                        className="qa-button qa-button--ghost"
                        to="#"
                        onClick={(event) => event.preventDefault()}
                      >
                        Converted
                      </Link>
                    ) : (
                      <button
                        className="qa-button qa-button--primary"
                        disabled={convertingDraftId === draft.id}
                        onClick={() => void handleConvertDraft(draft.id)}
                      >
                        {convertingDraftId === draft.id
                          ? 'Converting...'
                          : 'Convert to Main Story'}
                      </button>
                    )}
                    {!draft.mainStoryTemplateId && (
                      <button
                        className="qa-button qa-button--ghost"
                        disabled={deletingDraftId === draft.id}
                        onClick={() => void handleDeleteDraft(draft.id)}
                      >
                        {deletingDraftId === draft.id
                          ? 'Deleting...'
                          : 'Delete'}
                      </button>
                    )}
                  </div>
                </article>
              ))}
            </div>
          )}
        </section>
      </div>
    </div>
  );
};

export default MainStoryGenerator;
