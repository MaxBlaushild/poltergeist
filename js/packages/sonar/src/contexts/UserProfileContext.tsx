import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { UserProfile } from '@poltergeist/types';

export interface UserProfileContextType {
  userProfiles: UserProfile[] | null;
  loading: boolean;
  error: Error | null;
}

const UserProfileContext = createContext<UserProfileContextType>({
  userProfiles: null,
  loading: false,
  error: null
});

export const UserProfileProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const { apiClient } = useAPI();
  const [userProfiles, setUserProfiles] = useState<UserProfile[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchUserProfiles = async () => {
      try {
        const fetchedUserProfiles = await apiClient.get<UserProfile[]>('/sonar/userProfiles');
        setUserProfiles(fetchedUserProfiles);
      } catch (err) {
        setError(err);
      } finally {
        setLoading(false);
      }
    };

    fetchUserProfiles();
  }, [apiClient]);

  return (
    <UserProfileContext.Provider value={{ userProfiles, loading, error }}>
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
