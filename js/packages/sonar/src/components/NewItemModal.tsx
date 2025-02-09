import React from 'react';
import { useInventory } from '@poltergeist/contexts';
import { XMarkIcon } from '@heroicons/react/20/solid';
import "./ItemModal.css"

const NewItemModal = () => {
  const { presentedInventoryItem, setPresentedInventoryItem, inventoryItems } = useInventory();
  return presentedInventoryItem ? (
    <div className="ItemModal flex flex-col items-center z-10 gap-4">
      <div className="w-full flex item-start justify-start">
          <XMarkIcon 
          className="w-6 h-6"
          onClick={() => setPresentedInventoryItem(null)}
          />
      </div>
      <h1 className="text-xl font-bold"> You got a {presentedInventoryItem.name}!</h1>
      <img
        className="w-3/4 h-3/4 border-2 border-black rounded-lg"
        src={presentedInventoryItem.imageUrl}
        alt={presentedInventoryItem.name}
      />
      <p className="text-sm font-bold">{presentedInventoryItem.flavorText}</p>
      <p className="text-sm">
        {presentedInventoryItem.effectText}
      </p>
    </div>
  ) : null;
};

export default NewItemModal;
