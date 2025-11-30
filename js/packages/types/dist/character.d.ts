import { MovementPattern } from './movementPattern';
export interface Character {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    name: string;
    description: string;
    mapIconUrl: string;
    dialogueImageUrl: string;
    geometry?: string;
    movementPatternId: string;
    movementPattern: MovementPattern;
}
