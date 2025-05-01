import React from 'react';
import { Modal, ModalSize } from './shared/Modal.tsx';
import { useQuestLogContext } from '../contexts/QuestLogContext.tsx';
import { CompletedTask } from '../contexts/CompletedTaskContext.tsx';
import { PointOfInterestChallenge, QuestNode } from '@poltergeist/types';

interface CompletedTaskModalProps {
  completedTask: CompletedTask;
}
export const CompletedTaskModal = ({ completedTask }: CompletedTaskModalProps) => {
  const { reward } = completedTask;
  const { quests } = useQuestLogContext();

  if (!quests || !completedTask.quest) {
    return null;
  }

  console.log(completedTask.quest);

  const findNextNode = (node: QuestNode): QuestNode | undefined => {
    // First check if this node contains the completed challenge
    for (const objective of node.objectives) {
      if (objective.challenge.id === completedTask.challenge.id) {
        // If we found the completed challenge, return the next node in the sequence
        return node.children[completedTask.challenge.id];
      }
    }

    // If not found in this node, recursively check children
    for (const child of Object.values(node.children)) {
      const result = findNextNode(child);
      if (result) return result;
    }

    return undefined;
  }
  const nextNode = findNextNode(completedTask.quest.rootNode);
  const isFinished = nextNode === undefined;

  return <Modal size={ModalSize.FULLSCREEN}>
    <div className="flex flex-col items-center gap-6 p-8 rounded-lg">
      <div className="animate-bounce bg-white rounded-lg shadow-md p-4">
        <h1 className="text-4xl font-bold text-amber-500 text-center">Victory!</h1>
      </div>
      
      <div className="w-full text-center space-y-2">
        <div className="bg-white rounded-lg shadow-md p-4">
          <p className="text-lg text-gray-600">You've completed:</p>
          <p className="text-xl font-semibold text-gray-800">
            {completedTask.challenge.question}
          </p>
          {isFinished && (
            <p className="text-lg font-medium text-emerald-600 mt-2">
              Quest Complete: {completedTask.quest.name}
            </p>
          )}
        </div>
      </div>

      {!isFinished && <div className="w-full text-center space-y-2">
        <div className="bg-white rounded-lg shadow-md p-4">
          <p className="text-lg text-gray-600">Next up:</p>
          <p className="text-xl font-semibold text-gray-800">
            {completedTask.challenge.question}
          </p>
        </div>
      </div>}

      <div className="w-full bg-white rounded-xl p-6 space-y-4 shadow-md">
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
      </div>

      <div className="w-full bg-white rounded-xl p-6 shadow-md">
        <h2 className="text-2xl font-bold text-center text-emerald-500 mb-4">New Item!</h2>
        <div className="flex items-center gap-6">
          <div className="relative">
            <div className="absolute -inset-1 bg-amber-300 rounded-lg blur opacity-25 animate-pulse"></div>
            <img 
              src={reward.imageUrl}
              alt={reward.name}
              className="relative w-20 h-20 object-cover rounded-lg border-2 border-amber-400"
            />
          </div>
          <span className="text-xl font-bold text-gray-800">{reward.name}</span>
        </div>
      </div>
    </div>
  </Modal>;
};
