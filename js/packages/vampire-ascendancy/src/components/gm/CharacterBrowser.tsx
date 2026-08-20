import { useMemo, useState } from 'react';

interface CharacterLike {
  id: string;
  name: string;
  title: string;
  house?: string;
  preEventInfo?: string;
  tags?: string[];
}

// Search by name/bio, filter by trait tags (musical, gambler, risk taker,
// …), and a fuller preview per row (title, house, bio snippet, tags) than
// a bare name in a dropdown ever gave. Tag filtering is "any of" — picking
// multiple tags broadens the list, it doesn't require every tag on one
// character. Used for the Character Pool tab's curation picker and a
// player's self-select at RSVP time (both multi/single via selectedIds),
// and the Mysteries tab's per-character secrets editor (selectedId).
export function CharacterBrowser<T extends CharacterLike>({
  characters,
  selectedId,
  selectedIds,
  onSelect,
  emptyMessage = 'No characters match.',
}: {
  characters: T[];
  // Single-select mode: which one row is highlighted.
  selectedId?: string;
  // Multi-select mode: which rows are highlighted/checkmarked. onSelect is
  // still called with the clicked character either way — the caller
  // decides whether that means "set" or "toggle in a set".
  selectedIds?: Set<string>;
  onSelect: (character: T) => void;
  emptyMessage?: string;
}) {
  const [query, setQuery] = useState('');
  const [activeTags, setActiveTags] = useState<Set<string>>(new Set());

  const allTags = useMemo(
    () => [...new Set(characters.flatMap((c) => c.tags ?? []))].sort((a, b) => a.localeCompare(b)),
    [characters]
  );

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return characters.filter((c) => {
      if (activeTags.size > 0 && !(c.tags ?? []).some((t) => activeTags.has(t))) return false;
      if (!q) return true;
      return c.name.toLowerCase().includes(q) || (c.preEventInfo || '').toLowerCase().includes(q);
    });
  }, [characters, query, activeTags]);

  const toggleTag = (tag: string) => {
    setActiveTags((prev) => {
      const next = new Set(prev);
      if (next.has(tag)) next.delete(tag);
      else next.add(tag);
      return next;
    });
  };

  return (
    <div className="flex flex-col gap-2">
      <input
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search name or bio…"
        className="w-full rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm"
      />

      {allTags.length > 0 && (
        <div className="flex gap-1.5 flex-wrap">
          {allTags.map((tag) => (
            <button
              key={tag}
              type="button"
              onClick={() => toggleTag(tag)}
              className={`px-2 py-0.5 rounded-full text-[10px] uppercase tracking-[0.15em] border transition-colors ${
                activeTags.has(tag)
                  ? 'bg-gold text-blood-ink border-gold'
                  : 'text-gold/70 border-gold/30 hover:border-gold/60'
              }`}
            >
              {tag}
            </button>
          ))}
          {activeTags.size > 0 && (
            <button
              type="button"
              onClick={() => setActiveTags(new Set())}
              className="px-2 py-0.5 text-[10px] uppercase tracking-[0.15em] text-bone/40 hover:text-bone/70"
            >
              Clear
            </button>
          )}
        </div>
      )}

      <div className="flex flex-col gap-1.5 max-h-96 overflow-y-auto rounded-md border border-blood/20 p-1.5">
        {filtered.length === 0 ? (
          <p className="text-bone/50 text-sm p-2">{characters.length === 0 ? 'No characters yet.' : emptyMessage}</p>
        ) : (
          filtered.map((c) => {
            const selected = selectedId === c.id || (selectedIds?.has(c.id) ?? false);
            return (
              <div
                key={c.id}
                onClick={() => onSelect(c)}
                className={`rounded-md border p-2.5 cursor-pointer transition-colors ${
                  selected ? 'border-blood-bright bg-blood/20' : 'border-blood/20 bg-black/30 hover:border-blood/40'
                }`}
              >
                <p className="text-bone text-sm font-semibold flex items-center flex-wrap gap-1.5">
                  {selected && <span className="text-gold">✓</span>}
                  {c.name}
                  {c.house && (
                    <span className="text-[10px] uppercase tracking-[0.15em] rounded-full border border-blood/40 px-1.5 py-0.5 text-bone/60 font-normal">
                      {c.house}
                    </span>
                  )}
                </p>
                {c.title && <p className="text-xs text-bone/60 italic">{c.title}</p>}
                {c.preEventInfo && <p className="text-xs text-bone/60 mt-0.5 line-clamp-2">{c.preEventInfo}</p>}
                {c.tags && c.tags.length > 0 && (
                  <div className="flex gap-1 flex-wrap mt-1">
                    {c.tags.map((t) => (
                      <span
                        key={t}
                        className="text-[10px] uppercase tracking-[0.15em] rounded-full border border-gold/30 px-1.5 py-0.5 text-gold/80"
                      >
                        {t}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
