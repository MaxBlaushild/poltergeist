import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
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

export const MainStoryDistrictRunView = () => {
  const { apiClient } = useAPI();
  const navigate = useNavigate();
  const { id } = useParams();
  const [run, setRun] = useState<MainStoryDistrictRun | null>(null);
  const [template, setTemplate] = useState<MainStoryTemplate | null>(null);
  const [district, setDistrict] = useState<District | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [retrying, setRetrying] = useState(false);

  const loadRun = useCallback(async () => {
    if (!id) {
      setError('Missing district run ID.');
      setLoading(false);
      return;
    }

    setLoading(true);
    try {
      const runResponse = await apiClient.get<MainStoryDistrictRun>(
        `/sonar/mainStoryDistrictRuns/${id}`
      );
      setRun(runResponse);

      const [templateResponse, districtsResponse] = await Promise.all([
        apiClient.get<MainStoryTemplate>(
          `/sonar/mainStoryTemplates/${runResponse.mainStoryTemplateId}`
        ),
        apiClient.get<District[]>('/sonar/districts'),
      ]);

      setTemplate(templateResponse);
      setDistrict(
        districtsResponse.find(
          (candidate) => candidate.id === runResponse.districtId
        ) ?? null
      );
      setError(null);
    } catch (loadError) {
      console.error('Failed to load main story district run', loadError);
      setError(
        extractApiErrorMessage(
          loadError,
          'Failed to load that main story district run.'
        )
      );
    } finally {
      setLoading(false);
    }
  }, [apiClient, id]);

  useEffect(() => {
    void loadRun();
  }, [loadRun]);

  useEffect(() => {
    if (!run || (run.status !== 'queued' && run.status !== 'in_progress')) {
      return;
    }
    const interval = window.setInterval(() => {
      void loadRun();
    }, 5000);
    return () => window.clearInterval(interval);
  }, [loadRun, run]);

  const completedBeatCount = useMemo(
    () =>
      run?.beatRuns?.filter((beat) => beat.status === 'completed').length ?? 0,
    [run]
  );

  const handleDeleteRun = async () => {
    if (!run) {
      return;
    }
    const confirmed = window.confirm(
      'Delete this live district chain and all quests plus generated questgiver characters from this run?'
    );
    if (!confirmed) {
      return;
    }

    setDeleting(true);
    try {
      await apiClient.delete(`/sonar/mainStoryDistrictRuns/${run.id}`);
      navigate('/main-story-templates');
    } catch (deleteError) {
      console.error('Failed to delete main story district run', deleteError);
      setError(
        extractApiErrorMessage(
          deleteError,
          'Failed to delete this main story district run.'
        )
      );
    } finally {
      setDeleting(false);
    }
  };

  const handleRetryRun = async () => {
    if (!run) {
      return;
    }
    const confirmed = window.confirm(
      'Retry this run from its first failed beat onward? Completed earlier beats will be kept, and the failed beat will be re-attempted in a different zone first when possible.'
    );
    if (!confirmed) {
      return;
    }

    setRetrying(true);
    try {
      const retried = await apiClient.post<MainStoryDistrictRun>(
        `/sonar/mainStoryDistrictRuns/${run.id}/retry`,
        {}
      );
      setRun(retried);
      setError(
        retried.status === 'failed' ? retried.errorMessage ?? null : null
      );
    } catch (retryError) {
      console.error('Failed to retry main story district run', retryError);
      setError(
        extractApiErrorMessage(
          retryError,
          'Failed to retry this main story district run.'
        )
      );
    } finally {
      setRetrying(false);
    }
  };

  return (
    <div className="qa-theme">
      <div className="qa-shell">
        <section className="qa-hero">
          <div>
            <div className="qa-kicker">District Run</div>
            <h1 className="qa-title">
              {template?.name || 'Main Story District Chain'}
            </h1>
            <p className="qa-subtitle">
              Inspect the full live district chain generated from this template,
              including beat placement, questgiver assignments, and created
              quests. If something looks off, you can delete the entire run from
              here and try again.
            </p>
          </div>
          <div className="qa-hero-actions">
            <Link to="/main-story-templates" className="qa-btn qa-btn-outline">
              Back to Templates
            </Link>
            {run?.status === 'failed' ? (
              <button
                type="button"
                className="qa-btn qa-btn-primary"
                onClick={() => void handleRetryRun()}
                disabled={retrying || deleting || !run}
              >
                {retrying ? 'Retrying...' : 'Retry Run'}
              </button>
            ) : null}
            <button
              type="button"
              className="qa-btn qa-btn-danger"
              onClick={() => void handleDeleteRun()}
              disabled={deleting || retrying || !run}
            >
              {deleting ? 'Cleaning Up...' : 'Clean Up Run'}
            </button>
          </div>
        </section>

        {error && <div className="qa-card">{error}</div>}

        {loading ? (
          <div className="qa-card">Loading district run...</div>
        ) : !run ? (
          <div className="qa-card">District run not found.</div>
        ) : (
          <>
            <section className="qa-card">
              <div className="qa-card-header">
                <div>
                  <h2 className="qa-card-title">Run Summary</h2>
                  <div className="qa-meta">
                    {district?.name ?? 'Unknown district'} • Started{' '}
                    {formatDate(run.createdAt)} • Updated{' '}
                    {formatDate(run.updatedAt)}
                  </div>
                </div>
                <div className="qa-actions">
                  <span className={statusChipClass(run.status)}>
                    {run.status.replace(/_/g, ' ')}
                  </span>
                </div>
              </div>
              <div className="qa-stat-grid">
                <div className="qa-stat">
                  <div className="qa-stat-label">Completed Beats</div>
                  <div className="qa-stat-value">
                    {completedBeatCount}/{run.beatRuns.length}
                  </div>
                </div>
                <div className="qa-stat">
                  <div className="qa-stat-label">Generated Questgivers</div>
                  <div className="qa-stat-value">
                    {run.generatedCharacterIds?.length ?? 0}
                  </div>
                </div>
                <div className="qa-stat">
                  <div className="qa-stat-label">District</div>
                  <div className="qa-stat-value">
                    {district?.name ?? run.districtId}
                  </div>
                </div>
                <div className="qa-stat">
                  <div className="qa-stat-label">Template</div>
                  <div className="qa-stat-value">
                    {template?.name ?? run.mainStoryTemplateId}
                  </div>
                </div>
              </div>
              {run.errorMessage ? (
                <div className="qa-meta" style={{ color: 'var(--qa-danger)', marginTop: 16 }}>
                  {run.errorMessage}
                </div>
              ) : null}
            </section>

            <section className="qa-card">
              <div className="qa-card-header">
                <div>
                  <h2 className="qa-card-title">Beat Chain</h2>
                  <div className="qa-meta">
                    The full linked chain of live quests created from this run.
                  </div>
                </div>
              </div>
              <div className="qa-tree">
                {run.beatRuns.map((beatRun) => (
                  <div className="qa-node" key={`${run.id}-${beatRun.orderIndex}`}>
                    <div className="qa-node-card">
                      <div className="qa-card-header">
                        <div>
                          <div className="qa-node-title">
                            Beat {beatRun.orderIndex}: {beatRun.chapterTitle}
                          </div>
                          <div className="qa-meta">
                            {[
                              `Act ${beatRun.act}`,
                              beatRun.storyRole,
                              beatRun.zoneName,
                              beatRun.pointOfInterestName,
                            ]
                              .filter(Boolean)
                              .join(' • ')}
                          </div>
                        </div>
                        <div className="qa-actions">
                          <span className={statusChipClass(beatRun.status)}>
                            {beatRun.status.replace(/_/g, ' ')}
                          </span>
                        </div>
                      </div>

                      <div className="qa-stat-grid" style={{ marginTop: 14 }}>
                        <div className="qa-stat">
                          <div className="qa-stat-label">Questgiver</div>
                          <div className="qa-stat-value">
                            {beatRun.questGiverCharacterName || 'None'}
                          </div>
                        </div>
                        <div className="qa-stat">
                          <div className="qa-stat-label">Quest</div>
                          <div className="qa-stat-value">
                            {beatRun.questName || 'Not created'}
                          </div>
                        </div>
                        <div className="qa-stat">
                          <div className="qa-stat-label">POI</div>
                          <div className="qa-stat-value">
                            {beatRun.pointOfInterestName || 'Not placed'}
                          </div>
                        </div>
                        <div className="qa-stat">
                          <div className="qa-stat-label">Zone</div>
                          <div className="qa-stat-value">
                            {beatRun.zoneName || 'Not resolved'}
                          </div>
                        </div>
                      </div>

                      <div className="qa-actions" style={{ marginTop: 14 }}>
                        {beatRun.questId ? (
                          <Link
                            to={`/quests?id=${beatRun.questId}`}
                            className="qa-btn qa-btn-outline"
                          >
                            Open Quest
                          </Link>
                        ) : null}
                        {beatRun.questArchetypeId ? (
                          <Link
                            to={`/quest-archetypes?id=${beatRun.questArchetypeId}`}
                            className="qa-btn qa-btn-ghost"
                          >
                            Open Archetype
                          </Link>
                        ) : null}
                      </div>

                      {beatRun.errorMessage ? (
                        <div
                          className="qa-meta"
                          style={{ color: 'var(--qa-danger)', marginTop: 12 }}
                        >
                          {beatRun.errorMessage}
                        </div>
                      ) : null}
                    </div>
                  </div>
                ))}
              </div>
            </section>
          </>
        )}
      </div>
    </div>
  );
};

export default MainStoryDistrictRunView;
