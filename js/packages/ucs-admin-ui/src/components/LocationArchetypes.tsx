import React, { useEffect, useMemo, useState } from 'react';
import { useCandidates } from '@poltergeist/hooks';
import {
  Candidate,
  LocationArchetype,
  LocationArchetypeChallenge,
  QuestNodeSubmissionType,
} from '@poltergeist/types';
import { useAPI } from '@poltergeist/contexts';
import { useQuestArchetypes } from '../contexts/questArchetypes.tsx';
import './questArchetypeTheme.css';

const buildEmptyArchetype = (): LocationArchetype => ({
  id: '',
  name: '',
  includedTypes: [],
  excludedTypes: [],
  challenges: [],
  createdAt: new Date(),
  updatedAt: new Date(),
});

const clampPreview = (items: string[], limit: number) => {
  if (items.length <= limit) {
    return { preview: items, remaining: 0 };
  }
  return { preview: items.slice(0, limit), remaining: items.length - limit };
};

const mergeUnique = (list: string[], values: string[]) => {
  const next = [...list];
  values.forEach((value) => {
    if (!next.includes(value)) {
      next.push(value);
    }
  });
  return next;
};

type PlaceTypeImporterProps = {
  label: string;
  onApplyTypes: (types: string[]) => void;
};

const PlaceTypeImporter: React.FC<PlaceTypeImporterProps> = ({
  label,
  onApplyTypes,
}) => {
  const [query, setQuery] = useState('');
  const [selectedCandidate, setSelectedCandidate] = useState<Candidate | null>(
    null
  );
  const trimmedQuery = query.trim();
  const { candidates, loading } = useCandidates(trimmedQuery);

  useEffect(() => {
    if (!trimmedQuery) {
      setSelectedCandidate(null);
    }
  }, [trimmedQuery]);

  const selectedTypes = selectedCandidate?.types ?? [];
  const preview = clampPreview(selectedTypes, 6);

  return (
    <div className="qa-panel" style={{ marginTop: 12 }}>
      <div className="qa-meta">{label}</div>
      <div className="qa-combobox" style={{ marginTop: 8 }}>
        <input
          type="text"
          className="qa-input"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Search Google Maps..."
        />
        {trimmedQuery.length > 0 && (
          <div className="qa-combobox-list">
            {loading && <div className="qa-combobox-empty">Searching...</div>}
            {!loading && candidates.length === 0 && (
              <div className="qa-combobox-empty">No matches.</div>
            )}
            {!loading &&
              candidates.map((candidate) => (
                <button
                  key={candidate.place_id}
                  type="button"
                  className={`qa-combobox-option ${
                    selectedCandidate?.place_id === candidate.place_id
                      ? 'is-active'
                      : ''
                  }`}
                  onClick={() => setSelectedCandidate(candidate)}
                >
                  <div className="qa-option-title">{candidate.name}</div>
                  <div className="qa-option-sub">
                    {candidate.formatted_address}
                  </div>
                </button>
              ))}
          </div>
        )}
      </div>

      {selectedCandidate && (
        <div className="qa-panel" style={{ marginTop: 12 }}>
          <div className="qa-meta">Selected place</div>
          <div className="qa-option-title" style={{ marginTop: 6 }}>
            {selectedCandidate.name}
          </div>
          <div className="qa-option-sub">
            {selectedCandidate.formatted_address}
          </div>
          <div className="qa-inline" style={{ marginTop: 10 }}>
            {selectedTypes.length === 0 ? (
              <span className="qa-empty">
                No place types found on this result.
              </span>
            ) : (
              preview.preview.map((type) => (
                <span key={type} className="qa-chip accent">
                  {type}
                </span>
              ))
            )}
            {preview.remaining > 0 && (
              <span className="qa-chip muted">+{preview.remaining} more</span>
            )}
          </div>
          <div className="qa-inline" style={{ marginTop: 12 }}>
            <button
              type="button"
              className="qa-btn qa-btn-ghost"
              disabled={selectedTypes.length === 0}
              onClick={() => {
                onApplyTypes(selectedTypes);
                setQuery('');
                setSelectedCandidate(null);
              }}
            >
              Add {selectedTypes.length} type
              {selectedTypes.length === 1 ? '' : 's'}
            </button>
            <button
              type="button"
              className="qa-btn qa-btn-outline"
              onClick={() => setSelectedCandidate(null)}
            >
              Clear
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

type GeneratedLocationChallenge = {
  id: string;
  question: string;
  submissionType: QuestNodeSubmissionType;
  proficiency?: string | null;
  difficulty?: number | null;
};

const LocationArchetypes: React.FC = () => {
  const {
    locationArchetypes,
    createLocationArchetype,
    updateLocationArchetype,
    deleteLocationArchetype,
    placeTypes,
    refreshLocationArchetypes,
  } = useQuestArchetypes();
  const { apiClient } = useAPI();

  const submissionTypeOptions = [
    { value: 'photo', label: 'Photo' },
    { value: 'text', label: 'Text' },
    { value: 'video', label: 'Video' },
  ];

  const [searchQuery, setSearchQuery] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [createDraft, setCreateDraft] = useState<LocationArchetype>(
    buildEmptyArchetype()
  );
  const [createIncludedQuery, setCreateIncludedQuery] = useState('');
  const [createExcludedQuery, setCreateExcludedQuery] = useState('');
  const [createChallengeQuery, setCreateChallengeQuery] = useState('');
  const [createChallengeInputType, setCreateChallengeInputType] =
    useState<QuestNodeSubmissionType>('photo');
  const [createChallengeProficiency, setCreateChallengeProficiency] =
    useState('');
  const [createChallengeDifficulty, setCreateChallengeDifficulty] =
    useState<number>(0);
  const [createGeneratedChallenges, setCreateGeneratedChallenges] = useState<
    GeneratedLocationChallenge[]
  >([]);
  const [createGenerating, setCreateGenerating] = useState(false);
  const [createGenerateError, setCreateGenerateError] = useState<string | null>(
    null
  );
  const [queueingBatchGeneration, setQueueingBatchGeneration] = useState(false);
  const [batchGenerationMessage, setBatchGenerationMessage] = useState<
    string | null
  >(null);
  const [batchGenerationError, setBatchGenerationError] = useState<
    string | null
  >(null);

  const [editDraft, setEditDraft] = useState<LocationArchetype | null>(null);
  const [editIncludedQuery, setEditIncludedQuery] = useState('');
  const [editExcludedQuery, setEditExcludedQuery] = useState('');
  const [editChallengeQuery, setEditChallengeQuery] = useState('');
  const [editChallengeInputType, setEditChallengeInputType] =
    useState<QuestNodeSubmissionType>('photo');
  const [editChallengeProficiency, setEditChallengeProficiency] = useState('');
  const [editChallengeDifficulty, setEditChallengeDifficulty] =
    useState<number>(0);
  const [editGeneratedChallenges, setEditGeneratedChallenges] = useState<
    GeneratedLocationChallenge[]
  >([]);
  const [editGenerating, setEditGenerating] = useState(false);
  const [editGenerateError, setEditGenerateError] = useState<string | null>(
    null
  );

  const queueBatchGeneration = async () => {
    setQueueingBatchGeneration(true);
    setBatchGenerationMessage(null);
    setBatchGenerationError(null);
    try {
      const response = await apiClient.post<{ queued: boolean; count: number }>(
        '/sonar/locationArchetypes/generate',
        {
          count: 50,
        }
      );
      setBatchGenerationMessage(
        `Queued ${response.count} new archetypes. Refreshing the list shortly.`
      );
      window.setTimeout(() => {
        refreshLocationArchetypes().catch((error) => {
          console.error(
            'Failed to refresh location archetypes after generation',
            error
          );
        });
      }, 6000);
    } catch (error) {
      console.error('Failed to queue location archetype generation', error);
      setBatchGenerationError('Failed to queue location archetype generation.');
    } finally {
      setQueueingBatchGeneration(false);
    }
  };

  const filteredArchetypes = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    if (!query) return locationArchetypes;
    return locationArchetypes.filter((archetype) =>
      archetype.name.toLowerCase().includes(query)
    );
  }, [locationArchetypes, searchQuery]);

  const resetCreate = () => {
    setCreateDraft(buildEmptyArchetype());
    setCreateIncludedQuery('');
    setCreateExcludedQuery('');
    setCreateChallengeQuery('');
    setCreateChallengeInputType('photo');
    setCreateChallengeProficiency('');
    setCreateChallengeDifficulty(0);
    setCreateGeneratedChallenges([]);
    setCreateGenerating(false);
    setCreateGenerateError(null);
  };

  const openEdit = (archetype: LocationArchetype) => {
    setEditDraft({
      ...archetype,
      includedTypes: [...archetype.includedTypes],
      excludedTypes: [...archetype.excludedTypes],
      challenges: archetype.challenges.map((challenge) => ({
        ...challenge,
        submissionType: (challenge.submissionType ??
          'photo') as QuestNodeSubmissionType,
        proficiency: challenge.proficiency ?? '',
        difficulty: challenge.difficulty ?? 0,
      })),
    });
    setEditIncludedQuery('');
    setEditExcludedQuery('');
    setEditChallengeQuery('');
    setEditChallengeInputType('photo');
    setEditChallengeProficiency('');
    setEditChallengeDifficulty(0);
    setEditGeneratedChallenges([]);
    setEditGenerating(false);
    setEditGenerateError(null);
  };

  const addUnique = (list: string[], value: string) => {
    if (list.includes(value)) return list;
    return [...list, value];
  };

  const addChallengeUnique = (
    list: LocationArchetypeChallenge[],
    challenge: LocationArchetypeChallenge
  ) => {
    const proficiencyKey = (challenge.proficiency ?? '').toLowerCase();
    const difficultyKey = String(challenge.difficulty ?? 0);
    const key = `${challenge.question.toLowerCase()}|${challenge.submissionType}|${proficiencyKey}|${difficultyKey}`;
    if (
      list.some(
        (item) =>
          `${item.question.toLowerCase()}|${item.submissionType}|${(item.proficiency ?? '').toLowerCase()}|${String(
            item.difficulty ?? 0
          )}` === key
      )
    ) {
      return list;
    }
    return [...list, challenge];
  };

  const formatSubmissionType = (value?: QuestNodeSubmissionType) => {
    return (value ?? 'photo').toUpperCase();
  };

  const mergeUniqueChallenges = (
    list: LocationArchetypeChallenge[],
    additions: LocationArchetypeChallenge[]
  ) => {
    let next = [...list];
    additions.forEach((challenge) => {
      next = addChallengeUnique(next, challenge);
    });
    return next;
  };

  const clampChallengePreview = (
    items: LocationArchetypeChallenge[],
    limit: number
  ) => {
    if (items.length <= limit) {
      return { preview: items, remaining: 0 };
    }
    return { preview: items.slice(0, limit), remaining: items.length - limit };
  };

  const createIncludedOptions = placeTypes
    .filter((type) =>
      type.toLowerCase().includes(createIncludedQuery.trim().toLowerCase())
    )
    .filter((type) => !createDraft.includedTypes.includes(type))
    .slice(0, 8);

  const createExcludedOptions = placeTypes
    .filter((type) =>
      type.toLowerCase().includes(createExcludedQuery.trim().toLowerCase())
    )
    .filter((type) => !createDraft.excludedTypes.includes(type))
    .slice(0, 8);

  const editIncludedOptions = editDraft
    ? placeTypes
        .filter((type) =>
          type.toLowerCase().includes(editIncludedQuery.trim().toLowerCase())
        )
        .filter((type) => !editDraft.includedTypes.includes(type))
        .slice(0, 8)
    : [];

  const editExcludedOptions = editDraft
    ? placeTypes
        .filter((type) =>
          type.toLowerCase().includes(editExcludedQuery.trim().toLowerCase())
        )
        .filter((type) => !editDraft.excludedTypes.includes(type))
        .slice(0, 8)
    : [];

  const hydrateGeneratedChallenges = (
    challenges: {
      question: string;
      submissionType: string;
      proficiency?: string | null;
      difficulty?: number | null;
    }[]
  ): GeneratedLocationChallenge[] => {
    const seed = Date.now();
    return challenges.map((challenge, index) => ({
      id: `${seed}-${index}-${Math.random().toString(16).slice(2, 6)}`,
      question: challenge.question,
      submissionType: (challenge.submissionType ||
        'photo') as QuestNodeSubmissionType,
      proficiency: challenge.proficiency ?? null,
      difficulty:
        typeof challenge.difficulty === 'number' &&
        Number.isFinite(challenge.difficulty)
          ? challenge.difficulty
          : 0,
    }));
  };

  const requestGeneratedChallenges = async (
    draft: LocationArchetype | null,
    setGenerated: React.Dispatch<
      React.SetStateAction<GeneratedLocationChallenge[]>
    >,
    setGenerating: React.Dispatch<React.SetStateAction<boolean>>,
    setError: React.Dispatch<React.SetStateAction<string | null>>
  ) => {
    if (!draft) return;
    setGenerating(true);
    setError(null);
    try {
      const response = await apiClient.post<{
        challenges: {
          question: string;
          submissionType: string;
          proficiency?: string | null;
          difficulty?: number | null;
        }[];
      }>('/sonar/locationArchetypes/challenges/generate', {
        name: draft.name,
        includedTypes: draft.includedTypes,
        excludedTypes: draft.excludedTypes,
        allowedSubmissionTypes: submissionTypeOptions.map(
          (option) => option.value
        ),
        count: 10,
      });
      setGenerated(hydrateGeneratedChallenges(response.challenges ?? []));
    } catch (error) {
      console.error('Failed to generate location challenges', error);
      setError('Failed to generate challenges.');
    } finally {
      setGenerating(false);
    }
  };

  return (
    <div className="qa-theme">
      <div className="qa-shell">
        <header className="qa-hero">
          <div>
            <div className="qa-kicker">Location Blueprints</div>
            <h1 className="qa-title">Location Archetypes</h1>
            <p className="qa-subtitle">
              Decide where quests can spawn by shaping included and excluded
              place types, then attach reusable challenge prompts.
            </p>
          </div>
          <div className="qa-hero-actions">
            <input
              type="text"
              className="qa-input"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search archetypes..."
            />
            <button
              className="qa-btn qa-btn-outline"
              onClick={queueBatchGeneration}
              disabled={queueingBatchGeneration}
            >
              {queueingBatchGeneration
                ? 'Queueing...'
                : 'Generate 50 Archetypes'}
            </button>
            <button
              className="qa-btn qa-btn-primary"
              onClick={() => setShowCreateModal(true)}
            >
              New Location Archetype
            </button>
          </div>
        </header>

        {batchGenerationMessage && (
          <p className="qa-muted" style={{ marginTop: 12 }}>
            {batchGenerationMessage}
          </p>
        )}
        {batchGenerationError && (
          <p className="qa-muted" style={{ marginTop: 12, color: '#f99b9b' }}>
            {batchGenerationError}
          </p>
        )}

        <section className="qa-grid">
          {filteredArchetypes.length === 0 ? (
            <div className="qa-panel">
              <div className="qa-card-title">No archetypes found</div>
              <p className="qa-muted" style={{ marginTop: 8 }}>
                Try a different search or create a new location archetype.
              </p>
            </div>
          ) : (
            filteredArchetypes.map((archetype, index) => {
              const includedPreview = clampPreview(archetype.includedTypes, 6);
              const excludedPreview = clampPreview(archetype.excludedTypes, 6);
              const challengePreview = clampChallengePreview(
                archetype.challenges,
                6
              );
              return (
                <article
                  key={archetype.id}
                  className="qa-card"
                  style={{ animationDelay: `${index * 0.06}s` }}
                >
                  <div className="qa-card-header">
                    <div>
                      <h3 className="qa-card-title">{archetype.name}</h3>
                      <div className="qa-meta">
                        {archetype.includedTypes.length} included ·{' '}
                        {archetype.excludedTypes.length} excluded ·{' '}
                        {archetype.challenges.length} challenges
                      </div>
                    </div>
                    <div className="qa-actions">
                      <button
                        className="qa-btn qa-btn-outline"
                        onClick={() => openEdit(archetype)}
                      >
                        Edit
                      </button>
                      <button
                        className="qa-btn qa-btn-danger"
                        onClick={() => {
                          if (
                            window.confirm('Delete this location archetype?')
                          ) {
                            deleteLocationArchetype(archetype.id);
                          }
                        }}
                      >
                        Delete
                      </button>
                    </div>
                  </div>

                  <div className="qa-stat-grid">
                    <div className="qa-stat">
                      <div className="qa-stat-label">Included Types</div>
                      <div className="qa-stat-value">
                        {archetype.includedTypes.length}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Excluded Types</div>
                      <div className="qa-stat-value">
                        {archetype.excludedTypes.length}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Challenge Prompts</div>
                      <div className="qa-stat-value">
                        {archetype.challenges.length}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Updated</div>
                      <div className="qa-stat-value">
                        {new Date(archetype.updatedAt).toLocaleDateString()}
                      </div>
                    </div>
                  </div>

                  <div className="qa-divider" />

                  <div className="qa-panel">
                    <div className="qa-meta">Included Types</div>
                    <div className="qa-inline" style={{ marginTop: 10 }}>
                      {includedPreview.preview.length === 0 ? (
                        <span className="qa-empty">None selected.</span>
                      ) : (
                        includedPreview.preview.map((type) => (
                          <span key={type} className="qa-chip accent">
                            {type}
                          </span>
                        ))
                      )}
                      {includedPreview.remaining > 0 && (
                        <span className="qa-chip muted">
                          +{includedPreview.remaining} more
                        </span>
                      )}
                    </div>
                  </div>

                  <div className="qa-panel" style={{ marginTop: 16 }}>
                    <div className="qa-meta">Excluded Types</div>
                    <div className="qa-inline" style={{ marginTop: 10 }}>
                      {excludedPreview.preview.length === 0 ? (
                        <span className="qa-empty">None selected.</span>
                      ) : (
                        excludedPreview.preview.map((type) => (
                          <span key={type} className="qa-chip danger">
                            {type}
                          </span>
                        ))
                      )}
                      {excludedPreview.remaining > 0 && (
                        <span className="qa-chip muted">
                          +{excludedPreview.remaining} more
                        </span>
                      )}
                    </div>
                  </div>

                  <div className="qa-panel" style={{ marginTop: 16 }}>
                    <div className="qa-meta">Challenge Prompts</div>
                    <div className="qa-inline" style={{ marginTop: 10 }}>
                      {challengePreview.preview.length === 0 ? (
                        <span className="qa-empty">No challenges yet.</span>
                      ) : (
                        challengePreview.preview.map((challenge) => (
                          <span
                            key={`${challenge.question}-${challenge.submissionType}-${challenge.proficiency ?? ''}-${challenge.difficulty ?? 0}`}
                            className="qa-chip success"
                          >
                            {challenge.question} ·{' '}
                            {formatSubmissionType(challenge.submissionType)}
                            {challenge.proficiency
                              ? ` · ${challenge.proficiency}`
                              : ''}
                            {challenge.difficulty !== undefined &&
                            challenge.difficulty !== null
                              ? ` · Difficulty: ${challenge.difficulty}`
                              : ''}
                          </span>
                        ))
                      )}
                      {challengePreview.remaining > 0 && (
                        <span className="qa-chip muted">
                          +{challengePreview.remaining} more
                        </span>
                      )}
                    </div>
                  </div>
                </article>
              );
            })
          )}
        </section>
      </div>

      {showCreateModal && (
        <div className="qa-modal">
          <div className="qa-modal-card">
            <h2 className="qa-modal-title">Create Location Archetype</h2>
            <form
              className="qa-form-grid"
              onSubmit={async (event) => {
                event.preventDefault();
                await createLocationArchetype(createDraft);
                resetCreate();
                setShowCreateModal(false);
              }}
            >
              <div className="qa-field">
                <div className="qa-label">Name</div>
                <input
                  type="text"
                  className="qa-input"
                  value={createDraft.name}
                  onChange={(e) =>
                    setCreateDraft({ ...createDraft, name: e.target.value })
                  }
                  required
                />
              </div>

              <div className="qa-field">
                <div className="qa-label">Included Types</div>
                <div className="qa-combobox">
                  <input
                    type="text"
                    className="qa-input"
                    value={createIncludedQuery}
                    onChange={(e) => setCreateIncludedQuery(e.target.value)}
                    placeholder="Search place types..."
                  />
                  {createIncludedQuery.trim().length > 0 && (
                    <div className="qa-combobox-list">
                      {createIncludedOptions.length === 0 ? (
                        <div className="qa-combobox-empty">No matches.</div>
                      ) : (
                        createIncludedOptions.map((type) => (
                          <button
                            key={type}
                            type="button"
                            className="qa-combobox-option"
                            onClick={() => {
                              setCreateDraft({
                                ...createDraft,
                                includedTypes: addUnique(
                                  createDraft.includedTypes,
                                  type
                                ),
                              });
                              setCreateIncludedQuery('');
                            }}
                          >
                            {type}
                          </button>
                        ))
                      )}
                    </div>
                  )}
                </div>
                <PlaceTypeImporter
                  label="Import included types from a place"
                  onApplyTypes={(types) =>
                    setCreateDraft((prev) => ({
                      ...prev,
                      includedTypes: mergeUnique(prev.includedTypes, types),
                    }))
                  }
                />
                <div className="qa-panel" style={{ marginTop: 12 }}>
                  {createDraft.includedTypes.length === 0 ? (
                    <div className="qa-empty">No included types yet.</div>
                  ) : (
                    createDraft.includedTypes.map((type) => (
                      <div
                        key={type}
                        className="qa-inline"
                        style={{ marginBottom: 8 }}
                      >
                        <span className="qa-chip accent">{type}</span>
                        <button
                          type="button"
                          className="qa-btn qa-btn-text"
                          onClick={() =>
                            setCreateDraft({
                              ...createDraft,
                              includedTypes: createDraft.includedTypes.filter(
                                (value) => value !== type
                              ),
                            })
                          }
                        >
                          Remove
                        </button>
                      </div>
                    ))
                  )}
                </div>
              </div>

              <div className="qa-field">
                <div className="qa-label">Excluded Types</div>
                <div className="qa-combobox">
                  <input
                    type="text"
                    className="qa-input"
                    value={createExcludedQuery}
                    onChange={(e) => setCreateExcludedQuery(e.target.value)}
                    placeholder="Search place types..."
                  />
                  {createExcludedQuery.trim().length > 0 && (
                    <div className="qa-combobox-list">
                      {createExcludedOptions.length === 0 ? (
                        <div className="qa-combobox-empty">No matches.</div>
                      ) : (
                        createExcludedOptions.map((type) => (
                          <button
                            key={type}
                            type="button"
                            className="qa-combobox-option"
                            onClick={() => {
                              setCreateDraft({
                                ...createDraft,
                                excludedTypes: addUnique(
                                  createDraft.excludedTypes,
                                  type
                                ),
                              });
                              setCreateExcludedQuery('');
                            }}
                          >
                            {type}
                          </button>
                        ))
                      )}
                    </div>
                  )}
                </div>
                <PlaceTypeImporter
                  label="Import excluded types from a place"
                  onApplyTypes={(types) =>
                    setCreateDraft((prev) => ({
                      ...prev,
                      excludedTypes: mergeUnique(prev.excludedTypes, types),
                    }))
                  }
                />
                <div className="qa-panel" style={{ marginTop: 12 }}>
                  {createDraft.excludedTypes.length === 0 ? (
                    <div className="qa-empty">No excluded types yet.</div>
                  ) : (
                    createDraft.excludedTypes.map((type) => (
                      <div
                        key={type}
                        className="qa-inline"
                        style={{ marginBottom: 8 }}
                      >
                        <span className="qa-chip danger">{type}</span>
                        <button
                          type="button"
                          className="qa-btn qa-btn-text"
                          onClick={() =>
                            setCreateDraft({
                              ...createDraft,
                              excludedTypes: createDraft.excludedTypes.filter(
                                (value) => value !== type
                              ),
                            })
                          }
                        >
                          Remove
                        </button>
                      </div>
                    ))
                  )}
                </div>
              </div>

              <div className="qa-field">
                <div className="qa-label">Challenges</div>
                <div className="qa-inline">
                  <input
                    type="text"
                    className="qa-input"
                    value={createChallengeQuery}
                    onChange={(e) => setCreateChallengeQuery(e.target.value)}
                    placeholder="Add challenge prompt"
                  />
                  <select
                    className="qa-select"
                    value={createChallengeInputType}
                    onChange={(event) =>
                      setCreateChallengeInputType(
                        event.target.value as QuestNodeSubmissionType
                      )
                    }
                  >
                    {submissionTypeOptions.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                  <input
                    type="text"
                    className="qa-input"
                    value={createChallengeProficiency}
                    onChange={(event) =>
                      setCreateChallengeProficiency(event.target.value)
                    }
                    placeholder="Proficiency (optional)"
                  />
                  <input
                    type="number"
                    min={0}
                    className="qa-input"
                    value={createChallengeDifficulty}
                    onChange={(event) =>
                      setCreateChallengeDifficulty(
                        Number(event.target.value) || 0
                      )
                    }
                    placeholder="Difficulty"
                  />
                  <button
                    type="button"
                    className="qa-btn qa-btn-ghost"
                    onClick={() => {
                      const value = createChallengeQuery.trim();
                      const proficiency = createChallengeProficiency.trim();
                      const difficulty = Number.isFinite(
                        createChallengeDifficulty
                      )
                        ? Math.max(0, createChallengeDifficulty)
                        : 0;
                      if (!value) return;
                      setCreateDraft({
                        ...createDraft,
                        challenges: addChallengeUnique(createDraft.challenges, {
                          question: value,
                          submissionType: createChallengeInputType,
                          proficiency: proficiency ? proficiency : undefined,
                          difficulty,
                        }),
                      });
                      setCreateChallengeQuery('');
                      setCreateChallengeProficiency('');
                      setCreateChallengeDifficulty(0);
                    }}
                  >
                    Add
                  </button>
                </div>
                <div className="qa-panel" style={{ marginTop: 12 }}>
                  <div className="qa-meta">Challenge generator</div>
                  <p className="qa-muted" style={{ marginTop: 6 }}>
                    Generate 10 themed challenges based on the archetype
                    details.
                  </p>
                  <div className="qa-inline" style={{ marginTop: 10 }}>
                    <button
                      type="button"
                      className="qa-btn qa-btn-ghost"
                      onClick={() =>
                        requestGeneratedChallenges(
                          createDraft,
                          setCreateGeneratedChallenges,
                          setCreateGenerating,
                          setCreateGenerateError
                        )
                      }
                      disabled={createGenerating}
                    >
                      {createGenerating
                        ? 'Generating...'
                        : 'Generate 10 Challenges'}
                    </button>
                    {createGeneratedChallenges.length > 0 && (
                      <button
                        type="button"
                        className="qa-btn qa-btn-outline"
                        onClick={() => setCreateGeneratedChallenges([])}
                      >
                        Clear
                      </button>
                    )}
                  </div>
                  {createGenerateError && (
                    <div className="qa-error" style={{ marginTop: 8 }}>
                      {createGenerateError}
                    </div>
                  )}
                </div>
                {createGeneratedChallenges.length > 0 && (
                  <div className="qa-panel" style={{ marginTop: 12 }}>
                    <div className="qa-meta">Generated options</div>
                    <div
                      className="qa-generated-list"
                      style={{ marginTop: 10 }}
                    >
                      {createGeneratedChallenges.map((challenge) => (
                        <div key={challenge.id} className="qa-generated-row">
                          <div className="qa-generated-info">
                            <div className="qa-option-title">
                              {challenge.question}
                            </div>
                            <div className="qa-option-sub">
                              Input:{' '}
                              {formatSubmissionType(challenge.submissionType)}
                              {challenge.proficiency
                                ? ` · Proficiency: ${challenge.proficiency}`
                                : ''}
                              {challenge.difficulty !== undefined &&
                              challenge.difficulty !== null
                                ? ` · Difficulty: ${challenge.difficulty}`
                                : ''}
                            </div>
                          </div>
                          <button
                            type="button"
                            className="qa-btn qa-btn-ghost"
                            onClick={() => {
                              setCreateDraft((prev) => ({
                                ...prev,
                                challenges: addChallengeUnique(
                                  prev.challenges,
                                  {
                                    question: challenge.question,
                                    submissionType:
                                      challenge.submissionType as QuestNodeSubmissionType,
                                    proficiency:
                                      challenge.proficiency ?? undefined,
                                    difficulty: challenge.difficulty ?? 0,
                                  }
                                ),
                              }));
                            }}
                          >
                            Add
                          </button>
                        </div>
                      ))}
                    </div>
                    <div className="qa-inline" style={{ marginTop: 12 }}>
                      <button
                        type="button"
                        className="qa-btn qa-btn-outline"
                        onClick={() => {
                          setCreateDraft((prev) => ({
                            ...prev,
                            challenges: mergeUniqueChallenges(
                              prev.challenges,
                              createGeneratedChallenges.map((challenge) => ({
                                question: challenge.question,
                                submissionType:
                                  challenge.submissionType as QuestNodeSubmissionType,
                                proficiency: challenge.proficiency ?? undefined,
                                difficulty: challenge.difficulty ?? 0,
                              }))
                            ),
                          }));
                        }}
                      >
                        Add All
                      </button>
                    </div>
                  </div>
                )}
                <div className="qa-panel" style={{ marginTop: 12 }}>
                  {createDraft.challenges.length === 0 ? (
                    <div className="qa-empty">No challenges yet.</div>
                  ) : (
                    createDraft.challenges.map((challenge, index) => (
                      <div
                        key={`${challenge.question}-${challenge.submissionType}-${index}`}
                        className="qa-inline"
                        style={{ marginBottom: 8 }}
                      >
                        <span className="qa-chip success">
                          {challenge.question} ·{' '}
                          {formatSubmissionType(challenge.submissionType)} ·
                          Difficulty: {challenge.difficulty ?? 0}
                        </span>
                        <input
                          type="text"
                          className="qa-input"
                          style={{ minWidth: 180, flex: '1 1 200px' }}
                          value={challenge.proficiency ?? ''}
                          onChange={(event) => {
                            const value = event.target.value;
                            setCreateDraft((prev) => ({
                              ...prev,
                              challenges: prev.challenges.map((item, i) =>
                                i === index
                                  ? { ...item, proficiency: value }
                                  : item
                              ),
                            }));
                          }}
                          placeholder="Proficiency (optional)"
                        />
                        <input
                          type="number"
                          min={0}
                          className="qa-input"
                          style={{ width: 120 }}
                          value={challenge.difficulty ?? 0}
                          onChange={(event) => {
                            const value = Number(event.target.value) || 0;
                            setCreateDraft((prev) => ({
                              ...prev,
                              challenges: prev.challenges.map((item, i) =>
                                i === index
                                  ? { ...item, difficulty: value }
                                  : item
                              ),
                            }));
                          }}
                          placeholder="Difficulty"
                        />
                        <button
                          type="button"
                          className="qa-btn qa-btn-text"
                          onClick={() =>
                            setCreateDraft({
                              ...createDraft,
                              challenges: createDraft.challenges.filter(
                                (_, i) => i !== index
                              ),
                            })
                          }
                        >
                          Remove
                        </button>
                      </div>
                    ))
                  )}
                </div>
              </div>

              <div className="qa-footer">
                <button
                  type="button"
                  className="qa-btn qa-btn-outline"
                  onClick={() => {
                    setShowCreateModal(false);
                    resetCreate();
                  }}
                >
                  Cancel
                </button>
                <button type="submit" className="qa-btn qa-btn-primary">
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {editDraft && (
        <div className="qa-modal">
          <div className="qa-modal-card">
            <h2 className="qa-modal-title">Edit Location Archetype</h2>
            <div className="qa-form-grid">
              <div className="qa-field">
                <div className="qa-label">Name</div>
                <input
                  type="text"
                  className="qa-input"
                  value={editDraft.name}
                  onChange={(e) =>
                    setEditDraft({ ...editDraft, name: e.target.value })
                  }
                />
              </div>

              <div className="qa-field">
                <div className="qa-label">Included Types</div>
                <div className="qa-combobox">
                  <input
                    type="text"
                    className="qa-input"
                    value={editIncludedQuery}
                    onChange={(e) => setEditIncludedQuery(e.target.value)}
                    placeholder="Search place types..."
                  />
                  {editIncludedQuery.trim().length > 0 && (
                    <div className="qa-combobox-list">
                      {editIncludedOptions.length === 0 ? (
                        <div className="qa-combobox-empty">No matches.</div>
                      ) : (
                        editIncludedOptions.map((type) => (
                          <button
                            key={type}
                            type="button"
                            className="qa-combobox-option"
                            onClick={() => {
                              if (!editDraft) return;
                              setEditDraft({
                                ...editDraft,
                                includedTypes: addUnique(
                                  editDraft.includedTypes,
                                  type
                                ),
                              });
                              setEditIncludedQuery('');
                            }}
                          >
                            {type}
                          </button>
                        ))
                      )}
                    </div>
                  )}
                </div>
                <PlaceTypeImporter
                  label="Import included types from a place"
                  onApplyTypes={(types) =>
                    setEditDraft((prev) =>
                      prev
                        ? {
                            ...prev,
                            includedTypes: mergeUnique(
                              prev.includedTypes,
                              types
                            ),
                          }
                        : prev
                    )
                  }
                />
                <div className="qa-panel" style={{ marginTop: 12 }}>
                  {editDraft.includedTypes.length === 0 ? (
                    <div className="qa-empty">No included types yet.</div>
                  ) : (
                    editDraft.includedTypes.map((type) => (
                      <div
                        key={type}
                        className="qa-inline"
                        style={{ marginBottom: 8 }}
                      >
                        <span className="qa-chip accent">{type}</span>
                        <button
                          type="button"
                          className="qa-btn qa-btn-text"
                          onClick={() =>
                            setEditDraft({
                              ...editDraft,
                              includedTypes: editDraft.includedTypes.filter(
                                (value) => value !== type
                              ),
                            })
                          }
                        >
                          Remove
                        </button>
                      </div>
                    ))
                  )}
                </div>
              </div>

              <div className="qa-field">
                <div className="qa-label">Excluded Types</div>
                <div className="qa-combobox">
                  <input
                    type="text"
                    className="qa-input"
                    value={editExcludedQuery}
                    onChange={(e) => setEditExcludedQuery(e.target.value)}
                    placeholder="Search place types..."
                  />
                  {editExcludedQuery.trim().length > 0 && (
                    <div className="qa-combobox-list">
                      {editExcludedOptions.length === 0 ? (
                        <div className="qa-combobox-empty">No matches.</div>
                      ) : (
                        editExcludedOptions.map((type) => (
                          <button
                            key={type}
                            type="button"
                            className="qa-combobox-option"
                            onClick={() => {
                              if (!editDraft) return;
                              setEditDraft({
                                ...editDraft,
                                excludedTypes: addUnique(
                                  editDraft.excludedTypes,
                                  type
                                ),
                              });
                              setEditExcludedQuery('');
                            }}
                          >
                            {type}
                          </button>
                        ))
                      )}
                    </div>
                  )}
                </div>
                <PlaceTypeImporter
                  label="Import excluded types from a place"
                  onApplyTypes={(types) =>
                    setEditDraft((prev) =>
                      prev
                        ? {
                            ...prev,
                            excludedTypes: mergeUnique(
                              prev.excludedTypes,
                              types
                            ),
                          }
                        : prev
                    )
                  }
                />
                <div className="qa-panel" style={{ marginTop: 12 }}>
                  {editDraft.excludedTypes.length === 0 ? (
                    <div className="qa-empty">No excluded types yet.</div>
                  ) : (
                    editDraft.excludedTypes.map((type) => (
                      <div
                        key={type}
                        className="qa-inline"
                        style={{ marginBottom: 8 }}
                      >
                        <span className="qa-chip danger">{type}</span>
                        <button
                          type="button"
                          className="qa-btn qa-btn-text"
                          onClick={() =>
                            setEditDraft({
                              ...editDraft,
                              excludedTypes: editDraft.excludedTypes.filter(
                                (value) => value !== type
                              ),
                            })
                          }
                        >
                          Remove
                        </button>
                      </div>
                    ))
                  )}
                </div>
              </div>

              <div className="qa-field">
                <div className="qa-label">Challenges</div>
                <div className="qa-inline">
                  <input
                    type="text"
                    className="qa-input"
                    value={editChallengeQuery}
                    onChange={(e) => setEditChallengeQuery(e.target.value)}
                    placeholder="Add challenge prompt"
                  />
                  <select
                    className="qa-select"
                    value={editChallengeInputType}
                    onChange={(event) =>
                      setEditChallengeInputType(
                        event.target.value as QuestNodeSubmissionType
                      )
                    }
                  >
                    {submissionTypeOptions.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                  <input
                    type="text"
                    className="qa-input"
                    value={editChallengeProficiency}
                    onChange={(event) =>
                      setEditChallengeProficiency(event.target.value)
                    }
                    placeholder="Proficiency (optional)"
                  />
                  <input
                    type="number"
                    min={0}
                    className="qa-input"
                    value={editChallengeDifficulty}
                    onChange={(event) =>
                      setEditChallengeDifficulty(
                        Number(event.target.value) || 0
                      )
                    }
                    placeholder="Difficulty"
                  />
                  <button
                    type="button"
                    className="qa-btn qa-btn-ghost"
                    onClick={() => {
                      if (!editDraft) return;
                      const value = editChallengeQuery.trim();
                      const proficiency = editChallengeProficiency.trim();
                      const difficulty = Number.isFinite(
                        editChallengeDifficulty
                      )
                        ? Math.max(0, editChallengeDifficulty)
                        : 0;
                      if (!value) return;
                      setEditDraft({
                        ...editDraft,
                        challenges: addChallengeUnique(editDraft.challenges, {
                          question: value,
                          submissionType: editChallengeInputType,
                          proficiency: proficiency ? proficiency : undefined,
                          difficulty,
                        }),
                      });
                      setEditChallengeQuery('');
                      setEditChallengeProficiency('');
                      setEditChallengeDifficulty(0);
                    }}
                  >
                    Add
                  </button>
                </div>
                <div className="qa-panel" style={{ marginTop: 12 }}>
                  <div className="qa-meta">Challenge generator</div>
                  <p className="qa-muted" style={{ marginTop: 6 }}>
                    Generate 10 themed challenges based on the archetype
                    details.
                  </p>
                  <div className="qa-inline" style={{ marginTop: 10 }}>
                    <button
                      type="button"
                      className="qa-btn qa-btn-ghost"
                      onClick={() =>
                        requestGeneratedChallenges(
                          editDraft,
                          setEditGeneratedChallenges,
                          setEditGenerating,
                          setEditGenerateError
                        )
                      }
                      disabled={editGenerating}
                    >
                      {editGenerating
                        ? 'Generating...'
                        : 'Generate 10 Challenges'}
                    </button>
                    {editGeneratedChallenges.length > 0 && (
                      <button
                        type="button"
                        className="qa-btn qa-btn-outline"
                        onClick={() => setEditGeneratedChallenges([])}
                      >
                        Clear
                      </button>
                    )}
                  </div>
                  {editGenerateError && (
                    <div className="qa-error" style={{ marginTop: 8 }}>
                      {editGenerateError}
                    </div>
                  )}
                </div>
                {editGeneratedChallenges.length > 0 && (
                  <div className="qa-panel" style={{ marginTop: 12 }}>
                    <div className="qa-meta">Generated options</div>
                    <div
                      className="qa-generated-list"
                      style={{ marginTop: 10 }}
                    >
                      {editGeneratedChallenges.map((challenge) => (
                        <div key={challenge.id} className="qa-generated-row">
                          <div className="qa-generated-info">
                            <div className="qa-option-title">
                              {challenge.question}
                            </div>
                            <div className="qa-option-sub">
                              Input:{' '}
                              {formatSubmissionType(challenge.submissionType)}
                              {challenge.proficiency
                                ? ` · Proficiency: ${challenge.proficiency}`
                                : ''}
                              {challenge.difficulty !== undefined &&
                              challenge.difficulty !== null
                                ? ` · Difficulty: ${challenge.difficulty}`
                                : ''}
                            </div>
                          </div>
                          <button
                            type="button"
                            className="qa-btn qa-btn-ghost"
                            onClick={() => {
                              setEditDraft((prev) =>
                                prev
                                  ? {
                                      ...prev,
                                      challenges: addChallengeUnique(
                                        prev.challenges,
                                        {
                                          question: challenge.question,
                                          submissionType:
                                            challenge.submissionType as QuestNodeSubmissionType,
                                          proficiency:
                                            challenge.proficiency ?? undefined,
                                          difficulty: challenge.difficulty ?? 0,
                                        }
                                      ),
                                    }
                                  : prev
                              );
                            }}
                          >
                            Add
                          </button>
                        </div>
                      ))}
                    </div>
                    <div className="qa-inline" style={{ marginTop: 12 }}>
                      <button
                        type="button"
                        className="qa-btn qa-btn-outline"
                        onClick={() => {
                          setEditDraft((prev) =>
                            prev
                              ? {
                                  ...prev,
                                  challenges: mergeUniqueChallenges(
                                    prev.challenges,
                                    editGeneratedChallenges.map(
                                      (challenge) => ({
                                        question: challenge.question,
                                        submissionType:
                                          challenge.submissionType as QuestNodeSubmissionType,
                                        proficiency:
                                          challenge.proficiency ?? undefined,
                                        difficulty: challenge.difficulty ?? 0,
                                      })
                                    )
                                  ),
                                }
                              : prev
                          );
                        }}
                      >
                        Add All
                      </button>
                    </div>
                  </div>
                )}
                <div className="qa-panel" style={{ marginTop: 12 }}>
                  {editDraft.challenges.length === 0 ? (
                    <div className="qa-empty">No challenges yet.</div>
                  ) : (
                    editDraft.challenges.map((challenge, index) => (
                      <div
                        key={`${challenge.question}-${challenge.submissionType}-${index}`}
                        className="qa-inline"
                        style={{ marginBottom: 8 }}
                      >
                        <span className="qa-chip success">
                          {challenge.question} ·{' '}
                          {formatSubmissionType(challenge.submissionType)} ·
                          Difficulty: {challenge.difficulty ?? 0}
                        </span>
                        <input
                          type="text"
                          className="qa-input"
                          style={{ minWidth: 180, flex: '1 1 200px' }}
                          value={challenge.proficiency ?? ''}
                          onChange={(event) => {
                            const value = event.target.value;
                            setEditDraft((prev) =>
                              prev
                                ? {
                                    ...prev,
                                    challenges: prev.challenges.map(
                                      (item, i) =>
                                        i === index
                                          ? { ...item, proficiency: value }
                                          : item
                                    ),
                                  }
                                : prev
                            );
                          }}
                          placeholder="Proficiency (optional)"
                        />
                        <input
                          type="number"
                          min={0}
                          className="qa-input"
                          style={{ width: 120 }}
                          value={challenge.difficulty ?? 0}
                          onChange={(event) => {
                            const value = Number(event.target.value) || 0;
                            setEditDraft((prev) =>
                              prev
                                ? {
                                    ...prev,
                                    challenges: prev.challenges.map(
                                      (item, i) =>
                                        i === index
                                          ? { ...item, difficulty: value }
                                          : item
                                    ),
                                  }
                                : prev
                            );
                          }}
                          placeholder="Difficulty"
                        />
                        <button
                          type="button"
                          className="qa-btn qa-btn-text"
                          onClick={() =>
                            setEditDraft({
                              ...editDraft,
                              challenges: editDraft.challenges.filter(
                                (_, i) => i !== index
                              ),
                            })
                          }
                        >
                          Remove
                        </button>
                      </div>
                    ))
                  )}
                </div>
              </div>
            </div>
            <div className="qa-footer">
              <button
                className="qa-btn qa-btn-outline"
                onClick={() => {
                  setEditDraft(null);
                  setEditGeneratedChallenges([]);
                  setEditGenerateError(null);
                  setEditGenerating(false);
                }}
              >
                Cancel
              </button>
              <button
                className="qa-btn qa-btn-primary"
                onClick={async () => {
                  if (!editDraft) return;
                  await updateLocationArchetype(editDraft);
                  setEditDraft(null);
                  setEditGeneratedChallenges([]);
                  setEditGenerateError(null);
                  setEditGenerating(false);
                }}
              >
                Save Changes
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default LocationArchetypes;
