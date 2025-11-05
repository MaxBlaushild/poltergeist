import React, { useState, useEffect } from 'react';
import { Character, CharacterAction } from '@poltergeist/types';
import { useInventory } from '@poltergeist/contexts';
import { useAuth } from '@poltergeist/contexts';
import { useAPI } from '@poltergeist/contexts';
import './Shop.css';

interface ShopProps {
  character: Character;
  action: CharacterAction;
  onClose: () => void;
}

type Tab = 'buy' | 'sell';

export const Shop: React.FC<ShopProps> = ({
  character,
  action,
  onClose,
}) => {
  const { inventoryItems, getInventoryItemById, refreshOwnedInventoryItems, ownedInventoryItems } = useInventory();
  const { user } = useAuth();
  const { apiClient } = useAPI();
  const [activeTab, setActiveTab] = useState<Tab>('buy');
  const [isPurchasing, setIsPurchasing] = useState<number | null>(null);
  const [isSelling, setIsSelling] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const shopInventory = action.metadata?.inventory || [];

  // Handle keyboard navigation
  useEffect(() => {
    const handleKeyPress = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      }
    };

    window.addEventListener('keydown', handleKeyPress);
    return () => {
      window.removeEventListener('keydown', handleKeyPress);
    };
  }, [onClose]);

  const handlePurchase = async (itemId: number, price: number) => {
    if (!user) {
      setError('You must be logged in to purchase items');
      return;
    }

    if (user.gold < price) {
      setError('Insufficient gold');
      return;
    }

    setIsPurchasing(itemId);
    setError(null);
    setSuccessMessage(null);

    try {
      const response = await apiClient.post<{
        user: { gold: number };
        itemId: number;
        quantity: number;
        totalPrice: number;
      }>(`/sonar/character-actions/${action.id}/purchase`, {
        itemId,
        quantity: 1,
      });

      // Refresh inventory
      await refreshOwnedInventoryItems();
      
      // User is updated from the response
      setSuccessMessage(`Purchased ${getInventoryItemById(itemId)?.name || 'item'} for ${price} gold!`);
      
      // Clear success message after 3 seconds
      setTimeout(() => {
        setSuccessMessage(null);
      }, 3000);
    } catch (err: any) {
      const errorMessage = err?.response?.data?.error || err?.message || 'Failed to purchase item';
      setError(errorMessage);
    } finally {
      setIsPurchasing(null);
    }
  };

  const handleSell = async (itemId: number, sellValue: number, quantity: number = 1) => {
    if (!user) {
      setError('You must be logged in to sell items');
      return;
    }

    setIsSelling(itemId);
    setError(null);
    setSuccessMessage(null);

    try {
      const response = await apiClient.post<{
        user: { gold: number };
        itemId: number;
        quantity: number;
        totalSellValue: number;
      }>(`/sonar/character-actions/${action.id}/sell`, {
        itemId,
        quantity,
      });

      // Refresh inventory
      await refreshOwnedInventoryItems();
      
      // User is updated from the response
      const totalValue = sellValue * quantity;
      setSuccessMessage(`Sold ${quantity}x ${getInventoryItemById(itemId)?.name || 'item'} for ${totalValue} gold!`);
      
      // Clear success message after 3 seconds
      setTimeout(() => {
        setSuccessMessage(null);
      }, 3000);
    } catch (err: any) {
      const errorMessage = err?.response?.data?.error || err?.message || 'Failed to sell item';
      setError(errorMessage);
    } finally {
      setIsSelling(null);
    }
  };

  // Get user's items that can be sold (have sellValue)
  const getSellableItems = () => {
    return ownedInventoryItems
      .map(ownedItem => {
        const item = getInventoryItemById(ownedItem.inventoryItemId);
        if (!item || !item.sellValue) return null;
        return { ownedItem, item };
      })
      .filter((entry): entry is { ownedItem: typeof ownedInventoryItems[0]; item: NonNullable<ReturnType<typeof getInventoryItemById>> } => entry !== null);
  };

  const characterImageUrl = character.dialogueImageUrl || character.mapIconUrl;

  return (
    <div className="Shop__overlay" onClick={(e) => e.target === e.currentTarget && onClose()}>
      <div className="Shop__container">
        {/* Close Button */}
        <button
          className="Shop__close"
          onClick={onClose}
          aria-label="Close shop"
        >
          ✕
        </button>

        {/* Character Image */}
        {characterImageUrl && (
          <div className="Shop__character-image-container">
            <img
              src={characterImageUrl}
              alt={character.name}
              className="Shop__character-image"
            />
          </div>
        )}

        {/* Shop Content */}
        <div className="Shop__content">
          <div className="Shop__header">
            <h2 className="Shop__title">{character.name}'s Shop</h2>
            {user && (
              <div className="Shop__gold-display">
                <div className="Shop__gold-icon">💰</div>
                <span className="Shop__gold-amount">{user.gold}</span>
              </div>
            )}
          </div>

          {/* Tabs */}
          <div className="Shop__tabs" style={{ display: 'flex', gap: '10px', marginBottom: '20px', borderBottom: '2px solid #ddd' }}>
            <button
              className={`Shop__tab ${activeTab === 'buy' ? 'Shop__tab--active' : ''}`}
              onClick={() => setActiveTab('buy')}
              style={{
                padding: '10px 20px',
                border: 'none',
                background: activeTab === 'buy' ? '#007bff' : 'transparent',
                color: activeTab === 'buy' ? 'white' : '#333',
                cursor: 'pointer',
                borderBottom: activeTab === 'buy' ? '2px solid #007bff' : '2px solid transparent',
                marginBottom: '-2px'
              }}
            >
              Buy
            </button>
            <button
              className={`Shop__tab ${activeTab === 'sell' ? 'Shop__tab--active' : ''}`}
              onClick={() => setActiveTab('sell')}
              style={{
                padding: '10px 20px',
                border: 'none',
                background: activeTab === 'sell' ? '#007bff' : 'transparent',
                color: activeTab === 'sell' ? 'white' : '#333',
                cursor: 'pointer',
                borderBottom: activeTab === 'sell' ? '2px solid #007bff' : '2px solid transparent',
                marginBottom: '-2px'
              }}
            >
              Sell
            </button>
          </div>

          {/* Messages */}
          {error && (
            <div className="Shop__error-message">
              {error}
            </div>
          )}
          {successMessage && (
            <div className="Shop__success-message">
              {successMessage}
            </div>
          )}

          {/* Buy Tab */}
          {activeTab === 'buy' && (
            <div className="Shop__inventory">
              {shopInventory.length === 0 ? (
                <div className="Shop__empty">
                  This shop has no items for sale.
                </div>
              ) : (
                shopInventory.map((shopItem) => {
                  const item = getInventoryItemById(shopItem.itemId);
                  const canAfford = user ? user.gold >= shopItem.price : false;
                  const isPurchasingItem = isPurchasing === shopItem.itemId;

                  if (!item) {
                    return null;
                  }

                  return (
                    <div key={shopItem.itemId} className="Shop__item">
                      <div className="Shop__item-image-container">
                        <img
                          src={item.imageUrl}
                          alt={item.name}
                          className="Shop__item-image"
                        />
                      </div>
                      <div className="Shop__item-details">
                        <h3 className="Shop__item-name">{item.name}</h3>
                        <p className="Shop__item-description">{item.flavorText}</p>
                        <div className="Shop__item-footer">
                          <span className="Shop__item-price">💰 {shopItem.price}</span>
                          <button
                            className={`Shop__buy-button ${!canAfford ? 'Shop__buy-button--disabled' : ''}`}
                            onClick={() => handlePurchase(shopItem.itemId, shopItem.price)}
                            disabled={!canAfford || isPurchasingItem}
                          >
                            {isPurchasingItem ? 'Purchasing...' : 'Buy'}
                          </button>
                        </div>
                      </div>
                    </div>
                  );
                })
              )}
            </div>
          )}

          {/* Sell Tab */}
          {activeTab === 'sell' && (
            <div className="Shop__inventory">
              {getSellableItems().length === 0 ? (
                <div className="Shop__empty">
                  You have no items that can be sold.
                </div>
              ) : (
                getSellableItems().map(({ ownedItem, item }) => {
                  const isSellingItem = isSelling === item.id;
                  const sellValue = item.sellValue || 0;

                  return (
                    <div key={item.id} className="Shop__item">
                      <div className="Shop__item-image-container">
                        <img
                          src={item.imageUrl}
                          alt={item.name}
                          className="Shop__item-image"
                        />
                      </div>
                      <div className="Shop__item-details">
                        <h3 className="Shop__item-name">{item.name}</h3>
                        <p className="Shop__item-description">{item.flavorText}</p>
                        <p style={{ margin: '5px 0', color: '#666', fontSize: '14px' }}>
                          Owned: {ownedItem.quantity}
                        </p>
                        <div className="Shop__item-footer" style={{ display: 'flex', alignItems: 'center', gap: '10px', flexWrap: 'wrap' }}>
                          <span className="Shop__item-price">💰 {sellValue} each</span>
                          <div style={{ display: 'flex', alignItems: 'center', gap: '5px' }}>
                            {ownedItem.quantity > 1 && (
                              <>
                                <button
                                  onClick={() => handleSell(item.id, sellValue, 1)}
                                  disabled={isSellingItem}
                                  style={{
                                    padding: '5px 10px',
                                    background: '#28a745',
                                    color: 'white',
                                    border: 'none',
                                    borderRadius: '4px',
                                    cursor: 'pointer'
                                  }}
                                >
                                  Sell 1
                                </button>
                                <button
                                  onClick={() => handleSell(item.id, sellValue, ownedItem.quantity)}
                                  disabled={isSellingItem}
                                  style={{
                                    padding: '5px 10px',
                                    background: '#28a745',
                                    color: 'white',
                                    border: 'none',
                                    borderRadius: '4px',
                                    cursor: 'pointer'
                                  }}
                                >
                                  Sell All ({ownedItem.quantity})
                                </button>
                              </>
                            )}
                            {ownedItem.quantity === 1 && (
                              <button
                                onClick={() => handleSell(item.id, sellValue, 1)}
                                disabled={isSellingItem}
                                style={{
                                  padding: '5px 10px',
                                  background: '#28a745',
                                  color: 'white',
                                  border: 'none',
                                  borderRadius: '4px',
                                  cursor: 'pointer'
                                }}
                              >
                                {isSellingItem ? 'Selling...' : 'Sell'}
                              </button>
                            )}
                          </div>
                        </div>
                      </div>
                    </div>
                  );
                })
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

