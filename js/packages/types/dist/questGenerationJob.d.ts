import { Quest } from "./quest";
export type QuestGenerationJob = {
    id: string;
    createdAt: string;
    updatedAt: string;
    zoneQuestArchetypeId?: string | null;
    zoneId: string;
    questArchetypeId: string;
    questGiverCharacterId?: string | null;
    status: string;
    totalCount: number;
    completedCount: number;
    failedCount: number;
    errorMessage?: string | null;
    questIds: string[];
    quests?: Quest[];
};
