import { QuestNode } from './questNode';

export interface Quest {
  id: string;
  createdAt: string;
  updatedAt: string;
  name: string;
  description: string;
  imageUrl?: string;
  zoneId?: string | null;
  questArchetypeId?: string | null;
  questGiverCharacterId?: string | null;
  gold?: number;
  nodes?: QuestNode[];
}
