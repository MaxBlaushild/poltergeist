// Small presentational primitives shared across the Super Admin editors
// (Characters, Mysteries) — a labeled field wrapper, an add/remove ordered
// list section, and a round "×" remove button.

export const Field = ({ label, children }: { label: string; children: React.ReactNode }) => (
  <label className="flex flex-col gap-1">
    <span className="text-[11px] uppercase tracking-[0.15em] text-bone/50">{label}</span>
    {children}
  </label>
);

export const ListEditor = ({
  label,
  addLabel,
  onAdd,
  children,
}: {
  label: string;
  addLabel: string;
  onAdd: () => void;
  children: React.ReactNode;
}) => (
  <div className="flex flex-col gap-2">
    <div className="flex items-center justify-between">
      <span className="text-[11px] uppercase tracking-[0.15em] text-bone/50">{label}</span>
      <button onClick={onAdd} className="text-xs text-gold uppercase tracking-[0.15em]">
        {addLabel}
      </button>
    </div>
    {children}
  </div>
);

export const RemoveBtn = ({ onClick }: { onClick: () => void }) => (
  <button
    onClick={onClick}
    className="shrink-0 mt-1 w-6 h-6 rounded-full border border-blood/50 text-blood-bright text-xs leading-none"
    aria-label="Remove"
  >
    ×
  </button>
);
