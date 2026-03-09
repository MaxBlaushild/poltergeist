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
  damageAffinity: string;
  statusesToApply: SpellStatusTemplateForm[];
  statusesToRemove: string;
  effectData: string;
};

type SpellFormState = {
  name: string;
  description: string;
  iconUrl: string;
  abilityType: 'spell' | 'technique';
  abilityLevel: string;
  effectText: string;
  schoolOfMagic: string;
  manaCost: string;
  effects: SpellEffectForm[];
};

type BulkAbilityStatus = {
  jobId: string;
  status: string;
  source?: string;
  abilityType?: string;
  targetLevel?: number;
  totalCount: number;
  createdCount: number;
  effectCounts?: BulkEffectCountsPayload;
  error?: string;
  queuedAt?: string;
  startedAt?: string;
  completedAt?: string;
  updatedAt?: string;
};

type PromptSpellProgressionStatus = {
  jobId: string;
  status: string;
  prompt: string;
  abilityType?: string;
  createdCount: number;
  progressionId?: string;
  seedSpellId?: string;
  createdSpellIds?: string[];
  error?: string;
  queuedAt?: string;
  startedAt?: string;
  completedAt?: string;
  updatedAt?: string;
};

type BulkEffectCountsPayload = {
  dealDamage: number;
  dealDamageAllEnemies: number;
  restoreLifePartyMember: number;
  restoreLifeAllPartyMembers: number;
  applyBeneficialStatuses: number;
  removeDetrimentalStatuses: number;
};

type BulkEffectCountsForm = {
  dealDamage: string;
  dealDamageAllEnemies: string;
  restoreLifePartyMember: string;
  restoreLifeAllPartyMembers: string;
  applyBeneficialStatuses: string;
  removeDetrimentalStatuses: string;
};

const knownEffectTypes = [
  'deal_damage',
  'deal_damage_all_enemies',
  'restore_life_party_member',
  'restore_life_all_party_members',
  'revive_party_member',
  'revive_all_downed_party_members',
  'apply_beneficial_statuses',
  'remove_detrimental_statuses',
] as const;

const damageAffinityOptions = [
  'physical',
  'fire',
  'ice',
  'lightning',
  'poison',
  'arcane',
  'holy',
  'shadow',
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
  damageAffinity: 'physical',
  statusesToApply: [],
  statusesToRemove: '',
  effectData: '',
});

const emptyForm = (): SpellFormState => ({
  name: '',
  description: '',
  iconUrl: '',
  abilityType: 'spell',
  abilityLevel: '1',
  effectText: '',
  schoolOfMagic: '',
  manaCost: '0',
  effects: [emptyEffect()],
});

