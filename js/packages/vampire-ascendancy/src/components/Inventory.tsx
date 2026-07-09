import { useEffect, useState } from 'react';
import { getInventory, setInventoryTarget, getToken } from '../api';
import type { InventoryItem, InventoryTarget } from '../types';
import { VampireMark } from './VampireMark';

// "What you own." Lists the player's items; targeting items get a target picker,
// which locks once the closing quiz begins.
export const Inventory = () => {
  const token = getToken() || '';
  const [items, setItems] = useState<InventoryItem[] | null>(null);
  const [targets, setTargets] = useState<InventoryTarget[]>([]);
  const [locked, setLocked] = useState(false);

  const load = () => {
    getInventory(token)
      .then((d) => {
        setItems(d.items);
        setTargets(d.targets);
        setLocked(d.locked);
      })
      .catch(() => setItems([]));
  };
  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="pb-8">
      <header className="text-center mb-6">
        <p className="text-xs uppercase tracking-[0.4em] text-gold">Your Inventory</p>
        <h1 className="mt-3 font-display text-3xl font-bold text-bone">What You Carry</h1>
        {locked && (
          <p className="mt-2 text-xs uppercase tracking-[0.2em] text-blood-bright">
            Targets are locked — the quiz has begun
          </p>
        )}
      </header>

      {items === null ? (
        <p className="text-center text-bone/50">Opening your satchel…</p>
      ) : items.length === 0 ? (
        <div className="rounded-lg border border-blood/40 bg-black/40 p-8 text-center">
          <VampireMark className="w-12 h-12 mx-auto mb-3 opacity-80" />
          <p className="text-bone/85">You carry nothing yet. Items acquired tonight will appear here.</p>
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          {items.map((it) => (
            <ItemCard
              key={it.id}
              item={it}
              targets={targets}
              locked={locked}
              token={token}
              onChanged={load}
            />
          ))}
        </div>
      )}
    </div>
  );
};

const ItemCard = ({
  item,
  targets,
  locked,
  token,
  onChanged,
}: {
  item: InventoryItem;
  targets: InventoryTarget[];
  locked: boolean;
  token: string;
  onChanged: () => void;
}) => {
  const [saving, setSaving] = useState(false);

  const pickTarget = async (targetPlayerId: string) => {
    setSaving(true);
    try {
      await setInventoryTarget(token, item.id, targetPlayerId);
      onChanged();
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="rounded-lg border border-blood/40 bg-black/40 p-5">
      <h2 className="font-display text-xl text-bone mb-1">{item.name}</h2>
      {item.description && <p className="text-bone/80 leading-relaxed mb-2">{item.description}</p>}
      {item.effect && (
        <p className="text-sm text-gold/90 italic mb-1">{item.effect}</p>
      )}

      {item.targetsPlayer && (
        <div className="mt-3">
          <label className="block text-[11px] uppercase tracking-[0.2em] text-bone/50 mb-1">
            Target
          </label>
          <select
            value={item.targetPlayerId ?? ''}
            disabled={locked || saving}
            onChange={(e) => pickTarget(e.target.value)}
            className="w-full rounded-md bg-black/60 border border-blood/40 p-2.5 text-bone disabled:opacity-50"
          >
            <option value="">— Choose a target —</option>
            {targets.map((t) => (
              <option key={t.playerId} value={t.playerId}>
                {t.name}
              </option>
            ))}
          </select>
        </div>
      )}
    </div>
  );
};
