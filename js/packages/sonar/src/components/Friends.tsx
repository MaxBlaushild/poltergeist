import React, { useEffect, useState } from 'react';
import { useFriendContext } from '../contexts/FriendContext.tsx';
import { ChevronDownIcon, ChevronRightIcon, MagnifyingGlassIcon } from '@heroicons/react/20/solid';
import { Button, ButtonSize } from './shared/Button.tsx';
import { useAuth } from '@poltergeist/contexts';
import { useUserContext } from '../contexts/UserContext.tsx';

export const Friends: React.FC = () => {
  const { friends, friendInvites, searchResults, fetchFriends, fetchFriendInvites, acceptFriendInvite, searchForFriends, createFriendInvite, deleteFriendInvite } = useFriendContext();
  const { user } = useAuth();
  const [isFriendsOpen, setIsFriendsOpen] = useState(true);
  const [isReceivedInvitesOpen, setIsReceivedInvitesOpen] = useState(true);
  const [isSentInvitesOpen, setIsSentInvitesOpen] = useState(true);
  const [isSearchOpen, setIsSearchOpen] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [acceptingInvites, setAcceptingInvites] = useState<Set<string>>(new Set());
  const [rejectingInvites, setRejectingInvites] = useState<Set<string>>(new Set());
  const [sendingInvites, setSendingInvites] = useState<Set<string>>(new Set());
  const { setUsername } = useUserContext();

  const receivedInvites = friendInvites.filter(invite => invite.inviteeId === user?.id);
  const sentInvites = friendInvites.filter(invite => invite.inviterId === user?.id);

  useEffect(() => {
    fetchFriends();
    fetchFriendInvites();
  }, [fetchFriends, fetchFriendInvites]);

  const handleAcceptInvite = async (inviteId: string) => {
    setAcceptingInvites(prev => new Set(prev).add(inviteId));
    try {
      await acceptFriendInvite(inviteId);
      await fetchFriends(); // Refresh friends list
    } catch (error) {
      console.error('Failed to accept invite:', error);
    } finally {
      setAcceptingInvites(prev => {
        const next = new Set(prev);
        next.delete(inviteId);
        return next;
      });
    }
  };

  const handleRejectInvite = async (inviteId: string) => {
    setRejectingInvites(prev => new Set(prev).add(inviteId));
    try {
      await deleteFriendInvite(inviteId);
    } catch (error) {
      console.error('Failed to reject invite:', error);
    } finally {
      setRejectingInvites(prev => {
        const next = new Set(prev);
        next.delete(inviteId);
        return next;
      });
    }
  };

  const handleSendInvite = async (userId: string) => {
    setSendingInvites(prev => new Set(prev).add(userId));
    try {
      await createFriendInvite(userId);
      await fetchFriendInvites(); // Refresh invites list
    } catch (error) {
      console.error('Failed to send invite:', error);
    } finally {
      setSendingInvites(prev => {
        const next = new Set(prev);
        next.delete(userId);
        return next;
      });
    }
  };

  const handleSearchChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const query = e.target.value;
    setSearchQuery(query);
    if (query.trim()) {
      await searchForFriends(query);
    }
  };

  const isAlreadyFriend = (userId: string) => {
    return friends.some(friend => friend.id === userId);
  };

  const isCurrentUser = (userId: string) => {
    return user?.id === userId;
  };

  return (
    <div className="space-y-4">
      {/* Search Users Accordion */}
      <div className="border border-gray-300 rounded-lg overflow-hidden">
        <button
          className="w-full flex items-center justify-between p-4 bg-gray-50 hover:bg-gray-100 transition-colors"
          onClick={() => setIsSearchOpen(!isSearchOpen)}
        >
          <div className="flex items-center gap-2">
            <MagnifyingGlassIcon className="h-5 w-5" />
            <h3 className="font-bold text-lg">Find Friends</h3>
          </div>
          {isSearchOpen ? (
            <ChevronDownIcon className="h-5 w-5" />
          ) : (
            <ChevronRightIcon className="h-5 w-5" />
          )}
        </button>
        {isSearchOpen && (
          <div className="p-4">
            <div className="relative mb-4">
              <input
                type="text"
                placeholder="Search by username..."
                value={searchQuery}
                onChange={handleSearchChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            {searchQuery && (
              <div className="space-y-3">
                {searchResults?.length === 0 ? (
                  <p className="text-gray-500 text-sm text-center py-4">No users found</p>
                ) : (
                  searchResults?.map((searchUser) => (
                    <div
                      key={searchUser.id}
                      className="flex items-center justify-between p-3 bg-white border border-gray-200 rounded-lg"
                    >
                      <div className="flex items-center gap-3">
                        <div className="relative w-10 h-10">
                          <div className="w-10 h-10 rounded-full overflow-hidden bg-gray-300">
                            <img
                              src={searchUser.profilePictureUrl || '/blank-avatar.webp'}
                              alt={searchUser.username}
                              className="w-full h-full object-cover"
                            />
                          </div>
                          <div className={`absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-white ${searchUser.isActive ? 'bg-green-500' : 'bg-gray-400'}`}></div>
                        </div>
                        <div>
                          <p className="font-medium">{searchUser.username}</p>
                          {searchUser.name && (
                            <p className="text-xs text-gray-500">{searchUser.name}</p>
                          )}
                        </div>
                      </div>
                      <div className="ml-4">
                        {isCurrentUser(searchUser.id) ? (
                          <span className="text-sm text-gray-500">You</span>
                        ) : isAlreadyFriend(searchUser.id) ? (
                          <span className="text-sm text-green-600 font-medium">Friends</span>
                        ) : (
                          <Button
                            title={sendingInvites.has(searchUser.id) ? 'Sending...' : 'Invite Friend'}
                            buttonSize={ButtonSize.SMALL}
                            onClick={() => handleSendInvite(searchUser.id)}
                            disabled={sendingInvites.has(searchUser.id)}
                          />
                        )}
                      </div>
                    </div>
                  ))
                )}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Received Friend Invites Accordion */}
      <div className="border border-gray-300 rounded-lg overflow-hidden">
        <button
          className="w-full flex items-center justify-between p-4 bg-gray-50 hover:bg-gray-100 transition-colors"
          onClick={() => setIsReceivedInvitesOpen(!isReceivedInvitesOpen)}
        >
          <div className="flex items-center gap-2">
            <h3 className="font-bold text-lg">Received Invites</h3>
            {receivedInvites.length > 0 && (
              <span className="bg-blue-500 text-white text-xs font-bold px-2 py-1 rounded-full">
                {receivedInvites.length}
              </span>
            )}
          </div>
          {isReceivedInvitesOpen ? (
            <ChevronDownIcon className="h-5 w-5" />
          ) : (
            <ChevronRightIcon className="h-5 w-5" />
          )}
        </button>
        {isReceivedInvitesOpen && (
          <div className="p-4">
            {receivedInvites.length === 0 ? (
              <p className="text-gray-500 text-sm text-center py-4">No pending invites</p>
            ) : (
              <div className="space-y-3">
                {receivedInvites.map((invite) => (
                  <div
                    key={invite.id}
                    className="flex items-center justify-between p-3 bg-white border border-gray-200 rounded-lg"
                  >
                    <div className="flex items-center gap-3">
                      <div className="relative w-10 h-10">
                        <div className="w-10 h-10 rounded-full bg-gray-300 flex items-center justify-center overflow-hidden">
                          <img src={invite.inviter.profilePictureUrl || '/blank-avatar.webp'} alt={invite.inviter.username} className="w-full h-full object-cover" />
                        </div>
                        <div className={`absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-white ${invite.inviter.isActive ? 'bg-green-500' : 'bg-gray-400'}`}></div>
                      </div>
                      <div>
                        <p className="font-medium text-sm">{invite.inviter.username}</p>
                        <p className="text-xs text-gray-500">
                          {new Date(invite.createdAt).toLocaleDateString()}
                        </p>
                      </div>
                    </div>
                    <div className="flex gap-2 ml-2">
                      <Button
                        title={acceptingInvites.has(invite.id) ? 'Accepting...' : 'Accept'}
                        buttonSize={ButtonSize.SMALL}
                        onClick={() => handleAcceptInvite(invite.id)}
                        disabled={acceptingInvites.has(invite.id) || rejectingInvites.has(invite.id)}
                      />
                      <Button
                        title={rejectingInvites.has(invite.id) ? 'Rejecting...' : 'Reject'}
                        buttonSize={ButtonSize.SMALL}
                        onClick={() => handleRejectInvite(invite.id)}
                        disabled={acceptingInvites.has(invite.id) || rejectingInvites.has(invite.id)}
                      />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Sent Friend Invites Accordion */}
      <div className="border border-gray-300 rounded-lg overflow-hidden">
        <button
          className="w-full flex items-center justify-between p-4 bg-gray-50 hover:bg-gray-100 transition-colors"
          onClick={() => setIsSentInvitesOpen(!isSentInvitesOpen)}
        >
          <div className="flex items-center gap-2">
            <h3 className="font-bold text-lg">Sent Invites</h3>
            <span className="text-gray-600 text-sm">({sentInvites.length})</span>
          </div>
          {isSentInvitesOpen ? (
            <ChevronDownIcon className="h-5 w-5" />
          ) : (
            <ChevronRightIcon className="h-5 w-5" />
          )}
        </button>
        {isSentInvitesOpen && (
          <div className="p-4">
            {sentInvites.length === 0 ? (
              <p className="text-gray-500 text-sm text-center py-4">No sent invites</p>
            ) : (
              <div className="space-y-3">
                {sentInvites.map((invite) => (
                  <div
                    key={invite.id}
                    className="flex items-center justify-between p-3 bg-white border border-gray-200 rounded-lg"
                  >
                    <div className="flex items-center gap-3">
                      <div className="relative w-10 h-10">
                        <div className="w-10 h-10 rounded-full bg-gray-300 flex items-center justify-center overflow-hidden">
                          {invite.invitee.profilePictureUrl ? (
                            <img src={invite.invitee.profilePictureUrl} alt={invite.invitee.username} className="w-full h-full object-cover" />
                          ) : (
                            <span className="text-gray-600 text-sm">👤</span>
                          )}
                        </div>
                        <div className={`absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-white ${invite.invitee.isActive ? 'bg-green-500' : 'bg-gray-400'}`}></div>
                      </div>
                      <div>
                        <p className="font-medium text-sm">{invite.invitee.username}</p>
                        <p className="text-xs text-gray-500">
                          {new Date(invite.createdAt).toLocaleDateString()}
                        </p>
                      </div>
                    </div>
                    <div className="ml-4">
                      <Button
                        title={rejectingInvites.has(invite.id) ? 'Canceling...' : 'Cancel'}
                        buttonSize={ButtonSize.SMALL}
                        onClick={() => handleRejectInvite(invite.id)}
                        disabled={rejectingInvites.has(invite.id)}
                      />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Friends Accordion */}
      <div className="border border-gray-300 rounded-lg overflow-hidden">
        <button
          className="w-full flex items-center justify-between p-4 bg-gray-50 hover:bg-gray-100 transition-colors"
          onClick={() => setIsFriendsOpen(!isFriendsOpen)}
        >
          <div className="flex items-center gap-2">
            <h3 className="font-bold text-lg">Friends</h3>
            <span className="text-gray-600 text-sm">({friends.length})</span>
          </div>
          {isFriendsOpen ? (
            <ChevronDownIcon className="h-5 w-5" />
          ) : (
            <ChevronRightIcon className="h-5 w-5" />
          )}
        </button>
        {isFriendsOpen && (
          <div className="p-4">
            {friends.length === 0 ? (
              <p className="text-gray-500 text-sm text-center py-4">No friends yet</p>
            ) : (
              <div className="space-y-3">
                {friends.map((friend) => (
                  <div
                    key={friend.id}
                    className="flex items-center gap-3 p-3 bg-white border border-gray-200 rounded-lg"
                    onClick={() => setUsername(friend.username)}
                  >
                    <div className="relative w-10 h-10">
                      <div className="w-10 h-10 rounded-full overflow-hidden bg-gray-300">
                        <img
                          src={friend.profilePictureUrl || '/blank-avatar.webp'}
                          alt={friend.username}
                          className="w-full h-full object-cover"
                        />
                      </div>
                      <div className={`absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-white ${friend.isActive ? 'bg-green-500' : 'bg-gray-400'}`}></div>
                    </div>
                    <div className="flex-1">
                      <p className="font-medium">{friend.username}</p>
                      {friend.name && (
                        <p className="text-sm text-gray-500">{friend.name}</p>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};