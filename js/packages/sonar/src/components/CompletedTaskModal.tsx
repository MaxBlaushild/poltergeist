import React, { useEffect } from 'react';
import { Modal, ModalSize } from './shared/Modal.tsx';
import { useQuestLogContext } from '../contexts/QuestLogContext.tsx';
import { CompletedTask, useCompletedTaskContext } from '../contexts/CompletedTaskContext.tsx';
import { PointOfInterestChallenge, QuestNode } from '@poltergeist/types';
import { useInventory } from '@poltergeist/contexts';
import { usePointOfInterestContext } from '../contexts/PointOfInterestContext.tsx';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { useDiscoveriesContext } from '../contexts/DiscoveriesContext.tsx';
interface QuestNodes {
  currentNode: QuestNode;
  nextNode: QuestNode | null;
}

export const CompletedTaskModal = () => {
  const { getInventoryItemById } = useInventory();
  const { untrackQuest } = useQuestLogContext();
  const { currentUser } = useUserProfiles();
  const { removeCompletedTask, completedTask } = useCompletedTaskContext();
  const { discoveries } = useDiscoveriesContext();


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

  const reward = getInventoryItemById(completedTask?.challenge.inventoryItemId);

  return <Modal size={ModalSize.FREE} onClose={removeCompletedTask}>
    <div className="flex flex-col items-center gap-4 p-6 rounded-lg">
      <div className="animate-bounce bg-white rounded-lg shadow-md p-3">
        <h1 className="text-2xl font-bold text-amber-500 text-center">Victory!</h1>
      </div>
      
      <div className="w-full space-y-2">
        <div className="bg-white rounded-lg shadow-md p-3">
          <h2 className="text-xl font-bold text-center text-emerald-500 mb-3">Completed</h2>
          <div className="flex items-center gap-4">
            <img 
              src={currentNode?.pointOfInterest?.imageURL}
              alt={currentNode?.pointOfInterest?.name}
              className="w-16 h-16 object-cover rounded-lg flex-shrink-0"
            />
            <div className="flex-grow text-left">
              <p className="text-base font-semibold text-gray-900 mb-2">
                {currentNode?.pointOfInterest?.name}
              </p>
              <p className="text-base font-semibold text-gray-600">
                {completedTask.challenge.question}
              </p>
              {isFinished && (
                <p className="text-sm font-medium text-emerald-600 mt-2">
                  Quest Complete: {completedTask.quest.name}
                </p>
              )}
            </div>
          </div>
        </div>
      </div>

      {!isFinished && <div className="w-full space-y-2">
        <div className="bg-white rounded-lg shadow-md p-3">
          <h2 className="text-xl font-bold text-center text-emerald-500 mb-3">Next up</h2>
          <div className="flex items-center gap-4">
            <img 
              src={discoveriesForUser[nextNode?.pointOfInterest?.id] ? nextNode?.pointOfInterest?.imageURL : 'https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp'}
              alt={nextNode?.pointOfInterest?.name}
              className="w-16 h-16 object-cover rounded-lg flex-shrink-0"
            />
            <div className="flex-grow text-left">
              <p className="text-base font-semibold text-gray-900 mb-2">
                {nextNode?.pointOfInterest?.name}
              </p>
              <p className="text-base font-semibold text-gray-600">
                {nextNode?.objectives[0].challenge.question}
              </p>
            </div>
          </div>
        </div>
      </div>}

      {/* https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp */}

      {/* <div className="w-full bg-white rounded-xl p-6 space-y-4 shadow-md">
        <h2 className="text-2xl font-bold text-center text-emerald-500">Rewards</h2>
        
        <div className="flex flex-col gap-3">
          <div className="flex justify-between items-center bg-white p-3 rounded-lg shadow-sm">
            <span className="text-lg text-gray-600">Experience</span>
            <span className="text-xl font-bold text-amber-500">+100</span>
          </div>
          <div className="flex justify-between items-center bg-white p-3 rounded-lg shadow-sm">
            <span className="text-lg text-gray-600">Coins</span>
            <span className="text-xl font-bold text-amber-500">+50</span>
          </div>
        </div>
      </div> */}

      <div className="w-full bg-white rounded-xl p-4 shadow-md">
        <h2 className="text-xl font-bold text-center text-emerald-500 mb-3">Earned an item!</h2>
        <div className="flex items-center gap-4">
          <div className="relative flex-shrink-0">
            <div className="absolute -inset-1 bg-amber-300 rounded-lg blur opacity-25 animate-pulse"></div>
            <img 
              src={reward.imageUrl}
              alt={reward.name}
              className="relative w-16 h-16 object-cover rounded-lg border-2 border-amber-400"
            />
          </div>
          <span className="text-lg font-bold text-gray-800 text-left">{reward.name}</span>
        </div>
      </div>
    </div>
  </Modal>;
};
