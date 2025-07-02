import React, { useState, useEffect } from 'react';
import { useUsers } from '../hooks/useUsers.ts';
import { useInventory } from '@poltergeist/contexts';
import { useAPI } from '@poltergeist/contexts';
import { Link } from 'react-router-dom';

interface InventoryItemWithStats {
  id: number;
  name: string;
  imageUrl: string;
  flavorText: string;
  effectText: string;
  rarityTier: string;
  isCaptureType: boolean;
  itemType: string;
  equipmentSlot?: string;
  stats?: {
    strengthBonus: number;
    dexterityBonus: number;
    constitutionBonus: number;
    intelligenceBonus: number;
    wisdomBonus: number;
    charismaBonus: number;
  };
}

export const Armory = () => {
  const { users } = useUsers();
  const [selectedUser, setSelectedUser] = useState('');
  const [selectedItem, setSelectedItem] = useState('');
  const { inventoryItems } = useInventory();
  const [quantity, setQuantity] = useState(1);
  const { apiClient } = useAPI();
  const [adminItems, setAdminItems] = useState<InventoryItemWithStats[]>([]);
  const [isLoadingAdminItems, setIsLoadingAdminItems] = useState(false);
  const [activeTab, setActiveTab] = useState<'give-items' | 'view-items'>('give-items');

  // Fetch admin items with stats
  const fetchAdminItems = async () => {
    setIsLoadingAdminItems(true);
    try {
      const response = await apiClient.get<InventoryItemWithStats[]>('/sonar/admin/items');
      setAdminItems(response);
    } catch (error) {
      console.error('Failed to fetch admin items:', error);
    } finally {
      setIsLoadingAdminItems(false);
    }
  };

  useEffect(() => {
    if (activeTab === 'view-items') {
      fetchAdminItems();
    }
  }, [activeTab]);

  const handleSubmit = async () => {
    try {
      await apiClient.post('/sonar/users/giveItem', {
        userID: selectedUser,
        itemID: parseInt(selectedItem),
        quantity: quantity
      });
      console.log('Item given successfully');
      // Reset form
      setSelectedUser('');
      setSelectedItem('');
      setQuantity(1);
    } catch (error) {
      console.error('Failed to give item', error);
    }
  };

  const getRarityColor = (rarity: string) => {
    switch (rarity) {
      case 'Common': return 'text-gray-600 bg-gray-100';
      case 'Uncommon': return 'text-green-600 bg-green-100';
      case 'Epic': return 'text-purple-600 bg-purple-100';
      case 'Mythic': return 'text-yellow-600 bg-yellow-100';
      case 'Not Droppable': return 'text-red-600 bg-red-100';
      default: return 'text-gray-600 bg-gray-100';
    }
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'passive': return 'text-blue-600 bg-blue-100';
      case 'consumable': return 'text-orange-600 bg-orange-100';
      case 'equippable': return 'text-indigo-600 bg-indigo-100';
      default: return 'text-gray-600 bg-gray-100';
    }
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900 mb-4">Armory</h1>
        
        {/* Tab Navigation */}
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => setActiveTab('give-items')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'give-items'
                  ? 'border-indigo-500 text-indigo-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Give Items to Users
            </button>
            <button
              onClick={() => setActiveTab('view-items')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'view-items'
                  ? 'border-indigo-500 text-indigo-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              View All Items
            </button>
          </nav>
        </div>
      </div>

      {activeTab === 'give-items' && (
        <div className="max-w-md">
          <div className="bg-white p-6 rounded-lg shadow">
            <h2 className="text-xl font-semibold mb-4">Give Item to User</h2>
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
                  onChange={(e) => setQuantity(parseInt(e.target.value) || 1)}
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                />
              </div>

              <button
                onClick={handleSubmit}
                disabled={!selectedUser || !selectedItem}
                className="inline-flex justify-center rounded-md border border-transparent bg-indigo-600 py-2 px-4 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Give Item
              </button>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'view-items' && (
        <div>
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-xl font-semibold">All Inventory Items</h2>
            <Link
              to="/inventory/create"
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              Create New Item
            </Link>
          </div>

          {isLoadingAdminItems ? (
            <div className="text-center py-8">
              <div className="text-gray-500">Loading items...</div>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {adminItems.map((item) => (
                <div key={item.id} className="bg-white rounded-lg shadow-md overflow-hidden">
                  <div className="h-48 bg-gray-200 flex items-center justify-center">
                    {item.imageUrl ? (
                      <img 
                        src={item.imageUrl} 
                        alt={item.name}
                        className="h-full w-full object-cover"
                        onError={(e) => {
                          const target = e.target as HTMLImageElement;
                          target.style.display = 'none';
                        }}
                      />
                    ) : (
                      <div className="text-gray-400">No Image</div>
                    )}
                  </div>
                  
                  <div className="p-4">
                    <div className="flex items-start justify-between mb-2">
                      <h3 className="text-lg font-semibold text-gray-900">{item.name}</h3>
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getRarityColor(item.rarityTier)}`}>
                        {item.rarityTier}
                      </span>
                    </div>
                    
                    <div className="flex items-center gap-2 mb-3">
                      <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${getTypeColor(item.itemType)}`}>
                        {item.itemType}
                      </span>
                      {item.equipmentSlot && (
                        <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium text-gray-600 bg-gray-100">
                          {item.equipmentSlot.replace('_', ' ')}
                        </span>
                      )}
                      {item.isCaptureType && (
                        <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium text-pink-600 bg-pink-100">
                          Capture
                        </span>
                      )}
                    </div>

                    {item.flavorText && (
                      <p className="text-sm text-gray-600 mb-2 italic">"{item.flavorText}"</p>
                    )}
                    
                    {item.effectText && (
                      <p className="text-sm text-gray-700 mb-3">{item.effectText}</p>
                    )}

                    {item.stats && Object.values(item.stats).some(val => val !== 0) && (
                      <div className="mt-3 pt-3 border-t border-gray-200">
                        <h4 className="text-sm font-medium text-gray-900 mb-2">Stat Bonuses:</h4>
                                                 <div className="grid grid-cols-2 gap-1 text-xs">
                           {Object.entries(item.stats).map(([stat, value]) => {
                             const numValue = value as number;
                             if (numValue === 0) return null;
                             const statName = stat.replace('Bonus', '');
                             return (
                               <div key={stat} className="flex justify-between">
                                 <span className="capitalize">{statName}:</span>
                                 <span className={numValue > 0 ? 'text-green-600' : 'text-red-600'}>
                                   {numValue > 0 ? '+' : ''}{numValue}
                                 </span>
                               </div>
                             );
                           })}
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}

          {!isLoadingAdminItems && adminItems.length === 0 && (
            <div className="text-center py-8">
              <div className="text-gray-500 mb-4">No items found</div>
              <Link
                to="/inventory/create"
                className="text-indigo-600 hover:text-indigo-500"
              >
                Create the first item
              </Link>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default Armory;
