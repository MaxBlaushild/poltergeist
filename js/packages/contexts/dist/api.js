import { jsx as _jsx } from "react/jsx-runtime";
import { createContext, useContext, useMemo, useRef, useCallback, } from 'react';
import APIClient from '@poltergeist/api-client';
import { useLocation } from './location';
const APIContext = createContext({
    apiClient: new APIClient(''),
});
const getApiUrl = () => {
    return 'https://api.unclaimedstreets.com';
};
export const APIProvider = ({ children }) => {
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
    return (_jsx(APIContext.Provider, Object.assign({ value: value }, { children: children })));
};
export const useAPI = () => {
    const context = useContext(APIContext);
    if (context === null) {
        throw new Error('useAPI must be used within an APIProvider');
    }
    return context;
};
