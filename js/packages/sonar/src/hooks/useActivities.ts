import { useState, useEffect } from 'react';
import axios from 'axios';
import { useAPI } from '@poltergeist/contexts';
import { Activity } from '@poltergeist/types';

export const useActivities = () => {
  const { apiClient } = useAPI();
  const [activities, setActivities] = useState<Activity[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchActivities = async () => {
      try {
        const response = await apiClient.get<Activity[]>('/sonar/activities');
        setActivities(response);
        // setActivities([response[0], response[1], response[2]]);
        setLoading(false);
      } catch (err) {
        setError(err);
        setLoading(false);
      }
    };

    fetchActivities();
  }, []);

  return { activities, loading, error };
};

export default useActivities;
