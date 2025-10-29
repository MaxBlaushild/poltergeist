import React from 'react';
import { 
  PointOfInterestChallenge, 
  Quest, 
  QuestNode, 
  SubmissionResult,
  isLevelUpActivity,
  isReputationUpActivity,
  isChallengeCompletedActivity,
  ReputationUpActivityData,
  ChallengeCompletedActivityData
} from '@poltergeist/types';
import { createContext, useContext, useState, useEffect, useRef } from 'react';
import { useQuestLogContext } from './QuestLogContext.tsx';
import { useActivityFeedContext } from './ActivityFeedContext.tsx';

// Local extensions until shared types are rebuilt
type ChallengeCompletedActivityDataWithGold = ChallengeCompletedActivityData & { goldAwarded?: number };
type SubmissionResultWithGold = SubmissionResult & { goldAwarded?: number };

export interface CompletedTask {
  quest: Quest;
  challenge: PointOfInterestChallenge;
  result: SubmissionResult;
}

type ModalType = 'challenge' | 'levelUp' | 'reputationUp';

interface ModalQueueItem {
  type: ModalType;
  data?: any;
}

interface CompletedTaskContextType {
  completedTask: CompletedTask | null;
  removeCompletedTask: () => void;
  levelUp: boolean;
  setLevelUp: (levelUp: boolean) => void;
  reputationUp: boolean;
  setReputationUp: (reputationUp: boolean) => void;
  zoneId: string | null;
  zoneName: string | null;
  newReputationLevel: number | null;
  setZoneId: (zoneId: string | null) => void;
  // Modal queue management
  modalQueue: ModalQueueItem[];
  currentModalIndex: number;
  addToModalQueue: (modal: ModalQueueItem) => void;
  advanceModalQueue: () => void;
  clearModalQueue: () => void;
  getCurrentModal: () => ModalQueueItem | null;
}

export const CompletedTaskContext = createContext<CompletedTaskContextType>({
  completedTask: null,
  removeCompletedTask: () => {},
  levelUp: false,
  setLevelUp: () => {},
  reputationUp: false,
  setReputationUp: () => {},
  zoneId: null,
  zoneName: null,
  newReputationLevel: null,
  setZoneId: () => {},
  modalQueue: [],
  currentModalIndex: -1,
  addToModalQueue: () => {},
  advanceModalQueue: () => {},
  clearModalQueue: () => {},
  getCurrentModal: () => null,
});

