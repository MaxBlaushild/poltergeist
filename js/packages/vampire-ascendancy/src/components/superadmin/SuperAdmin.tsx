import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { adminListHouses } from '../../superAdminApi';
import { ApiError } from '../../api';
import { useUserAuth } from '../../userAuth';
import { SignInForm } from '../SignInForm';
import { SuperAdminCharacters } from './SuperAdminCharacters';
import { SuperAdminHouses } from './SuperAdminHouses';
import { SuperAdminItems } from './SuperAdminItems';
import { SuperAdminMysteries } from './SuperAdminMysteries';
import { SuperAdminUsers } from './SuperAdminUsers';

type Tab = 'mysteries' | 'characters' | 'houses' | 'items' | 'users';
type Status = 'checking' | 'ok' | 'forbidden' | 'signed-out';

// The shared content library editor — characters, houses, items, and quiz
// questions are global (read by every Toast), so editing them is restricted
// to super users here, not any Toast's Host/Co-Host GM console. Not scoped
// to any instance at all.
export const SuperAdmin = () => {
  const { auth } = useUserAuth();
  const [status, setStatus] = useState<Status>('checking');

  useEffect(() => {
    if (!auth) {
      setStatus('signed-out');
      return;
    }
    setStatus('checking');
    adminListHouses()
      .then(() => setStatus('ok'))
      .catch((err: unknown) => {
        setStatus(err instanceof ApiError && err.status === 401 ? 'signed-out' : 'forbidden');
      });
  }, [auth]);

  if (status === 'checking') return <Centered>Verifying access…</Centered>;
  if (status === 'signed-out') {
    return (
      <div className="min-h-screen flex items-center justify-center px-4">
        <SignInForm title="Sign in to the Super Admin dashboard" onSignedIn={() => setStatus('checking')} />
      </div>
    );
  }
  if (status === 'forbidden') {
    return (
      <Centered>
        <p className="text-bone/80">
          You don't have access to the shared content library. Ask an existing super user to grant it.
        </p>
      </Centered>
    );
  }
  return <SuperAdminConsole />;
};

const SuperAdminConsole = () => {
  const navigate = useNavigate();
  const { auth } = useUserAuth();
  const [tab, setTab] = useState<Tab>('mysteries');

  const tabs: { id: Tab; label: string }[] = [
    { id: 'mysteries', label: 'Mysteries' },
    { id: 'characters', label: 'Characters' },
    { id: 'houses', label: 'Houses' },
    { id: 'items', label: 'Items' },
    { id: 'users', label: 'Super Users' },
  ];

  return (
    <div className="min-h-screen max-w-3xl mx-auto px-4 py-6">
      <header className="flex items-center justify-between mb-5">
        <div>
          <p className="font-heading text-xs uppercase tracking-[0.35em] text-gold">Super Admin</p>
          <p className="text-bone/80 text-sm">Shared content library — {auth?.user.name}</p>
        </div>
        <button onClick={() => navigate('/toasts')} className="text-bone/50 text-sm uppercase tracking-[0.2em]">
          Leave
        </button>
      </header>

      <p className="mb-5 rounded-md border border-gold/50 bg-gold/10 p-3 text-sm text-bone/90">
        ⚠️ Everything here is shared across every Toast built on this story. Per-instance concerns —
        which characters/items/quiz questions a Toast includes, sigils, portraits — live in each
        Toast's own GM console instead.
      </p>

      <nav className="flex gap-2 mb-6 flex-wrap">
        {tabs.map((t) => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`px-4 py-2 rounded-md text-sm uppercase tracking-[0.15em] ${
              tab === t.id ? 'bg-blood text-bone' : 'text-bone/60 border border-blood/40'
            }`}
          >
            {t.label}
          </button>
        ))}
      </nav>

      {tab === 'mysteries' && <SuperAdminMysteries />}
      {tab === 'characters' && <SuperAdminCharacters />}
      {tab === 'houses' && <SuperAdminHouses />}
      {tab === 'items' && <SuperAdminItems />}
      {tab === 'users' && <SuperAdminUsers />}
    </div>
  );
};

const Centered = ({ children }: { children: React.ReactNode }) => (
  <div className="min-h-screen flex items-center justify-center px-6 text-center text-bone/80">
    {children}
  </div>
);
