import { useEffect, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';

export const usePlaceTypes = () => {
  const { apiClient } = useAPI();
  const [placeTypes, setPlaceTypes] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchPlaceTypes = async () => {
      try {
        const response = await apiClient.get<string[]>(`/sonar/placeTypes`);
        setPlaceTypes(response);
        setLoading(false);
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch place types'));
        setLoading(false);
      }
    };

    fetchPlaceTypes();
  }, []);

  return { placeTypes, loading, error };
};
