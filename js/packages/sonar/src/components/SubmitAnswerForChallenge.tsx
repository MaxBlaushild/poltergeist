import React, { useEffect, useState } from 'react';
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
import { useInventory } from '../contexts/InventoryContext.tsx';
import { ItemType } from '@poltergeist/types';
import { scrambleAndObscureWords } from '../utils/scrambleSentences.ts';

type SubmitAnswerForChallengeProps = {
  challenge: PointOfInterestChallenge;
  onSubmit: (immediate: boolean) => void;
};

export const SubmitAnswerForChallenge = (
  props: SubmitAnswerForChallengeProps
) => {
  const { usersTeam, attemptCapturePointOfInterest, getCurrentMatch, match } =
    useMatchContext();
  const { inventoryItems, consumeItem, setPresentedInventoryItem, setUsedItem } = useInventory();
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

  const hasBeenAnsweredCorrectly =
    props.challenge.pointOfInterestChallengeSubmissions?.some(
      (submission) => submission.isCorrect
    );

  let matchingRubyForChallenge;
  switch (props.challenge.tier) {
    case 1:
      matchingRubyForChallenge = inventoryItems.find(
        (item) => item.id === ItemType.FlawedRuby
      );
      break;
    case 2:
      matchingRubyForChallenge = inventoryItems.find(
        (item) => item.id === ItemType.Ruby
      );
      break;
    case 3:
      matchingRubyForChallenge = inventoryItems.find(
        (item) => item.id === ItemType.BrilliantRuby
      );
      break;
  }

  const matchingInventoryItem = usersTeam?.teamInventoryItems?.find(
    (item) => item.inventoryItemId === matchingRubyForChallenge?.id && item.quantity > 0
  );

  const isGoldenMonkeyActive = match?.inventoryItemEffects.some(
    (item) =>
      item.inventoryItemId === ItemType.CipherOfTheLaughingMonkey &&
      item.teamId !== usersTeam?.id &&
      new Date(item.expiresAt) > new Date()
  );

  useEffect(() => {
    if (judgement) {
      if (judgement.judgement.judgement.judgement) {
        setPresentedInventoryItem(judgement.item);
      }
    }
  }, [judgement]);

  return (
    <div className="w-full rounded-xl flex flex-col gap-3">
      {!judgement && !isLoading && (
        <>
          <p className="text-md text-left">
            {isGoldenMonkeyActive
              ? scrambleAndObscureWords(
                  props.challenge.question,
                  usersTeam?.id ?? ''
                )
              : props.challenge.question}
          </p>
          {!hasBeenAnsweredCorrectly && (
            <textarea
              className="w-full h-24"
              value={textSubmission}
              onChange={(e) => setTextSubmission(e.target.value)}
            />
          )}
          {!hasBeenAnsweredCorrectly && (
            <input
              id="file"
              type="file"
              className="w-full"
              onChange={(e) => setImageSubmission(e.target.files?.[0])}
            />
          )}
          <div className="flex flex-row justify-between gap-2">
            <Button
              title={hasBeenAnsweredCorrectly ? 'Locked' : 'Submit Answer'}
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
                      props.onSubmit(false);
                      setTimeout(() => {
                        setJudgement(undefined);
                      }, 3000);
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
            {!!matchingInventoryItem && !hasBeenAnsweredCorrectly && (
              <img
                src={matchingRubyForChallenge?.imageUrl}
                alt={matchingRubyForChallenge?.name}
                className="rounded-lg border-black border-2 h-12 w-12"
                onClick={() => {
                  consumeItem(matchingInventoryItem?.id, {
                    pointOfInterestId: props.challenge.pointOfInterestId,
                  });
                  setUsedItem(matchingRubyForChallenge!);
                  props.onSubmit(true);
                }
                }
              />
            )}
          </div>
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
          <XMarkIcon
            className="h-6 w-6"
            onClick={() => setJudgement(undefined)}
          />
          {judgement.judgement.judgement.judgement ? (
            <div className="flex flex-col items-center gap-3 mt-4 mb-4">
              <CheckBadgeIcon className="h-20 w-20 text-green-500" />
              <p>Correct</p>
            </div>
          ) : (
            <div className="flex flex-col items-center gap-3 mt-4 mb-4">
              <XCircleIcon className="h-20 w-20 text-red-500" />
              <p>{judgement.judgement.judgement.reason}</p>
            </div>
          )}
        </>
      )}
    </div>
  );
};
