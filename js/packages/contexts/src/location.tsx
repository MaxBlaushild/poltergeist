import { createContext, useContext, useState, useEffect, ReactNode, useCallback } from 'react';

interface Location {
    latitude: number | null;
    longitude: number | null;
    accuracy: number | null;
  }

interface LocationContextType {
  location: Location | null;
  isLoading: boolean;
  error: string | null;
  message: string | null;
  acknowledgeError: () => void;
}

const LocationContext = createContext<LocationContextType | undefined>(undefined);

export const LocationProvider = ({ children }: { children: ReactNode }) => {
  const [location, setLocation] = useState<Location | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [permissionStatus, setPermissionStatus] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const getSettingsInstructions = () => {
    // Improved browser/OS detection
    const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;
    const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent);
    const isAndroid = /Android/.test(navigator.userAgent);
    const isChrome = /Chrome/.test(navigator.userAgent) && !/Edge/.test(navigator.userAgent);
    const isFirefox = /Firefox/.test(navigator.userAgent);

    if (isAndroid && isChrome) {
      return 'To enable location: tap the lock icon in the address bar, or go to Settings > Site Settings > Location';
    } else if (isIOS) {
      return 'To enable location: go to Settings > Privacy > Location Services and enable for this browser';
    } else if (isMac) {
      return 'To enable location: go to System Preferences > Security & Privacy > Privacy > Location Services';
    } else if (isChrome) {
      return 'To enable location: click the lock icon in the address bar, or go to Chrome Settings > Privacy and Security > Site Settings > Location';
    } else if (isFirefox) {
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
      
      navigator.geolocation.getCurrentPosition(
        resolve,
        reject,
        {
          enableHighAccuracy: true,
          maximumAge: 0,
          timeout: 10000,
          ...options
        }
      );
    });
  };

  const calculateDistance = (lat1: number, lon1: number, lat2: number, lon2: number) => {
    const R = 6371e3; // Earth's radius in meters
    const φ1 = lat1 * Math.PI/180;
    const φ2 = lat2 * Math.PI/180;
    const Δφ = (lat2-lat1) * Math.PI/180;
    const Δλ = (lon2-lon1) * Math.PI/180;

    const a = Math.sin(Δφ/2) * Math.sin(Δφ/2) +
              Math.cos(φ1) * Math.cos(φ2) *
              Math.sin(Δλ/2) * Math.sin(Δλ/2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a));

    return R * c; // Distance in meters
  };

  const shouldUpdateLocation = useCallback((newLocation: Location, currentLocation: Location | null) => {
    if (!currentLocation?.latitude || !currentLocation?.longitude) return true;
    if (!newLocation.latitude || !newLocation.longitude) return false;

    const distance = calculateDistance(
      currentLocation.latitude,
      currentLocation.longitude,
      newLocation.latitude,
      newLocation.longitude
    );
    return distance >= 25; // Only update if distance is 25 meters or more
  }, []);

  useEffect(() => {
    const checkPermissionAndGetLocation = async () => {
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
        const status = await navigator.permissions.query({ name: 'geolocation' });
        setPermissionStatus(status.state);

        if (status.state === 'denied') {
          console.log('[DEBUG] Location Provider - Location access denied');
          setError(`Location access is denied. ${getSettingsInstructions()}`);
          setIsLoading(false);
          return;
        }

        if (status.state === 'granted' || status.state === 'prompt') {
          console.log('[DEBUG] Location Provider - Permission granted, getting location...');
          try {
            const position = await getLocation();
            const newLocation = {
              latitude: position?.coords?.latitude,
              longitude: position?.coords?.longitude,
              accuracy: position?.coords?.accuracy,
            };
            console.log('[DEBUG] Location Provider - Got location:', newLocation);
            if (shouldUpdateLocation(newLocation, location)) {
              setLocation(newLocation);
              console.log('[DEBUG] Location Provider - Location updated:', newLocation);
            }
            setError(null);
          } catch (locationError) {
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
          } finally {
            setIsLoading(false);
          }

          // Set up watch position with more lenient timeout
          const watchId = navigator.geolocation.watchPosition(
            (position) => {
              const newLocation = {
                latitude: position.coords.latitude,
                longitude: position.coords.longitude,
                accuracy: position.coords.accuracy,
              };
              if (shouldUpdateLocation(newLocation, location)) {
                setLocation(newLocation);
              }
              setError(null); // Clear any previous errors on success
            },
            (err) => {
              // Only update error if it's not a timeout
              if (err.code !== err.TIMEOUT) {
                setError(err.message);
              }
            },
            {
              enableHighAccuracy: true,
              maximumAge: 0,
              timeout: 20000, // More lenient timeout for watching
            }
          );

          return () => navigator.geolocation.clearWatch(watchId);
        } else {
          setError(`Location access is blocked. ${getSettingsInstructions()}`);
          setIsLoading(false);
        }
      } catch (err) {
        setError(`Error checking location permissions. ${getSettingsInstructions()}`);
        setIsLoading(false);
      }
    };

    checkPermissionAndGetLocation();
  }, []);

  const acknowledgeError = () => {  
    setError(null);
  };

  console.log('[DEBUG] Location Provider - Current state:', { location, error, isLoading, message });
  
  return (
    <LocationContext.Provider value={{ location, error, isLoading, message, acknowledgeError }}>
      {children}
    </LocationContext.Provider>
  );
};

export const useLocation = () => {
  const context = useContext(LocationContext);
  if (context === undefined) {
    return {
      location: null,
      error: null,
      isLoading: false,
      message: null,
      acknowledgeError: () => {}
    };
  }
  return context;
};
