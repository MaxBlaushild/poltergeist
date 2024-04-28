import './AnswerSurvey.css';
import { useAPI } from '@poltergeist/contexts';
import { Survey, Submission, SubmissionAnswer } from '@poltergeist/types';
import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Button } from './shared/Button.tsx';
import useActivities from '../hooks/useActivities.ts';
import { Modal } from './shared/Modal.tsx';
import { useSurveys } from '../hooks/useSurveys.ts';

interface ActivitySelection {
  [key: string]: boolean;
}

const stepTexts: string[] = ['Get ready', '3', '2', '1', 'Start!'];

export const AnswerSurvey: React.FC = () => {
  const { id } = useParams();
  const { apiClient } = useAPI();
  const [answers, setAnswers] = useState<ActivitySelection>({});
  const [surveyStarted, setSurveyStarted] = useState<boolean>(false);
  const [currentStepTextIndex, setCurrentStepTextIndex] = useState(-1);
  const [currentActivityIndex, setCurrentActivityIndex] = useState(0);
  const { loading, activities, error } = useActivities();
  const [helloWorldTop, setHelloWorldTop] = useState(-100); // Start off-screen
  const [selectedActivityIds, setSelectedActivityIds] = useState<string[]>([]);
  const { survey, submission, surveyLoading, surveyError, submissionLoading, submissionError } = useSurveys(id!);
  const [submissionSuccess, setSubmissionSuccess] = useState<boolean>(false);

  useEffect(() => {
    let timeout;
    if (surveyStarted) {
      const updateStep = (index) => {
        if (index < stepTexts.length + 1) {
          setCurrentStepTextIndex(index);
          timeout = setTimeout(() => updateStep(index + 1), 1000);
        }
      };
      updateStep(-1);
    }
    return () => {
      if (timeout) clearTimeout(timeout);
    };
  }, [surveyStarted, setCurrentStepTextIndex]);

  useEffect(() => {
    if (
      surveyStarted &&
      activities.length > 0 &&
      currentActivityIndex < activities.length &&
      currentStepTextIndex >= stepTexts.length
    ) {
      const interval = setInterval(
        () => {
          const activity = activities[currentActivityIndex];
          const newDiv = document.createElement('div');
          newDiv.textContent = activity.title;
          newDiv.style.position = 'absolute';
          newDiv.style.top = '0';

          // Randomize the starting x alignment
          const randomX = Math.floor(Math.random() * (window.innerWidth - 100)); // Subtract 100 to ensure it doesn't overflow the screen width
          newDiv.style.left = `${randomX}px`;

          newDiv.style.transition = 'top 3s ease-in, transform 0.3s'; // Increased the duration of the 'top' transition from 2s to 3s
          newDiv.style.padding = '16px';
          newDiv.style.backgroundColor = 'white';
          newDiv.style.border = '3px solid black';
          newDiv.style.borderRadius = '8px';
          newDiv.style.boxSizing = 'border-box';
          newDiv.style.display = 'flex';
          newDiv.style.flexDirection = 'column';
          newDiv.style.alignItems = 'center';
          newDiv.style.textAlign = 'center';
          newDiv.style.cursor = 'pointer';

          document.body.appendChild(newDiv);

          // Add event listener for all click-like events (click, touchend)
          const handleClickLikeEvents = (event) => {
            console.log('Element clicked');
            newDiv.style.transform = 'scale(1.4)';
            setSelectedActivityIds((prevIds) => [...prevIds, activity.id]); // Add activity id to selectedActivityIds
            setTimeout(() => {
              newDiv.style.transform = 'scale(1)';
              setTimeout(() => {
                newDiv.remove(); // Remove the div from the DOM when it is no longer needed
              }, 300);
            }, 300);
          };

          newDiv.addEventListener('click', handleClickLikeEvents);
          newDiv.addEventListener('touchend', handleClickLikeEvents);

          setTimeout(() => {
            newDiv.style.top = `${window.innerHeight}px`;
            setTimeout(() => {
              newDiv.remove(); // Ensure the div is removed when it reaches the bottom of the screen
            }, 3000); // Match the duration of the 'top' transition
          }, 50);

          setCurrentActivityIndex((prevIndex) => {
            if (prevIndex >= activities.length - 1) {
              clearInterval(interval);
              return prevIndex; // Keep the index at the last activity instead of resetting to 0
            }
            return prevIndex + 1;
          });
        },
        250 + Math.random() * 750
      );
      return () => clearInterval(interval);
    }
  }, [surveyStarted, activities, currentActivityIndex, currentStepTextIndex]);

  const currentStepText =
    currentStepTextIndex > -1 ? stepTexts[currentStepTextIndex] : undefined;
  const isSurveyFinished = currentActivityIndex === activities.length - 1;

  return (
    <div className="AnswerSurvey__background">
      {survey && !isSurveyFinished ? (
        <div
          className={`AnswerSurvey__modal ${surveyStarted ? 'animate-out' : ''}`}
        >
          <h2 className="AnswerSurvey__title">Greetings adventurer!</h2>
          <p className="AnswerSurvey__scrollText">
            You have been recruited by{' '}
            {survey?.user ? survey?.user?.name || 'Max' : 'Max'} to sail the
            high seas together.
          </p>
          <p className="AnswerSurvey__scrollText">
            Press start, then tap the things that you're interested in doing.
          </p>
          <Button title="Start" onClick={() => setSurveyStarted(true)} />
        </div>
      ) : null}
      {surveyStarted && currentStepText ? (
        <Modal>
          <h1>{currentStepText}</h1>
        </Modal>
      ) : null}
      {isSurveyFinished && !submissionSuccess ? (
        <div className="AnswerSurvey__activityList">
          <h2>Please take a second to scroll through and make sure you got everything right</h2>
          <div className="AnswerSurvey__activityListItems">
          {activities.map((activity) => (
            <div
              onClick={() => setSelectedActivityIds(prevIds => {
                const index = prevIds.indexOf(activity.id);
                if (index > -1) {
                  return prevIds.filter(id => id !== activity.id);
                } else {
                  return [...prevIds, activity.id];
                }
              })}
              className={
                selectedActivityIds.includes(activity.id)
                  ? 'AnswerSurvey__activity--selected'
                  : 'AnswerSurvey__activity'
              }
              key={activity.id}
            >
              {activity.title}
            </div>
          ))}
          </div>
          <Button title="Submit" onClick={async () => {
            const activityIds: string[] = [];
            const downs: boolean[] = [];

            activities.forEach(activity => {
              activityIds.push(activity.id);

              if (selectedActivityIds.includes(activity.id)) {
                downs.push(true);
              } else {
                downs.push(false);
              }
            });
            const submission = await apiClient.post<Submission>(`/sonar/surveys/${id}/submissions`, {
              activityIds,
              downs,
            });
            setSubmissionSuccess(true);
          }} />
        </div>
      ) : null}
      {submissionSuccess ? (
        <Modal>
          <h1>Thank you for your submission!</h1>
        </Modal>
      ) : null}
    </div>
  );
};
