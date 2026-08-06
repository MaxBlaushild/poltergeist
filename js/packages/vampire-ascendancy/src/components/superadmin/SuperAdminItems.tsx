import { useEffect, useMemo, useRef, useState } from 'react';
import {
  adminListItems,
  adminCreateItem,
  adminUpdateItem,
  adminDeleteItem,
  adminSetItemPhoto,
  adminDeleteItemPhoto,
} from '../../superAdminApi';
import type { GMItem, GMItemDraft } from '../../gmApi';
import { itemPhotoUrl } from '../../gmApi';
import { Card } from '../gm/GameSection';

// Load a picked image, downscale it (phone photos are multi-MB), return a JPEG data URL.
const resizeImage = (file: File, maxDim = 1200, quality = 0.8): Promise<string> =>
  new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const img = new Image();
      img.onload = () => {
        const scale = Math.min(1, maxDim / Math.max(img.width, img.height));
        const w = Math.round(img.width * scale);
        const h = Math.round(img.height * scale);
        const canvas = document.createElement('canvas');
        canvas.width = w;
        canvas.height = h;
        const cctx = canvas.getContext('2d');
        if (!cctx) return reject(new Error('no canvas'));
        cctx.drawImage(img, 0, 0, w, h);
        resolve(canvas.toDataURL('image/jpeg', quality));
      };
      img.onerror = reject;
      img.src = reader.result as string;
    };
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });

const PhotoThumb = ({ src, alt, className }: { src: string; alt: string; className?: string }) => {
  const [open, setOpen] = useState(false);
  return (
    <>
      <img src={src} alt={alt} onClick={() => setOpen(true)} className={`cursor-zoom-in ${className ?? ''}`} />
      {open && (
        <div
          className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-black/90 p-6"
          onClick={() => setOpen(false)}
          role="dialog"
          aria-label={`${alt} photo`}
        >
          <img src={src} alt={alt} className="max-h-[85vh] max-w-[92vw] rounded-lg border-2 border-blood/40 object-contain" />
          <p className="mt-3 font-display text-xl text-bone">{alt}</p>
        </div>
      )}
    </>
  );
};

const ItemPhoto = ({ item, onChanged }: { item: GMItem; onChanged: () => void }) => {
  const inputRef = useRef<HTMLInputElement>(null);
  const [busy, setBusy] = useState(false);
  const [ver, setVer] = useState(1);

  const pick = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file) return;
    setBusy(true);
    try {
      await adminSetItemPhoto(item.id, await resizeImage(file));
      setVer(Date.now());
      onChanged();
    } finally {
      setBusy(false);
    }
  };
  const remove = async () => {
    setBusy(true);
    try {
      await adminDeleteItemPhoto(item.id);
      onChanged();
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="flex items-center gap-2">
      {item.hasPhoto ? (
        <PhotoThumb src={itemPhotoUrl(item.id, ver)} alt={item.name} className="w-12 h-12 rounded object-cover border border-blood/40" />
      ) : (
        <div className="w-12 h-12 rounded border border-dashed border-blood/40 flex items-center justify-center text-bone/30">📷</div>
      )}
      <input ref={inputRef} type="file" accept="image/*" capture="environment" onChange={pick} className="hidden" />
      <button onClick={() => inputRef.current?.click()} disabled={busy} className="text-xs text-gold uppercase tracking-[0.15em] disabled:opacity-40">
        {busy ? '…' : item.hasPhoto ? '📷 Change' : '📷 Add photo'}
      </button>
      {item.hasPhoto && !busy && (
        <button onClick={remove} className="text-xs text-blood-bright uppercase tracking-[0.15em]">
          Remove
        </button>
      )}
    </div>
  );
};

