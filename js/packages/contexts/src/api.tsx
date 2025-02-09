import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from 'react';
import APIClient from '@poltergeist/api-client';

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
  const apiClient = new APIClient(baseURL);

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
