import React from 'react';
import {
  PointOfInterestGroup,
  QuestNode,
  QuestLog,
  Quest,
  hasDiscoveredPointOfInterest,
} from '@poltergeist/types';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { useDiscoveriesContext } from '../contexts/DiscoveriesContext.tsx';
import { TabItem } from './shared/TabNav.tsx';
import { TabNav } from './shared/TabNav.tsx';
import { useQuestLogContext } from '../contexts/QuestLogContext.tsx';

interface QuestNodeProps {
  node: QuestNode;
  onPointOfInterestClick: (e: React.MouseEvent<HTMLDivElement>, node: QuestNode) => void;
  discoveredPointsOfInterestIds: { [key: string]: boolean };
  darkMode?: boolean;
}

export const QuestNodeComponent = ({ node, onPointOfInterestClick, discoveredPointsOfInterestIds, darkMode = false }: QuestNodeProps) => {
  return (
    <div
      key={node.pointOfInterest.id}
      className="mb-2"
      onClick={(e) => onPointOfInterestClick(e, node)}
    >
      <div className="font-bold flex items-center gap-2">
        <img
          src={
            discoveredPointsOfInterestIds[node.pointOfInterest.id]
              ? node.pointOfInterest.imageURL
              : 'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp'
          }
          className="w-6 h-6 object-cover border-2 border-black"
          alt=""
        />
        <div className="flex flex-col">
          <span className={darkMode ? 'text-white' : 'text-black'}>{node.pointOfInterest.name}</span>
          {node.objectives.map((objective) => (
            <div
              key={objective.challenge.id}
              className={`text-sm ${darkMode ? 'text-white' : 'text-gray-600'} font-medium`}
            >
              <span>
                {objective.challenge.question}
                {objective.isCompleted && ' ✅'}
              </span>
            </div>
          ))}
        </div>
      </div>
      {node.objectives.map((objective) => {
        if (
          objective.nextNode &&
          objective.isCompleted
        ) {
          return <QuestNodeComponent 
            key={objective.nextNode.pointOfInterest.id}
            node={objective.nextNode} 
            onPointOfInterestClick={onPointOfInterestClick}
            discoveredPointsOfInterestIds={discoveredPointsOfInterestIds}
            darkMode={darkMode}
          />;
        }
        return null;
      })}
    </div>
  );
};

interface QuestProps {
  quest: Quest;
  onPointOfInterestClick: (
    e: React.MouseEvent<HTMLDivElement>,
    node: QuestNode
  ) => void;
}

export const QuestComponent = ({
  quest,
  onPointOfInterestClick,
}: QuestProps) => {
  const { currentUser } = useUserProfiles();
  const { discoveries } = useDiscoveriesContext();
  const { trackQuest, untrackQuest, trackedQuestIds } = useQuestLogContext();
  const isTracked = trackedQuestIds.includes(quest.id);

  const discovedPointsOfInterestIds = discoveries?.filter((discovery) => discovery.userId === currentUser?.id).reduce((acc, discovery) => {
    acc[discovery.pointOfInterestId] = true;
    return acc;
  }, {});

  return (
    <div className="flex flex-col items-center gap-4 font-medium">
      <h2 className="text-2xl font-extrabold">{quest.name}</h2>

      <img
        src={quest.imageUrl}
        alt={quest.name}
        className="w-full h-48 object-cover rounded-lg"
      />
      <div className="w-full text-left text-gray-700">
        {(() => {
          let completedQuests = 0;
          let totalQuests = 0;

          const countQuests = (node: QuestNode) => {
            totalQuests++;
            if (node.objectives.every((obj) => obj.isCompleted)) {
              completedQuests++;
            }
           
            node.objectives.forEach((objective) => {
              if (objective.nextNode) {
                countQuests(objective.nextNode);
              }
            });
          };

          countQuests(quest.rootNode);

          return (
            <div className="flex items-center justify-between gap-2">
              <div className="flex items-center gap-2">
                <span className="font-medium">Tasks completed:</span>
                <span className="font-bold">
                  {completedQuests}/{totalQuests}
                </span>
              </div>
              <button
                onClick={() => {
                  if (isTracked) {
                    untrackQuest(quest.id);
                  } else {
                    trackQuest(quest.id);
                  }
                }}
                className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
              >
                {isTracked ? 'Untrack Quest' : 'Track Quest'}
              </button>
            </div>
          );
        })()}
      </div>
      <TabNav tabs={['Description', 'Tasks']}>
        <TabItem key="Description">{quest.description}</TabItem>
        <TabItem key="Tasks">
          <div className="ml-6 mt-2 w-full text-left">
            <QuestNodeComponent
              node={quest.rootNode}
              onPointOfInterestClick={onPointOfInterestClick}
              discoveredPointsOfInterestIds={discovedPointsOfInterestIds}
            />
          </div>
        </TabItem>
      </TabNav>
    </div>
  );
};
