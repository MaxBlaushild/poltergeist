import { useEffect, useMemo, useState } from 'react';
import {
  adminListAllSecrets,
  adminListCharacters,
  adminListMysteryBeatOptions,
  adminCreateSecret,
  adminUpdateSecretBody,
  adminDeleteSecret,
} from '../../superAdminApi';
import type { AdminSecretRow, AdminCharacter, AdminMysteryBeatOption } from '../../superAdminApi';
import { ApiError } from '../../api';
import { Card } from '../gm/GameSection';
import { Field, RemoveBtn } from './SuperAdminShared';

type SortKey = 'character' | 'mystery' | 'beat';

// A flat, cross-cutting view of every secret in the system — every other
// secrets editor (the Mysteries tab's per-beat panel, the Mysteries tab's
// per-character content editor, the Characters tab's per-mystery editor) is
// scoped to one mystery and/or one character at a time; this is the one
// place to see, sort, filter, and author them all at once. All four write
// through the same underlying secret, so edits made anywhere stay in sync.
export const SuperAdminSecrets = () => {
  const [secrets, setSecrets] = useState<AdminSecretRow[] | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [query, setQuery] = useState('');
  const [mysteryFilter, setMysteryFilter] = useState('');
  const [missingBeatOnly, setMissingBeatOnly] = useState(false);
  const [sortKey, setSortKey] = useState<SortKey>('mystery');
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('asc');

  const reload = () =>
    adminListAllSecrets()
      .then((d) => setSecrets(d.secrets))
      .catch((e) => setLoadError(e instanceof ApiError ? e.message : 'Could not load secrets.'));

  useEffect(() => {
    reload();
  }, []);

  const mysteries = useMemo(() => {
    if (!secrets) return [];
    const byId = new Map<string, { id: string; name: string; isSubplot: boolean }>();
    for (const s of secrets) {
      if (!byId.has(s.mysteryId)) byId.set(s.mysteryId, { id: s.mysteryId, name: s.mysteryName, isSubplot: s.isSubplot });
    }
    return [...byId.values()].sort((a, b) => a.name.localeCompare(b.name));
  }, [secrets]);

  const toggleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setSortDir('asc');
    }
  };

  const visible = useMemo(() => {
    if (!secrets) return [];
    const q = query.trim().toLowerCase();
    let rows = secrets.filter((s) => {
      if (mysteryFilter && s.mysteryId !== mysteryFilter) return false;
      if (missingBeatOnly && s.beatId) return false;
      if (!q) return true;
      return (
        s.characterName.toLowerCase().includes(q) ||
        s.mysteryName.toLowerCase().includes(q) ||
        (s.beatTitle ?? '').toLowerCase().includes(q) ||
        s.body.toLowerCase().includes(q)
      );
    });
    const dir = sortDir === 'asc' ? 1 : -1;
    const keyed = (s: AdminSecretRow) =>
      sortKey === 'character' ? s.characterName : sortKey === 'mystery' ? s.mysteryName : s.beatTitle ?? '';
    rows = [...rows].sort((a, b) => keyed(a).localeCompare(keyed(b)) * dir);
    return rows;
  }, [secrets, query, mysteryFilter, missingBeatOnly, sortKey, sortDir]);

  if (!secrets) return <Card title="Secrets">{loadError || 'Loading…'}</Card>;

  const sortBtn = (key: SortKey, label: string) => (
    <button
      type="button"
      onClick={() => toggleSort(key)}
      className={`px-3 py-1.5 rounded-md text-xs uppercase tracking-[0.15em] border ${
        sortKey === key ? 'border-gold text-gold' : 'border-blood/40 text-bone/60'
      }`}
    >
      {label}
      {sortKey === key && <span className="ml-1">{sortDir === 'asc' ? '↑' : '↓'}</span>}
    </button>
  );

  return (
    <Card title={`Secrets (${secrets.length}) — every secret in the system, across every character and mystery`}>
      <CreateSecretForm onCreated={reload} />

      <div className="mt-5 mb-3 flex flex-col gap-2">
        <input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search by character, mystery, beat, or text…"
          className="w-full rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm"
        />
        <div className="flex items-center gap-2 flex-wrap">
          <select
            value={mysteryFilter}
            onChange={(e) => setMysteryFilter(e.target.value)}
            className="rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm"
          >
            <option value="">— every mystery / sub-plot —</option>
            {mysteries.map((m) => (
              <option key={m.id} value={m.id}>
                {m.name}
                {m.isSubplot ? ' (sub-plot)' : ''}
              </option>
            ))}
          </select>
          <label className="flex items-center gap-1.5 text-xs text-bone/60">
            <input
              type="checkbox"
              checked={missingBeatOnly}
              onChange={(e) => setMissingBeatOnly(e.target.checked)}
            />
            No beat only
          </label>
          <div className="flex items-center gap-1.5 ml-auto">
            <span className="text-[11px] uppercase tracking-[0.15em] text-bone/40">Sort</span>
            {sortBtn('character', 'Character')}
            {sortBtn('mystery', 'Mystery')}
            {sortBtn('beat', 'Beat')}
          </div>
        </div>
      </div>

      <p className="text-bone/40 text-xs mb-2">
        {visible.length} of {secrets.length} shown
      </p>

      <div className="flex flex-col gap-2">
        {visible.map((s) => (
          <SecretRowItem key={s.id} secret={s} onChanged={reload} />
        ))}
        {visible.length === 0 && <p className="text-bone/50 text-sm">No secrets match.</p>}
      </div>
    </Card>
  );
};

