import { useState, useEffect } from 'react';
import { useAPI, useAuth } from '@poltergeist/contexts';
import { PointOfInterest, UserLevel } from '@poltergeist/types';

export interface UseUserLevelResult {
  userLevel: UserLevel | null;
  loading: boolean;
  error: Error | null;
}

export const useUserLevel = (): UseUserLevelResult => {
  const { apiClient } = useAPI();
  const { user } = useAuth();
  const [userLevel, setUserLevel] = useState<UserLevel | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!user) {
      // Clear data when not authenticated
      setUserLevel(null);
      setLoading(false);
      return;
    }

    const fetchUserLevel = async () => {
      try {
        const fetchedUserLevel = await apiClient.get<UserLevel>(`/sonar/level`);
        setUserLevel(fetchedUserLevel);
      } catch (error: any) {
        // Silently handle auth errors
        if (error?.response?.status === 401 || error?.response?.status === 403) {
          setUserLevel(null);
          return;
        }
        setError(error as Error);
      } finally {
        setLoading(false);
      }
    };

    fetchUserLevel();
  }, [apiClient, user]);

  return {
    userLevel,
    loading,
    error,
  };
};
