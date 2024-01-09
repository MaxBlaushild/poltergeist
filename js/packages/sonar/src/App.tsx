import React from 'react';
// import './App.css';
import { Logister } from '@poltergeist/components';
import { AuthProvider } from '@poltergeist/contexts';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import { router } from './routes.ts';


function App() {
  return (
    <div className="App">
      <AuthProvider appName='Sonar' uriPrefix='/sonar'>
        <RouterProvider router={router} />
      </AuthProvider>
    </div>
  );
}

export default App;
