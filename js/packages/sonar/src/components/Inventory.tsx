import React, { useEffect, useState } from 'react';
import { useInventory } from '@poltergeist/contexts';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import {
  InventoryItem,
  ItemsRequiringTeamId,
  ItemsUsabledInMenu,
  Match,
  OwnedInventoryItem,
  Team,
} from '@poltergeist/types';
import Divider from './shared/Divider.tsx';
import { ArrowLeftIcon } from '@heroicons/react/20/solid';
import { Button, ButtonColor, ButtonSize } from './shared/Button.tsx';
import { generateColorFromTeamName } from '../utils/generateColor.ts';

type InventoryProps = {
  onClose: () => void;
  match?: Match | undefined;
  usersTeam?: Team | undefined;
}

export const Inventory = ({ onClose, match, usersTeam }: InventoryProps) => {
  const { inventoryItems, ownedInventoryItems, ownedInventoryItemsAreLoading, consumeItem, setUsedItem } =
    useInventory();
  const [selectedItem, setSelectedItem] = useState<
    OwnedInventoryItem | undefined
  >(undefined);
  const [itemBeingUsed, setItemBeingUsed] = useState<
    OwnedInventoryItem | undefined
  >(undefined);
  const [activeTab, setActiveTab] = useState<'inventory' | 'equipped'>('inventory');
  
  const selectedInventoryItem = inventoryItems?.find(
    (i) => i.id === selectedItem?.inventoryItemId
  );

  // Mock equipped items - you'll need to replace this with actual data from your backend
  const [equippedItems, setEquippedItems] = useState<{
    [slot: string]: OwnedInventoryItem | null;
  }>({
    head: null,
    chest: null,
    left_hand: null,
    right_hand: null,
    feet: null,
    gloves: null,
    neck: null,
    left_ring: null,
    right_ring: null,
    legs: null,
  });

  const equipItem = (item: OwnedInventoryItem, slot: string) => {
    // Unequip any existing item in that slot
    if (equippedItems[slot]) {
      // Add the unequipped item back to inventory
      console.log(`Unequipped ${equippedItems[slot]?.inventoryItemId} from ${slot}`);
    }
    
    // Equip the new item
    setEquippedItems(prev => ({
      ...prev,
      [slot]: item
    }));
    
    // Remove from inventory (you'll need to implement this with your backend)
    console.log(`Equipped ${item.inventoryItemId} to ${slot}`);
  };

  const unequipItem = (slot: string) => {
    const item = equippedItems[slot];
    if (item) {
      setEquippedItems(prev => ({
        ...prev,
        [slot]: null
      }));
      
      // Add back to inventory (you'll need to implement this with your backend)
      console.log(`Unequipped ${item.inventoryItemId} from ${slot}`);
    }
  };

  const getEquipmentSlotName = (slot: string) => {
    const slotNames: { [key: string]: string } = {
      head: 'Head',
      chest: 'Chest',
      left_hand: 'Left Hand',
      right_hand: 'Right Hand',
      feet: 'Feet',
      gloves: 'Gloves',
      neck: 'Neck',
      left_ring: 'Left Ring',
      right_ring: 'Right Ring',
      legs: 'Legs',
    };
    return slotNames[slot] || slot;
  };

  const renderEquipmentSlot = (slot: string) => {
    const equippedItem = equippedItems[slot];
    const slotName = getEquipmentSlotName(slot);
    
    return (
      <div key={slot} className="flex flex-col items-center gap-2">
        <div className="text-sm font-medium text-gray-600">{slotName}</div>
        <div className="aspect-square w-16 h-16 rounded-lg border-2 border-gray-300 flex justify-center items-center relative overflow-hidden bg-gray-100">
          {equippedItem ? (
            <>
              <img
                src={inventoryItems?.find(i => i.id === equippedItem.inventoryItemId)?.imageUrl}
                alt="Equipped item"
                className="max-h-full max-w-full"
              />
              <Button
                buttonSize={ButtonSize.SMALL}
                title="Unequip"
                onClick={() => unequipItem(slot)}
                className="absolute top-0 right-0 bg-red-500 text-white text-xs px-1 py-0.5 rounded"
              />
            </>
          ) : (
            <div className="text-gray-400 text-xs text-center">Empty</div>
          )}
        </div>
      </div>
    );
  };

  return (
    <div className="flex flex-col gap-4 w-full">
      {selectedItem || itemBeingUsed ? (
        <ArrowLeftIcon
          className="w-8 h-8 absolute left-4"
          onClick={() => {
            setSelectedItem(undefined);
            setItemBeingUsed(undefined);
          }}
        />
      ) : null}
      
      <h2 className="text-2xl font-bold">
        {selectedInventoryItem ? selectedInventoryItem?.name : 'Inventory'}
      </h2>

      {/* Tab Navigation */}
      {!selectedItem && !itemBeingUsed && (
        <div className="flex border-b border-gray-200">
          <button
            className={`px-4 py-2 font-medium ${
              activeTab === 'inventory'
                ? 'text-blue-600 border-b-2 border-blue-600'
                : 'text-gray-500 hover:text-gray-700'
            }`}
            onClick={() => setActiveTab('inventory')}
          >
            Inventory
          </button>
          <button
            className={`px-4 py-2 font-medium ${
              activeTab === 'equipped'
                ? 'text-blue-600 border-b-2 border-blue-600'
                : 'text-gray-500 hover:text-gray-700'
            }`}
            onClick={() => setActiveTab('equipped')}
          >
            Equipped
          </button>
        </div>
      )}

      {!selectedItem && !itemBeingUsed ? (
        activeTab === 'inventory' ? (
          // Inventory Tab
          <div className="grid grid-cols-3 gap-2 mt-4">
            {Array.from({ length: 12 }).map((_, index) => {
              const item = ownedInventoryItems?.[index];
              const inventoryItem = item
                ? inventoryItems.find((i) => i.id === item.inventoryItemId)
                : null;

              return (
                <div
                  key={index}
                  className="aspect-square rounded-lg border border-black flex justify-center items-center relative overflow-hidden"
                >
                  {inventoryItem ? (
                    <>
                      <img
                        onClick={() => setSelectedItem(item)}
                        src={inventoryItem.imageUrl}
                        alt={inventoryItem.name}
                        className="max-h-full max-w-full cursor-pointer"
                      />
                      <span className="absolute bottom-1 right-1 text-white bg-black px-1 rounded">
                        {item?.quantity}
                      </span>
                      {/* Show equip button for equippable items */}
                      {inventoryItem.equipmentSlot && (
                        <Button
                          buttonSize={ButtonSize.SMALL}
                          title="Equip"
                          onClick={(e) => {
                            e.stopPropagation();
                            equipItem(item, inventoryItem.equipmentSlot!);
                          }}
                          className="absolute top-1 left-1 bg-green-500 text-white text-xs px-1 py-0.5 rounded"
                        />
                      )}
                    </>
                  ) : null}
                </div>
              );
            })}
          </div>
        ) : (
          // Equipped Tab
          <div className="mt-4">
            <div className="grid grid-cols-5 gap-4">
              {Object.keys(equippedItems).map(slot => renderEquipmentSlot(slot))}
            </div>
            
            {/* Equipment Stats Summary */}
            <div className="mt-6 p-4 bg-gray-50 rounded-lg">
              <h3 className="text-lg font-semibold mb-3">Equipment Bonuses</h3>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <div className="font-medium">Combat</div>
                  <div>Attack: +0</div>
                  <div>Defense: +0</div>
                  <div>Health: +0</div>
                </div>
                <div>
                  <div className="font-medium">Attributes</div>
                  <div>Strength: +0</div>
                  <div>Agility: +0</div>
                  <div>Intelligence: +0</div>
                </div>
              </div>
            </div>
          </div>
        )
      ) : selectedItem && !itemBeingUsed ? (
        <div className="flex flex-col gap-4">
          <img
            src={selectedInventoryItem?.imageUrl}
            alt={selectedInventoryItem?.name}
            className="max-h-full max-w-full rounded-lg border border-black flex"
          />
          <div className="flex flex-col gap-2">
            <p className="text-lg font-bold">
              {selectedInventoryItem?.flavorText}
            </p>
            <p className="text-lg">{selectedInventoryItem?.effectText}</p>
            
            {/* Show equipment slot if it's an equippable item */}
            {selectedInventoryItem?.equipmentSlot && (
              <div className="text-sm text-gray-600">
                Equipment Slot: {getEquipmentSlotName(selectedInventoryItem.equipmentSlot)}
              </div>
            )}
            
            {/* Show item stats if available */}
            {selectedInventoryItem && (
              <div className="text-sm space-y-1">
                {selectedInventoryItem.plusStrength && <div>Strength: +{selectedInventoryItem.plusStrength}</div>}
                {selectedInventoryItem.plusAgility && <div>Agility: +{selectedInventoryItem.plusAgility}</div>}
                {selectedInventoryItem.plusIntelligence && <div>Intelligence: +{selectedInventoryItem.plusIntelligence}</div>}
                {selectedInventoryItem.defense && <div>Defense: +{selectedInventoryItem.defense}</div>}
                {selectedInventoryItem.weight && <div>Weight: {selectedInventoryItem.weight}</div>}
                {selectedInventoryItem.value && <div>Value: {selectedInventoryItem.value}</div>}
              </div>
            )}
            
            <div className="flex gap-2">
              {/* Equip button for equippable items */}
              {selectedInventoryItem?.equipmentSlot && (
                <Button
                  title="Equip"
                  onClick={() => {
                    equipItem(selectedItem, selectedInventoryItem.equipmentSlot!);
                    setSelectedItem(undefined);
                  }}
                />
              )}
              
              {/* Use button for consumable items */}
              {ItemsUsabledInMenu.includes(selectedInventoryItem!.id) ? (
                <Button
                  title="Use"
                  disabled={
                    !selectedItem?.quantity || selectedItem?.quantity === 0
                  }
                  onClick={() => {
                    if (
                      ItemsRequiringTeamId.includes(selectedItem.inventoryItemId)
                    ) {
                      setItemBeingUsed(selectedItem);
                    } else {
                      consumeItem(selectedItem!.id);
                      setUsedItem(selectedInventoryItem!);
                      onClose();
                    }
                  }}
                />
              ) : null}
            </div>
          </div>
        </div>
      ) : (
        <div>
          <table className="w-full mt-4">
            <thead>
              <tr>
                <th className="text-left">Team</th>
                <th className="text-center">Color</th>
                <th className="text-center">Action</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <br />
              </tr>
              {match?.teams
                .filter((t) => t.id !== usersTeam?.id)
                .map((team) => (
                  <tr key={team.id}>
                    <td className="text-left text-lg font-bold">{team.name}</td>
                    <td className="text-center">
                      <div
                        style={{
                          width: '32px',
                          height: '32px',
                          backgroundColor: generateColorFromTeamName(team.name),
                          borderRadius: '50%',
                          margin: 'auto',
                        }}
                      />
                    </td>
                    <td className="text-center text-xl font-bold p-2">
                      <Button
                        buttonSize={ButtonSize.SMALL}
                        title="Choose"
                        onClick={() => {
                          consumeItem(selectedItem!.id, {
                            targetTeamId: team.id,
                          });
                          setUsedItem(selectedInventoryItem!);
                          onClose();
                        }}
                      />
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};
