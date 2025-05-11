import { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { UserLevel, UserZoneReputation } from '@poltergeist/types';

export interface UseUserZoneReputationResult {
  userZoneReputation: UserZoneReputation | null;
  loading: boolean;
  error: Error | null;
}

export const useUserZoneReputation = (zoneId: string | undefined): UseUserZoneReputationResult => {
  const { apiClient } = useAPI();
  const [userZoneReputation, setUserZoneReputation] = useState<UserZoneReputation | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!zoneId) {
      setLoading(false);
      return;
    }

    const fetchUserZoneReputation = async () => {
      try {
        const fetchedUserZoneReputation = await apiClient.get<UserZoneReputation>(`/sonar/zones/${zoneId}/reputation`);
        setUserZoneReputation(fetchedUserZoneReputation);
      } catch (error) {
        setError(error as Error);
      } finally {
        setLoading(false);
      }
    };

    fetchUserZoneReputation();
  }, [apiClient, zoneId]);

  return {
    userZoneReputation,
    loading,
    error,
  };
};
