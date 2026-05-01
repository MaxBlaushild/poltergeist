import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAPI } from '@poltergeist/contexts';
import {
  QuestArchetypeSuggestionDraft,
  QuestArchetypeSuggestionJob,
  QuestArchetypeSuggestionNode,
} from '@poltergeist/types';
import { useQuestArchetypes } from '../contexts/questArchetypes.tsx';
import { useZoneKinds, zoneKindLabel } from './zoneKindHelpers.ts';
import './questArchetypeTheme.css';

type GeneratorFormState = {
  count: string;
  zoneKind: string;
  themePrompt: string;
  familyTagsText: string;
  familyMixTargets: Record<string, string>;
  characterTagsText: string;
  internalTagsText: string;
  requiredLocationArchetypeIds: string[];
  requiredLocationMetadataTagsText: string;
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

const emptyFamilyMixTargets = () =>
  QUEST_FAMILY_OPTIONS.reduce<Record<string, string>>((accumulator, family) => {
    accumulator[family.slug] = '0';
    return accumulator;
  }, {});

const emptyGeneratorForm = (): GeneratorFormState => ({
  count: '12',
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
  const [pageError, setPageError] = useState<string | null>(null);
  const [jobActionError, setJobActionError] = useState<string | null>(null);
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
        {
          count: Math.max(1, parseInt(form.count, 10) || 1),
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
        }
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
                <div className="qa-label">Required Location Archetypes</div>
                <div
                  className="qa-muted"
                  style={{ marginTop: 6, marginBottom: 10 }}
                >
                  Every generated draft should include each checked archetype at
                  least once.
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

            <div className="qa-actions" style={{ marginTop: 18 }}>
              <button
                className="qa-btn qa-btn-primary"
                onClick={() => void handleQueueJob()}
                disabled={
                  queueing ||
                  requestedFamilyMixCount >
                    Math.max(1, parseInt(form.count, 10) || 1)
                }
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
                      Drafts: {job.createdCount}/{job.count}
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
