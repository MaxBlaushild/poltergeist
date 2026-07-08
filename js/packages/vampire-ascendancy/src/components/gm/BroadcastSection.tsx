import { useEffect, useState } from 'react';
import { gmPushNotification, gmClearNotifications, gmListHouses, gmListPlayers } from '../../gmApi';
import type { GMPlayer } from '../../gmApi';
import type { House } from '../../types';
import { Card } from './GameSection';

type Scope = 'all' | 'house' | 'player';

export const BroadcastSection = () => {
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [scope, setScope] = useState<Scope>('all');
  const [targetId, setTargetId] = useState('');
  const [houses, setHouses] = useState<House[]>([]);
  const [players, setPlayers] = useState<GMPlayer[]>([]);
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);

  useEffect(() => {
    gmListHouses().then((d) => setHouses(d.houses));
    gmListPlayers().then((d) => setPlayers(d.players.filter((p) => p.character)));
  }, []);

  const push = async () => {
    if (!body.trim() || (scope !== 'all' && !targetId)) return;
    setBusy(true);
    setMsg(null);
    try {
      await gmPushNotification(title.trim(), body.trim(), scope, targetId);
      setTitle('');
      setBody('');
      setMsg('Announcement sent.');
    } catch {
      setMsg('Failed to send.');
    } finally {
      setBusy(false);
    }
  };

  const clear = async () => {
    setBusy(true);
    setMsg(null);
    try {
      await gmClearNotifications();
      setMsg('Cleared.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="flex flex-col gap-6">
      <Card title="Announcement">
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Title (optional, e.g. Act 2 Begins)"
          className="w-full mb-2 rounded-md bg-black/60 border border-blood/40 p-3 text-bone"
        />
        <textarea
          value={body}
          onChange={(e) => setBody(e.target.value)}
          placeholder="Message — appears full-screen on phones"
          rows={3}
          className="w-full mb-3 rounded-md bg-black/60 border border-blood/40 p-3 text-bone"
        />

        <div className="flex gap-2 mb-3">
          {(['all', 'house', 'player'] as Scope[]).map((sc) => (
            <button
              key={sc}
              onClick={() => {
                setScope(sc);
                setTargetId('');
              }}
              className={`px-3 py-2 rounded-md text-xs uppercase tracking-[0.15em] ${
                scope === sc ? 'bg-blood text-bone' : 'border border-blood/40 text-bone/70'
              }`}
            >
              {sc === 'all' ? 'Everyone' : sc}
            </button>
          ))}
        </div>

        {scope === 'house' && (
          <select
            value={targetId}
            onChange={(e) => setTargetId(e.target.value)}
            className="w-full mb-3 rounded-md bg-black/60 border border-blood/40 p-3 text-bone"
          >
            <option value="">— Select house —</option>
            {houses.map((h) => (
              <option key={h.id} value={h.id}>
                House of {h.name}
              </option>
            ))}
          </select>
        )}
        {scope === 'player' && (
          <select
            value={targetId}
            onChange={(e) => setTargetId(e.target.value)}
            className="w-full mb-3 rounded-md bg-black/60 border border-blood/40 p-3 text-bone"
          >
            <option value="">— Select player —</option>
            {players.map((p) => (
              <option key={p.id} value={p.id}>
                {p.character?.name}
                {p.guestLabel ? ` (${p.guestLabel})` : ''}
              </option>
            ))}
          </select>
        )}

        <button
          onClick={push}
          disabled={busy || !body.trim() || (scope !== 'all' && !targetId)}
          className="w-full py-2.5 rounded-md bg-blood text-bone uppercase tracking-[0.2em] text-sm disabled:opacity-40"
        >
          Send announcement
        </button>
        {msg && <p className="text-bone/60 text-sm mt-2">{msg}</p>}
      </Card>

      <Card title="Clear">
        <div className="flex items-center justify-between">
          <p className="text-bone/60 text-sm">Dismiss the current announcement on everyone's screens.</p>
          <button
            onClick={clear}
            disabled={busy}
            className="px-5 py-3 rounded-md border border-blood/50 text-blood-bright uppercase tracking-[0.15em] text-sm disabled:opacity-40"
          >
            Clear
          </button>
        </div>
      </Card>
    </div>
  );
};
