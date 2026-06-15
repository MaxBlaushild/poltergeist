import { createBrowserRouter, RouterProvider, Navigate } from 'react-router-dom';
import { PlayerShell } from './components/PlayerShell';
import { ConfirmLogin, SelectLogin } from './components/Login';
import { HousePage } from './components/HousePage';
import { GMAdmin } from './components/gm/GMAdmin';

const router = createBrowserRouter([
  // GM admin (passcode-gated) lives in the same app.
  { path: '/gm', element: <GMAdmin /> },
  // A guest's QR/link pre-selects their character; they confirm with a sigil.
  { path: '/c/:characterId', element: <ConfirmLogin /> },
  // The general "select your name" login, for anyone who lost their link.
  { path: '/login', element: <SelectLogin /> },
  // House overview (members + favor ledger).
  { path: '/house/:houseId', element: <HousePage /> },
  // The authenticated player app.
  { path: '/', element: <PlayerShell /> },
  { path: '*', element: <Navigate to="/login" replace /> },
]);

function App() {
  return (
    <div className="relative min-h-screen">
      <RouterProvider router={router} />
    </div>
  );
}

export default App;
