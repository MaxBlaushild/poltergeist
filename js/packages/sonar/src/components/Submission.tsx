import './Survey.css';
import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Survey as SurveyType } from '@poltergeist/types';
import axios from 'axios';
import { useAPI } from '@poltergeist/contexts';
import useSurvey from '../hooks/useSurvey.ts';
import { Modal, ModalSize } from './shared/Modal.tsx';
import { Button } from './shared/Button.tsx';
import ActivityCloud from './shared/ActivityCloud.tsx';
import { LameActivitySelector } from './shared/LameActivitySelector.tsx';
import Divider from './shared/Divider.tsx';
import useSubmission from '../hooks/useSubmission.ts';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';

interface SubmissionParams {
  id: string;
}

export const Submission: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { submission, loading, error } = useSubmission(id!);
  const navigate = useNavigate();

  return (
    <div className="Survey__background">
      {submission ? (
        <Modal size={ModalSize.FULLSCREEN}>
          <div className="flex flex-col gap-8 mt-4">
            <div className="flex flex-col gap-2">

              <img
                src={submission.user.profile.profilePictureUrl || 'default-profile.png'}
                alt={`${submission.user.name}'s profile`}
                className="h-20 w-20 rounded-full self-center"
              />
              <h1 className="text-2xl font-bold">{submission.user.name}</h1>
              <p>{new Date(submission.createdAt).toLocaleDateString()}</p>
              <p>
                {submission.answers.length} answer
                {submission.answers.length === 1 ? '' : 's'}
              </p>
            </div>
            <Divider />
            <LameActivitySelector
              openByDefault
              selectedActivityIds={submission.answers.map(
                (answer) => answer.activityId
              )}
            />
          </div>
        </Modal>
      ) : null}
    </div>
  );
};
