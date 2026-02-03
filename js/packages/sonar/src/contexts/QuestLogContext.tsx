import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  useRef,
} from 'react';
import { useAPI } from '@poltergeist/contexts';
import { PointOfInterest, PointOfInterestChallenge, Quest, QuestLog, QuestNode, Task } from '@poltergeist/types';
import { useUserProfiles } from './UserProfileContext.tsx';
import { useSubmissionsContext } from './SubmissionsContext.tsx';
import { useTagContext } from '@poltergeist/contexts';
import { useZoneContext } from '@poltergeist/contexts/dist/zones';
import { v4 as uuidv4 } from 'uuid';
const getAllPointsOfInterestIdsForQuest = (quest: Quest): string[] => {
  const pointsOfInterest: string[] = [];

  const traverseNode = (node: QuestNode) => {
    // Add current node's POI
    pointsOfInterest.push(node.pointOfInterest.id);

    // Recursively traverse child nodes
    node.objectives.forEach(objective => {
      if (objective.nextNode) {
        traverseNode(objective.nextNode);
      }
    });
  };

  // Start traversal from root node of each quest
  traverseNode(quest.rootNode);

  return pointsOfInterest;
};

interface QuestLogContextType {
  refreshQuestLog: () => Promise<void>;
  quests: Quest[];
  pointsOfInterest: PointOfInterest[];
  isRootNode: (pointOfInterest: PointOfInterest) => boolean;
  pendingTasks: Record<string, Task[]>;
  completedTasks: Record<string, Task[]>;
  trackedQuestIds: string[];
  trackedPointOfInterestIds: string[];
  trackQuest: (questID: string) => Promise<void>;
  untrackQuest: (questID: string) => Promise<void>;
  untrackAllQuests: () => Promise<void>;
}

interface QuestLogProviderProps {
  children: React.ReactNode;
}

export const QuestLogContext = createContext<QuestLogContextType | undefined>(
  undefined
);

export const useQuestLogContext = () => {
  const context = useContext(QuestLogContext);
  if (!context) {
    throw new Error(
      'useQuestLogContext must be used within a QuestLogContextProvider'
    );
  }
  return context;
};

