import React, { useState, useEffect } from "react";
import { LocationArchetype } from "@poltergeist/types";
import { useAPI } from "@poltergeist/contexts";
import { useQuestArchetypes } from "../contexts/questArchetypes.tsx";

const LocationArchetypes: React.FC = () => {
  const [selectedLocationArchetype, setSelectedLocationArchetype] = useState<LocationArchetype | null>(null);
  const { locationArchetypes, createLocationArchetype, updateLocationArchetype, deleteLocationArchetype, placeTypes } = useQuestArchetypes();
  const [newLocationArchetype, setNewLocationArchetype] = useState<LocationArchetype>({
    id: "",
    name: "",
    includedTypes: [],
    excludedTypes: [],
    challenges: [],
    createdAt: new Date(),
    updatedAt: new Date(),
  });
  const [showModal, setShowModal] = useState(false);
  const [newIncludedType, setNewIncludedType] = useState("");
  const [newExcludedType, setNewExcludedType] = useState("");
  const [newChallenge, setNewChallenge] = useState("");

  const handleNewLocationArchetypeChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setNewLocationArchetype({ ...newLocationArchetype, [event.target.name]: event.target.value });
  };

  const handleNewLocationArchetypeSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    await createLocationArchetype(newLocationArchetype);
    setNewLocationArchetype({ id: "", name: "", includedTypes: [], excludedTypes: [], challenges: [], createdAt: new Date(), updatedAt: new Date() });
  };

  const handleLocationArchetypeUpdate = async (locationArchetype: LocationArchetype) => {
    await updateLocationArchetype(locationArchetype);
  };

  const handleLocationArchetypeDelete = async (locationArchetypeId: string) => {
    await deleteLocationArchetype(locationArchetypeId);
  };

  return (
    <div>
      <h2>Location Archetypes</h2>
      <div className="grid grid-cols-1 gap-4 mb-4">
        {locationArchetypes.map((archetype) => (
          <div key={archetype.id} className="border rounded-lg p-4 bg-white shadow">
            <div className="flex justify-between items-center mb-3">
              <h3 className="text-lg font-semibold">{archetype.name}</h3>
              <div className="flex gap-2">
                <button
                  onClick={() => setSelectedLocationArchetype(archetype)}
                  className="p-2 text-blue-500 hover:text-blue-700"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                  </svg>
                </button>
                <button
                  onClick={() => handleLocationArchetypeDelete(archetype.id)}
                  className="p-2 text-red-500 hover:text-red-700"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clipRule="evenodd" />
                  </svg>
                </button>
              </div>
            </div>
            
            <div className="space-y-3">
              <div>
                <h4 className="font-medium text-gray-700 mb-1">Included Types:</h4>
                <div className="flex flex-wrap gap-2">
                  {archetype.includedTypes.map((type, index) => (
                    <span key={index} className="px-2 py-1 bg-blue-100 text-blue-800 rounded-full text-sm">
                      {type}
                    </span>
                  ))}
                </div>
              </div>
              
              <div>
                <h4 className="font-medium text-gray-700 mb-1">Excluded Types:</h4>
                <div className="flex flex-wrap gap-2">
                  {archetype.excludedTypes.map((type, index) => (
                    <span key={index} className="px-2 py-1 bg-red-100 text-red-800 rounded-full text-sm">
                      {type}
                    </span>
                  ))}
                </div>
              </div>
              
              <div>
                <h4 className="font-medium text-gray-700 mb-1">Challenges:</h4>
                <div className="flex flex-wrap gap-2">
                  {archetype.challenges.map((challenge, index) => (
                    <span key={index} className="px-2 py-1 bg-green-100 text-green-800 rounded-full text-sm">
                      {challenge}
                    </span>
                  ))}
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
      <button 
        className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
        onClick={() => setShowModal(true)}
      >
        Create New Location Archetype
      </button>

      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white rounded-lg p-6 w-full max-w-lg">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-xl font-semibold">Create New Location Archetype</h3>
              <button 
                onClick={() => setShowModal(false)}
                className="text-gray-500 hover:text-gray-700"
              >
                <span className="text-2xl">&times;</span>
              </button>
            </div>

            <form onSubmit={handleNewLocationArchetypeSubmit}>
              <div className="mb-4">
                <label className="block text-gray-700 text-sm font-bold mb-2">
                  Name
                </label>
                <input
                  type="text"
                  name="name"
                  value={newLocationArchetype.name}
                  onChange={handleNewLocationArchetypeChange}
                  required
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-blue-500"
                />
              </div>

              <div className="mb-4">
                <label className="block text-gray-700 text-sm font-bold mb-2">
                  Included Types
                </label>
                <div className="flex gap-2 mb-2">
                  <div className="relative flex-1">
                    <input
                      type="text"
                      value={newIncludedType}
                      onChange={(e) => setNewIncludedType(e.target.value)}
                      placeholder="Search place types..."
                      className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-blue-500"
                    />
                    {newIncludedType && (
                      <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-auto">
                        {placeTypes
                          .filter(type => 
                            type.toLowerCase().includes(newIncludedType.toLowerCase()) &&
                            !newLocationArchetype.includedTypes.includes(type)
                          )
                          .map((type, index) => (
                            <div
                              key={index}
                              className="px-3 py-2 hover:bg-gray-100 cursor-pointer"
                              onClick={() => {
                                setNewLocationArchetype({
                                  ...newLocationArchetype,
                                  includedTypes: [...newLocationArchetype.includedTypes, type]
                                });
                                setNewIncludedType('');
                              }}
                            >
                              {type}
                            </div>
                          ))
                        }
                      </div>
                    )}
                  </div>
                </div>
                <div className="space-y-2">
                  {newLocationArchetype.includedTypes.map((type, index) => (
                    <div key={index} className="flex items-center justify-between bg-gray-50 px-3 py-2 rounded">
                      <span>{type}</span>
                      <button
                        type="button"
                        className="text-red-500 hover:text-red-700"
                        onClick={() => {
                          const updatedTypes = [...newLocationArchetype.includedTypes];
                          updatedTypes.splice(index, 1);
                          setNewLocationArchetype({
                            ...newLocationArchetype,
                            includedTypes: updatedTypes
                          });
                        }}
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                          <path fillRule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clipRule="evenodd" />
                        </svg>
                      </button>
                    </div>
                  ))}
                </div>
              </div>

              <div className="mb-4">
                <label className="block text-gray-700 text-sm font-bold mb-2">
                  Excluded Types
                </label>
                <div className="flex gap-2 mb-2">
                  <div className="relative flex-1">
                    <input
                      type="text"
                      value={newExcludedType}
                      onChange={(e) => setNewExcludedType(e.target.value)}
                      placeholder="Search place types..."
                      className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-blue-500"
                    />
                    {newExcludedType && (
                      <div className="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-auto">
                        {placeTypes
                          .filter(type => 
                            type.toLowerCase().includes(newExcludedType.toLowerCase()) &&
                            !newLocationArchetype.excludedTypes.includes(type)
                          )
                          .map((type, index) => (
                            <div
                              key={index}
                              className="px-3 py-2 hover:bg-gray-100 cursor-pointer"
                              onClick={() => {
                                setNewLocationArchetype({
                                  ...newLocationArchetype,
                                  excludedTypes: [...newLocationArchetype.excludedTypes, type]
                                });
                                setNewExcludedType('');
                              }}
                            >
                              {type}
                            </div>
                          ))
                        }
                      </div>
                    )}
                  </div>
                </div>
                <div className="space-y-2">
                  {newLocationArchetype.excludedTypes.map((type, index) => (
                    <div key={index} className="flex items-center justify-between bg-gray-50 px-3 py-2 rounded">
                      <span>{type}</span>
                      <button
                        type="button"
                        className="text-red-500 hover:text-red-700"
                        onClick={() => {
                          const updatedTypes = [...newLocationArchetype.excludedTypes];
                          updatedTypes.splice(index, 1);
                          setNewLocationArchetype({
                            ...newLocationArchetype,
                            excludedTypes: updatedTypes
                          });
                        }}
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                          <path fillRule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clipRule="evenodd" />
                        </svg>
                      </button>
                    </div>
                  ))}
                </div>
              </div>

              <div className="mb-4">
                <label className="block text-gray-700 text-sm font-bold mb-2">
                  Challenges
                </label>
                <div className="flex gap-2 mb-2">
                  <input
                    type="text"
                    value={newChallenge}
                    onChange={(e) => setNewChallenge(e.target.value)}
                    placeholder="Add challenge"
                    className="flex-1 px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-blue-500"
                  />
                  <button
                    type="button"
                    className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
                    onClick={() => {
                      if (newChallenge.trim()) {
                        setNewLocationArchetype({
                          ...newLocationArchetype,
                          challenges: [...newLocationArchetype.challenges, newChallenge.trim()]
                        });
                        setNewChallenge('');
                      }
                    }}
                  >
                    Add
                  </button>
                </div>
                <div className="space-y-2">
                  {newLocationArchetype.challenges.map((challenge, index) => (
                    <div key={index} className="flex items-center justify-between bg-gray-50 px-3 py-2 rounded">
                      <span>{challenge}</span>
                      <button
                        type="button"
                        className="text-red-500 hover:text-red-700"
                        onClick={() => {
                          const updatedChallenges = [...newLocationArchetype.challenges];
                          updatedChallenges.splice(index, 1);
                          setNewLocationArchetype({
                            ...newLocationArchetype,
                            challenges: updatedChallenges
                          });
                        }}
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                          <path fillRule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clipRule="evenodd" />
                        </svg>
                      </button>
                    </div>
                  ))}
                </div>
              </div>

              <div className="flex justify-end gap-2">
                <button
                  type="submit"
                  className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
                >
                  Create Location Archetype
                </button>
                <button
                  type="button"
                  className="px-4 py-2 bg-gray-300 text-gray-700 rounded hover:bg-gray-400 transition-colors"
                  onClick={() => setShowModal(false)}
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default LocationArchetypes;
