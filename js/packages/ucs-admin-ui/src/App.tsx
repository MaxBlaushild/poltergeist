import React, { useEffect, useMemo, useState } from 'react';
import './App.css';
import {
  AuthProvider,
  APIProvider,
  MediaContextProvider,
  TagProvider,
  MapProvider,
  LocationProvider,
} from '@poltergeist/contexts';
import { Arenas } from './components/Arenas.tsx';
import { Arena } from './components/Arena.tsx';
import {
  createBrowserRouter,
  RouterProvider,
  useParams,
  Link,
  Outlet,
  useLocation,
} from 'react-router-dom';
import { LoaderFunctionArgs, redirect } from 'react-router-dom';
import {
  ArenaProvider,
  InventoryProvider,
  ZoneProvider,
} from '@poltergeist/contexts';
import { Login } from './components/Login.tsx';
import Armory from './components/Armory.tsx';
import { Zones } from './components/Zones.tsx';
import { Zone } from './components/Zone.tsx';
import { Districts } from './components/Districts.tsx';
import { DistrictEditor } from './components/District.tsx';
import { Place } from './components/Place.tsx';
import { Tags } from './components/Tags.tsx';
import LocationArchetypes from './components/LocationArchetypes.tsx';
import { QuestArchetypesProvider } from './contexts/questArchetypes.tsx';
import { QuestArchetypeComponent } from './components/QuestArchetype.tsx';
import QuestArchetypeGenerator from './components/QuestArchetypeGenerator.tsx';
import { ZoneQuestArchetypes } from './components/ZoneQuestArchetypes.tsx';
import { Users } from './components/Users.tsx';
import { Characters } from './components/Characters.tsx';
import { Parties } from './components/Parties.tsx';
import { InventoryItems } from './components/InventoryItems.tsx';
import { Bases } from './components/Bases.tsx';
import { TreasureChests } from './components/TreasureChests.tsx';
import NewUserStarterConfig from './components/NewUserStarterConfig.tsx';
import { FeteRooms } from './components/FeteRooms.tsx';
import { FeteTeams } from './components/FeteTeams.tsx';
import { FeteRoomLinkedListTeams } from './components/FeteRoomLinkedListTeams.tsx';
import { FeteRoomTeams } from './components/FeteRoomTeams.tsx';
import { UtilityClosetPuzzleAdmin } from './components/UtilityClosetPuzzle.tsx';
import { FlaggedPhotos } from './components/FlaggedPhotos.tsx';
import { PointOfInterest } from './components/PointOfInterest.tsx';
import { PointOfInterestEditor } from './components/PointOfInterestEditor.tsx';
import { Quests } from './components/Quests.tsx';
import { InsiderTrades } from './components/InsiderTrades.tsx';
import Feedback from './components/Feedback.tsx';
import ZoneSeedJobs from './components/ZoneSeedJobs.tsx';
import ZoneTagJobs from './components/ZoneTagJobs.tsx';
import { Scenarios } from './components/Scenarios.tsx';
import { Challenges } from './components/Challenges.tsx';
import ScenarioTemplates from './components/ScenarioTemplates.tsx';
import ChallengeTemplates from './components/ChallengeTemplates.tsx';
import Spells from './components/Spells.tsx';
import Monsters from './components/Monsters.tsx';
import HealingFountains from './components/HealingFountains.tsx';
import Tutorial from './components/Tutorial.tsx';
import { AdminHome } from './components/AdminHome.tsx';
import {
  adminNavItemMatchesPath,
  adminNavigationGroups,
  featuredAdminNavItems,
  findActiveAdminNavItem,
} from './adminNavigation.ts';

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

  return <ArenaProvider arenaId={id}>{children}</ArenaProvider>;
};

