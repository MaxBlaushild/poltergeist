import { useAPI } from '@poltergeist/contexts';
import { FeteRoomTeam, FeteRoom, FeteTeam } from '@poltergeist/types';
import React, { useState, useEffect } from 'react';

export const FeteRoomTeams = () => {
  const { apiClient } = useAPI();
  const [items, setItems] = useState<FeteRoomTeam[]>([]);
  const [rooms, setRooms] = useState<FeteRoom[]>([]);
  const [teams, setTeams] = useState<FeteTeam[]>([]);
  const [filteredItems, setFilteredItems] = useState<FeteRoomTeam[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [showCreateItem, setShowCreateItem] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [itemToDelete, setItemToDelete] = useState<FeteRoomTeam | null>(null);
  const [roomsLoading, setRoomsLoading] = useState(false);

  const [formData, setFormData] = useState({
    feteRoomId: '',
    teamId: '',
  });

  useEffect(() => {
    fetchItems();
    fetchRooms();
    fetchTeams();
  }, []);

  useEffect(() => {
    // Refetch rooms when modal opens to ensure they're up to date
    if (showCreateItem) {
      fetchRooms();
      fetchTeams();
    }
  }, [showCreateItem]);

  useEffect(() => {
    if (searchQuery === '') {
      setFilteredItems(items);
    } else {
      const filtered = items.filter(item => {
        const room = rooms.find(r => r.id === item.feteRoomId);
        const team = teams.find(t => t.id === item.teamId);
        return (
          room?.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
          team?.name?.toLowerCase().includes(searchQuery.toLowerCase())
        );
      });
      setFilteredItems(filtered);
    }
  }, [searchQuery, items, rooms, teams]);

  const fetchItems = async () => {
    try {
      const response = await apiClient.get<FeteRoomTeam[]>('/final-fete/room-teams');
      setItems(response);
      setFilteredItems(response);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching fete room teams:', error);
      setLoading(false);
    }
  };

  const fetchRooms = async () => {
    try {
      setRoomsLoading(true);
      const response = await apiClient.get<FeteRoom[]>('/final-fete/rooms');
      const roomsArray = Array.isArray(response) ? response : [];
      setRooms(roomsArray);
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
      teamId: '',
    });
  };

  const handleCreateItem = async () => {
    try {
      const submitData = {
        feteRoomId: formData.feteRoomId,
        teamId: formData.teamId,
      };

      const newItem = await apiClient.post<FeteRoomTeam>('/final-fete/room-teams', submitData);
      setItems([...items, newItem]);
      setShowCreateItem(false);
      resetForm();
    } catch (error) {
      console.error('Error creating fete room team:', error);
      alert('Error creating fete room team. Please check all required fields.');
    }
  };

  const handleDeleteItem = async (item: FeteRoomTeam) => {
    setItemToDelete(item);
    setShowDeleteConfirm(true);
  };

  const confirmDelete = async () => {
    if (!itemToDelete) return;
    
    try {
      await apiClient.delete(`/final-fete/room-teams/${itemToDelete.id}`);
      setItems(items.filter(i => i.id !== itemToDelete.id));
      setShowDeleteConfirm(false);
      setItemToDelete(null);
    } catch (error) {
      console.error('Error deleting fete room team:', error);
      alert('Error deleting fete room team.');
    }
  };

  if (loading) {
    return <div className="m-10">Loading fete room teams...</div>;
  }

  return (
    <div className="m-10">
      <h1 className="text-2xl font-bold mb-4">Fete Room Teams</h1>
      
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
          Create Room Team
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredItems.map((item) => {
          const room = rooms.find(r => r.id === item.feteRoomId);
          const team = teams.find(t => t.id === item.teamId);
          return (
            <div key={item.id} className="p-4 border rounded-lg bg-white shadow">
              <h2 className="text-lg font-semibold mb-2">Room Team</h2>
              <p className="text-sm text-gray-600">Room: {room?.name || item.feteRoomId}</p>
              <p className="text-sm text-gray-600">Team: {team?.name || item.teamId}</p>
              <div className="mt-4 flex gap-2">
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

      {showCreateItem && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg w-96 max-h-[80vh] overflow-auto">
            <h2 className="text-xl font-bold mb-4">Create Room Team</h2>
            
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
              <label className="block mb-2">Team ID *</label>
              <select
                value={formData.teamId}
                onChange={(e) => setFormData({ ...formData, teamId: e.target.value })}
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
                onClick={handleCreateItem}
                className="bg-blue-500 text-white px-4 py-2 rounded-md"
              >
                Create
              </button>
              <button
                onClick={() => {
                  setShowCreateItem(false);
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
            <p className="mb-4">Are you sure you want to delete this room team? This action cannot be undone.</p>
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

