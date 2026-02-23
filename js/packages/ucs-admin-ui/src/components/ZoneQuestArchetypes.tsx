import React, { useEffect, useState } from 'react';
import { useQuestArchetypes } from '../contexts/questArchetypes.tsx';
import { useAPI, useZoneContext } from '@poltergeist/contexts';
import { Character } from '@poltergeist/types';
import "./questArchetypeTheme.css";

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

  return (
    <div className="qa-theme">
      <div className="qa-shell">
        <header className="qa-hero">
          <div>
            <div className="qa-kicker">Zone Operations</div>
            <h1 className="qa-title">Zone Quest Archetypes</h1>
            <p className="qa-subtitle">
              Bind archetypes to specific zones, set quest volume targets, and assign quest givers so each area
              feels distinct.
            </p>
          </div>
          <div className="qa-hero-actions">
            <button className="qa-btn qa-btn-primary" onClick={() => setShouldShowModal(true)}>
              New Zone Assignment
            </button>
          </div>
        </header>

        <section className="qa-grid">
          {zoneQuestArchetypes?.length === 0 ? (
            <div className="qa-panel">
              <div className="qa-card-title">No zone assignments yet</div>
              <p className="qa-muted" style={{ marginTop: 8 }}>
                Assign an archetype to a zone to start generating quests.
              </p>
            </div>
          ) : (
            zoneQuestArchetypes.map((zoneQuestArchetype, index) => (
              <article
                key={zoneQuestArchetype.id}
                className="qa-card"
                style={{ animationDelay: `${index * 0.06}s` }}
              >
                <div className="qa-card-header">
                  <div>
                    <h2 className="qa-card-title">{zoneQuestArchetype.questArchetype.name}</h2>
                    <div className="qa-meta">Zone: {zoneQuestArchetype.zone.name}</div>
                  </div>
                  <div className="qa-actions">
                    <button
                      onClick={() => deleteZoneQuestArchetype(zoneQuestArchetype.id)}
                      className="qa-btn qa-btn-danger"
                    >
                      Delete
                    </button>
                  </div>
                </div>

                <div className="qa-stat-grid">
                  <div className="qa-stat">
                    <div className="qa-stat-label">Quests to Generate</div>
                    <div className="qa-stat-value">{zoneQuestArchetype.numberOfQuests}</div>
                  </div>
                  <div className="qa-stat">
                    <div className="qa-stat-label">Quest Giver</div>
                    <div className="qa-stat-value">{zoneQuestArchetype.character?.name ?? 'None'}</div>
                  </div>
                  <div className="qa-stat">
                    <div className="qa-stat-label">Archetype ID</div>
                    <div className="qa-stat-value">{zoneQuestArchetype.questArchetypeId.slice(0, 8)}…</div>
                  </div>
                </div>
              </article>
            ))
          )}
        </section>
      </div>

      {shouldShowModal && (
        <div className="qa-modal">
          <div className="qa-modal-card">
            <h2 className="qa-modal-title">Create Zone Quest Archetype</h2>

            <div className="qa-form-grid">
              <div className="qa-field">
                <div className="qa-label">Zone Search</div>
                <input
                  type="text"
                  className="qa-input"
                  value={zoneSearch}
                  onChange={(e) => setZoneSearch(e.target.value)}
                  placeholder="Search zones..."
                />
              </div>

              <div className="qa-field">
                <div className="qa-label">Zone</div>
                <select
                  className="qa-select"
                  value={selectedZoneId}
                  onChange={(e) => setSelectedZoneId(e.target.value)}
                >
                  <option value="">Select a zone</option>
                  {zones
                    .filter((z) => z.name.toLowerCase().includes(zoneSearch.toLowerCase()))
                    .map((z) => (
                      <option key={z.id} value={z.id}>
                        {z.name}
                      </option>
                    ))}
                </select>
              </div>

              <div className="qa-field">
                <div className="qa-label">Quest Archetype Search</div>
                <input
                  type="text"
                  className="qa-input"
                  value={questArchetypeSearch}
                  onChange={(e) => setQuestArchetypeSearch(e.target.value)}
                  placeholder="Search quest archetypes..."
                />
              </div>

              <div className="qa-field">
                <div className="qa-label">Quest Archetype</div>
                <select
                  className="qa-select"
                  value={selectedQuestArchetypeId}
                  onChange={(e) => setSelectedQuestArchetypeId(e.target.value)}
                >
                  <option value="">Select a quest archetype</option>
                  {questArchetypes
                    .filter((qa) => qa.name.toLowerCase().includes(questArchetypeSearch.toLowerCase()))
                    .map((qa) => (
                      <option key={qa.id} value={qa.id}>
                        {qa.name}
                      </option>
                    ))}
                </select>
              </div>

              <div className="qa-field">
                <div className="qa-label">Number of Quests</div>
                <input
                  type="number"
                  className="qa-input"
                  value={numberOfQuests}
                  onChange={(e) => setNumberOfQuests(parseInt(e.target.value) || 1)}
                  min="1"
                />
              </div>

              <div className="qa-field">
                <div className="qa-label">Character Search</div>
                <input
                  type="text"
                  className="qa-input"
                  value={characterSearch}
                  onChange={(e) => setCharacterSearch(e.target.value)}
                  placeholder="Search characters..."
                />
              </div>

              <div className="qa-field">
                <div className="qa-label">Quest Giver Character</div>
                <select
                  className="qa-select"
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

              <div className="qa-footer">
                <button className="qa-btn qa-btn-outline" onClick={() => setShouldShowModal(false)}>
                  Cancel
                </button>
                <button
                  className="qa-btn qa-btn-primary"
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
                >
                  Create
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
