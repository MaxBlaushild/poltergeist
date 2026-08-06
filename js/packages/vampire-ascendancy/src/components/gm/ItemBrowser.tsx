import { useMemo, useState } from 'react';
import { itemPhotoUrl } from '../../gmApi';
import { effectTag, filterAndSortItems, type SortMode } from '../../itemDisplay';

interface ItemLike {
  id: string;
  name: string;
  category: string;
  description: string;
  effect: string;
  targetsPlayer: boolean;
  hasPhoto: boolean;
  hfEffect: number;
  btSelf: number;
  btFromTarget: number;
  btDeductTarget: number;
  quizBtPct: number;
  doubleGameBt: boolean;
  immune: boolean;
  reflect: boolean;
  stripResistance: boolean;
}

// The shared item list + search/category/targeted/sort toolbar behind both
// the Items tab's "Assign an item" picker (click a row to select) and the
// Content tab's include/exclude picker (a checkbox per row via
// `renderTrailing`) — same card, same filters, same info, different action.
export function ItemBrowser<T extends ItemLike>({
  items,
  selectedId,
  onSelect,
  renderTrailing,
  emptyMessage = 'No items match.',
}: {
  items: T[];
  // Highlights this row and (if onSelect is set) makes the whole row
  // clickable — the Items tab's "pick one to assign" mode.
  selectedId?: string;
  onSelect?: (item: T) => void;
  // An extra control at the end of each row — the Content tab's "Included"
  // checkbox. Independent of onSelect/selectedId.
  renderTrailing?: (item: T) => React.ReactNode;
  emptyMessage?: string;
}) {
  const [query, setQuery] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [targetedOnly, setTargetedOnly] = useState(false);
  const [sortMode, setSortMode] = useState<SortMode>('name');

  const categories = useMemo(
    () => [...new Set(items.map((i) => i.category).filter(Boolean))].sort((a, b) => a.localeCompare(b)),
    [items]
  );

  const filtered = useMemo(
    () => filterAndSortItems(items, { query, category: categoryFilter, targetedOnly, sortMode }),
    [items, query, categoryFilter, targetedOnly, sortMode]
  );

  return (
    <div className="flex flex-col gap-2">
      <div className="flex gap-2 flex-wrap">
        <input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search name, category, description…"
          className="flex-1 min-w-[160px] rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm"
        />
        <select
          value={categoryFilter}
          onChange={(e) => setCategoryFilter(e.target.value)}
          className="rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm"
        >
          <option value="">All categories</option>
          {categories.map((c) => (
            <option key={c} value={c}>
              {c}
            </option>
          ))}
        </select>
        <select
          value={sortMode}
          onChange={(e) => setSortMode(e.target.value as SortMode)}
          className="rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm"
        >
          <option value="name">Sort: Name</option>
          <option value="category">Sort: Category</option>
          <option value="effect">Sort: Strongest effect</option>
        </select>
        <label className="flex items-center gap-1.5 text-xs text-bone/70 px-1">
          <input type="checkbox" checked={targetedOnly} onChange={(e) => setTargetedOnly(e.target.checked)} />
          Targets a player
        </label>
      </div>

      <div className="flex flex-col gap-1.5 max-h-96 overflow-y-auto rounded-md border border-blood/20 p-1.5">
        {filtered.length === 0 ? (
          <p className="text-bone/50 text-sm p-2">{items.length === 0 ? 'No items yet.' : emptyMessage}</p>
        ) : (
          filtered.map((it) => {
            const selected = selectedId === it.id;
            return (
              <div
                key={it.id}
                onClick={onSelect ? () => onSelect(it) : undefined}
                className={`flex items-start gap-2.5 rounded-md border p-2 transition-colors ${
                  onSelect ? 'cursor-pointer' : ''
                } ${selected ? 'border-blood-bright bg-blood/20' : 'border-blood/20 bg-black/30 hover:border-blood/40'}`}
              >
                {it.hasPhoto ? (
                  <img
                    src={itemPhotoUrl(it.id)}
                    alt={it.name}
                    className="w-12 h-12 rounded object-cover border border-blood/40 shrink-0"
                  />
                ) : (
                  <div className="w-12 h-12 rounded border border-blood/30 bg-black/40 shrink-0 flex items-center justify-center text-bone/30 text-[10px] uppercase tracking-wide">
                    No photo
                  </div>
                )}
                <div className="flex-1 min-w-0">
                  <p className="text-bone text-sm font-semibold flex items-center flex-wrap gap-1.5">
                    {it.name}
                    {it.category && (
                      <span className="text-[10px] uppercase tracking-[0.15em] rounded-full border border-blood/40 px-1.5 py-0.5 text-bone/60 font-normal">
                        {it.category}
                      </span>
                    )}
                    {it.targetsPlayer && (
                      <span className="text-[10px] uppercase tracking-[0.15em] text-blood-bright font-normal">targeted</span>
                    )}
                  </p>
                  {it.description && <p className="text-xs text-bone/60 mt-0.5 line-clamp-2">{it.description}</p>}
                  {it.effect && <p className="text-xs text-gold/90 italic mt-0.5 line-clamp-1">{it.effect}</p>}
                  {effectTag(it) && <p className="text-xs text-gold/70 mt-0.5">{effectTag(it)}</p>}
                </div>
                {renderTrailing && <div className="shrink-0 self-center">{renderTrailing(it)}</div>}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
