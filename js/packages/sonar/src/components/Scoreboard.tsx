import React, { useEffect, useState } from 'react';
import './Scoreboard.css';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import { ItemType, PointOfInterest, getHighestFirstCompletedChallenge } from '@poltergeist/types';
import { generateColorFromTeamName } from '../utils/generateColor.ts';
import Divider from './shared/Divider.tsx';
import { useNavigate } from 'react-router-dom';
import PersonListItem from './shared/PersonListItem.tsx';
import { ArrowLeftCircleIcon, ArrowLeftIcon } from '@heroicons/react/20/solid';
import { PointOfInterestChallenge } from '@poltergeist/types/dist/pointOfInterestChallenge';
import { getUniquePoiPairsWithinDistance } from '../utils/clusterPointsOfInterest.ts';

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

const ProfilePictureModal = ({ onExit, url }: { onExit: () => void, url: string }) => {

  return (
    <div 
      className="fixed inset-0 bg-black z-[100] flex flex-col items-center justify-center"
      onClick={() => {
        onExit();
      }}
    >
      <img
        src={url}
        alt="Profile Picture"
        className="w-screen h-screen object-contain"
      />
    </div>
  );
};

export const Scoreboard = () => {
  const { match } = useMatchContext();
  const navigate = useNavigate();
  const [selectedProfilePicture, setSelectedProfilePicture] = useState<string | null>(null);
  const [selectedTeamID, setSelectedTeamID] = useState<string | null>(null);
  const selectedTeam = match?.teams.find((team) => team.id === selectedTeamID);

  const uniquePoiPairs = match ? getUniquePoiPairsWithinDistance(match?.pointsOfInterest ?? []) : [];

  const scoreboard: { [key: string]: number } = {};
  const capturedPoints: { [key: string]: { poi: PointOfInterest, challenge: PointOfInterestChallenge }[] } = {};

  match?.pointsOfInterest.forEach((poi) => {
    const pointOneControllingInterest = getHighestFirstCompletedChallenge(poi);
    if (pointOneControllingInterest?.submission?.teamId) {
      capturedPoints[pointOneControllingInterest.submission.teamId] = [
        ...capturedPoints[pointOneControllingInterest.submission.teamId] ?? [],
        { poi, challenge: pointOneControllingInterest.challenge! },
      ];
    }
  });

  uniquePoiPairs?.forEach(([prevPoint, pointOfInterest]) => {
    const pointOneControllingInterest = getHighestFirstCompletedChallenge(prevPoint);
    const pointTwoControllingInterest =
    getHighestFirstCompletedChallenge(pointOfInterest);

    if (
      pointOneControllingInterest?.submission?.teamId ===
      pointTwoControllingInterest?.submission?.teamId
    ) {
      const teamId = pointOneControllingInterest?.submission?.teamId ?? '';
      scoreboard[teamId] = (scoreboard[teamId] || 0) + 1;
    }
  });

  match?.teams.forEach((team) => {
    team.ownedInventoryItems.forEach((item) => {
      if (item.inventoryItemId === ItemType.GoldCoin) {
        scoreboard[team.id] = (scoreboard[team.id] || 0) + item.quantity;
      }
      if (item.inventoryItemId === ItemType.Damage && !team.ownedInventoryItems.some((i) => i.inventoryItemId === ItemType.Entseed || i.inventoryItemId === ItemType.Witchflame)) {
        scoreboard[team.id] = (scoreboard[team.id] || 0) - (item.quantity * 2);
      }

      if (item.inventoryItemId === ItemType.Entseed) {
        scoreboard[team.id] = (scoreboard[team.id] || 0) + (item.quantity * 3);
      }
    });
  });

  return (
    <div className="flex flex-col items-center w-full gap-4">
      {!selectedTeamID && (
        <div className="flex flex-col items-center w-full gap-4">
          <h1 className="text-2xl font-bold">Leaderboard</h1>
          <table className="w-full mt-4">
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
              onClick={(u) => setSelectedProfilePicture(u.profilePictureUrl)}
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
                profilePictureUrl: poi.imageURL,
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
      {selectedProfilePicture && (
        <ProfilePictureModal onExit={() => setSelectedProfilePicture(null)} url={selectedProfilePicture} />
      )}
    </div>
  );
};
