import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
} from 'react';
import { useAPI, useAuth, useInventory, useMediaContext } from '@poltergeist/contexts';
import { PointOfInterestChallengeSubmission } from '@poltergeist/types/dist/pointOfInterestChallengeSubmission';
import { InventoryItem, SubmissionResult } from '@poltergeist/types';
import { useUserProfiles } from './UserProfileContext.tsx';


interface SubmissionsContextType {
  submissions: PointOfInterestChallengeSubmission[];
  fetchSubmissions: () => Promise<void>;
  setSubmissions: (submissions: PointOfInterestChallengeSubmission[]) => void;
  createSubmission: (
    challengeId: string,
    text: string | undefined,
    image?: File | undefined,
    teamId?: string | undefined,
    userId?: string | undefined,
  ) => Promise<SubmissionResult | undefined>;
}

interface SubmissionsContextProviderProps {
  children: React.ReactNode;
}

export const SubmissionsContext = createContext<
  SubmissionsContextType | undefined
>(undefined);

export const useSubmissionsContext = () => {
  const context = useContext(SubmissionsContext);
  if (!context) {
    throw new Error(
      'useSubmissionsContext must be used within a SubmissionsContextProvider'
    );
  }
  return context;
};

export const SubmissionsContextProvider: React.FC<
  SubmissionsContextProviderProps
> = ({ children }) => {
  const { apiClient } = useAPI();
  const { user } = useAuth();
  const [submissions, setSubmissions] = useState<PointOfInterestChallengeSubmission[]>(
    []
  );
  const { getPresignedUploadURL, uploadMedia } = useMediaContext();
  const { setPresentedInventoryItem } = useInventory();
  const fetchSubmissions = useCallback(async () => {
    try {
      const response = await apiClient.get<PointOfInterestChallengeSubmission[]>(
          `/sonar/pointsOfInterest/challenges/submissions`
      );
      setSubmissions(response);
    } catch (error: any) {
      // Silently handle auth errors
      if (error?.response?.status === 401 || error?.response?.status === 403) {
        setSubmissions([]);
        return;
      }
      console.error('Failed to fetch submissions:', error);
    }
  }, [apiClient]);

  const createSubmission = useCallback(async (
    challengeId: string,
    text: string | undefined,
    image?: File | undefined,
    teamId?: string | undefined,
    userId?: string | undefined,
  ): Promise<SubmissionResult | undefined> => {
    const key = `${teamId ? teamId : userId}/${challengeId}.webp`;
    let imageUrl = '';

    if (image) {
      const presignedUrl = await getPresignedUploadURL("crew-points-of-interest", key);
      if (!presignedUrl) return;
      await uploadMedia(presignedUrl, image);
      imageUrl = presignedUrl.split("?")[0];
    }

    var response = await apiClient.post<SubmissionResult>(`/sonar/pointOfInterest/challenge`, {
      teamId,
      userId,
      challengeId,
      textSubmission: text,
      imageSubmissionUrl: imageUrl,
    });

    fetchSubmissions();

    return response;
  }, [apiClient, getPresignedUploadURL, uploadMedia, setPresentedInventoryItem]);

  useEffect(() => {
    if (!user) {
      // Clear data when not authenticated
      setSubmissions([]);
      return;
    }

    fetchSubmissions();
    const interval = setInterval(() => {
      fetchSubmissions();
    }, 5000);

    return () => clearInterval(interval);
  }, [fetchSubmissions, user]);

  return (
    <SubmissionsContext.Provider
      value={{
        submissions,
        fetchSubmissions,
        setSubmissions,
        createSubmission,
      }}
    >
      {children}
    </SubmissionsContext.Provider>
  );
};
