import { useEffect, useMemo, useState } from 'react';
import { gmListCharacters, gmGetCharacterPool, gmSetCharacterPool, gmGetBeatCoverage } from '../../gmApi';
import type { GMCharacter, GMBeatCoverage } from '../../gmApi';
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
  const [coverage, setCoverage] = useState<GMBeatCoverage | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [note, setNote] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([gmListCharacters(), gmGetCharacterPool(), gmGetBeatCoverage()])
      .then(([c, p, cov]) => {
        setCharacters(c.characters.filter((ch) => ch.roleType === 'player'));
        setSelected(new Set(p.characterIds));
        setSaved(new Set(p.characterIds));
        setCoverage(cov);
      })
      .catch(() => setNote('Could not load the character pool.'))
      .finally(() => setLoading(false));
  }, []);

  // How many secrets among the currently-checked characters (not yet
  // saved — this recomputes live as the Host toggles) touch each beat of
  // the mystery/subplots, so gaps in coverage are visible while deciding,
  // not just after.
  const beatCounts = useMemo(() => {
    const counts = new Map<string, number>();
    if (!coverage) return counts;
    for (const link of coverage.secretBeatLinks) {
      if (!selected.has(link.characterId)) continue;
      counts.set(link.beatId, (counts.get(link.beatId) ?? 0) + 1);
    }
    return counts;
  }, [coverage, selected]);

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

      {coverage && coverage.beats.length > 0 && (
        <div className="mb-4">
          <p className="text-[11px] uppercase tracking-[0.15em] text-bone/50 mb-1.5">
            Story coverage — secrets among your current selection, per beat
          </p>
          <div className="flex flex-wrap gap-1.5">
            {coverage.beats.map((b) => {
              const count = beatCounts.get(b.id) ?? 0;
              return (
                <span
                  key={b.id}
                  className={`inline-flex items-center gap-1.5 pl-2.5 pr-1.5 py-1 rounded-full text-[11px] border ${
                    count === 0
                      ? 'border-blood-bright/50 text-blood-bright/90 bg-blood/10'
                      : 'border-gold/40 text-gold/90 bg-gold/10'
                  }`}
                >
                  {b.title || '(untitled beat)'}
                  <span
                    className={`rounded-full px-1.5 font-semibold ${
                      count === 0 ? 'bg-blood-bright/20' : 'bg-gold/20'
                    }`}
                  >
                    {count}
                  </span>
                </span>
              );
            })}
          </div>
          <p className="text-[11px] text-bone/40 mt-1.5">
            A beat at 0 isn't known by anyone in your current selection yet.
          </p>
        </div>
      )}

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
