import { jsx as _jsx } from "react/jsx-runtime";
import { createContext, useContext, useState, useEffect } from 'react';
const LocationContext = createContext(undefined);
export const LocationProvider = ({ children }) => {
    const [location, setLocation] = useState(null);
    const [error, setError] = useState(null);
    const [isLoading, setIsLoading] = useState(true);
    useEffect(() => {
        if (navigator.geolocation) {
            navigator.geolocation.getCurrentPosition((position) => {
                setLocation({
                    latitude: position.coords.latitude,
                    longitude: position.coords.longitude,
                    accuracy: position.coords.accuracy,
                });
                setIsLoading(false);
            }, (err) => {
                setError(err.message);
                setIsLoading(false);
            });
            const watchId = navigator.geolocation.watchPosition((position) => {
                setLocation({
                    latitude: position.coords.latitude,
                    longitude: position.coords.longitude,
                    accuracy: position.coords.accuracy,
                });
            }, (err) => {
                setError(err.message);
            }, {
                enableHighAccuracy: true,
                maximumAge: 0,
                timeout: 5000,
            });
            return () => {
                navigator.geolocation.clearWatch(watchId);
            };
        }
        else {
            setError('Geolocation is not supported by this browser.');
        }
    }, []);
    return (_jsx(LocationContext.Provider, Object.assign({ value: { location, error, isLoading } }, { children: children })));
};
export const useLocation = () => {
    const context = useContext(LocationContext);
    if (context === undefined) {
        throw new Error('useLocation must be used within a LocationProvider');
    }
    return context;
};
