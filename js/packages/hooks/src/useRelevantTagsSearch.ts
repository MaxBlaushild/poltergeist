import { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { PointOfInterest, Tag } from '@poltergeist/types';

export interface UseRelevantTagsSearchResult {
  relevantTags: Tag[] | null;
  loading: boolean;
  error: Error | null;
}

export const useRelevantTagsSearch = (query: string): UseRelevantTagsSearchResult => {
  const { apiClient } = useAPI();
  const [relevantTags, setRelevantTags] = useState<Tag[] | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchRelevantTags = async (query: string) => {
      try {
        setLoading(true);
        const fetchedRelevantTags = await apiClient.get<Tag[]>(`/sonar/search/tags?query=${encodeURIComponent(query)}`);
        setRelevantTags(fetchedRelevantTags);
      } catch (error) {
        setError(error as Error);
      } finally {
        setLoading(false);
      }
    };

    if (query) {
      fetchRelevantTags(query);
    } else {
      setLoading(false);
    }
  }, [apiClient, query]);

  return {
    relevantTags,
    loading,
    error,
  };
};
