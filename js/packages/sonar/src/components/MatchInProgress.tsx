import React, { useEffect, useRef, useState } from 'react';
import mapboxgl from 'mapbox-gl';
import polyline from '@mapbox/polyline';
import { createRoot } from 'react-dom/client';
import './MatchInProgress.css';
import { useMatchContext } from '../contexts/MatchContext.tsx';
import {
  PointOfInterest,
  Team,
  getControllingTeamForPoi,
  hasTeamDiscoveredPointOfInterest,
} from '@poltergeist/types';
import { useUserProfiles } from '../contexts/UserProfileContext.tsx';
import { Button, ButtonSize } from './shared/Button.tsx';
import { TabItem, TabNav } from './shared/TabNav.tsx';
import { generateColorFromTeamName } from '../utils/generateColor.ts';
import Divider from './shared/Divider.tsx';
import { useSwipeable } from 'react-swipeable';
import { PointOfInterestChallenge } from '@poltergeist/types/dist/pointOfInterestChallenge';
import { SubmitAnswerForChallenge } from './SubmitAnswerForChallenge.tsx';
import {
  XMarkIcon,
  LockClosedIcon,
  LockOpenIcon,
} from '@heroicons/react/20/solid';
import { PointOfInterestPanel } from './PointOfInterestPanel.tsx';
import { Drawer } from './Drawer.tsx';
import { Scoreboard } from './Scoreboard.tsx';
import { Inventory } from './Inventory.tsx';

mapboxgl.accessToken =
  'REDACTED';

const Marker = ({
  pointOfInterest,
  index,
  zoom,
  usersTeam,
  onClick,
}: {
  pointOfInterest: PointOfInterest;
  index: number;
  zoom: number;
  usersTeam: Team | null;
  onClick: (e: React.MouseEvent) => void;
}) => {
  const hasDiscovered = hasTeamDiscoveredPointOfInterest(
    usersTeam!,
    pointOfInterest
  );
  const imageUrl = hasDiscovered
    ? pointOfInterest.imageURL
    : `https://crew-points-of-interest.s3.amazonaws.com/unclaimed-pirate-fortress-${(index + 1) % 6}.png`;

  let pinSize = 2;
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
      pinSize = 40;
      break;
    case 18:
      pinSize = 40;
      break;
    case 19:
      pinSize = 40;
      break;

    default:
      pinSize = 40;
      break;
  }
  return (
    <button onClick={onClick} className="marker">
      <img
        src={imageUrl}
        alt={hasDiscovered ? pointOfInterest.name : 'Mystery fortress'}
        className={`w-${pinSize} h-${pinSize} rounded-lg border-2 border-black`}
      />
    </button>
  );
};

