import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAPI } from '@poltergeist/contexts';
import {
  District,
  MainStoryDistrictRun,
  MainStoryTemplate,
} from '@poltergeist/types';
import './questArchetypeTheme.css';

const formatDate = (value?: string | null) => {
  if (!value) return 'n/a';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString();
};

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

const isPendingStatus = (status?: string | null) =>
  status === 'queued' || status === 'in_progress';

export const MainStoryTemplates = () => {
  const { apiClient } = useAPI();
  const [templates, setTemplates] = useState<MainStoryTemplate[]>([]);
  const [districts, setDistricts] = useState<District[]>([]);
  const [runs, setRuns] = useState<MainStoryDistrictRun[]>([]);
  const [loading, setLoading] = useState(true);
  const [pageError, setPageError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [selectedDistrictByTemplate, setSelectedDistrictByTemplate] = useState<
    Record<string, string>
  >({});
  const [instantiatingTemplateId, setInstantiatingTemplateId] = useState<
    string | null
  >(null);

  const districtsById = useMemo(() => {
    const map = new Map<string, District>();
    districts.forEach((district) => {
      map.set(district.id, district);
    });
    return map;
  }, [districts]);

  const runsByTemplateId = useMemo(() => {
    const map = new Map<string, MainStoryDistrictRun[]>();
    runs.forEach((run) => {
      const existing = map.get(run.mainStoryTemplateId) ?? [];
      existing.push(run);
      map.set(run.mainStoryTemplateId, existing);
    });
    return map;
  }, [runs]);

  const loadPage = useCallback(async () => {
    setLoading(true);
    try {
      const [templatesResponse, districtsResponse, runsResponse] =
        await Promise.all([
          apiClient.get<MainStoryTemplate[]>('/sonar/mainStoryTemplates'),
          apiClient.get<District[]>('/sonar/districts'),
          apiClient.get<MainStoryDistrictRun[]>('/sonar/mainStoryDistrictRuns'),
        ]);
      setTemplates(templatesResponse);
      setDistricts(districtsResponse);
      setRuns(runsResponse);
      setPageError(null);
      setSelectedDistrictByTemplate((current) => {
        const next = { ...current };
        const fallbackDistrictId = districtsResponse[0]?.id ?? '';
        templatesResponse.forEach((template) => {
          if (!next[template.id]) {
            next[template.id] = fallbackDistrictId;
          }
        });
        return next;
      });
    } catch (error) {
      console.error('Failed to load main story templates page', error);
      setPageError(
        extractApiErrorMessage(
          error,
          'Failed to load main story templates and district runs.'
        )
      );
    } finally {
      setLoading(false);
    }
  }, [apiClient]);

  useEffect(() => {
    void loadPage();
  }, [loadPage]);

  useEffect(() => {
    const hasPendingRuns = runs.some((run) => isPendingStatus(run.status));
    if (!hasPendingRuns) {
      return;
    }
    const interval = window.setInterval(() => {
      void loadPage();
    }, 5000);
    return () => window.clearInterval(interval);
  }, [loadPage, runs]);

  const handleInstantiate = async (templateId: string) => {
    const districtId = selectedDistrictByTemplate[templateId];
    if (!districtId) {
      setActionError('Choose a district before instantiating a live chain.');
      return;
    }

    setInstantiatingTemplateId(templateId);
    setActionError(null);
    try {
      const created = await apiClient.post<MainStoryDistrictRun>(
        `/sonar/mainStoryTemplates/${templateId}/districtRuns`,
        { districtId }
      );
      setRuns((current) => [
        created,
        ...current.filter((run) => run.id !== created.id),
      ]);
      if (created.status === 'failed' && created.errorMessage) {
        setActionError(created.errorMessage);
      }
    } catch (error) {
      console.error('Failed to instantiate main story template', error);
      setActionError(
        extractApiErrorMessage(
          error,
          'Failed to instantiate that main story into the selected district.'
        )
      );
    } finally {
      setInstantiatingTemplateId(null);
    }
  };

  return (
    <div className="qa-theme">
      <div className="qa-shell">
        <section className="qa-hero">
          <div>
            <div className="qa-kicker">Main Story Templates</div>
            <h1 className="qa-title">Converted Campaign Templates</h1>
            <p className="qa-subtitle">
              Review converted campaign templates, inspect recent live district
              runs, and start building district-specific main-story quest
              chains from reusable story blueprints.
            </p>
          </div>
          <div className="qa-hero-actions">
            <Link to="/main-story-generator" className="qa-btn qa-btn-outline">
              Back to Generator
            </Link>
            <button
              type="button"
              className="qa-btn qa-btn-primary"
              onClick={() => void loadPage()}
              disabled={loading}
            >
              {loading ? 'Refreshing...' : 'Refresh'}
            </button>
          </div>
        </section>

        {pageError && <div className="qa-card">{pageError}</div>}
        {actionError && <div className="qa-card">{actionError}</div>}

        <section className="qa-card">
          <div className="qa-card-header">
            <div>
              <h2 className="qa-card-title">Live District Chains</h2>
              <div className="qa-meta">
                First pass: instantiation clones the recurring cast, places
                questgivers into district POIs, and creates a linked main-story
                quest chain for the selected district.
              </div>
            </div>
            <div className="qa-actions">
              <span className="qa-chip muted">
                {templates.length} templates available
              </span>
              <span className="qa-chip muted">{runs.length} total runs</span>
            </div>
          </div>
        </section>

        <section className="qa-grid">
          {loading ? (
            <div className="qa-card">Loading templates...</div>
          ) : templates.length === 0 ? (
            <div className="qa-card">
              No converted main story templates yet. Convert a draft from the
              generator first.
            </div>
          ) : (
            templates.map((template) => {
              const templateRuns = runsByTemplateId.get(template.id) ?? [];
              const selectedDistrictId =
                selectedDistrictByTemplate[template.id] ?? '';
              const selectedDistrict = districtsById.get(selectedDistrictId);
              return (
                <article className="qa-card" key={template.id}>
                  <div className="qa-card-header">
                    <div>
                      <h2 className="qa-card-title">{template.name}</h2>
                      <div className="qa-meta">
                        {template.premise || 'No premise yet.'}
                      </div>
                    </div>
                    <div className="qa-actions">
                      <span className="qa-chip accent">
                        {template.beats.length} beats
                      </span>
                      <span className="qa-chip muted">
                        Updated {formatDate(template.updatedAt)}
                      </span>
                    </div>
                  </div>

                  <div className="qa-stat-grid">
                    <div className="qa-stat">
                      <div className="qa-stat-label">Tone</div>
                      <div className="qa-stat-value">
                        {template.tone || 'Not specified'}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">District Fit</div>
                      <div className="qa-stat-value">
                        {template.districtFit || 'General fit'}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Theme Tags</div>
                      <div className="qa-stat-value">
                        {template.themeTags?.length
                          ? template.themeTags.join(', ')
                          : 'None'}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Recent Runs</div>
                      <div className="qa-stat-value">{templateRuns.length}</div>
                    </div>
                  </div>

                  <div className="qa-divider" />

                  <div className="qa-card-header">
                    <div>
                      <h3 className="qa-card-title" style={{ fontSize: 18 }}>
                        Instantiate in District
                      </h3>
                      <div className="qa-meta">
                        This creates live quests in that district and links them
                        into a main-story chain.
                      </div>
                    </div>
                    <div className="qa-actions">
                      <select
                        value={selectedDistrictId}
                        onChange={(event) =>
                          setSelectedDistrictByTemplate((current) => ({
                            ...current,
                            [template.id]: event.target.value,
                          }))
                        }
                        style={{
                          minWidth: 240,
                          borderRadius: 999,
                          padding: '8px 14px',
                          background: 'rgba(255,255,255,0.08)',
                          border: '1px solid rgba(255,255,255,0.16)',
                          color: 'var(--qa-ink)',
                        }}
                      >
                        <option value="">
                          {districts.length
                            ? 'Choose a district'
                            : 'No districts available'}
                        </option>
                        {districts.map((district) => (
                          <option key={district.id} value={district.id}>
                            {district.name}
                          </option>
                        ))}
                      </select>
                      <button
                        type="button"
                        className="qa-btn qa-btn-primary"
                        onClick={() => void handleInstantiate(template.id)}
                        disabled={
                          !selectedDistrictId ||
                          instantiatingTemplateId === template.id
                        }
                      >
                        {instantiatingTemplateId === template.id
                          ? 'Instantiating...'
                          : selectedDistrict
                            ? `Instantiate in ${selectedDistrict.name}`
                            : 'Instantiate'}
                      </button>
                    </div>
                  </div>

                  <div className="qa-divider" />

                  <div className="qa-card-header">
                    <div>
                      <h3 className="qa-card-title" style={{ fontSize: 18 }}>
                        Beat Outline
                      </h3>
                      <div className="qa-meta">
                        The live chain will follow this order.
                      </div>
                    </div>
                  </div>
                  <div className="qa-tree">
                    {template.beats.map((beat) => (
                      <div className="qa-node" key={`${template.id}-${beat.orderIndex}`}>
                        <div className="qa-node-card">
                          <div className="qa-card-header">
                            <div>
                              <div className="qa-node-title">
                                Beat {beat.orderIndex}: {beat.chapterTitle}
                              </div>
                              <div className="qa-meta">
                                {beat.chapterSummary || beat.name}
                              </div>
                            </div>
                            <div className="qa-actions">
                              <span className="qa-chip muted">
                                Act {beat.act}
                              </span>
                              <span className="qa-chip muted">
                                {beat.storyRole || 'story beat'}
                              </span>
                              {beat.questGiverCharacterName ? (
                                <span className="qa-chip accent">
                                  {beat.questGiverCharacterName}
                                </span>
                              ) : null}
                            </div>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>

                  <div className="qa-divider" />

                  <div className="qa-card-header">
                    <div>
                      <h3 className="qa-card-title" style={{ fontSize: 18 }}>
                        District Runs
                      </h3>
                      <div className="qa-meta">
                        Recent attempts to turn this template into live district
                        content.
                      </div>
                    </div>
                  </div>

                  {templateRuns.length === 0 ? (
                    <div className="qa-meta">
                      No district runs yet for this template.
                    </div>
                  ) : (
                    <div className="qa-grid">
                      {templateRuns.map((run) => {
                        const district = districtsById.get(run.districtId);
                        return (
                          <div className="qa-node-card" key={run.id}>
                            <div className="qa-card-header">
                              <div>
                                <div className="qa-node-title">
                                  {district?.name ?? 'Unknown district'}
                                </div>
                                <div className="qa-meta">
                                  Started {formatDate(run.createdAt)} • Updated{' '}
                                  {formatDate(run.updatedAt)}
                                </div>
                              </div>
                              <div className="qa-actions">
                                <span className={statusChipClass(run.status)}>
                                  {run.status.replace(/_/g, ' ')}
                                </span>
                                <span className="qa-chip muted">
                                  {run.beatRuns?.filter(
                                    (beat) => beat.status === 'completed'
                                  ).length ?? 0}
                                  /{template.beats.length} beats
                                </span>
                              </div>
                            </div>
                            {run.errorMessage ? (
                              <div className="qa-meta" style={{ color: 'var(--qa-danger)' }}>
                                {run.errorMessage}
                              </div>
                            ) : null}
                            <div className="qa-tree" style={{ marginTop: 14 }}>
                              {(run.beatRuns || []).map((beatRun) => (
                                <div className="qa-node" key={`${run.id}-${beatRun.orderIndex}`}>
                                  <div className="qa-node-card">
                                    <div className="qa-card-header">
                                      <div>
                                        <div className="qa-node-title">
                                          Beat {beatRun.orderIndex}:{' '}
                                          {beatRun.chapterTitle}
                                        </div>
                                        <div className="qa-meta">
                                          {[beatRun.zoneName, beatRun.pointOfInterestName]
                                            .filter(Boolean)
                                            .join(' • ') || 'Placement pending'}
                                        </div>
                                      </div>
                                      <div className="qa-actions">
                                        <span
                                          className={statusChipClass(
                                            beatRun.status
                                          )}
                                        >
                                          {beatRun.status.replace(/_/g, ' ')}
                                        </span>
                                      </div>
                                    </div>
                                    {beatRun.questName ? (
                                      <div className="qa-meta">
                                        Quest: {beatRun.questName}
                                      </div>
                                    ) : null}
                                    {beatRun.errorMessage ? (
                                      <div
                                        className="qa-meta"
                                        style={{ color: 'var(--qa-danger)' }}
                                      >
                                        {beatRun.errorMessage}
                                      </div>
                                    ) : null}
                                  </div>
                                </div>
                              ))}
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </article>
              );
            })
          )}
        </section>
      </div>
    </div>
  );
};

export default MainStoryTemplates;
