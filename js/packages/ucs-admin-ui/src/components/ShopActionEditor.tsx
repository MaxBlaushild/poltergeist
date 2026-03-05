import React, { useEffect, useMemo, useState } from 'react';
import { CharacterAction, InventoryItem, ShopInventoryItem, ShopMode } from '@poltergeist/types';
import { useAPI } from '@poltergeist/contexts';

export interface ShopActionSavePayload {
  inventory: ShopInventoryItem[];
  shopMode: ShopMode;
  shopItemTags: string[];
}

interface ShopActionEditorProps {
  action: CharacterAction | null;
  onSave: (payload: ShopActionSavePayload) => void;
  onCancel: () => void;
}

const normalizeShopTagList = (rawTags: unknown[]): string[] => {
  const seen = new Set<string>();
  const normalized: string[] = [];
  rawTags.forEach((entry) => {
    if (typeof entry !== 'string') return;
    const tag = entry.trim().toLowerCase();
    if (!tag || seen.has(tag)) return;
    seen.add(tag);
    normalized.push(tag);
  });
  return normalized.sort((a, b) => a.localeCompare(b));
};

const normalizeShopInventory = (rawInventory: unknown[]): ShopInventoryItem[] => {
  const seen = new Set<number>();
  const normalized: ShopInventoryItem[] = [];
  rawInventory.forEach((entry) => {
    if (!entry || typeof entry !== 'object') return;
    const itemId = Number((entry as { itemId?: unknown }).itemId);
    const price = Number((entry as { price?: unknown }).price);
    if (!Number.isFinite(itemId) || itemId <= 0 || !Number.isFinite(price) || price < 0) return;
    if (seen.has(itemId)) return;
    seen.add(itemId);
    normalized.push({ itemId: Math.floor(itemId), price: Math.floor(price) });
  });
  return normalized;
};

