import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { setGMInstanceId, gmGetState } from '../../gmApi';
import { ApiError } from '../../api';
import { useUserAuth } from '../../userAuth';
import { SignInForm } from '../SignInForm';
import type { GameState } from '../../types';
import { GameSection } from './GameSection';
import { SubmissionsSection } from './SubmissionsSection';
import { AwardsSection } from './AwardsSection';
import { PlayersSection } from './PlayersSection';
import { InvitesSection } from './InvitesSection';
import { BroadcastSection } from './BroadcastSection';
import { QuizSection } from './QuizSection';
import { GamesSection } from './GamesSection';
import { StandingsSection } from './StandingsSection';
import { BroadcastScreen } from './BroadcastScreen';
import { ItemsSection } from './ItemsSection';
import { ContentSection } from './ContentSection';
import { CoHostsSection } from './CoHostsSection';
import { GMReminder } from './GMReminder';

type Tab =
  | 'controls'
  | 'submissions'
  | 'awards'
  | 'games'
  | 'items'
  | 'content'
  | 'cohosts'
  | 'standings'
  | 'notifications'
  | 'broadcast'
  | 'invites'
  | 'players';

// Whether the signed-in account can administer this one Toast — checked by
// consequence (gmGetState either 200s or the server's withInstanceAdmin
// middleware rejects it), same pattern the old passcode gate used.
type AdminStatus = 'checking' | 'ok' | 'forbidden' | 'signed-out';

export const GMAdmin = () => {
  const { instanceId } = useParams();
  // Every gmApi call below targets this Toast — see setGMInstanceId.
  setGMInstanceId(instanceId || '');

  const { auth } = useUserAuth();
  const [status, setStatus] = useState<AdminStatus>('checking');

  useEffect(() => {
    if (!instanceId) return;
    if (!auth) {
      setStatus('signed-out');
      return;
    }
    setStatus('checking');
    gmGetState()
      .then(() => setStatus('ok'))
      .catch((err: unknown) => {
        setStatus(err instanceof ApiError && err.status === 401 ? 'signed-out' : 'forbidden');
      });
  }, [instanceId, auth]);

  if (!instanceId) return <Centered>No Toast selected.</Centered>;
  if (status === 'checking') return <Centered>Verifying the seal…</Centered>;
  if (status === 'signed-out') {
    return (
      <div className="min-h-screen flex items-center justify-center px-4">
        <SignInForm title="Sign in to administer this Toast" onSignedIn={() => setStatus('checking')} />
      </div>
    );
  }
  if (status === 'forbidden') {
    return (
      <Centered>
        <p className="text-bone/80">
          You are not a Host or Co-Host of this Toast. Ask whoever hosts it to invite you.
        </p>
      </Centered>
    );
  }
  return <GMConsole />;
};

const GMConsole = () => {
  const navigate = useNavigate();
  const { auth } = useUserAuth();
  const [tab, setTab] = useState<Tab>('controls');
  const [state, setState] = useState<GameState | null>(null);

  const refreshState = () => gmGetState().then(setState).catch(() => {});
  useEffect(() => {
    refreshState();
  }, []);

  const tabs: { id: Tab; label: string }[] = [
    { id: 'controls', label: 'Controls' },
    { id: 'submissions', label: 'Submissions' },
    { id: 'awards', label: 'HF Awards' },
    { id: 'games', label: 'Games' },
    { id: 'items', label: 'Items' },
    { id: 'content', label: 'Content' },
    { id: 'standings', label: 'Standings' },
    { id: 'notifications', label: 'Announcements' },
    { id: 'broadcast', label: 'Broadcast' },
    { id: 'invites', label: 'Invites' },
    { id: 'players', label: 'Players' },
    { id: 'cohosts', label: 'Co-Hosts' },
  ];

  return (
    <div className="min-h-screen max-w-3xl mx-auto px-4 py-6">
      <header className="flex items-center justify-between mb-5">
        <div>
          <p className="font-heading text-xs uppercase tracking-[0.35em] text-gold">Court Master</p>
          <p className="text-bone/80 text-sm">Acting as {auth?.user.name}</p>
        </div>
        <button
          onClick={() => navigate('/toasts')}
          className="text-bone/50 text-sm uppercase tracking-[0.2em]"
        >
          Leave
        </button>
      </header>

      <GMReminder />

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

      {tab === 'controls' && (
        <GameSection
          state={state}
          onChange={refreshState}
          quizSlot={
            state?.currentAct === 'quiz' ? (
              <QuizSection state={state} onChange={refreshState} />
            ) : undefined
          }
        />
      )}
      {tab === 'submissions' && <SubmissionsSection />}
      {tab === 'awards' && <AwardsSection />}
      {tab === 'games' && <GamesSection />}
      {tab === 'items' && <ItemsSection />}
      {tab === 'content' && <ContentSection />}
      {tab === 'standings' && <StandingsSection />}
      {tab === 'notifications' && <BroadcastSection />}
      {tab === 'broadcast' && <BroadcastScreen />}
      {tab === 'invites' && <InvitesSection />}
      {tab === 'players' && <PlayersSection />}
      {tab === 'cohosts' && <CoHostsSection />}
    </div>
  );
};

const Centered = ({ children }: { children: React.ReactNode }) => (
  <div className="min-h-screen flex items-center justify-center px-6 text-center text-bone/80">
    {children}
  </div>
);
