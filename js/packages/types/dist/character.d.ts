import { MovementPattern } from './movementPattern';
import { PointOfInterest } from './pointOfInterest';
export interface Character {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    name: string;
    description: string;
    mapIconUrl: string;
    dialogueImageUrl: string;
    locationId: string;
    location?: PointOfInterest;
    movementPatternId: string;
    movementPattern: MovementPattern;
}
