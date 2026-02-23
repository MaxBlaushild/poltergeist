import React, { useMemo, useState, useEffect } from "react";
import { useAPI } from "@poltergeist/contexts";
import { useQuestArchetypes } from "../contexts/questArchetypes.tsx";
import { LocationArchetype, QuestArchetype, QuestArchetypeNode, QuestArchetypeChallenge, InventoryItem } from "@poltergeist/types";
import "./questArchetypeTheme.css";

interface FlowNodeProps {
  node: QuestArchetypeNode;
  locationArchetypes: LocationArchetype[];
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
    unlockedLocationArchetypeId?: string | null
  ) => void;
  onEditChallenge: (challenge: QuestArchetypeChallenge) => void;
}

const FlowNode: React.FC<FlowNodeProps> = ({
  node,
  locationArchetypes,
  inventoryItems,
  depth,
  proficiencyOptions,
  onProficiencySearchChange,
  addChallengeToQuestArchetype,
  onEditChallenge,
}) => {
  const borderColor = depth % 2 === 0 ? 'rgba(255, 107, 74, 0.4)' : 'rgba(95, 211, 181, 0.35)';
  const locationName = locationArchetypes.find((la) => la.id === node.locationArchetypeId)?.name ?? 'Unknown location';
  const [isAdding, setIsAdding] = useState(false);
  const [rewardPoints, setRewardPoints] = useState<number>(0);
  const [rewardItemId, setRewardItemId] = useState<number>(0);
  const [challengeDifficulty, setChallengeDifficulty] = useState<number>(0);
  const [challengeProficiency, setChallengeProficiency] = useState<string>("");
  const [unlockedLocationArchetypeId, setUnlockedLocationArchetypeId] = useState<string>("");

  return (
    <div className="qa-flow-node" style={{ borderColor }}>
      <div className="qa-flow-node-card">
        <div className="qa-flow-node-header">
          <div>
            <div className="qa-flow-node-title">{depth === 0 ? 'Root Node' : `Node ${depth + 1}`}</div>
            <div className="qa-meta">{locationName}</div>
          </div>
          <button className="qa-btn qa-btn-primary" onClick={() => setIsAdding((prev) => !prev)}>
            {isAdding ? 'Close' : 'Add Challenge'}
          </button>
        </div>

        {isAdding && (
          <div className="qa-flow-form">
            <div className="qa-field">
              <div className="qa-label">Reward Points</div>
              <input
                type="number"
                min={0}
                className="qa-input"
                value={rewardPoints}
                onChange={(e) => setRewardPoints(parseInt(e.target.value) || 0)}
              />
            </div>
            <div className="qa-field">
              <div className="qa-label">Difficulty</div>
              <input
                type="number"
                min={0}
                className="qa-input"
                value={challengeDifficulty}
                onChange={(e) => setChallengeDifficulty(parseInt(e.target.value) || 0)}
              />
            </div>
            <div className="qa-field">
              <div className="qa-label">Reward Item</div>
              <select
                className="qa-select"
                value={rewardItemId || ''}
                onChange={(e) => setRewardItemId(parseInt(e.target.value) || 0)}
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
            <div className="qa-field">
              <div className="qa-label">Unlocked Location Type</div>
              <select
                className="qa-select"
                value={unlockedLocationArchetypeId}
                onChange={(e) => setUnlockedLocationArchetypeId(e.target.value)}
              >
                <option value="">None</option>
                {locationArchetypes.map((archetype) => (
                  <option key={archetype.id} value={archetype.id}>
                    {archetype.name}
                  </option>
                ))}
              </select>
            </div>
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
                    rewardPoints,
                    rewardItemId || null,
                    trimmed.length > 0 ? trimmed : null,
                    challengeDifficulty,
                    unlockedLocationArchetypeId || null
                  );
                  setRewardPoints(0);
                  setRewardItemId(0);
                  setChallengeDifficulty(0);
                  setChallengeProficiency("");
                  setUnlockedLocationArchetypeId("");
                  setIsAdding(false);
                }}
              >
                Add Challenge
              </button>
            </div>
          </div>
        )}

        {node.challenges && node.challenges.length > 0 ? (
          <div className="qa-flow-challenges">
            {node.challenges.map((challenge, index) => {
              const legacyItemId = !challenge.inventoryItemId
                ? inventoryItems?.find(item => item.id === challenge.reward)?.id
                : undefined;
              const rewardItemId = challenge.inventoryItemId ?? legacyItemId;
              const rewardItem = rewardItemId ? inventoryItems?.find(item => item.id === rewardItemId) : undefined;
              return (
                <div key={challenge.id} className="qa-flow-challenge-card">
                  <div className="qa-flow-challenge-header">
                    <div>
                      <div className="qa-flow-challenge-title">Challenge {index + 1}</div>
                      <div className="qa-inline" style={{ marginTop: 6 }}>
                        {challenge.reward > 0 && (
                          <span className="qa-chip accent">+{challenge.reward} pts</span>
                        )}
                        {challenge.difficulty !== undefined && challenge.difficulty !== null && (
                          <span className="qa-chip muted">Difficulty: {challenge.difficulty}</span>
                        )}
                        {rewardItem && (
                          <span className="qa-chip success">{rewardItem.name}</span>
                        )}
                        {challenge.proficiency && (
                          <span className="qa-chip muted">Proficiency: {challenge.proficiency}</span>
                        )}
                      </div>
                    </div>
                    <button className="qa-btn qa-btn-ghost" onClick={() => onEditChallenge(challenge)}>
                      Edit
                    </button>
                  </div>

                  {challenge.unlockedNode ? (
                    <div className="qa-flow-branch">
                      <div className="qa-flow-branch-label">Unlocks</div>
                      <FlowNode
                        node={challenge.unlockedNode}
                        locationArchetypes={locationArchetypes}
                        inventoryItems={inventoryItems}
                        depth={depth + 1}
                        proficiencyOptions={proficiencyOptions}
                        onProficiencySearchChange={onProficiencySearchChange}
                        addChallengeToQuestArchetype={addChallengeToQuestArchetype}
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
  locationArchetypes: LocationArchetype[]
): FlowMapLayout | null => {
  if (!root) return null;
  const nodes: FlowMapNode[] = [];
  const edges: FlowMapEdge[] = [];
  const visited = new Set<string>();
  let orderIndex = 0;
  let maxDepth = 0;

  const locationName = (node: QuestArchetypeNode) =>
    locationArchetypes.find((la) => la.id === node.locationArchetypeId)?.name ?? 'Unknown';

  const walk = (node: QuestArchetypeNode, depth: number) => {
    maxDepth = Math.max(maxDepth, depth);
    if (!visited.has(node.id)) {
      visited.add(node.id);
      nodes.push({
        id: node.id,
        depth,
        order: orderIndex,
        label: locationName(node),
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
    .filter(Boolean) as Array<{ fromX: number; fromY: number; toX: number; toY: number }>;

  const width = padding * 2 + Math.max(1, maxDepth + 1) * xSpacing;
  const height = padding * 2 + Math.max(1, nodes.length) * ySpacing;

  return {
    nodes: positionedNodes,
    edges: positionedEdges,
    width,
    height,
  };
};

export const QuestArchetypeComponent = () => {
  const { apiClient } = useAPI();
  const {
    questArchetypes,
    locationArchetypes,
    createQuestArchetype,
    updateQuestArchetype,
    deleteQuestArchetype,
    addChallengeToQuestArchetype,
    updateQuestArchetypeChallenge,
    deleteQuestArchetypeChallenge,
  } = useQuestArchetypes();
  const [inventoryItems, setInventoryItems] = useState<InventoryItem[]>([]);
  const [inventoryItemsLoading, setInventoryItemsLoading] = useState<boolean>(false);
  const [shouldShowModal, setShouldShowModal] = useState(false);
  const [name, setName] = useState("");
  const [locationArchetypeId, setLocationArchetypeId] = useState("");
  const [locationArchetypeQuery, setLocationArchetypeQuery] = useState("");
  const [defaultGold, setDefaultGold] = useState<number>(0);
  const [ archetypeRewards, setArchetypeRewards ] = useState<{ inventoryItemId: string; quantity: number }[]>([]);
  const [ editingArchetype, setEditingArchetype ] = useState<QuestArchetype | null>(null);
  const [ editGold, setEditGold ] = useState<number>(0);
  const [ editRewards, setEditRewards ] = useState<{ inventoryItemId: string; quantity: number }[]>([]);
  const [ editingChallenge, setEditingChallenge ] = useState<QuestArchetypeChallenge | null>(null);
  const [ editChallengeRewardPoints, setEditChallengeRewardPoints ] = useState<number>(0);
  const [ editChallengeRewardItemId, setEditChallengeRewardItemId ] = useState<number>(0);
  const [ editChallengeProficiency, setEditChallengeProficiency ] = useState<string>("");
  const [ editChallengeDifficulty, setEditChallengeDifficulty ] = useState<number>(0);
  const [ proficiencySearch, setProficiencySearch ] = useState<string>("");
  const [ proficiencyOptions, setProficiencyOptions ] = useState<string[]>([]);
  const [ archetypeSearch, setArchetypeSearch ] = useState<string>("");
  const [ selectedArchetypeId, setSelectedArchetypeId ] = useState<string>("");

  const filteredLocationArchetypes = locationArchetypes
    .filter((archetype) =>
      archetype.name.toLowerCase().includes(locationArchetypeQuery.trim().toLowerCase())
    )
    .slice(0, 8);

  const filteredArchetypes = useMemo(
    () =>
      questArchetypes.filter((archetype) =>
        archetype.name.toLowerCase().includes(archetypeSearch.trim().toLowerCase())
      ),
    [questArchetypes, archetypeSearch]
  );

  const selectedArchetype = useMemo(
    () => questArchetypes.find((archetype) => archetype.id === selectedArchetypeId) ?? null,
    [questArchetypes, selectedArchetypeId]
  );

  const flowMapLayout = useMemo(
    () => buildFlowMapLayout(selectedArchetype?.root ?? null, locationArchetypes),
    [selectedArchetype, locationArchetypes]
  );

  useEffect(() => {
    if (questArchetypes.length === 0) {
      setSelectedArchetypeId('');
      return;
    }
    const stillExists = questArchetypes.some((archetype) => archetype.id === selectedArchetypeId);
    if (!stillExists) {
      setSelectedArchetypeId(questArchetypes[0].id);
    }
  }, [questArchetypes, selectedArchetypeId]);

  useEffect(() => {
    const fetchInventoryItems = async () => {
      setInventoryItemsLoading(true);
      try {
        const response = await apiClient.get<InventoryItem[]>('/sonar/inventory-items');
        setInventoryItems(response);
      } catch (error) {
        console.error('Error fetching inventory items:', error);
      } finally {
        setInventoryItemsLoading(false);
      }
    };

    fetchInventoryItems();
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
              Build the backbone of every adventure. Define branching challenges, rewards, and proficiencies so
              generated quests feel crafted rather than random.
            </p>
          </div>
          <div className="qa-hero-actions">
            {inventoryItemsLoading && <span className="qa-chip muted">Loading items…</span>}
            <button className="qa-btn qa-btn-primary" onClick={() => setShouldShowModal(true)}>
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
                  <div className="qa-empty">No archetypes match that search.</div>
                ) : (
                  filteredArchetypes.map((questArchetype) => {
                    const rootLocation = locationArchetypes.find((la) =>
                      la.id === questArchetype.root?.locationArchetypeId
                    )?.name ?? 'Unknown';
                    return (
                      <button
                        key={questArchetype.id}
                        className={`qa-sidebar-item ${selectedArchetypeId === questArchetype.id ? 'is-active' : ''}`}
                        onClick={() => setSelectedArchetypeId(questArchetype.id)}
                      >
                        <div className="qa-sidebar-item-title">{questArchetype.name}</div>
                        <div className="qa-meta">
                          Root: {rootLocation} · {questArchetype.root?.challenges?.length ?? 0} challenges
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
                    <h2 className="qa-title" style={{ fontSize: 'clamp(26px, 3vw, 34px)' }}>
                      {selectedArchetype.name}
                    </h2>
                    <p className="qa-subtitle">
                      Craft the journey by stacking challenges and branching nodes. Each challenge can unlock a new
                      node to extend the quest.
                    </p>
                  </div>
                  <div className="qa-actions">
                    <button
                      className="qa-btn qa-btn-ghost"
                      onClick={() => {
                        setEditingArchetype(selectedArchetype);
                        setEditGold(selectedArchetype.defaultGold ?? 0);
                        setEditRewards(
                          (selectedArchetype.itemRewards ?? []).map((reward) => ({
                            inventoryItemId: reward.inventoryItemId ? String(reward.inventoryItemId) : '',
                            quantity: reward.quantity ?? 1,
                          }))
                        );
                      }}
                    >
                      Edit Rewards
                    </button>
                    <button
                      className="qa-btn qa-btn-danger"
                      onClick={() => {
                        if (window.confirm('Are you sure you want to delete this quest archetype?')) {
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
                      <div className="qa-stat-label">Default Gold</div>
                      <div className="qa-stat-value">{selectedArchetype.defaultGold ?? 0}</div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Root Location</div>
                      <div className="qa-stat-value">
                        {locationArchetypes.find((la) => la.id === selectedArchetype.root?.locationArchetypeId)?.name ?? 'Unknown'}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Challenges</div>
                      <div className="qa-stat-value">{selectedArchetype.root?.challenges?.length ?? 0}</div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Created</div>
                      <div className="qa-stat-value">
                        {new Date(selectedArchetype.createdAt).toLocaleDateString()}
                      </div>
                    </div>
                  </div>

                  <div className="qa-divider" />

                  <div className="qa-panel">
                    <div className="qa-meta">Quest Rewards</div>
                    {(selectedArchetype.itemRewards ?? []).length > 0 ? (
                      <div className="qa-inline" style={{ marginTop: 10 }}>
                        {(selectedArchetype.itemRewards ?? []).map((reward) => {
                          const item = inventoryItems.find((entry) => entry.id === reward.inventoryItemId);
                          return (
                            <span
                              key={reward.id ?? `${reward.inventoryItemId}-${reward.quantity}`}
                              className="qa-chip"
                            >
                              {reward.quantity}x {item?.name ?? `Item ${reward.inventoryItemId}`}
                            </span>
                          );
                        })}
                      </div>
                    ) : (
                      <div className="qa-empty" style={{ marginTop: 10 }}>
                        No item rewards configured.
                      </div>
                    )}
                  </div>
                </div>

                {flowMapLayout && (
                  <div className="qa-card qa-flow-map">
                    <div className="qa-card-title">Flow Map</div>
                    <p className="qa-muted" style={{ marginTop: 6 }}>
                      A mini-map of the quest flow. Each node represents a location archetype, connected by challenges.
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
                            <path d="M0,0 L0,6 L6,3 z" fill="rgba(255,255,255,0.65)" />
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
                    Start from the root and add challenges. Add a location to a challenge to branch into a new node.
                  </p>
                  {selectedArchetype.root ? (
                    <div className="qa-flow-canvas">
                      <FlowNode
                        node={selectedArchetype.root}
                        locationArchetypes={locationArchetypes}
                        inventoryItems={inventoryItems}
                        depth={0}
                        proficiencyOptions={proficiencyOptions}
                        onProficiencySearchChange={setProficiencySearch}
                        addChallengeToQuestArchetype={addChallengeToQuestArchetype}
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

      {shouldShowModal && (
        <div className="qa-modal">
          <div className="qa-modal-card">
            <h2 className="qa-modal-title">Create Quest Archetype</h2>
            <form
              className="qa-form-grid"
              onSubmit={(e) => {
                e.preventDefault();
                const normalizedRewards = archetypeRewards
                  .map((reward) => ({
                    inventoryItemId: Number(reward.inventoryItemId) || 0,
                    quantity: Number(reward.quantity) || 0,
                  }))
                  .filter((reward) => reward.inventoryItemId > 0 && reward.quantity > 0);
                createQuestArchetype(name, locationArchetypeId, defaultGold, normalizedRewards).then((created) => {
                  if (created?.id) {
                    setSelectedArchetypeId(created.id);
                  }
                });
                setArchetypeRewards([]);
                setLocationArchetypeId('');
                setLocationArchetypeQuery('');
                setShouldShowModal(false);
              }}
            >
              <div className="qa-field">
                <div className="qa-label">Name</div>
                <input
                  type="text"
                  className="qa-input"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </div>

              <div className="qa-field">
                <div className="qa-label">Location Archetype</div>
                <div className="qa-combobox">
                  <input
                    type="text"
                    className="qa-input"
                    value={locationArchetypeQuery}
                    onChange={(e) => {
                      const value = e.target.value;
                      setLocationArchetypeQuery(value);
                      const matched = locationArchetypes.find(
                        (archetype) => archetype.name.toLowerCase() === value.trim().toLowerCase()
                      );
                      setLocationArchetypeId(matched ? matched.id : '');
                    }}
                    placeholder="Search location archetypes..."
                  />
                  {locationArchetypeQuery.trim().length > 0 && (
                    <div className="qa-combobox-list">
                      {filteredLocationArchetypes.length === 0 ? (
                        <div className="qa-combobox-empty">No matches.</div>
                      ) : (
                        filteredLocationArchetypes.map((archetype) => (
                          <button
                            key={archetype.id}
                            type="button"
                            className="qa-combobox-option"
                            onClick={() => {
                              setLocationArchetypeId(archetype.id);
                              setLocationArchetypeQuery(archetype.name);
                            }}
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
                <div className="qa-label">Default Gold</div>
                <input
                  type="number"
                  min={0}
                  className="qa-input"
                  value={defaultGold}
                  onChange={(e) => setDefaultGold(parseInt(e.target.value) || 0)}
                />
              </div>

              <div className="qa-field">
                <div className="qa-label">Item Rewards</div>
                {archetypeRewards.length === 0 ? (
                  <div className="qa-empty">No item rewards yet.</div>
                ) : (
                  <div className="qa-form-grid">
                    {archetypeRewards.map((reward, index) => (
                      <div key={`reward-${index}`} className="qa-reward-row">
                        <select
                          className="qa-select"
                          value={reward.inventoryItemId}
                          onChange={(e) =>
                            setArchetypeRewards((prev) =>
                              prev.map((entry, rewardIndex) =>
                                rewardIndex === index ? { ...entry, inventoryItemId: e.target.value } : entry
                              )
                            )
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
                          value={reward.quantity}
                          onChange={(e) =>
                            setArchetypeRewards((prev) =>
                              prev.map((entry, rewardIndex) =>
                                rewardIndex === index
                                  ? { ...entry, quantity: parseInt(e.target.value) || 1 }
                                  : entry
                              )
                            )
                          }
                        />
                        <button
                          type="button"
                          className="qa-btn qa-btn-text"
                          onClick={() =>
                            setArchetypeRewards((prev) => prev.filter((_, rewardIndex) => rewardIndex !== index))
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
                  onClick={() => setArchetypeRewards((prev) => [...prev, { inventoryItemId: '', quantity: 1 }])}
                >
                  Add Item Reward
                </button>
              </div>

              <div className="qa-footer">
                <button
                  type="button"
                  className="qa-btn qa-btn-outline"
                  onClick={() => {
                    setShouldShowModal(false);
                    setArchetypeRewards([]);
                    setLocationArchetypeId('');
                    setLocationArchetypeQuery('');
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

      {editingArchetype && (
        <div className="qa-modal">
          <div className="qa-modal-card">
            <h2 className="qa-modal-title">Edit Quest Rewards</h2>
            <div className="qa-form-grid">
              <div className="qa-field">
                <div className="qa-label">Default Gold</div>
                <input
                  type="number"
                  min={0}
                  className="qa-input"
                  value={editGold}
                  onChange={(e) => setEditGold(parseInt(e.target.value) || 0)}
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Item Rewards</div>
                {editRewards.length === 0 ? (
                  <div className="qa-empty">No item rewards yet.</div>
                ) : (
                  <div className="qa-form-grid">
                    {editRewards.map((reward, index) => (
                      <div key={`edit-reward-${index}`} className="qa-reward-row">
                        <select
                          className="qa-select"
                          value={reward.inventoryItemId}
                          onChange={(e) =>
                            setEditRewards((prev) =>
                              prev.map((entry, rewardIndex) =>
                                rewardIndex === index ? { ...entry, inventoryItemId: e.target.value } : entry
                              )
                            )
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
                          value={reward.quantity}
                          onChange={(e) =>
                            setEditRewards((prev) =>
                              prev.map((entry, rewardIndex) =>
                                rewardIndex === index
                                  ? { ...entry, quantity: parseInt(e.target.value) || 1 }
                                  : entry
                              )
                            )
                          }
                        />
                        <button
                          type="button"
                          className="qa-btn qa-btn-text"
                          onClick={() =>
                            setEditRewards((prev) => prev.filter((_, rewardIndex) => rewardIndex !== index))
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
                  onClick={() => setEditRewards((prev) => [...prev, { inventoryItemId: '', quantity: 1 }])}
                >
                  Add Item Reward
                </button>
              </div>
            </div>
            <div className="qa-footer">
              <button
                className="qa-btn qa-btn-outline"
                onClick={() => {
                  setEditingArchetype(null);
                  setEditRewards([]);
                }}
              >
                Cancel
              </button>
              <button
                className="qa-btn qa-btn-primary"
                onClick={async () => {
                  const normalizedRewards = editRewards
                    .map((reward) => ({
                      inventoryItemId: Number(reward.inventoryItemId) || 0,
                      quantity: Number(reward.quantity) || 0,
                    }))
                    .filter((reward) => reward.inventoryItemId > 0 && reward.quantity > 0);
                  await updateQuestArchetype({
                    ...editingArchetype,
                    defaultGold: editGold,
                    itemRewards: normalizedRewards,
                  });
                  setEditingArchetype(null);
                  setEditRewards([]);
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
                  onChange={(e) => setEditChallengeRewardPoints(parseInt(e.target.value) || 0)}
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Difficulty</div>
                <input
                  type="number"
                  min={0}
                  className="qa-input"
                  value={editChallengeDifficulty}
                  onChange={(e) => setEditChallengeDifficulty(parseInt(e.target.value) || 0)}
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Reward Item</div>
                <select
                  className="qa-select"
                  value={editChallengeRewardItemId || ''}
                  onChange={(e) => setEditChallengeRewardItemId(parseInt(e.target.value) || 0)}
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
              <button className="qa-btn qa-btn-outline" onClick={() => setEditingChallenge(null)}>
                Cancel
              </button>
              <button
                className="qa-btn qa-btn-danger"
                onClick={async () => {
                  if (!editingChallenge) return;
                  const confirmDelete = window.confirm('Delete this challenge? This cannot be undone.');
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
                    inventoryItemId: editChallengeRewardItemId > 0 ? editChallengeRewardItemId : null,
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
