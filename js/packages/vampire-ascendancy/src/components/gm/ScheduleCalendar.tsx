import { useMemo, useState } from 'react';
import { gmSetGameSchedule } from '../../gmApi';
import type { GMGame } from '../../gmApi';
import { formatClock } from '../../theme';
import { Card } from './GameSection';
import { GM_NAMES } from './GMAdmin';

const DAY_START = 18 * 60; // 6 PM
const DAY_END = 24 * 60; // midnight
const PPM = 1.7; // pixels per minute
const SNAP = 15; // minutes
const DEFAULT_DUR = 30;

const height = (DAY_END - DAY_START) * PPM;
const clamp = (n: number, lo: number, hi: number) => Math.max(lo, Math.min(hi, n));
const snap = (n: number) => Math.round(n / SNAP) * SNAP;

type Drag = { id: string; startY: number; origStart: number; curStart: number; moved: boolean };

// A Google-Calendar-style evening timeline (6pm–midnight). Scheduled games are
// blocks you can drag to move; tap one to edit its time, length, and location.
export const ScheduleCalendar = ({ games, onChange }: { games: GMGame[]; onChange: () => void }) => {
  const [editingId, setEditingId] = useState<string | null>(null);
  const [drag, setDrag] = useState<Drag | null>(null);

  const scheduled = useMemo(
    () => games.filter((g) => g.startMinutes != null && g.endMinutes != null),
    [games]
  );
  const unscheduled = useMemo(() => games.filter((g) => g.startMinutes == null), [games]);

  // Lane-pack overlapping blocks so they sit side by side.
  const { laneOf, lanes } = useMemo(() => {
    const sorted = [...scheduled].sort((a, b) => a.startMinutes! - b.startMinutes!);
    const laneEnds: number[] = [];
    const laneOf: Record<string, number> = {};
    for (const g of sorted) {
      let lane = laneEnds.findIndex((end) => end <= g.startMinutes!);
      if (lane === -1) {
        lane = laneEnds.length;
        laneEnds.push(0);
      }
      laneEnds[lane] = g.endMinutes!;
      laneOf[g.id] = lane;
    }
    return { laneOf, lanes: Math.max(1, laneEnds.length) };
  }, [scheduled]);

  const hours: number[] = [];
  for (let m = DAY_START; m <= DAY_END; m += 60) hours.push(m);

  const onPointerDown = (e: React.PointerEvent, g: GMGame) => {
    (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
    setDrag({ id: g.id, startY: e.clientY, origStart: g.startMinutes!, curStart: g.startMinutes!, moved: false });
  };
  const onPointerMove = (e: React.PointerEvent, g: GMGame) => {
    if (!drag || drag.id !== g.id) return;
    const dy = e.clientY - drag.startY;
    const dur = g.endMinutes! - g.startMinutes!;
    const ns = clamp(snap(drag.origStart + dy / PPM), DAY_START, DAY_END - dur);
    setDrag({ ...drag, curStart: ns, moved: drag.moved || Math.abs(dy) > 4 });
  };
  const onPointerUp = async (g: GMGame) => {
    if (!drag || drag.id !== g.id) return;
    const d = drag;
    setDrag(null);
    if (d.moved && d.curStart !== d.origStart) {
      const dur = g.endMinutes! - g.startMinutes!;
      await gmSetGameSchedule(g.id, {
        startMinutes: d.curStart,
        endMinutes: d.curStart + dur,
        location: g.location,
        assignedGm: g.assignedGm,
        runNotes: g.runNotes,
      });
      onChange();
    } else if (!d.moved) {
      setEditingId(g.id);
    }
  };

  // First 30-min slot that doesn't overlap a scheduled block (for new placements).
  const firstFreeSlot = () => {
    for (let s = DAY_START; s + DEFAULT_DUR <= DAY_END; s += SNAP) {
      const clash = scheduled.some((g) => s < g.endMinutes! && s + DEFAULT_DUR > g.startMinutes!);
      if (!clash) return s;
    }
    return DAY_START;
  };

  const scheduleUnscheduled = async (g: GMGame) => {
    const s = firstFreeSlot();
    await gmSetGameSchedule(g.id, {
      startMinutes: s,
      endMinutes: s + DEFAULT_DUR,
      location: g.location,
      assignedGm: g.assignedGm,
      runNotes: g.runNotes,
    });
    onChange();
    setEditingId(g.id);
  };

  const editing = games.find((g) => g.id === editingId) || null;

  return (
    <Card title="Schedule · 6 PM – Midnight">
      <div className="flex gap-2">
        {/* hour gutter */}
        <div className="relative w-14 shrink-0" style={{ height }}>
          {hours.map((m) => (
            <div
              key={m}
              className="absolute right-1 -translate-y-1/2 text-[10px] uppercase tracking-wide text-bone/40"
              style={{ top: (m - DAY_START) * PPM }}
            >
              {formatClock(m)}
            </div>
          ))}
        </div>

        {/* timeline */}
        <div className="relative flex-1 rounded-md border border-blood/30 bg-black/30" style={{ height }}>
          {hours.map((m) => (
            <div
              key={m}
              className="absolute left-0 right-0 border-t border-blood/15"
              style={{ top: (m - DAY_START) * PPM }}
            />
          ))}
          {scheduled.map((g) => {
            const start = drag && drag.id === g.id ? drag.curStart : g.startMinutes!;
            const dur = g.endMinutes! - g.startMinutes!;
            const lane = laneOf[g.id] ?? 0;
            const dragging = drag?.id === g.id;
            return (
              <div
                key={g.id}
                onPointerDown={(e) => onPointerDown(e, g)}
                onPointerMove={(e) => onPointerMove(e, g)}
                onPointerUp={() => onPointerUp(g)}
                className={`absolute rounded-md border px-2 py-1 overflow-hidden cursor-grab active:cursor-grabbing select-none ${
                  g.status === 'played'
                    ? 'bg-green-900/40 border-green-500/40'
                    : 'bg-blood/30 border-blood-bright/50'
                } ${dragging ? 'opacity-90 ring-1 ring-gold z-10' : ''}`}
                style={{
                  top: (start - DAY_START) * PPM,
                  height: Math.max(dur * PPM - 2, 18),
                  left: `${(lane / lanes) * 100}%`,
                  width: `calc(${100 / lanes}% - 4px)`,
                }}
                title="Drag to move · tap to edit"
              >
                <p className="text-xs font-semibold text-bone leading-tight truncate">{g.name}</p>
                <p className="text-[10px] text-bone/70 leading-tight">
                  {formatClock(start)}–{formatClock(start + dur)}
                </p>
                {g.location && <p className="text-[10px] text-gold/80 leading-tight truncate">📍 {g.location}</p>}
                {g.assignedGm && <p className="text-[10px] text-bone/50 leading-tight truncate">👤 {g.assignedGm}</p>}
              </div>
            );
          })}
        </div>
      </div>

      {unscheduled.length > 0 && (
        <div className="mt-3">
          <p className="text-[11px] uppercase tracking-[0.15em] text-bone/50 mb-1.5">Unscheduled</p>
          <div className="flex flex-wrap gap-1.5">
            {unscheduled.map((g) => (
              <button
                key={g.id}
                onClick={() => scheduleUnscheduled(g)}
                className="rounded-full border border-blood/40 px-3 py-1 text-xs text-bone/80 hover:text-bone hover:border-blood-bright"
              >
                + {g.name}
              </button>
            ))}
          </div>
        </div>
      )}

      {editing && editing.startMinutes != null && (
        <SlotEditor
          key={editing.id}
          game={editing}
          onClose={() => setEditingId(null)}
          onChange={onChange}
        />
      )}
    </Card>
  );
};

const DURATIONS = [15, 30, 45, 60, 90, 120];
const startOptions: number[] = [];
for (let m = DAY_START; m <= DAY_END - SNAP; m += SNAP) startOptions.push(m);

const SlotEditor = ({ game, onClose, onChange }: { game: GMGame; onClose: () => void; onChange: () => void }) => {
  const [start, setStart] = useState(game.startMinutes!);
  const [dur, setDur] = useState(clamp(game.endMinutes! - game.startMinutes!, SNAP, DAY_END - DAY_START));
  const [location, setLocation] = useState(game.location);
  const [assignedGm, setAssignedGm] = useState(game.assignedGm);
  const [runNotes, setRunNotes] = useState(game.runNotes);
  const [busy, setBusy] = useState(false);

  const end = Math.min(start + dur, DAY_END);
  const field = 'rounded-md bg-black/60 border border-blood/40 p-2 text-bone text-sm';

  const save = async () => {
    setBusy(true);
    try {
      await gmSetGameSchedule(game.id, {
        startMinutes: start,
        endMinutes: end,
        location: location.trim(),
        assignedGm,
        runNotes,
      });
      onChange();
      onClose();
    } finally {
      setBusy(false);
    }
  };
  const unschedule = async () => {
    setBusy(true);
    try {
      await gmSetGameSchedule(game.id, {
        startMinutes: null,
        endMinutes: null,
        location: location.trim(),
        assignedGm,
        runNotes,
      });
      onChange();
      onClose();
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="mt-3 rounded-md border border-blood/40 bg-black/40 p-3 flex flex-col gap-2">
      <p className="text-sm text-bone font-semibold">{game.name}</p>
      <div className="grid grid-cols-2 gap-2">
        <label className="flex flex-col gap-1 text-[11px] text-bone/60">
          Start
          <select className={field} value={start} onChange={(e) => setStart(Number(e.target.value))}>
            {startOptions.map((m) => (
              <option key={m} value={m}>
                {formatClock(m)}
              </option>
            ))}
          </select>
        </label>
        <label className="flex flex-col gap-1 text-[11px] text-bone/60">
          Length
          <select className={field} value={dur} onChange={(e) => setDur(Number(e.target.value))}>
            {DURATIONS.map((d) => (
              <option key={d} value={d}>
                {d} min
              </option>
            ))}
          </select>
        </label>
      </div>
      <div className="grid grid-cols-2 gap-2">
        <label className="flex flex-col gap-1 text-[11px] text-bone/60">
          Location
          <input
            className={field}
            value={location}
            onChange={(e) => setLocation(e.target.value)}
            placeholder="e.g. The Ballroom"
          />
        </label>
        <label className="flex flex-col gap-1 text-[11px] text-bone/60">
          Run by (GM)
          <select className={field} value={assignedGm} onChange={(e) => setAssignedGm(e.target.value)}>
            <option value="">— unassigned —</option>
            {GM_NAMES.map((n) => (
              <option key={n} value={n}>
                {n}
              </option>
            ))}
          </select>
        </label>
      </div>
      <label className="flex flex-col gap-1 text-[11px] text-bone/60">
        How to run (GM-only notes)
        <textarea
          className={field}
          rows={3}
          value={runNotes}
          onChange={(e) => setRunNotes(e.target.value)}
          placeholder="Setup, rules, materials, scoring — only GMs see this."
        />
      </label>
      <p className="text-[11px] text-bone/40">{formatClock(start)} – {formatClock(end)}</p>
      <div className="flex items-center gap-2">
        <button
          onClick={save}
          disabled={busy}
          className="py-1.5 px-4 rounded-md bg-blood text-bone uppercase tracking-[0.15em] text-xs disabled:opacity-40"
        >
          Save
        </button>
        <button
          onClick={unschedule}
          disabled={busy}
          className="py-1.5 px-3 rounded-md border border-blood/50 text-blood-bright uppercase tracking-[0.15em] text-xs disabled:opacity-40"
        >
          Unschedule
        </button>
        <button onClick={onClose} className="ml-auto text-xs text-bone/50 uppercase tracking-[0.15em]">
          Close
        </button>
      </div>
    </div>
  );
};
