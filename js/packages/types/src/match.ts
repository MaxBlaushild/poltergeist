import { MatchInventoryItemEffect } from "./matchInventoryItemEffect";
import { MatchUser } from "./matchUser";
import { PointOfInterest } from "./pointOfInterest";
import { Team } from "./team";
import { User } from "./user";
import { VerificationCode } from "./verificationCode";

export type Match = {
  id: string;
  createdAt: Date;
  updatedAt: Date;
  creatorId: string;
  startedAt?: Date;
  endedAt?: Date;
  verificationCodes: VerificationCode[];
  pointsOfInterest: PointOfInterest[];
  teams: Team[];
  inventoryItemEffects: MatchInventoryItemEffect[];
  users: User[];
};
