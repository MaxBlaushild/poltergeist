import React, { useState, useEffect } from "react";
import { LocationArchetype } from "@poltergeist/types";
import { useAPI } from "@poltergeist/contexts";
import { useQuestArchetypes } from "../contexts/questArchetypes";


const LocationArchetypes: React.FC<LocationArchetypesProps> = () => {
  const [selectedLocationArchetype, setSelectedLocationArchetype] = useState<LocationArchetype | null>(null);
  const { locationArchetypes, createLocationArchetype, updateLocationArchetype, deleteLocationArchetype } = useQuestArchetypes();
  const [newLocationArchetype, setNewLocationArchetype] = useState<LocationArchetype>({
    id: "",
    name: "",
    description: "",
    placeTypes: [],
    createdAt: new Date(),
    updatedAt: new Date(),
  });

  const { createLocationArchetype, updateLocationArchetype, deleteLocationArchetype } = useQuestArchetypes();

  const handleLocationArchetypeChange = (locationArchetype: LocationArchetype) => {
    setSelectedLocationArchetype(locationArchetype);
    onLocationArchetypeChange(locationArchetype);
  };

  const handleLocationArchetypeDelete = (locationArchetypeId: string) => {
    onLocationArchetypeDelete(locationArchetypeId);
  };

  const handleNewLocationArchetypeChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setNewLocationArchetype({ ...newLocationArchetype, [event.target.name]: event.target.value });
  };

  const handleNewLocationArchetypeSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    await createLocationArchetype(newLocationArchetype);
    setNewLocationArchetype({ id: "", name: "", description: "", placeTypes: [], createdAt: new Date(), updatedAt: new Date() });
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
                  Description
                </label>
                <input
                  type="text"
                  name="description"
                  value={newLocationArchetype.description}
                  onChange={handleNewLocationArchetypeChange}
                  required
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-blue-500"
                />
              </div>

              <div className="mb-4">
                <label className="block text-gray-700 text-sm font-bold mb-2">
                  Place Types
                </label>
                <div className="flex gap-2 mb-2">
                  <input
                    type="text"
                    value={newPlaceType}
                    onChange={(e) => setNewPlaceType(e.target.value)}
                    placeholder="Add place type"
                    className="flex-1 px-3 py-2 border border-gray-300 rounded focus:outline-none focus:border-blue-500"
                  />
                  <button
                    type="button"
                    className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
                    onClick={() => {
                      if (newPlaceType.trim()) {
                        setNewLocationArchetype({
                          ...newLocationArchetype,
                          placeTypes: [...newLocationArchetype.placeTypes, newPlaceType.trim()]
                        });
                        setNewPlaceType('');
                      }
                    }}
                  >
                    Add
                  </button>
                </div>
                <div className="space-y-2">
                  {newLocationArchetype.placeTypes.map((placeType, index) => (
                    <div key={index} className="flex items-center justify-between bg-gray-50 px-3 py-2 rounded">
                      <span>{placeType}</span>
                      <button
                        type="button"
                        className="text-red-500 hover:text-red-700"
                        onClick={() => {
                          const updatedPlaceTypes = [...newLocationArchetype.placeTypes];
                          updatedPlaceTypes.splice(index, 1);
                          setNewLocationArchetype({
                            ...newLocationArchetype,
                            placeTypes: updatedPlaceTypes
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
