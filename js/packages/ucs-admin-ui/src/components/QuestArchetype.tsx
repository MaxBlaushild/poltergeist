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

  return (
    <div className={`border-l-2 ${borderColor} pl-4 mt-2`}>
      <div className={`${bgColor} p-3 rounded-md`}>
        <div className="text-sm">
          <span className="font-medium">
            {depth === 0 ? 'Challenge' : 'Sub-Challenge'} {index + 1}
          </span>
          <div className="text-gray-600">Reward Item: {inventoryItems?.find(item => item.id === challenge.reward)?.name}</div>
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
  const { questArchetypes, locationArchetypes, createQuestArchetype, deleteQuestArchetype, addChallengeToQuestArchetype } = useQuestArchetypes();
  const [inventoryItems, setInventoryItems] = useState<InventoryItem[]>([]);
  const [inventoryItemsLoading, setInventoryItemsLoading] = useState<boolean>(false);
  const [shouldShowModal, setShouldShowModal] = useState(false);
  const [selectedQuestArchetype, setSelectedQuestArchetype] = useState<QuestArchetype | null>(null);
  const [name, setName] = useState("");
  const [locationArchetypeId, setLocationArchetypeId] = useState("");
  const [defaultGold, setDefaultGold] = useState<number>(0);
  const [ selectedNode, setSelectedNode ] = useState<QuestArchetypeNode | null>(null);
  const [ reward, setReward ] = useState<number>(0);
  const [ unlockedLocationArchetypeId, setUnlockedLocationArchetypeId ] = useState<string>("");

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

          <div className="mt-2 text-sm text-gray-500">
            Created: {new Date(questArchetype.createdAt).toLocaleDateString()}
          </div>
          <button className="bg-blue-500 text-white px-4 py-2 rounded-md" onClick={() => 
            setSelectedNode(questArchetype.root)}>Edit Root Node</button>
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
                    Reward Item
                  </label>
                  <select
                    className="w-full px-3 py-2 border border-gray-300 rounded-md"
                    value={reward}
                    onChange={(e) => setReward(parseInt(e.target.value))}
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
                    await addChallengeToQuestArchetype(selectedNode.id, reward, unlockedLocationArchetypeId);
                    setReward(0);
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
            createQuestArchetype(name, locationArchetypeId, defaultGold);
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

            <div className="flex justify-end gap-2">
              <button
                type="button"
                className="px-4 py-2 text-gray-600 hover:text-gray-800"
                onClick={() => setShouldShowModal(false)}
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
  </div>;
};
