import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  useRef,
} from 'react';
import { useAPI, useLocation } from '@poltergeist/contexts';
import { PointOfInterest, PointOfInterestGroup, Quest, QuestLog, QuestNode } from '@poltergeist/types';
import { useUserProfiles } from './UserProfileContext.tsx';
import { useSubmissionsContext } from './SubmissionsContext.tsx';
import { useTagContext } from './TagContext.tsx';

interface QuestLogContextType {
  refreshQuestLog: () => Promise<void>;
  quests: Quest[];
  pointsOfInterest: PointOfInterest[];
  isRootNode: (pointOfInterest: PointOfInterest) => boolean;
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

  const refreshQuestLog = useCallback(async () => {
    if (!location?.latitude || !location?.longitude || !selectedTags?.length || selectedTags.length === 0) {
      return;
    }

    try {
      const fetchedQuestLog = await apiClient.get<QuestLog>(`/sonar/questlog?lat=${location?.latitude}&lng=${location?.longitude}&tags=${selectedTags.map(tag => tag.name).join(',')}`);
      setQuests(fetchedQuestLog.quests);
      const pointsOfInterest = getMapPointsOfInterest(fetchedQuestLog.quests);
      setPointsOfInterest(pointsOfInterest);
      lastFetchLocation.current = {
        lat: location.latitude,
        lng: location.longitude
      };
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
        return;
      }

    }
    console.log('refreshing quest log');
    console.log('selectedTags', selectedTags);
    refreshQuestLog();
  }, [apiClient, location?.latitude, location?.longitude, selectedTags]);

  const isRootNode = (pointOfInterest: PointOfInterest) => {
    return quests.some(quest => quest.rootNode.pointOfInterest.id === pointOfInterest.id);
  };

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
