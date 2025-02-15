import { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { PointOfInterest } from '@poltergeist/types';

export interface UsePointsOfInterestResult {
  pointsOfInterest: PointOfInterest[] | null;
  loading: boolean;
  error: Error | null;
}

export const usePointsOfInterest = (): UsePointsOfInterestResult => {
  const { apiClient } = useAPI();
  const [pointsOfInterest, setPointsOfInterest] = useState<PointOfInterest[] | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchPointsOfInterest = async () => {
      try {
        const fetchedPointsOfInterest = await apiClient.get<PointOfInterest[]>(`/sonar/pointsOfInterest`);
        setPointsOfInterest(fetchedPointsOfInterest);
      } catch (error) {
        setError(error as Error);
      } finally {
        setLoading(false);
      }
    };

    fetchPointsOfInterest();
  }, [apiClient]);

  return {
    pointsOfInterest,
    loading,
    error,
  };
};
