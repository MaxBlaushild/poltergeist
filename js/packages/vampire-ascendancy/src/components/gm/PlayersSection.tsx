import { useEffect, useState } from 'react';
import {
  gmListPlayers,
  gmListCharacters,
  gmListHouses,
  gmUpdatePlayer,
  gmGetCharacter,
  gmUpdateCharacter,
} from '../../gmApi';
import type { GMPlayer, GMCharacter, GMCharacterFull, GMMissionEdit } from '../../gmApi';
import type { House } from '../../types';
import { Card } from './GameSection';

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
      <p className="text-bone/50 text-sm">
        {players.length} player links · assign characters, edit names, and expand a row to edit that
        character's bios, secrets, and missions.
      </p>
      {sorted.map((p) => (
        <PlayerRow key={p.id} player={p} characters={assignable} houses={houses} onSaved={load} />
      ))}
    </div>
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
  const [editing, setEditing] = useState(false);

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
              onClick={() => setEditing((e) => !e)}
              className="text-xs text-gold uppercase tracking-[0.15em]"
            >
              {editing ? '▾ Close editor' : '▸ Edit character'}
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

      {editing && player.character && (
        <div className="mt-4 pt-4 border-t border-blood/30">
          <CharacterEditor characterId={player.character.id} houses={houses} onSaved={onSaved} />
        </div>
      )}
    </Card>
  );
};

const blankMission = (): GMMissionEdit => ({ tier: 'easy', rewardBt: 2, prompt: '', answerFormat: '' });

