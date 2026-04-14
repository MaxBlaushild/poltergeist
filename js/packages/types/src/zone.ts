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

export type ZoneAdminSummary = {
  id: string;
  name: string;
  latitude: number;
  longitude: number;
  createdAt: string;
  updatedAt: string;
  description: string;
  internalTags?: string[];
  zoneImportId?: string | null;
  importMetroName?: string | null;
  boundaryPointCount: number;
  pointOfInterestCount: number;
  questCount: number;
  zoneQuestArchetypeCount: number;
  challengeCount: number;
  scenarioCount: number;
  monsterCount: number;
  monsterEncounterCount: number;
  standardEncounterCount: number;
  bossEncounterCount: number;
  raidEncounterCount: number;
  treasureChestCount: number;
  healingFountainCount: number;
};
