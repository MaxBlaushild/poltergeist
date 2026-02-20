import { useAPI } from '@poltergeist/contexts';
import { Party, User } from '@poltergeist/types';
import React, { useCallback, useEffect, useMemo, useState } from 'react';

const MAX_PARTY_SIZE = 5;

const formatUserLabel = (user?: User | null) => {
  if (!user) return 'Unknown user';
  return user.username || user.name || user.id;
};

const getPartyLabel = (party: Party) => {
  const leaderLabel = formatUserLabel(party.leader);
  return `${leaderLabel} (${party.members?.length ?? 0} members)`;
};

export const Parties = () => {
  const { apiClient } = useAPI();
  const [parties, setParties] = useState<Party[]>([]);
  const [filteredParties, setFilteredParties] = useState<Party[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [selectedParty, setSelectedParty] = useState<Party | null>(null);
  const [leaderSelection, setLeaderSelection] = useState<string>('');

  const [showCreateParty, setShowCreateParty] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [partyToDelete, setPartyToDelete] = useState<Party | null>(null);

  const [memberSearchQuery, setMemberSearchQuery] = useState('');
  const [memberSearchResults, setMemberSearchResults] = useState<User[]>([]);
  const [searchingMembers, setSearchingMembers] = useState(false);

  const [createLeaderQuery, setCreateLeaderQuery] = useState('');
  const [createLeaderResults, setCreateLeaderResults] = useState<User[]>([]);
  const [createLeaderSearching, setCreateLeaderSearching] = useState(false);
  const [selectedLeader, setSelectedLeader] = useState<User | null>(null);

  const [createMemberQuery, setCreateMemberQuery] = useState('');
  const [createMemberResults, setCreateMemberResults] = useState<User[]>([]);
  const [createMemberSearching, setCreateMemberSearching] = useState(false);
  const [selectedMembers, setSelectedMembers] = useState<User[]>([]);

  useEffect(() => {
    fetchParties();
  }, []);

  useEffect(() => {
    if (!searchQuery) {
      setFilteredParties(parties);
      return;
    }

    const query = searchQuery.toLowerCase();
    const filtered = parties.filter((party) => {
      const leader = party.leader?.username?.toLowerCase() || party.leader?.name?.toLowerCase() || '';
      return leader.includes(query) || party.id.toLowerCase().includes(query);
    });
    setFilteredParties(filtered);
  }, [searchQuery, parties]);

  useEffect(() => {
    if (!selectedParty) return;
    const refreshed = parties.find((party) => party.id === selectedParty.id) || null;
    setSelectedParty(refreshed);
    if (refreshed?.leaderId) {
      setLeaderSelection(refreshed.leaderId);
    }
  }, [parties, selectedParty]);

  const fetchParties = async (selectId?: string) => {
    try {
      setLoading(true);
      const response = await apiClient.get<Party[]>('/sonar/admin/parties');
      const list = Array.isArray(response) ? response : [];
      setParties(list);
      if (selectId) {
        const selected = list.find((party) => party.id === selectId) || null;
        setSelectedParty(selected);
        if (selected?.leaderId) {
          setLeaderSelection(selected.leaderId);
        }
      }
    } catch (error) {
      console.error('Error fetching parties:', error);
    } finally {
      setLoading(false);
    }
  };

  const searchUsers = useCallback(async (query: string, setResults: (users: User[]) => void, setSearching: (value: boolean) => void) => {
    if (!query || query.length < 2) {
      setResults([]);
      return;
    }

    try {
      setSearching(true);
      const response = await apiClient.get<User[]>(`/sonar/users/search?query=${encodeURIComponent(query)}`);
      setResults(Array.isArray(response) ? response : []);
    } catch (error) {
      console.error('Error searching users:', error);
      setResults([]);
    } finally {
      setSearching(false);
    }
  }, [apiClient]);

  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (memberSearchQuery) {
        searchUsers(memberSearchQuery, setMemberSearchResults, setSearchingMembers);
      } else {
        setMemberSearchResults([]);
      }
    }, 300);

    return () => clearTimeout(timeoutId);
  }, [memberSearchQuery, searchUsers]);

  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (createLeaderQuery) {
        searchUsers(createLeaderQuery, setCreateLeaderResults, setCreateLeaderSearching);
      } else {
        setCreateLeaderResults([]);
      }
    }, 300);

    return () => clearTimeout(timeoutId);
  }, [createLeaderQuery, searchUsers]);

  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (createMemberQuery) {
        searchUsers(createMemberQuery, setCreateMemberResults, setCreateMemberSearching);
      } else {
        setCreateMemberResults([]);
      }
    }, 300);

    return () => clearTimeout(timeoutId);
  }, [createMemberQuery, searchUsers]);

  const membersInSelectedParty = useMemo(() => {
    if (!selectedParty) return new Set<string>();
    return new Set((selectedParty.members || []).map((member) => member.id));
  }, [selectedParty]);

  const availableMemberResults = memberSearchResults.filter((user) => !membersInSelectedParty.has(user.id));

  const resetCreateForm = () => {
    setCreateLeaderQuery('');
    setCreateLeaderResults([]);
    setSelectedLeader(null);
    setCreateMemberQuery('');
    setCreateMemberResults([]);
    setSelectedMembers([]);
  };

  const handleSelectParty = (party: Party) => {
    setSelectedParty(party);
    setLeaderSelection(party.leaderId || '');
    setMemberSearchQuery('');
    setMemberSearchResults([]);
  };

  const handleCreateParty = async () => {
    if (!selectedLeader) {
      alert('Select a leader to create a party.');
      return;
    }
    const totalMembers = selectedMembers.length + 1;
    if (totalMembers > MAX_PARTY_SIZE) {
      alert(`Party size cannot exceed ${MAX_PARTY_SIZE} members.`);
      return;
    }

    try {
      const response = await apiClient.post<Party>('/sonar/admin/parties', {
        leaderId: selectedLeader.id,
        memberIds: selectedMembers.map((member) => member.id),
      });
      await fetchParties(response.id);
      setShowCreateParty(false);
      resetCreateForm();
    } catch (error: any) {
      console.error('Error creating party:', error);
      const errorMessage = error.response?.data?.error || error.message || 'Failed to create party';
      alert(`Error: ${errorMessage}`);
    }
  };

  const handleSetLeader = async () => {
    if (!selectedParty || !leaderSelection) return;

    try {
      await apiClient.patch(`/sonar/admin/parties/${selectedParty.id}/leader`, {
        leaderId: leaderSelection,
      });
      await fetchParties(selectedParty.id);
    } catch (error: any) {
      console.error('Error setting party leader:', error);
      const errorMessage = error.response?.data?.error || error.message || 'Failed to set leader';
      alert(`Error: ${errorMessage}`);
    }
  };

  const handleAddMember = async (userId: string) => {
    if (!selectedParty) return;
    if ((selectedParty.members?.length ?? 0) >= MAX_PARTY_SIZE) {
      alert(`Party is already at the maximum size of ${MAX_PARTY_SIZE}.`);
      return;
    }

    try {
      await apiClient.post(`/sonar/admin/parties/${selectedParty.id}/members`, { userId });
      await fetchParties(selectedParty.id);
      setMemberSearchQuery('');
      setMemberSearchResults([]);
    } catch (error: any) {
      console.error('Error adding party member:', error);
      const errorMessage = error.response?.data?.error || error.message || 'Failed to add member';
      alert(`Error: ${errorMessage}`);
    }
  };

  const handleRemoveMember = async (userId: string) => {
    if (!selectedParty) return;

    try {
      await apiClient.delete(`/sonar/admin/parties/${selectedParty.id}/members/${userId}`);
      await fetchParties(selectedParty.id);
    } catch (error: any) {
      console.error('Error removing party member:', error);
      const errorMessage = error.response?.data?.error || error.message || 'Failed to remove member';
      alert(`Error: ${errorMessage}`);
    }
  };

  const handleDeleteParty = async () => {
    if (!partyToDelete) return;

    try {
      await apiClient.delete(`/sonar/admin/parties/${partyToDelete.id}`);
      await fetchParties();
      setShowDeleteConfirm(false);
      setPartyToDelete(null);
      if (selectedParty?.id === partyToDelete.id) {
        setSelectedParty(null);
      }
    } catch (error: any) {
      console.error('Error deleting party:', error);
      const errorMessage = error.response?.data?.error || error.message || 'Failed to delete party';
      alert(`Error: ${errorMessage}`);
    }
  };

  const partyMemberCount = selectedParty?.members?.length ?? 0;

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold">Parties</h1>
        <button
          onClick={() => setShowCreateParty(true)}
          className="bg-blue-500 text-white px-4 py-2 rounded-md"
        >
          Create Party
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-[320px_1fr] gap-6">
        <div className="bg-white border rounded-lg p-4">
          <div className="mb-4">
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search parties by leader or ID"
              className="w-full p-2 border rounded-md"
            />
          </div>

          {loading ? (
            <div>Loading parties...</div>
          ) : filteredParties.length === 0 ? (
            <div className="text-gray-600">No parties found</div>
          ) : (
            <ul className="space-y-2">
              {filteredParties.map((party) => (
                <li key={party.id}>
                  <button
                    onClick={() => handleSelectParty(party)}
                    className={`w-full text-left px-3 py-2 rounded-md border ${
                      selectedParty?.id === party.id
                        ? 'border-blue-500 bg-blue-50'
                        : 'border-gray-200 hover:border-blue-300'
                    }`}
                  >
                    <div className="font-medium">{getPartyLabel(party)}</div>
                    <div className="text-xs text-gray-500">{party.id}</div>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>

        <div className="bg-white border rounded-lg p-6">
          {!selectedParty ? (
            <div className="text-gray-600">Select a party to manage.</div>
          ) : (
            <div className="space-y-6">
              <div className="flex items-start justify-between">
                <div>
                  <h2 className="text-2xl font-semibold">Party Details</h2>
                  <p className="text-sm text-gray-500">{selectedParty.id}</p>
                </div>
                <button
                  onClick={() => {
                    setPartyToDelete(selectedParty);
                    setShowDeleteConfirm(true);
                  }}
                  className="bg-red-500 text-white px-3 py-2 rounded-md"
                >
                  Delete Party
                </button>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="p-4 border rounded-md">
                  <div className="text-sm text-gray-500 mb-2">Leader</div>
                  <div className="font-medium">{formatUserLabel(selectedParty.leader)}</div>
                  <div className="text-xs text-gray-500">{selectedParty.leaderId}</div>
                </div>
                <div className="p-4 border rounded-md">
                  <div className="text-sm text-gray-500 mb-2">Members</div>
                  <div className="font-medium">{partyMemberCount} / {MAX_PARTY_SIZE}</div>
                </div>
              </div>

              <div className="border rounded-md p-4">
                <h3 className="text-lg font-semibold mb-3">Set Leader</h3>
                <div className="flex flex-col md:flex-row gap-3">
                  <select
                    value={leaderSelection}
                    onChange={(e) => setLeaderSelection(e.target.value)}
                    className="border p-2 rounded-md flex-1"
                  >
                    {(selectedParty.members || []).map((member) => (
                      <option key={member.id} value={member.id}>
                        {formatUserLabel(member)}
                      </option>
                    ))}
                  </select>
                  <button
                    onClick={handleSetLeader}
                    className="bg-blue-500 text-white px-4 py-2 rounded-md"
                  >
                    Set Leader
                  </button>
                </div>
              </div>

              <div className="border rounded-md p-4">
                <h3 className="text-lg font-semibold mb-3">Members</h3>
                {(selectedParty.members || []).length === 0 ? (
                  <div className="text-gray-600">No members in this party.</div>
                ) : (
                  <ul className="space-y-2">
                    {(selectedParty.members || []).map((member) => (
                      <li key={member.id} className="flex items-center justify-between border rounded-md px-3 py-2">
                        <div>
                          <div className="font-medium">{formatUserLabel(member)}</div>
                          <div className="text-xs text-gray-500">{member.id}</div>
                        </div>
                        {member.id === selectedParty.leaderId ? (
                          <span className="text-xs font-semibold text-blue-600">Leader</span>
                        ) : (
                          <button
                            onClick={() => handleRemoveMember(member.id)}
                            className="text-red-600 text-sm"
                          >
                            Remove
                          </button>
                        )}
                      </li>
                    ))}
                  </ul>
                )}
              </div>

              <div className="border rounded-md p-4">
                <h3 className="text-lg font-semibold mb-3">Add Member</h3>
                <input
                  type="text"
                  value={memberSearchQuery}
                  onChange={(e) => setMemberSearchQuery(e.target.value)}
                  placeholder="Search users by username"
                  className="w-full p-2 border rounded-md"
                />

                {searchingMembers && <div className="mt-2 text-gray-600">Searching users...</div>}

                {!searchingMembers && memberSearchQuery && availableMemberResults.length === 0 && (
                  <div className="mt-2 text-gray-600">No users found</div>
                )}

                {availableMemberResults.length > 0 && (
                  <div className="mt-3 space-y-2">
                    {availableMemberResults.map((user) => (
                      <div key={user.id} className="flex items-center justify-between border rounded-md px-3 py-2">
                        <div>
                          <div className="font-medium">{formatUserLabel(user)}</div>
                          <div className="text-xs text-gray-500">{user.id}</div>
                        </div>
                        <button
                          onClick={() => handleAddMember(user.id)}
                          className="bg-blue-500 text-white px-3 py-1 rounded-md"
                        >
                          Add
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>

      {showCreateParty && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg w-[520px] max-h-[80vh] overflow-auto">
            <h2 className="text-xl font-bold mb-4">Create Party</h2>

            <div className="mb-4">
              <label className="block mb-2">Leader *</label>
              {selectedLeader ? (
                <div className="flex items-center justify-between border rounded-md px-3 py-2">
                  <div>
                    <div className="font-medium">{formatUserLabel(selectedLeader)}</div>
                    <div className="text-xs text-gray-500">{selectedLeader.id}</div>
                  </div>
                  <button
                    onClick={() => setSelectedLeader(null)}
                    className="text-red-600 text-sm"
                  >
                    Change
                  </button>
                </div>
              ) : (
                <>
                  <input
                    type="text"
                    value={createLeaderQuery}
                    onChange={(e) => setCreateLeaderQuery(e.target.value)}
                    placeholder="Search for leader"
                    className="w-full p-2 border rounded-md"
                  />
                  {createLeaderSearching && <div className="mt-2 text-gray-600">Searching users...</div>}
                  {!createLeaderSearching && createLeaderQuery && createLeaderResults.length === 0 && (
                    <div className="mt-2 text-gray-600">No users found</div>
                  )}
                  {createLeaderResults.length > 0 && (
                    <div className="mt-3 space-y-2">
                      {createLeaderResults.map((user) => (
                        <button
                          key={user.id}
                          type="button"
                          onClick={() => {
                            setSelectedLeader(user);
                            setSelectedMembers((prev) => prev.filter((member) => member.id !== user.id));
                            setCreateLeaderQuery('');
                            setCreateLeaderResults([]);
                          }}
                          className="w-full text-left border rounded-md px-3 py-2 hover:border-blue-300"
                        >
                          <div className="font-medium">{formatUserLabel(user)}</div>
                          <div className="text-xs text-gray-500">{user.id}</div>
                        </button>
                      ))}
                    </div>
                  )}
                </>
              )}
            </div>

            <div className="mb-4">
              <label className="block mb-2">Members ({selectedMembers.length}{selectedLeader ? ` + 1 leader` : ''})</label>
              <input
                type="text"
                value={createMemberQuery}
                onChange={(e) => setCreateMemberQuery(e.target.value)}
                placeholder="Search users to add"
                className="w-full p-2 border rounded-md"
              />
              {createMemberSearching && <div className="mt-2 text-gray-600">Searching users...</div>}
              {!createMemberSearching && createMemberQuery && createMemberResults.length === 0 && (
                <div className="mt-2 text-gray-600">No users found</div>
              )}
              {createMemberResults.length > 0 && (
                <div className="mt-3 space-y-2">
                  {createMemberResults.map((user) => {
                    const isLeader = selectedLeader?.id === user.id;
                    const isSelected = selectedMembers.some((member) => member.id === user.id);
                    const isAtCapacity = selectedLeader ? selectedMembers.length + 1 >= MAX_PARTY_SIZE : selectedMembers.length >= MAX_PARTY_SIZE;

                    return (
                      <div key={user.id} className="flex items-center justify-between border rounded-md px-3 py-2">
                        <div>
                          <div className="font-medium">{formatUserLabel(user)}</div>
                          <div className="text-xs text-gray-500">{user.id}</div>
                        </div>
                        <button
                          type="button"
                          onClick={() => {
                            if (isLeader || isSelected || isAtCapacity) return;
                            setSelectedMembers([...selectedMembers, user]);
                            setCreateMemberQuery('');
                            setCreateMemberResults([]);
                          }}
                          className={`px-3 py-1 rounded-md text-sm ${
                            isLeader || isSelected
                              ? 'bg-gray-200 text-gray-600'
                              : isAtCapacity
                              ? 'bg-gray-200 text-gray-600'
                              : 'bg-blue-500 text-white'
                          }`}
                        >
                          {isLeader ? 'Leader' : isSelected ? 'Added' : isAtCapacity ? 'Full' : 'Add'}
                        </button>
                      </div>
                    );
                  })}
                </div>
              )}

              {selectedMembers.length > 0 && (
                <div className="mt-4">
                  <div className="text-sm font-semibold mb-2">Selected Members</div>
                  <div className="space-y-2">
                    {selectedMembers.map((member) => (
                      <div key={member.id} className="flex items-center justify-between border rounded-md px-3 py-2">
                        <div>
                          <div className="font-medium">{formatUserLabel(member)}</div>
                          <div className="text-xs text-gray-500">{member.id}</div>
                        </div>
                        <button
                          type="button"
                          onClick={() => setSelectedMembers(selectedMembers.filter((m) => m.id !== member.id))}
                          className="text-red-600 text-sm"
                        >
                          Remove
                        </button>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>

            <div className="flex gap-2">
              <button
                onClick={handleCreateParty}
                className="bg-blue-500 text-white px-4 py-2 rounded-md"
              >
                Create
              </button>
              <button
                onClick={() => {
                  setShowCreateParty(false);
                  resetCreateForm();
                }}
                className="bg-gray-500 text-white px-4 py-2 rounded-md"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {showDeleteConfirm && partyToDelete && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg w-96">
            <h2 className="text-xl font-bold mb-4">Confirm Delete</h2>
            <p className="mb-4">Are you sure you want to delete this party? Members will be removed from the party.</p>
            <div className="flex gap-2">
              <button
                onClick={handleDeleteParty}
                className="bg-red-500 text-white px-4 py-2 rounded-md"
              >
                Delete
              </button>
              <button
                onClick={() => {
                  setShowDeleteConfirm(false);
                  setPartyToDelete(null);
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
