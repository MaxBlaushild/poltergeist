import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
  gmListPlayers,
  gmListCharacters,
  gmListHouses,
  gmUpdatePlayer,
  gmSetCharacterPortrait,
  gmProvisionPlayers,
} from '../../gmApi';
import type { GMPlayer, GMCharacter, GMProvisionedSeat } from '../../gmApi';
import type { House } from '../../types';
import { Card } from './GameSection';
import { DossierPreviewLoader } from './DossierPreview';

export const PlayersSection = () => {
  const [players, setPlayers] = useState<GMPlayer[]>([]);
  const [characters, setCharacters] = useState<GMCharacter[]>([]);
  const [houses, setHouses] = useState<House[]>([]);
  const [loading, setLoading] = useState(true);

  const load = () => {
    Promise.all([gmListPlayers(), gmListCharacters(), gmListHouses()])
      .then(([p, c, h]) => {
        setPlayers(p.players);
        setCharacters(c.characters);
        setHouses(h.houses);
      })
      .finally(() => setLoading(false));
  };
  useEffect(() => {
    load();
  }, []);

  if (loading) return <p className="text-bone/50">Loading roster…</p>;

  const assignable = characters.filter((c) => c.roleType === 'player');

  // Alphabetical by character name (unassigned slots last) so a GM can find a
  // person quickly.
  const sorted = [...players].sort((a, b) => {
    const an = a.character?.name ?? '';
    const bn = b.character?.name ?? '';
    if (!an) return bn ? 1 : 0;
    if (!bn) return -1;
    return an.localeCompare(bn);
  });

  return (
    <div className="flex flex-col gap-3">
      <ProvisionSeats onProvisioned={load} />
      <p className="text-bone/50 text-sm">
        {players.length} player links · assign characters, edit names, and set this Toast's portrait.
        Bios, secrets, and missions are shared content, edited in the{' '}
        <Link to="/admin" className="text-gold underline underline-offset-2">
          Super Admin dashboard
        </Link>
        .
      </p>
      {sorted.map((p) => (
        <PlayerRow key={p.id} player={p} characters={assignable} houses={houses} onSaved={load} />
      ))}
    </div>
  );
};

// Replaces the old `go run ./cmd/provision` CLI step: create a player slot +
// sigil for every included, non-optional character that doesn't have one yet.
// Safe to re-run as the roster firms up.
const ProvisionSeats = ({ onProvisioned }: { onProvisioned: () => void }) => {
  const [busy, setBusy] = useState(false);
  const [created, setCreated] = useState<GMProvisionedSeat[] | null>(null);
  const [includeOptional, setIncludeOptional] = useState(false);

  const run = async () => {
    setBusy(true);
    try {
      const res = await gmProvisionPlayers(includeOptional);
      setCreated(res.created);
      onProvisioned();
    } finally {
      setBusy(false);
    }
  };

  return (
    <Card title="Provision seats">
      <p className="text-bone/50 text-sm mb-3">
        Creates a login link + sigil for every included character that doesn't have a player slot yet.
        Safe to re-run — only fills gaps.
      </p>
      <div className="flex items-center gap-3 flex-wrap">
        <label className="flex items-center gap-2 text-sm text-bone/70">
          <input
            type="checkbox"
            checked={includeOptional}
            onChange={(e) => setIncludeOptional(e.target.checked)}
          />
          Include optional (✦) characters
        </label>
        <button
          onClick={run}
          disabled={busy}
          className="py-2 px-4 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
        >
          {busy ? 'Provisioning…' : 'Provision seats'}
        </button>
      </div>
      {created && (
        <div className="mt-3 text-sm">
          {created.length === 0 ? (
            <p className="text-bone/50">Everyone already has a seat.</p>
          ) : (
            <>
              <p className="text-gold mb-1">Created {created.length} new seat(s):</p>
              <ul className="text-bone/70 space-y-0.5">
                {created.map((c) => (
                  <li key={c.characterId}>
                    {c.characterName} — sigil {c.sigil}
                  </li>
                ))}
              </ul>
            </>
          )}
        </div>
      )}
    </Card>
  );
};

