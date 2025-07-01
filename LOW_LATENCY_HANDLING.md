# Low Latency Handling for Sonar UI

## Overview

This implementation provides graceful handling of low latency and poor network conditions in the Sonar application. When users experience slow connections or network issues, the app will display helpful warnings and suggestions to improve their service.

## Components

### 1. NetworkQualityContext (`js/packages/sonar/src/contexts/NetworkQualityContext.tsx`)

A React context that monitors network quality in real-time:

**Features:**
- **Automatic Latency Monitoring**: Checks connection quality every 30 seconds
- **Real-time Status Updates**: Tracks connection status (good/poor/critical/unknown)
- **Failure Detection**: Counts consecutive failures to detect persistent issues
- **Moving Average Calculation**: Smooths latency measurements over time
- **Online/Offline Detection**: Responds to browser network events

**Configuration:**
```typescript
const LATENCY_THRESHOLD_MS = 2000; // 2 seconds
const CRITICAL_LATENCY_THRESHOLD_MS = 5000; // 5 seconds  
const FAILURE_THRESHOLD = 3; // consecutive failures before showing warning
const CHECK_INTERVAL_MS = 30000; // check every 30 seconds
```

**State Interface:**
```typescript
interface NetworkQualityState {
  isSlowConnection: boolean;
  averageLatency: number;
  consecutiveFailures: number;
  isMonitoring: boolean;
  lastCheckTime: number | null;
  connectionStatus: 'good' | 'poor' | 'critical' | 'unknown';
}
```

### 2. LowLatencyWarning Component (`js/packages/sonar/src/components/LowLatencyWarning.tsx`)

A modal component that displays when network issues are detected:

**Features:**
- **Context-Aware Messages**: Different messages for poor vs critical connections
- **Connection Details**: Shows current status, response time, and failure count
- **Actionable Recommendations**: Specific suggestions based on connection severity
- **Interactive Actions**: 
  - Test Connection button to retry latency check
  - Continue Anyway to dismiss warning
  - Contact Provider link (*611) for immediate help
- **Modern UI**: Clean, responsive design with proper color coding

**Warning Levels:**
- **Warning (Yellow)**: Latency > 2 seconds - suggests optimization tips
- **Critical (Red)**: Latency > 5 seconds or 3+ consecutive failures - suggests immediate action

### 3. NetworkQualityIndicator Component (`js/packages/sonar/src/components/NetworkQualityIndicator.tsx`)

A small, non-intrusive indicator that can be placed anywhere in the UI:

**Features:**
- **Minimal Mode**: Just an icon showing connection status
- **Detailed Mode**: Shows status and latency information
- **Color-Coded**: Green (good), Yellow (slow), Red (poor), Gray (unknown)
- **Tooltip Support**: Hover for quick status information

### 4. Health Endpoint (`go/sonar/internal/server/server.go`)

A lightweight backend endpoint for latency testing:

```go
GET /sonar/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "service": "sonar"
}
```

## Integration

### App Level Integration

The NetworkQualityProvider is integrated at the app level in `App.tsx`:

```typescript
function App() {
  return (
    <APIProvider>
      <NetworkQualityProvider>  {/* Added here */}
        <MediaContextProvider>
          {/* ... other providers ... */}
          <RouterProvider router={router} />
          <LocationError />
          <LowLatencyWarning />  {/* Added here */}
        </MediaContextProvider>
      </NetworkQualityProvider>
    </APIProvider>
  );
}
```

### API Client Enhancement

The API client now includes a 10-second timeout for better error handling:

```typescript
this.client = axios.create({
  baseURL,
  timeout: 10000, // 10 second timeout for network quality monitoring
});
```

## Usage

### Using the Hook

```typescript
import { useNetworkQuality } from '../contexts/NetworkQualityContext';

function MyComponent() {
  const { 
    networkState, 
    startMonitoring, 
    stopMonitoring, 
    acknowledgeSlowConnection,
    performLatencyCheck 
  } = useNetworkQuality();

  // Access network state
  console.log(networkState.connectionStatus); // 'good' | 'poor' | 'critical' | 'unknown'
  console.log(networkState.averageLatency); // number in milliseconds
  console.log(networkState.isSlowConnection); // boolean
}
```

### Adding the Indicator

```typescript
import { NetworkQualityIndicator } from './NetworkQualityIndicator';

function Header() {
  return (
    <header>
      <h1>My App</h1>
      {/* Minimal indicator */}
      <NetworkQualityIndicator />
      
      {/* Detailed indicator */}
      <NetworkQualityIndicator showDetails={true} />
    </header>
  );
}
```

## How It Works

1. **Initialization**: When the app starts, NetworkQualityProvider begins monitoring
2. **Health Checks**: Every 30 seconds, makes a request to `/sonar/health` endpoint
3. **Latency Calculation**: Measures time between request start and completion
4. **Status Assessment**: Categorizes connection based on latency and failure count
5. **Warning Display**: Shows LowLatencyWarning modal when issues detected
6. **User Action**: User can test connection, acknowledge warning, or contact provider

## Connection Quality Thresholds

| Status | Latency | Failures | Action |
|--------|---------|----------|--------|
| Good | < 2s | 0 | None |
| Poor | 2s - 5s | 0-2 | Show optimization tips |
| Critical | > 5s or any | 3+ | Show urgent warnings |

## Benefits

1. **User Awareness**: Users understand when their connection is the issue
2. **Actionable Guidance**: Specific recommendations for improvement
3. **Reduced Support Load**: Users can self-diagnose network issues
4. **Better UX**: Graceful degradation instead of unexplained failures
5. **Provider Contact**: Direct path to upgrade service when needed

## Customization

### Adjusting Thresholds

Modify constants in `NetworkQualityContext.tsx`:

```typescript
const LATENCY_THRESHOLD_MS = 3000; // Increase threshold to 3 seconds
const CHECK_INTERVAL_MS = 60000; // Check every minute instead
```

### Custom Styling

The components use Tailwind CSS classes and can be customized:

```typescript
<NetworkQualityIndicator 
  showDetails={true} 
  className="my-custom-class"
/>
```

### Different Health Endpoint

Change the endpoint in `NetworkQualityContext.tsx`:

```typescript
const PING_ENDPOINT = '/api/ping'; // Use different endpoint
```

## Future Enhancements

1. **Bandwidth Testing**: Add download/upload speed measurements
2. **Regional Detection**: Different thresholds based on geographic location
3. **Connection Type Detection**: WiFi vs cellular specific handling
4. **Historical Tracking**: Store connection quality over time
5. **Predictive Warnings**: Warn before critical issues occur
6. **A/B Testing**: Test different threshold values and UI designs