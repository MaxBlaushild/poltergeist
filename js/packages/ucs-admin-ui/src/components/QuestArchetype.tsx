import React, { useState, useEffect } from "react";
import { useAPI } from "@poltergeist/contexts";
import { useQuestArchetypes } from "../contexts/questArchetypes.tsx";
import { LocationArchetype, QuestArchetype, QuestArchetypeNode, QuestArchetypeChallenge, InventoryItem } from "@poltergeist/types";
import "./questArchetypeTheme.css";

interface ChallengeNodeProps {
  challenge: QuestArchetypeChallenge;
  index: number;
  locationArchetypes: LocationArchetype[];
  depth: number;
  inventoryItems: InventoryItem[];
  onEditNode: (node: QuestArchetypeNode) => void;
  onEditChallenge: (challenge: QuestArchetypeChallenge) => void;
}

const ChallengeNode: React.FC<ChallengeNodeProps> = ({
  challenge,
  index,
  locationArchetypes,
  depth,
  inventoryItems,
  onEditNode,
  onEditChallenge,
}) => {
  const borderColor = depth % 2 === 0 ? 'rgba(255, 107, 74, 0.4)' : 'rgba(95, 211, 181, 0.35)';
  const legacyItemId = !challenge.inventoryItemId
    ? inventoryItems?.find(item => item.id === challenge.reward)?.id
    : undefined;
  const rewardItemId = challenge.inventoryItemId ?? legacyItemId;
  const rewardItem = rewardItemId ? inventoryItems?.find(item => item.id === rewardItemId) : undefined;

  return (
    <div className="qa-node" style={{ borderColor }}>
      <div className="qa-node-card">
        <div className="qa-node-title">
          {depth === 0 ? 'Challenge' : 'Sub-Challenge'} {index + 1}
        </div>
        <div className="qa-inline">
          {challenge.reward > 0 && (
            <span className="qa-chip accent">+{challenge.reward} pts</span>
          )}
          {rewardItem && (
            <span className="qa-chip success">{rewardItem.name}</span>
          )}
          {challenge.proficiency && (
            <span className="qa-chip muted">Proficiency: {challenge.proficiency}</span>
          )}
        </div>
        <div className="qa-inline" style={{ marginTop: 8 }}>
          <button
            className="qa-btn qa-btn-ghost"
            onClick={() => onEditChallenge(challenge)}
          >
            Edit Challenge
          </button>
        </div>
        {challenge.unlockedNode && (
          <div className="qa-panel" style={{ marginTop: 12 }}>
            <div className="qa-meta">Unlocks Node</div>
            <div className="qa-inline" style={{ marginTop: 8 }}>
              <span className="qa-chip">
                {locationArchetypes.find(la =>
                  la.id === challenge.unlockedNode?.locationArchetypeId
                )?.name ?? 'Unknown location'}
              </span>
              <button className="qa-btn qa-btn-outline" onClick={() => onEditNode(challenge.unlockedNode!)}>
                Edit Node
              </button>
            </div>
            <div className="qa-tree" style={{ marginTop: 12 }}>
              {challenge.unlockedNode.challenges?.map((subChallenge, i) => (
                <ChallengeNode
                  key={subChallenge.id}
                  challenge={subChallenge}
                  index={i}
                  locationArchetypes={locationArchetypes}
                  depth={depth + 1}
                  inventoryItems={inventoryItems}
                  onEditNode={onEditNode}
                  onEditChallenge={onEditChallenge}
                />
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
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
  } = useQuestArchetypes();
  const [inventoryItems, setInventoryItems] = useState<InventoryItem[]>([]);
  const [inventoryItemsLoading, setInventoryItemsLoading] = useState<boolean>(false);
  const [shouldShowModal, setShouldShowModal] = useState(false);
  const [name, setName] = useState("");
  const [locationArchetypeId, setLocationArchetypeId] = useState("");
  const [locationArchetypeQuery, setLocationArchetypeQuery] = useState("");
  const [defaultGold, setDefaultGold] = useState<number>(0);
  const [ selectedNode, setSelectedNode ] = useState<QuestArchetypeNode | null>(null);
  const [ rewardPoints, setRewardPoints ] = useState<number>(0);
  const [ rewardItemId, setRewardItemId ] = useState<number>(0);
  const [ challengeProficiency, setChallengeProficiency ] = useState<string>("");
  const [ unlockedLocationArchetypeId, setUnlockedLocationArchetypeId ] = useState<string>("");
  const [ archetypeRewards, setArchetypeRewards ] = useState<{ inventoryItemId: string; quantity: number }[]>([]);
  const [ editingArchetype, setEditingArchetype ] = useState<QuestArchetype | null>(null);
  const [ editGold, setEditGold ] = useState<number>(0);
  const [ editRewards, setEditRewards ] = useState<{ inventoryItemId: string; quantity: number }[]>([]);
  const [ editingChallenge, setEditingChallenge ] = useState<QuestArchetypeChallenge | null>(null);
  const [ editChallengeRewardPoints, setEditChallengeRewardPoints ] = useState<number>(0);
  const [ editChallengeRewardItemId, setEditChallengeRewardItemId ] = useState<number>(0);
  const [ editChallengeProficiency, setEditChallengeProficiency ] = useState<string>("");
  const [ proficiencySearch, setProficiencySearch ] = useState<string>("");
  const [ proficiencyOptions, setProficiencyOptions ] = useState<string[]>([]);

  const filteredLocationArchetypes = locationArchetypes
    .filter((archetype) =>
      archetype.name.toLowerCase().includes(locationArchetypeQuery.trim().toLowerCase())
    )
    .slice(0, 8);

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

        <section className="qa-grid">
          {questArchetypes.length === 0 ? (
            <div className="qa-panel">
              <div className="qa-card-title">No archetypes yet</div>
              <p className="qa-muted" style={{ marginTop: 8 }}>
                Create the first quest archetype to start generating quest chains for a zone.
              </p>
            </div>
          ) : (
            questArchetypes.map((questArchetype, index) => {
              const rootLocation = locationArchetypes.find((la) =>
                la.id === questArchetype.root?.locationArchetypeId
              )?.name ?? 'Unknown';
              const nodeId = questArchetype.root?.id ?? '';
              const nodeIdShort = nodeId ? `${nodeId.slice(0, 8)}…` : '—';
              const rewards = questArchetype.itemRewards ?? [];
              return (
                <article
                  key={questArchetype.id}
                  className="qa-card"
                  style={{ animationDelay: `${index * 0.06}s` }}
                >
                  <div className="qa-card-header">
                    <div>
                      <h3 className="qa-card-title">{questArchetype.name}</h3>
                      <div className="qa-meta">
                        Root: {rootLocation} · {questArchetype.root?.challenges?.length || 0} challenges
                      </div>
                    </div>
                    <div className="qa-actions">
                      <button
                        className="qa-btn qa-btn-ghost"
                        onClick={() => {
                          setEditingArchetype(questArchetype);
                          setEditGold(questArchetype.defaultGold ?? 0);
                          setEditRewards(
                            (questArchetype.itemRewards ?? []).map((reward) => ({
                              inventoryItemId: reward.inventoryItemId ? String(reward.inventoryItemId) : '',
                              quantity: reward.quantity ?? 1,
                            }))
                          );
                        }}
                      >
                        Edit Rewards
                      </button>
                      <button className="qa-btn qa-btn-outline" onClick={() => setSelectedNode(questArchetype.root)}>
                        Edit Root
                      </button>
                      <button
                        className="qa-btn qa-btn-danger"
                        onClick={() => {
                          if (window.confirm('Are you sure you want to delete this quest archetype?')) {
                            deleteQuestArchetype(questArchetype.id);
                          }
                        }}
                      >
                        Delete
                      </button>
                    </div>
                  </div>

                  <div className="qa-stat-grid">
                    <div className="qa-stat">
                      <div className="qa-stat-label">Default Gold</div>
                      <div className="qa-stat-value">{questArchetype.defaultGold ?? 0}</div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Root Node</div>
                      <div className="qa-stat-value">{rootLocation}</div>
                    </div>
                    <div className="qa-stat" title={nodeId}>
                      <div className="qa-stat-label">Node ID</div>
                      <div className="qa-stat-value">{nodeIdShort}</div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Created</div>
                      <div className="qa-stat-value">
                        {new Date(questArchetype.createdAt).toLocaleDateString()}
                      </div>
                    </div>
                  </div>

                  <div className="qa-divider" />

                  <div className="qa-tree">
                    <div className="qa-meta">Challenge Tree</div>
                    {questArchetype.root?.challenges?.length ? (
                      questArchetype.root.challenges.map((challenge, i) => (
                        <ChallengeNode
                          key={challenge.id}
                          challenge={challenge}
                          index={i}
                          locationArchetypes={locationArchetypes}
                          depth={0}
                          inventoryItems={inventoryItems}
                          onEditNode={() => setSelectedNode(challenge.unlockedNode!)}
                          onEditChallenge={(selected) => {
                            setEditingChallenge(selected);
                            setEditChallengeRewardPoints(selected.reward ?? 0);
                            const itemId = selected.inventoryItemId ?? 0;
                            setEditChallengeRewardItemId(itemId);
                            setEditChallengeProficiency(selected.proficiency ?? '');
                            setProficiencySearch(selected.proficiency ?? '');
                          }}
                        />
                      ))
                    ) : (
                      <div className="qa-empty">No challenges yet. Add one from the node editor.</div>
                    )}
                  </div>

                  <div className="qa-divider" />

                  <div className="qa-panel">
                    <div className="qa-meta">Quest Rewards</div>
                    {rewards.length > 0 ? (
                      <div className="qa-inline" style={{ marginTop: 10 }}>
                        {rewards.map((reward) => {
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
                </article>
              );
            })
          )}
        </section>
      </div>

      {selectedNode && (
        <div className="qa-modal">
          <div className="qa-modal-card">
            <h2 className="qa-modal-title">Edit Quest Node</h2>
            <div className="qa-form-grid">
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
                    setProficiencySearch(e.target.value);
                  }}
                  list="qa-proficiency-options"
                  placeholder="Optional proficiency (e.g. Persuasion)"
                />
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

              <button
                type="button"
                className="qa-btn qa-btn-primary"
                onClick={async () => {
                  if (!selectedNode) return;
                  await addChallengeToQuestArchetype(
                    selectedNode.id,
                    rewardPoints,
                    rewardItemId || null,
                    challengeProficiency,
                    unlockedLocationArchetypeId
                  );
                  setRewardPoints(0);
                  setRewardItemId(0);
                  setChallengeProficiency("");
                  setUnlockedLocationArchetypeId("");
                }}
              >
                Add Challenge
              </button>
            </div>
            <div className="qa-footer">
              <button className="qa-btn qa-btn-outline" onClick={() => setSelectedNode(null)}>
                Close
              </button>
            </div>
          </div>
        </div>
      )}

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
                createQuestArchetype(name, locationArchetypeId, defaultGold, normalizedRewards);
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
                className="qa-btn qa-btn-primary"
                onClick={async () => {
                  const trimmed = editChallengeProficiency.trim();
                  await updateQuestArchetypeChallenge(editingChallenge.id, {
                    reward: editChallengeRewardPoints,
                    inventoryItemId: editChallengeRewardItemId > 0 ? editChallengeRewardItemId : null,
                    proficiency: trimmed.length > 0 ? trimmed : null,
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
