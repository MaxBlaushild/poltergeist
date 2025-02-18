import { PointOfInterest } from "./pointOfInterest";
import { User } from "./user";
import { PointOfInterestDiscovery } from "./pointOfInterestDiscovery";
import { OwnedInventoryItem } from "./ownedInventoryItem";
export type Team = {
  id: string;
  createdAt: Date;
  updatedAt: Date;
  name: string;
  users: User[];
  pointOfInterestDiscoveries: PointOfInterestDiscovery[];
  ownedInventoryItems: OwnedInventoryItem[];
};
