import type { InventoryItem } from './inventoryItem';
import type { ResourceType } from './resourceType';

export type ResourceGatherRequirement = {
  id?: string;
  resourceId?: string;
  minLevel: number;
  maxLevel: number;
  requiredInventoryItemId: number;
  requiredInventoryItem?: InventoryItem | null;
  createdAt?: string;
  updatedAt?: string;
};

export type Resource = {
  id: string;
  zoneId: string;
  resourceTypeId: string;
  resourceType: ResourceType;
  gatherRequirements?: ResourceGatherRequirement[];
  quantity: number;
  latitude: number;
  longitude: number;
  invalidated?: boolean;
  gatheredByUser?: boolean;
  createdAt?: string;
  updatedAt?: string;
  zone?: {
    id: string;
    name: string;
  };
};
