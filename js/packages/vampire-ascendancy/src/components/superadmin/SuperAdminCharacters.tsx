import { useEffect, useState } from 'react';
import {
  adminListCharacters,
  adminListHouses,
  adminGetCharacter,
  adminUpdateCharacter,
  adminGenerateCharacterTags,
  adminListMysteries,
  adminGetMystery,
  adminGetCharacterContentForMystery,
  adminUpdateCharacterContentForMystery,
} from '../../superAdminApi';
import type { AdminCharacter, AdminMystery, AdminMysteryBeat, AdminMysterySecret } from '../../superAdminApi';
import type { GMCharacterFull } from '../../gmApi';
import type { House } from '../../types';
import { ApiError } from '../../api';
import { Card } from '../gm/GameSection';
import { Field, ListEditor, RemoveBtn } from './SuperAdminShared';

// The shared character roster: bio, tags, and — per mystery — secrets and
// pre-/post-Act-1 context. Missions stay Mysteries-tab-only for now (see
// MYSTERY_REQUIREMENTS.md); the rest is common enough to author
// character-first (pick who, then which mystery, then what they know) that
// it's also editable from here, not just mystery-first. Sigils, portraits,
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

const CharacterEditor = ({
  characterId,
  houses,
  onSaved,
}: {
  characterId: string;
  houses: House[];
  onSaved: () => void;
}) => {
  const [c, setC] = useState<Omit<
    GMCharacterFull,
    'sigil' | 'imageUrl' | 'playerName' | 'secrets' | 'preAct1Context' | 'postAct1Context' | 'missions'
  > | null>(null);
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
        bio: c.bio,
        tags: c.tags,
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
            Reads the saved pre-event bio, plus this character's secrets, missions, and post-Act-1
            context in every mystery they appear in — save your other edits first if you've changed
            them.
          </p>
          {c.tagsGenerationStatus === 'failed' && c.tagsGenerationError && (
            <p className="text-blood-bright text-xs">{c.tagsGenerationError}</p>
          )}
          {tagsError && <p className="text-blood-bright text-xs">{tagsError}</p>}
        </div>
      </Field>
      <Field label="Bio">
        <textarea className={input} rows={4} value={c.bio} onChange={(e) => set('bio', e.target.value)} />
      </Field>

      <CharacterSecretsByMystery characterId={c.id} />
      <p className="text-[11px] text-bone/40 -mt-1">
        Missions are still only edited from the Mysteries tab's per-character content editor.
      </p>

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

// Character-first secrets: pick which mystery/sub-plot, then edit this
// character's secrets for it (each with a beat) — the reverse direction of
// the Mysteries tab's own "walk the cast" editor, for when it's easier to
// start from a character you're already looking at. Both write to the same
// place (PUT .../content), so edits from either side stay in sync.
const CharacterSecretsByMystery = ({ characterId }: { characterId: string }) => {
  const [mysteries, setMysteries] = useState<AdminMystery[]>([]);
  const [mysteryId, setMysteryId] = useState('');

  useEffect(() => {
    adminListMysteries()
      .then((d) => setMysteries(d.mysteries))
      .catch(() => {});
  }, []);

  return (
    <Field label="Secrets — pick a mystery or sub-plot to add or edit this character's secrets for it">
      <select
        className="w-full rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm"
        value={mysteryId}
        onChange={(e) => setMysteryId(e.target.value)}
      >
        <option value="">— choose a mystery or sub-plot —</option>
        {mysteries.map((m) => (
          <option key={m.id} value={m.id}>
            {m.name}
            {m.isSubplot ? ' (sub-plot)' : ''}
          </option>
        ))}
      </select>
      {mysteryId && <SecretsForMystery key={mysteryId} characterId={characterId} mysteryId={mysteryId} />}
    </Field>
  );
};

const SecretsForMystery = ({ characterId, mysteryId }: { characterId: string; mysteryId: string }) => {
  const [beats, setBeats] = useState<AdminMysteryBeat[]>([]);
  const [secrets, setSecrets] = useState<AdminMysterySecret[] | null>(null);
  // Missions aren't shown here, but the save endpoint replaces a
  // character's whole mystery-scoped content in one call — load and echo
  // them back unchanged so saving from here doesn't clobber whatever's
  // already set on the Mysteries tab.
  const [missions, setMissions] = useState<Parameters<typeof adminUpdateCharacterContentForMystery>[2]['missions']>(
    []
  );
  const [preAct1Context, setPreAct1Context] = useState('');
  const [postAct1Context, setPostAct1Context] = useState('');
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState<string | null>(null);

  useEffect(() => {
    setSecrets(null);
    setNote(null);
    Promise.all([adminGetMystery(mysteryId), adminGetCharacterContentForMystery(mysteryId, characterId)])
      .then(([m, content]) => {
        setBeats(m.beats);
        setSecrets(content.secrets);
        setMissions(content.missions);
        setPreAct1Context(content.preAct1Context);
        setPostAct1Context(content.postAct1Context);
      })
      .catch((e) => setNote(e instanceof ApiError ? e.message : 'Could not load secrets for this mystery.'));
  }, [mysteryId, characterId]);

  if (!secrets) return <p className="text-bone/50 text-xs mt-1.5">{note || 'Loading…'}</p>;

  const input = 'w-full rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm';

  const save = async () => {
    setBusy(true);
    setNote(null);
    try {
      await adminUpdateCharacterContentForMystery(mysteryId, characterId, {
        secrets: secrets.map((s) => ({ body: s.body, beatId: s.beatId })),
        missions,
        preAct1Context,
        postAct1Context,
      });
      setNote('Saved.');
    } catch (e) {
      setNote(e instanceof Error ? e.message : 'Save failed.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="mt-2 rounded-md border border-gold/25 bg-black/20 p-3">
      {secrets.length === 0 && (
        <p className="text-gold/80 text-xs mb-2">
          No secrets yet for this mystery — this character can't be invited to a Toast running it
          until they have at least one.
        </p>
      )}
      <ListEditor
        label="Secrets"
        addLabel="+ Add secret"
        onAdd={() => setSecrets([...secrets, { ordinal: secrets.length + 1, body: '', beatId: null }])}
      >
        {secrets.map((s, i) => (
          <div key={i} className="flex gap-2 items-start">
            <span className="text-gold text-xs mt-2 w-4">{i + 1}</span>
            <div className="flex-1 flex flex-col gap-1.5">
              <textarea
                className={input}
                rows={2}
                value={s.body}
                onChange={(e) => setSecrets(secrets.map((x, j) => (j === i ? { ...x, body: e.target.value } : x)))}
              />
              <select
                className={input}
                value={s.beatId ?? ''}
                onChange={(e) =>
                  setSecrets(secrets.map((x, j) => (j === i ? { ...x, beatId: e.target.value || null } : x)))
                }
              >
                <option value="">— which beat does this reveal? —</option>
                {beats.map((b) => (
                  <option key={b.id} value={b.id}>
                    {b.title || b.description.slice(0, 60) || '(untitled beat)'}
                  </option>
                ))}
              </select>
            </div>
            <RemoveBtn onClick={() => setSecrets(secrets.filter((_, j) => j !== i))} />
          </div>
        ))}
      </ListEditor>

      <div className="mt-3 flex flex-col gap-3">
        <Field label="Pre-Act-1 context — who they are going into this mystery specifically">
          <textarea
            className={`${input} mt-1.5`}
            rows={3}
            value={preAct1Context}
            onChange={(e) => setPreAct1Context(e.target.value)}
          />
        </Field>
        <Field label="Post-Act-1 context — what happens to them after Act One, in this mystery">
          <textarea
            className={`${input} mt-1.5`}
            rows={3}
            value={postAct1Context}
            onChange={(e) => setPostAct1Context(e.target.value)}
          />
        </Field>
      </div>

      <div className="mt-3 flex items-center gap-3">
        <button
          type="button"
          onClick={save}
          disabled={busy}
          className="py-2 px-5 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
        >
          {busy ? 'Saving…' : 'Save secrets'}
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

