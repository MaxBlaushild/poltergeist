export type QuestNodeSubmissionType = 'text' | 'photo' | 'video' | (string & {});
export type QuestNodeObjectiveType = 'challenge' | 'scenario' | 'monster_encounter' | 'monster' | (string & {});
export interface QuestNodeObjective {
    id: string;
    type: QuestNodeObjectiveType;
    prompt: string;
    description?: string;
    imageUrl?: string;
    thumbnailUrl?: string;
    reward?: number;
    inventoryItemId?: number | null;
    submissionType?: QuestNodeSubmissionType;
    difficulty?: number;
    statTags?: string[];
    proficiency?: string | null;
}
export interface QuestNode {
    id: string;
    questId: string;
    orderIndex: number;
    submissionType?: QuestNodeSubmissionType;
    objectiveText?: string;
    objective?: QuestNodeObjective | null;
    pointOfInterestId?: string | null;
    scenarioId?: string | null;
    monsterId?: string | null;
    monsterEncounterId?: string | null;
    challengeId?: string | null;
    polygon?: string | null;
    polygonPoints?: [number, number][];
}
