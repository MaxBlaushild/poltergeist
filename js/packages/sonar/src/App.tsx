import React from 'react';
// import './App.css';
import { Logister } from '@poltergeist/components';
import { AuthProvider, APIProvider } from '@poltergeist/contexts';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import { router } from './routes.ts';
import { ActivityContextProvider } from './contexts/ActivityContext.tsx';
import { UserProfileProvider } from './contexts/UserProfileContext.tsx';
import { MatchContextProvider } from './contexts/MatchContext.tsx';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import { TouchBackend } from 'react-dnd-touch-backend';
import { MultiBackend, TouchTransition } from 'react-dnd-multi-backend';
import { MediaContextProvider } from './contexts/MediaContext.tsx';

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
          <MatchContextProvider>
            <AuthProvider appName="Sonar" uriPrefix="/sonar">
              <ActivityContextProvider>
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
              </ActivityContextProvider>
            </AuthProvider>
          </MatchContextProvider>
        </UserProfileProvider>
      </MediaContextProvider>
    </APIProvider>
  );
}

export default App;
