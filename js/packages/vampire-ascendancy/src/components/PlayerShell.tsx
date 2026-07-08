import { useCallback, useEffect, useState } from 'react';
import { Navigate, useNavigate } from 'react-router-dom';
import { getMe, getToken, clearToken, ApiError } from '../api';
import type { MeResponse } from '../types';
import { Dossier, type DossierSection } from './Dossier';
import { HowToPlay, EarnSpend } from './Rules';
import { Leaderboard } from './Leaderboard';
import { VampireMark } from './VampireMark';
import { NotificationTakeover } from './NotificationTakeover';
import { QuizTakeover } from './QuizTakeover';

const DISMISSED_KEY = 'vampireDismissedNotif';
const TAB_KEY = 'vampireTab';

type Tab = 'dossier' | 'chronicle' | 'missions' | 'secrets' | 'standings' | 'rules' | 'earn';

const TAB_LABEL: Record<Tab, string> = {
  dossier: 'Briefing',
  chronicle: 'The Night',
  missions: 'Missions',
  secrets: 'Secrets',
  standings: 'Standings',
  rules: 'How to Play',
  earn: 'Earn & Spend',
};

// Tabs that need no unlocked content — available the whole evening.
const ALWAYS_TABS: Tab[] = ['dossier', 'standings', 'rules', 'earn'];
// Personal story + tasks, revealed once the host opens the evening.
const UNLOCKED_TABS: Tab[] = ['chronicle', 'missions', 'secrets'];

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
  const [quizDismissedPart, setQuizDismissedPart] = useState<number | null>(null);
  const navigate = useNavigate();

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

  // Briefing, standings, and the rules are always available; personal story and
  // tasks only once content is unlocked. Keep a stable display order.
  const unlocked = me.gameState.contentUnlocked;
  const ORDER: Tab[] = ['dossier', 'chronicle', 'missions', 'secrets', 'standings', 'rules', 'earn'];
  const available = new Set<Tab>([...ALWAYS_TABS, ...(unlocked ? UNLOCKED_TABS : [])]);
  const tabs = ORDER.filter((t) => available.has(t));
  const activeTab = tabs.includes(tab) ? tab : 'dossier';

  const selectTab = (t: Tab) => {
    localStorage.setItem(TAB_KEY, t);
    setTab(t);
  };

  const logout = () => {
    clearToken();
    navigate('/login', { replace: true });
  };

  // Show a GM broadcast as a takeover until this player dismisses it.
  const showTakeover = me.notification && me.notification.id !== dismissedNotif;
  const dismissTakeover = () => {
    if (!me.notification) return;
    localStorage.setItem(DISMISSED_KEY, me.notification.id);
    setDismissedNotif(me.notification.id);
  };

  // The end quiz takes over the screen while a part is open and not yet dismissed.
  const activeQuizPart = me.gameState.quizPart1Open ? 1 : me.gameState.quizPart2Open ? 2 : null;
  const showQuiz = activeQuizPart !== null && quizDismissedPart !== activeQuizPart;

  return (
    <div className="min-h-screen max-w-2xl mx-auto px-4 pb-10">
      {showQuiz && activeQuizPart && (
        <QuizTakeover
          part={activeQuizPart as 1 | 2}
          onDone={() => setQuizDismissedPart(activeQuizPart)}
        />
      )}
      {!showQuiz && showTakeover && me.notification && (
        <NotificationTakeover notification={me.notification} onDismiss={dismissTakeover} />
      )}
      <TopNav tabs={tabs} active={activeTab} onSelect={selectTab} onLogout={logout} />
      {activeTab === 'standings' ? (
        <Leaderboard myHouse={me.character?.house?.name} />
      ) : activeTab === 'rules' ? (
        <HowToPlay />
      ) : activeTab === 'earn' ? (
        <EarnSpend />
      ) : (
        <Dossier me={me} reload={reload} section={activeTab as DossierSection} />
      )}
    </div>
  );
};

// A single menu button opens the full tab list. With up to seven tabs plus
// logout, an overflow menu stays legible on a phone where a flat bar would clip.
const TopNav = ({
  tabs,
  active,
  onSelect,
  onLogout,
}: {
  tabs: Tab[];
  active: Tab;
  onSelect: (t: Tab) => void;
  onLogout: () => void;
}) => {
  const [open, setOpen] = useState(false);

  const pick = (t: Tab) => {
    onSelect(t);
    setOpen(false);
  };

  return (
    <nav className="sticky top-0 z-20 -mx-4 px-3 py-2 mb-4 bg-blood-ink/95 backdrop-blur-sm border-b border-blood/30">
      <div className="flex items-center gap-3">
        <button
          onClick={() => setOpen((o) => !o)}
          aria-label="Menu"
          aria-expanded={open}
          className="p-2 -ml-1 rounded-md text-bone/80 hover:text-bone"
        >
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
            {open ? (
              <>
                <line x1="6" y1="6" x2="18" y2="18" />
                <line x1="6" y1="18" x2="18" y2="6" />
              </>
            ) : (
              <>
                <line x1="3" y1="6" x2="21" y2="6" />
                <line x1="3" y1="12" x2="21" y2="12" />
                <line x1="3" y1="18" x2="21" y2="18" />
              </>
            )}
          </svg>
        </button>
        <span className="font-heading uppercase tracking-[0.2em] text-sm text-bone">
          {TAB_LABEL[active]}
        </span>
      </div>

      {open && (
        <>
          {/* Tap-away scrim so the menu closes when the player taps the page. */}
          <div className="fixed inset-0 z-10" onClick={() => setOpen(false)} />
          <div className="absolute left-2 right-2 mt-2 z-20 rounded-lg border border-blood/40 bg-blood-ink shadow-xl overflow-hidden">
            {tabs.map((t) => (
              <button
                key={t}
                onClick={() => pick(t)}
                className={`w-full text-left px-4 py-3 uppercase tracking-[0.12em] text-sm transition-colors ${
                  active === t ? 'bg-blood text-bone' : 'text-bone/80 hover:bg-white/5 hover:text-bone'
                }`}
              >
                {TAB_LABEL[t]}
              </button>
            ))}
            <button
              onClick={() => {
                setOpen(false);
                onLogout();
              }}
              className="w-full text-left px-4 py-3 uppercase tracking-[0.12em] text-sm text-blood-bright border-t border-blood/30 hover:bg-white/5"
            >
              Log out / change character
            </button>
          </div>
        </>
      )}
    </nav>
  );
};

const Centered = ({ children }: { children: React.ReactNode }) => (
  <div className="min-h-screen flex items-center justify-center px-6 text-center">
    <div className="max-w-md text-bone/80">{children}</div>
  </div>
);
