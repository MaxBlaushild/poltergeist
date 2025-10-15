import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { ActivityFeed } from '@poltergeist/types';

interface ActivityFeedContextType {
  activities: ActivityFeed[];
  unseenActivities: ActivityFeed[];
  markActivitiesAsSeen: (activityIds: string[]) => Promise<void>;
  refetchActivities: () => Promise<void>;
}

const ActivityFeedContext = createContext<ActivityFeedContextType | undefined>(undefined);

export const useActivityFeedContext = () => {
  const context = useContext(ActivityFeedContext);
  if (!context) {
    throw new Error('useActivityFeedContext must be used within an ActivityFeedProvider');
  }
  return context;
};

interface ActivityFeedProviderProps {
  children: React.ReactNode;
}

export const ActivityFeedProvider: React.FC<ActivityFeedProviderProps> = ({ children }) => {
  const { apiClient } = useAPI();
  const [activities, setActivities] = useState<ActivityFeed[]>([]);
  const [unseenActivities, setUnseenActivities] = useState<ActivityFeed[]>([]);

  const refetchActivities = useCallback(async () => {
    try {
      const response = await apiClient.get<ActivityFeed[]>('/sonar/activities');
      setActivities(response);
      setUnseenActivities(response.filter(activity => !activity.seen));
    } catch (error) {
      console.error('Failed to fetch activities:', error);
    }
  }, [apiClient]);

  const markActivitiesAsSeen = useCallback(async (activityIds: string[]) => {
    try {
      await apiClient.post('/sonar/activities/markAsSeen', { activityIds });
      setActivities(prevActivities =>
        prevActivities.map(activity =>
          activityIds.includes(activity.id) ? { ...activity, seen: true } : activity
        )
      );
      setUnseenActivities(prevUnseen =>
        prevUnseen.filter(activity => !activityIds.includes(activity.id))
      );
    } catch (error) {
      console.error('Failed to mark activities as seen:', error);
    }
  }, [apiClient]);

  useEffect(() => {
    refetchActivities();
    
    // Poll for new activities every 3 seconds
    const interval = setInterval(() => {
      refetchActivities();
    }, 3000);

    return () => clearInterval(interval);
  }, [refetchActivities]);

  return (
    <ActivityFeedContext.Provider
      value={{
        activities,
        unseenActivities,
        markActivitiesAsSeen,
        refetchActivities,
      }}
    >
      {children}
    </ActivityFeedContext.Provider>
  );
};