export const ShopActionEditor: React.FC<ShopActionEditorProps> = ({
  action,
  onSave,
  onCancel,
}) => {
  const { apiClient } = useAPI();
  const [inventory, setInventory] = useState<ShopInventoryItem[]>([]);
  const [shopMode, setShopMode] = useState<ShopMode>('explicit');
  const [shopItemTags, setShopItemTags] = useState<string[]>([]);
  const [availableItems, setAvailableItems] = useState<InventoryItem[]>([]);
  const [isLoadingItems, setIsLoadingItems] = useState(false);
  const [selectedItemId, setSelectedItemId] = useState<number | null>(null);
  const [priceInput, setPriceInput] = useState<string>('');
  const [selectedTag, setSelectedTag] = useState<string>('');

  useEffect(() => {
    const metadata = action?.metadata ?? {};
    const nextMode: ShopMode = metadata.shopMode === 'tags' ? 'tags' : 'explicit';
    setShopMode(nextMode);
    setInventory(normalizeShopInventory(Array.isArray(metadata.inventory) ? metadata.inventory : []));
    setShopItemTags(
      normalizeShopTagList(Array.isArray(metadata.shopItemTags) ? metadata.shopItemTags : [])
    );
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

  const availableTags = useMemo(() => {
    const seen = new Set<string>();
    availableItems.forEach((item) => {
      (item.internalTags ?? []).forEach((rawTag) => {
        const tag = rawTag.trim().toLowerCase();
        if (tag) seen.add(tag);
      });
    });
    return Array.from(seen).sort((a, b) => a.localeCompare(b));
  }, [availableItems]);

  const tagMatchingCount = useMemo(() => {
    if (shopItemTags.length === 0) return 0;
    const selectedTagSet = new Set(shopItemTags.map((tag) => tag.toLowerCase()));
    return availableItems.filter((item) =>
      (item.internalTags ?? []).some((tag) => selectedTagSet.has(tag.toLowerCase()))
    ).length;
  }, [availableItems, shopItemTags]);

  const handleAddItem = () => {
    if (!selectedItemId || !priceInput) return;

    const price = parseInt(priceInput, 10);
    if (isNaN(price) || price < 0) {
      alert('Please enter a valid price (non-negative number)');
      return;
    }

    if (inventory.find((item) => item.itemId === selectedItemId)) {
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
    setInventory(inventory.filter((item) => item.itemId !== itemId));
  };

  const handleUpdatePrice = (itemId: number, newPrice: number) => {
    setInventory(inventory.map((item) =>
      item.itemId === itemId ? { ...item, price: newPrice } : item
    ));
  };

  const handleAddTag = () => {
    const normalized = selectedTag.trim().toLowerCase();
    if (!normalized) return;
    if (shopItemTags.includes(normalized)) return;
    setShopItemTags([...shopItemTags, normalized].sort((a, b) => a.localeCompare(b)));
    setSelectedTag('');
  };

  const handleRemoveTag = (tag: string) => {
    setShopItemTags(shopItemTags.filter((existing) => existing !== tag));
  };

  const handleSave = () => {
    onSave({
      inventory,
      shopMode,
      shopItemTags,
    });
  };

  const getItemName = (itemId: number) => {
    const item = availableItems.find((i) => i.id === itemId);
    return item ? item.name : `Item ${itemId}`;
  };

  const getItemImageUrl = (itemId: number) => {
    const item = availableItems.find((i) => i.id === itemId);
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

      <div style={{
        border: '1px solid #ccc',
        borderRadius: '8px',
        padding: '15px',
        backgroundColor: '#f9f9f9'
      }}>
        <h4 style={{ margin: '0 0 10px 0', fontSize: '14px', fontWeight: 'bold' }}>
          Shop Inventory Source
        </h4>
        <div style={{ display: 'flex', gap: '16px', alignItems: 'center', flexWrap: 'wrap' }}>
          <label style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer' }}>
            <input
              type="radio"
              name="shopMode"
              value="explicit"
              checked={shopMode === 'explicit'}
              onChange={() => setShopMode('explicit')}
            />
            Explicit item list
          </label>
          <label style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer' }}>
            <input
              type="radio"
              name="shopMode"
              value="tags"
              checked={shopMode === 'tags'}
              onChange={() => setShopMode('tags')}
            />
            Tag-based inventory
          </label>
        </div>
      </div>

      {shopMode === 'explicit' ? (
        <>
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
                  .filter((item) => !inventory.find((invItem) => invItem.itemId === item.id))
                  .map((item) => (
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
                {inventory.map((item) => (
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
                ))}
              </div>
            )}
          </div>
        </>
      ) : (
        <div style={{
          flex: 1,
          overflowY: 'auto',
          border: '1px solid #ccc',
          borderRadius: '8px',
          padding: '15px',
          backgroundColor: '#f9f9f9',
          display: 'flex',
          flexDirection: 'column',
          gap: '12px'
        }}>
          <h4 style={{ margin: 0, fontSize: '14px', fontWeight: 'bold' }}>
            Configure Item Tags
          </h4>
          <div style={{ color: '#666', fontSize: '13px' }}>
            Players see items matching any selected tag where item level is within +/- 15 of their current level.
          </div>
          <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}>
            <select
              value={selectedTag}
              onChange={(e) => setSelectedTag(e.target.value)}
              style={{
                flex: 1,
                padding: '8px',
                border: '1px solid #ccc',
                borderRadius: '4px',
                fontSize: '14px'
              }}
            >
              <option value="">Select an item tag...</option>
              {availableTags
                .filter((tag) => !shopItemTags.includes(tag))
                .map((tag) => (
                  <option key={tag} value={tag}>
                    {tag}
                  </option>
                ))}
            </select>
            <button
              onClick={handleAddTag}
              disabled={!selectedTag}
              style={{
                padding: '8px 20px',
                backgroundColor: '#4caf50',
                color: 'white',
                border: 'none',
                borderRadius: '6px',
                cursor: selectedTag ? 'pointer' : 'not-allowed',
                fontSize: '14px',
                fontWeight: 'bold',
                opacity: selectedTag ? 1 : 0.6
              }}
            >
              Add Tag
            </button>
          </div>

          <div style={{ color: '#666', fontSize: '12px' }}>
            {isLoadingItems
              ? 'Loading items and tags...'
              : `${shopItemTags.length} selected tag(s), ${tagMatchingCount} matching item(s) across all levels.`}
          </div>

          {shopItemTags.length === 0 ? (
            <div style={{
              textAlign: 'center',
              color: '#999',
              padding: '24px',
              fontStyle: 'italic',
              border: '1px dashed #ddd',
              borderRadius: '8px',
              backgroundColor: '#fff'
            }}>
              No tags selected yet.
            </div>
          ) : (
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
              {shopItemTags.map((tag) => (
                <span
                  key={tag}
                  style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: '8px',
                    backgroundColor: '#e3f2fd',
                    color: '#0d47a1',
                    border: '1px solid #bbdefb',
                    borderRadius: '16px',
                    padding: '6px 10px',
                    fontSize: '13px'
                  }}
                >
                  {tag}
                  <button
                    type="button"
                    onClick={() => handleRemoveTag(tag)}
                    style={{
                      border: 'none',
                      background: 'transparent',
                      color: '#0d47a1',
                      cursor: 'pointer',
                      fontSize: '12px',
                      fontWeight: 'bold',
                      padding: 0,
                    }}
                  >
                    x
                  </button>
                </span>
              ))}
            </div>
          )}
        </div>
      )}

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
