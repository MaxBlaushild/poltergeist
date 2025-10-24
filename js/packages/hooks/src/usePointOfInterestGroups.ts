import { useState, useEffect } from 'react';
import { useAPI, useAuth } from '@poltergeist/contexts';
import { PointOfInterestGroup } from '@poltergeist/types';

export interface UsePointOfInterestGroupsResult {
  pointOfInterestGroups: PointOfInterestGroup[] | null;
  loading: boolean;
  error: Error | null;
}

export const usePointOfInterestGroups = (type?: number): UsePointOfInterestGroupsResult => {
  const { apiClient } = useAPI();
  const { user } = useAuth();
  const [pointOfInterestGroups, setPointOfInterestGroups] = useState<PointOfInterestGroup[] | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!user) {
      setPointOfInterestGroups([]);
      setLoading(false);
      return;
    }

    const fetchPointOfInterestGroups = async () => {
      try {
        const fetchedPointOfInterestGroups = await apiClient.get<PointOfInterestGroup[]>(`/sonar/pointsOfInterest/groups${type ? `?type=${type}` : ''}`);
        setPointOfInterestGroups(fetchedPointOfInterestGroups);
      } catch (error) {
        setError(error as Error);
      } finally {
        setLoading(false);
      }
    };

    fetchPointOfInterestGroups();
  }, [user, type]);

  return {
    pointOfInterestGroups,
    loading,
    error,
  };
};
