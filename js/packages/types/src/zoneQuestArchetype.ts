import { Zone } from "./zone";
import { QuestArchetype } from "./questArchetype";
import { Character } from "./character";

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
  characterId?: string | null;
  character?: Character | null;
};
