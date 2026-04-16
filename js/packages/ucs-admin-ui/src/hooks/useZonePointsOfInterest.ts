import { useCallback, useEffect, useState } from 'react';
import { useAPI, useAuth } from '@poltergeist/contexts';
import { PointOfInterest } from '@poltergeist/types';

export const useZonePointsOfInterest = (zoneId: string) => {
  const { apiClient } = useAPI();
  const { user } = useAuth();
  const [pointsOfInterest, setPointsOfInterest] = useState<PointOfInterest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const refreshPointsOfInterest = useCallback(async () => {
    if (!user || !zoneId) {
      setPointsOfInterest([]);
      setLoading(false);
      return;
    }

    setLoading(true);
    try {
      const response = await apiClient.get<PointOfInterest[]>(
        `/sonar/zones/${zoneId}/pointsOfInterest`
      );
      setPointsOfInterest(response);
      setError(null);
    } catch (err) {
      setError(
        err instanceof Error
          ? err
          : new Error('Failed to fetch points of interest')
      );
    } finally {
      setLoading(false);
    }
  }, [apiClient, user, zoneId]);

  useEffect(() => {
    void refreshPointsOfInterest();
  }, [refreshPointsOfInterest]);

  return { pointsOfInterest, loading, error, refreshPointsOfInterest };
};
