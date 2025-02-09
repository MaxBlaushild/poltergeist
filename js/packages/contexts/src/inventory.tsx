import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react';
import { useAPI } from './api';
import { Category, Activity, InventoryItem } from '@poltergeist/types';

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
};

interface UseItemMetadata {
  targetTeamId?: string | null;
  pointOfInterestId?: string | null;
}

const InventoryContext = createContext<InventoryContextType>({
  inventoryItems: [],
  presentedInventoryItem: null,
  inventoryItemError: null,
  setPresentedInventoryItem: (item: InventoryItem | null) => {},
  inventoryItemsAreLoading: false,
  consumeItem: () => Promise.resolve(),
  useItemError: null,
  isUsingItem: false,
  usedItem: null,
  setUsedItem: (item: InventoryItem | null) => {},
});

export const useInventory = () => useContext(InventoryContext);

export const InventoryProvider = ({ children }: { children: ReactNode }) => {
  const { apiClient } = useAPI();
  const [inventoryItems, setInventoryItems] = useState<InventoryItem[]>([]);
  const [inventoryItemsAreLoading, setInventoryItemsAreLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [useItemError, setUseItemError] = useState<string | null>(null);
  const [isUsingItem, setIsUsingItem] = useState<boolean>(false);
  const [presentedInventoryItem, setPresentedInventoryItem] = useState<InventoryItem | null>(null);
  const [usedItem, setUsedItem] = useState<InventoryItem | null>(null);

  const fetchInventoryItems = async () => {
    setInventoryItemsAreLoading(true);
    setError(null);
    try {
      const response = await apiClient.get<InventoryItem[]>('/sonar/items');
      setInventoryItems(response);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch inventory items');
    } finally {
      setInventoryItemsAreLoading(false);
    }
  };

  useEffect(() => {
    fetchInventoryItems();
  }, []);

  const consumeItem = async (teamInventoryItemId: string, metadata: UseItemMetadata = {}) => {
    try {
      setIsUsingItem(true);
      await apiClient.post(`/sonar/inventory/${teamInventoryItemId}/use`, {
        ...metadata,
      });
    } catch (err) {
      setUseItemError(err instanceof Error ? err.message : 'Failed to use item');
    } finally {
      setIsUsingItem(false);
    }
  };

  return (
    <InventoryContext.Provider value={{ 
      inventoryItems, 
      inventoryItemsAreLoading, 
      setPresentedInventoryItem,
      presentedInventoryItem,
      inventoryItemError: error,
      consumeItem,
      useItemError,
      isUsingItem,
      usedItem,
      setUsedItem,
    }}>
      {children}
    </InventoryContext.Provider>
  );
};
