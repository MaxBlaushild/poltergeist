import { QuestDifficultyMode } from './questDifficulty';
import { QuestNodeSubmissionType } from './questNode';

export interface MainStoryBeatStep {
  source: 'location' | 'proximity' | (string & {});
  content: 'challenge' | 'scenario' | 'monster' | (string & {});
  locationConcept: string;
  locationArchetypeName?: string;
  locationArchetypeId?: string | null;
  locationMetadataTags: string[];
  distanceMeters?: number | null;
  templateConcept: string;
  potentialContent: string[];
  challengeQuestion?: string;
  challengeDescription?: string;
  challengeSubmissionType?: QuestNodeSubmissionType;
  challengeProficiency?: string | null;
  challengeStatTags?: string[];
  scenarioPrompt?: string;
  scenarioOpenEnded?: boolean;
  scenarioBeats?: string[];
  monsterTemplateNames?: string[];
  monsterTemplateIds?: string[];
  encounterTone?: string[];
}

export interface MainStoryBeatDraft {
  orderIndex: number;
  act: number;
  storyRole: string;
  chapterTitle: string;
  chapterSummary: string;
  purpose: string;
  whatChanges: string;
  introducedCharacterKeys: string[];
  requiredCharacterKeys: string[];
  introducedRevealKeys: string[];
  requiredRevealKeys: string[];
  requiredZoneTags: string[];
  requiredLocationArchetypeIds: string[];
  preferredContentMix: string[];
  questGiverCharacterKey: string;
  questGiverCharacterId?: string | null;
  questGiverCharacterName?: string;
  name: string;
  hook: string;
  description: string;
  acceptanceDialogue: string[];
  characterTags: string[];
  internalTags: string[];
  difficultyMode: QuestDifficultyMode;
  difficulty: number;
  monsterEncounterTargetLevel: number;
  whyThisScales: string;
  steps: MainStoryBeatStep[];
  challengeTemplateSeeds: string[];
  scenarioTemplateSeeds: string[];
  monsterTemplateSeeds: string[];
  warnings: string[];
  questArchetypeId?: string | null;
  questArchetypeName?: string;
}

export interface MainStorySuggestionJob {
  id: string;
  createdAt: string;
  updatedAt: string;
  status: string;
  count: number;
  questCount: number;
  themePrompt: string;
  districtFit: string;
  tone: string;
  familyTags: string[];
  characterTags: string[];
  internalTags: string[];
  requiredLocationArchetypeIds: string[];
  requiredLocationMetadataTags: string[];
  createdCount: number;
  errorMessage?: string | null;
}

export interface MainStoryTemplate {
  id: string;
  createdAt: string;
  updatedAt: string;
  name: string;
  premise: string;
  districtFit: string;
  tone: string;
  themeTags: string[];
  internalTags: string[];
  factionKeys: string[];
  characterKeys: string[];
  revealKeys: string[];
  climaxSummary: string;
  resolutionSummary: string;
  whyItWorks: string;
  beats: MainStoryBeatDraft[];
}

export interface MainStorySuggestionDraft {
  id: string;
  createdAt: string;
  updatedAt: string;
  jobId: string;
  status: string;
  name: string;
  premise: string;
  districtFit: string;
  tone: string;
  themeTags: string[];
  internalTags: string[];
  factionKeys: string[];
  characterKeys: string[];
  revealKeys: string[];
  climaxSummary: string;
  resolutionSummary: string;
  whyItWorks: string;
  beats: MainStoryBeatDraft[];
  warnings: string[];
  mainStoryTemplateId?: string | null;
  mainStoryTemplate?: MainStoryTemplate | null;
  convertedAt?: string | null;
}
