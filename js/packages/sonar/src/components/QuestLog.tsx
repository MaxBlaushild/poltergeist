import React, { useState } from 'react';
import { useQuestLogContext } from '../contexts/QuestLogContext.tsx';
import { useSubmissionsContext } from '../contexts/SubmissionsContext.tsx';
import { TabItem, TabNav } from './shared/TabNav.tsx';
import { Quest, QuestNode, PointOfInterest, hasDiscoveredPointOfInterest } from '@poltergeist/types';
import { useMap } from '@poltergeist/contexts';
import { useDiscoveriesContext } from '../contexts/DiscoveriesContext.tsx';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { QuestComponent } from './Quest.tsx';

interface QuestLogProps {
  onClose: (pointOfInterest?: PointOfInterest | null) => void;
}

export const QuestLog: React.FC<QuestLogProps> = ({ onClose }) => {
  const { quests } = useQuestLogContext();
  const { submissions } = useSubmissionsContext();
  const { setLocation, flyToLocation } = useMap();
  const { discoveries } = useDiscoveriesContext();
  const [selectedQuest, setSelectedQuest] = useState<Quest | null>(null);

  const [expandedQuests, setExpandedQuests] = useState<{
    [key: string]: boolean;
  }>({});

  const toggleQuest = (quest: Quest) => {
    setSelectedQuest(quest);
  };

  const onPointOfInterestClick = (
    e: React.MouseEvent<HTMLDivElement>,
    node: QuestNode
  ) => {
    e.stopPropagation();
    onClose(node.pointOfInterest);
    flyToLocation(
      parseFloat(node.pointOfInterest.lat),
      parseFloat(node.pointOfInterest.lng)
    );
  };

  const inProgressQuests = quests.filter((quest) => !quest.isCompleted);
  const completedQuests = quests.filter((quest) => quest.isCompleted);

  if (selectedQuest) {
    return (
      <div className="flex flex-col gap-4 p-4 pt-0 w-full">
        <button 
          className="text-left font-bold mb-4"
          onClick={() => setSelectedQuest(null)}
        >
          ← Back to Quests
        </button>
        <QuestComponent quest={selectedQuest} onPointOfInterestClick={onPointOfInterestClick} />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 p-4 pt-0 w-full">
      <h2 className="text-2xl font-bold">Quest Log</h2>
      <TabNav tabs={['Quests', 'Completed']}>
        <TabItem key="Quests">
          <QuestList
            quests={inProgressQuests}
            toggleQuest={toggleQuest}
            expandedQuests={expandedQuests}
          />
        </TabItem>
        <TabItem key="Completed">
          <QuestList
            quests={completedQuests}
            toggleQuest={toggleQuest}
            expandedQuests={expandedQuests}
          />
        </TabItem>
      </TabNav>
    </div>
  );
};

type QuestListProps = {
  quests: Quest[];
  toggleQuest: (quest: Quest) => void;
  expandedQuests: { [key: string]: boolean };
};

const QuestList = ({
  quests,
  toggleQuest,
  expandedQuests,
}: QuestListProps) => {
  const { currentUser } = useUserProfiles();
  const { discoveries } = useDiscoveriesContext();

  return (
    <>
      {quests.filter((quest, index, self) => 
        index === self.findIndex((q) => q.rootNode.pointOfInterest.id === quest.rootNode.pointOfInterest.id)
      ).map((quest) => {
        const isDiscovered = hasDiscoveredPointOfInterest(
          quest.rootNode.pointOfInterest.id,
          currentUser?.id ?? '',
          discoveries
        );
        
        return (
          <div
            key={quest.rootNode.pointOfInterest.id}
            className="rounded p-2 w-full cursor-pointer hover:bg-gray-100"
            onClick={() => {
                toggleQuest(quest);
            }}
          >
            <div className="w-full flex items-center gap-2">
              <img
                src={quest.imageUrl}
                className="w-6 h-6 object-cover border-2 border-black"
                alt=""
              />
              <span>{quest.name}</span>
              {/* {!isDiscovered && <span className="text-sm text-gray-500 ml-2">(Undiscovered)</span>} */}
            </div>
          </div>
        );
      })}
    </>
  );
};
