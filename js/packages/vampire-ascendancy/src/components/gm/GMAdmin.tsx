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
import { CharacterPoolSection } from './CharacterPoolSection';
import { BroadcastSection } from './BroadcastSection';
import { QuizSection } from './QuizSection';
import { GamesScheduleSection } from './GamesScheduleSection';
import { GamesScoringSection } from './GamesScoringSection';
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
  | 'gamesSetup'
  | 'gamesScoring'
  | 'items'
  | 'content'
  | 'cohosts'
  | 'standings'
  | 'notifications'
  | 'broadcast'
  | 'characterPool'
  | 'invites'
  | 'players';

// The console groups its twelve tabs into three sections, matching the
// actual arc of running one of these nights: get people in (Setup), pace
// the evening (Live), track the game (Scoring). Each group's first tab is
// its default when switching into it.
type Group = 'setup' | 'run' | 'scoring';
const GROUPS: { id: Group; label: string; tabs: { id: Tab; label: string }[] }[] = [
  {
    id: 'setup',
    label: 'Setup',
    tabs: [
      { id: 'content', label: 'Content' },
      { id: 'characterPool', label: 'Character Pool' },
      { id: 'invites', label: 'Invites' },
      { id: 'players', label: 'Players' },
      { id: 'gamesSetup', label: 'Games' },
      { id: 'cohosts', label: 'Co-Hosts' },
    ],
  },
  {
    id: 'run',
    label: 'Live',
    tabs: [
      { id: 'controls', label: 'Controls' },
      { id: 'notifications', label: 'Announcements' },
      { id: 'broadcast', label: 'Broadcast' },
    ],
  },
  {
    id: 'scoring',
    label: 'Scoring',
    tabs: [
      { id: 'submissions', label: 'Submissions' },
      { id: 'awards', label: 'HF Awards' },
      { id: 'gamesScoring', label: 'Games' },
      { id: 'items', label: 'Items' },
      { id: 'standings', label: 'Standings' },
    ],
  },
];

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
  // Default to Live → Controls — the single most-reached-for tab,
  // whether they're mid-event or just checking the current act.
  const [group, setGroup] = useState<Group>('run');
  const [tab, setTab] = useState<Tab>('controls');
  const [state, setState] = useState<GameState | null>(null);

  const refreshState = () => gmGetState().then(setState).catch(() => {});
  useEffect(() => {
    refreshState();
  }, []);

  const activeGroup = GROUPS.find((g) => g.id === group)!;
  const selectGroup = (g: Group) => {
    setGroup(g);
    setTab(GROUPS.find((x) => x.id === g)!.tabs[0].id);
  };

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

      <nav className="flex gap-2 mb-3">
        {GROUPS.map((g) => (
          <button
            key={g.id}
            onClick={() => selectGroup(g.id)}
            className={`flex-1 px-4 py-2.5 rounded-full text-sm uppercase tracking-[0.2em] font-semibold transition-colors ${
              group === g.id ? 'bg-gold text-blood-ink' : 'text-gold/70 border border-gold/30 hover:text-gold'
            }`}
          >
            {g.label}
          </button>
        ))}
      </nav>

      <nav className="flex gap-2 mb-6 flex-wrap">
        {activeGroup.tabs.map((t) => (
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
      {tab === 'gamesSetup' && <GamesScheduleSection />}
      {tab === 'gamesScoring' && <GamesScoringSection />}
      {tab === 'items' && <ItemsSection />}
      {tab === 'content' && <ContentSection />}
      {tab === 'standings' && <StandingsSection />}
      {tab === 'notifications' && <BroadcastSection />}
      {tab === 'broadcast' && <BroadcastScreen />}
      {tab === 'characterPool' && <CharacterPoolSection />}
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
