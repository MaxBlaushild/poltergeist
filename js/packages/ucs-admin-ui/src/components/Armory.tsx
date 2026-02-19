import React, { useMemo, useState } from 'react';
import { useUsers } from '../hooks/useUsers.ts';
import { useInventory } from '@poltergeist/contexts';
import { useAPI } from '@poltergeist/contexts';

type SelectOption = {
  value: string;
  label: string;
  secondary?: string;
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
    return options.filter((o) => {
      const hay = `${o.label} ${o.secondary ?? ''}`.toLowerCase();
      return hay.includes(q);
    });
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
              className="flex w-full flex-col items-start px-3 py-2 text-left text-sm hover:bg-indigo-50"
            >
              <span className="font-medium text-gray-900">{option.label}</span>
              {option.secondary && (
                <span className="text-xs text-gray-500">{option.secondary}</span>
              )}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export const Armory = () => {
  const { users } = useUsers();
  const [selectedUser, setSelectedUser] = useState('');
  const [selectedItem, setSelectedItem] = useState('');
  const { inventoryItems } = useInventory();
  const [quantity, setQuantity] = useState(1);
  const { apiClient } = useAPI();
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [statusKind, setStatusKind] = useState<'success' | 'error' | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async () => {
    try {
      setSubmitting(true);
      setStatusMessage(null);
      setStatusKind(null);
      await apiClient.post('/sonar/users/giveItem', {
        userID: selectedUser,
        itemID: parseInt(selectedItem),
        quantity: quantity
      });
      setStatusMessage('Item granted successfully.');
      setStatusKind('success');
    } catch (error) {
      setStatusMessage('Failed to grant item. Please try again.');
      setStatusKind('error');
      console.error('Failed to give item', error);
    } finally {
      setSubmitting(false);
    }
  };

  const userOptions = useMemo(() => {
    return (users ?? []).map((user) => {
      const username = user.username?.trim() ? `@${user.username}` : '';
      const display = username || user.name || user.phoneNumber;
      const secondary = username ? user.name : user.phoneNumber;
      return {
        value: user.id,
        label: display,
        secondary: secondary && secondary !== display ? secondary : undefined,
      };
    });
  }, [users]);

  const itemOptions = useMemo(() => {
    return (inventoryItems ?? []).map((item) => ({
      value: String(item.id),
      label: item.name,
    }));
  }, [inventoryItems]);

  return (
    <div className="px-4 py-6">
      <div className="mx-auto flex w-full max-w-xl flex-col gap-5 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
        <SearchableSelect
          label="Select User"
          placeholder="Search by username or name…"
          options={userOptions}
          value={selectedUser}
          onChange={setSelectedUser}
        />

        <SearchableSelect
          label="Select Item"
          placeholder="Search item name…"
          options={itemOptions}
          value={selectedItem}
          onChange={setSelectedItem}
        />

        <div>
          <label className="block text-sm font-medium text-gray-700">Quantity</label>
          <input
            type="number"
            min="1"
            value={quantity}
            onChange={(e) => setQuantity(parseInt(e.target.value))}
            className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          />
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
          onClick={handleSubmit}
          disabled={!selectedUser || !selectedItem || submitting}
          className="inline-flex justify-center rounded-md border border-transparent bg-indigo-600 py-2 px-4 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {submitting ? 'Granting…' : 'Submit'}
        </button>
      </div>
    </div>
  );
};

export default Armory;
