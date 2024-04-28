import { Activity } from './activity';
import { User } from './user';

export type Survey = {
  id: string;
  title: string;
  createdAt: string;
  updatedAt: string;
  referrerId: string;
  progenitorId: string;
  activities: Activity[];
  user: User;
};
