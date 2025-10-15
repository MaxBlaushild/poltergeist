import React, { useState } from 'react';
import { useParty } from '../contexts/PartyContext.tsx';
import { useAuth } from '@poltergeist/contexts';
import { User } from '@poltergeist/types';
import { 
  UserIcon, 
  ArrowRightOnRectangleIcon,
  CheckIcon,
  XMarkIcon,
  ChevronDownIcon,
  ChevronRightIcon
} from '@heroicons/react/24/outline';

export const Party: React.FC = () => {
  const { party, partyInvites, loading, setLeader, leaveParty, acceptPartyInvite, rejectPartyInvite } = useParty();
  const { user: currentUser } = useAuth();
  const [promotingMember, setPromotingMember] = useState<string | null>(null);
  const [isInvitesOpen, setIsInvitesOpen] = useState(true);
  const [isSentInvitesOpen, setIsSentInvitesOpen] = useState(true);

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="animate-spin rounded-full h-12 w-12 border-t-4 border-b-4 border-purple-500"></div>
      </div>
    );
  }

  const isLeader = party?.leaderId === currentUser?.id;
  const receivedInvites = partyInvites.filter(invite => invite.inviteeId === currentUser?.id);
  const sentInvites = partyInvites.filter(invite => invite.inviterId === currentUser?.id);

  const handlePromoteToLeader = async (member: User) => {
    if (!isLeader || !party) return;
    setPromotingMember(member.id);
    try {
      await setLeader(member);
    } catch (error) {
      console.error('Failed to promote member:', error);
    } finally {
      setPromotingMember(null);
    }
  };

  const handleLeaveParty = async () => {
    if (window.confirm(isLeader ? 'Are you sure you want to leave the party? You are the leader!' : 'Are you sure you want to leave the party?')) {
      await leaveParty();
    }
  };

  const handleAcceptInvite = async (inviteId: string) => {
    try {
      await acceptPartyInvite(inviteId);
    } catch (error) {
      console.error('Failed to accept invite:', error);
    }
  };

  const handleRejectInvite = async (inviteId: string) => {
    try {
      await rejectPartyInvite(inviteId);
    } catch (error) {
      console.error('Failed to reject invite:', error);
    }
  };

  return (
    <div className="p-4 space-y-4">
      {receivedInvites.length > 0 && (
        <div className="bg-gradient-to-br from-yellow-900 to-yellow-800 rounded-lg border-2 border-yellow-600 shadow-lg overflow-hidden">
          <button
            className="w-full flex items-center justify-between p-3 hover:bg-yellow-700/30 transition-colors"
            onClick={() => setIsInvitesOpen(!isInvitesOpen)}
          >
            <div className="flex items-center gap-2">
              {isInvitesOpen ? (
                <ChevronDownIcon className="h-5 w-5 text-yellow-300" />
              ) : (
                <ChevronRightIcon className="h-5 w-5 text-yellow-300" />
              )}
              <span className="font-bold text-yellow-100">Party Invites</span>
            </div>
            <span className="bg-yellow-600 text-white px-2 py-1 rounded-full text-xs font-bold">
              {receivedInvites.length}
            </span>
          </button>
          
          {isInvitesOpen && (
            <div className="p-3 space-y-2">
              {receivedInvites.map((invite) => (
                <div
                  key={invite.id}
                  className="bg-black/30 rounded-lg p-3 flex items-center gap-3"
                >
                  <div className="w-10 h-10 rounded-full overflow-hidden bg-gray-700 flex-shrink-0">
                    {invite.inviter.profilePictureUrl ? (
                      <img
                        src={invite.inviter.profilePictureUrl}
                        alt={invite.inviter.username}
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center">
                        <UserIcon className="h-6 w-6 text-gray-400" />
                      </div>
                    )}
                  </div>
                  <div className="flex-1">
                    <p className="text-yellow-100 font-semibold">{invite.inviter.username}</p>
                    <p className="text-yellow-300 text-xs">invited you to their party</p>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => handleAcceptInvite(invite.id)}
                      className="bg-green-600 hover:bg-green-700 text-white p-2 rounded transition-colors"
                      title="Accept"
                    >
                      <CheckIcon className="h-4 w-4" />
                    </button>
                    <button
                      onClick={() => handleRejectInvite(invite.id)}
                      className="bg-red-600 hover:bg-red-700 text-white p-2 rounded transition-colors"
                      title="Decline"
                    >
                      <XMarkIcon className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Current Party */}
      {party ? (
        <div className="bg-gradient-to-br from-gray-900 to-gray-800 rounded-lg border-2 border-gray-700 shadow-lg overflow-hidden">
          <div className="bg-gradient-to-r from-purple-900 to-blue-900 p-3 border-b-2 border-purple-700">
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-bold text-white flex items-center gap-2">
                <UserIcon className="h-6 w-6 text-purple-300" />
                Party ({party.members.length}/5)
              </h2>
              {isLeader && (
                <div className="flex items-center gap-1 bg-yellow-600 px-2 py-1 rounded">
                  {/* <CrownSolid className="h-4 w-4 text-white" /> */}
                  <span className="text-white text-xs font-bold">LEADER</span>
                </div>
              )}
            </div>
          </div>

          <div className="p-3 space-y-2">
            {party.members.map((member) => {
              const isMemberLeader = member.id === party.leaderId;
              const isCurrentUser = member.id === currentUser?.id;
              
              return (
                <div
                  key={member.id}
                  className={`rounded-lg p-3 flex items-center gap-3 ${
                    isMemberLeader
                      ? 'bg-gradient-to-r from-yellow-900/40 to-yellow-800/40 border-2 border-yellow-600/50'
                      : 'bg-black/30 border border-gray-700'
                  }`}
                >
                  {/* Profile Picture */}
                  <div className="relative">
                    <div className="w-12 h-12 rounded-lg overflow-hidden bg-gray-700 border-2 border-gray-600">
                      {member.profilePictureUrl ? (
                        <img
                          src={member.profilePictureUrl}
                          alt={member.username}
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        <div className="w-full h-full flex items-center justify-center">
                          <UserIcon className="h-7 w-7 text-gray-400" />
                        </div>
                      )}
                    </div>
                    {isMemberLeader && (
                      <div className="absolute -top-1 -right-1 bg-yellow-500 rounded-full p-1">
                        {/* <CrownSolid className="h-3 w-3 text-white" /> */}
                      </div>
                    )}
                  </div>

                  {/* Member Info */}
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <p className={`font-bold ${isMemberLeader ? 'text-yellow-300' : 'text-white'}`}>
                        {member.username}
                        {isCurrentUser && <span className="text-gray-400 text-sm ml-1">(You)</span>}
                      </p>
                    </div>
                    <p className="text-gray-400 text-xs">{member.name}</p>
                    {/* Health bar aesthetic */}
                    <div className="mt-1 w-full h-1.5 bg-gray-700 rounded-full overflow-hidden">
                      <div className="h-full bg-gradient-to-r from-green-500 to-green-400 w-full"></div>
                    </div>
                  </div>

                  {/* Actions */}
                  {isLeader && !isCurrentUser && !isMemberLeader && (
                    <button
                      onClick={() => handlePromoteToLeader(member)}
                      disabled={promotingMember === member.id}
                      className="bg-yellow-600 hover:bg-yellow-700 disabled:bg-gray-600 text-white px-3 py-1 rounded text-sm font-semibold transition-colors flex items-center gap-1"
                    >
                      {/* <CrownIcon className="h-4 w-4" /> */}
                      Promote
                    </button>
                  )}
                </div>
              );
            })}
          </div>

          {/* Party Actions */}
          <div className="p-3 border-t border-gray-700 bg-black/20">
            <button
              onClick={handleLeaveParty}
              className="w-full bg-red-600 hover:bg-red-700 text-white font-bold py-2 px-4 rounded transition-colors flex items-center justify-center gap-2"
            >
              <ArrowRightOnRectangleIcon className="h-5 w-5" />
              Leave Party
            </button>
          </div>
        </div>
      ) : (
        <div className="bg-gradient-to-br from-gray-900 to-gray-800 rounded-lg border-2 border-gray-700 shadow-lg p-8 text-center">
          <UserIcon className="h-16 w-16 text-gray-600 mx-auto mb-4" />
          <h3 className="text-xl font-bold text-gray-400 mb-2">No Active Party</h3>
          <p className="text-gray-500">You are not currently in a party. Accept an invite or get invited by a friend!</p>
        </div>
      )}

      {/* Sent Invites */}
      {sentInvites.length > 0 && (
        <div className="bg-gradient-to-br from-gray-900 to-gray-800 rounded-lg border-2 border-gray-700 shadow-lg overflow-hidden">
          <button
            className="w-full flex items-center justify-between p-3 hover:bg-gray-700/30 transition-colors"
            onClick={() => setIsSentInvitesOpen(!isSentInvitesOpen)}
          >
            <div className="flex items-center gap-2">
              {isSentInvitesOpen ? (
                <ChevronDownIcon className="h-5 w-5 text-gray-400" />
              ) : (
                <ChevronRightIcon className="h-5 w-5 text-gray-400" />
              )}
              <span className="font-bold text-gray-300">Pending Invites</span>
            </div>
            <span className="bg-gray-600 text-white px-2 py-1 rounded-full text-xs font-bold">
              {sentInvites.length}
            </span>
          </button>
          
          {isSentInvitesOpen && (
            <div className="p-3 space-y-2">
              {sentInvites.map((invite) => (
                <div
                  key={invite.id}
                  className="bg-black/30 rounded-lg p-3 flex items-center gap-3"
                >
                  <div className="w-10 h-10 rounded-full overflow-hidden bg-gray-700 flex-shrink-0">
                    {invite.invitee.profilePictureUrl ? (
                      <img
                        src={invite.invitee.profilePictureUrl}
                        alt={invite.invitee.username}
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center">
                        <UserIcon className="h-6 w-6 text-gray-400" />
                      </div>
                    )}
                  </div>
                  <div className="flex-1">
                    <p className="text-gray-300 font-semibold">{invite.invitee.username}</p>
                    <p className="text-gray-500 text-xs">Invite pending...</p>
                  </div>
                  <button
                    onClick={() => handleRejectInvite(invite.id)}
                    className="bg-gray-600 hover:bg-gray-700 text-white px-3 py-1 rounded text-sm transition-colors"
                  >
                    Cancel
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
};