import {
  ItemType,
  PointOfInterest,
  Team,
  getControllingTeamForPoi,
  hasTeamDiscoveredPointOfInterest,
} from '@poltergeist/types';
import React, { useState } from 'react';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { PointOfInterestChallenge } from '@poltergeist/types/dist/pointOfInterestChallenge';
import { TabItem, TabNav } from './shared/TabNav.tsx';
import { SubmitAnswerForChallenge } from './SubmitAnswerForChallenge.tsx';
import { Button } from './shared/Button.tsx';
import { LockClosedIcon, LockOpenIcon } from '@heroicons/react/20/solid';
import { StatusIndicator } from './shared/StatusIndicator.tsx';
import { useInventory } from '../contexts/InventoryContext.tsx';
import { scrambleAndObscureWords } from '../utils/scrambleSentences.ts';

const toRoman = (num: number): string => {
  const lookup: { [key: number]: string } = {
    1: 'I',
    2: 'II',
    3: 'III',
    4: 'IV',
    5: 'V',
    6: 'VI',
    7: 'VII',
    8: 'VIII',
    9: 'IX',
    10: 'X',
  };
  return lookup[num] || '';
};

export const PointOfInterestPanel = ({
  pointOfInterest,
  allTeams,
  onClose
}: {
  pointOfInterest: PointOfInterest;
  allTeams: Team[];
  onClose: (immediate: boolean) => void;
}) => {
  const {
    unlockPointOfInterest,
    attemptCapturePointOfInterest,
    usersTeam,
    getCurrentMatch,
    match,
  } = useMatchContext();
  const { consumeItem, setUsedItem, inventoryItems } = useInventory();
  const [buttonText, setButtonText] = useState<string>("I'm here!");
  const hasDiscovered = hasTeamDiscoveredPointOfInterest(
    usersTeam,
    pointOfInterest
  );
  const { submission, challenge } = getControllingTeamForPoi(pointOfInterest);
  const controllingTeam = allTeams.find(
    (team) => team.id === submission?.teamId
  );

  const goldenTelescope = usersTeam?.teamInventoryItems.find(
    (item) =>
      item.inventoryItemId === ItemType.GoldenTelescope && item.quantity > 0
  );

  const isGoldenMonkeyActive = match?.inventoryItemEffects.some(
    (item) =>
      item.inventoryItemId === ItemType.CipherOfTheLaughingMonkey &&
      item.teamId !== usersTeam?.id &&
      new Date(item.expiresAt) > new Date()
  );

  return (
    <div className="flex flex-col items-center gap-4">
      <h3 className="text-2xl font-bold">
        {hasDiscovered ? pointOfInterest.name : 'Uncharted Waters'}
      </h3>
      <img
        src={
          hasDiscovered
            ? pointOfInterest.imageURL
            : `https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp`
        }
        alt={pointOfInterest.name}
      />
        <StatusIndicator
          tier={challenge?.tier}
          teamName={controllingTeam?.name}
          yourTeamName={usersTeam?.name ?? ''}
        />
        {hasDiscovered && (
        <TabNav
          tabs={[
            'Info',
            ...pointOfInterest.pointOfInterestChallenges
              .sort((a, b) => a.tier - b.tier)
              .map((challenge) => (
                <div className="flex items-center gap-2">
                  <span key={`Tier ${toRoman(challenge.tier)}`}>
                    {toRoman(challenge.tier)}
                  </span>
                  {challenge.pointOfInterestChallengeSubmissions?.some(
                    (submission) => submission.isCorrect
                  ) ? (
                    <LockClosedIcon className="h-4 w-4" />
                  ) : (
                    <LockOpenIcon className="h-4 w-4" />
                  )}
                </div>
              )),
          ]}
        >
          <TabItem key="Info">
            <p className="text-md text-left">{pointOfInterest.description}</p>
          </TabItem>
          {pointOfInterest.pointOfInterestChallenges.map((challenge) => {
            return (
              <TabItem key={`Tier ${toRoman(challenge.tier)}`}>
                <SubmitAnswerForChallenge
                  challenge={challenge}
                  onSubmit={(immediate) => {
                    onClose(immediate);
                  }}
                />
              </TabItem>
            );
          })}
        </TabNav>
      )}
      {!hasDiscovered && (
        <p className="text-xl text-left">
          <span className="font-bold">Clue:</span>{' '}
          {isGoldenMonkeyActive
            ? scrambleAndObscureWords(pointOfInterest.clue, usersTeam?.id ?? '')
            : pointOfInterest.clue}
        </p>
      )}
      {!hasDiscovered && (
        <div className="flex gap-2 w-full">
          <Button
            onClick={() => {
              navigator.geolocation.getCurrentPosition(async (position) => {
                try {
                  console.log('position', position);
                  await unlockPointOfInterest(
                    pointOfInterest.id,
                    usersTeam?.id ?? '',
                    position.coords.latitude.toString(),
                    (position.coords.longitude * -1).toString()
                  );
                  getCurrentMatch();
                } catch (error) {
                  setButtonText('Wrong, dingus');
                  setTimeout(() => {
                    setButtonText("I'm here!");
                  }, 1000);
                }
              });
            }}
            title={buttonText}
          />
          {!!goldenTelescope && (
            <img
              src={`https://crew-points-of-interest.s3.amazonaws.com/telescope-better.png`}
              alt="Golden Telescope"
              className="rounded-lg border-black border-2 h-12 w-12"
              onClick={() => {
                consumeItem(goldenTelescope.id, {
                  pointOfInterestId: pointOfInterest.id,
                });
                setUsedItem(inventoryItems.find(item => item.id === ItemType.GoldenTelescope)!);
                onClose(true);
              }}
            />
          )}
        </div>
      )}
    </div>
  );
};
