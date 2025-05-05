import React from 'react';
import { Zone } from '@poltergeist/types';
type ZoneContextType = {
    zones: Zone[];
    selectedZone: Zone | null;
    setSelectedZone: (zone: Zone | null) => void;
    createZone: (zone: Zone) => void;
    deleteZone: (zone: Zone) => void;
};
export declare const ZoneContext: React.Context<ZoneContextType>;
export declare const ZoneProvider: ({ children }: {
    children: React.ReactNode;
}) => import("react/jsx-runtime").JSX.Element;
export declare const useZoneContext: () => ZoneContextType;
export {};
