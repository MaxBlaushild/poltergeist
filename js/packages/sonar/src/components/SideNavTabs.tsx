import React, { useState } from 'react';
import { Character } from './Character.tsx';
import { Friends } from './Friends.tsx';
import { FriendContextProvider } from '../contexts/FriendContext.tsx';
import { Party } from './Party.tsx';
import { useUserContext } from '../contexts/UserContext.tsx';

type TabType = 'character' | 'friends' | 'party';

export function SideNavTabs() {
  const [activeTab, setActiveTab] = useState<TabType>('character');
  const { user } = useUserContext();

  if (user) {
    return null;
  }

  return (
    <>
      {/* Tab Navigation */}
      <div className="px-4 mb-4">
        <div className="flex border-b border-gray-300">
          <button
            className={`flex-1 py-2 text-center font-medium transition-colors ${
              activeTab === 'character'
                ? 'border-b-2 border-blue-500 text-blue-500'
                : 'text-gray-600 hover:text-gray-800'
            }`}
            onClick={() => setActiveTab('character')}
          >
            Character
          </button>
          <button
            className={`flex-1 py-2 text-center font-medium transition-colors ${
              activeTab === 'party'
                ? 'border-b-2 border-blue-500 text-blue-500'
                : 'text-gray-600 hover:text-gray-800'
            }`}
            onClick={() => setActiveTab('party')}
          >
            Party
          </button>
          <button
            className={`flex-1 py-2 text-center font-medium transition-colors ${
              activeTab === 'friends'
                ? 'border-b-2 border-blue-500 text-blue-500'
                : 'text-gray-600 hover:text-gray-800'
            }`}
            onClick={() => setActiveTab('friends')}
          >
            Friends
          </button>

        </div>
      </div>

      {/* Tab Content */}
      <div className="px-4 mb-6 flex-1 overflow-y-auto">
        {activeTab === 'character' && (
          <div>
            <Character />
          </div>
        )}
        {activeTab === 'friends' && (
          <div>
            <FriendContextProvider>
                <Friends />
            </FriendContextProvider>
          </div>
        )}
        {activeTab === 'party' && (
          <div>
            <Party />
          </div>
        )}
      </div>
    </>
  );
}

