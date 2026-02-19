import { useAPI, useMediaContext } from '@poltergeist/contexts';
import { InventoryItem, Rarity } from '@poltergeist/types';
import React, { useMemo, useState, useEffect, useRef } from 'react';
import { useUsers } from '../hooks/useUsers.ts';

type SelectOption = {
  value: string;
  label: string;
  secondary?: string;
};

const SearchableSelect = ({
  label,
  placeholder,
  options,
  value,
  onChange,
}: {
  label: string;
  placeholder: string;
  options: SelectOption[];
  value: string;
  onChange: (value: string) => void;
}) => {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');

  const selected = options.find((o) => o.value === value);
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return options;
    return options.filter((o) => {
      const hay = `${o.label} ${o.secondary ?? ''}`.toLowerCase();
      return hay.includes(q);
    });
  }, [options, query]);

  const displayValue = open ? query : selected?.label ?? '';

  return (
    <div className="relative">
      <label className="block text-sm font-medium text-gray-700">{label}</label>
      <input
        value={displayValue}
        onChange={(e) => {
          setQuery(e.target.value);
          setOpen(true);
        }}
        onFocus={() => {
          setOpen(true);
          setQuery('');
        }}
        onBlur={() => {
          setTimeout(() => setOpen(false), 150);
        }}
        placeholder={placeholder}
        className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
      />
      {open && (
        <div className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md border border-gray-200 bg-white shadow-lg">
          {filtered.length === 0 && (
            <div className="px-3 py-2 text-sm text-gray-500">No matches found</div>
          )}
          {filtered.map((option) => (
            <button
              type="button"
              key={option.value}
              onMouseDown={(e) => e.preventDefault()}
              onClick={() => {
                onChange(option.value);
                setOpen(false);
                setQuery('');
              }}
              className="flex w-full flex-col items-start px-3 py-2 text-left text-sm hover:bg-indigo-50"
            >
              <span className="font-medium text-gray-900">{option.label}</span>
              {option.secondary && (
                <span className="text-xs text-gray-500">{option.secondary}</span>
              )}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export const InventoryItems = () => {
  const { apiClient } = useAPI();
  const { uploadMedia, getPresignedUploadURL } = useMediaContext();
  const { users } = useUsers();
  const [items, setItems] = useState<InventoryItem[]>([]);
  const [filteredItems, setFilteredItems] = useState<InventoryItem[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [showCreateItem, setShowCreateItem] = useState(false);
  const [showGenerateItem, setShowGenerateItem] = useState(false);
  const [editingItem, setEditingItem] = useState<InventoryItem | null>(null);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [itemToDelete, setItemToDelete] = useState<InventoryItem | null>(null);
  const [imageFile, setImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [useOutfitItem, setUseOutfitItem] = useState<InventoryItem | null>(null);
  const [useOutfitUser, setUseOutfitUser] = useState('');
  const [useOutfitSelfieUrl, setUseOutfitSelfieUrl] = useState('');
  const [useOutfitStatus, setUseOutfitStatus] = useState<string | null>(null);
  const [useOutfitStatusKind, setUseOutfitStatusKind] = useState<'success' | 'error' | null>(null);
  const [useOutfitSubmitting, setUseOutfitSubmitting] = useState(false);

  const [formData, setFormData] = useState({
    name: '',
    imageUrl: '',
    flavorText: '',
    effectText: '',
    rarityTier: 'Common' as string,
    isCaptureType: false,
    sellValue: undefined as number | undefined,
    unlockTier: undefined as number | undefined,
  });

  const [generationData, setGenerationData] = useState({
    name: '',
    description: '',
    rarityTier: 'Common' as string,
  });

  const userOptions = useMemo(() => {
    return (users ?? []).map((user) => {
      const username = user.username?.trim() ? `@${user.username}` : '';
      const display = username || user.name || user.phoneNumber;
      const secondary = username ? user.name : user.phoneNumber;
      return {
        value: user.id,
        label: display,
        secondary: secondary && secondary !== display ? secondary : undefined,
      };
    });
  }, [users]);

  useEffect(() => {
    fetchItems();
  }, []);

  useEffect(() => {
    const hasPending = items.some(item =>
      ['queued', 'in_progress'].includes(item.imageGenerationStatus || '')
    );
    if (!hasPending) return;

    const interval = setInterval(() => {
      fetchItems();
    }, 5000);

    return () => clearInterval(interval);
  }, [items]);

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
      sellValue: undefined,
      unlockTier: undefined,
    });
    setImageFile(null);
    setImagePreview(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const resetGenerationForm = () => {
    setGenerationData({
      name: '',
      description: '',
      rarityTier: 'Common',
    });
  };

  const handleCreateItem = async () => {
    try {
      let imageUrl = formData.imageUrl;

      // Upload image to S3 if a file is selected
      if (imageFile) {
        const getExtension = (filename: string): string => {
          return filename.split('.').pop()?.toLowerCase() || 'jpg';
        };
        const extension = getExtension(imageFile.name);
        const timestamp = Date.now();
        const imageKey = `inventory-items/${timestamp}-${Math.random().toString(36).substring(2, 15)}.${extension}`;

        const presignedUrl = await getPresignedUploadURL('crew-points-of-interest', imageKey);
        if (!presignedUrl) {
          alert('Failed to get upload URL. Please try again.');
          return;
        }

        const uploadSuccess = await uploadMedia(presignedUrl, imageFile);
        if (!uploadSuccess) {
          alert('Failed to upload image. Please try again.');
          return;
        }

        imageUrl = presignedUrl.split('?')[0];
      }

      const submitData = { ...formData, imageUrl };
      const newItem = await apiClient.post<InventoryItem>('/sonar/inventory-items', submitData);
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
      let imageUrl = formData.imageUrl;

      // Upload new image to S3 if a file is selected, otherwise keep existing imageUrl
      if (imageFile) {
        const getExtension = (filename: string): string => {
          return filename.split('.').pop()?.toLowerCase() || 'jpg';
        };
        const extension = getExtension(imageFile.name);
        const timestamp = Date.now();
        const imageKey = `inventory-items/${timestamp}-${Math.random().toString(36).substring(2, 15)}.${extension}`;

        const presignedUrl = await getPresignedUploadURL('crew-points-of-interest', imageKey);
        if (!presignedUrl) {
          alert('Failed to get upload URL. Please try again.');
          return;
        }

        const uploadSuccess = await uploadMedia(presignedUrl, imageFile);
        if (!uploadSuccess) {
          alert('Failed to upload image. Please try again.');
          return;
        }

        imageUrl = presignedUrl.split('?')[0];
      }

      const submitData = { ...formData, imageUrl };
      const updatedItem = await apiClient.put<InventoryItem>(`/sonar/inventory-items/${editingItem.id}`, submitData);
      setItems(items.map(i => i.id === editingItem.id ? updatedItem : i));
      setEditingItem(null);
      resetForm();
    } catch (error) {
      console.error('Error updating inventory item:', error);
      alert('Error updating inventory item. Please check all required fields.');
    }
  };

  const handleGenerateItem = async () => {
    try {
      const newItem = await apiClient.post<InventoryItem>('/sonar/inventory-items/generate', {
        name: generationData.name,
        description: generationData.description,
        rarityTier: generationData.rarityTier,
      });
      setItems([...items, newItem]);
      setShowGenerateItem(false);
      resetGenerationForm();
    } catch (error) {
      console.error('Error generating inventory item:', error);
      alert('Error generating inventory item. Please check all required fields.');
    }
  };

  const handleRegenerateImage = async (item: InventoryItem) => {
    try {
      const updated = await apiClient.post<InventoryItem>(`/sonar/inventory-items/${item.id}/regenerate`, {});
      setItems(items.map(i => i.id === item.id ? updated : i));
    } catch (error) {
      console.error('Error regenerating inventory item image:', error);
      alert('Error regenerating inventory item image.');
    }
  };

  const handleUseOutfit = async () => {
    if (!useOutfitItem) return;
    try {
      setUseOutfitSubmitting(true);
      setUseOutfitStatus(null);
      setUseOutfitStatusKind(null);
      await apiClient.post('/sonar/admin/useOutfitItem', {
        userID: useOutfitUser,
        itemID: useOutfitItem.id,
        selfieUrl: useOutfitSelfieUrl,
      });
      setUseOutfitStatus('Outfit generation queued.');
      setUseOutfitStatusKind('success');
    } catch (error) {
      console.error('Error using outfit item:', error);
      setUseOutfitStatus('Failed to start outfit generation.');
      setUseOutfitStatusKind('error');
    } finally {
      setUseOutfitSubmitting(false);
    }
  };

  const isOutfitName = (name?: string) =>
    (name || '').trim().toLowerCase().endsWith('outfit');

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
      sellValue: item.sellValue,
      unlockTier: item.unlockTier,
    });
    setImageFile(null);
    setImagePreview(item.imageUrl || null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      setImageFile(file);
      
      // Create preview URL
      const reader = new FileReader();
      reader.onloadend = () => {
        setImagePreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };

  const formatGenerationStatus = (status?: string) => {
    switch (status) {
      case 'queued':
        return 'Queued';
      case 'in_progress':
        return 'Generating';
      case 'complete':
        return 'Complete';
      case 'failed':
        return 'Failed';
      case 'none':
        return 'Not requested';
      default:
        return 'Unknown';
    }
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
              ID: {item.id}
            </p>

            <p style={{ margin: '5px 0', color: '#666' }}>
              Image Status: {formatGenerationStatus(item.imageGenerationStatus)}
            </p>
            {item.imageGenerationStatus === 'failed' && item.imageGenerationError && (
              <p style={{ margin: '5px 0', color: '#b91c1c', fontSize: '12px' }}>
                Error: {item.imageGenerationError}
              </p>
            )}
            
            <p style={{ margin: '5px 0', color: '#666' }}>
              Rarity: {item.rarityTier}
            </p>
            
            <p style={{ margin: '5px 0', color: '#666' }}>
              Capture Type: {item.isCaptureType ? 'Yes' : 'No'}
            </p>
            
            {item.sellValue !== undefined && item.sellValue !== null && (
              <p style={{ margin: '5px 0', color: '#666' }}>
                Sell Value: {item.sellValue} gold
              </p>
            )}

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
              {isOutfitName(item.name) && (
                <button
                  onClick={() => {
                    setUseOutfitItem(item);
                    setUseOutfitUser('');
                    setUseOutfitSelfieUrl('');
                    setUseOutfitStatus(null);
                    setUseOutfitStatusKind(null);
                  }}
                  className="bg-indigo-600 text-white px-4 py-2 rounded-md mr-2"
                >
                  Use Outfit
                </button>
              )}
              <button
                onClick={() => handleRegenerateImage(item)}
                className="bg-yellow-500 text-white px-4 py-2 rounded-md mr-2"
                disabled={['queued', 'in_progress'].includes(item.imageGenerationStatus || '')}
              >
                Regenerate Image
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

      {/* Create Item Buttons */}
      <div style={{ display: 'flex', gap: '10px' }}>
        <button
          className="bg-blue-500 text-white px-4 py-2 rounded-md"
          onClick={() => setShowCreateItem(true)}
        >
          Create Inventory Item
        </button>
        <button
          className="bg-green-600 text-white px-4 py-2 rounded-md"
          onClick={() => setShowGenerateItem(true)}
        >
          Generate Inventory Item
        </button>
      </div>

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
              <label style={{ display: 'block', marginBottom: '5px' }}>Image:</label>
              <input
                type="file"
                accept="image/*"
                ref={fileInputRef}
                onChange={handleImageChange}
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
              />
              {imagePreview && (
                <img 
                  src={imagePreview} 
                  alt="Preview" 
                  style={{ 
                    maxWidth: '100%', 
                    maxHeight: 200, 
                    borderRadius: 4, 
                    marginTop: '10px',
                    objectFit: 'contain'
                  }} 
                />
              )}
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

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Sell Value (gold):</label>
              <input
                type="number"
                min="0"
                value={formData.sellValue !== undefined ? formData.sellValue : ''}
                onChange={(e) => setFormData({ 
                  ...formData, 
                  sellValue: e.target.value === '' ? undefined : parseInt(e.target.value, 10) 
                })}
                placeholder="Leave empty if item cannot be sold"
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
              />
              <small style={{ color: '#666', fontSize: '12px' }}>
                Set the amount of gold this item sells for. Leave empty if the item cannot be sold.
              </small>
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Unlock Tier:</label>
              <input
                type="number"
                min="0"
                value={formData.unlockTier !== undefined ? formData.unlockTier : ''}
                onChange={(e) => setFormData({ 
                  ...formData, 
                  unlockTier: e.target.value === '' ? undefined : parseInt(e.target.value, 10) 
                })}
                placeholder="Leave empty if no unlock tier required"
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
              />
              <small style={{ color: '#666', fontSize: '12px' }}>
                Set the tier level required to unlock this item. Leave empty if no tier requirement.
              </small>
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

      {useOutfitItem && (
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
            width: '520px',
            maxHeight: '80vh',
            overflow: 'auto'
          }}>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold">Use Outfit</h2>
              <button
                onClick={() => setUseOutfitItem(null)}
                className="text-gray-500 hover:text-gray-700"
              >
                ✕
              </button>
            </div>

            <div className="mb-4 text-sm text-gray-600">
              Selected item: <span className="font-medium text-gray-900">{useOutfitItem.name}</span>
            </div>

            <div className="mb-4">
              <SearchableSelect
                label="User"
                placeholder="Search by username or name…"
                options={userOptions}
                value={useOutfitUser}
                onChange={setUseOutfitUser}
              />
            </div>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700">Selfie URL</label>
              <input
                type="text"
                value={useOutfitSelfieUrl}
                onChange={(e) => setUseOutfitSelfieUrl(e.target.value)}
                placeholder="https://..."
                className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
            </div>

            {useOutfitStatus && (
              <div
                className={`mb-4 rounded-md border px-3 py-2 text-sm ${
                  useOutfitStatusKind === 'error'
                    ? 'border-rose-200 bg-rose-50 text-rose-800'
                    : 'border-emerald-200 bg-emerald-50 text-emerald-800'
                }`}
              >
                {useOutfitStatus}
              </div>
            )}

            <div className="flex gap-2">
              <button
                onClick={handleUseOutfit}
                disabled={!useOutfitUser || !useOutfitSelfieUrl || useOutfitSubmitting}
                className="bg-indigo-600 text-white px-4 py-2 rounded-md disabled:opacity-60"
              >
                {useOutfitSubmitting ? 'Starting…' : 'Start Generation'}
              </button>
              <button
                onClick={() => setUseOutfitItem(null)}
                className="bg-gray-100 text-gray-700 px-4 py-2 rounded-md"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Generate Item Modal */}
      {showGenerateItem && (
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
            width: '500px',
            maxHeight: '80vh',
            overflow: 'auto'
          }}>
            <h2>Generate Inventory Item</h2>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Name *:</label>
              <input
                type="text"
                value={generationData.name}
                onChange={(e) => setGenerationData({ ...generationData, name: e.target.value })}
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
                required
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Description:</label>
              <textarea
                value={generationData.description}
                onChange={(e) => setGenerationData({ ...generationData, description: e.target.value })}
                style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px', minHeight: '80px' }}
              />
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '5px' }}>Rarity Tier *:</label>
              <select
                value={generationData.rarityTier}
                onChange={(e) => setGenerationData({ ...generationData, rarityTier: e.target.value })}
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

            <div style={{ marginTop: '20px', display: 'flex', gap: '10px' }}>
              <button
                onClick={handleGenerateItem}
                className="bg-green-600 text-white px-4 py-2 rounded-md"
              >
                Generate
              </button>
              <button
                onClick={() => {
                  setShowGenerateItem(false);
                  resetGenerationForm();
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
