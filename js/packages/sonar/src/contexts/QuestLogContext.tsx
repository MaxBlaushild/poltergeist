import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  useRef,
} from 'react';
import { useAPI, useLocation } from '@poltergeist/contexts';
import { PointOfInterest, PointOfInterestChallenge, Quest, QuestLog, QuestNode, Task } from '@poltergeist/types';
import { useUserProfiles } from './UserProfileContext.tsx';
import { useSubmissionsContext } from './SubmissionsContext.tsx';
import { useTagContext } from '@poltergeist/contexts';

const getAllPointsOfInterestIdsForQuest = (quest: Quest): string[] => {
  const pointsOfInterest: string[] = [];

  const traverseNode = (node: QuestNode) => {
    // Add current node's POI
    pointsOfInterest.push(node.pointOfInterest.id);

    // Recursively traverse child nodes
    Object.values(node.children).forEach(childNode => {
      traverseNode(childNode);
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
  const { selectedTags } = useTagContext();
  const [pointsOfInterest, setPointsOfInterest] = useState<PointOfInterest[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);
  const { location } = useLocation();
  const lastFetchLocation = useRef<{lat: number, lng: number} | null>(null);
  const lastFetchTags = useRef<string[]>([]);
  const [pendingTasks, setPendingTasks] = useState<Record<string, Task[]>>({});
  const [completedTasks, setCompletedTasks] = useState<Record<string, Task[]>>({});
  const [trackedQuestIds, setTrackedQuestIds] = useState<string[]>([]);
  const [trackedPointOfInterestIds, setTrackedPointOfInterestIds] = useState<string[]>([]);
  const refreshQuestLog = useCallback(async () => {
    if (!location?.latitude || !location?.longitude) {
      return;
    }

    try {
      const fetchedQuestLog = await apiClient.get<QuestLog>(`/sonar/questlog?lat=${location?.latitude}&lng=${location?.longitude}&${selectedTags.length ? `tags=${selectedTags.map(tag => tag.name).join(',')}` : ''}`);
      setQuests(fetchedQuestLog.quests);
      const pointsOfInterest = getMapPointsOfInterest(fetchedQuestLog.quests);
      setPointsOfInterest(pointsOfInterest);
      setPendingTasks(fetchedQuestLog.pendingTasks);
      setCompletedTasks(fetchedQuestLog.completedTasks);
      setTrackedQuestIds(fetchedQuestLog.trackedQuestIds);
      lastFetchLocation.current = {
        lat: location.latitude,
        lng: location.longitude
      };
      
      const trackedQuests = fetchedQuestLog.trackedQuestIds.map(id => quests.find(quest => quest.id === id)).filter(quest => quest !== undefined);
      const trackedPointsOfInterestIds = trackedQuests.flatMap(quest => getAllPointsOfInterestIdsForQuest(quest));
      setTrackedPointOfInterestIds(trackedPointsOfInterestIds);
      lastFetchTags.current = selectedTags.map(tag => tag.name);
    } catch (error) {
      setError(error as Error);
    } finally {
      setLoading(false);
    }
  }, [apiClient, location?.latitude, location?.longitude, selectedTags]);

  const fetchQuestLog = useCallback(async () => {
    if (!location?.latitude || !location?.longitude) {
      return;
    }

    // Only fetch if moved more than 100 meters or first fetch
    if (lastFetchLocation.current) {
      const R = 6371e3; // Earth's radius in meters
      const φ1 = lastFetchLocation.current.lat * Math.PI/180;
      const φ2 = location.latitude * Math.PI/180;
      const Δφ = (location.latitude - lastFetchLocation.current.lat) * Math.PI/180;
      const Δλ = (location.longitude - lastFetchLocation.current.lng) * Math.PI/180;

      const a = Math.sin(Δφ/2) * Math.sin(Δφ/2) +
              Math.cos(φ1) * Math.cos(φ2) *
              Math.sin(Δλ/2) * Math.sin(Δλ/2);
      const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a));
      const distance = R * c;

      if (distance < 100) { // Less than 100 meters moved
        if (lastFetchTags.current.length !== selectedTags.length) {
          refreshQuestLog();
        }
      }

    }
    refreshQuestLog();
  }, [apiClient, location?.latitude, location?.longitude, selectedTags]);

  const isRootNode = (pointOfInterest: PointOfInterest) => {
    return quests.some(quest => quest.rootNode.pointOfInterest.id === pointOfInterest.id);
  };

  const trackQuest = useCallback(async (questID: string) => {
    await apiClient.post(`/sonar/trackedPointOfInterestGroups`, { pointOfInterestGroupID: questID });
    setTrackedQuestIds([...trackedQuestIds, questID]);
    fetchQuestLog();
  }, [apiClient, trackedQuestIds]);

  const untrackQuest = useCallback(async (questID: string) => {
    await apiClient.delete(`/sonar/trackedPointOfInterestGroups/${questID}`);
    fetchQuestLog();
  }, [apiClient, trackedQuestIds]);

  const untrackAllQuests = useCallback(async () => {
    await apiClient.delete(`/sonar/trackedPointOfInterestGroups`);
    fetchQuestLog();
  }, [apiClient]);

  useEffect(() => {
    fetchQuestLog();
  }, [fetchQuestLog]);

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
    const addPointsFromNode = (node: QuestNode) => {
      pointsOfInterest.push(node.pointOfInterest);
      
      Object.entries(node.children).forEach(([childId, childNode]) => {
        if (node.objectives.some(obj => obj.challenge.id === childId && obj.isCompleted)) {
          addPointsFromNode(childNode);
        }
      });
    };

    addPointsFromNode(quest.rootNode);
  });
  
  return pointsOfInterest;
};