const parseIntSafe = (value: string, fallback = 0): number => {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const DEFAULT_BULK_ABILITY_COUNT = '8';

const buildSuggestedBulkEffectCounts = (total: number): BulkEffectCountsForm => {
  const clamped = Math.max(1, total);
  const weightedTypes: Array<{ key: keyof BulkEffectCountsPayload; weight: number }> = [
    { key: 'dealDamage', weight: 35 },
    { key: 'dealDamageAllEnemies', weight: 15 },
    { key: 'restoreLifePartyMember', weight: 18 },
    { key: 'restoreLifeAllPartyMembers', weight: 10 },
    { key: 'applyBeneficialStatuses', weight: 12 },
    { key: 'removeDetrimentalStatuses', weight: 10 },
  ];
  const totalWeight = weightedTypes.reduce((sum, entry) => sum + entry.weight, 0);
  const entries = weightedTypes.map((entry) => ({
    key: entry.key,
    count: Math.floor((clamped * entry.weight) / totalWeight),
    remainder: (clamped * entry.weight) % totalWeight,
  }));
  let assigned = entries.reduce((sum, entry) => sum + entry.count, 0);
  while (assigned < clamped) {
    entries.sort((a, b) => b.remainder - a.remainder);
    entries[assigned % entries.length].count += 1;
    assigned += 1;
  }
  const counts = entries.reduce<BulkEffectCountsPayload>(
    (acc, entry) => ({ ...acc, [entry.key]: entry.count }),
    {
      dealDamage: 0,
      dealDamageAllEnemies: 0,
      restoreLifePartyMember: 0,
      restoreLifeAllPartyMembers: 0,
      applyBeneficialStatuses: 0,
      removeDetrimentalStatuses: 0,
    }
  );

  return {
    dealDamage: String(counts.dealDamage),
    dealDamageAllEnemies: String(counts.dealDamageAllEnemies),
    restoreLifePartyMember: String(counts.restoreLifePartyMember),
    restoreLifeAllPartyMembers: String(counts.restoreLifeAllPartyMembers),
    applyBeneficialStatuses: String(counts.applyBeneficialStatuses),
    removeDetrimentalStatuses: String(counts.removeDetrimentalStatuses),
  };
};

const parseBulkEffectCounts = (counts: BulkEffectCountsForm): BulkEffectCountsPayload => ({
  dealDamage: Math.max(0, parseIntSafe(counts.dealDamage, 0)),
  dealDamageAllEnemies: Math.max(0, parseIntSafe(counts.dealDamageAllEnemies, 0)),
  restoreLifePartyMember: Math.max(0, parseIntSafe(counts.restoreLifePartyMember, 0)),
  restoreLifeAllPartyMembers: Math.max(0, parseIntSafe(counts.restoreLifeAllPartyMembers, 0)),
  applyBeneficialStatuses: Math.max(0, parseIntSafe(counts.applyBeneficialStatuses, 0)),
  removeDetrimentalStatuses: Math.max(0, parseIntSafe(counts.removeDetrimentalStatuses, 0)),
});

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
            damageAffinity: (effect.damageAffinity ?? 'physical').toString(),
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
    abilityType: spell.abilityType === 'technique' ? 'technique' : 'spell',
    abilityLevel: String(Math.max(1, spell.abilityLevel ?? 1)),
    effectText: spell.effectText ?? '',
    schoolOfMagic: spell.schoolOfMagic ?? '',
    manaCost: String(spell.abilityType === 'technique' ? 0 : (spell.manaCost ?? 0)),
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
      damageAffinity:
        effectType === 'deal_damage' || effectType === 'deal_damage_all_enemies'
          ? (effect.damageAffinity?.trim().toLowerCase() || 'physical')
          : undefined,
      statusesToApply,
      statusesToRemove,
      effectData,
    };
  });

  return {
    name: form.name.trim(),
    description: form.description.trim(),
    iconUrl: form.iconUrl.trim(),
    abilityType: form.abilityType,
    abilityLevel: Math.max(1, parseIntSafe(form.abilityLevel, 1)),
    effectText: form.effectText.trim(),
    schoolOfMagic: form.schoolOfMagic.trim(),
    manaCost: form.abilityType === 'technique' ? 0 : parseIntSafe(form.manaCost, 0),
    effects,
  };
};

