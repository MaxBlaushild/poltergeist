import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { Category, Activity } from '@poltergeist/types';

interface ActivityContextType {
  categories: Category[];
  activities: Activity[];
  fetchCategories: () => Promise<void>;
  fetchActivities: () => Promise<void>;
  createCategory: (categoryData: Partial<Category>) => Promise<void>;
  createActivity: (activityData: Partial<Activity>) => Promise<void>;
  areCategoriesLoading: boolean;
  areActivitiesLoading: boolean;
  isCreatingCategory: boolean;
  isCreatingActivity: boolean;
}

export const ActivityContext = createContext<ActivityContextType | undefined>(undefined);

export const useActivityContext = () => {
  const context = useContext(ActivityContext);
  if (!context) {
    throw new Error('useActivityContext must be used within a ActivityContextProvider');
  }
  return context;
};

export const ActivityContextProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { apiClient } = useAPI();
  const [categories, setCategories] = useState<Category[]>([]);
  const [activities, setActivities] = useState<Activity[]>([]);
  const [areCategoriesLoading, setAreCategoriesLoading] = useState(false);
  const [areActivitiesLoading, setAreActivitiesLoading] = useState(false);
  const [isCreatingCategory, setIsCreatingCategory] = useState(false);
  const [isCreatingActivity, setIsCreatingActivity] = useState(false);

  const fetchCategories = useCallback(async () => {
    try {
      setAreCategoriesLoading(true);
      const response = await apiClient.get<Category[]>('/sonar/categories');
      setCategories(response);
    } catch (error) {
      console.error('Failed to fetch categories', error);
    } finally {
      setAreCategoriesLoading(false);
    }
  }, [apiClient, setCategories, setAreCategoriesLoading]);

  const fetchActivities = useCallback(async () => {
    try {
      setAreActivitiesLoading(true);
      const response = await apiClient.get<Activity[]>('/sonar/activities');
      setActivities(response);
    } catch (error) {
      console.error('Failed to fetch activities', error);
    } finally {
      setAreActivitiesLoading(false);
    }
  }, [apiClient, setActivities, setAreActivitiesLoading]);

  const createCategory = async (categoryData: Partial<Category>) => {
    try {
      setIsCreatingCategory(true);
      const response = await apiClient.post<Category>('/sonar/categories', { name: categoryData.title });
      fetchActivities();
      fetchCategories();
    } catch (error) {
      console.error('Failed to create category', error);
    } finally {
      setIsCreatingCategory(false);
    }
  };

  const createActivity = async (activityData: Partial<Activity>) => {
    try {
      setIsCreatingActivity(true);
      await apiClient.post<Activity>('/sonar/activities', activityData);
      fetchActivities();
      fetchCategories();
    } catch (error) {
      console.error('Failed to create activity', error);
    } finally {
      setIsCreatingActivity(false);
    }
  };

  useEffect(() => {
    fetchCategories();
    fetchActivities();
  }, [fetchCategories, fetchActivities]);

  return (
    <ActivityContext.Provider value={{ 
        categories, 
        activities, 
        fetchCategories, 
        fetchActivities, 
        createCategory, 
        createActivity,
        areCategoriesLoading,
        areActivitiesLoading,
        isCreatingCategory,
        isCreatingActivity
     }}>
      {children}
    </ActivityContext.Provider>
  );
};

