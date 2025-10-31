import React from 'react';
import './App.css';
import { Logister } from '@poltergeist/components';
import {
  AuthProvider,
  APIProvider,
  MediaContextProvider,
} from '@poltergeist/contexts';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import { router } from './routes.tsx';
import { ActivityContextProvider } from './contexts/ActivityContext.tsx';
import { UserProfileProvider } from './contexts/UserProfileContext.tsx';
import { MatchContextProvider } from './contexts/MatchContext.tsx';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import { TouchBackend } from 'react-dnd-touch-backend';
import { MultiBackend, TouchTransition } from 'react-dnd-multi-backend';
import {
  InventoryProvider,
  LocationProvider,
  MapProvider,
  ZoneProvider,
} from '@poltergeist/contexts';
import mapboxgl from 'mapbox-gl';
import { LogContextProvider } from './contexts/LogContext.tsx';
import {
  PointOfInterestContext,
  PointOfInterestContextProvider,
} from './contexts/PointOfInterestContext.tsx';
import { SubmissionsContextProvider } from './contexts/SubmissionsContext.tsx';
import { DiscoveriesContextProvider } from './contexts/DiscoveriesContext.tsx';
import { LocationError } from './components/LocationError.tsx';

mapboxgl.accessToken =
  'REDACTED';

const HTML5toTouch = {
  backends: [
    {
      backend: HTML5Backend,
    },
    {
      backend: TouchBackend,
      options: { enableMouseEvents: true },
      preview: true,
      transition: TouchTransition,
    },
  ],
};

function App() {
  return (
    <LocationProvider>
      <APIProvider>
        <MediaContextProvider>
          <UserProfileProvider>
            <ZoneProvider>
            <InventoryProvider>
                <AuthProvider appName="Sonar" uriPrefix="/sonar">
                    <LogContextProvider>
                      <DiscoveriesContextProvider>
                        <SubmissionsContextProvider>
                            <DndProvider
                              backend={TouchBackend}
                              options={{
                                enableTouchEvents: true,
                                enableMouseEvents: true,
                                enableHoverOutsideTarget: true,
                              }}
                            >
                              <RouterProvider router={router} />
                              <LocationError />
                            </DndProvider>
                        </SubmissionsContextProvider>
                      </DiscoveriesContextProvider>
                    </LogContextProvider>
                </AuthProvider>
            </InventoryProvider>
            </ZoneProvider>
          </UserProfileProvider>
        </MediaContextProvider>
      </APIProvider>
    </LocationProvider>
  );
}

export default App;
