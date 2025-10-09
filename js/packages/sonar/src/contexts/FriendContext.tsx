import React, { createContext, useCallback, useContext, useState } from "react";
import { FriendInvite, User } from "@poltergeist/types";
import { useAPI } from "@poltergeist/contexts";

interface FriendContextType {
  friends: User[];
  friendInvites: FriendInvite[];
  searchResults: User[];
  fetchFriends: () => Promise<void>;
  fetchFriendInvites: () => Promise<void>;
  createFriendInvite: (inviteeID: string) => Promise<void>;
  acceptFriendInvite: (inviteID: string) => Promise<void>;
  searchForFriends: (query: string) => Promise<void>;
  deleteFriendInvite: (inviteID: string) => Promise<void>;
}

interface FriendContextProviderProps {
  children: React.ReactNode;
}

export const FriendContext = createContext<FriendContextType | undefined>(undefined);

export const FriendContextProvider: React.FC<FriendContextProviderProps> = ({ children }) => {
  const { apiClient } = useAPI();
  const [friends, setFriends] = useState<User[]>([]);
  const [friendInvites, setFriendInvites] = useState<FriendInvite[]>([]);
  const [searchResults, setSearchResults] = useState<User[]>([]);

  const searchForFriends = useCallback(async (query: string) => {
    if (!query.trim()) {
      setSearchResults([]);
      return;
    }
    const response = await apiClient.get<User[]>(`/sonar/users/search?query=${query}`);
    const filteredResults = response.filter(user => user.id && user.id !== '00000000-0000-0000-0000-000000000000');
    setSearchResults(filteredResults);
  }, [apiClient]);

  const fetchFriendInvites = useCallback(async () => {
    const response = await apiClient.get<FriendInvite[]>(`/sonar/friendInvites`);
    setFriendInvites(response);
  }, [apiClient]);

  const createFriendInvite = useCallback(async (inviteeID: string) => {
    const response = await apiClient.post<FriendInvite>(`/sonar/friendInvites/create`, { inviteeID });
    setFriendInvites([...friendInvites, response]);
  }, [apiClient, friendInvites]);

  const acceptFriendInvite = useCallback(async (inviteID: string) => {
    const response = await apiClient.post<FriendInvite>(`/sonar/friendInvites/accept`, { inviteID });
    setFriendInvites(friendInvites.filter((invite) => invite.id !== inviteID));
  }, [apiClient, friendInvites]);

  const fetchFriends = useCallback(async () => {
    const response = await apiClient.get<User[]>(`/sonar/friends`);
    setFriends(response);
  }, [apiClient]);

  const deleteFriendInvite = useCallback(async (inviteID: string) => {
    await apiClient.delete(`/sonar/friendInvites/${inviteID}`);
    setFriendInvites(friendInvites.filter((invite) => invite.id !== inviteID));
  }, [apiClient, friendInvites]);

  return <FriendContext.Provider value={{ 
    friends, 
    friendInvites, 
    fetchFriendInvites, 
    createFriendInvite, 
    fetchFriends,
    acceptFriendInvite,
    searchForFriends,
    searchResults,
    deleteFriendInvite
  }}>
    {children}
  </FriendContext.Provider>;
};

export const useFriendContext = () => {
  const context = useContext(FriendContext);
  if (!context) {
    throw new Error('useFriendContext must be used within a FriendContextProvider');
  }
  return context;
};  