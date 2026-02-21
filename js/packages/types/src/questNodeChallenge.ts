export interface QuestNodeChallenge {
  id: string;
  questNodeId: string;
  tier: number;
  question: string;
  reward: number;
  inventoryItemId?: number | null;
  difficulty?: number;
  statTags?: string[];
  proficiency?: string | null;
}
