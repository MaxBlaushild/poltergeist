import React from 'react';
import { PointOfInterestGroup, PointOfInterest, PointOfInterestChallenge } from '@poltergeist/types';
interface ArenaContextType {
    arena: PointOfInterestGroup | null;
    loading: boolean;
    error: Error | null;
    updateArena: (name: string, description: string) => Promise<void>;
    updateArenaImage: (id: string, image: File) => Promise<void>;
    createPointOfInterest: (name: string, description: string, lat: number, lng: number, image: File | null, clue: string) => Promise<void>;
    updatePointOfInterest: (id: string, arena: Partial<PointOfInterest>) => Promise<void>;
    updatePointOfInterestImage: (id: string, image: File) => Promise<void>;
    deletePointOfInterest: (id: string) => Promise<void>;
    updatePointOfInterestChallenge: (id: string, challenge: Partial<PointOfInterestChallenge>) => Promise<void>;
    deletePointOfInterestChallenge: (id: string) => Promise<void>;
    createPointOfInterestChallenge: (id: string, challenge: Partial<PointOfInterestChallenge>) => Promise<void>;
}
interface ArenaProviderProps {
    children: React.ReactNode;
    arenaId: string | undefined | null;
}
export declare const ArenaProvider: React.FC<ArenaProviderProps>;
export declare const useArena: () => ArenaContextType;
export {};
