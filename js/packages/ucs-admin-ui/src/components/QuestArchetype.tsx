import React, { useMemo, useState, useEffect, useCallback, useRef } from 'react';
import { useAPI, useZoneContext } from '@poltergeist/contexts';
import {
  QuestArchetypeDraft,
  QuestTemplateGeneratorDraft,
  useQuestArchetypes,
} from '../contexts/questArchetypes.tsx';
import {
  DialogueMessage,
  LocationArchetype,
  QuestArchetype,
  QuestArchetypeNode,
  QuestArchetypeNodeLocationSelectionMode,
  QuestArchetypeChallenge,
  QuestArchetypeChallengeTemplate,
  QuestGenerationJob,
  QuestArchetypeNodeType,
  QuestDifficultyMode,
  InventoryItem,
  Spell,
  Character,
} from '@poltergeist/types';
import {
  MaterialRewardsEditor,
  emptyMaterialReward,
  normalizeMaterialRewards,
  summarizeMaterialRewards,
} from './MaterialRewardsEditor.tsx';
import './questArchetypeTheme.css';
import { useSearchParams } from 'react-router-dom';
import { DialogueMessageListEditor } from './DialogueMessageListEditor.tsx';

interface FlowNodeProps {
  node: QuestArchetypeNode;
  locationArchetypes: LocationArchetype[];
  monsterTemplates: MonsterTemplateRecord[];
  scenarioTemplates: ScenarioTemplateRecord[];
  challengeTemplates: ChallengeTemplateRecord[];
  characters: Character[];
  inventoryItems: InventoryItem[];
  spells: Spell[];
  depth: number;
  addChallengeToQuestArchetype: (
    questArchetypeId: string,
    proficiency?: string | null,
    unlockedNode?: QuestArchetypeNodeDraft | null,
    challengeTemplateId?: string | null
  ) => void;
  onSaveNode: (nodeId: string, updates: QuestArchetypeNodeDraft) => void;
  onEditChallenge: (
    challenge: QuestArchetypeChallenge,
    allowsTemplate?: boolean
  ) => void;
}

type MonsterTemplateRecord = {
  id: string;
  name: string;
  monsterType?: 'monster' | 'boss' | 'raid';
  imageUrl?: string;
  thumbnailUrl?: string;
  archived?: boolean;
  description?: string;
  baseStrength?: number;
  baseDexterity?: number;
  baseConstitution?: number;
  baseIntelligence?: number;
  baseWisdom?: number;
  baseCharisma?: number;
  strongAgainstAffinity?: string | null;
  weakAgainstAffinity?: string | null;
  progressions?: MonsterTemplateProgressionRecord[];
  spells?: Spell[];
  imageGenerationStatus?: string;
  imageGenerationError?: string | null;
};

type MonsterTemplateProgressionRecord = {
  id: string;
  name: string;
  abilityType?: 'spell' | 'technique' | string;
};

type MonsterAbilityProgressionOption = {
  id: string;
  name: string;
  abilityType: 'spell' | 'technique';
  memberCount: number;
};

type ScenarioTemplateRecord = {
  id: string;
  prompt: string;
  imageUrl?: string;
  thumbnailUrl?: string;
  scaleWithUserLevel?: boolean;
  rewardMode?: 'explicit' | 'random';
  randomRewardSize?: 'small' | 'medium' | 'large';
  difficulty?: number | null;
  rewardExperience?: number;
  rewardGold?: number;
  openEnded?: boolean;
  failurePenaltyMode?: 'shared' | 'individual';
  failureHealthDrainType?: 'none' | 'flat' | 'percent';
  failureHealthDrainValue?: number;
  failureManaDrainType?: 'none' | 'flat' | 'percent';
  failureManaDrainValue?: number;
  failureStatuses?: unknown[];
  successRewardMode?: 'shared' | 'individual';
  successHealthRestoreType?: 'none' | 'flat' | 'percent';
  successHealthRestoreValue?: number;
  successManaRestoreType?: 'none' | 'flat' | 'percent';
  successManaRestoreValue?: number;
  successStatuses?: unknown[];
  options?: unknown[];
  itemRewards?: unknown[];
  itemChoiceRewards?: unknown[];
  spellRewards?: unknown[];
  createdAt?: string;
  updatedAt?: string;
};

type ChallengeTemplateRecord = QuestArchetypeChallengeTemplate & {
  locationArchetype?: LocationArchetype | null;
  description?: string;
  imageUrl?: string;
  thumbnailUrl?: string;
  scaleWithUserLevel?: boolean;
  rewardMode?: 'explicit' | 'random';
  randomRewardSize?: 'small' | 'medium' | 'large';
  rewardExperience?: number;
  reward?: number;
  inventoryItemId?: number | null;
  itemChoiceRewards?: unknown[];
  submissionType?: 'photo' | 'text' | 'video' | string;
  difficulty?: number | null;
  statTags?: string[];
  proficiency?: string | null;
  createdAt?: string;
  updatedAt?: string;
};

type QuestArchetypeInlineEditorState =
  | {
      kind: 'challenge';
      templateId: string;
      sourceLabel: string;
      lockedLocationArchetypeId?: string | null;
    }
  | {
      kind: 'scenario';
      templateId: string;
      sourceLabel: string;
    }
  | {
      kind: 'monster';
      templateId: string;
      sourceLabel: string;
    };

type ChallengeTemplateFormState = {
  locationArchetypeId: string;
  question: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  scaleWithUserLevel: boolean;
  rewardMode: 'explicit' | 'random';
  randomRewardSize: 'small' | 'medium' | 'large';
  rewardExperience: string;
  reward: string;
  inventoryItemId: string;
  itemChoiceRewardsJson: string;
  submissionType: 'photo' | 'text' | 'video';
  difficulty: string;
  statTags: string;
  proficiency: string;
};

type ScenarioTemplateFormState = {
  prompt: string;
  imageUrl: string;
  thumbnailUrl: string;
  scaleWithUserLevel: boolean;
  rewardMode: 'explicit' | 'random';
  randomRewardSize: 'small' | 'medium' | 'large';
  difficulty: string;
  rewardExperience: string;
  rewardGold: string;
  openEnded: boolean;
  failurePenaltyMode: 'shared' | 'individual';
  failureHealthDrainType: 'none' | 'flat' | 'percent';
  failureHealthDrainValue: string;
  failureManaDrainType: 'none' | 'flat' | 'percent';
  failureManaDrainValue: string;
  failureStatusesJson: string;
  successRewardMode: 'shared' | 'individual';
  successHealthRestoreType: 'none' | 'flat' | 'percent';
  successHealthRestoreValue: string;
  successManaRestoreType: 'none' | 'flat' | 'percent';
  successManaRestoreValue: string;
  successStatusesJson: string;
  optionsJson: string;
  itemRewardsJson: string;
  itemChoiceRewardsJson: string;
  spellRewardsJson: string;
};

type MonsterTemplateFormState = {
  monsterType: 'monster' | 'boss' | 'raid';
  name: string;
  description: string;
  imageUrl: string;
  thumbnailUrl: string;
  baseStrength: string;
  baseDexterity: string;
  baseConstitution: string;
  baseIntelligence: string;
  baseWisdom: string;
  baseCharisma: string;
  strongAgainstAffinity: string;
  weakAgainstAffinity: string;
  spellProgressionIds: string[];
  techniqueProgressionIds: string[];
};

type PaginatedResponse<T> = {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
};

const describeChallengeTemplate = (
  template:
    | ChallengeTemplateRecord
    | QuestArchetypeChallengeTemplate
    | null
    | undefined,
  locationArchetypes: LocationArchetype[]
) => {
  if (!template) {
    return 'Challenge template';
  }
  const locationLabel =
    locationArchetypes.find(
      (entry) => entry.id === template.locationArchetypeId
    )?.name ?? 'Unknown location';
  const question = template.question?.trim() || 'Untitled challenge';
  return `${question} @ ${locationLabel}`;
};

const resolveChallengeTemplateForChallenge = (
  challenge: QuestArchetypeChallenge,
  challengeTemplates: ChallengeTemplateRecord[]
) => {
  if (challenge.challengeTemplateId) {
    const latestTemplate = challengeTemplates.find(
      (template) => template.id === challenge.challengeTemplateId
    );
    if (latestTemplate) {
      return latestTemplate;
    }
  }
  return challenge.challengeTemplate ?? null;
};

const resolveChallengeProficiency = (
  challenge: QuestArchetypeChallenge,
  challengeTemplates: ChallengeTemplateRecord[]
) => {
  const template = resolveChallengeTemplateForChallenge(
    challenge,
    challengeTemplates
  );
  return template?.proficiency ?? challenge.proficiency ?? '';
};

const validateQuestArchetypeNodeEditor = (
  editor: QuestArchetypeNodeEditorState,
  sourceLabel: string
) => {
  if (
    editor.nodeType === 'challenge' &&
    editor.challengeTemplateId.trim().length === 0
  ) {
    return `${sourceLabel} challenge nodes require a challenge template.`;
  }
  return '';
};

const extractApiErrorMessage = (error: unknown, fallback: string) => {
  if (
    error &&
    typeof error === 'object' &&
    'response' in error &&
    error.response &&
    typeof error.response === 'object' &&
    'data' in error.response &&
    error.response.data &&
    typeof error.response.data === 'object' &&
    'error' in error.response.data &&
    typeof error.response.data.error === 'string'
  ) {
    return error.response.data.error;
  }
  return fallback;
};

const prettyJson = (value: unknown): string =>
  JSON.stringify(value ?? [], null, 2);

