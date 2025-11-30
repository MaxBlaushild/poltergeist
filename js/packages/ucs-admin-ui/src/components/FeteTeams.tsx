import { useAPI } from '@poltergeist/contexts';
import { FeteTeam } from '@poltergeist/types';
import React, { useState, useEffect } from 'react';

export const FeteTeams = () => {
  const { apiClient } = useAPI();
  const [teams, setTeams] = useState<FeteTeam[]>([]);
  const [filteredTeams, setFilteredTeams] = useState<FeteTeam[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [showCreateTeam, setShowCreateTeam] = useState(false);
  const [editingTeam, setEditingTeam] = useState<FeteTeam | null>(null);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [teamToDelete, setTeamToDelete] = useState<FeteTeam | null>(null);

  const [formData, setFormData] = useState({
    name: '',
  });

  useEffect(() => {
    fetchTeams();
  }, []);

  useEffect(() => {
    if (searchQuery === '') {
      setFilteredTeams(teams);
    } else {
      const filtered = teams.filter(team =>
        team.name?.toLowerCase().includes(searchQuery.toLowerCase())
      );
      setFilteredTeams(filtered);
    }
  }, [searchQuery, teams]);

  const fetchTeams = async () => {
    try {
      const response = await apiClient.get<FeteTeam[]>('/final-fete/teams');
      setTeams(response);
      setFilteredTeams(response);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching fete teams:', error);
      setLoading(false);
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
    });
  };

  const handleCreateTeam = async () => {
    try {
      const submitData = {
        name: formData.name,
      };

      const newTeam = await apiClient.post<FeteTeam>('/final-fete/teams', submitData);
      setTeams([...teams, newTeam]);
      setShowCreateTeam(false);
      resetForm();
    } catch (error) {
      console.error('Error creating fete team:', error);
      alert('Error creating fete team. Please check all required fields.');
    }
  };

  const handleUpdateTeam = async () => {
    if (!editingTeam) return;
    
    try {
      const submitData: any = {};
      if (formData.name) submitData.name = formData.name;

      const updatedTeam = await apiClient.put<FeteTeam>(`/final-fete/teams/${editingTeam.id}`, submitData);
      setTeams(teams.map(t => t.id === editingTeam.id ? updatedTeam : t));
      setEditingTeam(null);
      resetForm();
    } catch (error) {
      console.error('Error updating fete team:', error);
      alert('Error updating fete team.');
    }
  };

  const handleDeleteTeam = async (team: FeteTeam) => {
    setTeamToDelete(team);
    setShowDeleteConfirm(true);
  };

  const confirmDelete = async () => {
    if (!teamToDelete) return;
    
    try {
      await apiClient.delete(`/final-fete/teams/${teamToDelete.id}`);
      setTeams(teams.filter(t => t.id !== teamToDelete.id));
      setShowDeleteConfirm(false);
      setTeamToDelete(null);
    } catch (error) {
      console.error('Error deleting fete team:', error);
      alert('Error deleting fete team.');
    }
  };

  const handleEditTeam = (team: FeteTeam) => {
    setEditingTeam(team);
    setFormData({
      name: team.name,
    });
  };

  if (loading) {
    return <div className="m-10">Loading fete teams...</div>;
  }

  return (
    <div className="m-10">
      <h1 className="text-2xl font-bold mb-4">Fete Teams</h1>
      
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
          onClick={() => setShowCreateTeam(true)}
        >
          Create Fete Team
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredTeams.map((team) => (
          <div key={team.id} className="p-4 border rounded-lg bg-white shadow">
            <h2 className="text-lg font-semibold mb-2">{team.name}</h2>
            <p className="text-sm text-gray-600">ID: {team.id}</p>
            <div className="mt-4 flex gap-2">
              <button
                onClick={() => handleEditTeam(team)}
                className="bg-blue-500 text-white px-4 py-2 rounded-md"
              >
                Edit
              </button>
              <button
                onClick={() => handleDeleteTeam(team)}
                className="bg-red-500 text-white px-4 py-2 rounded-md"
              >
                Delete
              </button>
            </div>
          </div>
        ))}
      </div>

      {(showCreateTeam || editingTeam) && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg w-96 max-h-[80vh] overflow-auto">
            <h2 className="text-xl font-bold mb-4">{editingTeam ? 'Edit Fete Team' : 'Create Fete Team'}</h2>
            
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

            <div className="flex gap-2">
              <button
                onClick={() => {
                  if (editingTeam) {
                    handleUpdateTeam();
                  } else {
                    handleCreateTeam();
                  }
                }}
                className="bg-blue-500 text-white px-4 py-2 rounded-md"
              >
                {editingTeam ? 'Update' : 'Create'}
              </button>
              <button
                onClick={() => {
                  setShowCreateTeam(false);
                  setEditingTeam(null);
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

      {showDeleteConfirm && teamToDelete && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg w-96">
            <h2 className="text-xl font-bold mb-4">Confirm Delete</h2>
            <p className="mb-4">Are you sure you want to delete "{teamToDelete.name}"? This action cannot be undone.</p>
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
                  setTeamToDelete(null);
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

