import { useAPI } from '@poltergeist/contexts';
import { Character, Zone, MovementPatternType, Location, CharacterAction, DialogueMessage, ShopInventoryItem } from '@poltergeist/types';
import React, { useState, useEffect } from 'react';
import { CharacterMapPicker } from './CharacterMapPicker.tsx';
import { DialogueActionEditor } from './DialogueActionEditor.tsx';
import { ShopActionEditor } from './ShopActionEditor.tsx';

export const Characters = () => {
  const { apiClient } = useAPI();
  const [characters, setCharacters] = useState<Character[]>([]);
  const [filteredCharacters, setFilteredCharacters] = useState<Character[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [showCreateCharacter, setShowCreateCharacter] = useState(false);
  const [editingCharacter, setEditingCharacter] = useState<Character | null>(null);
  const [availableZones, setAvailableZones] = useState<Zone[]>([]);
  
  // Dialogue management state
  const [selectedCharacterForDialogue, setSelectedCharacterForDialogue] = useState<Character | null>(null);
  const [characterActions, setCharacterActions] = useState<CharacterAction[]>([]);
  const [editingAction, setEditingAction] = useState<CharacterAction | null>(null);
  const [showDialogueEditor, setShowDialogueEditor] = useState(false);
  const [showDialogueManager, setShowDialogueManager] = useState(false);
  const [showShopEditor, setShowShopEditor] = useState(false);

  // Form state
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    mapIconUrl: '',
    dialogueImageUrl: '',
    movementPattern: {
      movementPatternType: 'static' as MovementPatternType,
      zoneId: '',
      startingLatitude: 0,
      startingLongitude: 0,
      path: [] as Location[]
    }
  });

  useEffect(() => {
    fetchCharacters();
    fetchZones();
  }, []);

  useEffect(() => {
    if (searchQuery === '') {
      setFilteredCharacters(characters);
    } else {
      const filtered = characters.filter(character =>
        character.name?.toLowerCase().includes(searchQuery.toLowerCase())
      );
      setFilteredCharacters(filtered);
    }
  }, [searchQuery, characters]);

  const fetchCharacters = async () => {
    try {
      const response = await apiClient.get<Character[]>('/sonar/characters');
      setCharacters(response);
      setFilteredCharacters(response);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching characters:', error);
      setLoading(false);
    }
  };

  const fetchZones = async () => {
    try {
      const response = await apiClient.get<Zone[]>('/sonar/zones');
      setAvailableZones(response);
    } catch (error) {
      console.error('Error fetching zones:', error);
    }
  };

  // Dialogue management functions
  const fetchCharacterActions = async (characterId: string) => {
    try {
      const response = await apiClient.get<CharacterAction[]>(`/sonar/characters/${characterId}/actions`);
      setCharacterActions(response);
    } catch (error) {
      console.error('Error fetching character actions:', error);
    }
  };

  const createCharacterAction = async (characterId: string, actionType: 'talk' | 'shop', dialogue?: DialogueMessage[], metadata?: any) => {
    try {
      const newAction = await apiClient.post<CharacterAction>('/sonar/character-actions', {
        characterId,
        actionType,
        dialogue: dialogue || [],
        metadata: metadata || {}
      });
      setCharacterActions([...characterActions, newAction]);
      return newAction;
    } catch (error) {
      console.error('Error creating character action:', error);
      throw error;
    }
  };

  const updateCharacterAction = async (actionId: string, dialogue?: DialogueMessage[], metadata?: any) => {
    try {
      const updates: any = {};
      if (dialogue !== undefined) {
        updates.dialogue = dialogue;
      }
      if (metadata !== undefined) {
        updates.metadata = metadata;
      }
      const updatedAction = await apiClient.put<CharacterAction>(`/sonar/character-actions/${actionId}`, updates);
      setCharacterActions(characterActions.map(a => a.id === actionId ? updatedAction : a));
      return updatedAction;
    } catch (error) {
      console.error('Error updating character action:', error);
      throw error;
    }
  };

  const deleteCharacterAction = async (actionId: string) => {
    try {
      await apiClient.delete(`/sonar/character-actions/${actionId}`);
      setCharacterActions(characterActions.filter(a => a.id !== actionId));
    } catch (error) {
      console.error('Error deleting character action:', error);
    }
  };

  const handleManageDialogue = async (character: Character) => {
    setSelectedCharacterForDialogue(character);
    setShowDialogueManager(true);
    await fetchCharacterActions(character.id);
  };

  const handleCreateNewAction = () => {
    setEditingAction(null);
    setShowDialogueEditor(true);
  };

  const handleEditAction = (action: CharacterAction) => {
    setEditingAction(action);
    if (action.actionType === 'shop') {
      setShowShopEditor(true);
    } else {
      setShowDialogueEditor(true);
    }
  };

  const handleSaveDialogue = async (dialogue: DialogueMessage[]) => {
    if (!selectedCharacterForDialogue) return;

    try {
      if (editingAction) {
        await updateCharacterAction(editingAction.id, dialogue);
      } else {
        await createCharacterAction(selectedCharacterForDialogue.id, 'talk', dialogue);
      }
      setShowDialogueEditor(false);
      setEditingAction(null);
      await fetchCharacterActions(selectedCharacterForDialogue.id);
    } catch (error) {
      console.error('Error saving dialogue:', error);
    }
  };

  const handleSaveShop = async (inventory: ShopInventoryItem[]) => {
    if (!selectedCharacterForDialogue) return;

    try {
      if (editingAction) {
        await updateCharacterAction(editingAction.id, undefined, { inventory });
      } else {
        await createCharacterAction(selectedCharacterForDialogue.id, 'shop', [], { inventory });
      }
      setShowShopEditor(false);
      setEditingAction(null);
      await fetchCharacterActions(selectedCharacterForDialogue.id);
    } catch (error) {
      console.error('Error saving shop:', error);
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      mapIconUrl: '',
      dialogueImageUrl: '',
      movementPattern: {
        movementPatternType: 'static',
        zoneId: '',
        startingLatitude: 0,
        startingLongitude: 0,
        path: []
      }
    });
  };

  const handleCreateCharacter = async () => {
    try {
      const newCharacter = await apiClient.post<Character>('/sonar/characters', formData);
      setCharacters([...characters, newCharacter]);
      setShowCreateCharacter(false);
      resetForm();
    } catch (error) {
      console.error('Error creating character:', error);
    }
  };

  const handleUpdateCharacter = async () => {
    if (!editingCharacter) return;
    
    try {
      const updatedCharacter = await apiClient.put<Character>(`/sonar/characters/${editingCharacter.id}`, formData);
      setCharacters(characters.map(c => c.id === editingCharacter.id ? updatedCharacter : c));
      setEditingCharacter(null);
      resetForm();
    } catch (error) {
      console.error('Error updating character:', error);
    }
  };

  const handleDeleteCharacter = async (character: Character) => {
    try {
      await apiClient.delete(`/sonar/characters/${character.id}`);
      setCharacters(characters.filter(c => c.id !== character.id));
    } catch (error) {
      console.error('Error deleting character:', error);
    }
  };

  const handleEditCharacter = (character: Character) => {
    setEditingCharacter(character);
    setFormData({
      name: character.name,
      description: character.description,
      mapIconUrl: character.mapIconUrl,
      dialogueImageUrl: character.dialogueImageUrl,
      movementPattern: {
        movementPatternType: character.movementPattern.movementPatternType,
        zoneId: character.movementPattern.zoneId || '',
        startingLatitude: character.movementPattern.startingLatitude,
        startingLongitude: character.movementPattern.startingLongitude,
        path: character.movementPattern.path || []
      }
    });
  };

  const addWaypoint = () => {
    setFormData({
      ...formData,
      movementPattern: {
        ...formData.movementPattern,
        path: [...formData.movementPattern.path, { latitude: 0, longitude: 0 }]
      }
    });
  };

  const updateWaypoint = (index: number, field: 'latitude' | 'longitude', value: number) => {
    const newPath = [...formData.movementPattern.path];
    newPath[index][field] = value;
    setFormData({
      ...formData,
      movementPattern: {
        ...formData.movementPattern,
        path: newPath
      }
    });
  };

  const removeWaypoint = (index: number) => {
    setFormData({
      ...formData,
      movementPattern: {
        ...formData.movementPattern,
        path: formData.movementPattern.path.filter((_, i) => i !== index)
      }
    });
  };

  if (loading) {
    return <div className="m-10">Loading characters...</div>;
  }

  return (
    <div className="m-10">
      <h1 className="text-2xl font-bold mb-4">Characters</h1>
      
      {/* Search */}
      <div className="mb-4">
        <input
          type="text"
          placeholder="Search characters..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full p-2 border rounded-md"
        />
      </div>

      {/* Characters Grid */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
        gap: '20px',
        padding: '20px'
      }}>
        {filteredCharacters.map((character) => (
          <div 
            key={character.id}
            style={{
              padding: '20px',
              border: '1px solid #ccc',
              borderRadius: '8px',
              backgroundColor: '#fff',
              boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
            }}
          >
            <h2 style={{ 
              margin: '0 0 15px 0',
              color: '#333'
            }}>{character.name}</h2>
            
            <p style={{ margin: '5px 0', color: '#666' }}>
              Description: {character.description || 'No description'}
            </p>
            
            <p style={{ margin: '5px 0', color: '#666' }}>
              Movement: {character.movementPattern.movementPatternType}
            </p>

            <p style={{ margin: '5px 0', color: '#666' }}>
              Dialogue Image URL: {character.dialogueImageUrl || '—'}
            </p>
            {character.dialogueImageUrl && (
              <img
                src={character.dialogueImageUrl}
                alt={`${character.name} dialogue`}
                style={{ maxWidth: '100%', maxHeight: 120, borderRadius: 4 }}
              />
            )}

            <div style={{ marginTop: '15px' }}>
              <button
                onClick={() => handleEditCharacter(character)}
                className="bg-blue-500 text-white px-4 py-2 rounded-md mr-2"
              >
                Edit
              </button>
              <button
                onClick={() => handleManageDialogue(character)}
                className="bg-green-500 text-white px-4 py-2 rounded-md mr-2"
              >
                Manage Dialogue
              </button>
              <button
                onClick={() => handleDeleteCharacter(character)}
                className="bg-red-500 text-white px-4 py-2 rounded-md"
              >
                Delete
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Create Character Button */}
      <button
        className="bg-blue-500 text-white px-4 py-2 rounded-md"
        onClick={() => setShowCreateCharacter(true)}
      >
        Create Character
      </button>

      {/* Create/Edit Character Modal */}
      {(showCreateCharacter || editingCharacter) && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
          backgroundColor: 'rgba(0,0,0,0.5)',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          zIndex: 1000
        }}>
          <div style={{
            backgroundColor: '#fff',
            padding: '30px',
            borderRadius: '8px',
            width: '600px',
            maxHeight: '80vh',
            overflow: 'auto'
          }}>
            <h2>{editingCharacter ? 'Edit Character' : 'Create Character'}</h2>
            
            {/* Character Fields */}
            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Name:</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Description:</label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px', minHeight: '60px' }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Map Icon URL:</label>
              <input
                type="text"
                value={formData.mapIconUrl}
                onChange={(e) => setFormData({ ...formData, mapIconUrl: e.target.value })}
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Dialogue Image URL:</label>
              <input
                type="text"
                value={formData.dialogueImageUrl}
                onChange={(e) => setFormData({ ...formData, dialogueImageUrl: e.target.value })}
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
              />
            </div>

            {/* Character Position Section */}
            <div style={{ marginBottom: '15px', padding: '15px', border: '1px solid #eee', borderRadius: '4px' }}>
              <h3 style={{ margin: '0 0 15px 0' }}>Character Position</h3>
              <CharacterMapPicker
                latitude={formData.movementPattern.startingLatitude}
                longitude={formData.movementPattern.startingLongitude}
                onChange={(lat, lng) => {
                  setFormData({
                    ...formData,
                    movementPattern: {
                      ...formData.movementPattern,
                      startingLatitude: lat,
                      startingLongitude: lng
                    }
                  });
                }}
              />
            </div>

            {/* Movement Pattern Section */}
            <div style={{ marginBottom: '15px', padding: '15px', border: '1px solid #eee', borderRadius: '4px' }}>
              <h3 style={{ margin: '0 0 15px 0' }}>Movement Pattern</h3>
              
              <div style={{ marginBottom: '15px' }}>
                <label style={{ display: 'block', marginBottom: '5px' }}>Movement Type:</label>
                <select
                  value={formData.movementPattern.movementPatternType}
                  onChange={(e) => setFormData({
                    ...formData,
                    movementPattern: {
                      ...formData.movementPattern,
                      movementPatternType: e.target.value as MovementPatternType
                    }
                  })}
                  style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
                >
                  <option value="static">Static</option>
                  <option value="random">Random</option>
                  <option value="path">Path</option>
                </select>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '15px', marginBottom: '15px' }}>
                <div>
                  <label style={{ display: 'block', marginBottom: '5px' }}>Starting Latitude:</label>
                  <input
                    type="number"
                    step="any"
                    value={formData.movementPattern.startingLatitude}
                    onChange={(e) => setFormData({
                      ...formData,
                      movementPattern: {
                        ...formData.movementPattern,
                        startingLatitude: parseFloat(e.target.value) || 0
                      }
                    })}
                    style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
                  />
                </div>
                <div>
                  <label style={{ display: 'block', marginBottom: '5px' }}>Starting Longitude:</label>
                  <input
                    type="number"
                    step="any"
                    value={formData.movementPattern.startingLongitude}
                    onChange={(e) => setFormData({
                      ...formData,
                      movementPattern: {
                        ...formData.movementPattern,
                        startingLongitude: parseFloat(e.target.value) || 0
                      }
                    })}
                    style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
                  />
                </div>
              </div>

              <div style={{ marginBottom: '15px' }}>
                <label style={{ display: 'block', marginBottom: '5px' }}>Zone:</label>
                <select
                  value={formData.movementPattern.zoneId}
                  onChange={(e) => setFormData({
                    ...formData,
                    movementPattern: {
                      ...formData.movementPattern,
                      zoneId: e.target.value
                    }
                  })}
                  style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
                >
                  <option value="">Select a zone (optional)</option>
                  {availableZones.map(zone => (
                    <option key={zone.id} value={zone.id}>{zone.name}</option>
                  ))}
                </select>
              </div>

              {/* Path Waypoints (only for path type) */}
              {formData.movementPattern.movementPatternType === 'path' && (
                <div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '10px' }}>
                    <label>Path Waypoints:</label>
                    <button
                      type="button"
                      onClick={addWaypoint}
                      className="bg-green-500 text-white px-3 py-1 rounded-md text-sm"
                    >
                      Add Waypoint
                    </button>
                  </div>
                  
                  {formData.movementPattern.path.map((waypoint, index) => (
                    <div key={index} style={{ display: 'grid', gridTemplateColumns: '1fr 1fr auto', gap: '10px', marginBottom: '10px' }}>
                      <input
                        type="number"
                        step="any"
                        placeholder="Latitude"
                        value={waypoint.latitude}
                        onChange={(e) => updateWaypoint(index, 'latitude', parseFloat(e.target.value) || 0)}
                        style={{ padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
                      />
                      <input
                        type="number"
                        step="any"
                        placeholder="Longitude"
                        value={waypoint.longitude}
                        onChange={(e) => updateWaypoint(index, 'longitude', parseFloat(e.target.value) || 0)}
                        style={{ padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
                      />
                      <button
                        type="button"
                        onClick={() => removeWaypoint(index)}
                        className="bg-red-500 text-white px-3 py-1 rounded-md text-sm"
                      >
                        Remove
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Modal Buttons */}
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '10px' }}>
              <button
                onClick={() => {
                  setShowCreateCharacter(false);
                  setEditingCharacter(null);
                  resetForm();
                }}
                className="bg-gray-500 text-white px-4 py-2 rounded-md"
              >
                Cancel
              </button>
              <button
                onClick={editingCharacter ? handleUpdateCharacter : handleCreateCharacter}
                className="bg-blue-500 text-white px-4 py-2 rounded-md"
              >
                {editingCharacter ? 'Update' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Dialogue Manager Modal */}
      {showDialogueManager && selectedCharacterForDialogue && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
          backgroundColor: 'rgba(0,0,0,0.5)',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          zIndex: 1000
        }}>
          <div style={{
            backgroundColor: '#fff',
            padding: '30px',
            borderRadius: '8px',
            width: '800px',
            maxHeight: '80vh',
            overflow: 'auto'
          }}>
            <h2 style={{ margin: '0 0 20px 0' }}>
              Manage Dialogue - {selectedCharacterForDialogue.name}
            </h2>

            {/* Character Actions List */}
            <div style={{ marginBottom: '20px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '15px' }}>
                <h3 style={{ margin: 0 }}>Existing Actions</h3>
                <div style={{ display: 'flex', gap: '10px' }}>
                  <button
                    onClick={() => {
                      setEditingAction(null);
                      setShowDialogueEditor(true);
                    }}
                    className="bg-blue-500 text-white px-4 py-2 rounded-md"
                  >
                    Create Talk Action
                  </button>
                  <button
                    onClick={() => {
                      setEditingAction(null);
                      setShowShopEditor(true);
                    }}
                    className="bg-green-500 text-white px-4 py-2 rounded-md"
                  >
                    Create Shop Action
                  </button>
                </div>
              </div>

              {characterActions.length === 0 ? (
                <div style={{
                  padding: '40px',
                  textAlign: 'center',
                  color: '#999',
                  fontStyle: 'italic',
                  border: '1px dashed #ccc',
                  borderRadius: '8px'
                }}>
                  No actions yet. Create one to get started.
                </div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                  {characterActions.map((action) => (
                    <div
                      key={action.id}
                      style={{
                        padding: '15px',
                        border: '1px solid #ccc',
                        borderRadius: '8px',
                        backgroundColor: '#f9f9f9'
                      }}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <div style={{ flex: 1 }}>
                          <div style={{ fontWeight: 'bold', marginBottom: '5px' }}>
                            Type: {action.actionType}
                          </div>
                          <div style={{ color: '#666', fontSize: '14px' }}>
                            {action.actionType === 'talk' ? (
                              action.dialogue.length > 0 ? (
                                <>
                                  Preview: {action.dialogue[0].text.substring(0, 100)}
                                  {action.dialogue[0].text.length > 100 ? '...' : ''}
                                </>
                              ) : (
                                'No dialogue messages'
                              )
                            ) : action.actionType === 'shop' ? (
                              action.metadata?.inventory ? (
                                <>
                                  Shop with {action.metadata.inventory.length} item{action.metadata.inventory.length !== 1 ? 's' : ''}
                                </>
                              ) : (
                                'Shop with no items'
                              )
                            ) : (
                              'Unknown action type'
                            )}
                          </div>
                          <div style={{ color: '#999', fontSize: '12px', marginTop: '5px' }}>
                            {action.actionType === 'talk' ? (
                              <>
                                {action.dialogue.length} message{action.dialogue.length !== 1 ? 's' : ''}
                              </>
                            ) : action.actionType === 'shop' ? (
                              <>
                                {action.metadata?.inventory?.length || 0} item{action.metadata?.inventory?.length !== 1 ? 's' : ''}
                              </>
                            ) : null}
                          </div>
                        </div>
                        <div style={{ display: 'flex', gap: '10px' }}>
                          <button
                            onClick={() => handleEditAction(action)}
                            className="bg-blue-500 text-white px-3 py-1 rounded-md"
                          >
                            Edit
                          </button>
                          <button
                            onClick={() => deleteCharacterAction(action.id)}
                            className="bg-red-500 text-white px-3 py-1 rounded-md"
                          >
                            Delete
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Modal Buttons */}
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '10px' }}>
              <button
                onClick={() => {
                  setShowDialogueManager(false);
                  setSelectedCharacterForDialogue(null);
                  setCharacterActions([]);
                }}
                className="bg-gray-500 text-white px-4 py-2 rounded-md"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Dialogue Editor Modal */}
      {showDialogueEditor && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
          backgroundColor: 'rgba(0,0,0,0.5)',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          zIndex: 2000
        }}>
          <div style={{
            backgroundColor: '#fff',
            padding: '30px',
            borderRadius: '8px',
            width: '900px',
            maxHeight: '90vh',
            overflow: 'hidden'
          }}>
            <h2 style={{ margin: '0 0 20px 0' }}>
              {editingAction ? 'Edit Dialogue Action' : 'Create New Dialogue Action'}
            </h2>
            <DialogueActionEditor
              action={editingAction}
              onSave={handleSaveDialogue}
              onCancel={() => {
                setShowDialogueEditor(false);
                setEditingAction(null);
              }}
            />
          </div>
        </div>
      )}

      {/* Shop Editor Modal */}
      {showShopEditor && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
          backgroundColor: 'rgba(0,0,0,0.5)',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          zIndex: 2000
        }}>
          <div style={{
            backgroundColor: '#fff',
            padding: '30px',
            borderRadius: '8px',
            width: '900px',
            maxHeight: '90vh',
            overflow: 'hidden'
          }}>
            <h2 style={{ margin: '0 0 20px 0' }}>
              {editingAction ? 'Edit Shop Action' : 'Create New Shop Action'}
            </h2>
            <ShopActionEditor
              action={editingAction}
              onSave={handleSaveShop}
              onCancel={() => {
                setShowShopEditor(false);
                setEditingAction(null);
              }}
            />
          </div>
        </div>
      )}
    </div>
  );
};
