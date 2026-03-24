import { Zone } from './zone';

export type District = {
  id: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string | null;
  zones: Zone[];
};
