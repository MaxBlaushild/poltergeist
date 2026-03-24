import React, { useCallback, useEffect, useState } from 'react';
import { useQuestArchetypes } from '../contexts/questArchetypes.tsx';
import { useAPI, useZoneContext } from '@poltergeist/contexts';
import { Character, QuestGenerationJob } from '@poltergeist/types';
import './questArchetypeTheme.css';

export const ZoneQuestArchetypes = () => {
  const { zones } = useZoneContext();
  const { apiClient } = useAPI();
  const {
    zoneQuestArchetypes,
    createZoneQuestArchetype,
    deleteZoneQuestArchetype,
    questArchetypes,
  } = useQuestArchetypes();
  const [shouldShowModal, setShouldShowModal] = useState(false);
  const [zoneSearch, setZoneSearch] = useState('');
  const [questArchetypeSearch, setQuestArchetypeSearch] = useState('');
  const [characterSearch, setCharacterSearch] = useState('');
  const [selectedZoneId, setSelectedZoneId] = useState('');
  const [selectedQuestArchetypeId, setSelectedQuestArchetypeId] = useState('');
  const [numberOfQuests, setNumberOfQuests] = useState(1);
  const [characters, setCharacters] = useState<Character[]>([]);
  const [selectedCharacterId, setSelectedCharacterId] = useState('');
  const [generationJobsByArchetype, setGenerationJobsByArchetype] = useState<
    Record<string, QuestGenerationJob[]>
  >({});
  const [generationErrors, setGenerationErrors] = useState<
    Record<string, string | null>
  >({});
  const [generationLoading, setGenerationLoading] = useState<
    Record<string, boolean>
  >({});
  const [generating, setGenerating] = useState<Record<string, boolean>>({});
  const [generationPolling, setGenerationPolling] = useState(false);

  useEffect(() => {
    const fetchCharacters = async () => {
      try {
        const response = await apiClient.get<Character[]>('/sonar/characters');
        setCharacters(response);
      } catch (error) {
        console.error('Error fetching characters:', error);
      }
    };
    fetchCharacters();
  }, [apiClient]);

  const isPendingStatus = useCallback(
    (status: string) => status === 'queued' || status === 'in_progress',
    []
  );

  const statusChipClass = (status: string) => {
    switch (status) {
      case 'completed':
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

  const formatStatus = (status: string) => status.replace(/_/g, ' ');

  const formatTimestamp = (value: string) => {
    if (!value) return '';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return value;
    }
    return date.toLocaleString();
  };

  const updatePollingState = useCallback(
    (jobsMap: Record<string, QuestGenerationJob[]>) => {
      const hasPending = Object.values(jobsMap).some((jobs) =>
        jobs.some((job) => isPendingStatus(job.status))
      );
      setGenerationPolling(hasPending);
    },
    [isPendingStatus]
  );

  const fetchGenerationJobsForArchetype = useCallback(
    async (zoneQuestArchetypeId: string) => {
      setGenerationLoading((prev) => ({
        ...prev,
        [zoneQuestArchetypeId]: true,
      }));
      try {
        const jobs = await apiClient.get<QuestGenerationJob[]>(
          `/sonar/zoneQuestArchetypes/${zoneQuestArchetypeId}/questGenerations`
        );
        setGenerationJobsByArchetype((prev) => {
          const next = { ...prev, [zoneQuestArchetypeId]: jobs };
          updatePollingState(next);
          return next;
        });
        setGenerationErrors((prev) => ({
          ...prev,
          [zoneQuestArchetypeId]: null,
        }));
      } catch (error) {
        console.error('Failed to fetch quest generation jobs', error);
        setGenerationErrors((prev) => ({
          ...prev,
          [zoneQuestArchetypeId]: 'Failed to load generation status.',
        }));
      } finally {
        setGenerationLoading((prev) => ({
          ...prev,
          [zoneQuestArchetypeId]: false,
        }));
      }
    },
    [apiClient, updatePollingState]
  );

  useEffect(() => {
    if (!zoneQuestArchetypes || zoneQuestArchetypes.length === 0) {
      setGenerationJobsByArchetype({});
      setGenerationPolling(false);
      return;
    }
    zoneQuestArchetypes.forEach((zoneQuestArchetype) => {
      fetchGenerationJobsForArchetype(zoneQuestArchetype.id);
    });
  }, [zoneQuestArchetypes, fetchGenerationJobsForArchetype]);

  useEffect(() => {
    if (
      !generationPolling ||
      !zoneQuestArchetypes ||
      zoneQuestArchetypes.length === 0
    ) {
      return;
    }
    const interval = setInterval(() => {
      zoneQuestArchetypes.forEach((zoneQuestArchetype) => {
        fetchGenerationJobsForArchetype(zoneQuestArchetype.id);
      });
    }, 4000);
    return () => clearInterval(interval);
  }, [generationPolling, zoneQuestArchetypes, fetchGenerationJobsForArchetype]);

  const handleGenerateQuests = async (zoneQuestArchetypeId: string) => {
    setGenerating((prev) => ({ ...prev, [zoneQuestArchetypeId]: true }));
    setGenerationErrors((prev) => ({ ...prev, [zoneQuestArchetypeId]: null }));
    try {
      const job = await apiClient.post<QuestGenerationJob>(
        `/sonar/zoneQuestArchetypes/${zoneQuestArchetypeId}/generate`,
        {}
      );
      setGenerationJobsByArchetype((prev) => {
        const existing = prev[zoneQuestArchetypeId] ?? [];
        const next = { ...prev, [zoneQuestArchetypeId]: [job, ...existing] };
        updatePollingState(next);
        return next;
      });
    } catch (error) {
      console.error('Failed to generate quests', error);
      setGenerationErrors((prev) => ({
        ...prev,
        [zoneQuestArchetypeId]: 'Failed to queue quest generation.',
      }));
    } finally {
      setGenerating((prev) => ({ ...prev, [zoneQuestArchetypeId]: false }));
    }
  };

  return (
    <div className="qa-theme">
      <div className="qa-shell">
        <header className="qa-hero">
          <div>
            <div className="qa-kicker">Zone Operations</div>
            <h1 className="qa-title">Zone Quest Archetypes</h1>
            <p className="qa-subtitle">
              Bind archetypes to specific zones, set quest volume targets, and
              assign quest givers so each area feels distinct.
            </p>
          </div>
          <div className="qa-hero-actions">
            <button
              className="qa-btn qa-btn-primary"
              onClick={() => setShouldShowModal(true)}
            >
              New Zone Assignment
            </button>
          </div>
        </header>

        <section className="qa-grid">
          {zoneQuestArchetypes?.length === 0 ? (
            <div className="qa-panel">
              <div className="qa-card-title">No zone assignments yet</div>
              <p className="qa-muted" style={{ marginTop: 8 }}>
                Assign an archetype to a zone to start generating quests.
              </p>
            </div>
          ) : (
            zoneQuestArchetypes.map((zoneQuestArchetype, index) => (
              <article
                key={zoneQuestArchetype.id}
                className="qa-card"
                style={{ animationDelay: `${index * 0.06}s` }}
              >
                <div className="qa-card-header">
                  <div>
                    <h2 className="qa-card-title">
                      {zoneQuestArchetype.questArchetype.name}
                    </h2>
                    <div className="qa-meta">
                      Zone: {zoneQuestArchetype.zone.name}
                    </div>
                  </div>
                  <div className="qa-actions">
                    <button
                      onClick={() =>
                        deleteZoneQuestArchetype(zoneQuestArchetype.id)
                      }
                      className="qa-btn qa-btn-danger"
                    >
                      Delete
                    </button>
                  </div>
                </div>

                <div className="qa-stat-grid">
                  <div className="qa-stat">
                    <div className="qa-stat-label">Quests to Generate</div>
                    <div className="qa-stat-value">
                      {zoneQuestArchetype.numberOfQuests}
                    </div>
                  </div>
                  <div className="qa-stat">
                    <div className="qa-stat-label">Quest Giver</div>
                    <div className="qa-stat-value">
                      {zoneQuestArchetype.character?.name ??
                        ((zoneQuestArchetype.questArchetype.characterTags
                          ?.length ?? 0) > 0
                          ? `Auto: ${zoneQuestArchetype.questArchetype.characterTags?.join(', ')}`
                          : 'None')}
                    </div>
                  </div>
                  <div className="qa-stat">
                    <div className="qa-stat-label">Archetype ID</div>
                    <div className="qa-stat-value">
                      {zoneQuestArchetype.questArchetypeId.slice(0, 8)}…
                    </div>
                  </div>
                </div>

                <div className="qa-divider" />

                <div>
                  <div className="qa-card-title" style={{ fontSize: 16 }}>
                    Quest Generation
                  </div>
                  <p className="qa-muted" style={{ marginTop: 6 }}>
                    Queue a fresh batch of quests and track completion progress
                    in real time.
                  </p>
                  <div className="qa-actions" style={{ marginTop: 12 }}>
                    <button
                      className="qa-btn qa-btn-primary"
                      onClick={() =>
                        handleGenerateQuests(zoneQuestArchetype.id)
                      }
                      disabled={generating[zoneQuestArchetype.id]}
                    >
                      {generating[zoneQuestArchetype.id]
                        ? 'Queueing...'
                        : `Generate ${zoneQuestArchetype.numberOfQuests} Quests`}
                    </button>
                    <button
                      className="qa-btn qa-btn-outline"
                      onClick={() =>
                        fetchGenerationJobsForArchetype(zoneQuestArchetype.id)
                      }
                      disabled={generationLoading[zoneQuestArchetype.id]}
                    >
                      {generationLoading[zoneQuestArchetype.id]
                        ? 'Refreshing...'
                        : 'Refresh Status'}
                    </button>
                  </div>

                  {generationErrors[zoneQuestArchetype.id] && (
                    <div className="qa-chip danger" style={{ marginTop: 12 }}>
                      {generationErrors[zoneQuestArchetype.id]}
                    </div>
                  )}

                  {(generationJobsByArchetype[zoneQuestArchetype.id]?.length ??
                    0) === 0 ? (
                    <div className="qa-panel" style={{ marginTop: 14 }}>
                      <div className="qa-muted">
                        No quest generation jobs yet.
                      </div>
                    </div>
                  ) : (
                    <div className="qa-tree" style={{ marginTop: 14 }}>
                      {(
                        generationJobsByArchetype[zoneQuestArchetype.id] ?? []
                      ).map((job) => (
                        <div key={job.id} className="qa-node-card">
                          <div
                            className="qa-card-header"
                            style={{ marginBottom: 8 }}
                          >
                            <div>
                              <div
                                className="qa-card-title"
                                style={{ fontSize: 16 }}
                              >
                                Job {job.id.slice(0, 6)}…
                              </div>
                              <div className="qa-meta">
                                Started {formatTimestamp(job.createdAt)}
                              </div>
                            </div>
                            <div className={statusChipClass(job.status)}>
                              {formatStatus(job.status)}
                            </div>
                          </div>
                          <div className="qa-meta">
                            Progress: {job.completedCount}/{job.totalCount} ·
                            Failed: {job.failedCount}
                          </div>
                          {job.errorMessage && (
                            <div
                              className="qa-chip danger"
                              style={{ marginTop: 8 }}
                            >
                              {job.errorMessage}
                            </div>
                          )}
                          {job.quests && job.quests.length > 0 ? (
                            <div style={{ marginTop: 12 }}>
                              <div className="qa-stat-label">
                                Generated Quests
                              </div>
                              <div className="qa-tree" style={{ marginTop: 8 }}>
                                {job.quests.map((quest) => (
                                  <div key={quest.id} className="qa-node-card">
                                    <div className="qa-node-title">
                                      {quest.name || 'Untitled Quest'}
                                    </div>
                                    <div className="qa-meta">
                                      ID: {quest.id.slice(0, 8)}…
                                    </div>
                                  </div>
                                ))}
                              </div>
                            </div>
                          ) : job.questIds && job.questIds.length > 0 ? (
                            <div style={{ marginTop: 12 }}>
                              <div className="qa-stat-label">
                                Generated Quest IDs
                              </div>
                              <div className="qa-tree" style={{ marginTop: 8 }}>
                                {job.questIds.map((questId) => (
                                  <div key={questId} className="qa-node-card">
                                    <div className="qa-node-title">
                                      Quest {questId.slice(0, 8)}…
                                    </div>
                                    <div className="qa-meta">ID: {questId}</div>
                                  </div>
                                ))}
                              </div>
                            </div>
                          ) : null}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </article>
            ))
          )}
        </section>
      </div>

      {shouldShowModal && (
        <div className="qa-modal">
          <div className="qa-modal-card">
            <h2 className="qa-modal-title">Create Zone Quest Archetype</h2>

            <div className="qa-form-grid">
              <div className="qa-field">
                <div className="qa-label">Zone Search</div>
                <input
                  type="text"
                  className="qa-input"
                  value={zoneSearch}
                  onChange={(e) => setZoneSearch(e.target.value)}
                  placeholder="Search zones..."
                />
              </div>

              <div className="qa-field">
                <div className="qa-label">Zone</div>
                <select
                  className="qa-select"
                  value={selectedZoneId}
                  onChange={(e) => setSelectedZoneId(e.target.value)}
                >
                  <option value="">Select a zone</option>
                  {zones
                    .filter((z) =>
                      z.name.toLowerCase().includes(zoneSearch.toLowerCase())
                    )
                    .map((z) => (
                      <option key={z.id} value={z.id}>
                        {z.name}
                      </option>
                    ))}
                </select>
              </div>

              <div className="qa-field">
                <div className="qa-label">Quest Archetype Search</div>
                <input
                  type="text"
                  className="qa-input"
                  value={questArchetypeSearch}
                  onChange={(e) => setQuestArchetypeSearch(e.target.value)}
                  placeholder="Search quest archetypes..."
                />
              </div>

              <div className="qa-field">
                <div className="qa-label">Quest Archetype</div>
                <select
                  className="qa-select"
                  value={selectedQuestArchetypeId}
                  onChange={(e) => setSelectedQuestArchetypeId(e.target.value)}
                >
                  <option value="">Select a quest archetype</option>
                  {questArchetypes
                    .filter((qa) =>
                      qa.name
                        .toLowerCase()
                        .includes(questArchetypeSearch.toLowerCase())
                    )
                    .map((qa) => (
                      <option key={qa.id} value={qa.id}>
                        {qa.name}
                      </option>
                    ))}
                </select>
              </div>

              <div className="qa-field">
                <div className="qa-label">Number of Quests</div>
                <input
                  type="number"
                  className="qa-input"
                  value={numberOfQuests}
                  onChange={(e) =>
                    setNumberOfQuests(parseInt(e.target.value) || 1)
                  }
                  min="1"
                />
              </div>

              <div className="qa-field">
                <div className="qa-label">Character Search</div>
                <input
                  type="text"
                  className="qa-input"
                  value={characterSearch}
                  onChange={(e) => setCharacterSearch(e.target.value)}
                  placeholder="Search characters..."
                />
              </div>

              <div className="qa-field">
                <div className="qa-label">Quest Giver Character</div>
                <select
                  className="qa-select"
                  value={selectedCharacterId}
                  onChange={(e) => setSelectedCharacterId(e.target.value)}
                >
                  <option value="">Auto-match from template tags</option>
                  {characters
                    .filter((character) =>
                      character.name
                        .toLowerCase()
                        .includes(characterSearch.toLowerCase())
                    )
                    .map((character) => (
                      <option key={character.id} value={character.id}>
                        {character.name}
                      </option>
                    ))}
                </select>
              </div>

              <div className="qa-footer">
                <button
                  className="qa-btn qa-btn-outline"
                  onClick={() => setShouldShowModal(false)}
                >
                  Cancel
                </button>
                <button
                  className="qa-btn qa-btn-primary"
                  onClick={async () => {
                    if (
                      selectedZoneId &&
                      selectedQuestArchetypeId &&
                      numberOfQuests
                    ) {
                      await createZoneQuestArchetype(
                        selectedZoneId,
                        selectedQuestArchetypeId,
                        numberOfQuests,
                        selectedCharacterId || null
                      );
                      setShouldShowModal(false);
                    }
                  }}
                  disabled={
                    !selectedZoneId ||
                    !selectedQuestArchetypeId ||
                    !numberOfQuests
                  }
                >
                  Create
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
