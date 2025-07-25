import React, { useEffect, useState, useMemo } from 'react';
import mapboxgl from 'mapbox-gl';
import { createRoot } from 'react-dom/client';
import { PointOfInterest, PointOfInterestDiscovery, hasDiscoveredPointOfInterest } from '@poltergeist/types';
import { PointOfInterestMarker } from '../components/PointOfInterestMarker.tsx';
import { useLocation, useMap, useTagContext } from '@poltergeist/contexts';
import { useQuestLogContext } from '../contexts/QuestLogContext.tsx';

interface UsePointOfInterestMarkersProps {
  pointsOfInterest: PointOfInterest[];
  discoveries: PointOfInterestDiscovery[];
  entityId: string;
  needsDiscovery?: boolean;
  trackedPointOfInterestIds: string[];
}

interface UsePointOfInterestMarkersReturn {
  markers: mapboxgl.Marker[];
  selectedPointOfInterest: PointOfInterest | null;
  setSelectedPointOfInterest: (poi: PointOfInterest | null) => void;
}

export const usePointOfInterestMarkers = ({
  pointsOfInterest,
  discoveries,
  trackedPointOfInterestIds,
  entityId,
  needsDiscovery = false,
}: UsePointOfInterestMarkersProps): UsePointOfInterestMarkersReturn => {
  const { map, zoom } = useMap();
  const [markers, setMarkers] = useState<mapboxgl.Marker[]>([]);
  const [previousZoom, setPreviousZoom] = useState(0);
  const [previousUnlockedPoiCount, setPreviousUnlockedPoiCount] = useState(-1);
  const [previousTrackedPointOfInterestIds, setPreviousTrackedPointOfInterestIds] = useState<string[]>([]);
  const [selectedPointOfInterest, setSelectedPointOfInterest] = useState<PointOfInterest | null>(null);
  const { location } = useLocation();
  const { tagGroups } = useTagContext();

  const createPoiMarker = (pointOfInterest: PointOfInterest, index: number) => {
    const markerDiv = document.createElement('div');

    const hasDiscovered = hasDiscoveredPointOfInterest(
      pointOfInterest.id,
      entityId,
      discoveries
    );

    createRoot(markerDiv).render(
      <PointOfInterestMarker
        pointOfInterest={pointOfInterest}
        index={index}
        zoom={zoom}
        tagGroups={tagGroups}
        hasDiscovered={hasDiscovered}
        borderColor={'black'}
        isTrackedQuest={trackedPointOfInterestIds.includes(pointOfInterest.id)}
        usersLocation={location}
        onClick={(e) => {
          setSelectedPointOfInterest(pointOfInterest);
        }}
      />
    );

    let lat = parseFloat(pointOfInterest.lat);
    let lng = parseFloat(pointOfInterest.lng);

    const marker = new mapboxgl.Marker(markerDiv)
      .setLngLat([lng, lat])
      .addTo(map.current!);

    return marker;
  };

  const createMarkers = () => {
    const unlockedPoiCount = pointsOfInterest.filter((poi) => 
      hasDiscoveredPointOfInterest(poi.id, entityId, discoveries))?.length;

    if (Math.abs(zoom - previousZoom) < 1 && unlockedPoiCount === previousUnlockedPoiCount && previousTrackedPointOfInterestIds.length === trackedPointOfInterestIds.length) return;
      
    setPreviousUnlockedPoiCount(unlockedPoiCount!);
    setPreviousZoom(zoom);
    setPreviousTrackedPointOfInterestIds(trackedPointOfInterestIds);
    markers.forEach((marker) => marker.remove());
    setMarkers([]);
    const newMarkers = pointsOfInterest.map((poi, i) => createPoiMarker(poi, i));
    setMarkers(newMarkers);
  };

  useEffect(() => {
    if (map.current && entityId && map.current?.isStyleLoaded()) {
      createMarkers();
    } else {
      // Only create timer for first load when map isn't ready
      const timer = setInterval(() => {
        if (map.current && entityId && map.current?.isStyleLoaded()) {
          createMarkers();
          clearInterval(timer);
        }
      }, 100);

      return () => clearInterval(timer);
    }
  }, [pointsOfInterest, map, zoom, discoveries, entityId, map?.current, trackedPointOfInterestIds.length]);

  return { markers, selectedPointOfInterest, setSelectedPointOfInterest };
};
