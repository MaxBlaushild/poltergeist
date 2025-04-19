import { useAPI } from '@poltergeist/contexts';
import { useState, useEffect } from 'react';
import { Candidate } from '@poltergeist/types';

interface UsePlacesResult {
  candidates: Candidate[];
  loading: boolean;
  error: Error | null;
}

export const useCandidates = (query: string): UsePlacesResult => {
  const [candidates, setCandidates] = useState<Candidate[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);
  const { apiClient } = useAPI();

  useEffect(() => {
    const fetchPlaces = async () => {
      try {
        const data = await apiClient.get<Candidate[]>(
          `/sonar/google/places?query=${query}`
        );
        setCandidates(data);
      } catch (err) {
        setError(err as Error);
      } finally {
        setLoading(false);
      }
    };

    if (query) {
      fetchPlaces();
    }
  }, [query]);

  return {
    candidates,
    loading,
    error
  };
};
