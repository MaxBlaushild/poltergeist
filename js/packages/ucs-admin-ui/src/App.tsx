import React from 'react';
import './App.css';
import {
  AuthProvider,
  APIProvider,
  MediaContextProvider,
  TagProvider,
} from '@poltergeist/contexts';
import { Arenas } from './components/Arenas.tsx';
import { Arena } from './components/Arena.tsx';
import { createBrowserRouter, RouterProvider, useParams } from 'react-router-dom';
import { LoaderFunctionArgs, redirect } from 'react-router-dom';
import { ArenaProvider, InventoryProvider } from '@poltergeist/contexts';
import { Login } from './components/Login.tsx';
import Armory from './components/Armory.tsx';
import { Zones } from './components/Zones.tsx';
import { ZoneProvider } from './contexts/zones.tsx';
import { Zone } from './components/Zone.tsx';

function onlyAuthenticated({ request }: LoaderFunctionArgs) {
  if (!localStorage.getItem('token')) {
    let params = new URLSearchParams();
    params.set('from', new URL(request.url).pathname);
    return redirect('/login?' + params.toString());
  }
  return null;
}

function onlyUnauthenticated({ request }: LoaderFunctionArgs) {
  if (localStorage.getItem('token')) {
    return redirect('/');
  }
  return null;
}

const ArenaWrapper = ({ children }: { children: React.ReactNode }) => {
  const { id } = useParams();
  
  return (
    <ArenaProvider arenaId={id}>
      {children}
    </ArenaProvider>
  );
};

const router = createBrowserRouter([
  {
    path: "/",
    element: <Arenas />,
    loader: onlyAuthenticated,
  },
  {
    path: "/login",
    element: <Login />,
    loader: onlyUnauthenticated,
  },
  {
    path: "/arena/:id",
    element: (
      <ArenaWrapper>
        <Arena />
      </ArenaWrapper>
    ),
    loader: onlyAuthenticated,
  },
  {
    path: "/armory",
    element: <Armory />,
    loader: onlyAuthenticated,
  },
  {
    path: "/zones",
    element: (
        <Zones />
    ),
    loader: onlyAuthenticated,
  },
  {
    path: "/zones/:id",
    element: <Zone />,
    loader: onlyAuthenticated,
  },
]);

const App = () => {
  return (
    <APIProvider>
      <TagProvider>
        <ZoneProvider>
        <MediaContextProvider>
          <AuthProvider
            appName="UCS Admin Dashboard"
          uriPrefix="/sonar"
        >
          <InventoryProvider>
            <RouterProvider router={router} />
          </InventoryProvider>
        </AuthProvider>
      </MediaContextProvider>
      </ZoneProvider>
      </TagProvider>
    </APIProvider>
  );
};



export default App;
