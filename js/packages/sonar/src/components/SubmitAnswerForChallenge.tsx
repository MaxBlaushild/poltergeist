import React, { useState } from 'react';
import './SubmitAnswerForChallenge.css';
import { PointOfInterestChallenge } from '@poltergeist/types/dist/pointOfInterestChallenge';
import { Button } from './shared/Button.tsx';
import {
  CapturePointOfInterestResponse,
  useMatchContext,
} from '../contexts/MatchContext.tsx';
import {
  XMarkIcon,
  CheckBadgeIcon,
  XCircleIcon,
} from '@heroicons/react/20/solid';
import Divider from './shared/Divider.tsx';
import { PointOfInterestChallengeSubmission } from '@poltergeist/types/dist/pointOfInterestChallengeSubmission';
import { Oval } from 'react-loader-spinner';

type SubmitAnswerForChallengeProps = {
  challenge: PointOfInterestChallenge;
  onSubmit: () => void;
};

export const SubmitAnswerForChallenge = (
  props: SubmitAnswerForChallengeProps
) => {
  const { usersTeam, attemptCapturePointOfInterest } = useMatchContext();
  const [textSubmission, setTextSubmission] = useState<string | undefined>(
    undefined
  );
  const [imageSubmission, setImageSubmission] = useState<File | undefined>(
    undefined
  );
  const [judgement, setJudgement] = useState<
    CapturePointOfInterestResponse | undefined
  >(undefined);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  return (
    <div className="w-full rounded-xl flex flex-col gap-3">
      {!judgement && !isLoading && (
        <>
          <p className="text-md text-left">{props.challenge.question}</p>
          <textarea
            className="w-full h-24"
            value={textSubmission}
            onChange={(e) => setTextSubmission(e.target.value)}
          />
          <input
            id="file"
            type="file"
            className="w-full"
            onChange={(e) => setImageSubmission(e.target.files?.[0])}
          />
          <Button
            title={
              props.challenge.pointOfInterestChallengeSubmissions?.some(
                (submission) => submission.isCorrect
              )
                ? 'Ya Got Scooped'
                : 'Submit Answer'
            }
            disabled={props.challenge.pointOfInterestChallengeSubmissions?.some(
              (submission) => submission.isCorrect
            )}
            onClick={async () => {
              if (usersTeam) {
                try {
                  setIsLoading(true);
                  try {
                    const judgement = await attemptCapturePointOfInterest(
                      usersTeam.id,
                      props.challenge.id,
                      textSubmission,
                      imageSubmission
                    );
                    setJudgement(judgement);
                  } catch (e) {
                    console.log(e);
                  } finally {
                    setIsLoading(false);
                    setTextSubmission(undefined);
                    setImageSubmission(undefined);
                  }

                } catch {}
              }
            }}
          />
        </>
      )}
      {isLoading && (
        <div className="flex justify-center items-center mt-10 mb-10">
          <Oval
            visible={true}
          height="80"
          width="80"
          color={'#fa9eb5'}
          ariaLabel="oval-loading"
          wrapperStyle={{}}
          wrapperClass=""
        />
        </div>
      )}
      {judgement && (
        <>
        <XMarkIcon className="h-6 w-6" onClick={() => setJudgement(undefined)} />
          {judgement.judgement.judgement ? (
            <div className="flex flex-col items-center gap-3 mt-4 mb-4">
              <CheckBadgeIcon className="h-20 w-20 text-green-500" />
              <p>Correct</p>
            </div>
          ) : (
            <div className="flex flex-col items-center gap-3 mt-4 mb-4">
              <XCircleIcon className="h-20 w-20 text-red-500" />
              <p>{judgement.judgement.reason}</p>
            </div>
          )}
        </>
      )}
    </div>
  );
};
