import { useState, useEffect } from 'react';
export const useLocation = () => {
    const [location, setLocation] = useState({
        latitude: null,
        longitude: null,
        accuracy: null,
    });
    const [error, setError] = useState(null);
    useEffect(() => {
        if (navigator.geolocation) {
            navigator.geolocation.getCurrentPosition((position) => {
                setLocation({
                    latitude: position.coords.latitude,
                    longitude: position.coords.longitude,
                    accuracy: position.coords.accuracy,
                });
            }, (err) => {
                setError(err.message);
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
            // Cleanup the watch when the component is unmounted
            return () => {
                navigator.geolocation.clearWatch(watchId);
            };
        }
        else {
            setError('Geolocation is not supported by this browser.');
        }
    }, []);
    return { location, error };
};
export default useLocation;
