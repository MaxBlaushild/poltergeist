import React, { useEffect, useState, useMemo } from 'react';
import mapboxgl from 'mapbox-gl';
import { createRoot } from 'react-dom/client';
import { PointOfInterest, PointOfInterestDiscovery, hasDiscoveredPointOfInterest } from '@poltergeist/types';
import { PointOfInterestMarker } from '../components/PointOfInterestMarker.tsx';
import { useMap } from '@poltergeist/contexts';

interface UsePointOfInterestMarkersProps {
  pointsOfInterest: PointOfInterest[];
  discoveries: PointOfInterestDiscovery[];
  entityId: string;
}

interface UsePointOfInterestMarkersReturn {
  markers: mapboxgl.Marker[];
  selectedPointOfInterest: PointOfInterest | null;
  setSelectedPointOfInterest: (poi: PointOfInterest | null) => void;
}

export const usePointOfInterestMarkers = ({
  pointsOfInterest,
  discoveries,
  entityId,
}: UsePointOfInterestMarkersProps): UsePointOfInterestMarkersReturn => {
  const { map, zoom } = useMap();
  const [markers, setMarkers] = useState<mapboxgl.Marker[]>([]);
  const [previousZoom, setPreviousZoom] = useState(0);
  const [previousUnlockedPoiCount, setPreviousUnlockedPoiCount] = useState(-1);
  const [selectedPointOfInterest, setSelectedPointOfInterest] = useState<PointOfInterest | null>(null);

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
    }, {} as Record<string, { newLat: number; newLng: number }>);
  }, [pointsOfInterest.length]);

  useEffect(() => {
    if ((map.current && entityId && map.current?.isStyleLoaded())) {
      const unlockedPoiCount = pointsOfInterest.filter((poi) => 
        hasDiscoveredPointOfInterest(poi.id, entityId, discoveries))?.length;

      if (Math.abs(zoom - previousZoom) < 1 && unlockedPoiCount === previousUnlockedPoiCount) return;
        
      setPreviousUnlockedPoiCount(unlockedPoiCount!);
      setPreviousZoom(zoom);
      markers.forEach((marker) => marker.remove());
      setMarkers([]);

      pointsOfInterest.forEach((pointOfInterest, i) => {
        const markerDiv = document.createElement('div');

        const hasDiscovered = hasDiscoveredPointOfInterest(
          pointOfInterest.id,
          entityId,
          discoveries
        );

        createRoot(markerDiv).render(
          <PointOfInterestMarker
            pointOfInterest={pointOfInterest}
            index={i}
            zoom={zoom}
            hasDiscovered={!!hasDiscovered}
            borderColor={'black'}
            onClick={(e) => {
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
  }, [pointsOfInterest, map, zoom, discoveries, entityId, memoizedAlternativeCoordinates]);

  return { markers, selectedPointOfInterest, setSelectedPointOfInterest };
};
