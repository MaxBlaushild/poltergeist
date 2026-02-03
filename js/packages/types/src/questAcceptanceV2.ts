export interface QuestAcceptanceV2 {
  id: string;
  userId: string;
  questId: string;
  acceptedAt: string;
  turnedInAt?: string | null;
}
