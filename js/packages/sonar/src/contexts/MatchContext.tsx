import React, { createContext, useContext, useState, useCallback, useEffect } from 'react';
import { useAPI, useAuth } from '@poltergeist/contexts';
import { Match } from '@poltergeist/types';

interface MatchContextType {
  match: Match | null;

  createMatch: (pointsOfInterestIds: string[]) => Promise<void>;
  getMatch: (matchId: string) => Promise<void>;
  areMatchesLoading: boolean;
  isCreatingMatch: boolean;
  isInviting: boolean;
  getCurrentMatch: () => Promise<void>;
}

export const MatchContext = createContext<MatchContextType | undefined>(undefined);

export const useMatchContext = () => {
  const context = useContext(MatchContext);
  if (!context) {
    throw new Error('useMatchContext must be used within a MatchContextProvider');
  }
  return context;
};

export const MatchContextProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { apiClient } = useAPI();
  const [match, setMatch] = useState<Match | null>(null);
  const [areMatchesLoading, setAreMatchesLoading] = useState(false);
  const [isCreatingMatch, setIsCreatingMatch] = useState(false);
  const [isInviting, setIsInviting] = useState(false);
  const { user } = useAuth();
  const userID = user?.id;

  useEffect(() => {
    getCurrentMatch();
  }, [userID]);

  const getCurrentMatch = useCallback(async () => {
    const response = await apiClient.get<Match>('/sonar/matches/current');
    setMatch(response);
  }, [apiClient, setMatch]);

  const createMatch = useCallback(async (pointsOfInterestIds: string[]) => {
    try {
      setIsCreatingMatch(true);
      const response = await apiClient.post<Match>('/sonar/matches', {
        pointsOfInterestIds
      });
      getCurrentMatch();
    } catch (error) {
      console.error('Failed to create match', error);
    } finally {
      setIsCreatingMatch(false);
    }
  }, [apiClient, setMatch, setIsCreatingMatch, getCurrentMatch]);

  const getMatch = useCallback(async (matchId: string) => {
    try {
      const response = await apiClient.get<Match>(`/sonar/matches/${matchId}`);
      setMatch(response);
    } catch (error) {
      console.error('Failed to get match', error);
    }
  }, [apiClient, setMatch]);

  return (
    <MatchContext.Provider value={{
      match,
      createMatch,
      getMatch,
      getCurrentMatch,
      areMatchesLoading,
      isCreatingMatch,
      isInviting
    }}>
      {children}
    </MatchContext.Provider>
  );
};
