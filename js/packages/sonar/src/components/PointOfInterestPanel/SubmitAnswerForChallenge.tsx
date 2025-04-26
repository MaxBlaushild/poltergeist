import React, { useEffect, useState } from 'react';
import { PointOfInterestChallenge } from '@poltergeist/types/dist/pointOfInterestChallenge';
import { Button } from '../shared/Button.tsx';
import {
  XMarkIcon,
  CheckBadgeIcon,
  XCircleIcon,
} from '@heroicons/react/20/solid';
import Divider from '../shared/Divider.tsx';
import { PointOfInterestChallengeSubmission } from '@poltergeist/types/dist/pointOfInterestChallengeSubmission';
import { Oval } from 'react-loader-spinner';
import { useInventory } from '@poltergeist/contexts';
import { useLocation } from '@poltergeist/contexts';
import {
  ItemType,
  Match,
  MatchInventoryItemEffect,
  OwnedInventoryItem,
  PointOfInterest,
  Team,
  User,
} from '@poltergeist/types';
import { scrambleAndObscureWords } from '../../utils/scrambleSentences.ts';
import {
  CapturePointOfInterestResponse,
  useSubmissionsContext,
} from '../../contexts/SubmissionsContext.tsx';
import { mapCaptureTiers } from '../../utils/mapCaptureTiers.ts';
import { useUserProfiles } from '../../contexts/UserProfileContext.tsx';
import { useQuestLogContext } from '../../contexts/QuestLogContext.tsx';

type SubmitAnswerForChallengeProps = {
  pointOfInterest: PointOfInterest;
  challenge: PointOfInterestChallenge;
  match?: Match | undefined;
  usersTeam?: Team | undefined;
  onSubmit: (immediate: boolean) => void;
};

export const SubmitAnswerForChallenge = (
  props: SubmitAnswerForChallengeProps
) => {
  const { submissions, setSubmissions } = useSubmissionsContext();
  const { currentUser } = useUserProfiles();
  const { location } = useLocation();
  const { inventoryItems, consumeItem, setUsedItem, ownedInventoryItems } =
    useInventory();
  const { refreshQuestLog } = useQuestLogContext();
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

  const isGoldenMonkeyActive = props.match?.inventoryItemEffects.some(
    (item) =>
      item.inventoryItemId === ItemType.CipherOfTheLaughingMonkey &&
      item.teamId !== props.usersTeam?.id &&
      new Date(item.expiresAt) > new Date()
  );

  const completed = completedForTier[props.challenge.tier] ?? false;

  // Calculate distance between user and POI using Haversine formula
  const isWithinRange = location?.latitude && location?.longitude && props.pointOfInterest.lat && props.pointOfInterest.lng ? (() => {
    const R = 6371e3; // Earth's radius in meters
    const φ1 = location?.latitude * Math.PI/180;
    const φ2 = parseFloat(props.pointOfInterest.lat) * Math.PI/180;
    const Δφ = (parseFloat(props.pointOfInterest.lat) - location?.latitude) * Math.PI/180;
    const Δλ = (parseFloat(props.pointOfInterest.lng) - location?.longitude) * Math.PI/180;

    const a = Math.sin(Δφ/2) * Math.sin(Δφ/2) +
            Math.cos(φ1) * Math.cos(φ2) *
            Math.sin(Δλ/2) * Math.sin(Δλ/2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a));
    const distance = R * c;

    return distance <= 100; // Within 100 meters
  })() : false;

  return (
    <div className="w-full rounded-xl flex flex-col gap-3">
      {correctness === undefined && !isLoading && (
        <>
          <p className="text-md text-left">
            {isGoldenMonkeyActive
              ? scrambleAndObscureWords(
                  props.challenge.question,
                  props.usersTeam?.id ?? ''
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
              title={completed ? 
                !props.usersTeam?.id ? 'Completed' : 'Locked' : 
                isWithinRange ? 
                  'Submit Answer' : 
                  'Too Far Away'}
              disabled={
                completed ||
                !isWithinRange ||
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
                    imageSubmission,
                    props.usersTeam?.id,
                    props.usersTeam ? undefined : currentUser?.id
                  );
                  setCorrectness(result?.correctness ?? false);
                  setReason(result?.reason ?? 'Failed for an unknown reason.');

                  if (result?.correctness) {
                    setSubmissions([...submissions, {
                      id: '',
                      isCorrect: true,
                      text: textSubmission ?? '',
                      teamId: props.usersTeam?.id,
                      userId: currentUser?.id,
                      pointOfInterestChallengeId: props.challenge.id,
                      imageUrl: imageSubmission?.name ?? '',
                      createdAt: new Date(),
                      updatedAt: new Date(),
                    }]);
                  }
                  props.onSubmit(false);
                  try {
                    refreshQuestLog();
                  } catch (e) {
                    console.log(e);
                  }

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
                onClick={async () => {
                  await consumeItem(matchingInventoryItem?.id, {
                    pointOfInterestId: props.challenge.pointOfInterestId,
                  });
                  setUsedItem(matchingRubyForChallenge!);
                  setSubmissions([...submissions, {
                    id: '',
                    isCorrect: true,
                    text: 'Used a ruby to answer the challenge',
                    teamId: props.usersTeam?.id,
                    userId: currentUser?.id,
                    pointOfInterestChallengeId: props.challenge.id,
                    imageUrl: '',
                    createdAt: new Date(),
                    updatedAt: new Date(),
                  }]);
                  try {
                    refreshQuestLog();
                  } catch (e) {
                    console.log(e);
                  }
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
