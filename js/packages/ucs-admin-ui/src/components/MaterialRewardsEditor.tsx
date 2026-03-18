import React from 'react';

export type MaterialRewardForm = {
  resourceKey: string;
  amount: number;
};

export const BASE_MATERIAL_OPTIONS: Array<{
  value: string;
  label: string;
}> = [
  { value: 'timber', label: 'Timber' },
  { value: 'stone', label: 'Stone' },
  { value: 'iron', label: 'Iron' },
  { value: 'herbs', label: 'Herbs' },
  { value: 'monster_parts', label: 'Monster Parts' },
  { value: 'arcane_dust', label: 'Arcane Dust' },
  { value: 'relic_shards', label: 'Relic Shards' },
];

export const emptyMaterialReward = (): MaterialRewardForm => ({
  resourceKey: BASE_MATERIAL_OPTIONS[0].value,
  amount: 1,
});

export const normalizeMaterialRewards = (
  rewards: MaterialRewardForm[]
): MaterialRewardForm[] => {
  const totals = new Map<string, number>();
  rewards.forEach((reward) => {
    const key = reward.resourceKey.trim();
    const amount = Number.isFinite(reward.amount) ? reward.amount : 0;
    if (!key || amount <= 0) {
      return;
    }
    totals.set(key, (totals.get(key) ?? 0) + amount);
  });
  return BASE_MATERIAL_OPTIONS.map((option) => {
    const amount = totals.get(option.value) ?? 0;
    if (amount <= 0) {
      return null;
    }
    return {
      resourceKey: option.value,
      amount,
    };
  }).filter((reward): reward is MaterialRewardForm => reward !== null);
};

export const materialRewardLabel = (resourceKey: string): string =>
  BASE_MATERIAL_OPTIONS.find((option) => option.value === resourceKey)?.label ??
  resourceKey;

export const summarizeMaterialRewards = (
  rewards?: MaterialRewardForm[] | null
): string => {
  const normalized = normalizeMaterialRewards(rewards ?? []);
  if (normalized.length === 0) {
    return 'No materials';
  }
  return normalized
    .map((reward) => `${reward.amount} ${materialRewardLabel(reward.resourceKey)}`)
    .join(' · ');
};

type MaterialRewardsEditorProps = {
  value: MaterialRewardForm[];
  onChange: (next: MaterialRewardForm[]) => void;
  disabled?: boolean;
  title?: string;
};

export const MaterialRewardsEditor = ({
  value,
  onChange,
  disabled = false,
  title = 'Material Rewards',
}: MaterialRewardsEditorProps) => {
  const addReward = () => {
    onChange([...value, emptyMaterialReward()]);
  };

  const updateReward = (
    index: number,
    next: Partial<MaterialRewardForm>
  ) => {
    onChange(
      value.map((reward, rewardIndex) =>
        rewardIndex === index ? { ...reward, ...next } : reward
      )
    );
  };

  const removeReward = (index: number) => {
    onChange(value.filter((_, rewardIndex) => rewardIndex !== index));
  };

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <label className="block text-sm font-medium">{title}</label>
        <button
          type="button"
          className="rounded border px-3 py-1 text-sm disabled:opacity-50"
          onClick={addReward}
          disabled={disabled}
        >
          Add Material
        </button>
      </div>
      {value.length === 0 ? (
        <div className="rounded border border-dashed px-3 py-3 text-sm text-gray-500">
          No material rewards configured.
        </div>
      ) : (
        value.map((reward, index) => (
          <div
            key={`${reward.resourceKey}-${index}`}
            className="grid gap-3 rounded border p-3 md:grid-cols-[minmax(0,1fr)_120px_auto]"
          >
            <label className="block text-sm">
              Material
              <select
                className="mt-1 w-full rounded border p-2"
                value={reward.resourceKey}
                disabled={disabled}
                onChange={(event) =>
                  updateReward(index, { resourceKey: event.target.value })
                }
              >
                {BASE_MATERIAL_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
            <label className="block text-sm">
              Amount
              <input
                className="mt-1 w-full rounded border p-2"
                type="number"
                min={1}
                step={1}
                value={reward.amount}
                disabled={disabled}
                onChange={(event) =>
                  updateReward(index, {
                    amount: Number.parseInt(event.target.value, 10) || 0,
                  })
                }
              />
            </label>
            <div className="flex items-end">
              <button
                type="button"
                className="rounded border border-red-300 px-3 py-2 text-sm text-red-700 disabled:opacity-50"
                onClick={() => removeReward(index)}
                disabled={disabled}
              >
                Remove
              </button>
            </div>
          </div>
        ))
      )}
    </div>
  );
};
