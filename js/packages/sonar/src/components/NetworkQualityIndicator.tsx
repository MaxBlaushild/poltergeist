import React from 'react';
import { useNetworkQuality } from '../contexts/NetworkQualityContext';
import { 
  WifiIcon, 
  SignalIcon,
  SignalSlashIcon,
  ExclamationTriangleIcon
} from '@heroicons/react/24/solid';

interface NetworkQualityIndicatorProps {
  showDetails?: boolean;
  className?: string;
}

export const NetworkQualityIndicator: React.FC<NetworkQualityIndicatorProps> = ({ 
  showDetails = false, 
  className = '' 
}) => {
  const { networkState } = useNetworkQuality();

  if (!networkState.isMonitoring) {
    return null;
  }

  const getIndicatorContent = () => {
    const { connectionStatus, averageLatency } = networkState;
    
    switch (connectionStatus) {
      case 'good':
        return {
          icon: <WifiIcon className="w-4 h-4 text-green-500" />,
          color: 'text-green-500',
          label: 'Good',
          bgColor: 'bg-green-50'
        };
      case 'poor':
        return {
          icon: <SignalIcon className="w-4 h-4 text-yellow-500" />,
          color: 'text-yellow-500',
          label: 'Slow',
          bgColor: 'bg-yellow-50'
        };
      case 'critical':
        return {
          icon: <SignalSlashIcon className="w-4 h-4 text-red-500" />,
          color: 'text-red-500',
          label: 'Poor',
          bgColor: 'bg-red-50'
        };
      default:
        return {
          icon: <ExclamationTriangleIcon className="w-4 h-4 text-gray-500" />,
          color: 'text-gray-500',
          label: 'Unknown',
          bgColor: 'bg-gray-50'
        };
    }
  };

  const { icon, color, label, bgColor } = getIndicatorContent();

  if (!showDetails) {
    // Minimal indicator - just the icon
    return (
      <div className={`flex items-center ${className}`} title={`Connection: ${label}`}>
        {icon}
      </div>
    );
  }

  // Detailed indicator with latency info
  return (
    <div className={`flex items-center space-x-2 px-2 py-1 rounded-lg ${bgColor} ${className}`}>
      {icon}
      <div className="flex flex-col text-xs">
        <span className={`font-medium ${color}`}>{label}</span>
        {networkState.averageLatency > 0 && (
          <span className="text-gray-600">
            {Math.round(networkState.averageLatency)}ms
          </span>
        )}
      </div>
    </div>
  );
};