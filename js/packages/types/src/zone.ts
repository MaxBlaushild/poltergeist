import { Point } from './point';

export type ZoneGenre = {
  id: string;
  name: string;
  sortOrder: number;
  active: boolean;
  promptSeed?: string;
  createdAt?: string;
  updatedAt?: string;
};

export type ZoneGenreScore = {
  genreId: string;
  genre: ZoneGenre;
  score: number;
};

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
  genreScores?: ZoneGenreScore[];
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
