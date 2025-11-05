import { useAPI } from '@poltergeist/contexts';
import { InventoryItem, Rarity } from '@poltergeist/types';
import React, { useState, useEffect } from 'react';

export const InventoryItems = () => {
  const { apiClient } = useAPI();
  const [items, setItems] = useState<InventoryItem[]>([]);
  const [filteredItems, setFilteredItems] = useState<InventoryItem[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [showCreateItem, setShowCreateItem] = useState(false);
  const [editingItem, setEditingItem] = useState<InventoryItem | null>(null);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [itemToDelete, setItemToDelete] = useState<InventoryItem | null>(null);

  const [formData, setFormData] = useState({
    name: '',
    imageUrl: '',
    flavorText: '',
    effectText: '',
    rarityTier: 'Common' as string,
    isCaptureType: false,
  });

  useEffect(() => {
    fetchItems();
  }, []);

  useEffect(() => {
    if (searchQuery === '') {
      setFilteredItems(items);
    } else {
      const filtered = items.filter(item =>
        item.name?.toLowerCase().includes(searchQuery.toLowerCase())
      );
      setFilteredItems(filtered);
    }
  }, [searchQuery, items]);

  const fetchItems = async () => {
    try {
      const response = await apiClient.get<InventoryItem[]>('/sonar/inventory-items');
      setItems(response);
      setFilteredItems(response);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching inventory items:', error);
      setLoading(false);
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      imageUrl: '',
      flavorText: '',
      effectText: '',
      rarityTier: 'Common',
      isCaptureType: false,
    });
  };

  const handleCreateItem = async () => {
    try {
      const newItem = await apiClient.post<InventoryItem>('/sonar/inventory-items', formData);
      setItems([...items, newItem]);
      setShowCreateItem(false);
      resetForm();
    } catch (error) {
      console.error('Error creating inventory item:', error);
      alert('Error creating inventory item. Please check all required fields.');
    }
  };

  const handleUpdateItem = async () => {
    if (!editingItem) return;
    
    try {
      const updatedItem = await apiClient.put<InventoryItem>(`/sonar/inventory-items/${editingItem.id}`, formData);
      setItems(items.map(i => i.id === editingItem.id ? updatedItem : i));
      setEditingItem(null);
      resetForm();
    } catch (error) {
      console.error('Error updating inventory item:', error);
      alert('Error updating inventory item. Please check all required fields.');
    }
  };

  const handleDeleteItem = async (item: InventoryItem) => {
    setItemToDelete(item);
    setShowDeleteConfirm(true);
  };

  const confirmDelete = async () => {
    if (!itemToDelete) return;
    
    try {
      await apiClient.delete(`/sonar/inventory-items/${itemToDelete.id}`);
      setItems(items.filter(i => i.id !== itemToDelete.id));
      setShowDeleteConfirm(false);
      setItemToDelete(null);
    } catch (error) {
      console.error('Error deleting inventory item:', error);
      alert('Error deleting inventory item.');
    }
  };

  const handleEditItem = (item: InventoryItem) => {
    setEditingItem(item);
    setFormData({
      name: item.name,
      imageUrl: item.imageUrl,
      flavorText: item.flavorText,
      effectText: item.effectText,
      rarityTier: item.rarityTier,
      isCaptureType: item.isCaptureType,
    });
  };

  if (loading) {
    return <div className="m-10">Loading inventory items...</div>;
  }

  return (
    <div className="m-10">
      <h1 className="text-2xl font-bold mb-4">Inventory Items</h1>
      
      {/* Search */}
      <div className="mb-4">
        <input
          type="text"
          placeholder="Search inventory items..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full p-2 border rounded-md"
        />
      </div>

      {/* Items Grid */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
        gap: '20px',
        padding: '20px'
      }}>
        {filteredItems.map((item) => (
          <div 
            key={item.id}
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
            }}>{item.name}</h2>
            
            <p style={{ margin: '5px 0', color: '#666' }}>
              Rarity: {item.rarityTier}
            </p>
            
            <p style={{ margin: '5px 0', color: '#666' }}>
              Capture Type: {item.isCaptureType ? 'Yes' : 'No'}
            </p>

            {item.imageUrl && (
              <img
                src={item.imageUrl}
                alt={item.name}
                style={{ maxWidth: '100%', maxHeight: 120, borderRadius: 4, marginTop: '10px' }}
              />
            )}

            <p style={{ margin: '10px 0', color: '#666', fontSize: '14px' }}>
              <strong>Flavor:</strong> {item.flavorText || '—'}
            </p>

            <p style={{ margin: '10px 0', color: '#666', fontSize: '14px' }}>
              <strong>Effect:</strong> {item.effectText || '—'}
            </p>

            <div style={{ marginTop: '15px' }}>
              <button
                onClick={() => handleEditItem(item)}
                className="bg-blue-500 text-white px-4 py-2 rounded-md mr-2"
              >
                Edit
              </button>
              <button
                onClick={() => handleDeleteItem(item)}
                className="bg-red-500 text-white px-4 py-2 rounded-md"
              >
                Delete
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Create Item Button */}
      <button
        className="bg-blue-500 text-white px-4 py-2 rounded-md"
        onClick={() => setShowCreateItem(true)}
      >
        Create Inventory Item
      </button>

      {/* Create/Edit Item Modal */}
      {(showCreateItem || editingItem) && (
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
            <h2>{editingItem ? 'Edit Inventory Item' : 'Create Inventory Item'}</h2>
            
            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Name *:</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
                required
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Image URL:</label>
              <input
                type="text"
                value={formData.imageUrl}
                onChange={(e) => setFormData({ ...formData, imageUrl: e.target.value })}
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Flavor Text:</label>
              <textarea
                value={formData.flavorText}
                onChange={(e) => setFormData({ ...formData, flavorText: e.target.value })}
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px', minHeight: '60px' }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Effect Text:</label>
              <textarea
                value={formData.effectText}
                onChange={(e) => setFormData({ ...formData, effectText: e.target.value })}
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px', minHeight: '60px' }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Rarity Tier *:</label>
              <select
                value={formData.rarityTier}
                onChange={(e) => setFormData({ ...formData, rarityTier: e.target.value })}
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
                required
              >
                <option value={Rarity.Common}>Common</option>
                <option value={Rarity.Uncommon}>Uncommon</option>
                <option value={Rarity.Epic}>Epic</option>
                <option value={Rarity.Mythic}>Mythic</option>
                <option value="Not Droppable">Not Droppable</option>
              </select>
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <input
                  type="checkbox"
                  checked={formData.isCaptureType}
                  onChange={(e) => setFormData({ ...formData, isCaptureType: e.target.checked })}
                />
                Is Capture Type
              </label>
            </div>

            <div style={{ marginTop: '20px', display: 'flex', gap: '10px' }}>
              <button
                onClick={() => {
                  if (editingItem) {
                    handleUpdateItem();
                  } else {
                    handleCreateItem();
                  }
                }}
                className="bg-blue-500 text-white px-4 py-2 rounded-md"
              >
                {editingItem ? 'Update' : 'Create'}
              </button>
              <button
                onClick={() => {
                  setShowCreateItem(false);
                  setEditingItem(null);
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
      {showDeleteConfirm && itemToDelete && (
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
            <p>Are you sure you want to delete "{itemToDelete.name}"? This action cannot be undone.</p>
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
                  setItemToDelete(null);
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

