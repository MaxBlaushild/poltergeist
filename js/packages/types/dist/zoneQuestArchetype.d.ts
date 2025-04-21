import { Zone } from "./zone";
import { QuestArchetype } from "./questArchetype";
export type ZoneQuestArchetype = {
    id: string;
    createdAt: string;
    updatedAt: string;
    deletedAt: string | null;
    zone: Zone;
    zoneId: string;
    questArchetype: QuestArchetype;
    questArchetypeId: string;
    numberOfQuests: number;
};
