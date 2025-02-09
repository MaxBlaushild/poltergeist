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
    consumeItem: () => Promise.resolve(),
    useItemError: null,
    isUsingItem: false,
    usedItem: null,
    setUsedItem: (item) => { },
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
    useEffect(() => {
        fetchInventoryItems();
    }, []);
    const consumeItem = (teamInventoryItemId, metadata = {}) => __awaiter(void 0, void 0, void 0, function* () {
        try {
            setIsUsingItem(true);
            yield apiClient.post(`/sonar/inventory/${teamInventoryItemId}/use`, Object.assign({}, metadata));
        }
        catch (err) {
            setUseItemError(err instanceof Error ? err.message : 'Failed to use item');
        }
        finally {
            setIsUsingItem(false);
        }
    });
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
        } }, { children: children })));
};
