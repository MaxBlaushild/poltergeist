import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react';
import { useAPI } from './api';
import { Category, Activity, InventoryItem, OwnedInventoryItem } from '@poltergeist/types';

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
};

interface UseItemMetadata {
  targetTeamId?: string | null;
  pointOfInterestId?: string | null;
  challengeId?: string | null;
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
  ownedInventoryItems: [],
  ownedInventoryItemsAreLoading: false,
  ownedInventoryItemsError: null,
  getInventoryItemById: (id: number) => null,
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
  const [ownedInventoryItems, setOwnedInventoryItems] = useState<OwnedInventoryItem[]>([]);
  const [ownedInventoryItemsAreLoading, setOwnedInventoryItemsAreLoading] = useState<boolean>(false);
  const [ownedInventoryItemsError, setOwnedInventoryItemsError] = useState<string | null>(null);

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

  const fetchOwnedInventoryItems = async () => {
    setOwnedInventoryItemsAreLoading(true);
    setOwnedInventoryItemsError(null);
    try {
      const response = await apiClient.get<OwnedInventoryItem[]>('/sonar/ownedInventoryItems');
      setOwnedInventoryItems(response.filter((item) => item.quantity > 0));
    } catch (err) {
      setOwnedInventoryItemsError(err instanceof Error ? err.message : 'Failed to fetch owned inventory items');
    } finally {
      setOwnedInventoryItemsAreLoading(false);
    }
  };

  useEffect(() => {
    fetchInventoryItems();
    fetchOwnedInventoryItems();
  }, []);

  const consumeItem = async (ownedInventoryItemId: string, metadata: UseItemMetadata = {}) => {
    try {
      setIsUsingItem(true);
      await apiClient.post(`/sonar/inventory/${ownedInventoryItemId}/use`, {
        ...metadata,
      });
    } catch (err) {
      setUseItemError(err instanceof Error ? err.message : 'Failed to use item');
    } finally {
      setIsUsingItem(false);
      fetchOwnedInventoryItems();
    }
  };

  const getInventoryItemById = (id: number) => {
    return inventoryItems.find((item) => item.id === id) || null;
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
      ownedInventoryItems,
      ownedInventoryItemsAreLoading,
      ownedInventoryItemsError,
      getInventoryItemById,
    }}>
      {children}
    </InventoryContext.Provider>
  );
};
