import { ChevronDownIcon, ChevronUpIcon } from '@heroicons/react/20/solid';
import './LameActivitySelector.css';
import React, { useState } from 'react';
import { Activity } from '@poltergeist/types';
import { Button } from './Button.tsx';
import TextInput from './TextInput.tsx';
import { useActivityContext } from '../../contexts/ActivityContext.tsx';

interface ActivitySelectorProps {
  selectedActivityIds: string[];
  onSelect?: (activityId: string) => void;
  openByDefault?: boolean;
  activitiesToFilterBy?: string[];
}

export const LameActivitySelector = ({
  selectedActivityIds,
  onSelect,
  openByDefault = false,
  activitiesToFilterBy,
}: ActivitySelectorProps) => {
  const { categories, createActivity, isCreatingActivity } = useActivityContext();
  const [selectedCategoryIds, setSelectedCategoryIds] = useState<string[]>([]);
  const [newActivityTitles, setNewActivityTitles] = useState<{ [categoryId: string]: string }>({});

  const toggleCategory = (categoryId: string) => {
    setSelectedCategoryIds((prevIds) => {
      const index = prevIds.indexOf(categoryId);
      if (index > -1) {
        return prevIds.filter((id) => id !== categoryId);
      } else {
        return [...prevIds, categoryId];
      }
    });
  };

  const getNewActivityTitleForCategory = (categoryId: string): string => {
    return newActivityTitles[categoryId] || '';
  };

  return (
    <>
      {categories.map((category) => {
        const isSelected = selectedCategoryIds.includes(category.id);
        const isOpen = openByDefault ? !isSelected : isSelected;
        const activitiesLeftOver = activitiesToFilterBy ? category.activities.filter((activity) => activitiesToFilterBy.includes(activity.id)) : category.activities;

        if (activitiesLeftOver?.length === 0) {
          return null;
        }

        return (
          <div key={category.id} className="w-full">
            <div
              className="flex justify-between items-center p-2 cursor-pointer"
              onClick={() => toggleCategory(category.id)}
            >
              <h3>{category.title}</h3>
              {isOpen ? (
                <ChevronUpIcon className="h-5 w-5 text-black" />
              ) : (
                <ChevronDownIcon className="h-5 w-5 text-black" />
              )}
            </div>
            {isOpen && (
              <div className="pl-4">
                {activitiesLeftOver?.filter((activity) => activity.categoryId === category.id)
                  .map((activity) => (
                    <div
                      key={activity.id}
                      className="flex justify-between items-center p-2"
                    >
                      <span>{activity.title}</span>
                      <input
                        type="checkbox"
                        disabled={!onSelect}
                        className="NewSurvey__checkbox"
                        checked={selectedActivityIds.includes(activity.id)}
                        onChange={() => onSelect && onSelect(activity.id)}
                      />
                    </div>
                  ))}
              {isOpen && (
                <div className="flex flex-col items-center space-y-2">
                  <TextInput
                    value={getNewActivityTitleForCategory(category.id)}
                    type="text"
                    placeholder="Activity Title"
                    onChange={(value) => setNewActivityTitles({ ...newActivityTitles, [category.id]: value })}
                  />
                  <Button
                    title='Add Activity'
                      className="mt-2 p-2 bg-blue-500 text-white rounded hover:bg-blue-600"
                      onClick={() => {
                        createActivity({
                          categoryId: category.id,
                          title: getNewActivityTitleForCategory(category.id),
                        });
                        setNewActivityTitles({ ...newActivityTitles, [category.id]: '' });
                      }}
                      disabled={isCreatingActivity}
                    />
                </div>
              )}
              </div>
            )}
          </div>
        );
      })}
    </>
  );
};
