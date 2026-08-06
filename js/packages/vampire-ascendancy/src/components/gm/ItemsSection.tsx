import { useEffect, useMemo, useState } from 'react';
import {
  gmListItems,
  gmListPlayers,
  gmListPlayerItems,
  gmAssignItem,
  gmRemovePlayerItem,
  gmTransferPlayerItem,
  itemPhotoUrl,
} from '../../gmApi';
import type { GMItem, GMPlayer, GMPlayerItem } from '../../gmApi';
import { Card } from './GameSection';

// A thumbnail that opens a full-size modal on click.
const PhotoThumb = ({ src, alt, className }: { src: string; alt: string; className?: string }) => {
  const [open, setOpen] = useState(false);
  return (
    <>
      <img
        src={src}
        alt={alt}
        onClick={() => setOpen(true)}
        className={`cursor-zoom-in ${className ?? ''}`}
      />
      {open && (
        <div
          className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-black/90 p-6"
          onClick={() => setOpen(false)}
          role="dialog"
          aria-label={`${alt} photo`}
        >
          <img
            src={src}
            alt={alt}
            className="max-h-[85vh] max-w-[92vw] rounded-lg border-2 border-blood/40 object-contain"
          />
          <p className="mt-3 font-display text-xl text-bone">{alt}</p>
        </div>
      )}
    </>
  );
};

type SortMode = 'name' | 'category' | 'effect';

