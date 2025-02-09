import { useState } from 'react';
import { useAPI } from '@poltergeist/contexts';

interface CreatePointOfInterestParams {
  name: string;
  description: string;
  latitude: number;
  longitude: number;
  groupId: string;
  imageUrl?: string;
  challenges?: Array<{
    question: string;
    tier: number;
    inventoryItemId: number;
  }>;
}

export const useCreatePointOfInterest = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const api = useAPI();

  const createPointOfInterest = async (params: CreatePointOfInterestParams) => {
    setLoading(true);
    setError(null);

    try {
      const response = await api.post('/points-of-interest', params);
      return response.data;
    } catch (err) {
      setError(err as Error);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return {
    createPointOfInterest,
    loading,
    error
  };
};
