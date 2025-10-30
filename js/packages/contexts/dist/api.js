import { jsx as _jsx } from "react/jsx-runtime";
import { createContext, useContext, useMemo, useCallback, } from 'react';
import APIClient from '@poltergeist/api-client';
import { useLocation } from './location';
const APIContext = createContext({
    apiClient: new APIClient(process.env.REACT_APP_API_URL || ''),
});
export const APIProvider = ({ children }) => {
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
    return (_jsx(APIContext.Provider, Object.assign({ value: { apiClient } }, { children: children })));
};
export const useAPI = () => {
    const context = useContext(APIContext);
    if (context === null) {
        throw new Error('useAPI must be used within an APIProvider');
    }
    return context;
};
