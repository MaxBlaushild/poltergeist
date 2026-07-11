import { useEffect, useRef, useState } from 'react';
import { getGMAuth, gmListGames } from '../../gmApi';
import type { GMGame } from '../../gmApi';
import { formatClock } from '../../theme';

const LEAD_MINUTES = 7; // warn this many minutes before a game starts
const FIRED_KEY = 'vampireGMFired';
const nowMinutes = () => {
  const d = new Date();
  return d.getHours() * 60 + d.getMinutes();
};

// A live reminder for the logged-in GM: when a game they're assigned to is ~7
// minutes out (through the end of its slot), it shows a banner and — if the GM
// granted permission — fires a browser notification even if the tab is backgrounded.
export const GMReminder = () => {
  const { name } = getGMAuth();
  const [games, setGames] = useState<GMGame[]>([]);
  const [, tick] = useState(0);
  const firedRef = useRef<Set<string>>(
    new Set((() => {
      try {
        return JSON.parse(localStorage.getItem(FIRED_KEY) || '[]');
      } catch {
        return [];
      }
    })())
  );

  useEffect(() => {
    if ('Notification' in window && Notification.permission === 'default') {
      Notification.requestPermission().catch(() => {});
    }
    let cancelled = false;
    const load = () => gmListGames().then((d) => !cancelled && setGames(d.games)).catch(() => {});
    load();
    const poll = setInterval(load, 20000);
    const clock = setInterval(() => tick((t) => t + 1), 30000); // refresh the countdown
    return () => {
      cancelled = true;
      clearInterval(poll);
      clearInterval(clock);
    };
  }, []);

  const now = nowMinutes();
  const mine = games
    .filter(
      (g) =>
        g.assignedGm &&
        g.assignedGm === name &&
        g.status !== 'played' &&
        g.startMinutes != null &&
        g.endMinutes != null &&
        now >= g.startMinutes - LEAD_MINUTES &&
        now < g.endMinutes
    )
    .sort((a, b) => a.startMinutes! - b.startMinutes!);

  // Fire a one-time browser notification as each reminder enters its window.
  const mineKey = mine.map((g) => g.id).join(',');
  useEffect(() => {
    if (!('Notification' in window) || Notification.permission !== 'granted') return;
    for (const g of mine) {
      if (firedRef.current.has(g.id)) continue;
      firedRef.current.add(g.id);
      localStorage.setItem(FIRED_KEY, JSON.stringify([...firedRef.current]));
      try {
        new Notification(`You run "${g.name}" soon`, {
          body: `${formatClock(g.startMinutes!)}${g.location ? ' · ' + g.location : ''}`,
        });
      } catch {
        /* ignore */
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [mineKey]);

  if (mine.length === 0) return null;

  return (
    <div className="mb-5 rounded-lg border-2 border-gold/70 bg-gold/10 p-3">
      <p className="font-heading text-gold text-xs uppercase tracking-[0.3em] mb-1">Your game</p>
      {mine.map((g) => {
        const mins = g.startMinutes! - now;
        return (
          <p key={g.id} className="text-bone">
            ⏰ <span className="font-semibold">{g.name}</span>{' '}
            <span className="text-gold">
              {mins > 0 ? `starts in ${mins} min` : 'happening now'}
            </span>{' '}
            <span className="text-bone/60">
              · {formatClock(g.startMinutes!)}
              {g.location && ` · 📍 ${g.location}`}
            </span>
          </p>
        );
      })}
    </div>
  );
};
