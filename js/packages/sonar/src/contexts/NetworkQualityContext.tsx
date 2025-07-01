import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
  useCallback,
} from 'react';
import { useAPI } from '@poltergeist/contexts';

interface NetworkQualityState {
  isSlowConnection: boolean;
  averageLatency: number;
  consecutiveFailures: number;
  isMonitoring: boolean;
  lastCheckTime: number | null;
  connectionStatus: 'good' | 'poor' | 'critical' | 'unknown';
}

interface NetworkQualityContextType {
  networkState: NetworkQualityState;
  startMonitoring: () => void;
  stopMonitoring: () => void;
  acknowledgeSlowConnection: () => void;
  performLatencyCheck: () => Promise<number>;
}

const NetworkQualityContext = createContext<NetworkQualityContextType | null>(null);

interface NetworkQualityProviderProps {
  children: ReactNode;
}

const LATENCY_THRESHOLD_MS = 2000; // 2 seconds
const CRITICAL_LATENCY_THRESHOLD_MS = 5000; // 5 seconds
const FAILURE_THRESHOLD = 3; // consecutive failures before showing warning
const CHECK_INTERVAL_MS = 30000; // check every 30 seconds
const PING_ENDPOINT = '/sonar/health'; // lightweight endpoint for latency checks

export const NetworkQualityProvider: React.FC<NetworkQualityProviderProps> = ({ children }) => {
  const { apiClient } = useAPI();
  const [networkState, setNetworkState] = useState<NetworkQualityState>({
    isSlowConnection: false,
    averageLatency: 0,
    consecutiveFailures: 0,
    isMonitoring: false,
    lastCheckTime: null,
    connectionStatus: 'unknown',
  });

  const [intervalId, setIntervalId] = useState<NodeJS.Timeout | null>(null);

  const performLatencyCheck = useCallback(async (): Promise<number> => {
    const startTime = performance.now();
    
    try {
      // Try to make a lightweight request to measure latency
      await apiClient.get(PING_ENDPOINT);
      const endTime = performance.now();
      return endTime - startTime;
    } catch (error) {
      // If the request fails, return a very high latency value
      return CRITICAL_LATENCY_THRESHOLD_MS + 1000;
    }
  }, [apiClient]);

  const updateNetworkState = useCallback((latency: number, isSuccess: boolean) => {
    setNetworkState(prevState => {
      const newFailures = isSuccess ? 0 : prevState.consecutiveFailures + 1;
      const newAverageLatency = isSuccess 
        ? (prevState.averageLatency * 0.7 + latency * 0.3) // Moving average
        : prevState.averageLatency;

      let connectionStatus: 'good' | 'poor' | 'critical' | 'unknown' = 'good';
      
      if (!isSuccess || latency > CRITICAL_LATENCY_THRESHOLD_MS) {
        connectionStatus = 'critical';
      } else if (latency > LATENCY_THRESHOLD_MS) {
        connectionStatus = 'poor';
      }

      const shouldShowWarning = newFailures >= FAILURE_THRESHOLD || 
                              connectionStatus === 'critical' ||
                              (connectionStatus === 'poor' && newAverageLatency > LATENCY_THRESHOLD_MS);

      return {
        ...prevState,
        averageLatency: newAverageLatency,
        consecutiveFailures: newFailures,
        lastCheckTime: Date.now(),
        connectionStatus,
        isSlowConnection: shouldShowWarning,
      };
    });
  }, []);

  const performCheck = useCallback(async () => {
    try {
      const latency = await performLatencyCheck();
      updateNetworkState(latency, true);
    } catch (error) {
      updateNetworkState(0, false);
    }
  }, [performLatencyCheck, updateNetworkState]);

  const startMonitoring = useCallback(() => {
    if (intervalId) {
      clearInterval(intervalId);
    }

    setNetworkState(prevState => ({ ...prevState, isMonitoring: true }));
    
    // Perform initial check
    performCheck();
    
    // Set up interval for regular checks
    const id = setInterval(performCheck, CHECK_INTERVAL_MS);
    setIntervalId(id);
  }, [intervalId, performCheck]);

  const stopMonitoring = useCallback(() => {
    if (intervalId) {
      clearInterval(intervalId);
      setIntervalId(null);
    }
    setNetworkState(prevState => ({ ...prevState, isMonitoring: false }));
  }, [intervalId]);

  const acknowledgeSlowConnection = useCallback(() => {
    setNetworkState(prevState => ({ 
      ...prevState, 
      isSlowConnection: false,
      consecutiveFailures: 0,
    }));
  }, []);

  // Start monitoring when component mounts
  useEffect(() => {
    startMonitoring();

    // Cleanup on unmount
    return () => {
      if (intervalId) {
        clearInterval(intervalId);
      }
    };
  }, []);

  // Also monitor online/offline events
  useEffect(() => {
    const handleOnline = () => {
      performCheck();
    };

    const handleOffline = () => {
      updateNetworkState(0, false);
    };

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, [performCheck, updateNetworkState]);

  const value: NetworkQualityContextType = {
    networkState,
    startMonitoring,
    stopMonitoring,
    acknowledgeSlowConnection,
    performLatencyCheck,
  };

  return (
    <NetworkQualityContext.Provider value={value}>
      {children}
    </NetworkQualityContext.Provider>
  );
};

export const useNetworkQuality = (): NetworkQualityContextType => {
  const context = useContext(NetworkQualityContext);
  if (context === null) {
    throw new Error('useNetworkQuality must be used within a NetworkQualityProvider');
  }
  return context;
};