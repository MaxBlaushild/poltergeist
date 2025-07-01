import React from 'react';
import { useNetworkQuality } from '../contexts/NetworkQualityContext';
import { Modal, ModalSize } from './shared/Modal';
import { 
  WifiIcon, 
  ExclamationTriangleIcon,
  SignalSlashIcon,
  ArrowPathIcon
} from '@heroicons/react/24/solid';

export const LowLatencyWarning = () => {
  const { networkState, acknowledgeSlowConnection, performLatencyCheck } = useNetworkQuality();

  if (!networkState.isSlowConnection) {
    return null;
  }

  const handleRetryConnection = async () => {
    try {
      await performLatencyCheck();
    } catch (error) {
      // Error is handled in the context
    }
  };

  const getWarningContent = () => {
    const { connectionStatus, averageLatency, consecutiveFailures } = networkState;
    
    if (connectionStatus === 'critical' || consecutiveFailures >= 3) {
      return {
        icon: <SignalSlashIcon className="w-16 h-16 text-red-500 mx-auto mb-4" />,
        title: "Connection Issues Detected",
        message: "We're experiencing significant connectivity problems that may affect your Sonar experience.",
        severity: "critical" as const
      };
    } else {
      return {
        icon: <ExclamationTriangleIcon className="w-16 h-16 text-yellow-500 mx-auto mb-4" />,
        title: "Slow Connection Detected", 
        message: "Your current connection may not provide the best Sonar experience.",
        severity: "warning" as const
      };
    }
  };

  const { icon, title, message, severity } = getWarningContent();

  return (
    <Modal size={ModalSize.HERO}>
      <div className="flex flex-col items-center p-6 max-w-md">
        {icon}
        
        <h2 className="text-xl font-bold mb-4 text-center">
          {title}
        </h2>
        
        <p className="text-gray-700 text-center mb-4">
          {message}
        </p>

        {/* Connection Details */}
        <div className="bg-gray-100 rounded-lg p-4 mb-6 w-full">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium text-gray-600">Connection Status:</span>
            <span className={`text-sm font-semibold ${
              networkState.connectionStatus === 'critical' ? 'text-red-600' :
              networkState.connectionStatus === 'poor' ? 'text-yellow-600' :
              'text-green-600'
            }`}>
              {networkState.connectionStatus.toUpperCase()}
            </span>
          </div>
          
          {networkState.averageLatency > 0 && (
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium text-gray-600">Average Response Time:</span>
              <span className="text-sm font-semibold text-gray-800">
                {Math.round(networkState.averageLatency)}ms
              </span>
            </div>
          )}

          {networkState.consecutiveFailures > 0 && (
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-gray-600">Failed Attempts:</span>
              <span className="text-sm font-semibold text-red-600">
                {networkState.consecutiveFailures}
              </span>
            </div>
          )}
        </div>

        {/* Recommendations */}
        <div className="bg-blue-50 border-l-4 border-blue-400 p-4 mb-6 w-full">
          <h3 className="text-sm font-medium text-blue-800 mb-2">
            <WifiIcon className="w-4 h-4 inline mr-1" />
            Recommendations:
          </h3>
          <ul className="text-sm text-blue-700 space-y-1">
            {severity === 'critical' ? (
              <>
                <li>• Check your internet connection</li>
                <li>• Try moving to a location with better signal</li>
                <li>• Consider upgrading your data plan</li>
                <li>• Contact your service provider if issues persist</li>
              </>
            ) : (
              <>
                <li>• Consider upgrading to a faster data plan</li>
                <li>• Move closer to your WiFi router</li>
                <li>• Close other apps using bandwidth</li>
                <li>• Try switching between WiFi and mobile data</li>
              </>
            )}
          </ul>
        </div>

        {/* Action Buttons */}
        <div className="flex flex-col sm:flex-row gap-3 w-full">
          <button
            onClick={handleRetryConnection}
            className="flex items-center justify-center bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg"
          >
            <ArrowPathIcon className="w-4 h-4 mr-2" />
            Test Connection
          </button>
          
          <button
            onClick={acknowledgeSlowConnection}
            className="bg-gray-300 hover:bg-gray-400 text-gray-800 px-4 py-2 rounded-lg"
          >
            Continue Anyway
          </button>
        </div>

        {/* Additional Help */}
        <div className="mt-4 text-center">
          <p className="text-xs text-gray-500 mb-2">
            Need better service for Sonar?
          </p>
          <button
            onClick={() => window.open('tel:*611', '_self')}
            className="text-blue-600 hover:text-blue-800 text-sm underline bg-transparent border-none"
          >
            Contact Your Provider (*611)
          </button>
        </div>
      </div>
    </Modal>
  );
};