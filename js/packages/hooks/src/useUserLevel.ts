import { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { PointOfInterest, UserLevel } from '@poltergeist/types';

export interface UseUserLevelResult {
  userLevel: UserLevel | null;
  loading: boolean;
  error: Error | null;
}

export const useUserLevel = (): UseUserLevelResult => {
  const { apiClient } = useAPI();
  const [userLevel, setUserLevel] = useState<UserLevel | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchPointsOfInterest = async () => {
      try {
        const fetchedUserLevel = await apiClient.get<UserLevel>(`/sonar/level`);
        setUserLevel(fetchedUserLevel);
      } catch (error) {
        setError(error as Error);
      } finally {
        setLoading(false);
      }
    };

    fetchPointsOfInterest();
  }, [apiClient]);

  return {
    userLevel,
    loading,
    error,
  };
};
