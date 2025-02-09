import { ReactNode } from 'react';
import { InventoryItem } from '@poltergeist/types';
interface InventoryContextType {
    inventoryItems: InventoryItem[];
    presentedInventoryItem: InventoryItem | null;
    inventoryItemError: string | null;
    setPresentedInventoryItem: (inventoryItem: InventoryItem | null) => void;
    inventoryItemsAreLoading: boolean;
    consumeItem: (teamInventoryItemId: string, metadata?: UseItemMetadata) => Promise<void>;
    useItemError: string | null;
    isUsingItem: boolean;
    usedItem: InventoryItem | null;
    setUsedItem: (inventoryItem: InventoryItem | null) => void;
}
interface UseItemMetadata {
    targetTeamId?: string | null;
    pointOfInterestId?: string | null;
}
export declare const useInventory: () => InventoryContextType;
export declare const InventoryProvider: ({ children }: {
    children: ReactNode;
}) => import("react/jsx-runtime").JSX.Element;
export {};
