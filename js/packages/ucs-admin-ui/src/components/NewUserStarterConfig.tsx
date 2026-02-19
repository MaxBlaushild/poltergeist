import React, { useEffect, useMemo, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { useInventory } from '@poltergeist/contexts';

type SelectOption = {
  value: string;
  label: string;
};

const SearchableSelect = ({
  label,
  placeholder,
  options,
  value,
  onChange,
}: {
  label: string;
  placeholder: string;
  options: SelectOption[];
  value: string;
  onChange: (value: string) => void;
}) => {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');

  const selected = options.find((o) => o.value === value);
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return options;
    return options.filter((o) => o.label.toLowerCase().includes(q));
  }, [options, query]);

  const displayValue = open ? query : selected?.label ?? '';

  return (
    <div className="relative">
      <label className="block text-sm font-medium text-gray-700">{label}</label>
      <input
        value={displayValue}
        onChange={(e) => {
          setQuery(e.target.value);
          setOpen(true);
        }}
        onFocus={() => {
          setOpen(true);
          setQuery('');
        }}
        onBlur={() => {
          setTimeout(() => setOpen(false), 150);
        }}
        placeholder={placeholder}
        className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
      />
      {open && (
        <div className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md border border-gray-200 bg-white shadow-lg">
          {filtered.length === 0 && (
            <div className="px-3 py-2 text-sm text-gray-500">No matches found</div>
          )}
          {filtered.map((option) => (
            <button
              type="button"
              key={option.value}
              onMouseDown={(e) => e.preventDefault()}
              onClick={() => {
                onChange(option.value);
                setOpen(false);
                setQuery('');
              }}
              className="flex w-full items-center px-3 py-2 text-left text-sm hover:bg-indigo-50"
            >
              <span className="font-medium text-gray-900">{option.label}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

type StarterItemRow = {
  id: string;
  inventoryItemId: string;
  quantity: number;
};

const makeRow = (inventoryItemId = '', quantity = 1): StarterItemRow => ({
  id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
  inventoryItemId,
  quantity,
});

export const NewUserStarterConfig = () => {
  const { apiClient } = useAPI();
  const { inventoryItems } = useInventory();
  const [gold, setGold] = useState(0);
  const [items, setItems] = useState<StarterItemRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [statusKind, setStatusKind] = useState<'success' | 'error' | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    const load = async () => {
      try {
        setLoading(true);
        const data = await apiClient.get('/sonar/admin/new-user-starter-config');
        const loadedGold = typeof data?.gold === 'number' ? data.gold : 0;
        const loadedItems = Array.isArray(data?.items) ? data.items : [];
        setGold(loadedGold);
        setItems(
          loadedItems.map((item: { inventoryItemId: number; quantity: number }) =>
            makeRow(String(item.inventoryItemId), item.quantity ?? 1),
          ),
        );
      } catch (error) {
        console.error('Failed to load starter config', error);
        setStatusMessage('Failed to load starter config.');
        setStatusKind('error');
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [apiClient]);

  const itemOptions = useMemo(() => {
    return (inventoryItems ?? []).map((item) => ({
      value: String(item.id),
      label: item.name,
    }));
  }, [inventoryItems]);

  const updateRow = (id: string, updates: Partial<StarterItemRow>) => {
    setItems((prev) => prev.map((row) => (row.id === id ? { ...row, ...updates } : row)));
  };

  const removeRow = (id: string) => {
    setItems((prev) => prev.filter((row) => row.id !== id));
  };

  const handleSave = async () => {
    try {
      setSaving(true);
      setStatusMessage(null);
      setStatusKind(null);

      const payloadItems = items
        .filter((row) => row.inventoryItemId && row.quantity > 0)
        .map((row) => ({
          inventoryItemId: parseInt(row.inventoryItemId),
          quantity: row.quantity,
        }));

      await apiClient.put('/sonar/admin/new-user-starter-config', {
        gold,
        items: payloadItems,
      });

      setStatusMessage('Starter config saved.');
      setStatusKind('success');
    } catch (error) {
      console.error('Failed to save starter config', error);
      setStatusMessage('Failed to save starter config.');
      setStatusKind('error');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="px-4 py-6">
      <div className="mx-auto flex w-full max-w-2xl flex-col gap-6 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
        <div>
          <h1 className="text-xl font-semibold text-gray-900">New User Starter Config</h1>
          <p className="mt-1 text-sm text-gray-500">
            Configure the gold and items new users receive on account creation.
          </p>
        </div>

        {loading ? (
          <div className="text-sm text-gray-500">Loading…</div>
        ) : (
          <>
            <div>
              <label className="block text-sm font-medium text-gray-700">Starting Gold</label>
              <input
                type="number"
                min="0"
                value={gold}
                onChange={(e) => setGold(parseInt(e.target.value || '0'))}
                className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
              />
            </div>

            <div className="flex flex-col gap-3">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-sm font-semibold text-gray-800">Starter Items</h2>
                  <p className="text-xs text-gray-500">Add one or more items with quantities.</p>
                </div>
                <button
                  type="button"
                  onClick={() => setItems((prev) => [...prev, makeRow()])}
                  className="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50"
                >
                  Add Item
                </button>
              </div>

              {items.length === 0 && (
                <div className="rounded-md border border-dashed border-gray-300 bg-gray-50 px-3 py-4 text-sm text-gray-500">
                  No starter items configured yet.
                </div>
              )}

              {items.map((row) => (
                <div
                  key={row.id}
                  className="flex flex-col gap-3 rounded-md border border-gray-200 bg-gray-50 p-3"
                >
                  <SearchableSelect
                    label="Inventory Item"
                    placeholder="Search item name…"
                    options={itemOptions}
                    value={row.inventoryItemId}
                    onChange={(value) => updateRow(row.id, { inventoryItemId: value })}
                  />
                  <div className="flex items-end justify-between gap-3">
                    <div className="flex-1">
                      <label className="block text-sm font-medium text-gray-700">Quantity</label>
                      <input
                        type="number"
                        min="1"
                        value={row.quantity}
                        onChange={(e) =>
                          updateRow(row.id, { quantity: parseInt(e.target.value || '1') })
                        }
                        className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                      />
                    </div>
                    <button
                      type="button"
                      onClick={() => removeRow(row.id)}
                      className="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-600 hover:bg-gray-100"
                    >
                      Remove
                    </button>
                  </div>
                </div>
              ))}
            </div>

            {statusMessage && (
              <div
                className={`rounded-md border px-3 py-2 text-sm ${
                  statusKind === 'success'
                    ? 'border-emerald-200 bg-emerald-50 text-emerald-800'
                    : 'border-rose-200 bg-rose-50 text-rose-800'
                }`}
              >
                {statusMessage}
              </div>
            )}

            <button
              onClick={handleSave}
              disabled={saving}
              className="inline-flex justify-center rounded-md border border-transparent bg-indigo-600 py-2 px-4 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {saving ? 'Saving…' : 'Save Starter Config'}
            </button>
          </>
        )}
      </div>
    </div>
  );
};

export default NewUserStarterConfig;
