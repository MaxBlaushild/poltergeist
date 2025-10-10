import { User } from "./user";

export type Party = {
  id: string;
  createdAt: string;
  updatedAt: string;
  leaderId: string;
  leader: User;
  members: User[];
};