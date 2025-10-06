import { UserProfile } from "./userProfile";

export type User = {
  phoneNumber: string;
  name: string;
  id: string;
  profilePictureUrl: string;
  partyId: string | null;
};
