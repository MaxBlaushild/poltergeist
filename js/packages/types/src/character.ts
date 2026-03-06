import { CharacterLocation } from './characterLocation';

export interface Character {
  id: string;
  createdAt: Date;
  updatedAt: Date;
  name: string;
  description: string;
  mapIconUrl: string;
  dialogueImageUrl: string;
  thumbnailUrl?: string;
  imageGenerationStatus?: string;
  imageGenerationError?: string | null;
  locations?: CharacterLocation[];
  geometry?: string;
  latitude?: number | null;
  longitude?: number | null;
  pointOfInterestId?: string | null;
  pointOfInterest?: { id: string; name: string } | null;
}
