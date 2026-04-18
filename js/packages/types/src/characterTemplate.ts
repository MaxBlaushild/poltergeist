import type { CharacterStoryVariant } from './character';
import type { ZoneGenre } from './zone';

export interface CharacterTemplate {
  id: string;
  createdAt?: Date | string;
  updatedAt?: Date | string;
  name: string;
  description: string;
  genreId: string;
  genre?: ZoneGenre | null;
  internalTags?: string[];
  mapIconUrl?: string;
  dialogueImageUrl?: string;
  thumbnailUrl?: string;
  storyVariants?: CharacterStoryVariant[];
  imageGenerationStatus?: string;
  imageGenerationError?: string | null;
}
