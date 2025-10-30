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
    console.log('[DEBUG] API Provider - getLocation called, returning:', location);
    return location;
  }, [location]); // Include location in dependencies
  
  console.log('[DEBUG] API Provider - Current location:', location);
  
  // Recreate apiClient when location changes
  const apiClient = useMemo(() => {
    console.log('[DEBUG] API Provider - Creating new API client with location:', location);
    const client = new APIClient(baseURL, getLocation);
    console.log('[DEBUG] API Provider - API client created:', client);
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
