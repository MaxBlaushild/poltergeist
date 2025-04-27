import { useAPI, useArena, useInventory, useMediaContext } from '@poltergeist/contexts';
import { usePointOfInterestGroups } from '@poltergeist/hooks';
import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
import { AddNewPointOfInterest } from './AddNewPointOfInterest.tsx';
import { PointOfInterestChallenge, PointOfInterestGroupType } from '@poltergeist/types';
import { useTagContext } from '@poltergeist/contexts';

export const Arena = () => {
  const { 
    arena, 
    loading, 
    error, 
    updateArena, 
    createPointOfInterest, 
    deletePointOfInterest,
    updatePointOfInterest,
    updatePointOfInterestImage, 
    updateArenaImage, 
    createPointOfInterestChallenge,
    deletePointOfInterestChallenge,
    updatePointOfInterestChallenge,
    createPointOfInterestChildren,
    deletePointOfInterestChildren,
    addTagToPointOfInterest,
    removeTagFromPointOfInterest
  } = useArena();
  const [selectedTagId, setSelectedTagId] = useState<string | null>(null);
  const { inventoryItems } = useInventory();
  const [editingArena, setEditingArena] = useState(false);
  const [editingArenaImage, setEditingArenaImage] = useState(false);
  const [editingPointId, setEditingPointId] = useState<string | null>(null);
  const [editingPointImageId, setEditingPointImageId] = useState<string | null>(null);
  const [arenaName, setArenaName] = useState(arena?.name);
  const [arenaDescription, setArenaDescription] = useState(arena?.description);
  const [editedPoint, setEditedPoint] = useState<any>(null);
  const [showNewPointModal, setShowNewPointModal] = useState(false);
  const [showNewChallengeModal, setShowNewChallengeModal] = useState(false);
  const [showNewChildModal, setShowNewChildModal] = useState(false);
  const [selectedPointId, setSelectedPointId] = useState<string | null>(null);
  const [selectedChallenge, setSelectedChallenge] = useState<PointOfInterestChallenge | null>(null);
  const [showEditChallengeModal, setShowEditChallengeModal] = useState(false);
  const [selectedImage, setSelectedImage] = useState<string | null>(null);
  const [arenaType, setArenaType] = useState<PointOfInterestGroupType | undefined>(arena?.type);
  const [newChild, setNewChild] = useState({
    pointOfInterestId: '',
    pointOfInterestChallengeId: '',
    pointOfInterestGroupMemberId: '',
  });
  const [newChildChallengeSelection, setNewChildChallengeSelection] = useState<PointOfInterestChallenge[]>([]);
  const [newChallenge, setNewChallenge] = useState<PointOfInterestChallenge>({
    question: '',
    inventoryItemId: 0,
    tier: 1,
    id: '',
    pointOfInterestId: '',
    pointOfInterestChallengeSubmissions: [],
    createdAt: new Date(),
    updatedAt: new Date()
  });
  const { uploadMedia, getPresignedUploadURL } = useMediaContext();
  const { apiClient } = useAPI();
  const { tagGroups } = useTagContext();
  const [newPoint, setNewPoint] = useState({
    name: '',
    description: '',
    lat: 0,
    lng: 0,
  });

  const tags = tagGroups.flatMap(group => group.tags);

  if (loading) {
    return <div className="p-4">Loading...</div>;
  }

  if (error) {
    return <div className="p-4 text-red-500">Error loading arena details</div>;
  }

  if (!arena) {
    return <div className="p-4">Arena not found</div>;
  }

  const handleArenaEdit = () => {
    setEditingArena(true);
    setArenaName(arena.name);
    setArenaDescription(arena.description);
    setArenaType(arena.type);
  };

  const handleArenaSave = async () => {
    try {
      if (!arenaName || !arenaDescription || !arenaType) {
        throw new Error('Name and description are required');
      }
      await updateArena(arenaName, arenaDescription, arenaType);
      setEditingArena(false);
    } catch (error) {
      console.error('Error saving arena:', error);
    }
  };

  const handleArenaImageUpdate = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file || !arena) return;

    try {
      await updateArenaImage(arena.id, file);
      setEditingArenaImage(false);
    } catch (error) {
      console.error('Error updating arena image:', error);
    }
  };

  const handlePointEdit = (point: any) => {
    setEditingPointId(point.id);
    setEditedPoint(point);
  };

  const handlePointSave = async (pointId: string) => {
    try {
      await updatePointOfInterest(pointId, editedPoint);
    } catch (error) {
      console.error('Error saving point:', error);
    }
  };

  const handlePointDelete = async (pointId: string) => {
    try {
      await deletePointOfInterest(pointId);
    } catch (error) {
      console.error('Error deleting point:', error);
    }
  };

  const handleNewPointSave = async (
    name: string,
    description: string,
    lat: number,
    lng: number,
    image: File | null,
    clue: string
  ) => {
    try {
      await createPointOfInterest(name, description, lat, lng, image, clue);
    } catch (error) {
      console.error('Error creating point:', error);
    } finally {
      setShowNewPointModal(false);
    }
  };

  const handleNewChildSave = async () => {
    try {
      if (newChild.pointOfInterestId && newChild.pointOfInterestGroupMemberId && newChild.pointOfInterestChallengeId) {
        await createPointOfInterestChildren(
          newChild.pointOfInterestId,
          newChild.pointOfInterestGroupMemberId,
          newChild.pointOfInterestChallengeId
        );
        setShowNewChildModal(false);
        setSelectedPointId(null);
        setNewChild({
          pointOfInterestId: '',
          pointOfInterestChallengeId: '',
          pointOfInterestGroupMemberId: '',
        });
      }
    } catch (error) {
      console.error('Error creating child point:', error);
    }
  };

  const handleNewChallengeSave = async () => {
    try {
      if (!selectedPointId) return;
      await createPointOfInterestChallenge(selectedPointId, newChallenge);
      setShowNewChallengeModal(false);
      setSelectedPointId(null);
      setNewChallenge({
        question: '',
        inventoryItemId: 0,
        tier: 1,
        id: '',
        pointOfInterestId: '',
        pointOfInterestChallengeSubmissions: [],
        createdAt: new Date(),
        updatedAt: new Date()
      });
    } catch (error) {
      console.error('Error creating challenge:', error);
    }
  };

  const handleImageUpdate = async (pointId: string, event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    try {
      await updatePointOfInterestImage(pointId, file);
      setEditingPointImageId(null);
    } catch (error) {
      console.error('Error updating image:', error);
    }
  };

  return (
    <div className="p-4">
      <div className="mb-6">
        <div className="flex gap-4 items-start">
          <div className="relative">
            {arena.imageUrl && (
              <>
                <img
                  src={arena.imageUrl}
                  alt={arena.name}
                  className="w-64 h-64 rounded-lg object-contain bg-gray-100 cursor-pointer"
                  onClick={() => setSelectedImage(arena.imageUrl)}
                />
                <button
                  onClick={() => setEditingArenaImage(true)}
                  className="absolute bottom-2 right-2 bg-blue-500 text-white p-1 rounded text-xs"
                >
                  Edit Image
                </button>
              </>
            )}
            {editingArenaImage && (
              <div className="absolute inset-0 bg-black bg-opacity-50 rounded-lg flex items-center justify-center">
                <div className="bg-white p-4 rounded">
                  <input
                    type="file"
                    accept="image/*"
                    onChange={handleArenaImageUpdate}
                    className="text-sm"
                  />
                  <button
                    onClick={() => setEditingArenaImage(false)}
                    className="mt-2 bg-gray-500 text-white px-2 py-1 rounded text-sm w-full"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            )}
          </div>
          <div className="flex-grow">
            {editingArena ? (
              <div>
                <input
                  type="text"
                  value={arenaName}
                  onChange={(e) => setArenaName(e.target.value)}
                  className="text-3xl font-bold mb-2 border rounded px-2 py-1 w-full"
                />
                <textarea
                  value={arenaDescription}
                  onChange={(e) => setArenaDescription(e.target.value)}
                  className="text-gray-600 text-lg mb-4 border rounded px-2 py-1 w-full"
                />
                <select
                  value={arenaType}
                  onChange={(e) => setArenaType(JSON.parse(e.target.value) as PointOfInterestGroupType)}
                  className="text-gray-600 text-lg mb-4 border rounded px-2 py-1 w-full"
                >
                  <option value={PointOfInterestGroupType.Arena}>Arena</option>
                  <option value={PointOfInterestGroupType.Quest}>Quest</option>
                </select>
                <div className="flex gap-2">
                  <button
                    onClick={handleArenaSave}
                    className="bg-blue-500 text-white px-4 py-2 rounded"
                  >
                    Save
                  </button>
                  <button
                    onClick={() => setEditingArena(false)}
                    className="bg-gray-500 text-white px-4 py-2 rounded"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            ) : (
              <div>
                <h1 className="text-3xl font-bold mb-2">{arena.name}</h1>
                <p>{arena.id}</p>
                <p className="text-gray-600 text-lg mb-4 w-[1400px]">
                  {arena.description}
                </p>
                <p className="text-gray-600 text-lg mb-4">
                  Type: {arena.type ? PointOfInterestGroupType[arena.type] : 'Unknown'}
                </p>
                <button
                  onClick={handleArenaEdit}
                  className="bg-blue-500 text-white px-4 py-2 rounded"
                >
                  Edit Arena
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      <div className="mt-8">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-2xl font-bold">Points of Interest</h2>
          <button
            onClick={() => setShowNewPointModal(true)}
            className="bg-green-500 text-white px-4 py-2 rounded"
          >
            Add New Point
          </button>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {arena.pointsOfInterest?.map((point) => (
            <div key={point.id} className="border rounded-lg p-4 shadow-sm">
              {editingPointId === point.id ? (
                <div>
                  <input
                    type="text"
                    value={editedPoint?.name}
                    onChange={(e) =>
                      setEditedPoint({ ...editedPoint, name: e.target.value })
                    }
                    className="text-xl font-semibold mb-2 border rounded px-2 py-1 w-full"
                  />
                  <textarea
                    value={editedPoint?.description}
                    onChange={(e) =>
                      setEditedPoint({
                        ...editedPoint,
                        description: e.target.value,
                      })
                    }
                    className="text-gray-600 mb-2 border rounded px-2 py-1 w-full"
                  />
                  <textarea
                    value={editedPoint?.clue}
                    onChange={(e) =>
                      setEditedPoint({
                        ...editedPoint,
                        clue: e.target.value,
                      })
                    }
                    placeholder="Enter clue"
                    className="text-gray-600 mb-2 border rounded px-2 py-1 w-full"
                  />
                  <div className="text-sm text-gray-500">
                    <input
                      type="number"
                      value={editedPoint?.lat}
                      onChange={(e) =>
                        setEditedPoint({
                          ...editedPoint,
                          lat: e.target.value,
                        })
                      }
                      className="border rounded px-2 py-1 w-full mb-1"
                      step="0.000001"
                    />
                    <input
                      type="number"
                      value={editedPoint?.lng}
                      onChange={(e) =>
                        setEditedPoint({
                          ...editedPoint,
                          lng: e.target.value,
                        })
                      }
                      className="border rounded px-2 py-1 w-full"
                      step="0.000001"
                    />
                  </div>
                  <div className="flex gap-2 mt-2">
                    <button
                      onClick={() => handlePointSave(point.id)}
                      className="bg-blue-500 text-white px-3 py-1 rounded text-sm"
                    >
                      Save
                    </button>
                    <button
                      onClick={() => setEditingPointId(null)}
                      className="bg-gray-500 text-white px-3 py-1 rounded text-sm"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              ) : (
                <div className="flex gap-4">
                  <div className="flex-shrink-0 relative">
                    {point.imageURL && (
                      <>
                        <img 
                          src={point.imageURL} 
                          alt={point.name}
                          className="w-32 h-32 object-cover rounded-lg cursor-pointer"
                          onClick={() => setSelectedImage(point.imageURL)}
                        />
                        <button
                          onClick={() => setEditingPointImageId(point.id)}
                          className="absolute bottom-2 right-2 bg-blue-500 text-white p-1 rounded text-xs"
                        >
                          Edit Image
                        </button>
                      </>
                    )}
                    {editingPointImageId === point.id && (
                      <div className="absolute inset-0 bg-black bg-opacity-50 rounded-lg flex items-center justify-center">
                        <div className="bg-white p-4 rounded">
                          <input
                            type="file"
                            accept="image/*"
                            onChange={(e) => handleImageUpdate(point.id, e)}
                            className="text-sm"
                          />
                          <button
                            onClick={() => setEditingPointImageId(null)}
                            className="mt-2 bg-gray-500 text-white px-2 py-1 rounded text-sm w-full"
                          >
                            Cancel
                          </button>
                        </div>
                      </div>
                    )}
                  </div>
                  <div>
                    <h3 className="text-xl font-semibold mb-2">{point.name}</h3>
                    <p className="text-gray-600 mb-2">{point.description}</p>
                    <p className="text-gray-600 mb-2">Clue: {point.clue}</p>
                    <p className="text-gray-600 mb-2">Original Name: {point.originalName}</p>
                    <p className="text-gray-600 mb-2">ID: {point.id}</p>
                    <p className="text-gray-600 mb-2">
                      Place ID: <a href={`/place/${point.googleMapsPlaceId}`} className="text-blue-500 hover:underline">{point.googleMapsPlaceId}</a>
                    </p>
                    <p className="text-gray-600 mb-2">Geometry: {point.geometry}</p>
                    <p className="text-gray-600 mb-2">Tags: {point.tags?.map(tag => tag.name).join(', ')}</p>
                    <div className="text-sm text-gray-500">
                      <p>Latitude: {point.lat}</p>
                      <p>Longitude: {point.lng}</p>
                    </div>
                    <div className="mt-2">
                      <div className="flex gap-2 items-center">
                        <select
                          className="border rounded px-2 py-1 text-sm"
                          onChange={(e) => setSelectedTagId(e.target.value)}
                        >
                          <option value="">Select a tag...</option>
                          {tags.map((tag) => (
                            <option key={tag.id} value={tag.id}>
                              {tag.name}
                            </option>
                          ))}
                        </select>
                        <button
                          onClick={() => addTagToPointOfInterest(selectedTagId, point.id)}
                          className="bg-green-500 text-white px-3 py-1 rounded text-sm"
                        >
                          Add Tag
                        </button>
                      </div>
                    </div>
                    <div className="flex gap-2 mt-2">
                      <button
                        onClick={() => handlePointEdit(point)}
                        className="bg-blue-500 text-white px-3 py-1 rounded text-sm"
                      >
                        Edit Point
                      </button>
                      <button
                        onClick={() => handlePointDelete(point.id)}
                        className="bg-red-500 text-white px-3 py-1 rounded text-sm"
                      >
                        Delete
                      </button>
                      <button
                        onClick={() => {
                          setSelectedPointId(point.id);
                          setShowNewChallengeModal(true);
                        }}
                        className="bg-purple-500 text-white px-3 py-1 rounded text-sm"
                      >
                        Add Challenge
                      </button>
                      <button
                        onClick={() => {
                          setNewChild({
                            pointOfInterestId: '',
                            pointOfInterestChallengeId: '',
                            pointOfInterestGroupMemberId: arena.groupMembers.find(member => member.pointOfInterestId === point.id)?.id || '',
                          });
                          setNewChildChallengeSelection(point.pointOfInterestChallenges);
                          setShowNewChildModal(true);
                        }}
                        className="bg-green-500 text-white px-3 py-1 rounded text-sm"
                      >
                        Add Child
                      </button>
                    </div>
                  </div>
                </div>
              )}
              {point.pointOfInterestChallenges &&
                point.pointOfInterestChallenges.length > 0 && (
                  <div className="mt-3">
                    <h4 className="font-semibold mb-1">Challenges:</h4>
                    <ul className="space-y-2">
                      {point.pointOfInterestChallenges.map((challenge, idx) => (
                        <li key={idx} className="flex items-center justify-between p-2 border rounded">
                          <div className="text-sm text-gray-600">
                            <div>{challenge.question}</div>
                            <div className="text-xs text-gray-500">
                              <span className="mr-4">Tier: {challenge.tier}</span>
                              <span>Item: {inventoryItems?.find(item => item.id === challenge.inventoryItemId)?.name || 'None'}</span>
                            </div>
                          </div>
                          <div className="flex gap-2">
                            <button
                              onClick={() => {
                                setNewChallenge(challenge);
                                setShowEditChallengeModal(true);
                              }}
                              className="bg-blue-500 text-white px-2 py-1 rounded text-xs"
                            >
                              Edit
                            </button>
                            <button
                              onClick={() => deletePointOfInterestChallenge(challenge.id)}
                              className="bg-red-500 text-white px-2 py-1 rounded text-xs"
                            >
                              Delete
                            </button>
                          </div>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              {arena.groupMembers.find(member => member.pointOfInterestId === point.id)?.children && (
                <div className="mt-3">
                  <h4 className="font-semibold mb-1">Children:</h4>
                  <ul className="space-y-2">
                    {arena.groupMembers.find(member => member.pointOfInterestId === point.id)?.children.map((child, idx) => (
                      <li key={idx} className="p-2 border rounded">
                        <div className="text-sm flex justify-between items-center">
                          <div>
                            <span className="font-medium">{arena.pointsOfInterest.find(p => p.id === child.pointOfInterestId)?.name}</span>
                            <span className="text-gray-600 mx-2">|</span>
                            <span className="text-gray-600">Tier {arena.pointsOfInterest.flatMap(p => p.pointOfInterestChallenges).find(c => c.id === child.pointOfInterestChallengeId)?.tier}</span>
                            <span className="text-gray-600 mx-2">|</span>
                            <span className="text-gray-600">{arena.pointsOfInterest.flatMap(p => p.pointOfInterestChallenges).find(c => c.id === child.pointOfInterestChallengeId)?.question}</span>
                          </div>
                          <button
                            onClick={() => deletePointOfInterestChildren(child.id)}
                            className="bg-red-500 text-white px-2 py-1 rounded text-xs"
                          >
                            Delete
                          </button>
                        </div>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      {showNewPointModal && (
        <AddNewPointOfInterest
          onSave={handleNewPointSave}
          onCancel={() => setShowNewPointModal(false)}
        />
      )}

      {showNewChildModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white p-6 rounded-lg w-96">
            <h3 className="text-xl font-bold mb-4">Add New Child Point</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">Select Child Point</label>
                <select
                  value={newChild.pointOfInterestId}
                  onChange={(e) => setNewChild({
                    ...newChild,
                    pointOfInterestId: e.target.value,
                  })}
                  className="mt-1 block w-full border rounded-md px-3 py-2"
                >
                  <option value="">Select a point...</option>
                  {arena?.pointsOfInterest?.map((point) => (
                    <option key={point.id} value={point.id}>
                      {point.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">Select Challenge</label>
                <select
                  value={newChild.pointOfInterestChallengeId}
                  onChange={(e) => setNewChild({
                    ...newChild,
                    pointOfInterestChallengeId: e.target.value
                  })}
                  className="mt-1 block w-full border rounded-md px-3 py-2"
                >
                  <option value="">Select a challenge...</option>
                  {newChildChallengeSelection.map((challenge) => (
                    <option key={challenge.id} value={challenge.id}>
                      Tier {challenge.tier} - {challenge.question}
                    </option>
                  ))}
                </select>
              </div>

              <div className="flex gap-2 mt-4">
                <button
                  onClick={handleNewChildSave}
                  className="bg-blue-500 text-white px-4 py-2 rounded"
                >
                  Save Child Point
                </button>
                <button
                  onClick={() => {
                    setShowNewChildModal(false);
                    setSelectedPointId(null);
                    setNewChildChallengeSelection([]);
                    setNewChild({
                      pointOfInterestId: '',
                      pointOfInterestChallengeId: '',
                      pointOfInterestGroupMemberId: '',
                    });
                  }}
                  className="bg-gray-500 text-white px-4 py-2 rounded"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {showEditChallengeModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white p-6 rounded-lg w-96">
            <h3 className="text-xl font-bold mb-4">Edit Challenge</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">Question</label>
                <textarea
                  value={newChallenge.question}
                  onChange={(e) => setNewChallenge({...newChallenge, question: e.target.value})}
                  className="mt-1 block w-full border rounded-md px-3 py-2"
                  rows={5}
                  placeholder="Enter your challenge question here. You can write multiple paragraphs of text to provide detailed instructions or context for the challenge."
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">Tier</label>
                <input
                  type="number"
                  value={newChallenge.tier}
                  onChange={(e) => setNewChallenge({...newChallenge, tier: parseInt(e.target.value)})}
                  className="mt-1 block w-full border rounded-md px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">Select Inventory Item</label>
                <div className="mt-1 grid grid-cols-2 gap-2 max-h-40 overflow-y-auto">
                  {inventoryItems.map((item) => (
                    <button
                      key={item.id}
                      onClick={() => setNewChallenge({
                        ...newChallenge,
                        inventoryItemId: newChallenge.inventoryItemId === item.id ? 0 : item.id
                      })}
                      className={`p-2 border rounded-md text-sm text-left hover:bg-gray-100 ${
                        newChallenge.inventoryItemId === item.id ? 'border-blue-500 bg-blue-50' : ''
                      }`}
                    >
                      {item.name}
                    </button>
                  ))}
                </div>
              </div>
              <div className="flex gap-2 mt-4">
                <button
                  onClick={async () => {
                    try {
                      await updatePointOfInterestChallenge(newChallenge.id, {
                        question: newChallenge.question,
                        tier: newChallenge.tier,
                        inventoryItemId: newChallenge.inventoryItemId
                      });
                      setShowEditChallengeModal(false);
                      setNewChallenge({
                        question: '',
                        inventoryItemId: 0,
                        tier: 1,
                        id: '',
                        pointOfInterestId: '',
                        pointOfInterestChallengeSubmissions: [],
                        createdAt: new Date(),
                        updatedAt: new Date()
                      });
                    } catch (error) {
                      console.error('Error updating challenge:', error);
                    }
                  }}
                  className="bg-blue-500 text-white px-4 py-2 rounded"
                >
                  Save Changes
                </button>
                <button
                  onClick={() => {
                    setShowEditChallengeModal(false);
                    setNewChallenge({
                      question: '',
                      inventoryItemId: 0,
                      tier: 1,
                      id: '',
                      pointOfInterestId: '',
                      pointOfInterestChallengeSubmissions: [],
                      createdAt: new Date(),
                      updatedAt: new Date()
                    });
                  }}
                  className="bg-gray-300 text-gray-700 px-4 py-2 rounded"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {showNewChallengeModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white p-6 rounded-lg w-96">
            <h3 className="text-xl font-bold mb-4">Add New Challenge</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">Question</label>
                <textarea
                  value={newChallenge.question}
                  onChange={(e) => setNewChallenge({...newChallenge, question: e.target.value})}
                  className="mt-1 block w-full border rounded-md px-3 py-2"
                  rows={5}
                  placeholder="Enter your challenge question here. You can write multiple paragraphs of text to provide detailed instructions or context for the challenge."
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">Tier</label>
                <input
                  type="number"
                  value={newChallenge.tier}
                  onChange={(e) => setNewChallenge({...newChallenge, tier: parseInt(e.target.value)})}
                  className="mt-1 block w-full border rounded-md px-3 py-2"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">Select Inventory Item</label>
                <div className="mt-1 grid grid-cols-2 gap-2 max-h-40 overflow-y-auto">
                  {inventoryItems.map((item) => (
                    <button
                      key={item.id}
                      onClick={() => setNewChallenge({
                        ...newChallenge, 
                        inventoryItemId: newChallenge.inventoryItemId === item.id ? 0 : item.id
                      })}
                      className={`p-2 border rounded-md text-sm text-left hover:bg-gray-100 ${
                        newChallenge.inventoryItemId === item.id ? 'border-blue-500 bg-blue-50' : ''
                      }`}
                    >
                      {item.name}
                    </button>
                  ))}
                </div>
              </div>
              <div className="flex gap-2 mt-4">
                <button
                  onClick={handleNewChallengeSave}
                  className="bg-blue-500 text-white px-4 py-2 rounded"
                >
                  Save Challenge
                </button>
                <button
                  onClick={() => {
                    setShowNewChallengeModal(false);
                    setSelectedPointId(null);
                  }}
                  className="bg-gray-500 text-white px-4 py-2 rounded"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {selectedImage && (
        <div 
          className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center p-4 z-50"
          onClick={() => setSelectedImage(null)}
        >
          <div className="relative max-w-4xl max-h-[90vh]">
            <img 
              src={selectedImage} 
              alt="Full size view"
              className="max-w-full max-h-[90vh] object-contain"
            />
            <button
              onClick={() => setSelectedImage(null)}
              className="absolute top-4 right-4 bg-white rounded-full p-2"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
