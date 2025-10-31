import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  useMemo,
  useCallback,
  ReactNode,
} from 'react';
import APIClient from '@poltergeist/api-client';
import { useLocation } from './location';

interface APIContextType {
  apiClient: APIClient;
}

const APIContext = createContext<APIContextType | null>({
  apiClient: new APIClient(process.env.REACT_APP_API_URL || ''),
});

interface APIProviderProps {
  children: ReactNode;
}

export const APIProvider: React.FC<APIProviderProps> = ({ children }) => {
  const baseURL = process.env.REACT_APP_API_URL || '';
  const { location } = useLocation();
  
  // Create stable getLocation function that always returns current location
  const getLocation = useCallback(() => {
    return location;
  }, [location]); // Include location in dependencies
  
  
  // Recreate apiClient when location changes
  const apiClient = useMemo(() => {
    const client = new APIClient(baseURL, getLocation);
    return client;
  }, [baseURL, getLocation, location]);

  return (
    <APIContext.Provider value={{ apiClient }}>{children}</APIContext.Provider>
  );
};

export const useAPI = (): APIContextType => {
  const context = useContext(APIContext);
  if (context === null) {
    throw new Error('useAPI must be used within an APIProvider');
  }
  return context;
};
