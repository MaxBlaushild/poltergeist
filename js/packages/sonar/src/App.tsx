import React from 'react';
import './App.css';
import { Logister } from '@poltergeist/components';
import {
  AuthProvider,
  APIProvider,
  MediaContextProvider,
} from '@poltergeist/contexts';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import { router } from './routes.ts';
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
} from '@poltergeist/contexts';
import mapboxgl from 'mapbox-gl';
import { LogContextProvider } from './contexts/LogContext.tsx';
import {
  PointOfInterestContext,
  PointOfInterestContextProvider,
} from './contexts/PointOfInterestContext.tsx';
import { SubmissionsContextProvider } from './contexts/SubmissionsContext.tsx';
import { DiscoveriesContextProvider } from './contexts/DiscoveriesContext.tsx';

mapboxgl.accessToken =
  'pk.eyJ1IjoibWF4YmxhdXNoaWxkIiwiYSI6ImNsenE2YWY2bDFmNnQyam9jOXJ4dHFocm4ifQ.tvO7DVEK_OLUyHfwDkUifA';

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
    <APIProvider>
      <MediaContextProvider>
        <UserProfileProvider>
          <InventoryProvider>
            <MatchContextProvider>
              <AuthProvider appName="Sonar" uriPrefix="/sonar">
                <ActivityContextProvider>
                  <LocationProvider>
                    <PointOfInterestContextProvider>
                      <LogContextProvider>
                        <DiscoveriesContextProvider>
                          <SubmissionsContextProvider>
                            <MapProvider>
                              <DndProvider
                              backend={TouchBackend}
                            options={{
                              enableTouchEvents: true,
                              enableMouseEvents: true,
                              enableHoverOutsideTarget: true,
                            }}
                          >
                            <RouterProvider router={router} />
                              </DndProvider>
                            </MapProvider>
                          </SubmissionsContextProvider>
                        </DiscoveriesContextProvider>
                      </LogContextProvider>
                    </PointOfInterestContextProvider>
                  </LocationProvider>
                </ActivityContextProvider>
              </AuthProvider>
            </MatchContextProvider>
          </InventoryProvider>
        </UserProfileProvider>
      </MediaContextProvider>
    </APIProvider>
  );
}

export default App;
