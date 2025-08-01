var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
import { jsx as _jsx } from "react/jsx-runtime";
import { createContext, useContext, useState, useEffect } from 'react';
import { useAPI } from './api';
;
const InventoryContext = createContext({
    inventoryItems: [],
    presentedInventoryItem: null,
    inventoryItemError: null,
    setPresentedInventoryItem: (item) => { },
    inventoryItemsAreLoading: false,
    consumeItem: () => Promise.resolve(undefined),
    useItemError: null,
    isUsingItem: false,
    usedItem: null,
    setUsedItem: (item) => { },
    ownedInventoryItems: [],
    ownedInventoryItemsAreLoading: false,
    ownedInventoryItemsError: null,
    getInventoryItemById: (id) => null,
});
export const useInventory = () => useContext(InventoryContext);
export const InventoryProvider = ({ children }) => {
    const { apiClient } = useAPI();
    const [inventoryItems, setInventoryItems] = useState([]);
    const [inventoryItemsAreLoading, setInventoryItemsAreLoading] = useState(false);
    const [error, setError] = useState(null);
    const [useItemError, setUseItemError] = useState(null);
    const [isUsingItem, setIsUsingItem] = useState(false);
    const [presentedInventoryItem, setPresentedInventoryItem] = useState(null);
    const [usedItem, setUsedItem] = useState(null);
    const [ownedInventoryItems, setOwnedInventoryItems] = useState([]);
    const [ownedInventoryItemsAreLoading, setOwnedInventoryItemsAreLoading] = useState(false);
    const [ownedInventoryItemsError, setOwnedInventoryItemsError] = useState(null);
    const fetchInventoryItems = () => __awaiter(void 0, void 0, void 0, function* () {
        setInventoryItemsAreLoading(true);
        setError(null);
        try {
            const response = yield apiClient.get('/sonar/items');
            setInventoryItems(response);
        }
        catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to fetch inventory items');
        }
        finally {
            setInventoryItemsAreLoading(false);
        }
    });
    const fetchOwnedInventoryItems = () => __awaiter(void 0, void 0, void 0, function* () {
        setOwnedInventoryItemsAreLoading(true);
        setOwnedInventoryItemsError(null);
        try {
            const response = yield apiClient.get('/sonar/ownedInventoryItems');
            setOwnedInventoryItems(response.filter((item) => item.quantity > 0));
        }
        catch (err) {
            setOwnedInventoryItemsError(err instanceof Error ? err.message : 'Failed to fetch owned inventory items');
        }
        finally {
            setOwnedInventoryItemsAreLoading(false);
        }
    });
    useEffect(() => {
        fetchInventoryItems();
        fetchOwnedInventoryItems();
    }, []);
    const consumeItem = (ownedInventoryItemId, metadata = {}) => __awaiter(void 0, void 0, void 0, function* () {
        try {
            setIsUsingItem(true);
            const result = yield apiClient.post(`/sonar/inventory/${ownedInventoryItemId}/use`, Object.assign({}, metadata));
            return result;
        }
        catch (err) {
            setUseItemError(err instanceof Error ? err.message : 'Failed to use item');
        }
        finally {
            setIsUsingItem(false);
            fetchOwnedInventoryItems();
        }
    });
    const getInventoryItemById = (id) => {
        return inventoryItems.find((item) => item.id === id) || null;
    };
    return (_jsx(InventoryContext.Provider, Object.assign({ value: {
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
        } }, { children: children })));
};
