import React from 'react';
import { generateColorFromTeamName } from '../utils/generateColor';
import { PointOfInterest, Team } from '@poltergeist/types';

export const PointOfInterestMarker = ({
    pointOfInterest,
    index,
    zoom,
    hasDiscovered,
    onClick,
    controllingTeam,
  }: {
    pointOfInterest: PointOfInterest;
    index: number;
    zoom: number;
    hasDiscovered: boolean;
    controllingTeam: Team | null;
    onClick: (e: React.MouseEvent) => void;
  }) => {
    const imageUrl = hasDiscovered
      ? pointOfInterest.imageURL
      : `https://crew-points-of-interest.s3.amazonaws.com/unclaimed-pirate-fortress-${(index % 6) + 1}.png`;
  
    let pinSize = 4;
    switch (Math.floor(zoom)) {
      case 0:
        pinSize = 4;
        break;
      case 1:
        pinSize = 4;
        break;
      case 2:
        pinSize = 4;
        break;
      case 3:
        pinSize = 4;
        break;
      case 4:
        pinSize = 4;
        break;
      case 5:
        pinSize = 4;
        break;
      case 6:
        pinSize = 4;
        break;
      case 7:
        pinSize = 4;
        break;
      case 8:
        pinSize = 4;
        break;
      case 9:
        pinSize = 5;
        break;
      case 10:
        pinSize = 5;
        break;
      case 11:
        pinSize = 6;
        break;
      case 12:
        pinSize = 8;
        break;
      case 13:
        pinSize = 8;
        break;
      case 14:
        pinSize = 16;
        break;
      case 15:
        pinSize = 16;
        break;
      case 16:
        pinSize = 24;
        break;
      case 17:
        pinSize = 24;
        break;
      case 18:
        pinSize = 24;
        break;
      case 19:
        pinSize = 24;
        break;
      default:
        pinSize = 24;
        break;
    }
    return (
      <button onClick={onClick} className="marker">
        <img
          src={imageUrl}
          alt={hasDiscovered ? pointOfInterest.name : 'Mystery fortress'}
          className={`w-${pinSize} h-${pinSize} rounded-lg border-2`}
          style={{
            borderColor: controllingTeam ? generateColorFromTeamName(controllingTeam.name) : 'black',
          }}
        />
      </button>
    );
  };