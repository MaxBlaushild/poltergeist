import { useState } from 'react';
import { useAPI } from '@poltergeist/contexts';

export const useGeneratePointsOfInterest = (zoneId: string) => {
  const { apiClient } = useAPI();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const generatePointsOfInterest = async (zoneId: string, placeType: string) => {
    try {
      setLoading(true);
      await apiClient.post(`/sonar/zones/${zoneId}/pointsOfInterest`, { placeType });
      setLoading(false);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to generate points of interest'));
      setLoading(false);
    }
  };

  return { loading, error, generatePointsOfInterest };
};
