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
import { useMatchContext } from './MatchContext.tsx';


interface LogContextType {
  auditItems: AuditItem[];
  fetchAuditItems: () => Promise<void>;
}

interface LogContextProviderProps {
  children: React.ReactNode;
}

export const LogContext = createContext<LogContextType | undefined>(
  undefined
);

export const useLogContext = () => {
  const context = useContext(LogContext);
  if (!context) {
    throw new Error(
      'useLogContext must be used within a LogContextProvider'
    );
  }
  return context;
};

export const LogContextProvider: React.FC<LogContextProviderProps> = ({
  children,
}) => {
  const { apiClient } = useAPI();
  const { currentUser } = useUserProfiles();
  const [auditItems, setAuditItems] = useState<AuditItem[]>([]);
  const { match } = useMatchContext();

  const fetchAuditItems = useCallback(async () => {
    try {
      const response = await apiClient.get<AuditItem[]>(
        `/sonar/chat${match?.id ? `?matchId=${match.id}` : ''}`
      );
      setAuditItems(response);
    } catch (error) {
      console.error('Failed to fetch audit items:', error);
    }
  }, [apiClient, match?.id]);

  useEffect(() => {
    fetchAuditItems();
  }, []);

  return (
    <LogContext.Provider
      value={{
        auditItems,
        fetchAuditItems,
      }}
    >
      {children}
    </LogContext.Provider>
  );
};
