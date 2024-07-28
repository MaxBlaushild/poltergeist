import './Survey.css';
import React, { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Survey as SurveyType } from '@poltergeist/types';
import axios from 'axios';
import { useAPI } from '@poltergeist/contexts';
import { useSurvey } from '../hooks/useSurvey.ts';
import { Modal, ModalSize } from './shared/Modal.tsx';
import { Button } from './shared/Button.tsx';
import ActivityCloud from './shared/ActivityCloud.tsx';
import { LameActivitySelector } from './shared/LameActivitySelector.tsx';
import Divider from './shared/Divider.tsx';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';

interface SurveyParams {
  id: string;
}

export const Survey: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { survey } = useSurvey(id!);
  const [toastText, setToastText] = useState<string | null>(null);
  const navigate = useNavigate();
  const { userProfiles } = useUserProfiles();

  return (
    <div className="Survey__background">
      {survey ? (
        <Modal size={ModalSize.FULLSCREEN}>
          <div className="flex flex-col gap-8 mt-4">
            <div className="flex flex-col gap-2">
              <h1 className="text-2xl font-bold">{survey.title}</h1>
              <p>{new Date(survey.createdAt).toLocaleDateString()}</p>
              <p>
                {survey.surveySubmissions.length} submission
                {survey.surveySubmissions.length === 1 ? '' : 's'}
              </p>
              <Button
                title="Copy share link"
                onClick={() => {
                  navigator.clipboard.writeText(
                    `${window.location.origin}/submit-answer/${id}`
                  );

                  setToastText('Link copied to clipboard');
                  setTimeout(() => {
                    setToastText(null);
                  }, 1500);
                }}
              />
            </div>
            {survey.surveySubmissions.length > 0 ? <Divider /> : null}
            {survey.surveySubmissions.length > 0 ? (
              <ul className="w-full">
                {survey.surveySubmissions
                  .sort(
                    (a, b) =>
                      new Date(b.createdAt).getTime() -
                      new Date(a.createdAt).getTime()
                  )
                  .map((submission) => (
                    <li
                      key={submission.id}
                      className="relative rounded-md p-3 text-sm/6 transition hover:bg-black/5 flex items-center"
                    >
                      <img
                        src={userProfiles?.find(profile => profile.vieweeId === submission.user.id)?.profilePictureUrl || 'default-profile.png'}
                        alt={`${submission.user.name}'s profile`}
                        className="h-9 w-9 rounded-full mr-4"
                      />
                      <div className="flex flex-col">
                        <a
                          href={`/submissions/${submission.id}`}
                          className="font-semibold text-black text-left"
                        >
                          <span className="absolute inset-0" />
                          {submission.user.name}
                        </a>
                        <ul
                          className="flex gap-2 text-black/50"
                          aria-hidden="true"
                        >
                          <li>
                            {new Date(submission.createdAt).toLocaleDateString()}
                          </li>
                          <li aria-hidden="true">&middot;</li>
                          <li>
                            {submission.answers.length} answer
                            {submission.answers.length === 1 ? '' : 's'}
                          </li>
                        </ul>
                      </div>
                    </li>
                  ))}
              </ul>
            ) : null}
            <Divider />
            <LameActivitySelector
              openByDefault
              selectedActivityIds={survey.activities.map(
                (activity) => activity.id
              )}
            />
          </div>
        </Modal>
      ) : null}
      {toastText ? <Modal size={ModalSize.TOAST}>{toastText}</Modal> : null}
    </div>
  );
};
