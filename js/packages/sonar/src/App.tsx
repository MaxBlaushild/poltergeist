import React from 'react';
// import './App.css';
import { Logister } from '@poltergeist/components';
import { AuthProvider, APIProvider } from '@poltergeist/contexts';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import { router } from './routes.ts';
import { ActivityContextProvider } from './contexts/ActivityContext.tsx';
import { UserProfileProvider } from './contexts/UserProfileContext.tsx';
import { MatchContextProvider } from './contexts/MatchContext.tsx';

function App() {
  return (
    <APIProvider>
      <MatchContextProvider>
        <UserProfileProvider>
          <AuthProvider appName="Sonar" uriPrefix="/sonar">
            <ActivityContextProvider>
              <RouterProvider router={router} />
            </ActivityContextProvider>
          </AuthProvider>
        </UserProfileProvider>
      </MatchContextProvider>
    </APIProvider>
  );
}

export default App;
