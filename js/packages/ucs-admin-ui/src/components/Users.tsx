import { useAPI } from '@poltergeist/contexts';
import { User, PointOfInterestDiscovery, PointOfInterestChallengeSubmission, ActivityFeed, PointOfInterest } from '@poltergeist/types';
import React, { useState, useEffect } from 'react';

export const Users = () => {
  const { apiClient } = useAPI();
  const [users, setUsers] = useState<User[]>([]);
  const [filteredUsers, setFilteredUsers] = useState<User[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [discoveries, setDiscoveries] = useState<PointOfInterestDiscovery[]>([]);
  const [submissions, setSubmissions] = useState<PointOfInterestChallengeSubmission[]>([]);
  const [activities, setActivities] = useState<ActivityFeed[]>([]);
  const [selectedDiscoveries, setSelectedDiscoveries] = useState<Set<string>>(new Set());
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [userToDelete, setUserToDelete] = useState<User | null>(null);
  const [showAddDiscoveryModal, setShowAddDiscoveryModal] = useState(false);
  const [availablePOIs, setAvailablePOIs] = useState<PointOfInterest[]>([]);
  const [selectedPOIsToAdd, setSelectedPOIsToAdd] = useState<Set<string>>(new Set());
  const [selectedUsers, setSelectedUsers] = useState<Set<string>>(new Set());
  const [showBulkDeleteConfirm, setShowBulkDeleteConfirm] = useState(false);
  const [editingGold, setEditingGold] = useState(false);
  const [goldInputValue, setGoldInputValue] = useState<string>('');

  useEffect(() => {
    fetchUsers();
    fetchPOIs();
  }, []);

  useEffect(() => {
    if (searchQuery === '') {
      setFilteredUsers(users);
    } else {
      const filtered = users.filter(user =>
        user.username?.toLowerCase().includes(searchQuery.toLowerCase())
      );
      setFilteredUsers(filtered);
    }
  }, [searchQuery, users]);

  const fetchUsers = async () => {
    try {
      const response = await apiClient.get<User[]>('/sonar/users');
      setUsers(response);
      setFilteredUsers(response);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching users:', error);
      setLoading(false);
    }
  };

  const fetchPOIs = async () => {
    try {
      const response = await apiClient.get<PointOfInterest[]>('/sonar/pointsOfInterest');
      setAvailablePOIs(response);
    } catch (error) {
      console.error('Error fetching POIs:', error);
    }
  };

  const selectUser = async (user: User) => {
    setSelectedUser(user);
    setSelectedDiscoveries(new Set());
    setEditingGold(false);
    setGoldInputValue('');
    
    try {
      const [discoveriesRes, submissionsRes, activitiesRes] = await Promise.all([
        apiClient.get<PointOfInterestDiscovery[]>(`/sonar/users/${user.id}/discoveries`),
        apiClient.get<PointOfInterestChallengeSubmission[]>(`/sonar/users/${user.id}/submissions`),
        apiClient.get<ActivityFeed[]>(`/sonar/users/${user.id}/activities`)
      ]);
      
      setDiscoveries(discoveriesRes);
      setSubmissions(submissionsRes);
      setActivities(activitiesRes);
    } catch (error) {
      console.error('Error fetching user details:', error);
    }
  };

  const updateUserGold = async () => {
    if (!selectedUser) return;
    
    const goldAmount = parseInt(goldInputValue);
    if (isNaN(goldAmount) || goldAmount < 0) {
      alert('Please enter a valid gold amount (>= 0)');
      return;
    }
    
    try {
      const updatedUser = await apiClient.patch<User>(`/sonar/users/${selectedUser.id}/gold`, {
        gold: goldAmount
      });
      
      // Update the user in the users list
      setUsers(users.map(u => u.id === selectedUser.id ? updatedUser : u));
      setSelectedUser(updatedUser);
      setEditingGold(false);
      setGoldInputValue('');
    } catch (error) {
      console.error('Error updating user gold:', error);
      alert('Failed to update gold amount');
    }
  };

  const handleDeleteUser = async () => {
    if (!userToDelete) return;
    
    try {
      await apiClient.delete(`/sonar/users/${userToDelete.id}`);
      setUsers(users.filter(u => u.id !== userToDelete.id));
      if (selectedUser?.id === userToDelete.id) {
        setSelectedUser(null);
      }
      setShowDeleteConfirm(false);
      setUserToDelete(null);
    } catch (error) {
      console.error('Error deleting user:', error);
    }
  };

  const toggleUserSelection = (userId: string) => {
    const newSelection = new Set(selectedUsers);
    if (newSelection.has(userId)) {
      newSelection.delete(userId);
    } else {
      newSelection.add(userId);
    }
    setSelectedUsers(newSelection);
  };

  const selectAllUsers = () => {
    setSelectedUsers(new Set(filteredUsers.map(u => u.id)));
  };

  const clearUserSelection = () => {
    setSelectedUsers(new Set());
  };

  const getUsersWithoutUsernames = () => {
    return filteredUsers.filter(user => !user.username || user.username.trim() === '');
  };

  const handleDeleteUsersWithoutUsernames = async () => {
    const usersWithoutUsernames = getUsersWithoutUsernames();
    if (usersWithoutUsernames.length === 0) return;
    
    try {
      await apiClient.delete('/sonar/users', {
        userIds: usersWithoutUsernames.map(u => u.id)
      });
      
      setUsers(users.filter(u => usersWithoutUsernames.every(wu => wu.id !== u.id)));
      if (selectedUser && usersWithoutUsernames.some(u => u.id === selectedUser.id)) {
        setSelectedUser(null);
      }
      setSelectedUsers(new Set());
    } catch (error) {
      console.error('Error deleting users without usernames:', error);
    }
  };

  const handleBulkDeleteUsers = async () => {
    if (selectedUsers.size === 0) return;
    
    try {
      await apiClient.delete('/sonar/users', {
        userIds: Array.from(selectedUsers)
      });
      
      setUsers(users.filter(u => !selectedUsers.has(u.id)));
      if (selectedUser && selectedUsers.has(selectedUser.id)) {
        setSelectedUser(null);
      }
      setShowBulkDeleteConfirm(false);
      setSelectedUsers(new Set());
    } catch (error) {
      console.error('Error deleting users:', error);
    }
  };

  const toggleDiscoverySelection = (discoveryId: string) => {
    const newSelection = new Set(selectedDiscoveries);
    if (newSelection.has(discoveryId)) {
      newSelection.delete(discoveryId);
    } else {
      newSelection.add(discoveryId);
    }
    setSelectedDiscoveries(newSelection);
  };

  const deleteSelectedDiscoveries = async () => {
    if (!selectedUser || selectedDiscoveries.size === 0) return;
    
    try {
      await Promise.all(
        Array.from(selectedDiscoveries).map(id =>
          apiClient.delete(`/sonar/users/${selectedUser.id}/discoveries/${id}`)
        )
      );
      
      setDiscoveries(discoveries.filter(d => !selectedDiscoveries.has(d.id)));
      setSelectedDiscoveries(new Set());
    } catch (error) {
      console.error('Error deleting discoveries:', error);
    }
  };

  const deleteAllDiscoveries = async () => {
    if (!selectedUser) return;
    
    try {
      await apiClient.delete(`/sonar/users/${selectedUser.id}/discoveries`);
      setDiscoveries([]);
      setSelectedDiscoveries(new Set());
    } catch (error) {
      console.error('Error deleting all discoveries:', error);
    }
  };

  const togglePOISelection = (poiId: string) => {
    const newSelection = new Set(selectedPOIsToAdd);
    if (newSelection.has(poiId)) {
      newSelection.delete(poiId);
    } else {
      newSelection.add(poiId);
    }
    setSelectedPOIsToAdd(newSelection);
  };

  const addSelectedDiscoveries = async () => {
    if (!selectedUser || selectedPOIsToAdd.size === 0) return;
    
    try {
      await apiClient.post(`/sonar/users/${selectedUser.id}/discoveries`, {
        pointOfInterestIds: Array.from(selectedPOIsToAdd)
      });
      
      // Refresh discoveries
      const discoveriesRes = await apiClient.get<PointOfInterestDiscovery[]>(`/sonar/users/${selectedUser.id}/discoveries`);
      setDiscoveries(discoveriesRes);
      setSelectedPOIsToAdd(new Set());
      setShowAddDiscoveryModal(false);
    } catch (error) {
      console.error('Error adding discoveries:', error);
    }
  };

  const deleteSubmission = async (submissionId: string) => {
    try {
      await apiClient.delete(`/sonar/submissions/${submissionId}`);
      setSubmissions(submissions.filter(s => s.id !== submissionId));
    } catch (error) {
      console.error('Error deleting submission:', error);
    }
  };

  const deleteAllSubmissions = async () => {
    if (!selectedUser) return;
    
    try {
      await apiClient.delete(`/sonar/users/${selectedUser.id}/submissions`);
      setSubmissions([]);
    } catch (error) {
      console.error('Error deleting all submissions:', error);
    }
  };

  const deleteActivity = async (activityId: string) => {
    try {
      await apiClient.delete(`/sonar/activities/${activityId}`);
      setActivities(activities.filter(a => a.id !== activityId));
    } catch (error) {
      console.error('Error deleting activity:', error);
    }
  };

  const deleteAllActivities = async () => {
    if (!selectedUser) return;
    
    try {
      await apiClient.delete(`/sonar/users/${selectedUser.id}/activities`);
      setActivities([]);
    } catch (error) {
      console.error('Error deleting all activities:', error);
    }
  };

  if (loading) {
    return <div className="p-4">Loading...</div>;
  }

  return (
    <div className="p-4">
      <h1 className="text-3xl font-bold mb-6">User Management</h1>
      
      {/* Search Bar */}
      <div className="mb-6">
        <div className="flex gap-4 items-center">
          <input
            type="text"
            placeholder="Search by username..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="flex-1 px-4 py-2 border rounded-lg"
          />
          {selectedUsers.size > 0 && (
            <div className="flex gap-2">
              <span className="text-sm text-gray-600">{selectedUsers.size} selected</span>
              <button
                onClick={() => setShowBulkDeleteConfirm(true)}
                className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600"
              >
                Delete Selected
              </button>
              <button
                onClick={clearUserSelection}
                className="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600"
              >
                Clear
              </button>
            </div>
          )}
          {filteredUsers.length > 0 && selectedUsers.size === 0 && (
            <button
              onClick={selectAllUsers}
              className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
            >
              Select All
            </button>
          )}
          {getUsersWithoutUsernames().length > 0 && (
            <button
              onClick={handleDeleteUsersWithoutUsernames}
              className="bg-orange-500 text-white px-4 py-2 rounded hover:bg-orange-600"
            >
              Delete Users Without Usernames ({getUsersWithoutUsernames().length})
            </button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-6">
        {/* Users List */}
        <div className="bg-white rounded-lg shadow">
          <div className="p-4 border-b">
            <h2 className="text-xl font-semibold">Users ({filteredUsers.length})</h2>
          </div>
          <div className="overflow-y-auto max-h-[calc(100vh-200px)]">
            {filteredUsers.map(user => (
              <div
                key={user.id}
                className={`p-4 border-b cursor-pointer hover:bg-gray-50 ${
                  selectedUser?.id === user.id ? 'bg-blue-50' : ''
                } ${selectedUsers.has(user.id) ? 'bg-yellow-50' : ''}`}
              >
                <div className="flex items-start gap-3">
                  <input
                    type="checkbox"
                    checked={selectedUsers.has(user.id)}
                    onChange={() => toggleUserSelection(user.id)}
                    onClick={(e) => e.stopPropagation()}
                    className="mt-1"
                  />
                  <div 
                    className="flex-grow cursor-pointer"
                    onClick={() => selectUser(user)}
                  >
                    <div className="flex items-center gap-2">
                      <div className="font-semibold">{user.username || 'No username'}</div>
                      <div className="bg-amber-100 border border-amber-400 rounded px-2 py-0.5">
                        <span className="text-xs font-bold text-amber-600">🪙 {user.gold}</span>
                      </div>
                    </div>
                    <div className="text-sm text-gray-600">{user.phoneNumber}</div>
                    <div className="text-xs text-gray-500">
                      Created: {new Date(user.createdAt).toLocaleDateString()}
                    </div>
                  </div>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      setUserToDelete(user);
                      setShowDeleteConfirm(true);
                    }}
                    className="bg-red-500 text-white px-3 py-1 rounded text-sm hover:bg-red-600"
                  >
                    Delete
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* User Details Panel */}
        <div className="bg-white rounded-lg shadow">
          {selectedUser ? (
            <div>
              <div className="p-4 border-b">
                <h2 className="text-xl font-semibold">{selectedUser.username || 'User Details'}</h2>
                <p className="text-sm text-gray-600">ID: {selectedUser.id}</p>
              </div>

              <div className="p-4 space-y-6 overflow-y-auto max-h-[calc(100vh-200px)]">
                {/* Gold Section */}
                <div>
                  <div className="flex justify-between items-center mb-3">
                    <h3 className="text-lg font-semibold">Gold</h3>
                  </div>
                  {!editingGold ? (
                    <div className="flex items-center gap-4">
                      <div className="bg-amber-100 border border-amber-400 rounded-lg p-4 flex items-center gap-3">
                        <div className="w-12 h-12 rounded-lg border border-amber-400 flex items-center justify-center">
                          <span className="text-lg font-bold text-amber-600">GOLD</span>
                        </div>
                        <span className="text-3xl font-bold text-gray-900">{selectedUser.gold}</span>
                      </div>
                      <button
                        onClick={() => {
                          setEditingGold(true);
                          setGoldInputValue(selectedUser.gold.toString());
                        }}
                        className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
                      >
                        Edit
                      </button>
                    </div>
                  ) : (
                    <div className="flex items-center gap-2">
                      <input
                        type="number"
                        min="0"
                        value={goldInputValue}
                        onChange={(e) => setGoldInputValue(e.target.value)}
                        className="px-4 py-2 border rounded-lg text-lg"
                        placeholder="Enter gold amount"
                      />
                      <button
                        onClick={updateUserGold}
                        className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"
                      >
                        Save
                      </button>
                      <button
                        onClick={() => {
                          setEditingGold(false);
                          setGoldInputValue('');
                        }}
                        className="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600"
                      >
                        Cancel
                      </button>
                    </div>
                  )}
                </div>

                {/* Discoveries Section */}
                <div>
                  <div className="flex justify-between items-center mb-3">
                    <h3 className="text-lg font-semibold">Discoveries ({discoveries.length})</h3>
                    <div className="space-x-2">
                      <button
                        onClick={() => setShowAddDiscoveryModal(true)}
                        className="bg-green-500 text-white px-3 py-1 rounded text-sm hover:bg-green-600"
                      >
                        Add
                      </button>
                      {selectedDiscoveries.size > 0 && (
                        <button
                          onClick={deleteSelectedDiscoveries}
                          className="bg-orange-500 text-white px-3 py-1 rounded text-sm hover:bg-orange-600"
                        >
                          Delete Selected ({selectedDiscoveries.size})
                        </button>
                      )}
                      {discoveries.length > 0 && (
                        <button
                          onClick={deleteAllDiscoveries}
                          className="bg-red-500 text-white px-3 py-1 rounded text-sm hover:bg-red-600"
                        >
                          Delete All
                        </button>
                      )}
                    </div>
                  </div>
                  <div className="space-y-2 max-h-64 overflow-y-auto">
                    {discoveries.map(discovery => (
                      <div key={discovery.id} className="flex items-center p-2 border rounded">
                        <input
                          type="checkbox"
                          checked={selectedDiscoveries.has(discovery.id)}
                          onChange={() => toggleDiscoverySelection(discovery.id)}
                          className="mr-3"
                        />
                        <div className="flex-grow">
                          <div className="text-sm font-medium">{discovery.pointOfInterest?.name || 'Unknown POI'}</div>
                          <div className="text-xs text-gray-500">
                            {new Date(discovery.createdAt).toLocaleDateString()}
                          </div>
                        </div>
                      </div>
                    ))}
                    {discoveries.length === 0 && (
                      <div className="text-gray-500 text-sm text-center py-4">No discoveries</div>
                    )}
                  </div>
                </div>

                {/* Submissions Section */}
                <div>
                  <div className="flex justify-between items-center mb-3">
                    <h3 className="text-lg font-semibold">Submissions ({submissions.length})</h3>
                    {submissions.length > 0 && (
                      <button
                        onClick={deleteAllSubmissions}
                        className="bg-red-500 text-white px-3 py-1 rounded text-sm hover:bg-red-600"
                      >
                        Delete All
                      </button>
                    )}
                  </div>
                  <div className="space-y-2 max-h-64 overflow-y-auto">
                    {submissions.map(submission => (
                      <div key={submission.id} className="flex justify-between items-start p-2 border rounded">
                        <div className="flex-grow">
                          <div className="text-sm">{submission.text}</div>
                          <div className="text-xs text-gray-500">
                            {new Date(submission.createdAt).toLocaleDateString()}
                          </div>
                        </div>
                        <button
                          onClick={() => deleteSubmission(submission.id)}
                          className="bg-red-500 text-white px-2 py-1 rounded text-xs hover:bg-red-600 ml-2"
                        >
                          Delete
                        </button>
                      </div>
                    ))}
                    {submissions.length === 0 && (
                      <div className="text-gray-500 text-sm text-center py-4">No submissions</div>
                    )}
                  </div>
                </div>

                {/* Activities Section */}
                <div>
                  <div className="flex justify-between items-center mb-3">
                    <h3 className="text-lg font-semibold">Activities ({activities.length})</h3>
                    {activities.length > 0 && (
                      <button
                        onClick={deleteAllActivities}
                        className="bg-red-500 text-white px-3 py-1 rounded text-sm hover:bg-red-600"
                      >
                        Delete All
                      </button>
                    )}
                  </div>
                  <div className="space-y-2 max-h-64 overflow-y-auto">
                    {activities.map(activity => (
                      <div key={activity.id} className="flex justify-between items-start p-2 border rounded">
                        <div className="flex-grow">
                          <div className="text-sm">{activity.activityType}</div>
                          <div className="text-xs text-gray-500">
                            {new Date(activity.createdAt).toLocaleDateString()}
                          </div>
                        </div>
                        <button
                          onClick={() => deleteActivity(activity.id)}
                          className="bg-red-500 text-white px-2 py-1 rounded text-xs hover:bg-red-600 ml-2"
                        >
                          Delete
                        </button>
                      </div>
                    ))}
                    {activities.length === 0 && (
                      <div className="text-gray-500 text-sm text-center py-4">No activities</div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          ) : (
            <div className="p-4 text-center text-gray-500">
              Select a user to view details
            </div>
          )}
        </div>
      </div>

      {/* Delete User Confirmation Modal */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg max-w-md">
            <h3 className="text-xl font-bold mb-4">Confirm Delete</h3>
            <p className="mb-6">
              Are you sure you want to delete user <strong>{userToDelete?.username}</strong>? This action cannot be undone.
            </p>
            <div className="flex gap-3 justify-end">
              <button
                onClick={() => {
                  setShowDeleteConfirm(false);
                  setUserToDelete(null);
                }}
                className="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600"
              >
                Cancel
              </button>
              <button
                onClick={handleDeleteUser}
                className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add Discovery Modal */}
      {showAddDiscoveryModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg max-w-2xl w-full max-h-[80vh] overflow-y-auto">
            <h3 className="text-xl font-bold mb-4">Add Discoveries</h3>
            <div className="mb-4">
              <div className="text-sm text-gray-600 mb-3">
                Select points of interest to add as discoveries ({selectedPOIsToAdd.size} selected)
              </div>
              <div className="space-y-2 max-h-96 overflow-y-auto">
                {availablePOIs.map(poi => {
                  const alreadyDiscovered = discoveries.some(d => d.pointOfInterestId === poi.id);
                  return (
                    <div
                      key={poi.id}
                      className={`flex items-center p-3 border rounded ${
                        alreadyDiscovered ? 'bg-gray-100 opacity-50' : 'hover:bg-gray-50'
                      }`}
                    >
                      <input
                        type="checkbox"
                        disabled={alreadyDiscovered}
                        checked={selectedPOIsToAdd.has(poi.id)}
                        onChange={() => togglePOISelection(poi.id)}
                        className="mr-3"
                      />
                      <div className="flex-grow">
                        <div className="font-medium">{poi.name}</div>
                        <div className="text-sm text-gray-600">{poi.description}</div>
                        {alreadyDiscovered && (
                          <div className="text-xs text-green-600">Already discovered</div>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
            <div className="flex gap-3 justify-end">
              <button
                onClick={() => {
                  setShowAddDiscoveryModal(false);
                  setSelectedPOIsToAdd(new Set());
                }}
                className="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600"
              >
                Cancel
              </button>
              <button
                onClick={addSelectedDiscoveries}
                disabled={selectedPOIsToAdd.size === 0}
                className={`px-4 py-2 rounded ${
                  selectedPOIsToAdd.size > 0
                    ? 'bg-green-500 text-white hover:bg-green-600'
                    : 'bg-gray-300 text-gray-500 cursor-not-allowed'
                }`}
              >
                Add Selected ({selectedPOIsToAdd.size})
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Bulk Delete Confirmation Modal */}
      {showBulkDeleteConfirm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg max-w-md">
            <h3 className="text-xl font-bold mb-4">Confirm Bulk Delete</h3>
            <p className="mb-6">
              Are you sure you want to delete <strong>{selectedUsers.size} users</strong>? This action cannot be undone and will delete all their data including discoveries, submissions, activities, and relationships.
            </p>
            <div className="flex gap-3 justify-end">
              <button
                onClick={() => {
                  setShowBulkDeleteConfirm(false);
                }}
                className="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600"
              >
                Cancel
              </button>
              <button
                onClick={handleBulkDeleteUsers}
                className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600"
              >
                Delete {selectedUsers.size} Users
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

