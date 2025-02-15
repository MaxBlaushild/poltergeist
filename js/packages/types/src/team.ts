import { PointOfInterest } from "./pointOfInterest";
import { User } from "./user";
import { TeamInventoryItem } from "./teamInventoryItem";
import { PointOfInterestDiscovery } from "./pointOfInterestDiscovery";
export type Team = {
  id: string;
  createdAt: Date;
  updatedAt: Date;
  name: string;
  users: User[];
  teamInventoryItems: TeamInventoryItem[];
  pointOfInterestDiscoveries: PointOfInterestDiscovery[];
};
