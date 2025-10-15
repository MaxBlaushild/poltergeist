import React, { createContext, useContext, useEffect, useState, useCallback } from "react";
import { Party, PartyInvite, User } from "@poltergeist/types";
import { useAPI } from "@poltergeist/contexts";

interface PartyContextType {
  party: Party | null;
  partyInvites: PartyInvite[];
  setParty: (party: Party) => void;
  loading: boolean;
  error: Error | null;
  setLeader: (leader: User) => void;
  leaveParty: () => void;
  fetchParty: () => void;
  inviteToParty: (invitee: User) => void;
  fetchPartyInvites: () => void;
  acceptPartyInvite: (inviteId: string) => void;
  rejectPartyInvite: (inviteId: string) => void;
}

export const PartyContext = createContext<PartyContextType | undefined>(undefined);

export const PartyProvider = ({ children }: { children: React.ReactNode }) => {
  const [party, setParty] = useState<Party | null>(null);
  const [partyInvites, setPartyInvites] = useState<PartyInvite[]>([]);
  const { apiClient } = useAPI();
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchParty = useCallback(async () => {
    try {
      const party = await apiClient.get<Party>('/sonar/party');
      setParty(party);
    } catch (error) {
      setError(error as Error);
    } finally {
      setLoading(false);
    }
  }, [apiClient]);

  const fetchPartyInvites = useCallback(async () => {
    const response = await apiClient.get<PartyInvite[]>(`/sonar/partyInvites`);
    setPartyInvites(response);
  }, [apiClient]);

  useEffect(() => {
    fetchParty();
    fetchPartyInvites();

    const pollInterval = setInterval(() => {
      fetchParty();
      fetchPartyInvites();
    }, 5000);

    return () => clearInterval(pollInterval);
  }, [fetchParty, fetchPartyInvites]);

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
      fetchPartyInvites();
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

  const acceptPartyInvite = async (inviteId: string) => {
    try {
      await apiClient.post('/sonar/partyInvites/accept', { inviteID: inviteId });
      setPartyInvites(partyInvites.filter((invite) => invite.id !== inviteId));
      fetchParty();
    } catch (error) {
      setError(error as Error);
    }
  };

  const rejectPartyInvite = async (inviteId: string) => {
    try {
      await apiClient.post(`/sonar/partyInvites/reject`, { inviteID: inviteId });
      setPartyInvites(partyInvites.filter((invite) => invite.id !== inviteId));
    } catch (error) {
      setError(error as Error);
    }
  };

  return (
    <PartyContext.Provider value={{ 
      party, 
      partyInvites,
      setParty, 
      loading, 
      error, 
      setLeader, 
      leaveParty, 
      fetchParty,
      inviteToParty,
      fetchPartyInvites,
      acceptPartyInvite,
      rejectPartyInvite
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