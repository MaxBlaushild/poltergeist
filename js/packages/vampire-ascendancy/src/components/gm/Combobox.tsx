import { useMemo, useState } from 'react';

export type ComboOption = { id: string; label: string; sub?: string };

// Search-and-select with removable pills. Multi-select: typing filters the list,
// clicking adds a pill, the × removes it. Options already chosen are hidden.
export const Combobox = ({
  options,
  selected,
  onChange,
  placeholder = 'Search…',
  max = 8,
}: {
  options: ComboOption[];
  selected: string[];
  onChange: (ids: string[]) => void;
  placeholder?: string;
  max?: number;
}) => {
  const [query, setQuery] = useState('');
  const [open, setOpen] = useState(false);

  const byId = useMemo(() => Object.fromEntries(options.map((o) => [o.id, o])), [options]);
  const selectedSet = new Set(selected);
  const q = query.trim().toLowerCase();
  const filtered = options
    .filter((o) => !selectedSet.has(o.id))
    .filter((o) => !q || `${o.label} ${o.sub ?? ''}`.toLowerCase().includes(q))
    .slice(0, max);

  const add = (id: string) => {
    onChange([...selected, id]);
    setQuery('');
  };
  const remove = (id: string) => onChange(selected.filter((s) => s !== id));

  const onKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && q && filtered.length) {
      // Enter adds the top match.
      e.preventDefault();
      add(filtered[0].id);
    } else if (e.key === 'Backspace' && query === '' && selected.length) {
      // Backspace on an empty query removes the last pill.
      e.preventDefault();
      remove(selected[selected.length - 1]);
    } else if (e.key === 'Escape') {
      setOpen(false);
    }
  };

  return (
    <div className="relative">
      <div className="flex flex-wrap items-center gap-1.5 rounded-md bg-black/60 border border-blood/40 p-2 min-h-[42px]">
        {selected.map((id) => (
          <span
            key={id}
            className="inline-flex items-center gap-1 rounded-full bg-blood/30 border border-blood/50 pl-2.5 pr-1 py-0.5 text-xs text-bone"
          >
            {byId[id]?.label ?? id}
            <button
              type="button"
              onClick={() => remove(id)}
              aria-label={`Remove ${byId[id]?.label ?? id}`}
              className="w-4 h-4 leading-none rounded-full text-blood-bright hover:text-bone hover:bg-blood/50"
            >
              ×
            </button>
          </span>
        ))}
        <input
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setOpen(true);
          }}
          onFocus={() => setOpen(true)}
          onBlur={() => setTimeout(() => setOpen(false), 120)}
          onKeyDown={onKeyDown}
          placeholder={selected.length ? '' : placeholder}
          className="flex-1 min-w-[90px] bg-transparent text-bone text-sm outline-none placeholder:text-bone/40"
        />
      </div>
      {open && filtered.length > 0 && (
        <div className="absolute z-30 mt-1 w-full max-h-56 overflow-y-auto rounded-md border border-blood/40 bg-blood-ink shadow-xl">
          {filtered.map((o, i) => (
            <button
              key={o.id}
              type="button"
              // onMouseDown (not onClick) so the pick lands before the input blurs.
              onMouseDown={(e) => {
                e.preventDefault();
                add(o.id);
              }}
              className={`w-full text-left px-3 py-2 text-sm text-bone/90 hover:bg-white/5 ${
                // Highlight the top match — the one Enter will add.
                i === 0 && q ? 'bg-white/5' : ''
              }`}
            >
              {o.label}
              {o.sub ? <span className="text-bone/40"> · {o.sub}</span> : null}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};
