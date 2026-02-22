import React, { useState, useEffect } from "react";
import { useAPI } from "@poltergeist/contexts";
import { useQuestArchetypes } from "../contexts/questArchetypes.tsx";
import { LocationArchetype, QuestArchetype, QuestArchetypeNode, QuestArchetypeChallenge, InventoryItem } from "@poltergeist/types";

interface ChallengeNodeProps {
  challenge: QuestArchetypeChallenge;
  index: number;
  locationArchetypes: LocationArchetype[];
  depth: number;
  inventoryItems: InventoryItem[];
  onEdit: (challenge: QuestArchetypeNode) => void;
}

const ChallengeNode: React.FC<ChallengeNodeProps> = ({ challenge, index, locationArchetypes, depth, inventoryItems, onEdit }) => {
  const borderColors = [
    'border-gray-200',
    'border-blue-200',
    'border-green-200',
    'border-purple-200',
    'border-yellow-200',
  ];
  const bgColors = [
    'bg-gray-50',
    'bg-blue-50',
    'bg-green-50',
    'bg-purple-50',
    'bg-yellow-50',
  ];

  const borderColor = borderColors[depth % borderColors.length];
  const bgColor = bgColors[depth % bgColors.length];
  const legacyItemId = !challenge.inventoryItemId
    ? inventoryItems?.find(item => item.id === challenge.reward)?.id
    : undefined;
  const rewardItemId = challenge.inventoryItemId ?? legacyItemId;
  const rewardItem = rewardItemId ? inventoryItems?.find(item => item.id === rewardItemId) : undefined;

  return (
    <div className={`border-l-2 ${borderColor} pl-4 mt-2`}>
      <div className={`${bgColor} p-3 rounded-md`}>
        <div className="text-sm">
          <span className="font-medium">
            {depth === 0 ? 'Challenge' : 'Sub-Challenge'} {index + 1}
          </span>
          {challenge.reward > 0 && (
            <div className="text-gray-600">Reward Points: {challenge.reward}</div>
          )}
          {rewardItem && (
            <div className="text-gray-600">Reward Item: {rewardItem.name}</div>
          )}
          {challenge.proficiency && (
            <div className="text-gray-600">Proficiency: {challenge.proficiency}</div>
          )}
          {challenge.unlockedNode && (
            <div className="mt-2">
              <div className="text-gray-600 font-medium">Unlocks Node:</div>
              <div className="bg-white p-3 rounded-md mt-1 shadow-sm">
                <div className="text-gray-700">
                  Location Type: {locationArchetypes.find(la => 
                    la.id === challenge.unlockedNode?.locationArchetypeId
                  )?.name}
                </div>
                <button className="bg-blue-500 text-white px-4 py-2 rounded-md" onClick={() => onEdit(challenge.unlockedNode!)}>Edit</button>
                {challenge.unlockedNode.challenges?.map((subChallenge, i) => (
                  <ChallengeNode
                    key={subChallenge.id}
                    challenge={subChallenge}
                    index={i}
                    locationArchetypes={locationArchetypes}
                    depth={depth + 1}
                    inventoryItems={inventoryItems}
                    onEdit={onEdit}
                  />
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export const QuestArchetypeComponent = () => {
  const { apiClient } = useAPI();
  const { questArchetypes, locationArchetypes, createQuestArchetype, updateQuestArchetype, deleteQuestArchetype, addChallengeToQuestArchetype } = useQuestArchetypes();
  const [inventoryItems, setInventoryItems] = useState<InventoryItem[]>([]);
  const [inventoryItemsLoading, setInventoryItemsLoading] = useState<boolean>(false);
  const [shouldShowModal, setShouldShowModal] = useState(false);
  const [selectedQuestArchetype, setSelectedQuestArchetype] = useState<QuestArchetype | null>(null);
  const [name, setName] = useState("");
  const [locationArchetypeId, setLocationArchetypeId] = useState("");
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

  return <div>
    <h1 className="text-2xl font-bold">Quest Archetype</h1>
    <div className="grid grid-cols-1 gap-4 mt-4 mb-8">
      {questArchetypes.map((questArchetype) => (
        <div 
          key={questArchetype.id}
          className="bg-white shadow rounded-lg p-6 border border-gray-200"
        >
          <div className="flex justify-between items-center">
            <h3 className="text-xl font-semibold text-gray-900">{questArchetype.name}</h3>
            <button
              onClick={() => {
                if (window.confirm('Are you sure you want to delete this quest archetype?')) {
                  deleteQuestArchetype(questArchetype.id);
                }
              }}
              className="bg-red-500 hover:bg-red-600 text-white px-3 py-1 rounded-md text-sm"
            >
              Delete
            </button>
          </div>
          
          <div className="mt-4 bg-gray-50 p-4 rounded-md">
            <h4 className="font-medium text-gray-700 mb-2">Root Node</h4>
            <div className="space-y-2">
              <div>
                <span className="text-sm text-gray-500">Location Type: </span>
                <span className="text-sm font-medium text-gray-900">
                  {locationArchetypes.find(la => la.id === questArchetype.root?.locationArchetypeId)?.name || 'Unknown'}
                </span>
              </div>
              
              <div>
                <span className="text-sm text-gray-500">Challenges: </span>
                <span className="text-sm font-medium text-gray-900">
                  {questArchetype.root?.challenges?.length || 0}
                </span>
              </div>

              <div>
                <span className="text-sm text-gray-500">Node ID: </span>
                <span className="text-sm font-medium text-gray-900">
                  {questArchetype.root?.id}
                </span>
              </div>
            </div>
          </div>

          <div className="mt-4">
            <h4 className="font-medium text-gray-700 mb-2">Challenge Tree</h4>
            {questArchetype.root && (
              <div className="pl-4">
                {questArchetype.root.challenges?.map((challenge, i) => (
                  <ChallengeNode
                    key={challenge.id}
                    challenge={challenge}
                    index={i}
                    locationArchetypes={locationArchetypes}
                    depth={0}
                    inventoryItems={inventoryItems}
                    onEdit={() => setSelectedNode(challenge.unlockedNode!)}
                  />
                ))}
              </div>
            )}
          </div>

          <div className="mt-4 bg-gray-50 p-4 rounded-md">
            <h4 className="font-medium text-gray-700 mb-2">Rewards</h4>
            <div className="text-sm text-gray-600">Default Gold: {questArchetype.defaultGold ?? 0}</div>
            {questArchetype.itemRewards && questArchetype.itemRewards.length > 0 ? (
              <div className="mt-2 text-sm text-gray-600">
                {questArchetype.itemRewards.map((reward) => {
                  const item = inventoryItems.find((entry) => entry.id === reward.inventoryItemId);
                  return (
                    <div key={reward.id ?? `${reward.inventoryItemId}-${reward.quantity}`}>
                      {reward.quantity}x {item?.name ?? `Item ${reward.inventoryItemId}`}
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className="mt-2 text-sm text-gray-500">No item rewards.</div>
            )}
          </div>

          <div className="mt-2 text-sm text-gray-500">
            Created: {new Date(questArchetype.createdAt).toLocaleDateString()}
          </div>
          <div className="mt-3 flex gap-2">
            <button className="bg-blue-500 text-white px-4 py-2 rounded-md" onClick={() => 
              setSelectedNode(questArchetype.root)}>Edit Root Node</button>
            <button
              className="bg-gray-600 text-white px-4 py-2 rounded-md"
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
          </div>
        </div>
      ))}
    </div>
    <button className="bg-blue-500 text-white px-4 py-2 rounded-md" onClick={() => setShouldShowModal(true)}>Create Quest Archetype</button>
    {selectedNode && (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
        <div className="bg-white p-6 rounded-lg w-96">
          <h2 className="text-xl font-bold mb-4">Edit Root Node</h2>
          <div className="space-y-2">
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Add Challenge
              </label>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm text-gray-600 mb-1">
                    Reward Points
                  </label>
                  <input
                    type="number"
                    min={0}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    value={rewardPoints}
                    onChange={(e) => setRewardPoints(parseInt(e.target.value) || 0)}
                  />
                </div>

                <div>
                  <label className="block text-sm text-gray-600 mb-1">
                    Reward Item
                  </label>
                  <select
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    value={rewardItemId}
                    onChange={(e) => setRewardItemId(parseInt(e.target.value))}
                  >
                    <option value="">Select an item</option>
                    {inventoryItems.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.name}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm text-gray-600 mb-1">
                    Proficiency
                  </label>
                  <input
                    type="text"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    value={challengeProficiency}
                    onChange={(e) => setChallengeProficiency(e.target.value)}
                    placeholder="Optional proficiency (e.g. Persuasion)"
                  />
                </div>

                <div>
                  <label className="block text-sm text-gray-600 mb-1">
                    Unlocked Location Type (Optional)
                  </label>
                  <select
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
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
                  className="bg-green-500 text-white px-4 py-2 rounded-md"
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
            </div>
            <button className="bg-blue-500 text-white px-4 py-2 rounded-md" onClick={() => 
              setSelectedNode(null)}>Cancel</button>
            <button className="bg-blue-500 text-white px-4 py-2 rounded-md" onClick={() => 
              setSelectedNode(null)}>Save</button>
          </div>
        </div>
      </div>
    )}
    {shouldShowModal && (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
        <div className="bg-white p-6 rounded-lg w-96">
          <h2 className="text-xl font-bold mb-4">Create Quest Archetype</h2>
          <form onSubmit={(e) => {
            e.preventDefault();
            const normalizedRewards = archetypeRewards
              .map((reward) => ({
                inventoryItemId: Number(reward.inventoryItemId) || 0,
                quantity: Number(reward.quantity) || 0,
              }))
              .filter((reward) => reward.inventoryItemId > 0 && reward.quantity > 0);
            createQuestArchetype(name, locationArchetypeId, defaultGold, normalizedRewards);
            setArchetypeRewards([]);
            setShouldShowModal(false);
          }}>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Name
              </label>
              <input
                type="text"
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Location Archetype
              </label>
              <select
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                value={locationArchetypeId}
                onChange={(e) => setLocationArchetypeId(e.target.value)}
              >
                <option value="">Select a location archetype</option>
                {locationArchetypes.map((archetype) => (
                  <option key={archetype.id} value={archetype.id}>
                    {archetype.name}
                  </option>
                ))}
              </select>
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Default Gold
              </label>
              <input
                type="number"
                min={0}
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                value={defaultGold}
                onChange={(e) => setDefaultGold(parseInt(e.target.value) || 0)}
              />
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Item Rewards
              </label>
              {archetypeRewards.length === 0 ? (
                <div className="text-xs text-gray-500">No item rewards yet.</div>
              ) : (
                <div className="space-y-2">
                  {archetypeRewards.map((reward, index) => (
                    <div key={`reward-${index}`} className="flex gap-2">
                      <select
                        className="flex-1 px-3 py-2 border border-gray-300 rounded-md"
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
                        className="w-20 px-2 py-2 border border-gray-300 rounded-md"
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
                        className="px-2 py-2 text-sm text-red-600"
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
                className="mt-2 px-3 py-1 text-sm text-blue-600"
                onClick={() => setArchetypeRewards((prev) => [...prev, { inventoryItemId: '', quantity: 1 }])}
              >
                Add Item Reward
              </button>
            </div>

            <div className="flex justify-end gap-2">
              <button
                type="button"
                className="px-4 py-2 text-gray-600 hover:text-gray-800"
                onClick={() => {
                  setShouldShowModal(false);
                  setArchetypeRewards([]);
                }}
              >
                Cancel
              </button>
              <button
                type="submit"
                className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600"
              >
                Create
              </button>
            </div>
          </form>
        </div>
      </div>
    )}
    {editingArchetype && (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
        <div className="bg-white p-6 rounded-lg w-96">
          <h2 className="text-xl font-bold mb-4">Edit Quest Rewards</h2>
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Default Gold
            </label>
            <input
              type="number"
              min={0}
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
              value={editGold}
              onChange={(e) => setEditGold(parseInt(e.target.value) || 0)}
            />
          </div>
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Item Rewards
            </label>
            {editRewards.length === 0 ? (
              <div className="text-xs text-gray-500">No item rewards yet.</div>
            ) : (
              <div className="space-y-2">
                {editRewards.map((reward, index) => (
                  <div key={`edit-reward-${index}`} className="flex gap-2">
                    <select
                      className="flex-1 px-3 py-2 border border-gray-300 rounded-md"
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
                      className="w-20 px-2 py-2 border border-gray-300 rounded-md"
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
                      className="px-2 py-2 text-sm text-red-600"
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
              className="mt-2 px-3 py-1 text-sm text-blue-600"
              onClick={() => setEditRewards((prev) => [...prev, { inventoryItemId: '', quantity: 1 }])}
            >
              Add Item Reward
            </button>
          </div>
          <div className="flex justify-end gap-2">
            <button
              className="px-4 py-2 text-gray-600 hover:text-gray-800"
              onClick={() => {
                setEditingArchetype(null);
                setEditRewards([]);
              }}
            >
              Cancel
            </button>
            <button
              className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600"
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
  </div>;
};
