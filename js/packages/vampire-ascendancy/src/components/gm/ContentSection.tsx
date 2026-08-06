import { useEffect, useState } from 'react';
import { gmListLibraryItems, gmSetItemIncluded } from '../../gmApi';
import type { GMLibraryItem } from '../../gmApi';
import { ApiError } from '../../api';
import { Card } from './GameSection';
import { ItemBrowser } from './ItemBrowser';

// Content tab: choose which of the shared story's items are included in
// this Toast. (Which characters are "in" is implicit now — see the Invites
// tab: a character is in play once someone's been invited to it. The quiz
// is fixed — it's the mystery's payoff and isn't meant to be trimmed per
// event, so it isn't customizable here at all.) Toggling off an item
// that's still assigned to a player is blocked (409) rather than silently
// allowed — the error names exactly what's blocking it.
export const ContentSection = () => {
  return (
    <div className="flex flex-col gap-4">
      <p className="text-bone/50 text-sm">
        Trim items to fit your event. Everything starts included; toggling something off that's still
        assigned to a player is blocked — reassign it first.
      </p>
      <ItemsPanel />
    </div>
  );
};

const ItemsPanel = () => {
  const [rows, setRows] = useState<GMLibraryItem[] | null>(null);
  const [error, setError] = useState<Record<string, string>>({});

  const load = () => gmListLibraryItems().then((d) => setRows(d.items));
  useEffect(() => {
    load();
  }, []);

  const toggle = async (row: GMLibraryItem) => {
    setError((e) => ({ ...e, [row.id]: '' }));
    try {
      await gmSetItemIncluded(row.id, !row.included);
      load();
    } catch (err) {
      setError((e) => ({ ...e, [row.id]: err instanceof ApiError ? err.message : 'Could not save.' }));
    }
  };

  if (!rows) return <Card title="Items">Loading…</Card>;
  const included = rows.filter((r) => r.included).length;

  return (
    <Card title={`Items (${included}/${rows.length} included)`}>
      <ItemBrowser
        items={rows}
        renderTrailing={(row) => (
          <div className="flex flex-col items-end gap-1">
            <label className="flex items-center gap-1.5 text-xs text-bone/70 whitespace-nowrap">
              <input type="checkbox" checked={row.included} onChange={() => toggle(row)} />
              Included
            </label>
            {error[row.id] && <p className="text-blood-bright text-[11px] text-right">{error[row.id]}</p>}
          </div>
        )}
      />
    </Card>
  );
};
