import './FunActivitySelector.css';
import { useAPI } from '@poltergeist/contexts';
import { Activity, Survey } from '@poltergeist/types';
import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Button } from './Button.tsx';
import { Modal, ModalSize } from './Modal.tsx';
import { useSurvey } from '../../hooks/useSurvey.ts';
import ActivityCloud from './ActivityCloud.tsx';
import { LameActivitySelector } from './LameActivitySelector.tsx';
import Divider from './Divider.tsx';
import { Scroll } from './Scroll.tsx';

interface ActivitySelection {
  [key: string]: boolean;
}

interface ActivitySelectorProps {
  name: string;
  survey: Survey;
  confirmationContent: React.ReactNode;
  activities: Activity[];
  onActivitiesSelected: (activityIds: string[]) => Promise<void>;
}

const stepTexts: string[] = ['Get ready', '3', '2', '1', 'Start!'];

export const FunActivitySelector: React.FC<ActivitySelectorProps> = ({
  activities,
  onActivitiesSelected,
  confirmationContent,
  name,
}: ActivitySelectorProps) => {
  const { id } = useParams();
  const { apiClient } = useAPI();
  const [answers, setAnswers] = useState<ActivitySelection>({});
  const [surveyStarted, setSurveyStarted] = useState<boolean>(false);
  const [currentStepTextIndex, setCurrentStepTextIndex] = useState(-1);
  const [currentActivityIndex, setCurrentActivityIndex] = useState(0);
  const [helloWorldTop, setHelloWorldTop] = useState(-100); // Start off-screen
  const [selectedActivityIds, setSelectedActivityIds] = useState<string[]>([]);
  const [submissionSuccess, setSubmissionSuccess] = useState<boolean>(false);
  const [shouldShowActivities, setShouldShowActivities] = useState(false);
  const [shouldShowSkipButton, setShouldShowSkipButton] = useState(false);
  const [collectedDivs, setCollectedDivs] = useState<HTMLDivElement[]>([]);

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

      setTimeout(() => {
        setShouldShowSkipButton(true);
      }, 10000);
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
      currentStepTextIndex >= stepTexts.length &&
      !shouldShowActivities
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

          newDiv.style.transition = 'top 6s ease-in, transform 0.3s'; // Increased the duration of the 'top' transition from 2s to 3s
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

          setCollectedDivs(prevDivs => [...prevDivs, newDiv]);

          // Add event listener for all click-like events (click, touchend)
          const handleClickLikeEvents = (event) => {
            newDiv.style.transform = 'scale(1.4)';
            setSelectedActivityIds((prevIds) => {
              if (!prevIds.includes(activity.id)) {
                return [...prevIds, activity.id];
              }
              return prevIds;
            }); // Add activity id to selectedActivityIds if not already present
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
              if (currentActivityIndex === activities.length - 1) {
                setTimeout(() => {
                  setShouldShowActivities(true);
                }, 300);
              }
            }, 6000); // Match the duration of the 'top' transition
          }, 50);

          setCurrentActivityIndex((prevIndex) => {
            if (prevIndex >= activities.length - 1) {
              clearInterval(interval);
              return prevIndex; // Keep the index at the last activity instead of resetting to 0
            }
            return prevIndex + 1;
          });
        },
        500 + Math.random() * 1000
      );
      return () => clearInterval(interval);
    }
  }, [surveyStarted, activities, currentActivityIndex, currentStepTextIndex, shouldShowActivities, setCollectedDivs]);

  const currentStepText =
    currentStepTextIndex > -1 ? stepTexts[currentStepTextIndex] : undefined;

  return (
    <div className="ActivitySelector__background">
      {activities && !shouldShowActivities ? (
          <Scroll shouldScrollOut={surveyStarted}>
          <div className="AnswerSurvey__scrollContent">
      <p className="AnswerSurvey__scrollText text-sm/4 font-bold">
        {name} wants to
        hang out with you!
      </p>
      <p className="AnswerSurvey__scrollText text-sm/4">
        Press start, then tap the things that you're interested in doing.
      </p>
      <p className="AnswerSurvey__scrollText text-sm/4">
        You'll have a chance to touch up your results at the end.
      </p>
    </div>
            <Button title="Start" onClick={() => setSurveyStarted(true)} />
          </Scroll>

      ) : null}
      {surveyStarted && currentStepText ? (
        <Modal>
          <h1>{currentStepText}</h1>
        </Modal>
      ) : null}
      {shouldShowActivities && !submissionSuccess ? (
        <Modal size={ModalSize.FULLSCREEN}>
          <div className='flex flex-col items-start w-full gap-8 mt-4'>
          <h2 className='text-center font-bold'>
            Please take a moment to review your selections
          </h2>
          <Divider />
          <LameActivitySelector
            selectedActivityIds={selectedActivityIds}
            activitiesToFilterBy={activities.map((activity) => activity.id)}
            onSelect={(activityId) =>
              setSelectedActivityIds((prevIds) => {
                const index = prevIds.indexOf(activityId);
                if (index > -1) {
                  return prevIds.filter((id) => id !== activityId);
                } else {
                  return [...prevIds, activityId];
                }
              })
            }
          />
          <Button
            title="Submit"
            onClick={async () => {
              await onActivitiesSelected(selectedActivityIds);
              setSubmissionSuccess(true);
            }}
          />
          </div>
        </Modal>
      ) : null}
      {submissionSuccess ? <Modal>{confirmationContent}</Modal> : null}
      {shouldShowSkipButton && !shouldShowActivities ? (
        <div className='FunActivitySelector__skip'>
              <Button title="Skip" onClick={() => {
                setShouldShowActivities(true)
                collectedDivs.forEach(div => div.remove());
                setCollectedDivs([]);
              }} />
        </div>
      ) : null}
    </div>
  );
};
