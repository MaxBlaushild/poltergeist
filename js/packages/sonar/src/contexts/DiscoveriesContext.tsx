import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
} from 'react';
import { useAPI, useLocation } from '@poltergeist/contexts';
import { PointOfInterestDiscovery } from '@poltergeist/types';
import { useUserProfiles } from './UserProfileContext.tsx';
import { useMatchContext } from './MatchContext.tsx';

interface DiscoveriesContextType {
  discoveries: PointOfInterestDiscovery[];
  fetchDiscoveries: () => Promise<void>;
  setDiscoveries: (discoveries: PointOfInterestDiscovery[]) => void;
  discoverPointOfInterest: (pointOfInterestId: string) => Promise<void>;
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
  const { currentUser } = useUserProfiles();
  const { location } = useLocation();
  const [discoveries, setDiscoveries] = useState<PointOfInterestDiscovery[]>(
    []
  );
  const { usersTeam } = useMatchContext();

  const fetchDiscoveries = useCallback(async () => {
    try {
      const response = await apiClient.get<PointOfInterestDiscovery[]>(
        `/sonar/pointsOfInterest/discoveries`
      );
      setDiscoveries(response);
    } catch (error) {
      console.error('Failed to fetch discoveries:', error);
    }
  }, [apiClient, usersTeam?.id]);

  const discoverPointOfInterest = async (pointOfInterestId: string) => {
    await apiClient.post(`/sonar/pointOfInterest/unlock`, {
      pointOfInterestId,
      teamId: usersTeam?.id,
      userId: usersTeam ? undefined : currentUser?.id,
      lat: location?.latitude?.toString(),
      lng: location?.longitude?.toString(),
    });

    setDiscoveries([...discoveries, {
      id: '',
      createdAt: new Date(),
      updatedAt: new Date(),
      teamId: usersTeam?.id,
      userId: usersTeam ? undefined : currentUser?.id,
      pointOfInterestId: pointOfInterestId,
    }]);
  };

  useEffect(() => {
    fetchDiscoveries();
  }, [usersTeam?.id]);

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
