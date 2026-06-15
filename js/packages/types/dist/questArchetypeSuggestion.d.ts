import { QuestDifficultyMode } from './questDifficulty';
import { QuestNodeSubmissionType } from './questNode';
import { QuestArchetype } from './questArchetype';
export interface QuestArchetypeSuggestionStep {
    source: 'location' | 'proximity' | (string & {});
    content: 'challenge' | 'scenario' | 'monster' | 'exposition' | (string & {});
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
    expositionTitle?: string;
    expositionDescription?: string;
    expositionSpeakerName?: string;
    expositionPortraitUrl?: string;
    expositionDialogue?: string[];
    monsterTemplateNames?: string[];
    monsterTemplateIds?: string[];
    encounterTone?: string[];
}
export interface QuestArchetypeSuggestionNodeOutcome {
    outcome: 'success' | 'failure' | (string & {});
    nextNodeKey?: string;
}
export interface QuestArchetypeSuggestionNode extends QuestArchetypeSuggestionStep {
    nodeKey: string;
    outcomes?: QuestArchetypeSuggestionNodeOutcome[];
}
export interface QuestArchetypeSuggestionJob {
    id: string;
    createdAt: string;
    updatedAt: string;
    status: string;
    count: number;
    yeetIt: boolean;
    zoneKind: string;
    themePrompt: string;
    familyTags: string[];
    familyMixTargets?: Record<string, number>;
    characterTags: string[];
    internalTags: string[];
    requiredLocationArchetypeIds: string[];
    requiredLocationMetadataTags: string[];
    createdCount: number;
    errorMessage?: string | null;
}
export interface QuestArchetypeSuggestionPreset {
    count: number;
    zoneKind: string;
    themePrompt: string;
    familyTags: string[];
    familyMixTargets: Record<string, number>;
    characterTags: string[];
    internalTags: string[];
    requiredLocationArchetypeIds: string[];
    requiredLocationMetadataTags: string[];
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
    zoneKind: string;
    acceptanceDialogue: string[];
    characterTags: string[];
    internalTags: string[];
    difficultyMode: QuestDifficultyMode;
    difficulty: number;
    monsterEncounterTargetLevel: number;
    whyThisScales: string;
    steps: QuestArchetypeSuggestionStep[];
    nodes?: QuestArchetypeSuggestionNode[];
    challengeTemplateSeeds: string[];
    scenarioTemplateSeeds: string[];
    monsterTemplateSeeds: string[];
    warnings: string[];
    questArchetypeId?: string | null;
    questArchetype?: QuestArchetype | null;
    convertedAt?: string | null;
}
