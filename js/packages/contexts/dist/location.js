var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
import { jsx as _jsx } from "react/jsx-runtime";
import { createContext, useContext, useState, useEffect, useCallback } from 'react';
const LocationContext = createContext(undefined);
export const LocationProvider = ({ children }) => {
    const [location, setLocation] = useState(null);
    const [error, setError] = useState(null);
    const [isLoading, setIsLoading] = useState(true);
    const [permissionStatus, setPermissionStatus] = useState(null);
    const [message, setMessage] = useState(null);
    const getSettingsInstructions = () => {
        // Improved browser/OS detection
        const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;
        const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent);
        const isAndroid = /Android/.test(navigator.userAgent);
        const isChrome = /Chrome/.test(navigator.userAgent) && !/Edge/.test(navigator.userAgent);
        const isFirefox = /Firefox/.test(navigator.userAgent);
        if (isAndroid && isChrome) {
            return 'To enable location: tap the lock icon in the address bar, or go to Settings > Site Settings > Location';
        }
        else if (isIOS) {
            return 'To enable location: go to Settings > Privacy > Location Services and enable for this browser';
        }
        else if (isMac) {
            return 'To enable location: go to System Preferences > Security & Privacy > Privacy > Location Services';
        }
        else if (isChrome) {
            return 'To enable location: click the lock icon in the address bar, or go to Chrome Settings > Privacy and Security > Site Settings > Location';
        }
        else if (isFirefox) {
            return 'To enable location: click the lock icon in the address bar, or go to Firefox Settings > Privacy & Security > Permissions > Location';
        }
        return 'Please enable location services in your browser settings to use this feature';
    };
    const getLocation = (options = {}) => {
        return new Promise((resolve, reject) => {
            if (!navigator.geolocation) {
                reject(new Error('Geolocation is not supported by this browser'));
                return;
            }
            navigator.geolocation.getCurrentPosition(resolve, reject, Object.assign({ enableHighAccuracy: true, maximumAge: 0, timeout: 10000 }, options));
        });
    };
    const calculateDistance = (lat1, lon1, lat2, lon2) => {
        const R = 6371e3; // Earth's radius in meters
        const φ1 = lat1 * Math.PI / 180;
        const φ2 = lat2 * Math.PI / 180;
        const Δφ = (lat2 - lat1) * Math.PI / 180;
        const Δλ = (lon2 - lon1) * Math.PI / 180;
        const a = Math.sin(Δφ / 2) * Math.sin(Δφ / 2) +
            Math.cos(φ1) * Math.cos(φ2) *
                Math.sin(Δλ / 2) * Math.sin(Δλ / 2);
        const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
        return R * c; // Distance in meters
    };
    const shouldUpdateLocation = useCallback((newLocation, currentLocation) => {
        if (!(currentLocation === null || currentLocation === void 0 ? void 0 : currentLocation.latitude) || !(currentLocation === null || currentLocation === void 0 ? void 0 : currentLocation.longitude))
            return true;
        if (!newLocation.latitude || !newLocation.longitude)
            return false;
        const distance = calculateDistance(currentLocation.latitude, currentLocation.longitude, newLocation.latitude, newLocation.longitude);
        return distance >= 25; // Only update if distance is 25 meters or more
    }, []);
    useEffect(() => {
        const checkPermissionAndGetLocation = () => __awaiter(void 0, void 0, void 0, function* () {
            var _a, _b, _c;
            try {
                // Add HTTPS check
                if (window.location.protocol !== 'https:' && window.location.hostname !== 'localhost') {
                    setError('Geolocation requires a secure HTTPS connection');
                    setIsLoading(false);
                    return;
                }
                if (!navigator.geolocation) {
                    setError('Geolocation is not supported by this browser');
                    setIsLoading(false);
                    return;
                }
                // Check permissions first
                const status = yield navigator.permissions.query({ name: 'geolocation' });
                setPermissionStatus(status.state);
                if (status.state === 'denied') {
                    setError(`Location access is denied. ${getSettingsInstructions()}`);
                    setIsLoading(false);
                    return;
                }
                if (status.state === 'granted' || status.state === 'prompt') {
                    try {
                        const position = yield getLocation();
                        const newLocation = {
                            latitude: (_a = position === null || position === void 0 ? void 0 : position.coords) === null || _a === void 0 ? void 0 : _a.latitude,
                            longitude: (_b = position === null || position === void 0 ? void 0 : position.coords) === null || _b === void 0 ? void 0 : _b.longitude,
                            accuracy: (_c = position === null || position === void 0 ? void 0 : position.coords) === null || _c === void 0 ? void 0 : _c.accuracy,
                        };
                        if (shouldUpdateLocation(newLocation, location)) {
                            setLocation(newLocation);
                        }
                        setError(null);
                    }
                    catch (locationError) {
                        // Handle specific geolocation errors
                        let errorMessage = 'Error getting location. ';
                        switch (locationError.code) {
                            case locationError.TIMEOUT:
                                errorMessage += 'Request timed out. Please try again.';
                                break;
                            case locationError.POSITION_UNAVAILABLE:
                                errorMessage += 'Location information is unavailable.';
                                break;
                            case locationError.PERMISSION_DENIED:
                                errorMessage += getSettingsInstructions();
                                break;
                            default:
                                errorMessage += locationError.message;
                        }
                        setError(errorMessage);
                    }
                    finally {
                        setIsLoading(false);
                    }
                    // Set up watch position with more lenient timeout
                    const watchId = navigator.geolocation.watchPosition((position) => {
                        const newLocation = {
                            latitude: position.coords.latitude,
                            longitude: position.coords.longitude,
                            accuracy: position.coords.accuracy,
                        };
                        if (shouldUpdateLocation(newLocation, location)) {
                            setLocation(newLocation);
                        }
                        setError(null); // Clear any previous errors on success
                    }, (err) => {
                        // Only update error if it's not a timeout
                        if (err.code !== err.TIMEOUT) {
                            setError(err.message);
                        }
                    }, {
                        enableHighAccuracy: true,
                        maximumAge: 0,
                        timeout: 20000, // More lenient timeout for watching
                    });
                    return () => navigator.geolocation.clearWatch(watchId);
                }
                else {
                    setError(`Location access is blocked. ${getSettingsInstructions()}`);
                    setIsLoading(false);
                }
            }
            catch (err) {
                setError(`Error checking location permissions. ${getSettingsInstructions()}`);
                setIsLoading(false);
            }
        });
        checkPermissionAndGetLocation();
    }, []);
    const acknowledgeError = () => {
        setError(null);
    };
    return (_jsx(LocationContext.Provider, Object.assign({ value: { location, error, isLoading, message, acknowledgeError } }, { children: children })));
};
export const useLocation = () => {
    const context = useContext(LocationContext);
    if (context === undefined) {
        return {
            location: null,
            error: null,
            isLoading: false,
            message: null,
            acknowledgeError: () => { }
        };
    }
    return context;
};
