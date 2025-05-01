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

interface QuestAccordionProps {
  title: string;
  iconUrl?: string;
  quests: Quest[];
  onQuestClick: (quest: Quest) => void;
  isExpanded: boolean;
  onToggleExpand: () => void;
}

const QuestAccordion: React.FC<QuestAccordionProps> = ({
  title,
  iconUrl,
  quests,
  onQuestClick,
  isExpanded,
  onToggleExpand,
}) => {
  return (
    <div className="bg-white/90 border-2 border-gray-200 rounded-xl shadow-lg backdrop-blur-sm">
      <button
        className="w-full flex justify-between items-center p-4 hover:bg-gray-50/80 rounded-t-xl transition-colors"
        onClick={onToggleExpand}
      >
        <div className="flex items-center gap-2">
          {iconUrl ? (
            <img
              src={iconUrl}
              alt={title}
              className="w-6 h-6 object-cover rounded shadow-sm"
            />
          ) : title === "Tracked Quests" && (
            <span className="text-xl">⭐</span>
          )}
          <span className="font-semibold text-gray-800">
            {title.charAt(0).toUpperCase() + title.slice(1).toLowerCase()}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-gray-600 font-medium">({quests.length})</span>
          <ChevronDownIcon
            className={`w-5 h-5 transition-transform text-gray-600 ${
              isExpanded ? 'rotate-180' : ''
            }`}
          />
        </div>
      </button>
      {isExpanded && (
        <div className="p-4 pt-0">
          {quests.map((quest) => (
            <div
              key={quest.id}
              className="py-4 cursor-pointer hover:bg-gray-50/80 rounded-lg transition-colors px-3"
              onClick={() => onQuestClick(quest)}
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
};

interface QuestLogProps {
  onClose: (pointOfInterest?: PointOfInterest | null) => void;
}

export const QuestLog: React.FC<QuestLogProps> = ({ onClose }) => {
  const { quests, trackedQuestIds } = useQuestLogContext();
  const { submissions } = useSubmissionsContext();
  const { setLocation, flyToLocation } = useMap();
  const { discoveries } = useDiscoveriesContext();
  const { tagGroups } = useTagContext();
  const [selectedQuest, setSelectedQuest] = useState<Quest | null>(null);

  const [expandedGroups, setExpandedGroups] = useState<{
    [key: string]: boolean;
  }>({});

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

  const trackedQuests: Quest[] = [];
  const untaggedQuests: Quest[] = [];

  console.log(quests);

  quests.forEach((quest) => {
    const isTracked = trackedQuestIds.includes(quest.id);
    if (isTracked) {
      trackedQuests.push(quest);
      return;
    }
    const tags = getQuestTags(quest);
    
    if (tags.length === 0) {
      untaggedQuests.push(quest);
      return;
    }

    // Put quest in every bucket where it has at least one matching tag
    let addedToBucket = false;
    tagGroups.forEach((tagGroup) => {
      const hasMatchingTag = tags.some((tag) =>
        tagGroup.tags.some((groupTag) => groupTag.name === tag)
      );

      if (hasMatchingTag) {
        tagGroupBuckets[tagGroup.id].push(quest);
        addedToBucket = true;
      }
    });

    // If quest wasn't added to any buckets, put it in untagged
    if (!addedToBucket) {
      untaggedQuests.push(quest);
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

  console.log(tagGroups);
  Object.entries(tagGroupBuckets).forEach(([groupId, quests]) => {
    console.log(`Tag group ${groupId}:`, quests);
  });

  return (
    <div className="flex flex-col gap-4 p-4 pt-0 w-full">
      <h2 className="text-2xl font-bold">Quest Log</h2>
      <TabNav tabs={['Quests', 'Filters']}>
        <TabItem key="Quests">
          <div className="flex flex-col gap-4">
            {trackedQuests.length > 0 && (
              <QuestAccordion
                title="Tracked Quests"
                quests={trackedQuests}
                onQuestClick={setSelectedQuest}
                isExpanded={expandedGroups['tracked'] ?? false}
                onToggleExpand={() =>
                  setExpandedGroups((prev) => ({
                    ...prev,
                    tracked: !prev['tracked'],
                  }))
                }
              />
            )}
            {tagGroups.map((tagGroup) => {
              const groupQuests = tagGroupBuckets[tagGroup.id];
              if (groupQuests.length === 0) return null;

              return (
                <QuestAccordion
                  key={tagGroup.id}
                  title={tagGroup.name}
                  iconUrl={tagGroup.iconUrl}
                  quests={groupQuests}
                  onQuestClick={setSelectedQuest}
                  isExpanded={expandedGroups[tagGroup.id] ?? false}
                  onToggleExpand={() =>
                    setExpandedGroups((prev) => ({
                      ...prev,
                      [tagGroup.id]: !prev[tagGroup.id],
                    }))
                  }
                />
              );
            })}
            {untaggedQuests.length > 0 && (
              <QuestAccordion
                title="The Rest"
                quests={untaggedQuests}
                onQuestClick={setSelectedQuest}
                isExpanded={expandedGroups['untagged'] ?? false}
                onToggleExpand={() =>
                  setExpandedGroups((prev) => ({
                    ...prev,
                    untagged: !prev['untagged'],
                  }))
                }
              />
            )}
          </div>
        </TabItem>
        <TabItem key="Filters">
          <TagFilterComponent />
        </TabItem>
      </TabNav>
    </div>
  );
};
