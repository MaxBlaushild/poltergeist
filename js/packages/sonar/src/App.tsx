import React from 'react';
// import './App.css';
import { Logister } from '@poltergeist/components';
import { AuthProvider, APIProvider } from '@poltergeist/contexts';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import { router } from './routes.ts';
import { ActivityContextProvider } from './contexts/ActivityContext.tsx';

function App() {
  return (
    <APIProvider>
      <AuthProvider appName="Sonar" uriPrefix="/sonar">
        <ActivityContextProvider>
          <RouterProvider router={router} />
        </ActivityContextProvider>
      </AuthProvider>
    </APIProvider>
  );
}

export default App;
