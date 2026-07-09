import { useEffect, useMemo, useState } from 'react';
import {
  gmListItems,
  gmListPlayers,
  gmListPlayerItems,
  gmAssignItem,
  gmRemovePlayerItem,
  gmCreateItem,
  gmUpdateItem,
  gmDeleteItem,
  gmTransferPlayerItem,
} from '../../gmApi';
import type { GMItem, GMItemDraft, GMPlayer, GMPlayerItem } from '../../gmApi';
import { Card } from './GameSection';

const emptyDraft: GMItemDraft = {
  code: '',
  name: '',
  description: '',
  effect: '',
  targetsPlayer: false,
  hfEffect: 0,
  btSelf: 0,
  btFromTarget: 0,
  btDeductTarget: 0,
  quizBtPct: 0,
  doubleGameBt: false,
  immune: false,
  reflect: false,
  stripResistance: false,
};

const field = 'rounded-md bg-black/60 border border-blood/40 p-2.5 text-bone text-sm';

// GM Items tab: create/delete catalog items, assign them to players, and
// review/remove holdings.
export const ItemsSection = () => {
  const [items, setItems] = useState<GMItem[]>([]);
  const [players, setPlayers] = useState<GMPlayer[]>([]);
  const [held, setHeld] = useState<GMPlayerItem[]>([]);
  const [playerId, setPlayerId] = useState('');
  const [itemId, setItemId] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [catalogOpen, setCatalogOpen] = useState(false);
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

  const sortedItems = useMemo(
    () => [...items].sort((a, b) => a.name.localeCompare(b.name)),
    [items]
  );

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

  const deleteItem = async (it: GMItem) => {
    const holders = held.filter((h) => h.itemName === it.name).length;
    const msg = holders
      ? `Delete "${it.name}"? It is held by ${holders} player(s); those assignments will be removed too.`
      : `Delete "${it.name}" from the catalog?`;
    if (!window.confirm(msg)) return;
    setBusy(true);
    try {
      await gmDeleteItem(it.id);
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

  return (
    <div className="flex flex-col gap-4">
      <NewItemForm busy={busy} onCreated={load} />

      <Card title="Assign an item">
        <div className="flex flex-col gap-2">
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
          <select
            value={itemId}
            onChange={(e) => setItemId(e.target.value)}
            className="rounded-md bg-black/60 border border-blood/40 p-2.5 text-bone"
          >
            <option value="">— Select item —</option>
            {sortedItems.map((it) => (
              <option key={it.id} value={it.id}>
                {it.name}
                {effectTag(it) ? ` — ${effectTag(it)}` : ''}
              </option>
            ))}
          </select>
          <button
            onClick={assign}
            disabled={busy || !playerId || !itemId}
            className="py-2 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
          >
            Assign
          </button>
        </div>
      </Card>

      <Card title={`Catalog (${sortedItems.length})`}>
        <button
          onClick={() => setCatalogOpen((o) => !o)}
          className="text-xs text-gold uppercase tracking-[0.15em] mb-2"
        >
          {catalogOpen ? '▾ Hide catalog' : '▸ Show catalog'}
        </button>
        {!catalogOpen ? null : sortedItems.length === 0 ? (
          <p className="text-bone/50 text-sm">No items yet.</p>
        ) : (
          <div className="flex flex-col gap-1.5">
            {sortedItems.map((it) => (
              <div key={it.id} className="border-b border-blood/15 last:border-0 pb-1.5">
                <div className="flex items-center gap-2">
                  <div className="flex-1 min-w-0">
                    <p className="text-bone text-sm">
                      {it.name}
                      {it.targetsPlayer && <span className="ml-2 text-[10px] uppercase tracking-[0.15em] text-blood-bright">targeted</span>}
                    </p>
                    {effectTag(it) && <p className="text-xs text-gold/80">{effectTag(it)}</p>}
                  </div>
                  <button
                    onClick={() => setEditingId((cur) => (cur === it.id ? null : it.id))}
                    disabled={busy}
                    className="shrink-0 text-xs text-gold uppercase tracking-[0.15em] disabled:opacity-40"
                  >
                    {editingId === it.id ? 'Close' : 'Edit'}
                  </button>
                  <button
                    onClick={() => deleteItem(it)}
                    disabled={busy}
                    className="shrink-0 text-xs text-blood-bright uppercase tracking-[0.15em] disabled:opacity-40"
                  >
                    Delete
                  </button>
                </div>
                {editingId === it.id && (
                  <div className="mt-2">
                    <ItemForm
                      initial={toDraft(it)}
                      submitLabel="Save changes"
                      onSubmit={(draft) => gmUpdateItem(it.id, draft)}
                      onDone={() => {
                        setEditingId(null);
                        load();
                      }}
                    />
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
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
  players,
  busy,
  onRemove,
  onTransfer,
}: {
  holding: GMPlayerItem;
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

const toDraft = (it: GMItem): GMItemDraft => {
  const { id: _id, ...rest } = it;
  return rest;
};

// Collapsible wrapper for authoring a brand-new item.
const NewItemForm = ({ busy, onCreated }: { busy: boolean; onCreated: () => void }) => {
  const [open, setOpen] = useState(false);
  return (
    <Card title="Add a new item">
      <button onClick={() => setOpen((o) => !o)} className="text-xs text-gold uppercase tracking-[0.15em]">
        {open ? '▾ Hide form' : '▸ New item'}
      </button>
      {open && (
        <div className="mt-3">
          <ItemForm
            initial={emptyDraft}
            submitLabel="Create item"
            resetOnDone
            disabled={busy}
            onSubmit={(draft) => gmCreateItem(draft)}
            onDone={onCreated}
          />
        </div>
      )}
    </Card>
  );
};

// Shared create/edit form. onSubmit performs the API call; onDone refreshes.
const ItemForm = ({
  initial,
  submitLabel,
  onSubmit,
  onDone,
  resetOnDone,
  disabled,
}: {
  initial: GMItemDraft;
  submitLabel: string;
  onSubmit: (draft: GMItemDraft) => Promise<unknown>;
  onDone: () => void;
  resetOnDone?: boolean;
  disabled?: boolean;
}) => {
  const [draft, setDraft] = useState<GMItemDraft>(initial);
  const [saving, setSaving] = useState(false);
  const [note, setNote] = useState<string | null>(null);

  const set = <K extends keyof GMItemDraft>(k: K, v: GMItemDraft[K]) =>
    setDraft((d) => ({ ...d, [k]: v }));
  const num = (k: keyof GMItemDraft) => (e: React.ChangeEvent<HTMLInputElement>) =>
    set(k, (Number(e.target.value) || 0) as never);

  const submit = async () => {
    if (!draft.name.trim()) {
      setNote('Name is required.');
      return;
    }
    setSaving(true);
    setNote(null);
    try {
      await onSubmit(draft);
      if (resetOnDone) setDraft(emptyDraft);
      onDone();
    } catch (e) {
      setNote(e instanceof Error ? e.message : 'Save failed.');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="flex flex-col gap-2">
      <input
        className={field}
        placeholder="Name (required)"
        value={draft.name}
        onChange={(e) => set('name', e.target.value)}
      />
      <textarea
        className={field}
        rows={2}
        placeholder="Description (flavor, shown to the player)"
        value={draft.description}
        onChange={(e) => set('description', e.target.value)}
      />
      <input
        className={field}
        placeholder="Effect text (shown to the player)"
        value={draft.effect}
        onChange={(e) => set('effect', e.target.value)}
      />
      <label className="flex items-center gap-2 text-sm text-bone/80">
        <input
          type="checkbox"
          checked={draft.targetsPlayer}
          onChange={(e) => set('targetsPlayer', e.target.checked)}
        />
        Targets another player (shows a target picker)
      </label>

      <p className="text-[11px] uppercase tracking-[0.15em] text-gold mt-2">Auto-applied effects</p>
      <div className="grid grid-cols-2 gap-2">
        <NumIn label="House Favor" v={draft.hfEffect} onChange={num('hfEffect')} />
        <NumIn label="Flat BT (self)" v={draft.btSelf} onChange={num('btSelf')} />
        <NumIn label="Steal BT from target" v={draft.btFromTarget} onChange={num('btFromTarget')} />
        <NumIn label="Deduct BT from target" v={draft.btDeductTarget} onChange={num('btDeductTarget')} />
        <NumIn label="Quiz BT bonus %" v={draft.quizBtPct} onChange={num('quizBtPct')} />
      </div>
      <div className="flex flex-col gap-1.5 mt-1">
        <Chk label="Double game BT" v={draft.doubleGameBt} onChange={(v) => set('doubleGameBt', v)} />
        <Chk label="Immune to incoming steals/deducts" v={draft.immune} onChange={(v) => set('immune', v)} />
        <Chk label="Reflect incoming loss to attacker" v={draft.reflect} onChange={(v) => set('reflect', v)} />
        <Chk label="Strip target's resistance" v={draft.stripResistance} onChange={(v) => set('stripResistance', v)} />
      </div>

      <div className="flex items-center gap-3 mt-2">
        <button
          onClick={submit}
          disabled={saving || disabled}
          className="py-2 px-5 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40"
        >
          {saving ? 'Saving…' : submitLabel}
        </button>
        {note && <span className="text-bone/60 text-sm">{note}</span>}
      </div>
      <p className="text-[11px] text-bone/40">
        Steal/deduct effects need "Targets another player" checked, and the target is chosen by the
        player (or a GM) in their inventory before the quiz locks.
      </p>
    </div>
  );
};

const NumIn = ({
  label,
  v,
  onChange,
}: {
  label: string;
  v: number;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
}) => (
  <label className="flex flex-col gap-1 text-[11px] text-bone/60">
    {label}
    <input type="number" className={`${field} text-center`} value={v} onChange={onChange} />
  </label>
);

const Chk = ({
  label,
  v,
  onChange,
}: {
  label: string;
  v: boolean;
  onChange: (v: boolean) => void;
}) => (
  <label className="flex items-center gap-2 text-sm text-bone/80">
    <input type="checkbox" checked={v} onChange={(e) => onChange(e.target.checked)} />
    {label}
  </label>
);
