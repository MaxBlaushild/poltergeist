export type UserZoneReputationName = 'neutral' | 'friendly' | 'honored' | 'revered' | 'exalted' | 'legendary';

export interface UserZoneReputation {
  id: string;
  createdAt: string;
  updatedAt: string;
  userId: string;
  zoneId: string;
  level: number;
  totalReputation: number;
  reputationOnLevel: number;
  levelsGained: number;
  name: UserZoneReputationName;
  reputationToNextLevel: number;
}
