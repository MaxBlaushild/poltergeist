import { useEffect, useState } from 'react';
import {
  adminListMysteries,
  adminCreateMystery,
  adminGetMystery,
  adminUpdateMystery,
  adminListCharacters,
  adminGetCharacterContentForMystery,
  adminUpdateCharacterContentForMystery,
  adminListBeatSecrets,
  adminCreateBeatSecret,
  adminUpdateSecretBody,
  adminDeleteSecret,
} from '../../superAdminApi';
import type {
  AdminMystery,
  AdminMysteryFull,
  AdminMysteryBeat,
  AdminMysterySecret,
  AdminMysteryMission,
  AdminCharacter,
  AdminBeatSecret,
} from '../../superAdminApi';
import { Card } from '../gm/GameSection';
import { CharacterBrowser } from '../gm/CharacterBrowser';
import { SuperAdminQuiz } from './SuperAdminQuiz';
import { Field, ListEditor, RemoveBtn } from './SuperAdminShared';

// The underlying story an instance's players are solving — see
// MYSTERY_REQUIREMENTS.md. An instance picks exactly one mystery at
// creation (this component with isSubplot=false) plus zero or many
// subplots (isSubplot=true — a sibling, not a separate table: same row
// shape, same editor, just filtered/created as the other kind). Everything
// here (beats, quiz, character secrets/missions/context) is what a Host is
// choosing between on that one-time picker. Subplots skip the Quiz section
// (an instance's quiz is always just its main mystery's) and don't gate
// invite eligibility the way a mystery's secrets do.
export const SuperAdminMysteries = ({ isSubplot }: { isSubplot: boolean }) => {
  const noun = isSubplot ? 'subplot' : 'mystery';
  const [all, setAll] = useState<AdminMystery[]>([]);
  const [openId, setOpenId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [newName, setNewName] = useState('');
  const [creating, setCreating] = useState(false);

  const load = () => {
    adminListMysteries()
      .then((d) => setAll(d.mysteries))
      .finally(() => setLoading(false));
  };
  useEffect(() => {
    load();
  }, []);

  const mysteries = all.filter((m) => m.isSubplot === isSubplot);

  const create = async () => {
    const name = newName.trim();
    if (!name) return;
    setCreating(true);
    try {
      const res = await adminCreateMystery(name, isSubplot);
      setNewName('');
      load();
      setOpenId(res.id);
    } finally {
      setCreating(false);
    }
  };

  if (loading) return <Card title={isSubplot ? 'Sub-plots' : 'Mysteries'}>Loading…</Card>;

  return (
    <div className="flex flex-col gap-4">
      <Card title={`New ${noun}`}>
        <div className="flex gap-2">
          <input
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && create()}
            placeholder={isSubplot ? 'e.g. The Missing Ledger' : 'e.g. The Ashglass Inheritance'}
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

      <Card title={`${isSubplot ? 'Sub-plots' : 'Mysteries'} (${mysteries.length})`}>
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
          {mysteries.length === 0 && (
            <p className="text-bone/50 text-sm">No {noun}s yet — create one above.</p>
          )}
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
  // Which beats have their "who knows this" panel open — keyed by beat id,
  // so it only applies to beats that have already been saved (see below).
  const [expandedBeatIds, setExpandedBeatIds] = useState<Set<string>>(new Set());

  useEffect(() => {
    setM(null);
    adminGetMystery(mysteryId).then(setM).catch(() => setNote('Could not load mystery.'));
  }, [mysteryId]);

  if (!m) return <p className="text-bone/50 text-sm">{note || 'Loading…'}</p>;

  const set = <K extends keyof AdminMysteryFull>(k: K, v: AdminMysteryFull[K]) => setM({ ...m, [k]: v });

  const toggleBeat = (id: string) =>
    setExpandedBeatIds((cur) => {
      const next = new Set(cur);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });

  const save = async () => {
    setBusy(true);
    setNote(null);
    try {
      // id is passed through (not just body) so a beat that already exists
      // keeps its id — and any secret's beatId pointing at it — instead of
      // every save silently regenerating every beat's id.
      await adminUpdateMystery(m.id, {
        name: m.name,
        summary: m.summary,
        fullLore: m.fullLore,
        active: m.active,
        isSubplot: m.isSubplot,
        beats: m.beats.map((b) => ({ id: b.id || undefined, title: b.title, description: b.description })),
      });
      setNote('Saved.');
      onSaved();
      adminGetMystery(m.id).then(setM);
    } catch (e) {
      setNote(e instanceof Error ? e.message : 'Save failed.');
    } finally {
      setBusy(false);
    }
  };

  const noun = m.isSubplot ? 'subplot' : 'mystery';
  const input = 'w-full rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm';
  const sections: { id: MysterySection; label: string }[] = [
    { id: 'story', label: 'Story' },
    // Subplots don't have their own quiz — an instance's quiz is always
    // just its main mystery's (see MYSTERY_REQUIREMENTS.md).
    ...(m.isSubplot ? [] : [{ id: 'quiz' as const, label: 'Quiz' }]),
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
          <label className="flex items-center gap-2 text-sm text-bone/70">
            <input type="checkbox" checked={m.isSubplot} onChange={(e) => set('isSubplot', e.target.checked)} />
            Sub-plot — a Host can pick zero or many of these, in addition to their one required
            mystery. Doesn't gate invite eligibility or have its own quiz.
          </label>

          <ListEditor
            label="Beats — discoverable facts about the mystery"
            addLabel="+ Add beat"
            onAdd={() => set('beats', [...m.beats, { id: '', ordinal: m.beats.length + 1, title: '', description: '' }])}
          >
            {m.beats.map((b, i) => (
              <div key={i} className="flex flex-col gap-1">
                <div className="flex gap-2 items-start">
                  <span className="text-gold text-xs mt-2 w-4">{i + 1}</span>
                  <div className="flex-1 flex flex-col gap-1.5">
                    <input
                      className={input}
                      placeholder="Title"
                      value={b.title}
                      onChange={(e) => set('beats', m.beats.map((x, j) => (j === i ? { ...x, title: e.target.value } : x)))}
                    />
                    <textarea
                      className={input}
                      rows={2}
                      placeholder="Description — what this beat actually reveals"
                      value={b.description}
                      onChange={(e) =>
                        set('beats', m.beats.map((x, j) => (j === i ? { ...x, description: e.target.value } : x)))
                      }
                    />
                  </div>
                  <RemoveBtn onClick={() => set('beats', m.beats.filter((_, j) => j !== i))} />
                </div>
                {b.id ? (
                  <button
                    type="button"
                    onClick={() => toggleBeat(b.id)}
                    className="self-start ml-6 text-[11px] uppercase tracking-[0.15em] text-gold/70 hover:text-gold"
                  >
                    {expandedBeatIds.has(b.id) ? '▾ Hide who knows this' : '▸ Who knows this'}
                  </button>
                ) : (
                  <p className="ml-6 text-[11px] text-bone/30">Save to assign secrets to this beat.</p>
                )}
                {b.id && expandedBeatIds.has(b.id) && <BeatSecretsPanel mysteryId={m.id} beat={b} />}
              </div>
            ))}
          </ListEditor>
          <p className="text-[11px] text-bone/40">
            Renaming a beat's text keeps its id (and any secrets pointing at it) — only removing it
            un-sets those secrets' beat rather than deleting them.
          </p>

          <div className="flex items-center gap-3">
            <button
              onClick={save}
              disabled={busy}
              className="py-2 px-5 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
            >
              {busy ? 'Saving…' : `Save ${noun}`}
            </button>
            {note && <span className="text-bone/60 text-sm">{note}</span>}
          </div>
        </div>
      )}

      {section === 'quiz' && !m.isSubplot && <SuperAdminQuiz mysteryId={m.id} />}
      {section === 'secrets' && <CharacterContentEditor mysteryId={m.id} beats={m.beats} />}
    </div>
  );
};

