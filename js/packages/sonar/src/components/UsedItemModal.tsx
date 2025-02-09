import React, { useEffect } from 'react';
import { useInventory } from '@poltergeist/contexts';
import { XMarkIcon } from '@heroicons/react/20/solid';
import './ItemModal.css';

export const UsedItemModal = () => {
  const { usedItem, setUsedItem } = useInventory();
  const [fadeOut, setFadeOut] = React.useState(false);
  useEffect(() => {
    if (usedItem) {
      setFadeOut(true);
      setTimeout(() => {
        setUsedItem(null);
        setFadeOut(false);
      }, 3000);
    }
  }, [usedItem, setUsedItem]);

  return usedItem ? (
    <div className="ItemModal flex flex-col items-center z-10 gap-4">
      <h1 className="text-xl font-bold"> You used a {usedItem.name}!</h1>
      <img
        className={`w-3/4 h-3/4 border-2 border-black rounded-lg`}
        src={usedItem.imageUrl}
        alt={usedItem.name}
        style={{ transition: 'opacity 3s' }}
      />
    </div>
  ) : null;
};

export default UsedItemModal;
