import { PointOfInterest, Team, getControllingTeamForPoi, hasTeamDiscoveredPointOfInterest } from '@poltergeist/types';
import React, { useState } from 'react';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { PointOfInterestChallenge } from '@poltergeist/types/dist/pointOfInterestChallenge';
import { TabItem, TabNav } from './shared/TabNav.tsx';
import { SubmitAnswerForChallenge } from './SubmitAnswerForChallenge.tsx';
import { Button } from './shared/Button.tsx';
import { LockClosedIcon, LockOpenIcon } from '@heroicons/react/20/solid';
import { StatusIndicator } from './shared/StatusIndicator.tsx';

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
}: {
  pointOfInterest: PointOfInterest;
  allTeams: Team[];
}) => {
  const { unlockPointOfInterest, attemptCapturePointOfInterest, usersTeam, getCurrentMatch } =
    useMatchContext();
  const hasDiscovered = hasTeamDiscoveredPointOfInterest(
    usersTeam,
    pointOfInterest
  );
  const { submission, challenge } = getControllingTeamForPoi(pointOfInterest);
  const controllingTeam = allTeams.find(
    (team) => team.id === submission?.teamId
  );
  const [selectedChallenge, setSelectedChallenge] = useState<
    PointOfInterestChallenge | undefined
  >(undefined);

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
      {!selectedChallenge && (
        <StatusIndicator
          tier={challenge?.tier}
          teamName={controllingTeam?.name}
          yourTeamName={usersTeam?.name ?? ''}
        />
      )}
      {hasDiscovered && !selectedChallenge && (
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
                  onSubmit={() => setSelectedChallenge(undefined)}
                />
              </TabItem>
            );
          })}
        </TabNav>
      )}
      {hasDiscovered && selectedChallenge && (
        <SubmitAnswerForChallenge
          challenge={selectedChallenge}
          onSubmit={() => setSelectedChallenge(undefined)}
        />
      )}
      {!hasDiscovered && (
        <p className="text-xl text-left">
          <span className="font-bold">Clue:</span> {pointOfInterest.clue}
        </p>
      )}
      {!hasDiscovered && (
        <Button
          onClick={() => {
            navigator.geolocation.getCurrentPosition((position) => {
              unlockPointOfInterest(
                pointOfInterest.id,
                usersTeam?.id ?? '',
                pointOfInterest.lat,
                pointOfInterest.lng
              );
              getCurrentMatch();
            });
          }}
          title="I'm here!"
        />
      )}
    </div>
  );
};
