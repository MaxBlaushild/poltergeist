import { useEffect, useState } from 'react';
import {
  adminListMysteries,
  adminCreateMystery,
  adminGetMystery,
  adminUpdateMystery,
  adminListCharacters,
  adminGetCharacterContentForMystery,
  adminUpdateCharacterContentForMystery,
} from '../../superAdminApi';
import type { AdminMystery, AdminMysteryFull, AdminMysterySecret, AdminCharacter } from '../../superAdminApi';
import { Card } from '../gm/GameSection';
import { CharacterBrowser } from '../gm/CharacterBrowser';
import { SuperAdminQuiz } from './SuperAdminQuiz';
import { Field, ListEditor, RemoveBtn } from './SuperAdminShared';

// The underlying story an instance's players are solving — see
// MYSTERY_REQUIREMENTS.md. An instance picks exactly one mystery at
// creation and never changes it; everything here (beats, quiz, character
// secrets) is what a Host is choosing between on that one-time picker.
export const SuperAdminMysteries = () => {
  const [mysteries, setMysteries] = useState<AdminMystery[]>([]);
  const [openId, setOpenId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [newName, setNewName] = useState('');
  const [creating, setCreating] = useState(false);

  const load = () => {
    adminListMysteries()
      .then((d) => setMysteries(d.mysteries))
      .finally(() => setLoading(false));
  };
  useEffect(() => {
    load();
  }, []);

  const create = async () => {
    const name = newName.trim();
    if (!name) return;
    setCreating(true);
    try {
      const res = await adminCreateMystery(name);
      setNewName('');
      load();
      setOpenId(res.id);
    } finally {
      setCreating(false);
    }
  };

  if (loading) return <Card title="Mysteries">Loading…</Card>;

  return (
    <div className="flex flex-col gap-4">
      <Card title="New mystery">
        <div className="flex gap-2">
          <input
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && create()}
            placeholder="e.g. The Ashglass Inheritance"
            className="flex-1 rounded-md bg-black/60 border border-blood/40 p-2.5 text-bone"
          />
          <button
            onClick={create}
            disabled={creating || !newName.trim()}
            className="px-4 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
          >
            Create
          </button>
        </div>
      </Card>

      <Card title={`Mysteries (${mysteries.length})`}>
        <div className="flex flex-col gap-1.5">
          {mysteries.map((m) => (
            <div key={m.id} className="border-b border-blood/15 last:border-0 pb-1.5">
              <div className="flex items-center gap-2">
                <div className="flex-1 min-w-0">
                  <p className="text-bone text-sm">
                    {m.name}
                    {!m.active && (
                      <span className="ml-2 text-[10px] uppercase tracking-[0.15em] text-bone/40 border border-bone/20 rounded-full px-1.5 py-0.5">
                        inactive
                      </span>
                    )}
                  </p>
                  <p className="text-bone/50 text-xs mt-0.5">
                    {m.beatCount} beat{m.beatCount === 1 ? '' : 's'}
                    {m.summary && <span className="text-bone/40"> · {m.summary}</span>}
                  </p>
                </div>
                <button
                  onClick={() => setOpenId((cur) => (cur === m.id ? null : m.id))}
                  className="shrink-0 text-xs text-gold uppercase tracking-[0.15em]"
                >
                  {openId === m.id ? 'Close' : 'Edit'}
                </button>
              </div>
              {openId === m.id && (
                <div className="mt-2">
                  <MysteryEditor mysteryId={m.id} onSaved={load} />
                </div>
              )}
            </div>
          ))}
          {mysteries.length === 0 && <p className="text-bone/50 text-sm">No mysteries yet — create one above.</p>}
        </div>
      </Card>
    </div>
  );
};

type MysterySection = 'story' | 'quiz' | 'secrets';

