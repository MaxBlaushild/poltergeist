import { useAPI, useInventory, useZoneContext } from '@poltergeist/contexts';
import React, { useCallback, useEffect, useMemo, useState } from 'react';

type ScenarioRewardItem = {
  inventoryItemId: number;
  quantity: number;
};

type ScenarioOption = {
  id?: string;
  optionText: string;
  statTag: string;
  proficiencies: string[];
  difficulty?: number | null;
  rewardExperience: number;
  rewardGold: number;
  itemRewards: ScenarioRewardItem[];
};

type ScenarioRecord = {
  id: string;
  zoneId: string;
  latitude: number;
  longitude: number;
  prompt: string;
  imageUrl: string;
  thumbnailUrl: string;
  difficulty: number;
  rewardExperience: number;
  rewardGold: number;
  openEnded: boolean;
  options: ScenarioOption[];
  itemRewards: ScenarioRewardItem[];
  attemptedByUser?: boolean;
};

type ScenarioFormState = {
  zoneId: string;
  latitude: string;
  longitude: string;
  prompt: string;
  imageUrl: string;
  thumbnailUrl: string;
  difficulty: string;
  openEnded: boolean;
  rewardExperience: string;
  rewardGold: string;
  options: ScenarioOption[];
  itemRewards: ScenarioRewardItem[];
};

const statTags = [
  'strength',
  'dexterity',
  'constitution',
  'intelligence',
  'wisdom',
  'charisma',
] as const;

const emptyOption = (): ScenarioOption => ({
  optionText: '',
  statTag: 'charisma',
  proficiencies: [],
  difficulty: null,
  rewardExperience: 0,
  rewardGold: 0,
  itemRewards: [],
});

const emptyFormState = (): ScenarioFormState => ({
  zoneId: '',
  latitude: '',
  longitude: '',
  prompt: '',
  imageUrl: '',
  thumbnailUrl: '',
  difficulty: '24',
  openEnded: false,
  rewardExperience: '0',
  rewardGold: '0',
  options: [emptyOption()],
  itemRewards: [],
});

