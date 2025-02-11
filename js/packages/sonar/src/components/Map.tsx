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
import { PointOfInterestChallenge } from '@poltergeist/types';
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
import { PointOfInterestMarker } from './PointOfInterestMarker.tsx';
import { PointOfInterestChallengeSubmission } from '@poltergeist/types/dist/pointOfInterestChallengeSubmission';

mapboxgl.accessToken =
  'REDACTED';

interface MapProps {
  pointsOfInterest: PointOfInterest[];
  submissions: PointOfInterestChallengeSubmission[];
}

export const Map = ({ pointsOfInterest }: MapProps) => {
  const mapContainer = useRef<HTMLDivElement>(null);
  const map = useRef<mapboxgl.Map | null>(null);
  const { currentUser } = useUserProfiles();
  const { location } = useLocation();
  const [lng, setLng] = useState(-70.9);
  const [lat, setLat] = useState(42.35);
  const [zoom, setZoom] = useState(16);
  const [markers, setMarkers] = useState<mapboxgl.Marker[]>([]);
  const [previousZoom, setPreviousZoom] = useState(0);
  const [isMapInitialized, setIsMapInitialized] = useState(false);
  const [userLocator, setUserLocator] = useState<mapboxgl.Marker | null>(null);
  const [previousUnlockedPoiCount, setPreviousUnlockedPoiCount] = useState(-1);

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
    if (location?.longitude && location?.latitude && map?.current) {
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
  }, [location?.longitude, location?.latitude, map?.current]);

  const memoizedAlternativeCoordinates = useMemo(() => {
    return pointsOfInterest.reduce((acc, poi) => {
        const baseLat = parseFloat(poi.lat);
        const baseLng = parseFloat(poi.lng);
        const radius = 150 / 111000; // degrees per meter
        const angle = ((baseLat + baseLng) * 1000) % 360; // deterministic angle based on lat and lng
        const newLat = baseLat + radius * Math.cos((angle * Math.PI) / 180);
        const newLng = baseLng + radius * Math.sin((angle * Math.PI) / 180);
        acc[poi.id] = { newLat, newLng: newLng };
        return acc;
    }, {});
  }, [pointsOfInterest.length]);

  useEffect(() => {
    if ((match && map.current && usersTeam, map.current?.isStyleLoaded())) {
      const unlockedPoiCount = pointsOfInterest.filter((poi) => hasTeamDiscoveredPointOfInterest(usersTeam!, poi))?.length;
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
          <PointOfInterestMarker
            pointOfInterest={pointOfInterest}
            index={i}
            zoom={zoom}
            hasDiscovered={!!hasDiscovered}
            controllingTeam={controllingTeam ?? null}
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
        <div
          className="absolute top-20 right-4 z-10 bg-white rounded-lg p-2 border-2 border-black opacity-80"
          onClick={() => {
            console.log('clicked');
            if (location?.longitude && location?.latitude) {
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
    </div>
  );
};
