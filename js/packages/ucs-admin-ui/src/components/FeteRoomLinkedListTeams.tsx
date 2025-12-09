import { useAPI } from '@poltergeist/contexts';
import { FeteRoomLinkedListTeam, FeteRoom, FeteTeam } from '@poltergeist/types';
import React, { useState, useEffect } from 'react';

export const FeteRoomLinkedListTeams = () => {
  const { apiClient } = useAPI();
  const [items, setItems] = useState<FeteRoomLinkedListTeam[]>([]);
  const [rooms, setRooms] = useState<FeteRoom[]>([]);
  const [teams, setTeams] = useState<FeteTeam[]>([]);
  const [filteredItems, setFilteredItems] = useState<FeteRoomLinkedListTeam[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [showCreateItem, setShowCreateItem] = useState(false);
  const [editingItem, setEditingItem] = useState<FeteRoomLinkedListTeam | null>(null);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [itemToDelete, setItemToDelete] = useState<FeteRoomLinkedListTeam | null>(null);
  const [roomsLoading, setRoomsLoading] = useState(false);

  const [formData, setFormData] = useState({
    feteRoomId: '',
    firstTeamId: '',
    secondTeamId: '',
  });

  useEffect(() => {
    fetchItems();
    fetchRooms();
    fetchTeams();
  }, []);

  useEffect(() => {
    // Refetch rooms when modal opens to ensure they're up to date
    if (showCreateItem || editingItem) {
      fetchRooms();
      fetchTeams();
    }
  }, [showCreateItem, editingItem]);

  useEffect(() => {
    if (searchQuery === '') {
      setFilteredItems(items);
    } else {
      const filtered = items.filter(item => {
        const room = rooms.find(r => r.id === item.feteRoomId);
        const firstTeam = teams.find(t => t.id === item.firstTeamId);
        const secondTeam = teams.find(t => t.id === item.secondTeamId);
        return (
          room?.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
          firstTeam?.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
          secondTeam?.name?.toLowerCase().includes(searchQuery.toLowerCase())
        );
      });
      setFilteredItems(filtered);
    }
  }, [searchQuery, items, rooms, teams]);

  const fetchItems = async () => {
    try {
      const response = await apiClient.get<FeteRoomLinkedListTeam[]>('/final-fete/room-linked-list-teams');
      setItems(response);
      setFilteredItems(response);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching fete room linked list teams:', error);
      setLoading(false);
    }
  };

  const fetchRooms = async () => {
    try {
      setRoomsLoading(true);
      const response = await apiClient.get<FeteRoom[]>('/final-fete/rooms');
      const roomsArray = Array.isArray(response) ? response : [];
      setRooms(roomsArray);
      console.log('Fetched rooms:', roomsArray.length);
    } catch (error) {
      console.error('Error fetching fete rooms:', error);
      setRooms([]);
    } finally {
      setRoomsLoading(false);
    }
  };

  const fetchTeams = async () => {
    try {
      const response = await apiClient.get<FeteTeam[]>('/final-fete/teams');
      const teamsArray = Array.isArray(response) ? response : [];
      setTeams(teamsArray);
    } catch (error) {
      console.error('Error fetching fete teams:', error);
      setTeams([]);
    }
  };

  const resetForm = () => {
    setFormData({
      feteRoomId: '',
      firstTeamId: '',
      secondTeamId: '',
    });
  };

  const handleCreateItem = async () => {
    try {
      const submitData = {
        feteRoomId: formData.feteRoomId,
        firstTeamId: formData.firstTeamId,
        secondTeamId: formData.secondTeamId,
      };

      const newItem = await apiClient.post<FeteRoomLinkedListTeam>('/final-fete/room-linked-list-teams', submitData);
      setItems([...items, newItem]);
      setShowCreateItem(false);
      resetForm();
    } catch (error) {
      console.error('Error creating fete room linked list team:', error);
      alert('Error creating fete room linked list team. Please check all required fields.');
    }
  };

  const handleUpdateItem = async () => {
    if (!editingItem) return;
    
    try {
      const submitData: any = {};
      if (formData.feteRoomId) submitData.feteRoomId = formData.feteRoomId;
      if (formData.firstTeamId) submitData.firstTeamId = formData.firstTeamId;
      if (formData.secondTeamId) submitData.secondTeamId = formData.secondTeamId;

      const updatedItem = await apiClient.put<FeteRoomLinkedListTeam>(`/final-fete/room-linked-list-teams/${editingItem.id}`, submitData);
      setItems(items.map(i => i.id === editingItem.id ? updatedItem : i));
      setEditingItem(null);
      resetForm();
    } catch (error) {
      console.error('Error updating fete room linked list team:', error);
      alert('Error updating fete room linked list team.');
    }
  };

  const handleDeleteItem = async (item: FeteRoomLinkedListTeam) => {
    setItemToDelete(item);
    setShowDeleteConfirm(true);
  };

  const confirmDelete = async () => {
    if (!itemToDelete) return;
    
    try {
      await apiClient.delete(`/final-fete/room-linked-list-teams/${itemToDelete.id}`);
      setItems(items.filter(i => i.id !== itemToDelete.id));
      setShowDeleteConfirm(false);
      setItemToDelete(null);
    } catch (error) {
      console.error('Error deleting fete room linked list team:', error);
      alert('Error deleting fete room linked list team.');
    }
  };

  const handleEditItem = (item: FeteRoomLinkedListTeam) => {
    setEditingItem(item);
    setFormData({
      feteRoomId: item.feteRoomId,
      firstTeamId: item.firstTeamId,
      secondTeamId: item.secondTeamId,
    });
  };

  if (loading) {
    return <div className="m-10">Loading fete room linked list teams...</div>;
  }

  return (
    <div className="m-10">
      <h1 className="text-2xl font-bold mb-4">Fete Room Linked List Teams</h1>
      
      <div className="mb-4">
        <input
          type="text"
          placeholder="Search by room or team name..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full p-2 border rounded-md"
        />
      </div>

      <div className="mb-4">
        <button
          className="bg-blue-500 text-white px-4 py-2 rounded-md"
          onClick={() => setShowCreateItem(true)}
        >
          Create Linked List Item
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredItems.map((item) => {
          const room = rooms.find(r => r.id === item.feteRoomId);
          const firstTeam = teams.find(t => t.id === item.firstTeamId);
          const secondTeam = teams.find(t => t.id === item.secondTeamId);
          return (
            <div key={item.id} className="p-4 border rounded-lg bg-white shadow">
              <h2 className="text-lg font-semibold mb-2">Linked List Item</h2>
              <p className="text-sm text-gray-600">Room: {room?.name || item.feteRoomId}</p>
              <p className="text-sm text-gray-600">First Team: {firstTeam?.name || item.firstTeamId}</p>
              <p className="text-sm text-gray-600">Second Team: {secondTeam?.name || item.secondTeamId}</p>
              <div className="mt-4 flex gap-2">
                <button
                  onClick={() => handleEditItem(item)}
                  className="bg-blue-500 text-white px-4 py-2 rounded-md"
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
          );
        })}
      </div>

      {(showCreateItem || editingItem) && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg w-96 max-h-[80vh] overflow-auto">
            <h2 className="text-xl font-bold mb-4">{editingItem ? 'Edit Linked List Item' : 'Create Linked List Item'}</h2>
            
            <div className="mb-4">
              <label className="block mb-2">Fete Room ID *</label>
              <select
                value={formData.feteRoomId}
                onChange={(e) => setFormData({ ...formData, feteRoomId: e.target.value })}
                className="w-full p-2 border rounded-md"
                required
                disabled={roomsLoading}
              >
                <option value="">{roomsLoading ? 'Loading rooms...' : 'Select a room'}</option>
                {rooms.length === 0 && !roomsLoading && (
                  <option value="" disabled>No rooms available</option>
                )}
                {rooms.map(room => (
                  <option key={room.id} value={room.id}>{room.name}</option>
                ))}
              </select>
            </div>

            <div className="mb-4">
              <label className="block mb-2">First Team ID *</label>
              <select
                value={formData.firstTeamId}
                onChange={(e) => setFormData({ ...formData, firstTeamId: e.target.value })}
                className="w-full p-2 border rounded-md"
                required
              >
                <option value="">Select a team</option>
                {teams.map(team => (
                  <option key={team.id} value={team.id}>{team.name}</option>
                ))}
              </select>
            </div>

            <div className="mb-4">
              <label className="block mb-2">Second Team ID *</label>
              <select
                value={formData.secondTeamId}
                onChange={(e) => setFormData({ ...formData, secondTeamId: e.target.value })}
                className="w-full p-2 border rounded-md"
                required
              >
                <option value="">Select a team</option>
                {teams.map(team => (
                  <option key={team.id} value={team.id}>{team.name}</option>
                ))}
              </select>
            </div>

            <div className="flex gap-2">
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

      {showDeleteConfirm && itemToDelete && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg w-96">
            <h2 className="text-xl font-bold mb-4">Confirm Delete</h2>
            <p className="mb-4">Are you sure you want to delete this linked list item? This action cannot be undone.</p>
            <div className="flex gap-2">
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