const parseIntegerString = (value: string, fallback = 0): number => {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const parseJsonField = <T,>(label: string, value: string): T => {
  try {
    return JSON.parse(value) as T;
  } catch {
    throw new Error(`${label} must be valid JSON.`);
  }
};

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

const emptyChallengeTemplateForm = (): ChallengeTemplateFormState => ({
  locationArchetypeId: '',
  question: '',
  description: '',
  imageUrl: '',
  thumbnailUrl: '',
  scaleWithUserLevel: false,
  rewardMode: 'random',
  randomRewardSize: 'small',
  rewardExperience: '0',
  reward: '0',
  inventoryItemId: '',
  itemChoiceRewardsJson: '[]',
  submissionType: 'photo',
  difficulty: '0',
  statTags: '',
  proficiency: '',
});

const challengeTemplateFormFromRecord = (
  record: ChallengeTemplateRecord
): ChallengeTemplateFormState => ({
  locationArchetypeId: record.locationArchetypeId ?? '',
  question: record.question ?? '',
  description: record.description ?? '',
  imageUrl: record.imageUrl ?? '',
  thumbnailUrl: record.thumbnailUrl ?? '',
  scaleWithUserLevel: Boolean(record.scaleWithUserLevel),
  rewardMode: record.rewardMode === 'explicit' ? 'explicit' : 'random',
  randomRewardSize:
    record.randomRewardSize === 'medium' || record.randomRewardSize === 'large'
      ? record.randomRewardSize
      : 'small',
  rewardExperience: String(record.rewardExperience ?? 0),
  reward: String(record.reward ?? 0),
  inventoryItemId:
    record.inventoryItemId !== undefined && record.inventoryItemId !== null
      ? String(record.inventoryItemId)
      : '',
  itemChoiceRewardsJson: prettyJson(record.itemChoiceRewards),
  submissionType:
    record.submissionType === 'text' || record.submissionType === 'video'
      ? record.submissionType
      : 'photo',
  difficulty: String(record.difficulty ?? 0),
  statTags: (record.statTags ?? []).join(', '),
  proficiency: record.proficiency ?? '',
});

const buildChallengeTemplatePayloadFromForm = (
  form: ChallengeTemplateFormState
) => ({
  locationArchetypeId: form.locationArchetypeId,
  question: form.question.trim(),
  description: form.description.trim(),
  imageUrl: form.imageUrl.trim(),
  thumbnailUrl: form.thumbnailUrl.trim(),
  scaleWithUserLevel: form.scaleWithUserLevel,
  rewardMode: form.rewardMode,
  randomRewardSize: form.randomRewardSize,
  rewardExperience: parseIntegerString(form.rewardExperience, 0),
  reward: parseIntegerString(form.reward, 0),
  inventoryItemId: form.inventoryItemId.trim()
    ? parseIntegerString(form.inventoryItemId, 0)
    : null,
  itemChoiceRewards: parseJsonField<unknown[]>(
    'Challenge item choice rewards',
    form.itemChoiceRewardsJson
  ),
  submissionType: form.submissionType,
  difficulty: parseIntegerString(form.difficulty, 0),
  statTags: form.statTags
    .split(',')
    .map((entry) => entry.trim().toLowerCase())
    .filter(Boolean),
  proficiency: form.proficiency.trim(),
});

const emptyScenarioTemplateForm = (): ScenarioTemplateFormState => ({
  prompt: '',
  imageUrl: '',
  thumbnailUrl: '',
  scaleWithUserLevel: false,
  rewardMode: 'random',
  randomRewardSize: 'small',
  difficulty: '24',
  rewardExperience: '0',
  rewardGold: '0',
  openEnded: false,
  failurePenaltyMode: 'shared',
  failureHealthDrainType: 'none',
  failureHealthDrainValue: '0',
  failureManaDrainType: 'none',
  failureManaDrainValue: '0',
  failureStatusesJson: '[]',
  successRewardMode: 'shared',
  successHealthRestoreType: 'none',
  successHealthRestoreValue: '0',
  successManaRestoreType: 'none',
  successManaRestoreValue: '0',
  successStatusesJson: '[]',
  optionsJson: '[]',
  itemRewardsJson: '[]',
  itemChoiceRewardsJson: '[]',
  spellRewardsJson: '[]',
});

const scenarioTemplateFormFromRecord = (
  record: ScenarioTemplateRecord
): ScenarioTemplateFormState => ({
  prompt: record.prompt ?? '',
  imageUrl: record.imageUrl ?? '',
  thumbnailUrl: record.thumbnailUrl ?? '',
  scaleWithUserLevel: Boolean(record.scaleWithUserLevel),
  rewardMode: record.rewardMode === 'explicit' ? 'explicit' : 'random',
  randomRewardSize:
    record.randomRewardSize === 'medium' || record.randomRewardSize === 'large'
      ? record.randomRewardSize
      : 'small',
  difficulty: String(record.difficulty ?? 0),
  rewardExperience: String(record.rewardExperience ?? 0),
  rewardGold: String(record.rewardGold ?? 0),
  openEnded: Boolean(record.openEnded),
  failurePenaltyMode:
    record.failurePenaltyMode === 'individual' ? 'individual' : 'shared',
  failureHealthDrainType:
    record.failureHealthDrainType === 'flat' ||
    record.failureHealthDrainType === 'percent'
      ? record.failureHealthDrainType
      : 'none',
  failureHealthDrainValue: String(record.failureHealthDrainValue ?? 0),
  failureManaDrainType:
    record.failureManaDrainType === 'flat' ||
    record.failureManaDrainType === 'percent'
      ? record.failureManaDrainType
      : 'none',
  failureManaDrainValue: String(record.failureManaDrainValue ?? 0),
  failureStatusesJson: prettyJson(record.failureStatuses),
  successRewardMode:
    record.successRewardMode === 'individual' ? 'individual' : 'shared',
  successHealthRestoreType:
    record.successHealthRestoreType === 'flat' ||
    record.successHealthRestoreType === 'percent'
      ? record.successHealthRestoreType
      : 'none',
  successHealthRestoreValue: String(record.successHealthRestoreValue ?? 0),
  successManaRestoreType:
    record.successManaRestoreType === 'flat' ||
    record.successManaRestoreType === 'percent'
      ? record.successManaRestoreType
      : 'none',
  successManaRestoreValue: String(record.successManaRestoreValue ?? 0),
  successStatusesJson: prettyJson(record.successStatuses),
  optionsJson: prettyJson(record.options),
  itemRewardsJson: prettyJson(record.itemRewards),
  itemChoiceRewardsJson: prettyJson(record.itemChoiceRewards),
  spellRewardsJson: prettyJson(record.spellRewards),
});

const buildScenarioTemplatePayloadFromForm = (
  form: ScenarioTemplateFormState
) => ({
  prompt: form.prompt.trim(),
  imageUrl: form.imageUrl.trim(),
  thumbnailUrl: form.thumbnailUrl.trim(),
  scaleWithUserLevel: form.scaleWithUserLevel,
  rewardMode: form.rewardMode,
  randomRewardSize: form.randomRewardSize,
  difficulty: parseIntegerString(form.difficulty, 0),
  rewardExperience: parseIntegerString(form.rewardExperience, 0),
  rewardGold: parseIntegerString(form.rewardGold, 0),
  openEnded: form.openEnded,
  failurePenaltyMode: form.failurePenaltyMode,
  failureHealthDrainType: form.failureHealthDrainType,
  failureHealthDrainValue: parseIntegerString(form.failureHealthDrainValue, 0),
  failureManaDrainType: form.failureManaDrainType,
  failureManaDrainValue: parseIntegerString(form.failureManaDrainValue, 0),
  failureStatuses: parseJsonField<unknown[]>(
    'Scenario failure statuses',
    form.failureStatusesJson
  ),
  successRewardMode: form.successRewardMode,
  successHealthRestoreType: form.successHealthRestoreType,
  successHealthRestoreValue: parseIntegerString(
    form.successHealthRestoreValue,
    0
  ),
  successManaRestoreType: form.successManaRestoreType,
  successManaRestoreValue: parseIntegerString(form.successManaRestoreValue, 0),
  successStatuses: parseJsonField<unknown[]>(
    'Scenario success statuses',
    form.successStatusesJson
  ),
  options: parseJsonField<unknown[]>('Scenario options', form.optionsJson),
  itemRewards: parseJsonField<unknown[]>(
    'Scenario item rewards',
    form.itemRewardsJson
  ),
  itemChoiceRewards: parseJsonField<unknown[]>(
    'Scenario item choice rewards',
    form.itemChoiceRewardsJson
  ),
  spellRewards: parseJsonField<unknown[]>(
    'Scenario spell rewards',
    form.spellRewardsJson
  ),
});

const emptyMonsterTemplateForm = (): MonsterTemplateFormState => ({
  monsterType: 'monster',
  name: '',
  description: '',
  imageUrl: '',
  thumbnailUrl: '',
  baseStrength: '10',
  baseDexterity: '10',
  baseConstitution: '10',
  baseIntelligence: '10',
  baseWisdom: '10',
  baseCharisma: '10',
  strongAgainstAffinity: '',
  weakAgainstAffinity: '',
  spellProgressionIds: [],
  techniqueProgressionIds: [],
});

const monsterTemplateFormFromRecord = (
  template: MonsterTemplateRecord
): MonsterTemplateFormState => ({
  monsterType: template.monsterType ?? 'monster',
  name: template.name ?? '',
  description: template.description ?? '',
  imageUrl: template.imageUrl ?? '',
  thumbnailUrl: template.thumbnailUrl ?? '',
  baseStrength: String(template.baseStrength ?? 10),
  baseDexterity: String(template.baseDexterity ?? 10),
  baseConstitution: String(template.baseConstitution ?? 10),
  baseIntelligence: String(template.baseIntelligence ?? 10),
  baseWisdom: String(template.baseWisdom ?? 10),
  baseCharisma: String(template.baseCharisma ?? 10),
  strongAgainstAffinity: template.strongAgainstAffinity ?? '',
  weakAgainstAffinity: template.weakAgainstAffinity ?? '',
  spellProgressionIds: (template.progressions ?? [])
    .filter(
      (progression) => (progression.abilityType ?? 'spell') !== 'technique'
    )
    .map((progression) => progression.id),
  techniqueProgressionIds: (template.progressions ?? [])
    .filter(
      (progression) => (progression.abilityType ?? 'spell') === 'technique'
    )
    .map((progression) => progression.id),
});

const buildMonsterTemplatePayloadFromForm = (
  form: MonsterTemplateFormState
) => ({
  monsterType: form.monsterType,
  name: form.name.trim(),
  description: form.description.trim(),
  imageUrl: form.imageUrl.trim(),
  thumbnailUrl: form.thumbnailUrl.trim(),
  baseStrength: parseIntegerString(form.baseStrength, 10),
  baseDexterity: parseIntegerString(form.baseDexterity, 10),
  baseConstitution: parseIntegerString(form.baseConstitution, 10),
  baseIntelligence: parseIntegerString(form.baseIntelligence, 10),
  baseWisdom: parseIntegerString(form.baseWisdom, 10),
  baseCharisma: parseIntegerString(form.baseCharisma, 10),
  strongAgainstAffinity: form.strongAgainstAffinity.trim(),
  weakAgainstAffinity: form.weakAgainstAffinity.trim(),
  progressionIds: Array.from(
    new Set([...form.spellProgressionIds, ...form.techniqueProgressionIds])
  ),
});

const isPendingQuestGenerationStatus = (status?: string | null) =>
  status === 'queued' || status === 'in_progress';

const questGenerationStatusChipClass = (status?: string | null) => {
  switch (status) {
    case 'completed':
      return 'qa-chip success';
    case 'failed':
      return 'qa-chip danger';
    case 'in_progress':
      return 'qa-chip accent';
    case 'queued':
      return 'qa-chip muted';
    default:
      return 'qa-chip muted';
  }
};

const formatQuestGenerationStatus = (status?: string | null) =>
  (status || 'queued').replace(/_/g, ' ');

const questArchetypeFormHasExplicitCopy = (form: QuestArchetypeFormState) =>
  form.name.trim().length > 0 &&
  form.description.trim().length > 0 &&
  form.acceptanceDialogue.some((line) => (line.text ?? '').trim().length > 0);

type QuestArchetypeNodeLocationMode =
  | 'point_of_interest'
  | 'coordinates'
  | 'ditto';

type QuestArchetypeNodeEditorState = {
  nodeType: QuestArchetypeNodeType;
  locationMode: QuestArchetypeNodeLocationMode;
  locationArchetypeId: string;
  locationArchetypeQuery: string;
  locationSelectionMode: 'random' | 'closest';
  challengeTemplateId: string;
  scenarioTemplateId: string;
  fetchCharacterId: string;
  fetchRequirements: Array<{ inventoryItemId: string; quantity: number }>;
  storyFlagKey: string;
  monsterTemplateIds: string[];
  targetLevel: number;
  encounterProximityMeters: number;
  expositionTitle: string;
  expositionDescription: string;
  expositionDialogue: DialogueMessage[];
  expositionRewardMode: RewardMode;
  expositionRandomRewardSize: RandomRewardSize;
  expositionRewardExperience: number;
  expositionRewardGold: number;
  expositionMaterialRewards: ReturnType<typeof emptyMaterialReward>[];
  expositionItemRewards: Array<{ inventoryItemId: string; quantity: number }>;
  expositionSpellRewards: Array<{ spellId: string }>;
};

const emptyNodeEditorState = (): QuestArchetypeNodeEditorState => ({
  nodeType: 'challenge',
  locationMode: 'point_of_interest',
  locationArchetypeId: '',
  locationArchetypeQuery: '',
  locationSelectionMode: 'random',
  challengeTemplateId: '',
  scenarioTemplateId: '',
  fetchCharacterId: '',
  fetchRequirements: [],
  storyFlagKey: '',
  monsterTemplateIds: [],
  targetLevel: 1,
  encounterProximityMeters: 100,
  expositionTitle: '',
  expositionDescription: '',
  expositionDialogue: [],
  expositionRewardMode: 'random',
  expositionRandomRewardSize: 'small',
  expositionRewardExperience: 0,
  expositionRewardGold: 0,
  expositionMaterialRewards: [],
  expositionItemRewards: [],
  expositionSpellRewards: [],
});

const buildNodeEditorState = (
  node: QuestArchetypeNode,
  locationArchetypes: LocationArchetype[]
): QuestArchetypeNodeEditorState => ({
  nodeType:
    node.nodeType === 'monster_encounter'
      ? 'monster_encounter'
      : node.nodeType === 'scenario'
        ? 'scenario'
        : node.nodeType === 'exposition'
          ? 'exposition'
          : node.nodeType === 'fetch_quest'
            ? 'fetch_quest'
          : node.nodeType === 'story_flag'
            ? 'story_flag'
            : 'challenge',
  locationMode:
    node.locationSelectionMode === 'same_as_previous'
      ? 'ditto'
      : node.locationArchetypeId
        ? 'point_of_interest'
        : 'coordinates',
  scenarioTemplateId: node.scenarioTemplateId ?? '',
  fetchCharacterId: node.fetchCharacterId ?? '',
  fetchRequirements: (node.fetchRequirements ?? []).map((requirement) => ({
    inventoryItemId:
      requirement.inventoryItemId != null
        ? String(requirement.inventoryItemId)
        : '',
    quantity: requirement.quantity ?? 1,
  })),
  storyFlagKey: node.storyFlagKey ?? '',
  locationArchetypeId: node.locationArchetypeId ?? '',
  locationArchetypeQuery:
    locationArchetypes.find((entry) => entry.id === node.locationArchetypeId)
      ?.name ?? '',
  locationSelectionMode:
    node.locationSelectionMode === 'closest' ? 'closest' : 'random',
  challengeTemplateId: node.challengeTemplateId ?? '',
  monsterTemplateIds: [...(node.monsterTemplateIds ?? [])],
  targetLevel: node.targetLevel ?? 1,
  encounterProximityMeters: node.encounterProximityMeters ?? 100,
  expositionTitle: node.expositionTitle ?? '',
  expositionDescription: node.expositionDescription ?? '',
  expositionDialogue: node.expositionDialogue ?? [],
  expositionRewardMode:
    node.expositionRewardMode === 'explicit' ? 'explicit' : 'random',
  expositionRandomRewardSize:
    node.expositionRandomRewardSize === 'medium' ||
    node.expositionRandomRewardSize === 'large'
      ? node.expositionRandomRewardSize
      : 'small',
  expositionRewardExperience: node.expositionRewardExperience ?? 0,
  expositionRewardGold: node.expositionRewardGold ?? 0,
  expositionMaterialRewards: (node.expositionMaterialRewards ?? []).map(
    (reward) => ({
      resourceKey: reward.resourceKey,
      amount: reward.amount,
    })
  ),
  expositionItemRewards: (node.expositionItemRewards ?? []).map((reward) => ({
    inventoryItemId:
      reward.inventoryItemId != null ? String(reward.inventoryItemId) : '',
    quantity: reward.quantity ?? 1,
  })),
  expositionSpellRewards: (node.expositionSpellRewards ?? []).map((reward) => ({
    spellId: reward.spellId ?? '',
  })),
});

const resolvedLocationSelectionModeForDraft = (
  state: QuestArchetypeNodeEditorState
): QuestArchetypeNodeLocationSelectionMode | undefined => {
  if (state.nodeType === 'story_flag') {
    return undefined;
  }
  if (state.locationMode === 'ditto') {
    return 'same_as_previous';
  }
  if (state.locationMode === 'point_of_interest') {
    return state.locationSelectionMode;
  }
  return undefined;
};

const resolvedLocationArchetypeIdForDraft = (
  state: QuestArchetypeNodeEditorState
): string | null => {
  if (state.nodeType === 'story_flag') {
    return null;
  }
  if (state.locationMode === 'point_of_interest') {
    return state.locationArchetypeId || null;
  }
  return null;
};

const buildNodeDraft = (
  state: QuestArchetypeNodeEditorState
): QuestArchetypeNodeDraft => ({
  nodeType: state.nodeType,
  locationArchetypeId: resolvedLocationArchetypeIdForDraft(state),
  locationSelectionMode: resolvedLocationSelectionModeForDraft(state),
  challengeTemplateId:
    state.nodeType === 'challenge' ? state.challengeTemplateId || null : null,
  scenarioTemplateId:
    state.nodeType === 'scenario' ? state.scenarioTemplateId || null : null,
  fetchCharacterId:
    state.nodeType === 'fetch_quest' ? state.fetchCharacterId || null : null,
  fetchRequirements:
    state.nodeType === 'fetch_quest'
      ? state.fetchRequirements
          .map((requirement) => ({
            inventoryItemId: Number(requirement.inventoryItemId) || 0,
            quantity: Number(requirement.quantity) || 0,
          }))
          .filter(
            (requirement) =>
              requirement.inventoryItemId > 0 && requirement.quantity > 0
          )
      : undefined,
  storyFlagKey:
    state.nodeType === 'story_flag' ? state.storyFlagKey.trim() : undefined,
  monsterTemplateIds:
    state.nodeType === 'monster_encounter'
      ? state.monsterTemplateIds
      : undefined,
  targetLevel:
    state.nodeType === 'monster_encounter'
      ? Number(state.targetLevel) || 1
      : undefined,
  encounterProximityMeters:
    state.nodeType === 'challenge' ||
    state.nodeType === 'monster_encounter' ||
    state.nodeType === 'scenario' ||
    state.nodeType === 'exposition' ||
    state.nodeType === 'fetch_quest'
      ? Number(state.encounterProximityMeters) || 0
      : undefined,
  expositionTitle:
    state.nodeType === 'exposition' ? state.expositionTitle.trim() : undefined,
  expositionDescription:
    state.nodeType === 'exposition'
      ? state.expositionDescription.trim()
      : undefined,
  expositionDialogue:
    state.nodeType === 'exposition'
      ? state.expositionDialogue.map((message, index) => ({
          speaker: message.speaker === 'user' ? 'user' : 'character',
          text: (message.text ?? '').trim(),
          order: index,
          effect: normalizeDialogueEffect(message.effect),
          characterId:
            message.speaker === 'user'
              ? undefined
              : message.characterId?.trim() || undefined,
        }))
      : undefined,
  expositionRewardMode:
    state.nodeType === 'exposition' ? state.expositionRewardMode : undefined,
  expositionRandomRewardSize:
    state.nodeType === 'exposition'
      ? state.expositionRandomRewardSize
      : undefined,
  expositionRewardExperience:
    state.nodeType === 'exposition'
      ? Number(state.expositionRewardExperience) || 0
      : undefined,
  expositionRewardGold:
    state.nodeType === 'exposition'
      ? Number(state.expositionRewardGold) || 0
      : undefined,
  expositionMaterialRewards:
    state.nodeType === 'exposition'
      ? normalizeMaterialRewards(state.expositionMaterialRewards)
      : undefined,
  expositionItemRewards:
    state.nodeType === 'exposition'
      ? state.expositionItemRewards
          .map((reward) => ({
            inventoryItemId: Number(reward.inventoryItemId) || 0,
            quantity: Number(reward.quantity) || 0,
          }))
          .filter(
            (reward) => reward.inventoryItemId > 0 && reward.quantity > 0
          )
      : undefined,
  expositionSpellRewards:
    state.nodeType === 'exposition'
      ? state.expositionSpellRewards
          .map((reward) => ({
            spellId: reward.spellId.trim(),
          }))
          .filter((reward) => reward.spellId.length > 0)
      : undefined,
});

const describeQuestArchetypeNode = (
  node: QuestArchetypeNode | undefined | null,
  locationArchetypes: LocationArchetype[],
  monsterTemplates: MonsterTemplateRecord[],
  scenarioTemplates: ScenarioTemplateRecord[]
) => {
  const locationLabelForNode = (() => {
    if (node?.locationSelectionMode === 'same_as_previous') {
      return 'Same as previous';
    }
    const locationLabel = locationArchetypes.find(
      (entry) => entry.id === node?.locationArchetypeId
    )?.name;
    if (locationLabel) {
      return locationLabel;
    }
    if (node?.locationArchetypeId) {
      return 'Point of interest';
    }
    return 'Coordinates';
  })();
  if (!node) {
    return 'Unknown';
  }
  if (node.nodeType === 'scenario') {
    const scenarioLabel =
      scenarioTemplates.find((entry) => entry.id === node.scenarioTemplateId)
        ?.prompt ?? 'Scenario';
    return `${scenarioLabel} @ ${locationLabelForNode}`;
  }
  if (node.nodeType === 'monster_encounter') {
    const names = (node.monsterTemplateIds ?? [])
      .map(
        (templateId) =>
          monsterTemplates.find((entry) => entry.id === templateId)?.name
      )
      .filter(Boolean) as string[];
    if (names.length === 0) {
      return `Monster encounter @ ${locationLabelForNode}`;
    }
    const encounterLabel = `Encounter: ${names.slice(0, 3).join(', ')}${names.length > 3 ? '…' : ''}`;
    return `${encounterLabel} @ ${locationLabelForNode}`;
  }
  if (node.nodeType === 'exposition') {
    const expositionLabel = node.expositionTitle?.trim() || 'Exposition';
    return `${expositionLabel} @ ${locationLabelForNode}`;
  }
  if (node.nodeType === 'fetch_quest') {
    const characterLabel = node.fetchCharacter?.name?.trim() || 'Character';
    const requirements = (node.fetchRequirements ?? [])
      .map((requirement) => {
        return `${requirement.quantity}x item ${requirement.inventoryItemId}`;
      })
      .slice(0, 2)
      .join(', ');
    const baseLabel = requirements
      ? `Deliver ${requirements}${(node.fetchRequirements?.length ?? 0) > 2 ? '…' : ''} to ${characterLabel}`
      : `Fetch quest for ${characterLabel}`;
    return `${baseLabel} @ ${locationLabelForNode}`;
  }
  if (node.nodeType === 'story_flag') {
    const storyFlagKey = node.storyFlagKey?.trim() || 'story flag';
    return `Story flag: ${storyFlagKey}`;
  }
  const challengeLabel = node.challengeTemplate?.question?.trim() || 'Challenge';
  return `${challengeLabel} @ ${locationLabelForNode}`;
};

type QuestArchetypeNodeConfigFieldsProps = {
  editor: QuestArchetypeNodeEditorState;
  setEditor: React.Dispatch<
    React.SetStateAction<QuestArchetypeNodeEditorState>
  >;
  allowSameAsPreviousLocationMode?: boolean;
  prefix: string;
  locationArchetypes: LocationArchetype[];
  challengeTemplates: ChallengeTemplateRecord[];
  monsterTemplates: MonsterTemplateRecord[];
  scenarioTemplates: ScenarioTemplateRecord[];
  characters: Character[];
  inventoryItems: InventoryItem[];
  spells: Spell[];
};

const QuestArchetypeNodeConfigFields: React.FC<
  QuestArchetypeNodeConfigFieldsProps
> = ({
  editor,
  setEditor,
  allowSameAsPreviousLocationMode = false,
  prefix,
  locationArchetypes,
  challengeTemplates,
  monsterTemplates,
  scenarioTemplates,
  characters,
  inventoryItems,
  spells,
}) => {
  const showsLocationConfig = editor.nodeType !== 'story_flag';
  const allowsDittoOption =
    allowSameAsPreviousLocationMode || editor.locationMode === 'ditto';
  const filteredLocationArchetypes = locationArchetypes
    .filter((archetype) =>
      archetype.name
        .toLowerCase()
        .includes(editor.locationArchetypeQuery.trim().toLowerCase())
    )
    .slice(0, 8);
  const availableChallengeTemplates = challengeTemplates;
  const characterOptions = characters.map((character) => ({
    value: character.id,
    label: character.name || character.id,
  }));

  const setSelectedLocationArchetype = (nextId: string, nextQuery: string) =>
    setEditor((prev) => ({
      ...prev,
      locationArchetypeId: nextId,
      locationArchetypeQuery: nextQuery,
    }));

  return (
    <>
      <div className="qa-field">
        <div className="qa-label">{prefix} Node Type</div>
        <select
          className="qa-select"
          value={editor.nodeType}
          onChange={(e) =>
            setEditor((prev) => ({
              ...prev,
              nodeType: e.target.value as QuestArchetypeNodeType,
            }))
          }
        >
          <option value="challenge">Challenge</option>
          <option value="scenario">Scenario</option>
          <option value="monster_encounter">Monster Encounter</option>
          <option value="exposition">Exposition</option>
          <option value="fetch_quest">Fetch Quest</option>
          <option value="story_flag">Story Flag</option>
        </select>
      </div>

      {showsLocationConfig ? (
        <>
          <div className="qa-field">
            <div className="qa-label">{prefix} Location Mode</div>
            <div className="qa-helper">
              {editor.locationMode === 'point_of_interest'
                ? 'Choose a location archetype, then decide whether generation should use the closest matching point of interest to the previous node or quest giver, or a random one in the zone.'
                : editor.locationMode === 'ditto'
                  ? 'Reuse the exact same map position as the previous node.'
                  : 'Place this node at generated coordinates near the previous node instead of at a point of interest.'}
            </div>
            <select
              className="qa-select"
              value={editor.locationMode}
              onChange={(e) =>
                setEditor((prev) => ({
                  ...prev,
                  locationMode:
                    e.target.value === 'coordinates'
                      ? 'coordinates'
                      : e.target.value === 'ditto'
                        ? 'ditto'
                      : 'point_of_interest',
                  locationArchetypeId:
                    e.target.value === 'coordinates'
                      ? ''
                      : prev.locationArchetypeId,
                  locationArchetypeQuery:
                    e.target.value === 'coordinates'
                      ? ''
                      : prev.locationArchetypeQuery,
                  locationSelectionMode:
                    e.target.value === 'coordinates' ||
                    e.target.value === 'ditto'
                      ? 'random'
                      : prev.locationSelectionMode,
                }))
              }
            >
              <option value="point_of_interest">Point Of Interest</option>
              <option value="coordinates">Coordinates</option>
              {allowsDittoOption ? (
                <option value="ditto">Same As Previous Node</option>
              ) : null}
            </select>
          </div>
          {editor.locationMode === 'point_of_interest' ? (
            <>
              <div className="qa-field">
                <div className="qa-label">{prefix} Location Archetype</div>
                <div className="qa-combobox">
                  <input
                    type="text"
                    className="qa-input"
                    value={editor.locationArchetypeQuery}
                    onChange={(e) => {
                      const value = e.target.value;
                      const matched = locationArchetypes.find(
                        (archetype) =>
                          archetype.name.toLowerCase() ===
                          value.trim().toLowerCase()
                      );
                      setSelectedLocationArchetype(
                        matched ? matched.id : '',
                        value
                      );
                    }}
                    placeholder="Search location archetypes..."
                  />
                  {editor.locationArchetypeQuery.trim().length > 0 && (
                    <div className="qa-combobox-list">
                      {filteredLocationArchetypes.length === 0 ? (
                        <div className="qa-combobox-empty">No matches.</div>
                      ) : (
                        filteredLocationArchetypes.map((archetype) => (
                          <button
                            key={`${prefix}-location-${archetype.id}`}
                            type="button"
                            className="qa-combobox-option"
                            onClick={() =>
                              setSelectedLocationArchetype(
                                archetype.id,
                                archetype.name
                              )
                            }
                          >
                            {archetype.name}
                          </button>
                        ))
                      )}
                    </div>
                  )}
                </div>
              </div>
              <div className="qa-field">
                <div className="qa-label">{prefix} POI Selection</div>
                <select
                  className="qa-select"
                  value={editor.locationSelectionMode}
                  onChange={(e) =>
                    setEditor((prev) => ({
                      ...prev,
                      locationSelectionMode:
                        e.target.value === 'closest' ? 'closest' : 'random',
                    }))
                  }
                >
                  <option value="random">Random In Zone</option>
                  <option value="closest">Closest To Previous / Questgiver</option>
                </select>
              </div>
            </>
          ) : (
            <div className="qa-field">
              <div className="qa-label">
                {prefix} Proximity To Previous Node (m)
              </div>
              <input
                type="number"
                min={0}
                className="qa-input"
                value={editor.encounterProximityMeters}
                onChange={(e) =>
                  setEditor((prev) => ({
                    ...prev,
                    encounterProximityMeters: parseInt(e.target.value) || 0,
                  }))
                }
              />
            </div>
          )}
        </>
      ) : null}

      {editor.nodeType === 'challenge' ? (
        <div className="qa-field">
          <div className="qa-label">{prefix} Challenge Template</div>
          <div className="qa-helper">
            Required. The template defines the challenge content, and the
            location settings above only control where the node appears.
          </div>
          <select
            className="qa-select"
            value={editor.challengeTemplateId}
            onChange={(e) =>
              setEditor((prev) => ({
                ...prev,
                challengeTemplateId: e.target.value,
              }))
            }
          >
            <option value="">Select a challenge template</option>
            {availableChallengeTemplates.map((template) => (
              <option key={`${prefix}-challenge-${template.id}`} value={template.id}>
                {describeChallengeTemplate(template, locationArchetypes)}
              </option>
            ))}
          </select>
          {availableChallengeTemplates.length === 0 ? (
            <div className="qa-helper">No challenge templates are available yet.</div>
          ) : null}
        </div>
      ) : null}

      {editor.nodeType === 'scenario' ? (
        <div className="qa-field">
          <div className="qa-label">{prefix} Scenario Template</div>
          <select
            className="qa-select"
            value={editor.scenarioTemplateId}
            onChange={(e) =>
              setEditor((prev) => ({
                ...prev,
                scenarioTemplateId: e.target.value,
              }))
            }
          >
            <option value="">Select a scenario template</option>
            {scenarioTemplates.map((template) => (
              <option
                key={`${prefix}-scenario-${template.id}`}
                value={template.id}
              >
                {template.prompt}
              </option>
            ))}
          </select>
        </div>
      ) : null}

      {editor.nodeType === 'monster_encounter' ? (
        <>
          <div className="qa-field">
            <div className="qa-label">{prefix} Monster Templates</div>
            <select
              className="qa-select"
              multiple
              size={6}
              value={editor.monsterTemplateIds}
              onChange={(e) =>
                setEditor((prev) => ({
                  ...prev,
                  monsterTemplateIds: Array.from(e.target.selectedOptions).map(
                    (option) => option.value
                  ),
                }))
              }
            >
              {monsterTemplates.map((template) => (
                <option key={template.id} value={template.id}>
                  {template.name}
                  {template.monsterType ? ` (${template.monsterType})` : ''}
                </option>
              ))}
            </select>
            <div className="qa-helper">
              Hold Command or Ctrl to select multiple monster templates.
            </div>
          </div>
          <div className="qa-field">
            <div className="qa-label">{prefix} Monster Level</div>
            <div className="qa-helper">
              Monster encounter target level is configured on the quest
              template, not per child node.
            </div>
          </div>
        </>
      ) : null}

      {editor.nodeType === 'exposition' ? (
        <>
          <div className="qa-field">
            <div className="qa-label">{prefix} Exposition Title</div>
            <input
              type="text"
              className="qa-input"
              value={editor.expositionTitle}
              onChange={(e) =>
                setEditor((prev) => ({
                  ...prev,
                  expositionTitle: e.target.value,
                }))
              }
              placeholder="Whispers Beneath the Overpass"
            />
          </div>
          <div className="qa-field">
            <div className="qa-label">{prefix} Exposition Description</div>
            <textarea
              className="qa-textarea"
              rows={3}
              value={editor.expositionDescription}
              onChange={(e) =>
                setEditor((prev) => ({
                  ...prev,
                  expositionDescription: e.target.value,
                }))
              }
              placeholder="Optional internal summary for the scene."
            />
          </div>
          <div className="qa-field">
            <DialogueMessageListEditor
              label={`${prefix} Dialogue`}
              helperText="Every line in an exposition needs a speaking character."
              value={editor.expositionDialogue}
              onChange={(value) =>
                setEditor((prev) => ({
                  ...prev,
                  expositionDialogue: value,
                }))
              }
              characterOptions={characterOptions}
              requireCharacterSelection
            />
          </div>
          <div className="qa-grid qa-grid-2">
            <div className="qa-field">
              <div className="qa-label">{prefix} Reward Mode</div>
              <select
                className="qa-select"
                value={editor.expositionRewardMode}
                onChange={(e) =>
                  setEditor((prev) => ({
                    ...prev,
                    expositionRewardMode:
                      e.target.value === 'explicit' ? 'explicit' : 'random',
                  }))
                }
              >
                <option value="random">Random Reward</option>
                <option value="explicit">Explicit Reward</option>
              </select>
            </div>
            {editor.expositionRewardMode === 'random' ? (
              <div className="qa-field">
                <div className="qa-label">{prefix} Random Reward Size</div>
                <select
                  className="qa-select"
                  value={editor.expositionRandomRewardSize}
                  onChange={(e) =>
                    setEditor((prev) => ({
                      ...prev,
                      expositionRandomRewardSize:
                        e.target.value === 'large'
                          ? 'large'
                          : e.target.value === 'medium'
                            ? 'medium'
                            : 'small',
                    }))
                  }
                >
                  <option value="small">Small</option>
                  <option value="medium">Medium</option>
                  <option value="large">Large</option>
                </select>
              </div>
            ) : null}
          </div>
          {editor.expositionRewardMode === 'explicit' ? (
            <>
              <div className="qa-grid qa-grid-2">
                <div className="qa-field">
                  <div className="qa-label">{prefix} Reward Experience</div>
                  <input
                    type="number"
                    min={0}
                    className="qa-input"
                    value={editor.expositionRewardExperience}
                    onChange={(e) =>
                      setEditor((prev) => ({
                        ...prev,
                        expositionRewardExperience:
                          parseInt(e.target.value) || 0,
                      }))
                    }
                  />
                </div>
                <div className="qa-field">
                  <div className="qa-label">{prefix} Reward Gold</div>
                  <input
                    type="number"
                    min={0}
                    className="qa-input"
                    value={editor.expositionRewardGold}
                    onChange={(e) =>
                      setEditor((prev) => ({
                        ...prev,
                        expositionRewardGold: parseInt(e.target.value) || 0,
                      }))
                    }
                  />
                </div>
              </div>
              <div className="qa-field">
                <div className="qa-label">{prefix} Material Rewards</div>
                <MaterialRewardsEditor
                  value={editor.expositionMaterialRewards}
                  onChange={(rewards) =>
                    setEditor((prev) => ({
                      ...prev,
                      expositionMaterialRewards: rewards,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">{prefix} Item Rewards</div>
                {editor.expositionItemRewards.length === 0 ? (
                  <div className="qa-empty">
                    No item rewards configured for this exposition.
                  </div>
                ) : (
                  <div className="qa-stack">
                    {editor.expositionItemRewards.map((reward, index) => (
                      <div
                        key={`${prefix}-exposition-item-${index}`}
                        className="qa-inline"
                        style={{ alignItems: 'flex-end' }}
                      >
                        <label className="qa-field" style={{ flex: 1 }}>
                          <div className="qa-label">Inventory Item</div>
                          <select
                            className="qa-select"
                            value={reward.inventoryItemId}
                            onChange={(e) =>
                              setEditor((prev) => ({
                                ...prev,
                                expositionItemRewards:
                                  prev.expositionItemRewards.map(
                                    (entry, entryIndex) =>
                                      entryIndex === index
                                        ? {
                                            ...entry,
                                            inventoryItemId: e.target.value,
                                          }
                                        : entry
                                  ),
                              }))
                            }
                          >
                            <option value="">Select an item</option>
                            {inventoryItems.map((item) => (
                              <option key={item.id} value={item.id}>
                                {item.name || item.id}
                              </option>
                            ))}
                          </select>
                        </label>
                        <label className="qa-field" style={{ width: 120 }}>
                          <div className="qa-label">Quantity</div>
                          <input
                            type="number"
                            min={1}
                            className="qa-input"
                            value={reward.quantity}
                            onChange={(e) =>
                              setEditor((prev) => ({
                                ...prev,
                                expositionItemRewards:
                                  prev.expositionItemRewards.map(
                                    (entry, entryIndex) =>
                                      entryIndex === index
                                        ? {
                                            ...entry,
                                            quantity: parseInt(e.target.value) || 1,
                                          }
                                        : entry
                                  ),
                              }))
                            }
                          />
                        </label>
                        <button
                          type="button"
                          className="qa-btn qa-btn-ghost"
                          onClick={() =>
                            setEditor((prev) => ({
                              ...prev,
                              expositionItemRewards:
                                prev.expositionItemRewards.filter(
                                  (_, entryIndex) => entryIndex !== index
                                ),
                            }))
                          }
                        >
                          Remove
                        </button>
                      </div>
                    ))}
                  </div>
                )}
                <button
                  type="button"
                  className="qa-btn qa-btn-outline"
                  style={{ marginTop: 12 }}
                  onClick={() =>
                    setEditor((prev) => ({
                      ...prev,
                      expositionItemRewards: [
                        ...prev.expositionItemRewards,
                        { inventoryItemId: '', quantity: 1 },
                      ],
                    }))
                  }
                >
                  Add Item Reward
                </button>
              </div>
              <div className="qa-field">
                <div className="qa-label">{prefix} Spell Rewards</div>
                {editor.expositionSpellRewards.length === 0 ? (
                  <div className="qa-empty">
                    No spell rewards configured for this exposition.
                  </div>
                ) : (
                  <div className="qa-stack">
                    {editor.expositionSpellRewards.map((reward, index) => (
                      <div
                        key={`${prefix}-exposition-spell-${index}`}
                        className="qa-inline"
                        style={{ alignItems: 'flex-end' }}
                      >
                        <label className="qa-field" style={{ flex: 1 }}>
                          <div className="qa-label">Spell</div>
                          <select
                            className="qa-select"
                            value={reward.spellId}
                            onChange={(e) =>
                              setEditor((prev) => ({
                                ...prev,
                                expositionSpellRewards:
                                  prev.expositionSpellRewards.map(
                                    (entry, entryIndex) =>
                                      entryIndex === index
                                        ? {
                                            ...entry,
                                            spellId: e.target.value,
                                          }
                                        : entry
                                  ),
                              }))
                            }
                          >
                            <option value="">Select a spell</option>
                            {spells.map((spell) => (
                              <option key={spell.id} value={spell.id}>
                                {spell.name || spell.id}
                              </option>
                            ))}
                          </select>
                        </label>
                        <button
                          type="button"
                          className="qa-btn qa-btn-ghost"
                          onClick={() =>
                            setEditor((prev) => ({
                              ...prev,
                              expositionSpellRewards:
                                prev.expositionSpellRewards.filter(
                                  (_, entryIndex) => entryIndex !== index
                                ),
                            }))
                          }
                        >
                          Remove
                        </button>
                      </div>
                    ))}
                  </div>
                )}
                <button
                  type="button"
                  className="qa-btn qa-btn-outline"
                  style={{ marginTop: 12 }}
                  onClick={() =>
                    setEditor((prev) => ({
                      ...prev,
                      expositionSpellRewards: [
                        ...prev.expositionSpellRewards,
                        { spellId: '' },
                      ],
                    }))
                  }
                >
                  Add Spell Reward
                </button>
              </div>
            </>
          ) : null}
        </>
      ) : null}

      {editor.nodeType === 'story_flag' ? (
        <div className="qa-field">
          <div className="qa-label">{prefix} Story Flag Key</div>
          <div className="qa-helper">
            This node completes automatically once the user has this story flag
            set to true.
          </div>
          <input
            type="text"
            className="qa-input"
            value={editor.storyFlagKey}
            onChange={(e) =>
              setEditor((prev) => ({
                ...prev,
                storyFlagKey: e.target.value,
              }))
            }
            placeholder="harbor_gate_open"
          />
        </div>
      ) : null}
      {editor.nodeType === 'fetch_quest' ? (
        <>
          <div className="qa-field">
            <div className="qa-label">{prefix} Target Character</div>
            <div className="qa-helper">
              The user must deliver the required items to this character to
              continue.
            </div>
            <select
              className="qa-select"
              value={editor.fetchCharacterId}
              onChange={(e) =>
                setEditor((prev) => ({
                  ...prev,
                  fetchCharacterId: e.target.value,
                }))
              }
            >
              <option value="">Select a character</option>
              {characters.map((character) => (
                <option
                  key={`${prefix}-fetch-character-${character.id}`}
                  value={character.id}
                >
                  {character.name || character.id}
                </option>
              ))}
            </select>
          </div>
          <div className="qa-field">
            <div className="qa-label">{prefix} Required Items</div>
            {editor.fetchRequirements.length === 0 ? (
              <div className="qa-empty">
                No delivery requirements configured yet.
              </div>
            ) : (
              <div className="qa-stack">
                {editor.fetchRequirements.map((requirement, index) => (
                  <div
                    key={`${prefix}-fetch-requirement-${index}`}
                    className="qa-inline"
                    style={{ alignItems: 'flex-end' }}
                  >
                    <label className="qa-field" style={{ flex: 1 }}>
                      <div className="qa-label">Inventory Item</div>
                      <select
                        className="qa-select"
                        value={requirement.inventoryItemId}
                        onChange={(e) =>
                          setEditor((prev) => ({
                            ...prev,
                            fetchRequirements: prev.fetchRequirements.map(
                              (entry, entryIndex) =>
                                entryIndex === index
                                  ? {
                                      ...entry,
                                      inventoryItemId: e.target.value,
                                    }
                                  : entry
                            ),
                          }))
                        }
                      >
                        <option value="">Select an item</option>
                        {inventoryItems.map((item) => (
                          <option key={item.id} value={item.id}>
                            {item.name || item.id}
                          </option>
                        ))}
                      </select>
                    </label>
                    <label className="qa-field" style={{ width: 120 }}>
                      <div className="qa-label">Quantity</div>
                      <input
                        type="number"
                        min={1}
                        className="qa-input"
                        value={requirement.quantity}
                        onChange={(e) =>
                          setEditor((prev) => ({
                            ...prev,
                            fetchRequirements: prev.fetchRequirements.map(
                              (entry, entryIndex) =>
                                entryIndex === index
                                  ? {
                                      ...entry,
                                      quantity: parseInt(e.target.value) || 1,
                                    }
                                  : entry
                            ),
                          }))
                        }
                      />
                    </label>
                    <button
                      type="button"
                      className="qa-btn qa-btn-ghost"
                      onClick={() =>
                        setEditor((prev) => ({
                          ...prev,
                          fetchRequirements: prev.fetchRequirements.filter(
                            (_, entryIndex) => entryIndex !== index
                          ),
                        }))
                      }
                    >
                      Remove
                    </button>
                  </div>
                ))}
              </div>
            )}
            <button
              type="button"
              className="qa-btn qa-btn-outline"
              style={{ marginTop: 12 }}
              onClick={() =>
                setEditor((prev) => ({
                  ...prev,
                  fetchRequirements: [
                    ...prev.fetchRequirements,
                    { inventoryItemId: '', quantity: 1 },
                  ],
                }))
              }
            >
              Add Required Item
            </button>
          </div>
        </>
      ) : null}
    </>
  );
};

const FlowNode: React.FC<FlowNodeProps> = ({
  node,
  locationArchetypes,
  monsterTemplates,
  scenarioTemplates,
  challengeTemplates,
  characters,
  inventoryItems,
  spells,
  depth,
  addChallengeToQuestArchetype,
  onSaveNode,
  onEditChallenge,
}) => {
  const borderColor =
    depth % 2 === 0 ? 'rgba(255, 107, 74, 0.4)' : 'rgba(95, 211, 181, 0.35)';
  const nodeSummary = describeQuestArchetypeNode(
    node,
    locationArchetypes,
    monsterTemplates,
    scenarioTemplates
  );
  const [isAdding, setIsAdding] = useState(false);
  const [nodeEditor, setNodeEditor] = useState<QuestArchetypeNodeEditorState>(
    buildNodeEditorState(node, locationArchetypes)
  );
  const [childEditor, setChildEditor] = useState<QuestArchetypeNodeEditorState>(
    emptyNodeEditorState()
  );
  const [childEnabled, setChildEnabled] = useState<boolean>(false);

  useEffect(() => {
    setNodeEditor(buildNodeEditorState(node, locationArchetypes));
  }, [locationArchetypes, node]);

  return (
    <div className="qa-flow-node" style={{ borderColor }}>
      <div className="qa-flow-node-card">
        <div className="qa-flow-node-header">
          <div>
            <div className="qa-flow-node-title">
              {depth === 0 ? 'Root Node' : `Node ${depth + 1}`}
            </div>
            <div className="qa-meta">{nodeSummary}</div>
          </div>
          <button
            className="qa-btn qa-btn-primary"
            onClick={() => setIsAdding((prev) => !prev)}
          >
            {isAdding ? 'Close' : 'Add Branch'}
          </button>
        </div>
        <div className="qa-flow-form" style={{ marginTop: 12 }}>
          <QuestArchetypeNodeConfigFields
            editor={nodeEditor}
            setEditor={setNodeEditor}
            allowSameAsPreviousLocationMode={depth > 0}
            prefix="Current"
            locationArchetypes={locationArchetypes}
            challengeTemplates={challengeTemplates}
            monsterTemplates={monsterTemplates}
            scenarioTemplates={scenarioTemplates}
            characters={characters}
            inventoryItems={inventoryItems}
            spells={spells}
          />
          <div className="qa-flow-form-actions">
            <button
              className="qa-btn qa-btn-outline"
              onClick={() =>
                setNodeEditor(buildNodeEditorState(node, locationArchetypes))
              }
            >
              Reset
            </button>
            <button
              className="qa-btn qa-btn-primary"
              onClick={() => {
                const validationError = validateQuestArchetypeNodeEditor(
                  nodeEditor,
                  'Current'
                );
                if (validationError) {
                  window.alert(validationError);
                  return;
                }
                onSaveNode(node.id, buildNodeDraft(nodeEditor));
              }}
            >
              Save Node
            </button>
          </div>
        </div>

        {isAdding && (
          <div className="qa-flow-form">
            <div className="qa-field">
              <label className="qa-inline" style={{ alignItems: 'center' }}>
                <input
                  type="checkbox"
                  checked={childEnabled}
                  onChange={(e) => setChildEnabled(e.target.checked)}
                />
                <span className="qa-label" style={{ marginBottom: 0 }}>
                  Unlock a child node
                </span>
              </label>
            </div>
            {childEnabled &&
              (
                <QuestArchetypeNodeConfigFields
                  editor={childEditor}
                  setEditor={setChildEditor}
                  allowSameAsPreviousLocationMode
                  prefix="Child"
                  locationArchetypes={locationArchetypes}
                  challengeTemplates={challengeTemplates}
                  monsterTemplates={monsterTemplates}
                  scenarioTemplates={scenarioTemplates}
                  characters={characters}
                  inventoryItems={inventoryItems}
                  spells={spells}
                />
              )}
            <div className="qa-flow-form-actions">
              <button
                className="qa-btn qa-btn-outline"
                onClick={() => {
                  setIsAdding(false);
                }}
              >
                Cancel
              </button>
              <button
                className="qa-btn qa-btn-primary"
                onClick={async () => {
                  const validationError =
                    childEnabled
                      ? validateQuestArchetypeNodeEditor(childEditor, 'Child')
                      : '';
                  if (validationError) {
                    window.alert(validationError);
                    return;
                  }
                  await addChallengeToQuestArchetype(
                    node.id,
                    null,
                    childEnabled ? buildNodeDraft(childEditor) : null,
                    null
                  );
                  setChildEnabled(false);
                  setChildEditor(emptyNodeEditorState());
                  setIsAdding(false);
                }}
              >
                Add Branch
              </button>
            </div>
          </div>
        )}

        {node.challenges && node.challenges.length > 0 ? (
          <div className="qa-flow-challenges">
            {node.challenges.map((challenge, index) => {
              const challengeTemplate = resolveChallengeTemplateForChallenge(
                challenge,
                challengeTemplates
              );
              const challengeProficiencyValue = resolveChallengeProficiency(
                challenge,
                challengeTemplates
              );
              return (
                <div key={challenge.id} className="qa-flow-challenge-card">
                  <div className="qa-flow-challenge-header">
                    <div>
                      <div className="qa-flow-challenge-title">
                        {`Branch ${index + 1}`}
                      </div>
                      {challengeTemplate && (
                        <div className="qa-meta" style={{ marginTop: 6 }}>
                          {describeChallengeTemplate(
                            challengeTemplate,
                            locationArchetypes
                          )}
                        </div>
                      )}
                      <div className="qa-inline" style={{ marginTop: 6 }}>
                        {challengeProficiencyValue && (
                          <span className="qa-chip muted">
                            Proficiency: {challengeProficiencyValue}
                          </span>
                        )}
                      </div>
                    </div>
                    <button
                      className="qa-btn qa-btn-ghost"
                      onClick={() =>
                        onEditChallenge(
                          challenge,
                          Boolean(
                            challenge.challengeTemplateId ||
                              challenge.challengeTemplate ||
                              challenge.proficiency
                          )
                        )
                      }
                    >
                      Edit
                    </button>
                  </div>

                  {challenge.unlockedNode ? (
                    <div className="qa-flow-branch">
                      <div className="qa-flow-branch-label">Unlocks</div>
                      <FlowNode
                        node={challenge.unlockedNode}
                        locationArchetypes={locationArchetypes}
                        monsterTemplates={monsterTemplates}
                        scenarioTemplates={scenarioTemplates}
                        challengeTemplates={challengeTemplates}
                        characters={characters}
                        inventoryItems={inventoryItems}
                        spells={spells}
                        depth={depth + 1}
                        addChallengeToQuestArchetype={
                          addChallengeToQuestArchetype
                        }
                        onSaveNode={onSaveNode}
                        onEditChallenge={onEditChallenge}
                      />
                    </div>
                  ) : (
                    <div className="qa-flow-branch qa-flow-branch-terminal">
                      <div className="qa-meta">No further node unlocked.</div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        ) : (
          <div className="qa-empty" style={{ marginTop: 12 }}>
            No branches yet. Add the first branch to define the flow.
          </div>
        )}
      </div>
    </div>
  );
};

type FlowMapNode = {
  id: string;
  depth: number;
  order: number;
  label: string;
};

type FlowMapEdge = {
  from: string;
  to: string;
};

type FlowMapLayout = {
  nodes: Array<FlowMapNode & { x: number; y: number }>;
  edges: Array<{ fromX: number; fromY: number; toX: number; toY: number }>;
  width: number;
  height: number;
};

const buildFlowMapLayout = (
  root: QuestArchetypeNode | undefined | null,
  locationArchetypes: LocationArchetype[],
  monsterTemplates: MonsterTemplateRecord[],
  scenarioTemplates: ScenarioTemplateRecord[]
): FlowMapLayout | null => {
  if (!root) return null;
  const nodes: FlowMapNode[] = [];
  const edges: FlowMapEdge[] = [];
  const visited = new Set<string>();
  let orderIndex = 0;
  let maxDepth = 0;

  const walk = (node: QuestArchetypeNode, depth: number) => {
    maxDepth = Math.max(maxDepth, depth);
    if (!visited.has(node.id)) {
      visited.add(node.id);
      nodes.push({
        id: node.id,
        depth,
        order: orderIndex,
        label: describeQuestArchetypeNode(
          node,
          locationArchetypes,
          monsterTemplates,
          scenarioTemplates
        ),
      });
      orderIndex += 1;
    }
    node.challenges?.forEach((challenge) => {
      if (!challenge.unlockedNode) return;
      edges.push({ from: node.id, to: challenge.unlockedNode.id });
      walk(challenge.unlockedNode, depth + 1);
    });
  };

  walk(root, 0);

  const xSpacing = 140;
  const ySpacing = 90;
  const padding = 32;
  const positionedNodes = nodes.map((node) => ({
    ...node,
    x: padding + node.depth * xSpacing,
    y: padding + node.order * ySpacing,
  }));
  const positions = new Map(positionedNodes.map((node) => [node.id, node]));
  const positionedEdges = edges
    .map((edge) => {
      const from = positions.get(edge.from);
      const to = positions.get(edge.to);
      if (!from || !to) return null;
      return {
        fromX: from.x,
        fromY: from.y,
        toX: to.x,
        toY: to.y,
      };
    })
    .filter(Boolean) as Array<{
    fromX: number;
    fromY: number;
    toX: number;
    toY: number;
  }>;

  const width = padding * 2 + Math.max(1, maxDepth + 1) * xSpacing;
  const height = padding * 2 + Math.max(1, nodes.length) * ySpacing;

  return {
    nodes: positionedNodes,
    edges: positionedEdges,
    width,
    height,
  };
};

type QuestFlowRouteNode = {
  id: string;
  node: QuestArchetypeNode;
  depth: number;
  order: number;
  pathLabel: string;
  parentNodeId: string | null;
  incomingChallenge: QuestArchetypeChallenge | null;
  incomingChallengeIndex: number | null;
};

const buildQuestFlowRoute = (
  root: QuestArchetypeNode | undefined | null
): QuestFlowRouteNode[] => {
  if (!root) {
    return [];
  }
  const routes: QuestFlowRouteNode[] = [];
  const visited = new Set<string>();

  const walk = (
    node: QuestArchetypeNode,
    depth: number,
    pathLabel: string,
    parentNodeId: string | null,
    incomingChallenge: QuestArchetypeChallenge | null,
    incomingChallengeIndex: number | null
  ) => {
    if (!node || visited.has(node.id)) {
      return;
    }
    visited.add(node.id);
    routes.push({
      id: node.id,
      node,
      depth,
      order: routes.length + 1,
      pathLabel,
      parentNodeId,
      incomingChallenge,
      incomingChallengeIndex,
    });
    node.challenges?.forEach((challenge, index) => {
      if (!challenge.unlockedNode) {
        return;
      }
      walk(
        challenge.unlockedNode,
        depth + 1,
        `${pathLabel}.${index + 1}`,
        node.id,
        challenge,
        index
      );
    });
  };

  walk(root, 0, '1', null, null, null);
  return routes;
};

const questArchetypeNodeTypeLabel = (nodeType?: QuestArchetypeNodeType) => {
  switch (nodeType) {
    case 'challenge':
      return 'Challenge';
    case 'monster_encounter':
      return 'Monster';
    case 'scenario':
      return 'Scenario';
    case 'exposition':
      return 'Exposition';
    case 'fetch_quest':
      return 'Fetch Quest';
    case 'story_flag':
      return 'Story Flag';
    default:
      return 'Node';
  }
};

type RewardMode = 'explicit' | 'random';
type RandomRewardSize = 'small' | 'medium' | 'large';

interface QuestNodeInspectorProps {
  node: QuestArchetypeNode;
  pathLabel: string;
  depth: number;
  locationArchetypes: LocationArchetype[];
  monsterTemplates: MonsterTemplateRecord[];
  scenarioTemplates: ScenarioTemplateRecord[];
  challengeTemplates: ChallengeTemplateRecord[];
  characters: Character[];
  inventoryItems: InventoryItem[];
  spells: Spell[];
  addChallengeToQuestArchetype: (
    questArchetypeId: string,
    proficiency?: string | null,
    unlockedNode?: QuestArchetypeNodeDraft | null,
    challengeTemplateId?: string | null
  ) => void;
  onSaveNode: (nodeId: string, updates: QuestArchetypeNodeDraft) => void;
  onEditChallenge: (
    challenge: QuestArchetypeChallenge,
    allowsTemplate?: boolean
  ) => void;
  onEditChallengeTemplate: (
    templateId: string,
    sourceLabel: string,
    lockedLocationArchetypeId?: string | null
  ) => void;
  onEditScenarioTemplate: (templateId: string, sourceLabel: string) => void;
  onEditMonsterTemplate: (templateId: string, sourceLabel: string) => void;
  onSelectNode: (nodeId: string) => void;
}

const QuestNodeInspector: React.FC<QuestNodeInspectorProps> = ({
  node,
  pathLabel,
  depth,
  locationArchetypes,
  monsterTemplates,
  scenarioTemplates,
  challengeTemplates,
  characters,
  inventoryItems,
  spells,
  addChallengeToQuestArchetype,
  onSaveNode,
  onEditChallenge,
  onEditChallengeTemplate,
  onEditScenarioTemplate,
  onEditMonsterTemplate,
  onSelectNode,
}) => {
  const nodeSummary = describeQuestArchetypeNode(
    node,
    locationArchetypes,
    monsterTemplates,
    scenarioTemplates
  );
  const [isAdding, setIsAdding] = useState(false);
  const [nodeEditor, setNodeEditor] = useState<QuestArchetypeNodeEditorState>(
    buildNodeEditorState(node, locationArchetypes)
  );
  const [childEditor, setChildEditor] = useState<QuestArchetypeNodeEditorState>(
    emptyNodeEditorState()
  );
  const [childEnabled, setChildEnabled] = useState<boolean>(false);

  useEffect(() => {
    setNodeEditor(buildNodeEditorState(node, locationArchetypes));
    setChildEnabled(false);
    setChildEditor(emptyNodeEditorState());
    setIsAdding(false);
  }, [locationArchetypes, node]);
  const selectedChallengeTemplate = challengeTemplates.find(
    (template) => template.id === nodeEditor.challengeTemplateId
  );
  const selectedScenarioTemplate = scenarioTemplates.find(
    (template) => template.id === nodeEditor.scenarioTemplateId
  );
  const selectedMonsterTemplates = monsterTemplates.filter((template) =>
    nodeEditor.monsterTemplateIds.includes(template.id)
  );

  return (
    <div className="qa-inspector-stack">
      <div className="qa-card qa-node-focus">
        <div className="qa-node-focus-header">
          <div>
            <div className="qa-kicker">Step {pathLabel}</div>
            <h3 className="qa-node-focus-title">{nodeSummary}</h3>
            <p className="qa-muted">
              {questArchetypeNodeTypeLabel(node.nodeType)} node at depth {depth}
              .
            </p>
          </div>
          <div className="qa-inline">
            <span className="qa-chip accent">
              {node.challenges?.length ?? 0} outgoing
            </span>
            <span className="qa-chip muted">Inherits quest difficulty</span>
          </div>
        </div>
      </div>

      <div className="qa-card">
        <div className="qa-card-header">
          <div>
            <div className="qa-card-title">Node Setup</div>
            <div className="qa-meta">
              Define what this beat generates. Difficulty is inherited from the
              quest template.
            </div>
          </div>
          <div className="qa-inline">
            {nodeEditor.nodeType === 'challenge' && selectedChallengeTemplate && (
              <button
                className="qa-btn qa-btn-ghost"
                onClick={() =>
                  onEditChallengeTemplate(
                    selectedChallengeTemplate.id,
                    `Challenge node ${pathLabel}`,
                    selectedChallengeTemplate.locationArchetypeId || null
                  )
                }
              >
                Edit Challenge Template
              </button>
            )}
            {nodeEditor.nodeType === 'scenario' && selectedScenarioTemplate && (
              <button
                className="qa-btn qa-btn-ghost"
                onClick={() =>
                  onEditScenarioTemplate(
                    selectedScenarioTemplate.id,
                    `Scenario node ${pathLabel}`
                  )
                }
              >
                Edit Scenario Template
              </button>
            )}
            <button
              className="qa-btn qa-btn-outline"
              onClick={() =>
                setNodeEditor(buildNodeEditorState(node, locationArchetypes))
              }
            >
              Reset
            </button>
            <button
              className="qa-btn qa-btn-primary"
              onClick={() => {
                const validationError = validateQuestArchetypeNodeEditor(
                  nodeEditor,
                  'Node'
                );
                if (validationError) {
                  window.alert(validationError);
                  return;
                }
                onSaveNode(node.id, buildNodeDraft(nodeEditor));
              }}
            >
              Save Node
            </button>
          </div>
        </div>
        <div className="qa-form-grid" style={{ marginTop: 18 }}>
          <QuestArchetypeNodeConfigFields
            editor={nodeEditor}
            setEditor={setNodeEditor}
            allowSameAsPreviousLocationMode={depth > 0}
            prefix="Node"
            locationArchetypes={locationArchetypes}
            challengeTemplates={challengeTemplates}
            monsterTemplates={monsterTemplates}
            scenarioTemplates={scenarioTemplates}
            characters={characters}
            inventoryItems={inventoryItems}
            spells={spells}
          />
        </div>
        {nodeEditor.nodeType === 'monster_encounter' &&
          selectedMonsterTemplates.length > 0 && (
            <div style={{ marginTop: 18 }}>
              <div className="qa-label">Linked Monster Templates</div>
              <div
                className="qa-inline"
                style={{ marginTop: 10, flexWrap: 'wrap' }}
              >
                {selectedMonsterTemplates.map((template) => (
                  <button
                    key={`inline-monster-template-${template.id}`}
                    className="qa-btn qa-btn-ghost"
                    onClick={() =>
                      onEditMonsterTemplate(
                        template.id,
                        `Monster node ${pathLabel}`
                      )
                    }
                  >
                    Edit {template.name}
                  </button>
                ))}
              </div>
            </div>
          )}
      </div>

      <div className="qa-card">
        <div className="qa-card-header">
          <div>
            <div className="qa-card-title">Outgoing Branches</div>
            <div className="qa-meta">
              Each row is a possible next beat unlocked from this node.
            </div>
          </div>
          <button
            className="qa-btn qa-btn-primary"
            onClick={() => setIsAdding((prev) => !prev)}
          >
            {isAdding ? 'Close Composer' : 'Add Branch'}
          </button>
        </div>

        {isAdding && (
          <div className="qa-flow-form" style={{ marginTop: 18 }}>
            <div className="qa-field">
              <label className="qa-inline" style={{ alignItems: 'center' }}>
                <input
                  type="checkbox"
                  checked={childEnabled}
                  onChange={(e) => setChildEnabled(e.target.checked)}
                />
                <span className="qa-label" style={{ marginBottom: 0 }}>
                  Unlock a child node
                </span>
              </label>
            </div>
            {childEnabled && (
              <QuestArchetypeNodeConfigFields
                editor={childEditor}
                setEditor={setChildEditor}
                allowSameAsPreviousLocationMode
                prefix="Child"
                locationArchetypes={locationArchetypes}
                challengeTemplates={challengeTemplates}
                monsterTemplates={monsterTemplates}
                scenarioTemplates={scenarioTemplates}
                characters={characters}
                inventoryItems={inventoryItems}
                spells={spells}
              />
            )}
            <div className="qa-flow-form-actions">
              <button
                className="qa-btn qa-btn-outline"
                onClick={() => setIsAdding(false)}
              >
                Cancel
              </button>
              <button
                className="qa-btn qa-btn-primary"
                onClick={async () => {
                  const validationError =
                    childEnabled
                      ? validateQuestArchetypeNodeEditor(childEditor, 'Child')
                      : '';
                  if (validationError) {
                    window.alert(validationError);
                    return;
                  }
                  await addChallengeToQuestArchetype(
                    node.id,
                    null,
                    childEnabled ? buildNodeDraft(childEditor) : null,
                    null
                  );
                  setChildEnabled(false);
                  setChildEditor(emptyNodeEditorState());
                  setIsAdding(false);
                }}
              >
                Add Branch
              </button>
            </div>
          </div>
        )}

        {node.challenges && node.challenges.length > 0 ? (
          <div className="qa-route-branches">
            {node.challenges.map((challenge, index) => {
              const challengeTemplate = resolveChallengeTemplateForChallenge(
                challenge,
                challengeTemplates
              );
              const challengeProficiencyValue = resolveChallengeProficiency(
                challenge,
                challengeTemplates
              );
              const unlockedNodeLabel = challenge.unlockedNode
                ? describeQuestArchetypeNode(
                    challenge.unlockedNode,
                    locationArchetypes,
                    monsterTemplates,
                    scenarioTemplates
                  )
                : null;
              return (
                <div key={challenge.id} className="qa-route-branch-card">
                  <div className="qa-flow-challenge-header">
                    <div>
                      <div className="qa-flow-challenge-title">
                        {`Branch ${index + 1}`}
                      </div>
                      {challengeTemplate && (
                        <div className="qa-meta" style={{ marginTop: 6 }}>
                          {describeChallengeTemplate(
                            challengeTemplate,
                            locationArchetypes
                          )}
                        </div>
                      )}
                      <div className="qa-inline" style={{ marginTop: 8 }}>
                        {challengeProficiencyValue && (
                          <span className="qa-chip muted">
                            {challengeProficiencyValue}
                          </span>
                        )}
                        {challenge.unlockedNode && (
                          <span className="qa-chip">
                            Unlocks{' '}
                            {questArchetypeNodeTypeLabel(
                              challenge.unlockedNode.nodeType
                            )}
                          </span>
                        )}
                      </div>
                    </div>
                    <div className="qa-inline" style={{ flexWrap: 'wrap' }}>
                      <button
                        className="qa-btn qa-btn-ghost"
                        onClick={() =>
                          onEditChallenge(
                            challenge,
                            Boolean(
                              challenge.challengeTemplateId ||
                                challenge.challengeTemplate ||
                                challenge.proficiency
                            )
                          )
                        }
                      >
                        Edit
                      </button>
                      {challengeTemplate && (
                        <button
                          className="qa-btn qa-btn-outline"
                          onClick={() =>
                            onEditChallengeTemplate(
                              challengeTemplate.id,
                              `Branch ${pathLabel}.${index + 1}`,
                              challengeTemplate.locationArchetypeId
                            )
                          }
                        >
                          Edit Template
                        </button>
                      )}
                    </div>
                  </div>
                  {challenge.unlockedNode ? (
                    <div className="qa-route-branch-footer">
                      <div className="qa-meta">Next: {unlockedNodeLabel}</div>
                      <button
                        className="qa-btn qa-btn-outline"
                        onClick={() => onSelectNode(challenge.unlockedNode!.id)}
                      >
                        Open Node
                      </button>
                    </div>
                  ) : (
                    <div className="qa-meta" style={{ marginTop: 12 }}>
                      Ends this path.
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        ) : (
          <div className="qa-empty" style={{ marginTop: 18 }}>
            No outgoing beats yet. Use the composer above to add the next step.
          </div>
        )}
      </div>
    </div>
  );
};

type QuestArchetypeRewardRow = {
  inventoryItemId: string;
  quantity: number;
};

type QuestArchetypeSpellRewardRow = {
  spellId: string;
};

type QuestArchetypeFormState = {
  name: string;
  description: string;
  category: 'side' | 'main_story';
  questGiverCharacterId: string;
  acceptanceDialogue: DialogueMessage[];
  imageUrl: string;
  rootNode: QuestArchetypeNodeEditorState;
  locationArchetypeId: string;
  locationArchetypeQuery: string;
  difficultyMode: QuestDifficultyMode;
  difficulty: number;
  monsterEncounterTargetLevel: number;
  defaultGold: number;
  rewardMode: RewardMode;
  randomRewardSize: RandomRewardSize;
  rewardExperience: number;
  recurrenceFrequency: string;
  materialRewards: ReturnType<typeof emptyMaterialReward>[];
  requiredStoryFlagsText: string;
  setStoryFlagsText: string;
  clearStoryFlagsText: string;
  relationshipTrust: number;
  relationshipRespect: number;
  relationshipFear: number;
  relationshipDebt: number;
  itemRewards: QuestArchetypeRewardRow[];
  spellRewards: QuestArchetypeSpellRewardRow[];
  characterTagsText: string;
  internalTagsText: string;
};

const createEmptyQuestArchetypeForm = (): QuestArchetypeFormState => ({
  name: '',
  description: '',
  category: 'side',
  questGiverCharacterId: '',
  acceptanceDialogue: [],
  imageUrl: '',
  rootNode: emptyNodeEditorState(),
  locationArchetypeId: '',
  locationArchetypeQuery: '',
  difficultyMode: 'scale',
  difficulty: 1,
  monsterEncounterTargetLevel: 1,
  defaultGold: 0,
  rewardMode: 'random',
  randomRewardSize: 'small',
  rewardExperience: 0,
  recurrenceFrequency: '',
  materialRewards: [],
  requiredStoryFlagsText: '',
  setStoryFlagsText: '',
  clearStoryFlagsText: '',
  relationshipTrust: 0,
  relationshipRespect: 0,
  relationshipFear: 0,
  relationshipDebt: 0,
  itemRewards: [],
  spellRewards: [],
  characterTagsText: '',
  internalTagsText: '',
});

const buildQuestArchetypeFormFromRecord = (
  archetype: QuestArchetype,
  locationArchetypes: LocationArchetype[]
): QuestArchetypeFormState => ({
  name: archetype.name ?? '',
  description: archetype.description ?? '',
  category: archetype.category === 'main_story' ? 'main_story' : 'side',
  questGiverCharacterId: archetype.questGiverCharacterId ?? '',
  acceptanceDialogue: archetype.acceptanceDialogue ?? [],
  imageUrl: archetype.imageUrl ?? '',
  rootNode: buildNodeEditorState(archetype.root, locationArchetypes),
  locationArchetypeId: archetype.root?.locationArchetypeId ?? '',
  locationArchetypeQuery:
    locationArchetypes.find(
      (entry) => entry.id === archetype.root?.locationArchetypeId
    )?.name ?? '',
  difficultyMode: archetype.difficultyMode === 'fixed' ? 'fixed' : 'scale',
  difficulty: archetype.difficulty ?? 1,
  monsterEncounterTargetLevel: archetype.monsterEncounterTargetLevel ?? 1,
  defaultGold: archetype.defaultGold ?? 0,
  rewardMode: archetype.rewardMode === 'explicit' ? 'explicit' : 'random',
  randomRewardSize:
    archetype.randomRewardSize === 'medium' ||
    archetype.randomRewardSize === 'large'
      ? archetype.randomRewardSize
      : 'small',
  rewardExperience: archetype.rewardExperience ?? 0,
  recurrenceFrequency: archetype.recurrenceFrequency ?? '',
  materialRewards: (archetype.materialRewards ?? []).map((reward) => ({
    resourceKey: reward.resourceKey,
    amount: reward.amount,
  })),
  requiredStoryFlagsText: (archetype.requiredStoryFlags ?? []).join(', '),
  setStoryFlagsText: (archetype.setStoryFlags ?? []).join(', '),
  clearStoryFlagsText: (archetype.clearStoryFlags ?? []).join(', '),
  relationshipTrust: archetype.questGiverRelationshipEffects?.trust ?? 0,
  relationshipRespect: archetype.questGiverRelationshipEffects?.respect ?? 0,
  relationshipFear: archetype.questGiverRelationshipEffects?.fear ?? 0,
  relationshipDebt: archetype.questGiverRelationshipEffects?.debt ?? 0,
  itemRewards: (archetype.itemRewards ?? []).map((reward) => ({
    inventoryItemId: reward.inventoryItemId
      ? String(reward.inventoryItemId)
      : '',
    quantity: reward.quantity ?? 1,
  })),
  spellRewards: (archetype.spellRewards ?? []).map((reward) => ({
    spellId: reward.spellId ?? '',
  })),
  characterTagsText: (archetype.characterTags ?? []).join(', '),
  internalTagsText: (archetype.internalTags ?? []).join(', '),
});

const normalizeDialogueEffect = (effect?: DialogueMessage['effect']) => {
  switch (effect) {
    case 'angry':
    case 'surprised':
    case 'whisper':
    case 'shout':
    case 'mysterious':
    case 'determined':
      return effect;
    default:
      return undefined;
  }
};

const normalizeQuestArchetypeDraft = (
  form: QuestArchetypeFormState
): QuestArchetypeDraft => {
  const rewardMode = form.rewardMode;
  const acceptanceDialogue = form.acceptanceDialogue
    .map((line, index) => ({
      speaker: line.speaker === 'user' ? 'user' : 'character',
      text: (line.text ?? '').trim(),
      order: index,
      effect: normalizeDialogueEffect(line.effect),
    }))
    .filter((line) => line.text.length > 0);
  const characterTags = form.characterTagsText
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0);
  const internalTags = form.internalTagsText
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0);
  const requiredStoryFlags = form.requiredStoryFlagsText
    .split(',')
    .map((flag) => flag.trim())
    .filter((flag) => flag.length > 0);
  const setStoryFlags = form.setStoryFlagsText
    .split(',')
    .map((flag) => flag.trim())
    .filter((flag) => flag.length > 0);
  const clearStoryFlags = form.clearStoryFlagsText
    .split(',')
    .map((flag) => flag.trim())
    .filter((flag) => flag.length > 0);
  const questGiverRelationshipEffects = {
    trust: Math.max(-3, Math.min(3, Number(form.relationshipTrust) || 0)),
    respect: Math.max(-3, Math.min(3, Number(form.relationshipRespect) || 0)),
    fear: Math.max(-3, Math.min(3, Number(form.relationshipFear) || 0)),
    debt: Math.max(-3, Math.min(3, Number(form.relationshipDebt) || 0)),
  };

  return {
    name: form.name.trim(),
    description: form.description.trim(),
    category: form.category,
    questGiverCharacterId:
      form.category === 'main_story' && form.questGiverCharacterId
        ? form.questGiverCharacterId
        : null,
    acceptanceDialogue,
    imageUrl: form.imageUrl.trim(),
    rootNode: buildNodeDraft(form.rootNode),
    difficultyMode: form.difficultyMode,
    difficulty: Math.max(1, Number(form.difficulty) || 1),
    monsterEncounterTargetLevel: Math.max(
      1,
      Number(form.monsterEncounterTargetLevel) || 1
    ),
    defaultGold: Number(form.defaultGold) || 0,
    rewardMode,
    randomRewardSize: form.randomRewardSize,
    rewardExperience:
      rewardMode === 'explicit' ? Number(form.rewardExperience) || 0 : 0,
    recurrenceFrequency: form.recurrenceFrequency.trim() || null,
    materialRewards:
      rewardMode === 'explicit'
        ? normalizeMaterialRewards(form.materialRewards)
        : [],
    requiredStoryFlags,
    setStoryFlags,
    clearStoryFlags,
    questGiverRelationshipEffects,
    itemRewards:
      rewardMode === 'explicit'
        ? form.itemRewards
            .map((reward) => ({
              inventoryItemId: Number(reward.inventoryItemId) || 0,
              quantity: Number(reward.quantity) || 0,
            }))
            .filter(
              (reward) => reward.inventoryItemId > 0 && reward.quantity > 0
            )
        : [],
    spellRewards:
      rewardMode === 'explicit'
        ? form.spellRewards.filter((reward) => reward.spellId.trim().length > 0)
        : [],
    characterTags: form.category === 'main_story' ? [] : characterTags,
    internalTags,
  };
};

type GeneratorStepSource = 'location_archetype' | 'proximity';
type GeneratorStepContent = 'challenge' | 'scenario' | 'monster';

type QuestTemplateGeneratorStepFormState = {
  id: string;
  source: GeneratorStepSource;
  content: GeneratorStepContent;
  locationArchetypeId: string;
  proximityMeters: number;
};

type QuestTemplateGeneratorFormState = {
  name: string;
  themePrompt: string;
  characterTagsText: string;
  internalTagsText: string;
  steps: QuestTemplateGeneratorStepFormState[];
};

const createGeneratorStepId = () =>
  globalThis.crypto?.randomUUID?.() ??
  `quest-template-step-${Math.random().toString(36).slice(2, 10)}`;

const createQuestTemplateGeneratorStep = (
  source: GeneratorStepSource = 'location_archetype',
  content: GeneratorStepContent = 'challenge'
): QuestTemplateGeneratorStepFormState => ({
  id: createGeneratorStepId(),
  source,
  content:
    source === 'proximity' && content === 'challenge' ? 'scenario' : content,
  locationArchetypeId: '',
  proximityMeters: 100,
});

const createEmptyQuestTemplateGeneratorForm =
  (): QuestTemplateGeneratorFormState => ({
    name: '',
    themePrompt: '',
    characterTagsText: '',
    internalTagsText: '',
    steps: [
      createQuestTemplateGeneratorStep('location_archetype', 'challenge'),
    ],
  });

const normalizeQuestTemplateGeneratorDraft = (
  form: QuestTemplateGeneratorFormState
): QuestTemplateGeneratorDraft => ({
  name: form.name.trim(),
  themePrompt: form.themePrompt.trim(),
  characterTags: form.characterTagsText
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0),
  internalTags: form.internalTagsText
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0),
  steps: form.steps.map((step) => ({
    source: step.source,
    content: step.content,
    locationArchetypeId:
      step.source === 'location_archetype'
        ? step.locationArchetypeId || null
        : null,
    proximityMeters:
      step.source === 'proximity'
        ? Math.max(0, Number(step.proximityMeters) || 0)
        : null,
  })),
});

const validateQuestTemplateGeneratorForm = (
  form: QuestTemplateGeneratorFormState
): string | null => {
  if (form.steps.length === 0) {
    return 'Add at least one step.';
  }
  for (let index = 0; index < form.steps.length; index += 1) {
    const step = form.steps[index];
    if (index === 0 && step.source === 'proximity') {
      return 'The first step must use a location archetype anchor.';
    }
    if (step.source === 'location_archetype' && !step.locationArchetypeId) {
      return `Step ${index + 1} needs a location archetype.`;
    }
    if (step.source === 'proximity' && step.content === 'challenge') {
      return `Step ${index + 1} cannot be a proximity challenge.`;
    }
    if (
      step.source === 'proximity' &&
      (Number(step.proximityMeters) || 0) < 0
    ) {
      return `Step ${index + 1} needs a non-negative proximity.`;
    }
  }
  return null;
};

export const QuestArchetypeComponent = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const { apiClient } = useAPI();
  const { zones } = useZoneContext();
  const {
    questArchetypes,
    locationArchetypes,
    refreshQuestArchetypes,
    createQuestArchetype,
    generateQuestArchetypeTemplate,
    updateQuestArchetype,
    deleteQuestArchetype,
    addChallengeToQuestArchetype,
    updateQuestArchetypeChallenge,
    deleteQuestArchetypeChallenge,
    updateQuestArchetypeNode,
  } = useQuestArchetypes();
  const [characters, setCharacters] = useState<Character[]>([]);
  const [inventoryItems, setInventoryItems] = useState<InventoryItem[]>([]);
  const [monsterTemplates, setMonsterTemplates] = useState<
    MonsterTemplateRecord[]
  >([]);
  const [scenarioTemplates, setScenarioTemplates] = useState<
    ScenarioTemplateRecord[]
  >([]);
  const [challengeTemplates, setChallengeTemplates] = useState<
    ChallengeTemplateRecord[]
  >([]);
  const [spells, setSpells] = useState<Spell[]>([]);
  const [inventoryItemsLoading, setInventoryItemsLoading] =
    useState<boolean>(false);
  const [shouldShowModal, setShouldShowModal] = useState(false);
  const [shouldShowGeneratorModal, setShouldShowGeneratorModal] =
    useState(false);
  const [shouldShowGenerateQuestModal, setShouldShowGenerateQuestModal] =
    useState(false);
  const [createForm, setCreateForm] = useState<QuestArchetypeFormState>(
    createEmptyQuestArchetypeForm()
  );
  const [generatorForm, setGeneratorForm] =
    useState<QuestTemplateGeneratorFormState>(
      createEmptyQuestTemplateGeneratorForm()
    );
  const [editingArchetype, setEditingArchetype] =
    useState<QuestArchetype | null>(null);
  const [editForm, setEditForm] = useState<QuestArchetypeFormState>(
    createEmptyQuestArchetypeForm()
  );
  const [editingChallenge, setEditingChallenge] =
    useState<QuestArchetypeChallenge | null>(null);
  const [editChallengeTemplateId, setEditChallengeTemplateId] =
    useState<string>('');
  const [editingChallengeAllowsTemplate, setEditingChallengeAllowsTemplate] =
    useState<boolean>(false);
  const [editChallengeProficiency, setEditChallengeProficiency] =
    useState<string>('');
  const [inlineEditor, setInlineEditor] =
    useState<QuestArchetypeInlineEditorState | null>(null);
  const [inlineEditorError, setInlineEditorError] = useState<string>('');
  const [inlineEditorSuccess, setInlineEditorSuccess] = useState<string>('');
  const [inlineEditorSaving, setInlineEditorSaving] = useState<boolean>(false);
  const [challengeTemplateForm, setChallengeTemplateForm] =
    useState<ChallengeTemplateFormState>(emptyChallengeTemplateForm());
  const [scenarioTemplateForm, setScenarioTemplateForm] =
    useState<ScenarioTemplateFormState>(emptyScenarioTemplateForm());
  const [monsterTemplateForm, setMonsterTemplateForm] =
    useState<MonsterTemplateFormState>(emptyMonsterTemplateForm());
  const [proficiencySearch, setProficiencySearch] = useState<string>('');
  const [proficiencyOptions, setProficiencyOptions] = useState<string[]>([]);
  const [archetypeSearch, setArchetypeSearch] = useState<string>('');
  const [selectedArchetypeId, setSelectedArchetypeId] = useState<string>('');
  const [selectedArchetypeIds, setSelectedArchetypeIds] = useState<string[]>([]);
  const attemptedDeepLinkRefreshIdRef = useRef<string>('');
  const appliedDeepLinkedArchetypeIdRef = useRef<string>('');
  const [selectedNodeId, setSelectedNodeId] = useState<string>('');
  const [generateQuestZoneSearch, setGenerateQuestZoneSearch] =
    useState<string>('');
  const [selectedGenerateZoneId, setSelectedGenerateZoneId] =
    useState<string>('');
  const [isGeneratingQuest, setIsGeneratingQuest] = useState(false);
  const [generateQuestError, setGenerateQuestError] = useState<string>('');
  const [questGenerationJob, setQuestGenerationJob] =
    useState<QuestGenerationJob | null>(null);

  const generatorValidationError =
    validateQuestTemplateGeneratorForm(generatorForm);

  const toggleArchetypeSelection = useCallback((archetypeId: string) => {
    setSelectedArchetypeIds((current) =>
      current.includes(archetypeId)
        ? current.filter((id) => id !== archetypeId)
        : [...current, archetypeId]
    );
  }, []);

  const clearSelectedArchetypes = useCallback(() => {
    setSelectedArchetypeIds([]);
  }, []);

  const deleteSelectedArchetypes = useCallback(async () => {
    if (selectedArchetypeIds.length === 0) {
      return;
    }
    const archetypeLabel =
      selectedArchetypeIds.length === 1 ? 'quest archetype' : 'quest archetypes';
    if (
      !window.confirm(
        `Are you sure you want to delete ${selectedArchetypeIds.length} ${archetypeLabel}?`
      )
    ) {
      return;
    }
    const idsToDelete = [...selectedArchetypeIds];
    for (const archetypeId of idsToDelete) {
      await deleteQuestArchetype(archetypeId);
    }
    setSelectedArchetypeIds([]);
  }, [deleteQuestArchetype, selectedArchetypeIds]);

  const monsterProgressionOptions = useMemo(() => {
    const groups = new Map<string, MonsterAbilityProgressionOption>();
    for (const spell of spells) {
      const link = spell.progressionLinks?.[0];
      if (!link?.progressionId) {
        continue;
      }
      const abilityType =
        (spell.abilityType ?? 'spell') === 'technique' ? 'technique' : 'spell';
      const existing = groups.get(link.progressionId);
      if (existing) {
        existing.memberCount += 1;
        if (
          link.progression?.name &&
          link.progression.name.trim() &&
          existing.name.startsWith('Progression ')
        ) {
          existing.name = link.progression.name.trim();
        }
        continue;
      }
      groups.set(link.progressionId, {
        id: link.progressionId,
        name:
          link.progression?.name?.trim() || `Progression ${groups.size + 1}`,
        abilityType,
        memberCount: 1,
      });
    }
    return Array.from(groups.values()).sort((left, right) =>
      left.name.localeCompare(right.name)
    );
  }, [spells]);
  const monsterSpellProgressionOptions = useMemo(
    () =>
      monsterProgressionOptions.filter(
        (option) => option.abilityType !== 'technique'
      ),
    [monsterProgressionOptions]
  );
  const monsterTechniqueProgressionOptions = useMemo(
    () =>
      monsterProgressionOptions.filter(
        (option) => option.abilityType === 'technique'
      ),
    [monsterProgressionOptions]
  );

  const filteredArchetypes = useMemo(
    () =>
      questArchetypes.filter((archetype) =>
        archetype.name
          .toLowerCase()
          .includes(archetypeSearch.trim().toLowerCase())
      ),
    [questArchetypes, archetypeSearch]
  );

  const selectAllFilteredArchetypes = useCallback(() => {
    setSelectedArchetypeIds(
      Array.from(new Set(filteredArchetypes.map((archetype) => archetype.id)))
    );
  }, [filteredArchetypes]);

  const selectedArchetype = useMemo(
    () =>
      questArchetypes.find(
        (archetype) => archetype.id === selectedArchetypeId
      ) ?? null,
    [questArchetypes, selectedArchetypeId]
  );
  const charactersById = useMemo(
    () =>
      new Map(
        characters.map((character) => [character.id, character] as const)
      ),
    [characters]
  );
  const selectedArchetypeQuestGiver = useMemo(
    () =>
      selectedArchetype?.questGiverCharacterId
        ? charactersById.get(selectedArchetype.questGiverCharacterId) ?? null
        : null,
    [charactersById, selectedArchetype?.questGiverCharacterId]
  );
  const sortedCharacters = useMemo(
    () =>
      [...characters].sort((left, right) =>
        left.name.localeCompare(right.name)
      ),
    [characters]
  );
  const editingQuestGiver = useMemo(
    () =>
      editForm.questGiverCharacterId
        ? charactersById.get(editForm.questGiverCharacterId) ?? null
        : null,
    [charactersById, editForm.questGiverCharacterId]
  );

  const filteredGenerationZones = useMemo(() => {
    const query = generateQuestZoneSearch.trim().toLowerCase();
    const sorted = [...zones].sort((left, right) =>
      left.name.localeCompare(right.name)
    );
    if (!query) {
      return sorted.slice(0, 18);
    }
    return sorted
      .filter((zone) => zone.name.toLowerCase().includes(query))
      .slice(0, 18);
  }, [generateQuestZoneSearch, zones]);

  const selectedGenerateZone = useMemo(
    () => zones.find((zone) => zone.id === selectedGenerateZoneId) ?? null,
    [zones, selectedGenerateZoneId]
  );

  const generatedQuest = useMemo(
    () => questGenerationJob?.quests?.[0] ?? null,
    [questGenerationJob]
  );

  const flowMapLayout = useMemo(
    () =>
      buildFlowMapLayout(
        selectedArchetype?.root ?? null,
        locationArchetypes,
        monsterTemplates,
        scenarioTemplates
      ),
    [selectedArchetype, locationArchetypes, monsterTemplates, scenarioTemplates]
  );

  const flowRoutes = useMemo(
    () => buildQuestFlowRoute(selectedArchetype?.root ?? null),
    [selectedArchetype]
  );

  const selectedFlowNode = useMemo(
    () =>
      flowRoutes.find((route) => route.id === selectedNodeId) ??
      flowRoutes[0] ??
      null,
    [flowRoutes, selectedNodeId]
  );

  const activeChallengeTemplate = useMemo(
    () =>
      inlineEditor?.kind === 'challenge'
        ? challengeTemplates.find(
            (template) => template.id === inlineEditor.templateId
          ) ?? null
        : null,
    [challengeTemplates, inlineEditor]
  );

  const activeScenarioTemplate = useMemo(
    () =>
      inlineEditor?.kind === 'scenario'
        ? scenarioTemplates.find(
            (template) => template.id === inlineEditor.templateId
          ) ?? null
        : null,
    [inlineEditor, scenarioTemplates]
  );

  const activeMonsterTemplate = useMemo(
    () =>
      inlineEditor?.kind === 'monster'
        ? monsterTemplates.find(
            (template) => template.id === inlineEditor.templateId
          ) ?? null
        : null,
    [inlineEditor, monsterTemplates]
  );
  const deepLinkedArchetypeId = searchParams.get('id')?.trim() ?? '';
  const replaceDeepLinkedArchetypeId = useCallback((
    archetypeId?: string | null
  ) => {
    const normalizedArchetypeId = (archetypeId ?? '').trim();
    const currentArchetypeId = searchParams.get('id')?.trim() ?? '';
    if (normalizedArchetypeId === currentArchetypeId) {
      return;
    }
    const next = new URLSearchParams(searchParams);
    if (normalizedArchetypeId) {
      next.set('id', normalizedArchetypeId);
    } else {
      next.delete('id');
    }
    setSearchParams(next, { replace: true });
  }, [searchParams, setSearchParams]);

  useEffect(() => {
    if (!deepLinkedArchetypeId) {
      attemptedDeepLinkRefreshIdRef.current = '';
      appliedDeepLinkedArchetypeIdRef.current = '';
      return;
    }
    const hasDeepLinkedArchetype = questArchetypes.some(
      (archetype) => archetype.id === deepLinkedArchetypeId
    );
    if (hasDeepLinkedArchetype) {
      return;
    }
    if (attemptedDeepLinkRefreshIdRef.current === deepLinkedArchetypeId) {
      return;
    }
    attemptedDeepLinkRefreshIdRef.current = deepLinkedArchetypeId;
    void refreshQuestArchetypes();
  }, [deepLinkedArchetypeId, questArchetypes, refreshQuestArchetypes]);

  useEffect(() => {
    if (questArchetypes.length === 0) {
      setSelectedArchetypeId('');
      return;
    }
    if (deepLinkedArchetypeId) {
      const matchingArchetype = questArchetypes.find(
        (archetype) => archetype.id === deepLinkedArchetypeId
      );
      if (matchingArchetype) {
        if (
          appliedDeepLinkedArchetypeIdRef.current !== deepLinkedArchetypeId
        ) {
          appliedDeepLinkedArchetypeIdRef.current = deepLinkedArchetypeId;
          setSelectedArchetypeId(matchingArchetype.id);
          return;
        }
      }
      if (attemptedDeepLinkRefreshIdRef.current !== deepLinkedArchetypeId) {
        return;
      }
    }
    const stillExists = questArchetypes.some(
      (archetype) => archetype.id === selectedArchetypeId
    );
    if (!stillExists) {
      setSelectedArchetypeId(questArchetypes[0].id);
    }
  }, [deepLinkedArchetypeId, questArchetypes, selectedArchetypeId]);

  useEffect(() => {
    if (questArchetypes.length === 0) {
      return;
    }
    if (!selectedArchetypeId && deepLinkedArchetypeId) {
      return;
    }
    if (
      deepLinkedArchetypeId &&
      !questArchetypes.some((archetype) => archetype.id === deepLinkedArchetypeId)
    ) {
      return;
    }
    replaceDeepLinkedArchetypeId(selectedArchetypeId || null);
  }, [
    deepLinkedArchetypeId,
    questArchetypes,
    replaceDeepLinkedArchetypeId,
    selectedArchetypeId,
  ]);

  useEffect(() => {
    setSelectedArchetypeIds((current) =>
      current.filter((id) =>
        questArchetypes.some((archetype) => archetype.id === id)
      )
    );
  }, [questArchetypes]);

  useEffect(() => {
    if (flowRoutes.length === 0) {
      setSelectedNodeId('');
      return;
    }
    const stillExists = flowRoutes.some((route) => route.id === selectedNodeId);
    if (!stillExists) {
      setSelectedNodeId(flowRoutes[0].id);
    }
  }, [flowRoutes, selectedNodeId]);

  useEffect(() => {
    setInlineEditor(null);
    setInlineEditorError('');
    setInlineEditorSuccess('');
  }, [selectedArchetypeId, selectedNodeId]);

  useEffect(() => {
    if (!activeChallengeTemplate) {
      setChallengeTemplateForm(emptyChallengeTemplateForm());
      return;
    }
    const nextForm = challengeTemplateFormFromRecord(activeChallengeTemplate);
    if (
      inlineEditor?.kind === 'challenge' &&
      inlineEditor.lockedLocationArchetypeId
    ) {
      nextForm.locationArchetypeId = inlineEditor.lockedLocationArchetypeId;
    }
    setChallengeTemplateForm(nextForm);
  }, [activeChallengeTemplate, inlineEditor]);

  useEffect(() => {
    if (!activeScenarioTemplate) {
      setScenarioTemplateForm(emptyScenarioTemplateForm());
      return;
    }
    setScenarioTemplateForm(
      scenarioTemplateFormFromRecord(activeScenarioTemplate)
    );
  }, [activeScenarioTemplate]);

  useEffect(() => {
    if (!activeMonsterTemplate) {
      setMonsterTemplateForm(emptyMonsterTemplateForm());
      return;
    }
    setMonsterTemplateForm(
      monsterTemplateFormFromRecord(activeMonsterTemplate)
    );
  }, [activeMonsterTemplate]);

  useEffect(() => {
    if (!shouldShowGenerateQuestModal) {
      return;
    }
    if (
      selectedGenerateZoneId &&
      zones.some((zone) => zone.id === selectedGenerateZoneId)
    ) {
      return;
    }
    setSelectedGenerateZoneId(filteredGenerationZones[0]?.id ?? '');
  }, [
    filteredGenerationZones,
    selectedGenerateZoneId,
    shouldShowGenerateQuestModal,
    zones,
  ]);

  useEffect(() => {
    if (
      !shouldShowGenerateQuestModal ||
      !questGenerationJob?.id ||
      !isPendingQuestGenerationStatus(questGenerationJob.status)
    ) {
      return;
    }

    let active = true;

    const pollJob = async () => {
      try {
        const refreshed = await apiClient.get<QuestGenerationJob>(
          `/sonar/questGenerationJobs/${questGenerationJob.id}`
        );
        if (!active) {
          return;
        }
        setQuestGenerationJob(refreshed);
        if (refreshed.status === 'failed' && refreshed.errorMessage) {
          setGenerateQuestError(refreshed.errorMessage);
        } else if (refreshed.status === 'completed') {
          setGenerateQuestError('');
        }
      } catch (error) {
        if (!active) {
          return;
        }
        console.error('Failed to refresh quest generation job', error);
        setGenerateQuestError(
          extractApiErrorMessage(
            error,
            'Failed to refresh quest generation status.'
          )
        );
      }
    };

    pollJob();
    const interval = window.setInterval(pollJob, 3000);
    return () => {
      active = false;
      window.clearInterval(interval);
    };
  }, [
    apiClient,
    questGenerationJob?.id,
    questGenerationJob?.status,
    shouldShowGenerateQuestModal,
  ]);

  useEffect(() => {
    const fetchReferenceData = async () => {
      setInventoryItemsLoading(true);
      try {
        const [
          characterResponse,
          inventoryResponse,
          spellsResponse,
          monsterTemplateResponse,
          scenarioTemplateResponse,
          challengeTemplateResponse,
        ] = await Promise.all([
          apiClient.get<Character[]>('/sonar/characters'),
          apiClient.get<InventoryItem[]>('/sonar/inventory-items'),
          apiClient.get<Spell[]>('/sonar/spells'),
          apiClient.get<PaginatedResponse<MonsterTemplateRecord>>(
            '/sonar/admin/monster-templates?page=1&pageSize=500'
          ),
          apiClient.get<PaginatedResponse<ScenarioTemplateRecord>>(
            '/sonar/admin/scenario-templates?page=1&pageSize=500'
          ),
          apiClient.get<ChallengeTemplateRecord[]>(
            '/sonar/challenge-templates'
          ),
        ]);
        setCharacters(characterResponse ?? []);
        setInventoryItems(inventoryResponse);
        setSpells(spellsResponse);
        setMonsterTemplates(monsterTemplateResponse.items ?? []);
        setScenarioTemplates(scenarioTemplateResponse.items ?? []);
        setChallengeTemplates(challengeTemplateResponse ?? []);
      } catch (error) {
        console.error('Error fetching quest archetype reference data:', error);
      } finally {
        setInventoryItemsLoading(false);
      }
    };

    fetchReferenceData();
  }, [apiClient]);

  useEffect(() => {
    const query = proficiencySearch.trim();
    let active = true;
    const handle = window.setTimeout(async () => {
      try {
        const results = await apiClient.get<string[]>(
          `/sonar/proficiencies?query=${encodeURIComponent(query)}&limit=25`
        );
        if (!active) return;
        setProficiencyOptions(Array.isArray(results) ? results : []);
      } catch (error) {
        if (active) {
          console.error('Failed to load proficiencies', error);
          setProficiencyOptions([]);
        }
      }
    }, 200);
    return () => {
      active = false;
      window.clearTimeout(handle);
    };
  }, [apiClient, proficiencySearch]);

  const openGenerateQuestModal = () => {
    setGenerateQuestZoneSearch('');
    setSelectedGenerateZoneId('');
    setGenerateQuestError('');
    setQuestGenerationJob(null);
    setShouldShowGenerateQuestModal(true);
  };

  const handleGenerateQuest = async () => {
    if (!selectedArchetype || !selectedGenerateZoneId) {
      return;
    }
    setIsGeneratingQuest(true);
    setGenerateQuestError('');
    setQuestGenerationJob(null);
    try {
      const job = await apiClient.post<QuestGenerationJob>(
        `/sonar/questArchetypes/${selectedArchetype.id}/generate`,
        {
          zoneId: selectedGenerateZoneId,
        }
      );
      setQuestGenerationJob(job);
    } catch (error) {
      console.error('Failed to generate quest from archetype', error);
      setGenerateQuestError(
        extractApiErrorMessage(
          error,
          'Failed to generate a quest from this archetype.'
        )
      );
    } finally {
      setIsGeneratingQuest(false);
    }
  };

  const openChallengeEditor = (
    selected: QuestArchetypeChallenge,
    allowsTemplate = false
  ) => {
    setEditingChallenge(selected);
    setEditingChallengeAllowsTemplate(allowsTemplate);
    setEditChallengeTemplateId(
      allowsTemplate ? selected.challengeTemplateId ?? '' : ''
    );
    setEditChallengeProficiency(
      resolveChallengeProficiency(selected, challengeTemplates)
    );
    setProficiencySearch(
      resolveChallengeProficiency(selected, challengeTemplates)
    );
  };

  const openInlineChallengeTemplateEditor = (
    templateId: string,
    sourceLabel: string,
    lockedLocationArchetypeId?: string | null
  ) => {
    setInlineEditor({
      kind: 'challenge',
      templateId,
      sourceLabel,
      lockedLocationArchetypeId: lockedLocationArchetypeId ?? null,
    });
    setInlineEditorError('');
    setInlineEditorSuccess('');
  };

  const openInlineScenarioTemplateEditor = (
    templateId: string,
    sourceLabel: string
  ) => {
    setInlineEditor({
      kind: 'scenario',
      templateId,
      sourceLabel,
    });
    setInlineEditorError('');
    setInlineEditorSuccess('');
  };

  const openInlineMonsterTemplateEditor = (
    templateId: string,
    sourceLabel: string
  ) => {
    setInlineEditor({
      kind: 'monster',
      templateId,
      sourceLabel,
    });
    setInlineEditorError('');
    setInlineEditorSuccess('');
  };

  const handleSaveInlineChallengeTemplate = async () => {
    if (!activeChallengeTemplate || inlineEditor?.kind !== 'challenge') {
      return;
    }
    setInlineEditorSaving(true);
    setInlineEditorError('');
    setInlineEditorSuccess('');
    try {
      const payload = buildChallengeTemplatePayloadFromForm(
        challengeTemplateForm
      );
      const updated = await apiClient.put<ChallengeTemplateRecord>(
        `/sonar/challenge-templates/${activeChallengeTemplate.id}`,
        payload
      );
      setChallengeTemplates((prev) =>
        prev.map((template) =>
          template.id === updated.id ? { ...template, ...updated } : template
        )
      );
      setInlineEditorSuccess('Challenge template saved.');
    } catch (error) {
      console.error('Failed to save challenge template inline', error);
      setInlineEditorError(
        error instanceof Error
          ? error.message
          : extractApiErrorMessage(error, 'Failed to save challenge template.')
      );
    } finally {
      setInlineEditorSaving(false);
    }
  };

  const handleSaveInlineScenarioTemplate = async () => {
    if (!activeScenarioTemplate || inlineEditor?.kind !== 'scenario') {
      return;
    }
    setInlineEditorSaving(true);
    setInlineEditorError('');
    setInlineEditorSuccess('');
    try {
      const payload =
        buildScenarioTemplatePayloadFromForm(scenarioTemplateForm);
      const updated = await apiClient.put<ScenarioTemplateRecord>(
        `/sonar/scenario-templates/${activeScenarioTemplate.id}`,
        payload
      );
      setScenarioTemplates((prev) =>
        prev.map((template) =>
          template.id === updated.id ? { ...template, ...updated } : template
        )
      );
      setInlineEditorSuccess('Scenario template saved.');
    } catch (error) {
      console.error('Failed to save scenario template inline', error);
      setInlineEditorError(
        error instanceof Error
          ? error.message
          : extractApiErrorMessage(error, 'Failed to save scenario template.')
      );
    } finally {
      setInlineEditorSaving(false);
    }
  };

  const handleSaveInlineMonsterTemplate = async () => {
    if (!activeMonsterTemplate || inlineEditor?.kind !== 'monster') {
      return;
    }
    setInlineEditorSaving(true);
    setInlineEditorError('');
    setInlineEditorSuccess('');
    try {
      const payload = buildMonsterTemplatePayloadFromForm(monsterTemplateForm);
      if (!payload.name) {
        throw new Error('Monster template name is required.');
      }
      if (
        payload.baseStrength <= 0 ||
        payload.baseDexterity <= 0 ||
        payload.baseConstitution <= 0 ||
        payload.baseIntelligence <= 0 ||
        payload.baseWisdom <= 0 ||
        payload.baseCharisma <= 0
      ) {
        throw new Error('All monster base stats must be positive.');
      }
      if (
        payload.strongAgainstAffinity &&
        payload.strongAgainstAffinity === payload.weakAgainstAffinity
      ) {
        throw new Error(
          'Strong against and weak against affinities must be different.'
        );
      }

      const updated = await apiClient.put<MonsterTemplateRecord>(
        `/sonar/monster-templates/${activeMonsterTemplate.id}`,
        payload
      );
      setMonsterTemplates((prev) =>
        prev.map((template) =>
          template.id === updated.id ? { ...template, ...updated } : template
        )
      );
      setInlineEditorSuccess('Monster template saved.');
    } catch (error) {
      console.error('Failed to save monster template inline', error);
      setInlineEditorError(
        error instanceof Error
          ? error.message
          : extractApiErrorMessage(error, 'Failed to save monster template.')
      );
    } finally {
      setInlineEditorSaving(false);
    }
  };

  useEffect(() => {
    if (!editChallengeTemplateId) {
      return;
    }
    const template = challengeTemplates.find(
      (entry) => entry.id === editChallengeTemplateId
    );
    if (!template) {
      return;
    }
    setEditChallengeProficiency(template.proficiency ?? '');
    setProficiencySearch(template.proficiency ?? '');
  }, [challengeTemplates, editChallengeTemplateId]);

  return (
    <div className="qa-theme">
      <datalist id="qa-proficiency-options">
        {proficiencyOptions.map((option) => (
          <option key={option} value={option} />
        ))}
      </datalist>
      <div className="qa-shell">
        <header className="qa-hero">
          <div>
            <div className="qa-kicker">Quest Design Lab</div>
            <h1 className="qa-title">Quest Archetypes</h1>
            <p className="qa-subtitle">
              Build the backbone of every adventure. Define branching
              challenges, beats, and proficiencies so generated quests feel
              crafted rather than random.
            </p>
          </div>
          <div className="qa-hero-actions">
            {inventoryItemsLoading && (
              <span className="qa-chip muted">Loading items…</span>
            )}
            <button
              className="qa-btn qa-btn-outline"
              onClick={() => {
                setGeneratorForm(createEmptyQuestTemplateGeneratorForm());
                setShouldShowGeneratorModal(true);
              }}
            >
              Generate Template
            </button>
            <button
              className="qa-btn qa-btn-primary"
              onClick={() => {
                setCreateForm(createEmptyQuestArchetypeForm());
                setShouldShowModal(true);
              }}
            >
              New Archetype
            </button>
          </div>
        </header>

        <section className="qa-layout">
          <aside className="qa-sidebar">
            <div className="qa-card qa-sidebar-card">
              <div className="qa-card-title">Archetype Library</div>
              <p className="qa-muted" style={{ marginTop: 6 }}>
                Pick a quest archetype to shape its challenge flow.
              </p>
              <input
                className="qa-input qa-sidebar-search"
                placeholder="Search archetypes..."
                value={archetypeSearch}
                onChange={(e) => setArchetypeSearch(e.target.value)}
              />
              <div
                style={{
                  display: 'flex',
                  gap: 8,
                  flexWrap: 'wrap',
                  alignItems: 'center',
                  marginTop: 10,
                }}
              >
                <button
                  className="qa-btn qa-btn-ghost"
                  type="button"
                  onClick={selectAllFilteredArchetypes}
                  disabled={filteredArchetypes.length === 0}
                >
                  Select All
                </button>
                <button
                  className="qa-btn qa-btn-ghost"
                  type="button"
                  onClick={clearSelectedArchetypes}
                  disabled={selectedArchetypeIds.length === 0}
                >
                  Clear
                </button>
                <button
                  className="qa-btn qa-btn-danger"
                  type="button"
                  onClick={() => void deleteSelectedArchetypes()}
                  disabled={selectedArchetypeIds.length === 0}
                >
                  Delete Selected
                </button>
                <span className="qa-chip muted">
                  {selectedArchetypeIds.length} selected
                </span>
              </div>
              <div className="qa-sidebar-list">
                {filteredArchetypes.length === 0 ? (
                  <div className="qa-empty">
                    No archetypes match that search.
                  </div>
                ) : (
                  filteredArchetypes.map((questArchetype) => {
                    const rootLocation = describeQuestArchetypeNode(
                      questArchetype.root,
                      locationArchetypes,
                      monsterTemplates,
                      scenarioTemplates
                    );
                    return (
                      <button
                        key={questArchetype.id}
                        type="button"
                        className={`qa-sidebar-item ${selectedArchetypeId === questArchetype.id ? 'is-active' : ''}`}
                        onClick={() =>
                          setSelectedArchetypeId(questArchetype.id)
                        }
                      >
                        <div
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 10,
                            width: '100%',
                          }}
                        >
                          <input
                            type="checkbox"
                            checked={selectedArchetypeIds.includes(
                              questArchetype.id
                            )}
                            onChange={() =>
                              toggleArchetypeSelection(questArchetype.id)
                            }
                            onClick={(event) => event.stopPropagation()}
                          />
                          <div style={{ minWidth: 0, flex: 1 }}>
                        <div className="qa-sidebar-item-title">
                          {questArchetype.name}
                        </div>
                        <div className="qa-meta">
                          Root: {rootLocation} ·{' '}
                          {questArchetype.root?.challenges?.length ?? 0}{' '}
                          challenges
                        </div>
                          </div>
                        </div>
                      </button>
                    );
                  })
                )}
              </div>
            </div>
          </aside>

          <div className="qa-builder">
            {!selectedArchetype ? (
              <div className="qa-panel">
                <div className="qa-card-title">Select a quest archetype</div>
                <p className="qa-muted" style={{ marginTop: 8 }}>
                  Choose an archetype on the left to build its quest flow.
                </p>
              </div>
            ) : (
              <>
                <div className="qa-card qa-builder-header">
                  <div>
                    <div className="qa-kicker">Quest Flow Builder</div>
                    <h2
                      className="qa-title"
                      style={{ fontSize: 'clamp(26px, 3vw, 34px)' }}
                    >
                      {selectedArchetype.name}
                    </h2>
                    <p className="qa-subtitle">
                      Craft the journey by stacking challenges, expositions,
                      and branching nodes. Each beat can unlock a new node to
                      extend the quest.
                    </p>
                  </div>
                  <div className="qa-actions">
                    <button
                      className="qa-btn qa-btn-primary"
                      onClick={openGenerateQuestModal}
                      disabled={zones.length === 0}
                    >
                      Generate in Zone
                    </button>
                    <button
                      className="qa-btn qa-btn-ghost"
                      onClick={() => {
                        setEditingArchetype(selectedArchetype);
                        setEditForm(
                          buildQuestArchetypeFormFromRecord(
                            selectedArchetype,
                            locationArchetypes
                          )
                        );
                      }}
                    >
                      Edit Template
                    </button>
                    <button
                      className="qa-btn qa-btn-danger"
                      onClick={() => {
                        if (
                          window.confirm(
                            'Are you sure you want to delete this quest archetype?'
                          )
                        ) {
                          deleteQuestArchetype(selectedArchetype.id);
                        }
                      }}
                    >
                      Delete Archetype
                    </button>
                  </div>
                </div>

                <div className="qa-card qa-builder-summary">
                  <div className="qa-stat-grid">
                    <div className="qa-stat">
                      <div className="qa-stat-label">Reward Mode</div>
                      <div className="qa-stat-value">
                        {(
                          selectedArchetype.rewardMode ?? 'random'
                        ).toUpperCase()}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Root Node</div>
                      <div className="qa-stat-value">
                        {describeQuestArchetypeNode(
                          selectedArchetype.root,
                          locationArchetypes,
                          monsterTemplates,
                          scenarioTemplates
                        )}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">
                        {selectedArchetype.category === 'main_story'
                          ? 'Quest Giver'
                          : 'Quest Giver Tags'}
                      </div>
                      <div className="qa-stat-value">
                        {selectedArchetype.category === 'main_story'
                          ? selectedArchetypeQuestGiver?.name ?? 'Unassigned'
                          : (selectedArchetype.characterTags ?? []).length > 0
                            ? (selectedArchetype.characterTags ?? []).join(', ')
                            : 'Any'}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Internal Tags</div>
                      <div className="qa-stat-value">
                        {(selectedArchetype.internalTags ?? []).length > 0
                          ? (selectedArchetype.internalTags ?? []).join(', ')
                          : 'None'}
                      </div>
                    </div>
                    <div className="qa-stat">
                      <div className="qa-stat-label">Created</div>
                      <div className="qa-stat-value">
                        {new Date(
                          selectedArchetype.createdAt
                        ).toLocaleDateString()}
                      </div>
                    </div>
                  </div>

                  <div className="qa-divider" />

                  <div className="qa-panel">
                    <div className="qa-meta">Template Summary</div>
                    {selectedArchetype.description ? (
                      <p className="qa-muted" style={{ marginTop: 10 }}>
                        {selectedArchetype.description}
                      </p>
                    ) : (
                      <div className="qa-empty" style={{ marginTop: 10 }}>
                        No description configured.
                      </div>
                    )}
                    <div className="qa-inline" style={{ marginTop: 10 }}>
                      <span className="qa-chip muted">
                        Difficulty:{' '}
                        {selectedArchetype.difficultyMode === 'fixed'
                          ? selectedArchetype.difficulty ?? 1
                          : 'Scales'}
                      </span>
                      <span className="qa-chip muted">
                        Monster Level:{' '}
                        {selectedArchetype.monsterEncounterTargetLevel ?? 1}
                      </span>
                      <span className="qa-chip muted">
                        Gold: {selectedArchetype.defaultGold ?? 0}
                      </span>
                      <span className="qa-chip muted">
                        XP:{' '}
                        {selectedArchetype.rewardMode === 'explicit'
                          ? selectedArchetype.rewardExperience ?? 0
                          : 0}
                      </span>
                      <span className="qa-chip muted">
                        Materials:{' '}
                        {summarizeMaterialRewards(
                          selectedArchetype.materialRewards
                        )}
                      </span>
                      {selectedArchetype.recurrenceFrequency && (
                        <span className="qa-chip accent">
                          Repeats {selectedArchetype.recurrenceFrequency}
                        </span>
                      )}
                    </div>
                    <div style={{ marginTop: 14 }}>
                      <div className="qa-meta">Acceptance Dialogue</div>
                      {(selectedArchetype.acceptanceDialogue ?? []).length >
                      0 ? (
                        <div className="qa-inline" style={{ marginTop: 10 }}>
                          {(selectedArchetype.acceptanceDialogue ?? []).map(
                            (line, index) => (
                              <span
                                key={`${selectedArchetype.id}-dialogue-${index}`}
                                className="qa-chip"
                              >
                                {line.text}
                                {line.effect ? ` (${line.effect})` : ''}
                              </span>
                            )
                          )}
                        </div>
                      ) : (
                        <div className="qa-empty" style={{ marginTop: 10 }}>
                          No acceptance dialogue configured.
                        </div>
                      )}
                    </div>
                    {(selectedArchetype.itemRewards ?? []).length > 0 ? (
                      <div className="qa-inline" style={{ marginTop: 10 }}>
                        {(selectedArchetype.itemRewards ?? []).map((reward) => {
                          const item = inventoryItems.find(
                            (entry) => entry.id === reward.inventoryItemId
                          );
                          return (
                            <span
                              key={
                                reward.id ??
                                `${reward.inventoryItemId}-${reward.quantity}`
                              }
                              className="qa-chip"
                            >
                              {reward.quantity}x{' '}
                              {item?.name ?? `Item ${reward.inventoryItemId}`}
                            </span>
                          );
                        })}
                      </div>
                    ) : null}
                    {(selectedArchetype.spellRewards ?? []).length > 0 && (
                      <div className="qa-inline" style={{ marginTop: 10 }}>
                        {(selectedArchetype.spellRewards ?? []).map(
                          (reward) => (
                            <span
                              key={reward.id ?? reward.spellId}
                              className="qa-chip success"
                            >
                              {reward.spell?.name ?? reward.spellId}
                            </span>
                          )
                        )}
                      </div>
                    )}
                  </div>
                </div>

                {flowMapLayout && (
                  <div className="qa-card qa-flow-map">
                    <div className="qa-card-title">Flow Map</div>
                    <p className="qa-muted" style={{ marginTop: 6 }}>
                      A mini-map of the quest flow. Each node represents a
                      location archetype, connected by challenges.
                    </p>
                    <div className="qa-flow-map-canvas">
                      <svg
                        viewBox={`0 0 ${flowMapLayout.width} ${flowMapLayout.height}`}
                        role="img"
                        aria-label="Quest archetype flow map"
                      >
                        <defs>
                          <marker
                            id="qa-flow-arrow"
                            markerWidth="8"
                            markerHeight="8"
                            refX="6"
                            refY="3"
                            orient="auto"
                            markerUnits="strokeWidth"
                          >
                            <path
                              d="M0,0 L0,6 L6,3 z"
                              fill="rgba(255,255,255,0.65)"
                            />
                          </marker>
                        </defs>
                        {flowMapLayout.edges.map((edge, index) => (
                          <line
                            key={`${edge.fromX}-${edge.fromY}-${edge.toX}-${edge.toY}-${index}`}
                            x1={edge.fromX}
                            y1={edge.fromY}
                            x2={edge.toX}
                            y2={edge.toY}
                            stroke="rgba(255,255,255,0.3)"
                            strokeWidth="2"
                            markerEnd="url(#qa-flow-arrow)"
                          />
                        ))}
                        {flowMapLayout.nodes.map((node) => (
                          <g key={node.id}>
                            <circle
                              cx={node.x}
                              cy={node.y}
                              r="12"
                              fill="rgba(255,107,74,0.7)"
                              stroke="rgba(255,255,255,0.8)"
                              strokeWidth="1"
                            />
                            <text
                              x={node.x + 18}
                              y={node.y + 4}
                              fill="rgba(236,243,245,0.9)"
                              fontSize="11"
                              fontFamily="Space Grotesk, sans-serif"
                            >
                              {node.label}
                            </text>
                          </g>
                        ))}
                      </svg>
                    </div>
                  </div>
                )}

                <div className="qa-card qa-builder-flow">
                  <div className="qa-card-header">
                    <div>
                      <div className="qa-card-title">Quest Route</div>
                      <p className="qa-muted" style={{ marginTop: 6 }}>
                        Read the archetype as a route on the left, then tune one
                        beat at a time in the inspector.
                      </p>
                    </div>
                    <div className="qa-inline">
                      <span className="qa-chip accent">
                        {flowRoutes.length} nodes
                      </span>
                      <span className="qa-chip muted">
                        {flowRoutes.reduce(
                          (total, route) =>
                            total + (route.node.challenges?.length ?? 0),
                          0
                        )}{' '}
                        branches
                      </span>
                    </div>
                  </div>
                  {selectedArchetype.root ? (
                    <div className="qa-workbench">
                      <div className="qa-route-panel">
                        <div className="qa-route-list">
                          {flowRoutes.map((route) => {
                            const nodeTypeLabel = questArchetypeNodeTypeLabel(
                              route.node.nodeType
                            );
                            const routeSummary = describeQuestArchetypeNode(
                              route.node,
                              locationArchetypes,
                              monsterTemplates,
                              scenarioTemplates
                            );
                            return (
                              <button
                                key={route.id}
                                className={`qa-route-step ${selectedFlowNode?.id === route.id ? 'is-active' : ''}`}
                                style={{
                                  marginLeft: `${route.depth * 14}px`,
                                }}
                                onClick={() => setSelectedNodeId(route.id)}
                              >
                                <div className="qa-route-step-index">
                                  {route.pathLabel}
                                </div>
                                <div className="qa-route-step-body">
                                  <div className="qa-route-step-topline">
                                    <span className="qa-route-step-title">
                                      {routeSummary}
                                    </span>
                                    <span className="qa-chip muted">
                                      {nodeTypeLabel}
                                    </span>
                                  </div>
                                  <div className="qa-route-step-meta">
                                    {route.incomingChallenge
                                      ? `Unlocked by ${route.parentNodeId ? 'previous beat' : 'root'}`
                                      : 'Root beat'}
                                    {' · '}
                                    {route.node.challenges?.length ?? 0}{' '}
                                    outgoing
                                  </div>
                                  {route.incomingChallenge && (
                                    <div
                                      className="qa-inline"
                                      style={{ marginTop: 8 }}
                                    >
                                      {resolveChallengeProficiency(
                                        route.incomingChallenge,
                                        challengeTemplates
                                      ) && (
                                        <span className="qa-chip muted">
                                          {resolveChallengeProficiency(
                                            route.incomingChallenge,
                                            challengeTemplates
                                          )}
                                        </span>
                                      )}
                                    </div>
                                  )}
                                </div>
                              </button>
                            );
                          })}
                        </div>
                      </div>
                      <div className="qa-workbench-main">
                        {selectedFlowNode ? (
                          <>
                            <QuestNodeInspector
                              node={selectedFlowNode.node}
                              pathLabel={selectedFlowNode.pathLabel}
                              depth={selectedFlowNode.depth}
                              locationArchetypes={locationArchetypes}
                              monsterTemplates={monsterTemplates}
                              scenarioTemplates={scenarioTemplates}
                              challengeTemplates={challengeTemplates}
                              characters={characters}
                              inventoryItems={inventoryItems}
                              spells={spells}
                              addChallengeToQuestArchetype={
                                addChallengeToQuestArchetype
                              }
                              onSaveNode={updateQuestArchetypeNode}
                              onEditChallenge={openChallengeEditor}
                              onEditChallengeTemplate={
                                openInlineChallengeTemplateEditor
                              }
                              onEditScenarioTemplate={
                                openInlineScenarioTemplateEditor
                              }
                              onEditMonsterTemplate={
                                openInlineMonsterTemplateEditor
                              }
                              onSelectNode={setSelectedNodeId}
                            />
                            {inlineEditor?.kind === 'challenge' &&
                              activeChallengeTemplate && (
                                <div
                                  className="qa-card"
                                  style={{ marginTop: 18 }}
                                >
                                  <div className="qa-card-header">
                                    <div>
                                      <div className="qa-card-title">
                                        Challenge Template Editor
                                      </div>
                                      <div className="qa-meta">
                                        Editing{' '}
                                        {activeChallengeTemplate.question} for{' '}
                                        {inlineEditor.sourceLabel}.
                                      </div>
                                    </div>
                                    <div className="qa-inline">
                                      <span className="qa-chip muted">
                                        {locationArchetypes.find(
                                          (entry) =>
                                            entry.id ===
                                            activeChallengeTemplate.locationArchetypeId
                                        )?.name ?? 'Unknown location'}
                                      </span>
                                      <button
                                        className="qa-btn qa-btn-outline"
                                        onClick={() => setInlineEditor(null)}
                                      >
                                        Close
                                      </button>
                                    </div>
                                  </div>
                                  <div
                                    className="qa-form-grid"
                                    style={{ marginTop: 18 }}
                                  >
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Location Archetype
                                      </div>
                                      <select
                                        className="qa-select"
                                        value={
                                          challengeTemplateForm.locationArchetypeId
                                        }
                                        disabled={Boolean(
                                          inlineEditor.lockedLocationArchetypeId
                                        )}
                                        onChange={(e) =>
                                          setChallengeTemplateForm((prev) => ({
                                            ...prev,
                                            locationArchetypeId: e.target.value,
                                          }))
                                        }
                                      >
                                        <option value="">
                                          Select a location archetype
                                        </option>
                                        {locationArchetypes.map((archetype) => (
                                          <option
                                            key={`challenge-template-location-${archetype.id}`}
                                            value={archetype.id}
                                          >
                                            {archetype.name}
                                          </option>
                                        ))}
                                      </select>
                                      {inlineEditor.lockedLocationArchetypeId && (
                                        <div className="qa-helper">
                                          Locked to the location archetype used
                                          by this branch so the quest flow stays
                                          compatible.
                                        </div>
                                      )}
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Submission Type
                                      </div>
                                      <select
                                        className="qa-select"
                                        value={
                                          challengeTemplateForm.submissionType
                                        }
                                        onChange={(e) =>
                                          setChallengeTemplateForm((prev) => ({
                                            ...prev,
                                            submissionType: e.target.value as
                                              | 'photo'
                                              | 'text'
                                              | 'video',
                                          }))
                                        }
                                      >
                                        <option value="photo">Photo</option>
                                        <option value="text">Text</option>
                                        <option value="video">Video</option>
                                      </select>
                                    </div>
                                    <div
                                      className="qa-field"
                                      style={{ gridColumn: '1 / -1' }}
                                    >
                                      <div className="qa-label">Question</div>
                                      <input
                                        type="text"
                                        className="qa-input"
                                        value={challengeTemplateForm.question}
                                        onChange={(e) =>
                                          setChallengeTemplateForm((prev) => ({
                                            ...prev,
                                            question: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div
                                      className="qa-field"
                                      style={{ gridColumn: '1 / -1' }}
                                    >
                                      <div className="qa-label">
                                        Description
                                      </div>
                                      <textarea
                                        className="qa-textarea"
                                        rows={5}
                                        value={
                                          challengeTemplateForm.description
                                        }
                                        onChange={(e) =>
                                          setChallengeTemplateForm((prev) => ({
                                            ...prev,
                                            description: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">Image URL</div>
                                      <input
                                        type="text"
                                        className="qa-input"
                                        value={challengeTemplateForm.imageUrl}
                                        onChange={(e) =>
                                          setChallengeTemplateForm((prev) => ({
                                            ...prev,
                                            imageUrl: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Thumbnail URL
                                      </div>
                                      <input
                                        type="text"
                                        className="qa-input"
                                        value={
                                          challengeTemplateForm.thumbnailUrl
                                        }
                                        onChange={(e) =>
                                          setChallengeTemplateForm((prev) => ({
                                            ...prev,
                                            thumbnailUrl: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">Difficulty</div>
                                      <input
                                        type="number"
                                        min={0}
                                        className="qa-input"
                                        value={challengeTemplateForm.difficulty}
                                        onChange={(e) =>
                                          setChallengeTemplateForm((prev) => ({
                                            ...prev,
                                            difficulty: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Proficiency
                                      </div>
                                      <input
                                        type="text"
                                        className="qa-input"
                                        value={
                                          challengeTemplateForm.proficiency
                                        }
                                        onChange={(e) =>
                                          setChallengeTemplateForm((prev) => ({
                                            ...prev,
                                            proficiency: e.target.value,
                                          }))
                                        }
                                        list="qa-proficiency-options"
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Reward Mode
                                      </div>
                                      <select
                                        className="qa-select"
                                        value={challengeTemplateForm.rewardMode}
                                        onChange={(e) =>
                                          setChallengeTemplateForm((prev) => ({
                                            ...prev,
                                            rewardMode: e.target.value as
                                              | 'explicit'
                                              | 'random',
                                          }))
                                        }
                                      >
                                        <option value="random">Random</option>
                                        <option value="explicit">
                                          Explicit
                                        </option>
                                      </select>
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Random Reward Size
                                      </div>
                                      <select
                                        className="qa-select"
                                        value={
                                          challengeTemplateForm.randomRewardSize
                                        }
                                        onChange={(e) =>
                                          setChallengeTemplateForm((prev) => ({
                                            ...prev,
                                            randomRewardSize: e.target.value as
                                              | 'small'
                                              | 'medium'
                                              | 'large',
                                          }))
                                        }
                                      >
                                        <option value="small">Small</option>
                                        <option value="medium">Medium</option>
                                        <option value="large">Large</option>
                                      </select>
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">Reward</div>
                                      <input
                                        type="number"
                                        min={0}
                                        className="qa-input"
                                        value={challengeTemplateForm.reward}
                                        onChange={(e) =>
                                          setChallengeTemplateForm((prev) => ({
                                            ...prev,
                                            reward: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Reward Experience
                                      </div>
                                      <input
                                        type="number"
                                        min={0}
                                        className="qa-input"
                                        value={
                                          challengeTemplateForm.rewardExperience
                                        }
                                        onChange={(e) =>
                                          setChallengeTemplateForm((prev) => ({
                                            ...prev,
                                            rewardExperience: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Inventory Item
                                      </div>
                                      <select
                                        className="qa-select"
                                        value={
                                          challengeTemplateForm.inventoryItemId
                                        }
                                        onChange={(e) =>
                                          setChallengeTemplateForm((prev) => ({
                                            ...prev,
                                            inventoryItemId: e.target.value,
                                          }))
                                        }
                                      >
                                        <option value="">None</option>
                                        {inventoryItems.map((item) => (
                                          <option
                                            key={`challenge-template-item-${item.id}`}
                                            value={item.id}
                                          >
                                            {item.name}
                                          </option>
                                        ))}
                                      </select>
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">Stat Tags</div>
                                      <input
                                        type="text"
                                        className="qa-input"
                                        value={challengeTemplateForm.statTags}
                                        onChange={(e) =>
                                          setChallengeTemplateForm((prev) => ({
                                            ...prev,
                                            statTags: e.target.value,
                                          }))
                                        }
                                        placeholder="strength, dexterity"
                                      />
                                    </div>
                                    <div
                                      className="qa-field"
                                      style={{ gridColumn: '1 / -1' }}
                                    >
                                      <label className="qa-inline">
                                        <input
                                          type="checkbox"
                                          checked={
                                            challengeTemplateForm.scaleWithUserLevel
                                          }
                                          onChange={(e) =>
                                            setChallengeTemplateForm(
                                              (prev) => ({
                                                ...prev,
                                                scaleWithUserLevel:
                                                  e.target.checked,
                                              })
                                            )
                                          }
                                        />
                                        <span
                                          className="qa-label"
                                          style={{ marginBottom: 0 }}
                                        >
                                          Scale with user level
                                        </span>
                                      </label>
                                    </div>
                                    <div
                                      className="qa-field"
                                      style={{ gridColumn: '1 / -1' }}
                                    >
                                      <div className="qa-label">
                                        Item Choice Rewards JSON
                                      </div>
                                      <textarea
                                        className="qa-textarea"
                                        rows={8}
                                        value={
                                          challengeTemplateForm.itemChoiceRewardsJson
                                        }
                                        onChange={(e) =>
                                          setChallengeTemplateForm((prev) => ({
                                            ...prev,
                                            itemChoiceRewardsJson:
                                              e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                  </div>
                                  {inlineEditorError && (
                                    <div
                                      className="qa-chip danger"
                                      style={{ marginTop: 16 }}
                                    >
                                      {inlineEditorError}
                                    </div>
                                  )}
                                  {inlineEditorSuccess && (
                                    <div
                                      className="qa-chip success"
                                      style={{ marginTop: 16 }}
                                    >
                                      {inlineEditorSuccess}
                                    </div>
                                  )}
                                  <div className="qa-footer">
                                    <button
                                      className="qa-btn qa-btn-outline"
                                      onClick={() =>
                                        setChallengeTemplateForm(
                                          challengeTemplateFormFromRecord(
                                            activeChallengeTemplate
                                          )
                                        )
                                      }
                                    >
                                      Reset
                                    </button>
                                    <button
                                      className="qa-btn qa-btn-primary"
                                      disabled={inlineEditorSaving}
                                      onClick={
                                        handleSaveInlineChallengeTemplate
                                      }
                                    >
                                      {inlineEditorSaving
                                        ? 'Saving...'
                                        : 'Save Challenge Template'}
                                    </button>
                                  </div>
                                </div>
                              )}
                            {inlineEditor?.kind === 'scenario' &&
                              activeScenarioTemplate && (
                                <div
                                  className="qa-card"
                                  style={{ marginTop: 18 }}
                                >
                                  <div className="qa-card-header">
                                    <div>
                                      <div className="qa-card-title">
                                        Scenario Template Editor
                                      </div>
                                      <div className="qa-meta">
                                        Editing the linked scenario template for{' '}
                                        {inlineEditor.sourceLabel}.
                                      </div>
                                    </div>
                                    <button
                                      className="qa-btn qa-btn-outline"
                                      onClick={() => setInlineEditor(null)}
                                    >
                                      Close
                                    </button>
                                  </div>
                                  <div
                                    className="qa-form-grid"
                                    style={{ marginTop: 18 }}
                                  >
                                    <div
                                      className="qa-field"
                                      style={{ gridColumn: '1 / -1' }}
                                    >
                                      <div className="qa-label">Prompt</div>
                                      <textarea
                                        className="qa-textarea"
                                        rows={4}
                                        value={scenarioTemplateForm.prompt}
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            prompt: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">Image URL</div>
                                      <input
                                        type="text"
                                        className="qa-input"
                                        value={scenarioTemplateForm.imageUrl}
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            imageUrl: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Thumbnail URL
                                      </div>
                                      <input
                                        type="text"
                                        className="qa-input"
                                        value={
                                          scenarioTemplateForm.thumbnailUrl
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            thumbnailUrl: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">Difficulty</div>
                                      <input
                                        type="number"
                                        min={0}
                                        className="qa-input"
                                        value={scenarioTemplateForm.difficulty}
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            difficulty: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Reward Mode
                                      </div>
                                      <select
                                        className="qa-select"
                                        value={scenarioTemplateForm.rewardMode}
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            rewardMode: e.target.value as
                                              | 'explicit'
                                              | 'random',
                                          }))
                                        }
                                      >
                                        <option value="random">Random</option>
                                        <option value="explicit">
                                          Explicit
                                        </option>
                                      </select>
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Random Reward Size
                                      </div>
                                      <select
                                        className="qa-select"
                                        value={
                                          scenarioTemplateForm.randomRewardSize
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            randomRewardSize: e.target.value as
                                              | 'small'
                                              | 'medium'
                                              | 'large',
                                          }))
                                        }
                                      >
                                        <option value="small">Small</option>
                                        <option value="medium">Medium</option>
                                        <option value="large">Large</option>
                                      </select>
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Reward Experience
                                      </div>
                                      <input
                                        type="number"
                                        min={0}
                                        className="qa-input"
                                        value={
                                          scenarioTemplateForm.rewardExperience
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            rewardExperience: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Reward Gold
                                      </div>
                                      <input
                                        type="number"
                                        min={0}
                                        className="qa-input"
                                        value={scenarioTemplateForm.rewardGold}
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            rewardGold: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <label className="qa-inline">
                                        <input
                                          type="checkbox"
                                          checked={
                                            scenarioTemplateForm.openEnded
                                          }
                                          onChange={(e) =>
                                            setScenarioTemplateForm((prev) => ({
                                              ...prev,
                                              openEnded: e.target.checked,
                                            }))
                                          }
                                        />
                                        <span
                                          className="qa-label"
                                          style={{ marginBottom: 0 }}
                                        >
                                          Open ended
                                        </span>
                                      </label>
                                    </div>
                                    <div className="qa-field">
                                      <label className="qa-inline">
                                        <input
                                          type="checkbox"
                                          checked={
                                            scenarioTemplateForm.scaleWithUserLevel
                                          }
                                          onChange={(e) =>
                                            setScenarioTemplateForm((prev) => ({
                                              ...prev,
                                              scaleWithUserLevel:
                                                e.target.checked,
                                            }))
                                          }
                                        />
                                        <span
                                          className="qa-label"
                                          style={{ marginBottom: 0 }}
                                        >
                                          Scale with user level
                                        </span>
                                      </label>
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Failure Penalty Mode
                                      </div>
                                      <select
                                        className="qa-select"
                                        value={
                                          scenarioTemplateForm.failurePenaltyMode
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            failurePenaltyMode: e.target
                                              .value as 'shared' | 'individual',
                                          }))
                                        }
                                      >
                                        <option value="shared">Shared</option>
                                        <option value="individual">
                                          Individual
                                        </option>
                                      </select>
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Failure Health Drain
                                      </div>
                                      <select
                                        className="qa-select"
                                        value={
                                          scenarioTemplateForm.failureHealthDrainType
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            failureHealthDrainType: e.target
                                              .value as
                                              | 'none'
                                              | 'flat'
                                              | 'percent',
                                          }))
                                        }
                                      >
                                        <option value="none">None</option>
                                        <option value="flat">Flat</option>
                                        <option value="percent">Percent</option>
                                      </select>
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Failure Health Value
                                      </div>
                                      <input
                                        type="number"
                                        className="qa-input"
                                        value={
                                          scenarioTemplateForm.failureHealthDrainValue
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            failureHealthDrainValue:
                                              e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Failure Mana Drain
                                      </div>
                                      <select
                                        className="qa-select"
                                        value={
                                          scenarioTemplateForm.failureManaDrainType
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            failureManaDrainType: e.target
                                              .value as
                                              | 'none'
                                              | 'flat'
                                              | 'percent',
                                          }))
                                        }
                                      >
                                        <option value="none">None</option>
                                        <option value="flat">Flat</option>
                                        <option value="percent">Percent</option>
                                      </select>
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Failure Mana Value
                                      </div>
                                      <input
                                        type="number"
                                        className="qa-input"
                                        value={
                                          scenarioTemplateForm.failureManaDrainValue
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            failureManaDrainValue:
                                              e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Success Reward Mode
                                      </div>
                                      <select
                                        className="qa-select"
                                        value={
                                          scenarioTemplateForm.successRewardMode
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            successRewardMode: e.target
                                              .value as 'shared' | 'individual',
                                          }))
                                        }
                                      >
                                        <option value="shared">Shared</option>
                                        <option value="individual">
                                          Individual
                                        </option>
                                      </select>
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Success Health Restore
                                      </div>
                                      <select
                                        className="qa-select"
                                        value={
                                          scenarioTemplateForm.successHealthRestoreType
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            successHealthRestoreType: e.target
                                              .value as
                                              | 'none'
                                              | 'flat'
                                              | 'percent',
                                          }))
                                        }
                                      >
                                        <option value="none">None</option>
                                        <option value="flat">Flat</option>
                                        <option value="percent">Percent</option>
                                      </select>
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Success Health Value
                                      </div>
                                      <input
                                        type="number"
                                        className="qa-input"
                                        value={
                                          scenarioTemplateForm.successHealthRestoreValue
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            successHealthRestoreValue:
                                              e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Success Mana Restore
                                      </div>
                                      <select
                                        className="qa-select"
                                        value={
                                          scenarioTemplateForm.successManaRestoreType
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            successManaRestoreType: e.target
                                              .value as
                                              | 'none'
                                              | 'flat'
                                              | 'percent',
                                          }))
                                        }
                                      >
                                        <option value="none">None</option>
                                        <option value="flat">Flat</option>
                                        <option value="percent">Percent</option>
                                      </select>
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Success Mana Value
                                      </div>
                                      <input
                                        type="number"
                                        className="qa-input"
                                        value={
                                          scenarioTemplateForm.successManaRestoreValue
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            successManaRestoreValue:
                                              e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div
                                      className="qa-field"
                                      style={{ gridColumn: '1 / -1' }}
                                    >
                                      <div className="qa-label">
                                        Failure Statuses JSON
                                      </div>
                                      <textarea
                                        className="qa-textarea"
                                        rows={6}
                                        value={
                                          scenarioTemplateForm.failureStatusesJson
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            failureStatusesJson: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div
                                      className="qa-field"
                                      style={{ gridColumn: '1 / -1' }}
                                    >
                                      <div className="qa-label">
                                        Success Statuses JSON
                                      </div>
                                      <textarea
                                        className="qa-textarea"
                                        rows={6}
                                        value={
                                          scenarioTemplateForm.successStatusesJson
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            successStatusesJson: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div
                                      className="qa-field"
                                      style={{ gridColumn: '1 / -1' }}
                                    >
                                      <div className="qa-label">
                                        Options JSON
                                      </div>
                                      <textarea
                                        className="qa-textarea"
                                        rows={8}
                                        value={scenarioTemplateForm.optionsJson}
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            optionsJson: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div
                                      className="qa-field"
                                      style={{ gridColumn: '1 / -1' }}
                                    >
                                      <div className="qa-label">
                                        Item Rewards JSON
                                      </div>
                                      <textarea
                                        className="qa-textarea"
                                        rows={6}
                                        value={
                                          scenarioTemplateForm.itemRewardsJson
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            itemRewardsJson: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div
                                      className="qa-field"
                                      style={{ gridColumn: '1 / -1' }}
                                    >
                                      <div className="qa-label">
                                        Item Choice Rewards JSON
                                      </div>
                                      <textarea
                                        className="qa-textarea"
                                        rows={6}
                                        value={
                                          scenarioTemplateForm.itemChoiceRewardsJson
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            itemChoiceRewardsJson:
                                              e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div
                                      className="qa-field"
                                      style={{ gridColumn: '1 / -1' }}
                                    >
                                      <div className="qa-label">
                                        Spell Rewards JSON
                                      </div>
                                      <textarea
                                        className="qa-textarea"
                                        rows={6}
                                        value={
                                          scenarioTemplateForm.spellRewardsJson
                                        }
                                        onChange={(e) =>
                                          setScenarioTemplateForm((prev) => ({
                                            ...prev,
                                            spellRewardsJson: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                  </div>
                                  {inlineEditorError && (
                                    <div
                                      className="qa-chip danger"
                                      style={{ marginTop: 16 }}
                                    >
                                      {inlineEditorError}
                                    </div>
                                  )}
                                  {inlineEditorSuccess && (
                                    <div
                                      className="qa-chip success"
                                      style={{ marginTop: 16 }}
                                    >
                                      {inlineEditorSuccess}
                                    </div>
                                  )}
                                  <div className="qa-footer">
                                    <button
                                      className="qa-btn qa-btn-outline"
                                      onClick={() =>
                                        setScenarioTemplateForm(
                                          scenarioTemplateFormFromRecord(
                                            activeScenarioTemplate
                                          )
                                        )
                                      }
                                    >
                                      Reset
                                    </button>
                                    <button
                                      className="qa-btn qa-btn-primary"
                                      disabled={inlineEditorSaving}
                                      onClick={handleSaveInlineScenarioTemplate}
                                    >
                                      {inlineEditorSaving
                                        ? 'Saving...'
                                        : 'Save Scenario Template'}
                                    </button>
                                  </div>
                                </div>
                              )}
                            {inlineEditor?.kind === 'monster' &&
                              activeMonsterTemplate && (
                                <div
                                  className="qa-card"
                                  style={{ marginTop: 18 }}
                                >
                                  <div className="qa-card-header">
                                    <div>
                                      <div className="qa-card-title">
                                        Monster Template Editor
                                      </div>
                                      <div className="qa-meta">
                                        Editing the monster template linked from{' '}
                                        {inlineEditor.sourceLabel}.
                                      </div>
                                    </div>
                                    <div className="qa-inline">
                                      <span className="qa-chip muted">
                                        {activeMonsterTemplate.monsterType ??
                                          'monster'}
                                      </span>
                                      <button
                                        className="qa-btn qa-btn-outline"
                                        onClick={() => setInlineEditor(null)}
                                      >
                                        Close
                                      </button>
                                    </div>
                                  </div>
                                  <div
                                    className="qa-form-grid"
                                    style={{ marginTop: 18 }}
                                  >
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Monster Type
                                      </div>
                                      <select
                                        className="qa-select"
                                        value={monsterTemplateForm.monsterType}
                                        onChange={(e) =>
                                          setMonsterTemplateForm((prev) => ({
                                            ...prev,
                                            monsterType: e.target.value as
                                              | 'monster'
                                              | 'boss'
                                              | 'raid',
                                          }))
                                        }
                                      >
                                        <option value="monster">
                                          Standard Monster
                                        </option>
                                        <option value="boss">
                                          Boss Monster
                                        </option>
                                        <option value="raid">
                                          Raid Monster
                                        </option>
                                      </select>
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">Name</div>
                                      <input
                                        type="text"
                                        className="qa-input"
                                        value={monsterTemplateForm.name}
                                        onChange={(e) =>
                                          setMonsterTemplateForm((prev) => ({
                                            ...prev,
                                            name: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div
                                      className="qa-field"
                                      style={{ gridColumn: '1 / -1' }}
                                    >
                                      <div className="qa-label">
                                        Description
                                      </div>
                                      <textarea
                                        className="qa-textarea"
                                        rows={4}
                                        value={monsterTemplateForm.description}
                                        onChange={(e) =>
                                          setMonsterTemplateForm((prev) => ({
                                            ...prev,
                                            description: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">Image URL</div>
                                      <input
                                        type="text"
                                        className="qa-input"
                                        value={monsterTemplateForm.imageUrl}
                                        onChange={(e) =>
                                          setMonsterTemplateForm((prev) => ({
                                            ...prev,
                                            imageUrl: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Thumbnail URL
                                      </div>
                                      <input
                                        type="text"
                                        className="qa-input"
                                        value={monsterTemplateForm.thumbnailUrl}
                                        onChange={(e) =>
                                          setMonsterTemplateForm((prev) => ({
                                            ...prev,
                                            thumbnailUrl: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Base Strength
                                      </div>
                                      <input
                                        type="number"
                                        min={1}
                                        className="qa-input"
                                        value={monsterTemplateForm.baseStrength}
                                        onChange={(e) =>
                                          setMonsterTemplateForm((prev) => ({
                                            ...prev,
                                            baseStrength: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Base Dexterity
                                      </div>
                                      <input
                                        type="number"
                                        min={1}
                                        className="qa-input"
                                        value={
                                          monsterTemplateForm.baseDexterity
                                        }
                                        onChange={(e) =>
                                          setMonsterTemplateForm((prev) => ({
                                            ...prev,
                                            baseDexterity: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Base Constitution
                                      </div>
                                      <input
                                        type="number"
                                        min={1}
                                        className="qa-input"
                                        value={
                                          monsterTemplateForm.baseConstitution
                                        }
                                        onChange={(e) =>
                                          setMonsterTemplateForm((prev) => ({
                                            ...prev,
                                            baseConstitution: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Base Intelligence
                                      </div>
                                      <input
                                        type="number"
                                        min={1}
                                        className="qa-input"
                                        value={
                                          monsterTemplateForm.baseIntelligence
                                        }
                                        onChange={(e) =>
                                          setMonsterTemplateForm((prev) => ({
                                            ...prev,
                                            baseIntelligence: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Base Wisdom
                                      </div>
                                      <input
                                        type="number"
                                        min={1}
                                        className="qa-input"
                                        value={monsterTemplateForm.baseWisdom}
                                        onChange={(e) =>
                                          setMonsterTemplateForm((prev) => ({
                                            ...prev,
                                            baseWisdom: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Base Charisma
                                      </div>
                                      <input
                                        type="number"
                                        min={1}
                                        className="qa-input"
                                        value={monsterTemplateForm.baseCharisma}
                                        onChange={(e) =>
                                          setMonsterTemplateForm((prev) => ({
                                            ...prev,
                                            baseCharisma: e.target.value,
                                          }))
                                        }
                                      />
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Strong Against
                                      </div>
                                      <select
                                        className="qa-select"
                                        value={
                                          monsterTemplateForm.strongAgainstAffinity
                                        }
                                        onChange={(e) =>
                                          setMonsterTemplateForm((prev) => ({
                                            ...prev,
                                            strongAgainstAffinity:
                                              e.target.value,
                                          }))
                                        }
                                      >
                                        <option value="">None</option>
                                        {damageAffinityOptions.map(
                                          (affinity) => (
                                            <option
                                              key={`monster-strong-${affinity}`}
                                              value={affinity}
                                            >
                                              {affinity}
                                            </option>
                                          )
                                        )}
                                      </select>
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Weak Against
                                      </div>
                                      <select
                                        className="qa-select"
                                        value={
                                          monsterTemplateForm.weakAgainstAffinity
                                        }
                                        onChange={(e) =>
                                          setMonsterTemplateForm((prev) => ({
                                            ...prev,
                                            weakAgainstAffinity: e.target.value,
                                          }))
                                        }
                                      >
                                        <option value="">None</option>
                                        {damageAffinityOptions.map(
                                          (affinity) => (
                                            <option
                                              key={`monster-weak-${affinity}`}
                                              value={affinity}
                                            >
                                              {affinity}
                                            </option>
                                          )
                                        )}
                                      </select>
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Spell Progressions
                                      </div>
                                      <select
                                        className="qa-select"
                                        multiple
                                        size={8}
                                        value={
                                          monsterTemplateForm.spellProgressionIds
                                        }
                                        onChange={(e) =>
                                          setMonsterTemplateForm((prev) => ({
                                            ...prev,
                                            spellProgressionIds: Array.from(
                                              e.target.selectedOptions
                                            ).map((option) => option.value),
                                          }))
                                        }
                                      >
                                        {monsterSpellProgressionOptions.map(
                                          (progression) => (
                                            <option
                                              key={`monster-spell-progression-${progression.id}`}
                                              value={progression.id}
                                            >
                                              {progression.name} (
                                              {progression.memberCount})
                                            </option>
                                          )
                                        )}
                                      </select>
                                    </div>
                                    <div className="qa-field">
                                      <div className="qa-label">
                                        Technique Progressions
                                      </div>
                                      <select
                                        className="qa-select"
                                        multiple
                                        size={8}
                                        value={
                                          monsterTemplateForm.techniqueProgressionIds
                                        }
                                        onChange={(e) =>
                                          setMonsterTemplateForm((prev) => ({
                                            ...prev,
                                            techniqueProgressionIds: Array.from(
                                              e.target.selectedOptions
                                            ).map((option) => option.value),
                                          }))
                                        }
                                      >
                                        {monsterTechniqueProgressionOptions.map(
                                          (progression) => (
                                            <option
                                              key={`monster-technique-progression-${progression.id}`}
                                              value={progression.id}
                                            >
                                              {progression.name} (
                                              {progression.memberCount})
                                            </option>
                                          )
                                        )}
                                      </select>
                                    </div>
                                  </div>
                                  {inlineEditorError && (
                                    <div
                                      className="qa-chip danger"
                                      style={{ marginTop: 16 }}
                                    >
                                      {inlineEditorError}
                                    </div>
                                  )}
                                  {inlineEditorSuccess && (
                                    <div
                                      className="qa-chip success"
                                      style={{ marginTop: 16 }}
                                    >
                                      {inlineEditorSuccess}
                                    </div>
                                  )}
                                  <div className="qa-footer">
                                    <button
                                      className="qa-btn qa-btn-outline"
                                      onClick={() =>
                                        setMonsterTemplateForm(
                                          monsterTemplateFormFromRecord(
                                            activeMonsterTemplate
                                          )
                                        )
                                      }
                                    >
                                      Reset
                                    </button>
                                    <button
                                      className="qa-btn qa-btn-primary"
                                      disabled={inlineEditorSaving}
                                      onClick={handleSaveInlineMonsterTemplate}
                                    >
                                      {inlineEditorSaving
                                        ? 'Saving...'
                                        : 'Save Monster Template'}
                                    </button>
                                  </div>
                                </div>
                              )}
                          </>
                        ) : (
                          <div className="qa-empty">
                            Select a node to inspect it.
                          </div>
                        )}
                      </div>
                    </div>
                  ) : (
                    <div className="qa-empty" style={{ marginTop: 12 }}>
                      No root node available.
                    </div>
                  )}
                </div>
              </>
            )}
          </div>
        </section>
      </div>

      {shouldShowGenerateQuestModal && selectedArchetype && (
        <div className="qa-modal">
          <div className="qa-modal-card">
            <h2 className="qa-modal-title">Generate Quest In Zone</h2>
            <p className="qa-muted" style={{ marginBottom: 16 }}>
              Spin up one concrete quest from{' '}
              <strong>{selectedArchetype.name}</strong> in a specific zone. The
              archetype will keep its explicit copy and structure, while quest
              generation chooses the best-fit characters, points of interest,
              and exact placements.
            </p>

            <div className="qa-form-grid">
              <div className="qa-field" style={{ gridColumn: '1 / -1' }}>
                <div className="qa-label">Zone Search</div>
                <input
                  type="text"
                  className="qa-input"
                  value={generateQuestZoneSearch}
                  onChange={(e) => setGenerateQuestZoneSearch(e.target.value)}
                  placeholder="Search zones..."
                />
              </div>

              <div className="qa-field">
                <div className="qa-label">Pick Zone</div>
                {filteredGenerationZones.length === 0 ? (
                  <div className="qa-empty">No zones match that search.</div>
                ) : (
                  <div
                    className="qa-sidebar-list"
                    style={{ maxHeight: 280, marginTop: 8 }}
                  >
                    {filteredGenerationZones.map((zone) => (
                      <button
                        key={zone.id}
                        type="button"
                        className={`qa-sidebar-item ${selectedGenerateZoneId === zone.id ? 'is-active' : ''}`}
                        onClick={() => {
                          setSelectedGenerateZoneId(zone.id);
                          setGenerateQuestError('');
                          setQuestGenerationJob(null);
                        }}
                      >
                        <div className="qa-sidebar-item-title">{zone.name}</div>
                        <div className="qa-meta">
                          {(zone.internalTags ?? []).length > 0
                            ? (zone.internalTags ?? []).join(', ')
                            : 'No internal tags'}
                        </div>
                      </button>
                    ))}
                  </div>
                )}
              </div>

              <div className="qa-field">
                <div className="qa-label">Generation Brief</div>
                <div className="qa-panel" style={{ minHeight: 280 }}>
                  <div className="qa-card-header" style={{ marginBottom: 0 }}>
                    <div className="qa-card-title">
                      {selectedGenerateZone?.name ?? 'Select a zone'}
                    </div>
                    {questGenerationJob && (
                      <span
                        className={questGenerationStatusChipClass(
                          questGenerationJob.status
                        )}
                      >
                        {formatQuestGenerationStatus(questGenerationJob.status)}
                      </span>
                    )}
                  </div>
                  <p className="qa-muted" style={{ marginTop: 8 }}>
                    {selectedArchetype.category === 'main_story'
                      ? `Quest giver: ${selectedArchetypeQuestGiver?.name ?? 'Unassigned'}`
                      : (selectedArchetype.characterTags ?? []).length > 0
                        ? `Quest giver tags: ${selectedArchetype.characterTags?.join(', ')}`
                        : 'Quest giver: no tag preference'}
                  </p>
                  <p className="qa-muted" style={{ marginTop: 8 }}>
                    {(selectedArchetype.internalTags ?? []).length > 0
                      ? `Template tags: ${selectedArchetype.internalTags?.join(', ')}`
                      : 'Template tags: none'}
                  </p>
                  {selectedGenerateZone && (
                    <div className="qa-inline" style={{ marginTop: 12 }}>
                      <span className="qa-chip muted">
                        Zone Tags:{' '}
                        {(selectedGenerateZone.internalTags ?? []).length > 0
                          ? selectedGenerateZone.internalTags?.join(', ')
                          : 'None'}
                      </span>
                    </div>
                  )}
                  {questGenerationJob && (
                    <div className="qa-inline" style={{ marginTop: 12 }}>
                      <span className="qa-chip muted">
                        Progress: {questGenerationJob.completedCount}/
                        {questGenerationJob.totalCount}
                      </span>
                      <span className="qa-chip muted">
                        Failed: {questGenerationJob.failedCount}
                      </span>
                    </div>
                  )}
                  {generatedQuest && (
                    <div className="qa-panel" style={{ marginTop: 16 }}>
                      <div className="qa-card-title">Quest generated</div>
                      <div className="qa-meta" style={{ marginTop: 6 }}>
                        {generatedQuest.name || 'Untitled Quest'} ·{' '}
                        {generatedQuest.id.slice(0, 8)}…
                      </div>
                      <p className="qa-muted" style={{ marginTop: 10 }}>
                        {generatedQuest.description ||
                          'The quest was created successfully.'}
                      </p>
                      <div className="qa-inline" style={{ marginTop: 10 }}>
                        <span className="qa-chip success">
                          Zone: {selectedGenerateZone?.name}
                        </span>
                        {generatedQuest.questGiverCharacterId && (
                          <span className="qa-chip accent">
                            Quest giver assigned
                          </span>
                        )}
                        <span className="qa-chip muted">
                          Open it from the Quests screen.
                        </span>
                      </div>
                    </div>
                  )}
                  {questGenerationJob &&
                    isPendingQuestGenerationStatus(
                      questGenerationJob.status
                    ) && (
                      <div className="qa-panel" style={{ marginTop: 16 }}>
                        <div className="qa-card-title">Generation queued</div>
                        <p className="qa-muted" style={{ marginTop: 10 }}>
                          This quest is being generated in the background. You
                          can keep this modal open to watch it finish.
                        </p>
                      </div>
                    )}
                </div>
              </div>

              {generateQuestError && (
                <div
                  className="qa-chip danger"
                  style={{ gridColumn: '1 / -1', marginTop: 4 }}
                >
                  {generateQuestError}
                </div>
              )}
            </div>

            <div className="qa-footer">
              <button
                type="button"
                className="qa-btn qa-btn-outline"
                onClick={() => {
                  setShouldShowGenerateQuestModal(false);
                  setGenerateQuestZoneSearch('');
                  setSelectedGenerateZoneId('');
                  setGenerateQuestError('');
                  setQuestGenerationJob(null);
                }}
              >
                Close
              </button>
              <button
                type="button"
                className="qa-btn qa-btn-primary"
                disabled={
                  !selectedGenerateZoneId ||
                  isGeneratingQuest ||
                  isPendingQuestGenerationStatus(questGenerationJob?.status)
                }
                onClick={handleGenerateQuest}
              >
                {isGeneratingQuest
                  ? 'Queueing...'
                  : isPendingQuestGenerationStatus(questGenerationJob?.status)
                    ? 'Generating...'
                    : 'Generate Quest'}
              </button>
            </div>
          </div>
        </div>
      )}

      {shouldShowGeneratorModal && (
        <div className="qa-modal">
          <div className="qa-modal-card">
            <h2 className="qa-modal-title">Generate Quest Template</h2>
            <p className="qa-muted" style={{ marginBottom: 16 }}>
              Build an ordered quest flow from location and proximity steps. The
              generator will create a quest template with nodes in this exact
              order.
            </p>
            <form
              className="qa-form-grid"
              onSubmit={async (e) => {
                e.preventDefault();
                const validationError =
                  validateQuestTemplateGeneratorForm(generatorForm);
                if (validationError) {
                  window.alert(validationError);
                  return;
                }
                const created = await generateQuestArchetypeTemplate(
                  normalizeQuestTemplateGeneratorDraft(generatorForm)
                );
                if (created?.id) {
                  setSelectedArchetypeId(created.id);
                }
                setGeneratorForm(createEmptyQuestTemplateGeneratorForm());
                setShouldShowGeneratorModal(false);
              }}
            >
              <div className="qa-field">
                <div className="qa-label">Name</div>
                <input
                  type="text"
                  className="qa-input"
                  value={generatorForm.name}
                  onChange={(e) =>
                    setGeneratorForm((prev) => ({
                      ...prev,
                      name: e.target.value,
                    }))
                  }
                  placeholder="Optional generated template name"
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Theme Prompt</div>
                <textarea
                  className="qa-textarea"
                  rows={4}
                  value={generatorForm.themePrompt}
                  onChange={(e) =>
                    setGeneratorForm((prev) => ({
                      ...prev,
                      themePrompt: e.target.value,
                    }))
                  }
                  placeholder="Describe the kind of quest you want, tone, factions, stakes, and any notable beats."
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Character Tags</div>
                <input
                  type="text"
                  className="qa-input"
                  value={generatorForm.characterTagsText}
                  onChange={(e) =>
                    setGeneratorForm((prev) => ({
                      ...prev,
                      characterTagsText: e.target.value,
                    }))
                  }
                  placeholder="merchant, outlaw, druid"
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Internal Tags</div>
                <input
                  type="text"
                  className="qa-input"
                  value={generatorForm.internalTagsText}
                  onChange={(e) =>
                    setGeneratorForm((prev) => ({
                      ...prev,
                      internalTagsText: e.target.value,
                    }))
                  }
                  placeholder="city, mystery, waterfront"
                />
              </div>

              <div className="qa-field" style={{ gridColumn: '1 / -1' }}>
                <div className="qa-label">Add Step</div>
                <div className="qa-inline" style={{ flexWrap: 'wrap' }}>
                  <button
                    type="button"
                    className="qa-btn qa-btn-ghost"
                    onClick={() =>
                      setGeneratorForm((prev) => ({
                        ...prev,
                        steps: [
                          ...prev.steps,
                          createQuestTemplateGeneratorStep(
                            'location_archetype',
                            'challenge'
                          ),
                        ],
                      }))
                    }
                  >
                    Location Challenge
                  </button>
                  <button
                    type="button"
                    className="qa-btn qa-btn-ghost"
                    onClick={() =>
                      setGeneratorForm((prev) => ({
                        ...prev,
                        steps: [
                          ...prev.steps,
                          createQuestTemplateGeneratorStep(
                            'location_archetype',
                            'scenario'
                          ),
                        ],
                      }))
                    }
                  >
                    Location Scenario
                  </button>
                  <button
                    type="button"
                    className="qa-btn qa-btn-ghost"
                    onClick={() =>
                      setGeneratorForm((prev) => ({
                        ...prev,
                        steps: [
                          ...prev.steps,
                          createQuestTemplateGeneratorStep(
                            'location_archetype',
                            'monster'
                          ),
                        ],
                      }))
                    }
                  >
                    Location Monster
                  </button>
                  <button
                    type="button"
                    className="qa-btn qa-btn-ghost"
                    onClick={() =>
                      setGeneratorForm((prev) => ({
                        ...prev,
                        steps: [
                          ...prev.steps,
                          createQuestTemplateGeneratorStep(
                            'proximity',
                            'scenario'
                          ),
                        ],
                      }))
                    }
                  >
                    Nearby Scenario
                  </button>
                  <button
                    type="button"
                    className="qa-btn qa-btn-ghost"
                    onClick={() =>
                      setGeneratorForm((prev) => ({
                        ...prev,
                        steps: [
                          ...prev.steps,
                          createQuestTemplateGeneratorStep(
                            'proximity',
                            'monster'
                          ),
                        ],
                      }))
                    }
                  >
                    Nearby Monster
                  </button>
                </div>
              </div>

              <div className="qa-field" style={{ gridColumn: '1 / -1' }}>
                <div className="qa-label">Ordered Steps</div>
                {generatorForm.steps.length === 0 ? (
                  <div className="qa-empty">No steps yet.</div>
                ) : (
                  <div className="qa-flow-challenges">
                    {generatorForm.steps.map((step, index) => (
                      <div key={step.id} className="qa-flow-challenge-card">
                        <div className="qa-flow-challenge-header">
                          <div>
                            <div className="qa-flow-challenge-title">
                              Step {index + 1}
                            </div>
                            <div className="qa-meta">
                              {step.source === 'location_archetype'
                                ? 'Location archetype anchored'
                                : `Within ${step.proximityMeters}m of previous node`}
                            </div>
                          </div>
                          <div className="qa-inline">
                            <button
                              type="button"
                              className="qa-btn qa-btn-text"
                              disabled={index === 0}
                              onClick={() =>
                                setGeneratorForm((prev) => {
                                  const steps = [...prev.steps];
                                  [steps[index - 1], steps[index]] = [
                                    steps[index],
                                    steps[index - 1],
                                  ];
                                  return { ...prev, steps };
                                })
                              }
                            >
                              Up
                            </button>
                            <button
                              type="button"
                              className="qa-btn qa-btn-text"
                              disabled={
                                index === generatorForm.steps.length - 1
                              }
                              onClick={() =>
                                setGeneratorForm((prev) => {
                                  const steps = [...prev.steps];
                                  [steps[index], steps[index + 1]] = [
                                    steps[index + 1],
                                    steps[index],
                                  ];
                                  return { ...prev, steps };
                                })
                              }
                            >
                              Down
                            </button>
                            <button
                              type="button"
                              className="qa-btn qa-btn-text"
                              onClick={() =>
                                setGeneratorForm((prev) => ({
                                  ...prev,
                                  steps: prev.steps.filter(
                                    (entry) => entry.id !== step.id
                                  ),
                                }))
                              }
                            >
                              Remove
                            </button>
                          </div>
                        </div>

                        <div className="qa-form-grid">
                          <div className="qa-field">
                            <div className="qa-label">Anchor Type</div>
                            <select
                              className="qa-select"
                              value={step.source}
                              onChange={(e) =>
                                setGeneratorForm((prev) => ({
                                  ...prev,
                                  steps: prev.steps.map((entry) =>
                                    entry.id !== step.id
                                      ? entry
                                      : {
                                          ...entry,
                                          source: e.target
                                            .value as GeneratorStepSource,
                                          content:
                                            e.target.value === 'proximity' &&
                                            entry.content === 'challenge'
                                              ? 'scenario'
                                              : entry.content,
                                        }
                                  ),
                                }))
                              }
                            >
                              <option value="location_archetype">
                                Location Archetype
                              </option>
                              <option value="proximity" disabled={index === 0}>
                                Proximity To Previous Node
                              </option>
                            </select>
                          </div>
                          <div className="qa-field">
                            <div className="qa-label">Content</div>
                            <select
                              className="qa-select"
                              value={step.content}
                              onChange={(e) =>
                                setGeneratorForm((prev) => ({
                                  ...prev,
                                  steps: prev.steps.map((entry) =>
                                    entry.id !== step.id
                                      ? entry
                                      : {
                                          ...entry,
                                          content: e.target
                                            .value as GeneratorStepContent,
                                          source:
                                            entry.source === 'proximity' &&
                                            e.target.value === 'challenge'
                                              ? 'location_archetype'
                                              : entry.source,
                                        }
                                  ),
                                }))
                              }
                            >
                              <option value="challenge">Challenge</option>
                              <option value="scenario">Scenario</option>
                              <option value="monster">Monster</option>
                            </select>
                          </div>

                          {step.source === 'location_archetype' ? (
                            <div
                              className="qa-field"
                              style={{ gridColumn: '1 / -1' }}
                            >
                              <div className="qa-label">Location Archetype</div>
                              <select
                                className="qa-select"
                                value={step.locationArchetypeId}
                                onChange={(e) =>
                                  setGeneratorForm((prev) => ({
                                    ...prev,
                                    steps: prev.steps.map((entry) =>
                                      entry.id !== step.id
                                        ? entry
                                        : {
                                            ...entry,
                                            locationArchetypeId: e.target.value,
                                          }
                                    ),
                                  }))
                                }
                              >
                                <option value="">
                                  Select a location archetype
                                </option>
                                {locationArchetypes.map((archetype) => (
                                  <option
                                    key={archetype.id}
                                    value={archetype.id}
                                  >
                                    {archetype.name}
                                  </option>
                                ))}
                              </select>
                            </div>
                          ) : (
                            <div className="qa-field">
                              <div className="qa-label">Proximity (m)</div>
                              <input
                                type="number"
                                min={0}
                                className="qa-input"
                                value={step.proximityMeters}
                                onChange={(e) =>
                                  setGeneratorForm((prev) => ({
                                    ...prev,
                                    steps: prev.steps.map((entry) =>
                                      entry.id !== step.id
                                        ? entry
                                        : {
                                            ...entry,
                                            proximityMeters:
                                              parseInt(e.target.value) || 0,
                                          }
                                    ),
                                  }))
                                }
                              />
                            </div>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
                {generatorValidationError && (
                  <div className="qa-helper" style={{ marginTop: 8 }}>
                    {generatorValidationError}
                  </div>
                )}
              </div>

              <div className="qa-footer">
                <button
                  type="button"
                  className="qa-btn qa-btn-outline"
                  onClick={() => {
                    setShouldShowGeneratorModal(false);
                    setGeneratorForm(createEmptyQuestTemplateGeneratorForm());
                  }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="qa-btn qa-btn-primary"
                  disabled={Boolean(generatorValidationError)}
                >
                  Generate
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {shouldShowModal && (
        <div className="qa-modal">
          <div className="qa-modal-card">
            <h2 className="qa-modal-title">Create Quest Archetype</h2>
            <p className="qa-muted" style={{ marginBottom: 16 }}>
              Define the explicit quest copy up front. Side quests can still
              auto-match quest givers by character tags, while main-story quests
              use a fixed quest giver.
            </p>
            <form
              className="qa-form-grid"
              onSubmit={async (e) => {
                e.preventDefault();
                const validationError = validateQuestArchetypeNodeEditor(
                  createForm.rootNode,
                  'Root'
                );
                if (validationError) {
                  window.alert(validationError);
                  return;
                }
                const created = await createQuestArchetype(
                  normalizeQuestArchetypeDraft(createForm)
                );
                if (created?.id) {
                  setSelectedArchetypeId(created.id);
                  setEditingArchetype(created);
                  setEditForm(
                    buildQuestArchetypeFormFromRecord(
                      created,
                      locationArchetypes
                    )
                  );
                }
                setCreateForm(createEmptyQuestArchetypeForm());
                setShouldShowModal(false);
              }}
            >
              <div className="qa-field">
                <div className="qa-label">Name</div>
                <input
                  type="text"
                  className="qa-input"
                  value={createForm.name}
                  onChange={(e) =>
                    setCreateForm((prev) => ({ ...prev, name: e.target.value }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Description</div>
                <textarea
                  className="qa-input"
                  rows={3}
                  value={createForm.description}
                  onChange={(e) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      description: e.target.value,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <DialogueMessageListEditor
                  label="Acceptance Dialogue"
                  helperText="These lines appear before the quest is accepted."
                  value={createForm.acceptanceDialogue}
                  onChange={(acceptanceDialogue) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      acceptanceDialogue,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Image URL</div>
                <input
                  type="text"
                  className="qa-input"
                  value={createForm.imageUrl}
                  onChange={(e) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      imageUrl: e.target.value,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Quest Category</div>
                <select
                  className="qa-select"
                  value={createForm.category}
                  onChange={(e) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      category: e.target.value as 'side' | 'main_story',
                    }))
                  }
                >
                  <option value="side">Side Quest</option>
                  <option value="main_story">Main Story</option>
                </select>
              </div>
              <div className="qa-field">
                <div className="qa-label">
                  {createForm.category === 'main_story'
                    ? 'Quest Giver'
                    : 'Quest Giver Tags'}
                </div>
                {createForm.category === 'main_story' ? (
                  <>
                    <select
                      className="qa-select"
                      value={createForm.questGiverCharacterId}
                      onChange={(e) =>
                        setCreateForm((prev) => ({
                          ...prev,
                          questGiverCharacterId: e.target.value,
                        }))
                      }
                    >
                      <option value="">Select a character</option>
                      {sortedCharacters.map((character) => (
                        <option key={character.id} value={character.id}>
                          {character.name}
                        </option>
                      ))}
                    </select>
                    <div className="qa-helper">
                      {editingQuestGiver
                        ? `Main-story archetypes use a specific quest giver. Current: ${editingQuestGiver.name}`
                        : 'Main-story archetypes use a specific quest giver.'}
                    </div>
                  </>
                ) : (
                  <>
                    <input
                      type="text"
                      className="qa-input"
                      value={createForm.characterTagsText}
                      onChange={(e) =>
                        setCreateForm((prev) => ({
                          ...prev,
                          characterTagsText: e.target.value,
                        }))
                      }
                      placeholder="merchant, scholar, ranger"
                    />
                    <div className="qa-helper">
                      Used to auto-match a character when quests are generated
                      from this archetype.
                    </div>
                  </>
                )}
              </div>
              <div className="qa-field">
                <div className="qa-label">Internal Tags</div>
                <input
                  type="text"
                  className="qa-input"
                  value={createForm.internalTagsText}
                  onChange={(e) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      internalTagsText: e.target.value,
                    }))
                  }
                  placeholder="story_arc, faction, tutorial"
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Required Story Flags</div>
                <input
                  type="text"
                  className="qa-input"
                  value={createForm.requiredStoryFlagsText}
                  onChange={(e) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      requiredStoryFlagsText: e.target.value,
                    }))
                  }
                  placeholder="met_the_warden, chapter_2_started"
                />
                <div className="qa-helper">
                  Players must already have these flags to receive this quest.
                </div>
              </div>
              <div className="qa-field">
                <div className="qa-label">Set Story Flags On Turn-In</div>
                <input
                  type="text"
                  className="qa-input"
                  value={createForm.setStoryFlagsText}
                  onChange={(e) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      setStoryFlagsText: e.target.value,
                    }))
                  }
                  placeholder="warden_warned, chapter_2_complete"
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Clear Story Flags On Turn-In</div>
                <input
                  type="text"
                  className="qa-input"
                  value={createForm.clearStoryFlagsText}
                  onChange={(e) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      clearStoryFlagsText: e.target.value,
                    }))
                  }
                  placeholder="chapter_2_started"
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Quest Giver Relationship Effects</div>
                <div className="qa-inline-grid">
                  <input
                    type="number"
                    className="qa-input"
                    min={-3}
                    max={3}
                    value={createForm.relationshipTrust}
                    onChange={(e) =>
                      setCreateForm((prev) => ({
                        ...prev,
                        relationshipTrust: Number(e.target.value) || 0,
                      }))
                    }
                    placeholder="Trust"
                  />
                  <input
                    type="number"
                    className="qa-input"
                    min={-3}
                    max={3}
                    value={createForm.relationshipRespect}
                    onChange={(e) =>
                      setCreateForm((prev) => ({
                        ...prev,
                        relationshipRespect: Number(e.target.value) || 0,
                      }))
                    }
                    placeholder="Respect"
                  />
                  <input
                    type="number"
                    className="qa-input"
                    min={-3}
                    max={3}
                    value={createForm.relationshipFear}
                    onChange={(e) =>
                      setCreateForm((prev) => ({
                        ...prev,
                        relationshipFear: Number(e.target.value) || 0,
                      }))
                    }
                    placeholder="Fear"
                  />
                  <input
                    type="number"
                    className="qa-input"
                    min={-3}
                    max={3}
                    value={createForm.relationshipDebt}
                    onChange={(e) =>
                      setCreateForm((prev) => ({
                        ...prev,
                        relationshipDebt: Number(e.target.value) || 0,
                      }))
                    }
                    placeholder="Debt"
                  />
                </div>
                <div className="qa-helper">
                  Applied to the player's relationship with the quest giver when
                  the quest is turned in.
                </div>
              </div>
              <div
                className="qa-field"
                style={{ gridColumn: '1 / -1', marginTop: 8 }}
              >
                <div className="qa-label">Root Node</div>
                <div className="qa-helper" style={{ marginBottom: 12 }}>
                  Configure the first beat using the same shared node editor
                  used throughout the quest flow.
                </div>
                <div className="qa-form-grid">
                  <QuestArchetypeNodeConfigFields
                    editor={createForm.rootNode}
                    setEditor={(updater) =>
                      setCreateForm((prev) => ({
                        ...prev,
                        rootNode:
                          typeof updater === 'function'
                            ? updater(prev.rootNode)
                            : updater,
                      }))
                    }
                    prefix="Root"
                    locationArchetypes={locationArchetypes}
                    challengeTemplates={challengeTemplates}
                    monsterTemplates={monsterTemplates}
                    scenarioTemplates={scenarioTemplates}
                    characters={sortedCharacters}
                    inventoryItems={inventoryItems}
                    spells={spells}
                  />
                </div>
              </div>
              <div className="qa-field">
                <div className="qa-label">Reward Mode</div>
                <select
                  className="qa-select"
                  value={createForm.rewardMode}
                  onChange={(e) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      rewardMode: e.target.value as RewardMode,
                    }))
                  }
                >
                  <option value="random">Random</option>
                  <option value="explicit">Explicit</option>
                </select>
              </div>
              <div className="qa-field">
                <div className="qa-label">Random Reward Size</div>
                <select
                  className="qa-select"
                  value={createForm.randomRewardSize}
                  disabled={createForm.rewardMode !== 'random'}
                  onChange={(e) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      randomRewardSize: e.target.value as RandomRewardSize,
                    }))
                  }
                >
                  <option value="small">Small</option>
                  <option value="medium">Medium</option>
                  <option value="large">Large</option>
                </select>
              </div>
              <div className="qa-field">
                <div className="qa-label">Default Gold</div>
                <input
                  type="number"
                  min={0}
                  className="qa-input"
                  value={createForm.defaultGold}
                  onChange={(e) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      defaultGold: parseInt(e.target.value) || 0,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Reward Experience</div>
                <input
                  type="number"
                  min={0}
                  className="qa-input"
                  disabled={createForm.rewardMode !== 'explicit'}
                  value={createForm.rewardExperience}
                  onChange={(e) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      rewardExperience: parseInt(e.target.value) || 0,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Recurrence</div>
                <select
                  className="qa-select"
                  value={createForm.recurrenceFrequency}
                  onChange={(e) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      recurrenceFrequency: e.target.value,
                    }))
                  }
                >
                  <option value="">None</option>
                  <option value="daily">Daily</option>
                  <option value="weekly">Weekly</option>
                  <option value="monthly">Monthly</option>
                </select>
              </div>
              <div className="qa-field">
                <div className="qa-label">Difficulty Mode</div>
                <select
                  className="qa-select"
                  value={createForm.difficultyMode}
                  onChange={(e) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      difficultyMode: e.target.value as QuestDifficultyMode,
                    }))
                  }
                >
                  <option value="scale">Scale With User Level</option>
                  <option value="fixed">Fixed Difficulty</option>
                </select>
              </div>
              <div className="qa-field">
                <div className="qa-label">Difficulty</div>
                <input
                  type="number"
                  min={1}
                  className="qa-input"
                  disabled={createForm.difficultyMode !== 'fixed'}
                  value={createForm.difficulty}
                  onChange={(e) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      difficulty: Math.max(1, parseInt(e.target.value) || 1),
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Monster Encounter Target Level</div>
                <input
                  type="number"
                  min={1}
                  className="qa-input"
                  value={createForm.monsterEncounterTargetLevel}
                  onChange={(e) =>
                    setCreateForm((prev) => ({
                      ...prev,
                      monsterEncounterTargetLevel: Math.max(
                        1,
                        parseInt(e.target.value) || 1
                      ),
                    }))
                  }
                />
                <div className="qa-helper">
                  Used for all monster encounter child nodes in this quest
                  template.
                </div>
              </div>
              <div className="qa-field">
                <div className="qa-label">Material Rewards</div>
                <MaterialRewardsEditor
                  value={createForm.materialRewards}
                  onChange={(materialRewards) =>
                    setCreateForm((prev) => ({ ...prev, materialRewards }))
                  }
                  disabled={createForm.rewardMode !== 'explicit'}
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Item Rewards</div>
                {createForm.itemRewards.length === 0 ? (
                  <div className="qa-empty">No item rewards yet.</div>
                ) : (
                  <div className="qa-form-grid">
                    {createForm.itemRewards.map((reward, index) => (
                      <div
                        key={`create-reward-${index}`}
                        className="qa-reward-row"
                      >
                        <select
                          className="qa-select"
                          value={reward.inventoryItemId}
                          disabled={createForm.rewardMode !== 'explicit'}
                          onChange={(e) =>
                            setCreateForm((prev) => ({
                              ...prev,
                              itemRewards: prev.itemRewards.map(
                                (entry, rewardIndex) =>
                                  rewardIndex === index
                                    ? {
                                        ...entry,
                                        inventoryItemId: e.target.value,
                                      }
                                    : entry
                              ),
                            }))
                          }
                        >
                          <option value="">Select an item</option>
                          {inventoryItems.map((item) => (
                            <option key={item.id} value={item.id}>
                              {item.name}
                            </option>
                          ))}
                        </select>
                        <input
                          type="number"
                          min={1}
                          className="qa-input"
                          disabled={createForm.rewardMode !== 'explicit'}
                          value={reward.quantity}
                          onChange={(e) =>
                            setCreateForm((prev) => ({
                              ...prev,
                              itemRewards: prev.itemRewards.map(
                                (entry, rewardIndex) =>
                                  rewardIndex === index
                                    ? {
                                        ...entry,
                                        quantity: parseInt(e.target.value) || 1,
                                      }
                                    : entry
                              ),
                            }))
                          }
                        />
                        <button
                          type="button"
                          className="qa-btn qa-btn-text"
                          onClick={() =>
                            setCreateForm((prev) => ({
                              ...prev,
                              itemRewards: prev.itemRewards.filter(
                                (_, rewardIndex) => rewardIndex !== index
                              ),
                            }))
                          }
                        >
                          Remove
                        </button>
                      </div>
                    ))}
                  </div>
                )}
                <button
                  type="button"
                  className="qa-btn qa-btn-ghost"
                  onClick={() =>
                    setCreateForm((prev) => ({
                      ...prev,
                      itemRewards: [
                        ...prev.itemRewards,
                        { inventoryItemId: '', quantity: 1 },
                      ],
                    }))
                  }
                >
                  Add Item Reward
                </button>
              </div>
              <div className="qa-field">
                <div className="qa-label">Spell Rewards</div>
                {createForm.spellRewards.length === 0 ? (
                  <div className="qa-empty">No spell rewards yet.</div>
                ) : (
                  <div className="qa-form-grid">
                    {createForm.spellRewards.map((reward, index) => (
                      <div
                        key={`create-spell-${index}`}
                        className="qa-reward-row"
                      >
                        <select
                          className="qa-select"
                          value={reward.spellId}
                          disabled={createForm.rewardMode !== 'explicit'}
                          onChange={(e) =>
                            setCreateForm((prev) => ({
                              ...prev,
                              spellRewards: prev.spellRewards.map(
                                (entry, rewardIndex) =>
                                  rewardIndex === index
                                    ? { ...entry, spellId: e.target.value }
                                    : entry
                              ),
                            }))
                          }
                        >
                          <option value="">Select a spell</option>
                          {spells.map((spell) => (
                            <option key={spell.id} value={spell.id}>
                              {spell.name}
                            </option>
                          ))}
                        </select>
                        <button
                          type="button"
                          className="qa-btn qa-btn-text"
                          onClick={() =>
                            setCreateForm((prev) => ({
                              ...prev,
                              spellRewards: prev.spellRewards.filter(
                                (_, rewardIndex) => rewardIndex !== index
                              ),
                            }))
                          }
                        >
                          Remove
                        </button>
                      </div>
                    ))}
                  </div>
                )}
                <button
                  type="button"
                  className="qa-btn qa-btn-ghost"
                  onClick={() =>
                    setCreateForm((prev) => ({
                      ...prev,
                      spellRewards: [...prev.spellRewards, { spellId: '' }],
                    }))
                  }
                >
                  Add Spell Reward
                </button>
              </div>
              <div className="qa-footer">
                <button
                  type="button"
                  className="qa-btn qa-btn-outline"
                  onClick={() => {
                    setShouldShowModal(false);
                    setCreateForm(createEmptyQuestArchetypeForm());
                  }}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="qa-btn qa-btn-primary"
                  disabled={
                    !questArchetypeFormHasExplicitCopy(createForm) ||
                    (createForm.category === 'main_story' &&
                      !createForm.questGiverCharacterId)
                  }
                >
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {editingArchetype && (
        <div className="qa-modal">
          <div className="qa-modal-card">
            <h2 className="qa-modal-title">Edit Quest Template</h2>
            <div className="qa-form-grid">
              <div className="qa-field">
                <div className="qa-label">Name</div>
                <input
                  type="text"
                  className="qa-input"
                  value={editForm.name}
                  onChange={(e) =>
                    setEditForm((prev) => ({ ...prev, name: e.target.value }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Description</div>
                <textarea
                  className="qa-input"
                  rows={3}
                  value={editForm.description}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      description: e.target.value,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <DialogueMessageListEditor
                  label="Acceptance Dialogue"
                  helperText="These lines appear before the quest is accepted."
                  value={editForm.acceptanceDialogue}
                  onChange={(acceptanceDialogue) =>
                    setEditForm((prev) => ({
                      ...prev,
                      acceptanceDialogue,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Image URL</div>
                <input
                  type="text"
                  className="qa-input"
                  value={editForm.imageUrl}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      imageUrl: e.target.value,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Quest Category</div>
                <select
                  className="qa-select"
                  value={editForm.category}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      category: e.target.value as 'side' | 'main_story',
                    }))
                  }
                >
                  <option value="side">Side Quest</option>
                  <option value="main_story">Main Story</option>
                </select>
              </div>
              <div className="qa-field">
                <div className="qa-label">
                  {editForm.category === 'main_story'
                    ? 'Quest Giver'
                    : 'Quest Giver Tags'}
                </div>
                {editForm.category === 'main_story' ? (
                  <>
                    <select
                      className="qa-select"
                      value={editForm.questGiverCharacterId}
                      onChange={(e) =>
                        setEditForm((prev) => ({
                          ...prev,
                          questGiverCharacterId: e.target.value,
                        }))
                      }
                    >
                      <option value="">Select a character</option>
                      {sortedCharacters.map((character) => (
                        <option key={character.id} value={character.id}>
                          {character.name}
                        </option>
                      ))}
                    </select>
                    <div className="qa-helper">
                      Main-story archetypes use a specific quest giver.
                    </div>
                  </>
                ) : (
                  <>
                    <input
                      type="text"
                      className="qa-input"
                      value={editForm.characterTagsText}
                      onChange={(e) =>
                        setEditForm((prev) => ({
                          ...prev,
                          characterTagsText: e.target.value,
                        }))
                      }
                      placeholder="merchant, scholar, ranger"
                    />
                    <div className="qa-helper">
                      Used to auto-match a character when quests are generated
                      from this archetype.
                    </div>
                  </>
                )}
              </div>
              <div className="qa-field">
                <div className="qa-label">Internal Tags</div>
                <input
                  type="text"
                  className="qa-input"
                  value={editForm.internalTagsText}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      internalTagsText: e.target.value,
                    }))
                  }
                  placeholder="story_arc, faction, tutorial"
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Required Story Flags</div>
                <input
                  type="text"
                  className="qa-input"
                  value={editForm.requiredStoryFlagsText}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      requiredStoryFlagsText: e.target.value,
                    }))
                  }
                  placeholder="met_the_warden, chapter_2_started"
                />
                <div className="qa-helper">
                  Players must already have these flags to receive this quest.
                </div>
              </div>
              <div className="qa-field">
                <div className="qa-label">Set Story Flags On Turn-In</div>
                <input
                  type="text"
                  className="qa-input"
                  value={editForm.setStoryFlagsText}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      setStoryFlagsText: e.target.value,
                    }))
                  }
                  placeholder="warden_warned, chapter_2_complete"
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Clear Story Flags On Turn-In</div>
                <input
                  type="text"
                  className="qa-input"
                  value={editForm.clearStoryFlagsText}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      clearStoryFlagsText: e.target.value,
                    }))
                  }
                  placeholder="chapter_2_started"
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Quest Giver Relationship Effects</div>
                <div className="qa-inline-grid">
                  <input
                    type="number"
                    className="qa-input"
                    min={-3}
                    max={3}
                    value={editForm.relationshipTrust}
                    onChange={(e) =>
                      setEditForm((prev) => ({
                        ...prev,
                        relationshipTrust: Number(e.target.value) || 0,
                      }))
                    }
                    placeholder="Trust"
                  />
                  <input
                    type="number"
                    className="qa-input"
                    min={-3}
                    max={3}
                    value={editForm.relationshipRespect}
                    onChange={(e) =>
                      setEditForm((prev) => ({
                        ...prev,
                        relationshipRespect: Number(e.target.value) || 0,
                      }))
                    }
                    placeholder="Respect"
                  />
                  <input
                    type="number"
                    className="qa-input"
                    min={-3}
                    max={3}
                    value={editForm.relationshipFear}
                    onChange={(e) =>
                      setEditForm((prev) => ({
                        ...prev,
                        relationshipFear: Number(e.target.value) || 0,
                      }))
                    }
                    placeholder="Fear"
                  />
                  <input
                    type="number"
                    className="qa-input"
                    min={-3}
                    max={3}
                    value={editForm.relationshipDebt}
                    onChange={(e) =>
                      setEditForm((prev) => ({
                        ...prev,
                        relationshipDebt: Number(e.target.value) || 0,
                      }))
                    }
                    placeholder="Debt"
                  />
                </div>
                <div className="qa-helper">
                  Applied to the player's relationship with the quest giver when
                  the quest is turned in.
                </div>
              </div>
              <div className="qa-field">
                <div className="qa-label">Reward Mode</div>
                <select
                  className="qa-select"
                  value={editForm.rewardMode}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      rewardMode: e.target.value as RewardMode,
                    }))
                  }
                >
                  <option value="random">Random</option>
                  <option value="explicit">Explicit</option>
                </select>
              </div>
              <div className="qa-field">
                <div className="qa-label">Random Reward Size</div>
                <select
                  className="qa-select"
                  value={editForm.randomRewardSize}
                  disabled={editForm.rewardMode !== 'random'}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      randomRewardSize: e.target.value as RandomRewardSize,
                    }))
                  }
                >
                  <option value="small">Small</option>
                  <option value="medium">Medium</option>
                  <option value="large">Large</option>
                </select>
              </div>
              <div className="qa-field">
                <div className="qa-label">Default Gold</div>
                <input
                  type="number"
                  min={0}
                  className="qa-input"
                  value={editForm.defaultGold}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      defaultGold: parseInt(e.target.value) || 0,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Reward Experience</div>
                <input
                  type="number"
                  min={0}
                  className="qa-input"
                  disabled={editForm.rewardMode !== 'explicit'}
                  value={editForm.rewardExperience}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      rewardExperience: parseInt(e.target.value) || 0,
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Recurrence</div>
                <select
                  className="qa-select"
                  value={editForm.recurrenceFrequency}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      recurrenceFrequency: e.target.value,
                    }))
                  }
                >
                  <option value="">None</option>
                  <option value="daily">Daily</option>
                  <option value="weekly">Weekly</option>
                  <option value="monthly">Monthly</option>
                </select>
              </div>
              <div className="qa-field">
                <div className="qa-label">Difficulty Mode</div>
                <select
                  className="qa-select"
                  value={editForm.difficultyMode}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      difficultyMode: e.target.value as QuestDifficultyMode,
                    }))
                  }
                >
                  <option value="scale">Scale With User Level</option>
                  <option value="fixed">Fixed Difficulty</option>
                </select>
              </div>
              <div className="qa-field">
                <div className="qa-label">Difficulty</div>
                <input
                  type="number"
                  min={1}
                  className="qa-input"
                  disabled={editForm.difficultyMode !== 'fixed'}
                  value={editForm.difficulty}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      difficulty: Math.max(1, parseInt(e.target.value) || 1),
                    }))
                  }
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Monster Encounter Target Level</div>
                <input
                  type="number"
                  min={1}
                  className="qa-input"
                  value={editForm.monsterEncounterTargetLevel}
                  onChange={(e) =>
                    setEditForm((prev) => ({
                      ...prev,
                      monsterEncounterTargetLevel: Math.max(
                        1,
                        parseInt(e.target.value) || 1
                      ),
                    }))
                  }
                />
                <div className="qa-helper">
                  Used for all monster encounter child nodes in this quest
                  template.
                </div>
              </div>
              <div className="qa-field">
                <div className="qa-label">Material Rewards</div>
                <MaterialRewardsEditor
                  value={editForm.materialRewards}
                  onChange={(materialRewards) =>
                    setEditForm((prev) => ({ ...prev, materialRewards }))
                  }
                  disabled={editForm.rewardMode !== 'explicit'}
                />
              </div>
              <div className="qa-field">
                <div className="qa-label">Item Rewards</div>
                {editForm.itemRewards.length === 0 ? (
                  <div className="qa-empty">No item rewards yet.</div>
                ) : (
                  <div className="qa-form-grid">
                    {editForm.itemRewards.map((reward, index) => (
                      <div
                        key={`edit-reward-${index}`}
                        className="qa-reward-row"
                      >
                        <select
                          className="qa-select"
                          value={reward.inventoryItemId}
                          disabled={editForm.rewardMode !== 'explicit'}
                          onChange={(e) =>
                            setEditForm((prev) => ({
                              ...prev,
                              itemRewards: prev.itemRewards.map(
                                (entry, rewardIndex) =>
                                  rewardIndex === index
                                    ? {
                                        ...entry,
                                        inventoryItemId: e.target.value,
                                      }
                                    : entry
                              ),
                            }))
                          }
                        >
                          <option value="">Select an item</option>
                          {inventoryItems.map((item) => (
                            <option key={item.id} value={item.id}>
                              {item.name}
                            </option>
                          ))}
                        </select>
                        <input
                          type="number"
                          min={1}
                          className="qa-input"
                          disabled={editForm.rewardMode !== 'explicit'}
                          value={reward.quantity}
                          onChange={(e) =>
                            setEditForm((prev) => ({
                              ...prev,
                              itemRewards: prev.itemRewards.map(
                                (entry, rewardIndex) =>
                                  rewardIndex === index
                                    ? {
                                        ...entry,
                                        quantity: parseInt(e.target.value) || 1,
                                      }
                                    : entry
                              ),
                            }))
                          }
                        />
                        <button
                          type="button"
                          className="qa-btn qa-btn-text"
                          onClick={() =>
                            setEditForm((prev) => ({
                              ...prev,
                              itemRewards: prev.itemRewards.filter(
                                (_, rewardIndex) => rewardIndex !== index
                              ),
                            }))
                          }
                        >
                          Remove
                        </button>
                      </div>
                    ))}
                  </div>
                )}
                <button
                  type="button"
                  className="qa-btn qa-btn-ghost"
                  onClick={() =>
                    setEditForm((prev) => ({
                      ...prev,
                      itemRewards: [
                        ...prev.itemRewards,
                        { inventoryItemId: '', quantity: 1 },
                      ],
                    }))
                  }
                >
                  Add Item Reward
                </button>
              </div>
              <div className="qa-field">
                <div className="qa-label">Spell Rewards</div>
                {editForm.spellRewards.length === 0 ? (
                  <div className="qa-empty">No spell rewards yet.</div>
                ) : (
                  <div className="qa-form-grid">
                    {editForm.spellRewards.map((reward, index) => (
                      <div
                        key={`edit-spell-${index}`}
                        className="qa-reward-row"
                      >
                        <select
                          className="qa-select"
                          value={reward.spellId}
                          disabled={editForm.rewardMode !== 'explicit'}
                          onChange={(e) =>
                            setEditForm((prev) => ({
                              ...prev,
                              spellRewards: prev.spellRewards.map(
                                (entry, rewardIndex) =>
                                  rewardIndex === index
                                    ? { ...entry, spellId: e.target.value }
                                    : entry
                              ),
                            }))
                          }
                        >
                          <option value="">Select a spell</option>
                          {spells.map((spell) => (
                            <option key={spell.id} value={spell.id}>
                              {spell.name}
                            </option>
                          ))}
                        </select>
                        <button
                          type="button"
                          className="qa-btn qa-btn-text"
                          onClick={() =>
                            setEditForm((prev) => ({
                              ...prev,
                              spellRewards: prev.spellRewards.filter(
                                (_, rewardIndex) => rewardIndex !== index
                              ),
                            }))
                          }
                        >
                          Remove
                        </button>
                      </div>
                    ))}
                  </div>
                )}
                <button
                  type="button"
                  className="qa-btn qa-btn-ghost"
                  onClick={() =>
                    setEditForm((prev) => ({
                      ...prev,
                      spellRewards: [...prev.spellRewards, { spellId: '' }],
                    }))
                  }
                >
                  Add Spell Reward
                </button>
              </div>
            </div>
            <div className="qa-footer">
              <button
                className="qa-btn qa-btn-outline"
                onClick={() => {
                  setEditingArchetype(null);
                  setEditForm(createEmptyQuestArchetypeForm());
                }}
              >
                Cancel
              </button>
              <button
                className="qa-btn qa-btn-primary"
                disabled={
                  !questArchetypeFormHasExplicitCopy(editForm) ||
                  (editForm.category === 'main_story' &&
                    !editForm.questGiverCharacterId)
                }
                onClick={async () => {
                  if (!editingArchetype) return;
                  const draft = normalizeQuestArchetypeDraft(editForm);
                  await updateQuestArchetype({
                    ...editingArchetype,
                    name: draft.name,
                    description: draft.description,
                    category: draft.category,
                    questGiverCharacterId: draft.questGiverCharacterId,
                    acceptanceDialogue: draft.acceptanceDialogue,
                    imageUrl: draft.imageUrl,
                    difficultyMode: draft.difficultyMode,
                    difficulty: draft.difficulty,
                    monsterEncounterTargetLevel:
                      draft.monsterEncounterTargetLevel,
                    defaultGold: draft.defaultGold ?? 0,
                    rewardMode: draft.rewardMode,
                    randomRewardSize: draft.randomRewardSize,
                    rewardExperience: draft.rewardExperience ?? 0,
                    recurrenceFrequency: draft.recurrenceFrequency ?? null,
                    materialRewards: draft.materialRewards,
                    requiredStoryFlags: draft.requiredStoryFlags,
                    setStoryFlags: draft.setStoryFlags,
                    clearStoryFlags: draft.clearStoryFlags,
                    questGiverRelationshipEffects:
                      draft.questGiverRelationshipEffects,
                    itemRewards: draft.itemRewards,
                    spellRewards: draft.spellRewards,
                    characterTags: draft.characterTags,
                    internalTags: draft.internalTags,
                  });
                  setEditingArchetype(null);
                  setEditForm(createEmptyQuestArchetypeForm());
                }}
              >
                Save
              </button>
            </div>
          </div>
        </div>
      )}

      {editingChallenge && (
        <div className="qa-modal">
          <div className="qa-modal-card">
            <h2 className="qa-modal-title">Edit Challenge</h2>
            <div className="qa-form-grid">
              {editingChallengeAllowsTemplate && (
                <div className="qa-field">
                  <div className="qa-label">Challenge Template</div>
                  <select
                    className="qa-select"
                    value={editChallengeTemplateId}
                    onChange={(e) => setEditChallengeTemplateId(e.target.value)}
                  >
                    <option value="">Select a challenge template</option>
                    {challengeTemplates.map((template) => (
                      <option key={template.id} value={template.id}>
                        {describeChallengeTemplate(
                          template,
                          locationArchetypes
                        )}
                      </option>
                    ))}
                  </select>
                </div>
              )}
              <div className="qa-field">
                <div className="qa-label">Proficiency</div>
                <input
                  type="text"
                  className="qa-input"
                  value={editChallengeProficiency}
                  disabled={
                    editingChallengeAllowsTemplate &&
                    Boolean(editChallengeTemplateId)
                  }
                  onChange={(e) => {
                    setEditChallengeProficiency(e.target.value);
                    setProficiencySearch(e.target.value);
                  }}
                  list="qa-proficiency-options"
                  placeholder="Optional proficiency (e.g. Persuasion)"
                />
              </div>
            </div>
            <div className="qa-footer">
              <button
                className="qa-btn qa-btn-outline"
                onClick={() => {
                  setEditingChallenge(null);
                  setEditingChallengeAllowsTemplate(false);
                }}
              >
                Cancel
              </button>
              <button
                className="qa-btn qa-btn-danger"
                onClick={async () => {
                  if (!editingChallenge) return;
                  const confirmDelete = window.confirm(
                    'Delete this challenge? This cannot be undone.'
                  );
                  if (!confirmDelete) return;
                  await deleteQuestArchetypeChallenge(editingChallenge.id);
                  setEditingChallenge(null);
                  setEditingChallengeAllowsTemplate(false);
                }}
              >
                Delete
              </button>
              <button
                className="qa-btn qa-btn-primary"
                onClick={async () => {
                  const trimmed = editChallengeProficiency.trim();
                  await updateQuestArchetypeChallenge(editingChallenge.id, {
                    challengeTemplateId: editingChallengeAllowsTemplate
                      ? editChallengeTemplateId || null
                      : undefined,
                    proficiency: trimmed.length > 0 ? trimmed : null,
                  });
                  setEditingChallenge(null);
                  setEditingChallengeAllowsTemplate(false);
                }}
              >
                Save
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
