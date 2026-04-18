import React, {
  createContext,
  useContext,
  useMemo,
  useRef,
  useCallback,
  ReactNode,
} from 'react';
import APIClient from '@poltergeist/api-client';
import { useLocation } from './location';

interface APIContextType {
  apiClient: APIClient;
}

const APIContext = createContext<APIContextType | null>({
  apiClient: new APIClient(''),
});

interface APIProviderProps {
  children: ReactNode;
}

const getApiUrl = () => {
  return 'https://api.unclaimedstreets.com';
};

export const APIProvider: React.FC<APIProviderProps> = ({ children }) => {
  const baseURL = getApiUrl();
  const { location } = useLocation();
  const locationRef = useRef(location);

  locationRef.current = location;

  // Keep the API client stable while still reading the latest location header.
  const getLocation = useCallback(() => locationRef.current, []);

  const apiClient = useMemo(() => {
    return new APIClient(baseURL, getLocation);
  }, [baseURL, getLocation]);

  const value = useMemo(() => ({ apiClient }), [apiClient]);

  return (
    <APIContext.Provider value={value}>{children}</APIContext.Provider>
  );
};

export const useAPI = (): APIContextType => {
  const context = useContext(APIContext);
  if (context === null) {
    throw new Error('useAPI must be used within an APIProvider');
  }
  return context;
};
