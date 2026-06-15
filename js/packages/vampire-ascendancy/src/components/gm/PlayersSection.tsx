import { useEffect, useState } from 'react';
import { gmListPlayers, gmListCharacters, gmUpdatePlayer } from '../../gmApi';
import type { GMPlayer, GMCharacter } from '../../gmApi';
import { Card } from './GameSection';

export const PlayersSection = () => {
  const [players, setPlayers] = useState<GMPlayer[]>([]);
  const [characters, setCharacters] = useState<GMCharacter[]>([]);
  const [loading, setLoading] = useState(true);

  const load = () => {
    Promise.all([gmListPlayers(), gmListCharacters()])
      .then(([p, c]) => {
        setPlayers(p.players);
        setCharacters(c.characters);
      })
      .finally(() => setLoading(false));
  };
  useEffect(() => {
    load();
  }, []);

  if (loading) return <p className="text-bone/50">Loading players…</p>;

  // Only standard players are assignable to guests.
  const assignable = characters.filter((c) => c.roleType === 'player');

  return (
    <div className="flex flex-col gap-3">
      <p className="text-bone/50 text-sm">
        {players.length} player links · assign characters and toggle who is active before the
        evening begins.
      </p>
      {players.map((p) => (
        <PlayerRow key={p.id} player={p} characters={assignable} onSaved={load} />
      ))}
    </div>
  );
};

const PlayerRow = ({
  player,
  characters,
  onSaved,
}: {
  player: GMPlayer;
  characters: GMCharacter[];
  onSaved: () => void;
}) => {
  const [label, setLabel] = useState(player.guestLabel);
  const [characterId, setCharacterId] = useState(player.character?.id ?? '');
  const [active, setActive] = useState(player.active);
  const [busy, setBusy] = useState(false);
  const [copied, setCopied] = useState(false);

  const dirty =
    label !== player.guestLabel ||
    characterId !== (player.character?.id ?? '') ||
    active !== player.active;

  const save = async () => {
    setBusy(true);
    try {
      await gmUpdatePlayer(player.id, {
        characterId: characterId || null,
        guestLabel: label,
        active,
      });
      onSaved();
    } finally {
      setBusy(false);
    }
  };

  const copyLink = () => {
    // The link pre-selects the character; the sigil is the credential.
    const link = player.character
      ? `${window.location.origin}/c/${player.character.id}`
      : `${window.location.origin}/login`;
    navigator.clipboard?.writeText(link);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <Card title={player.character?.name ?? 'Unassigned'}>
      <div className="flex flex-col gap-2">
        <select
          value={characterId}
          onChange={(e) => setCharacterId(e.target.value)}
          className="rounded-md bg-black/60 border border-blood/40 p-2.5 text-bone"
        >
          <option value="">— Unassigned —</option>
          {characters.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name} ({c.house})
            </option>
          ))}
        </select>
        <input
          value={label}
          onChange={(e) => setLabel(e.target.value)}
          placeholder="Guest name (optional)"
          className="rounded-md bg-black/60 border border-blood/40 p-2.5 text-bone"
        />
        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 text-sm text-bone/70">
            <input type="checkbox" checked={active} onChange={(e) => setActive(e.target.checked)} />
            Active
          </label>
          {player.character?.sigil && (
            <span className="text-xs text-gold">sigil {player.character.sigil}</span>
          )}
          <span className="text-xs text-bone/40">· {player.btTotal} BT</span>
          <button onClick={copyLink} className="ml-auto text-xs text-blood-bright uppercase tracking-[0.15em]">
            {copied ? 'Copied!' : 'Copy link'}
          </button>
        </div>
        {dirty && (
          <button
            onClick={save}
            disabled={busy}
            className="mt-1 py-2 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
          >
            Save
          </button>
        )}
      </div>
    </Card>
  );
};
