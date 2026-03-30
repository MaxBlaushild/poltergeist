import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAPI } from '@poltergeist/contexts';
import {
  QuestArchetypeSuggestionDraft,
  QuestArchetypeSuggestionJob,
} from '@poltergeist/types';
import './questArchetypeTheme.css';

type GeneratorFormState = {
  count: string;
  themePrompt: string;
  familyTagsText: string;
  characterTagsText: string;
  internalTagsText: string;
  requiredLocationMetadataTagsText: string;
};

const emptyGeneratorForm = (): GeneratorFormState => ({
  count: '12',
  themePrompt: '',
  familyTagsText: '',
  characterTagsText: '',
  internalTagsText: '',
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
    typeof error.response.error !== 'string' &&
    typeof (error.response.data as { error?: unknown }).error === 'string'
  ) {
    return (error.response.data as { error: string }).error;
  }
  return fallback;
};

export const QuestArchetypeGenerator = () => {
  const { apiClient } = useAPI();
  const [form, setForm] = useState<GeneratorFormState>(emptyGeneratorForm);
  const [jobs, setJobs] = useState<QuestArchetypeSuggestionJob[]>([]);
  const [selectedJobId, setSelectedJobId] = useState<string>('');
  const [drafts, setDrafts] = useState<QuestArchetypeSuggestionDraft[]>([]);
  const [loadingJobs, setLoadingJobs] = useState(false);
  const [loadingDrafts, setLoadingDrafts] = useState(false);
  const [queueing, setQueueing] = useState(false);
  const [pageError, setPageError] = useState<string | null>(null);
  const [jobActionError, setJobActionError] = useState<string | null>(null);
  const [convertingDraftId, setConvertingDraftId] = useState<string | null>(null);
  const [deletingDraftId, setDeletingDraftId] = useState<string | null>(null);

  const selectedJob = useMemo(
    () => jobs.find((job) => job.id === selectedJobId) ?? null,
    [jobs, selectedJobId]
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
        console.error('Failed to load quest archetype suggestion drafts', error);
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
        {
          count: Math.max(1, parseInt(form.count, 10) || 1),
          themePrompt: form.themePrompt.trim(),
          familyTags: parseTags(form.familyTagsText),
          characterTags: parseTags(form.characterTagsText),
          internalTags: parseTags(form.internalTagsText),
          requiredLocationMetadataTags: parseTags(
            form.requiredLocationMetadataTagsText
          ),
        }
      );
      setJobs((current) => [created, ...current.filter((job) => job.id !== created.id)]);
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
      await apiClient.post(`/sonar/questArchetypeSuggestionDrafts/${draftId}/convert`, {});
      if (selectedJobId) {
        await fetchDrafts(selectedJobId);
      }
    } catch (error) {
      console.error('Failed to convert quest archetype suggestion draft', error);
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
      await apiClient.delete(`/sonar/questArchetypeSuggestionDrafts/${draftId}`);
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

  return (
    <div className="qa-theme">
      <div className="qa-shell">
        <header className="qa-hero">
          <div>
            <div className="qa-kicker">Questing</div>
            <h1 className="qa-title">Quest Archetype Generator</h1>
            <p className="qa-subtitle">
              Generate batches of draft archetype bundles, review node-by-node
              content, and convert the strongest ones into live quest archetypes.
            </p>
          </div>
        </header>

        <section className="qa-grid" style={{ gridTemplateColumns: '1.2fr 1fr' }}>
          <div className="qa-panel">
            <div className="qa-card-title">Queue Suggestion Job</div>
            <p className="qa-muted" style={{ marginTop: 8 }}>
              Use this to generate reusable draft bundles with explicit node
              content, location metadata tags, and suggested template copy.
            </p>
            <div className="qa-form-grid" style={{ marginTop: 18 }}>
              <div className="qa-field">
                <div className="qa-label">Batch Size</div>
                <input
                  className="qa-input"
                  value={form.count}
                  onChange={(event) =>
                    setForm((current) => ({ ...current, count: event.target.value }))
                  }
                  type="number"
                  min={1}
                  max={100}
                />
              </div>
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
            </div>

            <div className="qa-actions" style={{ marginTop: 18 }}>
              <button
                className="qa-btn qa-btn-primary"
                onClick={() => void handleQueueJob()}
                disabled={queueing}
              >
                {queueing ? 'Queueing...' : 'Queue Draft Job'}
              </button>
              <button
                className="qa-btn qa-btn-outline"
                onClick={() => void fetchJobs()}
                disabled={loadingJobs}
              >
                {loadingJobs ? 'Refreshing...' : 'Refresh Jobs'}
              </button>
            </div>

            {(jobActionError || pageError) && (
              <div className="qa-chip danger" style={{ marginTop: 14 }}>
                {jobActionError || pageError}
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
                        <div className="qa-meta">{formatDate(job.createdAt)}</div>
                      </div>
                      <div className={statusChipClass(job.status)}>
                        {formatStatus(job.status)}
                      </div>
                    </div>
                    <div className="qa-meta">
                      Drafts: {job.createdCount}/{job.count}
                    </div>
                    {job.themePrompt && (
                      <div className="qa-muted" style={{ marginTop: 8 }}>
                        {job.themePrompt}
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
                  ? `Showing drafts for job ${selectedJob.id.slice(0, 8)}...`
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

          {!loadingDrafts && selectedJob && drafts.length === 0 && (
            <div className="qa-muted" style={{ marginTop: 16 }}>
              {isPendingStatus(selectedJob.status)
                ? 'This job is still generating drafts.'
                : 'No drafts were generated for this job.'}
            </div>
          )}

          {!loadingDrafts && drafts.length > 0 && (
            <div className="qa-tree" style={{ marginTop: 18 }}>
              {drafts.map((draft) => (
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
                      <div className="qa-stat-label">Route Steps</div>
                      <div className="qa-stat-value">{draft.steps.length}</div>
                    </div>
                  </div>

                  <div style={{ marginTop: 18 }}>
                    <div className="qa-meta">
                      Character tags: {(draft.characterTags ?? []).join(', ') || 'none'}
                    </div>
                    <div className="qa-meta" style={{ marginTop: 6 }}>
                      Internal tags: {(draft.internalTags ?? []).join(', ') || 'none'}
                    </div>
                    <p style={{ marginTop: 12 }}>{draft.description}</p>
                  </div>

                  {draft.acceptanceDialogue?.length > 0 && (
                    <div style={{ marginTop: 18 }}>
                      <div className="qa-stat-label">Acceptance Dialogue</div>
                      <div className="qa-tree" style={{ marginTop: 8 }}>
                        {draft.acceptanceDialogue.map((line, index) => (
                          <div key={`${draft.id}-dialogue-${index}`} className="qa-node-card">
                            {line}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  <div style={{ marginTop: 18 }}>
                    <div className="qa-stat-label">Node Plan</div>
                    <div className="qa-tree" style={{ marginTop: 10 }}>
                      {draft.steps.map((step, index) => (
                        <div key={`${draft.id}-step-${index}`} className="qa-node-card">
                          <div className="qa-card-title" style={{ fontSize: 15 }}>
                            Step {index + 1}: {step.source} {step.content}
                          </div>
                          <div className="qa-meta" style={{ marginTop: 8 }}>
                            Location concept: {step.locationConcept}
                          </div>
                          {step.locationArchetypeName && (
                            <div className="qa-meta" style={{ marginTop: 6 }}>
                              Location archetype: {step.locationArchetypeName}
                            </div>
                          )}
                          {step.distanceMeters != null && (
                            <div className="qa-meta" style={{ marginTop: 6 }}>
                              Distance: {step.distanceMeters}m
                            </div>
                          )}
                          <div className="qa-meta" style={{ marginTop: 6 }}>
                            Metadata tags: {(step.locationMetadataTags ?? []).join(', ')}
                          </div>
                          <div className="qa-meta" style={{ marginTop: 6 }}>
                            Template concept: {step.templateConcept}
                          </div>

                          {step.content === 'challenge' && (
                            <div style={{ marginTop: 12 }}>
                              <div className="qa-stat-label">Challenge Template Draft</div>
                              <div className="qa-meta" style={{ marginTop: 6 }}>
                                Question: {step.challengeQuestion}
                              </div>
                              <div className="qa-meta" style={{ marginTop: 6 }}>
                                Description: {step.challengeDescription}
                              </div>
                              <div className="qa-meta" style={{ marginTop: 6 }}>
                                Submission: {step.challengeSubmissionType || 'photo'}
                              </div>
                            </div>
                          )}

                          {step.content === 'scenario' && (
                            <div style={{ marginTop: 12 }}>
                              <div className="qa-stat-label">Scenario Template Draft</div>
                              <div className="qa-meta" style={{ marginTop: 6 }}>
                                Prompt: {step.scenarioPrompt}
                              </div>
                              <div className="qa-meta" style={{ marginTop: 6 }}>
                                Beats: {(step.scenarioBeats ?? []).join(', ') || 'none'}
                              </div>
                            </div>
                          )}

                          {step.content === 'monster' && (
                            <div style={{ marginTop: 12 }}>
                              <div className="qa-stat-label">Monster Encounter Draft</div>
                              <div className="qa-meta" style={{ marginTop: 6 }}>
                                Templates: {(step.monsterTemplateNames ?? []).join(', ') || 'none'}
                              </div>
                              <div className="qa-meta" style={{ marginTop: 6 }}>
                                Tone: {(step.encounterTone ?? []).join(', ') || 'none'}
                              </div>
                            </div>
                          )}

                          {(step.potentialContent ?? []).length > 0 && (
                            <div style={{ marginTop: 12 }}>
                              <div className="qa-stat-label">Potential Content</div>
                              <div className="qa-tree" style={{ marginTop: 8 }}>
                                {step.potentialContent.map((item, ideaIndex) => (
                                  <div
                                    key={`${draft.id}-step-${index}-idea-${ideaIndex}`}
                                    className="qa-node-card"
                                  >
                                    {item}
                                  </div>
                                ))}
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
                          <div key={`${draft.id}-warning-${index}`} className="qa-chip danger">
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
                          Challenge seeds: {draft.challengeTemplateSeeds.join(' | ')}
                        </div>
                      ) : null}
                      {draft.scenarioTemplateSeeds?.length ? (
                        <div className="qa-meta" style={{ marginTop: 8 }}>
                          Scenario seeds: {draft.scenarioTemplateSeeds.join(' | ')}
                        </div>
                      ) : null}
                      {draft.monsterTemplateSeeds?.length ? (
                        <div className="qa-meta" style={{ marginTop: 8 }}>
                          Monster seeds: {draft.monsterTemplateSeeds.join(' | ')}
                        </div>
                      ) : null}
                    </div>
                  )}

                  <div className="qa-actions" style={{ marginTop: 18 }}>
                    {draft.questArchetypeId ? (
                      <Link to="/quest-archetypes" className="qa-btn qa-btn-outline">
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
                        {deletingDraftId === draft.id ? 'Deleting...' : 'Delete Draft'}
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

export default QuestArchetypeGenerator;
