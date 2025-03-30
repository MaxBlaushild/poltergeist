import { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';

const useHasCurrentMatch = () => {
  const [hasCurrentMatch, setHasCurrentMatch] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const [matchID, setMatchID] = useState<string>('');
  const { apiClient } = useAPI();

  useEffect(() => {
    const fetchHasCurrentMatch = async () => {
      try {
        const response = await apiClient.get<{ hasCurrentMatch: boolean, matchID: string }>(
          `/sonar/matches/hasCurrentMatch`
        );
        setHasCurrentMatch(response.hasCurrentMatch);
        setMatchID(response.matchID);
      } catch (err) {
        setError('Failed to fetch has current match');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchHasCurrentMatch();
  }, [apiClient]);

  return { hasCurrentMatch, loading, error, matchID };
};

export default useHasCurrentMatch;
