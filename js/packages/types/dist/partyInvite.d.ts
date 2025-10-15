import { User } from "./user";
export type PartyInvite = {
    id: string;
    createdAt: string;
    updatedAt: string;
    inviterId: string;
    inviteeId: string;
    invitee: User;
    inviter: User;
};