const Navigation = ({
  pathname,
  onNavigate,
}: {
  pathname: string;
  onNavigate?: () => void;
}) => {
  const [query, setQuery] = useState('');

  const filteredGroups = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    if (!normalizedQuery) {
      return adminNavigationGroups;
    }
    return adminNavigationGroups
      .map((group) => {
        const groupMatch =
          group.label.toLowerCase().includes(normalizedQuery) ||
          group.description.toLowerCase().includes(normalizedQuery);
        return {
          ...group,
          items: groupMatch
            ? group.items
            : group.items.filter(
                (item) =>
                  item.label.toLowerCase().includes(normalizedQuery) ||
                  item.description.toLowerCase().includes(normalizedQuery)
              ),
        };
      })
      .filter((group) => group.items.length > 0);
  }, [query]);

  return (
    <div className="admin-sidebar__inner">
      <Link to="/" className="admin-brand" onClick={onNavigate}>
        <div className="admin-brand__mark">US</div>
        <div>
          <div className="admin-brand__eyebrow">Unclaimed Streets</div>
          <div className="admin-brand__title">Admin Dashboard</div>
        </div>
      </Link>

      <div className="admin-sidebar__search">
        <label htmlFor="admin-nav-search" className="admin-sidebar__search-label">
          Find a tool
        </label>
        <input
          id="admin-nav-search"
          type="search"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Zones, quests, users..."
          className="admin-sidebar__search-input"
        />
      </div>

      <div className="admin-nav-home">
        <Link
          to="/"
          onClick={onNavigate}
          className={`admin-nav-link admin-nav-link--home ${
            pathname === '/' ? 'is-active' : ''
          }`}
        >
          <span>Home</span>
          <small>Overview, workflows, and grouped quick access.</small>
        </Link>
      </div>

      <div className="admin-nav-groups">
        {filteredGroups.map((group) => (
          <section key={group.id} className="admin-nav-group">
            <div className="admin-nav-group__header">
              <div className="admin-nav-group__title">{group.label}</div>
              <div className="admin-nav-group__description">
                {group.description}
              </div>
            </div>
            <div className="admin-nav-group__links">
              {group.items.map((item) => {
                const isActive = adminNavItemMatchesPath(item, pathname);
                return (
                  <Link
                    key={item.id}
                    to={item.path}
                    onClick={onNavigate}
                    className={`admin-nav-link ${isActive ? 'is-active' : ''}`}
                  >
                    <span>{item.label}</span>
                    <small>{item.description}</small>
                  </Link>
                );
              })}
            </div>
          </section>
        ))}
        {filteredGroups.length === 0 && (
          <div className="admin-nav-empty">
            No admin areas match <strong>{query}</strong>.
          </div>
        )}
      </div>
    </div>
  );
};

