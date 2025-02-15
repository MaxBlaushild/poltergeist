import { Match, PointOfInterest } from '@poltergeist/types';
import { calculateDistance } from './calculateDistance.ts';
const distanceThreshold = 300;

const distances = new Map();

export const getMemoizedDistance = (
  poi1: PointOfInterest,
  poi2: PointOfInterest
) => {
  const key = `${poi1.id}-${poi2.id}`;
  const reverseKey = `${poi2.id}-${poi1.id}`;
  if (distances.has(key)) {
    return distances.get(key);
  } else if (distances.has(reverseKey)) {
    return distances.get(reverseKey);
  } else {
    const distance = calculateDistance(poi1, poi2);
    distances.set(key, distance);
    return distance;
  }
};

const pairMemo = new Map();

export const getUniquePoiPairsWithinDistance = (pointsOfInterest: PointOfInterest[]) => {
  if (pairMemo.has(pointsOfInterest.length)) {
    return pairMemo.get(pointsOfInterest.length);
  }

  const pairs: [PointOfInterest, PointOfInterest][] = [];
  const uniquePairsMap = new Map();

  pointsOfInterest.forEach((poi1, index, array) => {
    array.slice(index + 1).forEach((poi2) => {
      const distance = getMemoizedDistance(poi1, poi2);
      if (distance < distanceThreshold) {
        const pairKey = `${poi1.id}-${poi2.id}`;
        const reversePairKey = `${poi2.id}-${poi1.id}`;
        if (
          !uniquePairsMap.has(pairKey) &&
          !uniquePairsMap.has(reversePairKey)
        ) {
          pairs.push([poi1, poi2]);
          uniquePairsMap.set(pairKey, true);
        }
      }
    });
  });

  pairMemo.set(pointsOfInterest.length, pairs);
  return pairs;
};
