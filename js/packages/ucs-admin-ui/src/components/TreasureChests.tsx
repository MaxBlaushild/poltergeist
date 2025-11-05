import { useAPI, useInventory, useZoneContext } from '@poltergeist/contexts';
import { TreasureChest, Zone, InventoryItem } from '@poltergeist/types';
import React, { useState, useEffect } from 'react';

interface TreasureChestItemForm {
  inventoryItemId: number;
  quantity: number;
}

export const TreasureChests = () => {
  const { apiClient } = useAPI();
  const { zones } = useZoneContext();
  const { inventoryItems } = useInventory();
  const [chests, setChests] = useState<TreasureChest[]>([]);
  const [filteredChests, setFilteredChests] = useState<TreasureChest[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [showCreateChest, setShowCreateChest] = useState(false);
  const [editingChest, setEditingChest] = useState<TreasureChest | null>(null);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [chestToDelete, setChestToDelete] = useState<TreasureChest | null>(null);

  const [formData, setFormData] = useState({
    latitude: '',
    longitude: '',
    zoneId: '',
    gold: '' as string | number,
    items: [] as TreasureChestItemForm[],
  });

  useEffect(() => {
    fetchChests();
  }, []);

  useEffect(() => {
    if (searchQuery === '') {
      setFilteredChests(chests);
    } else {
      const filtered = chests.filter(chest => {
        const zone = zones.find(z => z.id === chest.zoneId);
        return zone?.name?.toLowerCase().includes(searchQuery.toLowerCase());
      });
      setFilteredChests(filtered);
    }
  }, [searchQuery, chests, zones]);

  const fetchChests = async () => {
    try {
      const response = await apiClient.get<TreasureChest[]>('/sonar/treasure-chests');
      setChests(response);
      setFilteredChests(response);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching treasure chests:', error);
      setLoading(false);
    }
  };

  const resetForm = () => {
    setFormData({
      latitude: '',
      longitude: '',
      zoneId: '',
      gold: '',
      items: [],
    });
  };

  const handleCreateChest = async () => {
    try {
      const submitData = {
        latitude: parseFloat(formData.latitude),
        longitude: parseFloat(formData.longitude),
        zoneId: formData.zoneId,
        gold: formData.gold === '' ? undefined : parseInt(formData.gold.toString(), 10),
        items: formData.items,
      };

      const newChest = await apiClient.post<TreasureChest>('/sonar/treasure-chests', submitData);
      setChests([...chests, newChest]);
      setShowCreateChest(false);
      resetForm();
    } catch (error) {
      console.error('Error creating treasure chest:', error);
      alert('Error creating treasure chest. Please check all required fields.');
    }
  };

  const handleUpdateChest = async () => {
    if (!editingChest) return;
    
    try {
      const submitData: any = {};
      if (formData.latitude) submitData.latitude = parseFloat(formData.latitude);
      if (formData.longitude) submitData.longitude = parseFloat(formData.longitude);
      if (formData.zoneId) submitData.zoneId = formData.zoneId;
      if (formData.gold !== '') submitData.gold = parseInt(formData.gold.toString(), 10);
      if (formData.items.length > 0) submitData.items = formData.items;

      const updatedChest = await apiClient.put<TreasureChest>(`/sonar/treasure-chests/${editingChest.id}`, submitData);
      setChests(chests.map(c => c.id === editingChest.id ? updatedChest : c));
      setEditingChest(null);
      resetForm();
    } catch (error) {
      console.error('Error updating treasure chest:', error);
      alert('Error updating treasure chest.');
    }
  };

  const handleDeleteChest = async (chest: TreasureChest) => {
    setChestToDelete(chest);
    setShowDeleteConfirm(true);
  };

  const confirmDelete = async () => {
    if (!chestToDelete) return;
    
    try {
      await apiClient.delete(`/sonar/treasure-chests/${chestToDelete.id}`);
      setChests(chests.filter(c => c.id !== chestToDelete.id));
      setShowDeleteConfirm(false);
      setChestToDelete(null);
    } catch (error) {
      console.error('Error deleting treasure chest:', error);
      alert('Error deleting treasure chest.');
    }
  };

  const handleEditChest = (chest: TreasureChest) => {
    setEditingChest(chest);
    setFormData({
      latitude: chest.latitude.toString(),
      longitude: chest.longitude.toString(),
      zoneId: chest.zoneId,
      gold: chest.gold !== null && chest.gold !== undefined ? chest.gold.toString() : '',
      items: chest.items.map(item => ({
        inventoryItemId: item.inventoryItemId,
        quantity: item.quantity,
      })),
    });
  };

  const addItem = () => {
    setFormData({
      ...formData,
      items: [...formData.items, { inventoryItemId: 0, quantity: 1 }],
    });
  };

  const removeItem = (index: number) => {
    setFormData({
      ...formData,
      items: formData.items.filter((_, i) => i !== index),
    });
  };

  const updateItem = (index: number, field: keyof TreasureChestItemForm, value: number) => {
    const newItems = [...formData.items];
    newItems[index] = { ...newItems[index], [field]: value };
    setFormData({ ...formData, items: newItems });
  };

  if (loading) {
    return <div className="m-10">Loading treasure chests...</div>;
  }

  return (
    <div className="m-10">
      <h1 className="text-2xl font-bold mb-4">Treasure Chests</h1>
      
      {/* Search */}
      <div className="mb-4">
        <input
          type="text"
          placeholder="Search by zone name..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full p-2 border rounded-md"
        />
      </div>

      {/* Chests Grid */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
        gap: '20px',
        padding: '20px'
      }}>
        {filteredChests.map((chest) => {
          const zone = zones.find(z => z.id === chest.zoneId);
          return (
            <div 
              key={chest.id}
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
              }}>Treasure Chest</h2>
              
              <p style={{ margin: '5px 0', color: '#666' }}>
                Zone: {zone?.name || 'Unknown'}
              </p>
              
              <p style={{ margin: '5px 0', color: '#666' }}>
                Location: {chest.latitude.toFixed(6)}, {chest.longitude.toFixed(6)}
              </p>
              
              {chest.gold !== null && chest.gold !== undefined && (
                <p style={{ margin: '5px 0', color: '#666' }}>
                  Gold: {chest.gold}
                </p>
              )}

              {chest.items && chest.items.length > 0 && (
                <div style={{ marginTop: '10px' }}>
                  <strong style={{ color: '#666' }}>Items:</strong>
                  <ul style={{ margin: '5px 0', paddingLeft: '20px', color: '#666' }}>
                    {chest.items.map((item, idx) => (
                      <li key={idx}>
                        {item.inventoryItem?.name || `Item ${item.inventoryItemId}`} x{item.quantity}
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              <div style={{ marginTop: '15px' }}>
                <button
                  onClick={() => handleEditChest(chest)}
                  className="bg-blue-500 text-white px-4 py-2 rounded-md mr-2"
                >
                  Edit
                </button>
                <button
                  onClick={() => handleDeleteChest(chest)}
                  className="bg-red-500 text-white px-4 py-2 rounded-md"
                >
                  Delete
                </button>
              </div>
            </div>
          );
        })}
      </div>

      {/* Create Chest Button */}
      <button
        className="bg-blue-500 text-white px-4 py-2 rounded-md"
        onClick={() => setShowCreateChest(true)}
      >
        Create Treasure Chest
      </button>

      {/* Create/Edit Chest Modal */}
      {(showCreateChest || editingChest) && (
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
            <h2>{editingChest ? 'Edit Treasure Chest' : 'Create Treasure Chest'}</h2>
            
            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Zone *:</label>
              <select
                value={formData.zoneId}
                onChange={(e) => setFormData({ ...formData, zoneId: e.target.value })}
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
                required
              >
                <option value="">Select a zone</option>
                {zones.map(zone => (
                  <option key={zone.id} value={zone.id}>{zone.name}</option>
                ))}
              </select>
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Latitude *:</label>
              <input
                type="number"
                step="any"
                value={formData.latitude}
                onChange={(e) => setFormData({ ...formData, latitude: e.target.value })}
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
                required
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Longitude *:</label>
              <input
                type="number"
                step="any"
                value={formData.longitude}
                onChange={(e) => setFormData({ ...formData, longitude: e.target.value })}
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
                required
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Gold (optional):</label>
              <input
                type="number"
                min="0"
                value={formData.gold}
                onChange={(e) => setFormData({ ...formData, gold: e.target.value === '' ? '' : parseInt(e.target.value, 10) })}
                placeholder="Leave empty for no gold"
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '10px' }}>
                <label style={{ display: 'block' }}>Items:</label>
                <button
                  type="button"
                  onClick={addItem}
                  className="bg-green-500 text-white px-3 py-1 rounded-md text-sm"
                >
                  Add Item
                </button>
              </div>
              {formData.items.map((item, index) => (
                <div key={index} style={{ 
                  display: 'flex', 
                  gap: '10px', 
                  marginBottom: '10px',
                  padding: '10px',
                  border: '1px solid #ccc',
                  borderRadius: '4px'
                }}>
                  <select
                    value={item.inventoryItemId}
                    onChange={(e) => updateItem(index, 'inventoryItemId', parseInt(e.target.value, 10))}
                    style={{ flex: 1, padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
                  >
                    <option value="0">Select item</option>
                    {inventoryItems.map(invItem => (
                      <option key={invItem.id} value={invItem.id}>{invItem.name}</option>
                    ))}
                  </select>
                  <input
                    type="number"
                    min="1"
                    value={item.quantity}
                    onChange={(e) => updateItem(index, 'quantity', parseInt(e.target.value, 10))}
                    placeholder="Qty"
                    style={{ width: '80px', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
                  />
                  <button
                    type="button"
                    onClick={() => removeItem(index)}
                    className="bg-red-500 text-white px-3 py-1 rounded-md"
                  >
                    Remove
                  </button>
                </div>
              ))}
            </div>

            <div style={{ marginTop: '20px', display: 'flex', gap: '10px' }}>
              <button
                onClick={() => {
                  if (editingChest) {
                    handleUpdateChest();
                  } else {
                    handleCreateChest();
                  }
                }}
                className="bg-blue-500 text-white px-4 py-2 rounded-md"
              >
                {editingChest ? 'Update' : 'Create'}
              </button>
              <button
                onClick={() => {
                  setShowCreateChest(false);
                  setEditingChest(null);
                  resetForm();
                }}
                className="bg-gray-500 text-white px-4 py-2 rounded-md"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      {showDeleteConfirm && chestToDelete && (
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
            width: '400px'
          }}>
            <h2>Confirm Delete</h2>
            <p>Are you sure you want to delete this treasure chest? This action cannot be undone.</p>
            <div style={{ marginTop: '20px', display: 'flex', gap: '10px' }}>
              <button
                onClick={confirmDelete}
                className="bg-red-500 text-white px-4 py-2 rounded-md"
              >
                Delete
              </button>
              <button
                onClick={() => {
                  setShowDeleteConfirm(false);
                  setChestToDelete(null);
                }}
                className="bg-gray-500 text-white px-4 py-2 rounded-md"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