// GM Items tab: this Toast's included items — assign them to players and
// review/transfer/remove holdings. The catalog itself (creating/editing/
// deleting items, reference photos) is shared content, edited only by super
// users — see the Super Admin dashboard.
export const ItemsSection = () => {
  const [items, setItems] = useState<GMItem[]>([]);
  const [players, setPlayers] = useState<GMPlayer[]>([]);
  const [held, setHeld] = useState<GMPlayerItem[]>([]);
  const [playerId, setPlayerId] = useState('');
  const [itemId, setItemId] = useState('');
  const [query, setQuery] = useState('');
  const [categoryFilter, setCategoryFilter] = useState('');
  const [targetedOnly, setTargetedOnly] = useState(false);
  const [sortMode, setSortMode] = useState<SortMode>('name');
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);

  const load = () => {
    Promise.all([gmListItems(), gmListPlayers(), gmListPlayerItems()])
      .then(([i, p, h]) => {
        setItems(i.items);
        setPlayers(p.players);
        setHeld(h.playerItems);
      })
      .finally(() => setLoading(false));
  };
  useEffect(() => {
    load();
  }, []);

  // Only assign to slots that have a character.
  const assignablePlayers = useMemo(
    () =>
      players
        .filter((p) => p.character)
        .sort((a, b) => (a.character?.name ?? '').localeCompare(b.character?.name ?? '')),
    [players]
  );

  const categories = useMemo(
    () => [...new Set(items.map((i) => i.category).filter(Boolean))].sort((a, b) => a.localeCompare(b)),
    [items]
  );

  const sortedItems = useMemo(() => {
    const q = query.trim().toLowerCase();
    const filtered = items.filter((i) => {
      if (categoryFilter && i.category !== categoryFilter) return false;
      if (targetedOnly && !i.targetsPlayer) return false;
      if (!q) return true;
      return (
        i.name.toLowerCase().includes(q) ||
        (i.category || '').toLowerCase().includes(q) ||
        (i.description || '').toLowerCase().includes(q)
      );
    });
    const byName = (a: GMItem, b: GMItem) => a.name.localeCompare(b.name);
    if (sortMode === 'category') {
      return [...filtered].sort((a, b) => (a.category || '').localeCompare(b.category || '') || byName(a, b));
    }
    if (sortMode === 'effect') {
      return [...filtered].sort((a, b) => effectScore(b) - effectScore(a) || byName(a, b));
    }
    return [...filtered].sort(byName);
  }, [items, query, categoryFilter, targetedOnly, sortMode]);

  const assign = async () => {
    if (!playerId || !itemId) return;
    setBusy(true);
    try {
      await gmAssignItem(playerId, itemId);
      setItemId('');
      load();
    } finally {
      setBusy(false);
    }
  };

  const remove = async (id: string) => {
    setBusy(true);
    try {
      await gmRemovePlayerItem(id);
      load();
    } finally {
      setBusy(false);
    }
  };

  const transfer = async (id: string, toPlayerId: string) => {
    setBusy(true);
    try {
      await gmTransferPlayerItem(id, toPlayerId);
      load();
    } finally {
      setBusy(false);
    }
  };

  if (loading) return <p className="text-bone/50">Loading items…</p>;

  // Group holdings by owner for a readable review list.
  const byPlayer = new Map<string, GMPlayerItem[]>();
  for (const h of held) {
    if (!byPlayer.has(h.playerName)) byPlayer.set(h.playerName, []);
    byPlayer.get(h.playerName)!.push(h);
  }
  const owners = [...byPlayer.entries()].sort((a, b) => a[0].localeCompare(b[0]));
  const itemByName = new Map(items.map((i) => [i.name, i]));

  return (
    <div className="flex flex-col gap-4">
      <p className="rounded-md border border-gold/50 bg-gold/10 p-3 text-sm text-bone/90">
        ⚠️ Trading items is allowed, but must be recorded by a GM (use “Transfer” on a holding).
      </p>

      <Card title="Assign an item">
        <div className="flex flex-col gap-3">
          <select
            value={playerId}
            onChange={(e) => setPlayerId(e.target.value)}
            className="rounded-md bg-black/60 border border-blood/40 p-2.5 text-bone"
          >
            <option value="">— Select player —</option>
            {assignablePlayers.map((p) => (
              <option key={p.id} value={p.id}>
                {p.character?.name}
                {p.character?.house ? ` (${p.character.house})` : ''}
              </option>
            ))}
          </select>

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
            {sortedItems.length === 0 ? (
              <p className="text-bone/50 text-sm p-2">No items match.</p>
            ) : (
              sortedItems.map((it) => (
                <button
                  key={it.id}
                  type="button"
                  onClick={() => setItemId(it.id)}
                  className={`flex items-start gap-2.5 rounded-md border p-2 text-left transition-colors ${
                    itemId === it.id
                      ? 'border-blood-bright bg-blood/20'
                      : 'border-blood/20 bg-black/30 hover:border-blood/50'
                  }`}
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
                    {it.description && (
                      <p className="text-xs text-bone/60 mt-0.5 line-clamp-2">{it.description}</p>
                    )}
                    {it.effect && <p className="text-xs text-gold/90 italic mt-0.5 line-clamp-1">{it.effect}</p>}
                    {effectTag(it) && <p className="text-xs text-gold/70 mt-0.5">{effectTag(it)}</p>}
                  </div>
                </button>
              ))
            )}
          </div>

          <button
            onClick={assign}
            disabled={busy || !playerId || !itemId}
            className="py-2 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
          >
            Assign
          </button>
        </div>
      </Card>

      {owners.length === 0 ? (
        <p className="text-bone/50 text-sm">No items assigned yet.</p>
      ) : (
        owners.map(([owner, list]) => (
          <Card key={owner} title={owner}>
            <div className="flex flex-col gap-3">
              {list.map((h) => (
                <HoldingRow
                  key={h.id}
                  holding={h}
                  photoItem={itemByName.get(h.itemName)}
                  players={assignablePlayers}
                  busy={busy}
                  onRemove={remove}
                  onTransfer={transfer}
                />
              ))}
            </div>
          </Card>
        ))
      )}
    </div>
  );
};

