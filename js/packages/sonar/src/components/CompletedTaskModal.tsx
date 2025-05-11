import React, { useEffect } from 'react';
import { Modal, ModalSize } from './shared/Modal.tsx';
import { useQuestLogContext } from '../contexts/QuestLogContext.tsx';
import { useCompletedTaskContext } from '../contexts/CompletedTaskContext.tsx';
import { QuestNode } from '@poltergeist/types';
import { useInventory } from '@poltergeist/contexts';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { useDiscoveriesContext } from '../contexts/DiscoveriesContext.tsx';
import { useZoneContext } from '@poltergeist/contexts';

interface QuestNodes {
  currentNode: QuestNode;
  nextNode: QuestNode | null;
}

export const CompletedTaskModal = () => {
  const { zones } = useZoneContext();
  const { getInventoryItemById } = useInventory();
  const { untrackQuest } = useQuestLogContext();
  const { currentUser } = useUserProfiles();
  const { removeCompletedTask, completedTask, setLevelUp, setReputationUp } = useCompletedTaskContext();
  const { discoveries } = useDiscoveriesContext();

  const reputedZone = zones.find((zone) => zone.id === completedTask?.result?.zoneID);

  const findNextNode = (node: QuestNode): QuestNodes | null | undefined => {
    for (const objective of node.objectives) {
      if (objective.challenge.id === completedTask?.challenge.id) {
        return {
          currentNode: node,
          nextNode: objective.nextNode ?? null,
        };
      }

      if (objective.nextNode) {
        const result = findNextNode(objective.nextNode);
        if (result) return result;
      }
    }
    return undefined;
  }

  const discoveriesForUser = discoveries.filter((discovery) => discovery.userId === currentUser?.id).reduce((acc, discovery) => {
    acc[discovery.pointOfInterestId] = true;
    return acc;
  }, {});

  const nodes = completedTask ? findNextNode(completedTask?.quest.rootNode) : { currentNode: null, nextNode: null };
  const nextNode = nodes?.nextNode;
  const currentNode = nodes?.currentNode;
  const isFinished = !nextNode;

  useEffect(() => {
    if (isFinished && completedTask) {
      untrackQuest(completedTask?.quest.id);
    }
  }, [completedTask, untrackQuest, isFinished]);

  if (!completedTask) return null;

  return <Modal size={ModalSize.FREE} onClose={() => {
    removeCompletedTask();
    setTimeout(() => {
      if (completedTask?.result.levelUp) {
        setLevelUp(true);
      }
      if (completedTask?.result.reputationUp) {
        setReputationUp(true);
      }
    }, 200);
  }}>
    <div className="flex flex-col items-center gap-2 p-4 rounded-lg">
      <h1 className="text-xl font-bold text-amber-500">Victory!</h1>
      
      <div className="w-full bg-white rounded-lg shadow-md p-2">
        <div className="flex items-center gap-2">
          <img 
            src={currentNode?.pointOfInterest?.imageURL}
            alt={currentNode?.pointOfInterest?.name}
            className="w-12 h-12 object-cover rounded-lg"
          />
          <div className="flex-grow">
            <p className="text-sm font-semibold text-gray-900">{currentNode?.pointOfInterest?.name}</p>
            <p className="text-xs text-gray-600">{completedTask.challenge.question}</p>
            {isFinished && (
              <p className="text-xs text-emerald-600">Quest Complete: {completedTask.quest.name}</p>
            )}
          </div>
        </div>
      </div>

      {!isFinished && (
        <div className="w-full bg-white rounded-lg shadow-md p-2">
          <div className="flex items-center gap-2">
            <img 
              src={discoveriesForUser[nextNode?.pointOfInterest?.id] ? nextNode?.pointOfInterest?.imageURL : 'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp'}
              alt={nextNode?.pointOfInterest?.name}
              className="w-12 h-12 object-cover rounded-lg"
            />
            <div className="flex-grow">
              <p className="text-sm font-semibold text-gray-900">{nextNode?.pointOfInterest?.name}</p>
              <p className="text-xs text-gray-600">{nextNode?.objectives[0].challenge.question}</p>
            </div>
          </div>
        </div>
      )}

      <div className="w-full grid grid-cols-2 gap-2">
        {completedTask.result.experienceAwarded > 0 && (
          <div className="bg-white rounded-lg p-2 shadow-md">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg border border-blue-400 flex items-center justify-center">
                <span className="text-sm font-bold text-blue-600">XP</span>
              </div>
              <span className="text-sm font-bold">+{completedTask.result.experienceAwarded}</span>
            </div>
          </div>
        )}

        {completedTask.result.reputationAwarded > 0 && reputedZone && (
          <div className="bg-white rounded-lg p-2 shadow-md">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg border border-purple-400 flex items-center justify-center">
                <span className="text-sm font-bold text-purple-600">REP</span>
              </div>
              <div>
                <span className="text-sm font-bold">+{completedTask.result.reputationAwarded}</span>
                <p className="text-xs text-gray-600">{reputedZone.name}</p>
              </div>
            </div>
          </div>
        )}
      </div>

      {completedTask.result.itemsAwarded.length > 0 && (
        <div className="w-full bg-white rounded-lg p-2 shadow-md">
          <div className="flex flex-wrap gap-2 justify-center">
            {completedTask.result.itemsAwarded.map((reward, index) => (
              <div key={index} className="flex items-center gap-1">
                <img 
                  src={reward.imageUrl}
                  alt={reward.name}
                  className="w-8 h-8 object-cover rounded-lg border border-amber-400"
                />
                <span className="text-xs font-bold">{reward.name}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  </Modal>;
};
