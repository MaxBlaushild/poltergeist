import { useState, useEffect } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { ImageGeneration } from '@poltergeist/types';

const useImageGenerations = () => {
  const [imageGenerations, setImageGenerations] = useState<ImageGeneration[] | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');
  const { apiClient } = useAPI();

  useEffect(() => {
    const fetchImageGenerations = async () => {
      try {
        const response = await apiClient.get<ImageGeneration[]>(
          `/sonar/generations/complete`
        );
        setImageGenerations(response);
      } catch (err) {
        setError('Failed to fetch image generations');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchImageGenerations();
  }, [apiClient]);

  return { imageGenerations, loading, error };
};

export default useImageGenerations;
