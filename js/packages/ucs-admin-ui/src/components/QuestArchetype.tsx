import React, { useMemo, useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';
import {
  QuestArchetypeDraft,
  QuestTemplateGeneratorDraft,
  useQuestArchetypes,
} from '../contexts/questArchetypes.tsx';
import {
  LocationArchetype,
  QuestArchetype,
  QuestArchetypeNode,
  QuestArchetypeChallenge,
  QuestArchetypeNodeEncounterItemReward,
  QuestArchetypeNodeType,
  InventoryItem,
  Spell,
} from '@poltergeist/types';
import {
  MaterialRewardsEditor,
  emptyMaterialReward,
  normalizeMaterialRewards,
  summarizeMaterialRewards,
} from './MaterialRewardsEditor.tsx';
import './questArchetypeTheme.css';

interface FlowNodeProps {
  node: QuestArchetypeNode;
  locationArchetypes: LocationArchetype[];
  monsterTemplates: MonsterTemplateRecord[];
  scenarioTemplates: ScenarioTemplateRecord[];
  inventoryItems: InventoryItem[];
  depth: number;
  proficiencyOptions: string[];
  onProficiencySearchChange: (value: string) => void;
  addChallengeToQuestArchetype: (
    questArchetypeId: string,
    rewardPoints: number,
    inventoryItemId?: number | null,
    proficiency?: string | null,
    difficulty?: number | null,
    unlockedNode?: QuestArchetypeNodeDraft | null
  ) => void;
  onSaveNode: (nodeId: string, updates: QuestArchetypeNodeDraft) => void;
  onEditChallenge: (challenge: QuestArchetypeChallenge) => void;
}

type MonsterTemplateRecord = {
  id: string;
  name: string;
  monsterType?: 'monster' | 'boss' | 'raid';
  imageUrl?: string;
  thumbnailUrl?: string;
};

type ScenarioTemplateRecord = {
  id: string;
  prompt: string;
};

type PaginatedResponse<T> = {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
};

type NodeRewardMode = 'explicit' | 'random';
type QuestArchetypeNodeEditorState = {
  nodeType: QuestArchetypeNodeType;
  locationArchetypeId: string;
  locationArchetypeQuery: string;
  scenarioTemplateId: string;
  monsterTemplateIds: string[];
  targetLevel: number;
  encounterRewardMode: NodeRewardMode;
  encounterRandomRewardSize: RandomRewardSize;
  encounterRewardExperience: number;
  encounterRewardGold: number;
  encounterProximityMeters: number;
  encounterMaterialRewards: ReturnType<typeof emptyMaterialReward>[];
  encounterItemRewards: Array<{ inventoryItemId: string; quantity: number }>;
  difficulty: number;
};

const emptyNodeEditorState = (): QuestArchetypeNodeEditorState => ({
  nodeType: 'location',
  locationArchetypeId: '',
  locationArchetypeQuery: '',
  scenarioTemplateId: '',
  monsterTemplateIds: [],
  targetLevel: 1,
  encounterRewardMode: 'random',
  encounterRandomRewardSize: 'small',
  encounterRewardExperience: 0,
  encounterRewardGold: 0,
  encounterProximityMeters: 100,
  encounterMaterialRewards: [],
  encounterItemRewards: [],
  difficulty: 0,
});

const buildNodeEditorState = (
  node: QuestArchetypeNode,
  locationArchetypes: LocationArchetype[]
): QuestArchetypeNodeEditorState => ({
  nodeType:
    node.nodeType === 'monster_encounter'
      ? 'monster_encounter'
      : node.nodeType === 'scenario'
        ? 'scenario'
        : 'location',
  scenarioTemplateId: node.scenarioTemplateId ?? '',
  locationArchetypeId: node.locationArchetypeId ?? '',
  locationArchetypeQuery:
    locationArchetypes.find((entry) => entry.id === node.locationArchetypeId)
      ?.name ?? '',
  monsterTemplateIds: [...(node.monsterTemplateIds ?? [])],
  targetLevel: node.targetLevel ?? 1,
  encounterRewardMode:
    node.encounterRewardMode === 'explicit' ? 'explicit' : 'random',
  encounterRandomRewardSize:
    node.encounterRandomRewardSize === 'medium' ||
    node.encounterRandomRewardSize === 'large'
      ? node.encounterRandomRewardSize
      : 'small',
  encounterRewardExperience: node.encounterRewardExperience ?? 0,
  encounterRewardGold: node.encounterRewardGold ?? 0,
  encounterProximityMeters: node.encounterProximityMeters ?? 100,
  encounterMaterialRewards: (node.encounterMaterialRewards ?? []).map(
    (reward) => ({
      resourceKey: reward.resourceKey,
      amount: reward.amount,
    })
  ),
  encounterItemRewards: (node.encounterItemRewards ?? []).map((reward) => ({
    inventoryItemId: reward.inventoryItemId
      ? String(reward.inventoryItemId)
      : '',
    quantity: reward.quantity ?? 1,
  })),
  difficulty: node.difficulty ?? 0,
});

const buildNodeDraft = (
  state: QuestArchetypeNodeEditorState
): QuestArchetypeNodeDraft => ({
  nodeType: state.nodeType,
  locationArchetypeId:
    state.nodeType === 'location' || state.locationArchetypeId
      ? state.locationArchetypeId || null
      : null,
  scenarioTemplateId:
    state.nodeType === 'scenario' ? state.scenarioTemplateId || null : null,
  monsterTemplateIds:
    state.nodeType === 'monster_encounter'
      ? state.monsterTemplateIds
      : undefined,
  targetLevel:
    state.nodeType === 'monster_encounter' ? Number(state.targetLevel) || 1 : undefined,
  encounterRewardMode:
    state.nodeType === 'monster_encounter' ? state.encounterRewardMode : undefined,
  encounterRandomRewardSize:
    state.nodeType === 'monster_encounter'
      ? state.encounterRandomRewardSize
      : undefined,
  encounterRewardExperience:
    state.nodeType === 'monster_encounter'
      ? Number(state.encounterRewardExperience) || 0
      : undefined,
  encounterRewardGold:
    state.nodeType === 'monster_encounter'
      ? Number(state.encounterRewardGold) || 0
      : undefined,
  encounterProximityMeters:
    state.nodeType === 'monster_encounter' || state.nodeType === 'scenario'
      ? Number(state.encounterProximityMeters) || 0
      : undefined,
  encounterMaterialRewards:
    state.nodeType === 'monster_encounter'
      ? normalizeMaterialRewards(state.encounterMaterialRewards)
      : undefined,
  encounterItemRewards:
    state.nodeType === 'monster_encounter'
      ? state.encounterItemRewards
          .map((reward) => ({
            inventoryItemId: Number(reward.inventoryItemId) || 0,
            quantity: Number(reward.quantity) || 0,
          }))
          .filter(
            (reward): reward is QuestArchetypeNodeEncounterItemReward =>
              reward.inventoryItemId > 0 && reward.quantity > 0
          )
      : undefined,
  difficulty: Number(state.difficulty) || 0,
});

const describeQuestArchetypeNode = (
  node: QuestArchetypeNode | undefined | null,
  locationArchetypes: LocationArchetype[],
  monsterTemplates: MonsterTemplateRecord[],
  scenarioTemplates: ScenarioTemplateRecord[]
) => {
  if (!node) {
    return 'Unknown';
  }
  if (node.nodeType === 'scenario') {
    const scenarioLabel =
      scenarioTemplates.find((entry) => entry.id === node.scenarioTemplateId)
        ?.prompt ?? 'Scenario';
    const locationLabel = locationArchetypes.find(
      (entry) => entry.id === node.locationArchetypeId
    )?.name;
    return locationLabel ? `${scenarioLabel} @ ${locationLabel}` : scenarioLabel;
  }
  if (node.nodeType === 'monster_encounter') {
    const names = (node.monsterTemplateIds ?? [])
      .map(
        (templateId) =>
          monsterTemplates.find((entry) => entry.id === templateId)?.name
      )
      .filter(Boolean) as string[];
    if (names.length === 0) {
      const locationLabel = locationArchetypes.find(
        (entry) => entry.id === node.locationArchetypeId
      )?.name;
      return locationLabel ? `Monster encounter @ ${locationLabel}` : 'Monster encounter';
    }
    const encounterLabel = `Encounter: ${names.slice(0, 3).join(', ')}${names.length > 3 ? '…' : ''}`;
    const locationLabel = locationArchetypes.find(
      (entry) => entry.id === node.locationArchetypeId
    )?.name;
    return locationLabel ? `${encounterLabel} @ ${locationLabel}` : encounterLabel;
  }
  return (
    locationArchetypes.find((la) => la.id === node.locationArchetypeId)?.name ??
    'Unknown location'
  );
};

const FlowNode: React.FC<FlowNodeProps> = ({
  node,
  locationArchetypes,
  monsterTemplates,
  scenarioTemplates,
  inventoryItems,
  depth,
  proficiencyOptions,
  onProficiencySearchChange,
  addChallengeToQuestArchetype,
  onSaveNode,
  onEditChallenge,
}) => {
  const borderColor =
    depth % 2 === 0 ? 'rgba(255, 107, 74, 0.4)' : 'rgba(95, 211, 181, 0.35)';
  const nodeSummary = describeQuestArchetypeNode(
    node,
    locationArchetypes,
    monsterTemplates,
    scenarioTemplates
  );
  const isEncounterNode = node.nodeType === 'monster_encounter';
  const isScenarioNode = node.nodeType === 'scenario';
  const isBranchOnlyNode = isEncounterNode || isScenarioNode;
  const [isAdding, setIsAdding] = useState(false);
  const [rewardPoints, setRewardPoints] = useState<number>(0);
  const [rewardItemId, setRewardItemId] = useState<number>(0);
  const [challengeDifficulty, setChallengeDifficulty] = useState<number>(0);
  const [challengeProficiency, setChallengeProficiency] = useState<string>('');
  const [nodeEditor, setNodeEditor] = useState<QuestArchetypeNodeEditorState>(
    buildNodeEditorState(node, locationArchetypes)
  );
  const [childEditor, setChildEditor] = useState<QuestArchetypeNodeEditorState>(
    emptyNodeEditorState()
  );
  const [childEnabled, setChildEnabled] = useState<boolean>(false);

  useEffect(() => {
    setNodeEditor(buildNodeEditorState(node, locationArchetypes));
  }, [locationArchetypes, node]);

  const renderNodeConfigFields = (
    editor: QuestArchetypeNodeEditorState,
    setEditor: React.Dispatch<React.SetStateAction<QuestArchetypeNodeEditorState>>,
    prefix: string
  ) => {
    const filteredLocationArchetypes = locationArchetypes
      .filter((archetype) =>
        archetype.name
          .toLowerCase()
          .includes(editor.locationArchetypeQuery.trim().toLowerCase())
      )
      .slice(0, 8);

    return (
      <>
        <div className="qa-field">
          <div className="qa-label">{prefix} Node Type</div>
          <select
            className="qa-select"
            value={editor.nodeType}
            onChange={(e) =>
              setEditor((prev) => ({
                ...prev,
                nodeType: e.target.value as QuestArchetypeNodeType,
              }))
            }
          >
            <option value="location">Location</option>
            <option value="scenario">Scenario</option>
            <option value="monster_encounter">Monster Encounter</option>
          </select>
        </div>
        {editor.nodeType === 'location' ? (
          <div className="qa-field">
            <div className="qa-label">{prefix} Location Archetype</div>
            <div className="qa-combobox">
              <input
                type="text"
                className="qa-input"
                value={editor.locationArchetypeQuery}
                onChange={(e) => {
                  const value = e.target.value;
                  const matched = locationArchetypes.find(
                    (archetype) =>
                      archetype.name.toLowerCase() === value.trim().toLowerCase()
                  );
                  setEditor((prev) => ({
                    ...prev,
                    locationArchetypeQuery: value,
                    locationArchetypeId: matched ? matched.id : '',
                  }));
                }}
                placeholder="Search location archetypes..."
              />
              {editor.locationArchetypeQuery.trim().length > 0 && (
                <div className="qa-combobox-list">
                  {filteredLocationArchetypes.length === 0 ? (
                    <div className="qa-combobox-empty">No matches.</div>
                  ) : (
                    filteredLocationArchetypes.map((archetype) => (
                      <button
                        key={`${prefix}-${archetype.id}`}
                        type="button"
                        className="qa-combobox-option"
                        onClick={() =>
                          setEditor((prev) => ({
                            ...prev,
                            locationArchetypeId: archetype.id,
                            locationArchetypeQuery: archetype.name,
                          }))
                        }
                      >
                        {archetype.name}
                      </button>
                    ))
                  )}
                </div>
              )}
            </div>
          </div>
        ) : editor.nodeType === 'scenario' ? (
          <>
            <div className="qa-field">
              <div className="qa-label">{prefix} Scenario Template</div>
              <select
                className="qa-select"
                value={editor.scenarioTemplateId}
                onChange={(e) =>
                  setEditor((prev) => ({
                    ...prev,
                    scenarioTemplateId: e.target.value,
                  }))
                }
              >
                <option value="">Select a scenario template</option>
                {scenarioTemplates.map((template) => (
                  <option key={`${prefix}-scenario-${template.id}`} value={template.id}>
                    {template.prompt}
                  </option>
                ))}
              </select>
            </div>
            <div className="qa-field">
              <div className="qa-label">{prefix} Proximity To Previous Node (m)</div>
              <input
                type="number"
                min={0}
                className="qa-input"
                value={editor.encounterProximityMeters}
                onChange={(e) =>
                  setEditor((prev) => ({
                    ...prev,
                    encounterProximityMeters: parseInt(e.target.value) || 0,
                  }))
                }
              />
            </div>
          </>
        ) : (
          <>
            <div className="qa-field">
              <div className="qa-label">{prefix} Monster Templates</div>
              <select
                className="qa-select"
                multiple
                size={6}
                value={editor.monsterTemplateIds}
                onChange={(e) =>
                  setEditor((prev) => ({
                    ...prev,
                    monsterTemplateIds: Array.from(e.target.selectedOptions).map(
                      (option) => option.value
                    ),
                  }))
                }
              >
                {monsterTemplates.map((template) => (
                  <option key={template.id} value={template.id}>
                    {template.name}
                    {template.monsterType ? ` (${template.monsterType})` : ''}
                  </option>
                ))}
              </select>
              <div className="qa-helper">
                Hold Command or Ctrl to select multiple monster templates.
              </div>
            </div>
            <div className="qa-field">
              <div className="qa-label">{prefix} Target Level</div>
              <input
                type="number"
                min={1}
                className="qa-input"
                value={editor.targetLevel}
                onChange={(e) =>
                  setEditor((prev) => ({
                    ...prev,
                    targetLevel: parseInt(e.target.value) || 1,
                  }))
                }
              />
            </div>
            <div className="qa-field">
              <div className="qa-label">{prefix} Proximity To Previous Node (m)</div>
              <input
                type="number"
                min={0}
                className="qa-input"
                value={editor.encounterProximityMeters}
                onChange={(e) =>
                  setEditor((prev) => ({
                    ...prev,
                    encounterProximityMeters: parseInt(e.target.value) || 0,
                  }))
                }
              />
            </div>
            <div className="qa-field">
              <div className="qa-label">{prefix} Reward Mode</div>
              <select
                className="qa-select"
                value={editor.encounterRewardMode}
                onChange={(e) =>
                  setEditor((prev) => ({
                    ...prev,
                    encounterRewardMode: e.target.value as NodeRewardMode,
                  }))
                }
              >
                <option value="random">Random</option>
                <option value="explicit">Explicit</option>
              </select>
            </div>
            <div className="qa-field">
              <div className="qa-label">{prefix} Random Reward Size</div>
              <select
                className="qa-select"
                value={editor.encounterRandomRewardSize}
                disabled={editor.encounterRewardMode !== 'random'}
                onChange={(e) =>
                  setEditor((prev) => ({
                    ...prev,
                    encounterRandomRewardSize: e.target.value as RandomRewardSize,
                  }))
                }
              >
                <option value="small">Small</option>
                <option value="medium">Medium</option>
                <option value="large">Large</option>
              </select>
            </div>
            <div className="qa-field">
              <div className="qa-label">{prefix} Reward Experience</div>
              <input
                type="number"
                min={0}
                className="qa-input"
                disabled={editor.encounterRewardMode !== 'explicit'}
                value={editor.encounterRewardExperience}
                onChange={(e) =>
                  setEditor((prev) => ({
                    ...prev,
                    encounterRewardExperience: parseInt(e.target.value) || 0,
                  }))
                }
              />
            </div>
            <div className="qa-field">
              <div className="qa-label">{prefix} Reward Gold</div>
              <input
                type="number"
                min={0}
                className="qa-input"
                disabled={editor.encounterRewardMode !== 'explicit'}
                value={editor.encounterRewardGold}
                onChange={(e) =>
                  setEditor((prev) => ({
                    ...prev,
                    encounterRewardGold: parseInt(e.target.value) || 0,
                  }))
                }
              />
            </div>
            <div className="qa-field">
              <div className="qa-label">{prefix} Material Rewards</div>
              <MaterialRewardsEditor
                value={editor.encounterMaterialRewards}
                onChange={(encounterMaterialRewards) =>
                  setEditor((prev) => ({ ...prev, encounterMaterialRewards }))
                }
                disabled={editor.encounterRewardMode !== 'explicit'}
              />
            </div>
            <div className="qa-field">
              <div className="qa-label">{prefix} Item Rewards</div>
              {editor.encounterItemRewards.length === 0 ? (
                <div className="qa-empty">No item rewards yet.</div>
              ) : (
                <div className="qa-form-grid">
                  {editor.encounterItemRewards.map((reward, index) => (
                    <div key={`${prefix}-encounter-reward-${index}`} className="qa-reward-row">
                      <select
                        className="qa-select"
                        value={reward.inventoryItemId}
                        disabled={editor.encounterRewardMode !== 'explicit'}
                        onChange={(e) =>
                          setEditor((prev) => ({
                            ...prev,
                            encounterItemRewards: prev.encounterItemRewards.map(
                              (entry, rewardIndex) =>
                                rewardIndex === index
                                  ? { ...entry, inventoryItemId: e.target.value }
                                  : entry
                            ),
                          }))
                        }
                      >
                        <option value="">Select an item</option>
                        {inventoryItems.map((item) => (
                          <option key={`${prefix}-item-${item.id}`} value={item.id}>
                            {item.name}
                          </option>
                        ))}
                      </select>
                      <input
                        type="number"
                        min={1}
                        className="qa-input"
                        disabled={editor.encounterRewardMode !== 'explicit'}
                        value={reward.quantity}
                        onChange={(e) =>
                          setEditor((prev) => ({
                            ...prev,
                            encounterItemRewards: prev.encounterItemRewards.map(
                              (entry, rewardIndex) =>
                                rewardIndex === index
                                  ? {
                                      ...entry,
                                      quantity: parseInt(e.target.value) || 1,
                                    }
                                  : entry
                            ),
                          }))
                        }
                      />
                      <button
                        type="button"
                        className="qa-btn qa-btn-text"
                        onClick={() =>
                          setEditor((prev) => ({
                            ...prev,
                            encounterItemRewards: prev.encounterItemRewards.filter(
                              (_, rewardIndex) => rewardIndex !== index
                            ),
                          }))
                        }
                      >
                        Remove
                      </button>
                    </div>
                  ))}
                </div>
              )}
              <button
                type="button"
                className="qa-btn qa-btn-ghost"
                onClick={() =>
                  setEditor((prev) => ({
                    ...prev,
                    encounterItemRewards: [
                      ...prev.encounterItemRewards,
                      { inventoryItemId: '', quantity: 1 },
                    ],
                  }))
                }
              >
                Add Item Reward
              </button>
            </div>
          </>
        )}
        <div className="qa-field">
          <div className="qa-label">{prefix} Difficulty</div>
          <input
            type="number"
            min={0}
            className="qa-input"
            value={editor.difficulty}
            onChange={(e) =>
              setEditor((prev) => ({
                ...prev,
                difficulty: parseInt(e.target.value) || 0,
              }))
            }
          />
        </div>
      </>
    );
  };

  return (
    <div className="qa-flow-node" style={{ borderColor }}>
      <div className="qa-flow-node-card">
        <div className="qa-flow-node-header">
          <div>
            <div className="qa-flow-node-title">
              {depth === 0 ? 'Root Node' : `Node ${depth + 1}`}
            </div>
            <div className="qa-meta">{nodeSummary}</div>
          </div>
          <button
            className="qa-btn qa-btn-primary"
            onClick={() => setIsAdding((prev) => !prev)}
          >
            {isAdding ? 'Close' : isBranchOnlyNode ? 'Add Child Node' : 'Add Challenge'}
          </button>
        </div>
        <div className="qa-flow-form" style={{ marginTop: 12 }}>
          {renderNodeConfigFields(nodeEditor, setNodeEditor, 'Current')}
          <div className="qa-flow-form-actions">
            <button
              className="qa-btn qa-btn-outline"
              onClick={() =>
                setNodeEditor(buildNodeEditorState(node, locationArchetypes))
              }
            >
              Reset
            </button>
            <button
              className="qa-btn qa-btn-primary"
              onClick={() => onSaveNode(node.id, buildNodeDraft(nodeEditor))}
            >
              Save Node
            </button>
          </div>
        </div>

        {isAdding && (
          <div className="qa-flow-form">
            {!isBranchOnlyNode && (
              <>
                <div className="qa-field">
                  <div className="qa-label">Reward Points</div>
                  <input
                    type="number"
                    min={0}
                    className="qa-input"
                    value={rewardPoints}
                    onChange={(e) =>
                      setRewardPoints(parseInt(e.target.value) || 0)
                    }
                  />
                </div>
                <div className="qa-field">
                  <div className="qa-label">Difficulty</div>
                  <input
                    type="number"
                    min={0}
                    className="qa-input"
                    value={challengeDifficulty}
                    onChange={(e) =>
                      setChallengeDifficulty(parseInt(e.target.value) || 0)
                    }
                  />
                </div>
                <div className="qa-field">
                  <div className="qa-label">Reward Item</div>
                  <select
                    className="qa-select"
                    value={rewardItemId || ''}
                    onChange={(e) =>
                      setRewardItemId(parseInt(e.target.value) || 0)
                    }
                  >
                    <option value="">Select an item</option>
                    {inventoryItems.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.name}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="qa-field">
                  <div className="qa-label">Proficiency</div>
                  <input
                    type="text"
                    className="qa-input"
                    value={challengeProficiency}
                    onChange={(e) => {
                      setChallengeProficiency(e.target.value);
                      onProficiencySearchChange(e.target.value);
                    }}
                    list="qa-proficiency-options"
                    placeholder="Optional proficiency (e.g. Persuasion)"
                  />
                  {proficiencyOptions.length === 0 && (
                    <div className="qa-helper">No matching proficiencies yet.</div>
                  )}
                </div>
              </>
            )}
            <div className="qa-field">
              <label className="qa-inline" style={{ alignItems: 'center' }}>
                <input
                  type="checkbox"
                  checked={childEnabled}
                  onChange={(e) => setChildEnabled(e.target.checked)}
                />
                <span className="qa-label" style={{ marginBottom: 0 }}>
                  Unlock a child node
                </span>
              </label>
            </div>
            {childEnabled && renderNodeConfigFields(childEditor, setChildEditor, 'Child')}
            <div className="qa-flow-form-actions">
              <button
                className="qa-btn qa-btn-outline"
                onClick={() => {
                  setIsAdding(false);
                }}
              >
                Cancel
              </button>
              <button
                className="qa-btn qa-btn-primary"
                onClick={async () => {
                  const trimmed = challengeProficiency.trim();
                  await addChallengeToQuestArchetype(
                    node.id,
                    isBranchOnlyNode ? 0 : rewardPoints,
                    isBranchOnlyNode ? null : rewardItemId || null,
                    isBranchOnlyNode ? null : trimmed.length > 0 ? trimmed : null,
                    isBranchOnlyNode ? 0 : challengeDifficulty,
                    childEnabled ? buildNodeDraft(childEditor) : null
                  );
                  setRewardPoints(0);
                  setRewardItemId(0);
                  setChallengeDifficulty(0);
                  setChallengeProficiency('');
                  setChildEnabled(false);
                  setChildEditor(emptyNodeEditorState());
                  setIsAdding(false);
                }}
              >
                {isBranchOnlyNode ? 'Add Child Node' : 'Add Challenge'}
              </button>
            </div>
          </div>
        )}

        {node.challenges && node.challenges.length > 0 ? (
          <div className="qa-flow-challenges">
            {node.challenges.map((challenge, index) => {
              const legacyItemId = !challenge.inventoryItemId
                ? inventoryItems?.find((item) => item.id === challenge.reward)
                    ?.id
                : undefined;
              const rewardItemId = challenge.inventoryItemId ?? legacyItemId;
              const rewardItem = rewardItemId
                ? inventoryItems?.find((item) => item.id === rewardItemId)
                : undefined;
              return (
                <div key={challenge.id} className="qa-flow-challenge-card">
                  <div className="qa-flow-challenge-header">
                    <div>
                      <div className="qa-flow-challenge-title">
                        {isBranchOnlyNode ? `Branch ${index + 1}` : `Challenge ${index + 1}`}
                      </div>
                      <div className="qa-inline" style={{ marginTop: 6 }}>
                        {!isBranchOnlyNode && challenge.reward > 0 && (
                          <span className="qa-chip accent">
                            +{challenge.reward} pts
                          </span>
                        )}
                        {!isBranchOnlyNode &&
                          challenge.difficulty !== undefined &&
                          challenge.difficulty !== null && (
                            <span className="qa-chip muted">
                              Difficulty: {challenge.difficulty}
                            </span>
                          )}
                        {!isBranchOnlyNode && rewardItem && (
                          <span className="qa-chip success">
                            {rewardItem.name}
                          </span>
                        )}
                        {!isBranchOnlyNode && challenge.proficiency && (
                          <span className="qa-chip muted">
                            Proficiency: {challenge.proficiency}
                          </span>
                        )}
                      </div>
                    </div>
                    <button
                      className="qa-btn qa-btn-ghost"
                      onClick={() => onEditChallenge(challenge)}
                    >
                      Edit
                    </button>
                  </div>

                  {challenge.unlockedNode ? (
                    <div className="qa-flow-branch">
                      <div className="qa-flow-branch-label">Unlocks</div>
                      <FlowNode
                        node={challenge.unlockedNode}
                        locationArchetypes={locationArchetypes}
                        monsterTemplates={monsterTemplates}
                        scenarioTemplates={scenarioTemplates}
                        inventoryItems={inventoryItems}
                        depth={depth + 1}
                        proficiencyOptions={proficiencyOptions}
                        onProficiencySearchChange={onProficiencySearchChange}
                        addChallengeToQuestArchetype={
                          addChallengeToQuestArchetype
                        }
                        onSaveNode={onSaveNode}
                        onEditChallenge={onEditChallenge}
                      />
                    </div>
                  ) : (
                    <div className="qa-flow-branch qa-flow-branch-terminal">
                      <div className="qa-meta">No further node unlocked.</div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        ) : (
          <div className="qa-empty" style={{ marginTop: 12 }}>
            No challenges yet. Add the first challenge to define the flow.
          </div>
        )}
      </div>
    </div>
  );
};

type FlowMapNode = {
  id: string;
  depth: number;
  order: number;
  label: string;
};

type FlowMapEdge = {
  from: string;
  to: string;
};

type FlowMapLayout = {
  nodes: Array<FlowMapNode & { x: number; y: number }>;
  edges: Array<{ fromX: number; fromY: number; toX: number; toY: number }>;
  width: number;
  height: number;
};

const buildFlowMapLayout = (
  root: QuestArchetypeNode | undefined | null,
  locationArchetypes: LocationArchetype[],
  monsterTemplates: MonsterTemplateRecord[],
  scenarioTemplates: ScenarioTemplateRecord[]
): FlowMapLayout | null => {
  if (!root) return null;
  const nodes: FlowMapNode[] = [];
  const edges: FlowMapEdge[] = [];
  const visited = new Set<string>();
  let orderIndex = 0;
  let maxDepth = 0;

  const walk = (node: QuestArchetypeNode, depth: number) => {
    maxDepth = Math.max(maxDepth, depth);
    if (!visited.has(node.id)) {
      visited.add(node.id);
      nodes.push({
        id: node.id,
        depth,
        order: orderIndex,
        label: describeQuestArchetypeNode(
          node,
          locationArchetypes,
          monsterTemplates,
          scenarioTemplates
        ),
      });
      orderIndex += 1;
    }
    node.challenges?.forEach((challenge) => {
      if (!challenge.unlockedNode) return;
      edges.push({ from: node.id, to: challenge.unlockedNode.id });
      walk(challenge.unlockedNode, depth + 1);
    });
  };

  walk(root, 0);

  const xSpacing = 140;
  const ySpacing = 90;
  const padding = 32;
  const positionedNodes = nodes.map((node) => ({
    ...node,
    x: padding + node.depth * xSpacing,
    y: padding + node.order * ySpacing,
  }));
  const positions = new Map(positionedNodes.map((node) => [node.id, node]));
  const positionedEdges = edges
    .map((edge) => {
      const from = positions.get(edge.from);
      const to = positions.get(edge.to);
      if (!from || !to) return null;
      return {
        fromX: from.x,
        fromY: from.y,
        toX: to.x,
        toY: to.y,
      };
    })
    .filter(Boolean) as Array<{
    fromX: number;
    fromY: number;
    toX: number;
    toY: number;
  }>;

  const width = padding * 2 + Math.max(1, maxDepth + 1) * xSpacing;
  const height = padding * 2 + Math.max(1, nodes.length) * ySpacing;

  return {
    nodes: positionedNodes,
    edges: positionedEdges,
    width,
    height,
  };
};

type RewardMode = 'explicit' | 'random';
type RandomRewardSize = 'small' | 'medium' | 'large';

type QuestArchetypeRewardRow = {
  inventoryItemId: string;
  quantity: number;
};

type QuestArchetypeSpellRewardRow = {
  spellId: string;
};

type QuestArchetypeFormState = {
  name: string;
  description: string;
  acceptanceDialogueText: string;
  imageUrl: string;
  locationArchetypeId: string;
  locationArchetypeQuery: string;
  rootDifficulty: number;
  defaultGold: number;
  rewardMode: RewardMode;
  randomRewardSize: RandomRewardSize;
  rewardExperience: number;
  recurrenceFrequency: string;
  materialRewards: ReturnType<typeof emptyMaterialReward>[];
  itemRewards: QuestArchetypeRewardRow[];
  spellRewards: QuestArchetypeSpellRewardRow[];
  characterTagsText: string;
  internalTagsText: string;
};

const createEmptyQuestArchetypeForm = (): QuestArchetypeFormState => ({
  name: '',
  description: '',
  acceptanceDialogueText: '',
  imageUrl: '',
  locationArchetypeId: '',
  locationArchetypeQuery: '',
  rootDifficulty: 0,
  defaultGold: 0,
  rewardMode: 'random',
  randomRewardSize: 'small',
  rewardExperience: 0,
  recurrenceFrequency: '',
  materialRewards: [],
  itemRewards: [],
  spellRewards: [],
  characterTagsText: '',
  internalTagsText: '',
});

const buildQuestArchetypeFormFromRecord = (
  archetype: QuestArchetype,
  locationArchetypes: LocationArchetype[]
): QuestArchetypeFormState => ({
  name: archetype.name ?? '',
  description: archetype.description ?? '',
  acceptanceDialogueText: (archetype.acceptanceDialogue ?? []).join('\n'),
  imageUrl: archetype.imageUrl ?? '',
  locationArchetypeId: archetype.root?.locationArchetypeId ?? '',
  locationArchetypeQuery:
    locationArchetypes.find(
      (entry) => entry.id === archetype.root?.locationArchetypeId
    )?.name ?? '',
  rootDifficulty: archetype.root?.difficulty ?? 0,
  defaultGold: archetype.defaultGold ?? 0,
  rewardMode: archetype.rewardMode === 'explicit' ? 'explicit' : 'random',
  randomRewardSize:
    archetype.randomRewardSize === 'medium' ||
    archetype.randomRewardSize === 'large'
      ? archetype.randomRewardSize
      : 'small',
  rewardExperience: archetype.rewardExperience ?? 0,
  recurrenceFrequency: archetype.recurrenceFrequency ?? '',
  materialRewards: (archetype.materialRewards ?? []).map((reward) => ({
    resourceKey: reward.resourceKey,
    amount: reward.amount,
  })),
  itemRewards: (archetype.itemRewards ?? []).map((reward) => ({
    inventoryItemId: reward.inventoryItemId
      ? String(reward.inventoryItemId)
      : '',
    quantity: reward.quantity ?? 1,
  })),
  spellRewards: (archetype.spellRewards ?? []).map((reward) => ({
    spellId: reward.spellId ?? '',
  })),
  characterTagsText: (archetype.characterTags ?? []).join(', '),
  internalTagsText: (archetype.internalTags ?? []).join(', '),
});

const normalizeQuestArchetypeDraft = (
  form: QuestArchetypeFormState
): QuestArchetypeDraft => {
  const rewardMode = form.rewardMode;
  const acceptanceDialogue = form.acceptanceDialogueText
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line.length > 0);
  const characterTags = form.characterTagsText
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0);
  const internalTags = form.internalTagsText
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0);

  return {
    name: form.name.trim(),
    description: form.description.trim(),
    acceptanceDialogue,
    imageUrl: form.imageUrl.trim(),
    rootNode: {
      nodeType: 'location',
      locationArchetypeId: form.locationArchetypeId,
      difficulty: Number(form.rootDifficulty) || 0,
    },
    rootDifficulty: Number(form.rootDifficulty) || 0,
    defaultGold: Number(form.defaultGold) || 0,
    rewardMode,
    randomRewardSize: form.randomRewardSize,
    rewardExperience:
      rewardMode === 'explicit' ? Number(form.rewardExperience) || 0 : 0,
    recurrenceFrequency: form.recurrenceFrequency.trim() || null,
    materialRewards:
      rewardMode === 'explicit'
        ? normalizeMaterialRewards(form.materialRewards)
        : [],
    itemRewards:
      rewardMode === 'explicit'
        ? form.itemRewards
            .map((reward) => ({
              inventoryItemId: Number(reward.inventoryItemId) || 0,
              quantity: Number(reward.quantity) || 0,
            }))
            .filter(
              (reward) => reward.inventoryItemId > 0 && reward.quantity > 0
            )
        : [],
    spellRewards:
      rewardMode === 'explicit'
        ? form.spellRewards.filter((reward) => reward.spellId.trim().length > 0)
        : [],
    characterTags,
    internalTags,
  };
};

type GeneratorStepSource = 'location_archetype' | 'proximity';
type GeneratorStepContent = 'challenge' | 'scenario' | 'monster';

type QuestTemplateGeneratorStepFormState = {
  id: string;
  source: GeneratorStepSource;
  content: GeneratorStepContent;
  locationArchetypeId: string;
  proximityMeters: number;
};

type QuestTemplateGeneratorFormState = {
  name: string;
  themePrompt: string;
  characterTagsText: string;
  internalTagsText: string;
  steps: QuestTemplateGeneratorStepFormState[];
};

const createGeneratorStepId = () =>
  globalThis.crypto?.randomUUID?.() ??
  `quest-template-step-${Math.random().toString(36).slice(2, 10)}`;

const createQuestTemplateGeneratorStep = (
  source: GeneratorStepSource = 'location_archetype',
  content: GeneratorStepContent = 'challenge'
): QuestTemplateGeneratorStepFormState => ({
  id: createGeneratorStepId(),
  source,
  content: source === 'proximity' && content === 'challenge' ? 'scenario' : content,
  locationArchetypeId: '',
  proximityMeters: 100,
});

const createEmptyQuestTemplateGeneratorForm =
  (): QuestTemplateGeneratorFormState => ({
    name: '',
    themePrompt: '',
    characterTagsText: '',
    internalTagsText: '',
    steps: [createQuestTemplateGeneratorStep('location_archetype', 'challenge')],
  });

const normalizeQuestTemplateGeneratorDraft = (
  form: QuestTemplateGeneratorFormState
): QuestTemplateGeneratorDraft => ({
  name: form.name.trim(),
  themePrompt: form.themePrompt.trim(),
  characterTags: form.characterTagsText
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0),
  internalTags: form.internalTagsText
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0),
  steps: form.steps.map((step) => ({
    source: step.source,
    content: step.content,
    locationArchetypeId:
      step.source === 'location_archetype' ? step.locationArchetypeId || null : null,
    proximityMeters:
      step.source === 'proximity' ? Math.max(0, Number(step.proximityMeters) || 0) : null,
  })),
});

const validateQuestTemplateGeneratorForm = (
  form: QuestTemplateGeneratorFormState
): string | null => {
  if (form.steps.length === 0) {
    return 'Add at least one step.';
  }
  for (let index = 0; index < form.steps.length; index += 1) {
    const step = form.steps[index];
    if (index === 0 && step.source === 'proximity') {
      return 'The first step must use a location archetype anchor.';
    }
    if (step.source === 'location_archetype' && !step.locationArchetypeId) {
      return `Step ${index + 1} needs a location archetype.`;
    }
    if (step.source === 'proximity' && step.content === 'challenge') {
      return `Step ${index + 1} cannot be a proximity challenge.`;
    }
    if (step.source === 'proximity' && (Number(step.proximityMeters) || 0) < 0) {
      return `Step ${index + 1} needs a non-negative proximity.`;
    }
  }
  return null;
};

export const QuestArchetypeComponent = () => {
  const { apiClient } = useAPI();
  const {
    questArchetypes,
    locationArchetypes,
    createQuestArchetype,
    generateQuestArchetypeTemplate,
    updateQuestArchetype,
    deleteQuestArchetype,
    addChallengeToQuestArchetype,
    updateQuestArchetypeChallenge,
    deleteQuestArchetypeChallenge,
    updateQuestArchetypeNode,
  } = useQuestArchetypes();
  const [inventoryItems, setInventoryItems] = useState<InventoryItem[]>([]);
  const [monsterTemplates, setMonsterTemplates] = useState<
    MonsterTemplateRecord[]
  >([]);
  const [scenarioTemplates, setScenarioTemplates] = useState<
    ScenarioTemplateRecord[]
  >([]);
  const [spells, setSpells] = useState<Spell[]>([]);
  const [inventoryItemsLoading, setInventoryItemsLoading] =
    useState<boolean>(false);
  const [shouldShowModal, setShouldShowModal] = useState(false);
  const [shouldShowGeneratorModal, setShouldShowGeneratorModal] =
    useState(false);
  const [createForm, setCreateForm] = useState<QuestArchetypeFormState>(
    createEmptyQuestArchetypeForm()
  );
  const [generatorForm, setGeneratorForm] =
    useState<QuestTemplateGeneratorFormState>(
      createEmptyQuestTemplateGeneratorForm()
    );
  const [editingArchetype, setEditingArchetype] =
    useState<QuestArchetype | null>(null);
  const [editForm, setEditForm] = useState<QuestArchetypeFormState>(
    createEmptyQuestArchetypeForm()
  );
  const [editingChallenge, setEditingChallenge] =
    useState<QuestArchetypeChallenge | null>(null);
  const [editChallengeRewardPoints, setEditChallengeRewardPoints] =
    useState<number>(0);
  const [editChallengeRewardItemId, setEditChallengeRewardItemId] =
    useState<number>(0);
  const [editChallengeProficiency, setEditChallengeProficiency] =
    useState<string>('');
  const [editChallengeDifficulty, setEditChallengeDifficulty] =
    useState<number>(0);
  const [proficiencySearch, setProficiencySearch] = useState<string>('');
  const [proficiencyOptions, setProficiencyOptions] = useState<string[]>([]);
  const [archetypeSearch, setArchetypeSearch] = useState<string>('');
  const [selectedArchetypeId, setSelectedArchetypeId] = useState<string>('');

  const filteredCreateLocationArchetypes = locationArchetypes
    .filter((archetype) =>
      archetype.name
        .toLowerCase()
        .includes(createForm.locationArchetypeQuery.trim().toLowerCase())
    )
    .slice(0, 8);

  const generatorValidationError = validateQuestTemplateGeneratorForm(
    generatorForm
  );

  const filteredArchetypes = useMemo(
    () =>
      questArchetypes.filter((archetype) =>
        archetype.name
          .toLowerCase()
          .includes(archetypeSearch.trim().toLowerCase())
      ),
    [questArchetypes, archetypeSearch]
  );

  const selectedArchetype = useMemo(
    () =>
      questArchetypes.find(
        (archetype) => archetype.id === selectedArchetypeId
      ) ?? null,
    [questArchetypes, selectedArchetypeId]
  );

  const flowMapLayout = useMemo(
    () =>
      buildFlowMapLayout(
        selectedArchetype?.root ?? null,
        locationArchetypes,
        monsterTemplates,
        scenarioTemplates
      ),
    [selectedArchetype, locationArchetypes, monsterTemplates, scenarioTemplates]
  );

  useEffect(() => {
    if (questArchetypes.length === 0) {
      setSelectedArchetypeId('');
      return;
    }
    const stillExists = questArchetypes.some(
      (archetype) => archetype.id === selectedArchetypeId
    );
    if (!stillExists) {
      setSelectedArchetypeId(questArchetypes[0].id);
    }
  }, [questArchetypes, selectedArchetypeId]);

  useEffect(() => {
    const fetchReferenceData = async () => {
      setInventoryItemsLoading(true);
      try {
        const [
          inventoryResponse,
          spellsResponse,
          monsterTemplateResponse,
          scenarioTemplateResponse,
        ] =
          await Promise.all([
            apiClient.get<InventoryItem[]>('/sonar/inventory-items'),
            apiClient.get<Spell[]>('/sonar/spells'),
            apiClient.get<PaginatedResponse<MonsterTemplateRecord>>(
              '/sonar/admin/monster-templates?page=1&pageSize=500'
            ),
            apiClient.get<PaginatedResponse<ScenarioTemplateRecord>>(
              '/sonar/admin/scenario-templates?page=1&pageSize=500'
            ),
          ]);
        setInventoryItems(inventoryResponse);
        setSpells(spellsResponse);
        setMonsterTemplates(monsterTemplateResponse.items ?? []);
        setScenarioTemplates(scenarioTemplateResponse.items ?? []);
      } catch (error) {
        console.error('Error fetching quest archetype reference data:', error);
      } finally {
        setInventoryItemsLoading(false);
      }
    };

    fetchReferenceData();
  }, [apiClient]);

  useEffect(() => {
    const query = proficiencySearch.trim();
    let active = true;
    const handle = window.setTimeout(async () => {
      try {
        const results = await apiClient.get<string[]>(
          `/sonar/proficiencies?query=${encodeURIComponent(query)}&limit=25`
        );
        if (!active) return;
        setProficiencyOptions(Array.isArray(results) ? results : []);
      } catch (error) {
        if (active) {
          console.error('Failed to load proficiencies', error);
          setProficiencyOptions([]);
        }
      }
    }, 200);
    return () => {
      active = false;
      window.clearTimeout(handle);
    };
  }, [apiClient, proficiencySearch]);

  const openChallengeEditor = (selected: QuestArchetypeChallenge) => {
    setEditingChallenge(selected);
    setEditChallengeRewardPoints(selected.reward ?? 0);
    const itemId = selected.inventoryItemId ?? 0;
    setEditChallengeRewardItemId(itemId);
    setEditChallengeProficiency(selected.proficiency ?? '');
    setEditChallengeDifficulty(selected.difficulty ?? 0);
    setProficiencySearch(selected.proficiency ?? '');
  };

  return (
    <div className="qa-theme">
      <datalist id="qa-proficiency-options">
        {proficiencyOptions.map((option) => (
          <option key={option} value={option} />
        ))}
      </datalist>
      <div className="qa-shell">
        <header className="qa-hero">
          <div>
            <div className="qa-kicker">Quest Design Lab</div>
            <h1 className="qa-title">Quest Archetypes</h1>
            <p className="qa-subtitle">
              Build the backbone of every adventure. Define branching
              challenges, rewards, and proficiencies so generated quests feel
              crafted rather than random.
            </p>
          </div>
          <div className="qa-hero-actions">
            {inventoryItemsLoading && (
              <span className="qa-chip muted">Loading items…</span>
            )}
            <button
              className="qa-btn qa-btn-outline"
              onClick={() => {
                setGeneratorForm(createEmptyQuestTemplateGeneratorForm());
                setShouldShowGeneratorModal(true);
              }}
            >
              Generate Template
            </button>
            <button
              className="qa-btn qa-btn-primary"
              onClick={() => {
                setCreateForm(createEmptyQuestArchetypeForm());
                setShouldShowModal(true);
              }}
            >
              New Archetype
            </button>
          </div>
        </header>

        <section className="qa-layout">
          <aside className="qa-sidebar">
            <div className="qa-card qa-sidebar-card">
              <div className="qa-card-title">Archetype Library</div>
              <p className="qa-muted" style={{ marginTop: 6 }}>
                Pick a quest archetype to shape its challenge flow.
              </p>
              <input
                className="qa-input qa-sidebar-search"
                placeholder="Search archetypes..."
                value={archetypeSearch}
                onChange={(e) => setArchetypeSearch(e.target.value)}
              />
              <div className="qa-sidebar-list">
                {filteredArchetypes.length === 0 ? (
                  <div className="qa-empty">
                    No archetypes match that search.
                  </div>
                ) : (
                  filteredArchetypes.map((questArchetype) => {
                    const rootLocation = describeQuestArchetypeNode(
                      questArchetype.root,
                      locationArchetypes,
                      monsterTemplates,
                      scenarioTemplates
                    );
                    return (
                      <button
                        key={questArchetype.id}
                        className={`qa-sidebar-item ${selectedArchetypeId === questArchetype.id ? 'is-active' : ''}`}
                        onClick={() =>
                          setSelectedArchetypeId(questArchetype.id)
                        }
                      >
                        <div className="qa-sidebar-item-title">
                          {questArchetype.name}
                        </div>
                        <div className="qa-meta">
                          Root: {rootLocation} ·{' '}
                          {questArchetype.root?.challenges?.length ?? 0}{' '}
                          challenges
                        </div>
                      </button>
                    );
                  })
                )}
              </div>
            </div>
          </aside>

          <div className="qa-builder">
            {!selectedArchetype ? (
              <div className="qa-panel">
                <div className="qa-card-title">Select a quest archetype</div>
                <p className="qa-muted" style={{ marginTop: 8 }}>
                  Choose an archetype on the left to build its challenge flow.
                </p>
              </div>
            ) : (
              <>
                <div className="qa-card qa-builder-header">
                  <div>
                    <div className="qa-kicker">Quest Flow Builder</div>
                    <h2
                      className="qa-title"
                      style={{ fontSize: 'clamp(26px, 3vw, 34px)' }}
                    >
                      {selectedArchetype.name}
                    </h2>
                    <p className="qa-subtitle">
                      Craft the journey by stacking challenges and branching
                      nodes. Each challenge can unlock a new node to extend the
                      quest.
                    </p>
                  </div>
                  <div className="qa-actions">
                    <button
                      className="qa-btn qa-btn-ghost"
                      onClick={() => {
                        setEditingArchetype(selectedArchetype);
                        setEditForm(
                          buildQuestArchetypeFormFromRecord(
                            selectedArchetype,
                            locationArchetypes
                          )
                        );
                      }}
                    >
                      Edit Template
                    </button>
                    <button
                      className="qa-btn qa-btn-danger"
                      onClick={() => {
                        if (
                          window.confirm(
                            'Are you sure you want to delete this quest archetype?'
                          )
                        ) {
                          deleteQuestArchetype(selectedArchetype.id);
                        }
                      }}
                    >
                      Delete Archetype
                    </button>
                  </div>
                </div>

                <div className="qa-card qa-builder-summary">
                  <div className="qa-stat-grid">
                    <div className="qa-stat">
                      <div className="qa-stat-label">Reward Mode</div>
                      <div className="qa-stat-value">
                        {(
                          selectedArchetype.rewardMode ?? 'random'
                        ).toUpperCase()}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Root Node</div>
                      <div className="qa-stat-value">
                        {describeQuestArchetypeNode(
                          selectedArchetype.root,
                          locationArchetypes,
                          monsterTemplates,
                          scenarioTemplates
                        )}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Quest Giver Tags</div>
                      <div className="qa-stat-value">
                        {(selectedArchetype.characterTags ?? []).length > 0
                          ? (selectedArchetype.characterTags ?? []).join(', ')
                          : 'Any'}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Internal Tags</div>
                      <div className="qa-stat-value">
                        {(selectedArchetype.internalTags ?? []).length > 0
                          ? (selectedArchetype.internalTags ?? []).join(', ')
                          : 'None'}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Created</div>
                      <div className="qa-stat-value">
                        {new Date(
                          selectedArchetype.createdAt
                        ).toLocaleDateString()}
                      </div>
                    </div>
                  </div>

                  <div className="qa-divider" />

                  <div className="qa-panel">
                    <div className="qa-meta">Template Summary</div>
                    {selectedArchetype.description ? (
                      <p className="qa-muted" style={{ marginTop: 10 }}>
                        {selectedArchetype.description}
                      </p>
                    ) : (
                      <div className="qa-empty" style={{ marginTop: 10 }}>
                        No description configured.
                      </div>
                    )}
                    <div className="qa-inline" style={{ marginTop: 10 }}>
                      <span className="qa-chip muted">
                        Gold: {selectedArchetype.defaultGold ?? 0}
                      </span>
                      <span className="qa-chip muted">
                        XP:{' '}
                        {selectedArchetype.rewardMode === 'explicit'
                          ? selectedArchetype.rewardExperience ?? 0
                          : 0}
                      </span>
                      <span className="qa-chip muted">
                        Materials:{' '}
                        {summarizeMaterialRewards(
                          selectedArchetype.materialRewards
                        )}
                      </span>
                      {selectedArchetype.recurrenceFrequency && (
                        <span className="qa-chip accent">
                          Repeats {selectedArchetype.recurrenceFrequency}
                        </span>
                      )}
                    </div>
                    {(selectedArchetype.itemRewards ?? []).length > 0 ? (
                      <div className="qa-inline" style={{ marginTop: 10 }}>
                        {(selectedArchetype.itemRewards ?? []).map((reward) => {
                          const item = inventoryItems.find(
                            (entry) => entry.id === reward.inventoryItemId
                          );
                          return (
                            <span
                              key={
                                reward.id ??
                                `${reward.inventoryItemId}-${reward.quantity}`
                              }
                              className="qa-chip"
                            >
                              {reward.quantity}x{' '}
                              {item?.name ?? `Item ${reward.inventoryItemId}`}
                            </span>
                          );
                        })}
                      </div>
                    ) : null}
                    {(selectedArchetype.spellRewards ?? []).length > 0 && (
                      <div className="qa-inline" style={{ marginTop: 10 }}>
                        {(selectedArchetype.spellRewards ?? []).map(
                          (reward) => (
                            <span
                              key={reward.id ?? reward.spellId}
                              className="qa-chip success"
                            >
                              {reward.spell?.name ?? reward.spellId}
                            </span>
                          )
                        )}
                      </div>
                    )}
                  </div>
                </div>

                {flowMapLayout && (
                  <div className="qa-card qa-flow-map">
                    <div className="qa-card-title">Flow Map</div>
                    <p className="qa-muted" style={{ marginTop: 6 }}>
                      A mini-map of the quest flow. Each node represents a
                      location archetype, connected by challenges.
                    </p>
                    <div className="qa-flow-map-canvas">
                      <svg
                        viewBox={`0 0 ${flowMapLayout.width} ${flowMapLayout.height}`}
                        role="img"
                        aria-label="Quest archetype flow map"
                      >
                        <defs>
                          <marker
                            id="qa-flow-arrow"
                            markerWidth="8"
                            markerHeight="8"
                            refX="6"
                            refY="3"
                            orient="auto"
                            markerUnits="strokeWidth"
                          >
                            <path
                              d="M0,0 L0,6 L6,3 z"
                              fill="rgba(255,255,255,0.65)"
                            />
                          </marker>
                        </defs>
                        {flowMapLayout.edges.map((edge, index) => (
                          <line
                            key={`${edge.fromX}-${edge.fromY}-${edge.toX}-${edge.toY}-${index}`}
                            x1={edge.fromX}
                            y1={edge.fromY}
                            x2={edge.toX}
                            y2={edge.toY}
                            stroke="rgba(255,255,255,0.3)"
                            strokeWidth="2"
                            markerEnd="url(#qa-flow-arrow)"
                          />
                        ))}
                        {flowMapLayout.nodes.map((node) => (
                          <g key={node.id}>
                            <circle
                              cx={node.x}
                              cy={node.y}
                              r="12"
                              fill="rgba(255,107,74,0.7)"
                              stroke="rgba(255,255,255,0.8)"
                              strokeWidth="1"
                            />
                            <text
                              x={node.x + 18}
                              y={node.y + 4}
                              fill="rgba(236,243,245,0.9)"
                              fontSize="11"
                              fontFamily="Space Grotesk, sans-serif"
                            >
                              {node.label}
                            </text>
                          </g>
                        ))}
                      </svg>
                    </div>
                  </div>
                )}

                <div className="qa-card qa-builder-flow">
                  <div className="qa-card-title">Quest Flow</div>
                  <p className="qa-muted" style={{ marginTop: 6 }}>
                    Start from the root and add challenges. Add a location to a
                    challenge to branch into a new node.
                  </p>
                  {selectedArchetype.root ? (
                    <div className="qa-flow-canvas">
                      <FlowNode
                        node={selectedArchetype.root}
                        locationArchetypes={locationArchetypes}
                        monsterTemplates={monsterTemplates}
                        scenarioTemplates={scenarioTemplates}
                        inventoryItems={inventoryItems}
                        depth={0}
                        proficiencyOptions={proficiencyOptions}
                        onProficiencySearchChange={setProficiencySearch}
                        addChallengeToQuestArchetype={
                          addChallengeToQuestArchetype
                        }
                        onSaveNode={updateQuestArchetypeNode}
                        onEditChallenge={openChallengeEditor}
                      />
                    </div>
                  ) : (
                    <div className="qa-empty" style={{ marginTop: 12 }}>
                      No root node available.
                    </div>
                  )}
                </div>
              </>
            )}
          </div>
        </section>
      </div>

      {shouldShowGeneratorModal && (
        <div className="qa-modal">
          <div className="qa-modal-card">
            <h2 className="qa-modal-title">Generate Quest Template</h2>
            <p className="qa-muted" style={{ marginBottom: 16 }}>
              Build an ordered quest flow from location and proximity steps. The
              generator will create a quest template with nodes in this exact order.
            </p>
            <form
              className="qa-form-grid"
              onSubmit={async (e) => {
                e.preventDefault();
                const validationError =
                  validateQuestTemplateGeneratorForm(generatorForm);
                if (validationError) {
                  window.alert(validationError);
                  return;
                }
                const created = await generateQuestArchetypeTemplate(
                  normalizeQuestTemplateGeneratorDraft(generatorForm)
                );
                if (created?.id) {
                  setSelectedArchetypeId(created.id);
                }
                setGeneratorForm(createEmptyQuestTemplateGeneratorForm());
                setShouldShowGeneratorModal(false);
              }}
            >
              <div className="qa-field">
                <div className="qa-label">Name</div>
                <input
                  type="text"
                  className="qa-input"
                  value={generatorForm.name}
                  onChange={(e) =>
                    setGeneratorForm((prev) => ({ ...prev, name: e.target.value }))
                  }
                  placeholder="Optional generated template name"
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Theme Prompt</div>
                <textarea
                  className="qa-textarea"
                  rows={4}
                  value={generatorForm.themePrompt}
                  onChange={(e) =>
                    setGeneratorForm((prev) => ({
                      ...prev,
                      themePrompt: e.target.value,
                    }))
                  }
                  placeholder="Describe the kind of quest you want, tone, factions, stakes, and any notable beats."
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Character Tags</div>
                <input
                  type="text"
                  className="qa-input"
                  value={generatorForm.characterTagsText}
                  onChange={(e) =>
                    setGeneratorForm((prev) => ({
                      ...prev,
                      characterTagsText: e.target.value,
                    }))
                  }
                  placeholder="merchant, outlaw, druid"
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Internal Tags</div>
                <input
                  type="text"
                  className="qa-input"
                  value={generatorForm.internalTagsText}
                  onChange={(e) =>
                    setGeneratorForm((prev) => ({
                      ...prev,
                      internalTagsText: e.target.value,
                    }))
                  }
                  placeholder="city, mystery, waterfront"
                />
              </div>

              <div className="qa-field" style={{ gridColumn: '1 / -1' }}>
                <div className="qa-label">Add Step</div>
                <div className="qa-inline" style={{ flexWrap: 'wrap' }}>
                  <button
                    type="button"
                    className="qa-btn qa-btn-ghost"
                    onClick={() =>
                      setGeneratorForm((prev) => ({
                        ...prev,
                        steps: [
                          ...prev.steps,
                          createQuestTemplateGeneratorStep(
                            'location_archetype',
                            'challenge'
                          ),
                        ],
                      }))
                    }
                  >
                    Location Challenge
                  </button>
                  <button
                    type="button"
                    className="qa-btn qa-btn-ghost"
                    onClick={() =>
                      setGeneratorForm((prev) => ({
                        ...prev,
                        steps: [
                          ...prev.steps,
                          createQuestTemplateGeneratorStep(
                            'location_archetype',
                            'scenario'
                          ),
                        ],
                      }))
                    }
                  >
                    Location Scenario
                  </button>
                  <button
                    type="button"
                    className="qa-btn qa-btn-ghost"
                    onClick={() =>
                      setGeneratorForm((prev) => ({
                        ...prev,
                        steps: [
                          ...prev.steps,
                          createQuestTemplateGeneratorStep(
                            'location_archetype',
                            'monster'
                          ),
                        ],
                      }))
                    }
                  >
                    Location Monster
                  </button>
                  <button
                    type="button"
                    className="qa-btn qa-btn-ghost"
                    onClick={() =>
                      setGeneratorForm((prev) => ({
                        ...prev,
                        steps: [
                          ...prev.steps,
                          createQuestTemplateGeneratorStep('proximity', 'scenario'),
                        ],
                      }))
                    }
                  >
                    Nearby Scenario
                  </button>
                  <button
                    type="button"
                    className="qa-btn qa-btn-ghost"
                    onClick={() =>
                      setGeneratorForm((prev) => ({
                        ...prev,
                        steps: [
                          ...prev.steps,
                          createQuestTemplateGeneratorStep('proximity', 'monster'),
                        ],
                      }))
                    }
                  >
                    Nearby Monster
                  </button>
                </div>
              </div>

              <div className="qa-field" style={{ gridColumn: '1 / -1' }}>
                <div className="qa-label">Ordered Steps</div>
                {generatorForm.steps.length === 0 ? (
                  <div className="qa-empty">No steps yet.</div>
                ) : (
                  <div className="qa-flow-challenges">
                    {generatorForm.steps.map((step, index) => (
                      <div key={step.id} className="qa-flow-challenge-card">
                        <div className="qa-flow-challenge-header">
                          <div>
                            <div className="qa-flow-challenge-title">
                              Step {index + 1}
                            </div>
                            <div className="qa-meta">
                              {step.source === 'location_archetype'
                                ? 'Location archetype anchored'
                                : `Within ${step.proximityMeters}m of previous node`}
                            </div>
                          </div>
                          <div className="qa-inline">
                            <button
                              type="button"
                              className="qa-btn qa-btn-text"
                              disabled={index === 0}
                              onClick={() =>
                                setGeneratorForm((prev) => {
                                  const steps = [...prev.steps];
                                  [steps[index - 1], steps[index]] = [
                                    steps[index],
                                    steps[index - 1],
                                  ];
                                  return { ...prev, steps };
                                })
                              }
                            >
                              Up
                            </button>
                            <button
                              type="button"
                              className="qa-btn qa-btn-text"
                              disabled={index === generatorForm.steps.length - 1}
                              onClick={() =>
                                setGeneratorForm((prev) => {
                                  const steps = [...prev.steps];
                                  [steps[index], steps[index + 1]] = [
                                    steps[index + 1],
                                    steps[index],
                                  ];
                                  return { ...prev, steps };
                                })
                              }
                            >
                              Down
                            </button>
                            <button
                              type="button"
                              className="qa-btn qa-btn-text"
                              onClick={() =>
                                setGeneratorForm((prev) => ({
                                  ...prev,
                                  steps: prev.steps.filter(
                                    (entry) => entry.id !== step.id
                                  ),
                                }))
                              }
                            >
                              Remove
                            </button>
                          </div>
                        </div>

                        <div className="qa-form-grid">
                          <div className="qa-field">
                            <div className="qa-label">Anchor Type</div>
                            <select
                              className="qa-select"
                              value={step.source}
                              onChange={(e) =>
                                setGeneratorForm((prev) => ({
                                  ...prev,
                                  steps: prev.steps.map((entry) =>
                                    entry.id !== step.id
                                      ? entry
                                      : {
                                          ...entry,
                                          source: e.target.value as GeneratorStepSource,
                                          content:
                                            e.target.value === 'proximity' &&
                                            entry.content === 'challenge'
                                              ? 'scenario'
                                              : entry.content,
                                        }
                                  ),
                                }))
                              }
                            >
                              <option value="location_archetype">
                                Location Archetype
                              </option>
                              <option value="proximity" disabled={index === 0}>
                                Proximity To Previous Node
                              </option>
                            </select>
                          </div>
                          <div className="qa-field">
                            <div className="qa-label">Content</div>
                            <select
                              className="qa-select"
                              value={step.content}
                              onChange={(e) =>
                                setGeneratorForm((prev) => ({
                                  ...prev,
                                  steps: prev.steps.map((entry) =>
                                    entry.id !== step.id
                                      ? entry
                                      : {
                                          ...entry,
                                          content: e.target.value as GeneratorStepContent,
                                          source:
                                            entry.source === 'proximity' &&
                                            e.target.value === 'challenge'
                                              ? 'location_archetype'
                                              : entry.source,
                                        }
                                  ),
                                }))
                              }
                            >
                              <option value="challenge">Challenge</option>
                              <option value="scenario">Scenario</option>
                              <option value="monster">Monster</option>
                            </select>
                          </div>

                          {step.source === 'location_archetype' ? (
                            <div className="qa-field" style={{ gridColumn: '1 / -1' }}>
                              <div className="qa-label">Location Archetype</div>
                              <select
                                className="qa-select"
                                value={step.locationArchetypeId}
                                onChange={(e) =>
                                  setGeneratorForm((prev) => ({
                                    ...prev,
                                    steps: prev.steps.map((entry) =>
                                      entry.id !== step.id
                                        ? entry
                                        : {
                                            ...entry,
                                            locationArchetypeId: e.target.value,
                                          }
                                    ),
                                  }))
                                }
                              >
                                <option value="">Select a location archetype</option>
                                {locationArchetypes.map((archetype) => (
                                  <option key={archetype.id} value={archetype.id}>
                                    {archetype.name}
                                  </option>
                                ))}
                              </select>
                            </div>
                          ) : (
                            <div className="qa-field">
                              <div className="qa-label">Proximity (m)</div>
                              <input
                                type="number"
                                min={0}
                                className="qa-input"
                                value={step.proximityMeters}
                                onChange={(e) =>
                                  setGeneratorForm((prev) => ({
                                    ...prev,
                                    steps: prev.steps.map((entry) =>
                                      entry.id !== step.id
                                        ? entry
                                        : {
                                            ...entry,
                                            proximityMeters:
                                              parseInt(e.target.value) || 0,
                                          }
                                    ),
                                  }))
                                }
                              />
                            </div>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
                {generatorValidationError && (
                  <div className="qa-helper" style={{ marginTop: 8 }}>
                    {generatorValidationError}
                  </div>
                )}
              </div>

              <div className="qa-footer">
                <button
                  type="button"
                  className="qa-btn qa-btn-outline"
                  onClick={() => {
                    setShouldShowGeneratorModal(false);
                    setGeneratorForm(createEmptyQuestTemplateGeneratorForm());
                  }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="qa-btn qa-btn-primary"
                  disabled={Boolean(generatorValidationError)}
                >
                  Generate
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {shouldShowModal && (
        <div className="qa-modal">
          <div className="qa-modal-card">
            <h2 className="qa-modal-title">Create Quest Archetype</h2>
            <p className="qa-muted" style={{ marginBottom: 16 }}>
              Define the reusable quest template, then use zone assignment to
              auto-match quest givers by character tags.
            </p>
            <div className="qa-empty">
              The full template editor is being shown in the edit flow below for
              now. Create a basic archetype here, then open `Edit Template` to
              fill in dialogue, rewards, tags, and recurrence.
            </div>
            <form
              className="qa-form-grid"
              onSubmit={async (e) => {
                e.preventDefault();
                const created = await createQuestArchetype(
                  normalizeQuestArchetypeDraft(createForm)
                );
                if (created?.id) {
                  setSelectedArchetypeId(created.id);
                  setEditingArchetype(created);
                  setEditForm(
                    buildQuestArchetypeFormFromRecord(
                      created,
                      locationArchetypes
                    )
                  );
                }
                setCreateForm(createEmptyQuestArchetypeForm());
                setShouldShowModal(false);
              }}
            >
              <div className="qa-field">
                <div className="qa-label">Name</div>
                <input
                  type="text"
                  className="qa-input"
                  value={createForm.name}
                  onChange={(e) =>
                    setCreateForm((prev) => ({ ...prev, name: e.target.value }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Location Archetype</div>
                <div className="qa-combobox">
                  <input
                    type="text"
                    className="qa-input"
                    value={createForm.locationArchetypeQuery}
                    onChange={(e) => {
                      const value = e.target.value;
                      const matched = locationArchetypes.find(
                        (archetype) =>
                          archetype.name.toLowerCase() ===
                          value.trim().toLowerCase()
                      );
                      setCreateForm((prev) => ({
                        ...prev,
                        locationArchetypeQuery: value,
                        locationArchetypeId: matched ? matched.id : '',
                      }));
                    }}
                    placeholder="Search location archetypes..."
                  />
                  {createForm.locationArchetypeQuery.trim().length > 0 && (
                    <div className="qa-combobox-list">
                      {filteredCreateLocationArchetypes.length === 0 ? (
                        <div className="qa-combobox-empty">No matches.</div>
                      ) : (
                        filteredCreateLocationArchetypes.map((archetype) => (
                          <button
                            key={archetype.id}
                            type="button"
                            className="qa-combobox-option"
                            onClick={() =>
                              setCreateForm((prev) => ({
                                ...prev,
                                locationArchetypeId: archetype.id,
                                locationArchetypeQuery: archetype.name,
                              }))
                            }
                          >
                            {archetype.name}
                          </button>
                        ))
                      )}
                    </div>
                  )}
                </div>
              </div>
              <div className="qa-field">
                <div className="qa-label">Root Difficulty</div>
                <input
                  type="number"
                  min={0}
                  className="qa-input"
                  value={createForm.rootDifficulty}
                  onChange={(e) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      rootDifficulty: parseInt(e.target.value) || 0,
                    }))
                  }
                />
              </div>
              <div className="qa-footer">
                <button
                  type="button"
                  className="qa-btn qa-btn-outline"
                  onClick={() => {
                    setShouldShowModal(false);
                    setCreateForm(createEmptyQuestArchetypeForm());
                  }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="qa-btn qa-btn-primary"
                  disabled={
                    !createForm.name.trim() || !createForm.locationArchetypeId
                  }
                >
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {editingArchetype && (
        <div className="qa-modal">
          <div className="qa-modal-card">
            <h2 className="qa-modal-title">Edit Quest Template</h2>
            <div className="qa-form-grid">
              <div className="qa-field">
                <div className="qa-label">Name</div>
                <input
                  type="text"
                  className="qa-input"
                  value={editForm.name}
                  onChange={(e) =>
                    setEditForm((prev) => ({ ...prev, name: e.target.value }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Description</div>
                <textarea
                  className="qa-input"
                  rows={3}
                  value={editForm.description}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      description: e.target.value,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Acceptance Dialogue</div>
                <textarea
                  className="qa-input"
                  rows={4}
                  value={editForm.acceptanceDialogueText}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      acceptanceDialogueText: e.target.value,
                    }))
                  }
                  placeholder="One line per dialogue bubble"
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Image URL</div>
                <input
                  type="text"
                  className="qa-input"
                  value={editForm.imageUrl}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      imageUrl: e.target.value,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Quest Giver Tags</div>
                <input
                  type="text"
                  className="qa-input"
                  value={editForm.characterTagsText}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      characterTagsText: e.target.value,
                    }))
                  }
                  placeholder="merchant, scholar, ranger"
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Internal Tags</div>
                <input
                  type="text"
                  className="qa-input"
                  value={editForm.internalTagsText}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      internalTagsText: e.target.value,
                    }))
                  }
                  placeholder="story_arc, faction, tutorial"
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Reward Mode</div>
                <select
                  className="qa-select"
                  value={editForm.rewardMode}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      rewardMode: e.target.value as RewardMode,
                    }))
                  }
                >
                  <option value="random">Random</option>
                  <option value="explicit">Explicit</option>
                </select>
              </div>
              <div className="qa-field">
                <div className="qa-label">Random Reward Size</div>
                <select
                  className="qa-select"
                  value={editForm.randomRewardSize}
                  disabled={editForm.rewardMode !== 'random'}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      randomRewardSize: e.target.value as RandomRewardSize,
                    }))
                  }
                >
                  <option value="small">Small</option>
                  <option value="medium">Medium</option>
                  <option value="large">Large</option>
                </select>
              </div>
              <div className="qa-field">
                <div className="qa-label">Default Gold</div>
                <input
                  type="number"
                  min={0}
                  className="qa-input"
                  value={editForm.defaultGold}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      defaultGold: parseInt(e.target.value) || 0,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Reward Experience</div>
                <input
                  type="number"
                  min={0}
                  className="qa-input"
                  disabled={editForm.rewardMode !== 'explicit'}
                  value={editForm.rewardExperience}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      rewardExperience: parseInt(e.target.value) || 0,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Recurrence</div>
                <select
                  className="qa-select"
                  value={editForm.recurrenceFrequency}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      recurrenceFrequency: e.target.value,
                    }))
                  }
                >
                  <option value="">None</option>
                  <option value="daily">Daily</option>
                  <option value="weekly">Weekly</option>
                  <option value="monthly">Monthly</option>
                </select>
              </div>
              <div className="qa-field">
                <div className="qa-label">Root Difficulty</div>
                <input
                  type="number"
                  min={0}
                  className="qa-input"
                  value={editForm.rootDifficulty}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      rootDifficulty: parseInt(e.target.value) || 0,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Material Rewards</div>
                <MaterialRewardsEditor
                  value={editForm.materialRewards}
                  onChange={(materialRewards) =>
                    setEditForm((prev) => ({ ...prev, materialRewards }))
                  }
                  disabled={editForm.rewardMode !== 'explicit'}
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Item Rewards</div>
                {editForm.itemRewards.length === 0 ? (
                  <div className="qa-empty">No item rewards yet.</div>
                ) : (
                  <div className="qa-form-grid">
                    {editForm.itemRewards.map((reward, index) => (
                      <div
                        key={`edit-reward-${index}`}
                        className="qa-reward-row"
                      >
                        <select
                          className="qa-select"
                          value={reward.inventoryItemId}
                          disabled={editForm.rewardMode !== 'explicit'}
                          onChange={(e) =>
                            setEditForm((prev) => ({
                              ...prev,
                              itemRewards: prev.itemRewards.map(
                                (entry, rewardIndex) =>
                                  rewardIndex === index
                                    ? {
                                        ...entry,
                                        inventoryItemId: e.target.value,
                                      }
                                    : entry
                              ),
                            }))
                          }
                        >
                          <option value="">Select an item</option>
                          {inventoryItems.map((item) => (
                            <option key={item.id} value={item.id}>
                              {item.name}
                            </option>
                          ))}
                        </select>
                        <input
                          type="number"
                          min={1}
                          className="qa-input"
                          disabled={editForm.rewardMode !== 'explicit'}
                          value={reward.quantity}
                          onChange={(e) =>
                            setEditForm((prev) => ({
                              ...prev,
                              itemRewards: prev.itemRewards.map(
                                (entry, rewardIndex) =>
                                  rewardIndex === index
                                    ? {
                                        ...entry,
                                        quantity: parseInt(e.target.value) || 1,
                                      }
                                    : entry
                              ),
                            }))
                          }
                        />
                        <button
                          type="button"
                          className="qa-btn qa-btn-text"
                          onClick={() =>
                            setEditForm((prev) => ({
                              ...prev,
                              itemRewards: prev.itemRewards.filter(
                                (_, rewardIndex) => rewardIndex !== index
                              ),
                            }))
                          }
                        >
                          Remove
                        </button>
                      </div>
                    ))}
                  </div>
                )}
                <button
                  type="button"
                  className="qa-btn qa-btn-ghost"
                  onClick={() =>
                    setEditForm((prev) => ({
                      ...prev,
                      itemRewards: [
                        ...prev.itemRewards,
                        { inventoryItemId: '', quantity: 1 },
                      ],
                    }))
                  }
                >
                  Add Item Reward
                </button>
              </div>
              <div className="qa-field">
                <div className="qa-label">Spell Rewards</div>
                {editForm.spellRewards.length === 0 ? (
                  <div className="qa-empty">No spell rewards yet.</div>
                ) : (
                  <div className="qa-form-grid">
                    {editForm.spellRewards.map((reward, index) => (
                      <div
                        key={`edit-spell-${index}`}
                        className="qa-reward-row"
                      >
                        <select
                          className="qa-select"
                          value={reward.spellId}
                          disabled={editForm.rewardMode !== 'explicit'}
                          onChange={(e) =>
                            setEditForm((prev) => ({
                              ...prev,
                              spellRewards: prev.spellRewards.map(
                                (entry, rewardIndex) =>
                                  rewardIndex === index
                                    ? { ...entry, spellId: e.target.value }
                                    : entry
                              ),
                            }))
                          }
                        >
                          <option value="">Select a spell</option>
                          {spells.map((spell) => (
                            <option key={spell.id} value={spell.id}>
                              {spell.name}
                            </option>
                          ))}
                        </select>
                        <button
                          type="button"
                          className="qa-btn qa-btn-text"
                          onClick={() =>
                            setEditForm((prev) => ({
                              ...prev,
                              spellRewards: prev.spellRewards.filter(
                                (_, rewardIndex) => rewardIndex !== index
                              ),
                            }))
                          }
                        >
                          Remove
                        </button>
                      </div>
                    ))}
                  </div>
                )}
                <button
                  type="button"
                  className="qa-btn qa-btn-ghost"
                  onClick={() =>
                    setEditForm((prev) => ({
                      ...prev,
                      spellRewards: [...prev.spellRewards, { spellId: '' }],
                    }))
                  }
                >
                  Add Spell Reward
                </button>
              </div>
            </div>
            <div className="qa-footer">
              <button
                className="qa-btn qa-btn-outline"
                onClick={() => {
                  setEditingArchetype(null);
                  setEditForm(createEmptyQuestArchetypeForm());
                }}
              >
                Cancel
              </button>
              <button
                className="qa-btn qa-btn-primary"
                onClick={async () => {
                  if (!editingArchetype) return;
                  const draft = normalizeQuestArchetypeDraft(editForm);
                  await updateQuestArchetype({
                    ...editingArchetype,
                    name: draft.name,
                    description: draft.description,
                    acceptanceDialogue: draft.acceptanceDialogue,
                    imageUrl: draft.imageUrl,
                    defaultGold: draft.defaultGold ?? 0,
                    rewardMode: draft.rewardMode,
                    randomRewardSize: draft.randomRewardSize,
                    rewardExperience: draft.rewardExperience ?? 0,
                    recurrenceFrequency: draft.recurrenceFrequency ?? null,
                    materialRewards: draft.materialRewards,
                    itemRewards: draft.itemRewards,
                    spellRewards: draft.spellRewards,
                    characterTags: draft.characterTags,
                    internalTags: draft.internalTags,
                  });
                  await updateQuestArchetypeNode(editingArchetype.root.id, {
                    difficulty: editForm.rootDifficulty,
                  });
                  setEditingArchetype(null);
                  setEditForm(createEmptyQuestArchetypeForm());
                }}
              >
                Save
              </button>
            </div>
          </div>
        </div>
      )}

      {editingChallenge && (
        <div className="qa-modal">
          <div className="qa-modal-card">
            <h2 className="qa-modal-title">Edit Challenge</h2>
            <div className="qa-form-grid">
              <div className="qa-field">
                <div className="qa-label">Reward Points</div>
                <input
                  type="number"
                  min={0}
                  className="qa-input"
                  value={editChallengeRewardPoints}
                  onChange={(e) =>
                    setEditChallengeRewardPoints(parseInt(e.target.value) || 0)
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Difficulty</div>
                <input
                  type="number"
                  min={0}
                  className="qa-input"
                  value={editChallengeDifficulty}
                  onChange={(e) =>
                    setEditChallengeDifficulty(parseInt(e.target.value) || 0)
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Reward Item</div>
                <select
                  className="qa-select"
                  value={editChallengeRewardItemId || ''}
                  onChange={(e) =>
                    setEditChallengeRewardItemId(parseInt(e.target.value) || 0)
                  }
                >
                  <option value="">Select an item</option>
                  {inventoryItems.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="qa-field">
                <div className="qa-label">Proficiency</div>
                <input
                  type="text"
                  className="qa-input"
                  value={editChallengeProficiency}
                  onChange={(e) => {
                    setEditChallengeProficiency(e.target.value);
                    setProficiencySearch(e.target.value);
                  }}
                  list="qa-proficiency-options"
                  placeholder="Optional proficiency (e.g. Persuasion)"
                />
              </div>
            </div>
            <div className="qa-footer">
              <button
                className="qa-btn qa-btn-outline"
                onClick={() => setEditingChallenge(null)}
              >
                Cancel
              </button>
              <button
                className="qa-btn qa-btn-danger"
                onClick={async () => {
                  if (!editingChallenge) return;
                  const confirmDelete = window.confirm(
                    'Delete this challenge? This cannot be undone.'
                  );
                  if (!confirmDelete) return;
                  await deleteQuestArchetypeChallenge(editingChallenge.id);
                  setEditingChallenge(null);
                }}
              >
                Delete
              </button>
              <button
                className="qa-btn qa-btn-primary"
                onClick={async () => {
                  const trimmed = editChallengeProficiency.trim();
                  await updateQuestArchetypeChallenge(editingChallenge.id, {
                    reward: editChallengeRewardPoints,
                    inventoryItemId:
                      editChallengeRewardItemId > 0
                        ? editChallengeRewardItemId
                        : null,
                    proficiency: trimmed.length > 0 ? trimmed : null,
                    difficulty: editChallengeDifficulty,
                  });
                  setEditingChallenge(null);
                }}
              >
                Save
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
