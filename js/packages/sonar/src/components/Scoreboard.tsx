import React, { useEffect, useState } from 'react';
import './Scoreboard.css';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { PointOfInterest, getControllingTeamForPoi } from '@poltergeist/types';
import { generateColorFromTeamName } from '../utils/generateColor.ts';
import Divider from './shared/Divider.tsx';
import { useNavigate } from 'react-router-dom';
import PersonListItem from './shared/PersonListItem.tsx';
import { ArrowLeftCircleIcon, ArrowLeftIcon } from '@heroicons/react/20/solid';
import { PointOfInterestChallenge } from '@poltergeist/types/dist/pointOfInterestChallenge';

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

export const Scoreboard = () => {
  const { match } = useMatchContext();
  const navigate = useNavigate();
  const [selectedTeamID, setSelectedTeamID] = useState<string | null>(null);
  const selectedTeam = match?.teams.find((team) => team.id === selectedTeamID);

  const poiPairs = match?.pointsOfInterest.flatMap((poi, index, array) =>
    array.slice(index + 1).map((otherPoi) => [poi, otherPoi])
  );

  const scoreboard: { [key: string]: number } = {};
  const capturedPoints: { [key: string]: { poi: PointOfInterest, challenge: PointOfInterestChallenge }[] } = {};

  match?.pointsOfInterest.forEach((poi) => {
    const pointOneControllingInterest = getControllingTeamForPoi(poi);
    if (pointOneControllingInterest?.submission?.teamId) {
      capturedPoints[pointOneControllingInterest.submission.teamId] = [
        ...capturedPoints[pointOneControllingInterest.submission.teamId] ?? [],
        { poi, challenge: pointOneControllingInterest.challenge! },
      ];
    }
  });

  poiPairs?.forEach(([prevPoint, pointOfInterest]) => {
    const pointOneControllingInterest = getControllingTeamForPoi(prevPoint);
    const pointTwoControllingInterest =
      getControllingTeamForPoi(pointOfInterest);

    if (
      pointOneControllingInterest?.submission?.teamId ===
      pointTwoControllingInterest?.submission?.teamId
    ) {
      const teamId = pointOneControllingInterest?.submission?.teamId ?? '';
      scoreboard[teamId] = (scoreboard[teamId] || 0) + 1;
    }
  });

  return (
    <div className="flex flex-col items-center w-full gap-4">
      {!selectedTeamID && (
        <div className="flex flex-col items-center w-full gap-4">
          <h1 className="text-2xl font-bold">Leaderboard</h1>
          <Divider />
          <table className="w-full">
            <thead>
              <tr>
                <th className="text-left">Team</th>
                <th className="text-center">Color</th>
                <th className="text-center">Score</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <br />
              </tr>
              {match?.teams
                .sort((a, b) => scoreboard[b.id] - scoreboard[a.id])
                .map((team) => (
                  <tr key={team.id} onClick={() => setSelectedTeamID(team.id)}>
                    <td className="text-left text-lg font-bold">
                      {team.name}
                    </td>
                    <td className="text-center">
                      <div
                        style={{
                          width: '32px',
                          height: '32px',
                          backgroundColor: generateColorFromTeamName(team.name),
                          borderRadius: '50%',
                          margin: 'auto',
                        }}
                      />
                    </td>
                    <td className="text-center text-xl font-bold">
                      {scoreboard[team.id] ?? 0}
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
      )}
      {selectedTeamID && (
        <div className="flex flex-col items-center w-full gap-4">
          <div className="flex flex-row items-center w-full justify-center gap-4">
            <ArrowLeftIcon className="w-8 h-8 absolute left-4" onClick={() => setSelectedTeamID(null)} />
            <h2 className="text-2xl font-bold text-center">{selectedTeam?.name}</h2>
          </div>
          <Divider color={generateColorFromTeamName(selectedTeam?.name ?? '')} />
          <div className="w-full flex flex-col items-left mt-4">
            <h3 className="text-lg font-bold text-left">Team members</h3>
            {selectedTeam?.users.map((user) => (
              <PersonListItem
              key={user.id}
              user={user}
              onClick={(u) => {}}
              actionArea={() => <div></div>}
              />
            ))}
          </div>
          <div className="w-full flex flex-col items-left mt-4">
            <h3 className="text-lg font-bold text-left">Captured points</h3>
            {capturedPoints[selectedTeam?.id ?? '']?.map(({ poi, challenge }) => (
              <PersonListItem
              key={poi.id}
              user={{
                id: poi.id,
                name: poi.name,
                profile: {
                  id: poi.id,
                  createdAt: "",
                  updatedAt: "",
                  viewerId: "",
                  vieweeId: "",
                  profilePictureUrl: poi.imageURL,
                },
                phoneNumber: "",
              }}
              onClick={(u) => {}}
              actionArea={() => <div>
                <p className="text-lg font-bold">Tier {toRoman(challenge.tier)}</p>
                </div>}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
