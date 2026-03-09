import { useAPI } from '@poltergeist/contexts';
import { User, PointOfInterestDiscovery, PointOfInterestChallengeSubmission, ActivityFeed, PointOfInterest } from '@poltergeist/types';
import React, { useState, useEffect } from 'react';

type AdminCharacterStats = {
  health: number;
  maxHealth: number;
  mana: number;
  maxMana: number;
};

type UserCharacterProfileResponse = {
  stats?: Partial<AdminCharacterStats>;
};

type UserProfilePlaceholderResponse = {
  thumbnailUrl?: string;
  status?: string;
  exists?: boolean;
  requestedAt?: string;
  lastModified?: string;
  appliedUserCount?: number;
};

const defaultUserPlaceholderPrompt =
  'A polished fantasy RPG profile portrait avatar. Head-and-shoulders, centered composition, expressive face, clean background, no text, no logos, game-ready artwork.';

const placeholderStatusClassName = (status?: string) => {
  switch ((status || '').trim()) {
    case 'completed':
      return 'bg-emerald-600';
    case 'queued':
      return 'bg-amber-500';
    case 'in_progress':
      return 'bg-blue-600';
    case 'failed':
      return 'bg-rose-600';
    default:
      return 'bg-gray-500';
  }
};

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
  const [statusName, setStatusName] = useState('');
  const [statusDescription, setStatusDescription] = useState('');
  const [statusEffect, setStatusEffect] = useState('');
  const [statusPositive, setStatusPositive] = useState(true);
  const [statusDurationMinutes, setStatusDurationMinutes] = useState('60');
  const [statusStrengthMod, setStatusStrengthMod] = useState('0');
  const [statusDexterityMod, setStatusDexterityMod] = useState('0');
  const [statusConstitutionMod, setStatusConstitutionMod] = useState('0');
  const [statusIntelligenceMod, setStatusIntelligenceMod] = useState('0');
  const [statusWisdomMod, setStatusWisdomMod] = useState('0');
  const [statusCharismaMod, setStatusCharismaMod] = useState('0');
  const [grantingStatus, setGrantingStatus] = useState(false);
  const [statusGrantMessage, setStatusGrantMessage] = useState<string | null>(null);
  const [statusGrantKind, setStatusGrantKind] = useState<'success' | 'error' | null>(null);
  const [resourceStats, setResourceStats] = useState<AdminCharacterStats | null>(null);
  const [resourceLoading, setResourceLoading] = useState(false);
  const [resourceAmountHealth, setResourceAmountHealth] = useState('0');
  const [resourceAmountMana, setResourceAmountMana] = useState('0');
  const [resourceSubmitting, setResourceSubmitting] = useState(false);
  const [resourceMessage, setResourceMessage] = useState<string | null>(null);
  const [resourceMessageKind, setResourceMessageKind] = useState<'success' | 'error' | null>(null);
  const [profilePlaceholderPrompt, setProfilePlaceholderPrompt] = useState(defaultUserPlaceholderPrompt);
  const [profilePlaceholderStatus, setProfilePlaceholderStatus] = useState('unknown');
  const [profilePlaceholderUrl, setProfilePlaceholderUrl] = useState('');
  const [profilePlaceholderExists, setProfilePlaceholderExists] = useState(false);
  const [profilePlaceholderRequestedAt, setProfilePlaceholderRequestedAt] = useState<string | null>(null);
  const [profilePlaceholderLastModified, setProfilePlaceholderLastModified] = useState<string | null>(null);
  const [profilePlaceholderMessage, setProfilePlaceholderMessage] = useState<string | null>(null);
  const [profilePlaceholderError, setProfilePlaceholderError] = useState<string | null>(null);
  const [profilePlaceholderBusy, setProfilePlaceholderBusy] = useState(false);
  const [profilePlaceholderStatusLoading, setProfilePlaceholderStatusLoading] = useState(false);
  const [profilePlaceholderPreviewNonce, setProfilePlaceholderPreviewNonce] = useState(0);

  const applyUsersResponse = React.useCallback((nextUsers: User[]) => {
    setUsers(nextUsers);
    setFilteredUsers(nextUsers);
    setSelectedUser((prev) =>
      prev ? nextUsers.find((user) => user.id === prev.id) ?? prev : prev
    );
  }, []);

  const refreshProfilePlaceholderStatus = React.useCallback(
    async (showMessage = false) => {
      try {
        setProfilePlaceholderStatusLoading(true);
        setProfilePlaceholderError(null);
        const response = await apiClient.get<UserProfilePlaceholderResponse>(
          '/sonar/admin/users/profile-picture-placeholder/status'
        );
        const url = (response?.thumbnailUrl || '').trim();
        if (url) {
          setProfilePlaceholderUrl(url);
        }
        setProfilePlaceholderStatus((response?.status || 'unknown').trim() || 'unknown');
        setProfilePlaceholderExists(Boolean(response?.exists));
        setProfilePlaceholderRequestedAt(response?.requestedAt || null);
        setProfilePlaceholderLastModified(response?.lastModified || null);
        setProfilePlaceholderPreviewNonce(Date.now());
        if ((response?.appliedUserCount || 0) > 0) {
          const refreshedUsers = await apiClient.get<User[]>('/sonar/users');
          applyUsersResponse(refreshedUsers);
        }
        if (showMessage) {
          setProfilePlaceholderMessage('Profile placeholder status refreshed.');
        } else if ((response?.appliedUserCount || 0) > 0) {
          setProfilePlaceholderMessage(
            `Generated placeholder applied to ${response?.appliedUserCount} user${response?.appliedUserCount === 1 ? '' : 's'} without profile images.`
          );
        }
      } catch (error) {
        console.error('Error loading profile placeholder status:', error);
        setProfilePlaceholderError(
          error instanceof Error
            ? error.message
            : 'Failed to load profile placeholder status.'
        );
      } finally {
        setProfilePlaceholderStatusLoading(false);
      }
    },
    [apiClient, applyUsersResponse]
  );

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

  useEffect(() => {
    void refreshProfilePlaceholderStatus();
  }, [refreshProfilePlaceholderStatus]);

  useEffect(() => {
    if (
      profilePlaceholderStatus !== 'queued' &&
      profilePlaceholderStatus !== 'in_progress'
    ) {
      return;
    }

    const interval = window.setInterval(() => {
      void refreshProfilePlaceholderStatus();
    }, 4000);
    return () => window.clearInterval(interval);
  }, [profilePlaceholderStatus, refreshProfilePlaceholderStatus]);

  const fetchUsers = async () => {
    try {
      const response = await apiClient.get<User[]>('/sonar/users');
      applyUsersResponse(response);
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

  const normalizeResourceStats = (stats?: Partial<AdminCharacterStats> | null): AdminCharacterStats | null => {
    if (!stats) return null;
    const health = Number.isFinite(stats.health as number) ? Number(stats.health) : 0;
    const maxHealth = Number.isFinite(stats.maxHealth as number) ? Number(stats.maxHealth) : health;
    const mana = Number.isFinite(stats.mana as number) ? Number(stats.mana) : 0;
    const maxMana = Number.isFinite(stats.maxMana as number) ? Number(stats.maxMana) : mana;
    return {
      health,
      maxHealth: Math.max(maxHealth, 1),
      mana,
      maxMana: Math.max(maxMana, 1),
    };
  };

  const selectUser = async (user: User) => {
    setSelectedUser(user);
    setSelectedDiscoveries(new Set());
    setEditingGold(false);
    setGoldInputValue('');
    setStatusName('');
    setStatusDescription('');
    setStatusEffect('');
    setStatusPositive(true);
    setStatusDurationMinutes('60');
    setStatusStrengthMod('0');
    setStatusDexterityMod('0');
    setStatusConstitutionMod('0');
    setStatusIntelligenceMod('0');
    setStatusWisdomMod('0');
    setStatusCharismaMod('0');
    setStatusGrantMessage(null);
    setStatusGrantKind(null);
    setResourceAmountHealth('0');
    setResourceAmountMana('0');
    setResourceMessage(null);
    setResourceMessageKind(null);
    setResourceStats(null);
    setResourceLoading(true);
    
    try {
      const [discoveriesRes, submissionsRes, activitiesRes, characterProfileRes] = await Promise.all([
        apiClient.get<PointOfInterestDiscovery[]>(`/sonar/users/${user.id}/discoveries`),
        apiClient.get<PointOfInterestChallengeSubmission[]>(`/sonar/users/${user.id}/submissions`),
        apiClient.get<ActivityFeed[]>(`/sonar/users/${user.id}/activities`),
        apiClient.get<UserCharacterProfileResponse>(`/sonar/users/${user.id}/character`)
      ]);
      
      setDiscoveries(discoveriesRes);
      setSubmissions(submissionsRes);
      setActivities(activitiesRes);
      setResourceStats(normalizeResourceStats(characterProfileRes?.stats));
    } catch (error) {
      console.error('Error fetching user details:', error);
      setResourceStats(null);
    } finally {
      setResourceLoading(false);
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

  const parseModifierValue = (value: string) => {
    const parsed = parseInt(value, 10);
    return Number.isNaN(parsed) ? 0 : parsed;
  };

  const grantStatus = async () => {
    if (!selectedUser) return;

    const trimmedName = statusName.trim();
    if (!trimmedName) {
      setStatusGrantMessage('Status name is required.');
      setStatusGrantKind('error');
      return;
    }

    const durationMinutes = parseInt(statusDurationMinutes, 10);
    if (Number.isNaN(durationMinutes) || durationMinutes <= 0) {
      setStatusGrantMessage('Duration must be a positive number of minutes.');
      setStatusGrantKind('error');
      return;
    }

    try {
      setGrantingStatus(true);
      setStatusGrantMessage(null);
      setStatusGrantKind(null);

      await apiClient.post(`/sonar/admin/users/${selectedUser.id}/statuses`, {
        name: trimmedName,
        description: statusDescription.trim(),
        effect: statusEffect.trim(),
        positive: statusPositive,
        durationSeconds: durationMinutes * 60,
        strengthMod: parseModifierValue(statusStrengthMod),
        dexterityMod: parseModifierValue(statusDexterityMod),
        constitutionMod: parseModifierValue(statusConstitutionMod),
        intelligenceMod: parseModifierValue(statusIntelligenceMod),
        wisdomMod: parseModifierValue(statusWisdomMod),
        charismaMod: parseModifierValue(statusCharismaMod),
      });

      setStatusGrantMessage('Status granted successfully.');
      setStatusGrantKind('success');
      setStatusName('');
      setStatusDescription('');
      setStatusEffect('');
      setStatusPositive(true);
      setStatusDurationMinutes('60');
      setStatusStrengthMod('0');
      setStatusDexterityMod('0');
      setStatusConstitutionMod('0');
      setStatusIntelligenceMod('0');
      setStatusWisdomMod('0');
      setStatusCharismaMod('0');
    } catch (error) {
      console.error('Error granting status:', error);
      setStatusGrantMessage('Failed to grant status.');
      setStatusGrantKind('error');
    } finally {
      setGrantingStatus(false);
    }
  };

  const parseResourceAmount = (value: string) => {
    const parsed = parseInt(value, 10);
    if (Number.isNaN(parsed) || parsed < 0) return 0;
    return parsed;
  };

  const adjustResources = async (healthDelta: number, manaDelta: number, successMessage: string) => {
    if (!selectedUser) return;
    if (healthDelta === 0 && manaDelta === 0) {
      setResourceMessage('Enter at least one non-zero amount.');
      setResourceMessageKind('error');
      return;
    }

    try {
      setResourceSubmitting(true);
      setResourceMessage(null);
      setResourceMessageKind(null);
      const response = await apiClient.post<AdminCharacterStats>(`/sonar/admin/users/${selectedUser.id}/resources`, {
        healthDelta,
        manaDelta,
      });
      setResourceStats(normalizeResourceStats(response));
      setResourceMessage(successMessage);
      setResourceMessageKind('success');
    } catch (error) {
      console.error('Error adjusting user resources:', error);
      setResourceMessage('Failed to adjust resources.');
      setResourceMessageKind('error');
    } finally {
      setResourceSubmitting(false);
    }
  };

  const applyDamageAndDrain = async () => {
    const healthAmount = parseResourceAmount(resourceAmountHealth);
    const manaAmount = parseResourceAmount(resourceAmountMana);
    await adjustResources(-healthAmount, -manaAmount, 'Damage/drain applied.');
  };

  const restoreHealthAndMana = async () => {
    const healthAmount = parseResourceAmount(resourceAmountHealth);
    const manaAmount = parseResourceAmount(resourceAmountMana);
    await adjustResources(healthAmount, manaAmount, 'Resources restored.');
  };

  const generateProfilePlaceholder = async () => {
    const prompt = profilePlaceholderPrompt.trim();
    if (!prompt) {
      setProfilePlaceholderError('Prompt is required.');
      return;
    }

    try {
      setProfilePlaceholderBusy(true);
      setProfilePlaceholderError(null);
      setProfilePlaceholderMessage(null);
      await apiClient.post<UserProfilePlaceholderResponse>(
        '/sonar/admin/users/profile-picture-placeholder',
        { prompt }
      );
      setProfilePlaceholderMessage('Profile placeholder queued for generation.');
      await refreshProfilePlaceholderStatus();
    } catch (error) {
      console.error('Error generating profile placeholder:', error);
      setProfilePlaceholderError(
        error instanceof Error
          ? error.message
          : 'Failed to generate profile placeholder.'
      );
    } finally {
      setProfilePlaceholderBusy(false);
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

  const usersWithoutProfilePicturesCount = users.filter(
    (user) => !user.profilePictureUrl?.trim()
  ).length;

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

      <div className="mb-6 rounded-lg border border-gray-200 bg-white p-4 shadow">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-xl font-semibold">Default Profile Placeholder</h2>
            <p className="text-sm text-gray-600">
              Applies the generated S3 image to every user whose profile image is blank.
            </p>
            <p className="text-xs text-gray-500 mt-1">
              Users without profile images: {usersWithoutProfilePicturesCount}
            </p>
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => void refreshProfilePlaceholderStatus(true)}
              disabled={profilePlaceholderStatusLoading}
              className={`px-3 py-2 rounded text-white ${
                profilePlaceholderStatusLoading
                  ? 'bg-gray-400 cursor-not-allowed'
                  : 'bg-gray-600 hover:bg-gray-700'
              }`}
            >
              {profilePlaceholderStatusLoading ? 'Refreshing...' : 'Refresh Status'}
            </button>
            <button
              onClick={generateProfilePlaceholder}
              disabled={profilePlaceholderBusy || profilePlaceholderStatusLoading}
              className={`px-3 py-2 rounded text-white ${
                profilePlaceholderBusy || profilePlaceholderStatusLoading
                  ? 'bg-gray-400 cursor-not-allowed'
                  : 'bg-violet-600 hover:bg-violet-700'
              }`}
            >
              {profilePlaceholderBusy ? 'Working...' : 'Generate Placeholder'}
            </button>
          </div>
        </div>

        <div className="mt-4 rounded-lg border border-gray-200 p-4 space-y-4">
          <div className="flex flex-wrap items-start gap-4">
            <div>
              <p className="text-sm font-medium text-gray-700 mb-2">Generated Preview</p>
              {profilePlaceholderExists ? (
                <img
                  src={`${profilePlaceholderUrl}?v=${profilePlaceholderPreviewNonce}`}
                  alt="Generated profile placeholder preview"
                  className="w-24 h-24 object-cover rounded-lg border bg-gray-50"
                />
              ) : (
                <div className="w-24 h-24 rounded-lg border bg-gray-50 text-xs text-gray-500 flex items-center justify-center text-center px-2">
                  No generated placeholder
                </div>
              )}
            </div>

            <div className="min-w-[220px] flex-1">
              <div className="flex items-center gap-2">
                <span
                  className={`inline-flex rounded-full px-2 py-0.5 text-xs font-semibold text-white ${placeholderStatusClassName(
                    profilePlaceholderStatus
                  )}`}
                >
                  {profilePlaceholderStatus || 'unknown'}
                </span>
                <span className="text-xs text-gray-500 break-all">
                  {profilePlaceholderUrl || 'No generated S3 object yet'}
                </span>
              </div>
              <p className="text-xs text-gray-500 mt-2">
                Requested:{' '}
                {profilePlaceholderRequestedAt
                  ? new Date(profilePlaceholderRequestedAt).toLocaleString()
                  : 'never'}
                {' · '}
                Last updated:{' '}
                {profilePlaceholderLastModified
                  ? new Date(profilePlaceholderLastModified).toLocaleString()
                  : 'unknown'}
              </p>
            </div>
          </div>

          <label className="block text-sm font-medium text-gray-700">
            Generation Prompt
            <textarea
              value={profilePlaceholderPrompt}
              onChange={(event) => setProfilePlaceholderPrompt(event.target.value)}
              placeholder="Prompt used to generate the shared user placeholder portrait."
              className="mt-1 min-h-[96px] w-full rounded-lg border border-gray-300 px-3 py-2"
            />
          </label>

          {profilePlaceholderMessage ? (
            <div className="rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-800">
              {profilePlaceholderMessage}
            </div>
          ) : null}
          {profilePlaceholderError ? (
            <div className="rounded-md border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-800">
              {profilePlaceholderError}
            </div>
          ) : null}
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

                {/* Resources Section */}
                <div>
                  <div className="flex justify-between items-center mb-3">
                    <h3 className="text-lg font-semibold">Health & Mana</h3>
                  </div>
                  <div className="space-y-3 rounded-lg border border-gray-200 p-4">
                    {resourceLoading ? (
                      <div className="text-sm text-gray-500">Loading resources...</div>
                    ) : resourceStats ? (
                      <div className="space-y-3">
                        <div>
                          <div className="flex justify-between text-sm mb-1">
                            <span className="font-medium text-red-700">Health</span>
                            <span className="text-gray-700">
                              {resourceStats.health} / {resourceStats.maxHealth}
                            </span>
                          </div>
                          <div className="h-2 rounded bg-red-100 overflow-hidden">
                            <div
                              className="h-full bg-red-500"
                              style={{
                                width: `${Math.max(
                                  0,
                                  Math.min(
                                    100,
                                    (resourceStats.health / Math.max(resourceStats.maxHealth, 1)) * 100,
                                  ),
                                )}%`,
                              }}
                            />
                          </div>
                        </div>

                        <div>
                          <div className="flex justify-between text-sm mb-1">
                            <span className="font-medium text-blue-700">Mana</span>
                            <span className="text-gray-700">
                              {resourceStats.mana} / {resourceStats.maxMana}
                            </span>
                          </div>
                          <div className="h-2 rounded bg-blue-100 overflow-hidden">
                            <div
                              className="h-full bg-blue-500"
                              style={{
                                width: `${Math.max(
                                  0,
                                  Math.min(
                                    100,
                                    (resourceStats.mana / Math.max(resourceStats.maxMana, 1)) * 100,
                                  ),
                                )}%`,
                              }}
                            />
                          </div>
                        </div>
                      </div>
                    ) : (
                      <div className="text-sm text-gray-500">Resource values unavailable.</div>
                    )}

                    <div className="grid grid-cols-2 gap-3">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Health amount</label>
                        <input
                          type="number"
                          min="0"
                          value={resourceAmountHealth}
                          onChange={(e) => setResourceAmountHealth(e.target.value)}
                          className="w-full px-3 py-2 border rounded-lg"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Mana amount</label>
                        <input
                          type="number"
                          min="0"
                          value={resourceAmountMana}
                          onChange={(e) => setResourceAmountMana(e.target.value)}
                          className="w-full px-3 py-2 border rounded-lg"
                        />
                      </div>
                    </div>

                    {resourceMessage && (
                      <div
                        className={`rounded-md border px-3 py-2 text-sm ${
                          resourceMessageKind === 'success'
                            ? 'border-emerald-200 bg-emerald-50 text-emerald-800'
                            : 'border-rose-200 bg-rose-50 text-rose-800'
                        }`}
                      >
                        {resourceMessage}
                      </div>
                    )}

                    <div className="flex gap-2">
                      <button
                        onClick={applyDamageAndDrain}
                        disabled={resourceSubmitting || resourceLoading}
                        className={`px-4 py-2 rounded text-white ${
                          resourceSubmitting || resourceLoading
                            ? 'bg-gray-400 cursor-not-allowed'
                            : 'bg-red-600 hover:bg-red-700'
                        }`}
                      >
                        Apply Damage/Drain
                      </button>
                      <button
                        onClick={restoreHealthAndMana}
                        disabled={resourceSubmitting || resourceLoading}
                        className={`px-4 py-2 rounded text-white ${
                          resourceSubmitting || resourceLoading
                            ? 'bg-gray-400 cursor-not-allowed'
                            : 'bg-emerald-600 hover:bg-emerald-700'
                        }`}
                      >
                        Restore
                      </button>
                    </div>
                  </div>
                </div>

                {/* Statuses Section */}
                <div>
                  <div className="flex justify-between items-center mb-3">
                    <h3 className="text-lg font-semibold">Grant Status</h3>
                  </div>
                  <div className="space-y-3 rounded-lg border border-gray-200 p-4">
                    <div className="grid grid-cols-2 gap-3">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
                        <input
                          type="text"
                          value={statusName}
                          onChange={(e) => setStatusName(e.target.value)}
                          placeholder="Inspired"
                          className="w-full px-3 py-2 border rounded-lg"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Duration (minutes)</label>
                        <input
                          type="number"
                          min="1"
                          value={statusDurationMinutes}
                          onChange={(e) => setStatusDurationMinutes(e.target.value)}
                          className="w-full px-3 py-2 border rounded-lg"
                        />
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                      <input
                        type="text"
                        value={statusDescription}
                        onChange={(e) => setStatusDescription(e.target.value)}
                        placeholder="A surge of confidence."
                        className="w-full px-3 py-2 border rounded-lg"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Effect</label>
                      <input
                        type="text"
                        value={statusEffect}
                        onChange={(e) => setStatusEffect(e.target.value)}
                        placeholder="+2 Strength"
                        className="w-full px-3 py-2 border rounded-lg"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">Type</label>
                      <div className="flex gap-2">
                        <button
                          type="button"
                          onClick={() => setStatusPositive(true)}
                          className={`px-3 py-2 rounded border text-sm font-medium ${
                            statusPositive
                              ? 'border-emerald-300 bg-emerald-100 text-emerald-800'
                              : 'border-gray-300 bg-white text-gray-700'
                          }`}
                        >
                          ↑ Buff
                        </button>
                        <button
                          type="button"
                          onClick={() => setStatusPositive(false)}
                          className={`px-3 py-2 rounded border text-sm font-medium ${
                            !statusPositive
                              ? 'border-rose-300 bg-rose-100 text-rose-800'
                              : 'border-gray-300 bg-white text-gray-700'
                          }`}
                        >
                          ↓ Debuff
                        </button>
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">Stat Modifiers</label>
                      <div className="grid grid-cols-3 gap-3">
                        <div>
                          <label className="block text-xs text-gray-600 mb-1">STR</label>
                          <input
                            type="number"
                            value={statusStrengthMod}
                            onChange={(e) => setStatusStrengthMod(e.target.value)}
                            className="w-full px-2 py-2 border rounded-lg"
                          />
                        </div>
                        <div>
                          <label className="block text-xs text-gray-600 mb-1">DEX</label>
                          <input
                            type="number"
                            value={statusDexterityMod}
                            onChange={(e) => setStatusDexterityMod(e.target.value)}
                            className="w-full px-2 py-2 border rounded-lg"
                          />
                        </div>
                        <div>
                          <label className="block text-xs text-gray-600 mb-1">CON</label>
                          <input
                            type="number"
                            value={statusConstitutionMod}
                            onChange={(e) => setStatusConstitutionMod(e.target.value)}
                            className="w-full px-2 py-2 border rounded-lg"
                          />
                        </div>
                        <div>
                          <label className="block text-xs text-gray-600 mb-1">INT</label>
                          <input
                            type="number"
                            value={statusIntelligenceMod}
                            onChange={(e) => setStatusIntelligenceMod(e.target.value)}
                            className="w-full px-2 py-2 border rounded-lg"
                          />
                        </div>
                        <div>
                          <label className="block text-xs text-gray-600 mb-1">WIS</label>
                          <input
                            type="number"
                            value={statusWisdomMod}
                            onChange={(e) => setStatusWisdomMod(e.target.value)}
                            className="w-full px-2 py-2 border rounded-lg"
                          />
                        </div>
                        <div>
                          <label className="block text-xs text-gray-600 mb-1">CHA</label>
                          <input
                            type="number"
                            value={statusCharismaMod}
                            onChange={(e) => setStatusCharismaMod(e.target.value)}
                            className="w-full px-2 py-2 border rounded-lg"
                          />
                        </div>
                      </div>
                    </div>

                    {statusGrantMessage && (
                      <div
                        className={`rounded-md border px-3 py-2 text-sm ${
                          statusGrantKind === 'success'
                            ? 'border-emerald-200 bg-emerald-50 text-emerald-800'
                            : 'border-rose-200 bg-rose-50 text-rose-800'
                        }`}
                      >
                        {statusGrantMessage}
                      </div>
                    )}

                    <button
                      onClick={grantStatus}
                      disabled={grantingStatus || statusName.trim() === ''}
                      className={`px-4 py-2 rounded text-white ${
                        grantingStatus || statusName.trim() === ''
                          ? 'bg-gray-400 cursor-not-allowed'
                          : 'bg-indigo-600 hover:bg-indigo-700'
                      }`}
                    >
                      {grantingStatus ? 'Granting...' : 'Grant Status'}
                    </button>
                  </div>
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
