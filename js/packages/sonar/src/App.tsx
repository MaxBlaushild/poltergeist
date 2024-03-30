import React from 'react';
// import './App.css';
import { Logister } from '@poltergeist/components';
import { AuthProvider, APIProvider } from '@poltergeist/contexts';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import { router } from './routes.ts';


function App() {
  return (
    <div className="App">
      <APIProvider>
        <AuthProvider appName='Sonar' uriPrefix='/sonar'>
          <RouterProvider router={router} />
        </AuthProvider>
      </APIProvider>
    </div>
  );
}

export default App;
