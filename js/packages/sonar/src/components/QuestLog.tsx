import React, { useState } from 'react';
import { useQuestLogContext } from '../contexts/QuestLogContext.tsx';
import { useSubmissionsContext } from '../contexts/SubmissionsContext.tsx';
import { TabItem, TabNav } from './shared/TabNav.tsx';
import {
  Quest,
  QuestNode,
  PointOfInterest,
  hasDiscoveredPointOfInterest,
  getQuestTags,
  TagGroup,
} from '@poltergeist/types';
import { useMap, useTagContext } from '@poltergeist/contexts';
import { useDiscoveriesContext } from '../contexts/DiscoveriesContext.tsx';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { QuestComponent } from './Quest.tsx';
import { CheckCircleIcon, ChevronDownIcon } from '@heroicons/react/24/solid';
import { TagFilterComponent } from './TagFilter.tsx';

interface QuestLogProps {
  onClose: (pointOfInterest?: PointOfInterest | null) => void;
}

export const QuestLog: React.FC<QuestLogProps> = ({ onClose }) => {
  const { quests } = useQuestLogContext();
  const { submissions } = useSubmissionsContext();
  const { setLocation, flyToLocation } = useMap();
  const { discoveries } = useDiscoveriesContext();
  const { tagGroups } = useTagContext();
  const [selectedQuest, setSelectedQuest] = useState<Quest | null>(null);

  const [expandedGroups, setExpandedGroups] = useState<{
    [key: string]: boolean;
  }>({});

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

  const tagGroupBuckets = tagGroups.reduce((acc, tagGroup) => {
    acc[tagGroup.id] = [];
    return acc;
  }, {});

  quests.forEach((quest) => {
    const tags = getQuestTags(quest);
    // Find which tag group has the most matching tags for this quest
    let maxMatchingTags = 0;
    let bestMatchingGroup: TagGroup | null = null;

    tagGroups.forEach((tagGroup) => {
      const matchingTags = tags.filter((tag) =>
        tagGroup.tags.some((groupTag) => groupTag.name === tag)
      ).length;

      if (matchingTags > maxMatchingTags) {
        maxMatchingTags = matchingTags;
        bestMatchingGroup = tagGroup;
      }
    });

    if (bestMatchingGroup) {
      tagGroupBuckets[bestMatchingGroup.id].push(quest);
    }
  });
  if (selectedQuest) {
    return (
      <div className="flex flex-col gap-4 p-4 pt-0 w-full">
        <button
          className="text-left font-bold mb-4"
          onClick={() => setSelectedQuest(null)}
        >
          ← Back to Quests
        </button>
        <QuestComponent
          quest={selectedQuest}
          onPointOfInterestClick={onPointOfInterestClick}
        />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 p-4 pt-0 w-full">
      <h2 className="text-2xl font-bold">Quest Log</h2>
      <TabNav tabs={['Quests', 'Filters']}>
        <TabItem key="Quests">
          <div className="flex flex-col gap-4">
            {tagGroups.map((tagGroup) => {
              const groupQuests = tagGroupBuckets[tagGroup.id];
              if (groupQuests.length === 0) return null;

              return (
                <div key={tagGroup.id} className="bg-white/90 border-2 border-gray-200 rounded-xl shadow-lg backdrop-blur-sm">
                  <button
                    className="w-full flex justify-between items-center p-4 hover:bg-gray-50/80 rounded-t-xl transition-colors"
                    onClick={() => {
                      setExpandedGroups((prev) => ({
                        ...prev,
                        [tagGroup.id]: !prev[tagGroup.id],
                      }));
                    }}
                  >
                    <div className="flex items-center gap-2">
                      <img
                        src={tagGroup.iconUrl}
                        alt={tagGroup.name}
                        className="w-6 h-6 object-cover rounded shadow-sm"
                      />
                      <span className="font-semibold text-gray-800">
                        {tagGroup.name.charAt(0).toUpperCase() +
                          tagGroup.name.slice(1).toLowerCase()}
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-gray-600 font-medium">
                        ({groupQuests.length})
                      </span>
                      <ChevronDownIcon
                        className={`w-5 h-5 transition-transform text-gray-600 ${
                          expandedGroups[tagGroup.id] ? 'rotate-180' : ''
                        }`}
                      />
                    </div>
                  </button>
                  {expandedGroups[tagGroup.id] && (
                    <div className="p-4 pt-0">
                      {groupQuests.map((quest) => (
                        <div
                          key={quest.id}
                          className="py-4 cursor-pointer hover:bg-gray-50/80 rounded-lg transition-colors px-3"
                          onClick={() => setSelectedQuest(quest)}
                        >
                          <div className="flex items-center gap-4">
                            <div className="flex items-center gap-2">
                              {quest.isCompleted ? (
                                <CheckCircleIcon className="w-5 h-5 text-green-500 drop-shadow-sm" />
                              ) : (
                                <CheckCircleIcon className="w-5 h-5 text-gray-300" />
                              )}
                              <h3 className="font-medium text-gray-800">{quest.name}</h3>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </TabItem>
        <TabItem key="Filters">
          <TagFilterComponent />
        </TabItem>
      </TabNav>
    </div>
  );
};