// A beat's "who knows this" panel — the beat-centric complement of
// CharacterContentEditor/ContentForCharacter below. Lets a super user assign
// secrets to characters right where a beat is authored, instead of having
// to leave the Story tab and walk the cast one character at a time. Writes
// happen immediately per-secret (not deferred to "Save"), same posture as
// the character-centric editor's own save button.
const BeatSecretsPanel = ({ mysteryId, beat }: { mysteryId: string; beat: AdminMysteryBeat }) => {
  const [characters, setCharacters] = useState<AdminCharacter[]>([]);
  const [secrets, setSecrets] = useState<AdminBeatSecret[] | null>(null);
  const [newCharacterId, setNewCharacterId] = useState('');
  const [newBody, setNewBody] = useState('');
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState<string | null>(null);

  const load = () => {
    adminListBeatSecrets(mysteryId, beat.id)
      .then((d) => setSecrets(d.secrets))
      .catch(() => setNote('Could not load secrets for this beat.'));
  };

  useEffect(() => {
    adminListCharacters()
      .then((d) => setCharacters(d.characters.filter((c) => c.roleType === 'player')))
      .catch(() => {});
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [mysteryId, beat.id]);

  const characterName = (id: string) => characters.find((c) => c.id === id)?.name || '(unknown character)';

  const addSecret = async () => {
    if (!newCharacterId || !newBody.trim()) return;
    setBusy(true);
    setNote(null);
    try {
      await adminCreateBeatSecret(mysteryId, beat.id, newCharacterId, newBody.trim());
      setNewBody('');
      load();
    } catch (e) {
      setNote(e instanceof Error ? e.message : 'Could not add secret.');
    } finally {
      setBusy(false);
    }
  };

  const removeSecret = async (id: string) => {
    setBusy(true);
    setNote(null);
    try {
      await adminDeleteSecret(id);
      load();
    } catch (e) {
      setNote(e instanceof Error ? e.message : 'Could not remove secret.');
    } finally {
      setBusy(false);
    }
  };

  const input = 'w-full rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm';

  return (
    <div className="ml-6 mt-1 rounded-md border border-gold/25 bg-black/20 p-3 flex flex-col gap-2">
      <p className="text-[11px] uppercase tracking-[0.2em] text-gold/70">Who knows this</p>
      {secrets === null ? (
        <p className="text-bone/50 text-xs">{note || 'Loading…'}</p>
      ) : (
        <>
          {secrets.length === 0 && <p className="text-bone/40 text-xs">No one yet.</p>}
          {secrets.map((s) => (
            <div key={s.id} className="flex gap-2 items-start">
              <div className="flex-1">
                <p className="text-gold/80 text-xs mb-0.5">{characterName(s.characterId)}</p>
                <BeatSecretBodyEditor secret={s} onSaved={load} />
              </div>
              <RemoveBtn onClick={() => removeSecret(s.id)} />
            </div>
          ))}
        </>
      )}
      <div className="flex gap-2 items-start pt-2 border-t border-gold/10">
        <select
          className={`${input} w-36 shrink-0`}
          value={newCharacterId}
          onChange={(e) => setNewCharacterId(e.target.value)}
        >
          <option value="">— character —</option>
          {characters.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name}
            </option>
          ))}
        </select>
        <textarea
          className={input}
          rows={2}
          placeholder="What they know…"
          value={newBody}
          onChange={(e) => setNewBody(e.target.value)}
        />
        <button
          onClick={addSecret}
          disabled={busy || !newCharacterId || !newBody.trim()}
          className="shrink-0 px-3 py-2 rounded-md border border-gold/50 text-gold text-xs uppercase tracking-[0.15em] disabled:opacity-40"
        >
          + Add
        </button>
      </div>
      {note && <p className="text-blood-bright text-xs">{note}</p>}
    </div>
  );
};

