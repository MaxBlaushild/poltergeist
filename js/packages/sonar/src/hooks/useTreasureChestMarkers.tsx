import { useEffect, useState } from 'react';
import mapboxgl from 'mapbox-gl';
import { createRoot } from 'react-dom/client';
import { TreasureChestMarker } from '../components/TreasureChestMarker.tsx';
import { useMap } from '@poltergeist/contexts';
import { TreasureChest } from '@poltergeist/types';

interface UseTreasureChestMarkersProps {
  treasureChests: TreasureChest[];
}

interface UseTreasureChestMarkersReturn {
  markers: mapboxgl.Marker[];
  selectedTreasureChest: TreasureChest | null;
  setSelectedTreasureChest: (chest: TreasureChest | null) => void;
}

export const useTreasureChestMarkers = ({
  treasureChests,
}: UseTreasureChestMarkersProps): UseTreasureChestMarkersReturn => {
  const { map, zoom } = useMap();
  const [markers, setMarkers] = useState<mapboxgl.Marker[]>([]);
  const [previousZoom, setPreviousZoom] = useState(0);
  const [selectedTreasureChest, setSelectedTreasureChest] = useState<TreasureChest | null>(null);

  const createChestMarker = (treasureChest: TreasureChest, index: number) => {
    const markerDiv = document.createElement('div');

    createRoot(markerDiv).render(
      <TreasureChestMarker
        treasureChest={treasureChest}
        index={index}
        zoom={zoom}
        onClick={(e) => {
          setSelectedTreasureChest(treasureChest);
        }}
      />
    );

    const marker = new mapboxgl.Marker(markerDiv)
      .setLngLat([treasureChest.longitude, treasureChest.latitude])
      .addTo(map.current!);

    return marker;
  };

  const createMarkers = () => {
    if (Math.abs(zoom - previousZoom) < 1 && treasureChests.length === markers.length) return;
      
    setPreviousZoom(zoom);
    markers.forEach((marker) => marker.remove());
    setMarkers([]);
    const newMarkers = treasureChests.map((chest, i) => createChestMarker(chest, i));
    setMarkers(newMarkers);
  };

  useEffect(() => {
    if (map.current && map.current?.isStyleLoaded()) {
      createMarkers();
    } else {
      // Only create timer for first load when map isn't ready
      const timer = setInterval(() => {
        if (map.current && map.current?.isStyleLoaded()) {
          createMarkers();
          clearInterval(timer);
        }
      }, 100);

      return () => clearInterval(timer);
    }
  }, [treasureChests, map, zoom, map?.current]);

  return { markers, selectedTreasureChest, setSelectedTreasureChest };
};

