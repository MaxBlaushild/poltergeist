import type {
  ExpositionMaterialReward,
  ExpositionRandomRewardSize,
  ExpositionRewardMode,
} from './exposition';
import type { DialogueMessage } from './characterAction';

export interface ExpositionTemplateItemReward {
  inventoryItemId: number;
  quantity: number;
}

export interface ExpositionTemplateSpellReward {
  spellId: string;
}

export interface ExpositionTemplate {
  id: string;
  createdAt?: Date | string;
  updatedAt?: Date | string;
  title: string;
  description: string;
  dialogue: DialogueMessage[];
  requiredStoryFlags?: string[];
  imageUrl?: string;
  thumbnailUrl?: string;
  rewardMode?: ExpositionRewardMode;
  randomRewardSize?: ExpositionRandomRewardSize;
  rewardExperience?: number;
  rewardGold?: number;
  materialRewards?: ExpositionMaterialReward[];
  itemRewards?: ExpositionTemplateItemReward[];
  spellRewards?: ExpositionTemplateSpellReward[];
}