const parseIntValue = (value: string, fallback = 0): number => {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parseFloatValue = (value: string, fallback = 0): number => {
  const parsed = Number.parseFloat(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parseCsv = (value: string): string[] => {
  return value
    .split(',')
    .map((part) => part.trim())
    .filter(Boolean);
};

export const Scenarios = () => {
  const { apiClient } = useAPI();
  const { zones } = useZoneContext();
  const { inventoryItems } = useInventory();

  const [loading, setLoading] = useState(true);
  const [records, setRecords] = useState<ScenarioRecord[]>([]);
  const [query, setQuery] = useState('');
  const [error, setError] = useState<string | null>(null);

  const [showModal, setShowModal] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<ScenarioFormState>(emptyFormState);

  const [deleteId, setDeleteId] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.get<ScenarioRecord[]>('/sonar/scenarios');
      setRecords(response);
    } catch (err) {
      console.error('Error loading scenarios:', err);
      setError('Failed to load scenarios.');
    } finally {
      setLoading(false);
    }
  }, [apiClient]);

  useEffect(() => {
    void load();
  }, [load]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return records;
    return records.filter((record) => {
      const zoneName = zones.find((zone) => zone.id === record.zoneId)?.name ?? '';
      return (
        record.prompt.toLowerCase().includes(q) ||
        zoneName.toLowerCase().includes(q) ||
        record.id.toLowerCase().includes(q)
      );
    });
  }, [query, records, zones]);

  const openCreate = () => {
    setEditingId(null);
    setForm(emptyFormState());
    setShowModal(true);
  };

  const openEdit = (record: ScenarioRecord) => {
    setEditingId(record.id);
    setForm({
      zoneId: record.zoneId,
      latitude: record.latitude.toString(),
      longitude: record.longitude.toString(),
      prompt: record.prompt,
      imageUrl: record.imageUrl,
      thumbnailUrl: record.thumbnailUrl ?? '',
      difficulty: record.difficulty.toString(),
      openEnded: record.openEnded,
      rewardExperience: record.rewardExperience.toString(),
      rewardGold: record.rewardGold.toString(),
      options: record.options.length > 0 ? record.options : [emptyOption()],
      itemRewards: record.itemRewards,
    });
    setShowModal(true);
  };

  const closeModal = () => {
    setShowModal(false);
    setEditingId(null);
    setForm(emptyFormState());
  };

  const formPayload = () => ({
    zoneId: form.zoneId,
    latitude: parseFloatValue(form.latitude),
    longitude: parseFloatValue(form.longitude),
    prompt: form.prompt.trim(),
    imageUrl: form.imageUrl.trim(),
    thumbnailUrl: form.thumbnailUrl.trim(),
    difficulty: parseIntValue(form.difficulty, 24),
    openEnded: form.openEnded,
    rewardExperience: form.openEnded ? parseIntValue(form.rewardExperience) : 0,
    rewardGold: form.openEnded ? parseIntValue(form.rewardGold) : 0,
    options: form.openEnded
      ? []
      : form.options.map((option) => ({
          optionText: option.optionText.trim(),
          statTag: option.statTag,
          proficiencies: option.proficiencies,
          difficulty: option.difficulty,
          rewardExperience: option.rewardExperience,
          rewardGold: option.rewardGold,
          itemRewards: option.itemRewards,
        })),
    itemRewards: form.openEnded ? form.itemRewards : [],
  });

  const save = async () => {
    try {
      const payload = formPayload();
      if (!payload.zoneId || !payload.prompt || !payload.imageUrl || !payload.thumbnailUrl) {
        alert('Zone, prompt, image URL, and thumbnail URL are required.');
        return;
      }
      if (payload.openEnded === false && payload.options.length === 0) {
        alert('Non-open-ended scenarios need at least one option.');
        return;
      }

      if (editingId) {
        const updated = await apiClient.put<ScenarioRecord>(`/sonar/scenarios/${editingId}`, payload);
        setRecords((prev) => prev.map((record) => (record.id === updated.id ? updated : record)));
      } else {
        const created = await apiClient.post<ScenarioRecord>('/sonar/scenarios', payload);
        setRecords((prev) => [created, ...prev]);
      }
      closeModal();
    } catch (err) {
      console.error('Error saving scenario:', err);
      alert('Failed to save scenario. Check required fields and try again.');
    }
  };

  const confirmDelete = async () => {
    if (!deleteId) return;
    try {
      await apiClient.delete(`/sonar/scenarios/${deleteId}`);
      setRecords((prev) => prev.filter((record) => record.id !== deleteId));
      setDeleteId(null);
    } catch (err) {
      console.error('Error deleting scenario:', err);
      alert('Failed to delete scenario.');
    }
  };

  const updateOption = (index: number, next: Partial<ScenarioOption>) => {
    setForm((prev) => {
      const options = [...prev.options];
      options[index] = { ...options[index], ...next };
      return { ...prev, options };
    });
  };

  const addOption = () => {
    setForm((prev) => ({ ...prev, options: [...prev.options, emptyOption()] }));
  };

  const removeOption = (index: number) => {
    setForm((prev) => ({
      ...prev,
      options: prev.options.filter((_, i) => i !== index),
    }));
  };

  const addOptionItemReward = (optionIndex: number) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.itemRewards = [...option.itemRewards, { inventoryItemId: 0, quantity: 1 }];
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const updateOptionItemReward = (
    optionIndex: number,
    rewardIndex: number,
    next: Partial<ScenarioRewardItem>
  ) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      const rewards = [...option.itemRewards];
      rewards[rewardIndex] = { ...rewards[rewardIndex], ...next };
      option.itemRewards = rewards;
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const removeOptionItemReward = (optionIndex: number, rewardIndex: number) => {
    setForm((prev) => {
      const options = [...prev.options];
      const option = options[optionIndex];
      option.itemRewards = option.itemRewards.filter((_, i) => i !== rewardIndex);
      options[optionIndex] = option;
      return { ...prev, options };
    });
  };

  const addScenarioItemReward = () => {
    setForm((prev) => ({
      ...prev,
      itemRewards: [...prev.itemRewards, { inventoryItemId: 0, quantity: 1 }],
    }));
  };

  const updateScenarioItemReward = (index: number, next: Partial<ScenarioRewardItem>) => {
    setForm((prev) => {
      const rewards = [...prev.itemRewards];
      rewards[index] = { ...rewards[index], ...next };
      return { ...prev, itemRewards: rewards };
    });
  };

  const removeScenarioItemReward = (index: number) => {
    setForm((prev) => ({
      ...prev,
      itemRewards: prev.itemRewards.filter((_, i) => i !== index),
    }));
  };

  if (loading) {
    return <div className="m-10">Loading scenarios...</div>;
  }

  return (
    <div className="m-10">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Scenarios</h1>
        <button className="bg-blue-500 text-white px-4 py-2 rounded-md" onClick={openCreate}>
          Create Scenario
        </button>
      </div>

      {error && <div className="mb-3 text-red-600">{error}</div>}

      <div className="mb-4">
        <input
          type="text"
          placeholder="Search scenarios by prompt, zone, or ID..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="w-full p-2 border rounded-md"
        />
      </div>

      <div className="grid gap-4" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))' }}>
        {filtered.map((record) => {
          const zoneName = zones.find((zone) => zone.id === record.zoneId)?.name ?? record.zoneId;
          return (
            <div key={record.id} className="border rounded-md p-4 bg-white shadow-sm">
              <div className="text-xs text-gray-500 mb-1">{record.id}</div>
              <div className="font-semibold mb-2">{record.openEnded ? 'Open-Ended' : 'Choice'} Scenario</div>
              <div className="text-sm text-gray-700 mb-1">Zone: {zoneName}</div>
              <div className="text-sm text-gray-700 mb-1">
                Location: {record.latitude.toFixed(5)}, {record.longitude.toFixed(5)}
              </div>
              <div className="text-sm text-gray-700 mb-2">Difficulty: {record.difficulty}</div>
              <div className="text-sm text-gray-800 mb-3 line-clamp-3">{record.prompt}</div>
              <div className="flex gap-2">
                <button className="bg-blue-500 text-white px-3 py-1 rounded-md" onClick={() => openEdit(record)}>
                  Edit
                </button>
                <button className="bg-red-500 text-white px-3 py-1 rounded-md" onClick={() => setDeleteId(record.id)}>
                  Delete
                </button>
              </div>
            </div>
          );
        })}
      </div>

      {showModal && (
        <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-md p-6 w-full max-w-5xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-semibold mb-4">{editingId ? 'Edit Scenario' : 'Create Scenario'}</h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-4">
              <label className="text-sm">
                Zone
                <select
                  value={form.zoneId}
                  onChange={(e) => setForm((prev) => ({ ...prev, zoneId: e.target.value }))}
                  className="w-full border rounded-md p-2"
                >
                  <option value="">Select zone</option>
                  {zones.map((zone) => (
                    <option key={zone.id} value={zone.id}>
                      {zone.name}
                    </option>
                  ))}
                </select>
              </label>
              <label className="text-sm">
                Difficulty
                <input
                  value={form.difficulty}
                  onChange={(e) => setForm((prev) => ({ ...prev, difficulty: e.target.value }))}
                  className="w-full border rounded-md p-2"
                  type="number"
                  min={0}
                />
              </label>
              <label className="text-sm">
                Latitude
                <input
                  value={form.latitude}
                  onChange={(e) => setForm((prev) => ({ ...prev, latitude: e.target.value }))}
                  className="w-full border rounded-md p-2"
                  type="number"
                  step="any"
                />
              </label>
              <label className="text-sm">
                Longitude
                <input
                  value={form.longitude}
                  onChange={(e) => setForm((prev) => ({ ...prev, longitude: e.target.value }))}
                  className="w-full border rounded-md p-2"
                  type="number"
                  step="any"
                />
              </label>
              <label className="text-sm md:col-span-2">
                Image URL
                <input
                  value={form.imageUrl}
                  onChange={(e) => setForm((prev) => ({ ...prev, imageUrl: e.target.value }))}
                  className="w-full border rounded-md p-2"
                />
              </label>
              <label className="text-sm md:col-span-2">
                Thumbnail URL
                <input
                  value={form.thumbnailUrl}
                  onChange={(e) => setForm((prev) => ({ ...prev, thumbnailUrl: e.target.value }))}
                  className="w-full border rounded-md p-2"
                />
              </label>
            </div>

            <label className="text-sm block mb-4">
              Prompt
              <textarea
                value={form.prompt}
                onChange={(e) => setForm((prev) => ({ ...prev, prompt: e.target.value }))}
                className="w-full border rounded-md p-2 min-h-[90px]"
              />
            </label>

            <label className="inline-flex items-center gap-2 mb-4">
              <input
                type="checkbox"
                checked={form.openEnded}
                onChange={(e) =>
                  setForm((prev) => ({
                    ...prev,
                    openEnded: e.target.checked,
                    options: e.target.checked ? [emptyOption()] : prev.options,
                  }))
                }
              />
              Open-ended scenario (freeform response)
            </label>

            {form.openEnded ? (
              <div className="border rounded-md p-3 mb-4">
                <div className="font-medium mb-2">Scenario Rewards</div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-3">
                  <label className="text-sm">
                    Reward Experience
                    <input
                      value={form.rewardExperience}
                      onChange={(e) => setForm((prev) => ({ ...prev, rewardExperience: e.target.value }))}
                      className="w-full border rounded-md p-2"
                      type="number"
                      min={0}
                    />
                  </label>
                  <label className="text-sm">
                    Reward Gold
                    <input
                      value={form.rewardGold}
                      onChange={(e) => setForm((prev) => ({ ...prev, rewardGold: e.target.value }))}
                      className="w-full border rounded-md p-2"
                      type="number"
                      min={0}
                    />
                  </label>
                </div>
                <div className="flex items-center justify-between mb-2">
                  <div className="font-medium">Item Rewards</div>
                  <button className="bg-green-600 text-white px-3 py-1 rounded-md" type="button" onClick={addScenarioItemReward}>
                    Add Item
                  </button>
                </div>
                {form.itemRewards.map((reward, index) => (
                  <div key={index} className="grid grid-cols-1 md:grid-cols-3 gap-2 mb-2">
                    <select
                      value={reward.inventoryItemId}
                      onChange={(e) =>
                        updateScenarioItemReward(index, {
                          inventoryItemId: parseIntValue(e.target.value, 0),
                        })
                      }
                      className="border rounded-md p-2"
                    >
                      <option value={0}>Select item</option>
                      {inventoryItems.map((item) => (
                        <option key={item.id} value={item.id}>
                          {item.name}
                        </option>
                      ))}
                    </select>
                    <input
                      value={reward.quantity}
                      onChange={(e) =>
                        updateScenarioItemReward(index, {
                          quantity: parseIntValue(e.target.value, 1),
                        })
                      }
                      className="border rounded-md p-2"
                      type="number"
                      min={1}
                    />
                    <button
                      type="button"
                      className="bg-red-500 text-white px-3 py-1 rounded-md"
                      onClick={() => removeScenarioItemReward(index)}
                    >
                      Remove
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <div className="border rounded-md p-3 mb-4">
                <div className="flex items-center justify-between mb-3">
                  <div className="font-medium">Response Options</div>
                  <button className="bg-green-600 text-white px-3 py-1 rounded-md" type="button" onClick={addOption}>
                    Add Option
                  </button>
                </div>
                {form.options.map((option, optionIndex) => (
                  <div key={optionIndex} className="border rounded-md p-3 mb-3">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      <label className="text-sm md:col-span-2">
                        Option Text
                        <input
                          value={option.optionText}
                          onChange={(e) => updateOption(optionIndex, { optionText: e.target.value })}
                          className="w-full border rounded-md p-2"
                        />
                      </label>
                      <label className="text-sm">
                        Stat Tag
                        <select
                          value={option.statTag}
                          onChange={(e) => updateOption(optionIndex, { statTag: e.target.value })}
                          className="w-full border rounded-md p-2"
                        >
                          {statTags.map((tag) => (
                            <option key={tag} value={tag}>
                              {tag}
                            </option>
                          ))}
                        </select>
                      </label>
                      <label className="text-sm">
                        Difficulty Override (optional)
                        <input
                          value={option.difficulty ?? ''}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              difficulty: e.target.value === '' ? null : parseIntValue(e.target.value, 0),
                            })
                          }
                          className="w-full border rounded-md p-2"
                          type="number"
                          min={0}
                        />
                      </label>
                      <label className="text-sm md:col-span-2">
                        Proficiencies (comma separated)
                        <input
                          value={option.proficiencies.join(', ')}
                          onChange={(e) =>
                            updateOption(optionIndex, { proficiencies: parseCsv(e.target.value) })
                          }
                          className="w-full border rounded-md p-2"
                        />
                      </label>
                      <label className="text-sm">
                        Reward Experience
                        <input
                          value={option.rewardExperience}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              rewardExperience: parseIntValue(e.target.value, 0),
                            })
                          }
                          className="w-full border rounded-md p-2"
                          type="number"
                          min={0}
                        />
                      </label>
                      <label className="text-sm">
                        Reward Gold
                        <input
                          value={option.rewardGold}
                          onChange={(e) =>
                            updateOption(optionIndex, {
                              rewardGold: parseIntValue(e.target.value, 0),
                            })
                          }
                          className="w-full border rounded-md p-2"
                          type="number"
                          min={0}
                        />
                      </label>
                    </div>

                    <div className="mt-3">
                      <div className="flex items-center justify-between mb-2">
                        <div className="font-medium text-sm">Option Item Rewards</div>
                        <button
                          type="button"
                          className="bg-green-600 text-white px-2 py-1 rounded-md text-xs"
                          onClick={() => addOptionItemReward(optionIndex)}
                        >
                          Add Item
                        </button>
                      </div>
                      {option.itemRewards.map((reward, rewardIndex) => (
                        <div key={rewardIndex} className="grid grid-cols-1 md:grid-cols-3 gap-2 mb-2">
                          <select
                            value={reward.inventoryItemId}
                            onChange={(e) =>
                              updateOptionItemReward(optionIndex, rewardIndex, {
                                inventoryItemId: parseIntValue(e.target.value, 0),
                              })
                            }
                            className="border rounded-md p-2"
                          >
                            <option value={0}>Select item</option>
                            {inventoryItems.map((item) => (
                              <option key={item.id} value={item.id}>
                                {item.name}
                              </option>
                            ))}
                          </select>
                          <input
                            value={reward.quantity}
                            onChange={(e) =>
                              updateOptionItemReward(optionIndex, rewardIndex, {
                                quantity: parseIntValue(e.target.value, 1),
                              })
                            }
                            className="border rounded-md p-2"
                            type="number"
                            min={1}
                          />
                          <button
                            type="button"
                            className="bg-red-500 text-white px-3 py-1 rounded-md"
                            onClick={() => removeOptionItemReward(optionIndex, rewardIndex)}
                          >
                            Remove
                          </button>
                        </div>
                      ))}
                    </div>

                    <div className="mt-3">
                      <button
                        type="button"
                        className="bg-red-500 text-white px-3 py-1 rounded-md"
                        onClick={() => removeOption(optionIndex)}
                      >
                        Remove Option
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}

            <div className="flex gap-2">
              <button className="bg-blue-500 text-white px-4 py-2 rounded-md" onClick={save}>
                {editingId ? 'Update Scenario' : 'Create Scenario'}
              </button>
              <button className="bg-gray-500 text-white px-4 py-2 rounded-md" onClick={closeModal}>
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {deleteId && (
        <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-md p-6 max-w-sm w-full">
            <h3 className="text-lg font-semibold mb-2">Delete Scenario</h3>
            <p className="text-sm text-gray-700 mb-4">This action cannot be undone.</p>
            <div className="flex gap-2">
              <button className="bg-red-500 text-white px-4 py-2 rounded-md" onClick={confirmDelete}>
                Delete
              </button>
              <button className="bg-gray-500 text-white px-4 py-2 rounded-md" onClick={() => setDeleteId(null)}>
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