export const Spells = () => {
  const { apiClient } = useAPI();

  const [loading, setLoading] = useState(true);
  const [spells, setSpells] = useState<Spell[]>([]);
  const [generatingIconSpellId, setGeneratingIconSpellId] = useState<string | null>(null);
  const [generatingProgressionSpellId, setGeneratingProgressionSpellId] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [error, setError] = useState<string | null>(null);

  const [showModal, setShowModal] = useState(false);
  const [editingSpell, setEditingSpell] = useState<Spell | null>(null);
  const [form, setForm] = useState<SpellFormState>(emptyForm());

  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [bulkAbilityCount, setBulkAbilityCount] = useState(DEFAULT_BULK_ABILITY_COUNT);
  const [bulkAbilityTargetLevel, setBulkAbilityTargetLevel] = useState('25');
  const [bulkAbilityType, setBulkAbilityType] = useState<'spell' | 'technique'>('spell');
  const [bulkAbilityBusy, setBulkAbilityBusy] = useState(false);
  const [bulkAbilityJob, setBulkAbilityJob] = useState<BulkAbilityStatus | null>(null);
  const [bulkAbilityError, setBulkAbilityError] = useState<string | null>(null);
  const [bulkAbilityMessage, setBulkAbilityMessage] = useState<string | null>(null);
  const [bulkEffectCounts, setBulkEffectCounts] = useState<BulkEffectCountsForm>(
    buildSuggestedBulkEffectCounts(parseIntSafe(DEFAULT_BULK_ABILITY_COUNT, 8))
  );
  const [progressionPrompt, setProgressionPrompt] = useState('');
  const [progressionPromptAbilityType, setProgressionPromptAbilityType] =
    useState<'spell' | 'technique'>('spell');
  const [progressionPromptBusy, setProgressionPromptBusy] = useState(false);
  const [progressionPromptError, setProgressionPromptError] = useState<string | null>(null);
  const [progressionPromptMessage, setProgressionPromptMessage] = useState<string | null>(null);
  const [progressionPromptJob, setProgressionPromptJob] =
    useState<PromptSpellProgressionStatus | null>(null);

  const load = useCallback(async (suppressLoading = false) => {
    try {
      if (!suppressLoading) {
        setLoading(true);
      }
      setError(null);
      const response = await apiClient.get<Spell[]>('/sonar/spells');
      setSpells(Array.isArray(response) ? response : []);
    } catch (err) {
      console.error('Failed to load spells', err);
      setError('Failed to load spells.');
    } finally {
      if (!suppressLoading) {
        setLoading(false);
      }
    }
  }, [apiClient]);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    const hasPendingGeneration = spells.some((spell) =>
      ['queued', 'in_progress'].includes(spell.imageGenerationStatus || '')
    );
    if (!hasPendingGeneration) return;

    const interval = setInterval(() => {
      void load(true);
    }, 5000);

    return () => clearInterval(interval);
  }, [spells, load]);

  const filtered = useMemo(() => {
    const query = search.trim().toLowerCase();
    if (!query) return spells;
    return spells.filter((spell) => {
      return (
        spell.name.toLowerCase().includes(query) ||
        (spell.abilityType ?? 'spell').toLowerCase().includes(query) ||
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

  const handleGenerateIcon = async (spell: Spell) => {
    try {
      setGeneratingIconSpellId(spell.id);
      const updated = await apiClient.post<Spell>(`/sonar/spells/${spell.id}/generate-icon`, {});
      setSpells((prev) => prev.map((current) => (current.id === spell.id ? updated : current)));
    } catch (err) {
      console.error('Failed to generate spell icon', err);
      alert('Failed to queue spell icon generation.');
    } finally {
      setGeneratingIconSpellId(null);
    }
  };

  const handleGenerateProgression = async (spell: Spell) => {
    try {
      setGeneratingProgressionSpellId(spell.id);
      const result = await apiClient.post<{ createdCount?: number }>(
        `/sonar/spells/${spell.id}/generate-progression`,
        {}
      );
      const createdCount = typeof result?.createdCount === 'number' ? result.createdCount : 0;
      if (createdCount > 0) {
        setBulkAbilityMessage(`Generated ${createdCount} progression spell(s) from ${spell.name}.`);
      } else {
        setBulkAbilityMessage(`No missing progression bands for ${spell.name}.`);
      }
      setBulkAbilityError(null);
      await load(true);
    } catch (err) {
      console.error('Failed to generate spell progression', err);
      setBulkAbilityError(
        err instanceof Error ? err.message : 'Failed to generate spell progression.'
      );
    } finally {
      setGeneratingProgressionSpellId(null);
    }
  };

  const refreshBulkAbilityJobStatus = useCallback(
    async (jobId: string) => {
      try {
        const status = await apiClient.get<BulkAbilityStatus>(
          `/sonar/spells/bulk-generate/${jobId}/status`
        );
        setBulkAbilityJob(status);
        if (status.status === 'completed') {
          setBulkAbilityBusy(false);
          setBulkAbilityError(null);
          setBulkAbilityMessage(
            `Created ${status.createdCount} ${status.abilityType === 'technique' ? 'technique' : 'spell'}(s).`
          );
          await load(true);
        } else if (status.status === 'failed') {
          setBulkAbilityBusy(false);
          setBulkAbilityError(
            status.error || `Failed to generate ${status.abilityType === 'technique' ? 'techniques' : 'spells'}.`
          );
        }
      } catch (err) {
        console.error('Failed to refresh bulk ability status', err);
      }
    },
    [apiClient, load]
  );

  const refreshProgressionPromptJobStatus = useCallback(
    async (jobId: string, abilityType: 'spell' | 'technique') => {
      try {
        const path =
          abilityType === 'technique'
            ? `/sonar/techniques/progression-generate/${jobId}/status`
            : `/sonar/spells/progression-generate/${jobId}/status`;
        const status = await apiClient.get<PromptSpellProgressionStatus>(
          path
        );
        setProgressionPromptJob(status);
        const resolvedType =
          (status.abilityType === 'technique' ? 'technique' : 'spell') as
            | 'spell'
            | 'technique';
        if (status.status === 'completed') {
          setProgressionPromptBusy(false);
          setProgressionPromptError(null);
          setProgressionPromptMessage(
            `Created ${status.createdCount} ${resolvedType === 'technique' ? 'technique' : 'spell'}(s) across a progression.`
          );
          await load(true);
        } else if (status.status === 'failed') {
          setProgressionPromptBusy(false);
          setProgressionPromptError(
            status.error ||
              `Failed to generate ${resolvedType === 'technique' ? 'technique' : 'spell'} progression.`
          );
        }
      } catch (err) {
        console.error('Failed to refresh spell progression prompt job status', err);
      }
    },
    [apiClient, load]
  );

  const handleGenerateProgressionFromPrompt = async () => {
    const trimmedPrompt = progressionPrompt.trim();
    if (trimmedPrompt.length < 12) {
      setProgressionPromptError('Prompt must be at least 12 characters.');
      return;
    }
    if (trimmedPrompt.length > 2000) {
      setProgressionPromptError('Prompt must be at most 2000 characters.');
      return;
    }

    try {
      setProgressionPromptBusy(true);
      setProgressionPromptError(null);
      setProgressionPromptMessage(null);
      setProgressionPromptJob(null);

      const abilityType = progressionPromptAbilityType;
      const path =
        abilityType === 'technique'
          ? '/sonar/techniques/progression-generate'
          : '/sonar/spells/progression-generate';
      const response = await apiClient.post<PromptSpellProgressionStatus>(
        path,
        { prompt: trimmedPrompt, abilityType }
      );
      const resolvedType =
        (response.abilityType === 'technique' ? 'technique' : abilityType) as
          | 'spell'
          | 'technique';
      setProgressionPromptJob({ ...response, abilityType: resolvedType });
      if (response.status === 'completed') {
        setProgressionPromptBusy(false);
        setProgressionPromptMessage(
          `Created ${response.createdCount} ${resolvedType === 'technique' ? 'technique' : 'spell'}(s) across a progression.`
        );
        await load(true);
      } else if (response.status === 'failed') {
        setProgressionPromptBusy(false);
        setProgressionPromptError(
          response.error ||
            `Failed to generate ${resolvedType === 'technique' ? 'technique' : 'spell'} progression.`
        );
      }
    } catch (err) {
      console.error('Failed to queue spell progression prompt generation', err);
      setProgressionPromptBusy(false);
      setProgressionPromptError(
        err instanceof Error
          ? err.message
          : `Failed to queue ${progressionPromptAbilityType === 'technique' ? 'technique' : 'spell'} progression generation.`
      );
    }
  };

  const handleBulkGenerateAbilities = async () => {
    const count = Number.parseInt(bulkAbilityCount, 10);
    if (!Number.isFinite(count) || count < 1 || count > 100) {
      setBulkAbilityError('Count must be between 1 and 100.');
      return;
    }
    const targetLevel = Number.parseInt(bulkAbilityTargetLevel, 10);
    if (!Number.isFinite(targetLevel) || targetLevel < 1 || targetLevel > 100) {
      setBulkAbilityError('Target level must be between 1 and 100.');
      return;
    }
    const effectCounts = parseBulkEffectCounts(bulkEffectCounts);
    const totalConfiguredCount = Object.values(effectCounts).reduce((sum, value) => sum + value, 0);
    if (totalConfiguredCount !== count) {
      setBulkAbilityError(`Effect counts must add up to ${count}.`);
      return;
    }

    try {
      setBulkAbilityBusy(true);
      setBulkAbilityError(null);
      setBulkAbilityMessage(null);
      setBulkAbilityJob(null);

      const path =
        bulkAbilityType === 'technique'
          ? '/sonar/techniques/bulk-generate'
          : '/sonar/spells/bulk-generate';
      const response = await apiClient.post<BulkAbilityStatus>(path, {
        count,
        abilityType: bulkAbilityType,
        targetLevel,
        effectCounts,
      });
      setBulkAbilityJob(response);
      if (response.status === 'completed') {
        setBulkAbilityBusy(false);
        setBulkAbilityMessage(
          `Created ${response.createdCount} ${bulkAbilityType === 'technique' ? 'technique' : 'spell'}(s).`
        );
        await load(true);
      } else if (response.status === 'failed') {
        setBulkAbilityBusy(false);
        setBulkAbilityError(
          response.error ||
            `Failed to generate ${bulkAbilityType === 'technique' ? 'techniques' : 'spells'}.`
        );
      }
    } catch (err) {
      console.error('Failed to bulk generate abilities', err);
      setBulkAbilityBusy(false);
      setBulkAbilityError(
        err instanceof Error
          ? err.message
          : `Failed to generate ${bulkAbilityType === 'technique' ? 'techniques' : 'spells'}.`
      );
    }
  };

  const formatGenerationStatus = (status?: string) => {
    switch ((status || '').trim()) {
      case 'queued':
        return 'Queued';
      case 'in_progress':
        return 'In Progress';
      case 'complete':
        return 'Complete';
      case 'failed':
        return 'Failed';
      default:
        return 'None';
    }
  };

  const configuredEffectCountTotal = useMemo(
    () =>
      Object.values(parseBulkEffectCounts(bulkEffectCounts)).reduce(
        (sum, value) => sum + value,
        0
      ),
    [bulkEffectCounts]
  );

  useEffect(() => {
    if (!bulkAbilityJob?.jobId) {
      return;
    }
    if (
      bulkAbilityJob.status !== 'queued' &&
      bulkAbilityJob.status !== 'in_progress'
    ) {
      return;
    }
    const interval = window.setInterval(() => {
      void refreshBulkAbilityJobStatus(bulkAbilityJob.jobId);
    }, 3000);
    return () => window.clearInterval(interval);
  }, [bulkAbilityJob, refreshBulkAbilityJobStatus]);

  useEffect(() => {
    if (!progressionPromptJob?.jobId) {
      return;
    }
    if (
      progressionPromptJob.status !== 'queued' &&
      progressionPromptJob.status !== 'in_progress'
    ) {
      return;
    }
    const interval = window.setInterval(() => {
      const abilityType =
        progressionPromptJob.abilityType === 'technique' ? 'technique' : 'spell';
      void refreshProgressionPromptJobStatus(progressionPromptJob.jobId, abilityType);
    }, 3000);
    return () => window.clearInterval(interval);
  }, [progressionPromptJob, refreshProgressionPromptJobStatus]);

  return (
    <div className="p-6 bg-gray-100 min-h-screen">
      <div className="max-w-7xl mx-auto">
        <div className="qa-card mb-6">
          <div className="flex items-center justify-between gap-3">
            <div>
              <h1 className="qa-card-title">Spells & Techniques</h1>
              <p className="text-sm text-gray-600">Manage ability definitions and effect payloads.</p>
            </div>
            <div className="flex flex-wrap items-end gap-2">
              <label className="text-xs text-gray-600">
                Count
                <input
                  type="number"
                  min={1}
                  max={100}
                  value={bulkAbilityCount}
                  onChange={(e) => setBulkAbilityCount(e.target.value)}
                  className="mt-1 w-24 rounded-md border border-gray-300 px-2 py-2 text-sm"
                  aria-label="Bulk ability count"
                />
              </label>
              <label className="text-xs text-gray-600">
                Target Level
                <input
                  type="number"
                  min={1}
                  max={100}
                  value={bulkAbilityTargetLevel}
                  onChange={(e) => setBulkAbilityTargetLevel(e.target.value)}
                  className="mt-1 w-28 rounded-md border border-gray-300 px-2 py-2 text-sm"
                  aria-label="Bulk ability target level"
                  placeholder="Target lvl"
                />
              </label>
              <label className="text-xs text-gray-600">
                Ability Type
                <select
                  className="mt-1 rounded-md border border-gray-300 px-2 py-2 text-sm"
                  value={bulkAbilityType}
                  onChange={(e) =>
                    setBulkAbilityType(e.target.value === 'technique' ? 'technique' : 'spell')
                  }
                >
                  <option value="spell">Spells</option>
                  <option value="technique">Techniques</option>
                </select>
              </label>
              <button
                className="qa-btn qa-btn-secondary"
                onClick={handleBulkGenerateAbilities}
                disabled={bulkAbilityBusy}
              >
                {bulkAbilityBusy ? 'Generating...' : 'Generate Bulk'}
              </button>
              <button className="qa-btn qa-btn-primary" onClick={openCreate}>
                Create Ability
              </button>
            </div>
          </div>
          <div className="mt-4 grid grid-cols-2 md:grid-cols-6 gap-2">
            <label className="text-xs text-gray-600">
              Deal Damage Count
              <input
                type="number"
                min={0}
                max={100}
                value={bulkEffectCounts.dealDamage}
                onChange={(e) =>
                  setBulkEffectCounts((prev) => ({ ...prev, dealDamage: e.target.value }))
                }
                className="mt-1 w-full rounded-md border border-gray-300 px-2 py-1 text-sm"
              />
            </label>
            <label className="text-xs text-gray-600">
              AoE Damage Count
              <input
                type="number"
                min={0}
                max={100}
                value={bulkEffectCounts.dealDamageAllEnemies}
                onChange={(e) =>
                  setBulkEffectCounts((prev) => ({
                    ...prev,
                    dealDamageAllEnemies: e.target.value,
                  }))
                }
                className="mt-1 w-full rounded-md border border-gray-300 px-2 py-1 text-sm"
              />
            </label>
            <label className="text-xs text-gray-600">
              Heal One Ally Count
              <input
                type="number"
                min={0}
                max={100}
                value={bulkEffectCounts.restoreLifePartyMember}
                onChange={(e) =>
                  setBulkEffectCounts((prev) => ({
                    ...prev,
                    restoreLifePartyMember: e.target.value,
                  }))
                }
                className="mt-1 w-full rounded-md border border-gray-300 px-2 py-1 text-sm"
              />
            </label>
            <label className="text-xs text-gray-600">
              Heal All Allies Count
              <input
                type="number"
                min={0}
                max={100}
                value={bulkEffectCounts.restoreLifeAllPartyMembers}
                onChange={(e) =>
                  setBulkEffectCounts((prev) => ({
                    ...prev,
                    restoreLifeAllPartyMembers: e.target.value,
                  }))
                }
                className="mt-1 w-full rounded-md border border-gray-300 px-2 py-1 text-sm"
              />
            </label>
            <label className="text-xs text-gray-600">
              Apply Buff Status Count
              <input
                type="number"
                min={0}
                max={100}
                value={bulkEffectCounts.applyBeneficialStatuses}
                onChange={(e) =>
                  setBulkEffectCounts((prev) => ({
                    ...prev,
                    applyBeneficialStatuses: e.target.value,
                  }))
                }
                className="mt-1 w-full rounded-md border border-gray-300 px-2 py-1 text-sm"
              />
            </label>
            <label className="text-xs text-gray-600">
              Remove Debuffs Count
              <input
                type="number"
                min={0}
                max={100}
                value={bulkEffectCounts.removeDetrimentalStatuses}
                onChange={(e) =>
                  setBulkEffectCounts((prev) => ({
                    ...prev,
                    removeDetrimentalStatuses: e.target.value,
                  }))
                }
                className="mt-1 w-full rounded-md border border-gray-300 px-2 py-1 text-sm"
              />
            </label>
          </div>
          <p className="mt-2 text-xs text-gray-500">
            Configure exact counts per effect type. Total configured must equal the bulk count.
          </p>
          <div className="mt-2 flex items-center gap-3 text-xs text-gray-600">
            <span>
              Configured: {configuredEffectCountTotal}/{Number.parseInt(bulkAbilityCount, 10) || 0}
            </span>
            <button
              className="qa-btn qa-btn-secondary"
              onClick={() =>
                setBulkEffectCounts(
                  buildSuggestedBulkEffectCounts(Math.max(1, Number.parseInt(bulkAbilityCount, 10) || 1))
                )
              }
              disabled={bulkAbilityBusy}
            >
              Auto-Fill Counts
            </button>
          </div>
          {bulkAbilityJob && (
            <div className="mt-3 flex flex-wrap items-center gap-3 text-sm text-gray-700">
              <span className="font-semibold uppercase tracking-wide">
                {bulkAbilityJob.status.replace('_', ' ')}
              </span>
              <span>
                Type: {bulkAbilityJob.abilityType === 'technique' ? 'Technique' : 'Spell'}
              </span>
              {typeof bulkAbilityJob.targetLevel === 'number' ? (
                <span>Target Level: {bulkAbilityJob.targetLevel}</span>
              ) : null}
              <span>
                Progress: {bulkAbilityJob.createdCount}/{bulkAbilityJob.totalCount}
              </span>
              <span>Job: {bulkAbilityJob.jobId}</span>
              {bulkAbilityJob.updatedAt ? <span>Updated: {bulkAbilityJob.updatedAt}</span> : null}
            </div>
          )}
          <div className="mt-4 border-t border-gray-200 pt-4">
            <div className="text-sm font-semibold text-gray-800">
              Generate Full Progression From Prompt
            </div>
            <p className="mt-1 text-xs text-gray-600">
              Describe one idea and generate linked level bands (10/25/50/70) for spells or techniques.
            </p>
            <div className="mt-2">
              <label className="text-xs text-gray-600">
                Ability Type
                <select
                  className="mt-1 rounded-md border border-gray-300 px-2 py-2 text-sm"
                  value={progressionPromptAbilityType}
                  onChange={(e) =>
                    setProgressionPromptAbilityType(e.target.value === 'technique' ? 'technique' : 'spell')
                  }
                  disabled={progressionPromptBusy}
                >
                  <option value="spell">Spell</option>
                  <option value="technique">Technique</option>
                </select>
              </label>
            </div>
            <textarea
              className="mt-2 w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
              rows={3}
              placeholder={
                progressionPromptAbilityType === 'technique'
                  ? 'Example: A precise spear style that grows from quick thrusts into a battlefield-cleaving master form.'
                  : 'Example: A fire spell that starts as a tiny ember and evolves into an explosive inferno.'
              }
              value={progressionPrompt}
              onChange={(e) => setProgressionPrompt(e.target.value)}
              disabled={progressionPromptBusy}
            />
            <div className="mt-2 flex items-center gap-2">
              <button
                className="qa-btn qa-btn-secondary"
                onClick={handleGenerateProgressionFromPrompt}
                disabled={progressionPromptBusy}
              >
                {progressionPromptBusy
                  ? 'Generating...'
                  : `Generate ${progressionPromptAbilityType === 'technique' ? 'Technique' : 'Spell'} Progression`}
              </button>
              {progressionPromptJob ? (
                <span className="text-xs text-gray-600">
                  Job {progressionPromptJob.jobId} ·{' '}
                  {(progressionPromptJob.abilityType === 'technique' ? 'Technique' : 'Spell')} ·{' '}
                  {progressionPromptJob.status.replace('_', ' ')}
                </span>
              ) : null}
            </div>
            {progressionPromptMessage ? (
              <p className="mt-2 text-sm text-emerald-700">{progressionPromptMessage}</p>
            ) : null}
            {progressionPromptError ? (
              <p className="mt-2 text-sm text-red-700">{progressionPromptError}</p>
            ) : null}
          </div>
          {bulkAbilityMessage ? (
            <p className="mt-2 text-sm text-emerald-700">{bulkAbilityMessage}</p>
          ) : null}
          {bulkAbilityError ? (
            <p className="mt-2 text-sm text-red-700">{bulkAbilityError}</p>
          ) : null}
        </div>

        <div className="qa-card mb-6">
          <input
            className="block w-full border border-gray-300 rounded-md p-2"
            placeholder="Search by name, type, school, or effect text..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>

        {loading ? (
          <div className="qa-card">Loading abilities...</div>
        ) : error ? (
          <div className="qa-card text-red-600">{error}</div>
        ) : filtered.length === 0 ? (
          <div className="qa-card text-gray-600">No abilities found.</div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
            {filtered.map((spell) => {
              const progressionLink = spell.progressionLinks?.[0];
              return (
                <div key={spell.id} className="qa-card">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <div className="text-lg font-semibold">{spell.name}</div>
                      <div className="text-sm text-gray-600">
                        {(spell.abilityType ?? 'spell') === 'technique'
                          ? `${spell.schoolOfMagic} · Lvl ${Math.max(1, spell.abilityLevel ?? 1)} · Technique`
                          : `${spell.schoolOfMagic} · Lvl ${Math.max(1, spell.abilityLevel ?? 1)} · Mana ${spell.manaCost}`}
                      </div>
                      <div className="text-xs text-gray-500 mt-1">
                        Icon Status: {formatGenerationStatus(spell.imageGenerationStatus)}
                      </div>
                      {progressionLink ? (
                        <div className="text-xs text-gray-500 mt-1">
                          Spell Progression: {progressionLink.progression?.name ?? progressionLink.progressionId} ·
                          {' '}Level Band {progressionLink.levelBand}
                        </div>
                      ) : null}
                      {spell.imageGenerationStatus === 'failed' && spell.imageGenerationError ? (
                        <div className="text-xs text-red-600 mt-1">Error: {spell.imageGenerationError}</div>
                      ) : null}
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
                    <button
                      className="qa-btn qa-btn-secondary"
                      onClick={() => handleGenerateIcon(spell)}
                      disabled={
                        generatingIconSpellId === spell.id ||
                        ['queued', 'in_progress'].includes(spell.imageGenerationStatus || '')
                      }
                    >
                      {generatingIconSpellId === spell.id ? 'Queueing...' : 'Generate Icon'}
                    </button>
                    <button
                      className="qa-btn qa-btn-secondary"
                      onClick={() => handleGenerateProgression(spell)}
                      disabled={
                        generatingProgressionSpellId === spell.id ||
                        (spell.abilityType ?? 'spell') === 'technique'
                      }
                    >
                      {generatingProgressionSpellId === spell.id ? 'Generating...' : 'Generate Level Bands'}
                    </button>
                    <button className="qa-btn qa-btn-danger" onClick={() => setDeleteId(spell.id)}>
                      Delete
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}

        {showModal && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
            <div className="bg-white w-full max-w-5xl rounded-lg shadow-lg max-h-[92vh] overflow-y-auto">
              <div className="p-5 border-b flex items-center justify-between">
                <h2 className="text-xl font-semibold">
                  {editingSpell ? `Edit ${editingSpell.name}` : 'Create Ability'}
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
                    Ability Type
                    <select
                      className="w-full border rounded-md p-2"
                      value={form.abilityType}
                      onChange={(e) =>
                        setForm((prev) => ({
                          ...prev,
                          abilityType: e.target.value === 'technique' ? 'technique' : 'spell',
                          manaCost: e.target.value === 'technique' ? '0' : prev.manaCost,
                        }))
                      }
                    >
                      <option value="spell">Spell</option>
                      <option value="technique">Technique</option>
                    </select>
                  </label>
                  <label className="text-sm">
                    Mana Cost
                    <input
                      className="w-full border rounded-md p-2"
                      type="number"
                      min={0}
                      value={form.manaCost}
                      disabled={form.abilityType === 'technique'}
                      onChange={(e) => setForm((prev) => ({ ...prev, manaCost: e.target.value }))}
                    />
                  </label>
                  <label className="text-sm">
                    Ability Level
                    <input
                      className="w-full border rounded-md p-2"
                      type="number"
                      min={1}
                      value={form.abilityLevel}
                      onChange={(e) =>
                        setForm((prev) => ({
                          ...prev,
                          abilityLevel: e.target.value,
                        }))
                      }
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
                        <div className="grid grid-cols-1 md:grid-cols-4 gap-3 mb-3">
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
                          {normalizeEffectType(effect) === 'deal_damage' ||
                          normalizeEffectType(effect) === 'deal_damage_all_enemies' ? (
                            <label className="text-sm">
                              Damage Affinity
                              <select
                                className="w-full border rounded-md p-2"
                                value={effect.damageAffinity || 'physical'}
                                onChange={(e) =>
                                  updateEffect(effectIndex, {
                                    damageAffinity: e.target.value,
                                  })
                                }
                              >
                                {damageAffinityOptions.map((affinity) => (
                                  <option key={affinity} value={affinity}>
                                    {affinity}
                                  </option>
                                ))}
                              </select>
                            </label>
                          ) : (
                            <div />
                          )}
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
