// Shared item-display helpers — used by the Items tab's assign picker and
// the Content tab's include/exclude picker, so both show items exactly the
// same way.

interface ItemEffects {
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

// A one-line summary of an item's auto-applied effects (HF + BT tally).
// Free-text effects that aren't auto-resolved don't show here — see the
// item's own `effect` field for those.
export const effectTag = (it: ItemEffects): string => {
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

// Rough "how much does this item do" magnitude, for a strongest-effect
// sort. Boolean effects (immune, reflect, …) count as one unit each —
// there's no principled way to compare them to a numeric BT/HF swing, so
// this is a sort order, not a real valuation.
export const effectScore = (it: ItemEffects): number =>
  Math.abs(it.hfEffect) +
  Math.abs(it.btSelf) +
  Math.abs(it.btFromTarget) +
  Math.abs(it.btDeductTarget) +
  Math.abs(it.quizBtPct) +
  (it.doubleGameBt ? 1 : 0) +
  (it.immune ? 1 : 0) +
  (it.reflect ? 1 : 0) +
  (it.stripResistance ? 1 : 0);

export type SortMode = 'name' | 'category' | 'effect';

// Shared filter+sort over a list of items with { name, category,
// description, targetsPlayer } — a search query (name/category/
// description), an exact-category filter, a targets-a-player filter, then
// one of the three sort modes.
export function filterAndSortItems<
  T extends ItemEffects & { name: string; category: string; description: string; targetsPlayer: boolean },
>(items: T[], opts: { query: string; category: string; targetedOnly: boolean; sortMode: SortMode }): T[] {
  const q = opts.query.trim().toLowerCase();
  const filtered = items.filter((i) => {
    if (opts.category && i.category !== opts.category) return false;
    if (opts.targetedOnly && !i.targetsPlayer) return false;
    if (!q) return true;
    return (
      i.name.toLowerCase().includes(q) ||
      (i.category || '').toLowerCase().includes(q) ||
      (i.description || '').toLowerCase().includes(q)
    );
  });
  const byName = (a: T, b: T) => a.name.localeCompare(b.name);
  if (opts.sortMode === 'category') {
    return [...filtered].sort((a, b) => (a.category || '').localeCompare(b.category || '') || byName(a, b));
  }
  if (opts.sortMode === 'effect') {
    return [...filtered].sort((a, b) => effectScore(b) - effectScore(a) || byName(a, b));
  }
  return [...filtered].sort(byName);
}
