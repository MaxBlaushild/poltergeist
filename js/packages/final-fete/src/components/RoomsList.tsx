import { useAPI, useAuth } from '@poltergeist/contexts';
import type { FeteRoom, FeteTeam, FeteRoomLinkedListTeam, FeteRoomTeam } from '@poltergeist/types';
import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

export const RoomsList = () => {
  const { apiClient } = useAPI();
  const { logout } = useAuth();
  const navigate = useNavigate();
  const [rooms, setRooms] = useState<FeteRoom[]>([]);
  const [teams, setTeams] = useState<FeteTeam[]>([]);
  const [linkedListTeams, setLinkedListTeams] = useState<FeteRoomLinkedListTeam[]>([]);
  const [roomTeams, setRoomTeams] = useState<FeteRoomTeam[]>([]);
  const [userTeam, setUserTeam] = useState<FeteTeam | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showLockConfirm, setShowLockConfirm] = useState(false);
  const [roomToLock, setRoomToLock] = useState<FeteRoom | null>(null);

  useEffect(() => {
    fetchRooms();
    fetchTeams();
    fetchUserTeam();
    fetchLinkedListTeams();
    fetchRoomTeams();
  }, []);

  const fetchRooms = async () => {
    try {
      setLoading(true);
      const response = await apiClient.get<FeteRoom[]>('/final-fete/rooms');
      setRooms(Array.isArray(response) ? response : []);
      setError(null);
    } catch (err) {
      console.error('Error fetching rooms:', err);
      setError('Failed to load rooms');
    } finally {
      setLoading(false);
    }
  };

  const fetchTeams = async () => {
    try {
      const response = await apiClient.get<FeteTeam[]>('/final-fete/teams');
      setTeams(Array.isArray(response) ? response : []);
    } catch (err) {
      console.error('Error fetching teams:', err);
    }
  };

  const fetchUserTeam = async () => {
    try {
      const response = await apiClient.get<FeteTeam | null>('/final-fete/teams/current');
      setUserTeam(response);
    } catch (err) {
      console.error('Error fetching user team:', err);
      setUserTeam(null);
    }
  };

  const fetchLinkedListTeams = async () => {
    try {
      const response = await apiClient.get<FeteRoomLinkedListTeam[]>('/final-fete/room-linked-list-teams');
      setLinkedListTeams(Array.isArray(response) ? response : []);
    } catch (err) {
      console.error('Error fetching linked list teams:', err);
      setLinkedListTeams([]);
    }
  };

  const fetchRoomTeams = async () => {
    try {
      const response = await apiClient.get<FeteRoomTeam[]>('/final-fete/room-teams');
      setRoomTeams(Array.isArray(response) ? response : []);
    } catch (err) {
      console.error('Error fetching room teams:', err);
      setRoomTeams([]);
    }
  };

  const handleToggleRoom = async (roomId: string) => {
    try {
      await apiClient.post(`/final-fete/rooms/${roomId}/toggle`, {});
      // Refresh rooms after toggle
      fetchRooms();
      setShowLockConfirm(false);
      setRoomToLock(null);
    } catch (err) {
      console.error('Error toggling room:', err);
      alert('Failed to toggle room. Please try again.');
    }
  };

  const handleUnlockRoomClick = (room: FeteRoom) => {
    setRoomToLock(room);
    setShowLockConfirm(true);
  };

  const handleLockRoomClick = (roomId: string) => {
    handleToggleRoom(roomId);
  };

  if (loading) {
    return <div className="p-4 md:p-6 lg:p-10 text-[#00ff00]">Loading rooms...</div>;
  }

  if (error) {
    return <div className="p-4 md:p-6 lg:p-10 text-red-500" style={{ textShadow: '0 0 10px #ff0000' }}>{error}</div>;
  }

  return (
    <div className="p-4 md:p-6 lg:p-10 text-[#00ff00]">
      <div className="flex flex-col md:flex-row md:justify-between md:items-center mb-4 gap-4">
        <div>
          <h1 className="text-xl md:text-2xl font-bold text-[#00ff00]">Bunker Rooms</h1>
          {userTeam && (
            <p className="text-sm text-[#00ff00] mt-1 opacity-80">
              Your Team: <span className="font-semibold text-[#00ff00]">{userTeam.name}</span>
            </p>
          )}
        </div>
        <div className="flex gap-2 w-full md:w-auto">
          <button
            onClick={() => navigate('/scan-qr')}
            className="matrix-button matrix-button-primary flex-1 md:flex-none min-h-[44px]"
          >
            Import Room Controls
          </button>
        </div>
      </div>
      
      {(() => {
        // Filter rooms to only show those where the user's team has access
        const accessibleRoomIds = userTeam 
          ? roomTeams
              .filter(rt => rt.teamId === userTeam.id)
              .map(rt => rt.feteRoomId)
          : [];
        
        const filteredRooms = userTeam 
          ? rooms.filter(room => accessibleRoomIds.includes(room.id))
          : [];

        if (filteredRooms.length === 0) {
          return <p className="text-[#00ff00] opacity-70">No rooms available.</p>;
        }

        return (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredRooms.map((room) => {
            const currentTeam = teams.find(t => t.id === room.currentTeamId);
            const isUserTeam = userTeam && room.currentTeamId && userTeam.id === room.currentTeamId;
            // Find the linked list item where current team is the first team to get the next team
            const linkedListItem = linkedListTeams.find(
              item => item.feteRoomId === room.id && item.firstTeamId === room.currentTeamId
            );
            const nextTeam = linkedListItem ? teams.find(t => t.id === linkedListItem.secondTeamId) : null;
            return (
              <div key={room.id} className="p-3 md:p-4 border-2 border-[#00ff00] rounded-lg bg-black/80 backdrop-blur-sm matrix-card">
                <h2 className="text-lg font-semibold mb-2 text-[#00ff00]">{room.name}</h2>
                <p className="text-sm text-[#00ff00] mb-1 opacity-80">
                  Status: <span className={room.open ? 'text-[#00ff00]' : 'text-red-500'} style={room.open ? { textShadow: '0 0 10px #00ff00' } : { textShadow: '0 0 10px #ff0000' }}>
                    {room.open ? 'Open' : 'Closed'}
                  </span>
                </p>
                <p className="text-sm text-[#00ff00] mb-1 opacity-80">
                  Current Team:{' '}
                  {isUserTeam ? (
                    <span className="inline-flex items-center gap-1">
                      <span className="pulsing-green-highlight px-2 py-1 rounded font-semibold text-[#00ff00]">
                        Your Team: {currentTeam?.name || userTeam?.name || 'Unknown'}
                      </span>
                    </span>
                  ) : (
                    <span className="opacity-70">{currentTeam?.name || room.currentTeamId || 'None'}</span>
                  )}
                </p>
                {nextTeam && (
                  <p className="text-sm text-[#00ff00] mb-4 opacity-80">
                    Next Team: <span className="opacity-70">{nextTeam.name}</span>
                  </p>
                )}
                {!nextTeam && (
                  <p className="text-sm text-[#00ff00] mb-4 opacity-50 italic">
                    Next Team: Unknown
                  </p>
                )}
                {isUserTeam && (
                  <div className="flex gap-2">
                    <button
                      onClick={() => room.open ? handleLockRoomClick(room.id) : handleUnlockRoomClick(room)}
                      className={`w-full md:w-auto px-4 py-2 rounded-md min-h-[44px] matrix-button ${
                        room.open
                          ? 'matrix-button-danger'
                          : 'matrix-button-success'
                      }`}
                    >
                      {room.open ? 'Lock Room' : 'Unlock Room'}
                    </button>
                  </div>
                )}
              </div>
            );
            })}
          </div>
        );
      })()}

      {showLockConfirm && roomToLock && (
        <>
          {/* Backdrop */}
          <div 
            className="fixed inset-0 bg-black/80 z-50 md:hidden"
            onClick={() => {
              setShowLockConfirm(false);
              setRoomToLock(null);
            }}
          />
          
          {/* Mobile Bottom Sheet */}
          <div className="fixed bottom-0 left-0 right-0 z-50 md:hidden animate-slide-up">
            <div className="bg-black/95 backdrop-blur-sm border-t-2 border-[#00ff00] rounded-t-lg shadow-[0_0_30px_rgba(0,255,0,0.5)] p-6">
              <h2 className="text-xl font-bold text-[#00ff00] mb-4">Confirm Unlock Room</h2>
              <p className="text-[#00ff00] mb-2 opacity-90">
                You are about to unlock <span className="font-semibold">{roomToLock.name}</span>.
              </p>
              <p className="text-[#00ff00] mb-4 opacity-90">
                <span className="font-semibold text-red-500" style={{ textShadow: '0 0 10px #ff0000' }}>Warning:</span> You will lose access to this room and this action cannot be undone. Make sure you're ready to pass the room to the next team.
              </p>
              <div className="flex gap-3">
                <button
                  onClick={() => handleToggleRoom(roomToLock.id)}
                  className="flex-1 matrix-button matrix-button-success min-h-[44px]"
                >
                  Unlock Room
                </button>
                <button
                  onClick={() => {
                    setShowLockConfirm(false);
                    setRoomToLock(null);
                  }}
                  className="flex-1 matrix-button matrix-button-secondary min-h-[44px]"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>

          {/* Desktop Modal */}
          <div className="hidden md:flex fixed inset-0 bg-black/80 z-50 items-center justify-center p-4">
            <div className="bg-black/95 backdrop-blur-sm border-2 border-[#00ff00] rounded-lg shadow-[0_0_30px_rgba(0,255,0,0.5)] p-6 max-w-md w-full matrix-card">
              <h2 className="text-xl font-bold text-[#00ff00] mb-4">Confirm Unlock Room</h2>
              <p className="text-[#00ff00] mb-2 opacity-90">
                You are about to unlock <span className="font-semibold">{roomToLock.name}</span>.
              </p>
              <p className="text-[#00ff00] mb-6 opacity-90">
                <span className="font-semibold text-red-500" style={{ textShadow: '0 0 10px #ff0000' }}>Warning:</span> You will lose access to this room and this action cannot be undone. Make sure you're ready to pass the room to the next team.
              </p>
              <div className="flex gap-3">
                <button
                  onClick={() => handleToggleRoom(roomToLock.id)}
                  className="flex-1 matrix-button matrix-button-success min-h-[44px]"
                >
                  Unlock Room
                </button>
                <button
                  onClick={() => {
                    setShowLockConfirm(false);
                    setRoomToLock(null);
                  }}
                  className="flex-1 matrix-button matrix-button-secondary min-h-[44px]"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </>
      )}

      {/* Logout button - bottom aligned */}
      <div className="fixed bottom-4 left-4 right-4 md:left-auto md:right-4 md:w-auto">
        <button
          onClick={() => {
            logout();
            navigate('/login');
          }}
          className="matrix-button matrix-button-secondary w-full md:w-auto min-h-[44px]"
        >
          Logout
        </button>
      </div>
    </div>
  );
};

