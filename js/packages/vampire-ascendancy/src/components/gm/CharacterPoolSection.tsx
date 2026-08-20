import { useEffect, useState } from 'react';
import { gmListCharacters, gmGetCharacterPool, gmSetCharacterPool } from '../../gmApi';
import type { GMCharacter } from '../../gmApi';
import { ApiError } from '../../api';
import { Card } from './GameSection';
import { CharacterBrowser } from './CharacterBrowser';

// Character Pool tab: which of this Toast's mystery-eligible characters
// (the same set the old Invites picker used to offer) a player can choose
// from when they self-select their character after accepting an invite.
// Invites no longer name a character — this is where that "who's
// available" decision lives instead, curated once for the whole Toast
// rather than per-invite.
export const CharacterPoolSection = () => {
  const [characters, setCharacters] = useState<GMCharacter[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [saved, setSaved] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([gmListCharacters(), gmGetCharacterPool()])
      .then(([c, p]) => {
        setCharacters(c.characters.filter((ch) => ch.roleType === 'player'));
        setSelected(new Set(p.characterIds));
        setSaved(new Set(p.characterIds));
      })
      .catch(() => setNote('Could not load the character pool.'))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <Card title="Character Pool">Loading…</Card>;

  const toggle = (c: GMCharacter) =>
    setSelected((cur) => {
      const next = new Set(cur);
      if (next.has(c.id)) next.delete(c.id);
      else next.add(c.id);
      return next;
    });

  const dirty = selected.size !== saved.size || [...selected].some((id) => !saved.has(id));

  const save = async () => {
    setBusy(true);
    setNote(null);
    try {
      await gmSetCharacterPool([...selected]);
      setSaved(new Set(selected));
      setNote('Saved.');
    } catch (e) {
      setNote(e instanceof ApiError ? e.message : 'Could not save the pool.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <Card title={`Character Pool — ${selected.size} of ${characters.length} selectable`}>
      <p className="text-bone/60 text-sm mb-3">
        Choose which characters players can pick for themselves after accepting an invite. Only
        characters with secrets written for this Toast's mystery are eligible at all — the same set
        that used to populate the old "who is this invite for" picker.
      </p>
      <div className="flex gap-3 mb-2">
        <button
          type="button"
          onClick={() => setSelected(new Set(characters.map((c) => c.id)))}
          className="text-xs text-gold uppercase tracking-[0.15em]"
        >
          Select all
        </button>
        <button
          type="button"
          onClick={() => setSelected(new Set())}
          className="text-xs text-bone/50 uppercase tracking-[0.15em]"
        >
          Clear all
        </button>
      </div>
      <CharacterBrowser
        characters={characters}
        selectedIds={selected}
        onSelect={toggle}
        emptyMessage="No characters match."
      />
      {characters.length === 0 && (
        <p className="text-gold/80 text-xs mt-2">
          No characters are eligible yet — write secrets for this mystery's characters from the Super
          Admin dashboard's Mysteries tab.
        </p>
      )}
      <div className="mt-3 flex items-center gap-3">
        <button
          onClick={save}
          disabled={busy || !dirty}
          className="py-2 px-5 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
        >
          {busy ? 'Saving…' : 'Save pool'}
        </button>
        {note && <span className="text-bone/60 text-sm">{note}</span>}
      </div>
    </Card>
  );
};
