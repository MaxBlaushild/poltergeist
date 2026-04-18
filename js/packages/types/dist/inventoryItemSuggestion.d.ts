import { InventoryItem } from './inventoryItem';
import type { ZoneGenre } from './zone';
export interface InventoryItemSuggestionJob {
    id: string;
    createdAt: string;
    updatedAt: string;
    genreId: string;
    genre?: ZoneGenre | null;
    status: string;
    count: number;
    themePrompt: string;
    categories: string[];
    rarityTiers: string[];
    equipSlots: string[];
    statTags: string[];
    benefitTags: string[];
    statusNames: string[];
    internalTags: string[];
    minItemLevel: number;
    maxItemLevel: number;
    createdCount: number;
    errorMessage?: string | null;
}
export interface InventoryItemSuggestionPayload {
    category: string;
    whyItFits: string;
    item: InventoryItem;
}
export interface InventoryItemSuggestionDraft {
    id: string;
    createdAt: string;
    updatedAt: string;
    jobId: string;
    status: string;
    name: string;
    category: string;
    rarityTier: string;
    itemLevel: number;
    equipSlot?: string | null;
    whyItFits: string;
    internalTags: string[];
    warnings: string[];
    payload: InventoryItemSuggestionPayload;
    inventoryItemId?: number | null;
    inventoryItem?: InventoryItem | null;
    convertedAt?: string | null;
}
