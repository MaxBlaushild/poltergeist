import React from 'react';
import { Zone } from '@poltergeist/types';
export declare const calculateDistance: (poi1: any, poi2: any) => number;
export declare const isXMetersAway: (poi1: any, poi2: any, x: any) => boolean;
type ZoneContextType = {
    zones: Zone[];
    selectedZone: Zone | null;
    setSelectedZone: (zone: Zone | null) => void;
    createZone: (zone: Zone) => void;
    deleteZone: (zone: Zone) => void;
    findZoneAtCoordinate: (lng: number, lat: number) => Zone | null;
    editZone: (name: string, description: string, id: string) => void;
    refreshZones: () => Promise<void>;
};
export declare const ZoneProvider: React.FC<{
    children: React.ReactNode;
}>;
export declare const useZoneContext: () => ZoneContextType;
export {};
