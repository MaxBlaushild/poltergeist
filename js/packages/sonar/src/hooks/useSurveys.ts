import { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { Survey } from '@poltergeist/types';

export const useSurveys = () => {
  const [surveys, setSurveys] = useState<Survey[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);
  const { apiClient } = useAPI();

  useEffect(() => {
    const fetchSurveys = async () => {
      try {
        const response = await apiClient.get<Survey[]>('/sonar/surveys');
        setSurveys(response);
      } catch (err) {
        setError(err);
        console.error('Failed to fetch surveys:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchSurveys();
  }, [apiClient]);

  return { surveys, isLoading, error };
};