const MysteryEditor = ({ mysteryId, onSaved }: { mysteryId: string; onSaved: () => void }) => {
  const [m, setM] = useState<AdminMysteryFull | null>(null);
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState<string | null>(null);
  const [section, setSection] = useState<MysterySection>('story');

  useEffect(() => {
    setM(null);
    adminGetMystery(mysteryId).then(setM).catch(() => setNote('Could not load mystery.'));
  }, [mysteryId]);

  if (!m) return <p className="text-bone/50 text-sm">{note || 'Loading mystery…'}</p>;

  const set = <K extends keyof AdminMysteryFull>(k: K, v: AdminMysteryFull[K]) => setM({ ...m, [k]: v });

  const save = async () => {
    setBusy(true);
    setNote(null);
    try {
      await adminUpdateMystery(m.id, {
        name: m.name,
        summary: m.summary,
        fullLore: m.fullLore,
        active: m.active,
        beats: m.beats.map((b) => ({ body: b.body })),
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
  const sections: { id: MysterySection; label: string }[] = [
    { id: 'story', label: 'Story' },
    { id: 'quiz', label: 'Quiz' },
    { id: 'secrets', label: 'Characters' },
  ];

  return (
    <div className="flex flex-col gap-3">
      <nav className="flex gap-1.5">
        {sections.map((s) => (
          <button
            key={s.id}
            onClick={() => setSection(s.id)}
            className={`px-3 py-1.5 rounded-md text-xs uppercase tracking-[0.15em] ${
              section === s.id ? 'bg-blood text-bone' : 'text-bone/60 border border-blood/30'
            }`}
          >
            {s.label}
          </button>
        ))}
      </nav>

      {section === 'story' && (
        <div className="flex flex-col gap-3">
          <Field label="Name">
            <input className={input} value={m.name} onChange={(e) => set('name', e.target.value)} />
          </Field>
          <Field label="Summary — shown on the &quot;Host a Toast&quot; picker">
            <textarea className={input} rows={2} value={m.summary} onChange={(e) => set('summary', e.target.value)} />
          </Field>
          <Field label="Full lore — the detailed &quot;what actually happened&quot;, GM reference only">
            <textarea className={input} rows={6} value={m.fullLore} onChange={(e) => set('fullLore', e.target.value)} />
          </Field>
          <label className="flex items-center gap-2 text-sm text-bone/70">
            <input type="checkbox" checked={m.active} onChange={(e) => set('active', e.target.checked)} />
            Active — shows up on the "Host a Toast" picker
          </label>

          <ListEditor
            label="Beats — discoverable facts about the mystery"
            addLabel="+ Add beat"
            onAdd={() => set('beats', [...m.beats, { id: '', ordinal: m.beats.length + 1, body: '' }])}
          >
            {m.beats.map((b, i) => (
              <div key={i} className="flex gap-2 items-start">
                <span className="text-gold text-xs mt-2 w-4">{i + 1}</span>
                <textarea
                  className={input}
                  rows={2}
                  value={b.body}
                  onChange={(e) => set('beats', m.beats.map((x, j) => (j === i ? { ...x, body: e.target.value } : x)))}
                />
                <RemoveBtn onClick={() => set('beats', m.beats.filter((_, j) => j !== i))} />
              </div>
            ))}
          </ListEditor>
          <p className="text-[11px] text-bone/40">
            Removing a beat that a secret still points at un-sets that secret's beat rather than
            deleting the secret — check the Character secrets tab afterward if you remove one.
          </p>

          <div className="flex items-center gap-3">
            <button
              onClick={save}
              disabled={busy}
              className="py-2 px-5 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
            >
              {busy ? 'Saving…' : 'Save mystery'}
            </button>
            {note && <span className="text-bone/60 text-sm">{note}</span>}
          </div>
        </div>
      )}

      {section === 'quiz' && <SuperAdminQuiz mysteryId={m.id} />}
      {section === 'secrets' && <CharacterSecretsEditor mysteryId={m.id} beats={m.beats} />}
    </div>
  );
};

// Walk the cast and decide what each of them knows — and what happens to
// them — for this mystery. Deliberately reached from the mystery, not the
// character (see MYSTERY_REQUIREMENTS.md's "Super Admin UI"). A character
// with zero secrets here can't be invited to a Toast running this mystery.
const CharacterSecretsEditor = ({ mysteryId, beats }: { mysteryId: string; beats: AdminMysteryFull['beats'] }) => {
  const [characters, setCharacters] = useState<AdminCharacter[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    adminListCharacters()
      .then((d) => setCharacters(d.characters.filter((c) => c.roleType === 'player')))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <p className="text-bone/50 text-sm">Loading cast…</p>;

  const selected = characters.find((c) => c.id === selectedId) || null;

  return (
    <div className="flex flex-col gap-3">
      <CharacterBrowser characters={characters} selectedId={selectedId ?? undefined} onSelect={(c) => setSelectedId(c.id)} />
      {selected && <ContentForCharacter key={selected.id} mysteryId={mysteryId} character={selected} beats={beats} />}
    </div>
  );
};

const ContentForCharacter = ({
  mysteryId,
  character,
  beats,
}: {
  mysteryId: string;
  character: AdminCharacter;
  beats: AdminMysteryFull['beats'];
}) => {
  const [secrets, setSecrets] = useState<AdminMysterySecret[] | null>(null);
  const [postAct1Context, setPostAct1Context] = useState('');
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState<string | null>(null);

  useEffect(() => {
    setSecrets(null);
    setPostAct1Context('');
    adminGetCharacterContentForMystery(mysteryId, character.id)
      .then((d) => {
        setSecrets(d.secrets);
        setPostAct1Context(d.postAct1Context);
      })
      .catch(() => setNote('Could not load this character.'));
  }, [mysteryId, character.id]);

  if (!secrets) return <p className="text-bone/50 text-sm">{note || 'Loading…'}</p>;

  const input = 'w-full rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm';

  const save = async () => {
    setBusy(true);
    setNote(null);
    try {
      await adminUpdateCharacterContentForMystery(mysteryId, character.id, {
        secrets: secrets.map((s) => ({ body: s.body, beatId: s.beatId })),
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
    <Card title={`${character.name} in this mystery`}>
      {secrets.length === 0 && (
        <p className="text-gold/80 text-xs mb-2">
          No secrets yet — {character.name} can't be invited to a Toast running this mystery until
          they have at least one.
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
                    {b.body.slice(0, 60) || '(untitled beat)'}
                  </option>
                ))}
              </select>
            </div>
            <RemoveBtn onClick={() => setSecrets(secrets.filter((_, j) => j !== i))} />
          </div>
        ))}
      </ListEditor>

      <Field label="Post-Act-1 context — what happens to them after Act One, in this mystery">
        <textarea
          className={`${input} mt-1.5`}
          rows={4}
          value={postAct1Context}
          onChange={(e) => setPostAct1Context(e.target.value)}
        />
      </Field>

      <div className="mt-3 flex items-center gap-3">
        <button
          onClick={save}
          disabled={busy}
          className="py-2 px-5 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
        >
          {busy ? 'Saving…' : 'Save'}
        </button>
        {note && <span className="text-bone/60 text-sm">{note}</span>}
      </div>
    </Card>
  );
};
