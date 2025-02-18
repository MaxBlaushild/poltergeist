import { useEffect, useState } from 'react';
import { generateColorFromTeamName } from '../utils/generateColor.ts';
import {
  PointOfInterest,
  PointOfInterestDiscovery,
  Team,
  getHighestFirstCompletedChallenge,
  hasDiscoveredPointOfInterest,
} from '@poltergeist/types';
import { getUniquePoiPairsWithinDistance } from '../utils/clusterPointsOfInterest.ts';
import { useMap } from '@poltergeist/contexts';
import { useMatchContext } from '../contexts/MatchContext.tsx';

interface UseMapLinesProps {
  pointsOfInterest: PointOfInterest[];
  usersTeam: Team;
}

export const useMapLines = ({
  pointsOfInterest,
  usersTeam,
}) => {
  const { match } = useMatchContext();
  const { zoom, map } = useMap();
  const [previousZoom, setPreviousZoom] = useState(zoom);
  const [lines, setLines] = useState<any[]>([]);
  const [previousUnlockedPoiCount, setPreviousUnlockedPoiCount] = useState(0);

  useEffect(() => {
    if ((pointsOfInterest?.length > 0 && map.current && usersTeam, map.current?.isStyleLoaded())) {
      const unlockedPoiCount = pointsOfInterest.filter((poi) => 
        hasDiscoveredPointOfInterest(poi.id, usersTeam?.id ?? '', usersTeam?.pointOfInterestDiscoveries ?? [])
      )?.length;

      if (Math.abs(zoom - previousZoom) < 1 && unlockedPoiCount === previousUnlockedPoiCount) return;
      
      setPreviousUnlockedPoiCount(unlockedPoiCount!);
      setPreviousZoom(zoom);

      const uniquePairs = getUniquePoiPairsWithinDistance(pointsOfInterest);

      uniquePairs.forEach(([prevPoint, pointOfInterest]) => {
        const hasDiscoveredPointOne = hasDiscoveredPointOfInterest(
          pointOfInterest.id,
          usersTeam?.id ?? '',
          usersTeam?.pointOfInterestDiscoveries ?? []
        );

        const hasDiscoveredPointTwo = hasDiscoveredPointOfInterest(
          prevPoint.id,
          usersTeam?.id ?? '',
          usersTeam?.pointOfInterestDiscoveries ?? []
        );

        if (!hasDiscoveredPointOne || !hasDiscoveredPointTwo) {
          return;
        }

        if (!map.current?.getSource(`${prevPoint.id}-${pointOfInterest.id}`)) {
          map.current?.addSource(`${prevPoint.id}-${pointOfInterest.id}`, {
            type: 'geojson',
            data: {
              type: 'Feature',
              properties: {},
              geometry: {
                type: 'LineString',
                coordinates: [
                  [parseFloat(prevPoint.lng), parseFloat(prevPoint.lat)],
                  [
                    parseFloat(pointOfInterest.lng),
                    parseFloat(pointOfInterest.lat),
                  ],
                ],
              },
            },
          });

          const pointOneControllingInterest = getHighestFirstCompletedChallenge(prevPoint);
          const pointTwoControllingInterest = getHighestFirstCompletedChallenge(pointOfInterest);

          let color = 'grey';
          let opacity = 0.5;
          if (
            pointOneControllingInterest?.submission?.teamId ===
              pointTwoControllingInterest?.submission?.teamId &&
            pointOneControllingInterest?.submission?.teamId
          ) {
            const controllerTeam = match?.teams.find(
              (team) =>
                team.id === pointOneControllingInterest?.submission?.teamId
            );
            color = generateColorFromTeamName(controllerTeam?.name ?? '');
            opacity = 1;
          }

          map.current?.addLayer({
            id: `${prevPoint.id}-${pointOfInterest.id}`,
            type: 'line',
            source: `${prevPoint.id}-${pointOfInterest.id}`,
            layout: {},
            paint: {
              'line-color': color,
              'line-width': 5,
              'line-opacity': opacity,
            },
          });
        }
      });
    }
  }, [pointsOfInterest, map, zoom, usersTeam, previousZoom, setPreviousZoom, previousUnlockedPoiCount, setPreviousUnlockedPoiCount]);
};