const Layout = () => {
  const location = useLocation();
  const isLoggedIn = Boolean(localStorage.getItem('token'));
  const [navOpen, setNavOpen] = useState(false);
  const activeNavItem = findActiveAdminNavItem(location.pathname);

  useEffect(() => {
    setNavOpen(false);
  }, [location.pathname]);

  if (!isLoggedIn) {
    return <Outlet />;
  }

  return (
    <div className="admin-shell">
      <button
        type="button"
        className={`admin-sidebar-backdrop ${navOpen ? 'is-visible' : ''}`}
        aria-label="Close navigation"
        onClick={() => setNavOpen(false)}
      />
      <aside className={`admin-sidebar ${navOpen ? 'is-open' : ''}`}>
        <Navigation pathname={location.pathname} onNavigate={() => setNavOpen(false)} />
      </aside>
      <div className="admin-main">
        <header className="admin-topbar">
          <button
            type="button"
            className="admin-topbar__menu"
            onClick={() => setNavOpen((open) => !open)}
          >
            {navOpen ? 'Close' : 'Browse'}
          </button>
          <div className="admin-topbar__title">
            <div className="admin-topbar__eyebrow">
              {activeNavItem?.group.label ?? 'Overview'}
            </div>
            <h1>{location.pathname === '/' ? 'Admin Home' : activeNavItem?.label ?? 'Admin Dashboard'}</h1>
            <p>
              {location.pathname === '/'
                ? 'Use the new grouped navigation or start from one of the workflow launchpads below.'
                : activeNavItem?.description ??
                  'Manage world content, quests, systems, and live operations.'}
            </p>
          </div>
          <div className="admin-topbar__quicklinks">
            {featuredAdminNavItems.slice(0, 5).map((item) => (
              <Link
                key={item.id}
                to={item.path}
                className={`admin-topbar__quicklink ${
                  adminNavItemMatchesPath(item, location.pathname) ? 'is-active' : ''
                }`}
              >
                {item.label}
              </Link>
            ))}
          </div>
        </header>
        <div className="admin-page">
          <Outlet />
        </div>
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
        path: '/',
        element: <AdminHome />,
        loader: onlyAuthenticated,
      },
      {
        path: '/arenas',
        element: <Arenas />,
        loader: onlyAuthenticated,
      },
      {
        path: '/login',
        element: <Login />,
        loader: onlyUnauthenticated,
      },
      {
        path: '/arena/:id',
        element: (
          <ArenaWrapper>
            <Arena />
          </ArenaWrapper>
        ),
        loader: onlyAuthenticated,
      },
      {
        path: '/armory',
        element: <Armory />,
        loader: onlyAuthenticated,
      },
      {
        path: '/zones',
        element: <Zones />,
        loader: onlyAuthenticated,
      },
      {
        path: '/zones/:id',
        element: <Zone />,
        loader: onlyAuthenticated,
      },
      {
        path: '/districts',
        element: <Districts />,
        loader: onlyAuthenticated,
      },
      {
        path: '/districts/:id',
        element: <DistrictEditor />,
        loader: onlyAuthenticated,
      },
      {
        path: '/place/:id',
        element: <Place />,
        loader: onlyAuthenticated,
      },
      {
        path: '/tags',
        element: <Tags />,
        loader: onlyAuthenticated,
      },
      {
        path: '/location-archetypes',
        element: <LocationArchetypes />,
        loader: onlyAuthenticated,
      },
      {
        path: '/quest-archetypes',
        element: <QuestArchetypeComponent />,
        loader: onlyAuthenticated,
      },
      {
        path: '/quest-archetype-generator',
        element: <QuestArchetypeGenerator />,
        loader: onlyAuthenticated,
      },
      {
        path: '/zone-quest-archetypes',
        element: <ZoneQuestArchetypes />,
        loader: onlyAuthenticated,
      },
      {
        path: '/users',
        element: <Users />,
        loader: onlyAuthenticated,
      },
      {
        path: '/parties',
        element: <Parties />,
        loader: onlyAuthenticated,
      },
      {
        path: '/characters',
        element: <Characters />,
        loader: onlyAuthenticated,
      },
      {
        path: '/inventory-items',
        element: <InventoryItems />,
        loader: onlyAuthenticated,
      },
      {
        path: '/bases',
        element: <Bases />,
        loader: onlyAuthenticated,
      },
      {
        path: '/starter-config',
        element: <NewUserStarterConfig />,
        loader: onlyAuthenticated,
      },
      {
        path: '/tutorial',
        element: <Tutorial />,
        loader: onlyAuthenticated,
      },
      {
        path: '/treasure-chests',
        element: <TreasureChests />,
        loader: onlyAuthenticated,
      },
      {
        path: '/healing-fountains',
        element: <HealingFountains />,
        loader: onlyAuthenticated,
      },
      {
        path: '/fete-rooms',
        element: <FeteRooms />,
        loader: onlyAuthenticated,
      },
      {
        path: '/fete-teams',
        element: <FeteTeams />,
        loader: onlyAuthenticated,
      },
      {
        path: '/fete-room-linked-list-teams',
        element: <FeteRoomLinkedListTeams />,
        loader: onlyAuthenticated,
      },
      {
        path: '/fete-room-teams',
        element: <FeteRoomTeams />,
        loader: onlyAuthenticated,
      },
      {
        path: '/utility-closet-puzzle',
        element: <UtilityClosetPuzzleAdmin />,
        loader: onlyAuthenticated,
      },
      {
        path: '/flagged-photos',
        element: <FlaggedPhotos />,
        loader: onlyAuthenticated,
      },
      {
        path: '/points-of-interest',
        element: <PointOfInterest />,
        loader: onlyAuthenticated,
      },
      {
        path: '/points-of-interest/:id',
        element: <PointOfInterestEditor />,
        loader: onlyAuthenticated,
      },
      {
        path: '/quests',
        element: <Quests />,
        loader: onlyAuthenticated,
      },
      {
        path: '/insider-trades',
        element: <InsiderTrades />,
        loader: onlyAuthenticated,
      },
      {
        path: '/feedback',
        element: <Feedback />,
        loader: onlyAuthenticated,
      },
      {
        path: '/zone-seeding',
        element: <ZoneSeedJobs />,
        loader: onlyAuthenticated,
      },
      {
        path: '/zone-tagging',
        element: <ZoneTagJobs />,
        loader: onlyAuthenticated,
      },
      {
        path: '/scenarios',
        element: <Scenarios />,
        loader: onlyAuthenticated,
      },
      {
        path: '/scenario-templates',
        element: <ScenarioTemplates />,
        loader: onlyAuthenticated,
      },
      {
        path: '/challenges',
        element: <Challenges />,
        loader: onlyAuthenticated,
      },
      {
        path: '/challenge-templates',
        element: <ChallengeTemplates />,
        loader: onlyAuthenticated,
      },
      {
        path: '/spells',
        element: <Spells />,
        loader: onlyAuthenticated,
      },
      {
        path: '/monsters',
        element: <Monsters />,
        loader: onlyAuthenticated,
      },
    ],
  },
]);

const App = () => {
  return (
    <LocationProvider>
      <APIProvider>
        <AuthProvider appName="UCS Admin Dashboard" uriPrefix="/sonar">
          <MediaContextProvider>
            <TagProvider>
              <ZoneProvider>
                <MapProvider>
                  <QuestArchetypesProvider>
                    <InventoryProvider>
                      <RouterProvider router={router} />
                    </InventoryProvider>
                  </QuestArchetypesProvider>
                </MapProvider>
              </ZoneProvider>
            </TagProvider>
          </MediaContextProvider>
        </AuthProvider>
      </APIProvider>
    </LocationProvider>
  );
};

export default App;
