import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
} from 'react';
import { useAPI, useAuth } from '@poltergeist/contexts';
import { AuditItem, InventoryItem, Match, Team } from '@poltergeist/types';
import { useUserProfiles } from './UserProfileContext.tsx';
import { useMediaContext } from '@poltergeist/contexts';
import { PointOfInterestChallengeSubmission } from '@poltergeist/types/dist/pointOfInterestChallengeSubmission';
import { useSearchParams } from 'react-router-dom';

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
  addUserToMatch: (userId: string) => Promise<void>;
  editTeamName: (teamId: string, name: string) => Promise<void>;
  usersTeam: Team | undefined;
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
  const { uploadMedia, getPresignedUploadURL } = useMediaContext();
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
  const usersTeam = match?.teams.find((team) => team.users.some((user) => user.id === userID));

  const createTeam = useCallback(async () => {
    if (!match) return;
    try {
      setIsCreatingTeam(true);
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

  const addUserToMatch = useCallback(async (userId: string) => {
    await apiClient.post(`/sonar/matches/${match?.id}/invite`, {
      userId,
    });
    getCurrentMatch();
  }, [apiClient, match?.id]);

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
      const matchId = new URLSearchParams(window.location.search).get('matchId');
      const response = await apiClient.get<Match>(`/sonar/matches/${matchId ? matchId + '/users' : 'current'}?timestamp=${new Date().getTime()}`);
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
      window.location.href = '/match/in-progress';
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

  const editTeamName = useCallback(async (teamId: string, name: string) => {
    await apiClient.post(`/sonar/teams/${teamId}/edit`, {
      name,
    });
    getCurrentMatch();
  }, [apiClient, getCurrentMatch]);

  useEffect(() => {
    getCurrentMatch();
  }, []);

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
        usersTeam,
        editTeamName,
        addUserToMatch,
      }}
    >
      {children}
    </MatchContext.Provider>
  );
};
