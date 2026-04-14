import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAPI } from '@poltergeist/contexts';
import {
  District,
  MainStoryDistrictRun,
  MainStoryTemplate,
  LocationArchetype,
  QuestArchetype,
  QuestArchetypeChallenge,
  QuestArchetypeNode,
  Zone,
} from '@poltergeist/types';
import { useQuestArchetypes } from '../contexts/questArchetypes.tsx';
import MainStoryTemplateEditor from './MainStoryTemplateEditor.tsx';
import './questArchetypeTheme.css';

type MonsterTemplateSummary = {
  id: string;
  name: string;
};

type ScenarioTemplateSummary = {
  id: string;
  prompt: string;
};

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

const questArchetypeNodeTypeLabel = (
  nodeType?: QuestArchetypeNode['nodeType']
) => {
  switch (nodeType) {
    case 'challenge':
      return 'Challenge';
    case 'exposition':
      return 'Exposition';
    case 'fetch_quest':
      return 'Fetch Quest';
    case 'story_flag':
      return 'Story Flag';
    case 'scenario':
      return 'Scenario';
    case 'monster_encounter':
      return 'Monster Encounter';
    default:
      return 'Node';
  }
};

const summarizeQuestArchetypeRoot = (
  archetype: QuestArchetype | null | undefined
) => {
  if (!archetype?.root) {
    return 'No root node';
  }
  const nodeType = archetype.root.nodeType || 'node';
  if (nodeType === 'scenario') {
    return 'Starts with a scenario node';
  }
  if (nodeType === 'monster_encounter') {
    return 'Starts with a monster encounter node';
  }
  if (nodeType === 'exposition') {
    return 'Starts with an exposition node';
  }
  if (nodeType === 'fetch_quest') {
    return 'Starts with a fetch quest node';
  }
  if (nodeType === 'story_flag') {
    return 'Starts with a story flag node';
  }
  if (nodeType === 'challenge') {
    return 'Starts with a challenge node';
  }
  return `Starts with a ${nodeType} node`;
};

const countQuestArchetypeNodes = (
  archetype: QuestArchetype | null | undefined
) => {
  const countNode = (
    node: QuestArchetype['root'] | null | undefined
  ): number => {
    if (!node) {
      return 0;
    }
    return (
      1 +
      (node.challenges ?? []).reduce(
        (sum, challenge) => sum + countNode(challenge.unlockedNode),
        0
      )
    );
  };
  return countNode(archetype?.root);
};

const countQuestArchetypeChallenges = (
  archetype: QuestArchetype | null | undefined
) => {
  const countNodeChallenges = (
    node: QuestArchetype['root'] | null | undefined
  ): number => {
    if (!node) {
      return 0;
    }
    return (node.challenges ?? []).reduce(
      (sum, challenge) => sum + 1 + countNodeChallenges(challenge.unlockedNode),
      0
    );
  };
  return countNodeChallenges(archetype?.root);
};

const describeQuestArchetypeNode = (
  node: QuestArchetypeNode | null | undefined,
  locationArchetypesById: Map<string, LocationArchetype>,
  monsterTemplatesById: Map<string, MonsterTemplateSummary>,
  scenarioTemplatesById: Map<string, ScenarioTemplateSummary>
) => {
  if (!node) {
    return 'Unknown';
  }
  const locationLabel =
    node.locationSelectionMode === 'same_as_previous'
      ? 'Same as previous'
      : node.locationArchetypeId
        ? locationArchetypesById.get(node.locationArchetypeId)?.name ??
          'Point of interest'
        : 'Coordinates';
  if (node.nodeType === 'scenario') {
    const scenarioLabel = node.scenarioTemplateId
      ? scenarioTemplatesById.get(node.scenarioTemplateId)?.prompt ?? 'Scenario'
      : 'Scenario';
    return `${scenarioLabel} @ ${locationLabel}`;
  }
  if (node.nodeType === 'monster_encounter') {
    const names = (node.monsterTemplateIds ?? [])
      .map((templateId) => monsterTemplatesById.get(templateId)?.name)
      .filter(Boolean) as string[];
    const encounterLabel =
      names.length > 0
        ? `Encounter: ${names.slice(0, 3).join(', ')}${names.length > 3 ? '…' : ''}`
        : 'Monster encounter';
    return `${encounterLabel} @ ${locationLabel}`;
  }
  if (node.nodeType === 'exposition') {
    const expositionLabel = node.expositionTitle?.trim() || 'Exposition';
    return `${expositionLabel} @ ${locationLabel}`;
  }
  if (node.nodeType === 'fetch_quest') {
    const characterLabel = node.fetchCharacter?.name?.trim() || 'Character';
    const requirementCount = node.fetchRequirements?.length ?? 0;
    const label =
      requirementCount > 0
        ? `Deliver ${requirementCount} item${requirementCount === 1 ? '' : 's'} to ${characterLabel}`
        : `Fetch quest for ${characterLabel}`;
    return `${label} @ ${locationLabel}`;
  }
  if (node.nodeType === 'story_flag') {
    return `Story flag: ${node.storyFlagKey?.trim() || 'story flag'}`;
  }
  const challengeLabel =
    node.challengeTemplate?.question?.trim() || 'Challenge';
  return `${challengeLabel} @ ${locationLabel}`;
};

