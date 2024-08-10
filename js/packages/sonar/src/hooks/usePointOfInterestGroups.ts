import { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { PointOfInterestGroup } from '@poltergeist/types';

export interface UsePointOfInterestGroupsResult {
  pointOfInterestGroups: PointOfInterestGroup[] | null;
  loading: boolean;
  error: Error | null;
}

export const usePointOfInterestGroups = (): UsePointOfInterestGroupsResult => {
  const { apiClient } = useAPI();
  const [pointOfInterestGroups, setPointOfInterestGroups] = useState<PointOfInterestGroup[] | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchPointOfInterestGroups = async () => {
      try {
        const fetchedPointOfInterestGroups = await apiClient.get<PointOfInterestGroup[]>('/sonar/pointsOfInterest/groups');
        setPointOfInterestGroups(fetchedPointOfInterestGroups);
      } catch (error) {
        setError(error);
      } finally {
        setLoading(false);
      }
    };

    fetchPointOfInterestGroups();
  }, [apiClient]);

  return {
    pointOfInterestGroups,
    loading,
    error,
  };
};
