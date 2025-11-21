import { Character } from './character';

export type ActionType = 'talk' | 'shop' | 'giveQuest';

export interface DialogueMessage {
  speaker: 'character' | 'user';
  text: string;
  order: number;
}

export interface ShopInventoryItem {
  itemId: number;
  price: number;
}

export interface CharacterAction {
  id: string;
  createdAt: Date;
  updatedAt: Date;
  characterId: string;
  character?: Character;
  actionType: ActionType;
  dialogue: DialogueMessage[];
  metadata?: {
    inventory?: ShopInventoryItem[];
    pointOfInterestGroupId?: string;
    [key: string]: any;
  };
}