export const QuestLogContextProvider: React.FC<QuestLogProviderProps> = ({ children }) => {
  const { apiClient } = useAPI();
  const [quests, setQuests] = useState<Quest[]>([]);
  const { selectedZone } = useZoneContext();
  const { selectedTags } = useTagContext();
  const [pointsOfInterest, setPointsOfInterest] = useState<PointOfInterest[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);
  const lastFetchTags = useRef<string[]>([]);
  const lastFetchZoneID = useRef<string | null>(null);
  const isFetching = useRef<boolean>(false);
  const fetchPromise = useRef<Promise<void> | null>(null);
  const [pendingTasks, setPendingTasks] = useState<Record<string, Task[]>>({});
  const [completedTasks, setCompletedTasks] = useState<Record<string, Task[]>>({});
  const [trackedQuestIds, setTrackedQuestIds] = useState<string[]>([]);
  const [trackedPointOfInterestIds, setTrackedPointOfInterestIds] = useState<string[]>([]);

  const refreshQuestLog = useCallback(async () => {
    if (isFetching.current) {
      return;
    }

    // If there's already a fetch in progress, wait for it
    if (fetchPromise.current) {
      await fetchPromise.current;
      return;
    }

    try {
      isFetching.current = true;
      setLoading(true);
      
      // Create a new fetch promise
      fetchPromise.current = (async () => {
        const fetchedQuestLog = await apiClient.get<QuestLog>(
          `/sonar/questlog?zoneId=${selectedZone?.id ?? uuidv4()}${
            selectedTags.length ? `&tags=${selectedTags.map(tag => tag.name).join(',')}` : ''
          }`
        );

        setQuests(fetchedQuestLog.quests);
        const pointsOfInterest = getMapPointsOfInterest(fetchedQuestLog.quests);
        setPointsOfInterest(pointsOfInterest);
        setPendingTasks(fetchedQuestLog.pendingTasks);
        setCompletedTasks(fetchedQuestLog.completedTasks);
        setTrackedQuestIds(fetchedQuestLog.trackedQuestIds);
        
        const trackedQuests = fetchedQuestLog.trackedQuestIds
          .map(id => fetchedQuestLog.quests.find(quest => quest.id === id))
          .filter((quest): quest is Quest => quest !== undefined);
        
        const trackedPointsOfInterestIds = trackedQuests.flatMap(quest => 
          getAllPointsOfInterestIdsForQuest(quest)
        );
        
        setTrackedPointOfInterestIds(trackedPointsOfInterestIds);
        
        // Update last fetch state
        lastFetchZoneID.current = selectedZone?.id ?? null;
        lastFetchTags.current = selectedTags.map(tag => tag.name);
      })();

      await fetchPromise.current;
    } catch (error) {
      console.error('Error fetching quest log:', error);
      setError(error as Error);
    } finally {
      setLoading(false);
      isFetching.current = false;
      fetchPromise.current = null;
    }
  }, [apiClient, selectedTags, selectedZone]);

  const shouldFetchQuestLog = useCallback(() => {
    if (isFetching.current) return false;
    if (!lastFetchZoneID.current) return true;
    if (selectedZone && lastFetchZoneID.current !== selectedZone.id) return true;
    if (lastFetchTags.current.length !== selectedTags.length) return true;
    if (selectedTags.some(tag => !lastFetchTags.current.includes(tag.name))) return true;
    return false;
  }, [selectedZone, selectedTags]);

  useEffect(() => {
    if (shouldFetchQuestLog()) {
      refreshQuestLog();
    }
  }, [shouldFetchQuestLog, refreshQuestLog]);

  const isRootNode = (pointOfInterest: PointOfInterest) => {
    return quests.some(quest => quest.rootNode.pointOfInterest.id === pointOfInterest.id);
  };

  const trackQuest = useCallback(async (questID: string) => {
    try {
      await apiClient.post(`/sonar/trackedQuests`, { questId: questID });
      // Don't update state optimistically, wait for refreshQuestLog
      await refreshQuestLog();
    } catch (error) {
      console.error('Error tracking quest:', error);
    }
  }, [apiClient, refreshQuestLog]);

  const untrackQuest = useCallback(async (questID: string) => {
    try {
      await apiClient.delete(`/sonar/trackedQuests/${questID}`);
      await refreshQuestLog();
    } catch (error) {
      console.error('Error untracking quest:', error);
    }
  }, [apiClient, refreshQuestLog]);

  const untrackAllQuests = useCallback(async () => {
    try {
      await apiClient.delete(`/sonar/trackedQuests`);
      await refreshQuestLog();
    } catch (error) {
      console.error('Error untracking all quests:', error);
    }
  }, [apiClient, refreshQuestLog]);

  return (
    <QuestLogContext.Provider
      value={{
        refreshQuestLog,
        quests,
        pointsOfInterest,
        isRootNode,
        pendingTasks,
        completedTasks,
        trackedQuestIds,
        trackQuest,
        untrackQuest,
        untrackAllQuests,
        trackedPointOfInterestIds,
      }}
    >
      {children}
    </QuestLogContext.Provider>
  );
};

const getMapPointsOfInterest = (quests: Quest[]) => {
  const pointsOfInterest: PointOfInterest[] = [];

  quests.forEach((quest) => {
    // Only include quests that are in the quest log (which means they've been accepted if they have a quest giver)
    const addPointsFromNode = (node: QuestNode) => {
      pointsOfInterest.push(node.pointOfInterest);
      
      node.objectives.forEach(objective => {
        if (objective.nextNode && objective.isCompleted) {
          addPointsFromNode(objective.nextNode);
        }
      });
    };

    addPointsFromNode(quest.rootNode);
  });
  
  return pointsOfInterest;
};
