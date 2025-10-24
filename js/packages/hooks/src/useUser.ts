import { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { User, UserLevel } from '@poltergeist/types';

export interface UseUserResult {
  user: User | null;
  loading: boolean;
  error: Error | null;
}

export const useUser = (username: string | null): UseUserResult => {
  const { apiClient } = useAPI();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchUsers = async () => {
      if (!username) {
        setUser(null);
        setLoading(false);
        return;
      }
      
      try {
        const fetchedUser = await apiClient.get<User>(`/sonar/users/byUsername/${username}`);
        setUser(fetchedUser);
      } catch (error) {
        setError(error as Error);
      } finally {
        setLoading(false);
      }
    };

    fetchUsers();
  }, [apiClient, username]);

  return {
    user,
    loading,
    error,
  };
};
