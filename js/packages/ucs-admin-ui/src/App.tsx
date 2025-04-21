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
import { createBrowserRouter, RouterProvider, useParams, Link, Outlet } from 'react-router-dom';
import { LoaderFunctionArgs, redirect } from 'react-router-dom';
import { ArenaProvider, InventoryProvider } from '@poltergeist/contexts';
import { Login } from './components/Login.tsx';
import Armory from './components/Armory.tsx';
import { Zones } from './components/Zones.tsx';
import { ZoneProvider } from './contexts/zones.tsx';
import { Zone } from './components/Zone.tsx';
import { Place } from './components/Place.tsx';
import { Tags } from './components/Tags.tsx';
import LocationArchetypes from './components/LocationArchetypes.tsx';
import { QuestArchetypesProvider } from './contexts/questArchetypes.tsx';
import { QuestArchetypeComponent } from './components/QuestArchetype.tsx';
import { ZoneQuestArchetypes } from './components/ZoneQuestArchetypes.tsx';

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

const Navigation = () => {
  const isLoggedIn = localStorage.getItem('token');
  
  if (!isLoggedIn) return null;

  return (
    <nav className="bg-gray-800 p-4">
      <div className="container mx-auto flex gap-4">
        <Link to="/" className="text-white hover:text-gray-300">Arenas</Link>
        <Link to="/armory" className="text-white hover:text-gray-300">Armory</Link>
        <Link to="/zones" className="text-white hover:text-gray-300">Zones</Link>
        <Link to="/tags" className="text-white hover:text-gray-300">Tags</Link>
        <Link to="/location-archetypes" className="text-white hover:text-gray-300">Location Archetypes</Link>
        <Link to="/quest-archetypes" className="text-white hover:text-gray-300">Quest Archetypes</Link>
        <Link to="/zone-quest-archetypes" className="text-white hover:text-gray-300">Zone Quest Archetypes</Link>
      </div>
    </nav>
  );
};

// Create a new Layout component that includes Navigation and Outlet
const Layout = () => {
  return (
    <div className="min-h-screen flex flex-col">
      <Navigation />
      <div className="flex-1">
        <Outlet />
      </div>
    </div>
  );
};

// Update the router configuration to use the Layout
const router = createBrowserRouter([
  {
    element: <Layout />,
    children: [
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
      {
        path: "/place/:id",
        element: <Place />,
        loader: onlyAuthenticated,
      },
      {
        path: "/tags",
        element: <Tags />,
        loader: onlyAuthenticated,
      },
      {
        path: "/location-archetypes",
        element: (
            <LocationArchetypes />
        ),
        loader: onlyAuthenticated,
      },
      {
        path: "/quest-archetypes",
        element: <QuestArchetypeComponent />,
        loader: onlyAuthenticated,
      },
      {
        path: "/zone-quest-archetypes",
        element: <ZoneQuestArchetypes />,
        loader: onlyAuthenticated,
      },
    ]
  }
]);

const App = () => {
  return (
    <APIProvider>
      <TagProvider>
        <ZoneProvider>
          <QuestArchetypesProvider>
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
          </QuestArchetypesProvider>
        </ZoneProvider>
      </TagProvider>
    </APIProvider>
  );
};

export default App;
