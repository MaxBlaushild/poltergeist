import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
} from 'react';
import { useAPI, useAuth, useLocation } from '@poltergeist/contexts';
import { PointOfInterest } from '@poltergeist/types';
import { useUserProfiles } from './UserProfileContext.tsx';

interface PointOfInterestContextType {
  fetchPointsOfInterest: () => Promise<void>;
  pointsOfInterest: PointOfInterest[];
}

interface PointOfInterestProviderProps {
  children: React.ReactNode;
}

export const PointOfInterestContext = createContext<PointOfInterestContextType | undefined>(
  undefined
);

export const usePointOfInterestContext = () => {
  const context = useContext(PointOfInterestContext);
  if (!context) {
    throw new Error(
      'usePointOfInterestContext must be used within a PointOfInterestContextProvider'
    );
  }
  return context;
};

export const PointOfInterestContextProvider: React.FC<PointOfInterestProviderProps> = ({ children }) => {
  const { apiClient } = useAPI();
  const { user } = useAuth();
  const [pointsOfInterest, setPointsOfInterest] = useState<PointOfInterest[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchPointsOfInterest = async () => {
    try {
      const fetchedPointsOfInterest = await apiClient.get<PointOfInterest[]>(`/sonar/pointsOfInterest`);
      setPointsOfInterest(fetchedPointsOfInterest);
    } catch (error: any) {
      // Silently handle auth errors
      if (error?.response?.status === 401 || error?.response?.status === 403) {
        setPointsOfInterest([]);
        return;
      }
      setError(error as Error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!user) {
      // Clear data when not authenticated
      setPointsOfInterest([]);
      return;
    }

    fetchPointsOfInterest();
    const interval = setInterval(() => {
      fetchPointsOfInterest();
    }, 5000);

    return () => clearInterval(interval);
  }, [user]);

  return (
    <PointOfInterestContext.Provider
      value={{
        fetchPointsOfInterest,
        pointsOfInterest,
      }}
    >
      {children}
    </PointOfInterestContext.Provider>
  );
};
