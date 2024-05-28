import './NewSurvey.css';
import React, { useContext, useEffect, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import { Category, Activity, Survey } from '@poltergeist/types';
import { redirect, useNavigate } from 'react-router-dom';
import { Button } from './shared/Button.tsx';
import { LameActivitySelector } from './shared/LameActivitySelector.tsx';
import useActivities from '../hooks/useActivities.ts';
import { Modal, ModalSize } from './shared/Modal.tsx';
import TextInput from './shared/TextInput.tsx';
import { generateRandomName } from '../utils/generateName.ts';
import ActivityCloud from './shared/ActivityCloud.tsx';
import useCategories from '../hooks/useCategories.ts';
import { ChevronUpIcon, ChevronDownIcon } from '@heroicons/react/20/solid';
import Divider from './shared/Divider.tsx';
import { Scroll, ScrollAwayTime } from './shared/Scroll.tsx';
import { useSurveys } from '../hooks/useSurveys.ts';

const ConfirmationContent: React.FC = () => {
  return <div>Thank you for your service</div>;
};

export const NewSurvey: React.FC = () => {
  const { loading, activities, error } = useActivities();
  const { surveys, isLoading: surveysLoading } = useSurveys();
  const { categories } = useCategories();
  const { apiClient } = useAPI();
  const [name, setName] = useState('Untitled survey');
  const navigate = useNavigate();
  const [selectedActivityIds, setSelectedActivityIds] = useState<string[]>([]);
  const [selectedCategoryIds, setSelectedCategoryIds] = useState<string[]>([]);
  const [shouldStartSurvey, setShouldStartSurvey] = useState<boolean>(false);
  const [shouldShowForm, setShouldShowForm] = useState<boolean>(false);

  const handleCategorySelect = (categoryId: string) => {
    setSelectedCategoryIds((prevIds) => {
      const index = prevIds.indexOf(categoryId);
      if (index > -1) {
        return prevIds.filter((id) => id !== categoryId);
      } else {
        return [...prevIds, categoryId];
      }
    });
  };

  const handleActivitySelect = (activityId: string) => {
    setSelectedActivityIds((prevIds) => {
      const index = prevIds.indexOf(activityId);
      if (index > -1) {
        return prevIds.filter((id) => id !== activityId);
      } else {
        return [...prevIds, activityId];
      }
    });
  };

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

  useEffect(() => {
    if (shouldStartSurvey) {
      setTimeout(() => {
        setShouldShowForm(true);
      }, ScrollAwayTime);
    }
  }, [shouldStartSurvey]);

  return (
    <div className="NewSurvey__background">
      {!surveysLoading && surveys.length === 0 && <Scroll shouldScrollOut={shouldStartSurvey}>
        <p className='font-bold text-sm/4'>Create a survey to send to your friends to see what they want to do.</p>
        <p className='text-sm/4'>Select activities that you'd like to see if they're interested in and we'll give you a link to share out.</p>
        <Button title="Get started" onClick={() => setShouldStartSurvey(true)} />
      </Scroll>}
      {((!surveysLoading && surveys.length > 0) || shouldShowForm) && <Modal size={ModalSize.FULLSCREEN}>
        <div className="flex flex-col items-start w-full gap-8 mt-4">
          <TextInput
            value={name}
            onChange={setName}
            placeholder="Summons title"
          />
          <Divider />
          <p>
            Select the activities you'd like to see if you're friends are
            interested in
          </p>
          <LameActivitySelector
            selectedActivityIds={selectedActivityIds}
            onSelect={handleActivitySelect}
          />
          <Button
            disabled={selectedActivityIds.length === 0}
            title="Create summons"
            onClick={async () => {
              const survey = await apiClient.post<Survey>(`/sonar/surveys`, {
                activityIds: selectedActivityIds,
                name: name,
              });
              navigate(`/surveys/${survey.id}`);
            }}
          />
        </div>
      </Modal>}
    </div>
  );
};
