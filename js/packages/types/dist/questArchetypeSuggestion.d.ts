import { QuestDifficultyMode } from './questDifficulty';
import { QuestNodeSubmissionType } from './questNode';
import { QuestArchetype } from './questArchetype';
export interface QuestArchetypeSuggestionStep {
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
export interface QuestArchetypeSuggestionJob {
    id: string;
    createdAt: string;
    updatedAt: string;
    status: string;
    count: number;
    themePrompt: string;
    familyTags: string[];
    characterTags: string[];
    internalTags: string[];
    requiredLocationArchetypeIds: string[];
    requiredLocationMetadataTags: string[];
    createdCount: number;
    errorMessage?: string | null;
}
export interface QuestArchetypeSuggestionDraft {
    id: string;
    createdAt: string;
    updatedAt: string;
    jobId: string;
    status: string;
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
    steps: QuestArchetypeSuggestionStep[];
    challengeTemplateSeeds: string[];
    scenarioTemplateSeeds: string[];
    monsterTemplateSeeds: string[];
    warnings: string[];
    questArchetypeId?: string | null;
    questArchetype?: QuestArchetype | null;
    convertedAt?: string | null;
}
