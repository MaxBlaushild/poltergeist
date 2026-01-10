import { useAPI } from '@poltergeist/contexts';
import { FeteRoomTeam, FeteRoom, FeteTeam } from '@poltergeist/types';
import React, { useState, useEffect, useMemo } from 'react';

export const FeteRoomTeams = () => {
  const { apiClient } = useAPI();
  const [roomTeams, setRoomTeams] = useState<FeteRoomTeam[]>([]);
  const [rooms, setRooms] = useState<FeteRoom[]>([]);
  const [teams, setTeams] = useState<FeteTeam[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [togglingStates, setTogglingStates] = useState<Record<string, boolean>>({});
  const [error, setError] = useState<string | null>(null);

  // Optimize data structure: Create a map for quick lookup of room-team relationships
  // Key format: `${teamId}-${roomId}`, value: FeteRoomTeam (for ID lookup)
  const roomTeamMap = useMemo(() => {
    const map: Record<string, FeteRoomTeam> = {};
    roomTeams.forEach(rt => {
      const key = `${rt.teamId}-${rt.feteRoomId}`;
      map[key] = rt;
    });
    return map;
  }, [roomTeams]);

  // Filter teams based on search query (rooms are shown for all teams)
  const filteredTeams = useMemo(() => {
    if (searchQuery === '') return teams;
    const query = searchQuery.toLowerCase();
    return teams.filter(team => 
      team.name?.toLowerCase().includes(query)
    );
  }, [teams, searchQuery]);

  useEffect(() => {
    fetchAllData();
  }, []);

  const fetchAllData = async () => {
    setLoading(true);
    setError(null);
    try {
      await Promise.all([fetchRoomTeams(), fetchRooms(), fetchTeams()]);
    } catch (err) {
      console.error('Error fetching data:', err);
      setError('Failed to load data. Please refresh the page.');
    } finally {
      setLoading(false);
    }
  };

  const fetchRoomTeams = async () => {
    try {
      const response = await apiClient.get<FeteRoomTeam[]>('/final-fete/room-teams');
      setRoomTeams(Array.isArray(response) ? response : []);
    } catch (error) {
      console.error('Error fetching fete room teams:', error);
      setRoomTeams([]);
      throw error;
    }
  };

  const fetchRooms = async () => {
    try {
      const response = await apiClient.get<FeteRoom[]>('/final-fete/rooms');
      setRooms(Array.isArray(response) ? response : []);
    } catch (error) {
      console.error('Error fetching fete rooms:', error);
      setRooms([]);
      throw error;
    }
  };

  const fetchTeams = async () => {
    try {
      const response = await apiClient.get<FeteTeam[]>('/final-fete/teams');
      setTeams(Array.isArray(response) ? response : []);
    } catch (error) {
      console.error('Error fetching fete teams:', error);
      setTeams([]);
      throw error;
    }
  };

  // Check if a room is unlocked for a team
  const isRoomUnlocked = (teamId: string, roomId: string): boolean => {
    const key = `${teamId}-${roomId}`;
    return key in roomTeamMap;
  };

  // Get the relationship ID for a room-team pair
  const getRelationshipId = (teamId: string, roomId: string): string | null => {
    const key = `${teamId}-${roomId}`;
    return roomTeamMap[key]?.id || null;
  };

  // Toggle room unlock status for a team
  const handleToggleRoom = async (teamId: string, roomId: string) => {
    const toggleKey = `${teamId}-${roomId}`;
    const isUnlocked = isRoomUnlocked(teamId, roomId);
    
    // Set loading state for this specific toggle
    setTogglingStates(prev => ({ ...prev, [toggleKey]: true }));
    setError(null);

    try {
      if (isUnlocked) {
        // Lock the room: delete the relationship
        const relationshipId = getRelationshipId(teamId, roomId);
        if (!relationshipId) {
          throw new Error('Relationship ID not found');
        }
        
        await apiClient.delete(`/final-fete/room-teams/${relationshipId}`);
        
        // Optimistically update state
        setRoomTeams(prev => prev.filter(rt => rt.id !== relationshipId));
      } else {
        // Unlock the room: create the relationship
        const newRelationship = await apiClient.post<FeteRoomTeam>('/final-fete/room-teams', {
          teamId,
          feteRoomId: roomId,
        });
        
        // Optimistically update state
        setRoomTeams(prev => [...prev, newRelationship]);
      }
    } catch (error: any) {
      console.error('Error toggling room unlock:', error);
      setError(error?.message || `Failed to ${isUnlocked ? 'lock' : 'unlock'} room. Please try again.`);
      // Refresh data to sync with server state
      await fetchRoomTeams();
    } finally {
      // Clear loading state
      setTogglingStates(prev => {
        const updated = { ...prev };
        delete updated[toggleKey];
        return updated;
      });
    }
  };

  const isToggling = (teamId: string, roomId: string): boolean => {
    const toggleKey = `${teamId}-${roomId}`;
    return togglingStates[toggleKey] || false;
  };

  if (loading) {
    return (
      <div className="m-10">
        <div className="text-xl font-semibold mb-4">Loading team room unlocks...</div>
        <div className="text-gray-600">Please wait while we fetch the data.</div>
      </div>
    );
  }

  return (
    <div className="m-10">
      <div className="mb-6">
        <h1 className="text-3xl font-bold mb-2">Team Room Unlocks</h1>
        <p className="text-gray-600">Manage which teams have unlocked which final fete rooms</p>
      </div>

      {error && (
        <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-md">
          <div className="flex items-center justify-between">
            <span className="text-red-800">{error}</span>
            <button
              onClick={() => setError(null)}
              className="text-red-600 hover:text-red-800"
            >
              ×
            </button>
          </div>
        </div>
      )}

      <div className="mb-4">
        <input
          type="text"
          placeholder="Search by team name..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full max-w-md p-3 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      <div className="mb-4 flex items-center gap-4">
        <button
          onClick={fetchAllData}
          className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-md transition-colors"
          disabled={loading}
        >
          Refresh Data
        </button>
        <div className="text-sm text-gray-600">
          {roomTeams.length} unlock{roomTeams.length !== 1 ? 's' : ''} across {teams.length} team{teams.length !== 1 ? 's' : ''} and {rooms.length} room{rooms.length !== 1 ? 's' : ''}
        </div>
      </div>

      {filteredTeams.length === 0 ? (
        <div className="p-8 border border-gray-200 rounded-lg bg-gray-50 text-center">
          <p className="text-gray-600">
            {searchQuery ? 'No teams match your search.' : 'No teams found.'}
          </p>
        </div>
      ) : (
        <div className="space-y-6">
          {filteredTeams.map((team) => {
            const teamUnlockedRooms = roomTeams
              .filter(rt => rt.teamId === team.id)
              .map(rt => rt.feteRoomId);
            const unlockedCount = teamUnlockedRooms.length;
            const totalRooms = rooms.length;

            return (
              <div
                key={team.id}
                className="border border-gray-300 rounded-lg bg-white shadow-sm hover:shadow-md transition-shadow"
              >
                <div className="p-4 bg-gray-50 border-b border-gray-200 rounded-t-lg">
                  <div className="flex items-center justify-between">
                    <div>
                      <h2 className="text-xl font-semibold text-gray-800">{team.name}</h2>
                      <p className="text-sm text-gray-600 mt-1">
                        {unlockedCount} of {totalRooms} room{totalRooms !== 1 ? 's' : ''} unlocked
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      {totalRooms > 0 && (
                        <div className="w-32 h-2 bg-gray-200 rounded-full overflow-hidden">
                          <div
                            className="h-full bg-green-500 transition-all duration-300"
                            style={{ width: `${(unlockedCount / totalRooms) * 100}%` }}
                          />
                        </div>
                      )}
                    </div>
                  </div>
                </div>

                <div className="p-4">
                  {rooms.length === 0 ? (
                    <p className="text-gray-500 italic">No rooms available.</p>
                  ) : (
                    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
                      {rooms.map((room) => {
                        const unlocked = isRoomUnlocked(team.id, room.id);
                        const toggling = isToggling(team.id, room.id);

                        return (
                          <label
                            key={room.id}
                            className={`
                              flex items-center gap-3 p-3 border rounded-md cursor-pointer transition-all
                              ${unlocked
                                ? 'border-green-300 bg-green-50 hover:bg-green-100'
                                : 'border-gray-300 bg-white hover:bg-gray-50'
                              }
                              ${toggling ? 'opacity-50 cursor-wait' : ''}
                            `}
                          >
                            <input
                              type="checkbox"
                              checked={unlocked}
                              onChange={() => handleToggleRoom(team.id, room.id)}
                              disabled={toggling}
                              className="w-5 h-5 text-green-600 border-gray-300 rounded focus:ring-2 focus:ring-green-500 disabled:opacity-50"
                            />
                            <div className="flex-1 min-w-0">
                              <div className="flex items-center gap-2">
                                <span className={`font-medium ${unlocked ? 'text-green-800' : 'text-gray-700'}`}>
                                  {room.name}
                                </span>
                                {toggling && (
                                  <span className="text-xs text-gray-500">(updating...)</span>
                                )}
                              </div>
                              <div className="flex items-center gap-2 mt-1">
                                {unlocked ? (
                                  <>
                                    <span className="inline-block w-2 h-2 bg-green-500 rounded-full"></span>
                                    <span className="text-xs text-green-700">Unlocked</span>
                                  </>
                                ) : (
                                  <>
                                    <span className="inline-block w-2 h-2 bg-gray-400 rounded-full"></span>
                                    <span className="text-xs text-gray-500">Locked</span>
                                  </>
                                )}
                              </div>
                            </div>
                          </label>
                        );
                      })}
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

