import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useAPI, useInventory } from '@poltergeist/contexts';
import {
  Character,
  CharacterTemplate,
  DialogueMessage,
  QuestArchetype,
  Spell,
  User,
} from '@poltergeist/types';
import { DialogueMessageListEditor } from './DialogueMessageListEditor.tsx';

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
  baseQuestArchetypeId?: string | null;
  baseQuestGiverCharacterId?: string | null;
  baseQuestGiverCharacterTemplateId?: string | null;
  dialogue?: DialogueMessage[];
  loadoutDialogue?: DialogueMessage[];
  postMonsterDialogue?: DialogueMessage[];
  baseKitDialogue?: DialogueMessage[];
  postBasePlacementDialogue?: DialogueMessage[];
  postBaseDialogue?: DialogueMessage[];
  scenarioPrompt?: string;
  scenarioImageUrl?: string;
  imageGenerationStatus?: string;
  imageGenerationError?: string | null;
  monsterEncounterId?: string | null;
  monsterRewardExperience?: number;
  monsterRewardGold?: number;
  monsterItemRewards?: Array<{ inventoryItemId?: number; quantity?: number }>;
  rewardExperience?: number;
  rewardGold?: number;
  options?: TutorialOptionResponse[];
  itemRewards?: Array<{ inventoryItemId?: number; quantity?: number }>;
  spellRewards?: Array<{ spellId?: string }>;
};

type MonsterEncounterOptionResponse = {
  id: string;
  name: string;
};

type AdminMonsterEncounterListResponse = {
  items?: MonsterEncounterOptionResponse[];
};

const statTagOptions: SelectOption[] = [
  { value: 'strength', label: 'Strength' },
  { value: 'dexterity', label: 'Dexterity' },
  { value: 'constitution', label: 'Constitution' },
  { value: 'intelligence', label: 'Intelligence' },
  { value: 'wisdom', label: 'Wisdom' },
  { value: 'charisma', label: 'Charisma' },
];

