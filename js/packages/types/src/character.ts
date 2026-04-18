import type { DialogueMessage } from './characterAction';
import { CharacterLocation } from './characterLocation';
import type { ZoneGenre } from './zone';

export interface CharacterStoryVariant {
  id?: string;
  createdAt?: Date;
  updatedAt?: Date;
  priority?: number;
  requiredStoryFlags?: string[];
  description?: string;
  dialogue?: DialogueMessage[];
}

export interface Character {
  id: string;
  createdAt: Date;
  updatedAt: Date;
  name: string;
  description: string;
  genreId: string;
  genre?: ZoneGenre | null;
  internalTags?: string[];
  mapIconUrl: string;
  dialogueImageUrl: string;
  thumbnailUrl?: string;
  imageGenerationStatus?: string;
  imageGenerationError?: string | null;
  storyVariants?: CharacterStoryVariant[];
  locations?: CharacterLocation[];
  pointOfInterestId?: string | null;
  pointOfInterest?: { id: string; name: string } | null;
}
