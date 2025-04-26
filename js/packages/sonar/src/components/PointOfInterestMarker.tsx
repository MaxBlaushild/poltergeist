import React from 'react';
import { generateColorFromTeamName } from '../utils/generateColor';
import { PointOfInterest, TagGroup, Team } from '@poltergeist/types';
import { useLocation } from '@poltergeist/contexts';
import { tagsToFilter } from '../utils/tagFilter.ts';

export const PointOfInterestMarker = ({
  pointOfInterest,
  index,
  zoom,
  hasDiscovered,
  onClick,
  borderColor,
  usersLocation,
  tagGroups,
}: {
  pointOfInterest: PointOfInterest;
  index: number;
  zoom: number;
  hasDiscovered: boolean;
  borderColor: string | undefined;
  usersLocation: Location | null;
  onClick: (e: React.MouseEvent) => void;
  tagGroups: TagGroup[];
}) => {
  const tagGroup = tagGroups?.reduce<{group: TagGroup | null, matchCount: number}>((bestMatch, group) => {
    const matchCount = (group.tags || []).filter(tag => 
      pointOfInterest.tags?.some(t => t.id === tag.id && !tagsToFilter.includes(tag.name))
    ).length;
    return matchCount > bestMatch.matchCount ? {group, matchCount} : bestMatch;
  }, {group: null, matchCount: 0})?.group;
  const imageUrl = hasDiscovered
  ? pointOfInterest.imageURL
    : tagGroup?.iconUrl || `https://crew-points-of-interest.s3.amazonaws.com/question-mark.webp`;

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
      pinSize = 8;
      break;
    case 15:
      pinSize = 16;
      break;
    case 16:
      pinSize = 16;
      break;
    case 17:
      pinSize = 16;
      break;
    case 18:
      pinSize = 16;
      break;
    case 19:
      pinSize = 16;
      break;
    default:
      pinSize = 16;
      break;
  }
  let opacity = 1;
  if (usersLocation?.latitude && usersLocation?.longitude) {
    const R = 6371e3; // Earth's radius in meters
    const φ1 = usersLocation.latitude * Math.PI/180;
    const φ2 = parseFloat(pointOfInterest.lat) * Math.PI/180;
    const Δφ = (parseFloat(pointOfInterest.lat) - usersLocation.latitude) * Math.PI/180;
    const Δλ = (parseFloat(pointOfInterest.lng) - usersLocation.longitude) * Math.PI/180;

    const a = Math.sin(Δφ/2) * Math.sin(Δφ/2) +
            Math.cos(φ1) * Math.cos(φ2) *
            Math.sin(Δλ/2) * Math.sin(Δλ/2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a));
    const distance = R * c;

    if (distance < 50) { // Within 50 meters
      opacity = 0.3;
    }
  }

  return (
    <button onClick={onClick} className="marker">
      <img
        src={imageUrl}
        alt={pointOfInterest.name} 
        className={`w-${pinSize} h-${pinSize} rounded-lg border-2`}
        style={{
          borderColor,
          opacity
        }}
      />
    </button>
  );
};
