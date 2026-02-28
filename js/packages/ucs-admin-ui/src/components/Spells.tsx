import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { Spell, SpellEffect, SpellStatusTemplate } from '@poltergeist/types';

type SpellStatusTemplateForm = {
  name: string;
  description: string;
  effect: string;
  positive: boolean;
  durationSeconds: string;
  strengthMod: string;
  dexterityMod: string;
  constitutionMod: string;
  intelligenceMod: string;
  wisdomMod: string;
  charismaMod: string;
};

type SpellEffectForm = {
  type: string;
  customType: string;
  amount: string;
  statusesToApply: SpellStatusTemplateForm[];
  statusesToRemove: string;
  effectData: string;
};

type SpellFormState = {
  name: string;
  description: string;
  iconUrl: string;
  effectText: string;
  schoolOfMagic: string;
  manaCost: string;
  effects: SpellEffectForm[];
};

const knownEffectTypes = [
  'deal_damage',
  'restore_life_party_member',
  'restore_life_all_party_members',
  'apply_beneficial_statuses',
  'remove_detrimental_statuses',
] as const;

const emptyStatusTemplate = (): SpellStatusTemplateForm => ({
  name: '',
  description: '',
  effect: '',
  positive: true,
  durationSeconds: '60',
  strengthMod: '0',
  dexterityMod: '0',
  constitutionMod: '0',
  intelligenceMod: '0',
  wisdomMod: '0',
  charismaMod: '0',
});

const emptyEffect = (): SpellEffectForm => ({
  type: 'deal_damage',
  customType: '',
  amount: '0',
  statusesToApply: [],
  statusesToRemove: '',
  effectData: '',
});

const emptyForm = (): SpellFormState => ({
  name: '',
  description: '',
  iconUrl: '',
  effectText: '',
  schoolOfMagic: '',
  manaCost: '0',
  effects: [emptyEffect()],
});

const parseIntSafe = (value: string, fallback = 0): number => {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parseStatusTemplate = (
  template: SpellStatusTemplateForm
): SpellStatusTemplate | null => {
  const name = template.name.trim();
  const duration = parseIntSafe(template.durationSeconds, 0);
  if (!name || duration <= 0) return null;
  return {
    name,
    description: template.description.trim(),
    effect: template.effect.trim(),
    positive: template.positive,
    durationSeconds: duration,
    strengthMod: parseIntSafe(template.strengthMod, 0),
    dexterityMod: parseIntSafe(template.dexterityMod, 0),
    constitutionMod: parseIntSafe(template.constitutionMod, 0),
    intelligenceMod: parseIntSafe(template.intelligenceMod, 0),
    wisdomMod: parseIntSafe(template.wisdomMod, 0),
    charismaMod: parseIntSafe(template.charismaMod, 0),
  };
};

const parseEffectData = (raw: string): Record<string, unknown> | undefined => {
  const trimmed = raw.trim();
  if (!trimmed) return undefined;
  const parsed = JSON.parse(trimmed);
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    throw new Error('Effect data must be a JSON object');
  }
  return parsed as Record<string, unknown>;
};

const normalizeEffectType = (effect: SpellEffectForm): string => {
  if (effect.type === '__custom__') {
    return effect.customType.trim().toLowerCase();
  }
  return effect.type.trim().toLowerCase();
};

const formFromSpell = (spell: Spell): SpellFormState => {
  const effects =
    spell.effects?.length > 0
      ? spell.effects.map((effect) => {
          const rawType = (effect.type || '').toString().trim().toLowerCase();
          const isKnown = knownEffectTypes.includes(rawType as (typeof knownEffectTypes)[number]);
          return {
            type: isKnown ? rawType : '__custom__',
            customType: isKnown ? '' : rawType,
            amount:
              effect.amount !== undefined && effect.amount !== null
                ? String(effect.amount)
                : '0',
            statusesToApply: (effect.statusesToApply ?? []).map((status) => ({
              name: status.name ?? '',
              description: status.description ?? '',
              effect: status.effect ?? '',
              positive: status.positive ?? true,
              durationSeconds: String(status.durationSeconds ?? 60),
              strengthMod: String(status.strengthMod ?? 0),
              dexterityMod: String(status.dexterityMod ?? 0),
              constitutionMod: String(status.constitutionMod ?? 0),
              intelligenceMod: String(status.intelligenceMod ?? 0),
              wisdomMod: String(status.wisdomMod ?? 0),
              charismaMod: String(status.charismaMod ?? 0),
            })),
            statusesToRemove: (effect.statusesToRemove ?? []).join(', '),
            effectData: effect.effectData
              ? JSON.stringify(effect.effectData, null, 2)
              : '',
          };
        })
      : [emptyEffect()];

  return {
    name: spell.name ?? '',
    description: spell.description ?? '',
    iconUrl: spell.iconUrl ?? '',
    effectText: spell.effectText ?? '',
    schoolOfMagic: spell.schoolOfMagic ?? '',
    manaCost: String(spell.manaCost ?? 0),
    effects,
  };
};

