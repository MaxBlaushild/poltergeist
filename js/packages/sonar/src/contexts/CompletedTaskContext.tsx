import { PointOfInterestChallenge, Quest, InventoryItem } from '@poltergeist/types';
import { createContext, useContext } from 'react';

export interface CompletedTask {
  quest: Quest;
  challenge: PointOfInterestChallenge;
  reward: InventoryItem;
}

interface CompletedTaskContextType {
  completedTask: CompletedTask | null;
  setCompletedTask: (completedTask: CompletedTask) => void;
}

export const CompletedTaskContext = createContext<CompletedTaskContextType>({
  completedTask: null,
  setCompletedTask: () => {},
});

export const useCompletedTaskContext = () => {
  return useContext(CompletedTaskContext);
};
