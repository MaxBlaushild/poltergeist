import { Zone } from './zone';
export type MovementPatternType = 'static' | 'random' | 'path';
export interface Location {
    latitude: number;
    longitude: number;
}
export interface MovementPattern {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    movementPatternType: MovementPatternType;
    zoneId?: string;
    zone?: Zone;
    startingLatitude: number;
    startingLongitude: number;
    path: Location[];
}
