import React, { useEffect, useMemo, useState } from 'react';
import { useAPI, useInventory } from '@poltergeist/contexts';
import { Character, Spell } from '@poltergeist/types';

type SelectOption = {
  value: string;
  label: string;
};

type TutorialItemRewardRow = {
  id: string;
  inventoryItemId: string;
  quantity: number;
};

type TutorialSpellRewardRow = {
  id: string;
  spellId: string;
};

type TutorialOptionRow = {
  id: string;
  optionText: string;
  statTag: string;
  difficulty: number;
  rewardExperience: number;
  rewardGold: number;
  itemRewards: TutorialItemRewardRow[];
  spellRewards: TutorialSpellRewardRow[];
};

type TutorialOptionResponse = {
  optionText?: string;
  statTag?: string;
  difficulty?: number;
  rewardExperience?: number;
  rewardGold?: number;
  itemRewards?: Array<{ inventoryItemId?: number; quantity?: number }>;
  spellRewards?: Array<{ spellId?: string }>;
};

type TutorialConfigResponse = {
  characterId?: string | null;
  dialogue?: string[];
  scenarioPrompt?: string;
  scenarioImageUrl?: string;
  imageGenerationStatus?: string;
  imageGenerationError?: string | null;
  rewardExperience?: number;
  rewardGold?: number;
  options?: TutorialOptionResponse[];
  itemRewards?: Array<{ inventoryItemId?: number; quantity?: number }>;
  spellRewards?: Array<{ spellId?: string }>;
};

const statTagOptions: SelectOption[] = [
  { value: 'strength', label: 'Strength' },
  { value: 'dexterity', label: 'Dexterity' },
  { value: 'constitution', label: 'Constitution' },
  { value: 'intelligence', label: 'Intelligence' },
  { value: 'wisdom', label: 'Wisdom' },
  { value: 'charisma', label: 'Charisma' },
];