const createId = () =>
  `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;

const makeItemRewardRow = (
  inventoryItemId = '',
  quantity = 1
): TutorialItemRewardRow => ({
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
  input: Array<{ inventoryItemId?: number; quantity?: number }> | undefined
) =>
  Array.isArray(input)
    ? input
        .filter(
          (reward) =>
            (reward.inventoryItemId ?? 0) > 0 && (reward.quantity ?? 0) > 0
        )
        .map((reward) =>
          makeItemRewardRow(
            String(reward.inventoryItemId),
            reward.quantity ?? 1
          )
        )
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
    return options.filter((option) =>
      option.label.toLowerCase().includes(normalized)
    );
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
            <div className="px-3 py-2 text-sm text-gray-500">
              No matches found
            </div>
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

const formatUserOptionLabel = (user: User) => {
  const username = (user.username ?? '').trim();
  const name = (user.name ?? '').trim();
  const phoneNumber = (user.phoneNumber ?? '').trim();

  if (username && name && name.toLowerCase() !== username.toLowerCase()) {
    return `@${username} - ${name}`;
  }
  if (username) {
    return `@${username}`;
  }
  if (name && phoneNumber) {
    return `${name} - ${phoneNumber}`;
  }
  if (name) {
    return name;
  }
  return phoneNumber || user.id;
};

export const Tutorial = () => {
  const { apiClient } = useAPI();
  const { inventoryItems } = useInventory();
  const hasLoadedConfigRef = useRef(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [generatingImage, setGeneratingImage] = useState(false);
  const [referenceDataLoading, setReferenceDataLoading] = useState(false);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [statusKind, setStatusKind] = useState<'success' | 'error' | null>(
    null
  );
  const [users, setUsers] = useState<User[]>([]);
  const [characters, setCharacters] = useState<Character[]>([]);
  const [characterTemplates, setCharacterTemplates] = useState<
    CharacterTemplate[]
  >([]);
  const [questArchetypes, setQuestArchetypes] = useState<QuestArchetype[]>([]);
  const [spells, setSpells] = useState<Spell[]>([]);
  const [monsterEncounters, setMonsterEncounters] = useState<
    Array<{ id: string; name: string }>
  >([]);
  const [characterId, setCharacterId] = useState('');
  const [baseQuestArchetypeId, setBaseQuestArchetypeId] = useState('');
  const [baseQuestGiverSourceType, setBaseQuestGiverSourceType] = useState<
    'character' | 'template'
  >('character');
  const [baseQuestGiverCharacterId, setBaseQuestGiverCharacterId] =
    useState('');
  const [
    baseQuestGiverCharacterTemplateId,
    setBaseQuestGiverCharacterTemplateId,
  ] = useState('');
  const [dialogue, setDialogue] = useState<DialogueMessage[]>([]);
  const [loadoutDialogue, setLoadoutDialogue] = useState<DialogueMessage[]>([]);
  const [postMonsterDialogue, setPostMonsterDialogue] = useState<
    DialogueMessage[]
  >([]);
  const [baseKitDialogue, setBaseKitDialogue] = useState<DialogueMessage[]>([]);
  const [postBasePlacementDialogue, setPostBasePlacementDialogue] = useState<
    DialogueMessage[]
  >([]);
  const [postBaseDialogue, setPostBaseDialogue] = useState<DialogueMessage[]>(
    []
  );
  const [scenarioPrompt, setScenarioPrompt] = useState('');
  const [scenarioImageUrl, setScenarioImageUrl] = useState('');
  const [imageGenerationStatus, setImageGenerationStatus] = useState('none');
  const [imageGenerationError, setImageGenerationError] = useState<
    string | null
  >(null);
  const [options, setOptions] = useState<TutorialOptionRow[]>([]);
  const [monsterEncounterId, setMonsterEncounterId] = useState('');
  const [monsterRewardExperience, setMonsterRewardExperience] = useState(0);
  const [monsterRewardGold, setMonsterRewardGold] = useState(0);
  const [monsterItemRewards, setMonsterItemRewards] = useState<
    TutorialItemRewardRow[]
  >([]);
  const [baseQuestPreviewUserId, setBaseQuestPreviewUserId] = useState('');
  const [queueingBaseQuest, setQueueingBaseQuest] = useState(false);
  const imageGenerationActive =
    imageGenerationStatus === 'queued' ||
    imageGenerationStatus === 'in_progress';

  useEffect(() => {
    if (hasLoadedConfigRef.current) {
      return;
    }

    const loadConfig = async () => {
      try {
        setLoading(true);
        const config = await apiClient.get<TutorialConfigResponse>(
          '/sonar/admin/tutorial'
        );

        const sharedItemRewards = toItemRewardRows(config.itemRewards);
        const sharedSpellRewards = toSpellRewardRows(config.spellRewards);
        const sharedRewardExperience =
          typeof config.rewardExperience === 'number'
            ? config.rewardExperience
            : 0;
        const sharedRewardGold =
          typeof config.rewardGold === 'number' ? config.rewardGold : 0;

        setCharacterId(config.characterId ?? '');
        setBaseQuestArchetypeId(config.baseQuestArchetypeId ?? '');
        setBaseQuestGiverSourceType(
          config.baseQuestGiverCharacterTemplateId ? 'template' : 'character'
        );
        setBaseQuestGiverCharacterId(config.baseQuestGiverCharacterId ?? '');
        setBaseQuestGiverCharacterTemplateId(
          config.baseQuestGiverCharacterTemplateId ?? ''
        );
        setDialogue(Array.isArray(config.dialogue) ? config.dialogue : []);
        setLoadoutDialogue(
          Array.isArray(config.loadoutDialogue) ? config.loadoutDialogue : []
        );
        setPostMonsterDialogue(
          Array.isArray(config.postMonsterDialogue)
            ? config.postMonsterDialogue
            : []
        );
        setBaseKitDialogue(
          Array.isArray(config.baseKitDialogue) ? config.baseKitDialogue : []
        );
        setPostBasePlacementDialogue(
          Array.isArray(config.postBasePlacementDialogue)
            ? config.postBasePlacementDialogue
            : []
        );
        setPostBaseDialogue(
          Array.isArray(config.postBaseDialogue) ? config.postBaseDialogue : []
        );
        setScenarioPrompt(config.scenarioPrompt ?? '');
        setScenarioImageUrl(config.scenarioImageUrl ?? '');
        setImageGenerationStatus(
          (config.imageGenerationStatus ?? 'none').trim() || 'none'
        );
        setImageGenerationError(
          (config.imageGenerationError ?? '').trim() || null
        );
        setMonsterEncounterId(config.monsterEncounterId ?? '');
        setMonsterRewardExperience(
          typeof config.monsterRewardExperience === 'number'
            ? config.monsterRewardExperience
            : 0
        );
        setMonsterRewardGold(
          typeof config.monsterRewardGold === 'number'
            ? config.monsterRewardGold
            : 0
        );
        setMonsterItemRewards(toItemRewardRows(config.monsterItemRewards));
        setOptions(
          Array.isArray(config.options) && config.options.length > 0
            ? config.options.map((option) => {
                const useLegacySharedRewards =
                  optionNeedsLegacySharedRewards(option);
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
                        makeItemRewardRow(
                          reward.inventoryItemId,
                          reward.quantity
                        )
                      )
                    : toItemRewardRows(option.itemRewards),
                  spellRewards: useLegacySharedRewards
                    ? sharedSpellRewards.map((reward) =>
                        makeSpellRewardRow(reward.spellId)
                      )
                    : toSpellRewardRows(option.spellRewards),
                });
              })
            : [
                makeOptionRow({
                  optionText: 'I reach for my sword and check it out.',
                  statTag: 'strength',
                }),
                makeOptionRow({
                  optionText: 'I reach for my shield and check it out.',
                  statTag: 'constitution',
                }),
                makeOptionRow({
                  optionText: 'I reach for my spellbook and check it out.',
                  statTag: 'intelligence',
                }),
              ]
        );
        hasLoadedConfigRef.current = true;
      } catch (error) {
        console.error('Failed to load tutorial config', error);
        setStatusMessage('Failed to load tutorial config.');
        setStatusKind('error');
      } finally {
        setLoading(false);
      }
    };

    void loadConfig();
  }, [apiClient]);

  useEffect(() => {
    const loadReferenceData = async () => {
      try {
        setReferenceDataLoading(true);
        const [
          loadedUsers,
          loadedCharacters,
          loadedCharacterTemplates,
          loadedQuestArchetypes,
          loadedSpells,
          loadedMonsterEncounters,
        ] = await Promise.all([
          apiClient.get<User[]>('/sonar/users'),
          apiClient.get<Character[]>('/sonar/characters'),
          apiClient.get<CharacterTemplate[]>('/sonar/character-templates'),
          apiClient.get<QuestArchetype[]>('/sonar/questArchetypes'),
          apiClient.get<Spell[]>('/sonar/spells'),
          apiClient.get<AdminMonsterEncounterListResponse>(
            '/sonar/admin/monster-encounters?page=1&pageSize=250'
          ),
        ]);

        setUsers(Array.isArray(loadedUsers) ? loadedUsers : []);
        setCharacters(Array.isArray(loadedCharacters) ? loadedCharacters : []);
        setCharacterTemplates(
          Array.isArray(loadedCharacterTemplates)
            ? loadedCharacterTemplates
            : []
        );
        setQuestArchetypes(
          Array.isArray(loadedQuestArchetypes) ? loadedQuestArchetypes : []
        );
        setSpells(Array.isArray(loadedSpells) ? loadedSpells : []);
        setMonsterEncounters(
          Array.isArray(loadedMonsterEncounters?.items)
            ? loadedMonsterEncounters.items
                .filter((encounter) => encounter?.id && encounter?.name)
                .map((encounter) => ({
                  id: encounter.id,
                  name: encounter.name,
                }))
            : []
        );
      } catch (error) {
        console.error('Failed to load tutorial reference data', error);
        setStatusMessage(
          'Tutorial loaded, but some selector data failed to load.'
        );
        setStatusKind('error');
      } finally {
        setReferenceDataLoading(false);
      }
    };

    void loadReferenceData();
  }, [apiClient]);

  useEffect(() => {
    if (
      imageGenerationStatus !== 'queued' &&
      imageGenerationStatus !== 'in_progress'
    ) {
      return undefined;
    }

    const intervalId = window.setInterval(async () => {
      try {
        const config = await apiClient.get<TutorialConfigResponse>(
          '/sonar/admin/tutorial'
        );
        setScenarioImageUrl(config.scenarioImageUrl ?? '');
        setImageGenerationStatus(
          (config.imageGenerationStatus ?? 'none').trim() || 'none'
        );
        setImageGenerationError(
          (config.imageGenerationError ?? '').trim() || null
        );
      } catch (error) {
        console.error(
          'Failed to refresh tutorial image generation status',
          error
        );
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
    [characters]
  );

  const characterTemplateOptions = useMemo(
    () =>
      characterTemplates.map((template) => ({
        value: template.id,
        label: template.name,
      })),
    [characterTemplates]
  );

  const userOptions = useMemo(
    () =>
      users.map((user) => ({
        value: user.id,
        label: formatUserOptionLabel(user),
      })),
    [users]
  );

  const itemOptions = useMemo(
    () =>
      (inventoryItems ?? []).map((item) => ({
        value: String(item.id),
        label: item.name,
      })),
    [inventoryItems]
  );

  const spellOptions = useMemo(
    () =>
      spells.map((spell) => ({
        value: spell.id,
        label: spell.name,
      })),
    [spells]
  );

  const monsterEncounterOptions = useMemo(
    () =>
      monsterEncounters.map((encounter) => ({
        value: encounter.id,
        label: encounter.name,
      })),
    [monsterEncounters]
  );

  const questArchetypeOptions = useMemo(
    () =>
      questArchetypes.map((questArchetype) => ({
        value: questArchetype.id,
        label: questArchetype.name,
      })),
    [questArchetypes]
  );

  const updateOption = (id: string, updates: Partial<TutorialOptionRow>) => {
    setOptions((prev) =>
      prev.map((option) =>
        option.id === id ? { ...option, ...updates } : option
      )
    );
  };

  const removeOption = (id: string) => {
    setOptions((prev) => prev.filter((option) => option.id !== id));
  };

  const addOptionItemReward = (optionId: string) => {
    setOptions((prev) =>
      prev.map((option) =>
        option.id === optionId
          ? {
              ...option,
              itemRewards: [...option.itemRewards, makeItemRewardRow()],
            }
          : option
      )
    );
  };

  const updateOptionItemReward = (
    optionId: string,
    rewardId: string,
    updates: Partial<TutorialItemRewardRow>
  ) => {
    setOptions((prev) =>
      prev.map((option) =>
        option.id === optionId
          ? {
              ...option,
              itemRewards: option.itemRewards.map((reward) =>
                reward.id === rewardId ? { ...reward, ...updates } : reward
              ),
            }
          : option
      )
    );
  };

  const removeOptionItemReward = (optionId: string, rewardId: string) => {
    setOptions((prev) =>
      prev.map((option) =>
        option.id === optionId
          ? {
              ...option,
              itemRewards: option.itemRewards.filter(
                (reward) => reward.id !== rewardId
              ),
            }
          : option
      )
    );
  };

  const addOptionSpellReward = (optionId: string) => {
    setOptions((prev) =>
      prev.map((option) =>
        option.id === optionId
          ? {
              ...option,
              spellRewards: [...option.spellRewards, makeSpellRewardRow()],
            }
          : option
      )
    );
  };

  const updateOptionSpellReward = (
    optionId: string,
    rewardId: string,
    updates: Partial<TutorialSpellRewardRow>
  ) => {
    setOptions((prev) =>
      prev.map((option) =>
        option.id === optionId
          ? {
              ...option,
              spellRewards: option.spellRewards.map((reward) =>
                reward.id === rewardId ? { ...reward, ...updates } : reward
              ),
            }
          : option
      )
    );
  };

  const removeOptionSpellReward = (optionId: string, rewardId: string) => {
    setOptions((prev) =>
      prev.map((option) =>
        option.id === optionId
          ? {
              ...option,
              spellRewards: option.spellRewards.filter(
                (reward) => reward.id !== rewardId
              ),
            }
          : option
      )
    );
  };

  const addMonsterItemReward = () => {
    setMonsterItemRewards((prev) => [...prev, makeItemRewardRow()]);
  };

  const updateMonsterItemReward = (
    rewardId: string,
    updates: Partial<TutorialItemRewardRow>
  ) => {
    setMonsterItemRewards((prev) =>
      prev.map((reward) =>
        reward.id === rewardId ? { ...reward, ...updates } : reward
      )
    );
  };

  const removeMonsterItemReward = (rewardId: string) => {
    setMonsterItemRewards((prev) =>
      prev.filter((reward) => reward.id !== rewardId)
    );
  };

  const saveTutorialConfig = async ({
    showSuccessMessage = true,
  }: {
    showSuccessMessage?: boolean;
  } = {}) => {
    try {
      setSaving(true);
      setStatusMessage(null);
      setStatusKind(null);

      await apiClient.put('/sonar/admin/tutorial', {
        characterId: characterId || null,
        baseQuestArchetypeId: baseQuestArchetypeId || null,
        baseQuestGiverCharacterId:
          baseQuestGiverSourceType === 'character'
            ? baseQuestGiverCharacterId || null
            : null,
        baseQuestGiverCharacterTemplateId:
          baseQuestGiverSourceType === 'template'
            ? baseQuestGiverCharacterTemplateId || null
            : null,
        dialogue,
        loadoutDialogue,
        postMonsterDialogue,
        baseKitDialogue,
        postBasePlacementDialogue,
        postBaseDialogue,
        scenarioPrompt: scenarioPrompt.trim(),
        scenarioImageUrl: scenarioImageUrl.trim(),
        monsterEncounterId: monsterEncounterId || null,
        monsterRewardExperience: Math.max(0, monsterRewardExperience),
        monsterRewardGold: Math.max(0, monsterRewardGold),
        monsterItemRewards: monsterItemRewards
          .filter((reward) => reward.inventoryItemId && reward.quantity > 0)
          .map((reward) => ({
            inventoryItemId: Number.parseInt(reward.inventoryItemId, 10),
            quantity: reward.quantity,
          })),
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

      if (showSuccessMessage) {
        setStatusMessage('Tutorial config saved.');
        setStatusKind('success');
      }
      return true;
    } catch (error) {
      console.error('Failed to save tutorial config', error);
      setStatusMessage('Failed to save tutorial config.');
      setStatusKind('error');
      return false;
    } finally {
      setSaving(false);
    }
  };

  const handleSave = async () => {
    await saveTutorialConfig();
  };

  const handleQueueBaseQuestForUser = async () => {
    if (!baseQuestPreviewUserId) {
      setStatusMessage('Select a user first.');
      setStatusKind('error');
      return;
    }
    if (
      !baseQuestArchetypeId ||
      (baseQuestGiverSourceType === 'character'
        ? !baseQuestGiverCharacterId
        : !baseQuestGiverCharacterTemplateId)
    ) {
      setStatusMessage(
        'Configure the Home Base Quest archetype and source questgiver first.'
      );
      setStatusKind('error');
      return;
    }

    const saved = await saveTutorialConfig({ showSuccessMessage: false });
    if (!saved) {
      return;
    }

    try {
      setQueueingBaseQuest(true);
      setStatusMessage(null);
      setStatusKind(null);

      await apiClient.post('/sonar/admin/tutorial/instantiate-base-quest', {
        userId: baseQuestPreviewUserId,
      });

      const selectedUser = users.find(
        (user) => user.id === baseQuestPreviewUserId
      );
      setStatusMessage(
        `Queued the tutorial follow-up quest for ${
          selectedUser
            ? formatUserOptionLabel(selectedUser)
            : 'the selected user'
        }.`
      );
      setStatusKind('success');
    } catch (error) {
      console.error('Failed to queue tutorial follow-up quest', error);
      setStatusMessage('Failed to queue the tutorial follow-up quest.');
      setStatusKind('error');
    } finally {
      setQueueingBaseQuest(false);
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
        { scenarioPrompt: prompt }
      );
      setScenarioImageUrl(response.scenarioImageUrl ?? '');
      setImageGenerationStatus(
        (response.imageGenerationStatus ?? 'queued').trim() || 'queued'
      );
      setImageGenerationError(
        (response.imageGenerationError ?? '').trim() || null
      );
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
            Configure the first-run welcome dialogue and one-time tutorial
            scenario for newly registered users.
          </p>
        </div>

        {loading ? (
          <div className="text-sm text-gray-500">Loading…</div>
        ) : (
          <>
            <section className="rounded-lg border border-gray-200 p-4">
              <div className="mb-4">
                <h2 className="text-sm font-semibold text-gray-900">
                  Welcome Dialogue
                </h2>
                <p className="mt-1 text-xs text-gray-500">
                  The dialogue appears the first time a newly initialized user
                  lands on single player.
                </p>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    Character
                  </label>
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
                  {referenceDataLoading ? (
                    <p className="mt-1 text-xs text-gray-500">
                      Loading character options…
                    </p>
                  ) : null}
                </div>
              </div>

              <div className="mt-4">
                <DialogueMessageListEditor
                  label="Dialogue Lines"
                  helperText="Each line becomes one character message. Effects are played by the client."
                  value={dialogue}
                  onChange={setDialogue}
                />
              </div>
            </section>

            <section className="rounded-lg border border-gray-200 p-4">
              <div className="mb-4">
                <h2 className="text-sm font-semibold text-gray-900">
                  Tutorial Scenario
                </h2>
                <p className="mt-1 text-xs text-gray-500">
                  This scenario is spawned once at the user&apos;s current
                  location after the welcome dialogue finishes.
                </p>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700">
                    Scenario Prompt
                  </label>
                  <textarea
                    value={scenarioPrompt}
                    onChange={(event) => setScenarioPrompt(event.target.value)}
                    rows={3}
                    className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                  />
                </div>
                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700">
                    Scenario Image URL
                  </label>
                  <div className="mt-1 flex flex-col gap-3">
                    <div className="flex flex-col gap-3 md:flex-row">
                      <input
                        type="text"
                        value={scenarioImageUrl}
                        onChange={(event) =>
                          setScenarioImageUrl(event.target.value)
                        }
                        placeholder="https://..."
                        className="block flex-1 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                      />
                      <button
                        type="button"
                        onClick={handleGenerateScenarioImage}
                        disabled={
                          generatingImage || imageGenerationActive || saving
                        }
                        className="rounded-md bg-purple-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-purple-500 disabled:cursor-not-allowed disabled:opacity-60"
                      >
                        {generatingImage || imageGenerationActive
                          ? 'Generating…'
                          : 'Generate Image'}
                      </button>
                    </div>
                    <div className="rounded-md border border-gray-200 bg-gray-50 p-3">
                      <div className="mb-2 flex items-center justify-between">
                        <span className="text-sm font-medium text-gray-700">
                          Preview
                        </span>
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
                        <p className="mt-2 text-xs text-rose-600">
                          {imageGenerationError}
                        </p>
                      )}
                    </div>
                  </div>
                </div>
              </div>

              <div className="mt-6 flex flex-col gap-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-sm font-medium text-gray-800">
                      Scenario Options
                    </h3>
                    <p className="text-xs text-gray-500">
                      Each option has its own stat check and reward bundle.
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={() =>
                      setOptions((prev) => [...prev, makeOptionRow()])
                    }
                    className="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50"
                  >
                    Add Option
                  </button>
                </div>

                {options.map((option, index) => (
                  <div
                    key={option.id}
                    className="rounded-md border border-gray-200 bg-gray-50 p-4"
                  >
                    <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_180px_auto]">
                      <input
                        type="text"
                        value={option.optionText}
                        onChange={(event) =>
                          updateOption(option.id, {
                            optionText: event.target.value,
                          })
                        }
                        placeholder={`Option ${index + 1} text`}
                        className="block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                      />
                      <select
                        value={option.statTag}
                        onChange={(event) =>
                          updateOption(option.id, {
                            statTag: event.target.value,
                          })
                        }
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
                        <span className="block font-medium text-gray-700">
                          Difficulty
                        </span>
                        <input
                          type="number"
                          min="0"
                          value={option.difficulty}
                          onChange={(event) =>
                            updateOption(option.id, {
                              difficulty: Number.parseInt(
                                event.target.value || '0',
                                10
                              ),
                            })
                          }
                          className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                        />
                      </label>
                      <label className="text-sm">
                        <span className="block font-medium text-gray-700">
                          Reward Experience
                        </span>
                        <input
                          type="number"
                          min="0"
                          value={option.rewardExperience}
                          onChange={(event) =>
                            updateOption(option.id, {
                              rewardExperience: Number.parseInt(
                                event.target.value || '0',
                                10
                              ),
                            })
                          }
                          className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                        />
                      </label>
                      <label className="text-sm">
                        <span className="block font-medium text-gray-700">
                          Reward Gold
                        </span>
                        <input
                          type="number"
                          min="0"
                          value={option.rewardGold}
                          onChange={(event) =>
                            updateOption(option.id, {
                              rewardGold: Number.parseInt(
                                event.target.value || '0',
                                10
                              ),
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
                            <h4 className="text-sm font-medium text-gray-800">
                              Item Rewards
                            </h4>
                            <p className="text-xs text-gray-500">
                              Granted when this option succeeds.
                            </p>
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
                          <div
                            key={reward.id}
                            className="rounded-md border border-gray-200 bg-white p-3"
                          >
                            <SearchableSelect
                              label="Inventory Item"
                              placeholder="Search item name…"
                              options={itemOptions}
                              value={reward.inventoryItemId}
                              onChange={(value) =>
                                updateOptionItemReward(option.id, reward.id, {
                                  inventoryItemId: value,
                                })
                              }
                            />
                            <div className="mt-3 flex items-end gap-3">
                              <div className="flex-1">
                                <label className="block text-sm font-medium text-gray-700">
                                  Quantity
                                </label>
                                <input
                                  type="number"
                                  min="1"
                                  value={reward.quantity}
                                  onChange={(event) =>
                                    updateOptionItemReward(
                                      option.id,
                                      reward.id,
                                      {
                                        quantity: Number.parseInt(
                                          event.target.value || '1',
                                          10
                                        ),
                                      }
                                    )
                                  }
                                  className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                                />
                              </div>
                              <button
                                type="button"
                                onClick={() =>
                                  removeOptionItemReward(option.id, reward.id)
                                }
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
                            <h4 className="text-sm font-medium text-gray-800">
                              Spell Rewards
                            </h4>
                            <p className="text-xs text-gray-500">
                              Granted when this option succeeds.
                            </p>
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
                          <div
                            key={reward.id}
                            className="rounded-md border border-gray-200 bg-white p-3"
                          >
                            <SearchableSelect
                              label="Spell"
                              placeholder="Search spell name…"
                              options={spellOptions}
                              value={reward.spellId}
                              onChange={(value) =>
                                updateOptionSpellReward(option.id, reward.id, {
                                  spellId: value,
                                })
                              }
                            />
                            <button
                              type="button"
                              onClick={() =>
                                removeOptionSpellReward(option.id, reward.id)
                              }
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

            <section className="rounded-lg border border-gray-200 p-4">
              <div className="mb-4">
                <h2 className="text-sm font-semibold text-gray-900">
                  Loadout Step
                </h2>
                <p className="mt-1 text-xs text-gray-500">
                  After the scenario rewards are granted, the player is pushed
                  into inventory until they equip their new gear and use the
                  rewarded spellbook.
                </p>
              </div>

              <div className="mt-4">
                <DialogueMessageListEditor
                  label="Dialogue Lines"
                  helperText="Shown while the inventory drawer is forced open."
                  value={loadoutDialogue}
                  onChange={setLoadoutDialogue}
                />
              </div>
            </section>

            <section className="rounded-lg border border-gray-200 p-4">
              <div className="mb-4">
                <h2 className="text-sm font-semibold text-gray-900">
                  Post-Monster Dialogue
                </h2>
                <p className="mt-1 text-xs text-gray-500">
                  Shown after the tutorial monster celebration modal closes and
                  before the base-kit inventory step begins.
                </p>
              </div>

              <DialogueMessageListEditor
                label="Dialogue Lines"
                helperText="This uses the same dialogue presentation as the tutorial intro."
                value={postMonsterDialogue}
                onChange={setPostMonsterDialogue}
              />
            </section>

            <section className="rounded-lg border border-gray-200 p-4">
              <div className="mb-4">
                <h2 className="text-sm font-semibold text-gray-900">
                  Home Base Kit Step
                </h2>
                <p className="mt-1 text-xs text-gray-500">
                  Shown while the inventory drawer is forced open again and the
                  player must use the rewarded home base kit.
                </p>
              </div>

              <DialogueMessageListEditor
                label="Drawer Guidance"
                helperText="Use this to tell the player how to open and place their home base kit."
                value={baseKitDialogue}
                onChange={setBaseKitDialogue}
              />
            </section>

            <section className="rounded-lg border border-gray-200 p-4">
              <div className="mb-4">
                <h2 className="text-sm font-semibold text-gray-900">
                  Post-Base Placement Dialogue
                </h2>
                <p className="mt-1 text-xs text-gray-500">
                  Shown as soon as the player establishes their base. After
                  this dialogue closes, the tutorial requires them to use their
                  hearth before the final conversation can begin.
                </p>
              </div>

              <DialogueMessageListEditor
                label="Dialogue Lines"
                helperText="Use this to direct the player into their new base and point them toward the hearth."
                value={postBasePlacementDialogue}
                onChange={setPostBasePlacementDialogue}
              />
            </section>

            <section className="rounded-lg border border-gray-200 p-4">
              <div className="mb-4">
                <h2 className="text-sm font-semibold text-gray-900">
                  Hearth Objective
                </h2>
                <p className="mt-1 text-xs text-gray-500">
                  After the post-base placement dialogue, the player must use
                  their base&apos;s hearth to heal before the final tutorial
                  dialogue will appear.
                </p>
              </div>

              <div className="rounded-lg border border-dashed border-gray-300 bg-gray-50 px-4 py-3 text-sm text-gray-600">
                This step uses the in-game base UI and does not need separate
                dialogue lines.
              </div>
            </section>

            <section className="rounded-lg border border-gray-200 p-4">
              <div className="mb-4">
                <h2 className="text-sm font-semibold text-gray-900">
                  Home Base Quest
                </h2>
                <p className="mt-1 text-xs text-gray-500">
                  After the player places their base, the tutorial can clone a
                  questgiver and generate a private live quest from the selected
                  archetype near that base.
                </p>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <SearchableSelect
                  label="Quest Archetype"
                  placeholder="Search quest archetype name…"
                  options={questArchetypeOptions}
                  value={baseQuestArchetypeId}
                  onChange={setBaseQuestArchetypeId}
                />
                <SearchableSelect
                  label="Questgiver Source Type"
                  placeholder="Choose source type…"
                  options={[
                    {
                      value: 'character',
                      label: 'Live Character',
                    },
                    {
                      value: 'template',
                      label: 'Character Template',
                    },
                  ]}
                  value={baseQuestGiverSourceType}
                  onChange={(value) =>
                    setBaseQuestGiverSourceType(
                      value === 'template' ? 'template' : 'character'
                    )
                  }
                />
                <SearchableSelect
                  label={
                    baseQuestGiverSourceType === 'template'
                      ? 'Source Questgiver Template'
                      : 'Source Questgiver Character'
                  }
                  placeholder={
                    baseQuestGiverSourceType === 'template'
                      ? 'Search template name…'
                      : 'Search character name…'
                  }
                  options={
                    baseQuestGiverSourceType === 'template'
                      ? characterTemplateOptions
                      : characterOptions
                  }
                  value={
                    baseQuestGiverSourceType === 'template'
                      ? baseQuestGiverCharacterTemplateId
                      : baseQuestGiverCharacterId
                  }
                  onChange={(value) => {
                    if (baseQuestGiverSourceType === 'template') {
                      setBaseQuestGiverCharacterTemplateId(value);
                      setBaseQuestGiverCharacterId('');
                    } else {
                      setBaseQuestGiverCharacterId(value);
                      setBaseQuestGiverCharacterTemplateId('');
                    }
                  }}
                />
              </div>

              <div className="mt-4 rounded-lg border border-dashed border-gray-300 bg-gray-50 p-4">
                <div className="mb-3">
                  <h3 className="text-sm font-medium text-gray-900">
                    Run For User
                  </h3>
                  <p className="mt-1 text-xs text-gray-500">
                    Saves this tutorial config, then queues the follow-up home
                    base quest for a user who already has a base.
                  </p>
                </div>

                <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto] md:items-end">
                  <SearchableSelect
                    label="User"
                    placeholder="Search username, name, or phone…"
                    options={userOptions}
                    value={baseQuestPreviewUserId}
                    onChange={setBaseQuestPreviewUserId}
                  />
                  <button
                    type="button"
                    onClick={handleQueueBaseQuestForUser}
                    disabled={
                      saving ||
                      generatingImage ||
                      queueingBaseQuest ||
                      referenceDataLoading
                    }
                    className="rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {queueingBaseQuest ? 'Queueing…' : 'Queue Follow-Up Quest'}
                  </button>
                </div>

                {referenceDataLoading ? (
                  <p className="mt-2 text-xs text-gray-500">
                    Loading user options…
                  </p>
                ) : null}
              </div>
            </section>

            <section className="rounded-lg border border-gray-200 p-4">
              <div className="mb-4">
                <h2 className="text-sm font-semibold text-gray-900">
                  Post-Base Dialogue
                </h2>
                <p className="mt-1 text-xs text-gray-500">
                  Shown after the player uses their hearth and before the
                  Welcome to Unclaimed Streets overlay appears.
                </p>
              </div>

              <DialogueMessageListEditor
                label="Dialogue Lines"
                helperText="This is the final guided conversation before the player is released into the district."
                value={postBaseDialogue}
                onChange={setPostBaseDialogue}
              />
            </section>

            <section className="rounded-lg border border-gray-200 p-4">
              <div className="mb-4">
                <h2 className="text-sm font-semibold text-gray-900">
                  Tutorial Monster
                </h2>
                <p className="mt-1 text-xs text-gray-500">
                  When the loadout step is complete, the selected encounter is
                  cloned at the player&apos;s location as a private tutorial
                  fight.
                </p>
              </div>

              <div className="grid gap-4 md:grid-cols-3">
                <div className="md:col-span-3">
                  <SearchableSelect
                    label="Monster Encounter"
                    placeholder="Search encounter name…"
                    options={monsterEncounterOptions}
                    value={monsterEncounterId}
                    onChange={setMonsterEncounterId}
                  />
                  {referenceDataLoading ? (
                    <p className="mt-1 text-xs text-gray-500">
                      Loading encounter options…
                    </p>
                  ) : null}
                </div>
                <label className="text-sm">
                  <span className="block font-medium text-gray-700">
                    Reward Experience
                  </span>
                  <input
                    type="number"
                    min="0"
                    value={monsterRewardExperience}
                    onChange={(event) =>
                      setMonsterRewardExperience(
                        Number.parseInt(event.target.value || '0', 10)
                      )
                    }
                    className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                  />
                </label>
                <label className="text-sm">
                  <span className="block font-medium text-gray-700">
                    Reward Gold
                  </span>
                  <input
                    type="number"
                    min="0"
                    value={monsterRewardGold}
                    onChange={(event) =>
                      setMonsterRewardGold(
                        Number.parseInt(event.target.value || '0', 10)
                      )
                    }
                    className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                  />
                </label>
              </div>

              <div className="mt-4 flex flex-col gap-3">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-sm font-medium text-gray-800">
                      Item Rewards
                    </h3>
                    <p className="text-xs text-gray-500">
                      If set, these override the encounter&apos;s default item
                      rewards.
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={addMonsterItemReward}
                    className="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 hover:bg-gray-50"
                  >
                    Add Item
                  </button>
                </div>

                {monsterItemRewards.length === 0 && (
                  <div className="rounded-md border border-dashed border-gray-300 bg-gray-50 px-3 py-4 text-sm text-gray-500">
                    No monster item rewards configured.
                  </div>
                )}

                {monsterItemRewards.map((reward) => (
                  <div
                    key={reward.id}
                    className="rounded-md border border-gray-200 bg-white p-3"
                  >
                    <SearchableSelect
                      label="Inventory Item"
                      placeholder="Search item name…"
                      options={itemOptions}
                      value={reward.inventoryItemId}
                      onChange={(value) =>
                        updateMonsterItemReward(reward.id, {
                          inventoryItemId: value,
                        })
                      }
                    />
                    <div className="mt-3 flex items-end gap-3">
                      <div className="flex-1">
                        <label className="block text-sm font-medium text-gray-700">
                          Quantity
                        </label>
                        <input
                          type="number"
                          min="1"
                          value={reward.quantity}
                          onChange={(event) =>
                            updateMonsterItemReward(reward.id, {
                              quantity: Number.parseInt(
                                event.target.value || '1',
                                10
                              ),
                            })
                          }
                          className="mt-1 block w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                        />
                      </div>
                      <button
                        type="button"
                        onClick={() => removeMonsterItemReward(reward.id)}
                        className="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-600 hover:bg-gray-100"
                      >
                        Remove
                      </button>
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
