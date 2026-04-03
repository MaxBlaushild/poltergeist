import { QuestDifficultyMode } from './questDifficulty';
import { QuestNodeSubmissionType } from './questNode';
import { CharacterRelationshipState } from './quest';

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
  requiredStoryFlags: string[];
  setStoryFlags: string[];
  clearStoryFlags: string[];
  questGiverRelationshipEffects?: CharacterRelationshipState;
  worldChanges?: MainStoryWorldChange[];
  unlockedScenarios?: MainStoryUnlockedScenario[];
  unlockedChallenges?: MainStoryUnlockedChallenge[];
  unlockedMonsterEncounters?: MainStoryUnlockedEncounter[];
  questGiverAfterDescription?: string;
  questGiverAfterDialogue: string[];
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

export interface MainStoryWorldChange {
  type: 'move_character' | 'show_poi_text' | (string & {});
  targetKey: string;
  characterKey?: string;
  pointOfInterestHint?: string;
  destinationHint?: string;
  zoneTags?: string[];
  description?: string;
  clue?: string;
}

export interface MainStoryUnlockedScenario {
  name: string;
  prompt: string;
  pointOfInterestHint?: string;
  internalTags?: string[];
  difficulty?: number;
}

export interface MainStoryUnlockedChallenge {
  question: string;
  description: string;
  pointOfInterestHint?: string;
  submissionType?: QuestNodeSubmissionType;
  proficiency?: string | null;
  statTags?: string[];
  difficulty?: number;
}

export interface MainStoryUnlockedEncounter {
  name: string;
  description: string;
  pointOfInterestHint?: string;
  encounterType?: 'monster' | 'boss' | 'raid' | (string & {});
  monsterCount?: number;
  encounterTone?: string[];
  monsterTemplateHints?: string[];
  targetLevel?: number;
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

export interface MainStoryDistrictBeatRun {
  orderIndex: number;
  act: number;
  chapterTitle: string;
  storyRole: string;
  status: string;
  zoneId?: string | null;
  zoneName?: string;
  pointOfInterestId?: string | null;
  pointOfInterestName?: string;
  questId?: string | null;
  questName?: string;
  questArchetypeId?: string | null;
  questArchetypeName?: string;
  questGiverCharacterId?: string | null;
  questGiverCharacterName?: string;
  errorMessage?: string;
}

export interface MainStoryDistrictRun {
  id: string;
  createdAt: string;
  updatedAt: string;
  mainStoryTemplateId: string;
  districtId: string;
  status: string;
  beatRuns: MainStoryDistrictBeatRun[];
  generatedCharacterIds: string[];
  errorMessage?: string | null;
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