// One held item: shows the item + target, a Remove control, and a transfer
// picker that hands the item to a different player (the current owner loses it).
const HoldingRow = ({
  holding,
  photoItem,
  players,
  busy,
  onRemove,
  onTransfer,
}: {
  holding: GMPlayerItem;
  photoItem?: GMItem;
  players: GMPlayer[];
  busy: boolean;
  onRemove: (id: string) => void;
  onTransfer: (id: string, toPlayerId: string) => void;
}) => {
  const [to, setTo] = useState('');
  const others = players.filter((p) => p.id !== holding.playerId);

  return (
    <div className="border-b border-blood/15 last:border-0 pb-2">
      <div className="flex items-center gap-2">
        {photoItem?.hasPhoto && (
          <PhotoThumb
            src={itemPhotoUrl(photoItem.id)}
            alt={holding.itemName}
            className="w-10 h-10 rounded object-cover border border-blood/40 shrink-0"
          />
        )}
        <div className="flex-1 min-w-0">
          <p className="text-bone text-sm">{holding.itemName}</p>
          {holding.targetsPlayer && (
            <p className="text-xs text-bone/50">
              Target: {holding.targetName || <span className="text-bone/30">none set</span>}
            </p>
          )}
        </div>
        <button
          onClick={() => onRemove(holding.id)}
          disabled={busy}
          className="shrink-0 text-xs text-blood-bright uppercase tracking-[0.15em] disabled:opacity-40"
        >
          Remove
        </button>
      </div>
      <div className="flex items-center gap-2 mt-1.5">
        <select
          value={to}
          onChange={(e) => setTo(e.target.value)}
          className="flex-1 min-w-0 rounded-md bg-black/60 border border-blood/40 p-1.5 text-bone text-xs"
        >
          <option value="">— Transfer to… —</option>
          {others.map((p) => (
            <option key={p.id} value={p.id}>
              {p.character?.name}
              {p.character?.house ? ` (${p.character.house})` : ''}
            </option>
          ))}
        </select>
        <button
          onClick={() => {
            if (to) {
              onTransfer(holding.id, to);
              setTo('');
            }
          }}
          disabled={busy || !to}
          className="shrink-0 text-xs text-gold uppercase tracking-[0.15em] disabled:opacity-40"
        >
          Move
        </button>
      </div>
    </div>
  );
};

// Rough "how much does this item do" magnitude, for the picker's
// strongest-effect sort. Boolean effects (immune, reflect, …) count as one
// unit each — there's no principled way to compare them to a numeric BT/HF
// swing, so this is a sort order, not a real valuation.
const effectScore = (it: {
  hfEffect: number;
  btSelf: number;
  btFromTarget: number;
  btDeductTarget: number;
  quizBtPct: number;
  doubleGameBt: boolean;
  immune: boolean;
  reflect: boolean;
  stripResistance: boolean;
}): number =>
  Math.abs(it.hfEffect) +
  Math.abs(it.btSelf) +
  Math.abs(it.btFromTarget) +
  Math.abs(it.btDeductTarget) +
  Math.abs(it.quizBtPct) +
  (it.doubleGameBt ? 1 : 0) +
  (it.immune ? 1 : 0) +
  (it.reflect ? 1 : 0) +
  (it.stripResistance ? 1 : 0);

// A one-line summary of an item's auto-applied effects (HF + BT tally), for the
// dropdown and catalog. Free-text effects that aren't auto-resolved don't show.
const effectTag = (it: {
  hfEffect: number;
  btSelf: number;
  btFromTarget: number;
  btDeductTarget: number;
  quizBtPct: number;
  doubleGameBt: boolean;
  immune: boolean;
  reflect: boolean;
  stripResistance: boolean;
}): string => {
  const parts: string[] = [];
  if (it.hfEffect) parts.push(`${it.hfEffect > 0 ? '+' : ''}${it.hfEffect} HF`);
  if (it.btSelf) parts.push(`${it.btSelf > 0 ? '+' : ''}${it.btSelf} BT`);
  if (it.btFromTarget) parts.push(`steal ${it.btFromTarget} BT`);
  if (it.btDeductTarget) parts.push(`−${it.btDeductTarget} BT to target`);
  if (it.quizBtPct) parts.push(`+${it.quizBtPct}% quiz BT`);
  if (it.doubleGameBt) parts.push('double game BT');
  if (it.immune) parts.push('immune');
  if (it.reflect) parts.push('reflect');
  if (it.stripResistance) parts.push('strip resistance');
  return parts.join(' · ');
};
