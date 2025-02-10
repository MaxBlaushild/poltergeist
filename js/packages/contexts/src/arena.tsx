import React, { createContext, useContext, useState, useCallback, useEffect } from 'react';
import { PointOfInterestGroup, PointOfInterest, PointOfInterestChallenge } from '@poltergeist/types';
import { useMediaContext } from './media';
import { useAPI } from './api';

interface ArenaContextType {
  arena: PointOfInterestGroup | null;
  loading: boolean;
  error: Error | null;
  updateArena: (name: string, description: string) => Promise<void>;
  updateArenaImage: (id: string, image: File) => Promise<void>;
  createPointOfInterest: (
    name: string,
    description: string,
    lat: number,
    lng: number,
    image: File | null,
    clue: string
  ) => Promise<void>;
  updatePointOfInterest: (id: string, arena: Partial<PointOfInterest>) => Promise<void>;
  updatePointOfInterestImage: (id: string, image: File) => Promise<void>;
  deletePointOfInterest: (id: string) => Promise<void>;
  updatePointOfInterestChallenge: (id: string, challenge: Partial<PointOfInterestChallenge>) => Promise<void>;
  deletePointOfInterestChallenge: (id: string) => Promise<void>;
  createPointOfInterestChallenge: (id: string, challenge: Partial<PointOfInterestChallenge>) => Promise<void>;
  createPointOfInterestChildren: (pointOfInterestId: string, pointOfInterestGroupId: string, pointOfInterestChallengeId: string) => Promise<void>;
  deletePointOfInterestChildren: (id: string) => Promise<void>;
}

interface ArenaProviderProps {
  children: React.ReactNode;
  arenaId: string | undefined | null;
}

const ArenaContext = createContext<ArenaContextType | undefined>(undefined);

