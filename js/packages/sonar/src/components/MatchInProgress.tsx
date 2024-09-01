import React, { useEffect, useMemo, useRef, useState } from 'react';
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
  MapIcon,
  MapPinIcon,
} from '@heroicons/react/20/solid';
import { PointOfInterestPanel } from './PointOfInterestPanel.tsx';
import { Drawer } from './Drawer.tsx';
import { Scoreboard } from './Scoreboard.tsx';
import { Inventory } from './Inventory.tsx';
import { Log } from './Log.tsx';
import NewItemModal from './NewItemModal.tsx';
import UsedItemModal from './UsedItemModal.tsx';
import { getUniquePoiPairsWithinDistance } from '../utils/clusterPointsOfInterest.ts';

mapboxgl.accessToken =
  'pk.eyJ1IjoibWF4YmxhdXNoaWxkIiwiYSI6ImNsenE2YWY2bDFmNnQyam9jOXJ4dHFocm4ifQ.tvO7DVEK_OLUyHfwDkUifA';

const Marker = ({
  pointOfInterest,
  index,
  zoom,
  hasDiscovered,
  onClick,
}: {
  pointOfInterest: PointOfInterest;
  index: number;
  zoom: number;
  hasDiscovered: boolean;
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
  const [areMapOverlaysVisible, setAreMapOverlaysVisible] = useState(true);
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
    if (isLeaderboardVisible || isInventoryVisible || isPanelVisible) {
      setAreMapOverlaysVisible(false);
    }

    if (!isLeaderboardVisible && !isInventoryVisible && !isPanelVisible) {
      setTimeout(() => {
        setAreMapOverlaysVisible(true);
      }, 300);
    }
  }, [isLeaderboardVisible, isInventoryVisible, isPanelVisible]);

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
    console.log(match && map.current && usersTeam, map.current?.isStyleLoaded())
    if ((match && map.current && usersTeam, map.current?.isStyleLoaded())) {
      const uniquePairs = getUniquePoiPairsWithinDistance(match!);
      markers.forEach((marker) => marker.remove());
      setMarkers([]);

      uniquePairs.forEach(([prevPoint, pointOfInterest]) => {
        const hasDiscoveredPointOne = hasTeamDiscoveredPointOfInterest(
          usersTeam!,
          pointOfInterest
        );

        const hasDiscoveredPointTwo = hasTeamDiscoveredPointOfInterest(
          usersTeam!,
          prevPoint
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

        const hasDiscovered = hasTeamDiscoveredPointOfInterest(
          usersTeam!,
          pointOfInterest
        );

        createRoot(markerDiv).render(
          <Marker
            pointOfInterest={pointOfInterest}
            index={i}
            zoom={zoom}
            hasDiscovered={!!hasDiscovered}
            onClick={(e) => {
              e.stopPropagation();
              setSelectedPointOfInterest(pointOfInterest);
            }}
          />
        );

        let lat = parseFloat(pointOfInterest.lat);
        let lng = parseFloat(pointOfInterest.lng) * -1;

        if (!hasDiscovered) {
          const baseLat = parseFloat(pointOfInterest.lat);
          const baseLng = parseFloat(pointOfInterest.lng);
          const radius = 300 / 111000; // degrees per meter
          const angle = ((baseLat + baseLng) * 1000) % 360; // deterministic angle based on lat and lng
          const newLat = baseLat + radius * Math.cos((angle * Math.PI) / 180);
          const newLng = baseLng + radius * Math.sin((angle * Math.PI) / 180);
          lat = newLat;
          lng = newLng * -1;
        }

        const marker = new mapboxgl.Marker(markerDiv)
          .setLngLat([lng, lat])
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
      {areMapOverlaysVisible && (
        <div
          className="absolute top-20 right-4 z-10 bg-white rounded-lg p-2 border-2 border-black opacity-80"
          onClick={() => {
            if (navigator.geolocation) {
              navigator.geolocation.getCurrentPosition((position) => {
                const newCenter = [
                  position.coords.longitude,
                  position.coords.latitude,
                ];
                map.current?.flyTo({ center: newCenter, zoom: 15 });
              });
            }
          }}
        >
          <MapPinIcon className="w-6 h-6" />
        </div>
      )}
      {match && areMapOverlaysVisible && (
        <div className="absolute bottom-20 right-0 z-10 w-full p-2">
          <Log />
        </div>
      )}
      {match && (
        <Drawer isVisible={isPanelVisible} onClose={closePanel} peekHeight={0}>
          {selectedPointOfInterest && (
            <PointOfInterestPanel
              pointOfInterest={selectedPointOfInterest}
              allTeams={match.teams}
              onClose={(immediate) => {
                if (immediate) {
                  closePanel();
                } else {
                  setTimeout(() => {
                    closePanel();
                  }, 2000);
                }
              }}
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
            {!isLeaderboardVisible && !isInventoryVisible && (
              <Button
                onClick={() => setIsLeaderboardVisible(true)}
                title="Leaderboard"
              ></Button>
            )}
            {!isLeaderboardVisible && !isInventoryVisible && (
              <Button
                onClick={() => setIsInventoryVisible(true)}
                title="Inventory"
              ></Button>
            )}
          </div>
          {isLeaderboardVisible && <Scoreboard />}
          {isInventoryVisible && (
            <Inventory onClose={() => setIsInventoryVisible(false)} />
          )}
        </Drawer>
      )}
      {areMapOverlaysVisible && <NewItemModal />}
      {areMapOverlaysVisible && <UsedItemModal />}
    </div>
  );
};
