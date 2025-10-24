import { useEffect, useState } from 'react';
import { useAPI, useAuth } from '@poltergeist/contexts';

export const usePlaceTypes = () => {
  const { apiClient } = useAPI();
  const { user } = useAuth();
  const [placeTypes, setPlaceTypes] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!user) {
      setPlaceTypes([]);
      setLoading(false);
      return;
    }

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
  }, [user]);

  return { placeTypes, loading, error };
};
