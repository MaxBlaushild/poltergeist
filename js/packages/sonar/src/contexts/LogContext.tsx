import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
} from 'react';
import { useAPI, useAuth } from '@poltergeist/contexts';
import { AuditItem } from '@poltergeist/types';
import { useUserProfiles } from './UserProfileContext.tsx';

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
  const { user } = useAuth();
  const { currentUser } = useUserProfiles();
  const [auditItems, setAuditItems] = useState<AuditItem[]>([]);

  const fetchAuditItems = useCallback(async () => {
    try {
      const response = await apiClient.get<AuditItem[]>(
        `/sonar/chat`
      );
      setAuditItems(response);
    } catch (error: any) {
      // Silently handle auth errors
      if (error?.response?.status === 401 || error?.response?.status === 403) {
        setAuditItems([]);
        return;
      }
      console.error('Failed to fetch audit items:', error);
    }
  }, [apiClient]);

  useEffect(() => {
    if (!user) {
      // Clear data when not authenticated
      setAuditItems([]);
      return;
    }

    fetchAuditItems();
  }, [fetchAuditItems, user]);

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
