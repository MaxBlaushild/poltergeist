import { FeteRoom } from "./feteRoom";
import { FeteTeam } from "./feteTeam";

export interface FeteRoomTeam {
  id: string;
  createdAt: Date;
  updatedAt: Date;
  deletedAt: Date | null;
  feteRoomId: string;
  feteRoom?: FeteRoom;
  teamId: string;
  team?: FeteTeam;
}

