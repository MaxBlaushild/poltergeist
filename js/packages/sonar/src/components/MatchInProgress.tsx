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
import { useLocation } from '@poltergeist/contexts';
import { ImageBadge } from './shared/ImageBadge.tsx';

mapboxgl.accessToken =
  'pk.eyJ1IjoibWF4YmxhdXNoaWxkIiwiYSI6ImNsenE2YWY2bDFmNnQyam9jOXJ4dHFocm4ifQ.tvO7DVEK_OLUyHfwDkUifA';

const Marker = ({
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

export const MatchInProgress = () => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<mapboxgl.Map | null>(null);
  const { currentUser } = useUserProfiles();
  const { location } = useLocation();
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
  const [previousZoom, setPreviousZoom] = useState(0);
  const [isMapInitialized, setIsMapInitialized] = useState(false);
  
  const [userLocator, setUserLocator] = useState<mapboxgl.Marker | null>(null);
  const [previousUnlockedPoiCount, setPreviousUnlockedPoiCount] = useState(-1);
  const usersTeam = match?.teams.find((team) =>
    team.users.some((user) => user.id === currentUser?.id)
  );
  const otherTeams = match?.teams.filter((team) => team.id !== usersTeam?.id);

  useEffect(() => {
    if (!map?.current || !map.current?.isStyleLoaded()) return;

    let locator = userLocator;
    if (!locator) {
      const locatorDiv = document.createElement('div');
      createRoot(locatorDiv).render(
        <ImageBadge
          imageUrl={currentUser?.profilePictureUrl ?? '/blank-avatar.webp'} 
          onClick={() => {}}
          hasBorder={true}
        />
      );
      
      locator = new mapboxgl.Marker(locatorDiv);
      setUserLocator(locator);
    }

    // If we have location coordinates, use them
    if (location?.longitude && location?.latitude) {
      locator
        .setLngLat([location.longitude, location.latitude])
        .addTo(map.current);
    }
  }, [location, map, userLocator, currentUser]);

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
      setSelectedPointOfInterest(null);
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

    return () => map.current?.remove();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (isMapInitialized) return;
    if (location.longitude && location.latitude && map?.current) {
      setIsMapInitialized(true);
      map.current?.setCenter([location.longitude, location.latitude]);

      map.current.on('move', () => {
        setLng(map.current?.getCenter().lng ?? 0);
        setLat(map.current?.getCenter().lat ?? 0);
        setZoom(map.current?.getZoom() ?? 0);
      });
  
      map.current?.on('zoom', () => {
        setZoom(map.current?.getZoom() ?? 0);
      });
    }
  }, [location.longitude, location.latitude, map?.current]);

  const handleMapClick = () => {
    setIsPanelVisible(false);
    setTimeout(() => {
      setSelectedPointOfInterest(null);
    }, 300);
  };

  const handleDrawerClick = (event: React.MouseEvent) => {
    event.stopPropagation();
  };

  const memoizedAlternativeCoordinates = useMemo(() => {
    return match?.pointsOfInterest.reduce((acc, poi) => {
        const baseLat = parseFloat(poi.lat);
        const baseLng = parseFloat(poi.lng);
        const radius = 150 / 111000; // degrees per meter
        const angle = ((baseLat + baseLng) * 1000) % 360; // deterministic angle based on lat and lng
        const newLat = baseLat + radius * Math.cos((angle * Math.PI) / 180);
        const newLng = baseLng + radius * Math.sin((angle * Math.PI) / 180);
        acc[poi.id] = { newLat, newLng: newLng };
        return acc;
    }, {});
  }, [match]);

  useEffect(() => {
    if ((match && map.current && usersTeam, map.current?.isStyleLoaded())) {
      const unlockedPoiCount = match?.pointsOfInterest.filter((poi) => hasTeamDiscoveredPointOfInterest(usersTeam!, poi))?.length;
      if (Math.abs(zoom - previousZoom) < 1 && unlockedPoiCount === previousUnlockedPoiCount) return;
      setPreviousUnlockedPoiCount(unlockedPoiCount!);
      setPreviousZoom(zoom);
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
                  [parseFloat(prevPoint.lng), parseFloat(prevPoint.lat)],
                  [
                    parseFloat(pointOfInterest.lng),
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

        const controllingInterest = getControllingTeamForPoi(pointOfInterest);
        const controllingTeam = match?.teams.find(
          (team) =>
            team.id === controllingInterest?.submission?.teamId
        );

        createRoot(markerDiv).render(
          <Marker
            pointOfInterest={pointOfInterest}
            index={i}
            zoom={zoom}
            hasDiscovered={!!hasDiscovered}
            controllingTeam={controllingTeam}
            onClick={(e) => {
              e.stopPropagation();
              setSelectedPointOfInterest(pointOfInterest);
            }}
          />
        );



        let lat = parseFloat(pointOfInterest.lat);
        let lng = parseFloat(pointOfInterest.lng);

        if (!hasDiscovered) {
          const coords = memoizedAlternativeCoordinates?.[pointOfInterest.id];
          if (coords) {
            lat = coords.newLat;
            lng = coords.newLng;
          }
        }

        const marker = new mapboxgl.Marker(markerDiv)
          .setLngLat([lng, lat])
          .addTo(map.current!);

        setMarkers((prevMarkers) => [...prevMarkers, marker]);
      });
    }
  }, [match, map, zoom, usersTeam, memoizedAlternativeCoordinates]);

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
            console.log('clicked');
            if (location.longitude && location.latitude) {
              const newCenter = [
                location.longitude,
                location.latitude,
              ];
              map.current?.flyTo({ center: newCenter, zoom: 15 });
              return;
            }
            if (navigator.geolocation) {
              navigator.geolocation.getCurrentPosition((position) => {
                const newCenter = [
                  position.coords.longitude,
                  position.coords.latitude,
                ];
                map.current?.flyTo({ center: newCenter, zoom: 15 });
              }, (error) => {
                console.log('error', error);
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
