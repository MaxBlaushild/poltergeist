import { useAPI } from '@poltergeist/contexts';
import { FeteRoom, FeteTeam, HueLight } from '@poltergeist/types';
import React, { useState, useEffect } from 'react';

export const FeteRooms = () => {
  const { apiClient } = useAPI();
  const [rooms, setRooms] = useState<FeteRoom[]>([]);
  const [teams, setTeams] = useState<FeteTeam[]>([]);
  const [hueLights, setHueLights] = useState<HueLight[]>([]);
  const [filteredRooms, setFilteredRooms] = useState<FeteRoom[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [showCreateRoom, setShowCreateRoom] = useState(false);
  const [editingRoom, setEditingRoom] = useState<FeteRoom | null>(null);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [roomToDelete, setRoomToDelete] = useState<FeteRoom | null>(null);

  const [formData, setFormData] = useState({
    name: '',
    open: false,
    currentTeamId: '',
    hueLightId: '',
  });

  useEffect(() => {
    fetchRooms();
    fetchTeams();
    fetchHueLights();
  }, []);

  useEffect(() => {
    if (searchQuery === '') {
      setFilteredRooms(rooms);
    } else {
      const filtered = rooms.filter(room =>
        room.name?.toLowerCase().includes(searchQuery.toLowerCase())
      );
      setFilteredRooms(filtered);
    }
  }, [searchQuery, rooms]);

  const fetchRooms = async () => {
    try {
      const response = await apiClient.get<FeteRoom[]>('/final-fete/rooms');
      setRooms(response);
      setFilteredRooms(response);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching fete rooms:', error);
      setLoading(false);
    }
  };

  const fetchTeams = async () => {
    try {
      const response = await apiClient.get<FeteTeam[]>('/final-fete/teams');
      setTeams(response);
    } catch (error) {
      console.error('Error fetching fete teams:', error);
    }
  };

  const fetchHueLights = async () => {
    try {
      const response = await apiClient.get<HueLight[]>('/final-fete/hue-lights');
      setHueLights(response);
    } catch (error) {
      console.error('Error fetching hue lights:', error);
      // Don't show error to user, just log it - hue lights are optional
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      open: false,
      currentTeamId: '',
      hueLightId: '',
    });
  };

  const handleCreateRoom = async () => {
    try {
      const submitData: any = {
        name: formData.name,
        open: formData.open,
        currentTeamId: formData.currentTeamId,
      };

      if (formData.hueLightId) {
        submitData.hueLightId = parseInt(formData.hueLightId, 10);
      }

      const newRoom = await apiClient.post<FeteRoom>('/final-fete/rooms', submitData);
      setRooms([...rooms, newRoom]);
      setShowCreateRoom(false);
      resetForm();
    } catch (error) {
      console.error('Error creating fete room:', error);
      alert('Error creating fete room. Please check all required fields.');
    }
  };

  const handleUpdateRoom = async () => {
    if (!editingRoom) return;
    
    try {
      const submitData: any = {};
      if (formData.name) submitData.name = formData.name;
      if (formData.currentTeamId) submitData.currentTeamId = formData.currentTeamId;
      submitData.open = formData.open;
      
      if (formData.hueLightId) {
        submitData.hueLightId = parseInt(formData.hueLightId, 10);
      } else {
        submitData.hueLightId = null;
      }

      const updatedRoom = await apiClient.put<FeteRoom>(`/final-fete/rooms/${editingRoom.id}`, submitData);
      setRooms(rooms.map(r => r.id === editingRoom.id ? updatedRoom : r));
      setEditingRoom(null);
      resetForm();
    } catch (error) {
      console.error('Error updating fete room:', error);
      alert('Error updating fete room.');
    }
  };

  const handleDeleteRoom = async (room: FeteRoom) => {
    setRoomToDelete(room);
    setShowDeleteConfirm(true);
  };

  const confirmDelete = async () => {
    if (!roomToDelete) return;
    
    try {
      await apiClient.delete(`/final-fete/rooms/${roomToDelete.id}`);
      setRooms(rooms.filter(r => r.id !== roomToDelete.id));
      setShowDeleteConfirm(false);
      setRoomToDelete(null);
    } catch (error) {
      console.error('Error deleting fete room:', error);
      alert('Error deleting fete room.');
    }
  };

  const handleEditRoom = (room: FeteRoom) => {
    setEditingRoom(room);
    setFormData({
      name: room.name,
      open: room.open,
      currentTeamId: room.currentTeamId,
      hueLightId: room.hueLightId?.toString() || '',
    });
  };

  if (loading) {
    return <div className="m-10">Loading fete rooms...</div>;
  }

  return (
    <div className="m-10">
      <h1 className="text-2xl font-bold mb-4">Fete Rooms</h1>
      
      <div className="mb-4">
        <input
          type="text"
          placeholder="Search by name..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full p-2 border rounded-md"
        />
      </div>

      <div className="mb-4">
        <button
          className="bg-blue-500 text-white px-4 py-2 rounded-md"
          onClick={() => setShowCreateRoom(true)}
        >
          Create Fete Room
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredRooms.map((room) => {
          const currentTeam = teams.find(t => t.id === room.currentTeamId);
          const hueLight = room.hueLightId ? hueLights.find(l => l.id === room.hueLightId) : null;
          return (
            <div key={room.id} className="p-4 border rounded-lg bg-white shadow">
              <h2 className="text-lg font-semibold mb-2">{room.name}</h2>
              <p className="text-sm text-gray-600">Open: {room.open ? 'Yes' : 'No'}</p>
              <p className="text-sm text-gray-600">Current Team: {currentTeam?.name || room.currentTeamId}</p>
              {hueLight && (
                <p className="text-sm text-gray-600">Hue Light: {hueLight.name}</p>
              )}
              <div className="mt-4 flex gap-2">
                <button
                  onClick={() => handleEditRoom(room)}
                  className="bg-blue-500 text-white px-4 py-2 rounded-md"
                >
                  Edit
                </button>
                <button
                  onClick={() => handleDeleteRoom(room)}
                  className="bg-red-500 text-white px-4 py-2 rounded-md"
                >
                  Delete
                </button>
              </div>
            </div>
          );
        })}
      </div>

      {(showCreateRoom || editingRoom) && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg w-96 max-h-[80vh] overflow-auto">
            <h2 className="text-xl font-bold mb-4">{editingRoom ? 'Edit Fete Room' : 'Create Fete Room'}</h2>
            
            <div className="mb-4">
              <label className="block mb-2">Name *</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                className="w-full p-2 border rounded-md"
                required
              />
            </div>

            <div className="mb-4">
              <label className="block mb-2">Current Team ID *</label>
              <select
                value={formData.currentTeamId}
                onChange={(e) => setFormData({ ...formData, currentTeamId: e.target.value })}
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
              <label className="block mb-2">Hue Light</label>
              <select
                value={formData.hueLightId}
                onChange={(e) => setFormData({ ...formData, hueLightId: e.target.value })}
                className="w-full p-2 border rounded-md"
              >
                <option value="">None</option>
                {hueLights.map(light => (
                  <option key={light.id} value={light.id.toString()}>
                    {light.name} (ID: {light.id})
                  </option>
                ))}
              </select>
            </div>

            <div className="mb-4">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={formData.open}
                  onChange={(e) => setFormData({ ...formData, open: e.target.checked })}
                  className="mr-2"
                />
                Open
              </label>
            </div>

            <div className="flex gap-2">
              <button
                onClick={() => {
                  if (editingRoom) {
                    handleUpdateRoom();
                  } else {
                    handleCreateRoom();
                  }
                }}
                className="bg-blue-500 text-white px-4 py-2 rounded-md"
              >
                {editingRoom ? 'Update' : 'Create'}
              </button>
              <button
                onClick={() => {
                  setShowCreateRoom(false);
                  setEditingRoom(null);
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

      {showDeleteConfirm && roomToDelete && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg w-96">
            <h2 className="text-xl font-bold mb-4">Confirm Delete</h2>
            <p className="mb-4">Are you sure you want to delete "{roomToDelete.name}"? This action cannot be undone.</p>
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
                  setRoomToDelete(null);
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