const CharacterEditor = ({
  characterId,
  houses,
  onSaved,
}: {
  characterId: string;
  houses: House[];
  onSaved: () => void;
}) => {
  const [c, setC] = useState<GMCharacterFull | null>(null);
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState<string | null>(null);

  useEffect(() => {
    gmGetCharacter(characterId).then(setC).catch(() => setNote('Could not load character.'));
  }, [characterId]);

  if (!c) return <p className="text-bone/50 text-sm">{note || 'Loading character…'}</p>;

  const set = <K extends keyof GMCharacterFull>(k: K, v: GMCharacterFull[K]) =>
    setC({ ...c, [k]: v });

  const save = async () => {
    setBusy(true);
    setNote(null);
    try {
      await gmUpdateCharacter(c.id, {
        name: c.name,
        title: c.title,
        roleType: c.roleType,
        houseId: c.houseId,
        preEventInfo: c.preEventInfo,
        postAct1Context: c.postAct1Context,
        imageUrl: c.imageUrl,
        playerName: c.playerName,
        secrets: c.secrets.map((s) => s.body),
        missions: c.missions.map((m) => ({
          tier: m.tier,
          rewardBt: m.rewardBt,
          prompt: m.prompt,
          answerFormat: m.answerFormat,
        })),
      });
      setNote('Saved.');
      onSaved();
    } catch (e) {
      setNote(e instanceof Error ? e.message : 'Save failed.');
    } finally {
      setBusy(false);
    }
  };

  const input = 'w-full rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm';

  return (
    <div className="flex flex-col gap-3">
      <Field label="Name">
        <input className={input} value={c.name} onChange={(e) => set('name', e.target.value)} />
      </Field>
      <Field label="Title">
        <input className={input} value={c.title} onChange={(e) => set('title', e.target.value)} />
      </Field>
      <div className="grid grid-cols-2 gap-3">
        <Field label="House">
          <select className={input} value={c.houseId ?? ''} onChange={(e) => set('houseId', e.target.value || null)}>
            <option value="">— none —</option>
            {houses.map((h) => (
              <option key={h.id} value={h.id}>
                {h.name}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Role">
          <select className={input} value={c.roleType} onChange={(e) => set('roleType', e.target.value)}>
            <option value="player">player</option>
            <option value="gm">gm</option>
            <option value="npc">npc</option>
          </select>
        </Field>
      </div>
      <Field label="Player name">
        <input className={input} value={c.playerName} onChange={(e) => set('playerName', e.target.value)} />
      </Field>
      <Field label="Portrait URL">
        <input className={input} value={c.imageUrl} onChange={(e) => set('imageUrl', e.target.value)} placeholder="https://…" />
      </Field>
      <Field label="Pre-event bio">
        <textarea className={input} rows={4} value={c.preEventInfo} onChange={(e) => set('preEventInfo', e.target.value)} />
      </Field>
      <Field label="Post-act bio">
        <textarea className={input} rows={4} value={c.postAct1Context} onChange={(e) => set('postAct1Context', e.target.value)} />
      </Field>

      <ListEditor
        label="Secrets"
        addLabel="+ Add secret"
        onAdd={() => set('secrets', [...c.secrets, { ordinal: c.secrets.length + 1, body: '' }])}
      >
        {c.secrets.map((s, i) => (
          <div key={i} className="flex gap-2 items-start">
            <span className="text-gold text-xs mt-2 w-4">{i + 1}</span>
            <textarea
              className={input}
              rows={2}
              value={s.body}
              onChange={(e) =>
                set('secrets', c.secrets.map((x, j) => (j === i ? { ...x, body: e.target.value } : x)))
              }
            />
            <RemoveBtn onClick={() => set('secrets', c.secrets.filter((_, j) => j !== i))} />
          </div>
        ))}
      </ListEditor>

      <ListEditor
        label="Missions"
        addLabel="+ Add mission"
        onAdd={() => set('missions', [...c.missions, { ordinal: c.missions.length + 1, ...blankMission() }])}
      >
        {c.missions.map((m, i) => (
          <div key={i} className="rounded-md border border-blood/30 p-2 flex flex-col gap-2">
            <div className="flex items-center gap-2">
              <span className="text-gold text-xs w-4">{i + 1}</span>
              <select
                className={`${input} w-28`}
                value={m.tier}
                onChange={(e) => set('missions', c.missions.map((x, j) => (j === i ? { ...x, tier: e.target.value } : x)))}
              >
                <option value="easy">easy</option>
                <option value="medium">medium</option>
                <option value="hard">hard</option>
              </select>
              <input
                type="number"
                className={`${input} w-20`}
                value={m.rewardBt}
                onChange={(e) =>
                  set('missions', c.missions.map((x, j) => (j === i ? { ...x, rewardBt: Number(e.target.value) || 0 } : x)))
                }
              />
              <span className="text-bone/40 text-xs">BT</span>
              <RemoveBtn onClick={() => set('missions', c.missions.filter((_, j) => j !== i))} />
            </div>
            <textarea
              className={input}
              rows={2}
              placeholder="Mission prompt"
              value={m.prompt}
              onChange={(e) => set('missions', c.missions.map((x, j) => (j === i ? { ...x, prompt: e.target.value } : x)))}
            />
            <input
              className={input}
              placeholder="What to submit (answer format)"
              value={m.answerFormat}
              onChange={(e) => set('missions', c.missions.map((x, j) => (j === i ? { ...x, answerFormat: e.target.value } : x)))}
            />
          </div>
        ))}
      </ListEditor>

      <div className="flex items-center gap-3">
        <button
          onClick={save}
          disabled={busy}
          className="py-2 px-5 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
        >
          {busy ? 'Saving…' : 'Save character'}
        </button>
        {note && <span className="text-bone/60 text-sm">{note}</span>}
      </div>
    </div>
  );
};

const Field = ({ label, children }: { label: string; children: React.ReactNode }) => (
  <label className="flex flex-col gap-1">
    <span className="text-[11px] uppercase tracking-[0.15em] text-bone/50">{label}</span>
    {children}
  </label>
);

const ListEditor = ({
  label,
  addLabel,
  onAdd,
  children,
}: {
  label: string;
  addLabel: string;
  onAdd: () => void;
  children: React.ReactNode;
}) => (
  <div className="flex flex-col gap-2">
    <div className="flex items-center justify-between">
      <span className="text-[11px] uppercase tracking-[0.15em] text-bone/50">{label}</span>
      <button onClick={onAdd} className="text-xs text-gold uppercase tracking-[0.15em]">
        {addLabel}
      </button>
    </div>
    {children}
  </div>
);

const RemoveBtn = ({ onClick }: { onClick: () => void }) => (
  <button
    onClick={onClick}
    className="shrink-0 mt-1 w-6 h-6 rounded-full border border-blood/50 text-blood-bright text-xs leading-none"
    aria-label="Remove"
  >
    ×
  </button>
);
