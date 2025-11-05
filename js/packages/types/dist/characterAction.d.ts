import { Character } from './character';
export type ActionType = 'talk';
export interface DialogueMessage {
    speaker: 'character' | 'user';
    text: string;
    order: number;
}
export interface CharacterAction {
    id: string;
    createdAt: Date;
    updatedAt: Date;
    characterId: string;
    character?: Character;
    actionType: ActionType;
    dialogue: DialogueMessage[];
    metadata?: any;
}

