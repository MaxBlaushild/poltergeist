import type { User } from "./user";

type FriendInvite = {
    id: string;
    createdAt: string;
    updatedAt: string;
    inviterId: string;
    inviteeId: string;
    invitee: User;
    inviter: User;
  };
  
  export type { FriendInvite };