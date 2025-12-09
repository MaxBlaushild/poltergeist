import { useAPI } from '@poltergeist/contexts';
import { FeteTeam, User } from '@poltergeist/types';
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
  const [selectedTeam, setSelectedTeam] = useState<FeteTeam | null>(null);
  const [teamMembers, setTeamMembers] = useState<User[]>([]);
  const [userSearchQuery, setUserSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<User[]>([]);
  const [searchingUsers, setSearchingUsers] = useState(false);
  const [loadingMembers, setLoadingMembers] = useState(false);
  const [showRemoveConfirm, setShowRemoveConfirm] = useState(false);
  const [userToRemove, setUserToRemove] = useState<{ teamId: string; userId: string } | null>(null);

  const [formData, setFormData] = useState({
    name: '',
  });

  useEffect(() => {
    fetchTeams();
  }, []);

  useEffect(() => {
    if (selectedTeam) {
      fetchTeamUsers(selectedTeam.id);
    } else {
      setTeamMembers([]);
    }
  }, [selectedTeam]);

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

  const fetchTeamUsers = async (teamId: string) => {
    try {
      setLoadingMembers(true);
      const response = await apiClient.get<User[]>(`/final-fete/teams/${teamId}/users`);
      setTeamMembers(Array.isArray(response) ? response : []);
    } catch (error) {
      console.error('Error fetching team users:', error);
      setTeamMembers([]);
    } finally {
      setLoadingMembers(false);
    }
  };

  const searchUsers = async (query: string) => {
    if (!query || query.length < 2) {
      setSearchResults([]);
      return;
    }

    try {
      setSearchingUsers(true);
      const response = await apiClient.get<User[]>(`/final-fete/users/search?query=${encodeURIComponent(query)}`);
      setSearchResults(Array.isArray(response) ? response : []);
    } catch (error) {
      console.error('Error searching users:', error);
      setSearchResults([]);
    } finally {
      setSearchingUsers(false);
    }
  };

  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (userSearchQuery) {
        searchUsers(userSearchQuery);
      } else {
        setSearchResults([]);
      }
    }, 300); // Debounce search

    return () => clearTimeout(timeoutId);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [userSearchQuery]);

  const handleAddUserToTeam = async (teamId: string, userId: string) => {
    try {
      await apiClient.post(`/final-fete/teams/${teamId}/users`, { userId });
      // Refresh team members
      if (selectedTeam && selectedTeam.id === teamId) {
        await fetchTeamUsers(teamId);
      }
      // Clear search
      setUserSearchQuery('');
      setSearchResults([]);
      alert('User added to team successfully');
    } catch (error: any) {
      console.error('Error adding user to team:', error);
      const errorMessage = error.response?.data?.error || error.message || 'Failed to add user to team';
      alert(`Error: ${errorMessage}`);
    }
  };

  const handleRemoveUserFromTeam = (teamId: string, userId: string) => {
    setUserToRemove({ teamId, userId });
    setShowRemoveConfirm(true);
  };

  const confirmRemoveUser = async () => {
    if (!userToRemove) return;

    try {
      await apiClient.delete(`/final-fete/teams/${userToRemove.teamId}/users/${userToRemove.userId}`);
      // Refresh team members
      if (selectedTeam && selectedTeam.id === userToRemove.teamId) {
        await fetchTeamUsers(userToRemove.teamId);
      }
      setShowRemoveConfirm(false);
      setUserToRemove(null);
      alert('User removed from team successfully');
    } catch (error: any) {
      console.error('Error removing user from team:', error);
      const errorMessage = error.response?.data?.error || error.message || 'Failed to remove user from team';
      alert(`Error: ${errorMessage}`);
    }
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
            <div className="mt-4 flex gap-2 flex-wrap">
              <button
                onClick={() => setSelectedTeam(selectedTeam?.id === team.id ? null : team)}
                className={`px-4 py-2 rounded-md ${
                  selectedTeam?.id === team.id
                    ? 'bg-green-500 text-white'
                    : 'bg-gray-500 text-white'
                }`}
              >
                {selectedTeam?.id === team.id ? 'Hide Members' : 'Show Members'}
              </button>
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

      {selectedTeam && (
        <div className="mt-8 p-6 border rounded-lg bg-white shadow">
          <h2 className="text-xl font-bold mb-4">Team Members: {selectedTeam.name}</h2>
          
          {loadingMembers ? (
            <div className="text-gray-600">Loading members...</div>
          ) : (
            <>
              <div className="mb-4">
                <h3 className="text-lg font-semibold mb-2">Current Members ({teamMembers.length})</h3>
                {teamMembers.length === 0 ? (
                  <p className="text-gray-600">No members in this team</p>
                ) : (
                  <div className="space-y-2">
                    {teamMembers.map((user) => (
                      <div
                        key={user.id}
                        className="flex items-center justify-between p-3 border rounded-md bg-gray-50"
                      >
                        <div>
                          <p className="font-medium">
                            {user.username || user.name || 'Unnamed User'}
                          </p>
                          <p className="text-sm text-gray-600">
                            {user.phoneNumber} {user.username ? `(@${user.username})` : ''}
                          </p>
                          <p className="text-xs text-gray-500">ID: {user.id}</p>
                        </div>
                        <button
                          onClick={() => handleRemoveUserFromTeam(selectedTeam.id, user.id)}
                          className="bg-red-500 text-white px-3 py-1 rounded-md text-sm"
                        >
                          Remove
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              <div className="mt-6 border-t pt-4">
                <h3 className="text-lg font-semibold mb-2">Add User to Team</h3>
                <div className="relative">
                  <input
                    type="text"
                    placeholder="Search by phone number or username..."
                    value={userSearchQuery}
                    onChange={(e) => setUserSearchQuery(e.target.value)}
                    className="w-full p-2 border rounded-md"
                  />
                  {searchingUsers && (
                    <div className="absolute right-2 top-2 text-gray-500">Searching...</div>
                  )}
                </div>

                {userSearchQuery && searchResults.length > 0 && (
                  <div className="mt-2 border rounded-md bg-white shadow-lg max-h-60 overflow-auto">
                    {searchResults.map((user) => {
                      const isAlreadyMember = teamMembers.some(m => m.id === user.id);
                      return (
                        <div
                          key={user.id}
                          className={`p-3 border-b last:border-b-0 ${
                            isAlreadyMember ? 'bg-gray-100' : 'hover:bg-gray-50'
                          }`}
                        >
                          <div className="flex items-center justify-between">
                            <div>
                              <p className="font-medium">
                                {user.username || user.name || 'Unnamed User'}
                              </p>
                              <p className="text-sm text-gray-600">
                                {user.phoneNumber} {user.username ? `(@${user.username})` : ''}
                              </p>
                            </div>
                            {isAlreadyMember ? (
                              <span className="text-sm text-gray-500">Already a member</span>
                            ) : (
                              <button
                                onClick={() => handleAddUserToTeam(selectedTeam.id, user.id)}
                                className="bg-green-500 text-white px-3 py-1 rounded-md text-sm"
                              >
                                Add
                              </button>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}

                {userSearchQuery && !searchingUsers && searchResults.length === 0 && (
                  <div className="mt-2 text-gray-600 text-sm">No users found</div>
                )}
              </div>
            </>
          )}
        </div>
      )}

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

      {showRemoveConfirm && userToRemove && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg w-96">
            <h2 className="text-xl font-bold mb-4">Confirm Remove User</h2>
            <p className="mb-4">Are you sure you want to remove this user from the team?</p>
            <div className="flex gap-2">
              <button
                onClick={confirmRemoveUser}
                className="bg-red-500 text-white px-4 py-2 rounded-md"
              >
                Remove
              </button>
              <button
                onClick={() => {
                  setShowRemoveConfirm(false);
                  setUserToRemove(null);
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

