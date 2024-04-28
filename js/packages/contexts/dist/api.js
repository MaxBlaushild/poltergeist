import { jsx as _jsx } from 'react/jsx-runtime';
import { createContext, useContext } from 'react';
import APIClient from '@poltergeist/api-client';
const APIContext = createContext(null);
export const APIProvider = ({ children }) => {
  const baseURL = process.env.REACT_APP_API_URL || '';
  const apiClient = new APIClient(baseURL);
  return _jsx(
    APIContext.Provider,
    Object.assign({ value: { apiClient } }, { children: children })
  );
};
export const useAPI = () => {
  const context = useContext(APIContext);
  if (context === null) {
    throw new Error('useAPI must be used within an APIProvider');
  }
  return context;
};