// Auto-saves on blur (rather than a per-row Save button) since this list has
// no batch "Save" step of its own — each secret is its own row-level write.
const BeatSecretBodyEditor = ({ secret, onSaved }: { secret: AdminBeatSecret; onSaved: () => void }) => {
  const [body, setBody] = useState(secret.body);
  const [saving, setSaving] = useState(false);

  const commit = async () => {
    const trimmed = body.trim();
    if (!trimmed || trimmed === secret.body.trim()) {
      setBody(secret.body);
      return;
    }
    setSaving(true);
    try {
      await adminUpdateSecretBody(secret.id, trimmed);
      onSaved();
    } finally {
      setSaving(false);
    }
  };

  return (
    <textarea
      className="w-full rounded-md bg-black/60 border border-blood/30 p-1.5 text-bone text-xs disabled:opacity-60"
      rows={2}
      value={body}
      onChange={(e) => setBody(e.target.value)}
      onBlur={commit}
      disabled={saving}
    />
  );
};

// Walk the cast and decide what each of them knows, what they need to do,
// and what happens to them — for this mystery. Deliberately reached from
// the mystery, not the character (see MYSTERY_REQUIREMENTS.md's "Super
// Admin UI"). A character with zero secrets here can't be invited to a
// Toast running this mystery (missions don't gate invites the same way).
const CharacterContentEditor = ({ mysteryId, beats }: { mysteryId: string; beats: AdminMysteryFull['beats'] }) => {
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

const blankMission = (): AdminMysteryMission => ({ ordinal: 0, tier: 'easy', rewardBt: 2, prompt: '', answerFormat: '' });

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
  const [missions, setMissions] = useState<AdminMysteryMission[]>([]);
  const [postAct1Context, setPostAct1Context] = useState('');
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState<string | null>(null);

  useEffect(() => {
    setSecrets(null);
    setMissions([]);
    setPostAct1Context('');
    adminGetCharacterContentForMystery(mysteryId, character.id)
      .then((d) => {
        setSecrets(d.secrets);
        setMissions(d.missions);
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
        missions: missions.map((m) => ({
          tier: m.tier,
          rewardBt: m.rewardBt,
          prompt: m.prompt,
          answerFormat: m.answerFormat,
        })),
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
                    {b.title || b.description.slice(0, 60) || '(untitled beat)'}
                  </option>
                ))}
              </select>
            </div>
            <RemoveBtn onClick={() => setSecrets(secrets.filter((_, j) => j !== i))} />
          </div>
        ))}
      </ListEditor>

      <ListEditor
        label="Missions"
        addLabel="+ Add mission"
        onAdd={() => setMissions([...missions, { ...blankMission(), ordinal: missions.length + 1 }])}
      >
        {missions.map((mn, i) => (
          <div key={i} className="rounded-md border border-blood/30 p-2 flex flex-col gap-2">
            <div className="flex items-center gap-2">
              <span className="text-gold text-xs w-4">{i + 1}</span>
              <select
                className={`${input} w-28`}
                value={mn.tier}
                onChange={(e) => setMissions(missions.map((x, j) => (j === i ? { ...x, tier: e.target.value } : x)))}
              >
                <option value="easy">easy</option>
                <option value="medium">medium</option>
                <option value="hard">hard</option>
              </select>
              <input
                type="number"
                className={`${input} w-20`}
                value={mn.rewardBt}
                onChange={(e) =>
                  setMissions(missions.map((x, j) => (j === i ? { ...x, rewardBt: Number(e.target.value) || 0 } : x)))
                }
              />
              <span className="text-bone/40 text-xs">BT</span>
              <RemoveBtn onClick={() => setMissions(missions.filter((_, j) => j !== i))} />
            </div>
            <textarea
              className={input}
              rows={2}
              placeholder="Mission prompt"
              value={mn.prompt}
              onChange={(e) => setMissions(missions.map((x, j) => (j === i ? { ...x, prompt: e.target.value } : x)))}
            />
            <input
              className={input}
              placeholder="What to submit (answer format)"
              value={mn.answerFormat}
              onChange={(e) =>
                setMissions(missions.map((x, j) => (j === i ? { ...x, answerFormat: e.target.value } : x)))
              }
            />
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
