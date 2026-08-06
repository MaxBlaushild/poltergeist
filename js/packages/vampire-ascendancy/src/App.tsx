import { createBrowserRouter, RouterProvider, Navigate } from 'react-router-dom';
import { PlayerShell } from './components/PlayerShell';
import { ConfirmLogin, SelectLogin } from './components/Login';
import { HousePage } from './components/HousePage';
import { GMAdmin } from './components/gm/GMAdmin';
import { BroadcastPage } from './components/BroadcastPage';
import { Landing } from './components/Landing';
import { SignInPage } from './components/SignInPage';
import { MyToasts, CreateToast } from './components/MyToasts';
import { AcceptInvite } from './components/AcceptInvite';
import { RequireUser } from './components/RequireUser';
import { SuperAdmin } from './components/superadmin/SuperAdmin';

const router = createBrowserRouter([
  // Public front door.
  { path: '/', element: <Landing /> },
  { path: '/signin', element: <SignInPage /> },

  // Platform pages — Hosting/administering Toasts, not scoped to any one yet.
  { path: '/toasts', element: <RequireUser><MyToasts /></RequireUser> },
  { path: '/toasts/new', element: <RequireUser><CreateToast /></RequireUser> },
  { path: '/accept-invite/:token', element: <RequireUser><AcceptInvite /></RequireUser> },
  // Shared content library editor — super users only, not scoped to any Toast.
  { path: '/admin', element: <SuperAdmin /> },

  // Everything below is scoped to one Toast ("instance") via :instanceId.
  // GM admin (real Host/Co-Host account) lives in the same app.
  { path: '/e/:instanceId/gm', element: <GMAdmin /> },
  // Public projector screen — no auth, for casting to a TV.
  { path: '/e/:instanceId/broadcast', element: <BroadcastPage /> },
  // A guest's QR/link pre-selects their character; they confirm with a sigil.
  { path: '/e/:instanceId/c/:characterId', element: <ConfirmLogin /> },
  // The general "select your name" login, for anyone who lost their link.
  { path: '/e/:instanceId/login', element: <SelectLogin /> },
  // House overview (members + favor ledger).
  { path: '/e/:instanceId/house/:houseId', element: <HousePage /> },
  // The authenticated player app.
  { path: '/e/:instanceId', element: <PlayerShell /> },

  { path: '*', element: <Navigate to="/" replace /> },
]);

function App() {
  return (
    <div className="relative min-h-screen">
      <RouterProvider router={router} />
    </div>
  );
}

export default App;
