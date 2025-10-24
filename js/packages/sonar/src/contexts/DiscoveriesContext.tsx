import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
} from 'react';
import { useAPI, useAuth, useLocation } from '@poltergeist/contexts';
import { PointOfInterestDiscovery } from '@poltergeist/types';
import { useUserProfiles } from './UserProfileContext.tsx';

interface DiscoveriesContextType {
  discoveries: PointOfInterestDiscovery[];
  fetchDiscoveries: () => Promise<void>;
  setDiscoveries: (discoveries: PointOfInterestDiscovery[]) => void;
  discoverPointOfInterest: (
    pointOfInterestId: string,
    teamId?: string | undefined,
    userId?: string | undefined,
  ) => Promise<void>;
}

interface DiscoveriesContextProviderProps {
  children: React.ReactNode;
}

export const DiscoveriesContext = createContext<
  DiscoveriesContextType | undefined
>(undefined);

export const useDiscoveriesContext = () => {
  const context = useContext(DiscoveriesContext);
  if (!context) {
    throw new Error(
      'useDiscoveriesContext must be used within a DiscoveriesContextProvider'
    );
  }
  return context;
};

export const DiscoveriesContextProvider: React.FC<
  DiscoveriesContextProviderProps
> = ({ children }) => {
  const { apiClient } = useAPI();
  const { user } = useAuth();
  const { currentUser } = useUserProfiles();
  const { location } = useLocation();
  const [discoveries, setDiscoveries] = useState<PointOfInterestDiscovery[]>(
    []
  );

  const fetchDiscoveries = useCallback(async () => {
    try {
      const response = await apiClient.get<PointOfInterestDiscovery[]>(
        `/sonar/pointsOfInterest/discoveries`
      );
      setDiscoveries(response);
    } catch (error: any) {
      // Silently handle auth errors
      if (error?.response?.status === 401 || error?.response?.status === 403) {
        setDiscoveries([]);
        return;
      }
      console.error('Failed to fetch discoveries:', error);
    }
  }, [apiClient]);

  const discoverPointOfInterest = async (
    pointOfInterestId: string,
    teamId?: string | undefined,
    userId?: string | undefined,
  ) => {
    await apiClient.post(`/sonar/pointOfInterest/unlock`, {
      pointOfInterestId,
      teamId,
      userId,
      lat: location?.latitude?.toString(),
      lng: location?.longitude?.toString(),
    });

    setDiscoveries([...discoveries, {
      id: '',
      createdAt: new Date(),
      updatedAt: new Date(),
      teamId: teamId,
      userId: userId,
      pointOfInterestId: pointOfInterestId,
    }]);
  };

  useEffect(() => {
    if (!user) {
      // Clear data when not authenticated
      setDiscoveries([]);
      return;
    }

    fetchDiscoveries();
  }, [fetchDiscoveries, user]);

  return (
    <DiscoveriesContext.Provider
      value={{
        discoveries,
        fetchDiscoveries,
        setDiscoveries,
        discoverPointOfInterest,
      }}
    >
      {children}
    </DiscoveriesContext.Provider>
  );
};