const CreateSecretForm = ({ onCreated }: { onCreated: () => void }) => {
  const [characters, setCharacters] = useState<AdminCharacter[]>([]);
  const [options, setOptions] = useState<AdminMysteryBeatOption[]>([]);
  const [characterId, setCharacterId] = useState('');
  // Encodes both which mystery/sub-plot AND which beat as one composite
  // value, since a beat can be shared across several mysteries — picking
  // just a beat id would be ambiguous about which story the secret is for.
  const [beatKey, setBeatKey] = useState('');
  const [body, setBody] = useState('');
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([adminListCharacters(), adminListMysteryBeatOptions()])
      .then(([c, o]) => {
        setCharacters(c.characters.filter((ch) => ch.roleType === 'player'));
        setOptions(o.options);
      })
      .catch(() => {});
  }, []);

  const input = 'w-full rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm';

  const create = async () => {
    if (!characterId || !beatKey || !body.trim()) {
      setNote('Choose a character, a beat, and enter the secret text.');
      return;
    }
    const [mysteryId, beatId] = beatKey.split('::');
    setBusy(true);
    setNote(null);
    try {
      await adminCreateSecret({ characterId, mysteryId, beatId, body: body.trim() });
      setBody('');
      setNote('Created.');
      onCreated();
    } catch (e) {
      setNote(e instanceof ApiError ? e.message : 'Could not create secret.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="rounded-md border border-gold/25 bg-black/20 p-3">
      <p className="text-[11px] uppercase tracking-[0.15em] text-bone/50 mb-2">New secret</p>
      <div className="flex flex-col gap-1.5">
        <div className="grid grid-cols-2 gap-2">
          <Field label="Character">
            <select className={input} value={characterId} onChange={(e) => setCharacterId(e.target.value)}>
              <option value="">— choose a character —</option>
              {[...characters]
                .sort((a, b) => a.name.localeCompare(b.name))
                .map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
            </select>
          </Field>
          <Field label="Mystery / sub-plot + beat">
            <select className={input} value={beatKey} onChange={(e) => setBeatKey(e.target.value)}>
              <option value="">— choose a beat —</option>
              {options.map((o) => (
                <option key={`${o.mysteryId}::${o.beatId}`} value={`${o.mysteryId}::${o.beatId}`}>
                  {o.mysteryName}
                  {o.isSubplot ? ' (sub-plot)' : ''} — {o.beatTitle || '(untitled beat)'}
                </option>
              ))}
            </select>
          </Field>
        </div>
        <textarea
          className={input}
          rows={2}
          placeholder="What does this character know?"
          value={body}
          onChange={(e) => setBody(e.target.value)}
        />
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={create}
            disabled={busy}
            className="py-2 px-5 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
          >
            {busy ? 'Creating…' : 'Add secret'}
          </button>
          {note && <span className="text-bone/60 text-sm">{note}</span>}
        </div>
      </div>
    </div>
  );
};

const SecretRowItem = ({ secret, onChanged }: { secret: AdminSecretRow; onChanged: () => void }) => {
  const [body, setBody] = useState(secret.body);
  const [busy, setBusy] = useState(false);

  useEffect(() => setBody(secret.body), [secret.body]);

  const save = async () => {
    if (body === secret.body) return;
    setBusy(true);
    try {
      await adminUpdateSecretBody(secret.id, body);
      onChanged();
    } catch {
      setBody(secret.body);
    } finally {
      setBusy(false);
    }
  };

  const remove = async () => {
    setBusy(true);
    try {
      await adminDeleteSecret(secret.id);
      onChanged();
    } catch {
      setBusy(false);
    }
  };

  return (
    <div className="border-b border-blood/15 last:border-0 pb-2">
      <p className="text-xs text-bone/60 mb-1">
        <span className="text-bone">{secret.characterName}</span>
        <span className="text-bone/40 mx-1.5">•</span>
        {secret.mysteryName}
        {secret.isSubplot && <span className="text-gold/70 ml-1">(sub-plot)</span>}
        <span className="text-bone/40 mx-1.5">•</span>
        {secret.beatTitle ? (
          secret.beatTitle
        ) : (
          <span className="text-blood-bright">(no beat)</span>
        )}
      </p>
      <div className="flex gap-2 items-start">
        <textarea
          className="flex-1 rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm"
          rows={2}
          value={body}
          onChange={(e) => setBody(e.target.value)}
          onBlur={save}
          disabled={busy}
        />
        <RemoveBtn onClick={remove} />
      </div>
    </div>
  );
};
