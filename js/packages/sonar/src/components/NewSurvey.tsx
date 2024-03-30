import React, { useContext, useEffect, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { Category, Activity, Survey } from '@poltergeist/types';
import { redirect } from 'react-router-dom';

interface ActivitySelection {
  [key: string]: boolean;
}

export const NewSurvey: React.FC = () => {
  const { apiClient } = useAPI();
  const [categories, setCategories] = useState<Category[]>([]);
  const [selectedActivities, setSelectedActivities] = useState<ActivitySelection>({});

  useEffect(() => {
    const fetchCategories = async () => {
      try {
        const fetchedCategories = await apiClient.get<Category[]>('/sonar/categories');
        const initialSelectedActivities = fetchedCategories.reduce((acc, category) => {
          category.activities.forEach((activity: Activity) => {
            acc[activity.id] = true; // Set all activities within each category as selected by default
          });
          return acc;
        }, {});
        setCategories(fetchedCategories);
        setSelectedActivities(initialSelectedActivities);
      } catch (error) {
        console.error('Failed to fetch categories and activities', error);
      }
    };

    fetchCategories();
  }, [apiClient]);

  const handleActivityChange = (activityId: string) => {
    setSelectedActivities((prevSelectedActivities) => ({
      ...prevSelectedActivities,
      [activityId]: !prevSelectedActivities[activityId],
    }));
  };

  const handleCategoryChange = (categoryId: string) => {
    const category = categories.find(category => category.id === categoryId);
    if (category) {
      const newSelectedActivities = { ...selectedActivities };
      let allSelected = true;
      category.activities.forEach((activity) => {
        if (!selectedActivities[activity.id]) {
          allSelected = false;
        }
      });
      category.activities.forEach((activity) => {
        newSelectedActivities[activity.id] = !allSelected;
      })
      setSelectedActivities(newSelectedActivities);
    }
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    try {
      const activityIds = Object.keys(selectedActivities).filter(key => selectedActivities[key]);
      const newSurvey = await apiClient.post<Survey>('/sonar/surveys', { activityIds });
      alert('Activities submitted successfully');
     window.location.href = `/surveys/${newSurvey.id}`;
    } catch (error) {
      console.error('Failed to submit activities', error);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      {categories.map((category) => (
        <div key={category.id}>
          <label>
            <strong>{category.title}</strong>
            <input
              type="checkbox"
              name="category"
              value={category.id}
              checked={category.activities.every(activity => !!selectedActivities[activity.id])}
              onChange={() => handleCategoryChange(category.id)}
            />
          </label>
          {category.activities.map((activity) => (
            <div key={activity.id}>
              <label>
                {activity.title}
                <input
                  type="checkbox"
                  name="activity"
                  value={activity.id}
                  checked={!!selectedActivities[activity.id]}
                  onChange={() => handleActivityChange(activity.id)}
                />
              </label>
            </div>
          ))}
          <br />
        </div>
      ))}
      <button type="submit">Submit</button>
    </form>
  );
};


