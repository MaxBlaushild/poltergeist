import { ReactNode } from 'react';
import { InventoryItem, OwnedInventoryItem } from '@poltergeist/types';
interface InventoryContextType {
    inventoryItems: InventoryItem[];
    presentedInventoryItem: InventoryItem | null;
    inventoryItemError: string | null;
    setPresentedInventoryItem: (inventoryItem: InventoryItem | null) => void;
    inventoryItemsAreLoading: boolean;
    consumeItem: (ownedInventoryItemId: string, metadata?: UseItemMetadata) => Promise<void>;
    useItemError: string | null;
    isUsingItem: boolean;
    usedItem: InventoryItem | null;
    setUsedItem: (inventoryItem: InventoryItem | null) => void;
    ownedInventoryItems: OwnedInventoryItem[];
    ownedInventoryItemsAreLoading: boolean;
    ownedInventoryItemsError: string | null;
    getInventoryItemById: (id: number) => InventoryItem | null;
}
interface UseItemMetadata {
    targetTeamId?: string | null;
    pointOfInterestId?: string | null;
    challengeId?: string | null;
}
export declare const useInventory: () => InventoryContextType;
export declare const InventoryProvider: ({ children }: {
    children: ReactNode;
}) => import("react/jsx-runtime").JSX.Element;
export {};
