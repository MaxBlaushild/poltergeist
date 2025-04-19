import { useEffect, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { QuestArchType } from '@poltergeist/types';

export const useQuestArchtypes = () => {
  const { apiClient } = useAPI();
  const [questArchtypes, setQuestArchtypes] = useState<QuestArchType[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchQuestArchtypes = async () => {
      try {
        const response = await apiClient.get<QuestArchType[]>(`/sonar/quests/archTypes`);
        setQuestArchtypes(response);
        setLoading(false);
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch quest archtypes'));
        setLoading(false);
      }
    };

    fetchQuestArchtypes();
  }, []);

  return { questArchtypes, loading, error };
};
