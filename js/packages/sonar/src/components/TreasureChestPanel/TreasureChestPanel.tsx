import React, { useState } from 'react';
import { TreasureChest, OwnedInventoryItem, InventoryItem } from '@poltergeist/types';
import { useInventory, useLocation, useAPI } from '@poltergeist/contexts';
import { useUserProfiles } from '../../contexts/UserProfileContext.tsx';
import { Button } from '../shared/Button.tsx';

interface TreasureChestPanelProps {
  treasureChest: TreasureChest;
  onClose: (immediate: boolean) => void;
}

export const TreasureChestPanel = ({
  treasureChest,
  onClose,
}: TreasureChestPanelProps) => {
  const { currentUser } = useUserProfiles();
  const { inventoryItems, ownedInventoryItems, refreshOwnedInventoryItems } = useInventory();
  const { location } = useLocation();
  const { apiClient } = useAPI();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const imageUrl = 'https://crew-points-of-interest.s3.amazonaws.com/inventory-items/1762314753387-0gdf0170kq5m.png';

  // Calculate distance between user and chest (10m threshold)
  const calculateDistance = (lat1: number, lon1: number, lat2: number, lon2: number) => {
    const R = 6371e3; // Earth's radius in meters
    const φ1 = lat1 * Math.PI / 180;
    const φ2 = lat2 * Math.PI / 180;
    const Δφ = (lat2 - lat1) * Math.PI / 180;
    const Δλ = (lon2 - lon1) * Math.PI / 180;

    const a = Math.sin(Δφ/2) * Math.sin(Δφ/2) +
              Math.cos(φ1) * Math.cos(φ2) *
              Math.sin(Δλ/2) * Math.sin(Δλ/2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a));

    return R * c; // Distance in meters
  };

  const distanceInfo = location?.latitude && location?.longitude ? (() => {
    const distance = calculateDistance(
      location.latitude,
      location.longitude,
      treasureChest.latitude,
      treasureChest.longitude
    );

    return {
      distance: Math.round(distance),
      isWithinRange: distance <= 10
    };
  })() : null;

  const isWithinRange = distanceInfo?.isWithinRange ?? false;

  // Check if user has unlock item
  const hasUnlockItem = treasureChest.unlockTier != null ? (() => {
    for (const ownedItem of ownedInventoryItems) {
      if (ownedItem.quantity > 0) {
        const inventoryItem = inventoryItems.find(item => item.id === ownedItem.inventoryItemId);
        if (inventoryItem?.unlockTier != null && inventoryItem.unlockTier >= treasureChest.unlockTier) {
          return true;
        }
      }
    }
    return false;
  })() : true;

  // Determine button state
  const getButtonState = () => {
    if (treasureChest.openedByUser) {
      return { text: 'Already opened', disabled: true };
    }
    if (!isWithinRange) {
      return { text: 'Too far away', disabled: true };
    }
    if (treasureChest.unlockTier == null) {
      return { text: 'Unlocked', disabled: false };
    }
    if (hasUnlockItem) {
      return { text: 'Unlock', disabled: false };
    }
    return { text: 'Locked', disabled: true };
  };

  const buttonState = getButtonState();

  const handleOpenChest = async () => {
    if (buttonState.disabled) {
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await apiClient.post(`/sonar/treasure-chests/${treasureChest.id}/open`);
      
      // Refresh inventory
      await refreshOwnedInventoryItems();
      
      // Refresh user profile if needed
      if (response.user) {
        // The user data might be updated elsewhere, but we can trigger a refresh
      }

      // Close panel after successful open
      setTimeout(() => {
        onClose(true);
      }, 1000);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to open treasure chest');
      setIsLoading(false);
    }
  };

  return (
    <div className="flex flex-col items-center gap-4">
      <div className="flex flex-col items-center gap-2">
        <h3 className="text-2xl font-bold">Treasure Chest</h3>
      </div>
      <img
        src={imageUrl}
        alt="Treasure Chest"
        className="w-32 h-32 object-contain"
      />
      {error && (
        <div className="text-red-500 text-sm">{error}</div>
      )}
      {distanceInfo && (
        <div className="text-sm text-gray-600">
          Distance: {distanceInfo.distance}m
        </div>
      )}
      {treasureChest.gold != null && treasureChest.gold > 0 && (
        <div className="text-sm text-gray-600">
          Gold: {treasureChest.gold}
        </div>
      )}
      {treasureChest.items.length > 0 && (
        <div className="text-sm text-gray-600">
          Items: {treasureChest.items.length}
        </div>
      )}
      {treasureChest.unlockTier != null && (
        <div className="text-sm text-gray-600">
          Requires unlock tier: {treasureChest.unlockTier}
        </div>
      )}
      <Button
        onClick={handleOpenChest}
        title={buttonState.text}
        disabled={buttonState.disabled || isLoading}
      />
    </div>
  );
};

