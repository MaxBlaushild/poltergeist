import { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { User } from '@poltergeist/types';

export interface UseCurrentUserResult {
  currentUser: User | null;
  currentUserLoading: boolean;
  currentUserError: Error | null;
}

export const useCurrentUser = (): UseCurrentUserResult => {
  const { apiClient } = useAPI();
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [currentUserLoading, setCurrentUserLoading] = useState<boolean>(true);
  const [currentUserError, setCurrentUserError] = useState<Error | null>(null);

  useEffect(() => {
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

    fetchCurrentUser();
  }, [apiClient]);

  return {
    currentUser,
    currentUserLoading,
    currentUserError,
  };
};
