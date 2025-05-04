import React from 'react';
import { PointOfInterestChallenge, Quest, InventoryItem, QuestNode } from '@poltergeist/types';
import { createContext, useContext, useState } from 'react';
import { useQuestLogContext } from './QuestLogContext.tsx';

export interface CompletedTask {
  quest: Quest;
  challenge: PointOfInterestChallenge;
}

interface CompletedTaskContextType {
  completedTask: CompletedTask | null;
  setCompletedTask: (challenge: PointOfInterestChallenge) => void;
  removeCompletedTask: () => void;
}

export const CompletedTaskContext = createContext<CompletedTaskContextType>({
  completedTask: null,
  setCompletedTask: () => {},
  removeCompletedTask: () => {},
});

export const CompletedTaskProvider = ({ children }: { children: React.ReactNode }) => {
  const { quests } = useQuestLogContext();
  const [completedTask, _setCompletedTask] = useState<CompletedTask | null>(null);

  const setCompletedTask = (challenge: PointOfInterestChallenge) => {
    const searchNodeForChallenge = (quest: Quest, node: QuestNode): Quest | null => {
      for (const objective of node.objectives) {
        if (objective.challenge.id === challenge.id) {
          return quest;
        }

        if (objective.nextNode) {
          const result = searchNodeForChallenge(quest, objective.nextNode);
          if (result) {
            return result;
          }
        }
      }

      return null;
    }

    for (const quest of quests) {
      const completedQuest = searchNodeForChallenge(quest, quest.rootNode);
      if (completedQuest) {
        _setCompletedTask({ quest: completedQuest, challenge });
        return;
      }
    }
  };

  const removeCompletedTask = () => {
    _setCompletedTask(null);
  };

  return (
    <CompletedTaskContext.Provider value={{ completedTask, setCompletedTask, removeCompletedTask }}>
      {children}
    </CompletedTaskContext.Provider>
  );
};

export const useCompletedTaskContext = () => {
  return useContext(CompletedTaskContext);
};
