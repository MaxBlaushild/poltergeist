import { QuestNodeChallenge } from './questNodeChallenge';

export interface QuestNode {
  id: string;
  questId: string;
  orderIndex: number;
  pointOfInterestId?: string | null;
  polygon?: string | null;
  polygonPoints?: [number, number][];
  challenges?: QuestNodeChallenge[];
}
