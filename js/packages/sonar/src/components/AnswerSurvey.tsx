import './AnswerSurvey.css';
import { useAPI } from '@poltergeist/contexts';
import { Survey, Submission, SubmissionAnswer } from '@poltergeist/types';
import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Button } from './shared/Button.tsx';
import useActivities from '../hooks/useActivities.ts';
import { useSurvey } from '../hooks/useSurvey.ts';
import { FunActivitySelector } from './shared/FunActivitySelector.tsx';

const ConfirmationContent: React.FC = () => {
  return <h1>Thank you for your submission!</h1>;
};

export const AnswerSurvey: React.FC = () => {
  const { apiClient } = useAPI();
  const { id } = useParams();
  const { survey } = useSurvey(id!);
  const activities = survey?.activities || [];
  return (
    survey ? <FunActivitySelector
      name={survey?.user?.name || 'Someone'}
      survey={survey}
      activities={activities}
      confirmationContent={<ConfirmationContent />}
      onActivitiesSelected={async (activityIds) => {
        const downs: boolean[] = [];

        activities.forEach((activity) => {
          if (activityIds.includes(activity.id)) {
            downs.push(true);
          } else {
            downs.push(false);
          }
        });

        const submission = await apiClient.post<Submission>(
          `/sonar/surveys/${id}/submissions`,
          {
            activityIds,
            downs,
          }
        );
      }}
    /> : null
  );
};
