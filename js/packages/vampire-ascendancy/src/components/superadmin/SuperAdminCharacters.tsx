import { useEffect, useState } from 'react';
import {
  adminListCharacters,
  adminListHouses,
  adminGetCharacter,
  adminUpdateCharacter,
  adminGenerateCharacterTags,
} from '../../superAdminApi';
import type { AdminCharacter } from '../../superAdminApi';
import type { GMCharacterFull, GMMissionEdit } from '../../gmApi';
import type { House } from '../../types';
import { ApiError } from '../../api';
import { Card } from '../gm/GameSection';

// The shared character roster: bios, secrets, missions. Sigils, portraits,
// and the real guest playing a character are per-instance — see the GM
// console's Players tab for those.
export const SuperAdminCharacters = () => {
  const [characters, setCharacters] = useState<AdminCharacter[]>([]);
  const [houses, setHouses] = useState<House[]>([]);
  const [openId, setOpenId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState('');

  useEffect(() => {
    Promise.all([adminListCharacters(), adminListHouses()])
      .then(([c, h]) => {
        setCharacters(c.characters);
        setHouses(h.houses);
      })
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <Card title="Characters">Loading…</Card>;

  const sorted = [...characters].sort((a, b) => a.name.localeCompare(b.name));
  const q = query.trim().toLowerCase();
  const filtered = q ? sorted.filter((c) => c.name.toLowerCase().includes(q)) : sorted;

  return (
    <Card title={`Characters (${characters.length}) — shared, affects every Toast`}>
      <input
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search by name…"
        className="w-full mb-3 rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm"
      />
      <div className="flex flex-col gap-1.5">
        {filtered.map((c) => (
          <div key={c.id} className="border-b border-blood/15 last:border-0 pb-1.5">
            <div className="flex items-center gap-2">
              <div className="flex-1 min-w-0">
                <p className="text-bone text-sm">
                  {c.name}
                  {c.isOptional && <span className="text-gold/70 ml-1">✦</span>}
                  <span className="text-bone/40 mx-2">•</span>
                  <span className="text-bone/50">{c.house || c.roleType}</span>
                </p>
                {c.tags && c.tags.length > 0 && (
                  <p className="mt-0.5 text-[10px] text-gold/70 truncate">{c.tags.join(' · ')}</p>
                )}
              </div>
              <button
                onClick={() => setOpenId((cur) => (cur === c.id ? null : c.id))}
                className="shrink-0 text-xs text-gold uppercase tracking-[0.15em]"
              >
                {openId === c.id ? 'Close' : 'Edit'}
              </button>
            </div>
            {openId === c.id && (
              <div className="mt-2">
                <CharacterEditor characterId={c.id} houses={houses} onSaved={() => {}} />
              </div>
            )}
          </div>
        ))}
      </div>
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
  const [c, setC] = useState<Omit<GMCharacterFull, 'sigil' | 'imageUrl' | 'playerName'> | null>(null);
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState<string | null>(null);
  const [tagsError, setTagsError] = useState<string | null>(null);

  useEffect(() => {
    adminGetCharacter(characterId).then(setC).catch(() => setNote('Could not load character.'));
  }, [characterId]);

  // While a tag-generation job is in flight, poll for its result — same
  // "enqueue → poll → done" shape as quiz-grading. Stops once the status
  // leaves queued/generating (generated, failed, or never run).
  const generating = c?.tagsGenerationStatus === 'queued' || c?.tagsGenerationStatus === 'generating';
  useEffect(() => {
    if (!generating) return;
    const id = setInterval(() => {
      adminGetCharacter(characterId).then(setC).catch(() => {});
    }, 2000);
    return () => clearInterval(id);
  }, [generating, characterId]);

  if (!c) return <p className="text-bone/50 text-sm">{note || 'Loading character…'}</p>;

  const set = <K extends keyof typeof c>(k: K, v: (typeof c)[K]) => setC({ ...c, [k]: v });

  const generateTags = async () => {
    setTagsError(null);
    try {
      await adminGenerateCharacterTags(c.id);
      setC({ ...c, tagsGenerationStatus: 'queued', tagsGenerationError: '' });
    } catch (e) {
      setTagsError(e instanceof ApiError ? e.message : 'Could not start tag generation.');
    }
  };

  const save = async () => {
    setBusy(true);
    setNote(null);
    try {
      await adminUpdateCharacter(c.id, {
        name: c.name,
        title: c.title,
        roleType: c.roleType,
        houseId: c.houseId,
        preEventInfo: c.preEventInfo,
        postAct1Context: c.postAct1Context,
        tags: c.tags,
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
      <Field label="Tags — personality/trait labels for the Invites picker (comma-separated)">
        <div className="flex flex-col gap-1.5">
          <div className="flex gap-2 items-center">
            <TagsInput
              key={`${c.id}:${c.tags.join('|')}`}
              className={`${input} flex-1`}
              tags={c.tags}
              onChange={(tags) => set('tags', tags)}
            />
            <button
              onClick={generateTags}
              disabled={generating}
              className="shrink-0 px-3 py-2 rounded-md border border-gold/50 text-gold text-xs uppercase tracking-[0.15em] disabled:opacity-40 whitespace-nowrap"
            >
              {generating
                ? c.tagsGenerationStatus === 'queued'
                  ? 'Queued…'
                  : 'Generating…'
                : '✨ Generate with AI'}
            </button>
          </div>
          <p className="text-[11px] text-bone/40">
            Reads the saved bio, secrets, and missions — save your other edits first if you've changed
            them.
          </p>
          {c.tagsGenerationStatus === 'failed' && c.tagsGenerationError && (
            <p className="text-blood-bright text-xs">{c.tagsGenerationError}</p>
          )}
          {tagsError && <p className="text-blood-bright text-xs">{tagsError}</p>}
        </div>
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

// A comma-separated tag list, kept as free-typed text locally (not
// re-derived from the parsed array on every keystroke) so typing a
// trailing comma/space to start the next tag doesn't get collapsed away
// mid-edit. Parses back into the array on blur.
const TagsInput = ({
  tags,
  onChange,
  className,
}: {
  tags: string[];
  onChange: (tags: string[]) => void;
  className: string;
}) => {
  const [text, setText] = useState(tags.join(', '));

  const commit = () => {
    const parsed = text.split(',').map((t) => t.trim()).filter(Boolean);
    onChange(parsed);
    setText(parsed.join(', '));
  };

  return (
    <input
      className={className}
      value={text}
      onChange={(e) => setText(e.target.value)}
      onBlur={commit}
      onKeyDown={(e) => e.key === 'Enter' && commit()}
      placeholder="musical, gambler, aggressive, risk taker"
    />
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
