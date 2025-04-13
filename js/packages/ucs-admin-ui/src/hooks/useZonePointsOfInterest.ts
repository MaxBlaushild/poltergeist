import { useEffect, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { PointOfInterest } from '@poltergeist/types';


export const useZonePointsOfInterest = (zoneId: string) => {
  const { apiClient } = useAPI();
  const [pointsOfInterest, setPointsOfInterest] = useState<PointOfInterest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchPointsOfInterest = async () => {
      try {
        const response = await apiClient.get<PointOfInterest[]>(`/sonar/zones/${zoneId}/pointsOfInterest`);
        setPointsOfInterest(response);
        setLoading(false);
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch points of interest'));
        setLoading(false);
      }
    };

    fetchPointsOfInterest();
  }, []);

  return { pointsOfInterest, loading, error };
};
