import React, { useState, useEffect } from 'react';
import { CharacterAction, ShopInventoryItem, InventoryItem } from '@poltergeist/types';
import { useAPI } from '@poltergeist/contexts';

interface ShopActionEditorProps {
  action: CharacterAction | null;
  onSave: (inventory: ShopInventoryItem[]) => void;
  onCancel: () => void;
}

export const ShopActionEditor: React.FC<ShopActionEditorProps> = ({
  action,
  onSave,
  onCancel,
}) => {
  const { apiClient } = useAPI();
  const [inventory, setInventory] = useState<ShopInventoryItem[]>([]);
  const [availableItems, setAvailableItems] = useState<InventoryItem[]>([]);
  const [isLoadingItems, setIsLoadingItems] = useState(false);
  const [selectedItemId, setSelectedItemId] = useState<number | null>(null);
  const [priceInput, setPriceInput] = useState<string>('');

  useEffect(() => {
    if (action && action.metadata?.inventory) {
      setInventory(action.metadata.inventory);
    } else {
      setInventory([]);
    }
  }, [action]);

  useEffect(() => {
    const fetchItems = async () => {
      setIsLoadingItems(true);
      try {
        const response = await apiClient.get<InventoryItem[]>('/sonar/items');
        setAvailableItems(response);
      } catch (error) {
        console.error('Error fetching inventory items:', error);
      } finally {
        setIsLoadingItems(false);
      }
    };

    fetchItems();
  }, [apiClient]);

  const handleAddItem = () => {
    if (!selectedItemId || !priceInput) return;

    const price = parseInt(priceInput, 10);
    if (isNaN(price) || price < 0) {
      alert('Please enter a valid price (non-negative number)');
      return;
    }

    // Check if item already exists in inventory
    if (inventory.find(item => item.itemId === selectedItemId)) {
      alert('This item is already in the shop inventory');
      return;
    }

    const newItem: ShopInventoryItem = {
      itemId: selectedItemId,
      price: price,
    };

    setInventory([...inventory, newItem]);
    setSelectedItemId(null);
    setPriceInput('');
  };

  const handleRemoveItem = (itemId: number) => {
    setInventory(inventory.filter(item => item.itemId !== itemId));
  };

  const handleUpdatePrice = (itemId: number, newPrice: number) => {
    setInventory(inventory.map(item =>
      item.itemId === itemId ? { ...item, price: newPrice } : item
    ));
  };

  const handleSave = () => {
    onSave(inventory);
  };

  const getItemName = (itemId: number) => {
    const item = availableItems.find(i => i.id === itemId);
    return item ? item.name : `Item ${itemId}`;
  };

  const getItemImageUrl = (itemId: number) => {
    const item = availableItems.find(i => i.id === itemId);
    return item ? item.imageUrl : '';
  };

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      gap: '20px',
      minHeight: '300px',
      maxHeight: '600px',
      overflow: 'hidden'
    }}>
      <h3 style={{ margin: 0, fontSize: '18px', fontWeight: 'bold' }}>
        Edit Shop Inventory
      </h3>

      {/* Add Item Section */}
      <div style={{
        border: '1px solid #ccc',
        borderRadius: '8px',
        padding: '15px',
        backgroundColor: '#f9f9f9'
      }}>
        <h4 style={{ margin: '0 0 10px 0', fontSize: '14px', fontWeight: 'bold' }}>
          Add Item to Shop
        </h4>
        <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
          <select
            value={selectedItemId || ''}
            onChange={(e) => setSelectedItemId(e.target.value ? parseInt(e.target.value, 10) : null)}
            style={{
              flex: 1,
              padding: '8px',
              border: '1px solid #ccc',
              borderRadius: '4px',
              fontSize: '14px'
            }}
          >
            <option value="">Select an item...</option>
            {availableItems
              .filter(item => !inventory.find(invItem => invItem.itemId === item.id))
              .map(item => (
                <option key={item.id} value={item.id}>
                  {item.name}
                </option>
              ))}
          </select>
          <input
            type="number"
            placeholder="Price"
            value={priceInput}
            onChange={(e) => setPriceInput(e.target.value)}
            min="0"
            style={{
              width: '100px',
              padding: '8px',
              border: '1px solid #ccc',
              borderRadius: '4px',
              fontSize: '14px'
            }}
          />
          <button
            onClick={handleAddItem}
            disabled={!selectedItemId || !priceInput}
            style={{
              padding: '8px 20px',
              backgroundColor: '#4caf50',
              color: 'white',
              border: 'none',
              borderRadius: '6px',
              cursor: selectedItemId && priceInput ? 'pointer' : 'not-allowed',
              fontSize: '14px',
              fontWeight: 'bold',
              opacity: selectedItemId && priceInput ? 1 : 0.6
            }}
          >
            Add
          </button>
        </div>
      </div>

      {/* Inventory List */}
      <div style={{
        flex: 1,
        overflowY: 'auto',
        border: '1px solid #ccc',
        borderRadius: '8px',
        padding: '15px',
        backgroundColor: '#f9f9f9'
      }}>
        {inventory.length === 0 ? (
          <div style={{
            textAlign: 'center',
            color: '#999',
            padding: '40px',
            fontStyle: 'italic'
          }}>
            No items in shop inventory yet. Add one to get started.
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
            {inventory.map((item) => {
              const itemDetails = availableItems.find(i => i.id === item.itemId);
              return (
                <div
                  key={item.itemId}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '15px',
                    padding: '15px',
                    backgroundColor: 'white',
                    borderRadius: '8px',
                    border: '1px solid #e0e0e0'
                  }}
                >
                  {/* Item Image */}
                  {getItemImageUrl(item.itemId) && (
                    <img
                      src={getItemImageUrl(item.itemId)}
                      alt={getItemName(item.itemId)}
                      style={{
                        width: '60px',
                        height: '60px',
                        objectFit: 'contain',
                        border: '1px solid #ccc',
                        borderRadius: '4px',
                        backgroundColor: '#f9f9f9'
                      }}
                    />
                  )}

                  {/* Item Details */}
                  <div style={{ flex: 1 }}>
                    <div style={{ fontWeight: 'bold', marginBottom: '5px' }}>
                      {getItemName(item.itemId)}
                    </div>
                    <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
                      <label style={{ fontSize: '12px', color: '#666' }}>Price:</label>
                      <input
                        type="number"
                        value={item.price}
                        onChange={(e) => {
                          const newPrice = parseInt(e.target.value, 10);
                          if (!isNaN(newPrice) && newPrice >= 0) {
                            handleUpdatePrice(item.itemId, newPrice);
                          }
                        }}
                        min="0"
                        style={{
                          width: '80px',
                          padding: '4px 8px',
                          border: '1px solid #ccc',
                          borderRadius: '4px',
                          fontSize: '14px'
                        }}
                      />
                      <span style={{ fontSize: '12px', color: '#666' }}>gold</span>
                    </div>
                  </div>

                  {/* Remove Button */}
                  <button
                    onClick={() => handleRemoveItem(item.itemId)}
                    style={{
                      padding: '6px 12px',
                      backgroundColor: '#f44336',
                      color: 'white',
                      border: 'none',
                      borderRadius: '4px',
                      cursor: 'pointer',
                      fontSize: '12px',
                      fontWeight: 'bold'
                    }}
                  >
                    Remove
                  </button>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Action Buttons */}
      <div style={{ display: 'flex', gap: '10px', justifyContent: 'flex-end' }}>
        <button
          onClick={onCancel}
          style={{
            padding: '10px 20px',
            backgroundColor: '#757575',
            color: 'white',
            border: 'none',
            borderRadius: '6px',
            cursor: 'pointer',
            fontSize: '14px'
          }}
        >
          Cancel
        </button>
        <button
          onClick={handleSave}
          style={{
            padding: '10px 20px',
            backgroundColor: '#4caf50',
            color: 'white',
            border: 'none',
            borderRadius: '6px',
            cursor: 'pointer',
            fontSize: '14px',
            fontWeight: 'bold'
          }}
        >
          Save Shop Inventory
        </button>
      </div>
    </div>
  );
};

