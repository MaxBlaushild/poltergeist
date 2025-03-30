import React, { useState } from 'react';
import { useQuestLogContext } from '../contexts/QuestLogContext.tsx';
import { useSubmissionsContext } from '../contexts/SubmissionsContext.tsx';
import { TabItem, TabNav } from './shared/TabNav.tsx';
import { Quest, QuestNode, PointOfInterest, hasDiscoveredPointOfInterest } from '@poltergeist/types';
import { useMap } from '@poltergeist/contexts';
import { useDiscoveriesContext } from '../contexts/DiscoveriesContext.tsx';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';

interface QuestLogProps {
  onClose: (pointOfInterest?: PointOfInterest | null) => void;
}

export const QuestLog: React.FC<QuestLogProps> = ({ onClose }) => {
  const { quests } = useQuestLogContext();
  const { submissions } = useSubmissionsContext();
  const { setLocation, flyToLocation } = useMap();
  const { discoveries } = useDiscoveriesContext();

  const [expandedQuests, setExpandedQuests] = useState<{
    [key: string]: boolean;
  }>({});

  const toggleQuest = (questId: string) => {
    setExpandedQuests((prev) => ({
      ...prev,
      [questId]: !prev[questId],
    }));
  };

  const onClickQuest = (
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


  return (
    <div className="flex flex-col gap-4 p-4 pt-0 w-full">
      <h2 className="text-2xl font-bold">Quest Log</h2>
      <TabNav tabs={['Quests', 'Completed']}>
        <TabItem key="Quests">
          <QuestList
            quests={inProgressQuests}
            toggleQuest={toggleQuest}
            expandedQuests={expandedQuests}
            onClickQuest={onClickQuest}
          />
        </TabItem>
        <TabItem key="Completed">
          <QuestList
            quests={completedQuests}
            toggleQuest={toggleQuest}
            expandedQuests={expandedQuests}
            onClickQuest={onClickQuest}
          />
        </TabItem>
      </TabNav>
    </div>
  );
};

type QuestListProps = {
  quests: Quest[];
  toggleQuest: (questId: string) => void;
  expandedQuests: { [key: string]: boolean };
  onClickQuest: (e: React.MouseEvent<HTMLDivElement>, node: QuestNode) => void;
};

const QuestList = ({
  quests,
  toggleQuest,
  expandedQuests,
  onClickQuest,
}: QuestListProps) => {
  const { currentUser } = useUserProfiles();
  const { discoveries } = useDiscoveriesContext();

  return (
    <>
      {quests.map((quest) => {
        const isDiscovered = hasDiscoveredPointOfInterest(
          quest.rootNode.pointOfInterest.id,
          currentUser?.id ?? '',
          discoveries
        );
        
        return (
          <div
            key={quest.rootNode.pointOfInterest.id}
            className="rounded p-2 w-full"
          >
            <button
              className={`w-full flex items-center gap-2 ${!isDiscovered ? 'cursor-not-allowed' : ''}`}
              onClick={() => isDiscovered && toggleQuest(quest.rootNode.pointOfInterest.id)}
              disabled={!isDiscovered}
            >
              <span className={!isDiscovered ? 'opacity-0' : ''}>
                {expandedQuests[quest.rootNode.pointOfInterest.id] ? '-' : '+'}
              </span>
              <img
                src={isDiscovered ? quest.imageUrl : 'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp'}
                className="w-6 h-6 object-cover border-2 border-black"
                alt=""
              />
              <span>{quest.name}</span>
              {!isDiscovered && <span className="text-sm text-gray-500 ml-2">(Undiscovered)</span>}
            </button>

            {expandedQuests[quest.rootNode.pointOfInterest.id] && isDiscovered && (
              <div className="ml-6 mt-2 w-full text-left">
                {(() => {
                  const renderNode = (node: QuestNode) => {

                    const hasDiscoveredNode = hasDiscoveredPointOfInterest(
                      node.pointOfInterest.id,
                      currentUser?.id ?? '',
                      discoveries
                    );
                    return (
                      <div
                        key={node.pointOfInterest.id}
                        className="mb-2"
                        onClick={(e) => onClickQuest(e, node)}
                      >
                        <div className="font-semibold flex items-center gap-2">
                          <img
                            src={hasDiscoveredNode ? node.pointOfInterest.imageURL : "https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp"}
                            className="w-6 h-6 object-cover border-2 border-black"
                            alt=""
                          />
                          <div className="flex flex-col">
                            <span>{node.pointOfInterest.name}</span>
                            {node.objectives.map((objective) => (
                              <div
                                key={objective.challenge.id}
                                className="text-m text-gray-600 font-normal"
                              >
                                <span
                                >
                                  • {objective.challenge.question}
                                  {objective.isCompleted && ' ✅'}
                                </span>
                              </div>
                            ))}
                          </div>
                        </div>
                        {Object.entries(node.children).map(
                          ([childId, childNode]) => {
                            if (
                              node.objectives.some(
                                (obj) =>
                                  obj.challenge.id === childId && obj.isCompleted
                              )
                            ) {
                              return renderNode(childNode);
                            }
                            return null;
                          }
                        )}
                      </div>
                    );
                  };

                  return renderNode(quest.rootNode);
                })()}
              </div>
            )}
          </div>
        );
      })}
    </>
  );
};