const PlayerRow = ({
  player,
  characters,
  houses,
  onSaved,
}: {
  player: GMPlayer;
  characters: GMCharacter[];
  houses: House[];
  onSaved: () => void;
}) => {
  const [label, setLabel] = useState(player.guestLabel);
  const [characterId, setCharacterId] = useState(player.character?.id ?? '');
  const [active, setActive] = useState(player.active);
  const [busy, setBusy] = useState(false);
  const [copied, setCopied] = useState(false);
  const [editingPortrait, setEditingPortrait] = useState(false);
  const [previewing, setPreviewing] = useState(false);
  const { instanceId } = useParams();

  const dirty =
    label !== player.guestLabel ||
    characterId !== (player.character?.id ?? '') ||
    active !== player.active;

  const save = async () => {
    setBusy(true);
    try {
      await gmUpdatePlayer(player.id, { characterId: characterId || null, guestLabel: label, active });
      onSaved();
    } finally {
      setBusy(false);
    }
  };

  const copyLink = () => {
    const link = player.character
      ? `${window.location.origin}/e/${instanceId}/c/${player.character.id}`
      : `${window.location.origin}/e/${instanceId}/login`;
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
          placeholder="Player name (optional)"
          className="rounded-md bg-black/60 border border-blood/40 p-2.5 text-bone"
        />
        <div className="flex items-center gap-3 flex-wrap">
          <label className="flex items-center gap-2 text-sm text-bone/70">
            <input type="checkbox" checked={active} onChange={(e) => setActive(e.target.checked)} />
            Active
          </label>
          {player.character?.sigil && (
            <span className="text-xs text-gold">sigil {player.character.sigil}</span>
          )}
          <span className="text-xs text-bone/40">· {player.btTotal} BT</span>
          {player.character && (
            <button
              onClick={() => setEditingPortrait((e) => !e)}
              className="text-xs text-gold uppercase tracking-[0.15em]"
            >
              {editingPortrait ? '▾ Close portrait' : '▸ Portrait'}
            </button>
          )}
          {player.character && (
            <button
              onClick={() => setPreviewing(true)}
              className="text-xs text-gold/80 uppercase tracking-[0.15em]"
            >
              ↗ Preview dossier
            </button>
          )}
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
            Save assignment
          </button>
        )}
      </div>

      {editingPortrait && player.character && (
        <div className="mt-4 pt-4 border-t border-blood/30">
          <PortraitEditor characterId={player.character.id} onSaved={onSaved} />
        </div>
      )}

      {previewing && player.character && (
        <DossierPreviewLoader
          characterId={player.character.id}
          houses={houses}
          onClose={() => setPreviewing(false)}
        />
      )}
    </Card>
  );
};

// This Toast's portrait for a character — the one character edit a Host/
// Co-Host can make directly; bios/secrets/missions are shared content (see
// the Super Admin dashboard).
const PortraitEditor = ({ characterId, onSaved }: { characterId: string; onSaved: () => void }) => {
  const [url, setUrl] = useState('');
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState<string | null>(null);

  const save = async () => {
    setBusy(true);
    setNote(null);
    try {
      await gmSetCharacterPortrait(characterId, url);
      setNote('Saved.');
      onSaved();
    } catch (e) {
      setNote(e instanceof Error ? e.message : 'Save failed.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="flex flex-col gap-2">
      <label className="flex flex-col gap-1">
        <span className="text-[11px] uppercase tracking-[0.15em] text-bone/50">Portrait URL</span>
        <input
          className="w-full rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://…"
        />
      </label>
      <div className="flex items-center gap-3">
        <button
          onClick={save}
          disabled={busy}
          className="py-2 px-5 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
        >
          {busy ? 'Saving…' : 'Save portrait'}
        </button>
        {note && <span className="text-bone/60 text-sm">{note}</span>}
      </div>
    </div>
  );
};
