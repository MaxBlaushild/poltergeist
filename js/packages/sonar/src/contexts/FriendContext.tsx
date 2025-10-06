interface FriendContextType {
  friendInvites: FriendInvite[];
  fetchFriendInvites: () => Promise<void>;
  createFriendInvite: (inviteeID: string) => Promise<void>;
  deleteFriendInvite: (inviteID: string) => Promise<void>;
}

export const FriendContext = createContext<FriendContextType | undefined>(undefined);