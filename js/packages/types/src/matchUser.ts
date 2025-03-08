import { User } from "./user";

export type MatchUser = {
  id: string;
  matchId: string;
  userId: string;
  user: User;
};
