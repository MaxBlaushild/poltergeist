import { QuestNodeSubmissionType } from "./questNode";

export interface LocationArchetypeChallenge {
  question: string;
  submissionType: QuestNodeSubmissionType;
}

export interface LocationArchetype {
  id: string;
  name: string;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date;
  includedTypes: string[];
  excludedTypes: string[];
  challenges: LocationArchetypeChallenge[];
}
