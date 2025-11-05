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

export const Shop: React.FC<ShopProps> = ({
  character,
  action,
  onClose,
}) => {
  const { inventoryItems, getInventoryItemById, refreshOwnedInventoryItems } = useInventory();
  const { user } = useAuth();
  const { apiClient } = useAPI();
  const [isPurchasing, setIsPurchasing] = useState<number | null>(null);
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
      
      // Refresh user by refetching from API if needed
      // The user object will be updated via the response

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

          {/* Shop Inventory */}
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
        </div>
      </div>
    </div>
  );
};

