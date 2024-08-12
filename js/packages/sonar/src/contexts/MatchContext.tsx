import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
} from 'react';
import { useAPI, useAuth } from '@poltergeist/contexts';
import { Match, Team } from '@poltergeist/types';
import { useUserProfiles } from './UserProfileContext.tsx';

interface MatchContextType {
  match: Match | null;

  createMatch: (pointsOfInterestIds: string[]) => Promise<void>;
  getMatch: (matchId: string) => Promise<void>;
  isCreatingMatch: boolean;
  isCurrentMatchLoading: boolean;
  getCurrentMatch: () => Promise<void>;
  addUserToTeam: (teamId: string) => Promise<void>;
  createTeam: () => Promise<void>;
  startMatch: () => Promise<void>;
  isStartingMatch: boolean;
  leaveMatch: () => Promise<void>;
  isLeavingMatch: boolean;
  leaveMatchError: string | null;
  unlockPointOfInterest: (
    pointOfInterestId: string,
    teamId: string,
    lat: string,
    lng: string
  ) => Promise<void>;
}

export const MatchContext = createContext<MatchContextType | undefined>(
  undefined
);

export const useMatchContext = () => {
  const context = useContext(MatchContext);
  if (!context) {
    throw new Error(
      'useMatchContext must be used within a MatchContextProvider'
    );
  }
  return context;
};

export const MatchContextProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const { apiClient } = useAPI();
  const [match, setMatch] = useState<Match | null>(null);
  const [isCreatingMatch, setIsCreatingMatch] = useState(false);
  const { currentUser } = useUserProfiles();
  const [isCurrentMatchLoading, setIsCurrentMatchLoading] = useState(true);
  const [isCreatingTeam, setIsCreatingTeam] = useState(false);
  const [isJoiningTeam, setIsJoiningTeam] = useState(false);
  const [isStartingMatch, setIsStartingMatch] = useState(false);
  const [createTeamError, setCreateTeamError] = useState<string | null>(null);
  const [joinTeamError, setJoinTeamError] = useState<string | null>(null);
  const [leaveMatchError, setLeaveMatchError] = useState<string | null>(null);
  const [isLeavingMatch, setIsLeavingMatch] = useState(false);
  const userID = currentUser?.id;

  const createTeam = useCallback(async () => {
    if (!match) return;
    try {
      setIsCreatingTeam(true);
      console.log('userID', userID);
      const response = await apiClient.post<Team>(
        `/sonar/matches/${match?.id}/teams`,
        {
          userId: userID,
        }
      );
      getCurrentMatch();
    } catch (error) {
      console.error('Failed to create team', error);
    } finally {
      setIsCreatingTeam(false);
    }
  }, [apiClient, match?.id, userID]);

  const addUserToTeam = useCallback(
    async (teamId: string) => {
      if (!match) return;
      try {
        setIsJoiningTeam(true);
        const response = await apiClient.post<Team>(`/sonar/teams/${teamId}`, {
          userId: userID,
        });
        getCurrentMatch();
      } catch (error) {
        console.error('Failed to add user to team', error);
      } finally {
        setIsJoiningTeam(false);
      }
    },
    [apiClient, match?.id, userID]
  );

  const getCurrentMatch = useCallback(async () => {
    setIsCurrentMatchLoading(true);
    try {
      const response = await apiClient.get<Match>('/sonar/matches/current');
      setMatch(response);
      setIsCurrentMatchLoading(false);
    } catch (error) {
      setMatch(null);
      console.error('Failed to get current match', error);
    } finally {
      setIsCurrentMatchLoading(false);
    }
  }, [apiClient, setMatch]);

  const createMatch = useCallback(
    async (pointsOfInterestIds: string[]) => {
      try {
        setIsCreatingMatch(true);
        const response = await apiClient.post<Match>('/sonar/matches', {
          pointsOfInterestIds,
        });
        getCurrentMatch();
      } catch (error) {
        console.error('Failed to create match', error);
      } finally {
        setIsCreatingMatch(false);
      }
    },
    [apiClient, setIsCreatingMatch, getCurrentMatch]
  );

  const getMatch = useCallback(
    async (matchId: string) => {
      try {
        const response = await apiClient.get<Match>(
          `/sonar/matchesById/${matchId}`
        );
        setMatch(response);
      } catch (error) {
        console.error('Failed to get match', error);
      }
    },
    [apiClient, setMatch]
  );

  const startMatch = useCallback(async () => {
    try {
      await apiClient.post(`/sonar/matches/${match?.id}/start`);
      getCurrentMatch();
    } catch (error) {
      console.error('Failed to start match', error);
    }
  }, [apiClient, match?.id, getCurrentMatch]);

  const leaveMatch = useCallback(async () => {
    try {
      await apiClient.post(`/sonar/matches/${match?.id}/leave`);
      getCurrentMatch();
    } catch (error) {
      console.error('Failed to leave match', error);
    }
  }, [apiClient, match?.id, getCurrentMatch]);

  const unlockPointOfInterest = useCallback(
    async (
      pointOfInterestId: string,
      teamId: string,
      lat: string,
      lng: string
    ) => {
      await apiClient.post(`/sonar/pointOfInterest/unlock`, {
        pointOfInterestId,
        teamId,
        lat,
        lng,
      });
      getCurrentMatch();
    },
    [apiClient, userID]
  );

  return (
    <MatchContext.Provider
      value={{
        match,
        createMatch,
        getMatch,
        getCurrentMatch,
        isCreatingMatch,
        isCurrentMatchLoading,
        createTeam,
        addUserToTeam,
        startMatch,
        isStartingMatch,
        leaveMatch,
        isLeavingMatch,
        leaveMatchError,
        unlockPointOfInterest,
      }}
    >
      {children}
    </MatchContext.Provider>
  );
};
