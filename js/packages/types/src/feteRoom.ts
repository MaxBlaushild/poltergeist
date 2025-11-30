import { FeteTeam } from "./feteTeam";

export interface FeteRoom {
  id: string;
  createdAt: Date;
  updatedAt: Date;
  deletedAt: Date | null;
  name: string;
  open: boolean;
  currentTeamId: string;
  hueLightId?: number | null;
  currentTeam?: FeteTeam;
}