export const ArenaProvider: React.FC<ArenaProviderProps> = ({ children, arenaId }) => {
  const [arena, setArena] = useState<PointOfInterestGroup | null>(null);
  const { apiClient } = useAPI();
  const mediaContext = useMediaContext();
  if (!mediaContext) {
    throw new Error('ArenaProvider must be wrapped in a MediaProvider');
  }
  const { uploadMedia, getPresignedUploadURL } = mediaContext;
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const fetchArena = async (arenaId: string) => {
    setLoading(true);
    try {
      const response = await apiClient.get<PointOfInterestGroup>(`/sonar/pointsOfInterest/group/${arenaId}`);
      setArena(response);
    } catch (err) {
      console.error('Error fetching arena', err);
      setError(err instanceof Error ? err : new Error('An error occurred'));
    } finally {
      setLoading(false);
    }
  };

  const updateArena = async (name: string, description: string) => {
    setLoading(true);

    if (!arena) {
      return;
    }

    try {
      const response = await apiClient.patch(`/sonar/pointsOfInterest/group/${arenaId}`, {
        name,
        description,
      });

      setArena({
        ...arena,
        name,
        description,
      });
    } catch (err) {
      setError(err instanceof Error ? err : new Error('An error occurred'));
    } finally {
      setLoading(false);
    }
  };

  const updateArenaImage = async (id: string, image: File) => {
    setLoading(true);

    const imageKey = `arenas/${(image?.name || 'image.jpg').toLowerCase().replace(/\s+/g, '-')}`;
    let imageUrl = '';

    if (image) {
      const presignedUrl = await getPresignedUploadURL("crew-points-of-interest", imageKey);
      if (!presignedUrl) return;
      await uploadMedia(presignedUrl, image);
      imageUrl = presignedUrl.split("?")[0];
    }

    try {
      const response = await apiClient.patch(`/sonar/pointsofInterest/group/imageUrl/${id}`, {
        imageUrl,
      });
      fetchArena(arenaId!);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('An error occurred'));
    } finally {
      setLoading(false);
    }
  };

  const createPointOfInterest = async (
    name: string,
    description: string,
    lat: number,
    lng: number,
    image: File | null,
    clue: string
  ) => {
    if (!name || !description || !lat || !lng || !image || !clue || !arenaId) {
      return;
    }

    setLoading(true);

    const key = `${encodeURIComponent(name)}${image.name.substring(image.name.lastIndexOf('.'))}`;
    const imageUrl = await getPresignedUploadURL(
      'crew-points-of-interest',
      key
    );
    if (!imageUrl) {
      return;
    }
    const uploadResult = await uploadMedia(imageUrl, image);
    if (!uploadResult) {
      return;
    }

    try {
      const res = await apiClient.post(
        `/sonar/pointsOfInterest/group/${arenaId}`,
        {
          name,
          description,
          latitude: JSON.stringify(lat),
          longitude: JSON.stringify(lng),
          imageUrl: imageUrl.split("?")[0],
          clue,
          pointOfInterestGroupId: arenaId,
        }
      );
      fetchArena(arenaId);
    } catch (error) {
      console.error('Error creating point:', error);
    } finally {
      setLoading(false);
    }
  };

  const updatePointOfInterest = async (id: string, arena: Partial<PointOfInterest>) => {
    setLoading(true);
    try {
      const response = await apiClient.patch(`/sonar/pointsOfInterest/${id}`, {
        arena,
      });
      fetchArena(arenaId!);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('An error occurred'));
    } finally {
      setLoading(false);
    }
  };

  const updatePointOfInterestImage = async (id: string, image: File) => {
    setLoading(true);
    const imageKey = `arenas/${(image?.name || 'image.jpg').toLowerCase().replace(/\s+/g, '-')}`;
    let imageUrl = '';

    if (image) {
      const presignedUrl = await getPresignedUploadURL("crew-points-of-interest", imageKey);
      if (!presignedUrl) return;
      await uploadMedia(presignedUrl, image);
      imageUrl = presignedUrl.split("?")[0];
    }

    try {
      const response = await apiClient.patch(`/sonar/pointsofInterest/imageUrl/${id}`, {
        imageUrl,
      });
      fetchArena(arenaId!);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('An error occurred'));
    } finally {
      setLoading(false);
    }
  };

  const deletePointOfInterest = async (id: string) => {
    setLoading(true);
    try {
      const response = await apiClient.delete(`/sonar/pointsOfInterest/${id}`);
      fetchArena(arenaId!);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('An error occurred'));
    } finally {
      setLoading(false);
    }
  };

  const createPointOfInterestChallenge = async (id: string, challenge: Partial<PointOfInterestChallenge>) => {
    setLoading(true);
    try {
      const response = await apiClient.post(`/sonar/pointsOfInterest/challenge`, {
        ...challenge,
        pointOfInterestId: id,
      });
      fetchArena(arenaId!);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('An error occurred'));
    } finally {
      setLoading(false);
    }
  };

  const updatePointOfInterestChallenge = async (id: string, challenge: Partial<PointOfInterestChallenge>) => {
    setLoading(true);
    try {
      const response = await apiClient.patch(`/sonar/pointsOfInterest/challenge/${id}`, {
        ...challenge,
      });
      fetchArena(arenaId!);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('An error occurred'));
    } finally {
      setLoading(false);
    }
  };

  const deletePointOfInterestChallenge = async (id: string) => {
    setLoading(true);
    try {
      const response = await apiClient.delete(`/sonar/pointsOfInterest/challenge/${id}`);
      fetchArena(arenaId!);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('An error occurred'));
    } finally {
      setLoading(false);
    }
  };

  const createPointOfInterestChildren = async (pointOfInterestId: string, pointOfInterestGroupMemberId: string, pointOfInterestChallengeId: string) => {
    setLoading(true);
    try {
      const response = await apiClient.post(`/sonar/pointOfInterest/children`, {
        pointOfInterestId,
        pointOfInterestGroupMemberId,
        pointOfInterestChallengeId,
      });
      fetchArena(arenaId!);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('An error occurred'));
    } finally {
      setLoading(false);
    }
  };

  const deletePointOfInterestChildren = async (id: string) => {
    setLoading(true);
    try {
      const response = await apiClient.delete(`/sonar/pointOfInterest/children/${id}`);
      fetchArena(arenaId!);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('An error occurred'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (arenaId) {
      fetchArena(arenaId);
    }
  }, [arenaId]);

  return (
    <ArenaContext.Provider
      value={{
        arena,
        loading,
        error,
        updateArena,
        updateArenaImage,
        createPointOfInterest,
        updatePointOfInterest,
        updatePointOfInterestImage,
        deletePointOfInterest,
        createPointOfInterestChallenge,
        updatePointOfInterestChallenge,
        deletePointOfInterestChallenge,
        createPointOfInterestChildren,
        deletePointOfInterestChildren,
      }}
    >
      {children}
    </ArenaContext.Provider>
  );
};

export const useArena = () => {
  const context = useContext(ArenaContext);
  if (context === undefined) {
    throw new Error('useArena must be used within a ArenaProvider');
  }
  return context;
};

