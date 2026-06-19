import { useEffect, useState } from 'react';
import { gmListHouses, gmAwardHF, gmListPlayers, gmAwardBT } from '../../gmApi';
import type { GMPlayer } from '../../gmApi';
import type { House, HouseStanding } from '../../types';
import { formatHF } from '../../theme';
import { Card } from './GameSection';

export const AwardsSection = () => {
  return (
    <div className="flex flex-col gap-6">
      <HouseFavorForm />
      <BloodTokenForm />
    </div>
  );
};

const HouseFavorForm = () => {
  const [houses, setHouses] = useState<House[]>([]);
  const [standings, setStandings] = useState<HouseStanding[]>([]);
  const [houseId, setHouseId] = useState('');
  const [delta, setDelta] = useState('');
  const [reason, setReason] = useState('');
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);

  useEffect(() => {
    gmListHouses().then((d) => {
      setHouses(d.houses);
      if (d.houses[0]) setHouseId(d.houses[0].id);
    });
  }, []);

  const award = async () => {
    const n = Number(delta);
    if (!houseId || !n) return;
    setBusy(true);
    setMsg(null);
    try {
      const res = await gmAwardHF(houseId, n, reason);
      setStandings(res.standings);
      setDelta('');
      setReason('');
      setMsg('Recorded.');
    } catch {
      setMsg('Failed.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <Card title="House Favor">
      <select
        value={houseId}
        onChange={(e) => setHouseId(e.target.value)}
        className="w-full mb-2 rounded-md bg-black/60 border border-blood/40 p-3 text-bone"
      >
        {houses.map((h) => (
          <option key={h.id} value={h.id}>
            House of {h.name}
          </option>
        ))}
      </select>
      <div className="flex gap-2 mb-2">
        <input
          value={delta}
          onChange={(e) => setDelta(e.target.value.replace(/[^0-9-]/g, ''))}
          placeholder="±Favor"
          className="w-24 rounded-md bg-black/60 border border-blood/40 p-3 text-bone text-center"
        />
        <input
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          placeholder="Reason"
          className="flex-1 rounded-md bg-black/60 border border-blood/40 p-3 text-bone"
        />
      </div>
      <button
        onClick={award}
        disabled={busy || !delta}
        className="w-full py-2.5 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
      >
        Award Favor
      </button>
      {msg && <p className="text-bone/60 text-sm mt-2">{msg}</p>}
      {standings.length > 0 && (
        <div className="mt-3 text-sm text-bone/70">
          {standings.map((s) => (
            <div key={s.houseId} className="flex justify-between">
              <span>{s.name}</span>
              <span className="text-bone">{formatHF(s.favor)}</span>
            </div>
          ))}
        </div>
      )}
    </Card>
  );
};

const BloodTokenForm = () => {
  const [players, setPlayers] = useState<GMPlayer[]>([]);
  const [playerId, setPlayerId] = useState('');
  const [delta, setDelta] = useState('');
  const [reason, setReason] = useState('');
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);

  const loadPlayers = () =>
    gmListPlayers().then((d) => {
      const assigned = d.players.filter((p) => p.character);
      setPlayers(assigned);
      if (assigned[0] && !playerId) setPlayerId(assigned[0].id);
    });
  useEffect(() => {
    loadPlayers();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const award = async () => {
    const n = Number(delta);
    if (!playerId || !n) return;
    setBusy(true);
    setMsg(null);
    try {
      await gmAwardBT(playerId, n, reason);
      setDelta('');
      setReason('');
      await loadPlayers();
      setMsg('Recorded.');
    } catch {
      setMsg('Failed.');
    } finally {
      setBusy(false);
    }
  };

  const labelFor = (p: GMPlayer) =>
    `${p.character?.name ?? '—'}${p.guestLabel ? ` (${p.guestLabel})` : ''} · ${p.btTotal} BT`;

  return (
    <Card title="Blood Tokens (recorded)">
      <select
        value={playerId}
        onChange={(e) => setPlayerId(e.target.value)}
        className="w-full mb-2 rounded-md bg-black/60 border border-blood/40 p-3 text-bone"
      >
        {players.map((p) => (
          <option key={p.id} value={p.id}>
            {labelFor(p)}
          </option>
        ))}
      </select>
      <div className="flex gap-2 mb-2">
        <input
          value={delta}
          onChange={(e) => setDelta(e.target.value.replace(/[^0-9-]/g, ''))}
          placeholder="±BT"
          className="w-24 rounded-md bg-black/60 border border-blood/40 p-3 text-bone text-center"
        />
        <input
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          placeholder="Reason (e.g. won duel)"
          className="flex-1 rounded-md bg-black/60 border border-blood/40 p-3 text-bone"
        />
      </div>
      <button
        onClick={award}
        disabled={busy || !delta}
        className="w-full py-2.5 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
      >
        Award Tokens
      </button>
      {msg && <p className="text-bone/60 text-sm mt-2">{msg}</p>}
    </Card>
  );
};
