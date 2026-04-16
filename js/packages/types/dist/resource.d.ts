import type { ResourceType } from './resourceType';
export type Resource = {
    id: string;
    zoneId: string;
    resourceTypeId: string;
    resourceType: ResourceType;
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
