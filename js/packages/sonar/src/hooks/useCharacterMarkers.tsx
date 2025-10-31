import React, { useEffect, useState, useCallback, useRef } from 'react';
import mapboxgl from 'mapbox-gl';
import { createRoot } from 'react-dom/client';
import { Character } from '@poltergeist/types';
import { useMap } from '@poltergeist/contexts';
import { useZoneContext } from '@poltergeist/contexts/dist/zones';
import { useAPI } from '@poltergeist/contexts';

interface CharacterMarkerProps {
  character: Character;
  zoom: number;
}

const CharacterMarker: React.FC<CharacterMarkerProps> = ({ character, zoom }) => {
  let pinSize = 16;
  
  // Scale marker size based on zoom level
  switch (Math.floor(zoom)) {
    case 0:
    case 1:
    case 2:
    case 3:
    case 4:
    case 5:
    case 6:
    case 7:
    case 8:
      pinSize = 8;
      break;
    case 9:
    case 10:
    case 11:
    case 12:
    case 13:
    case 14:
      pinSize = 12;
      break;
    default:
      pinSize = 16;
      break;
  }

  return (
    <div
      style={{
        width: `${pinSize}px`,
        height: `${pinSize}px`,
        borderRadius: '50%',
        backgroundColor: '#FFD700',
        border: '2px solid #FFA500',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        cursor: 'pointer',
      }}
      title={character.name}
    >
      {character.mapIconUrl ? (
        <img
          src={character.mapIconUrl}
          alt={character.name}
          style={{
            width: `${pinSize - 4}px`,
            height: `${pinSize - 4}px`,
            borderRadius: '50%',
            objectFit: 'cover',
          }}
        />
      ) : (
        <span style={{ color: '#000', fontSize: `${pinSize * 0.6}px` }}>👤</span>
      )}
    </div>
  );
};

export const useCharacterMarkers = () => {
  const { map, zoom } = useMap();
  const { selectedZone } = useZoneContext();
  const { apiClient } = useAPI();
  const markersRef = useRef<mapboxgl.Marker[]>([]);
  const [characters, setCharacters] = useState<Character[]>([]);
  const [previousZoom, setPreviousZoom] = useState(0);
  const [previousSelectedZoneId, setPreviousSelectedZoneId] = useState<string | null>(null);
  const [previousCharactersCount, setPreviousCharactersCount] = useState(-1);
  const [isLoading, setIsLoading] = useState(false);

  // Fetch all characters
  useEffect(() => {
    const fetchCharacters = async () => {
      setIsLoading(true);
      try {
        const response = await apiClient.get<Character[]>('/sonar/characters');
        setCharacters(response);
      } catch (error) {
        console.error('Error fetching characters:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchCharacters();
  }, [apiClient]);

  const createMarker = useCallback((character: Character) => {
    const markerDiv = document.createElement('div');

    createRoot(markerDiv).render(
      <CharacterMarker character={character} zoom={zoom} />
    );

    const marker = new mapboxgl.Marker(markerDiv)
      .setLngLat([character.movementPattern.startingLongitude, character.movementPattern.startingLatitude])
      .addTo(map.current!);

    return marker;
  }, [zoom, map]);

  const createMarkers = useCallback(() => {
    // Filter characters for the selected zone
    const zoneCharacters = selectedZone 
      ? characters.filter((character) => character.movementPattern.zoneId === selectedZone.id)
      : [];
    
    const zoneCharactersCount = zoneCharacters.length;
    
    // Don't recreate markers if nothing has changed
    if (
      Math.abs(zoom - previousZoom) < 1 && 
      previousSelectedZoneId === selectedZone?.id &&
      previousCharactersCount === zoneCharactersCount
    ) {
      return;
    }

    setPreviousZoom(zoom);
    setPreviousSelectedZoneId(selectedZone?.id ?? null);
    setPreviousCharactersCount(zoneCharactersCount);

    // Remove existing markers
    markersRef.current.forEach((marker) => marker.remove());
    markersRef.current = [];

    // If no zone selected, don't show any character markers
    if (!selectedZone || isLoading) {
      return;
    }

    // Create new markers for characters in the selected zone
    const newMarkers = zoneCharacters.map((character) => createMarker(character));
    markersRef.current = newMarkers;
  }, [selectedZone, characters, zoom, previousZoom, previousSelectedZoneId, previousCharactersCount, isLoading, createMarker]);

  useEffect(() => {
    if (map.current && map.current?.isStyleLoaded() && !isLoading) {
      createMarkers();
    } else {
      // Wait for map to be ready
      const timer = setInterval(() => {
        if (map.current && map.current?.isStyleLoaded() && !isLoading) {
          createMarkers();
          clearInterval(timer);
        }
      }, 100);

      return () => clearInterval(timer);
    }
  }, [createMarkers, map, isLoading]);

  // Cleanup markers on unmount
  useEffect(() => {
    return () => {
      markersRef.current.forEach((marker) => marker.remove());
    };
  }, []);

  return { markers: markersRef.current };
};

