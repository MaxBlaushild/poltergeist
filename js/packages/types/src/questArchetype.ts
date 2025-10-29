import { LocationArchetype } from "./locationArchetype";

export interface QuestArchetypeChallenge {
  id: string;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date;
  reward: number;
  unlockedNodeId?: string;
  unlockedNode?: QuestArchetypeNode;
}

export interface QuestArchetypeNode {
  id: string;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date;
  locationArchetype: LocationArchetype;
  locationArchetypeId: string;
  challenges: QuestArchetypeChallenge[];
}

export interface QuestArchetype {
  id: string;
  name: string;
  defaultGold: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date;
  root: QuestArchetypeNode;
  rootId: string;
}
