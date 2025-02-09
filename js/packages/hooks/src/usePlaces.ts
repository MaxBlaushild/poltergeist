import { useAPI } from '@poltergeist/contexts';
import { useState, useEffect } from 'react';
import { Place } from '@poltergeist/types';

interface UsePlacesResult {
  places: Place[];
  loading: boolean;
  error: Error | null;
}

export const usePlaces = (address: string): UsePlacesResult => {
  const [places, setPlaces] = useState<Place[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);
  const { apiClient } = useAPI();

  useEffect(() => {
    const fetchPlaces = async () => {
      try {
        const data = await apiClient.get<Place[]>(
          `/sonar/mapbox/places?address=${address}`
        );
        setPlaces(data);
      } catch (err) {
        setError(err as Error);
      } finally {
        setLoading(false);
      }
    };

    if (address) {
      fetchPlaces();
    }
  }, [address]);

  return {
    places,
    loading,
    error
  };
};