const emptyDraft: GMItemDraft = {
  code: '',
  name: '',
  category: '',
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

// The shared item catalog: create, edit, delete, and photograph items —
// affects every Toast that includes them. Which items a given Toast
// includes is a per-instance toggle on the GM console's Content tab.
export const SuperAdminItems = () => {
  const [items, setItems] = useState<GMItem[]>([]);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [query, setQuery] = useState('');
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);

  const load = () => {
    adminListItems()
      .then((d) => setItems(d.items))
      .finally(() => setLoading(false));
  };
  useEffect(() => {
    load();
  }, []);

  const sorted = useMemo(() => [...items].sort((a, b) => a.name.localeCompare(b.name)), [items]);
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return sorted;
    return sorted.filter((i) => i.name.toLowerCase().includes(q) || (i.category || '').toLowerCase().includes(q));
  }, [sorted, query]);

  const deleteItem = async (it: GMItem) => {
    if (!window.confirm(`Delete "${it.name}" from the catalog? Any player holding it, in any Toast, loses it.`)) return;
    setBusy(true);
    try {
      await adminDeleteItem(it.id);
      load();
    } finally {
      setBusy(false);
    }
  };

  if (loading) return <Card title="Items">Loading…</Card>;

  return (
    <div className="flex flex-col gap-4">
      <NewItemForm busy={busy} onCreated={load} />

      <Card title={`Catalog (${sorted.length}) — shared, affects every Toast`}>
        <input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search items by name or category…"
          className="w-full mb-2 rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm"
        />
        {filtered.length === 0 ? (
          <p className="text-bone/50 text-sm">No items match.</p>
        ) : (
          <div className="flex flex-col gap-1.5">
            {filtered.map((it) => (
              <div key={it.id} className="border-b border-blood/15 last:border-0 pb-1.5">
                <div className="flex items-center gap-2">
                  <div className="flex-1 min-w-0">
                    <p className="text-bone text-sm">
                      {it.name}
                      {it.category && (
                        <span className="ml-2 text-[10px] uppercase tracking-[0.15em] rounded-full border border-blood/40 px-1.5 py-0.5 text-bone/60">
                          {it.category}
                        </span>
                      )}
                    </p>
                    <div className="mt-1.5">
                      <ItemPhoto item={it} onChanged={load} />
                    </div>
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
                      onSubmit={(draft) => adminUpdateItem(it.id, draft)}
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
    </div>
  );
};

const toDraft = (it: GMItem): GMItemDraft => {
  const { id: _id, hasPhoto: _p, ...rest } = it;
  return rest;
};

const NewItemForm = ({ busy, onCreated }: { busy: boolean; onCreated: () => void }) => {
  const [open, setOpen] = useState(false);
  return (
    <Card title="Add a new item">
      <button onClick={() => setOpen((o) => !o)} className="text-xs text-gold uppercase tracking-[0.15em]">
        {open ? '▾ Hide form' : '▸ New item'}
      </button>
      {open && (
        <div className="mt-3">
          <ItemForm initial={emptyDraft} submitLabel="Create item" resetOnDone disabled={busy} onSubmit={(draft) => adminCreateItem(draft)} onDone={onCreated} />
        </div>
      )}
    </Card>
  );
};

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

  const set = <K extends keyof GMItemDraft>(k: K, v: GMItemDraft[K]) => setDraft((d) => ({ ...d, [k]: v }));
  const num = (k: keyof GMItemDraft) => (e: React.ChangeEvent<HTMLInputElement>) => set(k, (Number(e.target.value) || 0) as never);

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
      <input className={field} placeholder="Name (required)" value={draft.name} onChange={(e) => set('name', e.target.value)} />
      <input className={field} placeholder="Category (e.g. War, Glory, Protection)" value={draft.category} onChange={(e) => set('category', e.target.value)} />
      <textarea className={field} rows={2} placeholder="Description (flavor, shown to the player)" value={draft.description} onChange={(e) => set('description', e.target.value)} />
      <input className={field} placeholder="Effect text (shown to the player)" value={draft.effect} onChange={(e) => set('effect', e.target.value)} />
      <label className="flex items-center gap-2 text-sm text-bone/80">
        <input type="checkbox" checked={draft.targetsPlayer} onChange={(e) => set('targetsPlayer', e.target.checked)} />
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
        <button onClick={submit} disabled={saving || disabled} className="py-2 px-5 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-sm disabled:opacity-40">
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

const NumIn = ({ label, v, onChange }: { label: string; v: number; onChange: (e: React.ChangeEvent<HTMLInputElement>) => void }) => (
  <label className="flex flex-col gap-1 text-[11px] text-bone/60">
    {label}
    <input type="number" className={`${field} text-center`} value={v} onChange={onChange} />
  </label>
);

const Chk = ({ label, v, onChange }: { label: string; v: boolean; onChange: (v: boolean) => void }) => (
  <label className="flex items-center gap-2 text-sm text-bone/80">
    <input type="checkbox" checked={v} onChange={(e) => onChange(e.target.checked)} />
    {label}
  </label>
);