export const MatchInProgress = () => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<mapboxgl.Map | null>(null);
  const { currentUser } = useUserProfiles();
  const { match } = useMatchContext();
  const [lng, setLng] = useState(-70.9);
  const [lat, setLat] = useState(42.35);
  const [zoom, setZoom] = useState(16);
  const [selectedPointOfInterest, setSelectedPointOfInterest] =
    useState<PointOfInterest | null>(null);
  const [isPanelVisible, setIsPanelVisible] = useState(false);
  const [isLeaderboardVisible, setIsLeaderboardVisible] = useState(false);
  const [isInventoryVisible, setIsInventoryVisible] = useState(false);
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null);
  const [markers, setMarkers] = useState<mapboxgl.Marker[]>([]);
  const usersTeam = match?.teams.find((team) =>
    team.users.some((user) => user.id === currentUser?.id)
  );
  const otherTeams = match?.teams.filter((team) => team.id !== usersTeam?.id);

  useEffect(() => {
    if (selectedPointOfInterest) {
      setIsPanelVisible(true);
    } else {
      setIsPanelVisible(false);
    }
  }, [selectedPointOfInterest]);

  const closePanel = () => {
    if (isPanelVisible) {
      setIsPanelVisible(false);
    }
  };

  useEffect(() => {
    map.current = new mapboxgl.Map({
      container: mapContainer.current!,
      style: 'mapbox://styles/maxblaushild/clzq7o8pr00ce01qgey4y0g31',
      center: [lng, lat],
      zoom: zoom,
    });

    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition((position) => {
        setLng(position.coords.longitude);
        setLat(position.coords.latitude);
        map.current?.setCenter([
          position.coords.longitude,
          position.coords.latitude,
        ]);
      });
    } else {
      console.error('Geolocation is not supported by this browser.');
    }

    map.current.on('move', () => {
      setLng(map.current?.getCenter().lng ?? 0);
      setLat(map.current?.getCenter().lat ?? 0);
      setZoom(map.current?.getZoom() ?? 0);
    });

    map.current?.on('zoom', () => {
      setZoom(map.current?.getZoom() ?? 0);
    });

    return () => map.current?.remove();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const handleMapClick = () => {
    setIsPanelVisible(false);
    setTimeout(() => {
      setSelectedPointOfInterest(null);
    }, 300);
  };

  const handleDrawerClick = (event: React.MouseEvent) => {
    event.stopPropagation();
  };

  useEffect(() => {
    if ((match && map.current && usersTeam, map.current?.isStyleLoaded())) {
      markers.forEach((marker) => marker.remove());
      setMarkers([]);

      const poiPairs = match!.pointsOfInterest.flatMap((poi, index, array) =>
        array.slice(index + 1).map((otherPoi) => [poi, otherPoi])
      );

      poiPairs.forEach(([prevPoint, pointOfInterest]) => {
        if (!map.current?.getSource(`${prevPoint.id}-${pointOfInterest.id}`)) {
          map.current?.addSource(`${prevPoint.id}-${pointOfInterest.id}`, {
            type: 'geojson',
            data: {
              type: 'Feature',
              properties: {},
              geometry: {
                type: 'LineString',
                coordinates: [
                  [parseFloat(prevPoint.lng) * -1, parseFloat(prevPoint.lat)],
                  [
                    parseFloat(pointOfInterest.lng) * -1,
                    parseFloat(pointOfInterest.lat),
                  ],
                ],
              },
            },
          });

          const pointOneControllingInterest =
            getControllingTeamForPoi(prevPoint);
          const pointTwoControllingInterest =
            getControllingTeamForPoi(pointOfInterest);

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

      match!.pointsOfInterest.forEach((pointOfInterest, i) => {
        const markerDiv = document.createElement('div');

        createRoot(markerDiv).render(
          <Marker
            pointOfInterest={pointOfInterest}
            index={i}
            zoom={zoom}
            usersTeam={usersTeam ?? null}
            onClick={(e) => {
              e.stopPropagation();
              setSelectedPointOfInterest(pointOfInterest);
            }}
          />
        );

        const marker = new mapboxgl.Marker(markerDiv)
          .setLngLat([
            parseFloat(pointOfInterest.lng) * -1,
            parseFloat(pointOfInterest.lat),
          ])
          .addTo(map.current!);

        setMarkers((prevMarkers) => [...prevMarkers, marker]);
      });
    }
  }, [match, map, zoom, usersTeam]);

  return (
    <div className="">
      <div
        ref={mapContainer}
        onClick={handleMapClick}
        style={{
          top: -70,
          left: 0,
          width: '100vw',
          height: '100vh',
          zIndex: 1,
        }}
      />
      {match && (
        <Drawer isVisible={isPanelVisible} onClose={closePanel} peekHeight={0}>
          {selectedPointOfInterest && (
            <PointOfInterestPanel
              pointOfInterest={selectedPointOfInterest}
              allTeams={match.teams}
            />
          )}
        </Drawer>
      )}
      {match && (
        <Drawer
          isVisible={isLeaderboardVisible || isInventoryVisible}
          onClose={() => {
            setIsLeaderboardVisible(false);
            setIsInventoryVisible(false);
          }}
          peekHeight={isPanelVisible ? 0 : 80}
        >
          <div className="flex justify-between w-full gap-4">
          {!isLeaderboardVisible && !isInventoryVisible && <Button onClick={() => setIsLeaderboardVisible(true)} title="Leaderboard"></Button>}
          {!isLeaderboardVisible && !isInventoryVisible && <Button onClick={() => setIsInventoryVisible(true)} title="Inventory"></Button>}
          </div>
          {isLeaderboardVisible && <Scoreboard />}
          {isInventoryVisible && <Inventory />}
        </Drawer>
      )}
    </div>
  );
};
