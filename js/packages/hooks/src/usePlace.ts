import { useAPI } from '@poltergeist/contexts';
import { useState, useEffect } from 'react';
import { GooglePlace } from '@poltergeist/types';

interface UsePlaceResult {
  place: GooglePlace | null;
  loading: boolean;
  error: Error | null;
}

export const usePlace = (placeId: string): UsePlaceResult => {
  const [place, setPlace] = useState<GooglePlace | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);
  const { apiClient } = useAPI();

  useEffect(() => {
    const fetchPlaces = async () => {
      try {
        const data = await apiClient.get<GooglePlace>(
          `/sonar/google/place/${placeId}`
        );
        setPlace(data);
      } catch (err) {
        setError(err as Error);
      } finally {
        setLoading(false);
      }
    };

      fetchPlaces();
  }, [placeId]);    

  return {
    place: place ?? null,
    loading,
    error
  };
};
