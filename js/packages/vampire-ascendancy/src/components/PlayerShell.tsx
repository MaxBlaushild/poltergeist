import { useCallback, useEffect, useRef, useState } from 'react';
import { Navigate, useNavigate } from 'react-router-dom';
import { getMe, getToken, clearToken, ApiError } from '../api';
import type { MeResponse } from '../types';
import { Summons } from './Summons';
import { Tournament } from './Tournament';
import { Dossier } from './Dossier';
import { Missions } from './Missions';
import { Inventory } from './Inventory';
import { PhysicalGames } from './PhysicalGames';
import { Leaderboard } from './Leaderboard';
import { VampireMark } from './VampireMark';
import { NotificationTakeover } from './NotificationTakeover';
import { QuizTakeover } from './QuizTakeover';
import { FinalReveal } from './FinalReveal';

const DISMISSED_KEY = 'vampireDismissedNotif';
const REVEAL_KEY = 'vampireRevealDismissed';
const TAB_KEY = 'vampireTab';

type Tab = 'tournament' | 'dossier' | 'missions' | 'inventory' | 'games' | 'standings';
// A view is a strip tab or the Summons, which lives in the hamburger menu.
type View = Tab | 'summons';

// The main tabs, always visible in a scrollable strip. Labels are short; the
// content, not the label, carries the description. The Summons is intentionally
// not here — it's the landing screen and is reached again via the menu.
const TABS: { id: Tab; label: string }[] = [
  { id: 'tournament', label: 'Rules' },
  { id: 'dossier', label: 'Dossier' },
  { id: 'missions', label: 'Missions' },
  { id: 'inventory', label: 'Items' },
  { id: 'games', label: 'Games' },
  { id: 'standings', label: 'Standings' },
];
const KNOWN_VIEWS: View[] = ['summons', ...TABS.map((t) => t.id)];

// PlayerShell owns the token, the /me poll, and the top navigation. The screens
// are presentational and read from the shared state.
export const PlayerShell = () => {
  const token = getToken();
  const [me, setMe] = useState<MeResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [view, setView] = useState<View>(() => (localStorage.getItem(TAB_KEY) as View) || 'summons');
  const [dismissedNotif, setDismissedNotif] = useState(() => localStorage.getItem(DISMISSED_KEY));
  const [quizDismissedPart, setQuizDismissedPart] = useState<number | null>(null);
  const [revealDismissed, setRevealDismissed] = useState(() => localStorage.getItem(REVEAL_KEY));
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

  const activeView: View = KNOWN_VIEWS.includes(view) ? view : 'summons';

  const selectView = (v: View) => {
    localStorage.setItem(TAB_KEY, v);
    setView(v);
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

  // The Final Reveal supersedes everything once the GM triggers it.
  const showReveal = me.gameState.currentAct === 'resolved' && !revealDismissed;
  const dismissReveal = () => {
    localStorage.setItem(REVEAL_KEY, '1');
    setRevealDismissed('1');
  };

  return (
    <div className="min-h-screen max-w-2xl mx-auto px-4 pb-10">
      {showReveal && <FinalReveal onDone={dismissReveal} />}
      {!showReveal && showQuiz && activeQuizPart && (
        <QuizTakeover
          part={activeQuizPart as 1 | 2}
          onDone={() => setQuizDismissedPart(activeQuizPart)}
        />
      )}
      {!showReveal && !showQuiz && showTakeover && me.notification && (
        <NotificationTakeover notification={me.notification} onDismiss={dismissTakeover} />
      )}
      <TopNav active={activeView} onSelect={selectView} onLogout={logout} />
      {activeView === 'summons' && <Summons />}
      {activeView === 'tournament' && <Tournament />}
      {activeView === 'dossier' && <Dossier me={me} />}
      {activeView === 'missions' && <Missions me={me} reload={reload} />}
      {activeView === 'inventory' && <Inventory />}
      {activeView === 'games' && <PhysicalGames />}
      {activeView === 'standings' && <Leaderboard myHouse={me.character?.house?.name} />}
    </div>
  );
};

// Horizontally-scrollable tab strip. The hamburger holds the Summons (the landing
// screen, reached again from here) and Sign out.
const TopNav = ({
  active,
  onSelect,
  onLogout,
}: {
  active: View;
  onSelect: (v: View) => void;
  onLogout: () => void;
}) => {
  const [menuOpen, setMenuOpen] = useState(false);
  const activeRef = useRef<HTMLButtonElement>(null);

  // Keep the active tab in view when it changes (e.g. off-screen on a phone).
  useEffect(() => {
    activeRef.current?.scrollIntoView({ inline: 'center', block: 'nearest' });
  }, [active]);

  return (
    <nav className="sticky top-0 z-20 -mx-4 px-2 py-2 mb-4 bg-blood-ink/95 backdrop-blur-sm border-b border-blood/30">
      <div className="flex items-center gap-1">
        <div className="flex-1 min-w-0 overflow-x-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          <div className="flex gap-1 w-max">
            {TABS.map((t) => (
              <button
                key={t.id}
                ref={active === t.id ? activeRef : undefined}
                onClick={() => onSelect(t.id)}
                className={`whitespace-nowrap px-3 py-2 rounded-md uppercase tracking-[0.1em] text-xs sm:text-sm transition-colors ${
                  active === t.id ? 'bg-blood text-bone' : 'text-bone/70 hover:text-bone'
                }`}
              >
                {t.label}
              </button>
            ))}
          </div>
        </div>
        <div className="relative shrink-0">
          <button
            onClick={() => setMenuOpen((o) => !o)}
            aria-label="Menu"
            aria-expanded={menuOpen}
            className="p-2 rounded-md text-bone/70 hover:text-bone"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <line x1="3" y1="6" x2="21" y2="6" />
              <line x1="3" y1="12" x2="21" y2="12" />
              <line x1="3" y1="18" x2="21" y2="18" />
            </svg>
          </button>
          {menuOpen && (
            <>
              <div className="fixed inset-0 z-10" onClick={() => setMenuOpen(false)} />
              <div className="absolute right-0 mt-2 z-20 min-w-[12rem] rounded-lg border border-blood/40 bg-blood-ink shadow-xl overflow-hidden">
                <button
                  onClick={() => {
                    setMenuOpen(false);
                    onSelect('summons');
                  }}
                  className={`w-full text-left px-4 py-3 uppercase tracking-[0.12em] text-sm ${
                    active === 'summons' ? 'bg-blood text-bone' : 'text-bone/80 hover:bg-white/5 hover:text-bone'
                  }`}
                >
                  The Summons
                </button>
                <button
                  onClick={() => {
                    setMenuOpen(false);
                    onLogout();
                  }}
                  className="w-full text-left px-4 py-3 uppercase tracking-[0.12em] text-sm text-blood-bright border-t border-blood/30 hover:bg-white/5"
                >
                  Sign out
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </nav>
  );
};

const Centered = ({ children }: { children: React.ReactNode }) => (
  <div className="min-h-screen flex items-center justify-center px-6 text-center">
    <div className="max-w-md text-bone/80">{children}</div>
  </div>
);
