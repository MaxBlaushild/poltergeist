import { useAPI, useInventory, useZoneContext } from '@poltergeist/contexts';
import { TreasureChest, Zone, InventoryItem } from '@poltergeist/types';
import React, { useState, useEffect, useRef } from 'react';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';

mapboxgl.accessToken = process.env.REACT_APP_MAPBOX_ACCESS_TOKEN || '';

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
  const [seeding, setSeeding] = useState(false);
  const [zoneQuery, setZoneQuery] = useState('');
  const [showZoneSuggestions, setShowZoneSuggestions] = useState(false);

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
    setZoneQuery('');
    setShowZoneSuggestions(false);
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

  const handleSeedTreasureChests = async () => {
    setSeeding(true);
    try {
      await apiClient.post('/sonar/admin/treasure-chests/seed');
      alert('Treasure chest seeding job queued successfully!');
      // Optionally refresh the chest list after a delay
      setTimeout(() => {
        fetchChests();
      }, 2000);
    } catch (error) {
      console.error('Error queueing seed treasure chests job:', error);
      alert('Error queueing seed treasure chests job.');
    } finally {
      setSeeding(false);
    }
  };

  const handleEditChest = (chest: TreasureChest) => {
    setEditingChest(chest);
    const zoneName = zones.find(z => z.id === chest.zoneId)?.name || '';
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
    setZoneQuery(zoneName);
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

  const filteredZones = zones.filter(zone =>
    zone.name.toLowerCase().includes(zoneQuery.toLowerCase())
  );

  return (
    <div className="m-10">
      <div className="flex flex-col gap-3 mb-4 md:flex-row md:items-center md:justify-between">
        <h1 className="text-2xl font-bold">Treasure Chests</h1>
        <div className="flex flex-wrap gap-2">
          <button
            className="bg-blue-500 text-white px-4 py-2 rounded-md"
            onClick={() => setShowCreateChest(true)}
          >
            Create Treasure Chest
          </button>
          <button
            className="bg-green-500 text-white px-4 py-2 rounded-md disabled:opacity-50 disabled:cursor-not-allowed"
            onClick={handleSeedTreasureChests}
            disabled={seeding}
          >
            {seeding ? 'Queuing...' : 'Seed Treasure Chests'}
          </button>
        </div>
      </div>
      
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
              <div style={{ position: 'relative' }}>
                <input
                  type="text"
                  value={zoneQuery}
                  onChange={(e) => {
                    const value = e.target.value;
                    setZoneQuery(value);
                    setShowZoneSuggestions(true);
                    if (value.trim() === '') {
                      setFormData({ ...formData, zoneId: '' });
                    }
                  }}
                  onFocus={() => setShowZoneSuggestions(true)}
                  onBlur={() => {
                    setTimeout(() => setShowZoneSuggestions(false), 120);
                  }}
                  placeholder="Type to filter zones..."
                  style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px' }}
                />
                {showZoneSuggestions && filteredZones.length > 0 && (
                  <div
                    style={{
                      position: 'absolute',
                      top: '100%',
                      left: 0,
                      right: 0,
                      backgroundColor: '#fff',
                      border: '1px solid #ccc',
                      borderRadius: '4px',
                      marginTop: '4px',
                      maxHeight: '200px',
                      overflowY: 'auto',
                      zIndex: 20,
                    }}
                  >
                    {filteredZones.map(zone => (
                      <button
                        type="button"
                        key={zone.id}
                        onClick={() => {
                          setFormData({ ...formData, zoneId: zone.id });
                          setZoneQuery(zone.name);
                          setShowZoneSuggestions(false);
                        }}
                        style={{
                          display: 'block',
                          width: '100%',
                          textAlign: 'left',
                          padding: '8px 10px',
                          background: 'none',
                          border: 'none',
                          cursor: 'pointer',
                        }}
                      >
                        {zone.name}
                      </button>
                    ))}
                  </div>
                )}
              </div>
              {!formData.zoneId && (
                <div style={{ marginTop: '6px', fontSize: '12px', color: '#999' }}>
                  Select a zone to continue.
                </div>
              )}
            </div>

            <div style={{ marginBottom: '15px' }}>
              <label style={{ display: 'block', marginBottom: '8px' }}>Placement *:</label>
              <TreasureChestMapPicker
                latitude={parseFloat(formData.latitude) || 0}
                longitude={parseFloat(formData.longitude) || 0}
                onChange={(lat, lng) =>
                  setFormData({
                    ...formData,
                    latitude: lat.toFixed(6),
                    longitude: lng.toFixed(6),
                  })
                }
              />
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

interface TreasureChestMapPickerProps {
  latitude: number;
  longitude: number;
  onChange: (lat: number, lng: number) => void;
}

const TreasureChestMapPicker: React.FC<TreasureChestMapPickerProps> = ({
  latitude,
  longitude,
  onChange,
}) => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<mapboxgl.Map | null>(null);
  const marker = useRef<mapboxgl.Marker | null>(null);
  const [isLoaded, setIsLoaded] = useState(false);
  const [locating, setLocating] = useState(false);
  const [locationError, setLocationError] = useState<string | null>(null);
  const locateTimeout = useRef<number | null>(null);
  const locateWatchId = useRef<number | null>(null);

  const defaultLat = 40.7128;
  const defaultLng = -74.0060;
  const initialLat = latitude || defaultLat;
  const initialLng = longitude || defaultLng;

  useEffect(() => {
    if (!mapContainer.current || map.current) return;

    map.current = new mapboxgl.Map({
      container: mapContainer.current,
      style: 'mapbox://styles/maxblaushild/clzq7o8pr00ce01qgey4y0g31',
      center: [initialLng, initialLat],
      zoom: 16,
    });

    map.current.addControl(new mapboxgl.NavigationControl());

    const el = document.createElement('div');
    el.className = 'custom-marker';
    el.style.width = '30px';
    el.style.height = '30px';
    el.style.backgroundImage = 'url(https://docs.mapbox.com/mapbox-gl-js/assets/custom_marker.png)';
    el.style.backgroundSize = 'cover';
    el.style.cursor = 'grab';

    marker.current = new mapboxgl.Marker({ element: el, draggable: true })
      .setLngLat([initialLng, initialLat])
      .addTo(map.current);

    marker.current.on('dragend', () => {
      const lngLat = marker.current!.getLngLat();
      onChange(lngLat.lat, lngLat.lng);
    });

    map.current.on('click', (e) => {
      if (marker.current) {
        marker.current.setLngLat([e.lngLat.lng, e.lngLat.lat]);
        onChange(e.lngLat.lat, e.lngLat.lng);
      }
    });

    map.current.on('load', () => {
      setIsLoaded(true);
      map.current?.resize();
    });

    return () => {
      if (locateTimeout.current) {
        window.clearTimeout(locateTimeout.current);
        locateTimeout.current = null;
      }
      if (locateWatchId.current !== null) {
        navigator.geolocation?.clearWatch(locateWatchId.current);
        locateWatchId.current = null;
      }
      if (map.current) {
        map.current.remove();
        map.current = null;
      }
    };
  }, [initialLat, initialLng, onChange]);

  useEffect(() => {
    if (map.current && isLoaded && marker.current) {
      const current = marker.current.getLngLat();
      if (
        Math.abs(current.lat - initialLat) > 0.0001 ||
        Math.abs(current.lng - initialLng) > 0.0001
      ) {
        marker.current.setLngLat([initialLng, initialLat]);
        map.current.easeTo({ center: [initialLng, initialLat] });
      }
    }
  }, [initialLat, initialLng, isLoaded]);

  const handleSnapToLocation = () => {
    if (!navigator.geolocation) {
      setLocationError('Geolocation is not supported in this browser.');
      return;
    }
    const startWatch = () => {
      setLocating(true);
      setLocationError(null);
      if (locateTimeout.current) {
        window.clearTimeout(locateTimeout.current);
      }
      if (locateWatchId.current !== null) {
        navigator.geolocation.clearWatch(locateWatchId.current);
        locateWatchId.current = null;
      }
      locateTimeout.current = window.setTimeout(() => {
        if (locateWatchId.current !== null) {
          navigator.geolocation.clearWatch(locateWatchId.current);
          locateWatchId.current = null;
        }
        setLocationError('Location request timed out.');
        setLocating(false);
        locateTimeout.current = null;
      }, 12000);
      locateWatchId.current = navigator.geolocation.watchPosition(
        (pos) => {
          const { latitude: lat, longitude: lng } = pos.coords;
          if (locateTimeout.current) {
            window.clearTimeout(locateTimeout.current);
            locateTimeout.current = null;
          }
          if (locateWatchId.current !== null) {
            navigator.geolocation.clearWatch(locateWatchId.current);
            locateWatchId.current = null;
          }
          onChange(lat, lng);
          if (marker.current) {
            marker.current.setLngLat([lng, lat]);
          }
          map.current?.easeTo({ center: [lng, lat], zoom: 16 });
          setLocating(false);
        },
        (err) => {
          if (locateTimeout.current) {
            window.clearTimeout(locateTimeout.current);
            locateTimeout.current = null;
          }
          if (locateWatchId.current !== null) {
            navigator.geolocation.clearWatch(locateWatchId.current);
            locateWatchId.current = null;
          }
          setLocationError(err.message || 'Unable to fetch location.');
          setLocating(false);
        },
        { enableHighAccuracy: true, maximumAge: 0 }
      );
    };

    const permissions = (navigator as any).permissions;
    if (permissions?.query) {
      permissions
        .query({ name: 'geolocation' })
        .then((status: { state?: string }) => {
          if (status.state === 'denied') {
            setLocationError('Location permission denied in browser settings.');
            setLocating(false);
            return;
          }
          startWatch();
        })
        .catch(() => startWatch());
    } else {
      startWatch();
    }
  };

  return (
    <div>
      <div
        ref={mapContainer}
        style={{
          width: '100%',
          height: '320px',
          borderRadius: '8px',
          border: '1px solid #ccc',
          overflow: 'hidden',
        }}
      />
      <div
        style={{
          marginTop: '8px',
          display: 'flex',
          flexWrap: 'wrap',
          gap: '10px',
          alignItems: 'center',
          justifyContent: 'space-between',
          fontSize: '14px',
          color: '#666',
        }}
      >
        <span>Latitude: {latitude ? latitude.toFixed(6) : 'Not set'}</span>
        <span>Longitude: {longitude ? longitude.toFixed(6) : 'Not set'}</span>
        <button
          type="button"
          onClick={handleSnapToLocation}
          className="bg-slate-800 text-white px-3 py-1 rounded-md text-sm"
        >
          {locating ? 'Locating...' : 'Use current location'}
        </button>
      </div>
      {locationError && (
        <p style={{ marginTop: '6px', color: '#c53030', fontSize: '12px' }}>
          {locationError}
        </p>
      )}
      <p
        style={{
          marginTop: '4px',
          fontSize: '12px',
          color: '#999',
          fontStyle: 'italic',
        }}
      >
        Click on the map or drag the marker to set the treasure chest location.
      </p>
    </div>
  );
};
