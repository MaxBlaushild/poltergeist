import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
} from 'react';
import { useAPI, useInventory, useMediaContext } from '@poltergeist/contexts';
import { PointOfInterestChallengeSubmission } from '@poltergeist/types/dist/pointOfInterestChallengeSubmission';
import { useMatchContext } from './MatchContext.tsx';
import { InventoryItem } from '@poltergeist/types';
import { useUserProfiles } from './UserProfileContext.tsx';

export type Judgement = {
  judgement: InnerJudgement;
  reason: string;
};

export type InnerJudgement = {
  judgement: InnerMostJudgement;
};

export type InnerMostJudgement = {
  judgement: boolean;
  reason: string;
};

export type CapturePointOfInterestResponse = {
  item: InventoryItem;
  judgement: InnerJudgement;
};

interface SubmissionsContextType {
  submissions: PointOfInterestChallengeSubmission[];
  fetchSubmissions: () => Promise<void>;
  setSubmissions: (submissions: PointOfInterestChallengeSubmission[]) => void;
  createSubmission: (
    challengeId: string,
    text: string | undefined,
    image?: File | undefined
  ) => Promise<{ correctness: boolean, reason: string } | undefined>;
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
  const [submissions, setSubmissions] = useState<PointOfInterestChallengeSubmission[]>(
    []
  );
  const { usersTeam } = useMatchContext();
  const { currentUser } = useUserProfiles();
  const { getPresignedUploadURL, uploadMedia } = useMediaContext();
  const { setPresentedInventoryItem } = useInventory();
  const fetchSubmissions = useCallback(async () => {
    try {
      const response = await apiClient.get<PointOfInterestChallengeSubmission[]>(
          `/sonar/pointsOfInterest/challenges/submissions`
      );
      setSubmissions(response);
    } catch (error) {
      console.error('Failed to fetch submissions:', error);
    }
  }, [apiClient]);

  const createSubmission = useCallback(async (
    challengeId: string,
    text: string | undefined,
    image?: File | undefined
  ): Promise<{ correctness: boolean, reason: string } | undefined> => {
    const key = `${usersTeam ? usersTeam.id : currentUser?.id}/${challengeId}.webp`;
    let imageUrl = '';

    if (image) {
      const presignedUrl = await getPresignedUploadURL("crew-points-of-interest", key);
      if (!presignedUrl) return;
      await uploadMedia(presignedUrl, image);
      imageUrl = presignedUrl.split("?")[0];
    }

    var response = await apiClient.post<CapturePointOfInterestResponse>(`/sonar/pointOfInterest/challenge`, {
      teamId: usersTeam?.id,
      userId: usersTeam ? undefined : currentUser?.id,
      challengeId,
      textSubmission: text,
      imageSubmissionUrl: imageUrl,
    });

    if (response.judgement.judgement.judgement) {
      setPresentedInventoryItem(response.item);
    }

    fetchSubmissions();

    return { correctness: response?.judgement?.judgement?.judgement ?? false, reason: response?.judgement?.judgement?.reason ?? 'Failed for an unknown reason.' };
  }, [apiClient, usersTeam, currentUser, getPresignedUploadURL, uploadMedia, setPresentedInventoryItem]);

  useEffect(() => {
    fetchSubmissions();
    const interval = setInterval(() => {
      fetchSubmissions();
    }, 5000);

    return () => clearInterval(interval);
  }, [fetchSubmissions]);

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
