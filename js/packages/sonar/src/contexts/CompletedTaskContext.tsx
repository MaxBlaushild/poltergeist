import React from 'react';
import { 
  PointOfInterestChallenge, 
  Quest, 
  QuestNode, 
  SubmissionResult,
  isLevelUpActivity,
  isReputationUpActivity,
  ReputationUpActivityData
} from '@poltergeist/types';
import { createContext, useContext, useState, useEffect } from 'react';
import { useQuestLogContext } from './QuestLogContext.tsx';
import { useActivityFeedContext } from './ActivityFeedContext.tsx';

export interface CompletedTask {
  quest: Quest;
  challenge: PointOfInterestChallenge;
  result: SubmissionResult;
}

interface CompletedTaskContextType {
  completedTask: CompletedTask | null;
  setCompletedTask: (challenge: PointOfInterestChallenge, result: SubmissionResult) => void;
  removeCompletedTask: () => void;
  levelUp: boolean;
  setLevelUp: (levelUp: boolean) => void;
  reputationUp: boolean;
  setReputationUp: (reputationUp: boolean) => void;
  zoneId: string | null;
  zoneName: string | null;
  newReputationLevel: number | null;
  setZoneId: (zoneId: string | null) => void;
}

export const CompletedTaskContext = createContext<CompletedTaskContextType>({
  completedTask: null,
  setCompletedTask: () => {},
  removeCompletedTask: () => {},
  levelUp: false,
  setLevelUp: () => {},
  reputationUp: false,
  setReputationUp: () => {},
  zoneId: null,
  zoneName: null,
  newReputationLevel: null,
  setZoneId: () => {},
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

  // Listen to activity feed for level ups and reputation changes
  useEffect(() => {
    for (const activity of unseenActivities) {
      if (isLevelUpActivity(activity)) {
        setLevelUp(true);
        // Mark as seen after a delay to allow modal to show
        setTimeout(() => {
          markActivitiesAsSeen([activity.id]);
        }, 3000);
      } else if (isReputationUpActivity(activity)) {
        const data = activity.data as ReputationUpActivityData;
        setReputationUp(true);
        setZoneId(data.zoneId);
        setZoneName(data.zoneName);
        setNewReputationLevel(data.newLevel);
        // Mark as seen after a delay to allow modal to show
        setTimeout(() => {
          markActivitiesAsSeen([activity.id]);
        }, 3000);
      }
    }
  }, [unseenActivities, markActivitiesAsSeen]);

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
      setCompletedTask, 
      removeCompletedTask, 
      levelUp, 
      setLevelUp, 
      reputationUp, 
      setReputationUp,
      zoneId,
      zoneName,
      newReputationLevel,
      setZoneId,
    }}>
      {children}
    </CompletedTaskContext.Provider>
  );
};

export const useCompletedTaskContext = () => {
  return useContext(CompletedTaskContext);
};
