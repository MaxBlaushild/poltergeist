import React from 'react';
import { generateColorFromTeamName } from '../../utils/generateColor.ts';

export type StatusCircleProps = {
  color: string;
};

export type StatusIndicatorProps = {
  tier?: number | null;
  teamName?: string | null;
  yourTeamName: string;
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
  tier,
  teamName,
  yourTeamName,
}: StatusIndicatorProps) => {
  let color = 'grey';
  let text = 'Unclaimed';

  if (tier && teamName) {
    color = generateColorFromTeamName(teamName);

    if (teamName === yourTeamName) {
      text = 'Owned by you';
    } else {
      text = `${teamName}`;
    }
  }

  const numCircles: number[] = [];
  if (!tier) {
    numCircles.push(1);
  } else {
    for (let i = 0; i < tier; i++) {
      numCircles.push(1);
    }
  }

  return (
    <div className="flex items-start gap-2 w-full">
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
  );
};
