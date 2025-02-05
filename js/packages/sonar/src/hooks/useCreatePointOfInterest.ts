import { useAPI } from '@poltergeist/contexts';

export interface CreatePointOfInterestPayload {
  name: string;
  description: string;
  imageUrl: string;
  lat: string;
  lon: string;
  clue: string;
  tierOne: string;
  tierTwo?: string;
  tierThree?: string;
  tierOneInventoryItemId?: number;
  tierTwoInventoryItemId?: number;
  tierThreeInventoryItemId?: number;
  pointOfInterestGroupMemberId?: string;
}

export const useCreatePointOfInterest = () => {
  const { apiClient } = useAPI();

  const createPointOfInterest = async (payload: CreatePointOfInterestPayload) => {
    try {
      await apiClient.post('/sonar/pointsOfInterest', payload as unknown as Record<string, unknown>);
      return null;
    } catch (error) {
      console.error('Error creating point of interest:', error);
      throw error;
    }
  };

  return { createPointOfInterest };
};
