import React, { useState } from 'react';
import { useUsers } from '../hooks/useUsers.ts';
import { useInventory } from '@poltergeist/contexts';
import { useAPI } from '@poltergeist/contexts';

export const Armory = () => {
  const { users } = useUsers();
  const [selectedUser, setSelectedUser] = useState('');
  const [selectedItem, setSelectedItem] = useState('');
  const { inventoryItems } = useInventory();
  const [quantity, setQuantity] = useState(1);
  const { apiClient } = useAPI();

  const handleSubmit = async () => {
    try {
      await apiClient.post('/sonar/users/giveItem', {
        userID: selectedUser,
        itemID: parseInt(selectedItem),
        quantity: quantity
      });
      console.log('Item given successfully');
    } catch (error) {
      console.error('Failed to give item', error);
    }
  };

  return (
    <div className="p-4">
      <div className="flex flex-col gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700">Select User</label>
          <select
            value={selectedUser}
            onChange={(e) => setSelectedUser(e.target.value)}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          >
            <option value="">Choose a user...</option>
            {users?.map((user) => (
              <option key={user.id} value={user.id}>
                {user.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700">Select Item</label>
          <select
            value={selectedItem}
            onChange={(e) => setSelectedItem(e.target.value)}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          >
            <option value="">Choose an item...</option>
            {inventoryItems?.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700">Quantity</label>
          <input
            type="number"
            min="1"
            value={quantity}
            onChange={(e) => setQuantity(parseInt(e.target.value))}
            className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
          />
        </div>

        <button
          onClick={handleSubmit}
          className="inline-flex justify-center rounded-md border border-transparent bg-indigo-600 py-2 px-4 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
        >
          Submit
        </button>
      </div>
    </div>
  );
};

export default Armory;
