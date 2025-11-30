import { FeteRoom } from "./feteRoom";
import { FeteTeam } from "./feteTeam";
export interface FeteRoomLinkedListTeam {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    deletedAt: Date | null;
    feteRoomId: string;
    feteRoom?: FeteRoom;
    firstTeamId: string;
    firstTeam?: FeteTeam;
    secondTeamId: string;
    secondTeam?: FeteTeam;
}
