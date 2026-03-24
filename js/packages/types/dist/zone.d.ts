import { Point } from './point';
export type Zone = {
    id: string;
    name: string;
    latitude: number;
    longitude: number;
    createdAt: string;
    updatedAt: string;
    description: string;
    internalTags?: string[];
    zoneImportId?: string | null;
    boundary: number[][];
    boundaryCoords: {
        latitude: number;
        longitude: number;
    }[];
    points: Point[];
};
