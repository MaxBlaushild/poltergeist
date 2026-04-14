import type { CharacterStoryVariant } from './character';

export interface CharacterTemplate {
  id: string;
  createdAt?: Date | string;
  updatedAt?: Date | string;
  name: string;
  description: string;
  internalTags?: string[];
  mapIconUrl?: string;
  dialogueImageUrl?: string;
  thumbnailUrl?: string;
  storyVariants?: CharacterStoryVariant[];
  imageGenerationStatus?: string;
  imageGenerationError?: string | null;
}
