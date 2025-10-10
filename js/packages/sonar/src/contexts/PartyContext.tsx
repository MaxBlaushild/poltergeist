import React, { createContext, useContext, useEffect, useState } from "react";
import { Party, User } from "@poltergeist/types";
import { useAPI } from "@poltergeist/contexts";

interface PartyContextType {
  party: Party | null;
  setParty: (party: Party) => void;
  loading: boolean;
  error: Error | null;
  setLeader: (leader: User) => void;
  leaveParty: () => void;
  fetchParty: () => void;
  inviteToParty: (invitee: User) => void;
}

export const PartyContext = createContext<PartyContextType | undefined>(undefined);

export const PartyProvider = ({ children }: { children: React.ReactNode }) => {
  const [party, setParty] = useState<Party | null>(null);
  const { apiClient } = useAPI();
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchParty = async () => {
    try {
      const party = await apiClient.get<Party>('/sonar/party');
      setParty(party);
    } catch (error) {
      setError(error as Error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchParty();
  }, [apiClient]);

  const setLeader = async (leader: User) => {
    try {
      await apiClient.post<Party>('/sonar/party/leader', { leader });
    } catch (error) {
      setError(error as Error);
    } finally {
      setLoading(false);
    }
    fetchParty();
  };

  const inviteToParty = async (invitee: User) => {
    try {
      await apiClient.post('/sonar/partyInvites', { inviteeID: invitee.id });
    } catch (error) {
      setError(error as Error);
    } finally {
      setLoading(false);
    }
  };

  const leaveParty = async () => {
    try {
      await apiClient.post('/sonar/party/leave');
    } catch (error) {
      setError(error as Error);
    } finally {
      setLoading(false);
    }
    setParty(null);
  };

  return (
    <PartyContext.Provider value={{ 
      party, 
      setParty, 
      loading, 
      error, 
      setLeader, 
      leaveParty, 
      fetchParty,
      inviteToParty
    }}>
      {children}
    </PartyContext.Provider>
  );
}

export const useParty = () => {
  const context = useContext(PartyContext);
  if (!context) {
    throw new Error('useParty must be used within a PartyProvider');
  }
  return context;
}