const describeQuestArchetypeChallenge = (
  challenge: QuestArchetypeChallenge
) => {
  const parts = [
    challenge.challengeTemplate?.question || 'Untitled challenge',
    challenge.proficiency || challenge.challengeTemplate?.proficiency || null,
    challenge.challengeTemplate?.submissionType || null,
  ].filter(Boolean);
  return parts.join(' · ');
};

export const MainStoryTemplates = () => {
  const { apiClient } = useAPI();
  const { questArchetypes, locationArchetypes, refreshQuestArchetypes } =
    useQuestArchetypes();
  const [templates, setTemplates] = useState<MainStoryTemplate[]>([]);
  const [districts, setDistricts] = useState<District[]>([]);
  const [runs, setRuns] = useState<MainStoryDistrictRun[]>([]);
  const [monsterTemplates, setMonsterTemplates] = useState<
    MonsterTemplateSummary[]
  >([]);
  const [scenarioTemplates, setScenarioTemplates] = useState<
    ScenarioTemplateSummary[]
  >([]);
  const [loading, setLoading] = useState(true);
  const [pageError, setPageError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [selectedDistrictByTemplate, setSelectedDistrictByTemplate] = useState<
    Record<string, string>
  >({});
  const [selectedZoneByTemplate, setSelectedZoneByTemplate] = useState<
    Record<string, string>
  >({});
  const [zoneQueryByTemplate, setZoneQueryByTemplate] = useState<
    Record<string, string>
  >({});
  const [openZoneComboboxTemplateId, setOpenZoneComboboxTemplateId] = useState<
    string | null
  >(null);
  const [instantiatingRunKey, setInstantiatingRunKey] = useState<string | null>(
    null
  );
  const [deletingTemplateId, setDeletingTemplateId] = useState<string | null>(
    null
  );
  const [editingTemplate, setEditingTemplate] = useState<
    MainStoryTemplate | 'new' | null
  >(null);
  const [savingTemplate, setSavingTemplate] = useState(false);
  const [retryingRunId, setRetryingRunId] = useState<string | null>(null);
  const [deletingRunId, setDeletingRunId] = useState<string | null>(null);
  const [expandedBeatKeys, setExpandedBeatKeys] = useState<string[]>([]);

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

  const questArchetypesById = useMemo(() => {
    const map = new Map<string, QuestArchetype>();
    questArchetypes.forEach((archetype) => {
      map.set(archetype.id, archetype);
    });
    return map;
  }, [questArchetypes]);

  const locationArchetypesById = useMemo(() => {
    const map = new Map<string, LocationArchetype>();
    locationArchetypes.forEach((archetype) => {
      map.set(archetype.id, archetype);
    });
    return map;
  }, [locationArchetypes]);

  const monsterTemplatesById = useMemo(() => {
    const map = new Map<string, MonsterTemplateSummary>();
    monsterTemplates.forEach((template) => {
      map.set(template.id, template);
    });
    return map;
  }, [monsterTemplates]);

  const scenarioTemplatesById = useMemo(() => {
    const map = new Map<string, ScenarioTemplateSummary>();
    scenarioTemplates.forEach((template) => {
      map.set(template.id, template);
    });
    return map;
  }, [scenarioTemplates]);

  const toggleBeatExpanded = useCallback((beatKey: string) => {
    setExpandedBeatKeys((current) =>
      current.includes(beatKey)
        ? current.filter((key) => key !== beatKey)
        : [...current, beatKey]
    );
  }, []);

  const renderQuestArchetypeNode = useCallback(
    (
      node: QuestArchetypeNode | null | undefined,
      depth = 0
    ): React.ReactNode => {
      if (!node) {
        return null;
      }

      const nodeSummary = describeQuestArchetypeNode(
        node,
        locationArchetypesById,
        monsterTemplatesById,
        scenarioTemplatesById
      );
      const scenarioPrompt =
        node.scenarioTemplateId &&
        scenarioTemplatesById.get(node.scenarioTemplateId)?.prompt;
      const monsterNames = (node.monsterTemplateIds ?? [])
        .map((templateId) => monsterTemplatesById.get(templateId)?.name)
        .filter(Boolean) as string[];
      const locationName =
        node.locationArchetypeId &&
        locationArchetypesById.get(node.locationArchetypeId)?.name;

      return (
        <div
          key={node.id}
          className="qa-flow-node"
          style={{
            borderColor:
              depth % 2 === 0
                ? 'rgba(255, 107, 74, 0.32)'
                : 'rgba(95, 211, 181, 0.28)',
            marginTop: depth === 0 ? 0 : 12,
          }}
        >
          <div className="qa-flow-node-card">
            <div className="qa-flow-node-header">
              <div>
                <div className="qa-flow-node-title">
                  {questArchetypeNodeTypeLabel(node.nodeType)} Node
                </div>
                <div className="qa-meta" style={{ marginTop: 4 }}>
                  {nodeSummary}
                </div>
              </div>
              <div className="qa-actions">
                <span className="qa-chip muted">
                  {node.challenges?.length ?? 0} branch
                  {(node.challenges?.length ?? 0) === 1 ? '' : 'es'}
                </span>
                {node.targetLevel ? (
                  <span className="qa-chip muted">Lvl {node.targetLevel}</span>
                ) : null}
                {node.encounterProximityMeters ? (
                  <span className="qa-chip muted">
                    {node.encounterProximityMeters}m
                  </span>
                ) : null}
              </div>
            </div>

            <div
              style={{
                display: 'grid',
                gap: 8,
                marginTop: 12,
              }}
            >
              {locationName ? (
                <div className="qa-meta">
                  <strong>Location:</strong> {locationName}
                </div>
              ) : null}
              {scenarioPrompt ? (
                <div className="qa-meta">
                  <strong>Scenario:</strong> {scenarioPrompt}
                </div>
              ) : null}
              {monsterNames.length > 0 ? (
                <div className="qa-meta">
                  <strong>Monsters:</strong> {monsterNames.join(', ')}
                </div>
              ) : null}
            </div>

            {(node.challenges?.length ?? 0) > 0 ? (
              <div className="qa-flow-challenges" style={{ marginTop: 14 }}>
                {node.challenges.map((challenge) => (
                  <div key={challenge.id} className="qa-flow-challenge-card">
                    <div className="qa-flow-challenge-header">
                      <div>
                        <div className="qa-flow-challenge-title">
                          {challenge.challengeTemplate?.question ||
                            'Challenge branch'}
                        </div>
                        <div className="qa-meta" style={{ marginTop: 4 }}>
                          {describeQuestArchetypeChallenge(challenge)}
                        </div>
                      </div>
                    </div>
                    {challenge.challengeTemplate?.description ? (
                      <div className="qa-meta" style={{ marginTop: 10 }}>
                        {challenge.challengeTemplate.description}
                      </div>
                    ) : null}
                    {challenge.unlockedNode ? (
                      <div className="qa-flow-branch" style={{ marginTop: 12 }}>
                        <div className="qa-flow-branch-label">
                          Unlocks next node
                        </div>
                        {renderQuestArchetypeNode(
                          challenge.unlockedNode,
                          depth + 1
                        )}
                      </div>
                    ) : null}
                  </div>
                ))}
              </div>
            ) : (
              <div className="qa-meta" style={{ marginTop: 12 }}>
                No challenge branches on this node.
              </div>
            )}
          </div>
        </div>
      );
    },
    [locationArchetypesById, monsterTemplatesById, scenarioTemplatesById]
  );

  const loadPage = useCallback(async () => {
    setLoading(true);
    try {
      const [templatesResponse, districtsResponse, runsResponse] =
        await Promise.all([
          apiClient.get<MainStoryTemplate[]>('/sonar/mainStoryTemplates'),
          apiClient.get<District[]>('/sonar/districts'),
          apiClient.get<MainStoryDistrictRun[]>('/sonar/mainStoryDistrictRuns'),
        ]);
      await refreshQuestArchetypes();
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
      setSelectedZoneByTemplate((current) => {
        const next = { ...current };
        const fallbackZoneId = districtsResponse[0]?.zones?.[0]?.id ?? '';
        templatesResponse.forEach((template) => {
          if (!next[template.id]) {
            next[template.id] = fallbackZoneId;
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
  }, [apiClient, refreshQuestArchetypes]);

  useEffect(() => {
    void loadPage();
  }, [loadPage]);

  useEffect(() => {
    let active = true;

    const loadReferenceData = async () => {
      try {
        const [monsterResponse, scenarioResponse] = await Promise.all([
          apiClient.get<{ items: MonsterTemplateSummary[] }>(
            '/sonar/admin/monster-templates?page=1&pageSize=500'
          ),
          apiClient.get<{ items: ScenarioTemplateSummary[] }>(
            '/sonar/admin/scenario-templates?page=1&pageSize=500'
          ),
        ]);
        if (!active) {
          return;
        }
        setMonsterTemplates(monsterResponse.items ?? []);
        setScenarioTemplates(scenarioResponse.items ?? []);
      } catch (error) {
        console.error(
          'Failed to load main story template reference data',
          error
        );
      }
    };

    void loadReferenceData();

    return () => {
      active = false;
    };
  }, [apiClient]);

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

  const resolveSelectedZone = useCallback(
    (
      templateId: string,
      district: District | null | undefined
    ): Zone | null => {
      const zones = district?.zones ?? [];
      if (zones.length === 0) {
        return null;
      }
      if (
        !Object.prototype.hasOwnProperty.call(
          selectedZoneByTemplate,
          templateId
        )
      ) {
        return zones[0] ?? null;
      }
      const selectedZoneId = selectedZoneByTemplate[templateId];
      if (!selectedZoneId) {
        return null;
      }
      return (
        zones.find((zone) => zone.id === selectedZoneId) ?? zones[0] ?? null
      );
    },
    [selectedZoneByTemplate]
  );

  const selectZoneForTemplate = useCallback(
    (templateId: string, zone: Zone | null) => {
      setSelectedZoneByTemplate((current) => ({
        ...current,
        [templateId]: zone?.id ?? '',
      }));
      setZoneQueryByTemplate((current) => ({
        ...current,
        [templateId]: zone?.name ?? '',
      }));
      setOpenZoneComboboxTemplateId((current) =>
        current === templateId ? null : current
      );
    },
    []
  );

  const handleInstantiate = async (templateId: string) => {
    const districtId = selectedDistrictByTemplate[templateId];
    if (!districtId) {
      setActionError('Choose a district before instantiating a live chain.');
      return;
    }

    setInstantiatingRunKey(`district:${templateId}`);
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
      setInstantiatingRunKey(null);
    }
  };

  const handleInstantiateZone = async (
    templateId: string,
    districtId: string,
    zoneId: string
  ) => {
    if (!districtId) {
      setActionError(
        'Choose a district before instantiating a zone-focused chain.'
      );
      return;
    }
    if (!zoneId) {
      setActionError(
        'Choose a zone before instantiating a zone-focused chain.'
      );
      return;
    }

    setInstantiatingRunKey(`zone:${templateId}`);
    setActionError(null);
    try {
      const created = await apiClient.post<MainStoryDistrictRun>(
        `/sonar/mainStoryTemplates/${templateId}/zoneRuns`,
        { districtId, zoneId }
      );
      setRuns((current) => [
        created,
        ...current.filter((run) => run.id !== created.id),
      ]);
      if (created.status === 'failed' && created.errorMessage) {
        setActionError(created.errorMessage);
      }
    } catch (error) {
      console.error('Failed to instantiate main story zone run', error);
      setActionError(
        extractApiErrorMessage(
          error,
          'Failed to instantiate that main story into the selected zone.'
        )
      );
    } finally {
      setInstantiatingRunKey(null);
    }
  };

  const handleRetryRun = async (run: MainStoryDistrictRun) => {
    const confirmed = window.confirm(
      'Retry this run from its first failed beat onward? Completed earlier beats will be kept, and the failed beat will be re-attempted in a different zone first when possible.'
    );
    if (!confirmed) {
      return;
    }

    setRetryingRunId(run.id);
    setActionError(null);
    try {
      const retried = await apiClient.post<MainStoryDistrictRun>(
        `/sonar/mainStoryDistrictRuns/${run.id}/retry`,
        {}
      );
      setRuns((current) =>
        current.map((candidate) =>
          candidate.id === retried.id ? retried : candidate
        )
      );
      if (retried.status === 'failed' && retried.errorMessage) {
        setActionError(retried.errorMessage);
      }
    } catch (error) {
      console.error('Failed to retry main story district run', error);
      setActionError(
        extractApiErrorMessage(error, 'Failed to retry that district run.')
      );
    } finally {
      setRetryingRunId(null);
    }
  };

  const handleDeleteRun = async (run: MainStoryDistrictRun) => {
    const confirmed = window.confirm(
      'Clean up this run and delete its live quests plus generated questgiver characters?'
    );
    if (!confirmed) {
      return;
    }

    setDeletingRunId(run.id);
    setActionError(null);
    try {
      await apiClient.delete(`/sonar/mainStoryDistrictRuns/${run.id}`);
      setRuns((current) =>
        current.filter((candidate) => candidate.id !== run.id)
      );
    } catch (error) {
      console.error('Failed to delete main story district run', error);
      setActionError(
        extractApiErrorMessage(error, 'Failed to clean up that district run.')
      );
    } finally {
      setDeletingRunId(null);
    }
  };

  const handleDeleteTemplate = async (
    template: MainStoryTemplate,
    templateRuns: MainStoryDistrictRun[]
  ) => {
    if (templateRuns.length > 0) {
      setActionError(
        'Clean up this template’s live runs before deleting the template itself.'
      );
      return;
    }

    const confirmed = window.confirm(
      'Delete this main story template and its converted artifacts? This will remove the template, its generated quest archetypes, linked story-world changes, generated recurring cast, and reset the source draft so it can be converted again later.'
    );
    if (!confirmed) {
      return;
    }

    setDeletingTemplateId(template.id);
    setActionError(null);
    try {
      await apiClient.delete(`/sonar/mainStoryTemplates/${template.id}`);
      setTemplates((current) =>
        current.filter((candidate) => candidate.id !== template.id)
      );
      setRuns((current) =>
        current.filter((run) => run.mainStoryTemplateId !== template.id)
      );
    } catch (error) {
      console.error('Failed to delete main story template', error);
      setActionError(
        extractApiErrorMessage(
          error,
          'Failed to delete that main story template.'
        )
      );
    } finally {
      setDeletingTemplateId(null);
    }
  };

  const handleSaveTemplate = async (template: MainStoryTemplate) => {
    setSavingTemplate(true);
    setActionError(null);
    try {
      const saved = template.id
        ? await apiClient.patch<MainStoryTemplate>(
            `/sonar/mainStoryTemplates/${template.id}`,
            template
          )
        : await apiClient.post<MainStoryTemplate>(
            '/sonar/mainStoryTemplates',
            template
          );
      setTemplates((current) => {
        const remaining = current.filter((entry) => entry.id !== saved.id);
        return [saved, ...remaining];
      });
      setEditingTemplate(null);
    } catch (error) {
      console.error('Failed to save main story template', error);
      setActionError(
        extractApiErrorMessage(
          error,
          'Failed to save that main story template.'
        )
      );
    } finally {
      setSavingTemplate(false);
    }
  };

  const renderRunCards = (
    template: MainStoryTemplate,
    templateRuns: MainStoryDistrictRun[]
  ) => {
    if (templateRuns.length === 0) {
      return null;
    }

    return (
      <div className="qa-grid">
        {templateRuns.map((run) => {
          const district = districtsById.get(run.districtId);
          const zone =
            run.zoneId && district
              ? district.zones.find(
                  (candidate) => candidate.id === run.zoneId
                ) ?? null
              : null;
          const isZoneRun = Boolean(run.zoneId);
          return (
            <div className="qa-node-card" key={run.id}>
              <div className="qa-card-header">
                <div>
                  <div className="qa-node-title">
                    {zone?.name ??
                      district?.name ??
                      (isZoneRun ? 'Unknown zone' : 'Unknown district')}
                  </div>
                  <div className="qa-meta">
                    {[
                      isZoneRun
                        ? district?.name ?? 'Unknown district'
                        : 'District-wide',
                      isZoneRun ? 'Zone-focused' : null,
                      `Started ${formatDate(run.createdAt)}`,
                      `Updated ${formatDate(run.updatedAt)}`,
                    ]
                      .filter(Boolean)
                      .join(' • ')}
                  </div>
                </div>
                <div className="qa-actions">
                  <span className="qa-chip muted">
                    {isZoneRun ? 'Zone Run' : 'District Run'}
                  </span>
                  <span className={statusChipClass(run.status)}>
                    {run.status.replace(/_/g, ' ')}
                  </span>
                  <span className="qa-chip muted">
                    {run.beatRuns?.filter((beat) => beat.status === 'completed')
                      .length ?? 0}
                    /{template.beats.length} beats
                  </span>
                </div>
              </div>
              {run.errorMessage ? (
                <div className="qa-meta" style={{ color: 'var(--qa-danger)' }}>
                  {run.errorMessage}
                </div>
              ) : null}
              <div className="qa-actions" style={{ marginTop: 14 }}>
                {run.status === 'failed' ? (
                  <button
                    type="button"
                    className="qa-btn qa-btn-primary"
                    onClick={() => void handleRetryRun(run)}
                    disabled={retryingRunId === run.id}
                  >
                    {retryingRunId === run.id ? 'Retrying...' : 'Retry Run'}
                  </button>
                ) : null}
                <Link
                  to={`/main-story-district-runs/${run.id}`}
                  className="qa-btn qa-btn-outline"
                >
                  View Run
                </Link>
                <button
                  type="button"
                  className="qa-btn qa-btn-danger"
                  onClick={() => void handleDeleteRun(run)}
                  disabled={deletingRunId === run.id}
                >
                  {deletingRunId === run.id ? 'Cleaning Up...' : 'Clean Up Run'}
                </button>
              </div>
              <div className="qa-tree" style={{ marginTop: 14 }}>
                {(run.beatRuns || []).map((beatRun) => (
                  <div
                    className="qa-node"
                    key={`${run.id}-${beatRun.orderIndex}`}
                  >
                    <div className="qa-node-card">
                      <div className="qa-card-header">
                        <div>
                          <div className="qa-node-title">
                            Beat {beatRun.orderIndex}: {beatRun.chapterTitle}
                          </div>
                          <div className="qa-meta">
                            {[beatRun.zoneName, beatRun.pointOfInterestName]
                              .filter(Boolean)
                              .join(' • ') || 'Placement pending'}
                          </div>
                        </div>
                        <div className="qa-actions">
                          <span className={statusChipClass(beatRun.status)}>
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
    );
  };

  return (
    <div className="qa-theme">
      <div className="qa-shell">
        <section className="qa-hero">
          <div>
            <div className="qa-kicker">Main Story Templates</div>
            <h1 className="qa-title">Converted Campaign Templates</h1>
            <p className="qa-subtitle">
              Review converted campaign templates, inspect recent live runs, and
              start building district-wide or zone-focused main-story quest
              chains from reusable story blueprints.
            </p>
          </div>
          <div className="qa-hero-actions">
            <Link to="/main-story-generator" className="qa-btn qa-btn-outline">
              Back to Generator
            </Link>
            <button
              type="button"
              className="qa-btn qa-btn-outline"
              onClick={() => setEditingTemplate('new')}
            >
              New Template
            </button>
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
              <h2 className="qa-card-title">Live Main Story Runs</h2>
              <div className="qa-meta">
                First pass: instantiation clones the recurring cast, places
                questgivers into POIs, and creates a linked main-story quest
                chain for the selected district or pinned zone.
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
              const availableZones = selectedDistrict?.zones ?? [];
              const selectedZone = resolveSelectedZone(
                template.id,
                selectedDistrict
              );
              const zoneQuery =
                zoneQueryByTemplate[template.id] ?? selectedZone?.name ?? '';
              const filteredZones = availableZones.filter((zone) =>
                zone.name.toLowerCase().includes(zoneQuery.trim().toLowerCase())
              );
              const districtRuns = templateRuns.filter((run) => !run.zoneId);
              const zoneRuns = templateRuns.filter((run) =>
                Boolean(run.zoneId)
              );
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
                      <button
                        type="button"
                        className="qa-btn qa-btn-outline"
                        onClick={() => setEditingTemplate(template)}
                      >
                        Edit Template
                      </button>
                      <button
                        type="button"
                        className="qa-btn qa-btn-danger"
                        onClick={() =>
                          void handleDeleteTemplate(template, templateRuns)
                        }
                        disabled={
                          deletingTemplateId === template.id ||
                          templateRuns.length > 0
                        }
                        title={
                          templateRuns.length > 0
                            ? 'Clean up all live runs for this template before deleting it.'
                            : undefined
                        }
                      >
                        {deletingTemplateId === template.id
                          ? 'Deleting...'
                          : 'Delete Template'}
                      </button>
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
                        onChange={(event) => {
                          const nextDistrictId = event.target.value;
                          const nextDistrict =
                            districtsById.get(nextDistrictId) ?? null;
                          setSelectedDistrictByTemplate((current) => ({
                            ...current,
                            [template.id]: nextDistrictId,
                          }));
                          selectZoneForTemplate(
                            template.id,
                            nextDistrict?.zones?.[0] ?? null
                          );
                        }}
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
                          instantiatingRunKey === `district:${template.id}`
                        }
                      >
                        {instantiatingRunKey === `district:${template.id}`
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
                        Instantiate in Zone
                      </h3>
                      <div className="qa-meta">
                        This pins every beat in the chain to one zone.
                      </div>
                    </div>
                    <div className="qa-actions">
                      <div
                        className="qa-combobox"
                        style={{ minWidth: 240, flex: '1 1 260px' }}
                      >
                        <input
                          type="text"
                          className="qa-input"
                          value={zoneQuery}
                          onFocus={() =>
                            setOpenZoneComboboxTemplateId(template.id)
                          }
                          onBlur={() => {
                            window.setTimeout(() => {
                              setOpenZoneComboboxTemplateId((current) =>
                                current === template.id ? null : current
                              );
                            }, 120);
                          }}
                          onChange={(event) => {
                            const value = event.target.value;
                            const matchedZone = availableZones.find(
                              (zone) =>
                                zone.name.toLowerCase() ===
                                value.trim().toLowerCase()
                            );
                            setZoneQueryByTemplate((current) => ({
                              ...current,
                              [template.id]: value,
                            }));
                            setSelectedZoneByTemplate((current) => ({
                              ...current,
                              [template.id]: matchedZone?.id ?? '',
                            }));
                            setOpenZoneComboboxTemplateId(template.id);
                          }}
                          placeholder={
                            !selectedDistrict
                              ? 'Choose a district first'
                              : availableZones.length === 0
                                ? 'No zones in district'
                                : 'Search zones...'
                          }
                          disabled={
                            !selectedDistrict || availableZones.length === 0
                          }
                        />
                        {openZoneComboboxTemplateId === template.id &&
                        selectedDistrict ? (
                          <div className="qa-combobox-list">
                            {filteredZones.length === 0 ? (
                              <div className="qa-combobox-empty">
                                No matching zones.
                              </div>
                            ) : (
                              filteredZones.map((zone) => (
                                <button
                                  key={`${template.id}-zone-${zone.id}`}
                                  type="button"
                                  className={`qa-combobox-option ${
                                    selectedZone?.id === zone.id
                                      ? 'is-active'
                                      : ''
                                  }`}
                                  onMouseDown={(event) => {
                                    event.preventDefault();
                                    selectZoneForTemplate(template.id, zone);
                                  }}
                                >
                                  {zone.name}
                                </button>
                              ))
                            )}
                          </div>
                        ) : null}
                      </div>
                      <button
                        type="button"
                        className="qa-btn qa-btn-primary"
                        onClick={() =>
                          void handleInstantiateZone(
                            template.id,
                            selectedDistrictId,
                            selectedZone?.id ?? ''
                          )
                        }
                        disabled={
                          !selectedDistrictId ||
                          !selectedZone ||
                          instantiatingRunKey === `zone:${template.id}`
                        }
                      >
                        {instantiatingRunKey === `zone:${template.id}`
                          ? 'Instantiating...'
                          : selectedZone
                            ? `Instantiate in ${selectedZone.name}`
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
                    {template.beats.map((beat) => {
                      const beatKey = `${template.id}-${beat.orderIndex}`;
                      const isExpanded = expandedBeatKeys.includes(beatKey);
                      const questArchetype = beat.questArchetypeId
                        ? questArchetypesById.get(beat.questArchetypeId) ?? null
                        : null;
                      return (
                        <div className="qa-node" key={beatKey}>
                          <div
                            className="qa-node-card"
                            style={{
                              background: isExpanded
                                ? 'linear-gradient(180deg, rgba(18, 30, 37, 0.92), rgba(10, 18, 23, 0.92))'
                                : undefined,
                            }}
                          >
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
                                <button
                                  type="button"
                                  className="qa-btn qa-btn-outline"
                                  onClick={() => toggleBeatExpanded(beatKey)}
                                >
                                  {isExpanded
                                    ? 'Hide Archetype Details'
                                    : 'Show Archetype Details'}
                                </button>
                              </div>
                            </div>
                            {isExpanded && (
                              <div style={{ marginTop: 16 }}>
                                <div
                                  style={{
                                    display: 'grid',
                                    gridTemplateColumns:
                                      'repeat(auto-fit, minmax(180px, 1fr))',
                                    gap: 12,
                                  }}
                                >
                                  <div className="qa-stat">
                                    <div className="qa-stat-label">
                                      Archetype
                                    </div>
                                    <div className="qa-stat-value">
                                      {questArchetype?.name ||
                                        beat.questArchetypeName ||
                                        'Not resolved'}
                                    </div>
                                  </div>
                                  <div className="qa-stat">
                                    <div className="qa-stat-label">
                                      Category
                                    </div>
                                    <div className="qa-stat-value">
                                      {questArchetype?.category || 'n/a'}
                                    </div>
                                  </div>
                                  <div className="qa-stat">
                                    <div className="qa-stat-label">Flow</div>
                                    <div className="qa-stat-value">
                                      {summarizeQuestArchetypeRoot(
                                        questArchetype
                                      )}
                                    </div>
                                  </div>
                                  <div className="qa-stat">
                                    <div className="qa-stat-label">Scale</div>
                                    <div className="qa-stat-value">
                                      {countQuestArchetypeNodes(questArchetype)}{' '}
                                      nodes ·{' '}
                                      {countQuestArchetypeChallenges(
                                        questArchetype
                                      )}{' '}
                                      challenges
                                    </div>
                                  </div>
                                </div>

                                {questArchetype?.description && (
                                  <div
                                    className="qa-meta"
                                    style={{ marginTop: 14 }}
                                  >
                                    {questArchetype.description}
                                  </div>
                                )}

                                {questArchetype &&
                                ((questArchetype.acceptanceDialogue ?? [])
                                  .length > 0 ||
                                  (questArchetype.characterTags ?? []).length >
                                    0 ||
                                  (questArchetype.internalTags ?? []).length >
                                    0) ? (
                                  <div
                                    style={{
                                      display: 'grid',
                                      gap: 10,
                                      marginTop: 14,
                                    }}
                                  >
                                    {(questArchetype.acceptanceDialogue ?? [])
                                      .length > 0 && (
                                      <div className="qa-meta">
                                        <strong>Acceptance Dialogue:</strong>{' '}
                                        {(
                                          questArchetype.acceptanceDialogue ??
                                          []
                                        )
                                          .map((line) =>
                                            line.effect
                                              ? `${line.text} (${line.effect})`
                                              : line.text
                                          )
                                          .join(' / ')}
                                      </div>
                                    )}
                                    {(questArchetype.characterTags ?? [])
                                      .length > 0 && (
                                      <div className="qa-meta">
                                        <strong>Character Tags:</strong>{' '}
                                        {(
                                          questArchetype.characterTags ?? []
                                        ).join(', ')}
                                      </div>
                                    )}
                                    {(questArchetype.internalTags ?? [])
                                      .length > 0 && (
                                      <div className="qa-meta">
                                        <strong>Internal Tags:</strong>{' '}
                                        {(
                                          questArchetype.internalTags ?? []
                                        ).join(', ')}
                                      </div>
                                    )}
                                  </div>
                                ) : null}

                                {questArchetype?.root ? (
                                  <>
                                    <div
                                      className="qa-divider"
                                      style={{ margin: '16px 0' }}
                                    />
                                    <div className="qa-card-header">
                                      <div>
                                        <h4
                                          className="qa-card-title"
                                          style={{ fontSize: 16 }}
                                        >
                                          Quest Flow
                                        </h4>
                                        <div className="qa-meta">
                                          Every node and unlocked branch in this
                                          beat’s quest archetype.
                                        </div>
                                      </div>
                                    </div>
                                    {renderQuestArchetypeNode(
                                      questArchetype.root
                                    )}
                                  </>
                                ) : null}
                              </div>
                            )}
                          </div>
                        </div>
                      );
                    })}
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

                  {districtRuns.length === 0 ? (
                    <div className="qa-meta">
                      No district runs yet for this template.
                    </div>
                  ) : (
                    renderRunCards(template, districtRuns)
                  )}

                  <div className="qa-divider" />

                  <div className="qa-card-header">
                    <div>
                      <h3 className="qa-card-title" style={{ fontSize: 18 }}>
                        Zone Runs
                      </h3>
                      <div className="qa-meta">
                        Recent attempts to pin every beat from this template to
                        a single zone.
                      </div>
                    </div>
                  </div>

                  {zoneRuns.length === 0 ? (
                    <div className="qa-meta">
                      No zone runs yet for this template.
                    </div>
                  ) : (
                    renderRunCards(template, zoneRuns)
                  )}
                </article>
              );
            })
          )}
        </section>
      </div>
      {editingTemplate !== null && (
        <MainStoryTemplateEditor
          initialTemplate={editingTemplate === 'new' ? null : editingTemplate}
          questArchetypes={questArchetypes}
          locationArchetypes={locationArchetypes}
          saving={savingTemplate}
          onCancel={() => setEditingTemplate(null)}
          onSave={handleSaveTemplate}
        />
      )}
    </div>
  );
};

export default MainStoryTemplates;