const payloadFromForm = (form: SpellFormState) => {
  const effects: SpellEffect[] = form.effects.map((effect, index) => {
    const effectType = normalizeEffectType(effect);
    if (!effectType) {
      throw new Error(`Effect #${index + 1} requires a type`);
    }

    const statusesToApply = effect.statusesToApply
      .map(parseStatusTemplate)
      .filter((status): status is SpellStatusTemplate => status !== null);
    const statusesToRemove = effect.statusesToRemove
      .split(',')
      .map((value) => value.trim())
      .filter(Boolean);
    const effectData = parseEffectData(effect.effectData);

    return {
      type: effectType,
      amount: parseIntSafe(effect.amount, 0),
      statusesToApply,
      statusesToRemove,
      effectData,
    };
  });

  return {
    name: form.name.trim(),
    description: form.description.trim(),
    iconUrl: form.iconUrl.trim(),
    effectText: form.effectText.trim(),
    schoolOfMagic: form.schoolOfMagic.trim(),
    manaCost: parseIntSafe(form.manaCost, 0),
    effects,
  };
};

export const Spells = () => {
  const { apiClient } = useAPI();

  const [loading, setLoading] = useState(true);
  const [spells, setSpells] = useState<Spell[]>([]);
  const [search, setSearch] = useState('');
  const [error, setError] = useState<string | null>(null);

  const [showModal, setShowModal] = useState(false);
  const [editingSpell, setEditingSpell] = useState<Spell | null>(null);
  const [form, setForm] = useState<SpellFormState>(emptyForm());

  const [deleteId, setDeleteId] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.get<Spell[]>('/sonar/spells');
      setSpells(Array.isArray(response) ? response : []);
    } catch (err) {
      console.error('Failed to load spells', err);
      setError('Failed to load spells.');
    } finally {
      setLoading(false);
    }
  }, [apiClient]);

  useEffect(() => {
    void load();
  }, [load]);

  const filtered = useMemo(() => {
    const query = search.trim().toLowerCase();
    if (!query) return spells;
    return spells.filter((spell) => {
      return (
        spell.name.toLowerCase().includes(query) ||
        spell.schoolOfMagic.toLowerCase().includes(query) ||
        spell.effectText.toLowerCase().includes(query)
      );
    });
  }, [search, spells]);

  const openCreate = () => {
    setEditingSpell(null);
    setForm(emptyForm());
    setShowModal(true);
  };

  const openEdit = (spell: Spell) => {
    setEditingSpell(spell);
    setForm(formFromSpell(spell));
    setShowModal(true);
  };

  const closeModal = () => {
    setShowModal(false);
    setEditingSpell(null);
    setForm(emptyForm());
  };

  const addEffect = () => {
    setForm((prev) => ({ ...prev, effects: [...prev.effects, emptyEffect()] }));
  };

  const removeEffect = (index: number) => {
    setForm((prev) => ({
      ...prev,
      effects: prev.effects.filter((_, effectIndex) => effectIndex !== index),
    }));
  };

  const updateEffect = (index: number, next: Partial<SpellEffectForm>) => {
    setForm((prev) => {
      const effects = [...prev.effects];
      effects[index] = { ...effects[index], ...next };
      return { ...prev, effects };
    });
  };

  const addEffectStatus = (effectIndex: number) => {
    setForm((prev) => {
      const effects = [...prev.effects];
      effects[effectIndex] = {
        ...effects[effectIndex],
        statusesToApply: [...effects[effectIndex].statusesToApply, emptyStatusTemplate()],
      };
      return { ...prev, effects };
    });
  };

  const updateEffectStatus = (
    effectIndex: number,
    statusIndex: number,
    next: Partial<SpellStatusTemplateForm>
  ) => {
    setForm((prev) => {
      const effects = [...prev.effects];
      const statuses = [...effects[effectIndex].statusesToApply];
      statuses[statusIndex] = { ...statuses[statusIndex], ...next };
      effects[effectIndex] = { ...effects[effectIndex], statusesToApply: statuses };
      return { ...prev, effects };
    });
  };

  const removeEffectStatus = (effectIndex: number, statusIndex: number) => {
    setForm((prev) => {
      const effects = [...prev.effects];
      effects[effectIndex] = {
        ...effects[effectIndex],
        statusesToApply: effects[effectIndex].statusesToApply.filter((_, index) => index !== statusIndex),
      };
      return { ...prev, effects };
    });
  };

  const save = async () => {
    try {
      const payload = payloadFromForm(form);
      if (!payload.name || !payload.schoolOfMagic) {
        alert('Name and school of magic are required.');
        return;
      }

      if (editingSpell) {
        const updated = await apiClient.put<Spell>(`/sonar/spells/${editingSpell.id}`, payload);
        setSpells((prev) => prev.map((spell) => (spell.id === updated.id ? updated : spell)));
      } else {
        const created = await apiClient.post<Spell>('/sonar/spells', payload);
        setSpells((prev) => [created, ...prev]);
      }
      closeModal();
    } catch (err) {
      console.error('Failed to save spell', err);
      const message = err instanceof Error ? err.message : 'Failed to save spell.';
      alert(message);
    }
  };

  const confirmDelete = async () => {
    if (!deleteId) return;
    try {
      await apiClient.delete(`/sonar/spells/${deleteId}`);
      setSpells((prev) => prev.filter((spell) => spell.id !== deleteId));
      setDeleteId(null);
    } catch (err) {
      console.error('Failed to delete spell', err);
      alert('Failed to delete spell.');
    }
  };

  return (
    <div className="p-6 bg-gray-100 min-h-screen">
      <div className="max-w-7xl mx-auto">
        <div className="qa-card mb-6">
          <div className="flex items-center justify-between gap-3">
            <div>
              <h1 className="qa-card-title">Spells</h1>
              <p className="text-sm text-gray-600">Manage spell definitions and effect payloads.</p>
            </div>
            <button className="qa-btn qa-btn-primary" onClick={openCreate}>
              Create Spell
            </button>
          </div>
        </div>

        <div className="qa-card mb-6">
          <input
            className="block w-full border border-gray-300 rounded-md p-2"
            placeholder="Search by name, school, or effect text..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>

        {loading ? (
          <div className="qa-card">Loading spells...</div>
        ) : error ? (
          <div className="qa-card text-red-600">{error}</div>
        ) : filtered.length === 0 ? (
          <div className="qa-card text-gray-600">No spells found.</div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {filtered.map((spell) => (
              <div key={spell.id} className="qa-card">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <div className="text-lg font-semibold">{spell.name}</div>
                    <div className="text-sm text-gray-600">{spell.schoolOfMagic} · Mana {spell.manaCost}</div>
                  </div>
                  {spell.iconUrl ? (
                    <img src={spell.iconUrl} alt={spell.name} className="w-12 h-12 rounded-md object-cover border" />
                  ) : null}
                </div>
                {spell.description ? <p className="text-sm text-gray-700 mt-3">{spell.description}</p> : null}
                {spell.effectText ? <p className="text-sm text-gray-700 mt-2">{spell.effectText}</p> : null}
                <div className="text-xs text-gray-500 mt-2">Effects: {spell.effects?.length ?? 0}</div>
                <div className="flex items-center gap-2 mt-4">
                  <button className="qa-btn qa-btn-secondary" onClick={() => openEdit(spell)}>
                    Edit
                  </button>
                  <button className="qa-btn qa-btn-danger" onClick={() => setDeleteId(spell.id)}>
                    Delete
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}

        {showModal && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
            <div className="bg-white w-full max-w-5xl rounded-lg shadow-lg max-h-[92vh] overflow-y-auto">
              <div className="p-5 border-b flex items-center justify-between">
                <h2 className="text-xl font-semibold">
                  {editingSpell ? `Edit ${editingSpell.name}` : 'Create Spell'}
                </h2>
                <button className="text-gray-600 hover:text-gray-900" onClick={closeModal}>
                  Close
                </button>
              </div>

              <div className="p-5 space-y-5">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <label className="text-sm">
                    Name
                    <input
                      className="w-full border rounded-md p-2"
                      value={form.name}
                      onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
                    />
                  </label>
                  <label className="text-sm">
                    School of Magic
                    <input
                      className="w-full border rounded-md p-2"
                      value={form.schoolOfMagic}
                      onChange={(e) =>
                        setForm((prev) => ({ ...prev, schoolOfMagic: e.target.value }))
                      }
                    />
                  </label>
                  <label className="text-sm">
                    Icon URL
                    <input
                      className="w-full border rounded-md p-2"
                      value={form.iconUrl}
                      onChange={(e) => setForm((prev) => ({ ...prev, iconUrl: e.target.value }))}
                    />
                  </label>
                  <label className="text-sm">
                    Mana Cost
                    <input
                      className="w-full border rounded-md p-2"
                      type="number"
                      min={0}
                      value={form.manaCost}
                      onChange={(e) => setForm((prev) => ({ ...prev, manaCost: e.target.value }))}
                    />
                  </label>
                </div>

                <label className="text-sm block">
                  Description
                  <textarea
                    className="w-full border rounded-md p-2 min-h-[84px]"
                    value={form.description}
                    onChange={(e) => setForm((prev) => ({ ...prev, description: e.target.value }))}
                  />
                </label>

                <label className="text-sm block">
                  Effect Text
                  <textarea
                    className="w-full border rounded-md p-2 min-h-[84px]"
                    value={form.effectText}
                    onChange={(e) => setForm((prev) => ({ ...prev, effectText: e.target.value }))}
                  />
                </label>

                <div className="border rounded-md p-4">
                  <div className="flex items-center justify-between mb-3">
                    <div className="font-semibold">Effects</div>
                    <button className="qa-btn qa-btn-secondary" type="button" onClick={addEffect}>
                      Add Effect
                    </button>
                  </div>

                  <div className="space-y-3">
                    {form.effects.map((effect, effectIndex) => (
                      <div key={effectIndex} className="border rounded-md p-3 bg-gray-50">
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-3 mb-3">
                          <label className="text-sm">
                            Effect Type
                            <select
                              className="w-full border rounded-md p-2"
                              value={effect.type}
                              onChange={(e) =>
                                updateEffect(effectIndex, {
                                  type: e.target.value,
                                })
                              }
                            >
                              {knownEffectTypes.map((type) => (
                                <option key={type} value={type}>
                                  {type}
                                </option>
                              ))}
                              <option value="__custom__">Custom</option>
                            </select>
                          </label>

                          {effect.type === '__custom__' ? (
                            <label className="text-sm">
                              Custom Type
                              <input
                                className="w-full border rounded-md p-2"
                                value={effect.customType}
                                onChange={(e) =>
                                  updateEffect(effectIndex, {
                                    customType: e.target.value,
                                  })
                                }
                              />
                            </label>
                          ) : (
                            <div />
                          )}

                          <label className="text-sm">
                            Amount
                            <input
                              className="w-full border rounded-md p-2"
                              type="number"
                              value={effect.amount}
                              onChange={(e) =>
                                updateEffect(effectIndex, {
                                  amount: e.target.value,
                                })
                              }
                            />
                          </label>
                        </div>

                        <label className="text-sm block mb-3">
                          Statuses to Remove (comma separated names)
                          <input
                            className="w-full border rounded-md p-2"
                            value={effect.statusesToRemove}
                            onChange={(e) =>
                              updateEffect(effectIndex, {
                                statusesToRemove: e.target.value,
                              })
                            }
                          />
                        </label>

                        <label className="text-sm block mb-3">
                          Effect Data (JSON object)
                          <textarea
                            className="w-full border rounded-md p-2 min-h-[84px] font-mono text-xs"
                            value={effect.effectData}
                            onChange={(e) =>
                              updateEffect(effectIndex, {
                                effectData: e.target.value,
                              })
                            }
                          />
                        </label>

                        <div className="border rounded-md p-3 bg-white">
                          <div className="flex items-center justify-between mb-2">
                            <div className="font-medium text-sm">Statuses to Apply</div>
                            <button
                              type="button"
                              className="qa-btn qa-btn-secondary"
                              onClick={() => addEffectStatus(effectIndex)}
                            >
                              Add Status
                            </button>
                          </div>

                          <div className="space-y-3">
                            {effect.statusesToApply.map((status, statusIndex) => (
                              <div key={statusIndex} className="border rounded-md p-3 bg-gray-50">
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                                  <label className="text-xs">
                                    Name
                                    <input
                                      className="w-full border rounded-md p-1.5"
                                      value={status.name}
                                      onChange={(e) =>
                                        updateEffectStatus(effectIndex, statusIndex, {
                                          name: e.target.value,
                                        })
                                      }
                                    />
                                  </label>
                                  <label className="text-xs">
                                    Duration (seconds)
                                    <input
                                      className="w-full border rounded-md p-1.5"
                                      type="number"
                                      min={1}
                                      value={status.durationSeconds}
                                      onChange={(e) =>
                                        updateEffectStatus(effectIndex, statusIndex, {
                                          durationSeconds: e.target.value,
                                        })
                                      }
                                    />
                                  </label>
                                  <label className="text-xs md:col-span-2">
                                    Description
                                    <input
                                      className="w-full border rounded-md p-1.5"
                                      value={status.description}
                                      onChange={(e) =>
                                        updateEffectStatus(effectIndex, statusIndex, {
                                          description: e.target.value,
                                        })
                                      }
                                    />
                                  </label>
                                  <label className="text-xs md:col-span-2">
                                    Effect
                                    <input
                                      className="w-full border rounded-md p-1.5"
                                      value={status.effect}
                                      onChange={(e) =>
                                        updateEffectStatus(effectIndex, statusIndex, {
                                          effect: e.target.value,
                                        })
                                      }
                                    />
                                  </label>
                                  <label className="text-xs inline-flex items-center gap-2">
                                    <input
                                      type="checkbox"
                                      checked={status.positive}
                                      onChange={(e) =>
                                        updateEffectStatus(effectIndex, statusIndex, {
                                          positive: e.target.checked,
                                        })
                                      }
                                    />
                                    Positive
                                  </label>
                                  <div className="grid grid-cols-3 md:grid-cols-6 gap-2 md:col-span-2">
                                    {[
                                      ['strengthMod', 'STR'],
                                      ['dexterityMod', 'DEX'],
                                      ['constitutionMod', 'CON'],
                                      ['intelligenceMod', 'INT'],
                                      ['wisdomMod', 'WIS'],
                                      ['charismaMod', 'CHA'],
                                    ].map(([key, label]) => (
                                      <label className="text-[11px]" key={key}>
                                        {label}
                                        <input
                                          className="w-full border rounded-md p-1"
                                          type="number"
                                          value={status[key as keyof SpellStatusTemplateForm] as string}
                                          onChange={(e) =>
                                            updateEffectStatus(effectIndex, statusIndex, {
                                              [key]: e.target.value,
                                            })
                                          }
                                        />
                                      </label>
                                    ))}
                                  </div>
                                </div>
                                <div className="mt-2">
                                  <button
                                    type="button"
                                    className="qa-btn qa-btn-danger"
                                    onClick={() => removeEffectStatus(effectIndex, statusIndex)}
                                  >
                                    Remove Status
                                  </button>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>

                        <div className="mt-3">
                          <button
                            type="button"
                            className="qa-btn qa-btn-danger"
                            onClick={() => removeEffect(effectIndex)}
                          >
                            Remove Effect
                          </button>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              <div className="p-5 border-t flex items-center justify-end gap-2">
                <button className="qa-btn qa-btn-secondary" onClick={closeModal}>
                  Cancel
                </button>
                <button className="qa-btn qa-btn-primary" onClick={save}>
                  {editingSpell ? 'Save Changes' : 'Create Spell'}
                </button>
              </div>
            </div>
          </div>
        )}

        {deleteId && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
            <div className="bg-white rounded-lg shadow-lg w-full max-w-md p-5">
              <h3 className="text-lg font-semibold mb-2">Delete Spell?</h3>
              <p className="text-sm text-gray-700 mb-4">This action cannot be undone.</p>
              <div className="flex justify-end gap-2">
                <button className="qa-btn qa-btn-secondary" onClick={() => setDeleteId(null)}>
                  Cancel
                </button>
                <button className="qa-btn qa-btn-danger" onClick={confirmDelete}>
                  Delete
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default Spells;
