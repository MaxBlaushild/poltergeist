import { useState } from 'react';
import { useAPI } from '@poltergeist/contexts';

export const useGeneratePointsOfInterest = (zoneId: string) => {
  const { apiClient } = useAPI();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const generatePointsOfInterest = async (zoneId: string, includedPlaceTypes: string[], excludedPlaceTypes: string[], numberOfPlaces: number) => {
    try {
      setLoading(true);
      await apiClient.post(`/sonar/zones/${zoneId}/pointsOfInterest`, { includedPlaceTypes, excludedPlaceTypes, numberOfPlaces });
      setLoading(false);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to generate points of interest'));
      setLoading(false);
    }
  };

  const refreshPointOfInterestImage = async (pointOfInterestID: string) => {
    try {
      setLoading(true);
      await apiClient.post(`/sonar/pointOfInterest/image/refresh`, { pointOfInterestID });
      setLoading(false);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to refresh point of interest image'));
      setLoading(false);
    }
  };

  const refreshPointOfInterest = async (pointOfInterestID: string) => {
    try {
      setLoading(true);
      await apiClient.post(`/sonar/pointOfInterest/refresh`, { pointOfInterestID });
      setLoading(false);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to refresh point of interest'));
      setLoading(false);
    }
  };

  const importPointOfInterest = async (placeID: string, zoneID: string) => {
    try {
      setLoading(true);
      await apiClient.post(`/sonar/pointOfInterest/import`, { placeID, zoneID });
      setLoading(false);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to import point of interest'));
      setLoading(false);
    }
  };

  const generateQuest = async (zoneId: string, questArchTypeID: string) => {
    try {
      setLoading(true);
      await apiClient.post(`/sonar/quests/${zoneId}/${questArchTypeID}/generate`);
      setLoading(false);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to generate quest'));
      setLoading(false);
    }
  };

  return { 
    loading, 
    error, 
    generatePointsOfInterest, 
    refreshPointOfInterestImage, 
    refreshPointOfInterest,
    importPointOfInterest,
    generateQuest,
  };
};
