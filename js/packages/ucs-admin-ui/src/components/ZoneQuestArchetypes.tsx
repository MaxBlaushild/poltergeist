import React, { useEffect, useState } from 'react';
import { useQuestArchetypes } from '../contexts/questArchetypes.tsx';
import { useAPI, useZoneContext } from '@poltergeist/contexts';
import { Character } from '@poltergeist/types';

export const ZoneQuestArchetypes = () => {
  const { zones } = useZoneContext();
  const { apiClient } = useAPI();
  const { zoneQuestArchetypes, createZoneQuestArchetype, deleteZoneQuestArchetype, questArchetypes } = useQuestArchetypes();
  const [shouldShowModal, setShouldShowModal] = useState(false);
  const [zoneSearch, setZoneSearch] = useState('');
  const [questArchetypeSearch, setQuestArchetypeSearch] = useState('');
  const [characterSearch, setCharacterSearch] = useState('');
  const [selectedZoneId, setSelectedZoneId] = useState('');
  const [selectedQuestArchetypeId, setSelectedQuestArchetypeId] = useState('');
  const [numberOfQuests, setNumberOfQuests] = useState(1);
  const [characters, setCharacters] = useState<Character[]>([]);
  const [selectedCharacterId, setSelectedCharacterId] = useState('');

  useEffect(() => {
    const fetchCharacters = async () => {
      try {
        const response = await apiClient.get<Character[]>('/sonar/characters');
        setCharacters(response);
      } catch (error) {
        console.error('Error fetching characters:', error);
      }
    };
    fetchCharacters();
  }, [apiClient]);

  return <div className="m-10">
    <h1 className="text-2xl font-bold">Zone Quest Archetypes</h1>
    <div className="flex flex-col gap-4">
      {zoneQuestArchetypes?.map((zoneQuestArchetype) => (
        <div key={zoneQuestArchetype.id} className="flex items-center justify-between p-4 bg-white rounded-lg shadow">
          <div className="flex flex-col">
            <h2 className="text-xl font-semibold">{zoneQuestArchetype.questArchetype.name}</h2>
            <div className="text-gray-600">
              <p>Zone: {zoneQuestArchetype.zone.name}</p>
              <p>Number of Quests: {zoneQuestArchetype.numberOfQuests}</p>
              <p>Character: {zoneQuestArchetype.character?.name ?? 'None'}</p>
            </div>
          </div>
          <button
            onClick={() => deleteZoneQuestArchetype(zoneQuestArchetype.id)}
            className="px-4 py-2 text-white bg-red-500 rounded hover:bg-red-600 transition-colors"
          >
            Delete
          </button>
        </div>
      ))}
      <button
        onClick={() => setShouldShowModal(true)}
        className="px-4 py-2 text-white bg-blue-500 rounded hover:bg-blue-600 transition-colors"
      >
        Create Zone Quest Archetype
      </button>

      {shouldShowModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white p-6 rounded-lg w-[500px]">
            <h2 className="text-xl font-bold mb-4">Create Zone Quest Archetype</h2>
            
            <div className="flex flex-col gap-4">
              <div>
                <label className="block mb-2">Zone Search</label>
                <input
                  type="text"
                  className="w-full p-2 border rounded"
                  value={zoneSearch}
                  onChange={(e) => setZoneSearch(e.target.value)}
                  placeholder="Search zones..."
                />
              </div>

              <div>
                <label className="block mb-2">Zone</label>
                <select 
                  className="w-full p-2 border rounded"
                  value={selectedZoneId}
                  onChange={(e) => setSelectedZoneId(e.target.value)}
                >
                  <option value="">Select a zone</option>
                  {zones
                    .filter((z) => 
                      z.name.toLowerCase().includes(zoneSearch.toLowerCase())
                    )
                    .map((z) => (
                      <option key={z.id} value={z.id}>
                        {z.name}
                      </option>
                    ))}
                </select>
              </div>

              <div>
                <label className="block mb-2">Quest Archetype Search</label>
                <input
                  type="text"
                  className="w-full p-2 border rounded"
                  value={questArchetypeSearch}
                  onChange={(e) => setQuestArchetypeSearch(e.target.value)}
                  placeholder="Search quest archetypes..."
                />
              </div>

              <div>
                <label className="block mb-2">Quest Archetype</label>
                <select
                  className="w-full p-2 border rounded"
                  value={selectedQuestArchetypeId}
                  onChange={(e) => setSelectedQuestArchetypeId(e.target.value)}
                >
                  <option value="">Select a quest archetype</option>
                  {questArchetypes
                    .filter((qa) =>
                      qa.name.toLowerCase().includes(questArchetypeSearch.toLowerCase())
                    )
                    .map((qa) => (
                      <option key={qa.id} value={qa.id}>
                        {qa.name}
                      </option>
                    ))}
                </select>
              </div>

              <div>
                <label className="block mb-2">Number of Quests</label>
                <input
                  type="number"
                  className="w-full p-2 border rounded"
                  value={numberOfQuests}
                  onChange={(e) => setNumberOfQuests(parseInt(e.target.value))}
                  min="1"
                />
              </div>

              <div>
                <label className="block mb-2">Character Search</label>
                <input
                  type="text"
                  className="w-full p-2 border rounded"
                  value={characterSearch}
                  onChange={(e) => setCharacterSearch(e.target.value)}
                  placeholder="Search characters..."
                />
              </div>

              <div>
                <label className="block mb-2">Quest Giver Character</label>
                <select
                  className="w-full p-2 border rounded"
                  value={selectedCharacterId}
                  onChange={(e) => setSelectedCharacterId(e.target.value)}
                >
                  <option value="">No character</option>
                  {characters
                    .filter((character) =>
                      character.name.toLowerCase().includes(characterSearch.toLowerCase())
                    )
                    .map((character) => (
                      <option key={character.id} value={character.id}>
                        {character.name}
                      </option>
                    ))}
                </select>
              </div>

              <div className="flex justify-end gap-2 mt-4">
                <button
                  onClick={() => setShouldShowModal(false)}
                  className="px-4 py-2 text-gray-600 bg-gray-200 rounded hover:bg-gray-300 transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={async () => {
                    if (selectedZoneId && selectedQuestArchetypeId && numberOfQuests) {
                      await createZoneQuestArchetype(
                        selectedZoneId,
                        selectedQuestArchetypeId,
                        numberOfQuests,
                        selectedCharacterId || null
                      );
                      setShouldShowModal(false);
                    }
                  }}
                  disabled={!selectedZoneId || !selectedQuestArchetypeId || !numberOfQuests}
                  className="px-4 py-2 text-white bg-blue-500 rounded hover:bg-blue-600 transition-colors disabled:bg-gray-400"
                >
                  Create
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  </div>;
};
