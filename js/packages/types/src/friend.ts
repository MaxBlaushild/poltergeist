import { User } from "./user";

type Friend = {
  id: string;
  createdAt: string;
  updatedAt: string;
  firstUserId: string;
  secondUserId: string;
  firstUser?: User;
  secondUser?: User;
};

export type { Friend };