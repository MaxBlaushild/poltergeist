import React, { useEffect, useState, useCallback, useRef } from 'react';
import mapboxgl from 'mapbox-gl';
import { createRoot } from 'react-dom/client';
import { Character } from '@poltergeist/types';
import { useMap, useLocation } from '@poltergeist/contexts';
import { useZoneContext } from '@poltergeist/contexts/dist/zones';
import { useAPI } from '@poltergeist/contexts';
import { calculateDistance } from '../utils/calculateDistance.ts';
import { getMarkerPixelSize } from '../utils/markerSize.ts';

interface CharacterMarkerProps {
  character: Character;
  zoom: number;
  onClick?: () => void;
  isClickable: boolean;
}

const CharacterMarker: React.FC<CharacterMarkerProps> = ({ character, zoom, onClick, isClickable }) => {
  const pixelSize = getMarkerPixelSize(zoom);

  return (
    <div
      onClick={isClickable ? onClick : undefined}
      style={{
        width: `${pixelSize}px`,
        height: `${pixelSize}px`,
        borderRadius: '50%',
        backgroundColor: '#FFD700',
        border: '2px solid #FFA500',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        cursor: isClickable ? 'pointer' : 'default',
        opacity: isClickable ? 1 : 0.6,
      }}
      title={character.name}
    >
      {character.mapIconUrl ? (
        <img
          src={character.mapIconUrl}
          alt={character.name}
          style={{
            width: `${pixelSize - 4}px`,
            height: `${pixelSize - 4}px`,
            borderRadius: '50%',
            objectFit: 'cover',
          }}
        />
      ) : (
        <span style={{ color: '#000', fontSize: `${pixelSize * 0.6}px` }}>👤</span>
      )}
    </div>
  );
};

interface UseCharacterMarkersReturn {
  markers: mapboxgl.Marker[];
  selectedCharacter: Character | null;
  setSelectedCharacter: (character: Character | null) => void;
}

export const useCharacterMarkers = (onCharacterClick?: (character: Character) => void): UseCharacterMarkersReturn => {
  const { map, zoom } = useMap();
  const { selectedZone } = useZoneContext();
  const { apiClient } = useAPI();
  const { location } = useLocation();
  const markersRef = useRef<mapboxgl.Marker[]>([]);
  const [characters, setCharacters] = useState<Character[]>([]);
  const [previousZoom, setPreviousZoom] = useState(0);
  const [previousSelectedZoneId, setPreviousSelectedZoneId] = useState<string | null>(null);
  const [previousCharactersCount, setPreviousCharactersCount] = useState(-1);
  const [previousLocationKey, setPreviousLocationKey] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedCharacter, setSelectedCharacter] = useState<Character | null>(null);

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

    // Check if user is within 10 meters
    let isClickable = false;
    if (location?.latitude && location?.longitude) {
      const userLocation = {
        lat: location.latitude,
        lng: location.longitude,
      };
      const characterLocation = {
        lat: character.movementPattern.startingLatitude,
        lng: character.movementPattern.startingLongitude,
      };
      const distance = calculateDistance(userLocation, characterLocation);
      isClickable = distance <= 10;
    }

    const handleClick = () => {
      if (isClickable && onCharacterClick) {
        onCharacterClick(character);
        setSelectedCharacter(character);
      }
    };

    createRoot(markerDiv).render(
      <CharacterMarker 
        character={character} 
        zoom={zoom} 
        onClick={handleClick}
        isClickable={isClickable}
      />
    );

    const marker = new mapboxgl.Marker(markerDiv)
      .setLngLat([character.movementPattern.startingLongitude, character.movementPattern.startingLatitude])
      .addTo(map.current!);

    return marker;
  }, [zoom, map, location, onCharacterClick]);

  const createMarkers = useCallback(() => {
    // Filter characters for the selected zone
    const zoneCharacters = selectedZone 
      ? characters.filter((character) => character.movementPattern.zoneId === selectedZone.id)
      : [];
    
    const zoneCharactersCount = zoneCharacters.length;
    // Round location to ~100m precision to avoid too frequent updates
    const locationKey = location?.latitude && location?.longitude 
      ? `${location.latitude.toFixed(3)},${location.longitude.toFixed(3)}` 
      : null;
    
    // Recreate markers if something significant has changed
    const shouldRecreate = 
      Math.abs(zoom - previousZoom) >= 1 || 
      previousSelectedZoneId !== selectedZone?.id ||
      previousCharactersCount !== zoneCharactersCount ||
      previousLocationKey !== locationKey; // Update when location changes

    if (!shouldRecreate) {
      return;
    }

    setPreviousZoom(zoom);
    setPreviousSelectedZoneId(selectedZone?.id ?? null);
    setPreviousCharactersCount(zoneCharactersCount);
    setPreviousLocationKey(locationKey);

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
  }, [selectedZone, characters, zoom, previousZoom, previousSelectedZoneId, previousCharactersCount, previousLocationKey, isLoading, createMarker, location]);

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

  return { 
    markers: markersRef.current,
    selectedCharacter,
    setSelectedCharacter,
  };
};

