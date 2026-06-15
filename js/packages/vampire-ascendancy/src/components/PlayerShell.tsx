import { useCallback, useEffect, useState } from 'react';
import { Navigate } from 'react-router-dom';
import { getMe, getToken, clearToken, ApiError } from '../api';
import type { MeResponse } from '../types';
import { Dossier } from './Dossier';
import { Leaderboard } from './Leaderboard';
import { VampireMark } from './VampireMark';
import { NotificationTakeover } from './NotificationTakeover';
import { QuizTakeover } from './QuizTakeover';

const DISMISSED_KEY = 'vampireDismissedNotif';
const TAB_KEY = 'vampireTab';

type Tab = 'dossier' | 'secrets' | 'missions' | 'standings';

const TAB_LABEL: Record<Tab, string> = {
  dossier: 'Dossier',
  secrets: 'Secrets',
  missions: 'Missions',
  standings: 'Standings',
};

// PlayerShell owns the token, the /me poll, and the single top navigation. The
// screens are presentational and read from the shared state.
export const PlayerShell = () => {
  const token = getToken();
  const [me, setMe] = useState<MeResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  // Remember the active tab so returning from a house page (which is its own
  // route) lands the player back on the tab they left, not always the dossier.
  const [tab, setTab] = useState<Tab>(() => (localStorage.getItem(TAB_KEY) as Tab) || 'dossier');
  const [dismissedNotif, setDismissedNotif] = useState(() => localStorage.getItem(DISMISSED_KEY));
  const [quizDismissed, setQuizDismissed] = useState(false);

  useEffect(() => {
    if (!token) {
      setLoading(false);
      return;
    }
    let cancelled = false;

    // Poll so unlocks, act changes, and verifications propagate without a manual
    // refresh. A failed poll keeps the last good state (flaky venue wifi).
    const load = (initial: boolean) => {
      getMe(token)
        .then((data) => {
          if (cancelled) return;
          setMe(data);
          setError(null);
          if (initial) setLoading(false);
        })
        .catch((err: unknown) => {
          if (cancelled) return;
          if (err instanceof ApiError && err.status === 401) {
            clearToken(); // stale/invalid token — back to login
            setError('invalid-token');
            setLoading(false);
          } else if (initial) {
            setError('load-failed');
            setLoading(false);
          }
        });
    };
    load(true);
    const id = setInterval(() => load(false), 5000);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, [token]);

  const reload = useCallback(() => {
    if (token) getMe(token).then(setMe).catch(() => {});
  }, [token]);

  // No token (or it just went stale) → send them to the login page.
  if (!token || error === 'invalid-token') return <Navigate to="/login" replace />;

  if (loading) return <Centered>Summoning your dossier…</Centered>;
  if (error || !me) {
    return (
      <Centered>
        <VampireMark className="w-16 h-16 mx-auto mb-4" />
        <p className="text-bone/85">The court could not be reached. Try again in a moment.</p>
      </Centered>
    );
  }

  // Standings is always available; the rest only once content is unlocked.
  const unlocked = me.gameState.contentUnlocked;
  const tabs: Tab[] = unlocked
    ? ['dossier', 'secrets', 'missions', 'standings']
    : ['dossier', 'standings'];
  const activeTab = tabs.includes(tab) ? tab : 'dossier';

  const selectTab = (t: Tab) => {
    localStorage.setItem(TAB_KEY, t);
    setTab(t);
  };

  // Show a GM broadcast as a takeover until this player dismisses it.
  const showTakeover = me.notification && me.notification.id !== dismissedNotif;
  const dismissTakeover = () => {
    if (!me.notification) return;
    localStorage.setItem(DISMISSED_KEY, me.notification.id);
    setDismissedNotif(me.notification.id);
  };

  // The end quiz takes over the screen while it's open and not yet dismissed.
  const showQuiz = me.gameState.quizOpen && !quizDismissed;

  return (
    <div className="min-h-screen max-w-2xl mx-auto px-4 pb-10">
      {showQuiz && <QuizTakeover onDone={() => setQuizDismissed(true)} />}
      {!showQuiz && showTakeover && me.notification && (
        <NotificationTakeover notification={me.notification} onDismiss={dismissTakeover} />
      )}
      <TopNav tabs={tabs} active={activeTab} onSelect={selectTab} />
      {activeTab === 'standings' ? (
        <Leaderboard myHouse={me.character?.house?.name} />
      ) : (
        <Dossier me={me} reload={reload} section={activeTab} />
      )}
    </div>
  );
};

const TopNav = ({
  tabs,
  active,
  onSelect,
}: {
  tabs: Tab[];
  active: Tab;
  onSelect: (t: Tab) => void;
}) => (
  <nav className="sticky top-0 z-10 -mx-4 px-4 py-2 mb-4 bg-blood-ink/90 backdrop-blur-sm border-b border-blood/30">
    <div className="flex gap-1 justify-center">
      {tabs.map((t) => (
        <button
          key={t}
          onClick={() => onSelect(t)}
          className={`px-3 py-2 rounded-md text-xs sm:text-sm uppercase tracking-[0.15em] transition-colors ${
            active === t ? 'bg-blood text-bone' : 'text-bone/80 hover:text-bone'
          }`}
        >
          {TAB_LABEL[t]}
        </button>
      ))}
    </div>
  </nav>
);

const Centered = ({ children }: { children: React.ReactNode }) => (
  <div className="min-h-screen flex items-center justify-center px-6 text-center">
    <div className="max-w-md text-bone/80">{children}</div>
  </div>
);