export const CompletedTaskProvider = ({ children }: { children: React.ReactNode }) => {
  const { quests } = useQuestLogContext();
  const { unseenActivities, markActivitiesAsSeen } = useActivityFeedContext();
  const [completedTask, _setCompletedTask] = useState<CompletedTask | null>(null);
  const [levelUp, setLevelUp] = useState(false);
  const [reputationUp, setReputationUp] = useState(false);
  const [zoneId, setZoneId] = useState<string | null>(null);
  const [zoneName, setZoneName] = useState<string | null>(null);
  const [newReputationLevel, setNewReputationLevel] = useState<number | null>(null);
  
  // Modal queue state
  const [modalQueue, setModalQueue] = useState<ModalQueueItem[]>([]);
  const [currentModalIndex, setCurrentModalIndex] = useState(-1);
  const processedActivities = useRef<Set<string>>(new Set());

  // Modal queue management functions
  const addToModalQueue = (modal: ModalQueueItem) => {
    setModalQueue(prev => [...prev, modal]);
    if (currentModalIndex === -1) {
      setCurrentModalIndex(0);
    }
  };

  const advanceModalQueue = () => {
    setCurrentModalIndex(prev => {
      const nextIndex = prev + 1;
      if (nextIndex >= modalQueue.length) {
        // Queue is empty, reset
        setModalQueue([]);
        return -1;
      }
      return nextIndex;
    });
  };

  const clearModalQueue = () => {
    setModalQueue([]);
    setCurrentModalIndex(-1);
  };

  const getCurrentModal = () => {
    if (currentModalIndex >= 0 && currentModalIndex < modalQueue.length) {
      return modalQueue[currentModalIndex];
    }
    return null;
  };

  // Listen to activity feed for level ups, reputation changes, and challenge completions
  useEffect(() => {
    for (const activity of unseenActivities) {
      // Skip if we've already processed this activity
      if (processedActivities.current.has(activity.id)) {
        continue;
      }

      if (isChallengeCompletedActivity(activity)) {
        const data = activity.data as ChallengeCompletedActivityDataWithGold;
        
        // Find the quest matching the questId from activity
        const quest = quests.find(q => q.id === data.questId);
        if (!quest) {
          // If quest not found, mark as seen and skip
          setTimeout(() => {
            markActivitiesAsSeen([activity.id]);
          }, 100);
          processedActivities.current.add(activity.id);
          continue;
        }

        // Find the challenge within the quest
        const findChallengeInNode = (node: QuestNode): PointOfInterestChallenge | null => {
          for (const objective of node.objectives) {
            if (objective.challenge.id === data.challengeId) {
              return objective.challenge;
            }
            if (objective.nextNode) {
              const found = findChallengeInNode(objective.nextNode);
              if (found) return found;
            }
          }
          return null;
        };

        const challenge = findChallengeInNode(quest.rootNode);
        if (!challenge) {
          // If challenge not found, mark as seen and skip
          setTimeout(() => {
            markActivitiesAsSeen([activity.id]);
          }, 100);
          processedActivities.current.add(activity.id);
          continue;
        }

        // Construct SubmissionResult from activity data
        const result: SubmissionResultWithGold = {
          successful: data.successful,
          reason: data.reason,
          questCompleted: data.questCompleted,
          experienceAwarded: data.experienceAwarded,
          reputationAwarded: data.reputationAwarded,
          itemsAwarded: data.itemsAwarded,
          goldAwarded: data.goldAwarded,
          zoneID: data.zoneId,
        };

        // Set the completed task
        _setCompletedTask({ quest, challenge, result });
        
        // Clear any existing queue and build a new one for this challenge completion
        clearModalQueue();
        
        // Add challenge completion modal to queue first
        addToModalQueue({ type: 'challenge' });
        
        // Mark as processed and seen after a delay to allow modal to show
        processedActivities.current.add(activity.id);
        setTimeout(() => {
          markActivitiesAsSeen([activity.id]);
        }, 3000);
        
        // Only process one challenge completion at a time
        break;
      } else if (isLevelUpActivity(activity)) {
        // Only add level up modal if we're not currently processing a challenge completion
        if (!completedTask) {
          clearModalQueue();
          addToModalQueue({ type: 'levelUp' });
        }
        setLevelUp(true);
        // Mark as processed and seen after a delay to allow modal to show
        processedActivities.current.add(activity.id);
        setTimeout(() => {
          markActivitiesAsSeen([activity.id]);
        }, 3000);
      } else if (isReputationUpActivity(activity)) {
        const data = activity.data as ReputationUpActivityData;
        // Only add reputation up modal if we're not currently processing a challenge completion
        if (!completedTask) {
          clearModalQueue();
          addToModalQueue({ type: 'reputationUp', data: { zoneId: data.zoneId, zoneName: data.zoneName, newLevel: data.newLevel } });
        }
        setReputationUp(true);
        setZoneId(data.zoneId);
        setZoneName(data.zoneName);
        setNewReputationLevel(data.newLevel);
        // Mark as processed and seen after a delay to allow modal to show
        processedActivities.current.add(activity.id);
        setTimeout(() => {
          markActivitiesAsSeen([activity.id]);
        }, 3000);
      }
    }
  }, [unseenActivities, markActivitiesAsSeen, quests, completedTask]);

  const setCompletedTask = (challenge: PointOfInterestChallenge, result: SubmissionResult) => {
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
        _setCompletedTask({ quest: completedQuest, challenge, result });
        return;
      }
    }
  };

  const removeCompletedTask = () => {
    _setCompletedTask(null);
  };

  return (
    <CompletedTaskContext.Provider value={{ 
      completedTask, 
      removeCompletedTask, 
      levelUp, 
      setLevelUp, 
      reputationUp, 
      setReputationUp,
      zoneId,
      zoneName,
      newReputationLevel,
      setZoneId,
      modalQueue,
      currentModalIndex,
      addToModalQueue,
      advanceModalQueue,
      clearModalQueue,
      getCurrentModal,
    }}>
      {children}
    </CompletedTaskContext.Provider>
  );
};

export const useCompletedTaskContext = () => {
  return useContext(CompletedTaskContext);
};
