import { useEffect, useState } from 'react';
import { useAPI, useAuth } from '@poltergeist/contexts';
import { QuestArchType } from '@poltergeist/types';

export const useQuestArchtypes = () => {
  const { apiClient } = useAPI();
  const { user } = useAuth();
  const [questArchtypes, setQuestArchtypes] = useState<QuestArchType[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!user) {
      setQuestArchtypes([]);
      setLoading(false);
      return;
    }

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
  }, [user]);

  return { questArchtypes, loading, error };
};