const createId = () => `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;

const makeItemRewardRow = (inventoryItemId = '', quantity = 1): TutorialItemRewardRow => ({
  id: createId(),
  inventoryItemId,
  quantity,
});

const makeSpellRewardRow = (spellId = ''): TutorialSpellRewardRow => ({
  id: createId(),
  spellId,
});

const makeOptionRow = ({
  optionText = '',
  statTag = 'strength',
  difficulty = 0,
  rewardExperience = 0,
  rewardGold = 0,
  itemRewards = [],
  spellRewards = [],
}: Partial<Omit<TutorialOptionRow, 'id'>> = {}): TutorialOptionRow => ({
  id: createId(),
  optionText,
  statTag,
  difficulty,
  rewardExperience,
  rewardGold,
  itemRewards,
  spellRewards,
});

const toItemRewardRows = (
  input: Array<{ inventoryItemId?: number; quantity?: number }> | undefined,
) =>
  Array.isArray(input)
    ? input
        .filter((reward) => (reward.inventoryItemId ?? 0) > 0 && (reward.quantity ?? 0) > 0)
        .map((reward) => makeItemRewardRow(String(reward.inventoryItemId), reward.quantity ?? 1))
    : [];

const toSpellRewardRows = (input: Array<{ spellId?: string }> | undefined) =>
  Array.isArray(input)
    ? input
        .map((reward) => (reward.spellId ?? '').trim())
        .filter(Boolean)
        .map((spellId) => makeSpellRewardRow(spellId))
    : [];

const optionNeedsLegacySharedRewards = (option: TutorialOptionResponse) =>
  (option.rewardExperience ?? 0) === 0 &&
  (option.rewardGold ?? 0) === 0 &&
  (!Array.isArray(option.itemRewards) || option.itemRewards.length === 0) &&
  (!Array.isArray(option.spellRewards) || option.spellRewards.length === 0);

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

  const selected = options.find((option) => option.value === value);
  const filtered = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    if (!normalized) return options;
    return options.filter((option) => option.label.toLowerCase().includes(normalized));
  }, [options, query]);

  return (
    <div className="relative">
      <label className="block text-sm font-medium text-gray-700">{label}</label>
      <input
        value={open ? query : selected?.label ?? ''}
        onChange={(event) => {
          setQuery(event.target.value);
          setOpen(true);
        }}
        onFocus={() => {
          setOpen(true);
          setQuery('');
        }}
        onBlur={() => {
          window.setTimeout(() => setOpen(false), 150);
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
              onMouseDown={(event) => event.preventDefault()}
              onClick={() => {
                onChange(option.value);
                setQuery('');
                setOpen(false);
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

export const Tutorial = () => {
  const { apiClient } = useAPI();
  const { inventoryItems } = useInventory();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [generatingImage, setGeneratingImage] = useState(false);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [statusKind, setStatusKind] = useState<'success' | 'error' | null>(null);
  const [characters, setCharacters] = useState<Character[]>([]);
  const [spells, setSpells] = useState<Spell[]>([]);
  const [characterId, setCharacterId] = useState('');
  const [dialogue, setDialogue] = useState<string[]>([]);
  const [scenarioPrompt, setScenarioPrompt] = useState('');
  const [scenarioImageUrl, setScenarioImageUrl] = useState('');
  const [imageGenerationStatus, setImageGenerationStatus] = useState('none');
  const [imageGenerationError, setImageGenerationError] = useState<string | null>(null);
  const [options, setOptions] = useState<TutorialOptionRow[]>([]);
  const imageGenerationActive =
    imageGenerationStatus === 'queued' || imageGenerationStatus === 'in_progress';

  useEffect(() => {
    const load = async () => {
      try {
        setLoading(true);
        const [config, loadedCharacters, loadedSpells] = await Promise.all([
          apiClient.get<TutorialConfigResponse>('/sonar/admin/tutorial'),
          apiClient.get<Character[]>('/sonar/characters'),
          apiClient.get<Spell[]>('/sonar/spells'),
        ]);

        const sharedItemRewards = toItemRewardRows(config.itemRewards);
        const sharedSpellRewards = toSpellRewardRows(config.spellRewards);
        const sharedRewardExperience =
          typeof config.rewardExperience === 'number' ? config.rewardExperience : 0;
        const sharedRewardGold = typeof config.rewardGold === 'number' ? config.rewardGold : 0;

        setCharacters(Array.isArray(loadedCharacters) ? loadedCharacters : []);
        setSpells(Array.isArray(loadedSpells) ? loadedSpells : []);
        setCharacterId(config.characterId ?? '');
        setDialogue(Array.isArray(config.dialogue) ? config.dialogue : []);
        setScenarioPrompt(config.scenarioPrompt ?? '');
        setScenarioImageUrl(config.scenarioImageUrl ?? '');
        setImageGenerationStatus((config.imageGenerationStatus ?? 'none').trim() || 'none');
        setImageGenerationError((config.imageGenerationError ?? '').trim() || null);
        setOptions(
          Array.isArray(config.options) && config.options.length > 0
            ? config.options.map((option) => {
                const useLegacySharedRewards = optionNeedsLegacySharedRewards(option);
                return makeOptionRow({
                  optionText: option.optionText ?? '',
                  statTag: option.statTag ?? 'strength',
                  difficulty: Math.max(0, option.difficulty ?? 0),
                  rewardExperience: useLegacySharedRewards
                    ? sharedRewardExperience
                    : Math.max(0, option.rewardExperience ?? 0),
                  rewardGold: useLegacySharedRewards
                    ? sharedRewardGold
                    : Math.max(0, option.rewardGold ?? 0),
                  itemRewards: useLegacySharedRewards
                    ? sharedItemRewards.map((reward) =>
                        makeItemRewardRow(reward.inventoryItemId, reward.quantity),
                      )
                    : toItemRewardRows(option.itemRewards),
                  spellRewards: useLegacySharedRewards
                    ? sharedSpellRewards.map((reward) => makeSpellRewardRow(reward.spellId))
                    : toSpellRewardRows(option.spellRewards),
                });
              })
            : [
                makeOptionRow({ optionText: 'I reach for my sword and check it out.', statTag: 'strength' }),
                makeOptionRow({
                  optionText: 'I reach for my shield and check it out.',
                  statTag: 'constitution',
                }),
                makeOptionRow({
                  optionText: 'I reach for my spellbook and check it out.',
                  statTag: 'intelligence',
                }),
              ],
        );
      } catch (error) {
        console.error('Failed to load tutorial config', error);
        setStatusMessage('Failed to load tutorial config.');
        setStatusKind('error');
      } finally {
        setLoading(false);
      }
    };

    load();
  }, [apiClient]);

  useEffect(() => {
    if (imageGenerationStatus !== 'queued' && imageGenerationStatus !== 'in_progress') {
      return undefined;
    }

    const intervalId = window.setInterval(async () => {
      try {
        const config = await apiClient.get<TutorialConfigResponse>('/sonar/admin/tutorial');
        setScenarioImageUrl(config.scenarioImageUrl ?? '');
        setImageGenerationStatus((config.imageGenerationStatus ?? 'none').trim() || 'none');
        setImageGenerationError((config.imageGenerationError ?? '').trim() || null);
      } catch (error) {
        console.error('Failed to refresh tutorial image generation status', error);
      }
    }, 1500);

    return () => window.clearInterval(intervalId);
  }, [apiClient, imageGenerationStatus]);

  const characterOptions = useMemo(
    () =>
      characters.map((character) => ({
        value: character.id,
        label: character.name,
      })),
    [characters],
  );

  const itemOptions = useMemo(
    () =>
      (inventoryItems ?? []).map((item) => ({
        value: String(item.id),
        label: item.name,
      })),
    [inventoryItems],
  );

  const spellOptions = useMemo(
    () =>
      spells.map((spell) => ({
        value: spell.id,
        label: spell.name,
      })),
    [spells],
  );

  const updateDialogueLine = (index: number, value: string) => {
    setDialogue((prev) => prev.map((line, lineIndex) => (lineIndex === index ? value : line)));
  };

  const removeDialogueLine = (index: number) => {
    setDialogue((prev) => prev.filter((_, lineIndex) => lineIndex !== index));
  };

  const updateOption = (id: string, updates: Partial<TutorialOptionRow>) => {
    setOptions((prev) => prev.map((option) => (option.id === id ? { ...option, ...updates } : option)));
  };

  const removeOption = (id: string) => {
    setOptions((prev) => prev.filter((option) => option.id !== id));
  };

  const addOptionItemReward = (optionId: string) => {
    setOptions((prev) =>
      prev.map((option) =>
        option.id === optionId
          ? { ...option, itemRewards: [...option.itemRewards, makeItemRewardRow()] }
          : option,
      ),
    );
  };

  const updateOptionItemReward = (
    optionId: string,
    rewardId: string,
    updates: Partial<TutorialItemRewardRow>,
  ) => {
    setOptions((prev) =>
      prev.map((option) =>
        option.id === optionId
          ? {
              ...option,
              itemRewards: option.itemRewards.map((reward) =>
                reward.id === rewardId ? { ...reward, ...updates } : reward,
              ),
            }
          : option,
      ),
    );
  };

  const removeOptionItemReward = (optionId: string, rewardId: string) => {
    setOptions((prev) =>
      prev.map((option) =>
        option.id === optionId
          ? {
              ...option,
              itemRewards: option.itemRewards.filter((reward) => reward.id !== rewardId),
            }
          : option,
      ),
    );
  };

  const addOptionSpellReward = (optionId: string) => {
    setOptions((prev) =>
      prev.map((option) =>
        option.id === optionId
          ? { ...option, spellRewards: [...option.spellRewards, makeSpellRewardRow()] }
          : option,
      ),
    );
  };

  const updateOptionSpellReward = (
    optionId: string,
    rewardId: string,
    updates: Partial<TutorialSpellRewardRow>,
  ) => {
    setOptions((prev) =>
      prev.map((option) =>
        option.id === optionId
          ? {
              ...option,
              spellRewards: option.spellRewards.map((reward) =>
                reward.id === rewardId ? { ...reward, ...updates } : reward,
              ),
            }
          : option,
      ),
    );
  };

  const removeOptionSpellReward = (optionId: string, rewardId: string) => {
    setOptions((prev) =>
      prev.map((option) =>
        option.id === optionId
          ? {
              ...option,
              spellRewards: option.spellRewards.filter((reward) => reward.id !== rewardId),
            }
          : option,
      ),
    );
  };

  const handleSave = async () => {
    try {
      setSaving(true);
      setStatusMessage(null);
      setStatusKind(null);

      await apiClient.put('/sonar/admin/tutorial', {
        characterId: characterId || null,
        dialogue: dialogue.map((line) => line.trim()).filter(Boolean),
        scenarioPrompt: scenarioPrompt.trim(),
        scenarioImageUrl: scenarioImageUrl.trim(),
        rewardExperience: 0,
        rewardGold: 0,
        itemRewards: [],
        spellRewards: [],
        options: options
          .map((option) => ({
            optionText: option.optionText.trim(),
            statTag: option.statTag,
            difficulty: Math.max(0, option.difficulty),
            rewardExperience: Math.max(0, option.rewardExperience),
            rewardGold: Math.max(0, option.rewardGold),
            itemRewards: option.itemRewards
              .filter((reward) => reward.inventoryItemId && reward.quantity > 0)
              .map((reward) => ({
                inventoryItemId: Number.parseInt(reward.inventoryItemId, 10),
                quantity: reward.quantity,
              })),
            spellRewards: option.spellRewards
              .filter((reward) => reward.spellId)
              .map((reward) => ({ spellId: reward.spellId })),
          }))
          .filter((option) => option.optionText),
      });

      setStatusMessage('Tutorial config saved.');
      setStatusKind('success');
    } catch (error) {
      console.error('Failed to save tutorial config', error);
      setStatusMessage('Failed to save tutorial config.');
      setStatusKind('error');
    } finally {
      setSaving(false);
    }
  };

  const handleGenerateScenarioImage = async () => {
    const prompt = scenarioPrompt.trim();
    if (!prompt) {
      setStatusMessage('Enter a scenario prompt before generating an image.');
      setStatusKind('error');
      return;
    }

    try {
      setGeneratingImage(true);
      setStatusMessage(null);
      setStatusKind(null);

      const response = await apiClient.post<TutorialConfigResponse>(
        '/sonar/admin/tutorial/generate-image',
        { scenarioPrompt: prompt },
      );
      setScenarioImageUrl(response.scenarioImageUrl ?? '');
      setImageGenerationStatus((response.imageGenerationStatus ?? 'queued').trim() || 'queued');
      setImageGenerationError((response.imageGenerationError ?? '').trim() || null);
      setStatusMessage('Tutorial scenario image generation queued.');
      setStatusKind('success');
    } catch (error) {
      console.error('Failed to generate tutorial scenario image', error);
      setStatusMessage('Failed to generate tutorial scenario image.');
      setStatusKind('error');
    } finally {
      setGeneratingImage(false);
    }
  };

  return (
    <div className="px-4 py-6">
      <div className="mx-auto flex w-full max-w-5xl flex-col gap-6 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
        <div>
          <h1 className="text-xl font-semibold text-gray-900">Tutorial</h1>
          <p className="mt-1 text-sm text-gray-500">
            Configure the first-run welcome dialogue and one-time tutorial scenario for newly
            registered users.
          </p>
        </div>

        {loading ? (
          <div className="text-sm text-gray-500">Loading…</div>
        ) : (
          <>
            <section className="rounded-lg border border-gray-200 p-4">
              <div className="mb-4">
                <h2 className="text-sm font-semibold text-gray-900">Welcome Dialogue</h2>
                <p className="mt-1 text-xs text-gray-500">
                  The dialogue appears the first time a newly initialized user lands on single
                  player.
                </p>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <label className="block text-sm font-medium text-gray-700">Character</label>
                  <select
                    value={characterId}
                    onChange={(event) => setCharacterId(event.target.value)}
                    className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                  >
                    <option value="">No character selected</option>
                    {characterOptions.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="mt-4 flex flex-col gap-3">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-sm font-medium text-gray-800">Dialogue Lines</h3>
                    <p className="text-xs text-gray-500">Each line becomes one character message.</p>
                  </div>
                  <button
                    type="button"
                    onClick={() => setDialogue((prev) => [...prev, ''])}
                    className="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50"
                  >
                    Add Line
                  </button>
                </div>

                {dialogue.length === 0 && (
                  <div className="rounded-md border border-dashed border-gray-300 bg-gray-50 px-3 py-4 text-sm text-gray-500">
                    No dialogue lines configured.
                  </div>
                )}

                {dialogue.map((line, index) => (
                  <div key={`dialogue-${index}`} className="flex gap-3">
                    <textarea
                      value={line}
                      onChange={(event) => updateDialogueLine(index, event.target.value)}
                      rows={2}
                      className="block flex-1 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                    />
                    <button
                      type="button"
                      onClick={() => removeDialogueLine(index)}
                      className="self-start rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-600 hover:bg-gray-100"
                    >
                      Remove
                    </button>
                  </div>
                ))}
              </div>
            </section>

            <section className="rounded-lg border border-gray-200 p-4">
              <div className="mb-4">
                <h2 className="text-sm font-semibold text-gray-900">Tutorial Scenario</h2>
                <p className="mt-1 text-xs text-gray-500">
                  This scenario is spawned once at the user&apos;s current location after the welcome
                  dialogue finishes.
                </p>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700">Scenario Prompt</label>
                  <textarea
                    value={scenarioPrompt}
                    onChange={(event) => setScenarioPrompt(event.target.value)}
                    rows={3}
                    className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                  />
                </div>
                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700">Scenario Image URL</label>
                  <div className="mt-1 flex flex-col gap-3">
                    <div className="flex flex-col gap-3 md:flex-row">
                      <input
                        type="text"
                        value={scenarioImageUrl}
                        onChange={(event) => setScenarioImageUrl(event.target.value)}
                        placeholder="https://..."
                        className="block flex-1 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                      />
                      <button
                        type="button"
                        onClick={handleGenerateScenarioImage}
                        disabled={generatingImage || imageGenerationActive || saving}
                        className="rounded-md bg-purple-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-purple-500 disabled:cursor-not-allowed disabled:opacity-60"
                      >
                        {generatingImage || imageGenerationActive ? 'Generating…' : 'Generate Image'}
                      </button>
                    </div>
                    <div className="rounded-md border border-gray-200 bg-gray-50 p-3">
                      <div className="mb-2 flex items-center justify-between">
                        <span className="text-sm font-medium text-gray-700">Preview</span>
                        <span className="text-xs text-gray-500">
                          {imageGenerationStatus === 'queued'
                            ? 'Queued'
                            : imageGenerationStatus === 'in_progress'
                              ? 'Generating…'
                              : imageGenerationStatus === 'failed'
                                ? 'Failed'
                                : imageGenerationStatus === 'complete'
                                  ? 'Ready'
                                  : 'Idle'}
                        </span>
                      </div>
                      {scenarioImageUrl.trim() ? (
                        <img
                          src={scenarioImageUrl.trim()}
                          alt="Tutorial scenario preview"
                          className="max-h-72 w-full rounded-md border border-gray-200 object-cover"
                        />
                      ) : (
                        <div className="flex h-48 items-center justify-center rounded-md border border-dashed border-gray-300 bg-white text-sm text-gray-500">
                          No tutorial image yet.
                        </div>
                      )}
                      {imageGenerationError && (
                        <p className="mt-2 text-xs text-rose-600">{imageGenerationError}</p>
                      )}
                    </div>
                  </div>
                </div>
              </div>

              <div className="mt-6 flex flex-col gap-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-sm font-medium text-gray-800">Scenario Options</h3>
                    <p className="text-xs text-gray-500">
                      Each option has its own stat check and reward bundle.
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={() => setOptions((prev) => [...prev, makeOptionRow()])}
                    className="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50"
                  >
                    Add Option
                  </button>
                </div>

                {options.map((option, index) => (
                  <div key={option.id} className="rounded-md border border-gray-200 bg-gray-50 p-4">
                    <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_180px_auto]">
                      <input
                        type="text"
                        value={option.optionText}
                        onChange={(event) => updateOption(option.id, { optionText: event.target.value })}
                        placeholder={`Option ${index + 1} text`}
                        className="block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                      />
                      <select
                        value={option.statTag}
                        onChange={(event) => updateOption(option.id, { statTag: event.target.value })}
                        className="block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                      >
                        {statTagOptions.map((statTag) => (
                          <option key={statTag.value} value={statTag.value}>
                            {statTag.label}
                          </option>
                        ))}
                      </select>
                      <button
                        type="button"
                        onClick={() => removeOption(option.id)}
                        className="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-600 hover:bg-gray-100"
                      >
                        Remove
                      </button>
                    </div>

                    <div className="mt-4 grid gap-3 md:grid-cols-3">
                      <label className="text-sm">
                        <span className="block font-medium text-gray-700">Difficulty</span>
                        <input
                          type="number"
                          min="0"
                          value={option.difficulty}
                          onChange={(event) =>
                            updateOption(option.id, {
                              difficulty: Number.parseInt(event.target.value || '0', 10),
                            })
                          }
                          className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                        />
                      </label>
                      <label className="text-sm">
                        <span className="block font-medium text-gray-700">Reward Experience</span>
                        <input
                          type="number"
                          min="0"
                          value={option.rewardExperience}
                          onChange={(event) =>
                            updateOption(option.id, {
                              rewardExperience: Number.parseInt(event.target.value || '0', 10),
                            })
                          }
                          className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                        />
                      </label>
                      <label className="text-sm">
                        <span className="block font-medium text-gray-700">Reward Gold</span>
                        <input
                          type="number"
                          min="0"
                          value={option.rewardGold}
                          onChange={(event) =>
                            updateOption(option.id, {
                              rewardGold: Number.parseInt(event.target.value || '0', 10),
                            })
                          }
                          className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                        />
                      </label>
                    </div>

                    <div className="mt-4 grid gap-4 lg:grid-cols-2">
                      <div className="flex flex-col gap-3">
                        <div className="flex items-center justify-between">
                          <div>
                            <h4 className="text-sm font-medium text-gray-800">Item Rewards</h4>
                            <p className="text-xs text-gray-500">Granted when this option succeeds.</p>
                          </div>
                          <button
                            type="button"
                            onClick={() => addOptionItemReward(option.id)}
                            className="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50"
                          >
                            Add Item
                          </button>
                        </div>

                        {option.itemRewards.length === 0 && (
                          <div className="rounded-md border border-dashed border-gray-300 bg-white px-3 py-4 text-sm text-gray-500">
                            No item rewards configured.
                          </div>
                        )}

                        {option.itemRewards.map((reward) => (
                          <div key={reward.id} className="rounded-md border border-gray-200 bg-white p-3">
                            <SearchableSelect
                              label="Inventory Item"
                              placeholder="Search item name…"
                              options={itemOptions}
                              value={reward.inventoryItemId}
                              onChange={(value) =>
                                updateOptionItemReward(option.id, reward.id, { inventoryItemId: value })
                              }
                            />
                            <div className="mt-3 flex items-end gap-3">
                              <div className="flex-1">
                                <label className="block text-sm font-medium text-gray-700">Quantity</label>
                                <input
                                  type="number"
                                  min="1"
                                  value={reward.quantity}
                                  onChange={(event) =>
                                    updateOptionItemReward(option.id, reward.id, {
                                      quantity: Number.parseInt(event.target.value || '1', 10),
                                    })
                                  }
                                  className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                                />
                              </div>
                              <button
                                type="button"
                                onClick={() => removeOptionItemReward(option.id, reward.id)}
                                className="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-600 hover:bg-gray-100"
                              >
                                Remove
                              </button>
                            </div>
                          </div>
                        ))}
                      </div>

                      <div className="flex flex-col gap-3">
                        <div className="flex items-center justify-between">
                          <div>
                            <h4 className="text-sm font-medium text-gray-800">Spell Rewards</h4>
                            <p className="text-xs text-gray-500">Granted when this option succeeds.</p>
                          </div>
                          <button
                            type="button"
                            onClick={() => addOptionSpellReward(option.id)}
                            className="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50"
                          >
                            Add Spell
                          </button>
                        </div>

                        {option.spellRewards.length === 0 && (
                          <div className="rounded-md border border-dashed border-gray-300 bg-white px-3 py-4 text-sm text-gray-500">
                            No spell rewards configured.
                          </div>
                        )}

                        {option.spellRewards.map((reward) => (
                          <div key={reward.id} className="rounded-md border border-gray-200 bg-white p-3">
                            <SearchableSelect
                              label="Spell"
                              placeholder="Search spell name…"
                              options={spellOptions}
                              value={reward.spellId}
                              onChange={(value) =>
                                updateOptionSpellReward(option.id, reward.id, { spellId: value })
                              }
                            />
                            <button
                              type="button"
                              onClick={() => removeOptionSpellReward(option.id, reward.id)}
                              className="mt-3 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-600 hover:bg-gray-100"
                            >
                              Remove
                            </button>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </section>

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

            <div className="flex justify-end">
              <button
                type="button"
                onClick={handleSave}
                disabled={saving || generatingImage}
                className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-60"
              >
                {saving ? 'Saving…' : 'Save Tutorial'}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
};

export default Tutorial;
