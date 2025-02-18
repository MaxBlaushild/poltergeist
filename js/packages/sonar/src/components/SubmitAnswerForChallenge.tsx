import React, { useEffect, useState } from 'react';
import './SubmitAnswerForChallenge.css';
import { PointOfInterestChallenge } from '@poltergeist/types/dist/pointOfInterestChallenge';
import { Button } from './shared/Button.tsx';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import {
  XMarkIcon,
  CheckBadgeIcon,
  XCircleIcon,
} from '@heroicons/react/20/solid';
import Divider from './shared/Divider.tsx';
import { PointOfInterestChallengeSubmission } from '@poltergeist/types/dist/pointOfInterestChallengeSubmission';
import { Oval } from 'react-loader-spinner';
import { useInventory } from '@poltergeist/contexts';
import {
  ItemType,
  MatchInventoryItemEffect,
  OwnedInventoryItem,
  PointOfInterest,
  Team,
  User,
} from '@poltergeist/types';
import { scrambleAndObscureWords } from '../utils/scrambleSentences.ts';
import {
  CapturePointOfInterestResponse,
  useSubmissionsContext,
} from '../contexts/SubmissionsContext.tsx';
import { mapCaptureTiers } from '../utils/mapCaptureTiers.ts';

type SubmitAnswerForChallengeProps = {
  pointOfInterest: PointOfInterest;
  challenge: PointOfInterestChallenge;
  onSubmit: (immediate: boolean) => void;
};

export const SubmitAnswerForChallenge = (
  props: SubmitAnswerForChallengeProps
) => {
  const { match, usersTeam } = useMatchContext();
  const { submissions } = useSubmissionsContext();
  const { inventoryItems, consumeItem, setUsedItem, ownedInventoryItems } =
    useInventory();
  const completedForTier = mapCaptureTiers(props.pointOfInterest, submissions);
  const { createSubmission } = useSubmissionsContext();
  const [textSubmission, setTextSubmission] = useState<string | undefined>(
    undefined
  );
  const [imageSubmission, setImageSubmission] = useState<File | undefined>(
    undefined
  );
  const [correctness, setCorrectness] = useState<boolean | undefined>(
    undefined
  );
  const [reason, setReason] = useState<string | undefined>(undefined);
  const [isLoading, setIsLoading] = useState<boolean>(false);

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

  const matchingInventoryItem = ownedInventoryItems?.find(
    (item) =>
      item.inventoryItemId === matchingRubyForChallenge?.id && item.quantity > 0
  );

  const isGoldenMonkeyActive = match?.inventoryItemEffects.some(
    (item) =>
      item.inventoryItemId === ItemType.CipherOfTheLaughingMonkey &&
      item.teamId !== usersTeam?.id &&
      new Date(item.expiresAt) > new Date()
  );

  const completed = completedForTier[props.challenge.tier] ?? false;

  return (
    <div className="w-full rounded-xl flex flex-col gap-3">
      {correctness === undefined && !isLoading && (
        <>
          <p className="text-md text-left">
            {isGoldenMonkeyActive
              ? scrambleAndObscureWords(
                  props.challenge.question,
                  usersTeam?.id ?? ''
                )
              : props.challenge.question}
          </p>
          {!completed && (
            <textarea
              className="w-full h-24"
              value={textSubmission}
              onChange={(e) => setTextSubmission(e.target.value)}
            />
          )}
          {!completed && (
            <input
              id="file"
              type="file"
              className="w-full"
              onChange={(e) => setImageSubmission(e.target.files?.[0])}
            />
          )}
          <div className="flex flex-row justify-between gap-2">
            <Button
              title={completed ? 'Locked' : 'Submit Answer'}
              disabled={
                completed ||
                props.challenge.pointOfInterestChallengeSubmissions?.some(
                  (submission) => submission.isCorrect
                )
              }
              onClick={async () => {
                setIsLoading(true);
                try {
                  const result = await createSubmission(
                    props.challenge.id,
                    textSubmission,
                    imageSubmission
                  );
                  setCorrectness(result?.correctness ?? false);
                  setReason(result?.reason ?? 'Failed for an unknown reason.');
                  props.onSubmit(false);
                  setTimeout(() => {
                    setCorrectness(undefined);
                    setReason(undefined);
                  }, 3000);
                } catch (e) {
                  console.log(e);
                } finally {
                  setIsLoading(false);
                  setTextSubmission(undefined);
                  setImageSubmission(undefined);
                }
              }}
            />
            {!!matchingInventoryItem && !completed && (
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
                }}
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
      {correctness !== undefined && (
        <>
          <XMarkIcon
            className="h-6 w-6"
            onClick={() => {
              setCorrectness(undefined);
              setReason(undefined);
            }}
          />
          {correctness ? (
            <div className="flex flex-col items-center gap-3 mt-4 mb-4">
              <CheckBadgeIcon className="h-20 w-20 text-green-500" />
              <p>Correct</p>
            </div>
          ) : (
            <div className="flex flex-col items-center gap-3 mt-4 mb-4">
              <XCircleIcon className="h-20 w-20 text-red-500" />
              <p>{reason}</p>
            </div>
          )}
        </>
      )}
    </div>
  );
};
