import React, { useEffect } from 'react';
import { generateColorFromTeamName } from '../../utils/generateColor.ts';
import { useMatchContext } from '../../contexts/MatchContext.tsx';
import { PointOfInterestEffectingItems } from '@poltergeist/types';
import { useInventory } from '@poltergeist/contexts';

export type StatusCircleProps = {
  color: string;
};

export type StatusIndicatorProps = {
  capturingEntityName?: string | null;
  captureTier?: number | null;
};

export const StatusCircle = ({
  status,
}: {
  status: 'discovered' | 'unclaimed' | 'claimed';
}) => {
  return (
    <div
      style={{
        width: '25px',
        height: '25px',
        backgroundColor: 'grey',
        borderRadius: '50%',
      }}
    ></div>
  );
};

export const StatusIndicator = ({
  capturingEntityName,
  captureTier,
}: StatusIndicatorProps) => {
  const { match, usersTeam } = useMatchContext();
  const { inventoryItems } = useInventory();
  const effectingItems = match?.inventoryItemEffects.filter(
    (item) =>
      PointOfInterestEffectingItems.includes(item.inventoryItemId) &&
      item.teamId !== usersTeam?.id &&
      new Date(item.expiresAt) > new Date()
  );
  let color = 'grey';
  let text = 'Unclaimed';

  if (capturingEntityName) {
    color = generateColorFromTeamName(capturingEntityName);
    text = `${capturingEntityName}`;
  }

  const numCircles: number[] = [];
  if (!captureTier) {
    numCircles.push(1);
  } else {
    for (let i = 0; i < captureTier; i++) {
      numCircles.push(1);
    }
  }

  return (
    <div className="flex justify-between gap-2 w-full">
      <div className="flex gap-2 items-center">
        <p className="text-md font-bold">{text}</p>
        <div className="flex space-x-2">
          {numCircles.map((circle, index) => (
            <div
              key={index}
              style={{
                width: '25px',
                height: '25px',
                backgroundColor: color,
                borderRadius: '50%',
              }}
            ></div>
          ))}
        </div>
      </div>
      <div>
        {effectingItems?.map((item) => {
          const inventoryItem = inventoryItems.find(
            (i) => i.id === item.inventoryItemId
          );

          if (!inventoryItem) return null;
          return (
            <img
              src={inventoryItem.imageUrl}
              alt={inventoryItem.name}
              className="rounded-lg w-8 h-8 border-black border-2"
            />
          );
        })}
      </div>
    </div>
  );
};
