import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useAPI, useAuth } from '@poltergeist/contexts';
import { User, UserProfile } from '@poltergeist/types';

export interface UserProfileContextType {
  currentUser: User | null;
  currentUserLoading: boolean;
  currentUserError: Error | null;
}

const UserProfileContext = createContext<UserProfileContextType>({
  currentUser: null,
  currentUserLoading: true,
  currentUserError: null
});

export const UserProfileProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const { apiClient } = useAPI();
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [currentUserLoading, setCurrentUserLoading] = useState<boolean>(true);
  const [currentUserError, setCurrentUserError] = useState<Error | null>(null);
  const { user } = useAuth();

    const fetchCurrentUser = async () => {
      try {
        const fetchedUser = await apiClient.get<User>('/sonar/whoami');
        setCurrentUser(fetchedUser);
      } catch (error) {
        setCurrentUserError(error);
      } finally {
        setCurrentUserLoading(false);
      }
    };

  useEffect(() => {
    fetchCurrentUser();
  }, [apiClient, user]);

  return (
    <UserProfileContext.Provider value={{ currentUser, currentUserLoading, currentUserError }}>
      {children}
    </UserProfileContext.Provider>
  );
};

export const useUserProfiles = () => {
  const context = useContext(UserProfileContext);
  if (!context) {
    throw new Error('useUserProfiles must be used within a UserProfileProvider');
  }
  return context;
};
