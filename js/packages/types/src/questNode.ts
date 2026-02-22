import { QuestNodeChallenge } from './questNodeChallenge';

export type QuestNodeSubmissionType = 'text' | 'photo' | 'video' | (string & {});

export interface QuestNode {
  id: string;
  questId: string;
  orderIndex: number;
  submissionType?: QuestNodeSubmissionType;
  pointOfInterestId?: string | null;
  polygon?: string | null;
  polygonPoints?: [number, number][];
  challenges?: QuestNodeChallenge[];
}
